package vsphere

import (
	"context"
	"testing"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
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
	simulator.Test(func(ctx context.Context, c *vim25.Client) {
		s := newTestSession(ctx, t, c)

		id, err := EnsureTagCategory(ctx, s, "test-category", "a test category", "SINGLE")
		if err != nil {
			t.Fatalf("EnsureTagCategory: %v", err)
		}
		if id == "" {
			t.Fatal("expected non-empty category ID")
		}
		// Idempotent: second call returns same category
		id2, err := EnsureTagCategory(ctx, s, "test-category", "a test category", "SINGLE")
		if err != nil {
			t.Fatalf("EnsureTagCategory (second call): %v", err)
		}
		if id2 != id {
			t.Fatalf("expected same category ID, got %q and %q", id, id2)
		}
	})
}

func TestEnsureTag(t *testing.T) {
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
	})
}
