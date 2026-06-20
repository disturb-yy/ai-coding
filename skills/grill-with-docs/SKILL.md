---
name: grll-with-me
description: Facilitate rigorous requirement analysis through guided questioning, codebase and documentation exploration, domain-language clarification, scenario stress testing, scope shaping, acceptance criteria drafting, and decision capture. Use when the user wants Codex to analyze a product, business, technical, or implementation requirement before coding; clarify vague ideas; challenge assumptions; derive user stories, rules, edge cases, risks, open questions, or a requirement specification; or update project context docs as decisions become clear.
---

# Grll With Me

Use this skill to turn an unclear request into an implementable requirement. The style is a collaborative grilling session: ask precise questions, recommend answers, verify what can be verified from the project, and continuously sharpen language, scope, constraints, and acceptance criteria.

## Core Behavior

Ask one important question at a time and wait for the user's answer before moving to the next question. For each question, include a recommended answer or a small set of likely options so the user can react instead of starting from a blank page.

If a question can be answered by exploring the codebase or existing documentation, inspect the source instead of asking. Prefer evidence from files over speculation.

Keep the conversation grounded in concrete examples. Use scenarios, counterexamples, boundary cases, and failure cases to expose hidden assumptions.

## Context Scan

Before deep questioning, inspect available context when it exists:

- Existing requirements, tickets, specs, README files, `CONTEXT.md`, `CONTEXT-MAP.md`, architecture docs, ADRs, API docs, `index.md`, and coding standards.
- Source files, tests, schemas, routes, UI flows, data models, logs, or configs that already express the domain.
- Existing glossary terms and product/business names.

If a repo has `CONTEXT-MAP.md`, use it to find the relevant bounded context. If there is one `CONTEXT.md`, treat it as the primary glossary for the current context.

Do not invent domain rules, user roles, APIs, data shapes, or policy constraints. Mark unknowns explicitly.

## Requirement Analysis Loop

### 1. Frame The Request

Restate the requirement in one sentence. Identify:

- Goal and expected outcome.
- Primary actor or user.
- Triggering situation.
- Affected domain, system, module, or workflow.
- Whether this is discovery, implementation planning, bug clarification, or scope negotiation.

### 2. Sharpen Language

Challenge fuzzy or overloaded terms immediately. When a term conflicts with the project glossary, call out the conflict and ask which meaning is intended.

Examples:

- "You said account. Do you mean Customer, User, Organization, or billing account?"
- "`CONTEXT.md` defines cancellation as full-order cancellation, but this scenario sounds like item-level cancellation. Which model should hold?"

When a term becomes clear, propose a canonical wording.

### 3. Walk The Decision Tree

Explore the requirement branch by branch:

- Actors and permissions.
- Inputs, outputs, data ownership, and lifecycle.
- Happy path, alternate paths, edge cases, and failure modes.
- Business rules, invariants, validation, and policy constraints.
- Integration points, dependencies, and non-goals.
- Observability, audit, migration, compatibility, and rollout needs when relevant.

Resolve dependencies between decisions one by one. Avoid asking a large batch of questions unless the user explicitly requests a checklist.

### 4. Stress-Test With Scenarios

Invent concrete scenarios that probe boundaries:

- Minimum valid case.
- Maximum or high-volume case.
- Missing, stale, duplicate, conflicting, or unauthorized data.
- Race conditions, retries, rollback, cancellation, and partial completion.
- Existing-data compatibility and migration behavior.

Ask whether each scenario should be supported, rejected, delayed, or treated as unknown.

### 5. Produce Working Artifacts

As clarity emerges, maintain a concise working summary. Depending on the user's goal, produce one or more of:

- Problem statement.
- Scope and non-goals.
- Actors and permissions.
- Glossary terms.
- User stories or use cases.
- Functional requirements.
- Business rules and invariants.
- Edge cases and error behavior.
- Acceptance criteria.
- Open questions.
- Decision log.
- Implementation-ready handoff notes.

Make uncertainty visible. Separate confirmed facts from assumptions and open questions.

## Documentation Updates

When the user wants project docs updated, or when a decision should be captured immediately, update files with tools instead of only printing text.

Use `CONTEXT.md` only for glossary and domain language. Keep it free of implementation details, specs, scratch notes, and decision records.

Offer an ADR only when all are true:

1. The decision is hard to reverse.
2. A future reader would wonder why it was chosen.
3. Real alternatives were considered and a trade-off was made.

Create files lazily. Do not create `CONTEXT.md`, `docs/adr/`, or requirement docs until there is real content to preserve or the user asks for an artifact.

## Output Formats

For interactive analysis, use short conversational turns:

```markdown
My read: ...

Question: ...

Recommended answer: ...
```

For a requirement summary, prefer:

```markdown
# Requirement Summary

## Goal
## Context
## Scope
## Non-Goals
## Actors
## Rules
## Scenarios
## Acceptance Criteria
## Open Questions
## Decisions
```

For implementation handoff, add:

```markdown
## Impacted Areas
## Suggested Approach
## Validation Plan
## Risks
```

## Guardrails

Do not rush into implementation during requirement analysis unless the user explicitly asks to proceed. Do not ask questions that can be answered from available files. Do not bury the user in a questionnaire. Do not accept vague terms when they affect design or behavior. Do not record guesses as facts. Do not update documentation with unstable decisions unless clearly labeled or confirmed.
