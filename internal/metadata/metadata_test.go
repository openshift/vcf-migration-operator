package metadata

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	fakekubeclient "k8s.io/client-go/kubernetes/fake"
)

func TestGenerateMetadata(t *testing.T) {
	tests := []struct {
		name           string
		failureDomains []configv1.VSpherePlatformFailureDomainSpec
		infra          *configv1.Infrastructure
		credentials    map[string]string
		wantErr        bool
		wantInfraID    string
		wantVCCount    int
		wantPrimary    string
	}{
		{
			name:    "nil infrastructure",
			infra:   nil,
			wantErr: true,
		},
		{
			name: "empty InfrastructureName",
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{InfrastructureName: ""},
			},
			wantErr: true,
		},
		{
			name: "single failure domain",
			failureDomains: []configv1.VSpherePlatformFailureDomainSpec{
				{
					Name:   "fd1",
					Server: "vcenter.example.com",
					Topology: configv1.VSpherePlatformTopology{
						Datacenter:     "dc1",
						ComputeCluster: "/dc1/host/cluster1",
						Datastore:      "/dc1/datastore/ds1",
						Networks:       []string{"VM Network"},
						ResourcePool:   "/dc1/host/cluster1/Resources",
						Folder:         "/dc1/vm/my-cluster",
					},
				},
			},
			infra: &configv1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{Name: "cluster", UID: types.UID("test-uid")},
				Status:     configv1.InfrastructureStatus{InfrastructureName: "my-cluster-abc"},
			},
			credentials: map[string]string{
				"vcenter.example.com": "admin:secret",
			},
			wantInfraID: "my-cluster-abc",
			wantVCCount: 1,
			wantPrimary: "vcenter.example.com",
		},
		{
			name: "multiple failure domains",
			failureDomains: []configv1.VSpherePlatformFailureDomainSpec{
				{Name: "fd1", Server: "vc1.example.com", Topology: configv1.VSpherePlatformTopology{Datacenter: "dc1"}},
				{Name: "fd2", Server: "vc2.example.com", Topology: configv1.VSpherePlatformTopology{Datacenter: "dc2"}},
			},
			infra: &configv1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{Name: "cluster", UID: types.UID("uid2")},
				Status:     configv1.InfrastructureStatus{InfrastructureName: "infra2"},
			},
			credentials: map[string]string{
				"vc1.example.com": "user1:pass1",
				"vc2.example.com": "user2:pass2",
			},
			wantInfraID: "infra2",
			wantVCCount: 2,
			wantPrimary: "vc1.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewMetadataManager(fakekubeclient.NewClientset())
			got, err := mgr.GenerateMetadata(context.Background(), tt.failureDomains, tt.infra, tt.credentials)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("GenerateMetadata: expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("GenerateMetadata: %v", err)
			}
			if got.InfraID != tt.wantInfraID {
				t.Errorf("InfraID = %q, want %q", got.InfraID, tt.wantInfraID)
			}
			if len(got.VCenters) != tt.wantVCCount {
				t.Errorf("VCenters count = %d, want %d", len(got.VCenters), tt.wantVCCount)
			}
			if got.VCenter != tt.wantPrimary {
				t.Errorf("VCenter = %q, want %q", got.VCenter, tt.wantPrimary)
			}
			if got.TerraformPlatform != "vsphere" {
				t.Errorf("TerraformPlatform = %q, want %q", got.TerraformPlatform, "vsphere")
			}
		})
	}
}

func TestSaveToSecretAndGet(t *testing.T) {
	ctx := context.Background()
	client := fakekubeclient.NewClientset()
	mgr := NewMetadataManager(client)

	md := &Metadata{
		ClusterName:       "test-cluster",
		InfraID:           "test-infra",
		VCenter:           "vc.example.com",
		Username:          "admin",
		Password:          "secret",
		TerraformPlatform: "vsphere",
		VCenters: []VCenters{
			{Server: "vc.example.com", Port: 443, Username: "admin", Password: "secret"},
		},
	}

	// Save
	if err := mgr.SaveToSecret(ctx, md, "test-ns", "test-metadata"); err != nil {
		t.Fatalf("SaveToSecret: %v", err)
	}

	// Verify Secret exists with correct label
	secret, err := client.CoreV1().Secrets("test-ns").Get(ctx, "test-metadata", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Get Secret: %v", err)
	}
	if secret.Labels["migration.openshift.io/metadata"] != "true" {
		t.Error("Secret missing metadata label")
	}

	// Get metadata back
	data, err := mgr.GetMetadataFromSecret(ctx, "test-ns", "test-metadata")
	if err != nil {
		t.Fatalf("GetMetadataFromSecret: %v", err)
	}

	var roundTripped Metadata
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if roundTripped.InfraID != md.InfraID {
		t.Errorf("roundTripped.InfraID = %q, want %q", roundTripped.InfraID, md.InfraID)
	}
	if roundTripped.VCenter != md.VCenter {
		t.Errorf("roundTripped.VCenter = %q, want %q", roundTripped.VCenter, md.VCenter)
	}

	// Update existing
	md.InfraID = "updated-infra"
	if err := mgr.SaveToSecret(ctx, md, "test-ns", "test-metadata"); err != nil {
		t.Fatalf("SaveToSecret (update): %v", err)
	}
	data, err = mgr.GetMetadataFromSecret(ctx, "test-ns", "test-metadata")
	if err != nil {
		t.Fatalf("GetMetadataFromSecret (after update): %v", err)
	}
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("Unmarshal (after update): %v", err)
	}
	if roundTripped.InfraID != "updated-infra" {
		t.Errorf("after update InfraID = %q, want %q", roundTripped.InfraID, "updated-infra")
	}
}

func TestSaveToSecretAndGetNotFound(t *testing.T) {
	mgr := NewMetadataManager(fakekubeclient.NewClientset())
	_, err := mgr.GetMetadataFromSecret(context.Background(), "test-ns", "nonexistent")
	if err == nil {
		t.Fatal("GetMetadataFromSecret: expected error for nonexistent secret")
	}
}

func TestSaveToSecretMissingKey(t *testing.T) {
	ctx := context.Background()
	client := fakekubeclient.NewClientset()
	// Create a secret without the metadata key
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "empty-secret", Namespace: "test-ns"},
		Data:       map[string][]byte{},
	}
	if _, err := client.CoreV1().Secrets("test-ns").Create(ctx, secret, metav1.CreateOptions{}); err != nil {
		t.Fatalf("Create Secret: %v", err)
	}

	mgr := NewMetadataManager(client)
	_, err := mgr.GetMetadataFromSecret(ctx, "test-ns", "empty-secret")
	if err == nil {
		t.Fatal("GetMetadataFromSecret: expected error for missing key")
	}
}

func TestGetMetadataSecretName(t *testing.T) {
	got := GetMetadataSecretName("my-migration")
	want := "my-migration-metadata"
	if got != want {
		t.Errorf("GetMetadataSecretName = %q, want %q", got, want)
	}
}

func TestParseCredentials(t *testing.T) {
	tests := []struct {
		input    string
		wantUser string
		wantPass string
	}{
		{"admin:secret", "admin", "secret"},
		{"user:pass:with:colons", "user", "pass:with:colons"},
		{"", "", ""},
		{"nocolon", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			u, p := parseCredentials(tt.input)
			if u != tt.wantUser || p != tt.wantPass {
				t.Errorf("parseCredentials(%q) = (%q, %q), want (%q, %q)", tt.input, u, p, tt.wantUser, tt.wantPass)
			}
		})
	}
}
