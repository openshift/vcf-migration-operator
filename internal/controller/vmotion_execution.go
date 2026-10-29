package controller

import (
	"context"
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

// TODO(SPLAT-2703): qualify the six-hour shutdown, relocation and power-on deadlines.
const vmotionActionTimeout = 6 * time.Hour

// Bound each read separately so an unavailable backend cannot hold the worker
// for the entire action deadline or prevent a later deletion handoff.
const vmotionReadTimeout = 30 * time.Second

const vmotionObservationInterval = 30 * time.Second

// updateVMotionExecution merges only action-owned fields into an authoritative
// read. Unlike executionContext, this permits result persistence during pause,
// spec-generation changes and deletion; it never grants mutation permission.
func (r *VmwareCloudFoundationMigrationReconciler) updateVMotionExecution(ctx context.Context, expected *migrationv1alpha1.VmwareCloudFoundationMigration, change func(*migrationv1alpha1.VmwareCloudFoundationMigration) (bool, error)) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		latest := &migrationv1alpha1.VmwareCloudFoundationMigration{}
		if err := r.permissionReader().Get(ctx, client.ObjectKeyFromObject(expected), latest); err != nil {
			return fmt.Errorf("reading VMotion execution: %w", err)
		}
		if !sameVMotionExecution(expected, latest) {
			return fmt.Errorf("merging VMotion execution: migration UID or frozen execution identity changed")
		}
		changed, err := change(latest)
		if err != nil || !changed {
			return err
		}
		now := metav1.Now()
		latest.Status.LastUpdateTime = &now
		return r.Status().Update(ctx, latest)
	})
}

func sameVMotionExecution(a, b *migrationv1alpha1.VmwareCloudFoundationMigration) bool {
	x, y := a.Status.NodeMigration, b.Status.NodeMigration
	return a.UID != "" && a.UID == b.UID && x != nil && y != nil &&
		x.Type == migrationv1alpha1.NodeMigrationTypeVMotion && y.Type == x.Type &&
		x.MigrationUID == a.UID && y.MigrationUID == a.UID && x.ExecutionID != "" && x.ExecutionID == y.ExecutionID &&
		x.PlanFingerprint != "" && x.PlanFingerprint == y.PlanFingerprint && x.FrozenPlan != nil && reflect.DeepEqual(x.FrozenPlan, y.FrozenPlan) &&
		sameFrozenNodeList(x.Workers, y.Workers) && sameFrozenNodeList(x.ControlPlane, y.ControlPlane)
}

func vmotionActionNode(s *migrationv1alpha1.NodeMigrationStatus, name string) *migrationv1alpha1.NodeMigrationProgress {
	for _, role := range []*migrationv1alpha1.RoleMigrationStatus{s.Workers, s.ControlPlane} {
		if role == nil {
			continue
		}
		for i := range role.Nodes {
			if role.Nodes[i].Name == name {
				return &role.Nodes[i]
			}
		}
	}
	return nil
}

func terminalVMotionObservation(state migrationv1alpha1.NodeMigrationObservationState) bool {
	return state == migrationv1alpha1.NodeMigrationObservationSucceeded || state == migrationv1alpha1.NodeMigrationObservationFailed
}

// submitVMotionAction is reserved for the future admitted node workflow, after
// fresh health/drain/feasibility gates. No production workflow calls it yet.
// Only a newly acknowledged claim can reach the mutating backend, even if a
// concurrent pause follows that claim (pause policy A).
func (r *VmwareCloudFoundationMigrationReconciler) submitVMotionAction(ctx context.Context, expected *migrationv1alpha1.VmwareCloudFoundationMigration, nodeName string, kind migrationv1alpha1.NodeMigrationActionKind, backend vsphere.VMotionMutationBackend) error {
	if backend == nil {
		return fmt.Errorf("submitting VMotion action: backend is not configured")
	}
	var claimed *migrationv1alpha1.VmwareCloudFoundationMigration
	err := r.updateVMotionExecution(ctx, expected, func(latest *migrationv1alpha1.VmwareCloudFoundationMigration) (bool, error) {
		claimed = nil
		s := latest.Status.NodeMigration
		if !canClaimVMotionAction(latest, expected) {
			return false, nil
		}
		if err := validateVMotionSelection(latest.Spec.NodeMigration); err != nil {
			return false, err
		}
		if latest.Spec.NodeMigration.Workers.VMotion.Mode != s.FrozenPlan.Mode {
			return false, fmt.Errorf("claiming VMotion action: selected mode changed")
		}
		if s.ActiveAction != nil && !terminalVMotionObservation(s.ActiveAction.ObservationState) {
			return false, nil
		}
		node := vmotionActionNode(s, nodeName)
		if node == nil || node.NodeUID == "" || node.VMBIOSUUID == "" {
			return false, fmt.Errorf("claiming VMotion action: node identity is not frozen")
		}
		previous := vmotionOutcome(node, kind)
		if previous == nil {
			return false, fmt.Errorf("claiming VMotion action: unsupported kind %q", kind)
		}
		if !canRetryVMotionAction(s, nodeName, kind, *previous) {
			return false, nil
		}
		if s.FrozenPlan.Mode == migrationv1alpha1.VMotionModeHot && kind != migrationv1alpha1.NodeMigrationActionRelocateVM {
			return false, fmt.Errorf("claiming VMotion action: Hot mode permits relocation only")
		}
		if kind == migrationv1alpha1.NodeMigrationActionPowerOnVM && (node.RelocationOutcome == nil || node.RelocationOutcome.State != migrationv1alpha1.NodeMigrationOutcomeSucceeded) {
			return false, fmt.Errorf("claiming VMotion power-on: relocation success is not recorded")
		}
		owner := s.FrozenPlan.SourceVCenter
		if kind == migrationv1alpha1.NodeMigrationActionPowerOnVM {
			owner = s.FrozenPlan.DestinationVCenter
		}
		now := metav1.Now()
		deadline := metav1.NewTime(now.Add(vmotionActionTimeout))
		s.ActiveAction = &migrationv1alpha1.NodeMigrationAction{
			ActionID: string(uuid.NewUUID()), ExecutionID: s.ExecutionID, Kind: kind,
			NodeName: nodeName, NodeUID: node.NodeUID, VMBIOSUUID: node.VMBIOSUUID, OwnerVCenter: owner,
			OperationIdentity: fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s/%s/%s/%s", s.PlanFingerprint, nodeName, kind, owner)))),
			SubmissionState:   migrationv1alpha1.NodeMigrationSubmissionClaimed, ObservationState: migrationv1alpha1.NodeMigrationObservationNotObserved,
			ClaimedAt: now, Deadline: &deadline,
		}
		claimed = latest.DeepCopy()
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("acknowledging VMotion action claim: %w", err)
	}
	if claimed == nil {
		return nil
	}
	a := claimed.Status.NodeMigration.ActiveAction
	submitCtx, cancel := context.WithDeadline(ctx, a.Deadline.Time)
	defer cancel()
	submission, submitErr := callVMotionMutation(submitCtx, claimed, backend)
	return r.persistVMotionSubmission(ctx, claimed, submission, submitErr)
}

func canClaimVMotionAction(latest, expected *migrationv1alpha1.VmwareCloudFoundationMigration) bool {
	s := latest.Status.NodeMigration
	return latest.Name == migrationv1alpha1.SingletonName && latest.Namespace == migrationv1alpha1.SingletonNamespace &&
		latest.DeletionTimestamp.IsZero() && latest.Generation == expected.Generation && latest.Spec.State == migrationv1alpha1.MigrationStateRunning &&
		s.Intervention == nil && reflect.DeepEqual(s.ActiveAction, expected.Status.NodeMigration.ActiveAction)
}

func canRetryVMotionAction(s *migrationv1alpha1.NodeMigrationStatus, nodeName string, kind migrationv1alpha1.NodeMigrationActionKind, previous *migrationv1alpha1.NodeMigrationOperationOutcome) bool {
	ack := s.RecoveryAcknowledgement
	if previous != nil && (previous.State != migrationv1alpha1.NodeMigrationOutcomeFailed || ack == nil || ack.ExecutionID != s.ExecutionID || ack.ActionID != previous.ActionID) {
		return false
	}
	a := s.ActiveAction
	return a == nil || a.ObservationState != migrationv1alpha1.NodeMigrationObservationFailed ||
		(ack != nil && ack.ActionID == a.ActionID && a.Kind == kind && a.NodeName == nodeName)
}

func callVMotionMutation(ctx context.Context, m *migrationv1alpha1.VmwareCloudFoundationMigration, backend vsphere.VMotionMutationBackend) (vsphere.VMotionSubmission, error) {
	s := m.Status.NodeMigration
	a := s.ActiveAction
	if a.Kind == migrationv1alpha1.NodeMigrationActionShutdownGuest {
		return backend.ShutdownGuest(ctx, a.OwnerVCenter, s.FrozenPlan.SourceDatacenter, a.VMBIOSUUID)
	}
	n := vmotionActionNode(s, a.NodeName)
	target, err := targetVMotionPlacement(m, n)
	if err != nil {
		return vsphere.VMotionSubmission{}, err
	}
	if a.Kind == migrationv1alpha1.NodeMigrationActionPowerOnVM {
		return backend.PowerOn(ctx, a.OwnerVCenter, target.Datacenter, a.VMBIOSUUID)
	}
	return backend.Relocate(ctx, a.OwnerVCenter, s.FrozenPlan.SourceDatacenter, a.VMBIOSUUID, target)
}

func targetVMotionPlacement(m *migrationv1alpha1.VmwareCloudFoundationMigration, n *migrationv1alpha1.NodeMigrationProgress) (vsphere.VMotionPlacement, error) {
	for _, fd := range m.Spec.FailureDomains {
		if fd.Name == n.TargetFailureDomain && sameVMotionVCenter(fd.Server, m.Status.NodeMigration.FrozenPlan.DestinationVCenter) && fd.Topology.Datacenter != "" {
			return vsphere.VMotionPlacement{VCenter: m.Status.NodeMigration.FrozenPlan.DestinationVCenter, Datacenter: fd.Topology.Datacenter,
				FailureDomain: n.TargetFailureDomain, ResourcePool: n.TargetResourcePool, Datastore: n.TargetDatastore, Network: n.TargetNetwork}, nil
		}
	}
	return vsphere.VMotionPlacement{}, fmt.Errorf("resolving frozen VMotion destination: target failure domain is missing")
}

func sameVMotionAction(a, b *migrationv1alpha1.NodeMigrationAction) bool {
	return a != nil && b != nil && a.ActionID == b.ActionID && a.ExecutionID == b.ExecutionID && a.Kind == b.Kind &&
		a.NodeName == b.NodeName && a.NodeUID == b.NodeUID && a.VMBIOSUUID == b.VMBIOSUUID &&
		a.OwnerVCenter == b.OwnerVCenter && a.OperationIdentity == b.OperationIdentity
}

func (r *VmwareCloudFoundationMigrationReconciler) persistVMotionSubmission(ctx context.Context, expected *migrationv1alpha1.VmwareCloudFoundationMigration, result vsphere.VMotionSubmission, submitErr error) error {
	return r.updateVMotionExecution(ctx, expected, func(latest *migrationv1alpha1.VmwareCloudFoundationMigration) (bool, error) {
		s := latest.Status.NodeMigration
		a := s.ActiveAction
		if !sameVMotionAction(a, expected.Status.NodeMigration.ActiveAction) || terminalVMotionObservation(a.ObservationState) || a.SubmissionState == migrationv1alpha1.NodeMigrationSubmissionSubmitted {
			return false, nil
		}
		now := metav1.Now()
		a.SubmissionState = migrationv1alpha1.NodeMigrationSubmissionUnknown
		a.Reason = "Submission outcome is uncertain; observe or request administrator revalidation, never replay"
		if submitErr != nil && result.State == "" && result.TaskReference == nil {
			a.ObservationState = migrationv1alpha1.NodeMigrationObservationFailed
			a.Reason = "Submission was rejected; administrator revalidation is required"
		} else if submitErr == nil && validVMotionSubmission(a, result) {
			a.SubmissionState = migrationv1alpha1.NodeMigrationSubmissionSubmitted
			a.SubmittedAt = &now
			if result.TaskReference != nil {
				a.TaskReference = &migrationv1alpha1.NodeMigrationTaskReference{OwnerVCenter: result.TaskReference.OwnerVCenter, TaskID: result.TaskReference.TaskID}
			}
			a.Reason = "Submission acknowledged; awaiting authoritative completion"
		} else {
			a.ObservationState = migrationv1alpha1.NodeMigrationObservationUnknown
		}
		if a.SubmissionState == migrationv1alpha1.NodeMigrationSubmissionUnknown {
			setVMotionIntervention(s, a.Reason, now)
			recordVMotionOutcome(s, now)
		}
		return true, nil
	})
}

func validVMotionSubmission(a *migrationv1alpha1.NodeMigrationAction, result vsphere.VMotionSubmission) bool {
	if result.State != vsphere.VMotionSubmissionAcknowledged {
		return false
	}
	if a.Kind == migrationv1alpha1.NodeMigrationActionShutdownGuest {
		return result.TaskReference == nil
	}
	return result.TaskReference != nil && result.TaskReference.OwnerVCenter == a.OwnerVCenter && result.TaskReference.TaskID != "" && len(result.TaskReference.TaskID) <= 512
}

func setVMotionIntervention(s *migrationv1alpha1.NodeMigrationStatus, reason string, now metav1.Time) {
	if s.Intervention != nil {
		return
	}
	a := s.ActiveAction
	s.Intervention = &migrationv1alpha1.NodeMigrationIntervention{ExecutionID: s.ExecutionID, ActionID: a.ActionID, NodeName: a.NodeName, Reason: reason, Since: now}
}

func vmotionOutcome(n *migrationv1alpha1.NodeMigrationProgress, kind migrationv1alpha1.NodeMigrationActionKind) **migrationv1alpha1.NodeMigrationOperationOutcome {
	switch kind {
	case migrationv1alpha1.NodeMigrationActionShutdownGuest:
		return &n.ShutdownOutcome
	case migrationv1alpha1.NodeMigrationActionRelocateVM:
		return &n.RelocationOutcome
	case migrationv1alpha1.NodeMigrationActionPowerOnVM:
		return &n.PowerOnOutcome
	default:
		return nil
	}
}

func recordVMotionOutcome(s *migrationv1alpha1.NodeMigrationStatus, now metav1.Time) {
	a := s.ActiveAction
	n := vmotionActionNode(s, a.NodeName)
	outcome := vmotionOutcome(n, a.Kind)
	state := migrationv1alpha1.NodeMigrationOutcomeUnknown
	if a.ObservationState == migrationv1alpha1.NodeMigrationObservationSucceeded {
		state = migrationv1alpha1.NodeMigrationOutcomeSucceeded
	}
	if a.ObservationState == migrationv1alpha1.NodeMigrationObservationFailed {
		state = migrationv1alpha1.NodeMigrationOutcomeFailed
	}
	*outcome = &migrationv1alpha1.NodeMigrationOperationOutcome{ActionID: a.ActionID, Kind: a.Kind, State: state,
		OwnerVCenter: a.OwnerVCenter, TaskReference: a.TaskReference.DeepCopy(), Reason: a.Reason, ObservedAt: &now}
	if terminalVMotionObservation(a.ObservationState) {
		(*outcome).CompletedAt = &now
	}
}

// observeVMotionAction cannot submit mutations. Observations are conditional on
// the exact action/task snapshot read before the external call; a concurrent
// terminal result, task binding or newer action wins over this writer.
func (r *VmwareCloudFoundationMigrationReconciler) observeVMotionAction(ctx context.Context, expected *migrationv1alpha1.VmwareCloudFoundationMigration, backend vsphere.VMotionReadOnlyBackend) error {
	if err := validateVMotionActionIdentity(expected); err != nil {
		return err
	}
	a := expected.Status.NodeMigration.ActiveAction
	if terminalVMotionObservation(a.ObservationState) {
		return nil
	}
	// Status writes also trigger the CR watch. Do not let our own observation
	// timestamp turn the polling interval into an immediate reconcile loop.
	if a.LastObservedAt != nil {
		age := time.Since(a.LastObservedAt.Time)
		if age >= 0 && age < vmotionObservationInterval {
			return nil
		}
	}
	readCtx, cancel := context.WithTimeout(ctx, vmotionReadTimeout)
	defer cancel()
	state, reason := observeVMotionBackend(readCtx, expected, backend)
	now := metav1.Now()
	return r.updateVMotionExecution(ctx, expected, func(latest *migrationv1alpha1.VmwareCloudFoundationMigration) (bool, error) {
		s := latest.Status.NodeMigration
		current := s.ActiveAction
		if !reflect.DeepEqual(current, a) {
			return false, nil
		}
		current.ObservationState = state
		current.LastObservedAt = &now
		current.Reason = reason
		if a.Deadline == nil || !now.Before(a.Deadline) {
			setVMotionIntervention(s, "Action deadline expired or is missing; late results do not authorize further mutations", now)
		}
		if state == migrationv1alpha1.NodeMigrationObservationUnknown || state == migrationv1alpha1.NodeMigrationObservationFailed {
			setVMotionIntervention(s, reason, now)
		}
		if state != migrationv1alpha1.NodeMigrationObservationRunning {
			recordVMotionOutcome(s, now)
		}
		if s.Intervention != nil && state == migrationv1alpha1.NodeMigrationObservationRunning {
			current.ObservationState = migrationv1alpha1.NodeMigrationObservationUnknown
			recordVMotionOutcome(s, now)
		}
		return true, nil
	})
}

func validateVMotionActionIdentity(m *migrationv1alpha1.VmwareCloudFoundationMigration) error {
	s := m.Status.NodeMigration
	if s == nil || s.ActiveAction == nil || s.FrozenPlan == nil {
		return fmt.Errorf("observing VMotion action: frozen execution or action is missing")
	}
	a := s.ActiveAction
	n := vmotionActionNode(s, a.NodeName)
	if a.ExecutionID != s.ExecutionID || a.ActionID == "" || n == nil || a.NodeUID != n.NodeUID || a.VMBIOSUUID != n.VMBIOSUUID || vmotionOutcome(n, a.Kind) == nil {
		return fmt.Errorf("observing VMotion action: action does not match frozen node/execution identity")
	}
	owner := s.FrozenPlan.SourceVCenter
	if a.Kind == migrationv1alpha1.NodeMigrationActionPowerOnVM {
		owner = s.FrozenPlan.DestinationVCenter
	}
	if a.OwnerVCenter != owner || (a.Kind == migrationv1alpha1.NodeMigrationActionShutdownGuest && a.TaskReference != nil) ||
		(a.TaskReference != nil && (a.TaskReference.OwnerVCenter != a.OwnerVCenter || a.TaskReference.TaskID == "")) {
		return fmt.Errorf("observing VMotion action: task or action has invalid owning-vCenter scope")
	}
	return nil
}

func observeVMotionBackend(ctx context.Context, m *migrationv1alpha1.VmwareCloudFoundationMigration, backend vsphere.VMotionReadOnlyBackend) (migrationv1alpha1.NodeMigrationObservationState, string) {
	a := m.Status.NodeMigration.ActiveAction
	unknown := migrationv1alpha1.NodeMigrationObservationUnknown
	if backend == nil {
		return unknown, "Observation backend is unavailable; administrator intervention is required"
	}
	if a.Kind == migrationv1alpha1.NodeMigrationActionShutdownGuest {
		vm, err := backend.InspectVM(ctx, a.OwnerVCenter, m.Status.NodeMigration.FrozenPlan.SourceDatacenter, a.VMBIOSUUID)
		if err != nil || !strings.EqualFold(vm.BIOSUUID, a.VMBIOSUUID) {
			return unknown, "Shutdown VM identity/power evidence is unavailable or mismatched"
		}
		if vm.PowerState == "poweredOff" {
			return migrationv1alpha1.NodeMigrationObservationSucceeded, "Graceful shutdown observed powered off"
		}
		if vm.PowerState == "poweredOn" && a.SubmissionState == migrationv1alpha1.NodeMigrationSubmissionSubmitted {
			return migrationv1alpha1.NodeMigrationObservationRunning, "Awaiting graceful shutdown; no hard power-off is permitted"
		}
		return unknown, "Shutdown submission or observed power state is uncertain; never replay"
	}
	if a.TaskReference == nil {
		return unknown, "Claim has no task reference; submission may have begun, never replay"
	}
	observation, err := backend.ObserveTask(ctx, a.TaskReference.OwnerVCenter, a.TaskReference.TaskID)
	if err != nil {
		return unknown, "Task evidence is unavailable; task absence does not prove inactivity"
	}
	switch observation.State {
	case vsphere.VMotionTaskRunning:
		return migrationv1alpha1.NodeMigrationObservationRunning, "Task is running"
	case vsphere.VMotionTaskSucceeded:
		return migrationv1alpha1.NodeMigrationObservationSucceeded, "Task completed successfully; node recovery requires separate verification"
	case vsphere.VMotionTaskFailed:
		return migrationv1alpha1.NodeMigrationObservationFailed, "Task failed; administrator revalidation is required"
	default:
		return unknown, "Task outcome cannot be established; never replay"
	}
}

func hasUnresolvedVMotionAction(m *migrationv1alpha1.VmwareCloudFoundationMigration) bool {
	s := m.Status.NodeMigration
	return s != nil && s.Type == migrationv1alpha1.NodeMigrationTypeVMotion &&
		(s.Intervention != nil || (s.ActiveAction != nil && s.ActiveAction.ObservationState != migrationv1alpha1.NodeMigrationObservationSucceeded))
}

func hasVMotionAction(m *migrationv1alpha1.VmwareCloudFoundationMigration) bool {
	s := m.Status.NodeMigration
	return s != nil && s.Type == migrationv1alpha1.NodeMigrationTypeVMotion && s.ActiveAction != nil
}

func (r *VmwareCloudFoundationMigrationReconciler) reconcileVMotionAction(ctx context.Context, m *migrationv1alpha1.VmwareCloudFoundationMigration) (bool, ctrl.Result, error) {
	if !hasVMotionAction(m) || (!m.DeletionTimestamp.IsZero() && m.Annotations[migrationv1alpha1.AnnotationAllowDeletion] == allowDeletionAnnotationValue) {
		return false, ctrl.Result{}, nil
	}
	if err := r.observeVMotionAction(ctx, m, r.VMotionBackend); err != nil {
		return true, ctrl.Result{}, err
	}
	if !m.DeletionTimestamp.IsZero() {
		return false, ctrl.Result{}, nil
	}
	latest := &migrationv1alpha1.VmwareCloudFoundationMigration{}
	if err := r.permissionReader().Get(ctx, client.ObjectKeyFromObject(m), latest); err != nil {
		return true, ctrl.Result{}, err
	}
	if latest.Spec.NodeMigration != nil && latest.Spec.NodeMigration.RecoveryRequest != nil {
		if err := r.recheckVMotionRecovery(ctx, latest, r.VMotionRecoveryCheck); err != nil {
			return true, ctrl.Result{}, err
		}
	}
	return true, ctrl.Result{RequeueAfter: vmotionObservationInterval}, nil
}
