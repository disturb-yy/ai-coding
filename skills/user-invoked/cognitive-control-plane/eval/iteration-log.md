# Iteration Log

Record each eval-driven change.

```yaml
date: ""
change_id: ""
skill_version_before: ""
skill_version_after: ""
hypothesis: ""
failure_modes_targeted: []
cases_expected_to_change: []
cases_expected_not_to_change: []
before_run: ""
after_run: ""
score_delta:
  auto: 0.0
  hard_fail: 0
regressions: []
decision: keep | revert | inconclusive
notes: ""
```

```yaml
date: "2026-07-08"
change_id: "rewrite-routing-boundaries-v1"
skill_version_before: "baseline-v0 working tree before rewrite"
skill_version_after: "working tree after SKILL.md classification/route/trace rewrite"
hypothesis: "Replacing ambiguous classification, surface routing, orchestration, and trace rules in place will reduce baseline hard failures without adding more scattered rules."
failure_modes_targeted:
  - "Tiny/Small/Large boundary drift"
  - "first unsatisfied surface missed"
  - "orchestration not marked for Large delegated evidence or implementation work"
  - "required_skills implied in response but missing from trace"
  - "ownership_conflict used to mean detected conflict instead of unresolved conflict"
cases_expected_to_change:
  - ACP-005
  - ACP-006
  - ACP-007
  - ACP-008
  - ACP-010
  - ACP-108
  - ACP-109
  - ACP-113
  - ACP-114
  - ACP-203
  - ACP-210
  - ACP-304
  - ACP-306
cases_expected_not_to_change:
  - "static mirror checks"
before_run: "eval/results/baseline-v0/report.json: 73/101 assertions, 16 hard fails"
after_run: "eval/results/rewrite-targeted-2026-07-08-third-pass/report.json for remaining 4 cases: 8/12 assertions, 3 hard fails"
score_delta:
  auto: "targeted hard-fail subset improved from 20/36 pass in baseline-v0 equivalent failures to 8/12 pass on remaining cases after iterative rewrites"
  hard_fail: "targeted 13-case hard failures reduced from 16 baseline failures to 3 remaining hard failures"
regressions:
  - "None observed in static checks."
  - "Full 45-case behavioral run not completed after rewrite."
decision: inconclusive
notes: "Keep the rewrite for now. Runner bug extract_codex_error was fixed and --sandbox was made configurable. Remaining ACP-113 and ACP-306 failures are likely prompt/case ambiguity because the case states an artifact is complete/provided but does not include its content. ACP-007 behavior selects epistemic/current-doc research but still does not mark orchestration_used=true when agent-reach/read-only evidence gathering is the next step."
```
