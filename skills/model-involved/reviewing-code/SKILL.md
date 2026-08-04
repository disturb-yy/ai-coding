---
name: reviewing-code
description: Review code changes, PRs, branches, diffs, or implementation artifacts for syntax, functional correctness, repository standards, and security. Use when the user asks for code review, security review, PR review, diff review, review since a ref, or wants findings aggregated from parallel review lanes before accepting code.
---

# Reviewing Code

## Localization Maintenance

The English files are canonical model-facing instructions. When changing this skill, update the matching Chinese mirror under `zh/` in the same change for human maintainers. Models and agents must not read `zh/` files as runtime instructions or task context; those files are human-readable mirrors only.

Run a scoped, evidence-backed review. Separate review lanes so syntax, behavior, standards, and security findings do not mask each other, then reconcile them into one ordered report.

## Role Contract

Act as the local [`code_reviewer`](role/code-reviewer.md) role copy. Read its linked
[handoff standard](../../role/handoff-standard.md) and
[reviewer enforcement](../../user-invoked/cognitive-control-plane/references/reviewer-enforcement.md)
before review. The role owns reviewer independence, the immutable target, no-edit boundary,
stopping conditions, and final fields; this skill owns the review method. Finish with the role's
verification matrix, findings, blocked/skipped checks, residual risk, pinned `review_target`, and
`gate_decision`. Do not review your own implementation.

## Conceptual Space

```yaml
conceptual_space:
  target_region: "Code changes or concrete implementation artifacts that need review before acceptance."
  deviation_region:
    - "General architecture critique without code or diff evidence."
    - "Implementation work after findings are accepted."
    - "Tool-only linting or SAST output without human review."
  priority_dimensions:
    - "Find material defects before style preferences."
    - "Separate evidence from speculation."
    - "Keep review lanes independent until aggregation."
    - "Verify claims against diff, source, tests, or referenced standards."
  entry_conditions:
    - "The user asks for code review, PR review, branch review, diff review, security review, or review since a ref."
    - "A concrete plan or implementation has code, diff, commit, PR, or files to inspect."
  exit_conditions:
    - "Every selected review lane has reported findings or an explicit no-findings result."
    - "Findings are deduplicated, severity-ranked, and tied to file/line, hunk, test, spec, or standard evidence."
    - "Residual risks and skipped lanes are explicit."
```

## Workflow

1. Frame the review target. Identify whether the input is a PR, branch/ref range, working tree diff, named files, or pasted code. If a diff base is needed and absent, ask for it; otherwise use the smallest available target.
   Completion criterion: the review target, diff command or file list, and any user-stated acceptance criteria are explicit.
2. Inventory evidence. Collect changed files, commits, nearby tests, repo standards, specs/issues/PRDs when available, and security-relevant surfaces such as auth, input handling, secrets, persistence, network calls, and dependency changes.
   Completion criterion: each selected lane has the files and evidence it needs, or the missing evidence is marked as a constraint.
3. Select review packs. Use `references/code/` for syntax, functionality, and standards review. Use `references/security/` when security-sensitive code, dependency/config changes, external inputs, credentials, permissions, data access, network calls, or the user asks for security review.
   Completion criterion: selected reference folders are named, and unselected folders have a reason.
4. Delegate review lanes. When subagents or subtasks are available, create one independent read-only subtask per selected reference folder; each subtask must read every file in that folder before reviewing. If a folder contains independent large checklists, split further by file. If subagents are unavailable, run the same lanes serially with separate notes.
   Completion criterion: every selected folder has one completed lane report, or every selected file has a completed split report.
5. Verify findings. Reject findings based only on vague suspicion. Confirm each material issue by reading the relevant source, diff hunk, test, config, standard, or spec. Use local tooling when it is available and relevant; report tool failures separately from review findings.
   Completion criterion: each retained finding has concrete evidence and a plausible failure mode.
6. Aggregate. Deduplicate overlapping findings, preserve the lane source, rank by severity, and output actionable review comments plus residual risk.
   Completion criterion: the user receives one review report with findings first, no pile of raw subtask reports.

## Reference Packs

- Read [`references/code/`](references/code/) for code review lanes: syntax analysis, functional analysis, and standards analysis.
- Read [`references/security/`](references/security/) for security review lanes: auth and access control, input and data handling, and secrets, dependencies, and supply chain.

Reference folders are review packs. A pack-level subtask must read every file inside its folder before producing findings. Split by file only when the folder is too large or the files map to independent specialists.

## Tool Roles

| Layer | Prefer | Good for | Must verify with |
| --- | --- | --- | --- |
| Diff and ownership | `git diff`, PR files, changed-file lists | Review scope, touched files, hunks, commits | Source reads and nearby tests |
| Code navigation and impact | CodeMap MCP when available | Call chains, routes, function impact, related tests | Source, tests, and diff hunks |
| Exact evidence | `rg`, source reads, tests, linters, typecheckers | Symbols, standards, failing behavior, syntax/type problems | The smallest relevant source or test set |
| External PR context | GitHub MCP when available | PR description, review threads, linked issues, checks | Local diff/source or fetched PR artifacts |

Tool output is evidence, not the review. A linter or scanner can support a finding, but the final report must explain why the issue matters.

## Review Lane Contract

Use this shape for each subagent or serial lane:

```yaml
role: code_review_lane
phase: review
objective: ""
review_pack: "references/code or references/security"
required_references:
  - "every file in the selected reference folder"
target:
  diff_command: ""
  files: []
  specs_or_standards: []
edits_allowed: false
expected_output:
  format: lane_report
  required_fields:
    - lane
    - evidence_checked
    - findings
    - no_findings_statement
    - skipped_checks
    - residual_risk
stop_if:
  - "The target diff or files are unavailable."
  - "Required reference files cannot be read."
  - "A finding would require guessing without source, test, spec, or standard evidence."
```

## Evidence Rules

- Cite file paths and line numbers when available. For PR-only artifacts, cite hunk, file, and commit or PR reference.
- Tie every issue to an observed behavior, broken contract, violated standard, unsafe data flow, or plausible exploit path.
- Separate hard failures from judgment calls. Syntax, type, broken tests, data leaks, and auth bypasses are hard failures when proven; style and design smells are judgment calls unless a repo standard makes them mandatory.
- Prefer existing repo standards over generic style rules. If the repo explicitly endorses a pattern, suppress generic objections to that pattern.
- Do not report generated files, vendored code, lockfile churn, or formatting-only changes unless they create a real defect or the user asks.
- Do not suggest rewrites in the review report unless the current code has a concrete defect or material risk.

## Severity

Use these labels consistently:

- `Critical`: exploitable security issue, data loss, auth bypass, secret exposure, destructive migration risk, or production outage path.
- `High`: likely functional failure, broken user-visible behavior, privilege/data boundary violation, or missing required behavior.
- `Medium`: edge-case bug, maintainability risk with near-term cost, incomplete validation, unsafe default, or standards violation that will cause defects.
- `Low`: local clarity, minor standards drift, weak tests, or low-risk cleanup.

## Output Shape

Lead with findings. Keep summaries secondary.

```text
Findings:
- Severity - file:line - title
  Evidence: ...
  Impact: ...
  Recommendation: ...
  Lane: code/functionality/security/etc.

No Findings:
- lane: what was checked

Skipped Or Blocked:
- ...

Residual Risk:
- ...

Review Summary:
- target reviewed
- references loaded
- tools used
```

If there are no issues, say so clearly and still name the lanes, evidence, and test/tool gaps.
