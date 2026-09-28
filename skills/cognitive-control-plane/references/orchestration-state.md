---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Orchestration State

Use for real delegated work, dependent stages, or a durable handoff. A single agent applying a skill or performing a source lookup does not need this layer.

## Responsibilities

- The coordinating agent owns the user goal, task boundaries, dependencies, and reconciliation. It may perform useful local execution while a worker handles independent work.
- A worker owns one bounded result and its validation, within the supplied write scope.
- A required independent reviewer inspects a stable version read-only and follows [reviewer enforcement](reviewer-enforcement.md).
- A host Scheduler exists only for managed queues; its leases and sessions follow [work-item scheduling](work-item-scheduling.md).

A role label does not create a process, permission, or independent actor. Claim delegation only after an actual worker starts. If optional delegation is unavailable, execute serially. If independence is required but unavailable, finish authorized preparation and report the remaining review gate accurately.

## Task Contract

Before delegation, provide enough information to work without guessing. Plain prose is sufficient unless a runtime requires a schema:

```yaml
objective: "One bounded result"
scope: [] # target artifacts and writable paths; read-only when none
inputs: [] # source locations, accepted decisions, relevant evidence
constraints: [] # permissions, ownership, explicit non-goals
required_skills: [] # only needed skills, resolved in the current environment
expected_output: "Artifact or findings with evidence"
validation: []
stop_if: [] # material scope expansion or an indispensable missing input
```

Use a real task id to track delegated work. Add dependencies and actor ids for multi-worker coordination; add producer identity and a pinned target when independent review is required. Add required references/tools only if the worker needs an exact capability. The portable host adapter's full schema is linked from [skill routing](skill-orchestration.md); do not manufacture its empty fields in ordinary conversation.

## Dependencies and Ownership

Parallelize only independent work. A producer must establish a shared API or schema before dependent implementation consumes it. Two writers must not own overlapping files or a shared mutable interface at the same time; narrow or serialize their scopes. Independent files can still conflict through generated artifacts or integration behavior.

Read-only discovery may overlap writing, but its conclusions must identify the inspected version. A required final review waits for a stable target. Before replacing a cancelled or failed writer, inspect partial changes: cancellation is not rollback.

## State and Reconciliation

For multiple tasks, keep a compact board: task id, owner, scope, dependency, state, result/evidence. Persist it only when needed for interruption, long-running coordination, or a consumer handoff. Reuse an existing project artifact when appropriate.

A durable handoff records the goal, constraints, completed work, artifact identities, validation, unresolved questions, and the next action. Include an event log for concurrent state transitions or audit requirements; ordinary serial work needs no scheduler-shaped database.

When a worker returns, compare its result with the contract, inspect relevant artifacts and verification evidence, resolve conflicts, and integrate the useful result. Self-reported skill/tool use is a lead, not proof. Recheck material dependent behavior after integration. Apply required review gates from their single [policy reference](reviewer-enforcement.md).

Finish when the authorized outputs are reconciled, applicable validation is reported, and any outstanding dependency has an explicit continuation or blocker. Return one coherent result to the user.
