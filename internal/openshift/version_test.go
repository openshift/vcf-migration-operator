package openshift

import (
	"context"
	"reflect"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	configclient "github.com/openshift/client-go/config/clientset/versioned"
	fakeconfigclient "github.com/openshift/client-go/config/clientset/versioned/fake"
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

			got, err := GetVSphereMultiVCenterSupport(context.Background(), client)
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
	tests := []struct {
		name   string
		client func() configclient.Interface
	}{
		{
			name: "fails when cluster version missing",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset(newFeatureGateForVersion("5.0.0", true))
			},
		},
		{
			name: "fails when client is nil",
			client: func() configclient.Interface {
				return nil
			},
		},
		{
			name: "fails when feature gate missing",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset(&configv1.ClusterVersion{
					ObjectMeta: metav1.ObjectMeta{Name: clusterVersionName},
					Status: configv1.ClusterVersionStatus{
						Desired: configv1.Release{Version: "5.0.0"},
					},
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.client()
			if _, err := GetVSphereMultiVCenterSupport(context.Background(), client); err == nil {
				t.Fatal("GetVSphereMultiVCenterSupport succeeded, want error")
			}
		})
	}
}

func TestIsFeatureGateEnabledForVersion(t *testing.T) {
	tests := []struct {
		name    string
		gate    *configv1.FeatureGate
		version string
		want    bool
		wantErr bool
	}{
		{
			name:    "returns true when gate enabled for version",
			gate:    newFeatureGateForVersion("5.0.0", true),
			version: "5.0.0",
			want:    true,
		},
		{
			name:    "returns false when gate not enabled for version",
			gate:    newFeatureGateForVersion("5.0.0", false),
			version: "5.0.0",
			want:    false,
		},
		{
			name:    "returns error when version not found in gate status",
			gate:    newFeatureGateForVersion("5.0.0", true),
			version: "4.19.0",
			wantErr: true,
		},
		{
			name:    "returns error when featuregate is nil",
			gate:    nil,
			version: "5.0.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isFeatureGateEnabledForVersion(tt.gate, tt.version, featureGateVSphereMultiVCenterDay2)
			if (err != nil) != tt.wantErr {
				t.Fatalf("isFeatureGateEnabledForVersion error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("isFeatureGateEnabledForVersion = %v, want %v", got, tt.want)
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

func TestGetFeatureSet(t *testing.T) {
	customSource := &configv1.CustomFeatureGates{
		Enabled:  []configv1.FeatureGateName{"ExampleOn"},
		Disabled: []configv1.FeatureGateName{"ExampleOff"},
	}
	emptyCustomSource := &configv1.CustomFeatureGates{
		Enabled:  []configv1.FeatureGateName{},
		Disabled: []configv1.FeatureGateName{},
	}

	tests := []struct {
		name         string
		client       func() configclient.Interface
		sourceCustom *configv1.CustomFeatureGates
		wantSet      configv1.FeatureSet
		wantCustom   *configv1.CustomFeatureGates
		wantErr      bool
	}{
		{
			name: "returns default feature set",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset(newFeatureGateWithSpec(configv1.FeatureSet(""), nil))
			},
			wantSet:    configv1.FeatureSet(""),
			wantCustom: nil,
		},
		{
			name: "returns custom no upgrade feature set",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset(newFeatureGateWithSpec(configv1.CustomNoUpgrade, customSource))
			},
			sourceCustom: customSource,
			wantSet:      configv1.CustomNoUpgrade,
			wantCustom: &configv1.CustomFeatureGates{
				Enabled:  []configv1.FeatureGateName{"ExampleOn"},
				Disabled: []configv1.FeatureGateName{"ExampleOff"},
			},
		},
		{
			name: "preserves non nil empty custom feature gate slices",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset(newFeatureGateWithSpec(configv1.CustomNoUpgrade, emptyCustomSource))
			},
			sourceCustom: emptyCustomSource,
			wantSet:      configv1.CustomNoUpgrade,
			wantCustom: &configv1.CustomFeatureGates{
				Enabled:  []configv1.FeatureGateName{},
				Disabled: []configv1.FeatureGateName{},
			},
		},
		{
			name: "fails when feature gate missing",
			client: func() configclient.Interface {
				return fakeconfigclient.NewClientset()
			},
			wantErr: true,
		},
		{
			name: "fails when client is nil",
			client: func() configclient.Interface {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSet, gotCustom, err := GetFeatureSet(context.Background(), tt.client())
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetFeatureSet error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if gotSet != tt.wantSet {
				t.Fatalf("FeatureSet = %q, want %q", gotSet, tt.wantSet)
			}
			if tt.sourceCustom != nil && gotCustom == tt.sourceCustom {
				t.Fatalf("GetFeatureSet returned source custom gates pointer %p; want defensive clone", gotCustom)
			}
			if !reflect.DeepEqual(gotCustom, tt.wantCustom) {
				t.Fatalf("CustomFeatureGates = %#v, want %#v", gotCustom, tt.wantCustom)
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

func newFeatureGateWithSpec(featureSet configv1.FeatureSet, custom *configv1.CustomFeatureGates) *configv1.FeatureGate {
	return &configv1.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{Name: featureGateName},
		Spec: configv1.FeatureGateSpec{
			FeatureGateSelection: configv1.FeatureGateSelection{
				FeatureSet:      featureSet,
				CustomNoUpgrade: custom,
			},
		},
	}
}
