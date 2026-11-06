# VMotion planning review fixes

Approved in conversation following review of PR #228 and dependent PR #231.

## Intent

Avoid misleading inventory failures when Node identities change during a read, and expose consistent lifecycle status from planning start.

## Design

- Compare pre/post-inventory Node identity sets without depending on list order. Changed name, UID, role, providerID, or BIOS UUID requires retry; never freeze a candidate built from mismatched snapshots or replace an existing frozen plan.
- Initialize StartTime and Ready=False before VMotion dispatch for Running migrations. StartTime means planning start. Preserve timestamps across retries/resume and keep the legacy execution guard unchanged.
- Keep continuous frozen-plan drift detection. Document future authentication/inventory optimization without weakening live checks.
- Record dispositions for review findings #1–#10 in developer documentation.

## Approaches

Selected: retry on the next reconcile when identity snapshots differ. Immediate inventory refetch adds work and can chase ongoing drift; removing post-inventory validation weakens safety.

Selected: shared lifecycle initialization before engine dispatch, rather than duplicated VMotion-only initialization. Keep execution-specific guards in legacy path.

## Constraints and non-goals

- No new npm dependencies; no dependency changes needed.
- Local commits in logical blocks with detailed descriptions; never push.
- Editing migration CR spec during migration is excluded from these fixes. CRD schema compatibility is distinct; no schema change or webhook needed here.
- No session cache, concurrency framework, timeout adjustment, generic settled predicate, or unrelated refactor.
- No assumed performance target: cluster/network-dependent inventory costs remain unmeasured; VM action duration is separate.
- Preserve unrelated .idea changes.

## Scope

Controller planner and lifecycle code, focused regression tests, developer review-disposition documentation. No API, generated schema, RBAC, fixture or dependency changes expected.
