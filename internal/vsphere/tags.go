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

	// datacenterType is the vSphere associable type name for datacenters.
	datacenterType = "Datacenter"
	// clusterComputeResourceType is the vSphere associable type name for clusters.
	clusterComputeResourceType = "ClusterComputeResource"
	// datastoreType is the vSphere associable type name for datastores.
	datastoreType = "Datastore"
	// folderType is the vSphere associable type name for folders.
	folderType = "Folder"
	// virtualMachineType is the vSphere associable type name for VMs.
	virtualMachineType = "VirtualMachine"
	// resourcePoolType is the vSphere associable type name for resource pools.
	resourcePoolType = "ResourcePool"
	// storagePodType is the vSphere associable type name for datastore clusters.
	storagePodType = "StoragePod"

	// clusterTagCategoryDescription is used for the per-cluster infra tag category,
	// matching openshift-install behavior.
	clusterTagCategoryDescription = "Added by openshift-install do not remove"
)

var (
	tagCategoryAssociableTypes = []string{
		datacenterType,
		clusterComputeResourceType,
		datastoreType,
		folderType,
	}
	requiredTagCategoryAssociableTypes = []string{
		datacenterType,
		clusterComputeResourceType,
	}
)

// isAlreadyExists returns true if the error indicates the resource already exists
// (e.g. vSphere API returns already_exists).
func isAlreadyExists(err error) bool {
	return err != nil && strings.Contains(err.Error(), "already_exists")
}

func validateExistingCategory(category *tags.Category, name, cardinality string) error {
	var incompatibilities []string

	if category.Cardinality != cardinality {
		incompatibilities = append(incompatibilities, fmt.Sprintf("cardinality %q does not match required %q", category.Cardinality, cardinality))
	}

	missingTypes := missingAssociableTypes(category.AssociableTypes, requiredTagCategoryAssociableTypes)
	if len(missingTypes) > 0 {
		incompatibilities = append(incompatibilities, fmt.Sprintf("missing required associable types %s", strings.Join(missingTypes, ", ")))
	}

	if len(incompatibilities) == 0 {
		return nil
	}

	return fmt.Errorf("existing tag category %q is incompatible: %s; update the category in the vSphere UI or delete it and let the operator recreate it", name, strings.Join(incompatibilities, "; "))
}

func missingAssociableTypes(existingTypes, requiredTypes []string) []string {
	existing := make(map[string]struct{}, len(existingTypes))
	for _, associableType := range existingTypes {
		existing[associableType] = struct{}{}
	}

	missing := make([]string, 0, len(requiredTypes))
	for _, requiredType := range requiredTypes {
		if _, ok := existing[requiredType]; !ok {
			missing = append(missing, requiredType)
		}
	}

	return missing
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
		if err := validateExistingCategory(existing, name, cardinality); err != nil {
			return "", err
		}
		log.V(2).Info("using existing tag category", "name", name, "id", existing.ID)
		return existing.ID, nil
	}

	cat := tags.Category{
		Name:            name,
		Description:     description,
		Cardinality:     cardinality,
		AssociableTypes: tagCategoryAssociableTypes,
	}

	id, err := s.TagManager.CreateCategory(ctx, &cat)
	if err != nil {
		if isAlreadyExists(err) {
			existing, getErr := s.TagManager.GetCategory(ctx, name)
			if getErr != nil {
				return "", fmt.Errorf("creating tag category %q (already exists but get failed): %w", name, getErr)
			}
			if err := validateExistingCategory(existing, name, cardinality); err != nil {
				return "", err
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

// AttachTag attaches a vSphere tag to the given managed object reference. It
// returns true when the tag was newly attached, and false when the attachment
// already existed or no tag ID was provided.
func AttachTag(ctx context.Context, s *Session, tagID string, obj object.Reference) (bool, error) {
	if obj == nil {
		return false, fmt.Errorf("cannot attach tag: object reference is nil")
	}
	if tagID == "" {
		return false, nil
	}
	if s == nil || s.TagManager == nil {
		return false, fmt.Errorf("session and TagManager must not be nil")
	}
	log := klog.FromContext(ctx)
	ref := obj.Reference()
	log.V(2).Info("attaching tag", "tagID", tagID, "object", ref)

	if err := s.TagManager.AttachTag(ctx, tagID, ref); err != nil {
		if isAlreadyExists(err) {
			log.V(2).Info("tag already attached", "tagID", tagID, "object", ref)
			return false, nil
		}
		return false, fmt.Errorf("attaching tag %s to %s: %w", tagID, ref, err)
	}
	return true, nil
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
	if regionTagID != "" {
		attached, err := AttachTag(ctx, s, regionTagID, datacenter)
		if err != nil {
			return fmt.Errorf("attaching region tag to datacenter %s: %w", ref, err)
		}
		if attached {
			log.V(1).Info("ensured region tag attached to datacenter", "regionTagID", regionTagID, "datacenter", ref)
		} else {
			log.V(1).Info("region tag already attached to datacenter", "regionTagID", regionTagID, "datacenter", ref)
		}
	}

	clusterRef := cluster.Reference()
	if zoneTagID != "" {
		attached, err := AttachTag(ctx, s, zoneTagID, cluster)
		if err != nil {
			return fmt.Errorf("attaching zone tag to cluster %s: %w", clusterRef, err)
		}
		if attached {
			log.V(1).Info("ensured zone tag attached to cluster", "zoneTagID", zoneTagID, "cluster", clusterRef)
		} else {
			log.V(1).Info("zone tag already attached to cluster", "zoneTagID", zoneTagID, "cluster", clusterRef)
		}
	}

	return nil
}

// EnsureClusterTag ensures the OpenShift cluster tag category and tag exist
// on the vCenter, matching openshift-install behavior. Returns the tag ID
// to pass to AttachTag after template import.
//
// The category name is "openshift-<infraID>" with SINGLE cardinality,
// associable to VirtualMachine, ResourcePool, Folder, Datastore, and
// StoragePod types. The tag name is the infraID itself.
//
// This function is idempotent: it creates the category and tag if they
// do not exist, or reuses existing ones.
//
// Adapted from openshift/installer pkg/infrastructure/vsphere/clusterapi/tags.go:createClusterTagID.
func EnsureClusterTag(ctx context.Context, s *Session, infraID string) (string, error) {
	if s == nil || s.TagManager == nil {
		return "", fmt.Errorf("session and TagManager must not be nil")
	}
	log := klog.FromContext(ctx)

	categoryName := fmt.Sprintf("openshift-%s", infraID)

	// The cluster infra tag category uses a different set of associable types
	// than the region/zone categories, matching the installer exactly.
	clusterAssociableTypes := []string{
		virtualMachineType,
		resourcePoolType,
		folderType,
		datastoreType,
		storagePodType,
	}

	// Get or create category.
	categoryID, err := ensureTagCategoryWithTypes(ctx, s, categoryName, clusterTagCategoryDescription, "SINGLE", clusterAssociableTypes)
	if err != nil {
		return "", fmt.Errorf("ensuring cluster tag category %q: %w", categoryName, err)
	}
	log.V(1).Info("ensured cluster tag category", "name", categoryName, "id", categoryID)

	// Get or create tag within the category.
	tagID, err := EnsureTag(ctx, s, categoryID, infraID, clusterTagCategoryDescription)
	if err != nil {
		return "", fmt.Errorf("ensuring cluster tag %q in category %q: %w", infraID, categoryName, err)
	}
	log.V(1).Info("ensured cluster tag", "name", infraID, "id", tagID, "category", categoryName)

	return tagID, nil
}

// ensureTagCategoryWithTypes creates or retrieves a tag category with the given
// associable types. Unlike EnsureTagCategory, this accepts the associable types
// as a parameter instead of using the package-level default.
func ensureTagCategoryWithTypes(ctx context.Context, s *Session, name, description, cardinality string, associableTypes []string) (string, error) {
	log := klog.FromContext(ctx)
	log.V(2).Info("ensuring tag category with custom types", "name", name, "cardinality", cardinality)

	existing, err := s.TagManager.GetCategory(ctx, name)
	if err == nil && existing != nil && existing.ID != "" {
		if existing.Cardinality != cardinality {
			return "", fmt.Errorf("existing tag category %q has cardinality %q, required %q", name, existing.Cardinality, cardinality)
		}
		log.V(2).Info("using existing tag category", "name", name, "id", existing.ID)
		return existing.ID, nil
	}

	cat := tags.Category{
		Name:            name,
		Description:     description,
		Cardinality:     cardinality,
		AssociableTypes: associableTypes,
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
