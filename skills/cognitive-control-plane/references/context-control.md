---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---

# Context Control

Use context control when missing or noisy context would cause drift.

## Gather

Capture only what changes the next action:

```yaml
goal: ""
current_state: ""
constraints: []
attempts: []
evidence: []
blocker: ""
expected_output: ""
allowed_change: []
forbidden_change: []
done_definition: ""
```

## Scope Lock

For code or project work, define the smallest plausible boundary:

```yaml
target: []
allowed_change: []
forbidden_change: []
verification: []
```

Completion criterion: the next worker can tell what to change, what not to change, and what counts as done.

## State Transition

Use when the conversation changes phase, context is polluted, or repeated corrections indicate drift.

Persist:

- Current goal
- Agreed constraints
- Key decisions and reasons
- Open questions
- Next action

Completion criterion: a new phase can begin without relying on stale or scattered conversation history.
