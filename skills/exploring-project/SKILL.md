---
name: exploring-project
description: Explore and understand an existing software project before changing it. Use when asked to inspect a codebase, find where behavior lives, explain project structure, identify entry points, trace a feature, or plan a scoped code change without generating heavyweight project documentation. Prefer CodeMap MCP first when available to identify relevant chains, functions, modules, routes, flows, and risks before reading source files.
---

# Exploring Project

## Localization Maintenance

- If this English skill is changed, update `SKILL.zh-CN.md` in the same change.
- Do not read `SKILL.zh-CN.md` as model operating instructions or task context. It is a localized human-facing copy only.

## Goal

Build enough project understanding to answer or change the requested area without reading the whole repository; use CodeMap and project guides to narrow the scope, then read the necessary source files to verify behavior.

## Clarification Gate

Before deep exploration, restate the requested outcome and identify the target behavior, feature, or module.

Ask clarifying questions only when ambiguity would change where to look or what to modify. Keep questions few and concrete, usually 1-3. For each question, provide your recommended answer so the user can accept or correct it quickly. If the request is clear enough, proceed with reasonable assumptions and state them briefly.

Prefer questions about user-visible behavior, affected feature/module, reproduction steps or examples, and constraints such as framework, API, tests, or compatibility.

Do not run a broad requirements interview. Ask only the clarifying questions needed to choose the exploration path or implementation scope; otherwise resolve uncertainty through targeted project exploration.

## CodeMap First Pass

When the `codemap` MCP server is available for the current project, use it before opening source files to map likely chains, functions, modules, routes, flows, and risk areas.

1. Start with `find_change_points` using the user's requirement or your clarified requirement summary. Use a small `top_k` such as 5 unless the target is broad.
2. If the request is feature-oriented, call `get_feature_map` or `get_navigation_hints` to identify entry files and related modules.
3. If the target names a route, module, flow, or function, use the focused tools: `search_route`, `search_module`, `related_modules`, `search_flow`, `call_graph`, or `impact_analysis`.
4. If CodeMap is unavailable, stale, unsupported for the project, or too vague, fall back to lightweight repository discovery with `rg`, `rg --files`, manifests, routes, imports, and tests.

| Target from user request | CodeMap tool | Verify by reading |
| --- | --- | --- |
| Broad requirement or change request | `find_change_points` | Candidate files, functions, routes, and tests |
| Feature or user workflow | `get_feature_map`, `get_navigation_hints` | Entry files, related modules, and tests |
| HTTP/API route | `search_route` | Handler, service/use case, and route tests |
| Module or package | `search_module`, `related_modules` | Module implementation, imports, and dependents |
| Flow or call chain | `search_flow`, `call_graph` | Caller/callee implementation and integration tests |
| Function impact | `impact_analysis` | Callers, behavior contracts, and affected tests |

Treat CodeMap output as a navigation aid, not proof. Verify important claims by reading the referenced implementation or tests before explaining behavior or editing code.

## Core Workflow

1. Start from existing project guides if present: `README`, `.agent/PROJECT_INDEX.md`, `.agent/NAVIGATION.md`, docs, architecture notes, or contribution guides.
2. Apply the clarification gate if the request is ambiguous enough to affect exploration or implementation.
3. Run the CodeMap first pass when available to identify likely chains, functions, and files before reading source code.
4. Identify the project shape with lightweight commands: file tree, manifests, package files, routes, tests, and main entry points.
5. Name the target feature, module, or behavior in business terms before opening many files.
6. Follow references from entry point to implementation: route/command/UI -> service/use case -> model/storage/integration -> tests.
7. Open only files that are relevant to the target area. Avoid repository-wide scanning unless the first pass fails.
8. Summarize findings as actionable context: relevant files, ownership boundaries, data/control flow, risks, and next edit locations.

## Checkpoints

- After the guide and CodeMap first pass, name the candidate files, functions, routes, and uncertainties before opening many files.
- Before explaining behavior or editing code, verify the relevant CodeMap leads by reading the referenced implementation or tests.
- Before finishing exploration, report the verified flow and the next edit location. If verification fails, say what was not found and continue with targeted discovery.

## Investigation Rules

- Prefer `rg` and `rg --files` for discovery.
- Prefer CodeMap over raw file reading for the first pass when it is available and relevant.
- Prefer existing indexes and generated docs over rediscovering everything.
- Use dependency edges, imports, route registrations, test names, and configuration files to narrow the search.
- Treat names from code as evidence, not proof; verify behavior in the implementation or tests.
- Keep uncertainty explicit. Say `unknown` or `not found` instead of guessing.
- Stop exploring once the next action is clear enough.

## Output Shape

When reporting exploration results, keep it concise:

```text
Project shape:
- ...

Relevant files:
- path: why it matters

CodeMap leads:
- tool result -> verified/not verified

Flow:
- entry -> implementation -> persistence/integration

Risks:
- ...

Next change location:
- ...
```

For code changes, use the findings to keep the modification scope minimal and avoid unrelated refactors.

## Examples

### Clear change request

Input:

```text
Fix the bug where order cancellation does not refund inventory.
```

Expected behavior:

```text
Restate the outcome. Use `find_change_points` for the requirement, then use focused CodeMap tools for any named routes, flows, modules, or functions. Read only the candidate implementation and tests needed to verify the cancellation -> inventory flow before proposing or making edits.
```

### Ambiguous request

Input:

```text
Improve the login flow.
```

Expected behavior:

```text
Ask 1-3 clarifying questions because "improve" could change the exploration path. For each question, provide a recommended answer, such as "Recommended: focus on the user-visible error after failed password login." After the user accepts or corrects the recommendation, run CodeMap first and verify with source/tests.
```

### CodeMap unavailable or weak

Input:

```text
Find where invoices are generated in this project.
```

Expected behavior:

```text
Try CodeMap first. If unavailable, stale, unsupported, or too vague, read `README`, `.agent/PROJECT_INDEX.md`, project docs, manifests, routes, and tests, then use `rg`/`rg --files` to locate invoice entry points. Verify the final answer by reading the relevant implementation or tests.
```
