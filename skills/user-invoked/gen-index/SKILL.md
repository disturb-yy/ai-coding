---
name: gen-index
description: Generate or refresh an agent map for an existing repository by triangulating Graphify, CodeMap, targeted rg/source reads, and human-maintained docs. Use when asked for PROJECT_INDEX.md, NAVIGATION.md, CHANGE_GUIDE.md, FEATURES.md, AI-readable repository indexes, root INDEX.md, repository navigation, feature catalogs, or evidence-backed project index docs.
---

# Gen Index

## Localization Maintenance

- If this English skill is changed, update `SKILL.zh-CN.md` in the same change.
- Do not read `SKILL.zh-CN.md` as model operating instructions or task context. It is a localized human-facing copy only.

## Purpose

Generate a concise agent map for an existing repository. The map must help future agents answer three questions quickly: what the project does, where important behavior starts, and what evidence supports each claim. Triangulate: use Graphify for architecture relationships, CodeMap for code navigation, and targeted `rg`/source reads for facts.

## Conceptual Space

```yaml
conceptual_space:
  target_region: "Evidence-backed, agent-facing repository maps that explain project purpose, important behavior entry points, navigation paths, change touch points, risk areas, and freshness limits by triangulating graph, code-map, source, and doc evidence."
  deviation_region:
    - "General architecture essays, README rewrites, onboarding tutorials, or exhaustive source inventories."
    - "Graphify generation, CodeMap generation, or graph/code-map analysis work beyond reading existing artifacts or query results as evidence."
    - "Directory-local guides for ordinary folders without durable subproject, package, service, module, or workflow boundaries."
    - "Root INDEX.md generation unless the user asks for it or the repository already treats it as a convention."
  priority_dimensions:
    - "Evidence before inference: cite Graphify, CodeMap, source files, docs, or explicit unknowns for major claims."
    - "Right tool for the layer: Graphify for architecture, CodeMap for routes/modules/flows/call chains, and `rg` plus source reads for verification."
    - "Navigation before completeness: optimize for future agents finding where to start, not for documenting every file."
    - "Business capability before directory shape: describe user-visible or domain capabilities when facts support them."
    - "Preservation before rewrite during refreshes: keep useful user-added notes, but verify or mark them."
  entry_conditions:
    - "The user asks for an AI-readable repository index, agent map, project index, navigation file, change guide, feature catalog, or root INDEX.md."
    - "The task is to refresh generated `.agent/` index files after Graphify, CodeMap, source, or docs changed."
    - "A codebase needs evidence-backed orientation artifacts before future agent work."
  exit_conditions:
    - "Target and companion files are written or intentionally skipped according to the output rules."
    - "Every listed repository path exists or is marked `unknown`, `generated`, `external`, or `planned`."
    - "Major project claims cite Graphify output, CodeMap output, source, docs, or an explicit unknown/needs-confirmation marker."
    - "The final report names changed files, assumptions, evidence limits, and stale-when refresh triggers."
  pre_output_check:
    - "Check that `.agent/FEATURES.md` is about business capabilities and `.agent/NAVIGATION.md` is about where to start."
    - "Check that existing generated files were treated as drafts, not authority."
    - "Check that missing Graphify or CodeMap facts are not silently replaced with broad source-scan guesses."
    - "Check that root or directory-local indexes were not created outside the scope rules."
  sedimentation:
    - "Move durable clarified vocabulary into `.agent/GLOSSARY.md` and durable decisions into `.agent/adr/*.md` when they affect future indexing."
    - "Keep stale, unverified, or contradicted prior index material in Unknowns or remove it instead of carrying it forward as fact."
    - "Record freshness limits so later Graphify, CodeMap, route, schema, or architecture changes have an obvious refresh trigger."
```

## Evidence Ladder

Use this evidence ladder. The input pass is complete when every major project claim in the draft can cite Graphify output, CodeMap output, source, docs, or a clearly marked unknown.

| Priority | Input | Use for | Completion criterion |
|----------|-------|---------|----------------------|
| 1 | Graphify artifacts, especially `graphify-out/graph.json`, `graphify-out/GRAPH_REPORT.md`, and `graphify-out/.graphify_analysis.json` | Architecture, communities, important nodes, cross-area relationships, feature flows, and suggested navigation questions | Relevant `graphify-out/` files are read when present. |
| 2 | CodeMap MCP results or artifacts, especially `.codemap/INDEX.md`, routes, modules, flows, callgraph, and impact files | Code entry points, module boundaries, routes, call chains, change points, and affected tests | Relevant CodeMap queries or `.codemap/` files are checked when present; absence or staleness is noted. |
| 3 | Targeted `rg` and source reads | Important behavior, entry points, tests, and path verification | Listed repository paths have been checked in source or are marked `unknown`, `generated`, `external`, or `planned`. |
| 4 | Human-maintained docs | Project purpose, domain language, architecture intent, contribution conventions, durable decisions | README, architecture docs, ADRs, glossaries, and contribution docs are checked when present. |
| 5 | Existing generated index files | Incremental structure and user-added notes | Prior `.agent/*.md` or `INDEX.md` content is treated as a draft, not as fact. |
| 6 | User clarification | Business vocabulary, feature boundaries, architecture intent, or durable decisions that cannot be inferred safely | Ask only focused questions that block accurate indexing; record durable answers in the generated files. |

## Tool Roles

| Tool | Use it for | Do not use it for | Context rule |
|------|------------|-------------------|--------------|
| Graphify | Architecture relationships, god nodes, communities, cross-document links, surprising connections, and high-level capability clusters | Final proof of source behavior or route handlers | Query or read the smallest relevant graph/report section before opening broad source. |
| CodeMap | Modules, routes, flows, call graph, impact analysis, and change touch points | Design intent, business vocabulary, or claims not represented in code structure | Prefer CodeMap before broad source reads when `.codemap/` or the MCP server is available. |
| `rg` / source reads | Path existence, exact symbols, route registrations, tests, config, and final evidence | Architecture inference by scanning the whole repository | Use targeted patterns from Graphify/CodeMap leads; avoid whole-repo dumps. |

## Outputs

- Primary output: `.agent/PROJECT_INDEX.md`.
- Companion outputs: `.agent/NAVIGATION.md`, `.agent/CHANGE_GUIDE.md`, and `.agent/FEATURES.md`.
- Optional outputs: `.agent/GLOSSARY.md`, `.agent/adr/*.md`, and `.agent/ARCHITECTURE.md` or `.agent/architecture/*.md` when clarification creates durable project knowledge.
- Root `INDEX.md`: generate only when the user explicitly asks for a root-level human-facing index or the repository already uses it as a convention.

## Workflow

| Step | Action | Completion criterion |
|------|--------|----------------------|
| **Target** | State the target index path. Default to `.agent/PROJECT_INDEX.md`; use root `INDEX.md` only when explicitly requested or already established by the repository. | The target files and any companion files are named before writing. |
| **Map** | Read Graphify output first for project shape. Do not start with repository-wide source scanning. | Available relevant files under `graphify-out/` have been checked, or their absence is noted. |
| **Navigate** | Use CodeMap MCP results or `.codemap/` artifacts for code shape when available. | Candidate modules, routes, flows, call chains, tests, and uncertainties are named before opening many source files. |
| **Verify** | Use targeted `rg`, source files, and human-maintained docs to verify behavior and capture intent. | Important entry points, tests, and high-risk paths in the draft have source or doc evidence. |
| **Preserve** | For incremental updates, read existing generated index files only to preserve useful structure and user-added notes. | Preserved notes are either verified, explicitly marked as unverified, or moved to Unknowns. |
| **Clarify** | If the gap is conceptual rather than factual, ask focused questions directly. | Blocking vocabulary, feature-boundary, architecture-intent, or decision gaps are answered or recorded as `needs confirmation`. |
| **Write** | Write navigation artifacts, not exhaustive code walkthroughs. | Files use stable section names, compact bullets/tables, and the shapes below. |
| **Check** | Verify listed repository paths with `rg --files` or `test -e <path>`, and verify major flows against CodeMap/source when available. | Every listed path exists or is marked `unknown`, `generated`, `external`, or `planned`; major flows are cited or marked `not found`. |
| **Report** | Summarize changed files, assumptions, evidence limits, and refresh needs. | The final response names files changed and any sections that need refresh after Graphify output changes. |

## Index Shape

Include these sections unless the project context makes one irrelevant:

- Purpose: what the project does and who uses it.
- System Map: major areas, responsibilities, and start files.
- Core Capabilities: business capabilities, entry points, main modules, and notes.
- Architecture: style, runtime units, data stores, integrations, and cross-cutting concerns.
- Navigation: common tasks and where to start.
- Risk Areas: auth, payment, migrations, schedulers, critical flows, or other high-impact areas.
- Freshness: generated date only if known from file metadata, Graphify output, CodeMap output, or user context; Graphify/CodeMap source paths used; and stale-when conditions. If the date is not known, write `unknown`.
- Evidence: Graphify artifacts, CodeMap artifacts/results, docs, `rg` patterns, and targeted source files used.
- Unknowns: facts that need confirmation or refreshed Graphify output.

## Companion File Shapes

Write companion files with these stable fields when the facts are available. Use `unknown` or `None found in Graphify` instead of omitting a field. Do not create a companion file just to restate root overview content.

## Scope And Placement

Generate root `.agent/` files as the project map. Generate directory-local `.agent/` files only for packages, modules, apps, services, libraries, or other meaningful subtrees where local guidance will stay accurate:

- Use directory-local `.agent/NAVIGATION.md` as the default subtree guide. Cover the directory's purpose, entry points, key files, related tests, common flows, neighboring modules, and local cautions.
- Add directory-local `.agent/FEATURES.md` only when the subtree owns distinct business capabilities or user workflows.
- Add directory-local `.agent/CHANGE_GUIDE.md` only when changes in the subtree follow recurring touch points, test commands, or risk patterns.
- Add directory-local `.agent/PROJECT_INDEX.md` only when the subtree is effectively an independent subproject, such as a monorepo app, service, package, or library.

Ordinary directories do not need every guide. Prefer one accurate `NAVIGATION.md` over several stale or duplicated files. Do not copy root project overview content into directory-local guides.

### `.agent/NAVIGATION.md`

For each feature or workflow, include:

```text
Feature:
Start From:
Related Modules:
Related Routes:
Related Flows:
Tests:
Risk:
Source Evidence:
CodeMap Evidence:
```

### `.agent/CHANGE_GUIDE.md`

Group common change types by touch points:

```text
Change Type:
Touch:
Typical Flow:
Tests:
Risk:
Evidence:
```

Include common project-specific change types such as adding an API, changing persistence, adding a job, adding an event, changing auth, or modifying a critical integration.

### `.agent/FEATURES.md`

For each business capability, include:

```text
Feature:
Description:
Entry Points:
Modules:
Routes:
Flows:
Tests:
Unknowns:
Evidence:
```

### Minimal Example

```text
# Feature Navigation

Feature: User Login
Start From: src/routes/login.ts
Related Modules: auth, users
Related Routes: POST /login
Related Flows: Login request -> AuthService -> UserRepository
Tests: tests/auth/login.test.ts
Risk: Auth behavior and session creation
Source Evidence: graphify-out/graph.json; src/routes/login.ts
CodeMap Evidence: .codemap/routes/index.md

# Feature Catalog

Feature: User Login
Description: Authenticates a user and creates a session.
Entry Points: src/routes/login.ts
Modules: auth, users
Routes: POST /login
Flows: Login request -> AuthService -> UserRepository
Tests: tests/auth/login.test.ts
Unknowns: unknown
Evidence: graphify-out/graph.json; .codemap/routes/index.md; src/routes/login.ts
```

## Rules

- Prefer business capabilities over raw directory lists.
- Use Graphify artifacts for architecture structure, CodeMap for code navigation, and targeted source reads for verification.
- Use existing user-maintained docs over inferred naming when they conflict with generated facts.
- Treat existing generated index files as previous outputs, not as authoritative inputs.
- Use `unknown`, `not found in Graphify`, `not found in CodeMap`, or `needs confirmation` instead of guessing.
- Keep `.agent/FEATURES.md` focused on business capabilities and `.agent/NAVIGATION.md` focused on where to start reading or changing code.
- Keep generated indexes compact: enough structure for navigation, not a full source inventory.

## Examples

### Example 1: Fresh Agent Index

Input:

```text
Generate agent indexes for this project from Graphify output.
```

Expected behavior:

```text
Read `graphify-out/graph.json`, `graphify-out/GRAPH_REPORT.md`,
`graphify-out/.graphify_analysis.json`, and available `.codemap/` route/module/flow
artifacts when present; verify important entry files with targeted `rg` and source
reads; write `.agent/PROJECT_INDEX.md`,
`.agent/NAVIGATION.md`, `.agent/CHANGE_GUIDE.md`, and `.agent/FEATURES.md`.
```

### Example 2: Incremental Update

Input:

```text
Update the project index after Graphify output changed.
```

Expected behavior:

```text
Treat existing `.agent/PROJECT_INDEX.md`, `.agent/NAVIGATION.md`,
`.agent/CHANGE_GUIDE.md`, and `.agent/FEATURES.md` as prior drafts only.
Preserve useful user-added notes, but verify structure and paths against
Graphify output, CodeMap output, and source files before rewriting.
```

### Example 3: Conceptual Gap

Input:

```text
Generate indexes, but the module names do not explain the business features.
```

Expected behavior:

```text
Ask focused questions to clarify business vocabulary and feature boundaries.
Record stable terms in `.agent/GLOSSARY.md` or decisions in `.agent/adr/*.md`,
then generate the indexes.
```

### Example 4: No Graphify Output Available

Input:

```text
Create an agent-facing repository index, but this project has no graphify-out folder.
```

Expected behavior:

```text
Note that Graphify output is absent. Use CodeMap when present, then README,
architecture docs, manifests, and targeted source reads as evidence. Mark missing
architecture facts as `not found in Graphify` and missing route/callgraph/flow facts
as `not found in CodeMap` instead of inferring them from a broad source scan.
```

### Example 5: Graphify, CodeMap, And rg Together

Input:

```text
Generate agent indexes for this Go service; it has graphify-out and .codemap.
```

Expected behavior:

```text
Use Graphify to identify core concepts and cross-area relationships. Use CodeMap
routes, modules, flows, and call graph to name candidate entry points and change
touch points. Use targeted `rg` and source reads to verify paths and major flows
before writing `.agent/PROJECT_INDEX.md`, `.agent/NAVIGATION.md`,
`.agent/CHANGE_GUIDE.md`, and `.agent/FEATURES.md`.
```
