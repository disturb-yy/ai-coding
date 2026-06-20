 ---
name: ooda-index
description: Generate OODA-compliant index files from project documentation, source trees, API definitions, business rules, or infrastructure modules. Use when the user asks Codex to create, update, standardize, or validate an index file such as index.md, API index, module index, infrastructure index, business capability index, or any documentation file that must summarize files, modules, functions, contracts, dependencies, and usage rules in an OODA-style observe-orient-decide-act loop.
---

# OODA Index

Use this skill to generate index documents that are accurate, traceable, and useful for later agents. The index must describe what exists, how it is organized, what contracts are exposed, and what rules must be followed when using or changing it.

## Output Contract

For substantive tasks, structure visible work in this order:

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

Keep each section concise. Tool calls happen outside the visible XML labels, but summaries and decisions must follow this order.

## Workflow

### `<observe>` Source Reconnaissance

Identify the requested index target and scan the relevant project context before writing:

- Existing index files, README files, architecture docs, coding standards, route/API docs, schemas, and module-level comments.
- Directory structure and naming patterns around the target scope.
- Source files that define public modules, exported functions, classes, types, configs, commands, endpoints, events, or data contracts.

State missing expected materials when relevant. Do not invent modules, signatures, rules, or relationships that are not visible in source or documentation.

### `<orient>` Index Standard Selection

Choose the index shape that best matches the target:

- **Project index**: summarize layers, major directories, entrypoints, build/test commands, ownership boundaries, and cross-layer dependency rules.
- **Module index**: summarize module purpose, public files, exports, internal dependencies, extension points, and usage cautions.
- **API index**: summarize callable APIs, signatures, parameters, return values, side effects, errors, idempotency, auth/config requirements, and examples only when source supports them.
- **Infrastructure index**: summarize reusable infrastructure capabilities such as logging, HTTP, cache, time, config, auth, persistence, retries, serialization, queues, and storage. Prefer documented infrastructure over hand-written utilities.
- **Business capability index**: summarize business chains, entities, workflows, invariants, policy rules, external integrations, and validation requirements.

If the user provides an existing local index convention, preserve it unless it conflicts with accuracy or the requested OODA standard.

### `<decide>` Index Plan

List the target file, source materials to use, selected index type, and section outline. Keep the plan scoped to the requested index. Include validation commands or static checks that will be used after writing.

### `<draft>` Index Draft

Draft the complete index content before editing files. The draft should be specific enough to review:

- Scope and purpose.
- Directory or module map.
- Public contracts with precise names and signatures when available.
- Inputs, outputs, return conventions, errors, side effects, and configuration.
- Dependency rules and "use this instead of hand-writing that" guidance.
- Update rules for keeping the index current.

Do not write files during this phase.

### `<precheck>` Static Consistency Gate

Review the draft before editing:

- Every named file, module, function, class, type, endpoint, command, and config exists or is explicitly marked as missing/unknown.
- Signatures and return conventions match source or documentation.
- The index does not mix public contracts with private implementation details unless that is the requested scope.
- The document is scannable for future coding agents: stable headings, compact bullets, and direct cross-references.
- No unsupported business rules, API guarantees, or dependency claims were invented.

If issues are found, state the correction that will be applied in `<act>`.

### `<act>` Tool-Based File Write

Create or update the target index file with file editing tools. Do not satisfy the request by only printing Markdown unless the user explicitly asks for text-only output.

When updating an existing index, preserve accurate existing content and remove or mark stale content only after checking the source.

### `<evaluate>` Validation And Self-Correction

Validate after writing. Prefer real checks when available:

- Reread the written index.
- Run documentation linting, markdown linting, tests, type checks, or project validation commands when appropriate.
- Cross-check sampled entries against source files and existing docs.

If validation fails or a real inconsistency is found, report the issue, fix the index, and validate again until it passes or an external blocker is reached.

## Index File Norms

Use clear Markdown unless the user requests another format. Prefer this section order when no existing convention exists:

1. Title and scope.
2. Quick map of directories, modules, or capabilities.
3. Public contracts or documented capabilities.
4. Dependency and usage rules.
5. Validation, testing, or update guidance.
6. Source references used to generate the index.

Use relative paths when writing inside a repository. Include exact signatures only when they can be verified. Mark uncertain items as `Unknown` or omit them; never fill gaps with guesses.

## Guardrails

Do not skip source scanning when relevant files are available. Do not invent API signatures, business rules, directory purposes, or dependency constraints. Do not edit files during `<draft>`. Do not leave a generated index unvalidated. Do not create extra documentation files unless they are part of the requested index output.
