---
name: problem-framer-diagnosing-problem
description: Use the problem_framer role with diagnosing-problem to frame an uncertain symptom, assumptions, evidence standard, and next handoff.
disable-model-invocation: true
---

# Problem Framer Diagnosing Problem

Run the task as `problem_framer` through the `diagnosing-problem` workflow.

## Workflow

1. Read [problem-framer.md](../../role/problem-framer.md) and its linked
   [handoff standard](../../role/handoff-standard.md). Treat them as the role contract.
2. Read and follow [diagnosing-problem](../../model-involved/diagnosing-problem/SKILL.md). Treat it
   as the diagnosis workflow.
3. Select the most useful interpretation, distinguish facts from assumptions, and set an evidence
   standard. Reject competing interpretations only with available evidence.
4. Stop when the issue becomes requirements clarification, required evidence is inaccessible, or the
   next investigation route is clear. Do not explore or implement the codebase.

## Handoff

Report the role contract's required fields:

- `framed_problem`
- `selected_interpretation`
- `rejected_interpretations`
- `assumptions`
- `evidence_standard`
- `handoff`
