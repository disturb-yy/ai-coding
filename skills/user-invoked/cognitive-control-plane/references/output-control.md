---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Output Control

Use output control when the work is ready for handoff, implementation, or machine consumption.

## Stage Gate

Do not enforce a strict schema until delivery:

```text
Discovery -> free exploration
Synthesis -> structured judgment
Delivery -> strict contract
```

Completion criterion: the output format matches the current phase.

## Delivery Contract

Pick the smallest contract the next consumer needs:

```yaml
deliverable_type: ""
consumer: ""
required_fields: []
forbidden_content: []
validation: []
```

Common contracts:

- Implementation plan: files, changes, tests, risks
- ADR: decision, context, options, consequences
- Review: findings, severity, file references, open questions
- Handoff: goal, state, decisions, constraints, next steps
- Machine output: schema, required fields, validation rule

Completion criterion: the next consumer can use the output without reinterpretation.
