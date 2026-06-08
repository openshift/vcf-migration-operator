package vsphere

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25"

	_ "github.com/vmware/govmomi/vapi/simulator"
)

func newTestSession(ctx context.Context, t *testing.T, c *vim25.Client) *Session {
	t.Helper()

	gc := &govmomi.Client{
		Client: c,
	}

	finder := find.NewFinder(c, true)
	dc, err := finder.DefaultDatacenter(ctx)
	if err != nil {
		t.Fatalf("finding default datacenter: %v", err)
	}
	finder.SetDatacenter(dc)

	rc := rest.NewClient(c)
	if err := rc.Login(ctx, simulator.DefaultLogin); err != nil {
		t.Fatalf("REST login: %v", err)
	}

	tagMgr := tags.NewManager(rc)

	return &Session{
		Client:     gc,
		Finder:     finder,
		TagManager: tagMgr,
		Datacenter: dc.Name(),
		restClient: rc,
	}
}

func TestEnsureTagCategory(t *testing.T) {
	t.Run("rejects nil session", func(t *testing.T) {
		_, err := EnsureTagCategory(context.Background(), nil, "test-category", "a test category", "SINGLE")
		if err == nil {
			t.Fatal("EnsureTagCategory succeeded, want nil session error")
		}
		if !strings.Contains(err.Error(), "session and TagManager must not be nil") {
			t.Fatalf("EnsureTagCategory error = %q, want nil session detail", err.Error())
		}
	})

	t.Run("creates and reuses category", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)

			id, err := EnsureTagCategory(ctx, s, "test-category", "a test category", "SINGLE")
			if err != nil {
				t.Fatalf("EnsureTagCategory: %v", err)
			}
			if id == "" {
				t.Fatal("expected non-empty category ID")
			}

			id2, err := EnsureTagCategory(ctx, s, "test-category", "a test category", "SINGLE")
			if err != nil {
				t.Fatalf("EnsureTagCategory (second call): %v", err)
			}
			if id2 != id {
				t.Fatalf("expected same category ID, got %q and %q", id, id2)
			}
		})
	})

	t.Run("rejects mismatched cardinality on reuse", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)
			createTestCategory(ctx, t, s, tags.Category{
				Name:        "test-category",
				Description: "created externally",
				Cardinality: "MULTIPLE",
				AssociableTypes: []string{
					datacenterType,
					clusterComputeResourceType,
					datastoreType,
					folderType,
				},
			})

			_, err := EnsureTagCategory(ctx, s, "test-category", "a test category", "SINGLE")
			if err == nil {
				t.Fatal("EnsureTagCategory succeeded, want cardinality mismatch error")
			}
			if !strings.Contains(err.Error(), `cardinality "MULTIPLE"`) {
				t.Fatalf("EnsureTagCategory error = %q, want cardinality mismatch detail", err.Error())
			}
			if !strings.Contains(err.Error(), "delete it and let the operator recreate it") {
				t.Fatalf("EnsureTagCategory error = %q, want resolution guidance", err.Error())
			}
		})
	})

	t.Run("rejects missing required associable types on reuse", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)
			createTestCategory(ctx, t, s, tags.Category{
				Name:            "test-category",
				Description:     "created externally",
				Cardinality:     "SINGLE",
				AssociableTypes: []string{datastoreType, folderType},
			})

			_, err := EnsureTagCategory(ctx, s, "test-category", "a test category", "SINGLE")
			if err == nil {
				t.Fatal("EnsureTagCategory succeeded, want associable type mismatch error")
			}
			if !strings.Contains(err.Error(), "missing required associable types") {
				t.Fatalf("EnsureTagCategory error = %q, want missing associable types detail", err.Error())
			}
			if !strings.Contains(err.Error(), datacenterType) || !strings.Contains(err.Error(), clusterComputeResourceType) {
				t.Fatalf("EnsureTagCategory error = %q, want missing type names", err.Error())
			}
		})
	})

	t.Run("allows compatible existing category with extra associable types", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)
			id := createTestCategory(ctx, t, s, tags.Category{
				Name:        "test-category",
				Description: "custom description",
				Cardinality: "SINGLE",
				AssociableTypes: []string{
					datacenterType,
					clusterComputeResourceType,
					datastoreType,
					folderType,
					"VirtualMachine",
				},
			})

			gotID, err := EnsureTagCategory(ctx, s, "test-category", "a test category", "SINGLE")
			if err != nil {
				t.Fatalf("EnsureTagCategory: %v", err)
			}
			if gotID != id {
				t.Fatalf("EnsureTagCategory ID = %q, want %q", gotID, id)
			}
		})
	})
}

func TestEnsureTag(t *testing.T) {
	t.Run("rejects nil session", func(t *testing.T) {
		_, err := EnsureTag(context.Background(), nil, "category-id", "test-tag", "a test tag")
		if err == nil {
			t.Fatal("EnsureTag succeeded, want nil session error")
		}
		if !strings.Contains(err.Error(), "session and TagManager must not be nil") {
			t.Fatalf("EnsureTag error = %q, want nil session detail", err.Error())
		}
	})

	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		s := newTestSession(ctx, t, c)

		catID, err := EnsureTagCategory(ctx, s, "test-category", "a test category", "SINGLE")
		if err != nil {
			t.Fatalf("EnsureTagCategory: %v", err)
		}

		tagID, err := EnsureTag(ctx, s, catID, "test-tag", "a test tag")
		if err != nil {
			t.Fatalf("EnsureTag: %v", err)
		}
		if tagID == "" {
			t.Fatal("expected non-empty tag ID")
		}
		// Idempotent: second call returns same tag
		tagID2, err := EnsureTag(ctx, s, catID, "test-tag", "a test tag")
		if err != nil {
			t.Fatalf("EnsureTag (second call): %v", err)
		}
		if tagID2 != tagID {
			t.Fatalf("expected same tag ID, got %q and %q", tagID, tagID2)
		}
	})
}

func TestCreateRegionAndZoneTags(t *testing.T) {
	t.Run("creates and reuses tags", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)

			regionTagID, zoneTagID, err := CreateRegionAndZoneTags(ctx, s, "us-east", "us-east-1a")
			if err != nil {
				t.Fatalf("CreateRegionAndZoneTags: %v", err)
			}
			if regionTagID == "" {
				t.Fatal("expected non-empty region tag ID")
			}
			if zoneTagID == "" {
				t.Fatal("expected non-empty zone tag ID")
			}

			regionTagID2, zoneTagID2, err := CreateRegionAndZoneTags(ctx, s, "us-east", "us-east-1a")
			if err != nil {
				t.Fatalf("CreateRegionAndZoneTags (second call): %v", err)
			}
			if regionTagID2 != regionTagID {
				t.Fatalf("expected same region tag ID, got %q and %q", regionTagID, regionTagID2)
			}
			if zoneTagID2 != zoneTagID {
				t.Fatalf("expected same zone tag ID, got %q and %q", zoneTagID, zoneTagID2)
			}
		})
	})

	t.Run("surfaces region category compatibility errors", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)
			createTestCategory(ctx, t, s, tags.Category{
				Name:            TagCategoryRegion,
				Description:     "created externally",
				Cardinality:     "MULTIPLE",
				AssociableTypes: append([]string(nil), tagCategoryAssociableTypes...),
			})

			_, _, err := CreateRegionAndZoneTags(ctx, s, "us-east", "us-east-1a")
			if err == nil {
				t.Fatal("CreateRegionAndZoneTags succeeded, want region category compatibility error")
			}
			if !strings.Contains(err.Error(), "creating region tag category") {
				t.Fatalf("CreateRegionAndZoneTags error = %q, want wrapped region category detail", err.Error())
			}
		})
	})
}

func TestAttachTag(t *testing.T) {
	t.Run("rejects nil object", func(t *testing.T) {
		_, err := AttachTag(context.Background(), &Session{}, "tag-id", nil)
		if err == nil {
			t.Fatal("AttachTag succeeded, want nil object error")
		}
		if !strings.Contains(err.Error(), "object reference is nil") {
			t.Fatalf("AttachTag error = %q, want nil object detail", err.Error())
		}
	})

	t.Run("skips empty tag ID", func(t *testing.T) {
		attached, err := AttachTag(context.Background(), &Session{}, "", &object.Datacenter{})
		if err != nil {
			t.Fatalf("AttachTag: %v", err)
		}
		if attached {
			t.Fatal("AttachTag reported attachment for empty tag ID")
		}
	})

	t.Run("matches already exists errors", func(t *testing.T) {
		if !isAlreadyExists(fmt.Errorf("server returned already_exists")) {
			t.Fatal("isAlreadyExists returned false, want true")
		}
		if isAlreadyExists(fmt.Errorf("not found")) {
			t.Fatal("isAlreadyExists returned true for unrelated error")
		}
		if isAlreadyExists(nil) {
			t.Fatal("isAlreadyExists returned true for nil error")
		}
	})
}

func TestAttachFailureDomainTags(t *testing.T) {
	t.Run("attaches expected tags and is idempotent", func(t *testing.T) {
		simulator.Test(func(ctx context.Context, c *vim25.Client) {
			s := newTestSession(ctx, t, c)

			regionTagID, zoneTagID, err := CreateRegionAndZoneTags(ctx, s, "us-east", "us-east-1a")
			if err != nil {
				t.Fatalf("CreateRegionAndZoneTags: %v", err)
			}

			dc, err := s.Finder.DefaultDatacenter(ctx)
			if err != nil {
				t.Fatalf("finding default datacenter: %v", err)
			}

			clusters, err := s.Finder.ClusterComputeResourceList(ctx, "*")
			if err != nil {
				t.Fatalf("listing clusters: %v", err)
			}
			if len(clusters) == 0 {
				t.Fatal("expected at least one cluster in simulator inventory")
			}
			cluster := clusters[0]

			if err := AttachFailureDomainTags(ctx, s, regionTagID, zoneTagID, dc, cluster); err != nil {
				t.Fatalf("AttachFailureDomainTags: %v", err)
			}

			dcTags, err := s.TagManager.ListAttachedTags(ctx, dc.Reference())
			if err != nil {
				t.Fatalf("listing attached datacenter tags: %v", err)
			}
			assertTagAttached(t, dcTags, regionTagID)

			clusterTags, err := s.TagManager.ListAttachedTags(ctx, cluster.Reference())
			if err != nil {
				t.Fatalf("listing attached cluster tags: %v", err)
			}
			assertTagAttached(t, clusterTags, zoneTagID)

			if err := AttachFailureDomainTags(ctx, s, regionTagID, zoneTagID, dc, cluster); err != nil {
				t.Fatalf("AttachFailureDomainTags (second call): %v", err)
			}
		})
	})

	t.Run("rejects nil datacenter", func(t *testing.T) {
		err := AttachFailureDomainTags(context.Background(), &Session{}, "region-tag", "zone-tag", nil, &object.ClusterComputeResource{})
		if err == nil {
			t.Fatal("AttachFailureDomainTags succeeded, want nil datacenter error")
		}
		if !strings.Contains(err.Error(), "datacenter is nil") {
			t.Fatalf("AttachFailureDomainTags error = %q, want nil datacenter detail", err.Error())
		}
	})

	t.Run("rejects nil cluster", func(t *testing.T) {
		err := AttachFailureDomainTags(context.Background(), &Session{}, "region-tag", "zone-tag", &object.Datacenter{}, nil)
		if err == nil {
			t.Fatal("AttachFailureDomainTags succeeded, want nil cluster error")
		}
		if !strings.Contains(err.Error(), "cluster is nil") {
			t.Fatalf("AttachFailureDomainTags error = %q, want nil cluster detail", err.Error())
		}
	})
}

func createTestCategory(ctx context.Context, t *testing.T, s *Session, category tags.Category) string {
	t.Helper()

	id, err := s.TagManager.CreateCategory(ctx, &category)
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	return id
}

func assertTagAttached(t *testing.T, attachedTags []string, wantTagID string) {
	t.Helper()

	for _, tagID := range attachedTags {
		if tagID == wantTagID {
			return
		}
	}

	t.Fatalf("attached tags = %v, want %q", attachedTags, wantTagID)
}
