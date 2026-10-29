# Durable VMotion execution (SPLAT-2703, implementation-plan PR 3)

## Scope and story mapping

PR 3 implements the action-state portion of
[SPLAT-2703](https://redhat.atlassian.net/browse/SPLAT-2703). That story's scope
also covers PR 2's runtime planning. SPLAT-2704's title says durable execution,
but its description and acceptance criteria describe PR 4's vSphere adapter.
SPLAT-2703 is the confirmed attribution for this change.

The authoritative contract is the [first-release design](../plans/upi-cross-vcenter-vmotion-design.md),
especially sections 7 and 9. This component does not enable migration execution.
No production code calls `submitVMotionAction`; later worker/control-plane
orchestration must first perform the fresh plan, feasibility, protected drain,
availability and health admission checks. The vSphere adapter remains PR 4 work.

## Claims and submission

`internal/controller/vmotion_execution.go` uses the existing CR status and
`APIReader`, with resourceVersion-protected status updates and bounded conflict
retries. No lock object, journal, service or new dependency is introduced.

Each write checks CR UID, execution ID, sanitized plan fingerprint, frozen plan
and frozen per-node identities/placement. Claim writes also require the canonical
namespace/name, the expected generation, Running, no deletion, no intervention,
and the exact previous action snapshot. Every retry rereads these gates. This
distinguishes an unrelated status conflict from a competing claim or pause.

Only a successful response to a newly created claim reaches a mutating backend.
A lost acknowledgement stops submission even if Kubernetes stored the claim.
A later reconcile never calls a mutation merely because it sees `Claimed` or
`Unknown`. This deliberately permits conservative intervention after a crash
between claim acknowledgement and the actual request: Kubernetes and vCenter
do not provide an atomic transaction or exactly-once execution.

The claim includes a new action UUID, node/VM identity, owning vCenter, an
operation digest derived from sanitized frozen inputs, and a deadline. The
submission call uses the claimed snapshot, so a later pause does not discard its
result or change the intended operation. Pause policy A allows a prior winning
claim to reach vCenter; pause is not cancellation.

Acknowledged relocation/power-on submissions require a nonempty bounded task ID
owned by the action's vCenter. Shutdown must have no task reference. Missing,
wrong-owner, contradictory or uncertain submission evidence requires
intervention. An ordinary backend error with no acknowledgement/task is a
definite rejection under the shared backend contract. Error strings and task
output are not copied into status; stored reasons are controller-owned text.

## Observation and deadlines

Reconcile handles an existing VMotion action before normal pause/spec-selection
and deletion workflow gates. It can persist matching results after generation
changes, including removal of the VMotion selector. The existing general
`executionContext` mutation guard remains intact.

Results merge only into their matching action. A concurrent task binding,
terminal result or newer action prevents a stale observation from overwriting
it. Unrelated status fields survive conflict retries. A skipped stale observation
is retried by the next reconcile; it does not authorize mutation.

Each observation and recovery verification has a separate 30-second read budget,
matching the existing long-wait polling interval. Backend/verifier implementations
must honor cancellation. A timed-out read leaves uncertainty/intervention intact,
returns the worker to reconciliation, and permits subsequent deletion handoff.
Observation timestamps also throttle polls: the CR watch sees status updates,
so a self-generated status event must not immediately call the backend again.
Spec/deletion events are still processed; scheduled observation resumes after
the polling interval. Tests use Go's virtual-time support rather than real waits.

Task observations use the stored owning vCenter, never a current destination
selection. Missing/unknown task evidence is uncertainty, not proof of inactivity.
Shutdown observes the frozen BIOS UUID in the frozen source datacenter and
requires actual `poweredOff`; request acknowledgement alone is not completion.
No hard power-off or source-recovery power-on operation exists here.

All three action deadlines initially use **six hours**, as requested for this
implementation. `TODO(SPLAT-2703)` beside the constant requires qualification of
this value. It is not a qualified release timeout. Expiry blocks further claims
and retains intervention, but observation continues. Late success updates its
operation outcome without clearing intervention or permitting another action.

Shutdown, relocation and power-on write separate per-node outcome fields.
Relocation success therefore survives a later power-on failure. Task success
does not mark the Node complete: final VM placement, fresh Node heartbeat and
applicable etcd recovery remain later workflow checks.

## Recovery and deletion

`internal/controller/vmotion_recovery.go` validates request/execution/action
identity and consumes a request only once. A read-only `VMotionRecoveryCheck`
must establish claimant quiescence and authoritative outcome/inactivity, unique expected VM location and
power, preserved Node identity, applicable health, and unchanged live plan and
inventory. It receives an execution snapshot and must return a terminal action
outcome, not a new operation to perform. Nil, unavailable, conflicting or
nonterminal evidence leaves intervention in place and does not consume a request.
Any resourceVersion change during verification requires a fresh recheck.

A `Claimed` action may still have a suspended original submitter. Recovery is
therefore blocked until its persisted submission deadline has expired; missing
deadlines fail closed. The mutation backend must check cancellation before
initiating a request, including after any internal wait. Expiry alone still
does not prove claimant quiescence or external inactivity: the recovery verifier
must establish both before consuming the request. Qualifying those backend and
evidence semantics remains a prerequisite to production recovery enablement.

There is no production recovery verifier yet, because the required workflow
health/placement checks belong to later PRs. Recovery requests therefore fail
closed until that integration is supplied. This is an explicit dependency,
not an administrator assertion treated as proof.

After successful verification, the bounded acknowledgement retains the consumed
request/execution/action IDs. Existing terminal outcomes are preserved. A failed
operation can be explicitly retried with a new action ID through the normal
admission and Running gates; an unknown or successful operation cannot be
blindly replayed. Pause still prevents new claims after a recovery acknowledgement.
The later workflow must verify a healthy destination without relocating again.
There is one outcome per kind rather than an attempt history; the acknowledgement
retains the prior action's identity when an explicitly authorized retry replaces
that kind's outcome.

Deletion observes an existing action while retaining the migration record; it
does not run preparation, legacy progression or new mutations. Unresolved action
state prevents completion-based finalizer release. The existing explicit
`migration.openshift.io/allow-deletion: "true"` handoff can release the record
even when the backend is unavailable. Handoff does not establish success, cancel
a task, recover a VM, or authorize cleanup.

## Verification boundary

Focused tests cover competing claims and resourceVersion conflicts, lost claim
acknowledgement, claim/call/task-persistence crash windows, pause before/after
claim, restart uncertainty, task ownership/absence, stale terminal/newer-action
writers, execution identity replacement, shutdown timeout/late completion,
independent relocation/power-on outcomes, recovery consumption and stale evidence,
and paused/deleting reconciliation with explicit handoff.

The tests use the real controller status logic with a Kubernetes fake client
and fake external backend. They do not qualify real cross-vCenter SDK retry
behavior, shutdown/boot timing, platform health checks or release combinations.
