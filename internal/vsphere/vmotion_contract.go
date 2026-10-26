package vsphere

import "context"

// VMotionManagedObjectReference identifies one vSphere object within a vCenter
// and datacenter scope.
type VMotionManagedObjectReference struct {
	VCenter    string
	Datacenter string
	Type       string
	Value      string
}

// VMotionSourceFailureDomain supplies one configured source failure domain's
// name and compute-resource inventory path for scoped reference resolution.
type VMotionSourceFailureDomain struct {
	Name                string
	ComputeResourcePath string
}

// VMotionInventory contains VM matches and each source failure domain's
// resolved compute-resource reference. Duplicate VM UUID matches are retained.
type VMotionInventory struct {
	VMs                           []VMotionVM
	SourceFailureDomainReferences map[string]VMotionManagedObjectReference
}

// VMotionPlacement is the frozen destination placement sent to native vSphere
// checks and relocation. Host is intentionally absent; destination DRS selects
// the host, which is recorded from observation after relocation.
type VMotionPlacement struct {
	VCenter       string
	Datacenter    string
	FailureDomain string
	ResourcePool  string
	Datastore     string
	Network       string
}

// VMotionTaskReference scopes a task ID to the vCenter that created it.
type VMotionTaskReference struct {
	OwnerVCenter string
	TaskID       string
}

// VMotionSubmissionState reports whether a mutation submission was
// acknowledged or remains ambiguous.
type VMotionSubmissionState string

const (
	// VMotionSubmissionAcknowledged means vCenter acknowledged the mutation.
	VMotionSubmissionAcknowledged VMotionSubmissionState = "Acknowledged"
	// VMotionSubmissionUnknown means the mutation may have reached vCenter; it must not be replayed.
	VMotionSubmissionUnknown VMotionSubmissionState = "Unknown"
)

// VMotionSubmission is returned after a mutating call. Acknowledged means the
// request was accepted; task-based operations then require TaskReference.
// TaskReference is absent for ShutdownGuest and may be absent when State is
// Unknown. Unknown means submission may have begun and must not be replayed.
type VMotionSubmission struct {
	State         VMotionSubmissionState
	TaskReference *VMotionTaskReference
}

// VMotionFinding is one bounded result from the native relocation checker.
type VMotionFinding struct {
	Severity string
	Message  string
}

// VMotionCheckResult contains per-VM/host findings; a successful API task alone
// does not imply compatibility.
type VMotionCheckResult struct {
	Findings []VMotionFinding
}

// VMotionTaskState is the latest authoritative state observed for a task.
type VMotionTaskState string

const (
	// VMotionTaskRunning means vCenter reports the task as active.
	VMotionTaskRunning VMotionTaskState = "Running"
	// VMotionTaskSucceeded means vCenter reports successful completion.
	VMotionTaskSucceeded VMotionTaskState = "Succeeded"
	// VMotionTaskFailed means vCenter reports unsuccessful completion.
	VMotionTaskFailed VMotionTaskState = "Failed"
	// VMotionTaskUnknown means task state cannot be authoritatively established.
	VMotionTaskUnknown VMotionTaskState = "Unknown"
)

// VMotionTaskObservation contains bounded task state, not full task output.
type VMotionTaskObservation struct {
	State  VMotionTaskState
	Reason string
}

// VMotionReadOnlyBackend contains inventory and task/VM observation calls. A
// controller may observe and persist the result of an already-claimed action
// across pause or generation changes, bound to the same CR UID, execution ID,
// and action ID. That permission never authorizes a new mutation.
type VMotionReadOnlyBackend interface {
	// ListVMotionVMInventory performs a read-only BIOS-UUID inventory search scoped to one datacenter.
	ListVMotionVMInventory(ctx context.Context, vcenter, datacenter string, sourceFailureDomains []VMotionSourceFailureDomain, biosUUIDs []string) (VMotionInventory, error)
	// CheckRelocate performs a read-only native feasibility check for the exact frozen placement.
	CheckRelocate(ctx context.Context, sourceVCenter, sourceDatacenter, biosUUID string, target VMotionPlacement) (VMotionCheckResult, error)
	// InspectVM performs a read-only post-operation identity, host and power-state observation.
	InspectVM(ctx context.Context, vcenter, datacenter, biosUUID string) (VMotionVM, error)
	// ObserveTask performs a read-only lookup within ownerVCenter; task IDs are never global.
	ObserveTask(ctx context.Context, ownerVCenter, taskID string) (VMotionTaskObservation, error)
}

// VMotionMutationBackend contains VM-changing calls. The controller invokes
// each method only after acknowledging a durable action claim. Ambiguous
// submission is represented as VMotionSubmissionUnknown and never authorizes a
// retry. A returned task reference carries the vCenter that created it; an
// ordinary error must indicate that submission was definitely rejected.
type VMotionMutationBackend interface {
	// ShutdownGuest requests only graceful guest shutdown and returns no task reference.
	ShutdownGuest(ctx context.Context, vcenter, datacenter, biosUUID string) (VMotionSubmission, error)
	// Relocate submits one relocation using target placement; an empty host means DRS chooses it.
	Relocate(ctx context.Context, sourceVCenter, sourceDatacenter, biosUUID string, target VMotionPlacement) (VMotionSubmission, error)
	// PowerOn submits normal destination power-on after a successful Cold relocation.
	PowerOn(ctx context.Context, vcenter, datacenter, biosUUID string) (VMotionSubmission, error)
}

// VMotionBackend is the combined adapter contract for PR 3 execution and PR 4
// vSphere operations.
type VMotionBackend interface {
	VMotionReadOnlyBackend
	VMotionMutationBackend
}
