package controller

import (
	"context"
	"fmt"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	machinev1 "github.com/openshift/api/machine/v1"
	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	machinefake "github.com/openshift/client-go/machine/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/openshift"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

func TestVMotionMachinePhaseClassification(t *testing.T) {
	for _, phase := range []*string{nil, ptr.To(""), ptr.To("FuturePhase"), ptr.To(machinev1beta1.PhaseProvisioning), ptr.To(machinev1beta1.PhaseProvisioned), ptr.To(machinev1beta1.PhaseDeleting), ptr.To(machinev1beta1.PhaseRunning), ptr.To(machinev1beta1.PhaseFailed)} {
		name := "absent"
		if phase != nil {
			name = *phase
		}
		t.Run(name, func(t *testing.T) {
			node := newVMotionPlanNode("upi-node", vmotionWorkerRole, "vsphere://11111111-1111-4111-8111-111111111111", "upi-uid", false)
			ids, err := collectVMotionNodeIdentities([]corev1.Node{node})
			if err != nil {
				t.Fatal(err)
			}
			machine := machinev1beta1.Machine{ObjectMeta: metav1.ObjectMeta{Name: "history", Namespace: openshift.MachineAPINamespace}, Status: machinev1beta1.MachineStatus{Phase: phase}}
			err = validateVMotionMachineEligibility(ids, []machinev1beta1.Machine{machine})
			if phase != nil && *phase == machinev1beta1.PhaseFailed {
				if err != nil {
					t.Fatalf("unassociated terminal history should remain a candidate: %v", err)
				}
			} else if err == nil {
				t.Fatal("uninitialized, unsettled or inconsistent Machine was considered quiescent")
			}
		})
	}
}

func quiescentVMotionMachineSet() machinev1beta1.MachineSet {
	return machinev1beta1.MachineSet{
		ObjectMeta: metav1.ObjectMeta{Name: "workers", Namespace: openshift.MachineAPINamespace, UID: "set-uid", Generation: 2},
		Spec:       machinev1beta1.MachineSetSpec{Replicas: ptr.To(int32(0)), Selector: metav1.LabelSelector{MatchLabels: map[string]string{"pool": "workers"}}, Template: machinev1beta1.MachineTemplateSpec{ObjectMeta: machinev1beta1.ObjectMeta{Labels: map[string]string{"pool": "workers"}}}},
		Status:     machinev1beta1.MachineSetStatus{ObservedGeneration: 2},
	}
}

func TestVMotionMachineSetQuiescence(t *testing.T) {
	for _, tt := range []struct {
		name     string
		mutate   func(*machinev1beta1.MachineSet)
		children []machinev1beta1.Machine
		wantErr  bool
	}{
		{"unused zero replicas", func(*machinev1beta1.MachineSet) {}, nil, false},
		{"before child exists", func(ms *machinev1beta1.MachineSet) { ms.Spec.Replicas = ptr.To(int32(1)) }, nil, true},
		{"generation not observed", func(ms *machinev1beta1.MachineSet) { ms.Status.ObservedGeneration = 1 }, nil, true},
		{"unknown desired count", func(ms *machinev1beta1.MachineSet) { ms.Spec.Replicas = nil }, nil, true},
		{"empty diagnostics", func(ms *machinev1beta1.MachineSet) {
			ms.Status.ErrorReason = ptr.To(machinev1beta1.MachineSetStatusError(""))
			ms.Status.ErrorMessage = ptr.To("")
		}, nil, false},
		{"empty selector", func(ms *machinev1beta1.MachineSet) { ms.Spec.Selector = metav1.LabelSelector{} }, nil, true},
		{"template selector mismatch", func(ms *machinev1beta1.MachineSet) { ms.Spec.Template.Labels = nil }, nil, true},
		{"deleting parent", func(ms *machinev1beta1.MachineSet) { now := metav1.Now(); ms.DeletionTimestamp = &now }, nil, true},
		{"child absent from zero status", func(*machinev1beta1.MachineSet) {}, []machinev1beta1.Machine{{ObjectMeta: metav1.ObjectMeta{Name: "pending", Namespace: openshift.MachineAPINamespace, Labels: map[string]string{"pool": "workers"}}}}, true},
		{"foreign selector child", func(*machinev1beta1.MachineSet) {}, []machinev1beta1.Machine{{ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: openshift.MachineAPINamespace, Labels: map[string]string{"pool": "other"}}}}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ms := quiescentVMotionMachineSet()
			tt.mutate(&ms)
			err := validateVMotionProvisioning(tt.children, []machinev1beta1.MachineSet{ms}, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("quiescence error=%v, wantErr=%t", err, tt.wantErr)
			}
		})
	}
}

func TestVMotionParentStatusRequiresActualConvergedChildren(t *testing.T) {
	for _, name := range []string{"settled", "missing child", "deleting child", "failed child", "stale generation", "stale ready", "stale condition", "progressing", "controller error"} {
		t.Run(name, func(t *testing.T) {
			ms := quiescentVMotionMachineSet()
			ms.Spec.Replicas = ptr.To(int32(3))
			ms.Status.Replicas, ms.Status.ReadyReplicas, ms.Status.AvailableReplicas = 3, 3, 3
			cpms := machinev1.ControlPlaneMachineSet{ObjectMeta: metav1.ObjectMeta{Name: "cluster", Namespace: ms.Namespace, UID: "active-cpms", Generation: 2}, Spec: machinev1.ControlPlaneMachineSetSpec{State: machinev1.ControlPlaneMachineSetStateActive, Replicas: ptr.To(int32(3)), Selector: ms.Spec.Selector}, Status: machinev1.ControlPlaneMachineSetStatus{ObservedGeneration: 2, Replicas: 3, ReadyReplicas: 3, UpdatedReplicas: 3, Conditions: []metav1.Condition{{Type: "Available", Status: metav1.ConditionTrue, ObservedGeneration: 2}, {Type: "Progressing", Status: metav1.ConditionFalse, ObservedGeneration: 2}, {Type: "Degraded", Status: metav1.ConditionFalse, ObservedGeneration: 2}}}}
			var children []machinev1beta1.Machine
			for i := range 3 {
				children = append(children, machinev1beta1.Machine{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("child-%d", i), Namespace: ms.Namespace, Labels: ms.Spec.Selector.MatchLabels}, Status: machinev1beta1.MachineStatus{Phase: ptr.To(machinev1beta1.PhaseRunning)}})
			}
			switch name {
			case "missing child":
				children = children[:2]
			case "deleting child":
				now := metav1.Now()
				children[0].DeletionTimestamp = &now
			case "failed child":
				children[0].Status.Phase = ptr.To(machinev1beta1.PhaseFailed)
			case "stale generation":
				ms.Status.ObservedGeneration = 1
				cpms.Status.ObservedGeneration = 1
			case "stale ready":
				ms.Status.ReadyReplicas = 2
				cpms.Status.ReadyReplicas = 2
			case "stale condition":
				cpms.Status.Conditions[0].ObservedGeneration = 1
			case "progressing":
				cpms.Status.Conditions[1].Status = metav1.ConditionTrue
			case "controller error":
				cpms.Status.Conditions = append(cpms.Status.Conditions, metav1.Condition{Type: "Error", Status: metav1.ConditionTrue, ObservedGeneration: 2})
			}
			if name != "stale condition" && name != "progressing" && name != "controller error" {
				err := validateVMotionProvisioning(children, []machinev1beta1.MachineSet{ms}, nil, nil)
				if (err != nil) != (name != "settled") {
					t.Fatalf("MachineSet status/raw membership disagreement: %v", err)
				}
			}
			err := validateVMotionProvisioning(children, nil, []machinev1.ControlPlaneMachineSet{cpms}, nil)
			if (err != nil) != (name != "settled") {
				t.Fatalf("CPMS status/raw membership disagreement: %v", err)
			}
		})
	}
}

func TestVMotionPendingParentBlocksFreezeWithoutMachine(t *testing.T) {
	h := newVMotionPlanHarness(t)
	ms := quiescentVMotionMachineSet()
	ms.Spec.Replicas = ptr.To(int32(1))
	if _, err := h.reconciler.MachineClient.MachineV1beta1().MachineSets(ms.Namespace).Create(context.Background(), &ms, metav1.CreateOptions{}); err != nil {
		t.Fatal(err)
	}
	h.reconciler.MachineClient.(*machinefake.Clientset).ClearActions()
	h.reconcile(t)
	h.reconcile(t)
	if h.migration.Status.NodeMigration != nil || h.inventoryCalls != 0 {
		t.Fatal("pending parent created a frozen plan before any child existed")
	}
	condition := apimeta.FindStatusCondition(h.migration.Status.Conditions, migrationv1alpha1.ConditionInfrastructurePrepared)
	if condition == nil || !strings.Contains(condition.Message, "MachineSet") {
		t.Fatalf("missing actionable parent diagnostic: %+v", condition)
	}
	for _, action := range h.reconciler.MachineClient.(*machinefake.Clientset).Actions() {
		if action.GetVerb() != "list" && action.GetVerb() != "get" {
			t.Fatalf("quiescence mutated Machine API resources: %s %s", action.GetVerb(), action.GetResource().Resource)
		}
	}
}

func TestVMotionParentAndRemediationContext(t *testing.T) {
	failed := machinev1beta1.Machine{ObjectMeta: metav1.ObjectMeta{Name: "old", Namespace: openshift.MachineAPINamespace, Labels: map[string]string{"pool": "history"}}, Status: machinev1beta1.MachineStatus{Phase: ptr.To(machinev1beta1.PhaseFailed)}}
	inactive := machinev1.ControlPlaneMachineSet{ObjectMeta: metav1.ObjectMeta{Name: "cluster", Namespace: openshift.MachineAPINamespace, UID: "cpms-uid"}, Spec: machinev1.ControlPlaneMachineSetSpec{State: machinev1.ControlPlaneMachineSetStateInactive}}
	for _, name := range []string{"unowned failed history", "inactive cpms history", "unresolved parent", "external remediation", "owned remediation", "unrelated healthcheck", "internal remediation cannot delete unowned failed history", "active cpms without children"} {
		t.Run(name, func(t *testing.T) {
			m := *failed.DeepCopy()
			cpms := *inactive.DeepCopy()
			checks := []machinev1beta1.MachineHealthCheck{{ObjectMeta: metav1.ObjectMeta{Name: "health", Namespace: openshift.MachineAPINamespace}, Spec: machinev1beta1.MachineHealthCheckSpec{Selector: metav1.LabelSelector{MatchLabels: map[string]string{"pool": "history"}}}}}
			wantErr := false
			switch name {
			case "unowned failed history":
				checks = nil
			case "inactive cpms history":
				m.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(&cpms, machinev1.GroupVersion.WithKind("ControlPlaneMachineSet"))}
				checks = nil
			case "unresolved parent":
				m.OwnerReferences = []metav1.OwnerReference{{Kind: "MachineSet", APIVersion: machinev1beta1.GroupVersion.String(), Name: "missing", UID: "missing-uid", Controller: ptr.To(true)}}
				wantErr = true
			case "external remediation":
				checks[0].Spec.RemediationTemplate = &corev1.ObjectReference{Name: "repair"}
				wantErr = true
			case "owned remediation":
				m.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(&cpms, machinev1.GroupVersion.WithKind("ControlPlaneMachineSet"))}
				wantErr = true
			case "unrelated healthcheck":
				checks[0].Spec.Selector.MatchLabels["pool"] = "other"
			case "active cpms without children":
				cpms.Spec.State = machinev1.ControlPlaneMachineSetStateActive
				cpms.Spec.Replicas = ptr.To(int32(3))
				wantErr = true
			}
			err := validateVMotionProvisioning([]machinev1beta1.Machine{m}, nil, []machinev1.ControlPlaneMachineSet{cpms}, checks)
			if (err != nil) != wantErr {
				t.Fatalf("parent/remediation error=%v, wantErr=%t", err, wantErr)
			}
		})
	}
}

func TestVMotionPlanningReadsQuiescenceBeforeInventory(t *testing.T) {
	for _, resource := range []string{"machinesets", "controlplanemachinesets", "machinehealthchecks"} {
		t.Run(resource, func(t *testing.T) {
			h := newVMotionPlanHarness(t)
			client := h.reconciler.MachineClient.(*machinefake.Clientset)
			client.PrependReactor("list", resource, func(ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, fmt.Errorf("required %s read failed", resource)
			})
			h.reconcile(t)
			if _, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: h.key}); err == nil {
				t.Fatal("required quiescence read error was swallowed")
			}
			if h.inventoryCalls != 0 {
				t.Fatal("inventory was called without complete quiescence reads")
			}
			got := &migrationv1alpha1.VmwareCloudFoundationMigration{}
			if err := h.apiClient.Get(context.Background(), h.key, got); err != nil {
				t.Fatal(err)
			}
			cond := apimeta.FindStatusCondition(got.Status.Conditions, migrationv1alpha1.ConditionInfrastructurePrepared)
			if got.Status.NodeMigration != nil || cond == nil || !strings.Contains(cond.Message, resource) {
				t.Fatalf("read failure did not preserve blocked planning: %+v", got.Status)
			}
		})
	}
}

func TestVMotionRechecksProvisioningAfterInventory(t *testing.T) {
	h := newVMotionPlanHarness(t)
	list := h.reconciler.vmotionInventoryFunc
	h.reconciler.vmotionInventoryFunc = func(ctx context.Context, vc *configv1.VSpherePlatformVCenterSpec, fds []configv1.VSpherePlatformFailureDomainSpec, ids []string) (vsphere.VMotionInventory, error) {
		inventory, err := list(ctx, vc, fds, ids)
		if err != nil {
			return inventory, err
		}
		pending := &machinev1beta1.Machine{ObjectMeta: metav1.ObjectMeta{Name: "new-before-phase", Namespace: openshift.MachineAPINamespace}}
		if _, err := h.reconciler.MachineClient.MachineV1beta1().Machines(pending.Namespace).Create(ctx, pending, metav1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
		return inventory, nil
	}
	h.reconcile(t)
	h.reconcile(t)
	if h.migration.Status.NodeMigration != nil {
		t.Fatal("froze a plan using quiescence evidence from before the inventory call")
	}
	cond := apimeta.FindStatusCondition(h.migration.Status.Conditions, migrationv1alpha1.ConditionInfrastructurePrepared)
	if cond == nil || !strings.Contains(cond.Message, "new-before-phase") {
		t.Fatalf("lost fresh provisioning evidence: %+v", cond)
	}
}

func TestVMotionHistoricalControllerIdentityMustResolveExactly(t *testing.T) {
	for _, field := range []string{"unchanged", "UID", "name", "kind", "version"} {
		t.Run(field, func(t *testing.T) {
			ms := quiescentVMotionMachineSet()
			owner := *metav1.NewControllerRef(&ms, machinev1beta1.GroupVersion.WithKind("MachineSet"))
			switch field {
			case "UID":
				owner.UID = "old-parent-uid"
			case "name":
				owner.Name = "other-parent"
			case "kind":
				owner.Kind = "OtherController"
			case "version":
				owner.APIVersion = machinev1.GroupVersion.String()
			}
			m := machinev1beta1.Machine{ObjectMeta: metav1.ObjectMeta{Name: "history", Namespace: ms.Namespace, OwnerReferences: []metav1.OwnerReference{owner}}, Status: machinev1beta1.MachineStatus{Phase: ptr.To(machinev1beta1.PhaseFailed)}}
			err := validateVMotionProvisioning([]machinev1beta1.Machine{m}, []machinev1beta1.MachineSet{ms}, nil, nil)
			if (err != nil) != (field != "unchanged") {
				t.Fatalf("controller identity %s: %v", field, err)
			}
		})
	}
}
