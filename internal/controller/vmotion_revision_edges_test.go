package controller

import (
	"context"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

func TestVMotionReviewExpiredReadCannotCommitSuccess(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
		b := &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() }, observe: func(ctx context.Context, _, _ string) (vsphere.VMotionTaskObservation, error) {
			<-ctx.Done()
			return vsphere.VMotionTaskObservation{State: vsphere.VMotionTaskSucceeded}, nil
		}}
		m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
		if err := r.observeVMotionAction(context.Background(), m, b); err != nil {
			t.Fatal(err)
		}
		if readExecution(t, r, m).Status.NodeMigration.ActiveAction.ObservationState != migrationv1alpha1.NodeMigrationObservationUnknown {
			t.Fatal("expired read committed success")
		}
	})
}

func TestVMotionReviewOutcomeInvalidIdentity(t *testing.T) {
	for _, missing := range []bool{true, false} {
		r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
		m = claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, &executionBackend{submit: func(context.Context) (vsphere.VMotionSubmission, error) { return acknowledgedRelocation() }})
		if missing {
			m.Status.NodeMigration.Workers.Nodes = nil
		} else {
			m.Status.NodeMigration.ActiveAction.Kind = "invalid"
		}
		recordVMotionOutcome(m.Status.NodeMigration, metav1.Now())
	}
}

func TestVMotionReviewDeadlineWirePrecision(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		time.Sleep(900 * time.Millisecond)
		r, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
		b := &executionBackend{submit: func(ctx context.Context) (vsphere.VMotionSubmission, error) {
			deadline, ok := ctx.Deadline()
			if !ok || deadline.Nanosecond() != 0 {
				t.Fatalf("claim deadline not wire precision: %s", deadline)
			}
			return acknowledgedRelocation()
		}}
		claimExecution(t, r, m, migrationv1alpha1.NodeMigrationActionRelocateVM, b)
	})
}

func TestVMotionReviewLegacyPlanRequiresUpgrade(t *testing.T) {
	_, m := newVMotionExecutionFixture(t, interceptor.Funcs{})
	candidate := m.Status.NodeMigration.DeepCopy()
	m.Status.NodeMigration.Workers.Nodes[0].TargetDatacenter = ""
	err := validateVMotionPlanStability(m.Status.NodeMigration, candidate)
	if err == nil || !strings.Contains(err.Error(), "requires upgrade/replan") || strings.Contains(err.Error(), "drifted") {
		t.Fatalf("legacy plan reported as drift: %v", err)
	}
}
