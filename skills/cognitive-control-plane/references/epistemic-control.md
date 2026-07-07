---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---

# Epistemic Control

Use epistemic control when wrong assumptions would dominate the outcome.

## Assumption Audit

```yaml
assumptions:
  - statement: ""
    confidence: 0.0
    evidence: []
    falsification: []
```

Prioritize assumptions that are:

- Load-bearing for the plan
- Weakly evidenced
- Cheap to verify
- Expensive if wrong

Completion criterion: every load-bearing assumption has evidence, uncertainty, or a falsification path.

## Decision Trace

Use decision trace instead of exposed chain-of-thought:

```yaml
conclusion: ""
evidence: []
assumptions: []
uncertainty: []
alternative_hypotheses: []
verification: []
```

Completion criterion: a reviewer can understand why the decision is reasonable without seeing hidden reasoning.
