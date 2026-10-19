---
name: cognitive-control-plane
description: "Unblock a task that is drifting, repeating failed attempts, or missing a consequential decision; critique a concrete plan or prepare a continuation handoff. Recover an actionable next step and resume authorized work. Clear tasks go directly to execution."
metadata:
  access:
    audience: model
    model_read: true
    model_write: true
    purpose: skill_runtime
---
# Cognitive Control Plane

Recover the next useful action, then carry the user's task through execution and validation. Resolving a blocker is an intermediate result unless diagnosis, planning, or a handoff is itself the requested deliverable.

## Establish the decision

Use the conversation and current artifacts to recover three things: the requested outcome, what is already established, and the unresolved decision that changes the next action. Keep this in working context; create a document only when the task needs one.

Carry forward accepted decisions and authorization. A correction updates the affected part of the task; completed work and the remaining objective still apply. A change of skill or phase does not require renewed approval. An explicit planning-only boundary still applies until the user authorizes implementation.

If preferences, configuration, usage habits, or connection details could affect the approach, read the configured memory index and the relevant records using [memory routing](references/memory-routing.md). Missing memory is unknown context; retrieve current facts or ask only if the missing fact matters.

## Resolve the blocker

Choose the row that changes the immediate action. Use another only when evidence reveals a different blocker.

| Blocker | Intervention | Ready when |
| --- | --- | --- |
| Goal or scope is ambiguous | Inspect available artifacts first. Ask a focused question for a consequential choice only the user can make; state a reasonable assumption for a reversible detail. Continue independent work while awaiting an answer. | The target, permitted action, and success condition are clear. |
| A factual or causal claim is uncertain | Name the claim and run the smallest check that distinguishes the plausible explanations. Check source applicability and freshness. | Evidence supports the next action, or a bounded, reversible action can proceed with the uncertainty stated. |
| A plan or artifact needs critique | Test it against the requested outcome. For each material defect identify the trigger, consequence, evidence, and correction. Label hypothetical risks; use the applicable review skill for code. | Supported defects have actionable corrections, and useful constraints are preserved. |
| Work repeats or drifts | Compare the latest result with the intended effect. Identify what changed since the last attempt; change the hypothesis or method when evidence did not improve. Check partial effects before retrying a write. | A new attempt has a reason to succeed, or a specific missing input or capability is identified. |
| Work must continue elsewhere | Produce a self-contained handoff with the outcome, constraints and authorization, current artifacts, completed work, verification, remaining work, and first next action. | The recipient can continue without rediscovering completed work or mistaking proposals for results. |

An unresolved premise blocks actions that depend on it. Continue useful independent work. When only the user can resolve a blocker, ask for the missing decision or input and explain what it prevents; elapsed time is not an answer.

## Resume execution

Use a specialist from the current skill catalog when the next action needs its procedure; read its actual instructions. A skill name is neither a worker nor proof of completed work. If an optional specialist is absent, use an adequate available method; report an indispensable missing capability and finish independent work.

Perform the authorized next action and validate the behavior that matters to the user. Continue until the requested outcome is reached or a concrete external dependency prevents further progress. Reopen diagnosis only when new evidence changes the decision. Avoid repeated routing, planning, or checks that cannot affect the next action.

Report the result, supporting evidence, and any material unverified part. Distinguish inspected source, tests, runtime checks, and independent review; none silently substitutes for another. Use the consumer's required format, and validate machine output against its actual schema.

## Conditional references

- **Actual delegation or concurrent workers:** read [coordination](references/orchestration-state.md) before dispatch. Delegate only when authorized; otherwise execute in the current session.
- **A required independent review:** read [review rules](references/reviewer-enforcement.md). Identify the applicable requirement before treating it as a gate.

Ordinary use needs no role catalog, task-size class, routing trace, scheduler, or host hook. Installation and optional host tooling are documented in the repository's `README.md`; they do not add prerequisites to this workflow.
