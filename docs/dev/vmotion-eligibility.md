# VMotion planning eligibility: Machine API evidence

This records the approved B/C review policy for the read-only planning foundation
(SPLAT-2703, PR #228). It concerns OpenShift Machine API resources. CAPI ownership
and Machine API-to-CAPI handover are outside this release's qualification scope.

## Read-only boundary

The planner reads every planned Node and the Machine API Machines, MachineSets,
ControlPlaneMachineSets (CPMS), and MachineHealthChecks (MHC). Required read errors
are retry errors, never empty-inventory or verified-unmanaged evidence. Conflicting
or unsettled observations produce actionable blocked planning status.

Eligibility is checked before source VM inventory and again afterward, using
fresh Node and Machine API reads before the candidate plan is persisted. This
closes the directly observed window where a new phase-less Machine can appear
during the vSphere call. It does not make cross-resource reads atomic or fence
future administrator/controller actions. Later execution must still perform fresh
admission and drift checks, and the supported maintenance procedure must keep
conflicting activity settled throughout migration.

The operator does not delete, adopt, scale, pause, repair, or clear associations
on these resources. An absence of active evidence is not cancellation of an
already submitted external operation or qualification of vCenter inactivity.

## B: phase and provisioning categories

Machine lifecycle phase is an observation, not an ownership or quiescence lock.
The five documented phases are `Provisioning`, `Provisioned`, `Running`, `Failed`,
and `Deleting`; the API field is an optional string, so absent, empty, and other
values must be handled explicitly.

| Observation | Planning policy and reason |
| --- | --- |
| Positive current Node UID/providerID/association | Managed; block irrespective of phase. First-release VMotion does not move Machine-managed Nodes. |
| Deletion timestamp, `Deleting`, `Provisioning`, `Provisioned` | Unsettled; block. Deletion can precede a phase update, a provisioning VM can already exist, and a provisioned VM may still join as a new Node. |
| Absent, empty or unrecognized phase | Indeterminate; block. Empty phase is a normal early reconciliation state; an unexpected nonempty phase is not a native actuation stop signal. |
| `Running` without a verified current Node association | Inconsistent/indeterminate; block. The plan covers all current Nodes, and Running alone proves neither readiness nor valid ownership. A genuinely current Machine-managed Node is already rejected above. |
| Unassociated `Failed` | Candidate historical residue only. The Machine controller stops ordinary actuation, but parent/GC/remediation context must also be settled. |

Ownership/conflict evaluation precedes phase classification, so an existing
association is never ignored merely because its Machine is Failed or has an
uninitialized phase. Read failures, missing controller objects, changed owner
UIDs, or unsupported controller references cannot establish inactive history.

### MachineSet intent before a Machine exists

A MachineSet can create a Machine as soon as its selected child count is below
desired capacity. No child phase exists in that window. The gate therefore checks:

- No parent deletion and a known, nonnegative desired count.
- A valid nonempty selector matching the template.
- Current observed generation and no nonempty terminal error diagnostics.
- Actual selected children, including adoptable orphans, rather than only status
  replica summaries. Machines controlled by another object are not counted as
  this MachineSet's children, matching the native controller's selection.
- No selected child in a pending/failed/deleting state; all remaining children
  must be Running.
- Desired, actual, status replicas, Ready and Available counts agree.

Zero-replica MachineSets with current settled evidence and no selected children
are allowed. A positive desired count is not itself proof that planned Nodes are
Machine-managed; it must still be checked against actual children and ownership.
Zero status replicas cannot hide an outstanding deleting child.

### CPMS and remediation context

An Inactive, non-deleting CPMS is allowed; it does not create or replace Machines
in normal reconciliation. Existing child lifecycles are still checked separately.
An Active CPMS requires a known desired count, valid selector, Running selected
children, matching desired/actual/ready/updated counts, no unavailable replicas,
current `Available=True`, `Progressing=False`, and `Degraded=False` conditions,
and no unresolved controller Error condition. Missing/replacement children or
stale status cannot be interpreted as quiescence.

Controller-owned Machine references must resolve to the observed MachineSet or
CPMS with the same namespace/name, UID, kind, and supported API identity. An
unresolved or unsupported controller is Indeterminate, not inactive history.

A matching MHC that can remediate a Machine blocks planning. Native internal
remediation skips ownerless Failed Machines, so an unowned Failed historical
object without an external remediation template remains eligible as residue.
External templates or a controller-owned target do not have that exception. A
pause annotation alone does not prove that previously requested remediation has
ended; this gate does not invent a remediation cancellation mechanism.

## C: resolved-absent annotations and conflicts

NodeLink writes `machine.openshift.io/machine` as `namespace/name` after discovering
a Machine by providerID or internal IP. The annotation does not contain a Machine
UID and is not NodeLink's discovery key. If discovery finds no Machine, NodeLink
returns without clearing the annotation. Therefore:

- Successful complete reads can resolve a parseable reference as **absent**.
  That historical hint is not by itself proof of current Machine ownership and
  does not permanently block the Node.
- This removes only one ownership signal. Unique Node/VM identity, no other
  supported association/conflict, a non-terminating Node, and settled B checks
  are all still required before a plan can freeze.
- Failed reads, malformed references, conflicting Node UID/name/providerID
  evidence, and ambiguous associations are **Indeterminate** and block planning.
- A live Machine annotation cannot override contradictory UID or providerID
  evidence. Namespace/name reuse does not recover the identity of a deleted
  Machine incarnation.
- A Machine internal-IP match without a positive identity association is a veto:
  NodeLink has an IP fallback path, but IP alone is not sufficient VM identity
  here. Such candidates block as Indeterminate; they are never adopted. The check
  covers both IPv4 and IPv6 internal addresses without address conversion.
- Terminating Nodes block independently. During native deletion, a Machine can
  disappear after requesting Node deletion while the Node still remains.

Historical annotations are retained unchanged. The current Node UID/providerID
and BIOS-UUID identity contract remains the execution identity; no IP-based VM
lookup, name-based VM adoption, or automatic annotation repair is introduced.

## Deferred limitation A

Changing/removing the engine selector after a read-only plan is frozen but before
an active action exists can still dispatch into the legacy workflow. This is a
known, explicitly deferred flaw, not authorization to abandon a frozen execution.
The B/C changes do not implement an execution-reset policy or engine-switch guard.

## Source basis and qualification

The policy was researched against Machine API Operator `release-5.0` at
`fcb64975b70e848f193ffd8e9beb1b2a413d84d1` and CPMS Operator `release-5.0` at
`c00cf58dd81954ca100757f6a2d2af8d84770fd5`, notably Machine reconciliation,
MachineSet child selection/status, MHC remediation, and NodeLink discovery.

Unit and simulator tests exercise these planning checks. They do not qualify an
exact OpenShift payload, concurrent external controllers, vCenter operations, or
the operational maintenance/quiescence procedure for release enablement.
