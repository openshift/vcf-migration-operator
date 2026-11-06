# PR #228 planning review disposition

This records the approved scope for the external ten-finding review of the
read-only VMotion planning foundation. StartTime means planning start, not the
first VM-changing action. Production VM actions remain outside this PR.

| Finding | Disposition | Rationale |
| --- | --- | --- |
| 1. Successful condition cannot become False | Deferred with engine switching | Reproduction requires changing migration CR engine selection during migration, excluded from this PR's supported behavior. Do not weaken shared concurrent-status protections for this case. |
| 2. Stale inventory during Node drift | Fixed | Compare canonical pre/post-inventory Node identity snapshots. Retry inconsistent observations instead of freezing a candidate or reporting a misleading missing-VM failure. Keep any existing frozen plan unchanged. |
| 3. Switching away from frozen VMotion plan | Accepted limitation | Already documented as deferred limitation A in `vmotion-eligibility.md`. Unsupported spec edits are not prevented and can dispatch the mutating legacy workflow. |
| 4. Missing StartTime/Ready | Fixed | Shared Running initialization precedes engine dispatch. Planning seeds Ready=False and records StartTime once, preserving it across retries and pause/resume. |
| 5. Backend/function signature difference | Not a demonstrated defect | Interface method accepts vCenter/datacenter; standalone helper accepts an established session. An adapter bridges these layers. PR #231 still uses the standalone helper for planning and explicitly defers the production vSphere adapter. No signature change needed here. |
| 6. Repeated authentication/inventory after freeze | Optimization deferred | Live observations establish drift, including source resource-pool owner changes. Stored fingerprint and CR generation alone cannot establish unchanged vSphere state. |
| 7. Sequential per-VM lookups | Optimization deferred | Cluster/network-dependent cost remains unmeasured. Reuse shared resource-pool observations within a read before considering bounded concurrency; no performance target assumed from API list limits. |
| 8. Repeated condition/event plumbing | Cleanup deferred | Helpers have different event, error and requeue behavior. No generic status-update abstraction introduced. |
| 9. Kind-specific settled checks | Keep current design | Machine, MachineSet, CPMS and MHC have distinct lifecycle semantics. Shared selector helper already exists; generic settled predicate would obscure safety rules. |
| 10. Repeated eligibility reads | Cleanup deferred | Reads before and after inventory are intentional. Extraction must preserve fresh post-inventory Node/Machine API reads and retry-versus-blocked classification. |

## Future inventory optimization

Keep continuous frozen-plan drift detection. Measure repeated authentication,
full-datacenter inventory retrieval and per-VM placement lookup costs before
changing the implementation. Session reuse, shared lookup reuse or polling
changes must preserve fresh eligibility/placement checks and bounded cancellation.
Unchanged CR generation or stored plan fingerprint is not a live inventory check.
VM relocation taking hours does not imply inventory reads should have the same
budget; qualify these separately.

## API boundary

Editing migration CR spec during migration is excluded from these fixes. CRD
schema compatibility remains an API upgrade concern, addressed by compatible
schema evolution, synchronized generated manifests and upgrade/validation tests.
A migration CR validating webhook does not intercept edits to the CRD object.
Conversion webhooks are relevant when API versions need representation conversion;
no schema or webhook change is required for these fixes.

## Evidence and limits

Identity-drift regression cases cover additions, removals, UID replacement,
providerID and role changes during inventory both before and after plan freeze,
plus successful retry once stable and list-order independence. Lifecycle cases
cover fresh and blocked planning, initially paused migrations, pause/resume,
preserved timestamps and a single start event.

Verified locally: `make test`, `make lint` (0 issues), `make build`, focused VMotion
race tests and `git diff --check`. No real cross-vCenter or cluster E2E
qualification performed. PR #231 inspected as of head
`4bde47205c5d6505db63587a3414240ce8ff0ff9`; future adapter compatibility must be
checked when that implementation exists.
