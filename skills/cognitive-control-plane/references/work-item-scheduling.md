---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Work-item Scheduling

Load only when a host actually manages a queue, leases, or fresh-session continuation. Ordinary conversation, a long task, or a user mentioning a ticket does not install a Scheduler.

The host's configured budget and state are authoritative. Do not apply this bundled profile's percentages to unrelated sessions or invent telemetry. For the portable adapter use [the machine map](../config/skill-orchestration-map.yaml) and [contract schema](../adapters/contract.schema.json). The scripts validate host decisions; this document grants no launch or write permission.

## Work Item Scheduler

The Scheduler is the host-side control process for durable work. It is not a
specialist skill and it is not the Runner. Normalize every accepted `issue`,
`request`, `transaction`, or `ticket` into one **work item**. A **run** is one
session attempt for that same item.

Before a run starts, the Scheduler must:

- confirm dependencies are terminal and accepted where required;
- acquire one live lease for the work item and recover only expired leases;
- resolve overlapping write ownership before dispatch;
- allocate a per-run budget; and
- create a self-contained run contract with its attempt number and any prior
  checkpoint.

The Runner may investigate, plan, execute, validate, and reflect within one
run. It must not claim another item, bypass a lease, or turn a checkpoint into
a new work item. The Scheduler writes the durable transition after reconciling
the run output.

Use these states:

```text
work item active: ready -> leased -> running -> validating
work item terminal: resolved | concluded | duplicate | blocked | escalated | cancelled
run: scheduled | leased | running | checkpointed | completed | expired | cancelled
```

`resolved` requires validation evidence. `concluded` requires evidence for a
no-change or diagnostic conclusion. `blocked` and `escalated` stop automatic
retry until an external dependency, permission, or decision changes. A
`transaction` also needs an idempotency key before a write-capable run.

## Budget And Continuation

Use the actual thresholds configured by the host. The bundled scheduler defaults below are a host profile, not a universal model context limit:

- at **40%**, persist a checkpoint containing completed evidence, artifacts,
  validation, next action, and residual risks;
- at **45%**, stop expanding scope and only finish the current atomic action,
  validate it, or prepare handoff;
- at **50%**, end the run. Do not rely on the session to finish naturally.

When the run ends checkpointed, is interrupted, or its lease expires, the
Scheduler creates the next attempt for the same work item and passes the latest
checkpoint. It must not duplicate the item or discard the prior run's event
log. If normalized budget telemetry is unavailable, do not claim enforcement;
record the limitation and use a conservative checkpoint/handoff instead.

Completion criterion: a later session can continue the same work item from
explicit state, and no two live runs hold its lease.

## Fresh-Session Dispatch

The host may poll or receive a run-completion event, then run the deterministic
`scripts/work-item-scheduler.js` against the persisted work-item snapshot. Its
output is a decision, not an implicit process launch:

- `dispatch` starts the first fresh session for a ready item;
- `checkpoint` tells the active Runner to persist its handoff before it ends;
- `continue` is allowed only after the previous run is `checkpointed` or
  `expired`, and requires a new run id, a higher attempt number, and a durable
  checkpoint reference;
- `verify`, `close`, `wait`, and `wait_for_human` never start a worker.

`scripts/work-item-loop.js` binds a `dispatch` or `continue` decision to a
validated portable contract and then invokes the adapter. It is dry-run by
default. With `--execute`, it starts a new `codex exec` or `opencode run`
process. It never uses native resume/continue flags, so a continuation is a
fresh session carrying explicit state rather than a hidden session dependency.

Completion criterion: an active old run cannot overlap a successor run, and a
new process can only start from an accepted scheduler decision and matching
contract.

## Programmatic Tool Calling Boundary

Programmatic Tool Calling (PTC) is an in-run optimization, not the durable
Scheduler. Use it inside one model execution for predictable, bounded tool
batches: repeated read calls, filtering, joining, aggregation, and mechanical
validation. Return a small structured candidate decision to the host.

Do not use PTC as the source of truth for leases, durable work-item state,
session launch, approvals, external writes, or a semantic decision that needs
fresh model judgment. The PTC runtime has no durable filesystem, network, or
process state; the host Scheduler must persist the event/result, evaluate the
deterministic policy, and explicitly launch or withhold a fresh Runner.

Completion criterion: every automatic session launch can be explained by
persisted state plus a deterministic scheduler decision, not by transient
in-run state.
