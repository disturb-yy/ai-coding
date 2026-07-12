# Cognitive Control Plane evaluation wrapper

You are evaluating the installed `cognitive-control-plane` skill.

Apply the skill naturally to the case below. Do not optimize for the expected answer; you are not shown the golden label.

The user-visible response should remain natural. After the response, emit one machine-readable trace block.

The trace is an evaluation adapter, not a production output contract. Do not expose hidden chain-of-thought. Report only observable decisions and actions.

Use exactly this format:

<CCP_EVAL_TRACE>
{
  "activated": true,
  "classification": "Tiny | Small | Large",
  "active_surface": "none | context | epistemic | adversarial | output",
  "surfaces_used": [],
  "references_read": [],
  "orchestration_used": false,
  "dependency_graph_created": false,
  "persistent_state_used": false,
  "required_skills": [],
  "next_action": "direct_answer | direct_execute | ask_blocking_question | route_skill | delegate_read_only | delegate_write | verify | deliver",
  "asked_user_question": false,
  "strict_schema_during_exploration": false,
  "stopped_routing": true,
  "task_contract_complete": false,
  "ownership_conflict": false,
  "behaviors": []
}
</CCP_EVAL_TRACE>

Rules:

- `active_surface` is the first surface that materially changed the next action.
- `surfaces_used` lists only surfaces actually applied, not every surface that could apply.
- `references_read` must reflect actual reads.
- `required_skills` lists explicitly required specialized skills.
- A bounded provided artifact is not Tiny. If the case says a snippet, function, plan, or artifact is provided and no repository access is needed, treat the omitted body as an adapter omission: classify `Small`, set `next_action` to `direct_answer`, and do not ask the user to paste it.
- High-risk implementation planning involving auth, permissions, tenants, payments, migrations, security, data model, or multi-subsystem ownership uses orchestration state before implementation or delegation. Set `orchestration_used` to `true` even when the immediate route is project exploration.
- `behaviors` may include short stable identifiers such as:
  - `criteria_before_critique`
  - `serialize_overlapping_writers`
  - `schema_precedes_dependents`
  - `inspect_partial_changes_before_replacement`
  - `serial_one_question_at_a_time`
  - `reconcile_before_accept`
  - `block_final_until_validation`
  - `stop_on_required_skill_unavailable`
  - `event_log_in_persistent_state`
  - `conservative_reflection`
  - `mandatory_review_after_high_risk_implementation`
  - `enforce_reviewer_independence`
  - `block_final_on_blocking_findings`
  - `dispatch_fix_for_blocking_findings`
  - `rereview_after_fix`
  - `pin_review_target_version`
  - `invalidate_review_on_artifact_change`

Never claim a behavior that is contradicted by the visible response or runtime actions.

## Case

Case ID: {{CASE_ID}}

{{CASE_PROMPT}}
