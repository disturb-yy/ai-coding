---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Adversarial Control

Use adversarial control when a plan is concrete enough to attack.

## Criterion-Based Critique

Choose criteria before critiquing:

```yaml
review_dimensions:
  - "Does it solve the real problem?"
  - "Does it duplicate existing capability?"
  - "Does it add unnecessary complexity?"
  - "Is the benefit verifiable?"
  - "Is there a simpler option?"
  - "Is maintenance cost acceptable?"
```

Completion criterion: criticism is tied to explicit criteria, not attitude.

## Red Team Review

Attack only the plan, then separate valid failures from noise:

```yaml
attack_report:
  valid_failures: []
  weak_or_irrelevant_attacks: []
  mitigations: []
  residual_risk: []
```

Completion criterion: the plan either changes, gains mitigations, or has explicit residual risk.

## Pre-Mortem

Prompt:

```text
Assume this plan failed six months after launch.
List the three most likely failure causes, the early warning signs that were missed, and the cheapest prevention step for each.
```

Completion criterion: each major failure mode has an early warning signal and prevention step.
