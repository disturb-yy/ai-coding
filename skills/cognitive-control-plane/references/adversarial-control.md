---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Adversarial Control

Use when a concrete plan, design, prompt, skill, or approach needs critique. Code/diff reviews use the applicable code-review skill.

Choose criteria from the user goal and artifact: correctness, scope, evidence, duplication, complexity, or maintenance cost as relevant. Inspect the artifact and identify concrete failure mechanisms. For each material issue, show the triggering condition, consequence, supporting location/evidence, and a proportionate correction.

Separate demonstrated contradictions from plausible risks. Challenge an uncertain premise conditionally while verifying it; do not demand every assumption be settled before reviewing unrelated parts. Retain safeguards that have a concrete purpose, even when their removal would shorten the document.

Use a pre-mortem only if scenario analysis adds evidence: likely failure, early signal, prevention. There is no required number of attacks and no quota of findings. Omit weak objections rather than filling a report section with them.

Completion criterion: supported defects have actionable mitigations, speculative risks are labeled, and the proposed change preserves the artifact's useful invariants. If optimization is authorized, proceed to the scoped edits and validation.
