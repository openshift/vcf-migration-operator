package controller

import (
	"context"
	"fmt"
	"strings"
	"testing"

	machinev1beta1 "github.com/openshift/api/machine/v1beta1"
	machinefake "github.com/openshift/client-go/machine/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/openshift/vcf-migration-operator/internal/openshift"
)

func TestVMotionIPOnlyAndContradictoryAssociationsAreIndeterminate(t *testing.T) {
	for _, name := range []string{"IPv4 only", "IPv6 only", "annotation versus providerID", "annotation versus NodeRef", "annotation with malformed providerID", "strong identity with matching IP"} {
		t.Run(name, func(t *testing.T) {
			providerID := "vsphere://11111111-1111-4111-8111-111111111111"
			node := newVMotionPlanNode("upi-node", vmotionWorkerRole, providerID, "upi-uid", false)
			ip := "192.0.2.10"
			if name == "IPv6 only" {
				ip = "2001:db8::10"
			}
			node.Status.Addresses = []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: ip}}
			m := machinev1beta1.Machine{ObjectMeta: metav1.ObjectMeta{Name: "historical-machine", Namespace: openshift.MachineAPINamespace}, Status: machinev1beta1.MachineStatus{Phase: ptr.To(machinev1beta1.PhaseFailed)}}
			want := "indeterminate"
			switch name {
			case "IPv4 only", "IPv6 only":
				m.Status.Addresses = node.Status.Addresses
			case "annotation versus providerID":
				node = withMachineAnnotation(node, m.Name)
				m.Spec.ProviderID = ptr.To("vsphere://22222222-2222-4222-8222-222222222222")
			case "annotation versus NodeRef":
				node = withMachineAnnotation(node, m.Name)
				m.Status.NodeRef = &corev1.ObjectReference{Name: "other-node", UID: "old-node-uid"}
			case "annotation with malformed providerID":
				node = withMachineAnnotation(node, m.Name)
				m.Spec.ProviderID = ptr.To("not-a-vsphere-identity")
			case "strong identity with matching IP":
				m.Spec.ProviderID = &providerID
				m.Status.Addresses = node.Status.Addresses
				want = "Machine-managed"
			}
			ids, err := collectVMotionNodeIdentities([]corev1.Node{node})
			if err != nil {
				t.Fatal(err)
			}
			if err := validateVMotionMachineEligibility(ids, []machinev1beta1.Machine{m}); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("eligibility error=%v, want %q", err, want)
			}
		})
	}
}

func TestVMotionTerminatingNodeCannotBePlanned(t *testing.T) {
	node := newVMotionPlanNode("terminating", vmotionWorkerRole, "vsphere://11111111-1111-4111-8111-111111111111", "node-uid", false)
	now := metav1.Now()
	node.DeletionTimestamp = &now
	if _, err := collectVMotionNodeIdentities([]corev1.Node{node}); err == nil || !strings.Contains(err.Error(), "terminating") {
		t.Fatalf("terminating Node was eligible: %v", err)
	}
}

func TestVMotionResolvedAbsentAnnotationRequiresCompleteReads(t *testing.T) {
	for _, failReads := range []bool{false, true} {
		t.Run(fmt.Sprintf("machineReadFailure=%t", failReads), func(t *testing.T) {
			h := newVMotionPlanHarness(t)
			node, err := h.kubeClient.CoreV1().Nodes().Get(context.Background(), "worker-a", metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			node.Annotations = map[string]string{vmotionMachineAnnotation: "openshift-machine-api/absent-machine"}
			if _, err := h.kubeClient.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{}); err != nil {
				t.Fatal(err)
			}
			if failReads {
				h.reconciler.MachineClient.(*machinefake.Clientset).PrependReactor("list", "machines", func(ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, fmt.Errorf("Machine read unavailable")
				})
				h.reconcile(t)
				if _, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: h.key}); err == nil {
					t.Fatal("failed read became resolved absence")
				}
				if h.inventoryCalls != 0 {
					t.Fatal("read failure was allowed past eligibility")
				}
			} else {
				h.reconcileUntilFrozen(t)
			}
			got, err := h.kubeClient.CoreV1().Nodes().Get(context.Background(), node.Name, metav1.GetOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if got.Annotations[vmotionMachineAnnotation] != "openshift-machine-api/absent-machine" {
				t.Fatal("planning changed the historical annotation")
			}
		})
	}
}
