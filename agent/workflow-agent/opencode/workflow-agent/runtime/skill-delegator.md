# Skill Delegator Runtime (OpenCode)

## Purpose

This file defines the exact procedure for executing a workflow skill in delegated mode on OpenCode. When workflow-skills.json says "opencode": "delegated", the workflow-agent MUST follow this file — not improvise.

## When this file applies

This file applies for EVERY phase where:

``json
"execution": { "opencode": "delegated" }
``

Currently delegated phases:
- **Project Understanding** → skill: project-understanding
- **Implementation** → skill: code-implementation
- **Testing** → skill: ut-generation

## Step-by-step delegation procedure

### Step 1: Read the skill definition

``text
Read: ~/.config/opencode/workflow-agent/skills/<skill-name>/SKILL.md
``

Extract:
- Skill purpose
- Required input fields
- Required output artifact_type
- Forbidden actions
- Execution rules

### Step 2: Read the skill config schema

``text
Read: ~/.config/opencode/workflow-agent/skills/<skill-name>/config.schema.json
``

Use this to construct the config block for the subAgent.

### Step 3: Collect inputs

Gather from previous artifacts:
- For **Project Understanding**: user_request from equirement_analysis artifact
- For **Solution Design**: equirement_analysis + project_understanding artifacts
- For **Implementation**: solution_design artifact
- For **Testing**: solution_design + implementation_result artifacts

### Step 4: Construct the subAgent prompt

Use this template EXACTLY. Replace {placeholders}:

``text
You are executing the workflow skill {skill_name} for phase {phase}.

## Skill Definition
{content of skill SKILL.md}

## Config
`json
{config_json}
`

## Input
`json
{input_json}
`

## Instructions
1. Read the skill definition above carefully.
2. Read the config schema before acting.
3. Execute ONLY what the skill definition allows.
4. Return your result as a valid JSON artifact with this exact envelope:

`json
{
  "artifact_type": "{artifact_type}",
  "created_at": "{iso_timestamp}",
  "phase": "{phase}",
  "content": {
    ... skill-specific content ...
  }
}
`

5. Do NOT modify files outside the skill's scope.
6. Do NOT decide workflow phase transitions.
7. Return ONLY the final artifact JSON. No extra commentary.
``

### Step 5: Call the task tool

Use the  + "	ask" +  tool with these exact parameters:

``
agent: "worker"
prompt: <the constructed prompt from Step 4>
``

Example:

``text
task(
  agent: "worker",
  prompt: "You are executing the workflow skill project-understanding for phase Project Understanding.

## Skill Definition
# project-understanding
... (full SKILL.md content) ...

## Config
`json
{"depth": "normal", "allow_source_fallback": true, "max_files": 8}
`

## Input
`json
{
  "user_request": "为股票行情页面新增涨跌家数统计指标...",
  "requirement_documents": []
}
`
..."
)
``

### Step 6: Wait for subAgent result

- The  + "	ask" +  tool returns when the subAgent completes.
- Extract the JSON artifact from the subAgent response.
- The subAgent's last message should contain the artifact.

### Step 7: Validate the artifact

Check:
1. rtifact_type matches the registry entry
2. Required fields are present (check skill SKILL.md output section)
3. JSON is valid and parseable

If validation fails:
- Log the failure with event=skill_return confidence=<low>
- If retry budget remains, re-spawn with corrected prompt
- If retries exhausted, fall back to direct mode and log event=fallback_local

### Step 8: Persist the artifact

Write to:
``text
.agent/workflow-artifacts/<workflow_id>/<artifact_type>.json
``

Log the event:
``text
[workflow-agent] event=skill_return phase=<phase> skill=<skill> mode=delegated confidence=<0.0-1.0> artifact=<path>
``

## Anti-patterns (DO NOT DO)

| Wrong | Right |
|---|---|
| Read source files yourself during delegated phase | Let subAgent read source files |
| grep/codegraph during delegated phase | Let subAgent search |
| Edit files directly during Implementation phase | Spawn subAgent to edit |
| Write tests directly during Testing phase | Spawn subAgent to write tests |
| Skip delegation because "it's faster direct" | ALWAYS delegate when registry says delegated |
| Start executing before reading this file | ALWAYS read this file first |

## Quick Reference Card

``text
FOR EACH DELEGATED PHASE:
  1. read skill SKILL.md
  2. read skill config.schema.json  
  3. collect input artifacts
  4. build prompt from template above
  5. task(agent="worker", prompt=<built>)
  6. validate result
  7. persist artifact
  8. log
``
