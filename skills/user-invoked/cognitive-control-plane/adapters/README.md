# Adapter Layer

The Cognitive Control Plane skill is portable policy. It decides the next action and emits a task contract, but it must not pretend that every host can start a subagent.

The adapter layer is the host-specific bridge:

1. Validate the portable contract in `contract.schema.json`.
2. Detect platform capabilities.
3. Convert the contract into a native task, subagent call, or handoff.
4. Report whether a real worker started.

Every task contract carries a stable `actor_id`. Review-phase contracts also carry `review_of_task_id`, `review_of_actor_id`, `review_iteration`, `supersedes_review_task_id`, and an immutable `review_target`. Adapters must reject a review whose actor matches the reviewed implementation actor, whose target is unversioned, or whose task is write-capable.

Real delegation is true only when the adapter returns a terminal launch record with a `task_id`. If a platform has no subagent API available to the current run, the adapter must return a handoff result and the orchestrator must stop before implementation instead of claiming that delegation happened.

## Result Semantics

- `started`: a host worker was actually created and a task id is available.
- `handoff`: the contract is valid, but no host worker was started.
- `unavailable`: the platform lacks the required capability.
- `invalid`: the contract is not safe to run.

## Platform Guides

- `opencode.md`: OpenCode task/subagent bridge.
- `codex.md`: Codex surfaces and fallback behavior.
- `claude-code.md`: Claude Code Task-tool bridge.

Use `scripts/ccp-adapter.js` for local validation and deterministic rendering.
