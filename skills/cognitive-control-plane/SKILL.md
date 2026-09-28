---
name: cognitive-control-plane
description: "Choose the next useful action when ambiguity, conflicting evidence, plan critique, or coordination blocks complex work. Use for process diagnosis and multi-stage handoffs; clear bounded tasks go directly to their applicable skill or execution."
metadata:
  access:
    audience: model
    model_read: true
    model_write: true
    purpose: skill_runtime
---
# Cognitive Control Plane

Resolve the process bottleneck, then continue the user's task. This skill selects an approach; it does not create a separate agent identity or require a handoff to another session.

## Choose the Minimum Intervention

Use existing context and authorization. A bounded read can establish scope before choosing a route. Classify only if it changes execution or a caller requests a trace:

- **Tiny**: a direct explanation or recall with no substantive artifact work.
- **Small**: bounded inspection, lookup, review, or edit with a clear goal and manageable local risk.
- **Large**: consequential uncertainty, interacting ownership boundaries, high-impact changes, or work that needs staged coordination or durable continuation.

Skill use, repository access, current-source lookup, file count, and a requested output format do not by themselves make work Large. Investigate an unknown in proportion to its possible impact. If the next action is already clear, load the applicable skill if needed and execute it in the same session.

## Select the Actual Bottleneck

Read only the reference that changes the next action. These are decision aids, not a required four-stage pipeline.

| Bottleneck | Action and reference | Enough to proceed when |
|---|---|---|
| Missing goal, scope, artifact location, or current system behavior | [Context control](references/context-control.md): retrieve available facts; ask for a material user choice only when needed. | The next bounded action and its success condition are known. |
| A load-bearing factual or causal claim lacks support | [Epistemic control](references/epistemic-control.md): check evidence or define a discriminating test. | The claim is supported, falsified, or its uncertainty is explicitly handled. |
| A concrete plan, design, skill, or approach needs critique | [Adversarial control](references/adversarial-control.md): judge against task-specific criteria. | Material failures, mitigations, and remaining uncertainty are clear. |
| A consumer needs a handoff or structured deliverable | [Output control](references/output-control.md): meet the consumer's actual contract. | The consumer can act without reconstructing missing context. |

When several concerns apply, resolve the one that blocks the next useful action. Missing artifact identity precedes inspecting it; an unverified premise blocks only conclusions that depend on it. Other review or evidence work can continue. Once a route is adequate, act; reopen routing only for new evidence, a changed goal, or a material boundary change.

## Skills, Roles, and Execution

For specialized procedures, use [skill routing](references/skill-orchestration.md). Resolve names and locations from the current environment. Routing to a skill means reading and applying it, not just naming it in the response. Existing acceptance carries forward; an authorized implementation request proceeds into implementation after routing.

Roles describe responsibilities. The same agent can explore, implement, and verify sequentially. Independent review requires a genuinely separate reviewer; relabeling the implementer is not independence.

Delegate when authorized and a bounded independent task, parallel evidence lane, or required independent review justifies the coordination cost. Otherwise work directly. A missing optional worker or preferred tool is a reason to adapt, not to stop authorized work. Honor explicit delegation requests within host limits; serialize overlapping writers.

Read [orchestration state](references/orchestration-state.md) only for actual delegation, dependency coordination, or durable handoff. Ordinary tool use and a skill route need no task board. Read [reviewer enforcement](references/reviewer-enforcement.md) when independent review is required or high-impact implementation needs a review decision.

## Finish the Work

Routing is an intermediate decision. Continue through the authorized action and relevant validation unless the user requested only a plan or routing decision. Report results, evidence, and material limitations at the requested level of detail. A missing approval or indispensable input blocks only its dependent action; complete independent work first.

Load conditional mechanics only when they apply:

- Host-managed queues, leases, or cross-session continuation: [work-item scheduling](references/work-item-scheduling.md).
- A caller explicitly requests machine routing fields: [trace semantics](references/trace-semantics.md). Ordinary responses need no trace or YAML.
- Editing this skill or its guards: [maintenance](references/maintenance.md), including regeneration of Chinese mirrors.
