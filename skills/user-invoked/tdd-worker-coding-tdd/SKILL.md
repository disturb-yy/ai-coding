---
name: tdd-worker-coding-tdd
description: Use the tdd_worker role with coding-tdd to implement one behavior through red-green-refactor evidence and final validation.
disable-model-invocation: true
---

# TDD Worker Coding TDD

Run the task as `tdd_worker` through the `coding-tdd` workflow.

## Workflow

1. Read [tdd-worker.md](../../role/tdd-worker.md) and its linked
   [handoff standard](../../role/handoff-standard.md). Treat them as the role contract.
2. Read and follow [coding-tdd](../../model-involved/coding-tdd/SKILL.md). Treat it as the
   test-first implementation workflow.
3. Keep to one observable behavior slice. Make the target test red before production changes, make
   it green with the smallest implementation, then refactor only while preserving green evidence.
4. Stop if the test cannot be made red, scope exceeds one behavior slice, validation tooling is
   unavailable, or a shared contract needs serial work. Preserve unrelated user changes.

## Handoff

Report the role contract's required fields:

- `failing_test`
- `implementation_slice`
- `artifact_version`
- `review_risk_tags`
- `green_validation`
- `refactor_validation`
- `final_entry_to_output_check`
