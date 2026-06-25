package openshift

import (
	"context"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekube "k8s.io/client-go/kubernetes/fake"

	"gopkg.in/yaml.v2"
	vsphereconfig "k8s.io/cloud-provider-vsphere/pkg/common/config"
)

func newCloudProviderConfigMap(servers map[string][]string) *corev1.ConfigMap {
	cfg := &cloudProviderConfigYAML{
		Global: vsphereconfig.GlobalYAML{
			SecretName:      "vsphere-creds",
			SecretNamespace: "kube-system",
		},
		Vcenter: make(map[string]*vsphereconfig.VirtualCenterConfigYAML),
	}

	for server, datacenters := range servers {
		cfg.Vcenter[server] = &vsphereconfig.VirtualCenterConfigYAML{
			VCenterIP:    server,
			VCenterPort:  443,
			InsecureFlag: true,
			Datacenters:  datacenters,
		}
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CloudProviderConfigMapName,
			Namespace: CloudProviderConfigMapNamespace,
		},
		Data: map[string]string{
			cloudProviderConfigKey: string(data),
		},
	}
}

func TestParseCloudProviderConfig(t *testing.T) {
	tests := []struct {
		name            string
		cm              *corev1.ConfigMap
		wantErr         bool
		wantVCenterKeys []string
		wantSecretName  string
	}{
		{
			name: "valid single vcenter",
			cm: newCloudProviderConfigMap(map[string][]string{
				"source.example.com": {"DC0"},
			}),
			wantVCenterKeys: []string{"source.example.com"},
			wantSecretName:  "vsphere-creds",
		},
		{
			name:    "nil configmap",
			cm:      nil,
			wantErr: true,
		},
		{
			name: "missing config key",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: CloudProviderConfigMapName},
				Data:       map[string]string{"other-key": "value"},
			},
			wantErr: true,
		},
		{
			name: "invalid yaml",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: CloudProviderConfigMapName},
				Data:       map[string]string{cloudProviderConfigKey: ":::invalid"},
			},
			wantErr: true,
		},
		{
			name: "empty vcenter map is initialized",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: CloudProviderConfigMapName},
				Data:       map[string]string{cloudProviderConfigKey: "global: {}"},
			},
			wantVCenterKeys: nil,
			wantSecretName:  "vsphere-creds",
		},
		{
			name: "injects default secret when globals empty",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: CloudProviderConfigMapName},
				Data: map[string]string{
					cloudProviderConfigKey: "vcenter:\n  srv1:\n    server: srv1\n    datacenters:\n      - DC0",
				},
			},
			wantVCenterKeys: []string{"srv1"},
			wantSecretName:  "vsphere-creds",
		},
		{
			name: "preserves existing secret ref",
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: CloudProviderConfigMapName},
				Data: map[string]string{
					cloudProviderConfigKey: "global:\n  secretName: custom-secret\n  secretNamespace: custom-ns",
				},
			},
			wantSecretName: "custom-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseCloudProviderConfig(tt.cm)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseCloudProviderConfig error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if cfg.Vcenter == nil {
				t.Fatal("Vcenter map is nil, expected initialized map")
			}

			if len(cfg.Vcenter) != len(tt.wantVCenterKeys) {
				t.Fatalf("vcenter count = %d, want %d", len(cfg.Vcenter), len(tt.wantVCenterKeys))
			}
			for _, key := range tt.wantVCenterKeys {
				if _, ok := cfg.Vcenter[key]; !ok {
					t.Fatalf("expected vcenter key %q not found", key)
				}
			}

			if cfg.Global.SecretName != tt.wantSecretName {
				t.Fatalf("Global.SecretName = %q, want %q", cfg.Global.SecretName, tt.wantSecretName)
			}
		})
	}
}

func TestGetCloudProviderConfig(t *testing.T) {
	tests := []struct {
		name    string
		cm      *corev1.ConfigMap
		wantErr bool
	}{
		{
			name: "returns existing configmap",
			cm: newCloudProviderConfigMap(map[string][]string{
				"source.example.com": {"DC0"},
			}),
		},
		{
			name:    "errors when configmap does not exist",
			cm:      nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *fakekube.Clientset
			if tt.cm != nil {
				client = fakekube.NewClientset(tt.cm)
			} else {
				client = fakekube.NewClientset()
			}
			mgr := NewConfigMapManager(client)

			cm, err := mgr.GetCloudProviderConfig(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetCloudProviderConfig error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if cm.Name != CloudProviderConfigMapName {
				t.Fatalf("configmap name = %q, want %q", cm.Name, CloudProviderConfigMapName)
			}
		})
	}
}

func TestAddTargetVCenterToConfig(t *testing.T) {
	tests := []struct {
		name            string
		existingServers map[string][]string
		failureDomains  []configv1.VSpherePlatformFailureDomainSpec
		wantServerCount int
		wantServers     []string
	}{
		{
			name: "adds new vcenter from failure domain",
			existingServers: map[string][]string{
				"source.example.com": {"DC0"},
			},
			failureDomains: []configv1.VSpherePlatformFailureDomainSpec{
				{
					Name:   "fd-target",
					Server: "target.example.com",
					Topology: configv1.VSpherePlatformTopology{
						Datacenter: "DC1",
					},
				},
			},
			wantServerCount: 2,
			wantServers:     []string{"source.example.com", "target.example.com"},
		},
		{
			name: "skips vcenter that already exists",
			existingServers: map[string][]string{
				"source.example.com": {"DC0"},
			},
			failureDomains: []configv1.VSpherePlatformFailureDomainSpec{
				{
					Name:   "fd-source",
					Server: "source.example.com",
					Topology: configv1.VSpherePlatformTopology{
						Datacenter: "DC0",
					},
				},
			},
			wantServerCount: 1,
			wantServers:     []string{"source.example.com"},
		},
		{
			name: "multiple failure domains on same target server",
			existingServers: map[string][]string{
				"source.example.com": {"DC0"},
			},
			failureDomains: []configv1.VSpherePlatformFailureDomainSpec{
				{
					Name:   "fd-a",
					Server: "target.example.com",
					Topology: configv1.VSpherePlatformTopology{
						Datacenter: "DC1",
					},
				},
				{
					Name:   "fd-b",
					Server: "target.example.com",
					Topology: configv1.VSpherePlatformTopology{
						Datacenter: "DC2",
					},
				},
			},
			wantServerCount: 2,
			wantServers:     []string{"source.example.com", "target.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := newCloudProviderConfigMap(tt.existingServers)
			client := fakekube.NewClientset(cm)
			mgr := NewConfigMapManager(client)

			updated, err := mgr.AddTargetVCenterToConfig(context.Background(), cm, tt.failureDomains)
			if err != nil {
				t.Fatalf("AddTargetVCenterToConfig error = %v", err)
			}

			cfg, err := parseCloudProviderConfig(updated)
			if err != nil {
				t.Fatalf("parseCloudProviderConfig error = %v", err)
			}

			if len(cfg.Vcenter) != tt.wantServerCount {
				t.Fatalf("vcenter count = %d, want %d", len(cfg.Vcenter), tt.wantServerCount)
			}

			for _, server := range tt.wantServers {
				if _, ok := cfg.Vcenter[server]; !ok {
					t.Fatalf("expected server %q not found in config", server)
				}
			}
		})
	}
}

func TestAddTargetVCenterToConfigErrors(t *testing.T) {
	t.Run("errors when config key is missing", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CloudProviderConfigMapName,
				Namespace: CloudProviderConfigMapNamespace,
			},
			Data: map[string]string{"wrong-key": "data"},
		}
		client := fakekube.NewClientset(cm)
		mgr := NewConfigMapManager(client)

		_, err := mgr.AddTargetVCenterToConfig(context.Background(), cm, []configv1.VSpherePlatformFailureDomainSpec{
			{Name: "fd", Server: "s.example.com", Topology: configv1.VSpherePlatformTopology{Datacenter: "DC0"}},
		})
		if err == nil {
			t.Fatal("expected error when config key is missing")
		}
	})

	t.Run("errors when configmap data is nil", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CloudProviderConfigMapName,
				Namespace: CloudProviderConfigMapNamespace,
			},
		}
		client := fakekube.NewClientset(cm)
		mgr := NewConfigMapManager(client)

		_, err := mgr.AddTargetVCenterToConfig(context.Background(), cm, []configv1.VSpherePlatformFailureDomainSpec{
			{Name: "fd", Server: "s.example.com", Topology: configv1.VSpherePlatformTopology{Datacenter: "DC0"}},
		})
		if err == nil {
			t.Fatal("expected error when configmap data is nil")
		}
	})
}

func TestRemoveSourceVCenterFromConfig(t *testing.T) {
	tests := []struct {
		name            string
		existingServers map[string][]string
		removeServer    string
		wantServerCount int
	}{
		{
			name: "removes existing source vcenter",
			existingServers: map[string][]string{
				"source.example.com": {"DC0"},
				"target.example.com": {"DC1"},
			},
			removeServer:    "source.example.com",
			wantServerCount: 1,
		},
		{
			name: "no-ops when server not in config",
			existingServers: map[string][]string{
				"source.example.com": {"DC0"},
			},
			removeServer:    "nonexistent.example.com",
			wantServerCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := newCloudProviderConfigMap(tt.existingServers)
			client := fakekube.NewClientset(cm)
			mgr := NewConfigMapManager(client)

			updated, err := mgr.RemoveSourceVCenterFromConfig(context.Background(), cm, tt.removeServer)
			if err != nil {
				t.Fatalf("RemoveSourceVCenterFromConfig error = %v", err)
			}

			cfg, err := parseCloudProviderConfig(updated)
			if err != nil {
				t.Fatalf("parseCloudProviderConfig error = %v", err)
			}

			if len(cfg.Vcenter) != tt.wantServerCount {
				t.Fatalf("vcenter count = %d, want %d", len(cfg.Vcenter), tt.wantServerCount)
			}

			if _, exists := cfg.Vcenter[tt.removeServer]; exists && tt.removeServer != "nonexistent.example.com" {
				t.Fatalf("server %q should have been removed", tt.removeServer)
			}
		})
	}
}

func TestRemoveSourceVCenterFromConfigErrors(t *testing.T) {
	t.Run("errors when config key is missing", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CloudProviderConfigMapName,
				Namespace: CloudProviderConfigMapNamespace,
			},
			Data: map[string]string{"wrong-key": "data"},
		}
		client := fakekube.NewClientset(cm)
		mgr := NewConfigMapManager(client)

		_, err := mgr.RemoveSourceVCenterFromConfig(context.Background(), cm, "source.example.com")
		if err == nil {
			t.Fatal("expected error when config key is missing")
		}
	})
}
