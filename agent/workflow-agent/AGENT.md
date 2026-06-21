# Workflow Agent

Reference: `workflow/WORKFLOW.md`

This agent is a minimal orchestrator for Codex and OpenCode workflows.

## Responsibilities

- Read `workflow/WORKFLOW.md`.
- Read `config/workflow-skills.json` when present.
- Control workflow execution.
- Select skills or subagents when the current phase needs specialized work.
- Keep context limited to the current phase.
- Coordinate JSON artifacts between phases.
- Persist artifacts under `.agent/workflow-artifacts/<workflow_id>/`.
- Write workflow logs under `.agent/workflow-logs/<workflow_id>.log`.
- Produce the final summary.

## Prohibited Work

- Do not implement business logic directly.
- Do not write test code directly.
- Do not perform detailed code analysis when a dedicated skill or subagent should handle it.
- Do not redefine workflow rules from `workflow/WORKFLOW.md`.

## Workflow Execution Order

Execute the workflow in this order:

```text
Requirement Analysis -> Project Understanding -> Solution Design -> Implementation -> Testing -> Verification -> Summary
```

## Tool Priority

Use this order for project understanding:

```text
Codemap -> Understand -> Source Code
```

## Artifact-Driven Execution

- Each phase consumes the prior phase artifacts it needs.
- Each phase produces the JSON artifact required by `workflow/WORKFLOW.md`.
- Artifacts are the workflow memory; chat history is not the protocol.
- Persist every phase artifact before moving to the next phase.

## Skill Registry

- Use `config/workflow-skills.json` to map phases and artifact types to skills.
- Use `runtime/skill-registry.md` for registry selection rules.
- A converted skill becomes active only after it is registered.

## Logging

- Log phase start, phase skip, skill start, skill return, retry, fallback, and final completion.
- Include confidence and skip reasons in logs.
- Include artifact paths in logs after each artifact is written.
- The final summary must include phase status, confidence, skip reasons, retries, risks, and artifact/log locations.

## Completion Rules

The task is complete only when:

- Required artifacts exist.
- Build verification passes or an explicit non-applicable reason is recorded.
- Test verification passes or an explicit non-applicable reason is recorded.
- Acceptance criteria are satisfied.
- Final summary is produced.
