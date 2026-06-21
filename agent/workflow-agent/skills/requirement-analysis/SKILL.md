# requirement-analysis

## Purpose

Extract structured requirements from the user request and explicit requirement documents.

## Input

Expected execution input:

```json
{
  "skill": "requirement-analysis",
  "config": {},
  "input": {
    "user_request": "",
    "requirement_documents": []
  }
}
```

## Output

Return a `requirement_analysis` artifact:

```json
{
  "artifact_type": "requirement_analysis",
  "created_at": "",
  "phase": "Requirement Analysis",
  "content": {
    "goal": "",
    "constraints": [],
    "assumptions": [],
    "acceptance_criteria": []
  }
}
```

## Responsibilities

- Identify the user's concrete goal.
- Extract explicit constraints.
- Record assumptions only when needed.
- Define measurable acceptance criteria.
- Preserve ambiguity as a clarification need instead of inventing facts.

## Forbidden Actions

- Do NOT design system architecture.
- Do NOT inspect the full repository.
- Do NOT modify code.
- Do NOT write tests.
- Do NOT call other skills directly.
- Do NOT change workflow state.
- Do NOT persist memory outside the artifact.

## Execution Rules

- Use only the user request and provided requirement documents.
- Keep context limited to requirement material.
- Output only the required structured artifact.
- If required information is missing, record it in `assumptions` or `acceptance_criteria` as appropriate.
