/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/openshift/vcf-migration-operator/internal/execution"
)

const (
	// mcoNamespace is the namespace containing Machine Config Operator pods.
	mcoNamespace = "openshift-machine-config-operator"
	// mcoPodPrefix is the pod name prefix for the machine-config-operator pod.
	mcoPodPrefix = "machine-config-operator-"
	// eventReasonControlPlaneRollout is the event reason while the control plane rolls out.
	eventReasonControlPlaneRollout = "ControlPlaneRollout"
)

// syncControllerConfig restarts Machine Config Operator pods by deleting pods
// in the openshift-machine-config-operator namespace with the well-known prefix.
// This forces the MCO to pick up updated configuration.
func syncControllerConfig(ctx context.Context, kubeClient kubernetes.Interface) error {
	log := klog.FromContext(ctx)
	log.V(1).Info("restarting MCO pods to sync controller config")

	pods, err := kubeClient.CoreV1().Pods(mcoNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("listing pods in %s: %w", mcoNamespace, err)
	}

	deleted := 0
	for i := range pods.Items {
		pod := &pods.Items[i]
		if len(pod.Name) >= len(mcoPodPrefix) && pod.Name[:len(mcoPodPrefix)] == mcoPodPrefix {
			if err := execution.Check(ctx); err != nil {
				return err
			}
			if err := kubeClient.CoreV1().Pods(mcoNamespace).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
				log.V(2).Info("failed to delete MCO pod", "pod", pod.Name, "err", err)
				continue
			}
			deleted++
			log.V(2).Info("deleted MCO pod", "pod", pod.Name)
		}
	}

	log.V(1).Info("MCO pod restart complete", "deleted", deleted)
	return nil
}
