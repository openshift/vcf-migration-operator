package controller

import (
	"context"
	"reflect"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/openshift"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

func TestVMotionPlanFingerprintIgnoresInputOrder(t *testing.T) {
	frozen := migrationv1alpha1.NodeMigrationFrozenPlan{
		SourceVCenter: cleanupTestSource, SourceDatacenter: vmotionPlanTestSourceDatacenter,
		DestinationVCenter: "target.example.com", Mode: migrationv1alpha1.VMotionModeHot,
	}
	mappings := []migrationv1alpha1.FailureDomainMapping{
		{Source: "source-b", Destination: "target-b"},
		{Source: "source-a", Destination: "target-a"},
	}
	nodes := []vmotionFingerprintNode{
		{Name: "worker-b", Role: vmotionWorkerRole, UID: "uid-b", TargetFailureDomain: "target-b"},
		{Name: "worker-a", Role: vmotionWorkerRole, UID: "uid-a", TargetFailureDomain: "target-a"},
		{Name: "control-plane-a", Role: vmotionControlPlaneRole, UID: "uid-c", TargetFailureDomain: "target-a"},
	}
	first, err := vmotionPlanFingerprint("migration-uid", frozen, mappings, nodes)
	if err != nil {
		t.Fatalf("vmotionPlanFingerprint: %v", err)
	}
	reorderedMappings := []migrationv1alpha1.FailureDomainMapping{mappings[1], mappings[0]}
	reorderedNodes := []vmotionFingerprintNode{nodes[2], nodes[1], nodes[0]}
	second, err := vmotionPlanFingerprint("migration-uid", frozen, reorderedMappings, reorderedNodes)
	if err != nil {
		t.Fatalf("vmotionPlanFingerprint reordered: %v", err)
	}
	if first != second {
		t.Fatalf("reordered fingerprint = %q, want %q", second, first)
	}
	reorderedMappings[0].Destination = "target-b"
	changed, err := vmotionPlanFingerprint("migration-uid", frozen, reorderedMappings, reorderedNodes)
	if err != nil {
		t.Fatalf("vmotionPlanFingerprint changed mapping: %v", err)
	}
	if changed == first {
		t.Fatal("changing a mapping did not change the fingerprint")
	}
}

func TestParseVMotionMachineReference(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "machine-a", want: "openshift-machine-api/machine-a"},
		{value: "namespace/machine-a", want: "namespace/machine-a"},
		{value: ""},
		{value: "/name"},
		{value: "ns/"},
		{value: "ns/name/extra"},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := parseVMotionMachineReference(tt.value)
			if (err != nil) != (tt.want == "") || got != tt.want {
				t.Fatalf("parseVMotionMachineReference(%q) = %q, %v, want %q", tt.value, got, err, tt.want)
			}
		})
	}
}

func TestMergeInitialNodeMigrationPlanPreservesFirstWriter(t *testing.T) {
	migration := newVMotionPlanMigration()
	latest := migration.DeepCopy()
	baseStatus := *migration.Status.DeepCopy()
	migration.Status.NodeMigration = &migrationv1alpha1.NodeMigrationStatus{
		ExecutionID: "first-execution", MigrationUID: migration.UID,
		FrozenPlan: &migrationv1alpha1.NodeMigrationFrozenPlan{SourceVCenter: cleanupTestSource},
	}
	if !mergeInitialNodeMigrationPlan(migration, latest, baseStatus) {
		t.Fatal("first plan publication was rejected")
	}
	want := latest.Status.NodeMigration.DeepCopy()
	migration.Status.NodeMigration.ExecutionID = "competing-execution"
	migration.Status.NodeMigration.FrozenPlan.SourceVCenter = "other.example.com"
	if mergeInitialNodeMigrationPlan(migration, latest, baseStatus) {
		t.Fatal("competing plan overwrote the first plan")
	}
	if !reflect.DeepEqual(latest.Status.NodeMigration, want) {
		t.Fatalf("committed plan = %+v, want first writer %+v", latest.Status.NodeMigration, want)
	}
	for _, field := range []string{"UID", "generation"} {
		t.Run(field, func(t *testing.T) {
			stale := latest.DeepCopy()
			stale.Status.NodeMigration = nil
			if field == "UID" {
				stale.UID = "replacement-uid"
			} else {
				stale.Generation++
			}
			if mergeInitialNodeMigrationPlan(migration, stale, baseStatus) || stale.Status.NodeMigration != nil {
				t.Fatal("stale reconcile published a plan")
			}
		})
	}
}

func TestVMotionReconcileFreezesDistinctFailureDomainMappings(t *testing.T) {
	h := newVMotionPlanHarness(t)
	ctx := context.Background()
	infrastructure, err := h.reconciler.ConfigClient.ConfigV1().Infrastructures().Get(ctx, openshift.InfrastructureName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting Infrastructure: %v", err)
	}
	secondSource := infrastructure.Spec.PlatformSpec.VSphere.FailureDomains[0]
	secondSource.Name = "source-fd-2"
	secondSource.Topology.ComputeCluster = "/SourceDC/host/SecondCluster"
	secondSource.Topology.ResourcePool = "/SourceDC/host/SecondCluster/Resources"
	infrastructure.Spec.PlatformSpec.VSphere.FailureDomains = append(infrastructure.Spec.PlatformSpec.VSphere.FailureDomains, secondSource)
	if _, err := h.reconciler.ConfigClient.ConfigV1().Infrastructures().Update(ctx, infrastructure, metav1.UpdateOptions{}); err != nil {
		t.Fatalf("updating Infrastructure: %v", err)
	}
	secondTarget := *h.migration.Spec.FailureDomains[0].DeepCopy()
	secondTarget.Name = vmotionPlanTestSecondTargetFD
	secondTarget.Topology.ComputeCluster = "/TargetDC/host/SecondCluster"
	secondTarget.Topology.ResourcePool = "/TargetDC/host/SecondCluster/Resources"
	secondTarget.Topology.Datastore = "/TargetDC/datastore/SecondDS"
	secondTarget.Topology.Networks = []string{"/TargetDC/network/SecondNetwork"}
	h.migration.Spec.FailureDomains = append(h.migration.Spec.FailureDomains, secondTarget)
	h.migration.Spec.FailureDomainMappings = append(h.migration.Spec.FailureDomainMappings,
		migrationv1alpha1.FailureDomainMapping{Source: "source-fd-2", Destination: vmotionPlanTestSecondTargetFD})
	if err := h.apiClient.Update(ctx, h.migration); err != nil {
		t.Fatalf("updating migration: %v", err)
	}
	inventoryFunc := h.reconciler.vmotionInventoryFunc
	h.reconciler.vmotionInventoryFunc = func(ctx context.Context, sourceVC *configv1.VSpherePlatformVCenterSpec, sourceFDs []configv1.VSpherePlatformFailureDomainSpec, ids []string) (vsphere.VMotionInventory, error) {
		inventory, err := inventoryFunc(ctx, sourceVC, sourceFDs, ids)
		secondOwner := h.sourceFDReference
		secondOwner.Value = "domain-second"
		inventory.SourceFailureDomainReferences["source-fd-2"] = secondOwner
		for i := range inventory.VMs {
			if inventory.VMs[i].BIOSUUID == "55555555-5555-4555-8555-555555555555" {
				inventory.VMs[i].ResourcePoolOwner = secondOwner
				inventory.VMs[i].ResourcePoolRef.Value = "resgroup-2"
			}
		}
		return inventory, err
	}
	h.reconcileUntilFrozen(t)
	workers := h.migration.Status.NodeMigration.Workers.Nodes
	if len(workers) != 2 || workers[0].TargetFailureDomain != "target-fd" || workers[1].TargetFailureDomain != vmotionPlanTestSecondTargetFD ||
		workers[1].TargetResourcePool != "/TargetDC/host/SecondCluster/Resources" ||
		workers[1].TargetDatastore != "/TargetDC/datastore/SecondDS" || workers[1].TargetNetwork != "/TargetDC/network/SecondNetwork" {
		t.Fatalf("worker placements = %+v, want distinct mapped destinations", workers)
	}
	if !reflect.DeepEqual(h.migration.Status.FailureDomainMappings, h.migration.Spec.FailureDomainMappings) {
		t.Fatalf("effective mappings = %+v, want supplied mappings", h.migration.Status.FailureDomainMappings)
	}
}
