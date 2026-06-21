# CodeMap Cognitive Layer Design

## Background

Current CodeMap already provides:

* Module Analysis
* Dependency Analysis
* Route Analysis
* Flow Analysis
* Call Graph Analysis
* Impact Analysis

Current MCP Tools:

```text
get_project_info
list_modules
search_module
related_modules
search_route
search_flow
call_graph
impact_analysis
```

These capabilities belong to the **Fact Layer**.

They answer:

* Where is the code?
* Which module owns this function?
* Who depends on this module?
* Who calls this function?
* What is the impact of modifying this function?

However, AI agents working on requirements usually need a higher-level understanding:

* Which feature does this requirement belong to?
* Which module should be modified?
* Where should investigation begin?
* Which files are likely entry points?

This belongs to the **Cognitive Layer**.

---

# Goal

Introduce two new MCP capabilities:

```text
get_feature_map
get_navigation_hints
```

These APIs expose project cognition directly from CodeMap and provide structured data for:

* PROJECT_INDEX.md generation
* CHANGE_GUIDE.md generation
* NAVIGATION.md generation
* AI requirement analysis
* AI code modification planning

---

# Architecture

```text
Source Code
      ↓
CodeMap Analysis
      ↓
Module Graph
Dependency Graph
Route Graph
Flow Graph
Call Graph
      ↓
Feature Map
Navigation Hints
      ↓
PROJECT_INDEX
      ↓
AI Agent
```

---

# MCP Tool: get_feature_map

## Purpose

Return the project's business feature map.

This API answers:

```text
What business capabilities exist in the project?
Which modules implement them?
Which routes expose them?
Which flows support them?
```

---

## Example Request

```json
{}
```

---

## Example Response

```json
{
  "features": [
    {
      "name": "User Login",
      "description": "Authenticate users and create sessions",

      "modules": [
        "auth",
        "user"
      ],

      "routes": [
        "POST /login"
      ],

      "flows": [
        "login_flow"
      ]
    },

    {
      "name": "Create Order",

      "modules": [
        "order",
        "inventory",
        "payment"
      ],

      "routes": [
        "POST /orders"
      ],

      "flows": [
        "create_order_flow"
      ]
    },

    {
      "name": "Refund Order",

      "modules": [
        "refund",
        "payment"
      ],

      "routes": [
        "POST /refunds"
      ],

      "flows": [
        "refund_flow"
      ]
    }
  ]
}
```

---

## Usage

User Requirement:

```text
Add order cancellation functionality
```

Agent:

```text
1. get_feature_map

2. Find related features:
   - Create Order
   - Refund Order

3. Identify related modules:
   - order
   - payment
   - refund
```

The agent can narrow the investigation scope before opening source code.

---

# MCP Tool: get_navigation_hints

## Purpose

Return project navigation guidance.

This API answers:

```text
Where should investigation start?
Which files are likely modification points?
Which routes and flows are related?
```

---

## Example Request

```json
{}
```

---

## Example Response

```json
{
  "features": [
    {
      "feature": "User Login",

      "start_files": [
        "internal/auth/controller.go",
        "internal/auth/service.go"
      ],

      "routes": [
        "POST /login"
      ],

      "related_modules": [
        "auth",
        "user"
      ],

      "related_flows": [
        "login_flow"
      ],

      "risk": [
        "token",
        "session"
      ]
    },

    {
      "feature": "Create Order",

      "start_files": [
        "internal/order/controller.go",
        "internal/order/service.go"
      ],

      "routes": [
        "POST /orders"
      ],

      "related_modules": [
        "order",
        "inventory",
        "payment"
      ],

      "related_flows": [
        "create_order_flow"
      ],

      "risk": [
        "payment",
        "inventory_consistency"
      ]
    }
  ]
}
```

---

## Usage

User Requirement:

```text
Add SMS login
```

Agent Workflow:

```text
1. get_navigation_hints

2. Locate:

Feature:
User Login

Start Files:
auth/controller.go
auth/service.go

3. search_module(auth)

4. call_graph(auth)

5. Open source code
```

This significantly reduces repository-wide scanning.

---

# Feature Extraction Strategy

Feature Map should be generated automatically.

No LLM is required.

---

## Input Sources

CodeMap already provides:

```text
Route
Flow
Module
```

---

## Rule

```text
Feature
=
Route
+
Flow
+
Module
```

---

## Example

Route:

```text
POST /login
```

Flow:

```text
login_flow
```

Modules:

```text
auth
user
```

Generated Feature:

```json
{
  "name": "User Login",
  "modules": [
    "auth",
    "user"
  ],
  "routes": [
    "POST /login"
  ],
  "flows": [
    "login_flow"
  ]
}
```

---

# Navigation Hint Extraction Strategy

Navigation Hints should also be generated automatically.

No LLM is required.

---

## Input Sources

```text
Route
Flow
CallGraph
Module
```

---

## Rule

Identify:

* Entry Route
* Entry Function
* Entry File
* Related Modules
* Related Flows
* Risk Areas

---

## Example

Route:

```text
POST /login
```

Entry Function:

```text
AuthController.Login
```

File:

```text
internal/auth/controller.go
```

Generated Hint:

```json
{
  "feature": "User Login",

  "start_files": [
    "internal/auth/controller.go"
  ],

  "related_modules": [
    "auth",
    "user"
  ]
}
```

---

# Database Design

## feature

```sql
CREATE TABLE feature (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT
);
```

---

## feature_module

```sql
CREATE TABLE feature_module (
    feature_id INTEGER NOT NULL,
    module_name TEXT NOT NULL
);
```

---

## feature_route

```sql
CREATE TABLE feature_route (
    feature_id INTEGER NOT NULL,
    route TEXT NOT NULL
);
```

---

## feature_flow

```sql
CREATE TABLE feature_flow (
    feature_id INTEGER NOT NULL,
    flow_name TEXT NOT NULL
);
```

---

## navigation_hint

```sql
CREATE TABLE navigation_hint (
    feature_name TEXT NOT NULL,
    start_file TEXT NOT NULL,
    route TEXT,
    flow TEXT,
    risk TEXT
);
```

---

# Relationship with PROJECT_INDEX

PROJECT_INDEX remains a human-readable cognitive document.

```text
PROJECT_INDEX
=
Presentation Layer
```

CodeMap remains:

```text
Fact Layer
```

New APIs become:

```text
Feature Layer
```

---

# Final Cognitive Architecture

```text
PROJECT_INDEX
      ↓
Feature Layer
(get_feature_map)

      ↓
Navigation Layer
(get_navigation_hints)

      ↓
Fact Layer
(CodeMap)

      ↓
Source Code
```

---

# Expected AI Workflow

```text
Requirement

      ↓

get_feature_map

      ↓

Identify Feature

      ↓

get_navigation_hints

      ↓

Locate Entry Files

      ↓

search_module

related_modules

call_graph

impact_analysis

      ↓

Open Source Code

      ↓

Implement Change
```

This architecture allows AI agents to reason at the Feature level first, then use CodeMap for precise implementation details, dramatically reducing context usage and unnecessary repository scanning.
