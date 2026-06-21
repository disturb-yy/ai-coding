# code-implementation

## Purpose

Implement code or file changes according to an approved solution design.

## Input

Expected execution input:

```json
{
  "skill": "code-implementation",
  "config": {},
  "input": {
    "solution_design": {},
    "minimal_source_context": {}
  }
}
```

## Output

Return an `implementation_result` artifact:

```json
{
  "artifact_type": "implementation_result",
  "created_at": "",
  "phase": "Implementation",
  "content": {
    "modified_files": [],
    "created_files": [],
    "change_summary": ""
  }
}
```

## Responsibilities

- Apply only the changes described by the solution design.
- Preserve existing project style.
- Keep edits scoped to the requested implementation.
- Record modified and created files.
- Record deviations from the design when unavoidable.

## Forbidden Actions

- Do NOT change workflow state.
- Do NOT call other skills directly.
- Do NOT perform unrelated refactors.
- Do NOT load unrelated source files.
- Do NOT persist memory outside the artifact.
- Do NOT skip artifact output.

## Execution Rules

- Use the solution design as the implementation authority.
- Use only minimal source context required for the target files.
- Prefer small, reviewable edits.
- Output only the required structured artifact after changes are applied.
