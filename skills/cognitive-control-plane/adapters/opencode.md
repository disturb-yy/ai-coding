# OpenCode Adapter

The OpenCode adapter turns a validated portable contract into a **fresh** OpenCode CLI session. It never resumes or continues an existing session.

## Launch

Preview without starting a client:

```bash
node scripts/ccp-adapter.js launch \
  --platform opencode \
  --workspace /absolute/workspace \
  contract.json
```

Start only after an explicit opt-in:

```bash
node scripts/ccp-adapter.js launch \
  --platform opencode \
  --workspace /absolute/workspace \
  --execute \
  contract.json
```

The native invocation is `opencode run --format json --dir WORKSPACE PROMPT`. The adapter passes arguments directly and does not use resume, continue, or dangerous-bypass flags.

## Result semantics

- `dry_run`: a launch was rendered but no OpenCode process was started.
- `started`: a fresh process started; `pid` is present. It is a local process identifier, **not** an OpenCode native session ID.
- `invalid` or `unavailable`: no process was started.

`render` stays a handoff and never represents a started session. A transaction work item requires `task.work_item.idempotency_key` before it may launch.
