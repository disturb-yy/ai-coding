---
name: ooda-coding
description: Enforce an OODA-E coding workflow for software implementation, bug fixes, refactors, tests, and business-code changes. Use when the user asks Codex to write or modify code and the work benefits from reading project documentation, coding standards, existing indexes such as index.md or infrastructure indexes, drafting before editing, using tools for file changes, and validating with tests, linters, type checks, builds, or static review in a self-correcting loop.
---

# OODA Coding

Use this skill for coding tasks that need disciplined context gathering, architecture alignment, draft-first implementation, tool-based edits, and validation. The operating principle is: observe the project, orient to its standards and indexes, decide a small plan, draft before editing, precheck like a reviewer, act with tools, and evaluate until the result is correct.

## Required Context Scan

Before designing or editing code, scan the project context that exists for the target scope:

1. User requirement and current repository state.
2. Existing index files such as `index.md`, module indexes, API indexes, or infrastructure indexes.
3. Project README, architecture docs, package manifests, route/API docs, schemas, test conventions, and coding standards.
4. Business-specific docs when present, especially `业务层/README.md`, `业务层/编码规范.md`, and `基础设施层/index.md`.
5. Source files and tests directly related to the requested change.

If an expected file is missing, state that it is missing and continue with the closest available source. Do not invent APIs, business rules, coding standards, or module boundaries.

## Output Contract

For substantive coding work, structure visible reasoning and result summaries in this exact order:

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

Keep user-facing text concise. Tool calls happen outside the visible labels, but the work summaries must follow this order.

## OODA-E Workflow

### `<observe>` Requirement And Project Reconnaissance

Summarize the requested code change in one sentence. Identify the affected module, layer, feature, API, or business chain.

Read the relevant docs, indexes, source files, tests, and configuration. When `index.md` or an infrastructure index exists, use it to locate public contracts and reusable capabilities before reading implementation details.

For business repositories, extract hidden rules from business documentation and cite the relevant existing infrastructure APIs, function signatures, input requirements, return conventions, and usage constraints.

### `<orient>` Standards, Contracts, And Dependency Alignment

Align the intended implementation with the discovered project rules:

- Naming, formatting, error handling, logging, dependency injection, module boundaries, and layering.
- Existing public contracts, types, schemas, route patterns, persistence patterns, and test style.
- Infrastructure capabilities already documented in indexes.

Ask internally whether the change is about to hand-write a common capability that should come from existing code or infrastructure, such as network requests, caching, time handling, logging, configuration, retries, serialization, authentication, persistence, validation, queueing, or storage. If yes, switch to the documented existing capability.

Prefer narrow changes that preserve current architecture. Avoid unrelated refactors unless required to complete the user request safely.

### `<decide>` Implementation Plan

List the concrete implementation steps:

- Target files.
- Imports, functions, classes, types, routes, schemas, or tests to change.
- Core logic and error/edge-case handling.
- Validation commands or static checks to run.

Keep the plan scoped to the requested behavior and discovered architecture.

### `<draft>` Brain Sandbox Draft

Before modifying files, draft the intended code or patch content in text. This is the static sandbox phase; do not call file-writing tools in this phase.

The draft must be complete enough to review:

- Imports and dependencies.
- Function signatures and type flow.
- Control flow, async behavior, and error handling.
- Infrastructure/API calls and parameter compatibility.
- Test cases or validation strategy.

For very small changes, a compact patch sketch is enough, but still draft before editing.

### `<precheck>` Static Quality Gate

Switch to a strict reviewer stance and inspect the draft before editing:

- Dependency health: all referenced functions, variables, types, files, and infrastructure APIs exist or are planned additions.
- Type and contract health: parameters, return values, schemas, and error conventions match source and indexes.
- Syntax health: no obvious syntax, scoping, async, import, module-format, or framework lifecycle errors.
- Standards health: naming, logging, layering, error handling, and tests match project conventions.
- Blast-radius health: no unrelated behavior changes, hidden global state, fragile mocks, or avoidable compatibility risk.

If issues are found, name them and state how they will be corrected in `<act>`.

### `<act>` Tool-Based File Write

Use available file editing tools to apply the corrected change to real files. Do not satisfy this phase by only printing Markdown code blocks unless the user explicitly asked for text-only output.

Respect the existing working tree. Do not revert user changes. Keep edits limited to the target scope.

### `<evaluate>` Dynamic Validation And Self-Correction

After editing, validate the result. Prefer real commands when available:

- Unit tests, integration tests, targeted test files, or snapshot tests.
- Type checks, linting, formatting checks, compilation, build, or project-specific validation.
- For small or command-blocked changes, reread modified files and perform a final static review.

If validation fails or a real defect is found, report the failed check and relevant error, then continue the self-correction loop without waiting for the user. Return to the appropriate earlier phase, fix the issue, and rerun validation until checks pass or a genuine external blocker is reached.

## Index Awareness

When an `index.md` or OODA-style index exists, treat it as the navigation and contract layer:

- Use it to find public APIs and intended usage before opening private implementation files.
- Keep implementation aligned with documented contracts.
- If the code change intentionally updates a public contract, update the relevant index or documentation when that is within the user request or necessary for consistency.
- If the index appears stale, verify against source and mention the mismatch before relying on it.

## Guardrails

Do not skip context scanning when relevant files exist. Do not invent missing business rules, API signatures, or coding standards. Do not hand-write existing infrastructure capabilities. Do not edit files during `<draft>`. Do not finish after failed validation unless the blocker is external and clearly stated. Do not broaden the task with unrelated refactors or documentation churn.
