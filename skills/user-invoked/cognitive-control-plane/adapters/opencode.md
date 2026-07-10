# OpenCode Adapter

OpenCode can host the Cognitive Control Plane skill and may expose task or subagent features outside this repository. The portable skill must only emit a contract; the OpenCode adapter owns the native launch.

## Mapping

- `delegate_read_only` maps to a read-only worker when OpenCode exposes one.
- `delegate_write` maps to a bounded write-capable worker only after ownership paths are checked.
- `route_skill` maps to loading the named skill in the active session when no separate worker is required.

## Required Launch Result

A real delegation must return:

```json
{
  "status": "started",
  "platform": "opencode",
  "task_id": "opaque-host-task-id",
  "subagent_started": true
}
```

If the OpenCode host does not expose a task API to the adapter, return `handoff` and preserve the contract for the user or outer runtime. Do not mark `orchestration_used` as completed worker execution solely because the model wrote "delegate".

## Current Repository Scope

The bundled OpenCode plugin in `scripts/opencode-plugin.mjs` is a guard plugin. It blocks protected reads and verifies mirrors. It does not start subagents.
