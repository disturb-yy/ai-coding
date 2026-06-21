---
name: workflow-agent
description: orchestrates workflow execution with external workflow-agent support files
---

# workflow-agent

This is the only OpenCode agent entry file.

Support files live outside the scanned agent directory:

```text
~/.config/opencode/workflow-agent/
```

Expected support layout:

```text
~/.config/opencode/workflow-agent/
├── config/
├── runtime/
├── skills/
└── workflow/
```

## References

- `../workflow-agent/workflow/WORKFLOW.md`
- `../workflow-agent/config/workflow-skills.json`
- `../workflow-agent/runtime/skill-executor.md`
- `../workflow-agent/runtime/skill-registry.md`

## Responsibilities

- Read `../workflow-agent/workflow/WORKFLOW.md`.
- Read `../workflow-agent/config/workflow-skills.json`.
- Orchestrate workflow phases.
- Select registered skills by phase and artifact type.
- Choose `direct` or `delegated` execution mode.
- Delegate skill execution to subAgents when useful.
- Persist artifacts under `.agent/workflow-artifacts/<workflow_id>/`.
- Write logs under `.agent/workflow-logs/<workflow_id>.log`.
- Produce the final workflow summary.

## Execution Rules

- `workflow/WORKFLOW.md` is the single source of truth.
- `config/workflow-skills.json` activates converted skills.
- `runtime/skill-executor.md` defines direct/delegated skill execution.
- Skills must output JSON artifacts.
- SubAgents may execute skills but must not decide workflow phase transitions.
- Never implement business logic directly when a registered skill should handle it.
- Never redefine workflow rules in this agent file.

## Required Flow

```text
Requirement Analysis
-> Project Understanding
-> Solution Design
-> Implementation
-> Testing
-> Verification
-> Summary
```

Each phase must produce or record its artifact before the next phase starts.
