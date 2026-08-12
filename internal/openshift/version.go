package openshift

import (
	"context"
	"fmt"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterVersionName                                          = "version"
	featureGateName                                             = "cluster"
	featureGateVSphereMultiVCenterDay2 configv1.FeatureGateName = "VSphereMultiVCenterDay2"
)

// VSphereMultiVCenterSupport captures the cluster version and feature-gate state
// relevant to migration preflight validation.
type VSphereMultiVCenterSupport struct {
	ClusterVersion     string
	FeatureGateEnabled bool
	UpgradeInProgress  bool
}

// GetVSphereMultiVCenterSupport returns the cluster version, feature-gate, and
// upgrade status needed for migration preflight checks.
func GetVSphereMultiVCenterSupport(ctx context.Context, client configclient.Interface, featureGateAccessor featuregates.FeatureGateAccess) (*VSphereMultiVCenterSupport, error) {
	if client == nil {
		return nil, fmt.Errorf("GetVSphereMultiVCenterSupport client must not be nil")
	}
	if featureGateAccessor == nil {
		return nil, fmt.Errorf("GetVSphereMultiVCenterSupport featureGateAccessor must not be nil")
	}

	clusterVersion, err := client.ConfigV1().ClusterVersions().Get(ctx, clusterVersionName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting clusterversion %q: %w", clusterVersionName, err)
	}

	version := strings.TrimSpace(clusterVersion.Status.Desired.Version)
	if version == "" {
		return nil, fmt.Errorf("clusterversion %q has empty status.desired.version", clusterVersionName)
	}

	currentFeatureGates, err := featureGateAccessor.CurrentFeatureGates()
	if err != nil {
		return nil, fmt.Errorf("getting current feature gates: %w", err)
	}

	return &VSphereMultiVCenterSupport{
		ClusterVersion:     version,
		FeatureGateEnabled: currentFeatureGates.Enabled(featureGateVSphereMultiVCenterDay2),
		UpgradeInProgress:  isClusterVersionProgressing(clusterVersion),
	}, nil
}

func isClusterVersionProgressing(clusterVersion *configv1.ClusterVersion) bool {
	if clusterVersion == nil {
		return false
	}

	for _, condition := range clusterVersion.Status.Conditions {
		if condition.Type == configv1.OperatorProgressing {
			return condition.Status == configv1.ConditionTrue
		}
	}

	return false
}
