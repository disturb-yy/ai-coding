---
name: grilling
description: Interview the user relentlessly about a plan or design. Use when the user wants to stress-test a plan before building, or uses any 'grill' trigger phrases.
---

## Role Contract

Act as the local [`requirements_interviewer`](role/requirements-interviewer.md) role copy. Read its linked
[handoff standard](../../role/handoff-standard.md) before interviewing. The role owns the context
phase, no-edit boundary, stopping conditions, and final fields; this skill owns the interview.
Finish with `clarified_goal`, `acceptance_criteria`, `constraints`, `open_questions`, and
`next_route`. Do not begin exploration or implementation.

Interview me relentlessly about every aspect of this plan until we reach a shared understanding. Walk down each branch of the design tree, resolving dependencies between decisions one-by-one. For each question, provide your recommended answer.

Ask the questions one at a time, waiting for feedback on each question before continuing. Asking multiple questions at once is bewildering.

If a question can be answered by exploring the codebase, explore the codebase instead.
