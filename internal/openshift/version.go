package openshift

import (
	"context"
	"fmt"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
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
func GetVSphereMultiVCenterSupport(ctx context.Context, client configclient.Interface) (*VSphereMultiVCenterSupport, error) {
	if client == nil {
		return nil, fmt.Errorf("GetVSphereMultiVCenterSupport client must not be nil")
	}

	clusterVersion, err := client.ConfigV1().ClusterVersions().Get(ctx, clusterVersionName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting clusterversion %q: %w", clusterVersionName, err)
	}

	version := strings.TrimSpace(clusterVersion.Status.Desired.Version)
	if version == "" {
		return nil, fmt.Errorf("clusterversion %q has empty status.desired.version", clusterVersionName)
	}

	featureGate, err := client.ConfigV1().FeatureGates().Get(ctx, featureGateName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting featuregate %q: %w", featureGateName, err)
	}

	enabled, err := isFeatureGateEnabledForVersion(featureGate, version, featureGateVSphereMultiVCenterDay2)
	if err != nil {
		return nil, fmt.Errorf("checking feature gate %q for version %q: %w", featureGateVSphereMultiVCenterDay2, version, err)
	}

	return &VSphereMultiVCenterSupport{
		ClusterVersion:     version,
		FeatureGateEnabled: enabled,
		UpgradeInProgress:  isClusterVersionProgressing(clusterVersion),
	}, nil
}

func isFeatureGateEnabledForVersion(featureGate *configv1.FeatureGate, version string, gateName configv1.FeatureGateName) (bool, error) {
	if featureGate == nil {
		return false, fmt.Errorf("featuregate must not be nil")
	}

	for i := range featureGate.Status.FeatureGates {
		gateVersion := &featureGate.Status.FeatureGates[i]
		if gateVersion.Version != version {
			continue
		}
		for _, enabled := range gateVersion.Enabled {
			if enabled.Name == gateName {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("no feature gate status found for version %q", version)
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
