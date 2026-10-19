---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Trace Semantics

Read only when a caller requests routing telemetry. Use the caller's schema; these field names preserve the bundled adapter vocabulary. A trace records decisions and observed actions, not hidden reasoning or invented execution.

- `classification`: the caller's task-size label, if its schema requires one. The conversational skill defines no size classes; use the caller's definitions or report the field as unknown.
- `active_surface`: the bottleneck shaping the immediate action, or `none` when execution is ready. `surfaces_used` lists only applied surfaces.
- `required_skills`: exact currently resolved skill ids needed by the next action, not a fixed allowlist. `references_read` contains actual reads.
- `next_action`: `direct_answer`, `direct_execute`, `ask_blocking_question`, `route_skill`, `delegate_read_only`, `delegate_write`, `verify`, or `deliver`. Skill routing is not delegation; research can be direct.
- `orchestration_used`: coordination, ownership, dependent stages, or durable continuation actually shaped the work. Reading a reference or using tools alone does not qualify.
- `stopped_routing`: a sufficient route has been chosen and control returns to concrete work; this does not mean the user task is finished.
- `ownership_conflict`: an unresolved conflict; after overlapping writers are serialized, it is false.
- Worker start, lease acquisition, checkpoint persistence, and review completion require runtime evidence. A proposed contract is not an executed action. In a routing-only simulation describe proposed behavior separately and leave unperformed action fields false.

Only host-managed work uses the work-item/run fields described in [work-item scheduling](work-item-scheduling.md). Report unknown telemetry as unknown where the schema permits it; never claim a configured threshold was enforced merely because it is documented.
