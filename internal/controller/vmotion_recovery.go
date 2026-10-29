package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
)

// recheckVMotionRecovery consumes one request only after a read-only verifier
// establishes claimant quiescence and authoritative outcome/inactivity, unique expected VM location and
// power, preserved Node identity, applicable health and unchanged live plan and
// inventory. A nil verifier fails closed until the later workflow supplies it.
// Clearing intervention never bypasses Running/admission gates for new claims.
func (r *VmwareCloudFoundationMigrationReconciler) recheckVMotionRecovery(ctx context.Context, expected *migrationv1alpha1.VmwareCloudFoundationMigration, verify func(context.Context, *migrationv1alpha1.VmwareCloudFoundationMigration) (migrationv1alpha1.NodeMigrationObservationState, error)) error {
	if err := validateVMotionActionIdentity(expected); err != nil {
		return err
	}
	s := expected.Status.NodeMigration
	if expected.Spec.NodeMigration == nil {
		return fmt.Errorf("rechecking VMotion recovery: request is missing")
	}
	request := expected.Spec.NodeMigration.RecoveryRequest
	if request == nil || request.RequestID == "" || request.ExecutionID != s.ExecutionID || request.ActionID != s.ActiveAction.ActionID {
		return fmt.Errorf("rechecking VMotion recovery: request does not identify the current execution/action")
	}
	ack := s.RecoveryAcknowledgement
	if ack != nil && ack.RequestID == request.RequestID {
		if ack.ExecutionID == request.ExecutionID && ack.ActionID == request.ActionID {
			return nil
		}
		return fmt.Errorf("rechecking VMotion recovery: request ID was already consumed for another action")
	}
	if !expected.DeletionTimestamp.IsZero() || s.Intervention == nil || s.Intervention.ExecutionID != s.ExecutionID || s.Intervention.ActionID != request.ActionID {
		return fmt.Errorf("rechecking VMotion recovery: no matching intervention is eligible for recovery")
	}
	// External inactivity cannot fence a claimant suspended before its request.
	// Expiry ends its context-bound submission authority; the verifier must still
	// establish claimant quiescence and authoritative external outcome/inactivity.
	if s.ActiveAction.SubmissionState == migrationv1alpha1.NodeMigrationSubmissionClaimed &&
		(s.ActiveAction.Deadline == nil || time.Now().Before(s.ActiveAction.Deadline.Time)) {
		return fmt.Errorf("rechecking VMotion recovery: original claimant may still submit before its deadline")
	}
	if verify == nil {
		return fmt.Errorf("rechecking VMotion recovery: authoritative recovery verifier is not configured")
	}
	readCtx, cancel := context.WithTimeout(ctx, vmotionReadTimeout)
	defer cancel()
	state, err := verify(readCtx, expected.DeepCopy())
	if err != nil || readCtx.Err() != nil {
		return fmt.Errorf("rechecking VMotion recovery: authoritative evidence is unavailable")
	}
	if !terminalVMotionObservation(state) {
		return fmt.Errorf("rechecking VMotion recovery: action outcome/inactivity remains uncertain")
	}
	if terminalVMotionObservation(s.ActiveAction.ObservationState) && s.ActiveAction.ObservationState != state {
		return fmt.Errorf("rechecking VMotion recovery: evidence conflicts with a committed terminal action outcome")
	}
	return r.updateVMotionExecution(ctx, expected, func(latest *migrationv1alpha1.VmwareCloudFoundationMigration) (bool, error) {
		if latest.ResourceVersion != expected.ResourceVersion || latest.Spec.NodeMigration == nil ||
			!reflect.DeepEqual(latest.Spec.NodeMigration.RecoveryRequest, request) {
			return false, fmt.Errorf("consuming VMotion recovery request: execution changed during verification; recheck required")
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
		return true, nil
	})
}
