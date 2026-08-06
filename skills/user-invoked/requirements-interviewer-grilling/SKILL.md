---
name: requirements-interviewer-grilling
description: Use the requirements_interviewer role with grilling to turn an unclear request into testable acceptance criteria and a next route.
disable-model-invocation: true
---

# Requirements Interviewer Grilling

Run the task as `requirements_interviewer` through the `grilling` workflow.

## Workflow

1. Read [requirements-interviewer.md](../../role/requirements-interviewer.md) and its linked
   [handoff standard](../../role/handoff-standard.md). Treat them as the role contract.
2. Read and follow [grilling](../../model-involved/grilling/SKILL.md). Treat it as the requirement
   clarification workflow.
3. Resolve only uncertainty that changes the goal, acceptance criteria, constraints, or route. Ask
   one high-value question at a time; do not ask for facts the repository or supplied sources answer.
4. Stop when the requirement is actionable, needs evidence from an existing source, or the user
   changes scope. Do not start exploration or implementation.

## Handoff

Report the role contract's required fields:

- `clarified_goal`
- `acceptance_criteria`
- `constraints`
- `open_questions`
- `next_route`
