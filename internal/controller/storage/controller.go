package storage

import (
	"context"
	"strings"

	configclient "github.com/openshift/client-go/config/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/storageplan"
	"github.com/openshift/vcf-migration-operator/internal/util"
)

// Controller defines the interface boundary between the main migration
// reconciler and storage migration orchestration.
type Controller interface {
	// EnsureStoragePlanned computes and persists per-namespace storage plans
	// (StorageWorkloadMigration resources) and updates ConditionStoragePlanned.
	EnsureStoragePlanned(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration) (ctrl.Result, error)

	// DeleteStaleStoragePlans deletes planning records not covered by plan.
	// A nil plan removes all planning records owned by this migration or its predecessors.
	DeleteStaleStoragePlans(ctx context.Context, migration *migrationv1alpha1.VmwareCloudFoundationMigration, plan *storageplan.Plan) error

	// BlockUnsupportedStorageMigration guards the execution boundary until storage
	// migration execution is implemented, reporting whether execution was blocked.
	BlockUnsupportedStorageMigration(migration *migrationv1alpha1.VmwareCloudFoundationMigration) bool

	// ShouldRefreshPlan reports whether storage planning should re-run before
	// execution starts because inventory may have changed.
	ShouldRefreshPlan(migration *migrationv1alpha1.VmwareCloudFoundationMigration) bool

	// CheckStorageManagement validates vSphere CSI driver and Storage operator
	// management states during preflight. It returns a non-fatal warning when
	// applicable.
	CheckStorageManagement(ctx context.Context, migrateStorage bool) (string, error)

	// SetupWatches registers storage-specific watches on the controller builder.
	SetupWatches(bldr *builder.Builder) *builder.Builder
}

// StorageController manages storage migration planning and workload migration
// plan lifecycle independently from the main migration reconciler.
type StorageController struct {
	Client client.Client
	// APIReader bypasses the cache for conflict retries and permission-checked reads.
	APIReader     client.Reader
	KubeClient    kubernetes.Interface
	ConfigClient  configclient.Interface
	DynamicClient dynamic.Interface
	Recorder      events.EventRecorder
}

var _ Controller = (*StorageController)(nil)

// NewStorageController creates a new StorageController with the given clients.
func NewStorageController(
	c client.Client,
	apiReader client.Reader,
	kubeClient kubernetes.Interface,
	configClient configclient.Interface,
	dynamicClient dynamic.Interface,
	recorder events.EventRecorder,
) *StorageController {
	return &StorageController{
		Client:        c,
		APIReader:     apiReader,
		KubeClient:    kubeClient,
		ConfigClient:  configClient,
		DynamicClient: dynamicClient,
		Recorder:      recorder,
	}
}

// SetupWatches registers storage-specific watches on the controller builder.
func (s *StorageController) SetupWatches(bldr *builder.Builder) *builder.Builder {
	return bldr.Watches(
		&corev1.PersistentVolumeClaim{},
		handler.EnqueueRequestsFromMapFunc(s.mapPVCToStoragePlanning),
		builder.OnlyMetadata,
		builder.WithPredicates(pvcMetadataChangedPredicate()),
	)
}

func (s *StorageController) permissionReader() client.Reader {
	return util.PermissionReader(s.APIReader, s.Client)
}

func boundConditionMessage(msg string) string {
	if len(msg) <= util.MaxConditionMessageBytes {
		return msg
	}
	const suffix = "; ...; see StorageWorkloadMigration status"
	maxPrefixLen := util.MaxConditionMessageBytes - len(suffix)
	if maxPrefixLen <= 0 {
		return util.TruncateToBytes(msg, util.MaxConditionMessageBytes)
	}
	prefix := msg[:maxPrefixLen]
	if lastSemi := strings.LastIndex(prefix, "; "); lastSemi > 0 {
		prefix = prefix[:lastSemi]
	}
	return prefix + suffix
}
