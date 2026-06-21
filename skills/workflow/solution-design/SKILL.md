# solution-design

## Purpose

Design an implementation plan from requirement and project understanding artifacts.

## Input

Expected execution input:

```json
{
  "skill": "solution-design",
  "config": {},
  "input": {
    "requirement_analysis": {},
    "project_understanding": {}
  }
}
```

## Output

Return a `solution_design` artifact:

```json
{
  "artifact_type": "solution_design",
  "created_at": "",
  "phase": "Solution Design",
  "content": {
    "files_to_modify": [],
    "files_to_create": [],
    "implementation_plan": [],
    "risks": []
  }
}
```

## Responsibilities

- Choose the smallest viable implementation plan.
- Identify files to modify and create.
- Define ordered implementation steps.
- Record risks and validation considerations.
- Keep design aligned with existing project patterns.

## Forbidden Actions

- Do NOT modify code.
- Do NOT write tests.
- Do NOT read unrelated source files.
- Do NOT call other skills directly.
- Do NOT change workflow state.
- Do NOT persist memory outside the artifact.

## Execution Rules

- Use only requirement and project understanding artifacts plus minimal relevant context.
- Avoid broad refactors unless explicitly required.
- Output only the required structured artifact.
- Include risks when validation is uncertain or context is incomplete.
