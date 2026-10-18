---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Memory Routing

Use the configured memory directory when configuration, user preferences, usage habits, or connection metadata could change the next action. Codex's global `AGENTS.md` block names the directory's `index.md`; the setup script can install or change that path. A skill invocation is not required for the global instruction to apply.

Read `index.md` first, then only the records whose topic and scope match the current task. A record is a lead, not current authority: check its source, last verified date, and review date. Verify a drift-prone fact against the current configuration or service before acting. If the index or a referenced record is unavailable, say so where it affects the result; do not invent a remembered setting or connection.

For a requested memory write, update, or retirement, use the available `memory-generator` skill and obey the host's storage rules. Keep each record focused, link it from `index.md`, and update the index in the same change when the host permits writing there. Search for an equivalent record before adding one. If the host requires a different write location, report that boundary; do not claim the configured index was updated. Store connection identifiers or configuration paths only; never persist credentials, tokens, or full connection strings.

A connection record may hold a service name, account alias, approved credential file path, and intended login origin. Keep the file path in that scoped record rather than the root index. The path is a locator, not authorization to inspect its contents or sign in. When a current task authorizes login, confirm the target service and permitted file access, then prefer a trusted local login tool that consumes the file directly and returns status without exposing secret values. If the tool needs a different format, inspect structure without printing values. Respect sandbox and host approval decisions; do not copy credentials into prompts, logs, command arguments, or memory.

The host work-item checkpoint is separate from this durable memory directory. Checkpoints preserve one task's continuation state; the memory index routes stable guidance across tasks.
