---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Output Control

Use when a consumer needs a handoff, implementation contract, or structured result.

Follow an explicit consumer schema at any phase, including exploration. Otherwise choose the lightest format that communicates the result; delivery does not automatically require strict JSON or a checklist.

A handoff needs the goal, relevant current state, accepted constraints, evidence/artifacts, remaining work, and next action. An implementation plan needs target changes and validation; a review needs supported findings and limitations. Include only fields the consumer will use.

For machine output, validate the actual required schema and distinguish proposed actions from completed results. For human output, make the result usable without reconstructing the conversation.

Completion criterion: the requested consumer can use the output, with missing evidence or blocked actions explicitly identified.
