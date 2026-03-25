package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
	"k8s.io/klog/v2"
)

const (
	// TagCategoryRegion is the vSphere tag category name used for OpenShift
	// region failure domains.
	TagCategoryRegion = "openshift-region"

	// TagCategoryZone is the vSphere tag category name used for OpenShift
	// zone failure domains.
	TagCategoryZone = "openshift-zone"

	// TagCategoryRegionDescription describes the region tag category.
	TagCategoryRegionDescription = "OpenShift region for failure domain topology"

	// TagCategoryZoneDescription describes the zone tag category.
	TagCategoryZoneDescription = "OpenShift zone for failure domain topology"
)

// isAlreadyExists returns true if the error indicates the resource already exists
// (e.g. vSphere API returns already_exists).
func isAlreadyExists(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already_exists")
}

// EnsureTagCategory returns the vSphere tag category ID for the given name,
// creating the category if it does not exist. If the category already exists
// (e.g. from a previous run or another cluster), it is looked up and its ID is
// returned so creation is idempotent.
func EnsureTagCategory(ctx context.Context, s *Session, name, description, cardinality string) (string, error) {
	if s == nil || s.TagManager == nil {
		return "", fmt.Errorf("session and TagManager must not be nil")
	}
	log := klog.FromContext(ctx)
	log.V(2).Info("ensuring tag category", "name", name, "cardinality", cardinality)

	existing, err := s.TagManager.GetCategory(ctx, name)
	if err == nil && existing != nil && existing.ID != "" {
		log.V(2).Info("using existing tag category", "name", name, "id", existing.ID)
		return existing.ID, nil
	}

	cat := tags.Category{
		Name:        name,
		Description: description,
		Cardinality: cardinality,
		AssociableTypes: []string{
			"Datacenter",
			"ClusterComputeResource",
			"Datastore",
			"Folder",
		},
	}

	id, err := s.TagManager.CreateCategory(ctx, &cat)
	if err != nil {
		if isAlreadyExists(err) {
			existing, getErr := s.TagManager.GetCategory(ctx, name)
			if getErr != nil {
				return "", fmt.Errorf("creating tag category %q (already exists but get failed): %w", name, getErr)
			}
			log.V(2).Info("tag category already existed, using existing", "name", name, "id", existing.ID)
			return existing.ID, nil
		}
		return "", fmt.Errorf("creating tag category %q: %w", name, err)
	}

	log.V(2).Info("created tag category", "name", name, "id", id)
	return id, nil
}

// EnsureTag returns the vSphere tag ID for the given name in the category,
// creating the tag if it does not exist. Idempotent when the tag already exists.
func EnsureTag(ctx context.Context, s *Session, categoryID, name, description string) (string, error) {
	if s == nil || s.TagManager == nil {
		return "", fmt.Errorf("session and TagManager must not be nil")
	}
	log := klog.FromContext(ctx)
	log.V(2).Info("ensuring tag", "name", name, "categoryID", categoryID)

	existing, err := s.TagManager.GetTagForCategory(ctx, name, categoryID)
	if err == nil && existing != nil && existing.ID != "" {
		log.V(2).Info("using existing tag", "name", name, "id", existing.ID)
		return existing.ID, nil
	}

	tag := tags.Tag{
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
	}

	id, err := s.TagManager.CreateTag(ctx, &tag)
	if err != nil {
		if isAlreadyExists(err) {
			existing, getErr := s.TagManager.GetTagForCategory(ctx, name, categoryID)
			if getErr != nil {
				return "", fmt.Errorf("creating tag %q (already exists but get failed): %w", name, getErr)
			}
			log.V(2).Info("tag already existed, using existing", "name", name, "id", existing.ID)
			return existing.ID, nil
		}
		return "", fmt.Errorf("creating tag %q in category %s: %w", name, categoryID, err)
	}

	log.V(2).Info("created tag", "name", name, "id", id)
	return id, nil
}

// AttachTag attaches a vSphere tag to the given managed object reference.
func AttachTag(ctx context.Context, s *Session, tagID string, obj object.Reference) error {
	if obj == nil {
		return fmt.Errorf("cannot attach tag: object reference is nil")
	}
	log := klog.FromContext(ctx)
	ref := obj.Reference()
	log.V(2).Info("attaching tag", "tagID", tagID, "object", ref)

	if err := s.TagManager.AttachTag(ctx, tagID, ref); err != nil {
		return fmt.Errorf("attaching tag %s to %s: %w", tagID, ref, err)
	}
	return nil
}

// CreateRegionAndZoneTags ensures the OpenShift region and zone tag categories
// and their corresponding tags exist, creating them if necessary. It returns the
// region and zone tag IDs. Idempotent when categories or tags already exist.
func CreateRegionAndZoneTags(ctx context.Context, s *Session, region, zone string) (regionTagID, zoneTagID string, err error) {
	log := klog.FromContext(ctx)
	log.V(1).Info("creating region and zone tags", "region", region, "zone", zone)

	regionCatID, err := EnsureTagCategory(ctx, s, TagCategoryRegion, TagCategoryRegionDescription, "SINGLE")
	if err != nil {
		return "", "", fmt.Errorf("creating region tag category: %w", err)
	}

	zoneCatID, err := EnsureTagCategory(ctx, s, TagCategoryZone, TagCategoryZoneDescription, "SINGLE")
	if err != nil {
		return "", "", fmt.Errorf("creating zone tag category: %w", err)
	}

	regionTagID, err = EnsureTag(ctx, s, regionCatID, region, fmt.Sprintf("OpenShift region %s", region))
	if err != nil {
		return "", "", fmt.Errorf("creating region tag: %w", err)
	}

	zoneTagID, err = EnsureTag(ctx, s, zoneCatID, zone, fmt.Sprintf("OpenShift zone %s", zone))
	if err != nil {
		return "", "", fmt.Errorf("creating zone tag: %w", err)
	}

	log.V(1).Info("created region and zone tags", "regionTagID", regionTagID, "zoneTagID", zoneTagID)
	return regionTagID, zoneTagID, nil
}

// AttachFailureDomainTags attaches the region tag to the datacenter and the zone
// tag to the cluster compute resource. This establishes the vSphere failure
// domain topology required by OpenShift.
func AttachFailureDomainTags(ctx context.Context, s *Session, regionTagID, zoneTagID string, datacenter *object.Datacenter, cluster *object.ClusterComputeResource) error {
	if datacenter == nil {
		return fmt.Errorf("cannot attach failure domain tags: datacenter is nil")
	}
	if cluster == nil {
		return fmt.Errorf("cannot attach failure domain tags: cluster is nil")
	}
	log := klog.FromContext(ctx)
	log.V(1).Info("attaching failure domain tags",
		"regionTagID", regionTagID,
		"zoneTagID", zoneTagID,
		"datacenter", datacenter.Reference(),
		"cluster", cluster.Reference(),
	)

	ref := datacenter.Reference()
	if err := s.TagManager.AttachTag(ctx, regionTagID, ref); err != nil {
		return fmt.Errorf("attaching region tag to datacenter %s: %w", ref, err)
	}

	clusterRef := cluster.Reference()
	if err := s.TagManager.AttachTag(ctx, zoneTagID, clusterRef); err != nil {
		return fmt.Errorf("attaching zone tag to cluster %s: %w", clusterRef, err)
	}

	return nil
}
