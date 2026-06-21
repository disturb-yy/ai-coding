---
name: workflow-skill-gen
description: Convert existing Codex/OpenCode skills, including coding and unit-test generation skills, into workflow-compatible skills with JSON artifacts, config schemas, minimal-context rules, executor contracts, and isolation constraints. Use when adapting an existing SKILL.md or skill directory to work under workflow-agent, skill-executor, Codex sequential execution, or OpenCode delegated/subAgent execution.
---

# Workflow Skill Gen

## Purpose

Convert an existing skill into a workflow-compatible skill without changing the global workflow.

Use this skill to adapt skills such as coding, implementation, review, diagnosis, or unit-test generation so they can be called by workflow-agent through the skill-executor contract.

## Input

Accept any of:

- An existing `SKILL.md`.
- A skill directory path.
- A description of an existing skill's behavior.
- A target workflow phase and artifact type.
- Existing config or examples, when available.

If the target phase or artifact type is missing, infer the safest mapping and record the assumption.

## Output

Produce a converted skill package or a conversion patch containing:

- Updated `SKILL.md`.
- `config.schema.json`.
- `example.json`.
- A `config/workflow-skills.json` registry entry.
- Optional `agents/openai.yaml` metadata when the skill is used by Codex.
- A short conversion summary.

The converted skill must return the workflow artifact envelope:

```json
{
  "artifact_type": "",
  "created_at": "",
  "phase": "",
  "content": {}
}
```

## Responsibilities

- Analyze existing skill behavior and preserve useful domain instructions.
- Map the skill to one workflow phase and artifact type.
- Generate workflow-compatible `SKILL.md`, `config.schema.json`, and `example.json`.
- Generate the registry entry needed to activate the converted skill.
- Add artifact envelope requirements and minimal-context rules.
- Remove workflow-control behavior from the converted skill.
- Keep the converted skill compatible with Codex direct execution and OpenCode delegated execution.

## Forbidden Actions

- Do NOT modify `workflow/WORKFLOW.md`.
- Do NOT make the converted skill decide workflow phase transitions.
- Do NOT make the converted skill call other skills directly.
- Do NOT make the converted skill depend on hidden memory.
- Do NOT remove useful domain-specific behavior unless it violates workflow isolation.
- Do NOT couple the converted skill to one platform when a platform-neutral contract is possible.

## Execution Rules

- Treat the existing skill as the source of domain behavior.
- Treat the workflow artifact contract as the source of integration behavior.
- Prefer minimal edits when converting an existing skill directory.
- Preserve scripts, references, and assets unless they violate workflow isolation.
- Generate strict JSON config schemas with `additionalProperties: false`.
- Generate examples using the skill executor input contract.
- Record assumptions when phase or artifact mapping is inferred.

## Conversion Workflow

1. Inspect the existing skill's purpose, inputs, outputs, responsibilities, forbidden actions, and execution rules.
2. Map the skill to one workflow phase and artifact type.
3. Preserve useful domain behavior, tools, scripts, references, and examples.
4. Remove workflow-control behavior from the skill.
5. Add strict input, output, config, context, and artifact rules.
6. Add direct/delegated execution compatibility when relevant.
7. Generate or update `config.schema.json` and `example.json`.
8. Generate the `config/workflow-skills.json` entry.
9. Validate that the converted skill is isolated and artifact-driven.

## Phase Mapping

Use the mapping in [conversion-contract.md](references/conversion-contract.md) when choosing target phases and artifact types.

Common mappings:

- Requirement extraction skill -> `Requirement Analysis` / `requirement_analysis`
- Codebase analysis skill -> `Project Understanding` / `project_understanding`
- Architecture or implementation planning skill -> `Solution Design` / `solution_design`
- Coding or patch generation skill -> `Implementation` / `implementation_result`
- Unit-test generation or test-fix skill -> `Testing` / `test_result`
- Build/test verification skill -> `Verification` / `verification_result`

## Required SKILL.md Shape

The converted skill must include:

- `# Skill Name`
- `## Purpose`
- `## Input`
- `## Output`
- `## Responsibilities`
- `## Forbidden Actions`
- `## Execution Rules`

Keep behavior-specific instructions, but express them as skill-local execution rules rather than workflow-control instructions.

## Config Schema Rules

Generate a strict JSON schema:

- Use `additionalProperties: false`.
- Include only options the skill itself owns.
- Do not include workflow lifecycle state.
- Do not include cross-skill routing decisions.
- Require at least one meaningful config key when possible.

Typical config keys:

- `depth`: `fast | normal | deep`
- `mode`: skill-specific mode such as `implement | refactor | fix` or `generate | update | fix`
- `allow_source_fallback`: boolean for analysis skills
- `run_tests`: boolean for test skills
- `max_files`: integer bound for source-reading or editing skills

## Context Rules

Converted skills must:

- Use previous artifacts instead of chat history.
- Read only context required for the current skill invocation.
- Prefer Codemap and Understand before source code when doing project understanding.
- Avoid full repository context.
- Avoid unrelated artifacts.

## Isolation Rules

Converted skills must not:

- Decide workflow phase transitions.
- Modify workflow lifecycle state.
- Call other skills directly.
- Persist hidden memory.
- Depend on platform-specific subAgent behavior.
- Produce free-form output instead of artifacts.

## Direct And Delegated Compatibility

Design the converted skill so it works in both modes:

- Codex direct/sequential execution.
- OpenCode delegated/subAgent execution.

The skill must not know which mode is used. The workflow-agent or platform adapter chooses execution mode.

## Registry Entry

Every conversion must include a registry entry that can be copied into `config/workflow-skills.json`.

Example for a coding skill:

```json
{
  "phase": "Implementation",
  "artifact_type": "implementation_result",
  "skill": "my-coding-skill",
  "path": "skills/my-coding-skill",
  "enabled": true,
  "priority": 100,
  "execution": {
    "codex": "direct",
    "opencode": "delegated"
  },
  "config": {
    "mode": "implement"
  }
}
```

Example for a UT generation skill:

```json
{
  "phase": "Testing",
  "artifact_type": "test_result",
  "skill": "my-ut-skill",
  "path": "skills/my-ut-skill",
  "enabled": true,
  "priority": 100,
  "execution": {
    "codex": "direct",
    "opencode": "delegated"
  },
  "config": {
    "mode": "generate",
    "run_tests": true
  }
}
```

## Validation Checklist

Before finishing, verify:

- `SKILL.md` has required sections.
- `config.schema.json` is valid JSON schema.
- `example.json` follows the skill executor input contract.
- Output artifact type matches the target phase.
- Registry entry phase and artifact type match the converted skill output.
- Forbidden actions prevent workflow-control behavior.
- Context rules enforce minimal context.
- No other skill is called directly.
