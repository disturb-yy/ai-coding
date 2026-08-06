---
name: implementation-worker-coding-project
description: Use the implementation_worker role with coding-project to complete a confirmed, bounded change in an existing repository.
disable-model-invocation: true
---

# Implementation Worker Coding Project

Run the confirmed task as `implementation_worker` through the `coding-project` workflow.

## Workflow

1. Read [implementation-worker.md](../../role/implementation-worker.md) and its linked
   [handoff standard](../../role/handoff-standard.md). Treat them as the role contract.
2. Read and follow [coding-project](../../model-involved/coding-project/SKILL.md). Treat it as the
   implementation workflow and its referenced language guidance as the coding standard.
3. Confirm the task is an approved, non-test-first repository change. Keep edits within the agreed
   ownership and preserve unrelated user changes.
4. Complete the narrow implementation and run the relevant project checks. Stop and report if a
   required tool is unavailable, validation is externally blocked, the task needs out-of-scope files,
   or a material risk is outside the contract.

## Handoff

Report the role contract's required fields:

- `changed_files`
- `artifact_version`
- `review_risk_tags`
- `validation_commands`
- `validation_results`
- `residual_risks`
