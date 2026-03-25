package openshift

import (
	"context"
	"fmt"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// ExcludedOperators is the set of ClusterOperator names that should be skipped
// when checking cluster-wide operator health.
var ExcludedOperators = map[string]bool{
	"machine-config": true,
}

// OperatorManager provides health-check operations for OpenShift ClusterOperators.
type OperatorManager struct {
	client configclient.Interface
}

// NewOperatorManager creates a new OperatorManager with the given config client.
func NewOperatorManager(client configclient.Interface) *OperatorManager {
	return &OperatorManager{client: client}
}

// CheckAllOperatorsHealthy lists all ClusterOperators and returns whether every
// non-excluded operator is healthy. An operator is considered healthy when its
// Available condition is True and its Degraded condition is not True.
// The names of any unhealthy operators are returned in the unhealthyOperators slice.
func (o *OperatorManager) CheckAllOperatorsHealthy(ctx context.Context) (healthy bool, unhealthyOperators []string, err error) {
	log := klog.FromContext(ctx)
	log.V(2).Info("checking all cluster operators for health")

	operators, err := o.client.ConfigV1().ClusterOperators().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, nil, fmt.Errorf("listing cluster operators: %w", err)
	}

	for i := range operators.Items {
		co := &operators.Items[i]
		if ExcludedOperators[co.Name] {
			log.V(3).Info("skipping excluded operator", "operator", co.Name)
			continue
		}

		if !isOperatorHealthy(co) {
			unhealthyOperators = append(unhealthyOperators, co.Name)
		}
	}

	if len(unhealthyOperators) > 0 {
		log.V(2).Info("unhealthy operators found", "operators", strings.Join(unhealthyOperators, ", "))
		return false, unhealthyOperators, nil
	}

	log.V(2).Info("all cluster operators are healthy")
	return true, nil, nil
}

// GetOperator retrieves a single ClusterOperator by name.
func (o *OperatorManager) GetOperator(ctx context.Context, name string) (*configv1.ClusterOperator, error) {
	co, err := o.client.ConfigV1().ClusterOperators().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting cluster operator %q: %w", name, err)
	}
	return co, nil
}

// IsOperatorHealthy checks whether a single ClusterOperator is healthy and returns
// a human-readable message describing the operator's state. An operator is healthy
// when Available is True and Degraded is not True.
func (o *OperatorManager) IsOperatorHealthy(ctx context.Context, name string) (healthy bool, message string, err error) {
	co, err := o.GetOperator(ctx, name)
	if err != nil {
		return false, "", err
	}

	if isOperatorHealthy(co) {
		return true, fmt.Sprintf("operator %q is healthy", name), nil
	}

	var parts []string
	for _, cond := range co.Status.Conditions {
		if cond.Type == configv1.OperatorAvailable && cond.Status != configv1.ConditionTrue {
			parts = append(parts, fmt.Sprintf("Available=%s (%s)", cond.Status, cond.Message))
		}
		if cond.Type == configv1.OperatorDegraded && cond.Status == configv1.ConditionTrue {
			parts = append(parts, fmt.Sprintf("Degraded=True (%s)", cond.Message))
		}
	}

	return false, fmt.Sprintf("operator %q is unhealthy: %s", name, strings.Join(parts, "; ")), nil
}

// isOperatorHealthy returns true when the ClusterOperator has Available=True
// and Degraded is not True.
func isOperatorHealthy(co *configv1.ClusterOperator) bool {
	available := false
	degraded := false

	for _, cond := range co.Status.Conditions {
		switch cond.Type {
		case configv1.OperatorAvailable:
			available = cond.Status == configv1.ConditionTrue
		case configv1.OperatorDegraded:
			degraded = cond.Status == configv1.ConditionTrue
		}
	}

	return available && !degraded
}
