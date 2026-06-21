# AI Development Workflow Specification

Version: 1.0

---

# Purpose

This document defines the standard workflow used by AI agents, skills, and automation systems to perform software development tasks.

The workflow is platform-independent and is designed to work across multiple AI coding environments.

Examples include:

- Codex
- OpenCode
- Claude Code
- Cursor
- Future AI Coding Systems

This document defines:

- Workflow lifecycle
- Workflow phases
- Artifact contracts
- Context management rules
- Verification requirements
- Failure handling
- Recovery strategy

This document does not define:

- Agent implementations
- Skill implementations
- Tool implementations
- Platform-specific behavior

---

# Core Principles

## Principle 1: Requirement Before Implementation

Implementation must never begin before requirements are understood.

Required execution order:

```text
Requirement Analysis
    ↓
Project Understanding
    ↓
Solution Design
    ↓
Implementation
    ↓
Testing
    ↓
Verification
    ↓
Summary
```

Do not skip phases without justification.

---

## Principle 2: Artifact Driven Workflow

Artifacts are the primary workflow memory.

Do not rely on:

- Chat history
- Long prompts
- Hidden context

Each phase must generate structured artifacts.

Artifacts must be reusable by later phases.

---

## Principle 3: Context Reduction

Each phase should receive only the context necessary to complete its objective.

Avoid:

- Entire repository context
- Entire workflow history
- Unrelated artifacts

Prefer:

- Relevant artifacts
- Relevant tool outputs
- Relevant files

---

## Principle 4: Tool First

Repository understanding should prioritize tools before source code.

Preferred order:

1. Codemap
2. Understand
3. Source Code

Avoid reading source code when higher-level information is sufficient.

---

## Principle 5: Evidence-Based Decisions

Do not guess.

All conclusions should be supported by:

- User requirements
- Workflow artifacts
- Tool outputs
- Existing implementation patterns

---

# Workflow Lifecycle

Workflow states:

```text
Created
    ↓
Planning
    ↓
Running
    ↓
Verifying
    ↓
Completed
```

Failure path:

```text
Running
    ↓
Failed
```

---

# Workflow Phases

## Phase 1: Requirement Analysis

### Goal

Understand the requested task.

### Inputs

- User request
- Requirement documents
- Specifications

### Outputs

Requirement Analysis Artifact

### Required Information

- Goal
- Constraints
- Assumptions
- Acceptance Criteria

### Completion Criteria

Must clearly answer:

- What should be built?
- What should not be changed?
- How will success be measured?

---

## Phase 2: Project Understanding

### Goal

Understand the existing project.

### Preferred Sources

1. Codemap
2. Understand
3. Source Code

### Outputs

Project Understanding Artifact

### Required Information

- Relevant modules
- Relevant packages
- Relevant files
- Candidate implementation locations
- Existing implementation patterns

### Completion Criteria

Must identify where implementation work should occur.

---

## Phase 3: Solution Design

### Goal

Produce an implementation strategy.

### Inputs

- Requirement Analysis Artifact
- Project Understanding Artifact

### Outputs

Solution Design Artifact

### Required Information

- Files to modify
- Files to create
- Implementation plan
- Test strategy
- Risks

### Completion Criteria

Must define:

- What changes are required
- Where changes occur
- How changes are validated

Implementation must not begin before design is complete.

---

## Phase 4: Implementation

### Goal

Apply source code changes.

### Inputs

- Solution Design Artifact
- Relevant project context

### Outputs

Implementation Artifact

### Required Information

- Modified files
- Created files
- Change summary

### Completion Criteria

Implementation completed according to design.

---

## Phase 5: Testing

### Goal

Create or update tests.

### Inputs

- Solution Design Artifact
- Implementation Artifact
- Git Diff

### Outputs

Test Artifact

### Required Information

- Test files
- Test scenarios
- Coverage notes

### Completion Criteria

Relevant functionality is covered.

---

## Phase 6: Verification

### Goal

Validate implementation quality.

### Verification Checklist

- Build passes
- Tests pass
- Acceptance criteria satisfied
- No obvious regressions

### Outputs

Verification Artifact

### Completion Criteria

All verification checks pass.

---

## Phase 7: Summary

### Goal

Generate final delivery report.

### Outputs

Workflow Summary Artifact

### Required Information

- What changed
- Why it changed
- Risks
- Follow-up work

### Completion Criteria

Final report generated.

---

# Artifact Specification

Artifacts are structured workflow outputs.

Every artifact must follow the format:

```json
{
  "artifact_type": "",
  "created_at": "",
  "phase": "",
  "content": {}
}
```

---

## requirement_analysis

```json
{
  "goal": "",
  "constraints": [],
  "assumptions": [],
  "acceptance_criteria": []
}
```

---

## project_understanding

```json
{
  "modules": [],
  "packages": [],
  "candidate_locations": [],
  "implementation_patterns": []
}
```

---

## solution_design

```json
{
  "files_to_modify": [],
  "files_to_create": [],
  "implementation_plan": [],
  "test_strategy": [],
  "risks": []
}
```

---

## implementation_result

```json
{
  "modified_files": [],
  "created_files": [],
  "change_summary": ""
}
```

---

## test_result

```json
{
  "test_files": [],
  "test_scenarios": [],
  "coverage_notes": []
}
```

---

## verification_result

```json
{
  "build_passed": true,
  "test_passed": true,
  "issues": []
}
```

---

## workflow_summary

```json
{
  "completed": true,
  "changed_files": [],
  "risks": [],
  "follow_ups": []
}
```

---

# Context Rules

## Requirement Analysis

### Allowed

- User request
- Requirement documents

### Forbidden

- Entire repository

---

## Project Understanding

### Allowed

- Requirement Analysis Artifact
- Codemap
- Understand

### Forbidden

- Unrelated modules

---

## Solution Design

### Allowed

- Requirement Analysis Artifact
- Project Understanding Artifact

### Forbidden

- Entire repository

---

## Implementation

### Allowed

- Solution Design Artifact
- Relevant files
- Relevant Codemap results
- Relevant Understand results

### Forbidden

- Entire workflow history
- Entire repository

---

## Testing

### Allowed

- Solution Design Artifact
- Implementation Artifact
- Git Diff

### Forbidden

- Unrelated artifacts

---

## Verification

### Allowed

- Build output
- Test output
- Acceptance criteria

### Forbidden

- Unrelated project history

---

# Verification Rules

Before a workflow can be completed:

- Build must pass
- Tests must pass
- Acceptance criteria must be satisfied

If any requirement fails:

Workflow status must remain incomplete.

---

# Fix Loop

Maximum retries:

```text
3
```

Execution flow:

```text
Identify Failure
    ↓
Analyze Root Cause
    ↓
Create Fix Plan
    ↓
Apply Fix
    ↓
Verify Again
```

If retry limit is exceeded:

```text
Workflow Status = Failed
```

---

# Logging Requirements

Every phase should produce structured logs.

Minimum format:

```json
{
  "timestamp": "",
  "phase": "",
  "status": "",
  "input_artifacts": [],
  "output_artifact": "",
  "summary": ""
}
```

Logs should support:

- Recovery
- Auditing
- Debugging
- Traceability

---

# Recovery Rules

Workflow execution should be resumable.

Recovery process:

```text
Load Workflow State
    ↓
Find Last Successful Phase
    ↓
Resume Execution
```

Completed phases should not be re-executed unless explicitly required.

---

# Completion Rules

A workflow is complete only when all phases succeed:

- Requirement Analysis completed
- Project Understanding completed
- Solution Design completed
- Implementation completed
- Testing completed
- Verification passed
- Summary generated

Otherwise:

```text
Workflow Status = Incomplete
```

---

# Operating Philosophy

The workflow exists to ensure:

- Correctness
- Traceability
- Recoverability
- Minimal context usage
- Consistent execution

Agents may vary.

Skills may vary.

Tools may vary.

The workflow remains the same.