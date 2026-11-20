package storage

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/execution"
	"github.com/openshift/vcf-migration-operator/internal/storageplan"
	"github.com/openshift/vcf-migration-operator/internal/util"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

const (
	// PlanningResync is the periodic requeue interval while waiting on planning or blocked storage plans.
	PlanningResync = 5 * time.Minute
	// MigrationUIDLabel labels operator-managed StorageWorkloadMigration resources with the migration UID.
	MigrationUIDLabel = "migration.openshift.io/migration-uid"

	// maxStorageBlockerMessageBytes matches the API limit on blocker messages.
	maxStorageBlockerMessageBytes = 1024

	policyCheckTimeout = 2 * time.Minute

	managementStateManaged   = "Managed"
	managementStateRemoved   = "Removed"
	managementStateUnmanaged = "Unmanaged"
)

var (
	// ClusterCSIDriverGVR identifies OpenShift ClusterCSIDriver resources.
	ClusterCSIDriverGVR = schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "clustercsidrivers"}
	// StorageOperatorGVR identifies OpenShift Storage operator resources.
	StorageOperatorGVR = schema.GroupVersionResource{Group: "operator.openshift.io", Version: "v1", Resource: "storages"}
)

type credentials struct {
	username string
	password string
}

// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch
// +kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=replicasets;statefulsets;daemonsets,verbs=get
// +kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get
// +kubebuilder:rbac:groups=migration.openshift.io,resources=storageworkloadmigrations,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=migration.openshift.io,resources=storageworkloadmigrations/status,verbs=get;update;patch

// BlockUnsupportedStorageMigration keeps the guard at the execution boundary
// until the complete storage workflow is implemented, preserving valid plans.
func (s *StorageController) BlockUnsupportedStorageMigration(migration *migrationv1alpha1.VmwareCloudFoundationMigration) bool {
	if !migration.Spec.MigrateStorage {
		return false
	}
	message := "Storage migration execution is not implemented; only non-disruptive storage planning is available"
	changed := util.SetCondition(migration, migrationv1alpha1.ConditionReady, metav1.ConditionFalse, migrationv1alpha1.ReasonStorageExecutionUnsupported, message)
	if changed && s.Recorder != nil {
		s.Recorder.Eventf(migration, nil, "Warning", migrationv1alpha1.ReasonStorageExecutionUnsupported, migrationv1alpha1.ConditionReady, "%s", message)
	}
	return true
}

// CheckStorageManagement validates vSphere CSI driver and Storage operator
// management states for preflight.
func (s *StorageController) CheckStorageManagement(ctx context.Context, migrateStorage bool) (string, error) {
	if !migrateStorage {
		return checkVSphereStorageManagement(ctx, s.DynamicClient)
	}
	state, err := getManagementState(ctx, s.DynamicClient, ClusterCSIDriverGVR, storageplan.Driver)
	if err != nil {
		return "", fmt.Errorf("reading vSphere CSI management state: %w", err)
	}
	if state != managementStateManaged {
		return "", fmt.Errorf("configuring storage migration requires ClusterCSIDriver/%s spec.managementState Managed (current: %s)", storageplan.Driver, state)
	}
	return "", nil
}

func checkVSphereStorageManagement(ctx context.Context, dynamicClient dynamic.Interface) (string, error) {
	log := klog.FromContext(ctx)

	csiState, err := getManagementState(ctx, dynamicClient, ClusterCSIDriverGVR, storageplan.Driver)
	if err != nil {
		return "", fmt.Errorf("vSphere CSI driver is not supported for migration; unable to read ClusterCSIDriver/%s: %w", storageplan.Driver, err)
	}
	log.V(1).Info("checked vSphere CSI driver management state", "clusterCSIDriver", storageplan.Driver, "managementState", csiState)
	if csiState != managementStateRemoved {
		return "", fmt.Errorf("vSphere CSI driver is not supported for migration; set ClusterCSIDriver/%s spec.managementState to Removed (current: %s)", storageplan.Driver, csiState)
	}

	storageState, err := getManagementState(ctx, dynamicClient, StorageOperatorGVR, "cluster")
	if err != nil {
		log.V(1).Info("unable to read Storage/cluster management state", "error", err)
		return "unable to read Storage/cluster managementState; consider setting Unmanaged or Removed (only ClusterCSIDriver Removed is required)", nil
	}
	log.V(1).Info("checked Storage operator management state", "storage", "cluster", "managementState", storageState)
	switch storageState {
	case managementStateUnmanaged, managementStateRemoved:
		return "", nil
	default:
		return fmt.Sprintf("Storage/cluster managementState is %s; consider setting Unmanaged or Removed (only ClusterCSIDriver Removed is required)", storageState), nil
	}
}

func getManagementState(ctx context.Context, dynamicClient dynamic.Interface, gvr schema.GroupVersionResource, name string) (string, error) {
	obj, err := dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	state, found, err := unstructured.NestedString(obj.Object, "spec", "managementState")
	if err != nil {
		return "", fmt.Errorf("reading spec.managementState: %w", err)
	}
	if !found || state == "" {
		return "", fmt.Errorf("spec.managementState not set")
	}
	return state, nil
}

// mapPVCToStoragePlanning enqueues the singleton migration for a PVC event
// only while a storage plan is being refreshed: storage migration is enabled,
// the migration is planning or running, and execution has not started.
// The watch handler requires the event object, but only the singleton matters.
func (s *StorageController) mapPVCToStoragePlanning(ctx context.Context, _ client.Object) []reconcile.Request {
	key := types.NamespacedName{Name: migrationv1alpha1.SingletonName, Namespace: migrationv1alpha1.SingletonNamespace}
	request := []reconcile.Request{{NamespacedName: key}}

	migration := &migrationv1alpha1.VmwareCloudFoundationMigration{}
	if err := s.Client.Get(ctx, key, migration); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		// Enqueue on read errors; Reconcile reports the failure.
		return request
	}
	if !s.ShouldRefreshPlan(migration) {
		return nil
	}
	return request
}

// ShouldRefreshPlan reports whether PVC changes can affect the storage plan.
func (s *StorageController) ShouldRefreshPlan(m *migrationv1alpha1.VmwareCloudFoundationMigration) bool {
	if !m.Spec.MigrateStorage {
		return false
	}
	if m.Status.StartTime != nil {
		return false
	}
	return m.Spec.State == migrationv1alpha1.MigrationStatePlanning || m.Spec.State == migrationv1alpha1.MigrationStateRunning
}

// pvcMetadataChangedPredicate ignores PVC updates that only change server-managed
// metadata such as resourceVersion or managedFields. Generation alone is not
// enough: the target-failure-domain override is an annotation.
func pvcMetadataChangedPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == nil || e.ObjectNew == nil {
				return true
			}
			return isPVCMetadataChanged(e.ObjectOld, e.ObjectNew)
		},
	}
}

// isPVCMetadataChanged compares the PVC metadata that storage planning reads.
func isPVCMetadataChanged(old, updated client.Object) bool {
	if old.GetGeneration() != updated.GetGeneration() {
		return true
	}
	if !reflect.DeepEqual(old.GetAnnotations(), updated.GetAnnotations()) {
		return true
	}
	if !reflect.DeepEqual(old.GetOwnerReferences(), updated.GetOwnerReferences()) {
		return true
	}
	return !reflect.DeepEqual(old.GetDeletionTimestamp(), updated.GetDeletionTimestamp())
}

// storagePlanSummary aggregates volume counts across all planned namespaces.
type storagePlanSummary struct {
	namespaces        int
	volumes           int
	blockedVolumes    int
	blockedNamespaces int
}

// EnsureStoragePlanned computes and persists per-namespace storage migration plans
// and updates ConditionStoragePlanned on m.
func (s *StorageController) EnsureStoragePlanned(ctx context.Context, m *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error) {
	var previous *metav1.Condition
	if condition := apimeta.FindStatusCondition(m.Status.Conditions, migrationv1alpha1.ConditionStoragePlanned); condition != nil {
		previous = condition.DeepCopy()
	}

	defer s.recordStoragePlannedEvent(m, previous)

	if !m.Spec.MigrateStorage {
		// Plans written while storage migration was enabled no longer apply.
		if err := s.DeleteStaleStoragePlans(ctx, m, nil /* plan */); err != nil {
			return ctrl.Result{}, err
		}
		util.SetCondition(m, migrationv1alpha1.ConditionStoragePlanned, metav1.ConditionTrue, migrationv1alpha1.ReasonNotConfigured, "Storage migration is not enabled")
		s.clearStoragePlanBlocked(m)
		return ctrl.Result{}, nil
	}

	in, err := s.collectStorageInventory(ctx, m)
	if err != nil {
		return ctrl.Result{}, err
	}

	plan := storageplan.Build(in)
	if err := s.checkPlanPolicies(ctx, m, in.Infrastructure, plan, in.StorageClasses, in.PVCs); err != nil {
		return ctrl.Result{}, err
	}

	summary, err := s.writeNamespacePlans(ctx, m, plan, in.Pods)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := s.DeleteStaleStoragePlans(ctx, m, plan); err != nil {
		return ctrl.Result{}, err
	}

	if summary.blockedVolumes > 0 || len(plan.Blockers) > 0 {
		message := storagePlanBlockedMessage(summary, plan.Blockers)
		util.SetCondition(m, migrationv1alpha1.ConditionStoragePlanned, metav1.ConditionFalse, migrationv1alpha1.ReasonPlanBlocked, message)
		util.SetCondition(m, migrationv1alpha1.ConditionReady, metav1.ConditionFalse, migrationv1alpha1.ReasonPlanBlocked, message)
		return ctrl.Result{RequeueAfter: PlanningResync}, nil
	}
	util.SetCondition(m, migrationv1alpha1.ConditionStoragePlanned, metav1.ConditionTrue, migrationv1alpha1.ReasonCompleted, fmt.Sprintf("Planned %d volumes in %d namespaces", summary.volumes, summary.namespaces))
	s.clearStoragePlanBlocked(m)
	return ctrl.Result{}, nil
}

// clearStoragePlanBlocked releases a Running migration's Ready condition from
// PlanBlocked once storage planning passes. reconcileWorkflow keeps that
// reason in place so a blocked plan does not flip Ready on every reconcile.
func (s *StorageController) clearStoragePlanBlocked(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
	if m.Spec.State != migrationv1alpha1.MigrationStateRunning {
		return
	}
	ready := apimeta.FindStatusCondition(m.Status.Conditions, migrationv1alpha1.ConditionReady)
	if ready == nil || ready.Reason != migrationv1alpha1.ReasonPlanBlocked {
		return
	}
	util.SetReadyRunning(m, s.Recorder)
}

// recordStoragePlannedEvent emits an event when the StoragePlanned condition
// was added or changed relative to previous.
func (s *StorageController) recordStoragePlannedEvent(m *migrationv1alpha1.VmwareCloudFoundationMigration, previous *metav1.Condition) {
	if s.Recorder == nil {
		return
	}
	condition := apimeta.FindStatusCondition(m.Status.Conditions, migrationv1alpha1.ConditionStoragePlanned)
	if condition == nil {
		return
	}
	if previous != nil && !util.StatusConditionChanged(*previous, *condition) {
		return
	}
	eventType := "Normal"
	if condition.Status == metav1.ConditionFalse {
		eventType = "Warning"
	}
	s.Recorder.Eventf(m, nil, eventType, condition.Reason, migrationv1alpha1.ConditionStoragePlanned, "%s", util.TruncateToBytes(condition.Message, util.MaxEventNoteBytes))
}

// collectStorageInventory lists the cluster resources needed to build a storage plan.
func (s *StorageController) collectStorageInventory(ctx context.Context, m *migrationv1alpha1.VmwareCloudFoundationMigration) (storageplan.Input, error) {
	pvs, err := s.KubeClient.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return storageplan.Input{}, fmt.Errorf("listing PV inventory: %w", err)
	}
	pvcs, err := s.KubeClient.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return storageplan.Input{}, fmt.Errorf("listing PVC inventory: %w", err)
	}
	pods, err := s.KubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return storageplan.Input{}, fmt.Errorf("listing storage consumers: %w", err)
	}
	classes, err := s.KubeClient.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return storageplan.Input{}, fmt.Errorf("listing StorageClasses: %w", err)
	}
	infra, err := s.ConfigClient.ConfigV1().Infrastructures().Get(ctx, "cluster", metav1.GetOptions{})
	if err != nil {
		return storageplan.Input{}, fmt.Errorf("getting infrastructure for storage planning: %w", err)
	}
	categories, err := s.getCSITopologyCategories(ctx)
	if err != nil {
		return storageplan.Input{}, err
	}
	return storageplan.Input{
		Migration:          m,
		Infrastructure:     infra,
		PVs:                pvs.Items,
		PVCs:               pvcs.Items,
		Pods:               pods.Items,
		StorageClasses:     classes.Items,
		TopologyCategories: categories,
	}, nil
}

// getCSITopologyCategories returns the topology categories configured on the
// vSphere ClusterCSIDriver.
func (s *StorageController) getCSITopologyCategories(ctx context.Context) ([]string, error) {
	csi, err := s.DynamicClient.Resource(ClusterCSIDriverGVR).Get(ctx, storageplan.Driver, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting CSI topology categories: %w", err)
	}
	categories, _, err := unstructured.NestedStringSlice(csi.Object, "spec", "driverConfig", "vSphere", "topologyCategories")
	if err != nil {
		return nil, fmt.Errorf("reading CSI topology categories: %w", err)
	}
	return categories, nil
}

// writeNamespacePlans resolves storage consumers for each planned namespace,
// writes its StorageWorkloadMigration, and summarizes volume and blocker counts.
// Namespaces are processed in sorted order so writes are deterministic. A
// failure in one namespace becomes a plan blocker so the other namespaces are
// still planned; only a lost execution permission aborts the pass.
func (s *StorageController) writeNamespacePlans(
	ctx context.Context,
	m *migrationv1alpha1.VmwareCloudFoundationMigration,
	plan *storageplan.Plan,
	pods []corev1.Pod,
) (storagePlanSummary, error) {
	namespaceNames := make([]string, 0, len(plan.NamespaceMigrationPlanInstances))
	for ns := range plan.NamespaceMigrationPlanInstances {
		namespaceNames = append(namespaceNames, ns)
	}
	sort.Strings(namespaceNames)

	log := klog.FromContext(ctx)

	podsByNamespace := map[string][]corev1.Pod{}
	for i := range pods {
		if plan.NamespaceMigrationPlanInstances[pods[i].Namespace] == nil {
			continue
		}
		podsByNamespace[pods[i].Namespace] = append(podsByNamespace[pods[i].Namespace], pods[i])
	}

	summary := storagePlanSummary{namespaces: len(namespaceNames)}
	for _, namespace := range namespaceNames {
		nsPlan := plan.NamespaceMigrationPlanInstances[namespace]
		blocked := countBlockedVolumes(nsPlan.Volumes)
		summary.volumes += len(nsPlan.Volumes)
		summary.blockedVolumes += blocked

		reason, err := s.writeNamespacePlan(ctx, m, namespace, nsPlan, podsByNamespace[namespace], blocked)
		if err != nil {
			if checkErr := execution.Check(ctx); checkErr != nil {
				return storagePlanSummary{}, checkErr
			}
			log.Error(err, "storage planning blocked for namespace", "namespace", namespace, "reason", reason)
			plan.Blockers = append(plan.Blockers, migrationv1alpha1.StorageWorkloadMigrationBlocker{
				Reason:  reason,
				Message: util.TruncateToBytes(fmt.Sprintf("namespace %s: %v", namespace, err), maxStorageBlockerMessageBytes),
			})
			summary.blockedNamespaces++
			continue
		}
		if blocked > 0 {
			summary.blockedNamespaces++
		}
	}
	return summary, nil
}

// writeNamespacePlan resolves consumers and writes the plan for one namespace.
// On failure it returns the blocker reason describing which step failed.
func (s *StorageController) writeNamespacePlan(ctx context.Context, m *migrationv1alpha1.VmwareCloudFoundationMigration, namespace string, nsPlan *storageplan.NamespacePlan, pods []corev1.Pod, blocked int) (string, error) {
	consumers, err := storageplan.Consumers(ctx, pods, namespace, nsPlan.Volumes, s.lookupStorageOwner)
	if err != nil {
		return "ConsumerResolutionFailed", err
	}
	nsPlan.Consumers = consumers

	if err := s.writeStoragePlan(ctx, m, namespace, nsPlan, blocked); err != nil {
		if apierrors.HasStatusCause(err, corev1.NamespaceTerminatingCause) {
			return "NamespaceTerminating", err
		}
		return "StoragePlanWriteFailed", err
	}
	return "", nil
}

// countBlockedVolumes returns the number of volumes with at least one blocker.
func countBlockedVolumes(volumes []migrationv1alpha1.StorageWorkloadMigrationVolume) int {
	count := 0
	for _, v := range volumes {
		if len(v.Blockers) > 0 {
			count++
		}
	}
	return count
}

// storagePlanBlockedMessage builds the StoragePlanned condition message for a
// plan with blocked volumes or plan-wide blockers.
func storagePlanBlockedMessage(summary storagePlanSummary, blockers []migrationv1alpha1.StorageWorkloadMigrationBlocker) string {
	parts := make([]string, 0, len(blockers)+2)
	parts = append(parts, fmt.Sprintf("%d volumes blocked in %d namespaces", summary.blockedVolumes, summary.blockedNamespaces))
	for _, b := range blockers {
		parts = append(parts, b.Reason+": "+b.Message)
	}
	parts = append(parts, "see StorageWorkloadMigration status")
	return boundConditionMessage(strings.Join(parts, "; "))
}

func (s *StorageController) lookupStorageOwner(ctx context.Context, ns string, owner metav1.OwnerReference) (metav1.Object, error) {
	opts := metav1.GetOptions{}
	switch owner.Kind {
	case "ReplicaSet":
		return s.KubeClient.AppsV1().ReplicaSets(ns).Get(ctx, owner.Name, opts)
	case "Deployment":
		return s.KubeClient.AppsV1().Deployments(ns).Get(ctx, owner.Name, opts)
	case "StatefulSet":
		return s.KubeClient.AppsV1().StatefulSets(ns).Get(ctx, owner.Name, opts)
	case "DaemonSet":
		return s.KubeClient.AppsV1().DaemonSets(ns).Get(ctx, owner.Name, opts)
	case "Job":
		return s.KubeClient.BatchV1().Jobs(ns).Get(ctx, owner.Name, opts)
	case "CronJob":
		return s.KubeClient.BatchV1().CronJobs(ns).Get(ctx, owner.Name, opts)
	default:
		return nil, fmt.Errorf("looking up unsupported controller kind %s", owner.Kind)
	}
}

// PlanName returns the deterministic StorageWorkloadMigration name for m.
func PlanName(m *migrationv1alpha1.VmwareCloudFoundationMigration) string {
	uid := string(m.UID)
	if len(uid) > 8 {
		uid = uid[:8]
	}
	return "vcf-storage-" + uid
}

func (s *StorageController) writeStoragePlan(
	ctx context.Context,
	m *migrationv1alpha1.VmwareCloudFoundationMigration,
	namespace string,
	plan *storageplan.NamespacePlan,
	blocked int,
) error {
	key := client.ObjectKey{Namespace: namespace, Name: PlanName(m)}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		swm := &migrationv1alpha1.StorageWorkloadMigration{}
		// Read through the API server: a stale cached ResourceVersion would
		// make every conflict retry fail the same way.
		err := s.permissionReader().Get(ctx, key, swm)
		ref := migrationv1alpha1.MigrationObjectReference{Name: m.Name, UID: m.UID}
		if apierrors.IsNotFound(err) {
			swm = &migrationv1alpha1.StorageWorkloadMigration{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: namespace,
					Labels: map[string]string{
						MigrationUIDLabel: string(m.UID),
					},
				},
				Spec: migrationv1alpha1.StorageWorkloadMigrationSpec{
					MigrationRef: ref,
				},
			}
			if err := execution.Check(ctx); err != nil {
				return err
			}
			if err = s.Client.Create(ctx, swm); err != nil {
				return fmt.Errorf("creating SWM %s: %w", key, err)
			}
		} else if err != nil {
			return fmt.Errorf("getting SWM %s: %w", key, err)
		}
		if swm.Spec.MigrationRef != ref {
			return fmt.Errorf("checking SWM %s: migration identity conflicts", key)
		}
		if swm.Status.Phase != "" && swm.Status.Phase != migrationv1alpha1.StorageWorkloadMigrationPhasePlanning {
			return nil
		}
		if swm.Labels[MigrationUIDLabel] != string(m.UID) {
			if swm.Labels == nil {
				swm.Labels = map[string]string{}
			}
			swm.Labels[MigrationUIDLabel] = string(m.UID)
			if err := execution.Check(ctx); err != nil {
				return err
			}
			if err := s.Client.Update(ctx, swm); err != nil {
				return err
			}
		}
		previous := swm.Status.DeepCopy()
		swm.Status.Phase = migrationv1alpha1.StorageWorkloadMigrationPhasePlanning
		swm.Status.Volumes = plan.Volumes
		swm.Status.Consumers = plan.Consumers
		status, reason, message := metav1.ConditionTrue, "NoBlockers", "All volumes can be migrated"
		if blocked > 0 {
			status = metav1.ConditionFalse
			reason = "Blocked"
			message = fmt.Sprintf("%d of %d volumes blocked", blocked, len(plan.Volumes))
		}
		apimeta.SetStatusCondition(&swm.Status.Conditions, metav1.Condition{
			Type:               migrationv1alpha1.StorageWorkloadMigrationConditionPlanned,
			Status:             status,
			Reason:             reason,
			Message:            message,
			ObservedGeneration: swm.Generation,
		})
		if reflect.DeepEqual(*previous, swm.Status) {
			return nil
		}
		if err := execution.Check(ctx); err != nil {
			return err
		}
		if err := s.Client.Status().Update(ctx, swm); err != nil {
			return fmt.Errorf("updating SWM %s status: %w", key, err)
		}
		return nil
	})
}

// DeleteStaleStoragePlans deletes planning-phase StorageWorkloadMigrations
// that the current plan no longer covers: plans for namespaces without volumes,
// and plans left by an earlier migration object with a different UID. A nil
// plan deletes every planning-phase plan, for use when storage migration is
// disabled or the migration record is released.
func (s *StorageController) DeleteStaleStoragePlans(ctx context.Context, m *migrationv1alpha1.VmwareCloudFoundationMigration, plan *storageplan.Plan) error {
	var list migrationv1alpha1.StorageWorkloadMigrationList
	if err := s.Client.List(ctx, &list, client.HasLabels{MigrationUIDLabel}); err != nil {
		return fmt.Errorf("listing SWMs: %w", err)
	}
	for i := range list.Items {
		swm := &list.Items[i]
		if !isStaleStoragePlan(m, plan, swm) {
			continue
		}
		if err := execution.Check(ctx); err != nil {
			return err
		}
		if err := s.Client.Delete(ctx, swm, client.Preconditions{UID: &swm.UID, ResourceVersion: &swm.ResourceVersion}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("deleting stale SWM %s/%s: %w", swm.Namespace, swm.Name, err)
		}
	}
	return nil
}

// isStaleStoragePlan reports whether swm is a planning record that plan no
// longer covers. Plans past the planning phase are never deleted here.
func isStaleStoragePlan(m *migrationv1alpha1.VmwareCloudFoundationMigration, plan *storageplan.Plan, swm *migrationv1alpha1.StorageWorkloadMigration) bool {
	planning := swm.Status.Phase == "" || swm.Status.Phase == migrationv1alpha1.StorageWorkloadMigrationPhasePlanning
	if !planning {
		return false
	}
	// Only touch objects the operator wrote: the label and reference agree.
	if swm.Labels[MigrationUIDLabel] != string(swm.Spec.MigrationRef.UID) || swm.Spec.MigrationRef.Name != m.Name {
		return false
	}
	// The migration is a singleton, so a different UID means a deleted predecessor.
	if swm.Spec.MigrationRef.UID != m.UID {
		return true
	}
	if plan == nil {
		return true
	}
	return plan.NamespaceMigrationPlanInstances[swm.Namespace] == nil
}

func (s *StorageController) checkPlanPolicies(ctx context.Context, m *migrationv1alpha1.VmwareCloudFoundationMigration, infra *configv1.Infrastructure, plan *storageplan.Plan, classes []storagev1.StorageClass, pvcs []corev1.PersistentVolumeClaim) error {
	policyCtx, cancel := context.WithTimeout(ctx, policyCheckTimeout)
	defer cancel()
	defer func() {
		// Log out on a fresh bounded context, as preflight does, so cached
		// session handles are released even when policyCtx has expired.
		logoutCtx, logoutCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer logoutCancel()
		vsphere.ClearSessions(logoutCtx)
	}()

	classMap := map[string]*storagev1.StorageClass{}
	for i := range classes {
		classMap[classes[i].Name] = &classes[i]
	}
	claims := map[string]*corev1.PersistentVolumeClaim{}
	for i := range pvcs {
		claims[util.GetObjectKey(pvcs[i].Name, pvcs[i].Namespace)] = &pvcs[i]
	}

	checker := newPolicyChecker(s, m)
	for namespace, nsPlan := range plan.NamespaceMigrationPlanInstances {
		for i := range nsPlan.Volumes {
			v := &nsPlan.Volumes[i]
			if v.TargetFailureDomain == "" {
				continue
			}
			pvc := claims[util.GetObjectKey(v.SourcePVC.Name, namespace)]
			if pvc == nil {
				continue
			}
			policy := customerStoragePolicy(pvc, classMap, infra.Status.InfrastructureName)
			if policy == "" {
				continue
			}
			fd := util.FindFailureDomainByName(m.Spec.FailureDomains, v.TargetFailureDomain)
			if fd == nil {
				continue
			}
			compatible, err := checker.isCompatible(policyCtx, fd, policy)
			if err != nil {
				return err
			}
			if !compatible {
				storageplan.AddBlocker(v, "StoragePolicyMissing", fmt.Sprintf("policy %s is missing or incompatible with target datastore %s on %s", policy, fd.Topology.Datastore, fd.Server))
			}
		}
	}
	return nil
}

// customerStoragePolicy returns the storage policy named by the PVC's
// StorageClass. It returns "" when there is no policy to check: the PVC has no
// known StorageClass, the class names no policy, or the class uses the default
// policy that the CSI driver operator creates in every vCenter.
func customerStoragePolicy(pvc *corev1.PersistentVolumeClaim, classes map[string]*storagev1.StorageClass, infraName string) string {
	if pvc.Spec.StorageClassName == nil {
		return ""
	}
	sc := classes[*pvc.Spec.StorageClassName]
	if sc == nil {
		return ""
	}
	policy, _ := storageplan.Parameter(sc, "storagepolicyname")
	if infraName != "" && policy == "openshift-storage-policy-"+infraName {
		return ""
	}
	return policy
}

// profileKey identifies a storage policy on a target vCenter.
type profileKey struct {
	server string
	policy string
}

// datastoreKey identifies a storage policy checked against one target datastore.
type datastoreKey struct {
	profileKey
	datacenter string
	datastore  string
}

// policyChecker checks storage policies against target vCenters. It caches
// credentials, policy profile IDs, and datastore compatibility so volumes
// sharing a policy and failure domain are checked once.
type policyChecker struct {
	controller *StorageController
	migration  *migrationv1alpha1.VmwareCloudFoundationMigration
	creds      map[string]credentials
	profiles   map[profileKey]string
	results    map[datastoreKey]bool
}

// newPolicyChecker returns a policyChecker with empty caches.
func newPolicyChecker(s *StorageController, m *migrationv1alpha1.VmwareCloudFoundationMigration) *policyChecker {
	return &policyChecker{
		controller: s,
		migration:  m,
		creds:      map[string]credentials{},
		profiles:   map[profileKey]string{},
		results:    map[datastoreKey]bool{},
	}
}

// isCompatible reports whether policy exists in fd's vCenter and matches fd's datastore.
func (c *policyChecker) isCompatible(ctx context.Context, fd *configv1.VSpherePlatformFailureDomainSpec, policy string) (bool, error) {
	key := datastoreKey{
		profileKey: profileKey{server: fd.Server, policy: policy},
		datacenter: fd.Topology.Datacenter,
		datastore:  fd.Topology.Datastore,
	}
	if compatible, ok := c.results[key]; ok {
		return compatible, nil
	}

	session, err := c.session(ctx, fd)
	if err != nil {
		return false, err
	}
	profile, err := c.profileID(ctx, session, key.profileKey)
	if err != nil {
		return false, err
	}
	compatible := false
	if profile != "" {
		compatible, err = vsphere.IsDatastoreCompatible(ctx, session, profile, fd.Topology.Datastore)
		if err != nil {
			return false, err
		}
	}
	c.results[key] = compatible
	return compatible, nil
}

// session returns a cached session for fd's vCenter and datacenter, resolving
// credentials once per server.
func (c *policyChecker) session(ctx context.Context, fd *configv1.VSpherePlatformFailureDomainSpec) (*vsphere.Session, error) {
	creds, ok := c.creds[fd.Server]
	if !ok {
		username, password, err := util.GetTargetCredentials(ctx, c.controller.KubeClient, c.migration, fd.Server)
		if err != nil {
			return nil, err
		}
		creds = credentials{username: username, password: password}
		c.creds[fd.Server] = creds
	}
	session, err := util.GetVSphereSession(ctx, fd.Server, fd.Topology.Datacenter, creds.username, creds.password)
	if err != nil {
		return nil, fmt.Errorf("connecting to target vCenter %s for storage policy checks: %w", fd.Server, err)
	}
	return session, nil
}

// profileID returns the policy's profile ID, or "" when the policy does not exist.
func (c *policyChecker) profileID(ctx context.Context, s *vsphere.Session, key profileKey) (string, error) {
	if profile, ok := c.profiles[key]; ok {
		return profile, nil
	}
	profile, err := vsphere.GetProfileIDByName(ctx, s, key.policy)
	if err != nil {
		return "", err
	}
	c.profiles[key] = profile
	return profile, nil
}
