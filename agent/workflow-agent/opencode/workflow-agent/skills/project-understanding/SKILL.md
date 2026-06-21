# project-understanding

## Purpose

Understand the relevant project area using Codemap and Understand before reading source code.

## Input

Expected execution input:

```json
{
  "skill": "project-understanding",
  "config": {},
  "input": {
    "requirement_analysis": {},
    "codemap_results": {},
    "understand_results": {}
  }
}
```

## Output

Return a `project_understanding` artifact:

```json
{
  "artifact_type": "project_understanding",
  "created_at": "",
  "phase": "Project Understanding",
  "content": {
    "modules": [],
    "packages": [],
    "candidate_locations": [],
    "implementation_patterns": []
  }
}
```

## Responsibilities

- Identify relevant modules and packages.
- Prefer Codemap results, then Understand results, then scoped source code if necessary.
- Identify candidate implementation locations.
- Summarize existing implementation patterns.
- Record unresolved unknowns.

## Forbidden Actions

- Do NOT load the full repository.
- Do NOT implement code changes.
- Do NOT write tests.
- Do NOT design the full solution beyond locating relevant areas.
- Do NOT call other skills directly.
- Do NOT change workflow state.
- Do NOT persist memory outside the artifact.

## Execution Rules

- Use the priority order: Codemap, Understand, Source Code.
- Read source only when higher-level tool output is insufficient.
- Keep source inspection limited to the smallest relevant scope.
- Output only the required structured artifact.
