---
name: gen-index
description: Generate or update agent-facing project indexes from .codemap files, CodeMap facts, targeted source reads, and human-maintained docs. Use when asked to create PROJECT_INDEX.md, NAVIGATION.md, CHANGE_GUIDE.md, AI-readable repository indexes, optional root INDEX.md, or clarification-backed project index docs.
---

# Gen Index

## Localization Maintenance

- If this English skill is changed, update `SKILL.zh-CN.md` in the same change.
- Do not read `SKILL.zh-CN.md` as model operating instructions or task context. It is a localized human-facing copy only.

## Purpose

Generate a concise agent-facing project index from CodeMap-generated facts, `.codemap/` files, targeted source reads, and human-maintained project documentation. Use grilling/domain-modeling only to clarify missing business vocabulary, feature boundaries, architecture intent, or durable decisions before writing the index.

## Inputs

Prefer inputs in this order:

1. CodeMap-generated files or MCP facts, including files under `.codemap/` when present. Treat `.codemap/architecture/`, `.codemap/callgraph/`, `.codemap/flows/`, `.codemap/modules/`, and `.codemap/routes/` as primary generated inputs for architecture, call chains, feature flows, module maps, and route entry points.
2. Targeted source reads for important behavior, entry points, tests, and file-path verification.
3. Human-maintained project docs: `README.md`, architecture docs, ADRs, glossary files, contribution guides, and other docs that capture project intent.
4. Existing generated index files, such as `.agent/PROJECT_INDEX.md`, `.agent/NAVIGATION.md`, `.agent/CHANGE_GUIDE.md`, or `INDEX.md`, only as prior drafts for incremental updates. Do not treat them as the fact layer.
5. User clarification only when project meaning cannot be inferred safely.

## Outputs

- Primary output: `.agent/PROJECT_INDEX.md`.
- Companion outputs: `.agent/NAVIGATION.md` and `.agent/CHANGE_GUIDE.md`.
- Optional outputs: `.agent/GLOSSARY.md`, `.agent/adr/*.md`, and `.agent/ARCHITECTURE.md` or `.agent/architecture/*.md` when clarification creates durable project knowledge.
- Root `INDEX.md`: generate only when the user explicitly asks for a root-level human-facing index or the repository already uses it as a convention.

## Workflow

1. State the target index path. Default to `.agent/PROJECT_INDEX.md`; use root `INDEX.md` only when explicitly requested or already established by the repository.
2. Read CodeMap outputs first, including relevant files under `.codemap/architecture/`, `.codemap/callgraph/`, `.codemap/flows/`, `.codemap/modules/`, and `.codemap/routes/` when present. Do not start with repository-wide source scanning.
3. Read targeted source files and human-maintained docs to verify behavior and capture intent.
4. If existing generated index files are present, read them only to preserve useful structure and user-added notes during incremental updates.
5. If the missing information is conceptual rather than factual, run a `/grilling` session using `/domain-modeling` to clarify it. If that path is unavailable, ask focused clarification questions directly and record durable answers in `.agent/GLOSSARY.md`, `.agent/adr/*.md`, or `.agent/ARCHITECTURE.md`.
6. Write the index as a navigation artifact, not an exhaustive code walkthrough.
7. Verify that listed paths exist or are explicitly marked `unknown`, `generated`, `external`, or `planned`. Use `rg --files` or `test -e <path>` for repository paths before finishing.
8. Report changed files, assumptions, and any sections that need refresh after CodeMap is regenerated.

## Index Shape

Include these sections unless the project context makes one irrelevant:

- Purpose: what the project does and who uses it.
- System Map: major areas, responsibilities, and start files.
- Core Capabilities: business capabilities, entry points, main modules, and notes.
- Architecture: style, runtime units, data stores, integrations, and cross-cutting concerns.
- Navigation: common tasks and where to start.
- Risk Areas: auth, payment, migrations, schedulers, critical flows, or other high-impact areas.
- Evidence: CodeMap files/facts, docs, and targeted source files used.
- Unknowns: facts that need confirmation or CodeMap regeneration.

## Rules

- Prefer business capabilities over raw directory lists.
- Use CodeMap facts, including `.codemap/architecture/`, `.codemap/callgraph/`, `.codemap/flows/`, `.codemap/modules/`, and `.codemap/routes/`, for structure and targeted source reads for verification.
- Use existing user-maintained docs over inferred naming when they conflict with generated facts.
- Treat existing generated index files as previous outputs, not as authoritative inputs.
- Use `unknown`, `not found in CodeMap`, or `needs confirmation` instead of guessing.
- Do not call another skill implicitly except for the explicit `/grilling` and `/domain-modeling` clarification path described above.

## Examples

### Example 1: Fresh Agent Index

Input:

```text
Generate agent indexes for this project from .codemap.
```

Expected behavior:

```text
Read `.codemap/architecture/`, `.codemap/modules/`, `.codemap/routes/`,
`.codemap/flows/`, and `.codemap/callgraph/`; verify important entry files
with targeted source reads; write `.agent/PROJECT_INDEX.md`,
`.agent/NAVIGATION.md`, and `.agent/CHANGE_GUIDE.md`.
```

### Example 2: Incremental Update

Input:

```text
Update the project index after regenerating CodeMap.
```

Expected behavior:

```text
Treat existing `.agent/PROJECT_INDEX.md`, `.agent/NAVIGATION.md`, and
`.agent/CHANGE_GUIDE.md` as prior drafts only. Preserve useful user-added
notes, but verify structure and paths against `.codemap/` and source files
before rewriting.
```

### Example 3: Conceptual Gap

Input:

```text
Generate indexes, but the module names do not explain the business features.
```

Expected behavior:

```text
Use `/grilling` with `/domain-modeling` to clarify business vocabulary and
feature boundaries. If unavailable, ask direct focused questions. Record stable
terms in `.agent/GLOSSARY.md` or decisions in `.agent/adr/*.md`, then generate
the indexes.
```
