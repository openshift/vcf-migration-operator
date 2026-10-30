# Node Migration API (SPLAT-2949)

**Kind:** `VmwareCloudFoundationMigration` · **Group/Version:** `migration.openshift.io/v1alpha1`

This API builds on the existing per-role progress and per-node map list, adding
bounded execution identity, action outcomes and recovery-request data to that
structure.
Counts and summary strings describe observations; they are not authority to
replay an operation.

> **Bounded runtime foundation.** The controller implements explicit VMotion
> dispatch, eligibility checks, and read-only frozen-plan creation. It does not
> run the node migration workflow. The durable action component is implemented
> and tested with fake backends; production action submission remains disconnected.
> Existing claims can be observed across pause, spec edits, and deletion. Recovery
> requires an injected read-only evidence verifier; no production verifier is
> configured. Control-plane health gates, cleanup, and qualification remain
> follow-up work. See [durable execution](vmotion-execution.md).

## Selection and validation

When `spec.nodeMigration` is omitted, the existing legacy MachineSet/CPMS
replacement selection and template requirement are retained. When present, both
roles are required and must explicitly select the same engine with `type`. There
is no default engine selection; omitting either role or its `type` is rejected.

```yaml
nodeMigration:
  workers:
    type: VMotion
    vmotion:
      mode: Hot
    maxUnavailable: 1
  controlPlane:
    type: VMotion
    vmotion:
      mode: Hot
    maxUnavailable: 1
```

`VMotion` requires an explicit `vmotion.mode` of `Hot` or `Cold`. Both roles
must use the same mode. There is no `Auto` mode, default, or automatic fallback;
an execution backend must fail rather than change modes if the selected mode is
unsupported.
VMotion limits worker `maxUnavailable` to `1` and consumes the effective
source-to-destination mapping resolved during preflight; it adds no single-FD or
one-to-one mapping restriction of its own. Each planned node must resolve to one
target FD or execution is blocked. Every destination failure domain must include
an existing `topology.template` path for this release, including VMotion
configurations. VMotion relocates existing VMs and does not consume the template
during relocation. Recreate worker availability may exceed one.
`maxUnavailable` is an integer node count from 1 through 32767 (default 1);
percent strings are invalid.

The `vmotion` block is required for VMotion and forbidden for Recreate. Mixed
VMotion/Recreate role selections and mixed Hot/Cold modes are rejected by
structural schema/CEL validation. `status.failureDomainMappings` is the
preflight-resolved mapping consumed by VMotion; the path does not recalculate
it or infer a mapping from failure-domain counts or names.

## Runtime selection and frozen plan

Any explicit VMotion role selection enters the VMotion path before legacy
template/probe and MachineSet/CPMS processing. Invalid partial or mixed engine
selections are blocked and never fall through to replacement. The executable
resource remains only `openshift-vcf-migration/cluster`.

The plan includes every Kubernetes Node; there is no selector. Only a cluster
with exactly three dedicated control-plane Nodes and at least one worker is
eligible. Dual-role and unclassified Nodes block the entire plan. Non-Node VMs,
including load balancers, are outside operator scope.

Each Node's `spec.providerID` must be a canonical `vsphere://<BIOS UUID>` and
must resolve case-insensitively to exactly one VM `config.uuid` in a container
view of the single source datacenter. Missing, malformed, zero, or ambiguous
identities block the entire plan. A single source failure domain receives every
Node through the same ownership check used for multi-FD clusters: resolve every
source FD's `topology.computeCluster` path in the configured source
vCenter/datacenter and bind its managed-object reference to the FD name. Reject
duplicate references. Match each VM's `resourcePool.owner` to exactly one
resolved FD by vCenter, reference type and ID, and datacenter scope; do not use
resource-pool path/name as a tiebreaker or require a particular compute-resource
type. If the Node has `topology.kubernetes.io/zone`, it must equal the matched
FD's `Zone`; missing labels are allowed. Missing or ambiguous evidence blocks
the whole plan.
The source inventory connection verifies TLS against the operator's system
trust roots; an untrusted certificate blocks plan creation.
This owner-reference workflow is agreed in principle but has not been
runtime-tested against the intended environment.

Machine ownership is checked per planned Node through `Machine.status.nodeRef`
UID, Machine `spec.providerID`, and a Node's `machine.openshift.io/machine`
annotation when it resolves to a live Machine. A resolved-absent annotation is
retained and permitted only after complete ownership, identity, non-termination
and quiescence checks. Contradictory live associations and IP-only candidates are
Indeterminate, not proof of ownership or being unmanaged.

Machines in Provisioning, Provisioned, or Deleting phase, or with a deletion
timestamp, block admission separately. Empty/unrecognized phases and inconsistent
Running evidence also block. Unassociated Failed history requires resolved
parent/remediation context. The planner reads MachineSets, CPMS and MHCs before
and after VM inventory, compares parent intent/status with real children, permits
settled zero-replica MachineSets and Inactive CPMS, and rejects unsettled capacity,
replacement or applicable remediation. These are read-only observations, not
external-operation fencing. See [the approved B/C policy](vmotion-eligibility.md)
for exact checks, CAPI scope, qualification limits and deferred engine-switch A.

The plan consumes the effective mapping in `status.failureDomainMappings` and
freezes its execution identity, source/destination, mode, Node/VM identities,
original cordon state, selected source FD and owner reference in the sanitized
plan fingerprint, selected target failure domain, resource pool, datastore,
and one destination network for the supported one-NIC VM. Plans block if either
the VM or destination specification is ambiguous about that one-NIC mapping.
Host is deliberately not pinned: the relocation request leaves
Host unset for destination-cluster placement, then the workflow records the
observed destination host in `status.nodeMigration.*.nodes[].targetHost` after
the move. Reconciles reuse the frozen plan and block on identity, mapping, or
placement drift. PR 2 stops after this read-only plan; it issues no VM actions.

### Shared action/backend contract

`internal/vsphere/vmotion_contract.go` is the shared contract for PRs 3 and 4.
Inventory, feasibility checks, VM inspection, and task observation are
read-only calls. `ShutdownGuest`, relocation, and power-on are mutating calls
and may only follow an acknowledged durable action claim. An uncertain submit
returns `Unknown`; that state is observed or escalated, never resubmitted. Task
IDs always travel with their owning vCenter, `ShutdownGuest` has no task
reference, and shutdown, relocation, and power-on outcomes remain independent.
After a claim, observing its task/VM and persisting that result remains allowed
across `Paused` or generation changes when the CR UID, execution ID, and action
ID still match; this observation permission never allows a new claim or
mutation. Result merging must not let stale writers replace a terminal or newer
action.

## Examples

### Legacy selection

Omit `nodeMigration` to retain legacy behavior. The failure domain must include
the template used by the existing replacement path.

```yaml
spec:
  failureDomains:
    - name: target-fd-1
      region: target-region
      zone: target-zone-1
      server: vcenter-target.example.com
      topology:
        datacenter: TargetDC
        computeCluster: /TargetDC/host/TargetCluster
        datastore: /TargetDC/datastore/TargetDS
        networks: ["VM Network"]
        resourcePool: /TargetDC/host/TargetCluster/Resources
        template: /TargetDC/vm/rhcos-template
        folder: /TargetDC/vm/my-cluster
```

### Recreate selection

Both roles must explicitly select Recreate. The template remains required.
Worker concurrency greater than one remains valid for this engine.

```yaml
nodeMigration:
  workers:
    type: Recreate
    maxUnavailable: 3
  controlPlane:
    type: Recreate
    maxUnavailable: 1
```

### Hot and Cold API selections

For VMotion, choose one explicit mode and use it for both roles. The target
failure domain must include `topology.template`, as required for all engines.

Hot selection:

```yaml
nodeMigration:
  workers: {type: VMotion, vmotion: {mode: Hot}, maxUnavailable: 1}
  controlPlane: {type: VMotion, vmotion: {mode: Hot}, maxUnavailable: 1}
```

Cold selection:

```yaml
nodeMigration:
  workers: {type: VMotion, vmotion: {mode: Cold}, maxUnavailable: 1}
  controlPlane: {type: VMotion, vmotion: {mode: Cold}, maxUnavailable: 1}
```

These examples describe valid API shapes only; they do not assert that the
current controller executes either mode.

### Recovery request shape

A request identifies a request, frozen execution, and action. It is a request
for runtime rechecking, not proof that the action is inactive or authorization
to replay it. CEL validates the bounded shape, not equality to the live action.

```yaml
nodeMigration:
  workers:
    type: VMotion
    vmotion: {mode: Cold}
    maxUnavailable: 1
  controlPlane:
    type: VMotion
    vmotion: {mode: Cold}
    maxUnavailable: 1
  recoveryRequest:
    requestID: 11111111-1111-4111-8111-111111111111
    executionID: 22222222-2222-4222-8222-222222222222
    actionID: 33333333-3333-4333-8333-333333333333
    reason: Administrator completed manual inspection
```

Changing `spec.state` to `Running`, timing out, or changing a request does not
clear intervention, prove inactivity, or authorize replay. Pause A permits
observation of existing work but starts no new actions; after a Cold relocation,
the VM may remain off until resume.

Expected recovery refusals are reported by the bounded `VMotionRecovery=False`
condition (`RecoveryBlocked`), preserving normal observation polling. A verified,
consumed request records `VMotionRecovery=True` (`RecoveryVerified`). These
conditions describe the recheck result, not mutation authorization. Transient
read/write failures remain reconcile errors; repeated identical policy refusals
do not continuously rewrite status or emit events.

## Status model

`status.nodeMigration` retains `type`, human-readable `progress`, `controlPlane`,
and `workers`. Each role retains requested/observed mode, counts, and `nodes` as
a map list keyed by Kubernetes Node `name`. Counts and strings are observations.

### Execution and frozen plan

The optional status fields `executionID`, `migrationUID`, `planFingerprint`,
and `frozenPlan` identify the CR UID and sanitized immutable plan. The execution
and action IDs are independent of generation and remain stable across pause
generation changes. The frozen plan contains source/destination vCenter
identifiers, source datacenter, and selected mode. The preflight-resolved
`status.failureDomainMappings` remains the source-to-destination mapping; each
per-node record freezes its selected target failure domain and placement. The
plan must not contain credentials, TLS material, ServiceLocator secrets, SDK
objects, full task results, or an event journal.

Each existing per-node entry adds `nodeUID`, `providerID`, `vmBIOSUUID`,
`originalSchedulingState` (`Schedulable` or `Unschedulable`), and frozen target
failure-domain/datacenter/resource-pool/datastore/network fields. The optional
`targetDatacenter` is populated when freezing each node and is required by the
VMotion claim path. Older plans missing it are blocked; execution never fills
it from the live spec. `targetHost` remains
empty until the workflow records the observed destination host after relocation;
it is not a pinned placement. An omitted scheduling state means the original
state was not recorded. Existing source/target inventory paths, failure-domain,
mode and timestamps remain available. `vmBIOSUUID` is the VM BIOS UUID and does
**not** mean vSphere `config.instanceUuid`.

### Action and outcomes

`activeAction` carries one stable action ID/kind, execution ID, Node/VM identity,
owning vCenter, non-secret operation identity, submission and observation
states, optional task reference, timestamps/deadline, and a sanitized reason.
Task IDs are scoped to their `ownerVCenter`. `ShutdownGuest` has no task
reference. A durable claim without a task reference means submission may have
begun and is not safe to retry.

Each per-node entry has separate `shutdownOutcome`, `relocationOutcome`, and
`powerOnOutcome` fields. A later power-on result must not overwrite relocation
success. Outcome/count/state fields do not independently authorize another
action. `intervention`, `recoveryAcknowledgement`, and `cleanup` are bounded
records; their presence or values are not runtime proof or authorization.
The action component enforces identity-bound claims and recovery consumption;
cleanup enforcement remains follow-up work.

`Preparing` is a coarse pre-action/scheduling-preparation state and does not
imply that only Cold mode drains. `Succeeded` means required action, identity,
placement, and recovery verification completed; a stale Node Ready condition
alone is insufficient. These semantics are API documentation, not implemented
controller guarantees until the corresponding runtime stage is merged.

## Validation and runtime boundary

The CRD enforces static enums, required fields, the engine/mode pairing,
conditional template and failure-domain limits, bounded list/string sizes,
task-reference ownership shape, outcome-kind pairing, UUID/identity formats,
and recovery-request shape. It intentionally does not freeze all spec fields,
bind task results to current generation, require a retained recovery request to
match whichever action later becomes active, or claim live ownership/health was
checked.

Runtime work still needs integration with fresh workflow admission and recovery
evidence, cleanup enforcement, and the version/security/operational qualification described in
[`upi-cross-vcenter-vmotion-design.md`](../plans/upi-cross-vcenter-vmotion-design.md).
