---
name: diagnosing-problem
description: Diagnose ambiguous problems before solving them. Use when the user asks to sharpen a problem statement, compare interpretations, expose assumptions, choose evidence standards, turn a vague request into a handoff, or diagnose a non-code conceptual or decision problem. Route code navigation to exploring-project and failures/root causes to problem-diagnosis.
---

# Diagnosing Problem

## Goal

Frame the problem. Convert ambiguity, hidden assumptions, and weak evidence
into one answerable problem statement, then stop with an answer, a narrowed problem, or
a handoff another skill can run without repeating the framing.

## Conceptual Space

```yaml
conceptual_space:
  target_region: "Framing unclear problems into answerable work: interpretation, assumptions, evidence standard, and next action."
  deviation_region:
    - "Answering from the first plausible interpretation when another interpretation would change the answer."
    - "Doing codebase exploration, bug diagnosis, or broad research before the problem is framed."
    - "Treating examples, analogies, search hits, or named concepts as proof."
    - "Producing a report when a framed problem or handoff is enough."
  priority_dimensions:
    - "Answerability before breadth."
    - "Explicit assumptions before confident prose."
    - "Evidence standard before evidence volume."
    - "Routing before doing another skill's work."
  entry_conditions:
    - "The request is ambiguous, conceptual, strategic, research-framing, decision-framing, or open-ended."
    - "Load-bearing terms need definitions before work can proceed."
    - "Another skill needs scope, assumptions, or evidence requirements before acting."
  exit_conditions:
    - "The selected interpretation and rejected alternatives are named when they matter."
    - "Load-bearing assumptions are accepted, tested, rejected, or marked unknown."
    - "The evidence standard has been met, downgraded, or named as missing."
    - "The result is an answer, narrowed problem, decision frame, research plan, or handoff."
  pre_output_check:
    - "Every important claim is evidence-backed, assumption-backed, or marked unknown."
    - "The output does not imply broader research, diagnosis, or code inspection than occurred."
    - "The next action is clear."
  sedimentation:
    - "Preserve the frame, assumptions, evidence standard, rejected interpretations, and open issues in handoffs."
    - "Do not create persistent docs unless the user or active workflow requests them."
```

## Steps

1. Frame. Restate the smallest answerable problem statement, name the expected output
   type, and pick a default interpretation if one is needed.
   Complete when the user request can be written as one problem statement plus, when
   relevant, one rejected interpretation.

2. Gate. Apply the gateway table below before gathering more context.
   Complete when each relevant gate is `pass`, `fail`, or `not_applicable`, and
   every failed gate has a next action.

3. Gather. Collect only the evidence needed to pass the failed or uncertain
   gates: user-provided material, local artifacts, official docs, primary
   sources, current web sources, or another skill's exploration.
   Complete when the evidence standard is met or the missing evidence is named.

4. Answer or hand off. Synthesize under the selected frame, with limits and
   confidence. If another skill should continue, emit the handoff shape instead
   of doing its work.
   Complete when the answer is bounded by its assumptions or the handoff has
   enough information for the next skill to start.

## Gateways

| Gateway | Pass condition | If it fails |
| --- | --- | --- |
| Routing | The task is problem framing, not code navigation, bug diagnosis, web research, wiki filing, or visual design. | Hand off to the matching skill with the current frame. |
| Scope | Problem statement, output type, and audience/use are clear enough to answer. | Ask 1-3 targeted questions or state a low-risk default. |
| Concept | Load-bearing terms have working definitions. | Define them, request definitions, or present alternative frames. |
| Assumption | Major assumptions are accepted, tested, rejected, or marked unknown. | List assumptions and identify which must be tested. |
| Evidence | Evidence quality matches claim risk, recency, and stability. | Gather targeted evidence, browse when required, or lower the claim to a hypothesis. |
| Falsification | For causal, diagnostic, strategic, or high-impact claims, at least one plausible alternative has been considered. | Add alternatives and what would distinguish them. |
| Handoff | The next agent or human can continue without reframing. | Include frame, assumptions, evidence, open issues, and success criteria. |

## Routing

Route before doing work that belongs elsewhere:

| Request shape | Route |
| --- | --- |
| Codebase structure, feature tracing, route/module/function location, or edit planning | `exploring-project` |
| Failure, crash, regression, flaky test, performance issue, or root cause analysis | `problem-diagnosis` |
| Internet research, public discussion lookup, URLs, or current external facts | Active web/search skill or tool policy |
| Wiki/vault answering, filing, or retrieval | Relevant wiki skill |
| Frontend visual planning or design contract | Relevant design skill |

When routing away, preserve the framed problem and assumptions.

## Evidence Standard

- Low-stakes conceptual clarification can rely on reasoning plus stated
  assumptions.
- Factual, current, source-specific, legal, financial, medical, policy, or
  spend-impacting claims need primary, official, current, or user-provided
  evidence.
- Search results, titles, examples, and analogies are leads, not proof.
- If evidence is missing, say what is missing and whether a provisional answer
  is still useful.

## Output

Use concise prose for ordinary answers:

```text
Framed problem: ...
Assumptions: ...
Answer: ...
Limits: ...
Next step: ...
```

Use JSON for handoffs or complex framing:

```json
{
  "artifact_type": "problem_framing",
  "created_at": "<ISO 8601 timestamp>",
  "content": {
    "original_problem": "",
    "framed_problem": "",
    "problem_type": [],
    "selected_interpretation": "",
    "rejected_interpretations": [],
    "assumptions": [
      {
        "statement": "",
        "status": "accepted|tested|unknown|rejected",
        "load_bearing": true,
        "evidence_or_reason": ""
      }
    ],
    "gateways": [
      {
        "name": "Routing|Scope|Concept|Assumption|Evidence|Falsification|Handoff",
        "status": "pass|fail|not_applicable",
        "notes": ""
      }
    ],
    "answer_status": "answered|narrowed|needs_evidence|needs_user_input|handoff",
    "answer_or_frame": "",
    "confidence": 0.0,
    "open_issues": [],
    "handoff": {
      "recommended_next_skill": "",
      "next_action": "",
      "success_criteria": "",
      "blocked_reason": ""
    }
  }
}
```

## Examples

- `Is this architecture scalable?`: frame "scalable" before answering
  throughput, latency, cost, operations, or team scale.
- `Why is the checkout test flaky?`: route to `problem-diagnosis` with the
  framed symptom and assumptions.
- `Should we adopt framework X?`: frame the decision, constraints, and evidence
  standard before research or recommendation.
