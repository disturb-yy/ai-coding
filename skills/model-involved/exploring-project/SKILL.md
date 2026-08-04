---
name: exploring-project
description: Explore a codebase before answering or changing it. Use when asked to inspect project structure, locate behavior, trace a feature, route, module, flow, or function, explain an implementation, or plan a scoped code change without heavyweight documentation. Use Graphify for architecture/cross-area context, CodeMap for code navigation, and targeted rg/source reads for verification.
---

# Exploring Project

## Localization Maintenance

- If this English skill is changed, update `SKILL.zh-CN.md` in the same change.
- Do not read `SKILL.zh-CN.md` as model operating instructions or task context. It is a localized human-facing copy only.

## Goal

Run a narrowing loop: frame the request, map the region, navigate to candidate code, verify the relevant path in source or tests, then stop when the answer or next edit location is clear. Do not read the whole repository unless targeted discovery fails.

## Role Contract

Act as the local [`codebase_explorer`](role/codebase-explorer.md) role copy. Read its linked
[handoff standard](../../role/handoff-standard.md) before exploring. The role owns the context
phase, read-only boundary, stopping conditions, and final fields; this skill owns the exploration
method. Finish with `target`, `relevant_files`, `flow`, `leads_checked`, `risks`, and
`next_change_location`. Do not edit files.

## Conceptual Space

```yaml
conceptual_space:
  target_region: "Scoped codebase exploration that turns an unclear request into a verified flow, explanation, or next edit location by using graph, code-map, and source evidence at the right layer."
  deviation_region:
    - "Repository-wide audits, exhaustive indexing, or creating architecture documentation and generated project maps unless the user explicitly asks for them."
    - "Code editing before the target path and local ownership boundaries have been verified."
    - "Speculation from names, docs, graph edges, CodeMap output, or search hits without source or test confirmation."
  priority_dimensions:
    - "Verified evidence over breadth."
    - "Smallest useful discovery path over whole-repository reading."
    - "Right layer first: Graphify for architecture, CodeMap for code structure, rg/source for facts."
    - "Actionable stopping point over comprehensive narration."
  entry_conditions:
    - "The user asks to inspect project structure, locate behavior, trace a feature, route, module, flow, or function, or explain an implementation."
    - "A planned code change needs discovery before the safe edit location is known."
    - "Another skill needs repository context before it can act safely."
  exit_conditions:
    - "The target and assumptions are explicit."
    - "Relevant graph, guide, CodeMap, or search leads have been checked or explicitly marked unavailable."
    - "Candidate files, functions, routes, tests, and uncertainties have been named."
    - "The relevant behavior is verified in source or tests, or a targeted search has shown it is not present."
    - "The user has a verified flow, answer, or next edit location with risks and nearby tests."
  pre_output_check:
    - "Every important claim is backed by a file, test, Graphify result, CodeMap result, targeted search, or stated as unknown/not found."
    - "The answer names what was checked and avoids implying broader coverage than was performed."
    - "Exploration stops once the next action is clear enough."
  sedimentation:
    - "Preserve generated guides, Graphify output, and CodeMap output as navigation leads, not proof."
    - "Do not create or expand project documentation from exploration unless the user requested documentation."
    - "If guides, maps, or tests are stale or missing, report that as a finding instead of silently broadening scope."
```

## Workflow

1. Frame the target. Restate the requested outcome and name the target behavior, feature, route, module, flow, or function. Ask 1-3 concrete questions only when ambiguity would change where to look or what to modify; include your recommended answer with each question. Complete this step when the target and assumptions are explicit.
2. Map the region. Use generated project guides and Graphify when available to identify the relevant system area, domain concepts, cross-area links, and likely boundaries. Complete this step when you can name the likely region or say Graphify/guides are unavailable or unhelpful.
3. Navigate code. Use CodeMap MCP results or `.codemap/` artifacts when available to identify candidate files, functions, routes, flows, tests, and remaining uncertainties. Complete this step before opening many source files.
4. Verify the path. Use targeted `rg`, source files, and tests to confirm the lead: route/command/UI -> service/use case -> model/storage/integration -> tests. Complete this step only when the relevant behavior is verified, or when targeted search proves the expected path is not present.
5. Stop with actionable context. For exploration, report the verified flow and evidence. For code changes, name the next edit location, nearby tests, ownership boundaries, and risks. Complete this step when the user can act without another discovery pass.

## Navigation Leads

Use the smallest useful guide before broad discovery. Guides can exist at the repository root or inside a package, module, app, service, or other meaningful subtree:

- Root `.agent/PROJECT_INDEX.md`, `README`, docs, architecture notes, and contribution guides for overall project context.
- Root or directory-local `.agent/FEATURES.md` for business capability and user workflow.
- Root or directory-local `.agent/NAVIGATION.md` for the directory's purpose, start files, related modules, routes, and tests.
- Root or directory-local `.agent/CHANGE_GUIDE.md` for change touch points, typical flow, tests, and risks.

Treat root guides as the project map. Treat directory-local guides as introductions to that directory and its subtree. Prefer the closest relevant `.agent/` guide once candidate directories are known; walk up to parent guides or root guides only when local guidance is missing or too narrow. Do not expect every directory to have guides.

Read only the sections relevant to the requested feature, module, route, flow, directory, or change type. Treat generated guides as navigation aids, not proof; verify key claims against Graphify, CodeMap, source files, or tests according to the tool roles below.

## Tool Roles

Use the tightest useful tool for the question layer:

| Layer | Prefer | Good for | Must verify with |
| --- | --- | --- | --- |
| Architecture, domain concepts, cross-module or code/document relationships | Graphify query/path/explain or `graphify-out/GRAPH_REPORT.md` | Core abstractions, communities, surprising connections, broad system region | CodeMap, source, tests, or docs before claiming implementation behavior |
| Code structure, routes, modules, flows, call chains, impact | CodeMap MCP or `.codemap/` artifacts | Candidate files, handlers, services, tests, change points | Source and tests |
| Exact symbol, error text, config, path existence, final proof | Targeted `rg`, `rg --files`, and source reads | Concrete lines and behavior | The smallest relevant implementation/test set |

When Graphify output exists and the question is broad, architectural, cross-area, or concept-oriented, use it before CodeMap or source reading. When the question is about a route, module, flow, function, or change point, use CodeMap before broad source reading. When the question names an exact symbol, error, path, or config key, go straight to targeted `rg` and source reads.

When the `codemap` MCP server is available for the current project, use it before opening source files:

| Target | CodeMap tool | Verify by reading |
| --- | --- | --- |
| Broad requirement or change request | `find_change_points` with a small `top_k`, usually 5 | Candidate files, functions, routes, and tests |
| Feature or user workflow | `get_feature_map`, `get_navigation_hints` | Entry files, related modules, and tests |
| HTTP/API route | `search_route` | Handler, service/use case, and route tests |
| Module or package | `search_module`, `related_modules` | Module implementation, imports, and dependents |
| Flow or call chain | `search_flow`, `call_graph` | Caller/callee implementation and integration tests |
| Function impact | `impact_analysis` | Callers, behavior contracts, and affected tests |

Convert `.agent` guide fields into focused CodeMap calls when possible: modules -> `search_module` or `related_modules`; routes -> `search_route`; flows -> `search_flow` or `call_graph`; broad changes -> `find_change_points`.

If CodeMap is unavailable, stale, unsupported, or too vague, fall back to lightweight discovery with `rg`, `rg --files`, manifests, route registrations, imports, tests, and main entry points. If Graphify is unavailable, stale, or too vague, do not replace it with broad source scanning; use guides, CodeMap, and targeted source reads for the scoped target.

## Evidence Rules

- Prefer `rg` and `rg --files` for exact discovery and verification.
- Prefer Graphify over raw source reading for broad architecture or cross-area discovery when available and relevant.
- Prefer CodeMap over raw source reading for route, module, flow, function, and impact discovery when available and relevant.
- Use dependency edges, imports, route registrations, test names, and configuration files to narrow the search.
- Treat names from code, generated docs, Graphify, and CodeMap as leads, not conclusions.
- Before explaining behavior or editing code, verify important claims by reading implementation or tests.
- Keep uncertainty explicit: say `unknown` or `not found` instead of guessing.
- Stop exploring once the answer or next action is clear enough.

## Checkpoints

- After the guide/Graphify/CodeMap pass, name candidate files, functions, routes, tests, and uncertainties before opening many files.
- Before explaining behavior or editing code, verify the relevant leads in source or tests.
- Before finishing, report the verified flow and next edit location. If verification fails, say what was not found and what graph, CodeMap, or targeted search you used.

## Output Shape

Use only the sections that help the current task:

```text
Project shape:
- ...

Relevant files:
- path: why it matters

Leads checked:
- guide/Graphify/CodeMap/search result -> verified/not verified

Flow:
- entry -> implementation -> persistence/integration -> tests

Risks:
- ...

Next change location:
- ...
```

For code changes, use the findings to keep the modification scope minimal and avoid unrelated refactors.

## Examples

### Clear Change Request

Input:

```text
Fix the bug where order cancellation does not refund inventory.
```

Expected behavior:

```text
Restate the refund outcome. Use guides and `find_change_points`, then focused CodeMap tools for any named route, flow, module, or function. Read only the candidate implementation and tests needed to verify the cancellation -> inventory flow before editing.
```

### Ambiguous Request

Input:

```text
Improve the login flow.
```

Expected behavior:

```text
Ask 1-3 questions because "improve" could change the exploration path. Include recommended answers, such as "Recommended: focus on the user-visible error after failed password login." After the user accepts or corrects the target, run the navigation and verification loop.
```

### CodeMap Unavailable Or Weak

Input:

```text
Find where invoices are generated in this project.
```

Expected behavior:

```text
Try CodeMap first. If unavailable or weak, read the smallest relevant guides and use `rg`/`rg --files` against invoice terms, routes, tests, and entry points. Verify the final answer by reading the relevant implementation or tests.
```

### Broad Architecture Or Cross-Area Question

Input:

```text
Explain how market sync connects API handlers, providers, repositories, and docs.
```

Expected behavior:

```text
Use Graphify first if `graphify-out/` exists to identify core concepts and cross-area links. Use CodeMap routes/modules/flows to locate candidate handlers, services, providers, repositories, and tests. Use targeted `rg` and source reads to verify the actual path before answering.
```
