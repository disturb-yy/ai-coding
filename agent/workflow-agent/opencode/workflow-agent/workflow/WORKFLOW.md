# Workflow Specification

Version: 1.0

`workflow/WORKFLOW.md` is the single source of truth for workflow behavior. Agent files and platform adapters must reference this file instead of redefining workflow logic.

## Scope

This specification defines a platform-independent AI software development workflow for Codex, OpenCode, and other coding systems.

It defines:

- Workflow lifecycle
- Required phase order
- JSON artifact contracts
- Context rules
- Verification rules
- Fix loop rules

It does not define:

- Agent implementation details
- Skill implementation details
- Platform-specific behavior
- Business logic for a target project

## Core Principles

- Execute phases in order.
- Communicate between phases with JSON artifacts.
- Keep context minimal for the current phase.
- Prefer repository understanding tools before source code.
- Verify with evidence before marking work complete.

## Workflow Lifecycle

Normal lifecycle:

```text
Created -> Planning -> Running -> Verifying -> Completed
```

Failure lifecycle:

```text
Running -> Failed
```

State meanings:

- `Created`: the workflow exists but no planning has started.
- `Planning`: requirements, project understanding, and solution design are being produced.
- `Running`: implementation and testing work is being performed.
- `Verifying`: build, test, and acceptance checks are being validated.
- `Completed`: all verification rules passed and summary artifact exists.
- `Failed`: the workflow cannot complete after the allowed fix loop.

## Workflow Phases

Phases must run in this order:

```text
1. Requirement Analysis
2. Project Understanding
3. Solution Design
4. Implementation
5. Testing
6. Verification
7. Summary
```

A phase may be skipped only when the workflow summary records a concrete reason and the skip does not violate verification rules.

### 1. Requirement Analysis

Goal: understand the requested task before inspecting implementation details.

Inputs:

- User request
- Requirement documents
- Explicit constraints

Output artifact: `requirement_analysis`

Required content:

- Goal
- Constraints
- Assumptions
- Acceptance criteria

Completion criteria:

- The requested outcome is clear.
- Non-goals and constraints are recorded.
- Success can be measured with acceptance criteria.

### 2. Project Understanding

Goal: identify the smallest relevant project area.

Tool priority:

1. Codemap
2. Understand
3. Source Code

Output artifact: `project_understanding`

Required content:

- Relevant modules
- Relevant packages or components
- Relevant files, when source files are needed
- Candidate implementation locations
- Existing implementation patterns

Completion criteria:

- The implementation area is identified.
- Any source code read is limited to the relevant scope.

### 3. Solution Design

Goal: define the implementation and validation strategy.

Inputs:

- `requirement_analysis`
- `project_understanding`

Output artifact: `solution_design`

Required content:

- Files to modify
- Files to create
- Implementation plan
- Test strategy
- Risks

Completion criteria:

- Required changes are known.
- Validation approach is known.
- Implementation has not started before this artifact exists.

### 4. Implementation

Goal: apply the planned changes.

Inputs:

- `solution_design`
- Minimal relevant source context

Output artifact: `implementation_result`

Required content:

- Modified files
- Created files
- Change summary
- Deviations from design, if any

Completion criteria:

- Changes are applied according to the design or deviations are justified.

### 5. Testing

Goal: create, update, or run the tests needed for the change.

Inputs:

- `solution_design`
- `implementation_result`
- Relevant diff or test output

Output artifact: `test_result`

Required content:

- Test files
- Test scenarios
- Commands run
- Coverage notes
- Failures, if any

Completion criteria:

- Relevant behavior is covered or a test gap is explicitly recorded.

### 6. Verification

Goal: validate that the workflow can be completed.

Inputs:

- `requirement_analysis`
- `implementation_result`
- `test_result`
- Build output
- Test output

Output artifact: `verification_result`

Completion criteria:

- Build passes.
- Tests pass.
- Acceptance criteria are satisfied.
- No unresolved blocking issue remains.

### 7. Summary

Goal: produce the final delivery record.

Inputs:

- All phase artifacts
- Final verification result

Output artifact: `workflow_summary`

Required content:

- Completed status
- Changed files
- Verification evidence
- Risks
- Follow-up work

Completion criteria:

- Summary exists.
- It accurately reflects completed work and remaining risk.

## Artifact Contract

Every phase output must be a JSON artifact with this envelope:

```json
{
  "artifact_type": "",
  "created_at": "",
  "phase": "",
  "content": {}
}
```

`artifact_type` must be one of:

- `requirement_analysis`
- `project_understanding`
- `solution_design`
- `implementation_result`
- `test_result`
- `verification_result`
- `workflow_summary`

## Artifact Storage

Artifacts must be persisted so a workflow can be audited or resumed.

Default directory:

```text
.agent/workflow-artifacts/<workflow_id>/
```

`workflow_id` should be stable for one user request. Recommended format:

```text
YYYYMMDD-HHMMSS-<task_slug>
```

Required files:

```text
.agent/workflow-artifacts/<workflow_id>/
├── 00-metadata.json
├── 01-requirement-analysis.json
├── 02-project-understanding.json
├── 03-solution-design.json
├── 04-implementation-result.json
├── 05-test-result.json
├── 06-verification-result.json
└── 07-workflow-summary.json
```

Artifact storage rules:

- Write one JSON file per completed phase.
- Do not overwrite a previous artifact without recording the retry attempt.
- For fix loop retries, append an attempt suffix:

```text
04-implementation-result.attempt-2.json
05-test-result.attempt-2.json
06-verification-result.attempt-2.json
```

- Store only structured artifacts and compact metadata, not full repository snapshots.
- If `.agent/` cannot be created, use `/tmp/workflow-agent-artifacts/<workflow_id>/` and record the fallback path in logs.

`00-metadata.json` must include:

```json
{
  "workflow_id": "",
  "created_at": "",
  "task_type": "",
  "status": "Created",
  "artifact_dir": "",
  "log_file": ""
}
```

### requirement_analysis

```json
{
  "artifact_type": "requirement_analysis",
  "created_at": "",
  "phase": "Requirement Analysis",
  "content": {
    "goal": "",
    "constraints": [],
    "assumptions": [],
    "acceptance_criteria": []
  }
}
```

### project_understanding

```json
{
  "artifact_type": "project_understanding",
  "created_at": "",
  "phase": "Project Understanding",
  "content": {
    "modules": [],
    "packages": [],
    "files": [],
    "candidate_locations": [],
    "implementation_patterns": []
  }
}
```

### solution_design

```json
{
  "artifact_type": "solution_design",
  "created_at": "",
  "phase": "Solution Design",
  "content": {
    "files_to_modify": [],
    "files_to_create": [],
    "implementation_plan": [],
    "test_strategy": [],
    "risks": []
  }
}
```

### implementation_result

```json
{
  "artifact_type": "implementation_result",
  "created_at": "",
  "phase": "Implementation",
  "content": {
    "modified_files": [],
    "created_files": [],
    "change_summary": "",
    "deviations": []
  }
}
```

### test_result

```json
{
  "artifact_type": "test_result",
  "created_at": "",
  "phase": "Testing",
  "content": {
    "test_files": [],
    "test_scenarios": [],
    "commands": [],
    "coverage_notes": [],
    "failures": []
  }
}
```

### verification_result

```json
{
  "artifact_type": "verification_result",
  "created_at": "",
  "phase": "Verification",
  "content": {
    "build_passed": false,
    "test_passed": false,
    "acceptance_criteria_satisfied": false,
    "commands": [],
    "issues": []
  }
}
```

### workflow_summary

```json
{
  "artifact_type": "workflow_summary",
  "created_at": "",
  "phase": "Summary",
  "content": {
    "completed": false,
    "changed_files": [],
    "verification_evidence": [],
    "risks": [],
    "follow_ups": []
  }
}
```

## Context Rules

Global rules:

- Do not load the entire repository.
- Use the smallest context needed for the current phase.
- Prefer artifacts over chat history.
- Prefer targeted tool output over raw source.
- Read source code only after Codemap and Understand are insufficient or source is required to implement or verify a change.

Repository understanding priority:

```text
Codemap -> Understand -> Source Code
```

Phase-specific context:

- Requirement Analysis: user request and requirement documents only.
- Project Understanding: requirement artifact, Codemap output, Understand output, and only relevant source when necessary.
- Solution Design: requirement and project understanding artifacts.
- Implementation: solution design and relevant files only.
- Testing: solution design, implementation result, relevant diff, and test files.
- Verification: build output, test output, and acceptance criteria.
- Summary: final artifacts and verification evidence.

## Skill Registry

Workflow-agent selects skills through a registry, not by hard-coding skill names in this file.

Default registry path:

```text
config/workflow-skills.json
```

If no project registry exists, use built-in defaults or `config/workflow-skills.example.json` as the template.

Registry entries bind a workflow phase and artifact type to a skill:

```json
{
  "phase": "Implementation",
  "artifact_type": "implementation_result",
  "skill": "my-coding-skill",
  "path": "skills/my-coding-skill",
  "execution": {
    "codex": "direct",
    "opencode": "delegated"
  }
}
```

Rules:

- A converted skill is available to workflow-agent only after it is added to the registry.
- The skill's output artifact type must match the registry `artifact_type`.
- Codex must be able to run the skill in `direct` mode.
- OpenCode may run the skill in `delegated` mode when subAgent execution is available.
- Workflow-agent may choose a different registered skill for the same phase when task constraints require it, but the produced artifact type must remain the same.

## Verification Rules

A workflow can be completed only when:

- Build passes.
- Tests pass.
- Acceptance criteria are satisfied.

If the project has no applicable build or test command, the verification artifact must record:

- Why the command is unavailable or not applicable.
- What substitute validation was performed.
- Remaining risk.

## Fix Loop

When verification fails, use this loop:

```text
failure -> root cause -> fix -> verify
```

Rules:

- Maximum retries: 3.
- Each retry must identify a root cause before applying a fix.
- Each retry must produce updated implementation, testing, and verification artifacts.
- If all retries fail, set lifecycle state to `Failed` and produce a partial `workflow_summary`.

## Logging Rules

Workflow execution must produce structured logs for phase communication, delegation, retries, skip decisions, confidence changes, and final summary.

Default log directory:

```text
.agent/workflow-logs/
```

Default log file:

```text
.agent/workflow-logs/<workflow_id>.log
```

If `.agent/` cannot be created, use `/tmp/workflow-agent-logs/<workflow_id>.log` and record the fallback path in `00-metadata.json`.

Every log line must use a grep-friendly kv format:

```text
[workflow-agent] key=value key=value
```

### Required Log Events

Workflow start:

```text
[workflow-agent] event=workflow_start workflow_id=<id> task_type=<type> status=Created artifact_dir=<path> log_file=<path>
```

Phase entry:

```text
[workflow-agent] event=phase_start phase=<phase> prev=<prev_phase> confidence=<0.0-1.0>
```

Phase skip:

```text
[workflow-agent] event=phase_skip phase=<phase> prev=<prev_phase> confidence=<0.0-1.0> skip_reason=<reason>
```

Skill start and return:

```text
[workflow-agent] event=skill_start phase=<phase> skill=<skill> mode=direct|delegated artifact_type=<artifact_type>
[workflow-agent] event=skill_return phase=<phase> skill=<skill> mode=direct|delegated confidence=<0.0-1.0> artifact=<path>
```

SubAgent delegation in OpenCode:

```text
[workflow-agent] event=subagent_spawn phase=<phase> skill=<skill> agent=<agent_name> scope=<scope>
[workflow-agent] event=subagent_return phase=<phase> skill=<skill> agent=<agent_name> confidence=<0.0-1.0>
```

Delegation failure and fallback:

```text
[workflow-agent] event=subagent_spawn_failed phase=<phase> skill=<skill> agent=<agent_name> reason=unsupported|timeout|model_404
[workflow-agent] event=fallback_local phase=<phase> skill=<skill> mode=direct
```

Confidence update:

```text
[workflow-agent] event=confidence phase=<phase> before=<0.0-1.0> after=<0.0-1.0> delta=<number> reason=<reason>
```

Fix loop retry:

```text
[workflow-agent] event=retry attempt=<n> from=Verification to=<phase> failure=<failure> root_cause=<root_cause>
```

Workflow completion:

```text
[workflow-agent] event=workflow_complete workflow_id=<id> status=Completed confidence=<0.0-1.0> summary=<artifact_path>
```

Workflow failure:

```text
[workflow-agent] event=workflow_failed workflow_id=<id> status=Failed attempts=<n> summary=<artifact_path>
```

### Logging Requirements

- Log every phase start, skip, skill invocation, skill return, retry, and final status.
- Record `confidence` for every phase start and phase completion.
- Record `skip_reason` whenever a phase is skipped.
- Record artifact file paths after each phase artifact is written.
- Keep the same log file for the whole workflow.
- Sanitize multiline summaries before writing log lines.

## Workflow Summary Requirements

The final `workflow_summary` artifact must include enough information to review the full execution without reading every artifact.

Required summary content:

```json
{
  "completed": false,
  "changed_files": [],
  "verification_evidence": [],
  "phase_summary": [
    {
      "phase": "",
      "status": "completed|skipped|failed",
      "skill": "",
      "artifact": "",
      "confidence": 0.0,
      "skip_reason": "",
      "notes": ""
    }
  ],
  "confidence_timeline": [],
  "retry_count": 0,
  "risks": [],
  "follow_ups": []
}
```

Summary rules:

- Include every phase, including skipped phases.
- Explain why each skipped phase was skipped.
- Include confidence per phase.
- Include retry count and final status.
- Include artifact and log locations.

## Completion Rules

The workflow is complete only when:

- All required phase artifacts exist.
- The verification artifact passes build, tests, and acceptance criteria.
- The summary artifact exists.
- Any skipped work, unavailable validation, or residual risk is explicitly recorded.
