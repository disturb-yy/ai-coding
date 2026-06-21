# Skill Executor Runtime

## Purpose

The skill executor runs one skill as a deterministic execution unit. It loads the skill definition, validates config, prepares minimal context, executes the skill logic, and returns a structured artifact.

## Responsibilities

- Load the target SKILL.md.
- Validate config against config.schema.json.
- Enforce context rules.
- Execute only the requested skill.
- Return the required artifact envelope.
- Return artifact metadata needed for artifact storage and workflow logs.

## Input Contract

`json
{
  "skill": "",
  "config": {},
  "input": {}
}
`

## Output Contract

`json
{
  "artifact_type": "",
  "created_at": "",
  "phase": "",
  "content": {}
}
`

## Execution Pipeline

`	ext
1. Validate input
2. Resolve skill registry entry
3. Load skill definition
4. Validate config schema
5. ENFORCE execution mode (delegated phases MUST use task tool)
6. Prepare minimal context
7. Resolve execution mode
8. Execute skill logic (direct OR delegated via task)
9. Return structured artifact
`

## Skill Registry

The executor receives or resolves a registry entry from config/workflow-skills.json.

The registry entry determines:

- Skill path
- Target phase
- Expected artifact type
- Default config
- Codex execution mode
- OpenCode execution mode

The executor must reject output when the returned rtifact_type does not match the registry entry.

## Execution Modes

The executor supports two execution modes with the same input and output contracts:

- direct: execute the skill in the current agent context.
- delegated: execute the skill through the  + "	ask" +  tool as a subAgent.

Mode selection is determined by workflow-skills.json execution.opencode field. **The agent MUST follow this field — it is not optional.**

Direct mode is the default when delegation is unavailable or unnecessary.

**Delegated mode is MANDATORY when the registry says delegated.** The agent MUST use the  + "	ask" +  tool. See untime/skill-delegator.md for exact procedure.

### Delegation enforcement rules

1. Before starting ANY phase, read config/workflow-skills.json for that phase.
2. If execution.opencode is delegated:
   - Read untime/skill-delegator.md
   - Call  + "	ask(agent=\"worker\", prompt=...)" + 
   - Do NOT execute any skill logic directly
3. If execution.opencode is direct:
   - Execute skill in current context
   - Do NOT spawn subAgents

**Violation: doing direct execution when registry says delegated is a workflow error.**

### Detection: are you violating delegation?

Ask yourself before any read/edit/bash during a delegated phase:
- "Am I reading a source file that the subAgent should read?"
- "Am I editing a file that the subAgent should edit?"
- "Am I running a command that the subAgent should run?"

If YES to any, STOP and spawn the subAgent.

## Context Rules

Allowed context:

- Previous artifact required by the current skill
- Current skill input
- Codemap results
- Understand results
- Minimal scoped source or test files when the skill explicitly requires them

Forbidden context:

- Full repository context
- Full workflow history
- Unrelated artifacts
- Hidden state

## Execution Rules

- Run only the skill named in the input.
- Validate the selected registry entry before loading the skill.
- Do not let a skill call another skill directly.
- Do not let a skill modify workflow lifecycle state.
- Do not let delegated workers decide workflow phase transitions.
- Do not persist memory outside the returned artifact.
- Do not accept config keys that are not allowed by the schema.
- Return failure as a structured artifact when execution cannot complete.
- Normalize direct and delegated outputs to the same artifact envelope.
- Return the artifact path after workflow-agent persists it.

## Artifact And Log Handoff

The executor does not own workflow storage, but its result must contain enough metadata for workflow-agent to persist and log the invocation.

Required handoff fields:

`json
{
  "artifact": {},
  "metadata": {
    "skill": "",
    "phase": "",
    "artifact_type": "",
    "mode": "direct",
    "confidence": 0.0,
    "summary": ""
  }
}
`

Workflow-agent writes the artifact to .agent/workflow-artifacts/<workflow_id>/ and records the returned metadata in .agent/workflow-logs/<workflow_id>.log.

## Workflow Integration

The workflow-agent is the only component allowed to decide phase transitions, select skills, choose execution mode, and manage lifecycle. The skill executor only executes an isolated skill request and returns its artifact.
