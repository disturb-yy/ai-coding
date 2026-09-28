---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Skill Routing

Use this reference to select specialized procedure. Skill selection and agent allocation are separate decisions; use the current session unless delegation adds value and is authorized.

## Resolve Capabilities

Use the current skill catalog and the selected skill's actual scope. Prefer its canonical id (for example, `exploring-project` over `explore-project`) and resolve its path at runtime. The routes below are useful defaults, not an allowlist. An explicitly requested skill or a better matching available specialist takes precedence.

If a preferred capability is absent, use an available equivalent when it preserves the task's evidence and permission requirements. State material substitutions. When the user explicitly requires a particular skill or an unavailable capability is indispensable, report the gap and continue only independent work. A generic search tool is not a substitute for inaccessible project evidence.

## Select by Work, Not Keywords

| Need | Procedure | Ready to continue when |
|---|---|---|
| Extended requirements interrogation or stress-testing with the user | `grilling`; a single missing preference usually needs only a concise question. | Material decisions and acceptance criteria are clear. |
| Ambiguous conceptual or decision problem | `diagnosing-problem` | The interpretation, evidence needed, and next investigation are actionable. |
| Runtime error, test failure, performance regression, or code bug | An available debugging/root-cause skill | Evidence identifies the cause or the next discriminating experiment. Conceptual framing is not a prerequisite. |
| Repository behavior, code location, call chain, or safe change points | `exploring-project` | Target paths, relevant behavior, and nearby validation are verified. |
| Code, PR, diff, branch, or security review | `reviewing-code` | Findings are evidence-backed and tied to the inspected version. |
| Concrete non-code plan, design, prompt, or skill critique | [Adversarial control](adversarial-control.md) | Task-specific failures and mitigations are established. |
| Authorized existing-repository implementation | `coding-project`; `coding-tdd` when test-first is requested | Scoped changes and relevant validation are complete. |
| Skill authoring, frontend/UI, research, or another specialist domain | The matching available skill | Its task-specific completion criteria are met. |

Retrieve facts the workspace can answer before asking the user. Use `grilling`'s question cadence when that skill applies; do not impose its full interview on every incomplete request. Read only the selected route, not every skill in the table.

## Handoff to Execution

Carry forward the goal, known evidence, accepted decisions, scope, and validation needs. Do not require a second planning approval merely because the role or skill changes. Stop routing and perform the authorized action.

For a real delegated task, use the single [task contract](orchestration-state.md) rather than copying a template per role. Add a skill or reference only when the worker actually needs it. The parent reconciles the returned artifacts and evidence; a worker's statement that it followed a skill does not prove correctness.

When a caller needs named role responsibilities or an existing role-based handoff, consult the [role catalog](../role/index.md) and only the selected role. These optional role descriptions do not require delegation, fixed output fields, or a pause between authorized phases; current runtime rules and the consumer's explicit contract govern.

For the bundled host adapter only, use [the machine map](../config/skill-orchestration-map.yaml) and [contract schema](../adapters/contract.schema.json). Its explicit host profile is not the default conversational workflow. Resolve skills from the active environment before filling the contract.
