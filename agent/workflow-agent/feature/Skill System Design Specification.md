
# Skill System Design Specification (V1)

## 1. Overview

This document defines a **Skill Execution System** for a workflow-based AI coding architecture.

It is designed to work with:

- workflow-agent (orchestrator)
- Codex adapter
- OpenCode adapter

---

## 2. Core Concept

A Skill is a **deterministic execution unit** that transforms input + config into a structured artifact.

### Skill Definition:

```
Skill = Definition + Config Schema + Execution Contract + Output Artifact Schema
```

Skills are NOT:

- Agents
- Prompts
- Free-form instructions
- Workflow controllers

---

## 3. System Architecture

```
User Request
    ↓
Workflow Engine (WORKFLOW.md)
    ↓
Skill Executor
    ↓
Codemap / Understand / Code
    ↓
Artifact Chain
    ↓
Final Output
```

---

## 4. Skill Execution Model

Every skill must follow this execution pipeline:

```
1. Load Skill Definition (SKILL.md)
2. Validate Config (config.schema.json)
3. Prepare Minimal Context
4. Execute Skill Logic
5. Output Structured Artifact
```

---

## 5. Directory Structure

```
agent/workflow-agent/
│
├── workflow/
│   └── WORKFLOW.md
│
├── skills/
│   ├── requirement-analysis/
│   │   ├── SKILL.md
│   │   ├── config.schema.json
│   │   └── example.json
│   │
│   ├── project-understanding/
│   │   ├── SKILL.md
│   │   ├── config.schema.json
│   │   └── example.json
│   │
│   ├── solution-design/
│   ├── code-implementation/
│   └── ut-generation/
│
├── runtime/
│   └── skill-executor.md
│
├── codex/
└── opencode/
```

---

## 6. Skill Definition Standard (SKILL.md)

Each skill MUST include:

### Required Sections

```
# Skill Name

## Purpose
Define what the skill does.

## Input
Define required inputs.

## Output
Define required artifact output.

## Responsibilities
List responsibilities.

## Forbidden Actions
List what the skill MUST NOT do.

## Execution Rules
Strict constraints for behavior.
```

---

## Example: requirement-analysis

```
## Forbidden Actions

- Do NOT design system architecture
- Do NOT modify code
- Do NOT access full repository
```

---

## 7. Config Schema (config.schema.json)

Each skill must define a strict JSON schema.

### Example:

```json
{
  "type": "object",
  "properties": {
    "depth": {
      "type": "string",
      "enum": ["fast", "normal", "deep"]
    },
    "ask_clarification": {
      "type": "boolean"
    },
    "output_style": {
      "type": "string",
      "enum": ["strict", "flexible"]
    }
  },
  "required": ["depth"]
}
```

---

## 8. Execution Contract

All skills MUST follow this contract:

### Input Format

```json
{
  "skill": "project-understanding",
  "config": {},
  "input": {}
}
```

---

### Output Format (Mandatory)

```json
{
  "artifact_type": "",
  "phase": "",
  "content": {}
}
```

---

## 9. Artifact System

Artifacts are the ONLY valid state mechanism.

### Base Artifact Format

```json
{
  "artifact_type": "",
  "created_at": "",
  "phase": "",
  "content": {}
}
```

---

## 10. Core Skill Definitions

---

### 10.1 requirement-analysis

**Goal:**
Extract structured requirements.

**Output:**

```json
{
  "goal": "",
  "constraints": [],
  "assumptions": [],
  "acceptance_criteria": []
}
```

---

### 10.2 project-understanding

**Goal:**
Understand codebase using Codemap + Understand.

**Output:**

```json
{
  "modules": [],
  "packages": [],
  "candidate_locations": [],
  "implementation_patterns": []
}
```

---

### 10.3 solution-design

**Goal:**
Design implementation plan.

**Output:**

```json
{
  "files_to_modify": [],
  "files_to_create": [],
  "implementation_plan": [],
  "risks": []
}
```

---

### 10.4 code-implementation

**Goal:**
Implement code changes based on design.

**Output:**

```json
{
  "modified_files": [],
  "created_files": [],
  "change_summary": ""
}
```

---

### 10.5 ut-generation

**Goal:**
Generate or fix unit tests.

**Output:**

```json
{
  "test_files": [],
  "coverage_notes": []
}
```

---

## 11. Skill Executor (runtime/skill-executor.md)

### Responsibilities

- Load skill definition
- Validate config schema
- Enforce context rules
- Execute skill
- Return artifact

---

### Context Rules

Allowed:

- previous artifact
- current skill input
- codemap results
- understand results

Forbidden:

- full repository context
- workflow history
- unrelated artifacts

---

### Execution Rules

```
Step 1: Validate input
Step 2: Load skill definition
Step 3: Apply config constraints
Step 4: Execute minimal context reasoning
Step 5: Output artifact
```

---

## 12. Workflow Integration Rules

Skills MUST NOT control workflow.

Only workflow-agent can:

- decide phase transitions
- call skills
- manage lifecycle

Skills only execute.

---

## 13. Codex / OpenCode Compatibility

### Codex Mode

- Linear execution
- No sub-agents
- Skills executed sequentially

### OpenCode Mode

- Supports sub-agent delegation
- Skills may run in parallel
- Workflow-agent coordinates execution

---

## 14. Critical Design Constraints

### 14.1 Single Source of Truth

```
WORKFLOW.md = global truth
```

---

### 14.2 Skill Isolation

Skills MUST NOT:

- call other skills directly
- modify workflow state
- persist memory

---

### 14.3 Artifact-Driven System

All state must flow through artifacts.

No hidden state allowed.

---

### 14.4 Context Minimization

Each skill must operate with minimal required context only.

---

## 15. Success Criteria

A correct implementation MUST satisfy:

- Skills are independently executable
- Workflow remains unchanged by skill updates
- Artifact chain is consistent across all phases
- Codex and OpenCode produce equivalent outputs
- No cross-skill coupling exists

---

## 16. Output Requirement for Implementer

The implementing AI must generate:

- All skill directories
- All SKILL.md files
- All config.schema.json files
- All example.json files
- Skill executor runtime specification

WITHOUT modifying workflow definition.

---

## End of Specification