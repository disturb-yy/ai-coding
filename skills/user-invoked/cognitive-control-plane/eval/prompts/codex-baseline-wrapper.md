# Behavioral evaluation adapter

Handle the user case exactly as you normally would with installed skills available.

This evaluation tests whether process control should activate at all. Do not invoke or apply a skill merely because this is an evaluation. Choose the normal next action for the case.

For safety, this run evaluates the **control decision only**:

- Do not modify files.
- Do not perform downstream implementation.
- Do not invoke a routed worker skill after deciding that it should receive the task.
- Do not continue into external research after deciding that research is the next action.
- You may read an installed skill and only the references actually needed to decide or apply the immediate control step.
- For a Tiny direct-answer case, give the direct answer.
- A bounded provided artifact is not Tiny. If the case says a snippet, function, plan, or artifact is provided and no repository access is needed, treat the omitted body as an adapter omission: classify `Small`, set `next_action` to `direct_answer`, and do not ask the user to paste it.
- High-risk implementation planning involving auth, permissions, tenants, payments, migrations, security, data model, or multi-subsystem ownership uses orchestration state before implementation or delegation. Set `trace.orchestration_used` to `true` even when the immediate route is project exploration.
- For a Small direct-execute case, identify the direct action but do not edit.
- For a Large case, apply only enough control to determine the concrete handoff, verification step, or deliverable, then stop.
- If context, assumptions, review, and the implementation contract are already accepted and the case says to start implementation, stop control-plane routing and route to the implementation skill. Set `trace.active_surface` to `none`, `trace.next_action` to `route_skill`, and include `coding-project` in `trace.required_skills` unless the case explicitly asks for TDD.
- If a case says a complete plan, design, Skill design, diff, PR, or implementation approach exists and asks for critique or review, treat the artifact boundary as known for the control decision. Use adversarial control with `criteria_before_critique`; do not ask for the omitted body.
- If a critique or stress-test request says the idea, requirements, architecture, or boundaries are vague or unclear, context is the first active surface. Do not skip to assumption audit or adversarial critique before the target is clear.
- If a case says the team is still exploring a cause, evidence is incomplete, or asks to organize current thinking for continued investigation, use `trace.active_surface: context`; do not convert general incompleteness into Epistemic.
- If a high-risk implementation case says current storage, retry semantics, rollback behavior, permissions, tenant boundaries, payment flow, or persistence model are not confirmed, use `trace.active_surface: context` with orchestration before implementation.
- If a case says current official docs, latest major version, migration notes, release notes, or breaking changes are load-bearing, use `trace.active_surface: epistemic`, set `trace.orchestration_used: true`, and prefer `trace.next_action: delegate_read_only`.
- If a case says a concrete plan depends on an unverified assumption, or assumptions are not written down, use `trace.active_surface: epistemic` before adversarial review. Do not critique before the assumption inventory has an evidence standard or falsification path.
- If a case says context is clear, assumptions are verified, review/red-team is complete, and asks for an implementation or handoff contract, use `trace.active_surface: output`, `trace.next_action: deliver`, and `trace.stopped_routing: true`.
- Use exact stable skill ids in `trace.required_skills`: `exploring-project`, not `explore-project`. Product or feature requirements with unclear rules, roles, or interactions require `grilling`, even if the visible response asks the first question. A symptom or phenomenon without a framed cause requires `diagnosing-problem` before repository exploration.

Do not expose hidden chain-of-thought. Report only observable decisions and the concise user-facing response.

Return only one JSON object matching the provided output schema.

Field rules:

- `case_id`: exactly `{{CASE_ID}}`
- `evidence_source`: exactly `self_report`
- `response`: the concise response you would show the user in this routing-only run
- `trace.activated`: whether the cognitive control-plane skill actually shaped the next action
- `trace.classification`: Tiny, Small, or Large
- `trace.active_surface`: the first surface that materially changed the next action; `none` if no surface was applied
- `trace.surfaces_used`: only surfaces actually applied
- `trace.references_read`: actual skill reference paths read, not references that merely could apply
- `trace.orchestration_used`: whether orchestration state shaped the next action
- `trace.required_skills`: specialized downstream skills explicitly required by the route
- `trace.next_action`: the immediate next action after control
- `trace.stopped_routing`: true when no additional control surface should run before the next action
- `trace.behaviors`: only stable observable behavior identifiers that truly occurred

Useful behavior identifiers include:

- `criteria_before_critique`
- `serialize_overlapping_writers`
- `schema_precedes_dependents`
- `inspect_partial_changes_before_replacement`
- `serial_one_question_at_a_time`
- `reconcile_before_accept`
- `require_worker_must_report`
- `check_required_capabilities_before_acceptance`
- `block_final_until_validation`
- `stop_on_required_skill_unavailable`
- `event_log_in_persistent_state`
- `conservative_reflection`

## User case

{{CASE_PROMPT}}
