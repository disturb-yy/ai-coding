# Adapter Layer

中文操作手册见 [README.zh-CN.md](README.zh-CN.md)。

The Cognitive Control Plane skill produces portable task contracts. The adapter layer validates a contract, selects a supported host, and can explicitly start a **fresh** worker session.

## Lifecycle

1. `validate` checks the portable contract.
2. `render` preserves a handoff only; it never starts a worker.
3. `launch` is dry-run by default and returns the candidate native command shape.
4. `launch --execute` starts a fresh Codex or OpenCode process after workspace and transaction idempotency checks.

```bash
node scripts/ccp-adapter.js validate contract.json
node scripts/ccp-adapter.js render --platform codex contract.json
node scripts/ccp-adapter.js launch --platform codex --workspace /absolute/workspace contract.json
node scripts/ccp-adapter.js launch --platform codex --workspace /absolute/workspace --execute contract.json
```

`launch` only supports `codex` and `opencode`. It requires an existing workspace and uses direct argument arrays:

- Codex: `codex exec --json -C WORKSPACE -s SANDBOX -a APPROVAL PROMPT`
- OpenCode: `opencode run --format json --dir WORKSPACE PROMPT`

No adapter command uses resume, continue, or dangerous-bypass flags. `--executable PATH` exists for controlled integration tests or a locally installed client path; it does not alter the platform-specific argument contract.

## Result semantics

- `started` is returned only after the operating system spawned the client and includes a local `pid`.
- `dry_run`, `handoff`, `invalid`, and `unavailable` never include a `pid` and do not claim a native session ID.
- `work_id` and `run_id` identify the portable work contract. They are not native client/session IDs.

For a `transaction` work item, a non-empty `task.work_item.idempotency_key` is required before launch. The scheduler remains responsible for dependency, lease, retry, checkpoint, and terminal-state decisions.

## Scheduler bridge

Use the loop bridge after the host has read its durable work-item state and
created the candidate run contract:

```bash
node scripts/work-item-loop.js \
  --platform codex \
  --workspace /absolute/workspace \
  state.json contract.json
```

The default result is a dry run. Add `--execute` only after the Scheduler has
persisted the lease and run contract. The bridge starts a worker only for a
`dispatch` or a safe `continue` decision. A continuation must use a new run id,
a higher attempt, and `resume_checkpoint_ref`; a still-running old session is
reported as `not_dispatched`.

## Platform guides

- [codex.md](codex.md)
- [opencode.md](opencode.md)
- [claude-code.md](claude-code.md)
