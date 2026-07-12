---
name: cognitive-control-plane
description: "Control-plane router for complex AI collaboration. Use when process control should decide the next move: unclear context, risky assumptions, plan critique, machine handoff, current-source evidence, codebase discovery, specialized skill routing, or multi-stage orchestration."
metadata:
  access:
    audience: model
    model_read: true
    model_write: true
    purpose: skill_runtime
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

Classify before routing or doing substantive work. Use the smallest class that is fully proven; any unresolved scope, evidence, ownership, or risk question upgrades the task.

### Tiny

Tiny is interaction with no substantive work artifact.

Use Tiny only for:

- plain term explanations
- status or next-step recall from current context
- choosing or confirming a process step
- locating information already named in the current conversation

Do not use Tiny for reading a provided code snippet, explaining behavior in a bounded artifact, editing a known file, producing a handoff, clarifying requirements, routing to a specialized skill, or deciding whether evidence is sufficient; those are at least Small.

If the user says the snippet, design, plan, library, repository, assumption set, or other artifact is provided, current, already known, or complete, treat that boundary as known for the control decision. When only the next action or routing decision is being made, do not ask the user to paste the omitted body, name, or assumption list; classify and route from the stated boundary.

Do not use Tiny for critique, stress-test, red-team, implementation-start, or accepted-contract prompts. Those prompts create a control decision even when the current adapter prompt is short.

Completion criterion: there is no artifact to inspect, no edit to make, no skill route to choose, and no worker would add value.

### Small

Small is bounded work with known inputs.

Use Small when every condition is true:

- the goal is clear
- the file, snippet, command, or artifact boundary is already known
- the user says the snippet, function, plan, or other artifact is already provided or complete, and no repository access is needed
- no repository discovery is needed
- no architecture, data, auth, payment, deployment, security, tenant, permission, or user-visible behavior risk is present
- no current documentation, web research, or external source validation is needed
- no independent review, red-team pass, persistent state, staged dependency, or real parallelism is needed
- the work is a local low-risk edit, a direct command, or bounded read-only analysis over provided material

If the user asks for multiple agents on a known one-line or single-artifact change, treat that as Small and decline the unnecessary parallelism. Do not upgrade solely because the user requested ceremony.

Do not use Small when the next action depends on `grilling`, `diagnosing-problem`, `exploring-project`, `reviewing-code`, `coding-project`, or `coding-tdd`; specialized skill routing is Large control-plane work even when the downstream task may be bounded.

Do not use Small for a stated complete plan, design, Skill design, implementation approach, diff, PR, or agent run requested for critique or review. A concrete code, security, or adversarial review is Large control-plane work, even when the artifact body is omitted by an eval adapter.

Do not use Small when an implementation contract has been accepted and the user asks to start implementation. That is Large specialized routing to `coding-project` or `coding-tdd`, not `direct_execute`.

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

Large is work where process control changes correctness.

Treat the task as Large when any signal appears:

- ambiguous or broad requirements
- unclear requirements where the user asks to make the requirement clear before implementation
- unknown scope, code location, call chain, tests, or change points, including "find where this lives" requests
- repository discovery is needed before a safe answer or edit
- specialized skill routing is needed
- multiple ownership boundaries, subsystems, agents, stages, or dependencies
- architecture, data model, auth, permission, tenant, payment, deployment, security, or user-visible behavior risk
- current official documentation, latest-version facts, breaking changes, web research, or external source validation is load-bearing
- design before implementation is needed
- independent review, code review, security review, red-team critique, pre-mortem, or plan attack is requested
- machine handoff, implementation contract, strict schema, or final delivery format is the work
- tests, migration, rollback, verification, or regression strategy materially affects correctness
- long-running or interruptible work needs persistent state

Uncertainty upgrades the task. Small must be proven; Large needs only one signal.

Completion criterion: the active bottleneck, required skills, orchestration need, ownership boundaries, and verification gates are explicit enough to choose the next action.

## Implementation Guard

For Large implementation work, the control-plane agent must not silently become the implementation worker.

Before editing source files, generated artifacts, schemas, migrations, tests, or implementation-facing docs, delegate with a task contract that names the phase, required skills, required MCP/tools, ownership boundaries, validation, and stop conditions.

Direct implementation is allowed only when **all** of these conditions are true:

- The change touches a single file.
- No schema, migration, API surface, or generated-artifact change.
- No cross-package or cross-module dependency.
- The implementation pattern exactly mirrors an existing, working example in the same codebase.
- You can state, in one sentence, what would go wrong if delegated instead of done directly.

If **any** condition fails, delegation is mandatory. Do not invent additional exceptions. When implementing directly under this exception, state which conditions are met and switch explicitly from routing to implementation before the first edit.

This guard does not block minimal routing reads, final verification commands, or tiny local edits that only update the control artifact itself.

Completion criterion: every Large implementation edit is preceded by a visible delegation contract; if the direct-implementation exception was used, every condition is met and the justification is explicit before the first edit.

## Read-only Exploration Decision

Routing to `exploring-project` does not by itself require a subagent. Before substantive repository discovery for a Large task, choose and state one of these execution modes:

- **direct minimal exploration**: the orchestrator performs a bounded, read-only trace because one entry path or a small, dependent evidence chain is sufficient; record the search scope and why a separate worker or parallel lane would not materially improve the next decision.
- **delegated exploration**: create a read-only `exploring-project` task when the evidence boundary is broad, independent areas can be checked in parallel, or a separate evidence report materially reduces implementation risk.

Do not claim a skill route and then silently replace it with unstructured direct browsing. A direct mode still loads the routed skill and reports the decision, search scope, evidence checked, and unresolved uncertainty. Parallel exploration is read-only unless ownership is explicitly assigned later.

Completion criterion: the discovery trace names `direct_minimal_exploration` or `delegated_read_only_exploration`, its scope, and why its level of delegation is proportionate.

## Reviewer Enforcement Gate

After any implementation or fix reaches terminal state, assess the delivered artifact for mandatory-review risk. Security-sensitive, cross-module, public-API, schema or migration, auth or permission, and deployment or rollback-critical changes require an independent `reviewing-code` task before acceptance. Passing tests do not waive this gate.

Preflight both the `reviewing-code` capability and an independent reviewer actor when a mandatory trigger is known or becomes known. If `reviewing-code` is unavailable, a host may use the explicit `independent_read_only_reviewer` fallback only when it starts a distinct, read-only actor under the reviewer-enforcement reference and records the unavailable skill as a deviation. If that fallback cannot start, the gate remains blocked; self-review is never a fallback.

The reviewer actor must differ from the actor that implemented or fixed the reviewed version. Pin the review to `base_sha`/`head_sha`/`diff_hash` or an equivalent immutable artifact id. Blocking findings prevent final acceptance and require a fix task; every fix changes the artifact version, invalidates the prior review, and requires another independent review. Continue until the latest version has no blocking findings or the loop is explicitly terminated. Termination is an unaccepted outcome, not a successful review. The review report separates a verification matrix from a severity-ranked findings table.

Read [`references/reviewer-enforcement.md`](references/reviewer-enforcement.md) whenever implementation risk, reviewer assignment, review findings, fixes, artifact freshness, re-review, or final acceptance is in scope.

Completion criterion: the final accepted artifact is the exact version independently reviewed with no unresolved blocking findings.

## Route

Pick the first unsatisfied surface that would materially change the next action. Do not skip an earlier bottleneck because a later surface is more visible.

1. **Context control**: use when the goal, current state, constraints, attempted paths, evidence inventory, blocker, project location, ownership boundary, phase state, or scope boundary is unclear. Read [`references/context-control.md`](references/context-control.md) when context quality is the bottleneck.
2. **Epistemic control**: use when a specific assumption, causal claim, confidence level, current-source fact, latest-version decision, or evidence standard determines correctness. Read [`references/epistemic-control.md`](references/epistemic-control.md) when wrong beliefs would cause bad work.
3. **Adversarial control**: use when a concrete plan, design, architecture, prompt, skill, implementation approach, diff, PR, or agent run needs attack. Read [`references/adversarial-control.md`](references/adversarial-control.md) when failure modes matter. Criteria come before critique.
4. **Output control**: use when discovery, context, assumptions, and review are sufficiently complete and the work is now a handoff, implementation contract, strict schema, machine-readable output, or final delivery. Read [`references/output-control.md`](references/output-control.md) when format and interface quality matter.

Surface order examples:

- "Still exploring", "still exploring the cause", "evidence is incomplete", "help determine next step", "organize current state", "organize current thinking so investigation can continue", "find where this lives", "make the requirement clear", "先把需求弄清楚", "先把目前思路整理一下，方便继续查", or "repository can answer" -> Context before anything else. Do not upgrade an unfinished investigation to Epistemic unless a specific assumption, causal claim, or evidence standard is the active bottleneck.
- High-risk implementation requests where current storage, retry semantics, rollback behavior, permissions, tenant boundaries, payment flow, or persistence model are not confirmed -> Context with orchestration before implementation. The first bottleneck is discovering current system behavior and ownership, not judging an isolated claim.
- "Current official docs", "latest major version", "breaking changes", "must be based on current official documentation", or "migration decision from release notes" -> Epistemic and read-only evidence orchestration. Set `orchestration_used: true` and `next_action: delegate_read_only`; do not merely `route_skill` to research.
- "we assume", "unverified assumption", "assumptions are not written down", "depends on an unverified assumption", "tear the proposal apart immediately", "must be based on evidence", or "root cause is..." -> Epistemic unless repository context is the only way to identify the claim. This remains Epistemic even when the user also asks for critique, attack, or review.
- "Review this branch", "code review this PR", "security review this diff", or "review since main" -> route to `reviewing-code` through skill orchestration once the review target is known.
- "Review this concrete plan", "complete Skill design", "完整的 Skill 设计方案", "red-team", "pre-mortem", "worth doing?", "duplicate?", or "too complex?" -> Adversarial only after context is clear and load-bearing assumptions are explicit or verified.
- "Discovery and design are complete", "context is clear", "assumptions are verified", "red-team review is complete", "generate implementation contract", "给实现 agent 的交付契约", "machine-readable", "accepted contract", or "final handoff" -> Output. Use `next_action: deliver` for the contract unless the user explicitly says to start implementation.
- A vague idea plus critique/stress-test request, such as "requirements and boundaries are unclear" or "the architecture idea is vague", routes to Context first. Do not mark it `none` and do not jump to Epistemic or Adversarial before the target is clear enough to inspect.
- A complete artifact plus critique/review request, such as "complete Skill design" or "完整的 Skill 设计方案", routes to Adversarial with `criteria_before_critique`. Do not downgrade to Context because the adapter omitted the artifact body.
- "Implementation contract is accepted" plus "start implementation" routes directly to `coding-project` unless the prompt explicitly asks for TDD. In that state `active_surface` is `none`, `next_action` is `route_skill`, and routing stops.

Use [`references/orchestration-state.md`](references/orchestration-state.md) when a Large next action involves delegation, read-only evidence gathering, high-risk implementation planning, specialized skill routing, multiple agents, parallel lanes, background work, staged implementation, ownership boundaries, persistent state, reviewer enforcement, or reconciliation. Orchestration state is the runtime layer; it does not replace the active surface.

Use [`references/skill-orchestration.md`](references/skill-orchestration.md) whenever the next action depends on a specialized skill. Required skill names must be explicit and exact, even if the next visible act is asking the first question or starting implementation. Use only stable skill ids from this list: `grilling`, `diagnosing-problem`, `exploring-project`, `reviewing-code`, `coding-project`, `coding-tdd`. Never write aliases such as `explore-project`. Unclear product or feature requirements involving rules, roles, interactions, or acceptance criteria require `grilling` and one question at a time; symptoms or phenomena without a framed cause require `diagnosing-problem` before code exploration; unknown repository paths or change points require `exploring-project`; code, PR, diff, branch, or security review requires `reviewing-code`; accepted implementation contracts require `coding-project` or `coding-tdd`.

When modifying this skill, read [`references/maintenance.md`](references/maintenance.md) before editing canonical files or mirrors.

## Operating Steps

1. Apply the Work Classification Gate and classify the task as Tiny, Small, or Large.
   Completion criterion: Small is chosen only when every Small condition is true; any Large signal or uncertainty upgrades the task; a user request for unnecessary process does not upgrade known Small work.
2. For Large work, choose the first unsatisfied surface and read only that reference.
   Completion criterion: exactly one active surface is named unless the next action is already implementation with no remaining control surface.
3. Decide whether orchestration state is required.
   Completion criterion: orchestration is used for Large delegation, read-only evidence gathering, high-risk implementation planning, specialized skill routing, staged or parallel work, ownership boundaries, persistent state, or reconciliation; it is not used for Small work just because the user asked for agents.
4. If specialized procedure is needed, read skill orchestration and name `required_skills`.
   Completion criterion: `grilling`, `diagnosing-problem`, `exploring-project`, `reviewing-code`, `coding-project`, or `coding-tdd` is explicit whenever skipping it would change the result.
5. Apply the selected surface or orchestration state until its completion criterion is met. After implementation or a fix, apply the Reviewer Enforcement Gate before acceptance.
   Completion criterion: the task has a clearer scope, stronger evidence, challenged plan, deliverable contract, explicit task board, or resolved ownership state.
6. Hand off to the concrete next action.
   Completion criterion: the next action is one of direct answer, direct execute, ask one blocking question, route skill, delegate read-only, delegate write, verify, or deliver; once implementation starts, stop control-plane routing and route to `coding-project` or `coding-tdd` instead of continuing to analyze.

## Trace Semantics

When a wrapper, task contract, or summary asks for a routing trace, use these meanings:

- `active_surface` is the first unsatisfied surface, not every relevant concern.
- `orchestration_used` means orchestration state materially shaped delegation, ownership, persistence, dependency ordering, high-risk implementation planning, evidence gathering, or reconciliation.
- `classification: Tiny` is only for no-artifact interaction. Bounded snippet analysis is `Small`; repository discovery, current-source evidence gathering, and any required downstream skill route are `Large`.
- If the prompt states a snippet, function, or artifact is provided and says no repository access is needed, classify from that stated boundary as `Small` and use `next_action: direct_answer`; do not ask for the artifact merely because an eval wrapper omitted its body.
- If the prompt states that a library, repository, proposal, assumption set, or other target exists but the adapter omits its concrete body or name, do not convert the route into a blocking context question when the control decision is already determined.
- `active_surface: none` is incompatible with required repository discovery, current-source evidence gathering, or adversarial review of a stated complete plan.
- Requests to find an existing code location, route, call chain, tests, or change points require `exploring-project`; classification is `Large`, active surface is `context`, and `next_action` is `route_skill`.
- Requests to review code, a branch, PR, diff, commit range, or security-sensitive implementation require `reviewing-code`; classification is `Large`, active surface is `adversarial` unless context or evidence is still missing, and `next_action` is `route_skill`.
- `ownership_conflict` means an unresolved conflict remains in the planned execution. If overlapping writers are rejected or serialized, the safe routed state has `ownership_conflict: false` and behavior `serialize_overlapping_writers`.
- High-risk implementation planning involving auth, permissions, tenants, payments, migrations, security, data model, or multi-subsystem ownership uses orchestration state before implementation or delegation; set `orchestration_used: true` even when the immediate route is project exploration.
- Current-source evidence work, including latest-version, official-documentation, or breaking-change checks, uses `active_surface: epistemic`; when the immediate next step is external research or read-only evidence gathering, set `orchestration_used: true` and `next_action: delegate_read_only`.
- When parallel write-capable workers target the same file, folder, or logical subsystem, reject parallel writes and serialize the tasks; the resolved trace has `ownership_conflict: false` and behavior `serialize_overlapping_writers`.
- Explicitly unstated, missing, or unverified load-bearing assumptions use `active_surface: epistemic`; do not route to `grilling` or adversarial review until the assumption inventory has an evidence standard or falsification path.
- `required_skills` records specialized procedure that the next action depends on, even when the response also asks the first user question or starts implementation. If any required skill is present, classification is `Large`. Use exact skill ids only. If it contains `exploring-project`, active surface is `context`. If it contains `grilling`, active surface is `context` and the route asks one question at a time. If the request is a symptom without a framed cause, use `diagnosing-problem` before `exploring-project` unless the user only asks to locate an already named code path.
- Current official documentation, latest-version, or breaking-change decisions use `active_surface: epistemic`; if the immediate next step is read-only evidence gathering, external research, `agent-reach`, or official documentation lookup, orchestration state is used. In trace terms this means `orchestration_used: true` and normally `next_action: delegate_read_only`, not `route_skill`.
- A review or attack request does not override an explicit unverified load-bearing assumption; use `active_surface: epistemic` until the assumption has evidence, uncertainty, or a falsification path.
- A stated complete plan, design, Skill design, or implementation approach requested for review uses `active_surface: adversarial` and behavior `criteria_before_critique`; do not downgrade to context merely because the artifact body is not repeated in the current adapter prompt. A stated complete code diff, PR, branch, or commit range requested for code or security review routes to `reviewing-code`.
- `next_action: deliver` is used for output contracts and handoffs; `next_action: route_skill` is used when implementation should move to a named skill.
- If an implementation contract is accepted and implementation should start, stop routing and set `next_action: route_skill` with `coding-project` or `coding-tdd` as the required skill.
- If the prompt asks to start implementation after context, assumptions, review, and the implementation contract are already accepted, classify as `Large`, set `active_surface: none`, set `required_skills: ["coding-project"]` unless test-first/TDD is explicit, set `next_action: route_skill`, and set `stopped_routing: true`.
- `event_log_in_persistent_state` means long or high-risk persistent state explicitly includes an `event_log` or event log section in addition to the task board, decision trace, review gates, validation log, and unresolved risks.
- `mandatory_review_after_high_risk_implementation` means a terminal implementation with any mandatory risk tag creates a dependent, read-only `reviewing-code` task before acceptance.
- Mandatory post-implementation review uses `next_action: delegate_read_only` because it must start a distinct actor; ordinary user-requested review may use `route_skill` when actor independence is not the active requirement.
- `direct_minimal_exploration` means the orchestrator records a bounded, read-only discovery scope and why delegation would not materially improve the next decision; it does not mean the routed exploration skill was skipped.
- `delegated_read_only_exploration` means a distinct read-only exploration task owns a broad or independently verifiable evidence lane.
- `preflight_reviewer_capability` means the availability of `reviewing-code` and an independent actor is checked before the review gate can be cleared.
- `independent_read_only_reviewer_fallback` means `reviewing-code` was unavailable, but a distinct host-launched read-only reviewer followed the reviewer-enforcement reference; it is never self-review.
- `review_report_separates_verification_and_findings` means review output has a verification matrix distinct from the severity-ranked findings table.
- `enforce_reviewer_independence` means the reviewer `actor_id` differs from the actor that implemented or fixed the pinned version; changing only the role or task id is insufficient.
- `block_final_on_blocking_findings` and `dispatch_fix_for_blocking_findings` mean reconciled blocking findings keep acceptance closed and create dependent fix work.
- `rereview_after_fix` means a fix invalidates the prior review and creates a new independent review for the new version.
- `pin_review_target_version` and `invalidate_review_on_artifact_change` mean review evidence names an immutable artifact identity and cannot clear a later version.

## Do Not

- Do not turn every task into a full four-stage ceremony.
- Do not delegate small work unless delegation adds clear value.
- Do not implement Large work as the control-plane agent unless the direct-implementation exception is explicit.
- Do not accept high-risk implementation without a valid independent review of the final artifact version.
- Do not let an implementation actor review its own work by changing role or task id.
- Do not treat a fix as closing a blocking finding until the fixed version is re-reviewed.
- Do not ask for information the repository, wiki, tests, logs, or source files can answer.
- Do not critique before assumptions are explicit.
- Do not force JSON, tables, or checklists while the task is still exploratory.
- Do not append new routing rules when an existing rule is wrong; delete or rewrite the stale rule so the skill keeps one source of truth.
