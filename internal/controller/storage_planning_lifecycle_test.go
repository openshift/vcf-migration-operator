package controller

import (
	"context"
	"fmt"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/controller/storage"
	"github.com/openshift/vcf-migration-operator/internal/util"
)

// addPlannedVolume adds a bound vSphere PV and PVC in namespace to the fixture.
func addPlannedVolume(t *testing.T, r *VmwareCloudFoundationMigrationReconciler, namespace string) {
	t.Helper()
	ctx := context.Background()
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{Name: namespace + "-pv", UID: types.UID("pv-uid-" + namespace)},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{Driver: vsphereCSIDriverName, VolumeHandle: namespace + "-disk"},
			},
			ClaimRef: &corev1.ObjectReference{Namespace: namespace, Name: "data", UID: types.UID("pvc-uid-" + namespace)},
		},
		Status: corev1.PersistentVolumeStatus{Phase: corev1.VolumeBound},
	}
	if _, err := r.KubeClient.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{}); err != nil {
		t.Fatalf("creating PV fixture: %v", err)
	}
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "data", Namespace: namespace, UID: pv.Spec.ClaimRef.UID},
		Spec:       corev1.PersistentVolumeClaimSpec{VolumeName: pv.Name, StorageClassName: ptr.To("thin-csi")},
	}
	if _, err := r.KubeClient.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{}); err != nil {
		t.Fatalf("creating PVC fixture: %v", err)
	}
}

func listStoragePlans(t *testing.T, r *VmwareCloudFoundationMigrationReconciler) []migrationv1alpha1.StorageWorkloadMigration {
	t.Helper()
	var list migrationv1alpha1.StorageWorkloadMigrationList
	if err := r.List(context.Background(), &list); err != nil {
		t.Fatalf("listing SWMs: %v", err)
	}
	return list.Items
}

func TestStoragePlanNamespaceWriteFailureBlocksOnlyThatNamespace(t *testing.T) {
	faults := interceptor.Funcs{
		Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
			if obj.GetNamespace() != "dying" {
				return c.Create(ctx, obj, opts...)
			}
			err := apierrors.NewForbidden(schema.GroupResource{Group: migrationv1alpha1.GroupVersion.Group, Resource: "storageworkloadmigrations"}, obj.GetName(),
				fmt.Errorf("unable to create new content in namespace dying because it is being terminated"))
			err.ErrStatus.Details.Causes = append(err.ErrStatus.Details.Causes, metav1.StatusCause{Type: corev1.NamespaceTerminatingCause, Field: "metadata.namespace"})
			return err
		},
	}
	r, m := planningFixture(t, func(r *VmwareCloudFoundationMigrationReconciler, _ *migrationv1alpha1.VmwareCloudFoundationMigration) {
		r.Client = interceptor.NewClient(r.Client.(client.WithWatch), faults)
	})
	addPlannedVolume(t, r, "dying")

	got := reconcilePlan(t, r, m)
	requirePlanCondition(t, got, migrationv1alpha1.ConditionStoragePlanned, migrationv1alpha1.ReasonPlanBlocked, metav1.ConditionFalse)
	condition := apimeta.FindStatusCondition(got.Status.Conditions, migrationv1alpha1.ConditionStoragePlanned)
	if !strings.Contains(condition.Message, "NamespaceTerminating: namespace dying") || !strings.Contains(condition.Message, "blocked in 1 namespaces") {
		t.Fatalf("StoragePlanned message = %q, want a NamespaceTerminating blocker for one namespace", condition.Message)
	}
	plans := listStoragePlans(t, r)
	if len(plans) != 1 || plans[0].Namespace != "app" || len(plans[0].Status.Volumes) != 1 {
		t.Fatalf("SWMs = %+v, want the healthy namespace planned", plans)
	}
}

func TestStoragePlanRecordsDeletedOwner(t *testing.T) {
	r, m := planningFixture(t)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app-pod", Namespace: "app", UID: "pod-uid",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: "gone", UID: "gone-uid", Controller: ptr.To(true)}},
		},
		Spec: corev1.PodSpec{Volumes: []corev1.Volume{{VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "data"},
		}}}},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	if _, err := r.KubeClient.CoreV1().Pods("app").Create(context.Background(), pod, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}

	got := reconcilePlan(t, r, m)
	requirePlanCondition(t, got, migrationv1alpha1.ConditionStoragePlanned, migrationv1alpha1.ReasonCompleted, metav1.ConditionTrue)
	plans := listStoragePlans(t, r)
	if len(plans) != 1 || len(plans[0].Status.Consumers) != 1 || plans[0].Status.Consumers[0].UID != "gone-uid" {
		t.Fatalf("SWMs = %+v, want the deleted ReplicaSet recorded as the consumer", plans)
	}
}

func TestStoragePlansReleasedWhenStorageDisabled(t *testing.T) {
	r, m := planningFixture(t)
	m = reconcilePlan(t, r, m)
	if len(listStoragePlans(t, r)) != 1 {
		t.Fatal("planning did not write a SWM")
	}
	m.Spec.MigrateStorage = false
	m.Spec.Storage = nil
	if _, err := r.StorageController.EnsureStoragePlanned(context.Background(), m); err != nil {
		t.Fatalf("EnsureStoragePlanned: %v", err)
	}
	requirePlanCondition(t, m, migrationv1alpha1.ConditionStoragePlanned, migrationv1alpha1.ReasonNotConfigured, metav1.ConditionTrue)
	if plans := listStoragePlans(t, r); len(plans) != 0 {
		t.Fatalf("SWMs = %+v, want none after storage migration was disabled", plans)
	}
}

func TestStoragePlansFromPredecessorAreDeleted(t *testing.T) {
	r, m := planningFixture(t)
	ctx := context.Background()
	predecessor := &migrationv1alpha1.StorageWorkloadMigration{
		ObjectMeta: metav1.ObjectMeta{Name: "vcf-storage-old", Namespace: "old", Labels: map[string]string{storage.MigrationUIDLabel: "old-uid"}},
		Spec:       migrationv1alpha1.StorageWorkloadMigrationSpec{MigrationRef: migrationv1alpha1.MigrationObjectReference{Name: m.Name, UID: "old-uid"}},
	}
	foreign := &migrationv1alpha1.StorageWorkloadMigration{
		ObjectMeta: metav1.ObjectMeta{Name: "foreign", Namespace: "old", Labels: map[string]string{storage.MigrationUIDLabel: "other-uid"}},
		Spec:       migrationv1alpha1.StorageWorkloadMigrationSpec{MigrationRef: migrationv1alpha1.MigrationObjectReference{Name: m.Name, UID: "old-uid"}},
	}
	for _, swm := range []*migrationv1alpha1.StorageWorkloadMigration{predecessor, foreign} {
		if err := r.Create(ctx, swm); err != nil {
			t.Fatal(err)
		}
	}

	reconcilePlan(t, r, m)
	if err := r.Get(ctx, client.ObjectKeyFromObject(predecessor), predecessor); !apierrors.IsNotFound(err) {
		t.Fatalf("predecessor SWM: %v, want deleted", err)
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(foreign), foreign); err != nil {
		t.Fatalf("SWM whose label and reference disagree was deleted: %v", err)
	}
}

func TestStoragePlansReleasedWithMigrationRecord(t *testing.T) {
	r, m := planningFixture(t)
	ctx := context.Background()
	m = reconcilePlan(t, r, m)
	if len(listStoragePlans(t, r)) != 1 {
		t.Fatal("planning did not write a SWM")
	}
	m.Annotations = map[string]string{migrationv1alpha1.AnnotationAllowDeletion: "true"}
	if err := r.Update(ctx, m); err != nil {
		t.Fatal(err)
	}
	if err := r.Delete(ctx, m); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(m)}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	if plans := listStoragePlans(t, r); len(plans) != 0 {
		t.Fatalf("SWMs = %+v, want none after the migration record was released", plans)
	}
}

// completeMigration records a completed compute-only migration at generation 1
// and returns it with spec changes applied at generation 2.
func completeMigration(t *testing.T, r *VmwareCloudFoundationMigrationReconciler, m *migrationv1alpha1.VmwareCloudFoundationMigration, change func(*migrationv1alpha1.VmwareCloudFoundationMigration)) *migrationv1alpha1.VmwareCloudFoundationMigration {
	t.Helper()
	ctx := context.Background()
	m.Spec.State = migrationv1alpha1.MigrationStateRunning
	m.Spec.MigrateStorage = false
	m.Spec.Storage = nil
	if err := r.Update(ctx, m); err != nil {
		t.Fatal(err)
	}
	for _, condition := range conditionOrder {
		util.SetCondition(m, condition, metav1.ConditionTrue, migrationv1alpha1.ReasonCompleted, "done")
	}
	now := metav1.Now()
	m.Status.StartTime = &now
	m.Status.CompletionTime = &now
	if err := r.Status().Update(ctx, m); err != nil {
		t.Fatal(err)
	}
	change(m)
	m.Generation++
	if err := r.Update(ctx, m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestCompletedMigrationRejectsEnabledStorage(t *testing.T) {
	r, m := planningFixture(t)
	m = completeMigration(t, r, m, func(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
		m.Spec.MigrateStorage = true
	})
	got := reconcilePlan(t, r, m)
	requirePlanCondition(t, got, migrationv1alpha1.ConditionReady, migrationv1alpha1.ReasonStorageExecutionUnsupported, metav1.ConditionFalse)
}

func TestCompletedMigrationReportsPlanningAfterStart(t *testing.T) {
	r, m := planningFixture(t)
	m = completeMigration(t, r, m, func(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
		m.Spec.State = migrationv1alpha1.MigrationStatePlanning
	})
	got := reconcilePlan(t, r, m)
	requirePlanCondition(t, got, migrationv1alpha1.ConditionReady, migrationv1alpha1.ReasonPlanningAfterStart, metav1.ConditionFalse)
}

func TestStatusKeepsCompletedReadyOverSameGenerationRejection(t *testing.T) {
	existing := metav1.Condition{Type: migrationv1alpha1.ConditionReady, Status: metav1.ConditionTrue, ObservedGeneration: 2}
	for _, reason := range []string{migrationv1alpha1.ReasonPlanningAfterStart, migrationv1alpha1.ReasonStorageExecutionUnsupported} {
		sameGeneration := metav1.Condition{Type: migrationv1alpha1.ConditionReady, Status: metav1.ConditionFalse, Reason: reason, ObservedGeneration: 2}
		if isReadyRejectionOfNewerSpec(sameGeneration, existing) {
			t.Fatalf("isReadyRejectionOfNewerSpec(%s, same generation) = true, want false", reason)
		}
	}
	paused := metav1.Condition{Type: migrationv1alpha1.ConditionReady, Status: metav1.ConditionFalse, Reason: migrationv1alpha1.ReasonPaused, ObservedGeneration: 3}
	if isReadyRejectionOfNewerSpec(paused, existing) {
		t.Fatal("isReadyRejectionOfNewerSpec(Paused) = true, want false")
	}
}
