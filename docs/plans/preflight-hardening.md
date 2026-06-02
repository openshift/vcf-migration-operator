# Preflight Checks Hardening and Test Coverage

## Overview

Harden the preflight validation in `internal/controller/preflight.go` with several correctness fixes, resilience improvements, and missing test coverage identified during code review.

## Current Baseline

- `runPreflightChecks()` in `preflight.go` runs a linear sequence of validations: failure domains, credentials, version/gate detection, CSI PV blocking, interfering resource blocking, source connectivity, and per-failure-domain target validation (inventory + privileges).
- Sessions are created via `getVSphereSession()` which delegates to `vsphere.GetOrCreate()` — a global cache keyed by server+datacenter+username. Sessions have a `Close()` method and there is a `ClearSessions()` utility, but neither is called after preflight.
- `checkInterferingRolloutResources()` calls `listDynamicResourceNames()` which propagates any List error directly, including 404s when the CRD doesn't exist on the cluster.
- `missingPrivileges()` is a pure function with no test coverage.
- There is no integration-level test for `runPreflightChecks()` itself — only its leaf functions are unit-tested.
- Source vCenter validation uses only `sourceVC.Datacenters[0]`, silently ignoring additional datacenters.

## Plan

### 1. Clarify `isVersionAtLeastFive` condition

In `internal/openshift/version.go:117`, the expression `major == 5 && minor >= 0` is always true for any valid minor. Replace with the equivalent and self-documenting `major >= 5`.

### 2. Handle missing CRDs in `checkInterferingRolloutResources`

In `internal/controller/preflight.go`, update `listDynamicResourceNames()` (or its callers in `checkInterferingRolloutResources`) to treat `apierrors.IsNotFound` and `meta.IsNoMatchError` responses as "no resources of this type exist" rather than a hard error. This prevents preflight from failing on clusters where `MachineAutoscaler` or `ClusterAutoscaler` CRDs are not installed.

Follow the existing pattern from `internal/openshift/pods.go:137-139` — check `IsNotFound`, log at V(2), and return an empty list.

### 3. Add `missingPrivileges` unit tests

`missingPrivileges()` at `preflight.go:356` is a pure function that maps `EntityPrivilege` results against a requested privilege list. Add table-driven tests covering:
- All privileges granted — returns nil
- Some privileges missing — returns sorted missing list
- No privileges granted — returns all requested
- Empty requested list — returns nil
- Privilege present but `IsGranted == false` — treated as missing

### 4. Add `checkInterferingRolloutResources` multi-blocker test

The existing tests only check one blocker type at a time. Add a test case with all three resource types present simultaneously, verifying the error message contains all three blocker descriptions.

### 5. Add `checkNoVSphereCSIPersistentVolumes` mixed-PV test

Add a test case with both CSI-backed and non-CSI PVs present, verifying that:
- The check fails (CSI PV present)
- The error message names only the CSI PV, not the non-CSI one

### 6. Validate source datacenter count

In `runPreflightChecks()` at line 98, after the empty check, add a guard that errors if `len(sourceVC.Datacenters) > 1`. The operator assumes a single-datacenter source; silently ignoring extras could mask misconfiguration.

### 7. Deduplicate target credential lookups per server

In the failure domain loop (line 111), multiple failure domains may share the same vCenter server. Cache `getTargetCredentials` results by server to avoid redundant secret reads. Use a `map[string]credentials` populated on first lookup per server.

### 8. Add a reconcile-scoped context timeout for vSphere operations

Wrap the context passed to vSphere session creation and inventory lookups with a timeout (e.g., 2 minutes) so that an unreachable vCenter doesn't hang the reconciler until the controller-runtime global timeout. Use `context.WithTimeout` at the top of `runPreflightChecks`.

### 9. Remove redundant `migration.Status.MigrationPath` assignment

`runPreflightChecks()` sets `migration.Status.MigrationPath = path` at line 76. `ensureInfrastructurePrepared()` sets it again at line 178. Remove the assignment inside `runPreflightChecks` — it should only return the path value, and let the caller be the single owner of the status write.

### 10. Block migration when cluster is upgrading or degraded

Add a preflight check that reads `ClusterVersion/version` and rejects migration if an upgrade is in progress (`status.conditions` reports `Progressing=True`). Also read `ClusterOperator` resources via the existing `OperatorManager` and block if any operator reports `Degraded=True`. A vCenter migration on a cluster that is mid-upgrade or unhealthy is not a supported scenario.

### 11. Validate target folder path in failure domains

`runPreflightChecks()` validates datacenter, cluster, datastore, networks, resource pool, and template for each failure domain, but skips `fd.Topology.Folder`. A bad folder path passes preflight and fails only after disruptive cluster mutations begin. Add a `session.Finder.Folder(ctx, fd.Topology.Folder)` check when the field is non-empty, following the same pattern as resource pool and template.

### 12. Enforce failure-domain name uniqueness

`AddTargetVCenter()` deduplicates by failure-domain name when updating `Infrastructure`, but preflight does not reject duplicate names in `migration.Spec.FailureDomains`. Add an early validation step that collects failure-domain names and returns an error if any name appears more than once. Add a matching unit test.

### 13. Add `runPreflightChecks` integration-level unit test

There is no test that exercises the full `runPreflightChecks()` orchestration — ordering, error propagation, condition-setting, and the happy path. Add a focused unit test that wires up fakes for the Kubernetes, config, and dynamic clients plus a govmomi simulator (or stub session), and verifies:
- Happy path returns the expected path and message
- A CSI PV blocker short-circuits before target validation
- An interfering resource blocker short-circuits before target validation
- Path selection (native vs legacy) propagates correctly

## Key Files

- `internal/openshift/version.go` — simplify `isVersionAtLeastFive` (#1), upgrade-in-progress check (#10)
- `internal/controller/preflight.go` — CRD-not-found handling (#2), remove redundant status write (#9), source datacenter guard (#6), credential caching (#7), context timeout (#8), folder validation (#11), failure-domain uniqueness (#12), cluster health gate (#10)
- `internal/controller/preflight_test.go` — all new test cases (#3, #4, #5, #12, #13), and update existing tests if function signatures change
- `internal/openshift/version_test.go` — verify `isVersionAtLeastFive` behavior is unchanged after simplification

## Verification

1. `go build ./...` — confirms compilation
2. `go test ./internal/controller/ -run TestMissingPrivileges` — new privilege tests
3. `go test ./internal/controller/ -run TestCheckInterferingRolloutResources` — multi-blocker and CRD-not-found cases
4. `go test ./internal/controller/ -run TestCheckNoVSphereCSIPersistentVolumes` — mixed PV case
5. `go test ./internal/openshift/ -run TestIsVersionAtLeastFive` — version parsing unchanged
6. `go test ./...` — full suite passes
7. `go vet ./...` — no new warnings

## Risks

- The `IsNotFound` / `IsNoMatchError` handling (#2) changes behavior on clusters without those CRDs. This is the correct behavior, but verify the dynamic fake client in tests can simulate this condition (it may require returning a `NotFound` error from the fake).
- Adding a context timeout (#8) could cause preflight to fail on slow-but-reachable vCenters. The 2-minute default should be generous enough; consider making it configurable later if needed.
- The credential caching (#7) changes the error surface slightly — a per-server error now blocks all failure domains for that server rather than just the first one encountered. This is actually better behavior.

## Review Notes (resolved)

1. **RESOLVED: Cluster upgrade / degraded state.** Added as plan item #10. Block migration if `ClusterVersion` reports `Progressing=True` or any `ClusterOperator` is `Degraded=True`. This eliminates the version-source ambiguity concern — migrating during an upgrade is not a supported scenario.

2. **RESOLVED: Missing folder validation.** Added as plan item #11. Validate `fd.Topology.Folder` when non-empty.

3. **RESOLVED: Failure-domain uniqueness.** Added as plan item #12. Reject duplicate failure-domain names early in preflight.

4. **RESOLVED: Integration-level test.** Added as plan item #13. Focused unit test exercising the full `runPreflightChecks()` orchestration.
