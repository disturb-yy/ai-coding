---
name: cognitive-control-plane
description: "Resolve a blocked next action when task scope, evidence, a plan, memory-sensitive context, or coordination is unclear. Use for process diagnosis and handoffs; execute clear bounded tasks directly."
metadata:
  access:
    audience: model
    model_read: true
    model_write: true
    purpose: skill_runtime
---
# Cognitive Control Plane

Find the decision that blocks useful work, resolve it, then continue the user's task in the same session. Use the smallest intervention that establishes a checkable next action. Existing user authorization carries forward.

## Route the blocker

Read only the reference that applies:

| When the next action depends on… | Read | Proceed when… |
| --- | --- | --- |
| Missing goal, scope, artifact, or current state | [Context](references/context-control.md) | The target, permitted action, and success condition are clear. |
| An uncertain factual or causal claim | [Evidence](references/epistemic-control.md) | The claim is checked or its uncertainty is handled. |
| Critique of a concrete plan, design, prompt, or skill | [Critique](references/adversarial-control.md) | Material failures and proportionate corrections are identified. |
| A handoff or required output format | [Output](references/output-control.md) | The consumer can act from the deliverable. |

When configuration, user preferences, usage habits, or connection metadata may change the approach, follow [memory routing](references/memory-routing.md). The configured index is also named by global host instructions so relevant memory can be read without invoking this skill.

If a specialized procedure is needed, follow [skill routing](references/skill-orchestration.md). Applying a skill is not delegation. Use the current session unless an authorized independent task or a required separate review justifies another actor. For real delegation or a durable handoff, use [coordination](references/orchestration-state.md); for a required independent review, use [review rules](references/reviewer-enforcement.md).

Once the blocker is resolved, perform the authorized work and relevant validation. Reopen routing only when new evidence, scope, or authority changes the next action. Report the result and any material limitation.
