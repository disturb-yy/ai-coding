---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Context Control

Use when missing or noisy context would change the next action.

Recover the goal, current state, relevant constraints, evidence already checked, and the remaining blocker from available conversation and artifacts. Retrieve repository facts directly; ask the user only for a material choice or inaccessible input. State a reasonable assumption when a minor ambiguity does not block safe progress.

When a configured memory index could contain relevant preferences, habits, configuration, or connection metadata, follow [memory routing](memory-routing.md) before selecting an approach. Verify time-sensitive facts against current sources.

Define the next bounded action: target, permitted changes, and a checkable success condition. This can be one sentence; a fixed inventory or YAML form is unnecessary. Carry forward accepted scope and existing authorization.

When phase changes or repeated corrections reveal drift, summarize the current goal, agreed decisions, unfinished work, and next action. Persist only if a handoff or interruption requires it.

Completion criterion: the next action can proceed without guessing a consequential requirement or redoing completed discovery.
