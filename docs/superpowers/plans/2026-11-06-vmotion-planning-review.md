# VMotion Planning Review Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix review findings #2 and #4 while retaining live drift detection and documenting remaining dispositions.

**Architecture:** Reject inconsistent pre/post-inventory Node identity snapshots with retry rather than freezing mismatched observations. Initialize shared Running lifecycle status before engine dispatch; leave legacy mutation guards in place.

**Tech Stack:** Go, controller-runtime, Kubernetes fake clients, existing test harness.

**Spec:** `docs/superpowers/specs/2026-11-06-vmotion-planning-review-design.md`

## Global Constraints

- No new npm dependencies; no dependency changes needed.
- Local commits in logical blocks with detailed descriptions; never push.
- Preserve unrelated .idea changes.
- Keep continuous frozen-plan drift detection and legacy execution guards.
- No CR edit guard, schema change, webhook, cache, concurrency framework or unrelated refactor.

## Review Focus

- Node list order changes alone must not trigger retries.
- Node replacement with same name but changed UID must not freeze mismatched observations.
- Existing frozen plans must survive transient identity drift unchanged.
- Repeated Running reconciles and pause/resume must preserve initial StartTime.
- Paused migrations must not acquire StartTime; legacy initialization remains functional.

### Task 1: Retry inconsistent identity snapshots

**Files:** modify `internal/controller/vmotion_plan.go`; test `internal/controller/vmotion_plan_test.go` (existing inventory callback/harness).

**Interfaces:** consumes `collectVMotionNodeIdentities(nodes []corev1.Node) ([]vmotionNodeIdentity, error)` and existing `retryVMotionPlan`; produces internal identity comparison only, no public API changes.

- [ ] Add `TestVMotionInventoryRetriesNodeIdentityDrift`: mutate Nodes during inventory callback (addition, removal, UID replacement, providerID change, role change); assert retry error and Retrying condition, no new plan, and existing frozen plan unchanged. Assert next stable reconcile can freeze initial plan.
- [ ] Add order-only case asserting successful freeze despite reordered Node lists.
- [ ] Run `go test ./internal/controller -run 'TestVMotionInventoryRetriesNodeIdentityDrift' -count=1`; confirm failure exposes missing snapshot consistency check.
- [ ] Add minimal order-independent name-keyed comparison of name, UID, role, providerID and normalized BIOS UUID. Preserve original identities; after fresh role/eligibility validation, retry before building candidate if identities differ. Retain all existing read-error and eligibility handling.
- [ ] Run focused regression plus existing VMotion planning/provisioning/stability tests; require PASS.
- [ ] Commit planner and tests with detailed rationale, retry behavior and validation evidence.

### Task 2: Initialize lifecycle status before VMotion dispatch

**Files:** modify `internal/controller/vmwarecloudfoundationmigration_controller.go`; test `internal/controller/vmotion_plan_test.go` and existing lifecycle tests as needed.

**Interfaces:** reuses `seedReadyCondition(migration)` and `reconcileWorkflow`; no new API.

- [ ] Add `TestVMotionPlanningInitializesLifecycleStatus`: first Running reconcile persists nonnil StartTime and Ready=False, including planning blocked before inventory. Repeated reconcile and pause/resume preserve timestamp; initially Paused migration does not start clock. Verify start event is not repeated.
- [ ] Run `go test ./internal/controller -run 'TestVMotionPlanningInitializesLifecycleStatus' -count=1`; require expected failure before fix.
- [ ] Move shared StartTime initialization and Ready seeding before VMotion dispatch within Running path; guard optional recorder if necessary. Remove duplicate seed call; leave legacy executionContext/checks untouched. Preserve persistence of seeded Ready on first VMotion status update.
- [ ] Run lifecycle, pause/resume and existing legacy controller tests; require PASS.
- [ ] Commit lifecycle code and tests with detailed status semantics and validation evidence.

### Task 3: Document dispositions and verify combined change

**Files:** create `docs/dev/vmotion-planning-review.md`; include approved spec and this plan in documentation commit.

**Interfaces:** documentation only.

- [ ] Record #1/#3 excluded engine-switch behavior; #2/#4 fixes; #5 standalone function versus adapter distinction and #231 evidence; #6/#7 deferred measured optimization retaining drift checks; #8/#10 deferred cleanup; #9 keep kind-specific semantics. State unsupported edits are not prevented and can enter legacy mutation path.
- [ ] Add future-work note to inventory documentation: optimize repeated authentication and inventory after measurement, preserving fresh eligibility/placement checks and bounded cancellation.
- [ ] Run `make test`, `make lint`, `make build`, focused VMotion race tests and `git diff --check`; report blockers accurately rather than claim unrun checks passed.
- [ ] Inspect combined diff for unrelated files, status semantics and identity-comparison edge cases; verify no dependency/schema/RBAC changes.
- [ ] Mark plan tasks with observed outcomes; commit documentation with detailed review decisions and verification evidence. Never push.

## Execution record

- [x] Task 1: ten identity-drift cases reproduced RED, then passed; list-order case passed. Committed as `fe71995b5`.
- [x] Task 2: missing lifecycle status reproduced RED, then passed; paused/blocked/retry/event cases covered. Committed as `294259df9`.
- [x] Task 3: developer dispositions documented; full tests, lint, build, race and diff checks run.
- Ruling: compare already-canonical sorted identity slices instead of allocating name-keyed maps; existing collector sorts by role/name. If collector ordering changes, comparison must be adapted.
- Review: direct execution and author self-review; no independent reviewer dispatched without explicit delegation authorization.

## Self-review

Spec coverage maps to tasks 1–3. All five review-focus cases assigned tests. No exported interfaces, generated files or dependencies added. Implementation remains in existing files/harness rather than introducing new framework.
