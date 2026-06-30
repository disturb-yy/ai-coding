---
name: ooda-e-coding
description: Enforce an OODA-E business coding workflow for implementation tasks that require reading business documentation, coding standards, infrastructure API indexes, drafting before editing, using tools to modify files, and validating changes in a self-correcting loop. Use when the user asks Codex to develop or modify business code under a project that contains business-layer README, coding standards, and infrastructure index documentation.
---

# OODA-E Coding

Use this skill for business development tasks that must follow a strict architecture-first coding loop. The operating principle is: standards over freedom, infrastructure over hand-written utilities, and validation as the final gate.

## Required Context Scan

Before designing or editing code for a concrete business requirement, read or use existing context for these project assets when they exist:

1. Business blueprint: `业务层/README.md`.
2. Coding standard: `业务层/编码规范.md`.
3. Infrastructure API index: `基础设施层/index.md`.

If a file does not exist, state that it is missing and continue with the closest available project documentation. Do not invent API signatures, business rules, or coding standards.

## Output Contract

For each business coding task, structure the work in this exact OODA-E order. Use the XML-style labels below for substantive replies that perform implementation work:

```xml
<observe>
</observe>
<orient>
</orient>
<decide>
</decide>
<draft>
</draft>
<precheck>
</precheck>
<act>
</act>
<evaluate>
</evaluate>
```

Keep user-facing content concise inside each label. Tool calls happen outside the visible text, but the reasoning and result summaries must follow this order.

## OODA-E Workflow

### `<observe>` Business Context And Infrastructure Reconnaissance

Summarize the requirement in one sentence and identify which business chain or module it belongs to based on the business blueprint. Extract hidden business rules that affect the implementation.

Query `基础设施层/index.md` and precisely cite the infrastructure APIs, function signatures, input requirements, and return conventions needed for the task. Prefer existing infrastructure APIs over custom code.

### `<orient>` Standards And Dependency Alignment

Check the planned logic against `业务层/编码规范.md`, including naming, exception handling, logging, layering, data access, and module boundaries.

Ask explicitly whether the implementation is about to hand-write any common capability that should come from infrastructure instead, such as network requests, caching, time handling, logging, configuration, retries, serialization, authentication, or persistence. If so, stop the hand-written approach and switch to the documented infrastructure API.

### `<decide>` Implementation Plan

List the concrete implementation steps, including target files, required imports, core business logic, and validation commands. Keep the plan scoped to the user request and the documented architecture.

### `<draft>` Brain Sandbox Draft

Before modifying files, draft the intended code or patch content in text. This is the static sandbox phase; do not call file-writing tools in this phase.

The draft should be complete enough to review imports, types, control flow, infrastructure calls, and error handling before any real edit is made.

### `<precheck>` Static Quality Gate

Switch to a strict reviewer stance and inspect the draft before editing:

- Dependency health: all functions, variables, types, and infrastructure APIs are imported or referenced correctly.
- Type health: parameters passed to infrastructure APIs match the signatures in `基础设施层/index.md`.
- Syntax health: no obvious syntax, scoping, async, or module-format errors.
- Standards health: naming, errors, logging, and layering comply with `业务层/编码规范.md`.

If issues are found, name them clearly and state that they will be corrected in `<act>`.

### `<act>` Tool-Based File Write

Use the available file editing tools to modify the real target files. Do not satisfy this phase by only printing Markdown code blocks. Apply the corrected final code to the project.

Keep the visible text short, usually one sentence describing which files are being changed.

### `<evaluate>` Dynamic Validation And Self-Correction

After writing files, validate the result. Prefer real terminal commands when available, such as compilation, linting, type checks, unit tests, or project-specific checks. If terminal execution is unavailable, reread the modified files and perform a final static review.

If validation fails or the final review finds a real defect, report the failed check and the relevant error, then continue the self-correction loop without waiting for the user. Return to the appropriate earlier phase, usually `<orient>` or `<act>`, fix the issue, and rerun validation until the checks pass or a genuine external blocker is reached.

## Guardrails

Do not skip the documentation scan when the required files exist. Do not invent missing business rules or API signatures. Do not hand-write infrastructure capabilities already documented in `基础设施层/index.md`. Do not edit files during `<draft>`. Do not finish after a failed validation unless the blocker is external and clearly stated.
