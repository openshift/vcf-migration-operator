package vsphere

import (
	"context"
	"testing"

	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25"

	_ "github.com/vmware/govmomi/vapi/simulator"
)

func TestCreateVMFolder(t *testing.T) {
	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		s := newTestSession(ctx, t, c)

		folder, err := CreateVMFolder(ctx, s, "test-folder")
		if err != nil {
			t.Fatalf("CreateVMFolder: %v", err)
		}
		if folder == nil {
			t.Fatal("expected non-nil folder")
		}
		if folder.InventoryPath == "" {
			t.Fatal("expected non-empty inventory path")
		}
	})
}

func TestGetVMFolder(t *testing.T) {
	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		s := newTestSession(ctx, t, c)

		_, err := CreateVMFolder(ctx, s, "lookup-folder")
		if err != nil {
			t.Fatalf("CreateVMFolder: %v", err)
		}

		folder, err := GetVMFolder(ctx, s, "lookup-folder")
		if err != nil {
			t.Fatalf("GetVMFolder: %v", err)
		}
		if folder == nil {
			t.Fatal("expected non-nil folder")
		}
	})
}

func TestDeleteVMFolder(t *testing.T) {
	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		s := newTestSession(ctx, t, c)

		folder, err := CreateVMFolder(ctx, s, "delete-folder")
		if err != nil {
			t.Fatalf("CreateVMFolder: %v", err)
		}

		if err := DeleteVMFolder(ctx, s, folder); err != nil {
			t.Fatalf("DeleteVMFolder: %v", err)
		}

		_, err = GetVMFolder(ctx, s, "delete-folder")
		if err == nil {
			t.Fatal("expected error looking up deleted folder")
		}
	})
}

func TestCreateVMFolderWithFullPath(t *testing.T) {
	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		s := newTestSession(ctx, t, c)

		fullPath := "/" + s.Datacenter + "/vm/full-path-folder"
		folder, err := CreateVMFolder(ctx, s, fullPath)
		if err != nil {
			t.Fatalf("CreateVMFolder with full path: %v", err)
		}
		if folder == nil {
			t.Fatal("expected non-nil folder")
		}

		found, err := GetVMFolder(ctx, s, "full-path-folder")
		if err != nil {
			t.Fatalf("GetVMFolder: %v", err)
		}
		if found == nil {
			t.Fatal("expected to find folder created with full path")
		}
	})
}

func TestParseFolderName(t *testing.T) {
	tests := []struct {
		name       string
		datacenter string
		folderPath string
		want       string
	}{
		{
			name:       "bare name",
			datacenter: "DC0",
			folderPath: "my-folder",
			want:       "my-folder",
		},
		{
			name:       "full inventory path",
			datacenter: "DC0",
			folderPath: "/DC0/vm/my-folder",
			want:       "my-folder",
		},
		{
			name:       "relative with sub-folder",
			datacenter: "DC0",
			folderPath: "parent/child",
			want:       "parent/child",
		},
		{
			name:       "empty",
			datacenter: "DC0",
			folderPath: "",
			want:       "",
		},
		{
			name:       "leading slash only",
			datacenter: "DC0",
			folderPath: "/my-folder",
			want:       "my-folder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFolderName(tt.datacenter, tt.folderPath)
			if got != tt.want {
				t.Errorf("parseFolderName(%q, %q) = %q, want %q", tt.datacenter, tt.folderPath, got, tt.want)
			}
		})
	}
}
