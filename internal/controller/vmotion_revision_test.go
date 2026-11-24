package controller

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

func TestVMotionReviewSubmissionSafety(t *testing.T) {
	for _, tc := range []struct {
		name   string
		result vsphere.VMotionSubmission
		err    error
	}{
		{name: "cancelled", err: fmt.Errorf("sending request: %w", context.Canceled)},
		{name: "deadline", err: context.DeadlineExceeded},
		{name: "uncertain task", result: vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown, TaskReference: &vsphere.VMotionTaskReference{OwnerVCenter: cleanupTestSource, TaskID: executionTestTaskID}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
			m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return tc.result, tc.err }})
			a := m.Status.NodeMigration.ActiveAction
			if a.ObservationState != migrationv1alpha1.NodeMigrationObservationUnknown {
				t.Fatalf("uncertain submission recorded as %s", a.ObservationState)
			}
			if tc.result.TaskReference != nil && a.TaskReference == nil {
				t.Fatal("valid uncertain task discarded")
			}
		})
	}
}

func TestVMotionReviewClaimedShutdownCannotComplete(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	m.Status.NodeMigration.Workers.Nodes[0].ShutdownOutcome = nil
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionShutdownGuest, &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil
	}})
	m.Status.NodeMigration.ActiveAction.SubmissionState = migrationv1alpha1.NodeMigrationSubmissionClaimed
	m.Status.NodeMigration.Intervention = nil
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	b := &executionBackend{inspect: func(context.Context, string, string, string) (vsphere.VMotionVM, error) {
		return vsphere.VMotionVM{BIOSUUID: executionTestBIOSUUID, PowerState: "poweredOff"}, nil
	}}
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	if readExecution(t, r, m).Status.NodeMigration.ActiveAction.ObservationState == migrationv1alpha1.NodeMigrationObservationSucceeded {
		t.Fatal("power evidence resolved live claimant")
	}
}

func TestVMotionReviewReadRetryAndSelectorRemoval(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() }})
	m.Spec.NodeMigration = nil
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	b := &executionBackend{observe: func(context.Context, string, string) (vsphere.VMotionTaskObservation, error) {
		return vsphere.VMotionTaskObservation{}, fmt.Errorf("temporary read failure")
	}}
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	m = readExecution(t, r, m)
	if m.Status.NodeMigration.Intervention != nil {
		t.Fatal("transient read permanently blocked action")
	}
	old := metav1.NewTime(time.Now().Add(-time.Minute))
	m.Status.NodeMigration.ActiveAction.LastObservedAt = &old
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	b.observe = func(context.Context, string, string) (vsphere.VMotionTaskObservation, error) {
		return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskSucceeded}, nil
	}
	r.VMotionBackend = b
	stop, _, err := r.reconcileVMotionAction(context.Background(), m)
	if err != nil || stop {
		t.Fatalf("terminal action blocked workflow: stop=%v err=%v status=%+v", stop, err, readExecution(t, r, m).Status.NodeMigration.ActiveAction)
	}
	if readExecution(t, r, m).Status.NodeMigration.ActiveAction.ObservationState != migrationv1alpha1.NodeMigrationObservationSucceeded {
		t.Fatal("selector removal lost observation")
	}
}

func TestVMotionReviewConsumedRecoveryCannotBeBlocked(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil
	}})
	m.Spec.NodeMigration.RecoveryRequest = &migrationv1alpha1.NodeMigrationRecoveryRequest{RequestID: executionTestRecoveryID, ExecutionID: executionTestID, ActionID: m.Status.NodeMigration.ActiveAction.ActionID}
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	stale := m.DeepCopy()
	m.Status.NodeMigration.RecoveryAcknowledgement = &migrationv1alpha1.NodeMigrationRecoveryAcknowledgement{RequestID: executionTestRecoveryID, ExecutionID: executionTestID, ActionID: m.Status.NodeMigration.ActiveAction.ActionID}
	m.Status.NodeMigration.Intervention = nil
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if err := r.recordBlockedVMotionRecovery(context.Background(), stale, "stale verifier"); err != nil {
		t.Fatal(err)
	}
	if len(readExecution(t, r, m).Status.Conditions) != 0 {
		t.Fatal("losing verifier published blocked condition")
	}
}

func TestVMotionReviewHotGuard(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	m.Status.NodeMigration.FrozenPlan.Mode = migrationv1alpha1.VMotionModeHot
	m.Status.NodeMigration.Workers.Nodes[0].ShutdownOutcome = nil
	m.Spec.NodeMigration.Workers.VMotion.Mode = migrationv1alpha1.VMotionModeHot
	m.Spec.NodeMigration.ControlPlane.VMotion.Mode = migrationv1alpha1.VMotionModeHot
	status := m.Status.DeepCopy()
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	m.Status = *status
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	err := r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionShutdownGuest, &executionBackend{})
	if err == nil || !strings.Contains(err.Error(), "Hot mode permits relocation only") {
		t.Fatalf("wrong guard: %v", err)
	}
}

func TestVMotionReviewInterventionEvent(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	recorder := events.NewFakeRecorder(4)
	r.Recorder = recorder
	claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil
	}})
	select {
	case <-recorder.Events:
	default:
		t.Fatal("intervention emitted no event")
	}
}
