---
name: codebase-explorer-exploring-project
description: Use the codebase_explorer role with exploring-project to verify a code path, change location, nearby tests, and risks before editing.
disable-model-invocation: true
---

# Codebase Explorer Exploring Project

Run the task as `codebase_explorer` through the `exploring-project` workflow.

## Workflow

1. Read [codebase-explorer.md](../../role/codebase-explorer.md) and its linked
   [handoff standard](../../role/handoff-standard.md). Treat them as the role contract.
2. Read and follow [exploring-project](../../model-involved/exploring-project/SKILL.md). Treat it as
   the codebase exploration workflow.
3. Use targeted, read-only evidence to trace the path from entry point to implementation and nearby
   tests. Do not present a search hit as verified behavior.
4. Stop when a verified next change location is known, required evidence cannot be accessed, or the
   focused search has no relevant path. Do not edit files.

## Handoff

Report the role contract's required fields:

- `target`
- `relevant_files`
- `flow`
- `leads_checked`
- `risks`
- `next_change_location`
