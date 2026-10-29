package controller

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

const (
	executionTestBIOSUUID   = "00000000-0000-4000-8000-000000000001"
	executionTestID         = "11111111-1111-4111-8111-111111111111"
	executionTestRecoveryID = "22222222-2222-4222-8222-222222222222"
	executionTestTargetDC   = "TargetDC"
	executionTestNetwork    = "VM Network"
	executionTestTargetVC   = "target.example.com"
	executionTestNode       = "worker-0"
	executionTestTaskID     = "task-1"
	executionTestResource   = "vmwarecloudfoundationmigrations"
)

type executionBackend struct {
	vsphere.VMotionBackend
	submit  func(context.Context) (vsphere.VMotionSubmission, error)
	observe func(context.Context, string, string) (vsphere.VMotionTaskObservation, error)
	inspect func(context.Context, string, string, string) (vsphere.VMotionVM, error)
}

func (b *executionBackend) Relocate(ctx context.Context, vc, dc, biosUUID string, target vsphere.VMotionPlacement) (vsphere.VMotionSubmission, error) {
	if vc != cleanupTestSource || dc != vmotionPlanTestSourceDatacenter || biosUUID != executionTestBIOSUUID ||
		target.VCenter != executionTestTargetVC || target.Datacenter != executionTestTargetDC || target.ResourcePool != "/TargetDC/host/cluster/Resources" {
		return vsphere.VMotionSubmission{}, fmt.Errorf("unexpected frozen relocation inputs")
	}
	return b.submit(ctx)
}

func (b *executionBackend) ShutdownGuest(ctx context.Context, _, _, _ string) (vsphere.VMotionSubmission, error) {
	return b.submit(ctx)
}

func (b *executionBackend) PowerOn(ctx context.Context, vc, dc, id string) (vsphere.VMotionSubmission, error) {
	if vc != executionTestTargetVC || dc != executionTestTargetDC || id != executionTestBIOSUUID {
		return vsphere.VMotionSubmission{}, fmt.Errorf("power-on must use destination identity and scope")
	}
	return b.submit(ctx)
}

func (b *executionBackend) ObserveTask(ctx context.Context, vc, task string) (vsphere.VMotionTaskObservation, error) {
	return b.observe(ctx, vc, task)
}

func (b *executionBackend) InspectVM(ctx context.Context, vc, dc, id string) (vsphere.VMotionVM, error) {
	return b.inspect(ctx, vc, dc, id)
}

func newVMotionExecutionFixture(t *testing.T, hooks interceptor.Funcs) (*VmwareCloudFoundationMigrationReconciler, *migrationv1alpha1.VmwareCloudFoundationMigration) {
	t.Helper()
	m := newVMotionPlanMigration()
	m.Generation = 1
	m.Status.NodeMigration = &migrationv1alpha1.NodeMigrationStatus{
		Type: migrationv1alpha1.NodeMigrationTypeVMotion, MigrationUID: m.UID,
		ExecutionID: executionTestID, PlanFingerprint: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		FrozenPlan: &migrationv1alpha1.NodeMigrationFrozenPlan{
			SourceVCenter: cleanupTestSource, SourceDatacenter: vmotionPlanTestSourceDatacenter, DestinationVCenter: executionTestTargetVC, Mode: migrationv1alpha1.VMotionModeCold,
		},
		ControlPlane: &migrationv1alpha1.RoleMigrationStatus{},
		Workers: &migrationv1alpha1.RoleMigrationStatus{Nodes: []migrationv1alpha1.NodeMigrationProgress{{
			Name: executionTestNode, NodeUID: "node-uid", VMBIOSUUID: executionTestBIOSUUID,
			ProviderID: "vsphere://00000000-0000-4000-8000-000000000001", State: migrationv1alpha1.NodeMigrationStatePreparing,
			TargetFailureDomain: m.Spec.FailureDomains[0].Name, TargetResourcePool: "/TargetDC/host/cluster/Resources",
			TargetDatastore: "/TargetDC/datastore/ds", TargetNetwork: executionTestNetwork,
			ShutdownOutcome: &migrationv1alpha1.NodeMigrationOperationOutcome{
				ActionID: "44444444-4444-4444-8444-444444444444", Kind: migrationv1alpha1.NodeMigrationActionShutdownGuest,
				State: migrationv1alpha1.NodeMigrationOutcomeSucceeded, OwnerVCenter: cleanupTestSource,
			},
		}}},
	}
	m.Spec.NodeMigration.Workers.VMotion.Mode = migrationv1alpha1.VMotionModeCold
	m.Spec.NodeMigration.ControlPlane.VMotion.Mode = migrationv1alpha1.VMotionModeCold
	m.Spec.FailureDomains[0].Topology.Datacenter = executionTestTargetDC
	s := runtime.NewScheme()
	if err := migrationv1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(m).WithStatusSubresource(m).WithInterceptorFuncs(hooks).Build()
	r := &VmwareCloudFoundationMigrationReconciler{Client: c, APIReader: c}
	return r, readExecution(t, r, m)
}

func readExecution(t *testing.T, r *VmwareCloudFoundationMigrationReconciler, m *migrationv1alpha1.VmwareCloudFoundationMigration) *migrationv1alpha1.VmwareCloudFoundationMigration {
	t.Helper()
	got := &migrationv1alpha1.VmwareCloudFoundationMigration{}
	if err := r.Get(context.Background(), client.ObjectKeyFromObject(m), got); err != nil {
		t.Fatal(err)
	}
	return got
}

func acknowledgedRelocation() (vsphere.VMotionSubmission, error) {
	return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionAcknowledged, TaskReference: &vsphere.VMotionTaskReference{OwnerVCenter: cleanupTestSource, TaskID: executionTestTaskID}}, nil
}

func TestVMotionClaimSingleWinner(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	var calls atomic.Int32
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		calls.Add(1)
		latest := readExecution(t, r, m)
		if latest.Status.NodeMigration.ActiveAction == nil {
			t.Error("mutation reached backend before durable claim")
		}
		return acknowledgedRelocation()
	}}
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- r.submitVMotionAction(context.Background(), m.DeepCopy(), executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	got := readExecution(t, r, m).Status.NodeMigration.ActiveAction
	if calls.Load() != 1 || got.TaskReference == nil || got.SubmissionState != migrationv1alpha1.NodeMigrationSubmissionSubmitted {
		t.Fatalf("calls=%d, action=%+v; want one acknowledged submission", calls.Load(), got)
	}
	if got.Deadline == nil || got.Deadline.Sub(got.ClaimedAt.Time) != 6*time.Hour {
		t.Fatalf("claim has no bounded six-hour deadline: %+v", got)
	}
}

func TestVMotionClaimFailureNeverSubmits(t *testing.T) {
	for _, ambiguous := range []bool{false, true} {
		t.Run(fmt.Sprintf("writeMayHaveSucceeded=%t", ambiguous), func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{SubResourceUpdate: func(ctx context.Context, c client.Client, sub string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				if ambiguous {
					if err := c.SubResource(sub).Update(ctx, obj, opts...); err != nil {
						return err
					}
				}
				return fmt.Errorf("lost claim acknowledgement")
			}})
			calls := 0
			b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { calls++; return acknowledgedRelocation() }}
			if err := r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err == nil {
				t.Fatal("expected claim acknowledgement error")
			}
			if calls != 0 {
				t.Fatalf("calls=%d after failed acknowledgement", calls)
			}
		})
	}
}

func TestVMotionClaimConflictRechecksPause(t *testing.T) {
	var once sync.Once
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{SubResourceUpdate: func(ctx context.Context, c client.Client, sub string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
		conflict := false
		once.Do(func() {
			conflict = true
			latest := &migrationv1alpha1.VmwareCloudFoundationMigration{}
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), latest); err != nil {
				t.Fatal(err)
			}
			latest.Spec.State = migrationv1alpha1.MigrationStatePaused
			latest.Generation++
			if err := c.Update(ctx, latest); err != nil {
				t.Fatal(err)
			}
		})
		if conflict {
			return apierrors.NewConflict(schema.GroupResource{Resource: executionTestResource}, obj.GetName(), fmt.Errorf("concurrent pause"))
		}
		return c.SubResource(sub).Update(ctx, obj, opts...)
	}})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		t.Fatal("submitted after pause")
		return acknowledgedRelocation()
	}}
	if err := r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err != nil {
		t.Fatal(err)
	}
	if readExecution(t, r, m).Status.NodeMigration.ActiveAction != nil {
		t.Fatal("claim persisted after concurrent pause")
	}
}

func TestVMotionUnknownSubmissionNeverReplays(t *testing.T) {
	var writes int
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{SubResourceUpdate: func(ctx context.Context, c client.Client, sub string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
		writes++
		if writes == 2 {
			return fmt.Errorf("crash before task persistence")
		}
		return c.SubResource(sub).Update(ctx, obj, opts...)
	}})
	calls := 0
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { calls++; return acknowledgedRelocation() }}
	if err := r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err == nil {
		t.Fatal("expected task-persistence error")
	}
	latest := readExecution(t, r, m)
	if latest.Status.NodeMigration.ActiveAction.TaskReference != nil {
		t.Fatal("task unexpectedly persisted")
	}
	if err := r.submitVMotionAction(context.Background(), latest, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("replayed uncertain action: %d calls", calls)
	}
}

func TestVMotionSubmissionValidation(t *testing.T) {
	for _, tt := range []struct {
		name   string
		result vsphere.VMotionSubmission
		err    error
	}{
		{"uncertain submission", vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil},
		{"wrong owner", vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionAcknowledged, TaskReference: &vsphere.VMotionTaskReference{OwnerVCenter: executionTestTargetVC, TaskID: executionTestTaskID}}, nil},
		{"missing task", vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionAcknowledged}, nil},
		{"rejected", vsphere.VMotionSubmission{}, fmt.Errorf("secret-bearing backend rejection")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
			b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return tt.result, tt.err }}
			if err := r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err != nil {
				t.Fatal(err)
			}
			got := readExecution(t, r, m).Status.NodeMigration
			if got.Intervention == nil {
				t.Fatal("invalid/failed submission did not block for intervention")
			}
			if got.ActiveAction.TaskReference != nil {
				t.Fatal("accepted invalid task reference")
			}
			if tt.err != nil && got.ActiveAction.Reason == tt.err.Error() {
				t.Fatal("stored unsanitized backend error")
			}
		})
	}
}

func claimExecution(t *testing.T, r *VmwareCloudFoundationMigrationReconciler, m *migrationv1alpha1.VmwareCloudFoundationMigration, kind migrationv1alpha1.NodeMigrationActionKind, b *executionBackend) *migrationv1alpha1.VmwareCloudFoundationMigration {
	t.Helper()
	if err := r.submitVMotionAction(context.Background(), m, executionTestNode, kind, b); err != nil {
		t.Fatal(err)
	}
	return readExecution(t, r, m)
}

func TestVMotionObservationSurvivesPauseAndGenerationChange(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() }}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	paused := m.DeepCopy()
	paused.Spec.State = migrationv1alpha1.MigrationStatePaused
	paused.Generation++
	if err := r.Update(context.Background(), paused); err != nil {
		t.Fatal(err)
	}
	b.observe = func(_ context.Context, vc, task string) (vsphere.VMotionTaskObservation, error) {
		if vc != cleanupTestSource || task != executionTestTaskID {
			t.Fatalf("task lookup in wrong scope: %s/%s", vc, task)
		}
		return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskSucceeded}, nil
	}
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	got := readExecution(t, r, m)
	if got.Status.NodeMigration.Workers.Nodes[0].RelocationOutcome.State != migrationv1alpha1.NodeMigrationOutcomeSucceeded {
		t.Fatalf("lost result after pause: %+v", got.Status.NodeMigration)
	}
	if got.Spec.State != migrationv1alpha1.MigrationStatePaused || got.Generation != 2 {
		t.Fatal("observation changed spec")
	}
}

func TestVMotionObservationStaleWriterCannotReplaceTerminalOrNewAction(t *testing.T) {
	for _, newAction := range []bool{false, true} {
		t.Run(fmt.Sprintf("newAction=%t", newAction), func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
			b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() }}
			m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
			b.observe = func(ctx context.Context, _, _ string) (vsphere.VMotionTaskObservation, error) {
				latest := readExecution(t, r, m)
				if newAction {
					latest.Status.NodeMigration.ActiveAction.ActionID = executionTestRecoveryID
				} else {
					latest.Status.NodeMigration.ActiveAction.ObservationState = migrationv1alpha1.NodeMigrationObservationSucceeded
				}
				latest.Status.NodeMigration.Progress = "concurrent unrelated status"
				if err := r.Status().Update(ctx, latest); err != nil {
					t.Fatal(err)
				}
				return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskRunning}, nil
			}
			if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
				t.Fatal(err)
			}
			got := readExecution(t, r, m).Status.NodeMigration
			if newAction && got.ActiveAction.ActionID != executionTestRecoveryID {
				t.Fatal("stale result replaced new action")
			}
			if !newAction && got.ActiveAction.ObservationState != migrationv1alpha1.NodeMigrationObservationSucceeded {
				t.Fatal("stale result overwrote terminal result")
			}
			if got.Progress != "concurrent unrelated status" {
				t.Fatal("lost unrelated status")
			}
		})
	}
}

func TestVMotionShutdownDeadlineAndLateCompletion(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	m.Status.NodeMigration.Workers.Nodes[0].ShutdownOutcome = nil
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionAcknowledged}, nil
	}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionShutdownGuest, b)
	past := metav1.NewTime(time.Now().Add(-time.Minute))
	m.Status.NodeMigration.ActiveAction.Deadline = &past
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	power := "poweredOn"
	b.inspect = func(_ context.Context, vc, dc, id string) (vsphere.VMotionVM, error) {
		if vc != cleanupTestSource || dc != vmotionPlanTestSourceDatacenter || id != executionTestBIOSUUID {
			t.Fatal("shutdown inspected wrong identity/scope")
		}
		return vsphere.VMotionVM{BIOSUUID: id, PowerState: power}, nil
	}
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	m = readExecution(t, r, m)
	if m.Status.NodeMigration.Intervention == nil || m.Status.NodeMigration.Workers.Nodes[0].ShutdownOutcome.State != migrationv1alpha1.NodeMigrationOutcomeUnknown {
		t.Fatal("timeout did not block uncertain shutdown")
	}
	power = "poweredOff"
	priorPoll := metav1.NewTime(time.Now().Add(-31 * time.Second))
	m.Status.NodeMigration.ActiveAction.LastObservedAt = &priorPoll
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	got := readExecution(t, r, m).Status.NodeMigration
	if got.Workers.Nodes[0].ShutdownOutcome.State != migrationv1alpha1.NodeMigrationOutcomeSucceeded || got.Intervention == nil {
		t.Fatal("late completion lost evidence or cleared intervention")
	}
	if err := r.submitVMotionAction(context.Background(), readExecution(t, r, m), executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err != nil {
		t.Fatal(err)
	}
	if readExecution(t, r, m).Status.NodeMigration.ActiveAction.Kind != migrationv1alpha1.NodeMigrationActionShutdownGuest {
		t.Fatal("late completion authorized next action")
	}
}

func TestVMotionUnknownClaimRestartRequiresIntervention(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil
	}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	// Simulate a restart after durable claim but before any submission result.
	m.Status.NodeMigration.ActiveAction.SubmissionState = migrationv1alpha1.NodeMigrationSubmissionClaimed
	m.Status.NodeMigration.ActiveAction.ObservationState = migrationv1alpha1.NodeMigrationObservationNotObserved
	m.Status.NodeMigration.Intervention = nil
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	got := readExecution(t, r, m).Status.NodeMigration
	if got.Intervention == nil || got.ActiveAction.ObservationState != migrationv1alpha1.NodeMigrationObservationUnknown {
		t.Fatal("restart treated unknown claim as replayable")
	}
}

func TestVMotionRelocationSuccessSurvivesPowerOnFailure(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() },
		observe: func(context.Context, string, string) (vsphere.VMotionTaskObservation, error) {
			return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskSucceeded}, nil
		}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	m = readExecution(t, r, m)
	b.submit = func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionAcknowledged, TaskReference: &vsphere.VMotionTaskReference{OwnerVCenter: executionTestTargetVC, TaskID: "task-power"}}, nil
	}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionPowerOnVM, b)
	b.observe = func(_ context.Context, vc, task string) (vsphere.VMotionTaskObservation, error) {
		if vc != executionTestTargetVC || task != "task-power" {
			t.Fatal("power-on task observed in source vCenter")
		}
		return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskFailed}, nil
	}
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	got := readExecution(t, r, m).Status.NodeMigration
	n := got.Workers.Nodes[0]
	if n.RelocationOutcome.State != migrationv1alpha1.NodeMigrationOutcomeSucceeded || n.PowerOnOutcome.State != migrationv1alpha1.NodeMigrationOutcomeFailed || got.Intervention == nil {
		t.Fatalf("independent results lost: %+v", n)
	}
}

func TestVMotionRecoveryIsActionBoundAndCannotBypassPause(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{}, fmt.Errorf("rejected")
	}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	m.Spec.NodeMigration.RecoveryRequest = &migrationv1alpha1.NodeMigrationRecoveryRequest{
		RequestID: executionTestRecoveryID, ExecutionID: m.Status.NodeMigration.ExecutionID, ActionID: "33333333-3333-4333-8333-333333333333",
	}
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	verified := 0
	verify := func(context.Context, *migrationv1alpha1.VmwareCloudFoundationMigration) (migrationv1alpha1.NodeMigrationObservationState, error) {
		verified++
		return migrationv1alpha1.NodeMigrationObservationFailed, nil
	}
	if err := r.recheckVMotionRecovery(context.Background(), m, verify); err == nil {
		t.Fatal("accepted stale request")
	}
	if verified != 0 {
		t.Fatal("ran verifier for wrong action")
	}
	m = readExecution(t, r, m)
	m.Spec.NodeMigration.RecoveryRequest.ActionID = m.Status.NodeMigration.ActiveAction.ActionID
	m.Spec.State = migrationv1alpha1.MigrationStatePaused
	m.Generation++
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if err := r.recheckVMotionRecovery(context.Background(), m, nil); err == nil {
		t.Fatal("cleared intervention without evidence verifier")
	}
	if err := r.recheckVMotionRecovery(context.Background(), m, verify); err != nil {
		t.Fatal(err)
	}
	m = readExecution(t, r, m)
	if m.Status.NodeMigration.Intervention != nil || m.Status.NodeMigration.RecoveryAcknowledgement == nil {
		t.Fatal("verified recovery was not consumed")
	}
	if err := r.recheckVMotionRecovery(context.Background(), m, verify); err != nil {
		t.Fatal(err)
	}
	if verified != 1 {
		t.Fatal("consumed acknowledgement was rechecked")
	}
	if err := r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err != nil {
		t.Fatal(err)
	}
	if readExecution(t, r, m).Status.NodeMigration.ActiveAction.ActionID != m.Status.NodeMigration.ActiveAction.ActionID {
		t.Fatal("recovery bypassed pause")
	}
	oldID := m.Status.NodeMigration.ActiveAction.ActionID
	m.Spec.State = migrationv1alpha1.MigrationStateRunning
	m.Generation++
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	b.submit = func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() }
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	if m.Status.NodeMigration.ActiveAction.ActionID == oldID {
		t.Fatal("authorized retry reused action ID")
	}
	if err := r.recheckVMotionRecovery(context.Background(), m, verify); err == nil {
		t.Fatal("old request authorized new action")
	}
}

func TestVMotionReconcileObservesDuringPauseAndDeletion(t *testing.T) {
	for _, deleting := range []bool{false, true} {
		t.Run(fmt.Sprintf("deleting=%t", deleting), func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
			b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() },
				observe: func(context.Context, string, string) (vsphere.VMotionTaskObservation, error) {
					return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskSucceeded}, nil
				}}
			m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
			m.Spec.State = migrationv1alpha1.MigrationStatePaused
			m.Generation++
			if err := r.Update(context.Background(), m); err != nil {
				t.Fatal(err)
			}
			if deleting {
				if err := r.Delete(context.Background(), m); err != nil {
					t.Fatal(err)
				}
			}
			r.VMotionBackend = b
			if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(m)}); err != nil {
				t.Fatal(err)
			}
			got := readExecution(t, r, m)
			if got.Status.NodeMigration.ActiveAction.ObservationState != migrationv1alpha1.NodeMigrationObservationSucceeded {
				t.Fatal("Reconcile discarded in-flight result")
			}
			if deleting {
				got.Annotations = map[string]string{migrationv1alpha1.AnnotationAllowDeletion: allowDeletionAnnotationValue}
				if err := r.Update(context.Background(), got); err != nil {
					t.Fatal(err)
				}
				if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(m)}); err != nil {
					t.Fatal(err)
				}
				if err := r.Get(context.Background(), client.ObjectKeyFromObject(m), got); !apierrors.IsNotFound(err) {
					t.Fatalf("deliberate handoff did not release finalizer: %v", err)
				}
			}
		})
	}
}

func TestVMotionClaimGuards(t *testing.T) {
	for _, tt := range []struct {
		name   string
		change func(*migrationv1alpha1.VmwareCloudFoundationMigration)
		kind   migrationv1alpha1.NodeMigrationActionKind
	}{
		{"paused", func(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
			m.Spec.State = migrationv1alpha1.MigrationStatePaused
		}, migrationv1alpha1.NodeMigrationActionRelocateVM},
		{"unsupported name", func(m *migrationv1alpha1.VmwareCloudFoundationMigration) { m.Name = "other" }, migrationv1alpha1.NodeMigrationActionRelocateVM},
		{"hot shutdown", func(m *migrationv1alpha1.VmwareCloudFoundationMigration) {
			m.Status.NodeMigration.FrozenPlan.Mode = migrationv1alpha1.VMotionModeHot
			m.Spec.NodeMigration.Workers.VMotion.Mode = migrationv1alpha1.VMotionModeHot
			m.Spec.NodeMigration.ControlPlane.VMotion.Mode = migrationv1alpha1.VMotionModeHot
		}, migrationv1alpha1.NodeMigrationActionShutdownGuest},
		{"power-on before relocation", func(*migrationv1alpha1.VmwareCloudFoundationMigration) {}, migrationv1alpha1.NodeMigrationActionPowerOnVM},
		{"unknown kind", func(*migrationv1alpha1.VmwareCloudFoundationMigration) {}, migrationv1alpha1.NodeMigrationActionKind("invalid")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
			tt.change(m)
			if tt.name == "unsupported name" {
				m.ResourceVersion = ""
				if err := r.Create(context.Background(), m); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := r.Update(context.Background(), m); err != nil {
					t.Fatal(err)
				}
				if err := r.Status().Update(context.Background(), m); err != nil {
					t.Fatal(err)
				}
			}
			b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
				t.Fatal("forbidden VM mutation")
				return acknowledgedRelocation()
			}}
			_ = r.submitVMotionAction(context.Background(), m, executionTestNode, tt.kind, b)
			if readExecution(t, r, m).Status.NodeMigration.ActiveAction != nil {
				t.Fatal("forbidden claim persisted")
			}
		})
	}
}

func TestVMotionPauseAfterWinningClaimStillPersistsSubmission(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(ctx context.Context) (vsphere.VMotionSubmission, error) {
		latest := readExecution(t, r, m)
		latest.Spec.State = migrationv1alpha1.MigrationStatePaused
		latest.Generation++
		if err := r.Update(ctx, latest); err != nil {
			t.Fatal(err)
		}
		return acknowledgedRelocation()
	}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	if m.Status.NodeMigration.ActiveAction.TaskReference == nil || m.Spec.State != migrationv1alpha1.MigrationStatePaused {
		t.Fatal("pause after claim lost submission evidence")
	}
}

func TestVMotionObservationConflictsAndTaskUncertainty(t *testing.T) {
	var conflict atomic.Bool
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{SubResourceUpdate: func(ctx context.Context, c client.Client, sub string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
		if conflict.Swap(false) {
			latest := &migrationv1alpha1.VmwareCloudFoundationMigration{}
			if err := c.Get(ctx, client.ObjectKeyFromObject(obj), latest); err != nil {
				t.Fatal(err)
			}
			latest.Status.NodeMigration.Progress = "keep concurrent progress"
			if err := c.SubResource(sub).Update(ctx, latest); err != nil {
				t.Fatal(err)
			}
			return apierrors.NewConflict(schema.GroupResource{Resource: executionTestResource}, obj.GetName(), fmt.Errorf("concurrent status"))
		}
		return c.SubResource(sub).Update(ctx, obj, opts...)
	}})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() },
		observe: func(context.Context, string, string) (vsphere.VMotionTaskObservation, error) {
			return vsphere.VMotionTaskObservation{}, fmt.Errorf("missing task with secret details")
		}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	conflict.Store(true)
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	got := readExecution(t, r, m).Status.NodeMigration
	if got.Progress != "keep concurrent progress" || got.Intervention == nil || got.Workers.Nodes[0].RelocationOutcome.State != migrationv1alpha1.NodeMigrationOutcomeUnknown {
		t.Fatalf("lost conflict merge/uncertainty: %+v", got)
	}
	if got.ActiveAction.Reason == "missing task with secret details" {
		t.Fatal("persisted raw backend error")
	}
	// A wrong-owner reference is rejected before any task lookup.
	latest := readExecution(t, r, m)
	latest.Status.NodeMigration.ActiveAction.TaskReference.OwnerVCenter = "wrong.example.com"
	if err := r.Status().Update(context.Background(), latest); err != nil {
		t.Fatal(err)
	}
	b.observe = func(context.Context, string, string) (vsphere.VMotionTaskObservation, error) {
		t.Fatal("looked up task on wrong owner")
		return vsphere.VMotionTaskObservation{}, nil
	}
	if err := r.observeVMotionAction(context.Background(), latest, b); err == nil {
		t.Fatal("accepted wrong-owner reference")
	}
}

func TestVMotionRecoveryRejectsUncertaintyAndChangedEvidence(t *testing.T) {
	for _, tt := range []struct {
		name   string
		state  migrationv1alpha1.NodeMigrationObservationState
		change bool
		err    error
	}{
		{"unresolved evidence", migrationv1alpha1.NodeMigrationObservationUnknown, false, nil},
		{"running", migrationv1alpha1.NodeMigrationObservationRunning, false, nil},
		{"read error", "", false, fmt.Errorf("unavailable health")},
		{"concurrent change", migrationv1alpha1.NodeMigrationObservationSucceeded, true, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
			b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
				return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil
			}}
			m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
			m.Spec.NodeMigration.RecoveryRequest = &migrationv1alpha1.NodeMigrationRecoveryRequest{
				RequestID: executionTestRecoveryID, ExecutionID: m.Status.NodeMigration.ExecutionID, ActionID: m.Status.NodeMigration.ActiveAction.ActionID,
			}
			if err := r.Update(context.Background(), m); err != nil {
				t.Fatal(err)
			}
			err := r.recheckVMotionRecovery(context.Background(), m, func(ctx context.Context, snapshot *migrationv1alpha1.VmwareCloudFoundationMigration) (migrationv1alpha1.NodeMigrationObservationState, error) {
				if tt.change {
					snapshot.Spec.State = migrationv1alpha1.MigrationStatePaused
					snapshot.Generation++
					if err := r.Update(ctx, snapshot); err != nil {
						t.Fatal(err)
					}
				}
				return tt.state, tt.err
			})
			if err == nil {
				t.Fatal("recovery accepted incomplete or stale evidence")
			}
			got := readExecution(t, r, m).Status.NodeMigration
			if got.Intervention == nil || got.RecoveryAcknowledgement != nil {
				t.Fatal("recovery cleared uncertainty or consumed invalid evidence")
			}
		})
	}
}

func TestVMotionResultRejectsReplacedExecutionIdentity(t *testing.T) {
	for _, change := range []string{"uid", "execution", "plan", "node"} {
		t.Run(change, func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
			b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() }}
			m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
			b.observe = func(ctx context.Context, _, _ string) (vsphere.VMotionTaskObservation, error) {
				latest := readExecution(t, r, m)
				switch change {
				case "uid":
					latest.UID = "replacement"
					if err := r.Update(ctx, latest); err != nil {
						t.Fatal(err)
					}
				case "execution":
					latest.Status.NodeMigration.ExecutionID = executionTestRecoveryID
				case "plan":
					latest.Status.NodeMigration.PlanFingerprint = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
				case "node":
					latest.Status.NodeMigration.Workers.Nodes[0].NodeUID = "replacement-node"
				}
				if err := r.Status().Update(ctx, latest); err != nil {
					t.Fatal(err)
				}
				return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskSucceeded}, nil
			}
			if err := r.observeVMotionAction(context.Background(), m, b); err == nil {
				t.Fatal("accepted result for replaced execution identity")
			}
			if readExecution(t, r, m).Status.NodeMigration.Workers.Nodes[0].RelocationOutcome != nil {
				t.Fatal("stale result replaced newer execution status")
			}
		})
	}
}

func TestVMotionContradictorySubmissionIsUncertain(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		result, _ := acknowledgedRelocation()
		return result, fmt.Errorf("error with conflicting acknowledgement")
	}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	if m.Status.NodeMigration.Workers.Nodes[0].RelocationOutcome.State != migrationv1alpha1.NodeMigrationOutcomeUnknown {
		t.Fatal("contradictory submission was classified as a definite rejection")
	}
}

func TestVMotionClaimRejectsNodeOutsideFrozenPlan(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		t.Fatal("mutation for unfrozen node")
		return acknowledgedRelocation()
	}}
	if err := r.submitVMotionAction(context.Background(), m, "unfrozen-node", migrationv1alpha1.NodeMigrationActionRelocateVM, b); err == nil {
		t.Fatal("accepted node outside frozen plan")
	}
	if readExecution(t, r, m).Status.NodeMigration.ActiveAction != nil {
		t.Fatal("claimed node outside frozen plan")
	}
}

func TestVMotionCrashAfterClaimBeforeExternalRequest(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { panic("process stopped before request") }}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected simulated crash")
			}
		}()
		_ = r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	}()
	m = readExecution(t, r, m)
	b.submit = func(context.Context) (vsphere.VMotionSubmission, error) {
		t.Fatal("successor replayed claim")
		return acknowledgedRelocation()
	}
	if err := r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err != nil {
		t.Fatal(err)
	}
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	if readExecution(t, r, m).Status.NodeMigration.Intervention == nil {
		t.Fatal("ambiguous crash window did not retain intervention")
	}
}

func TestVMotionDeletionHandoffDoesNotRequireBackend(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil
	}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	m.Annotations = map[string]string{migrationv1alpha1.AnnotationAllowDeletion: allowDeletionAnnotationValue}
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if err := r.Delete(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(m)}); err != nil {
		t.Fatal(err)
	}
	got := &migrationv1alpha1.VmwareCloudFoundationMigration{}
	if err := r.Get(context.Background(), client.ObjectKeyFromObject(m), got); !apierrors.IsNotFound(err) {
		t.Fatalf("handoff retained record: %v", err)
	}
}

func TestVMotionRecoveryCannotOutrunOriginalClaimant(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	entered, release := make(chan struct{}), make(chan struct{})
	var externalRequests atomic.Int32
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		close(entered)
		<-release
		externalRequests.Add(1)
		return acknowledgedRelocation()
	}}
	done := make(chan error, 1)
	go func() {
		done <- r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	}()
	<-entered
	defer func() {
		close(release)
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	m = readExecution(t, r, m)
	if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
		t.Fatal(err)
	}
	m = readExecution(t, r, m)
	m.Spec.NodeMigration.RecoveryRequest = &migrationv1alpha1.NodeMigrationRecoveryRequest{
		RequestID: executionTestRecoveryID, ExecutionID: m.Status.NodeMigration.ExecutionID, ActionID: m.Status.NodeMigration.ActiveAction.ActionID,
	}
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	verified := false
	err := r.recheckVMotionRecovery(context.Background(), m, func(context.Context, *migrationv1alpha1.VmwareCloudFoundationMigration) (migrationv1alpha1.NodeMigrationObservationState, error) {
		verified = true
		return migrationv1alpha1.NodeMigrationObservationFailed, nil
	})
	if err == nil || verified {
		t.Fatal("recovery accepted external inactivity while the original claimant could still submit")
	}
	if err := r.submitVMotionAction(context.Background(), readExecution(t, r, m), executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b); err != nil {
		t.Fatal(err)
	}
	if externalRequests.Load() != 0 || readExecution(t, r, m).Status.NodeMigration.RecoveryAcknowledgement != nil {
		t.Fatal("recovery admitted a second submitter")
	}
}

func TestVMotionObservationTimeoutStillAllowsDeletionHandoff(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
		b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() },
			observe: func(ctx context.Context, _, _ string) (vsphere.VMotionTaskObservation, error) {
				if _, bounded := ctx.Deadline(); !bounded {
					t.Error("observation has no bounded context")
					return vsphere.VMotionTaskObservation{}, fmt.Errorf("unbounded read")
				}
				<-ctx.Done()
				return vsphere.VMotionTaskObservation{}, ctx.Err()
			}}
		m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
		r.VMotionBackend = b
		started := time.Now()
		if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(m)}); err != nil {
			t.Fatal(err)
		}
		if elapsed := time.Since(started); elapsed != 30*time.Second {
			t.Fatalf("observation returned after %s, want bounded 30-second read", elapsed)
		}
		m = readExecution(t, r, m)
		if m.Status.NodeMigration.Intervention == nil {
			t.Fatal("timed-out read did not preserve uncertainty")
		}
		m.Annotations = map[string]string{migrationv1alpha1.AnnotationAllowDeletion: allowDeletionAnnotationValue}
		if err := r.Update(context.Background(), m); err != nil {
			t.Fatal(err)
		}
		if err := r.Delete(context.Background(), m); err != nil {
			t.Fatal(err)
		}
		if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(m)}); err != nil {
			t.Fatal(err)
		}
		if err := r.Get(context.Background(), client.ObjectKeyFromObject(m), m); !apierrors.IsNotFound(err) {
			t.Fatalf("handoff could not release after timeout: %v", err)
		}
	})
}

func TestVMotionRecoveryVerificationIsBounded(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
		b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
			return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil
		}}
		m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
		m.Spec.NodeMigration.RecoveryRequest = &migrationv1alpha1.NodeMigrationRecoveryRequest{RequestID: executionTestRecoveryID, ExecutionID: m.Status.NodeMigration.ExecutionID, ActionID: m.Status.NodeMigration.ActiveAction.ActionID}
		if err := r.Update(context.Background(), m); err != nil {
			t.Fatal(err)
		}
		started := time.Now()
		err := r.recheckVMotionRecovery(context.Background(), m, func(ctx context.Context, _ *migrationv1alpha1.VmwareCloudFoundationMigration) (migrationv1alpha1.NodeMigrationObservationState, error) {
			if _, bounded := ctx.Deadline(); !bounded {
				t.Error("recovery verifier has no bounded context")
				return "", fmt.Errorf("unbounded verification")
			}
			<-ctx.Done()
			return "", ctx.Err()
		})
		if err == nil || time.Since(started) != 30*time.Second {
			t.Fatalf("verification was not bounded: %v after %s", err, time.Since(started))
		}
		got := readExecution(t, r, m).Status.NodeMigration
		if got.Intervention == nil || got.RecoveryAcknowledgement != nil {
			t.Fatal("timed-out verification consumed recovery")
		}
	})
}

func TestVMotionStatusWatchDoesNotCreatePollingLoop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
		polls := 0
		b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() },
			observe: func(context.Context, string, string) (vsphere.VMotionTaskObservation, error) {
				polls++
				time.Sleep(2 * time.Second)
				return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskRunning}, nil
			}}
		m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
		r.VMotionBackend = b
		watcher, err := r.Client.(client.WithWatch).Watch(context.Background(), &migrationv1alpha1.VmwareCloudFoundationMigrationList{}, client.InNamespace(m.Namespace))
		if err != nil {
			t.Fatal(err)
		}
		defer watcher.Stop()
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(m)}
		if _, err := r.Reconcile(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		select {
		case <-watcher.ResultChan():
		default:
			t.Fatal("first observation did not publish a status event")
		}
		// Feed our own status event back to the reconciler, as the real watch does.
		if result, err := r.Reconcile(context.Background(), req); err != nil || result.RequeueAfter != 30*time.Second {
			t.Fatalf("self-event reconcile: %+v, %v", result, err)
		}
		if polls != 1 {
			t.Fatalf("self-generated status event triggered %d polls, want one", polls)
		}
		select {
		case <-watcher.ResultChan():
			t.Fatal("self-event created another immediate status event")
		default:
		}
		time.Sleep(30 * time.Second)
		if _, err := r.Reconcile(context.Background(), req); err != nil {
			t.Fatal(err)
		}
		if polls != 2 {
			t.Fatal("throttling prevented the next scheduled observation")
		}
	})
}

func TestVMotionExpiredClaimStillNeedsAuthoritativeRecovery(t *testing.T) {
	r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) {
		return vsphere.VMotionSubmission{State: vsphere.VMotionSubmissionUnknown}, nil
	}}
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	oldID := m.Status.NodeMigration.ActiveAction.ActionID
	m.Status.NodeMigration.ActiveAction.SubmissionState = migrationv1alpha1.NodeMigrationSubmissionClaimed
	past := metav1.NewTime(time.Now().Add(-time.Minute))
	m.Status.NodeMigration.ActiveAction.Deadline = &past
	if err := r.Status().Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	m.Spec.NodeMigration.RecoveryRequest = &migrationv1alpha1.NodeMigrationRecoveryRequest{RequestID: executionTestRecoveryID, ExecutionID: m.Status.NodeMigration.ExecutionID, ActionID: oldID}
	if err := r.Update(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if err := r.recheckVMotionRecovery(context.Background(), m, nil); err == nil {
		t.Fatal("deadline expiry alone authorized recovery")
	}
	if err := r.recheckVMotionRecovery(context.Background(), m, func(context.Context, *migrationv1alpha1.VmwareCloudFoundationMigration) (migrationv1alpha1.NodeMigrationObservationState, error) {
		return migrationv1alpha1.NodeMigrationObservationFailed, nil
	}); err != nil {
		t.Fatal(err)
	}
	m = readExecution(t, r, m)
	b.submit = func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() }
	m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	if m.Status.NodeMigration.ActiveAction.ActionID == oldID {
		t.Fatal("verified retry did not receive a new action ID")
	}
}

func TestVMotionTaskLateSuccessRetainsDeadlineIntervention(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
		state := vsphere.VMotionTaskRunning
		b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() },
			observe: func(context.Context, string, string) (vsphere.VMotionTaskObservation, error) {
				return vsphere.VMotionTaskObservation{State: state}, nil
			}}
		m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
		if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
			t.Fatal(err)
		}
		time.Sleep(6*time.Hour + time.Second)
		m = readExecution(t, r, m)
		state = vsphere.VMotionTaskSucceeded
		if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
			t.Fatal(err)
		}
		got := readExecution(t, r, m).Status.NodeMigration
		if got.Workers.Nodes[0].RelocationOutcome.State != migrationv1alpha1.NodeMigrationOutcomeSucceeded || got.Intervention == nil {
			t.Fatal("late task success lost evidence or authorized progression")
		}
	})
}

func TestVMotionColdRelocationRequiresSuccessfulShutdown(t *testing.T) {
	for _, state := range []migrationv1alpha1.NodeMigrationOutcomeState{"", migrationv1alpha1.NodeMigrationOutcomeFailed, migrationv1alpha1.NodeMigrationOutcomeUnknown, migrationv1alpha1.NodeMigrationOutcomeSucceeded} {
		t.Run(string(state), func(t *testing.T) {
			r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
			n := &m.Status.NodeMigration.Workers.Nodes[0]
			n.ShutdownOutcome = nil
			if state != "" {
				n.ShutdownOutcome = &migrationv1alpha1.NodeMigrationOperationOutcome{ActionID: executionTestRecoveryID, Kind: migrationv1alpha1.NodeMigrationActionShutdownGuest, State: state, OwnerVCenter: cleanupTestSource}
			}
			if err := r.Status().Update(context.Background(), m); err != nil {
				t.Fatal(err)
			}
			calls := 0
			b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { calls++; return acknowledgedRelocation() }}
			err := r.submitVMotionAction(context.Background(), m, executionTestNode, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
			if state == migrationv1alpha1.NodeMigrationOutcomeSucceeded {
				if err != nil || calls != 1 {
					t.Fatalf("successful shutdown did not permit relocation: calls=%d, err=%v", calls, err)
				}
			} else if err == nil || calls != 0 || readExecution(t, r, m).Status.NodeMigration.ActiveAction != nil {
				t.Fatalf("Cold relocation claimed without shutdown success: calls=%d, err=%v", calls, err)
			}
		})
	}
}
