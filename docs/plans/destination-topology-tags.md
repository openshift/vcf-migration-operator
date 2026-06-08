# Destination Topology Tag Creation

## Overview

As the VCF migration operator, create the required vCenter tag categories and topology tags on the destination so that OpenShift topology information (regions and zones) is represented consistently in the new vCenter.

## Current Baseline

The operator already implements this functionality in `ensureDestinationInitialized()` at `internal/controller/vmwarecloudfoundationmigration_controller.go:181-267`. The tag lifecycle is handled by four functions in `internal/vsphere/tags.go`:

- `EnsureTagCategory()` (line 39) — creates or retrieves `openshift-region` / `openshift-zone` categories with `SINGLE` cardinality. Handles `already_exists` errors idempotently.
- `EnsureTag()` (line 83) — creates or retrieves individual region/zone tags within a category. Also idempotent.
- `CreateRegionAndZoneTags()` (line 137) — orchestrates category + tag creation for both region and zone, returning both tag IDs.
- `AttachFailureDomainTags()` (line 168) — attaches the region tag to the datacenter and the zone tag to the cluster compute resource.

Preflight validation (`internal/controller/preflight.go:36-47`) checks that the target vCenter user holds `InventoryService.Tagging.AttachTag`, `InventoryService.Tagging.CreateCategory`, `InventoryService.Tagging.CreateTag` on the root folder, and `InventoryService.Tagging.ObjectAttachable` on target datacenter and cluster objects.

Unit tests exist in `internal/vsphere/tags_test.go` covering `EnsureTagCategory`, `EnsureTag`, and `CreateRegionAndZoneTags` idempotency using the govmomi simulator. There is no test for `AttachFailureDomainTags`.

## Acceptance Criteria Assessment

| Criterion | Status | Location |
|---|---|---|
| Creates `openshift-region` and `openshift-zone` tag categories when missing | **Done** | `tags.go:39-79` — `EnsureTagCategory` with get-or-create + `already_exists` fallback |
| Creates region and zone tags referenced by target failure domains | **Done** | `tags.go:83-117`, `tags.go:137-163` — `EnsureTag` / `CreateRegionAndZoneTags` |
| Attaches tags to destination datacenter and compute cluster | **Mostly done** | `tags.go:168-194` — `AttachFailureDomainTags` exists, but attachment behavior is not verified by tests |
| Tag operations are idempotent and handle pre-existing categories/tags safely | **Done** | Category/tag create paths are idempotent and tested; attachment idempotency needs explicit handling (see #2) |
| Failures are surfaced through migration status and logs | **Done** | `controller.go:185-264` — errors propagate to condition `DestinationInitialized` with `ReasonFailed`; progress messages update condition during each step |
| Documentation explains the tag model used by the migration workflow | **Gap** | No documentation exists explaining the tag model |

## Plan

### 1. Add `AttachFailureDomainTags` test coverage

`AttachFailureDomainTags()` is the only tag function without a test. Add a test in `internal/vsphere/tags_test.go` that:

- Creates region and zone tags via `CreateRegionAndZoneTags`
- Looks up the simulator's default datacenter and cluster
- Calls `AttachFailureDomainTags` and asserts no error
- **Verifies final state**: calls `s.TagManager.ListAttachedTags(ctx, dc.Reference())` and confirms the region tag ID is in the result; does the same for the cluster and zone tag ID. This catches regressions where tags are swapped or silently dropped.
- Calls `AttachFailureDomainTags` a second time (idempotency) and asserts no error
- Verifies nil-datacenter and nil-cluster return errors

### 2. Handle `AttachTag` idempotency explicitly

The govmomi single-object `AttachTag` (unlike the batch variants `AttachTagToMultipleObjects` / `AttachMultipleTagsToObject` which are documented as idempotent no-ops) does not guarantee silent success on duplicate attachments. The vSphere API may return an `already_exists`-style error.

**2a.** Route `AttachFailureDomainTags` through the package-level `AttachTag()` helper (line 120) instead of calling `s.TagManager.AttachTag()` directly. This makes `AttachTag()` the single place to handle idempotency.

**2b.** Teach the `AttachTag()` helper to tolerate `already_exists` errors:

```go
if err := s.TagManager.AttachTag(ctx, tagID, ref); err != nil {
    if isAlreadyExists(err) {
        log.V(2).Info("tag already attached", "tagID", tagID, "object", ref)
        return nil
    }
    return fmt.Errorf("attaching tag %s to %s: %w", tagID, ref, err)
}
```

Note: `isAlreadyExists()` is a string match on `"already_exists"`. Confirm the exact error format from the govmomi simulator in the new test from item #1 before relying on it.

**2c.** Deduplicate tag attachment calls in the controller. The controller already deduplicates folder creation with a `folderCreated` map (`controller.go:194-197`). Apply the same pattern for tag attachment — multiple failure domains sharing a datacenter will re-attach the same region tag on every iteration. Add a `tagAttached` map keyed by `(server, datacenter, region)` and `(server, cluster, zone)` to skip redundant API calls.

### 3. Add tag model documentation

Add a section to the operator's existing documentation (or a dedicated doc if none exists) explaining:

- The `openshift-region` and `openshift-zone` tag categories are created with `SINGLE` cardinality, meaning each object can have at most one region tag and one zone tag.
- Associable types are `Datacenter`, `ClusterComputeResource`, `Datastore`, and `Folder`.
- The region tag (from `fd.Region`) is attached to the target datacenter.
- The zone tag (from `fd.Zone`) is attached to the target compute cluster.
- This mirrors the tag model that the OpenShift vSphere cloud provider and CSI driver use to discover failure domain topology.
- Multiple failure domains may share the same region but different zones.
- The operator creates these tags during the `DestinationInitialized` phase, after preflight validates that the target user has the required tagging privileges.
- **Ownership model**: tag categories (`openshift-region`, `openshift-zone`) are shared per-vCenter infrastructure — multiple OpenShift clusters on the same vCenter reuse them. The operator creates categories if missing but treats them as shared resources, not owned. It will never delete or modify an existing category.
- **Brownfield behavior**: if categories already exist with incompatible settings (wrong cardinality or missing required associable types), the operator rejects them with an error explaining the mismatch and how to resolve it (see #5). Extra associable types and different descriptions are tolerated.

### 4. Log tag attachment details at V(1)

`AttachFailureDomainTags` logs the intent at V(1) (line 176) but does not log success per-object. Add a V(1) log line after each successful `AttachTag` call so that operators can confirm from logs which objects received which tags.

Since item #2 makes the `AttachTag` helper tolerate duplicates, use wording that distinguishes the two outcomes:

```text
"ensured region tag attached to datacenter" regionTagID=... datacenter=...    // new attachment
"region tag already attached to datacenter" regionTagID=... datacenter=...    // idempotent no-op
```

The `failureDomain` name should not be added to these log lines — the `vsphere` package should not couple to the controller's domain model. The controller already logs `"failure domain initialized", "name", fd.Name` at line 261, which provides the correlation.

### 5. Validate tag category shape on reuse

`EnsureTagCategory` at `tags.go:46-50` returns an existing category without checking whether its settings are compatible. Two checks are needed when a category already exists:

**5a. Cardinality**: if the existing category's `Cardinality` differs from the requested `SINGLE` (e.g., a third-party tool created it with `MULTIPLE`), return an error. Wrong cardinality causes topology ambiguity downstream.

**5b. Required associable types**: if the existing category is missing `Datacenter` or `ClusterComputeResource` from its `AssociableTypes`, return an error. Without these, tag attachment to those object types will fail at the API level with a confusing error.

**Compatibility policy**: reject if cardinality differs or required associable types are missing. Tolerate extra associable types beyond what the operator needs and different descriptions — these are harmless. The error message should explain the mismatch and how to resolve it (update the category via the vSphere UI or delete and let the operator recreate it).

## Key Files

- `internal/vsphere/tags.go` — tag category/tag/attachment operations (#2, #4, #5)
- `internal/vsphere/tags_test.go` — test coverage (#1)
- `internal/controller/vmwarecloudfoundationmigration_controller.go:181-267` — `ensureDestinationInitialized` orchestration (context for all items)
- `internal/controller/preflight.go:36-47` — privilege definitions for tag operations (reference)
- `api/v1alpha1/vmwarecloudfoundationmigration_types.go:94-95` — `ConditionDestinationInitialized` definition (reference)

## Verification

1. `go build ./...` — confirms compilation after changes
2. `go test ./internal/vsphere/ -run TestAttachFailureDomainTags` — new attachment test
3. `go test ./internal/vsphere/ -run TestEnsureTagCategory` — verify cardinality check doesn't break existing tests
4. `go test ./internal/vsphere/...` — full vsphere package test suite
5. `go test ./...` — full suite passes
6. `go vet ./...` — no new warnings

## Risks

- The category shape validation (#5) is a new failure mode for brownfield vCenters where `openshift-region` or `openshift-zone` was created with different settings. This is the correct behavior — silently using a misconfigured category is worse — but it will surface as a new error for those environments. The error message must explain how to resolve the conflict (update the category via the vSphere UI, or delete and let the operator recreate it).
- The `AttachTag` idempotency fix (#2b) depends on the vSphere API returning an `already_exists`-class error for duplicate attachments. `isAlreadyExists()` is a string match on `"already_exists"` — confirm the exact error format from the govmomi simulator in the test from item #1 before relying on it. If the simulator or real API uses a different error string, the check will miss it and surface a spurious error.
- The controller-side deduplication (#2c) changes the execution model: previously every failure domain triggered its own attachment call, now shared datacenter/cluster pairs are skipped after the first. This is correct but shifts the error surface — a failure on the first domain's attachment now blocks all domains sharing that object, rather than each failing independently.
