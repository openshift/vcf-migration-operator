package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
)

const (
	vmotionRecoveryCondition     = "VMotionRecovery"
	vmotionRecoveryBlockedReason = "RecoveryBlocked"
)

// recheckVMotionRecovery consumes one request only after a read-only verifier
// establishes claimant quiescence and authoritative outcome/inactivity, unique expected VM location and
// power, preserved Node identity, applicable health and unchanged live plan and
// inventory. A nil verifier fails closed until the later workflow supplies it.
// Clearing intervention never bypasses Running/admission gates for new claims.
// Expected refusals are status observations; operational failures remain errors.
func (r *VmwareCloudFoundationMigrationReconciler) recheckVMotionRecovery(ctx context.Context, expected *migrationv1alpha1.VmwareCloudFoundationMigration, verify func(context.Context, *migrationv1alpha1.VmwareCloudFoundationMigration) (migrationv1alpha1.NodeMigrationObservationState, error)) error {
	if err := validateVMotionActionIdentity(expected); err != nil {
		return err
	}
	s := expected.Status.NodeMigration
	if reason := vmotionRecoveryBlockReason(expected); reason != "" {
		return r.recordBlockedVMotionRecovery(ctx, expected, reason)
	}
	request := expected.Spec.NodeMigration.RecoveryRequest
	ack := s.RecoveryAcknowledgement
	if ack != nil && ack.RequestID == request.RequestID {
		return nil
	}
	if verify == nil {
		return r.recordBlockedVMotionRecovery(ctx, expected, "Authoritative recovery verifier is not configured; recovery remains blocked")
	}
	readCtx, cancel := context.WithTimeout(ctx, vmotionReadTimeout)
	defer cancel()
	state, err := verify(readCtx, expected.DeepCopy())
	if err != nil || readCtx.Err() != nil {
		if statusErr := r.recordBlockedVMotionRecovery(ctx, expected, "Authoritative recovery evidence is unavailable; verification must be retried"); statusErr != nil {
			return statusErr
		}
		return fmt.Errorf("rechecking VMotion recovery: authoritative evidence is unavailable")
	}
	if !terminalVMotionObservation(state) {
		return r.recordBlockedVMotionRecovery(ctx, expected, "Action outcome/inactivity remains uncertain; recovery remains blocked")
	}
	if terminalVMotionObservation(s.ActiveAction.ObservationState) && s.ActiveAction.ObservationState != state {
		return r.recordBlockedVMotionRecovery(ctx, expected, "Recovery evidence conflicts with a committed terminal action outcome")
	}
	stale := false
	err = r.updateVMotionExecution(ctx, expected, func(latest *migrationv1alpha1.VmwareCloudFoundationMigration) (bool, error) {
		stale = false
		if latest.ResourceVersion != expected.ResourceVersion || latest.Spec.NodeMigration == nil ||
			!reflect.DeepEqual(latest.Spec.NodeMigration.RecoveryRequest, request) {
			stale = true
			return false, nil
		}
		current := latest.Status.NodeMigration
		now := metav1.Now()
		if !terminalVMotionObservation(current.ActiveAction.ObservationState) {
			current.ActiveAction.ObservationState = state
			current.ActiveAction.LastObservedAt = &now
			current.ActiveAction.Reason = "Action outcome/inactivity established by explicit administrator-requested verification"
			recordVMotionOutcome(current, now)
		}
		current.RecoveryAcknowledgement = &migrationv1alpha1.NodeMigrationRecoveryAcknowledgement{
			RequestID: request.RequestID, ExecutionID: request.ExecutionID, ActionID: request.ActionID, ConsumedAt: now,
		}
		current.Intervention = nil
		r.setCondition(latest, vmotionRecoveryCondition, metav1.ConditionTrue, "RecoveryVerified", "Recovery request verified and consumed for the current execution/action; new mutations still require normal admission")
		return true, nil
	})
	if err != nil {
		return err
	}
	if stale {
		return r.recordBlockedVMotionRecovery(ctx, expected, "Execution changed during recovery verification; a fresh recheck is required")
	}
	if r.Recorder != nil {
		r.Recorder.Eventf(expected, nil, "Normal", "RecoveryVerified", "RecoveryVerified", "Recovery request verified and consumed for the current execution/action")
	}
	return nil
}

func vmotionRecoveryBlockReason(m *migrationv1alpha1.VmwareCloudFoundationMigration) string {
	s := m.Status.NodeMigration
	if m.Spec.NodeMigration == nil {
		return "Recovery request is missing"
	}
	request := m.Spec.NodeMigration.RecoveryRequest
	if request == nil || request.RequestID == "" || request.ExecutionID != s.ExecutionID || request.ActionID != s.ActiveAction.ActionID {
		return "Recovery request does not identify the current execution/action"
	}
	ack := s.RecoveryAcknowledgement
	if ack != nil && ack.RequestID == request.RequestID {
		if ack.ExecutionID == request.ExecutionID && ack.ActionID == request.ActionID {
			return ""
		}
		return "Recovery request ID was already consumed for another action"
	}
	if !m.DeletionTimestamp.IsZero() || s.Intervention == nil || s.Intervention.ExecutionID != s.ExecutionID || s.Intervention.ActionID != request.ActionID {
		return "No matching intervention is eligible for recovery"
	}
	// External inactivity cannot fence a claimant suspended before its request.
	// Expiry ends its context-bound submission authority; the verifier must still
	// establish claimant quiescence and authoritative external outcome/inactivity.
	if s.ActiveAction.SubmissionState == migrationv1alpha1.NodeMigrationSubmissionClaimed &&
		(s.ActiveAction.Deadline == nil || time.Now().Before(s.ActiveAction.Deadline.Time)) {
		return "Original claimant may still submit before its deadline; recovery remains blocked"
	}
	return ""
}

// recordBlockedVMotionRecovery reports an expected refusal without changing
// intervention, consuming a request, or switching observation to error backoff.
func (r *VmwareCloudFoundationMigrationReconciler) recordBlockedVMotionRecovery(ctx context.Context, expected *migrationv1alpha1.VmwareCloudFoundationMigration, message string) error {
	published := false
	err := r.updateVMotionExecution(ctx, expected, func(latest *migrationv1alpha1.VmwareCloudFoundationMigration) (bool, error) {
		published = false
		if !sameVMotionAction(latest.Status.NodeMigration.ActiveAction, expected.Status.NodeMigration.ActiveAction) {
			return false, nil
		}
		var request, currentRequest *migrationv1alpha1.NodeMigrationRecoveryRequest
		if expected.Spec.NodeMigration != nil {
			request = expected.Spec.NodeMigration.RecoveryRequest
		}
		if latest.Spec.NodeMigration != nil {
			currentRequest = latest.Spec.NodeMigration.RecoveryRequest
		}
		if !reflect.DeepEqual(request, currentRequest) {
			return false, nil
		}
		condition := apimeta.FindStatusCondition(latest.Status.Conditions, vmotionRecoveryCondition)
		if condition != nil && condition.Status == metav1.ConditionFalse && condition.Reason == vmotionRecoveryBlockedReason && condition.Message == message && condition.ObservedGeneration == latest.Generation {
			return false, nil
		}
		r.setCondition(latest, vmotionRecoveryCondition, metav1.ConditionFalse, vmotionRecoveryBlockedReason, message)
		published = true
		return true, nil
	})
	if err == nil && published && r.Recorder != nil {
		r.Recorder.Eventf(expected, nil, "Warning", vmotionRecoveryBlockedReason, vmotionRecoveryBlockedReason, "%s", message)
	}
	return err
}
