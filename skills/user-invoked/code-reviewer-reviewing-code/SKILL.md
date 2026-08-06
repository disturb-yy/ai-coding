---
name: code-reviewer-reviewing-code
description: Use the code_reviewer role with reviewing-code to assess an immutable code change and return an evidence-backed gate decision.
disable-model-invocation: true
---

# Code Reviewer Reviewing Code

Run the task as `code_reviewer` through the `reviewing-code` workflow.

## Workflow

1. Read [code-reviewer.md](../../role/code-reviewer.md), its linked
   [handoff standard](../../role/handoff-standard.md), and
   [reviewer enforcement](../cognitive-control-plane/references/reviewer-enforcement.md). Treat them
   as the role contract.
2. Read and follow [reviewing-code](../../model-involved/reviewing-code/SKILL.md). Treat it as the
   review workflow.
3. Pin an immutable review target and baseline. Select or explicitly skip syntax, functionality,
   standards, and security lanes; separate tool failures from code findings.
4. Stop if the target is unavailable or mutable, required evidence is inaccessible, the reviewer is
   the implementation actor, or the task becomes implementation. Do not edit files.

## Handoff

Report the role contract's required fields:

- `verification_matrix`
- `findings_table`
- `findings`
- `blocking_findings`
- `non_blocking_findings`
- `no_findings`
- `skipped_or_blocked`
- `residual_risk`
- `review_summary`
- `review_target`
- `gate_decision`
