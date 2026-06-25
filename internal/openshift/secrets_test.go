package openshift

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakekube "k8s.io/client-go/kubernetes/fake"
)

func newTestVSphereCredsSecret(creds map[string][2]string) *corev1.Secret {
	data := make(map[string][]byte)
	for server, up := range creds {
		data[server+".username"] = []byte(up[0])
		data[server+".password"] = []byte(up[1])
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VSphereCredsSecretName,
			Namespace: VSphereCredsSecretNamespace,
		},
		Data: data,
	}
}

func TestGetVSphereCredsSecret(t *testing.T) {
	tests := []struct {
		name    string
		secret  *corev1.Secret
		wantErr bool
	}{
		{
			name: "returns existing secret",
			secret: newTestVSphereCredsSecret(map[string][2]string{
				"source.example.com": {"admin", "pass1"},
			}),
		},
		{
			name:    "errors when secret does not exist",
			secret:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *fakekube.Clientset
			if tt.secret != nil {
				client = fakekube.NewClientset(tt.secret)
			} else {
				client = fakekube.NewClientset()
			}
			mgr := NewSecretManager(client)

			secret, err := mgr.GetVSphereCredsSecret(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetVSphereCredsSecret error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if secret.Name != VSphereCredsSecretName {
				t.Fatalf("secret name = %q, want %q", secret.Name, VSphereCredsSecretName)
			}
		})
	}
}

func TestGetCredentials(t *testing.T) {
	secret := newTestVSphereCredsSecret(map[string][2]string{
		"vc1.example.com": {"admin@vsphere.local", "s3cret"},
	})
	client := fakekube.NewClientset(secret)
	mgr := NewSecretManager(client)

	user, pass, err := mgr.GetCredentials(context.Background(), "vc1.example.com")
	if err != nil {
		t.Fatalf("GetCredentials error = %v", err)
	}
	if user != "admin@vsphere.local" {
		t.Fatal("returned username does not match expected value")
	}
	if pass != "s3cret" {
		t.Fatal("returned password does not match expected value")
	}
}

func TestAddTargetVCenterCreds(t *testing.T) {
	tests := []struct {
		name          string
		existingCreds map[string][2]string
		addServer     string
		addUser       string
		addPass       string
		wantKeyCount  int
		wantSkipped   bool
	}{
		{
			name: "adds new credentials",
			existingCreds: map[string][2]string{
				"source.example.com": {"admin", "pass1"},
			},
			addServer:    "target.example.com",
			addUser:      "admin2",
			addPass:      "pass2",
			wantKeyCount: 4,
		},
		{
			name: "skips when credentials already exist",
			existingCreds: map[string][2]string{
				"source.example.com": {"admin", "pass1"},
			},
			addServer:    "source.example.com",
			addUser:      "other-user",
			addPass:      "other-pass",
			wantKeyCount: 2,
			wantSkipped:  true,
		},
		{
			name:          "handles nil data map",
			existingCreds: nil,
			addServer:     "target.example.com",
			addUser:       "admin",
			addPass:       "pass",
			wantKeyCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var secret *corev1.Secret
			if tt.existingCreds != nil {
				secret = newTestVSphereCredsSecret(tt.existingCreds)
			} else {
				secret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      VSphereCredsSecretName,
						Namespace: VSphereCredsSecretNamespace,
					},
				}
			}

			client := fakekube.NewClientset(secret)
			mgr := NewSecretManager(client)

			updated, err := mgr.AddTargetVCenterCreds(context.Background(), secret, tt.addServer, tt.addUser, tt.addPass)
			if err != nil {
				t.Fatalf("AddTargetVCenterCreds error = %v", err)
			}

			if len(updated.Data) != tt.wantKeyCount {
				t.Fatalf("data key count = %d, want %d", len(updated.Data), tt.wantKeyCount)
			}

			if tt.wantSkipped {
				got := string(updated.Data[tt.addServer+".username"])
				if got != tt.existingCreds[tt.addServer][0] {
					t.Fatal("expected existing credentials to be preserved, but username was overwritten")
				}
			}
		})
	}
}

func TestRemoveSourceVCenterCreds(t *testing.T) {
	tests := []struct {
		name          string
		existingCreds map[string][2]string
		removeServer  string
		wantErr       bool
		wantKeyCount  int
	}{
		{
			name: "removes existing credentials",
			existingCreds: map[string][2]string{
				"source.example.com": {"admin", "pass1"},
				"target.example.com": {"admin2", "pass2"},
			},
			removeServer: "source.example.com",
			wantKeyCount: 2,
		},
		{
			name: "no-ops when keys do not exist",
			existingCreds: map[string][2]string{
				"source.example.com": {"admin", "pass1"},
			},
			removeServer: "nonexistent.example.com",
			wantKeyCount: 2,
		},
		{
			name:         "errors on nil secret",
			removeServer: "source.example.com",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var secret *corev1.Secret
			if tt.existingCreds != nil {
				secret = newTestVSphereCredsSecret(tt.existingCreds)
			}

			var client *fakekube.Clientset
			if secret != nil {
				client = fakekube.NewClientset(secret)
			} else {
				client = fakekube.NewClientset()
			}
			mgr := NewSecretManager(client)

			updated, err := mgr.RemoveSourceVCenterCreds(context.Background(), secret, tt.removeServer)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RemoveSourceVCenterCreds error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if len(updated.Data) != tt.wantKeyCount {
				t.Fatalf("data key count = %d, want %d", len(updated.Data), tt.wantKeyCount)
			}
		})
	}
}

func TestGetVCenterCredsFromSecret(t *testing.T) {
	tests := []struct {
		name     string
		secret   *corev1.Secret
		server   string
		wantUser string
		wantPass string
		wantErr  bool
	}{
		{
			name: "returns credentials for server",
			secret: newTestVSphereCredsSecret(map[string][2]string{
				"source.example.com": {"admin@vsphere.local", "secret123"},
			}),
			server:   "source.example.com",
			wantUser: "admin@vsphere.local",
			wantPass: "secret123",
		},
		{
			name: "errors when username key missing",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-secret",
					Namespace: "test-ns",
				},
				Data: map[string][]byte{
					"source.example.com.password": []byte("pass"),
				},
			},
			server:  "source.example.com",
			wantErr: true,
		},
		{
			name: "errors when password key missing",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-secret",
					Namespace: "test-ns",
				},
				Data: map[string][]byte{
					"source.example.com.username": []byte("admin"),
				},
			},
			server:  "source.example.com",
			wantErr: true,
		},
		{
			name:    "errors when secret not found",
			secret:  nil,
			server:  "source.example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *fakekube.Clientset
			if tt.secret != nil {
				client = fakekube.NewClientset(tt.secret)
			} else {
				client = fakekube.NewClientset()
			}
			mgr := NewSecretManager(client)

			namespace := "test-ns"
			name := "my-secret"
			if tt.secret != nil {
				namespace = tt.secret.Namespace
				name = tt.secret.Name
			}

			user, pass, err := mgr.GetVCenterCredsFromSecret(context.Background(), namespace, name, tt.server)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetVCenterCredsFromSecret error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if user != tt.wantUser {
				t.Fatal("returned username does not match expected value")
			}
			if pass != tt.wantPass {
				t.Fatal("returned password does not match expected value")
			}
		})
	}
}
