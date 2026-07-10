# Codex Adapter

Codex can use the Cognitive Control Plane skill as a routing policy. The adapter must check whether the current Codex surface exposes a multi-agent or task tool before claiming delegation.

## Mapping

- If a task/subagent tool is available, pass the portable contract as the worker prompt and require the worker to report the contract fields listed in `expected_output.must_report`.
- If no task/subagent tool is available, emit a handoff contract and stop before implementation.
- Eval wrappers that ask for `next_action` or `required_skills` are routing-only. They do not start subagents.

## Required Launch Result

```json
{
  "status": "started",
  "platform": "codex",
  "task_id": "opaque-host-task-id",
  "subagent_started": true
}
```

Without `task_id`, the result is a handoff, not real delegation.
