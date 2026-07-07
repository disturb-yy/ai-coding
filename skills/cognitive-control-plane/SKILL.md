---
name: cognitive-control-plane
description: "Control-plane router for complex AI collaboration. Use when process control would materially change the next action: ambiguous context, load-bearing assumptions, a concrete plan that needs challenge, multi-agent or staged handoff, or a deliverable contract for another consumer; also use when another skill needs to route work through context, epistemic, adversarial, output, or orchestration control."
---

# Cognitive Control Plane

The control plane is a thin router. It does not solve the task by default; it selects the control surface that should shape the next move, then hands execution to the right skill, worker, edit, question, or deliverable.

## Conceptual Space

```yaml
conceptual_space:
  target_region: "Complex, uncertain, high-risk, or multi-stage work where process control changes outcome quality."
  deviation_region:
    - "Simple lookup, one-line edit, command execution, or explanation with no meaningful risk."
    - "Implementation work after the route is clear; defer to coding, TDD, review, research, or writing skills."
  priority_dimensions:
    - "Preserve direction before speed."
    - "Expose assumptions before critiquing solutions."
    - "Challenge mature plans before delivery."
    - "Tighten format only at handoff or final output."
  entry_conditions:
    - "The request is vague, broad, high-risk, or missing operating context."
    - "The task involves architecture, product, prompt, skill, workflow, or code-change planning where wrong process would change the outcome."
    - "The user asks for review, risk analysis, decision support, failure analysis, or handoff."
    - "A long conversation changes phase or needs compact state."
  exit_conditions:
    - "The active control surface is named."
    - "The next action is routed to a concrete skill, worker, file edit, question, verification step, or deliverable."
    - "Known assumptions, constraints, and unresolved risks are explicit enough for that next action."
  pre_output_check:
    - "Do not run all surfaces by default."
    - "Do not use a rigid output schema during exploration."
    - "Do not keep routing once direct implementation or delivery should begin."
  sedimentation:
    - "If reusable process knowledge appears repeatedly, suggest the lowest durable home: notes, project guide, CLAUDE.md, or a narrower skill."
```

## Work Classification Gate

Classify before routing or doing substantive work.

### Tiny

Handle directly when the request is limited to clarification, status, location, recall from current context, or choosing the next process step.

Completion criterion: no worker or task skill would add value because there is no substantive task to execute.

### Small

Handle directly or with one bounded worker when every condition is true:

- goal is clear
- scope is known
- no architecture, data, auth, payment, deployment, security, or user-visible behavior risk
- no external research or current documentation is needed
- no multiple ownership boundaries are involved
- no independent review, red-team pass, persistent state, or parallelism is needed
- expected edit is local and low-risk, or the task is bounded read-only analysis

Use this compact task contract only when delegating:

```yaml
objective: ""
scope: []
edits_allowed: false
expected_output: ""
what_not_to_do: []
stop_if: "scope expands, risk appears, requirements are unclear, ownership boundaries multiply, or review becomes necessary"
```

Completion criterion: the task can finish without discovering scope, making architectural judgment, or expanding ownership.

### Large

Read [`references/orchestration-state.md`](references/orchestration-state.md) before delegation when any large signal appears:

- ambiguous or broad requirements
- unknown scope or "first find where this lives"
- multiple ownership boundaries, subsystems, agents, or staged dependencies
- architecture, data model, auth, permission, payment, deployment, security, or user-visible behavior risk
- current library documentation, web research, or external source validation is needed
- design before implementation is needed
- independent review, red-team critique, or pre-mortem is needed
- parallel lanes, background work, worktrees, staged handoff, or persistent state may help
- tests, migration, rollback, or regression strategy materially affects correctness

Uncertainty upgrades the task. Small must be proven; large needs only one signal.

Completion criterion: large or uncertain work has an orchestration state, ownership boundaries, and verification gate before implementation.

## Route

Pick the first surface that would materially change the next action:

1. **Context control**: use when the goal, current state, constraints, attempted paths, evidence, blocker, or scope boundary is unclear. Read [`references/context-control.md`](references/context-control.md) when context quality is the bottleneck.
2. **Epistemic control**: use when assumptions, evidence, confidence, causality, or decision trace determine correctness. Read [`references/epistemic-control.md`](references/epistemic-control.md) when wrong beliefs would cause bad work.
3. **Adversarial control**: use when a concrete plan, design, architecture, prompt, skill, or implementation approach needs attack. Read [`references/adversarial-control.md`](references/adversarial-control.md) when failure modes matter.
4. **Output control**: use when the work is ready for handoff, implementation, machine consumption, or final delivery. Read [`references/output-control.md`](references/output-control.md) when format and interface quality matter.

If multiple surfaces apply, start with the earliest unsatisfied one: Context -> Epistemic -> Adversarial -> Output.

When the next action needs multiple agents, parallel lanes, background work, or staged implementation, read [`references/orchestration-state.md`](references/orchestration-state.md). It is the runtime layer for scheduler-first execution, task contracts, ownership boundaries, persistent state, and conservative reflection.

When delegation depends on specialized skills, read [`references/skill-orchestration.md`](references/skill-orchestration.md) to choose the required skills and task-contract shape.

When modifying this skill, read [`references/maintenance.md`](references/maintenance.md) before editing canonical files or mirrors.

## Operating Steps

1. Apply the Work Classification Gate and classify the task as Tiny, Small, or Large.
   Completion criterion: Small is chosen only when every Small condition is true; any Large signal or uncertainty upgrades the task.
2. Select one active surface and, only if needed, read its reference.
   Completion criterion: the selected surface is sufficient for the next move; unused surfaces stay unloaded.
3. Choose execution mode: Tiny direct interaction, Small direct execution or bounded delegation, or Large orchestration-state delegation.
   Completion criterion: direct work, delegation, or orchestration is justified by task size and risk.
4. Apply the selected surface or orchestration state until its completion criterion is met.
   Completion criterion: the task has a clearer scope, stronger evidence, challenged plan, deliverable contract, or explicit task board.
5. Hand off to the concrete next action: ask a blocking question, call another skill, edit files, delegate bounded work, run verification, or produce the deliverable.
   Completion criterion: the user can see what will happen next and why.

## Do Not

- Do not turn every task into a full four-stage ceremony.
- Do not delegate small work unless delegation adds clear value.
- Do not ask for information the repository, wiki, tests, logs, or source files can answer.
- Do not critique before assumptions are explicit.
- Do not force JSON, tables, or checklists while the task is still exploratory.
