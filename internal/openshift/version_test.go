package openshift

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	fakeconfigclient "github.com/openshift/client-go/config/clientset/versioned/fake"
	"github.com/openshift/library-go/pkg/operator/configobserver/featuregates"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetVSphereMultiVCenterSupport(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		enabled         bool
		progressing     bool
		wantVersion     string
		wantGate        bool
		wantProgressing bool
	}{
		{
			name:        "reports gate enabled",
			version:     "5.0.0",
			enabled:     true,
			wantVersion: "5.0.0",
			wantGate:    true,
		},
		{
			name:        "reports gate disabled",
			version:     "5.0.1",
			enabled:     false,
			wantVersion: "5.0.1",
			wantGate:    false,
		},
		{
			name:        "reports gate enabled on 4.x",
			version:     "4.19.0",
			enabled:     true,
			wantVersion: "4.19.0",
			wantGate:    true,
		},
		{
			name:            "reports when cluster version is progressing",
			version:         "5.0.0",
			enabled:         true,
			progressing:     true,
			wantVersion:     "5.0.0",
			wantGate:        true,
			wantProgressing: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterVersion := &configv1.ClusterVersion{
				ObjectMeta: metav1.ObjectMeta{Name: clusterVersionName},
				Status: configv1.ClusterVersionStatus{
					Desired: configv1.Release{Version: tt.version},
				},
			}
			if tt.progressing {
				clusterVersion.Status.Conditions = append(clusterVersion.Status.Conditions, configv1.ClusterOperatorStatusCondition{
					Type:   configv1.OperatorProgressing,
					Status: configv1.ConditionTrue,
				})
			}
			featureGate := newFeatureGateForVersion(tt.version, tt.enabled)
			client := fakeconfigclient.NewClientset(clusterVersion, featureGate)
			enabled := []configv1.FeatureGateName{}
			disabled := []configv1.FeatureGateName{}
			if tt.enabled {
				enabled = append(enabled, featureGateVSphereMultiVCenterDay2)
			} else {
				disabled = append(disabled, featureGateVSphereMultiVCenterDay2)
			}

			got, err := GetVSphereMultiVCenterSupport(context.Background(), client, featuregates.NewHardcodedFeatureGateAccess(enabled, disabled))
			if err != nil {
				t.Fatalf("GetVSphereMultiVCenterSupport: %v", err)
			}
			if got.ClusterVersion != tt.wantVersion {
				t.Fatalf("ClusterVersion = %q, want %q", got.ClusterVersion, tt.wantVersion)
			}
			if got.FeatureGateEnabled != tt.wantGate {
				t.Fatalf("FeatureGateEnabled = %v, want %v", got.FeatureGateEnabled, tt.wantGate)
			}
			if got.UpgradeInProgress != tt.wantProgressing {
				t.Fatalf("UpgradeInProgress = %v, want %v", got.UpgradeInProgress, tt.wantProgressing)
			}
		})
	}
}

func TestGetVSphereMultiVCenterSupportErrors(t *testing.T) {
	readErr := errors.New("boom")
	closed := make(chan struct{})
	close(closed)

	tests := []struct {
		name                string
		client              func() configclient.Interface
		featureGateAccessor featuregates.FeatureGateAccess
	}{
		{
			name: "fails when cluster version missing",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset(newFeatureGateForVersion("5.0.0", true))
			},
			featureGateAccessor: featuregates.NewHardcodedFeatureGateAccess([]configv1.FeatureGateName{featureGateVSphereMultiVCenterDay2}, nil),
		},
		{
			name: "fails when client is nil",
			client: func() configclient.Interface {
				return nil
			},
			featureGateAccessor: featuregates.NewHardcodedFeatureGateAccess([]configv1.FeatureGateName{featureGateVSphereMultiVCenterDay2}, nil),
		},
		{
			name: "fails when accessor is nil",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset(&configv1.ClusterVersion{
					ObjectMeta: metav1.ObjectMeta{Name: clusterVersionName},
					Status: configv1.ClusterVersionStatus{
						Desired: configv1.Release{Version: "5.0.0"},
					},
				})
			},
		},
		{
			name: "fails when current feature gates cannot be read",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset(&configv1.ClusterVersion{
					ObjectMeta: metav1.ObjectMeta{Name: clusterVersionName},
					Status: configv1.ClusterVersionStatus{
						Desired: configv1.Release{Version: "5.0.0"},
					},
				})
			},
			featureGateAccessor: featuregates.NewHardcodedFeatureGateAccessForTesting(nil, nil, closed, readErr),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.client()
			if _, err := GetVSphereMultiVCenterSupport(context.Background(), client, tt.featureGateAccessor); err == nil {
				t.Fatal("GetVSphereMultiVCenterSupport succeeded, want error")
			}
		})
	}
}

func TestIsClusterVersionProgressing(t *testing.T) {
	tests := []struct {
		name           string
		clusterVersion *configv1.ClusterVersion
		want           bool
	}{
		{
			name: "returns true when progressing is true",
			clusterVersion: &configv1.ClusterVersion{
				Status: configv1.ClusterVersionStatus{
					Conditions: []configv1.ClusterOperatorStatusCondition{
						{Type: configv1.OperatorProgressing, Status: configv1.ConditionTrue},
					},
				},
			},
			want: true,
		},
		{
			name: "returns false when progressing is false",
			clusterVersion: &configv1.ClusterVersion{
				Status: configv1.ClusterVersionStatus{
					Conditions: []configv1.ClusterOperatorStatusCondition{
						{Type: configv1.OperatorProgressing, Status: configv1.ConditionFalse},
					},
				},
			},
			want: false,
		},
		{
			name: "returns false when no progressing condition exists",
			clusterVersion: &configv1.ClusterVersion{
				Status: configv1.ClusterVersionStatus{
					Conditions: []configv1.ClusterOperatorStatusCondition{
						{Type: configv1.OperatorAvailable, Status: configv1.ConditionTrue},
					},
				},
			},
			want: false,
		},
		{
			name:           "returns false when clusterversion is nil",
			clusterVersion: nil,
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClusterVersionProgressing(tt.clusterVersion)
			if got != tt.want {
				t.Fatalf("isClusterVersionProgressing = %v, want %v", got, tt.want)
			}
		})
	}
}

func newFeatureGateForVersion(version string, enabled bool) *configv1.FeatureGate {
	details := configv1.FeatureGateDetails{Version: version}
	if enabled {
		details.Enabled = []configv1.FeatureGateAttributes{
			{Name: featureGateVSphereMultiVCenterDay2},
		}
	}

	return &configv1.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{Name: featureGateName},
		Status: configv1.FeatureGateStatus{
			FeatureGates: []configv1.FeatureGateDetails{details},
		},
	}
}
