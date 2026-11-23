# Snapshot Checker Implementation Plan

**Goal:** Adapt multiarch-manager-operator's snapshot checker for VCF Migration Operator.

**Approved design:** Keep explicit snapshot and interactive selection, default application suffix to `main`, compare the CSV manager image digest against snapshot images, fail on errors or mismatch, and clean temporary resources. No CI wiring. User requested branch first, then implementation.

**Architecture:** POSIX shell with oc, yq and podman. Copy manifests from a stopped bundle container instead of unpacking image layers. No container command runs. Python unittest mocks external cluster and container operations; YAML parsing remains real.

## Task 1: Script and regression checks

- [x] Write `hack/test-check-snapshots.py`: matching and mismatching digests, registry rewriting, invalid selection/version, missing or ambiguous bundle/manager, command failure and cleanup.
- [x] Run tests before implementation and observe missing-script failure.
- [x] Create executable `hack/check-snapshots.sh` with `-v <application-suffix>` and `-s <snapshot-name>`. Default suffix `main`; other suffixes must be dash-separated versions.
- [x] Run regression checks, shell syntax check, and diff whitespace check.
- [x] Review changes; leave uncommitted for user.

## Verification

- 10 snapshot checker tests passed; missing-script failures observed before implementation.
- POSIX shell syntax and diff whitespace checks passed.
- `make test` timed out after 120 seconds during Go tests; generation, formatting, vet, and envtest setup ran first. No tracked files changed.
- Self-review only; no independent review or live Konflux/registry validation.

**Review focus:** Empty snapshot list, malformed digests, command failures, cleanup after failure, and registry names differing between CSV and snapshot.

**Constraints:** Use current oc namespace/context. Require existing registry credentials. No cluster mutation, image execution, dependency installation, or unrelated file changes.
