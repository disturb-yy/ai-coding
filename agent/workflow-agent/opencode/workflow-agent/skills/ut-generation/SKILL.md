# ut-generation

## Purpose

Generate, update, or fix unit tests for the implemented behavior.

## Input

Expected execution input:

```json
{
  "skill": "ut-generation",
  "config": {},
  "input": {
    "solution_design": {},
    "implementation_result": {},
    "test_context": {}
  }
}
```

## Output

Return a `test_result` artifact:

```json
{
  "artifact_type": "test_result",
  "created_at": "",
  "phase": "Testing",
  "content": {
    "test_files": [],
    "coverage_notes": []
  }
}
```

## Responsibilities

- Identify relevant test scenarios.
- Create or update unit tests when required.
- Fix tests when the implementation changes expected behavior.
- Record test files and coverage notes.
- Keep tests scoped to the implemented change.

## Forbidden Actions

- Do NOT implement production logic.
- Do NOT change workflow state.
- Do NOT call other skills directly.
- Do NOT load unrelated test suites.
- Do NOT persist memory outside the artifact.

## Execution Rules

- Use implementation and design artifacts as the testing authority.
- Keep test changes focused on acceptance criteria.
- Prefer existing test patterns.
- Output only the required structured artifact.
