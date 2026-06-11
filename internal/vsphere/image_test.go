package vsphere

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	_ "github.com/vmware/govmomi/vapi/simulator"
)

func TestTemplateNameForFailureDomain(t *testing.T) {
	tests := []struct {
		infraID string
		fdName  string
		want    string
	}{
		{"my-cluster-abc", "fd1", "my-cluster-abc-rhcos-fd1"},
		{"infra", "az-east", "infra-rhcos-az-east"},
	}
	for _, tt := range tests {
		got := TemplateNameForFailureDomain(tt.infraID, tt.fdName)
		if got != tt.want {
			t.Errorf("TemplateNameForFailureDomain(%q, %q) = %q, want %q", tt.infraID, tt.fdName, got, tt.want)
		}
	}
}

func TestValidateTemplateName(t *testing.T) {
	// Under limit.
	if err := ValidateTemplateName("short-name"); err != nil {
		t.Fatalf("ValidateTemplateName(short-name) = %v, want nil", err)
	}

	// Exactly at limit.
	name80 := strings.Repeat("a", maxTemplateNameLen)
	if err := ValidateTemplateName(name80); err != nil {
		t.Fatalf("ValidateTemplateName(80 chars) = %v, want nil", err)
	}

	// Over limit.
	name81 := strings.Repeat("a", maxTemplateNameLen+1)
	if err := ValidateTemplateName(name81); err == nil {
		t.Fatal("ValidateTemplateName(81 chars) = nil, want error")
	}
}

func TestResolveRHCOSOVAFromConfigMap(t *testing.T) {
	// Realistic stream metadata JSON fixture.
	streamJSON := `{
		"stream": "rhcos-4.18",
		"metadata": {"last-modified": "2025-01-01T00:00:00Z"},
		"architectures": {
			"x86_64": {
				"artifacts": {
					"vmware": {
						"release": "418.94.202501010000-0",
						"formats": {
							"ova": {
								"disk": {
									"location": "https://rhcos.mirror.openshift.com/art/storage/prod/streams/4.18/builds/418.94.202501010000-0/x86_64/rhcos-vmware.x86_64.ova",
									"sha256": "abc123def456"
								}
							}
						}
					}
				}
			}
		}
	}`

	t.Run("resolves OVA from valid ConfigMap", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "coreos-bootimages",
				Namespace: "openshift-machine-config-operator",
			},
			Data: map[string]string{
				"stream": streamJSON,
			},
		}

		ova, err := ResolveRHCOSOVAFromConfigMap(cm, "x86_64")
		if err != nil {
			t.Fatalf("ResolveRHCOSOVAFromConfigMap: %v", err)
		}

		expectedURL := "https://rhcos.mirror.openshift.com/art/storage/prod/streams/4.18/builds/418.94.202501010000-0/x86_64/rhcos-vmware.x86_64.ova"
		if ova.Location != expectedURL {
			t.Errorf("Location = %q, want %q", ova.Location, expectedURL)
		}
		if ova.Sha256 != "abc123def456" {
			t.Errorf("Sha256 = %q, want %q", ova.Sha256, "abc123def456")
		}
	})

	t.Run("rejects nil ConfigMap", func(t *testing.T) {
		_, err := ResolveRHCOSOVAFromConfigMap(nil, "x86_64")
		if err == nil {
			t.Fatal("expected error for nil ConfigMap")
		}
	})

	t.Run("rejects missing stream key", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			Data: map[string]string{"other": "data"},
		}
		_, err := ResolveRHCOSOVAFromConfigMap(cm, "x86_64")
		if err == nil {
			t.Fatal("expected error for missing stream key")
		}
		if !strings.Contains(err.Error(), "missing 'stream' key") {
			t.Fatalf("error = %q, want mention of 'stream' key", err.Error())
		}
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			Data: map[string]string{"stream": "not-json"},
		}
		_, err := ResolveRHCOSOVAFromConfigMap(cm, "x86_64")
		if err == nil {
			t.Fatal("expected error for malformed JSON")
		}
	})

	t.Run("rejects missing architecture", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			Data: map[string]string{"stream": streamJSON},
		}
		_, err := ResolveRHCOSOVAFromConfigMap(cm, "arm64")
		if err == nil {
			t.Fatal("expected error for missing arm64 architecture")
		}
	})
}

func TestDownloadOVA(t *testing.T) {
	content := "fake-ova-content-for-testing"
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))

	t.Run("downloads and verifies SHA256", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		}))
		defer server.Close()

		dir := t.TempDir()
		path, err := DownloadOVAToDir(context.Background(), server.URL+"/rhcos.ova", hash, dir)
		if err != nil {
			t.Fatalf("DownloadOVAToDir: %v", err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading downloaded file: %v", err)
		}
		if string(data) != content {
			t.Fatalf("downloaded content = %q, want %q", string(data), content)
		}
	})

	t.Run("returns cached file on second call", func(t *testing.T) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.Write([]byte(content))
		}))
		defer server.Close()

		dir := t.TempDir()
		_, err := DownloadOVAToDir(context.Background(), server.URL+"/rhcos.ova", hash, dir)
		if err != nil {
			t.Fatalf("first download: %v", err)
		}
		_, err = DownloadOVAToDir(context.Background(), server.URL+"/rhcos.ova", hash, dir)
		if err != nil {
			t.Fatalf("second download: %v", err)
		}

		if calls != 1 {
			t.Fatalf("expected 1 HTTP call (cached), got %d", calls)
		}
	})

	t.Run("rejects SHA256 mismatch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("wrong-content"))
		}))
		defer server.Close()

		dir := t.TempDir()
		_, err := DownloadOVAToDir(context.Background(), server.URL+"/rhcos.ova", hash, dir)
		if err == nil {
			t.Fatal("expected SHA256 mismatch error")
		}
		if !strings.Contains(err.Error(), "SHA256 mismatch") {
			t.Fatalf("error = %q, want SHA256 mismatch", err.Error())
		}
	})

	t.Run("rejects non-200 responses", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		dir := t.TempDir()
		_, err := DownloadOVAToDir(context.Background(), server.URL+"/missing.ova", "", dir)
		if err == nil {
			t.Fatal("expected HTTP error")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Fatalf("error = %q, want 404", err.Error())
		}
	})

	t.Run("downloads without hash verification", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		}))
		defer server.Close()

		dir := t.TempDir()
		path, err := DownloadOVAToDir(context.Background(), server.URL+"/rhcos.ova", "", dir)
		if err != nil {
			t.Fatalf("DownloadOVAToDir (no hash): %v", err)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("downloaded file not found: %v", err)
		}
	})

	t.Run("strips query params from filename", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(content))
		}))
		defer server.Close()

		dir := t.TempDir()
		path, err := DownloadOVAToDir(context.Background(), server.URL+"/rhcos.ova?sha256=abc", "", dir)
		if err != nil {
			t.Fatalf("DownloadOVAToDir: %v", err)
		}
		if filepath.Base(path) != "rhcos.ova" {
			t.Fatalf("filename = %q, want %q", filepath.Base(path), "rhcos.ova")
		}
	})
}

func TestFindTemplateByName(t *testing.T) {
	t.Run("rejects nil session", func(t *testing.T) {
		_, _, err := FindTemplateByName(context.Background(), nil, "test")
		if err == nil {
			t.Fatal("expected nil session error")
		}
	})

	t.Run("returns not found for missing template", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)
			_, found, err := FindTemplateByName(ctx, s, "nonexistent-template")
			if err != nil {
				t.Fatalf("FindTemplateByName: %v", err)
			}
			if found {
				t.Fatal("expected found=false for missing template")
			}
		})
	})

	t.Run("finds existing template", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)
			vm := createSimulatorVM(ctx, t, s, c, "test-template")
			if err := vm.MarkAsTemplate(ctx); err != nil {
				t.Fatalf("marking as template: %v", err)
			}

			path, found, err := FindTemplateByName(ctx, s, "test-template")
			if err != nil {
				t.Fatalf("FindTemplateByName: %v", err)
			}
			if !found {
				t.Fatal("expected found=true")
			}
			if path == "" {
				t.Fatal("expected non-empty inventory path")
			}
			if !strings.Contains(path, "test-template") {
				t.Fatalf("path %q does not contain template name", path)
			}
		})
	})

	t.Run("errors on non-template VM", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)
			_ = createSimulatorVM(ctx, t, s, c, "not-a-template")

			_, _, err := FindTemplateByName(ctx, s, "not-a-template")
			if err == nil {
				t.Fatal("expected error for non-template VM")
			}
			if !strings.Contains(err.Error(), "not a template") {
				t.Fatalf("error = %q, want 'not a template'", err.Error())
			}
		})
	})
}

func TestReadOVFFromOVA(t *testing.T) {
	t.Run("extracts OVF from valid OVA", func(t *testing.T) {
		ovaPath := createTestOVA(t, "test.ovf", "<Envelope/>")
		descriptor, err := readOVFFromOVA(ovaPath)
		if err != nil {
			t.Fatalf("readOVFFromOVA: %v", err)
		}
		if descriptor != "<Envelope/>" {
			t.Fatalf("OVF = %q, want %q", descriptor, "<Envelope/>")
		}
	})

	t.Run("errors on missing OVF", func(t *testing.T) {
		ovaPath := createTestOVA(t, "data.vmdk", "disk-data")
		_, err := readOVFFromOVA(ovaPath)
		if err == nil {
			t.Fatal("expected error for missing OVF")
		}
	})

	t.Run("errors on missing file", func(t *testing.T) {
		_, err := readOVFFromOVA("/nonexistent/path.ova")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}

func TestHashFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := "hello world"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	hash, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile: %v", err)
	}

	expected := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	if hash != expected {
		t.Fatalf("hash = %q, want %q", hash, expected)
	}
}

// createSimulatorVM creates a VM in the govmomi simulator with the minimal
// required configuration. Returns the VM object.
func createSimulatorVM(ctx context.Context, t *testing.T, s *Session, c *vim25.Client, name string) *object.VirtualMachine {
	t.Helper()

	dc, err := s.Finder.Datacenter(ctx, "DC0")
	if err != nil {
		t.Fatalf("finding datacenter: %v", err)
	}
	folders, err := dc.Folders(ctx)
	if err != nil {
		t.Fatalf("getting folders: %v", err)
	}

	pool, err := s.Finder.ResourcePool(ctx, "/DC0/host/DC0_C0/Resources")
	if err != nil {
		t.Fatalf("finding resource pool: %v", err)
	}

	spec := types.VirtualMachineConfigSpec{
		Name: name,
		Files: &types.VirtualMachineFileInfo{
			VmPathName: "[LocalDS_0]",
		},
	}

	task, err := folders.VmFolder.CreateVM(ctx, spec, pool, nil)
	if err != nil {
		t.Fatalf("creating VM %q: %v", name, err)
	}
	info, err := task.WaitForResult(ctx)
	if err != nil {
		t.Fatalf("waiting for VM %q creation: %v", name, err)
	}
	return object.NewVirtualMachine(c, info.Result.(types.ManagedObjectReference))
}

// createTestOVA creates a minimal OVA (tar archive) containing a single file
// with the given name and content. Returns the path to the OVA file.
func createTestOVA(t *testing.T, filename, content string) string {
	t.Helper()

	dir := t.TempDir()
	ovaPath := filepath.Join(dir, "test.ova")

	f, err := os.Create(ovaPath)
	if err != nil {
		t.Fatalf("creating OVA file: %v", err)
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	header := &tar.Header{
		Name: filename,
		Size: int64(len(content)),
		Mode: 0o600,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("writing tar header: %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatalf("writing tar content: %v", err)
	}

	return ovaPath
}
