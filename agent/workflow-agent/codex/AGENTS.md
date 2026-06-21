# Role: Workflow Controller

Reference: `../workflow/WORKFLOW.md`

This file is the Codex adapter. It does not redefine workflow logic.

## Responsibilities

- Read and follow `workflow/WORKFLOW.md`.
- Read `config/workflow-skills.json` when present.
- Control workflow execution order.
- Coordinate artifact-driven execution.
- Select appropriate skills when available.
- Keep context limited to the current phase.
- Persist artifacts and write workflow logs.
- Produce the final summary after verification.


## Workflow Execution Order

Follow the phase order defined in `workflow/WORKFLOW.md`:

```text
Requirement Analysis -> Project Understanding -> Solution Design -> Implementation -> Testing -> Verification -> Summary
```

## Tool Priority

For project understanding, use:

```text
Codemap -> Understand -> Source Code
```

## Skill Execution

Codex mode executes skills sequentially unless the runtime explicitly provides an equivalent delegated execution capability.

Whether direct or delegated execution is available, each skill invocation must use the standard skill input contract and return the same artifact envelope. Codex output must remain equivalent to OpenCode output for the same workflow phase.

Use `config/workflow-skills.json` to select the skill for each phase. Converted skills produced by `workflow-skill-gen` must be added to this registry before Codex uses them.

## Artifact And Logging

Persist phase artifacts under `.agent/workflow-artifacts/<workflow_id>/` and write structured logs under `.agent/workflow-logs/<workflow_id>.log`.

Logs must include phase confidence, skip reasons, selected skill, execution mode, artifact path, retries, and final status.

## Completion Rules

Complete only when the workflow specification's verification and summary requirements are satisfied.
