package controller

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/openshift"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

func TestRetryVMotionPlan(t *testing.T) {
	testVMotionPlanCondition(t, true)
}

func TestBlockVMotionPlan(t *testing.T) {
	testVMotionPlanCondition(t, false)
}

func testVMotionPlanCondition(t *testing.T, retry bool) {
	t.Helper()
	reason, eventReason := migrationv1alpha1.ReasonFailed, "ConditionFailed"
	wantResult := ctrl.Result{RequeueAfter: 30 * time.Second}
	if retry {
		reason, eventReason = migrationv1alpha1.ReasonRetrying, migrationv1alpha1.ReasonRetrying
		wantResult = ctrl.Result{}
	}
	tests := []struct {
		name          string
		priorMessage  string
		priorGen      int64
		wantUnchanged bool
		failStatus    bool
	}{
		{name: "fresh condition"},
		{name: "identical condition", priorMessage: "inventory unavailable", priorGen: 2, wantUnchanged: true},
		{name: "changed message", priorMessage: "previous error", priorGen: 2},
		{name: "changed generation", priorMessage: "inventory unavailable", priorGen: 1},
		{name: "status write fails", failStatus: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			migration := newVMotionPlanMigration()
			migration.Generation = 2
			migration.Status.LastUpdateTime = &metav1.Time{Time: time.Unix(1, 0)}
			if tt.priorMessage != "" {
				condition := metav1.Condition{
					Type: migrationv1alpha1.ConditionInfrastructurePrepared, Status: metav1.ConditionFalse,
					Reason: reason, Message: tt.priorMessage, ObservedGeneration: tt.priorGen,
				}
				apimeta.SetStatusCondition(&migration.Status.Conditions, condition)
			}
			r, c := reconcileErrReconciler(t, migration)
			key := client.ObjectKeyFromObject(migration)
			if err := c.Get(ctx, key, migration); err != nil {
				t.Fatalf("getting migration: %v", err)
			}
			before := migration.DeepCopy()
			if tt.failStatus {
				r.Client = faultClient(c, interceptor.Funcs{
					SubResourceUpdate: func(context.Context, client.Client, string, client.Object, ...client.SubResourceUpdateOption) error {
						return errReconcileFault
					},
				})
			}
			originalErr := fmt.Errorf("inventory unavailable")
			var result ctrl.Result
			var err error
			var wantErr error
			if retry {
				wantErr = originalErr
				result, err = r.retryVMotionPlan(ctx, migration, before.Status, originalErr)
			} else {
				result, err = r.blockVMotionPlan(ctx, migration, before.Status, originalErr.Error())
			}
			if tt.failStatus {
				if !errors.Is(err, errReconcileFault) || result != (ctrl.Result{}) {
					t.Fatalf("status failure result = %+v, %v, want original status error and no requeue", result, err)
				}
			} else if err != wantErr || result != wantResult {
				t.Fatalf("result = %+v, %v, want %+v, %v", result, err, wantResult, wantErr)
			}
			got := &migrationv1alpha1.VmwareCloudFoundationMigration{}
			if err := c.Get(ctx, key, got); err != nil {
				t.Fatalf("getting persisted migration: %v", err)
			}
			if tt.wantUnchanged || tt.failStatus {
				if got.ResourceVersion != before.ResourceVersion || !reflect.DeepEqual(got.Status, before.Status) {
					t.Fatal("unchanged condition or failed write changed persisted status")
				}
			} else {
				condition := apimeta.FindStatusCondition(got.Status.Conditions, migrationv1alpha1.ConditionInfrastructurePrepared)
				if condition == nil || condition.Status != metav1.ConditionFalse || condition.Reason != reason ||
					condition.Message != "inventory unavailable" || condition.ObservedGeneration != 2 {
					t.Fatalf("persisted condition = %+v, want current failure condition", condition)
				}
			}
			recorder := r.Recorder.(*events.FakeRecorder)
			select {
			case event := <-recorder.Events:
				if tt.wantUnchanged || !strings.HasPrefix(event, "Warning "+eventReason+" ") || !strings.Contains(event, "inventory unavailable") {
					t.Fatalf("unexpected event %q", event)
				}
			default:
				if !tt.wantUnchanged {
					t.Fatal("missing Warning event")
				}
			}
			if len(recorder.Events) != 0 {
				t.Fatal("duplicate Warning event")
			}
		})
	}
}

func TestVMotionMachineAnnotationNodes(t *testing.T) {
	const machineRef = "openshift-machine-api/worker-0"
	tests := []struct {
		name        string
		annotations []map[string]string
		want        map[string]map[types.UID]bool
		wantErr     bool
	}{
		{name: "no annotation", annotations: []map[string]string{nil}, want: map[string]map[types.UID]bool{}},
		{name: "empty annotation", annotations: []map[string]string{{vmotionMachineAnnotation: ""}}, want: map[string]map[types.UID]bool{}},
		{name: "namespace and name", annotations: []map[string]string{{vmotionMachineAnnotation: machineRef}}, want: map[string]map[types.UID]bool{machineRef: {"uid-0": true}}},
		{name: "bare name", annotations: []map[string]string{{vmotionMachineAnnotation: "worker-0"}}, want: map[string]map[types.UID]bool{machineRef: {"uid-0": true}}},
		{name: "unknown Machine", annotations: []map[string]string{{vmotionMachineAnnotation: "missing"}}, want: map[string]map[types.UID]bool{}},
		{name: "shared Machine", annotations: []map[string]string{{vmotionMachineAnnotation: "worker-0"}, {vmotionMachineAnnotation: machineRef}}, want: map[string]map[types.UID]bool{machineRef: {"uid-0": true, "uid-1": true}}},
		{name: "missing namespace", annotations: []map[string]string{{vmotionMachineAnnotation: "/name"}}, wantErr: true},
		{name: "missing name", annotations: []map[string]string{{vmotionMachineAnnotation: "ns/"}}, wantErr: true},
		{name: "extra separator", annotations: []map[string]string{{vmotionMachineAnnotation: "ns/name/extra"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identities := make([]vmotionNodeIdentity, len(tt.annotations))
			for i, annotations := range tt.annotations {
				identities[i].node = corev1.Node{ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("node-%d", i), UID: types.UID(fmt.Sprintf("uid-%d", i)), Annotations: annotations,
				}}
			}
			got, err := vmotionMachineAnnotationNodes(identities, map[string]struct{}{machineRef: {}})
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), `resolving Machine annotation on Node "node-0"`) || got != nil {
					t.Fatalf("vmotionMachineAnnotationNodes = %+v, %v, want contextual error and no partial associations", got, err)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("vmotionMachineAnnotationNodes = %+v, %v, want %+v", got, err, tt.want)
			}
		})
	}
}

func TestSameFrozenNodeList(t *testing.T) {
	const changed = "changed"
	baseline := &migrationv1alpha1.RoleMigrationStatus{Nodes: []migrationv1alpha1.NodeMigrationProgress{{
		Name: "worker-a", NodeUID: "uid-a", ProviderID: "vsphere://11111111-1111-4111-8111-111111111111",
		VMBIOSUUID: "11111111-1111-4111-8111-111111111111", RequestedMode: migrationv1alpha1.VMotionModeHot,
		OriginalSchedulingState: migrationv1alpha1.NodeMigrationSchedulingStateSchedulable,
		TargetFailureDomain:     "target-fd", TargetResourcePool: "/TargetDC/host/Cluster/Resources",
		TargetDatastore: "/TargetDC/datastore/DS", TargetNetwork: "/TargetDC/network/Network",
		SourceInventoryPath: "/SourceDC/vm/worker-a",
	}}}
	second := baseline.Nodes[0]
	second.Name, second.NodeUID = "worker-b", "uid-b"
	baseline.Nodes = append(baseline.Nodes, second)
	tests := []struct {
		name      string
		existing  *migrationv1alpha1.RoleMigrationStatus
		candidate *migrationv1alpha1.RoleMigrationStatus
		want      bool
	}{
		{name: "both nil fail closed"},
		{name: "existing nil", candidate: baseline},
		{name: "candidate nil", existing: baseline},
		{name: "empty lists", existing: &migrationv1alpha1.RoleMigrationStatus{}, candidate: &migrationv1alpha1.RoleMigrationStatus{}, want: true},
		{name: "different lengths", existing: baseline, candidate: &migrationv1alpha1.RoleMigrationStatus{Nodes: baseline.Nodes[:1]}},
		{name: "identical lists", existing: baseline, candidate: baseline.DeepCopy(), want: true},
		{name: "reordered nodes", existing: baseline, candidate: &migrationv1alpha1.RoleMigrationStatus{Nodes: []migrationv1alpha1.NodeMigrationProgress{second, baseline.Nodes[0]}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sameFrozenNodeList(tt.existing, tt.candidate); got != tt.want {
				t.Fatalf("sameFrozenNodeList = %t, want %t", got, tt.want)
			}
		})
	}
	changes := []struct {
		name   string
		mutate func(*migrationv1alpha1.NodeMigrationProgress)
	}{
		{name: "Name", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.Name = changed }},
		{name: "NodeUID", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.NodeUID = changed }},
		{name: "ProviderID", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.ProviderID = changed }},
		{name: "VMBIOSUUID", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.VMBIOSUUID = changed }},
		{name: "RequestedMode", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.RequestedMode = migrationv1alpha1.VMotionModeCold }},
		{name: "OriginalSchedulingState", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) {
			n.OriginalSchedulingState = migrationv1alpha1.NodeMigrationSchedulingStateUnschedulable
		}},
		{name: "TargetFailureDomain", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.TargetFailureDomain = changed }},
		{name: "TargetResourcePool", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.TargetResourcePool = changed }},
		{name: "TargetDatastore", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.TargetDatastore = changed }},
		{name: "TargetNetwork", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.TargetNetwork = changed }},
		{name: "SourceInventoryPath", mutate: func(n *migrationv1alpha1.NodeMigrationProgress) { n.SourceInventoryPath = changed }},
	}
	for _, tt := range changes {
		t.Run(tt.name, func(t *testing.T) {
			candidate := baseline.DeepCopy()
			tt.mutate(&candidate.Nodes[1])
			if sameFrozenNodeList(baseline, candidate) {
				t.Fatal("changed frozen node field was accepted")
			}
		})
	}
}

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
		{value: "namespace/machine.a", want: "namespace/machine.a"},
		{value: ""},
		{value: "/name"},
		{value: "ns/"},
		{value: "ns/name/extra"},
		{value: "namespace/bad name"},
		{value: "bad namespace/machine-a"},
		{value: "namespace/UPPERCASE"},
		{value: "namespace/-machine"},
		{value: "dotted.namespace/machine-a"},
		{value: strings.Repeat("a", 64) + "/machine-a"},
		{value: "namespace/" + strings.Repeat("a", 254)},
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
