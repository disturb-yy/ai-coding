---
name: project-understanding
description: >-
  Build a cognitive model of any project using CodeMap as the fact layer.
  Use when asked to "understand this project", "explain the codebase",
  "document the architecture", "generate a project overview", or "create
  an AI-navigable project index". Produces PROJECT_INDEX.md, CHANGE_GUIDE.md,
  NAVIGATION.md, and FEATURES.md. Avoids repository-wide scanning by always
  following PROJECT_INDEX → CodeMap → Source Code. Requires CodeMap MCP facts
  or tools; do not use when CodeMap is unavailable.
---

# Project Understanding

Build a cognitive model of the project using CodeMap. The goal is to help new engineers understand the system quickly and AI Agents locate relevant modules — not explain source code file-by-file.

## Core Principles

### Principle 1: Index → CodeMap → Source

Do not scan the entire repository. Always follow:

```
PROJECT_INDEX → CodeMap → Source Code
```

### Principle 2: CodeMap is the Fact Layer

CodeMap provides objective facts: modules, dependencies, routes, flows, call graphs, impact analysis. Never invent information that contradicts CodeMap.

### Principle 3: Focus on Business Capabilities

Prefer "User Login", "Create Order", "Refund Order" over "AuthController", "UserService", "OrderRepository". Think in Features first.

---

## Feature Detection Rules

```
Feature = Route + Flow + Module
```

**Example — POST /login:**

AuthController → AuthService → UserRepository = **Feature: User Login**

**Example — POST /orders:**

OrderService → PaymentService → InventoryService = **Feature: Create Order**

When naming a feature, priority: Business Meaning → Route Meaning → Flow Meaning → Module Name. Avoid technical names.

See [references/feature-detection.md](references/feature-detection.md) for full rules.

---

## Architecture Detection

Identify the architecture style from module dependency patterns:

- Layered Architecture: controller → service → repository
- Clean Architecture: adapter → usecase → domain
- DDD: application → domain → infrastructure
- Hexagonal, Event-Driven, CQRS

If unclear, set Architecture Style = Unknown. Do not guess.

See [references/architecture-patterns.md](references/architecture-patterns.md).

---

## Risk Analysis

| Risk Level | Pattern |
|---|---|
| **High** | auth, payment, transaction, migration, scheduler |
| **Medium** | cache, message queue, event bus |
| **Low** | dto, util, view |

Explain why a module is risky. See [references/risk-analysis.md](references/risk-analysis.md).

---

## Change Strategy

**New API:** controller → service → repository

**New Database Field:** entity → migration → repository → service

**New Event:** producer → consumer → handler

**New Background Job:** scheduler → worker → service

Always identify: affected modules, entry points, dependency impact.

See [references/change-strategy.md](references/change-strategy.md) and [references/modification-strategy.md](references/modification-strategy.md).

---

## Decision Table

| Scenario | Action |
|---|---|
| CodeMap unavailable | Stop and report that CodeMap facts are required |
| Route missing, flow/module available | Infer feature from flow + module; write `None found in CodeMap` for routes |
| Flow missing, route/module available | Infer feature from route + module; write `None found in CodeMap` for flows |
| Architecture unclear | Set Architecture Style = `Unknown`; do not guess |
| Risk unclear | Mark risk as `Low` and explain the uncertainty |

---

## Few-Shot Examples

### Example 1 — Normal API Feature

Input facts:

```yaml
route: POST /orders
flow: create_order_flow
modules: [order, payment, inventory]
```

Expected feature:

```yaml
name: Create Order
routes: [POST /orders]
flows: [create_order_flow]
modules: [order, payment, inventory]
entry_points: [POST /orders]
```

### Example 2 — Missing Route

Input facts:

```yaml
route: null
flow: billing_invoice_flow
modules: [billing, payment]
```

Expected feature:

```yaml
name: Billing Invoice
routes: None found in CodeMap
flows: [billing_invoice_flow]
modules: [billing, payment]
entry_points: Unknown
```

### Example 3 — Missing CodeMap

Input facts:

```yaml
codemap: unavailable
```

Expected action:

```text
Stop. Report that project-understanding requires CodeMap MCP facts/tools and must not fall back to repository-wide scanning.
```

---

## Workflow

Execute all steps in order. Never skip.

### Step 1 — Collect Facts

Use CodeMap MCP tools. Required: `get_project_info`, `list_modules`, `related_modules`, `search_route`, `search_flow`. Optional: `call_graph`, `impact_analysis`.

Default path for OpenCode and other MCP-native agents: call the CodeMap MCP tools directly and save equivalent JSON facts under `.agent/context/`.

Optional CLI helper: if the local environment provides `mcp-call`, `scripts/collect-codemap.sh` can collect the same facts into `.agent/context/`. Treat this script as a convenience wrapper, not as a required part of the workflow.

If CodeMap is unavailable, stop and report that this skill requires CodeMap facts; do not replace CodeMap with repository-wide scanning.

Build: Module List, Dependency Graph, Route List, Flow List. Do not generate documentation yet.

Optional script: `scripts/collect-codemap.sh` automates fact collection into `.agent/context/` only when `mcp-call` is installed.

Optional phase prompt: use [prompts/01-discovery.md](prompts/01-discovery.md) when collecting project facts manually through CodeMap tools.

Checkpoint: confirm `.agent/context/project.json`, `.agent/context/modules.json`, `.agent/context/routes.json`, and `.agent/context/flows.json` exist and are non-empty before continuing.

### Step 2 — Discover Features

Infer business features. For each identify: name, modules, routes, flows. Output an internal Feature Map.

```
Feature: User Login
Modules: auth, user
Routes: POST /login
Flows: login_flow
```

See examples in [examples/](examples/) for ecommerce, gateway, and SaaS projects.

Optional phase prompt: use [prompts/02-feature-extraction.md](prompts/02-feature-extraction.md) when feature grouping becomes noisy or ambiguous.

### Step 3 — Build Cognitive Model

Construct: Project, Modules, Features, Entrypoints, Risk Areas, Navigation Hints, Modification Patterns. Do not write markdown yet.

Script: `scripts/build-context.sh` merges collected facts into `.agent/context/context.json`.

Optional phase prompt: use [prompts/03-cognitive-model.md](prompts/03-cognitive-model.md) when converting raw facts into the internal model.

Checkpoint: run `jq 'has("project") and has("modules") and has("routes") and has("flows")' .agent/context/context.json` and continue only if it returns `true`.

### Step 4 — Generate Documents

Generate using templates in [templates/](templates/):

- **PROJECT_INDEX.md** — Project Overview, Architecture, Core Features, Module Map, Request/Data Flow, Entry Points, Risk Areas, AI Navigation Hints
- **CHANGE_GUIDE.md** — Add API, Add Database Field, Add Background Job, Add Event, Add Module (with recommended modification locations)
- **NAVIGATION.md** — Per feature: Start Files, Related Modules, Related Routes, Related Flows, Risk Notes
- **FEATURES.md** — Per feature: Name, Description, Modules, Routes, Flows, Entry Points

Keep output concise. Prefer cognitive understanding over implementation details.

Fill template placeholders with grounded facts from `.agent/context/context.json` and the internal feature map. Use `Unknown` for missing facts instead of guessing. When a placeholder represents a list, render concise Markdown bullets or a compact table. When a feature has no route or flow, write `None found in CodeMap`.

Optional phase prompts:

- [prompts/04-project-index.md](prompts/04-project-index.md) for `PROJECT_INDEX.md`
- [prompts/05-change-guide.md](prompts/05-change-guide.md) for `CHANGE_GUIDE.md`
- [prompts/06-navigation.md](prompts/06-navigation.md) for `NAVIGATION.md`

### Step 5 — Verify

Run `scripts/verify-context.sh` to confirm module/flow/route counts in `.agent/context/context.json`.

---

## Agent Behavior (When Using Generated Docs)

**Step 1:** Read PROJECT_INDEX.md → identify Feature, Module, Entry Point.

**Step 2:** Read NAVIGATION.md → locate Start Files, Related Modules.

**Step 3:** Use CodeMap: `search_module`, `related_modules`, `search_route`, `search_flow`, `call_graph`, `impact_analysis`.

**Step 4:** Open source code — only files relevant to the target feature. Avoid repository-wide scanning.

**Step 5:** Perform modification — keep changes minimal, avoid unrelated refactoring.

See [references/project-understanding-workflow.md](references/project-understanding-workflow.md) and [references/navigation-rules.md](references/navigation-rules.md).

---

## Output Files

| File | Purpose |
|---|---|
| `.agent/PROJECT_INDEX.md` | Cognitive entry: overview, architecture, features, module map, flows |
| `.agent/CHANGE_GUIDE.md` | Modification patterns: where to touch for each change type |
| `.agent/NAVIGATION.md` | Feature→file lookup: start files, related modules per feature |
| `.agent/FEATURES.md` | Feature catalog: name, modules, routes, flows per feature |

---

## Success Criteria

The generated documentation must let a new engineer or AI Agent answer — without reading the entire repository:

- What does this project do?
- What are the main features?
- Which module owns a feature?
- Where should a new requirement be implemented?
- Which modules are high risk?
- Where should code investigation start?
