# Claude Code Adapter

Claude Code deployments may expose a Task tool for subagents. The Cognitive Control Plane contract maps cleanly to that tool when available.

## Mapping

- `task.role` becomes the Task agent role.
- `task.objective`, `constraints`, `ownership`, `validation`, and `stop_if` become the worker prompt.
- `required_skills` must be explicitly named in the prompt, and the worker must report whether each required skill was loaded.

## Fallback

If the Task tool is unavailable, return `handoff` with the complete contract. The main agent may continue only with direct execution when the direct-implementation exception is satisfied; otherwise it must stop before implementation.

## Required Launch Result

```json
{
  "status": "started",
  "platform": "claude-code",
  "task_id": "opaque-host-task-id",
  "subagent_started": true
}
```
