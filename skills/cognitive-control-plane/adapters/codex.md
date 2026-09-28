# Codex Adapter

The Codex adapter turns a validated portable contract into a **fresh** Codex CLI session. It never resumes or continues an existing session.

## Launch

Preview the exact native command without starting a client:

```bash
node scripts/ccp-adapter.js launch \
  --platform codex \
  --workspace /absolute/workspace \
  contract.json
```

Start only after an explicit opt-in:

```bash
node scripts/ccp-adapter.js launch \
  --platform codex \
  --workspace /absolute/workspace \
  --sandbox workspace-write \
  --approval on-request \
  --execute \
  contract.json
```

The native invocation is `codex exec --json -C WORKSPACE -s SANDBOX -a APPROVAL PROMPT`. The adapter uses an argument array, never a shell command, and does not use resume, continue, or dangerous-bypass flags.

## Result semantics

- `dry_run`: a launch was rendered but no Codex process was started.
- `started`: a fresh process started; `pid` is present. It is a local process identifier, **not** a Codex native session ID.
- `invalid` or `unavailable`: no process was started.

`render` is handoff-only and always reports `subagent_started: false`, even if a `--task-id` value is supplied. A transaction work item requires `task.work_item.idempotency_key` before it may launch.
