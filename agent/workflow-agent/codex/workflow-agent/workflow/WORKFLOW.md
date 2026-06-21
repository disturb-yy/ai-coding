# Workflow Specification

Version: 2.0

`workflow/WORKFLOW.md` is the single source of truth for workflow behavior. Agent files and platform adapters must reference this file instead of redefining workflow logic.

## Scope

This specification defines a platform-independent AI software development workflow for Codex, OpenCode, and other coding systems.

It defines:

- Workflow lifecycle
- Task type routing (feature vs diagnosis)
- Required phase order per task type
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

- Execute phases in order per task type.
- Communicate between phases with JSON artifacts.
- Keep context minimal for the current phase.
- Prefer repository understanding tools before source code.
- Verify with evidence before marking work complete.

## Task Type Routing

The workflow-agent must determine `task_type` from the user request before starting any phase:

| task_type | Trigger | Description |
|---|---|---|
| `feature` | New functionality, modification, refactoring | Standard 7-phase feature delivery |
| `diagnosis` | Bug report, error, unexpected behavior, "locate the problem" | 5-phase problem investigation and fix |

Default: `feature` when intent is ambiguous.

The `task_type` field must be recorded in the final `workflow_summary`.

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
- `Planning`: requirements/project understanding/solution design (feature) or problem understanding/root cause (diagnosis) are being produced.
- `Running`: implementation and testing/fix work is being performed.
- `Verifying`: build, test, regression, and acceptance checks are being validated.
- `Completed`: all verification rules passed and summary artifact exists.
- `Failed`: the workflow cannot complete after the allowed fix loop.

---

# Feature Workflow (task_type: feature)

## Feature Phases

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

---

# Diagnosis Workflow (task_type: diagnosis)

## Diagnosis Phases

Phases must run in this order:

```text
1. Problem Understanding
2. Root Cause Analysis
3. Fix Implementation
4. Regression Check
5. Summary
```

A phase may be skipped only when the workflow summary records a concrete reason and the skip does not violate verification rules.

### 1. Problem Understanding

Goal: capture the problem precisely before investigating code.

Inputs:

- User bug report or error description
- Error logs, stack traces, screenshots
- Steps to reproduce
- Observed vs expected behavior

Output artifact: `problem_understanding`

Required content:

- Problem description (what, when, where)
- Reproduction steps
- Observed behavior
- Expected behavior
- Impact scope (affected users/features)
- Environment context (version, config, data)

Completion criteria:

- The problem can be reproduced or clearly described.
- Impact scope is understood.
- Investigation is not yet started.

### 2. Root Cause Analysis

Goal: trace from symptoms to root cause.

Tool priority:

1. Error logs and stack traces
2. Codemap (trace call chains)
3. Source code inspection (focused on the call chain)

Output artifact: `root_cause_analysis`

Required content:

- Call chain / data flow from trigger to failure
- Root cause (the exact line or logic that causes the problem)
- Contributing factors
- Evidence (log excerpts, variable values, conditions)
- Confidence level (0.0-1.0)

Completion criteria:

- A specific root cause is identified with evidence.
- Contributing factors are documented.
- Confidence is assessed.

### 3. Fix Implementation

Goal: apply the minimal fix for the root cause.

Inputs:

- `root_cause_analysis`
- Targeted source context around the root cause

Output artifact: `diagnosis_fix`

Required content:

- Modified files
- Created files (if any)
- Change summary
- Fix rationale (why this fix addresses the root cause)
- Deviations from root cause analysis, if any

Completion criteria:

- The fix is applied.
- Rationale is documented.
- Fix is minimal and targeted.

### 4. Regression Check

Goal: confirm the problem is resolved and no new issues introduced.

Inputs:

- `problem_understanding`
- `diagnosis_fix`
- Build output
- Test output (existing + any new tests)

Output artifact: `regression_check`

Required content:

- Reproduction test result (problem no longer occurs)
- Existing test results (no regressions)
- Build status
- Any remaining risk

Completion criteria:

- Original problem no longer occurs.
- Existing tests pass.
- Build passes.
- No new blocking issues.

### 5. Summary

Goal: produce the final delivery record.

Inputs:

- All diagnosis phase artifacts
- Final regression check result

Output artifact: `workflow_summary`

Required content:

- Completed status
- Changed files
- Root cause summary
- Fix summary
- Verification evidence
- Risks
- Follow-up work

Completion criteria:

- Summary exists.
- It accurately reflects completed work and remaining risk.

---

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

Feature:
- `requirement_analysis`
- `project_understanding`
- `solution_design`
- `implementation_result`
- `test_result`
- `verification_result`
- `workflow_summary`

Diagnosis:
- `problem_understanding`
- `root_cause_analysis`
- `diagnosis_fix`
- `regression_check`
- `workflow_summary`

All artifacts must be valid JSON. Content structure is defined by the registered skill for each phase.

---

## Artifact Examples

### Feature: requirement_analysis

```json
{
  "artifact_type": "requirement_analysis",
  "created_at": "2026-01-01T00:00:00Z",
  "phase": "Requirement Analysis",
  "content": {
    "goal": "Add advance/decline statistics to /indices API",
    "constraints": ["Reuse existing Tushare data source", "Don't break existing API consumers"],
    "assumptions": ["Market data API already provides up_num/down_num fields"],
    "acceptance_criteria": ["API returns advance/decline counts", "Frontend displays the numbers"]
  }
}
```

### Diagnosis: root_cause_analysis

```json
{
  "artifact_type": "root_cause_analysis",
  "created_at": "2026-01-01T00:00:00Z",
  "phase": "Root Cause Analysis",
  "content": {
    "call_chain": ["GetIndices() -> FetchAndSave() -> clsMarketDataResp unmarshal"],
    "root_cause": "clsMarketData.UpDownDis is nil when market is closed, causing nil pointer dereference at market_statistic_api.go:67",
    "contributing_factors": ["No nil check before field access", "Weekend/holiday market closure not handled"],
    "evidence": ["Stack trace: panic at market_statistic_api.go:67", "Reproduced on Saturday"],
    "confidence": 0.95
  }
}
```

## Skill Registry

Registry entries bind a workflow phase and artifact type to a skill:

```json
{
  "version": 1,
  "skills": [
    {
      "phase": "Implementation",
      "artifact_type": "implementation_result",
      "skill": "code-implementation",
      "path": "skills/code-implementation",
      "enabled": true,
      "priority": 100,
      "execution": {
        "codex": "direct",
        "opencode": "delegated"
      },
      "config": {}
    }
  ]
}
```

Selection rules:

- Match by `phase` and `artifact_type`.
- Ignore disabled entries.
- Prefer highest `priority`.
- Use `execution.codex` in Codex mode, `execution.opencode` in OpenCode mode.
- The skill's output artifact type must match the registry `artifact_type`.

## Context Rules

- Load only the artifacts required for the current phase.
- Keep source code access scoped to relevant files.
- Do not pass the full project context to a skill.

## Artifact Storage

Default artifact directory:

```text
.agent/workflow-artifacts/<workflow_id>/
```

Artifact file naming:

```text
.agent/workflow-artifacts/<workflow_id>/<artifact_type>.json
```

- Do not overwrite a previous artifact without recording the retry attempt.
- For fix loop retries, append an attempt suffix:
  ```text
  .agent/workflow-artifacts/<workflow_id>/<artifact_type>_retry_<n>.json
  ```

## Verification Rules

A workflow can be completed only when:

- Build passes.
- Tests pass.
- Acceptance criteria are satisfied (feature) OR original problem is resolved (diagnosis).

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

For diagnosis workflows, the Fix Loop is built into the phase order (Problem Understanding -> Root Cause Analysis -> Fix -> Regression Check) and retries apply if regression fails.

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
[workflow-agent] event=workflow_complete workflow_id=<id> task_type=<type> status=Completed confidence=<0.0-1.0> summary=<artifact_path>
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
  "task_type": "feature",
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

- Include `task_type`.
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
