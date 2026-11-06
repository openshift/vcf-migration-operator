package controller

import (
	"context"
	"reflect"
	"strings"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	migrationv1alpha1 "github.com/openshift/vcf-migration-operator/api/v1alpha1"
	"github.com/openshift/vcf-migration-operator/internal/vsphere"
)

func TestVMotionInventoryRetriesNodeIdentityDrift(t *testing.T) {
	for _, change := range []string{"addition", "removal", "uid", "providerID", "role"} {
		for _, frozen := range []bool{false, true} {
			t.Run(change+map[bool]string{false: "/initial", true: "/frozen"}[frozen], func(t *testing.T) {
				h := newVMotionPlanHarness(t)
				if frozen {
					h.reconcileUntilFrozen(t)
				} else {
					h.reconcile(t)
				}
				original := h.migration.Status.NodeMigration.DeepCopy()
				list := h.reconciler.vmotionInventoryFunc
				h.reconciler.vmotionInventoryFunc = func(ctx context.Context, vc *configv1.VSpherePlatformVCenterSpec, fds []configv1.VSpherePlatformFailureDomainSpec, ids []string) (vsphere.VMotionInventory, error) {
					inventory, err := list(ctx, vc, fds, ids)
					node, getErr := h.kubeClient.CoreV1().Nodes().Get(ctx, "worker-a", metav1.GetOptions{})
					if getErr != nil {
						t.Fatal(getErr)
					}
					switch change {
					case "addition":
						added := newVMotionPlanNode("worker-c", "worker", h.providerIDs[5], "new-node", false)
						_, getErr = h.kubeClient.CoreV1().Nodes().Create(ctx, &added, metav1.CreateOptions{})
						h.expectedInventoryCount++
					case "removal":
						getErr = h.kubeClient.CoreV1().Nodes().Delete(ctx, node.Name, metav1.DeleteOptions{})
						h.expectedInventoryCount--
					case "uid":
						node.UID = "replacement-node"
					case "providerID":
						node.Spec.ProviderID = h.providerIDs[5]
					case "role":
						other, e := h.kubeClient.CoreV1().Nodes().Get(ctx, "control-plane-a", metav1.GetOptions{})
						if e != nil {
							t.Fatal(e)
						}
						node.Labels, other.Labels = other.Labels, node.Labels
						_, getErr = h.kubeClient.CoreV1().Nodes().Update(ctx, other, metav1.UpdateOptions{})
					}
					if getErr != nil {
						t.Fatal(getErr)
					}
					if change != "addition" && change != "removal" {
						if _, e := h.kubeClient.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{}); e != nil {
							t.Fatal(e)
						}
					}
					return inventory, err
				}
				_, err := h.reconciler.Reconcile(context.Background(), ctrl.Request{NamespacedName: h.key})
				if err == nil || !strings.Contains(err.Error(), "identities changed during inventory") {
					t.Fatalf("Reconcile = %v, want identity drift retry", err)
				}
				if err := h.apiClient.Get(context.Background(), h.key, h.migration); err != nil {
					t.Fatal(err)
				}
				condition := apimeta.FindStatusCondition(h.migration.Status.Conditions, migrationv1alpha1.ConditionInfrastructurePrepared)
				if condition == nil || condition.Reason != migrationv1alpha1.ReasonRetrying {
					t.Fatalf("condition = %+v, want Retrying", condition)
				}
				if !reflect.DeepEqual(original, h.migration.Status.NodeMigration) {
					t.Fatal("identity drift changed frozen plan")
				}
				h.reconciler.vmotionInventoryFunc = list
				if !frozen {
					h.reconcileUntilFrozen(t)
				}
			})
		}
	}
}

func TestVMotionInventoryIgnoresNodeListOrder(t *testing.T) {
	h := newVMotionPlanHarness(t)
	h.reconcile(t)
	list := h.reconciler.vmotionInventoryFunc
	h.reconciler.vmotionInventoryFunc = func(ctx context.Context, vc *configv1.VSpherePlatformVCenterSpec, fds []configv1.VSpherePlatformFailureDomainSpec, ids []string) (vsphere.VMotionInventory, error) {
		inventory, err := list(ctx, vc, fds, ids)
		nodes, e := h.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for i := range nodes.Items {
			node := &nodes.Items[len(nodes.Items)-1-i]
			if e := h.kubeClient.CoreV1().Nodes().Delete(ctx, node.Name, metav1.DeleteOptions{}); e != nil {
				t.Fatal(e)
			}
			if _, e := h.kubeClient.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{}); e != nil {
				t.Fatal(e)
			}
		}
		return inventory, err
	}
	h.reconcileUntilFrozen(t)
	if h.migration.Status.NodeMigration.Workers.Total != 2 {
		t.Fatal("list reordering changed plan")
	}
}
