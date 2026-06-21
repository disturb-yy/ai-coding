# Workflow Skill Conversion Contract

Use this reference when converting an existing skill into a workflow-compatible skill.

## Skill Executor Input

```json
{
  "skill": "",
  "config": {},
  "input": {}
}
```

## Artifact Envelope

```json
{
  "artifact_type": "",
  "created_at": "",
  "phase": "",
  "content": {}
}
```

## Phase And Artifact Mapping

| Existing skill type | Workflow phase | Artifact type |
| --- | --- | --- |
| Requirement extraction | Requirement Analysis | requirement_analysis |
| Repository/codebase analysis | Project Understanding | project_understanding |
| Design/planning | Solution Design | solution_design |
| Coding/patch implementation | Implementation | implementation_result |
| Unit-test generation/fix | Testing | test_result |
| Build/test verification | Verification | verification_result |
| Delivery reporting | Summary | workflow_summary |

## Artifact Content Schemas

### requirement_analysis

```json
{
  "goal": "",
  "constraints": [],
  "assumptions": [],
  "acceptance_criteria": []
}
```

### project_understanding

```json
{
  "modules": [],
  "packages": [],
  "candidate_locations": [],
  "implementation_patterns": []
}
```

### solution_design

```json
{
  "files_to_modify": [],
  "files_to_create": [],
  "implementation_plan": [],
  "risks": []
}
```

### implementation_result

```json
{
  "modified_files": [],
  "created_files": [],
  "change_summary": ""
}
```

### test_result

```json
{
  "test_files": [],
  "coverage_notes": []
}
```

### verification_result

```json
{
  "build_passed": false,
  "test_passed": false,
  "acceptance_criteria_satisfied": false,
  "issues": []
}
```

## Conversion Rules

- Preserve the original skill's domain-specific behavior.
- Remove phase transition logic.
- Remove direct calls to other skills.
- Replace persistent memory with artifact input and output.
- Replace broad source access with scoped context rules.
- Make platform assumptions explicit only in adapter notes, not in skill logic.
- Keep the skill usable by both Codex direct execution and OpenCode delegated execution.

## Registry Entry Contract

Converted skills must provide a registry entry for `config/workflow-skills.json`.

```json
{
  "phase": "",
  "artifact_type": "",
  "skill": "",
  "path": "",
  "enabled": true,
  "priority": 100,
  "execution": {
    "codex": "direct",
    "opencode": "direct"
  },
  "config": {}
}
```

Rules:

- `phase` must match the workflow phase.
- `artifact_type` must match the skill output.
- `path` must point to the converted skill directory.
- Use `opencode: delegated` only when subAgent execution is useful and safe.
- `config` must validate against the converted skill's `config.schema.json`.
