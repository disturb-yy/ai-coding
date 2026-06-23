# Find Change Points Design

## Overview

### Purpose

`find_change_points` is a Decision Layer capability built on top of CodeMap.

Unlike existing Fact Layer tools, it does not answer:

```text
Where is the code?
Who calls this function?
Which module owns this route?
```

Instead, it answers:

```text
What is most likely to change for this requirement?
Where should implementation begin?
What are the potential risks?
Which files should be investigated first?
```

Its goal is to significantly reduce repository-wide scanning and allow AI agents to move from requirement analysis to implementation planning efficiently.

---

# Position in Cognitive Architecture

```text
Requirement

        ↓

find_change_points

        ↓

Candidate Features

        ↓

Candidate Modules

        ↓

Candidate Files

        ↓

Risk Analysis

        ↓

Implementation Plan
```

Integrated architecture:

```text
PROJECT_INDEX

        ↑

Feature Layer
(get_feature_map)

        ↑

Navigation Layer
(get_navigation_hints)

        ↑

Decision Layer
(find_change_points)

        ↑

Fact Layer
(Route / Flow / Module / CallGraph)

        ↑

Source Code
```

---

# MCP Tool Definition

## Tool Name

```text
find_change_points
```

## Description

Identify likely modules, files, routes, flows, and risks related to a requirement.

This tool should be used before opening source code or planning implementation.

---

# Request Schema

```json
{
  "requirement": "Add order cancellation functionality",
  "top_k": 5
}
```

## Fields

| Field       | Type    | Required | Description                            |
| ----------- | ------- | -------- | -------------------------------------- |
| requirement | string  | Yes      | User requirement text                  |
| top_k       | integer | No       | Maximum number of candidates to return |

---

# Response Schema

```json
{
  "requirement": "Add order cancellation functionality",

  "matched_features": [
    {
      "name": "Create Order",
      "score": 0.82,
      "reason": "Related order lifecycle functionality"
    }
  ],

  "candidate_modules": [
    {
      "module": "order",
      "score": 0.91,
      "reason": "Primary owner of order lifecycle logic"
    }
  ],

  "candidate_files": [
    {
      "file": "internal/order/service.go",
      "module": "order",
      "score": 0.88,
      "reason": "Contains order state transition logic"
    }
  ],

  "related_routes": [
    "POST /orders"
  ],

  "related_flows": [
    "create_order_flow"
  ],

  "risk_areas": [
    {
      "name": "inventory_consistency",
      "level": "high",
      "reason": "Inventory rollback may be required"
    }
  ],

  "next_actions": [
    "search_module(order)",
    "call_graph(order/service.go)",
    "impact_analysis(order)"
  ]
}
```

---

# Design Principles

## Rule-Based

No LLM dependency.

All decisions should be generated from:

```text
Route Graph
Flow Graph
Module Graph
Dependency Graph
Call Graph
Impact Graph
```

---

## Probabilistic

The tool does not attempt to provide exact modification locations.

Output should always represent:

```text
Likely change points
```

instead of:

```text
Guaranteed change points
```

Therefore response fields should use:

```text
candidate_modules
candidate_files
```

instead of:

```text
target_modules
target_files
```

---

# High-Level Workflow

```text
Requirement

    ↓

Keyword Extraction

    ↓

Feature Matching

    ↓

Module Scoring

    ↓

File Scoring

    ↓

Risk Analysis

    ↓

Change Plan Generation
```

---

# Requirement Processing

## Tokenization

Requirement text is normalized before matching.

Example:

```text
Add order cancellation functionality
```

Generated tokens:

```text
add
order
cancel
functionality
```

---

## Normalization

Examples:

```text
orders      -> order
cancelled   -> cancel
cancellation-> cancel
creating    -> create
```

---

## Synonym Expansion

Example dictionary:

```go
var synonymMap = map[string][]string{
    "cancel": {
        "cancel",
        "cancellation",
        "cancelled",
        "close",
        "terminate",
    },

    "create": {
        "create",
        "add",
        "new",
        "insert",
    },

    "delete": {
        "delete",
        "remove",
    },

    "update": {
        "update",
        "modify",
        "edit",
    },

    "login": {
        "login",
        "signin",
        "authenticate",
        "auth",
    },

    "refund": {
        "refund",
        "rollback",
        "return",
    }
}
```

---

# Feature Matching

## Inputs

Generated Feature Map:

```text
Route
+
Flow
+
Module
=
Feature
```

---

## Matching Targets

Each requirement token is matched against:

```text
feature.name
feature.description
feature.modules
feature.routes
feature.flows
```

---

## Scoring

Suggested weights:

| Match Type          | Score |
| ------------------- | ----- |
| Feature Name        | +5    |
| Feature Description | +3    |
| Module Name         | +4    |
| Route               | +3    |
| Flow Name           | +4    |
| Synonym Match       | +2    |

---

## Example

Requirement:

```text
cancel order
```

Feature:

```text
Create Order
```

Matches:

```text
order
orders
create_order_flow
order module
```

Result:

```text
score = 0.82
```

---

# Module Scoring

## Candidate Sources

Modules are collected from:

```text
Matched Features

Related Modules

Route Owners

Flow Owners

Dependency Neighbors

Impact Analysis
```

---

## Scoring Rules

| Source                   | Score |
| ------------------------ | ----- |
| Direct Requirement Match | +10   |
| Matched Feature          | +8    |
| Route Owner              | +7    |
| Flow Owner               | +7    |
| Related Module           | +5    |
| Call Graph Neighbor      | +4    |
| Impact Analysis          | +3    |

---

## Output

```json
{
  "module": "order",
  "score": 0.91,
  "reason": "Primary owner of order lifecycle"
}
```

---

# File Scoring

## Candidate Sources

Files are collected from:

```text
Route Handlers

Flow Entry Files

Module Files

Call Graph Files

Service Layer

Repository Layer

Controller Layer
```

---

# Requirement Type Classification

Requirement type influences file ranking.

---

## Supported Types

```go
type RequirementType string

const (
    ReqAddAPI
    ReqModifyLogic
    ReqFixBug
    ReqAddField
    ReqAddAuth
    ReqUnknown
)
```

---

## Classification Rules

### Add API

Keywords:

```text
add api
new endpoint
new route
create api
```

Priority:

```text
router
controller
service
dto
```

---

### Modify Logic

Keywords:

```text
change
modify
update
business rule
logic
```

Priority:

```text
service
usecase
domain
biz
```

---

### Fix Bug

Keywords:

```text
fix
bug
error
failed
wrong
```

Priority:

```text
handler
service
call graph path
```

---

### Add Field

Keywords:

```text
field
column
request
response
param
```

Priority:

```text
model
entity
dto
repository
migration
```

---

### Add Auth

Keywords:

```text
auth
login
token
permission
session
```

Priority:

```text
middleware
auth
service
controller
```

---

# Risk Analysis

## Risk Sources

### Critical Modules

Built-in risk dictionary:

```go
var riskModuleMap = map[string]string{
    "auth":       "security",
    "payment":    "payment",
    "order":      "business_state",
    "inventory":  "consistency",
    "user":       "user_data",
    "permission": "access_control",
    "token":      "security",
    "session":    "security",
}
```

---

### Dependency Risk

Rule:

```text
Dependency Count > 5

→ Medium Risk
```

---

### Impact Risk

Rule:

```text
Dependent Count > 5

→ High Risk
```

---

### Call Graph Risk

Rule:

```text
Caller Count > 3

→ High Risk
```

---

### Route Risk

Sensitive routes:

```text
/login
/pay
/refund
/delete
/admin
```

Automatically increase risk score.

---

# Internal Data Structures

## Request

```go
type FindChangePointsRequest struct {
    Requirement string `json:"requirement"`
    TopK        int    `json:"top_k,omitempty"`
}
```

---

## Response

```go
type FindChangePointsResponse struct {
    Requirement      string
    MatchedFeatures  []MatchedFeature
    CandidateModules []CandidateModule
    CandidateFiles   []CandidateFile
    RelatedRoutes    []string
    RelatedFlows     []string
    RiskAreas        []RiskArea
    NextActions      []string
}
```

---

## Candidate Module

```go
type CandidateModule struct {
    Module string
    Score  float64
    Reason string
}
```

---

## Candidate File

```go
type CandidateFile struct {
    File   string
    Module string
    Score  float64
    Reason string
}
```

---

## Risk Area

```go
type RiskArea struct {
    Name   string
    Level  string
    Reason string
}
```

---

# Package Layout

```text
internal/
├── cognitive/
│
├── types.go
├── tokenizer.go
├── feature.go
├── navigation.go
├── scoring.go
├── risk.go
└── change_points.go
```

---

# Core Service

```go
type CognitiveService struct {
    repo Repository
}
```

---

## Main Entry

```go
func (s *CognitiveService) FindChangePoints(
    ctx context.Context,
    req FindChangePointsRequest,
) (*FindChangePointsResponse, error)
```

---

# Execution Pipeline

```go
tokens := TokenizeRequirement(req.Requirement)

features := BuildFeatureMap()

matched := MatchFeatures(tokens, features)

hints := BuildNavigationHints()

modules := ScoreCandidateModules(
    tokens,
    matched,
    hints,
)

files := ScoreCandidateFiles(
    tokens,
    modules,
    hints,
)

risks := InferRiskAreas(
    modules,
    files,
)

actions := BuildNextActions(
    modules,
    files,
)
```

---

# MCP Registration

```go
registerFindChangePointsTool(
    server,
    cognitiveService,
)
```

Description:

```text
Find likely modules, files, routes, flows and risks
for implementing a requirement.
Use before opening source code or planning changes.
```

---

# Agent Workflow

Requirement:

```text
Add SMS Login
```

Execution:

```text
find_change_points

    ↓

User Login Feature

    ↓

Candidate Modules:
auth
user
sms

    ↓

Candidate Files:
auth/controller.go
auth/service.go
user/service.go

    ↓

Risk:
token
session
security
```

Follow-up:

```text
search_module(auth)

call_graph(AuthService.Login)

impact_analysis(auth)
```

---

# Success Criteria

The tool is considered successful if it can:

```text
Reduce repository scanning

Improve modification planning

Identify likely entry points

Surface risk areas

Generate actionable next steps
```

without requiring source-code-wide traversal by the AI agent.

---

# Core Formula

```text
Requirement
+
Feature Map
+
Navigation Hints
+
Fact Graph

=

Candidate Change Plan
```

This transforms CodeMap from a code discovery tool into an AI-assisted software change planning engine.
