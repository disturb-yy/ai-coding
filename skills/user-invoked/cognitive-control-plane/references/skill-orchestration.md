---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Skill Orchestration

Use this reference when delegating work that depends on specialized skills. It defines the default route from task state to required skill and the task-contract shape specialists must receive.

Machine-readable map: [`../config/skill-orchestration-map.yaml`](../config/skill-orchestration-map.yaml)

## Skill Registry

| skill | role | source | path | use |
|---|---|---|---|---|
| `grilling` | requirements_interviewer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md` | Clarify requirements, plans, designs, or problem statements one question at a time until executable. |
| `diagnosing-problem` | problem_framer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md` | Frame ambiguous problems into answerable statements, assumptions, evidence standards, and handoffs. |
| `exploring-project` | codebase_explorer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md` | Explore project structure, behavior paths, modules, routes, functions, tests, and change points. |
| `reviewing-code` | code_reviewer | `available_skill` | `/home/jadon/projects/ai-coding/skills/user-invoked/reviewing-code/SKILL.md` | Review code changes, PRs, branches, diffs, and security-sensitive implementation artifacts for syntax, functionality, standards, and security issues. |
| `coding-project` | implementation_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md` | Implement ordinary code changes, test changes, validation, generated artifacts, and language-aware project work. |
| `coding-tdd` | tdd_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md` | Execute test-first, red-green-refactor, regression-test-first, and behavior-sliced implementation. |
| `adversarial-control` | adversarial_reviewer | `file_reference` | `/home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md` | Attack concrete plans, designs, architecture, prompts, skills, implementation approaches, or agent traces before acceptance. |

`diagnosing-problem` frames the problem and handoff. Code navigation still routes to `exploring-project`. If a dedicated root-cause or runtime failure skill exists, route to it from the `diagnosing-problem` handoff.

## Routing Rules

Pick the first matching route:

| signal | required skill | delegate | next action |
|---|---|---|---|
| User goal, acceptance criteria, constraints, business rules, roles, or interactions are unclear, or the user asks to make requirements clear. | `grilling` | usually no | Ask one question at a time; if the repository can answer, explore instead of asking. |
| User asks why, what is wrong, how to locate a problem, or how to analyze a phenomenon, and the problem is not framed yet. | `diagnosing-problem` | optional read-only | Produce framed problem, assumptions, evidence standard, and handoff; do not use for feature requirements clarification. |
| Existing project structure, entry point, call chain, route, module, function, tests, or change points are unclear. | `exploring-project` | optional read-only | Choose and record direct minimal exploration or a delegated read-only report; produce verified flow, candidate files, risks, and nearby tests. |
| Code changes, a PR, branch, diff, commit range, or security-sensitive implementation artifact needs code or security review. | `reviewing-code` | optional read-only | Produce a verification matrix plus severity-ranked findings table with evidence, recommendations, skipped checks, and residual risk. |
| A terminal implementation or fix artifact is security-sensitive, cross-module, public-API, schema/migration, auth/permission, or deployment/rollback-critical. | `reviewing-code` | mandatory independent read-only | Preflight reviewer capability, pin the artifact version, then run independent review before acceptance; use the explicit independent-read-only fallback only when the dedicated skill is unavailable. |
| A concrete plan, design, architecture, prompt, skill, implementation approach, diff, PR, or agent run needs critique. | `adversarial-control` | optional read-only | Produce criterion-based critique, valid failures, mitigations, and residual risks. |
| Requirements are clear and existing code or tests need ordinary edits. | `coding-project` | optional bounded write | Implement narrowly and validate. |
| User asks for TDD, test-first, red-green-refactor, or regression test first. | `coding-tdd` | optional bounded write | Run red -> green -> refactor for each behavior slice. |

## Default Pipeline

```text
intake
  -> grilling                   # requirements unclear
  -> diagnosing-problem          # problem or phenomenon not framed
  -> exploring-project           # project path or change points unclear
  -> reviewing-code              # code, PR, diff, branch, or security review
  -> adversarial-control          # concrete plan or agent output needs attack
  -> coding-tdd | coding-project # choose by test-first requirement
  -> verification
  -> ExecutionRun / VerifiedExperience / KnowledgeAsset
```

## Phase Gates

### Grilling

Enter when:

- The user gave direction without acceptance criteria.
- Business rules, roles, edge cases, or forbidden behavior can change implementation.
- The user is discussing a plan, design, or requirement rather than asking for immediate execution.
- The user asks to make requirements clear, or says rules, roles, interactions, or scope are not thought through.

Exit when goal, acceptance criteria, scope, constraints, and next action are clear. If the repository can answer the question, route to `exploring-project` instead of asking the user.

### Diagnosing Problem

Enter when:

- The input is an ambiguous, open-ended, conceptual, strategic, decision, problem-location, or cause-analysis request.
- It is not clear whether the work should become code exploration, runtime diagnosis, external research, design discussion, or a direct answer.
- Another skill needs a problem statement, assumptions, evidence standard, or success criteria before acting.

Do not enter for product or feature requirements whose rules, roles, interactions, or acceptance criteria are unclear; route those to `grilling`.

Exit when the handoff has a framed problem, selected interpretation, rejected interpretations, load-bearing assumptions, evidence standard, and next action. If project navigation is required, route to `exploring-project`.

### Exploring Project

Enter when:

- The task needs project structure, routes, modules, call chains, functions, tests, or change points.
- Safe edit boundaries are unknown before coding.

Exit when candidate files, functions, routes, tests, risks, and key evidence are verified by source, tests, Graphify, CodeMap, or targeted search.

### Code Review

Enter when:

- Code changes, a PR, branch, diff, commit range, or implementation artifact needs review before acceptance.
- The user asks for code review, security review, PR review, diff review, branch review, or review since a ref.
- Syntax, functional correctness, repository standards, or security findings matter more than implementation.

If the review target or diff base is unclear, use context control or `exploring-project` first. If the artifact is not code or the user wants a plan/design/prompt/agent trace attacked, route to `adversarial-control`.

For mandatory post-implementation review, read [`reviewer-enforcement.md`](reviewer-enforcement.md). The reviewer actor must differ from the actor that implemented or fixed the pinned version.

Exit when code/security findings are evidence-backed, severity-ranked, deduplicated, include skipped checks and residual risk, and the review result is bound to the current artifact version. A mandatory gate exits to acceptance only when the latest valid review has no blocking findings.

### Coding Project

Enter when:

- Requirements are clear.
- Existing repository code, tests, dependencies, generated artifacts, or implementation docs need edits.
- There is no explicit test-first requirement.

Exit when narrow changes are complete and relevant validation passed, failed for a concrete unrelated reason, or is blocked with a concrete cause.

### Adversarial Review

Enter when:

- A concrete plan, design, architecture, prompt, skill, implementation approach, or agent run needs critique.
- The user asks for review, risk analysis, pre-mortem, red-team, or asks whether an agent violated process.
- Failure modes matter more than producing new implementation.

Do not enter while the requested critique rests on an explicit unverified load-bearing assumption; route to epistemic control first.
Do not use adversarial review for ordinary code, PR, branch, diff, commit-range, or security review; route those to `reviewing-code`.

Exit when the critique names its criteria, separates valid failures from weak attacks, gives mitigations, and leaves explicit residual risk.

### Coding TDD

Enter when:

- The user asks for TDD, test-first, red-green-refactor, or regression-test-first work.
- A visible behavior or small module can be sliced before implementation.

Exit when every slice has gone failing test -> minimal implementation -> green -> green-only refactor, and combined affected tests or final entry validation have run.

## SubAgent Contract Patterns

### Requirements Clarification

```yaml
task_id: "requirements-1"
actor_id: "requirements-actor-a"
role: requirements_interviewer
phase: context
objective: "Clarify the user requirement until it is implementable."
required_skills:
  - name: grilling
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md
    required: true
    reason: "Requirements still affect the implementation route and need one-question-at-a-time clarification."
required_references: []
required_mcp: []
required_tools: []
edits_allowed: false
expected_output:
  format: clarification_state
  required_fields:
    - clarified_goal
    - acceptance_criteria
    - constraints
    - open_questions
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "The repository or source artifacts can answer the question better than the user."
```

### Problem Framing

```yaml
task_id: "problem-frame-1"
actor_id: "problem-framer-a"
role: problem_framer
phase: context
objective: "Frame the problem and produce a handoff."
required_skills:
  - name: diagnosing-problem
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md
    required: true
    reason: "The task needs selected interpretation, assumptions, evidence standard, and handoff before execution."
required_references: []
required_mcp: []
required_tools: []
edits_allowed: false
expected_output:
  format: problem_framing
  required_fields:
    - framed_problem
    - selected_interpretation
    - rejected_interpretations
    - assumptions
    - evidence_standard
    - handoff
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "The task becomes concrete implementation before the problem frame is accepted."
```

### Project Exploration

```yaml
task_id: "exploration-1"
actor_id: "explorer-a"
role: codebase_explorer
phase: context
objective: "Locate entry points, call chain, candidate change points, and nearby tests."
required_skills:
  - name: exploring-project
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md
    required: true
    reason: "The project path and safe edit boundary must be verified before coding."
required_references: []
required_mcp:
  - name: CodeMap or Graphify
    required: false
    reason: "Use when available for architecture, call-chain, or cross-area navigation."
required_tools:
  - name: rg
    required: true
    reason: "Fast source search is required to verify candidate files and flows."
edits_allowed: false
expected_output:
  format: change_point_report
  required_fields:
    - target
    - relevant_files
    - flow
    - leads_checked
    - risks
    - next_change_location
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "Implementation is requested before candidate files, flows, risks, and tests are verified."
```

### Code Review

```yaml
task_id: "review-1"
actor_id: "reviewer-actor-b"
role: code_reviewer
phase: review
objective: "Review code changes, PRs, branches, diffs, commit ranges, or security-sensitive implementation artifacts before acceptance."
review_of_task_id: "implementation-1"
review_of_actor_id: "implementation-actor-a"
review_iteration: 1
supersedes_review_task_id: ""
review_fallback: none
review_target:
  kind: git_range
  base_sha: ""
  head_sha: ""
  diff_hash: ""
  stable_id: ""
required_skills:
  - name: reviewing-code
    source: available_skill
    path: /home/jadon/projects/ai-coding/skills/user-invoked/reviewing-code/SKILL.md
    required: true
    reason: "Code and security review require syntax, functionality, standards, and security lanes with evidence-backed aggregation."
required_references:
  - name: reviewer-enforcement
    source: file_reference
    path: references/reviewer-enforcement.md
    required: true
    reason: "Mandatory review needs reviewer independence, artifact freshness, and final-gate rules."
required_mcp:
  - name: GitHub
    required: false
    reason: "Use when the review target is a PR, review thread, check run, linked issue, or remote diff."
  - name: CodeMap or Graphify
    required: false
    reason: "Use when call chains, route impact, related tests, or cross-area risk need navigation."
required_tools:
  - name: git diff or PR file list
    required: true
    reason: "The review target must be pinned before findings can be accepted."
  - name: rg
    required: false
    reason: "Use for targeted source, standard, test, and config evidence checks."
edits_allowed: false
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: code_review_report
  required_fields:
    - verification_matrix
    - findings_table
    - gate_decision
    - review_target
    - residual_risk
    - review_summary
    - review_target
    - gate_decision
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "The review target, diff base, PR, branch, commit range, or file list is unavailable."
  - "The review target is mutable or lacks a stable version identifier."
  - "The assigned reviewer actor matches the actor that implemented or fixed the target version."
  - "Required review evidence is inaccessible."
  - "The task becomes implementation instead of review."
```

If `reviewing-code` is unavailable, do not silently omit the review. A host may create an `independent_read_only_reviewer` contract only if it can start a distinct read-only actor. Set `review_fallback: independent_read_only_reviewer`, omit the unavailable skill from `required_skills`, retain the required `reviewer-enforcement` reference, and list the unavailable skill under `deviations`. Otherwise emit a handoff and leave the gate blocked.

### Mandatory Post-implementation Review

Apply this pipeline after every implementation and fix task:

```text
terminal implementation or fix
  -> assess delivered artifact risk
  -> mandatory trigger present?
       -> no: continue normal verification
       -> yes: create independent reviewing-code task
  -> pin artifact version
  -> reconcile review
       -> no blocking findings: clear gate for this exact version
       -> blocking findings: block final, dispatch fix
  -> fix completed
  -> invalidate prior review, pin new version, re-review
  -> repeat until cleared or explicitly terminated without acceptance
```

Mandatory triggers are `security_sensitive`, `cross_module_change`, `public_api_change`, `schema_change`, `migration`, `auth_or_permission_change`, and `deployment_or_rollback_critical`. Passing tests never converts mandatory review to optional review.

For each review iteration, enforce:

```yaml
review_invariants:
  reviewer_independent: "review.actor_id != review.review_of_actor_id"
  reviewer_read_only: true
  target_immutable: true
  stale_review_clears_gate: false
  blocking_findings_allow_acceptance: false
  fix_requires_rereview: true
```

If a host cannot launch an independent reviewer, emit a handoff and keep the mandatory gate blocked. If the loop is explicitly terminated, report the outcome as terminated and unaccepted with unresolved findings and residual risk.

Use `next_action: delegate_read_only` for this mandatory post-implementation task because a distinct actor must actually be started. Keep `next_action: route_skill` for an ordinary user-requested code review when the current actor is only handing control to the installed skill and independent delegation is not itself required.

### Adversarial Review

```yaml
task_id: "adversarial-review-1"
actor_id: "adversarial-reviewer-a"
role: adversarial_reviewer
phase: review
objective: "Attack the concrete plan, design, implementation approach, diff, PR, skill, prompt, or agent run before acceptance."
required_skills:
  - name: none
    source: none
    path: ""
    required: false
    reason: "The review depends on a control-surface reference rather than a standalone skill."
required_references:
  - path: /home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md
    required: true
    reason: "The specialist must use criterion-based critique, red-team separation, and pre-mortem structure."
required_mcp:
  - name: GitHub
    required: false
    reason: "Use when the artifact under review is a PR, issue, commit, or review thread."
required_tools:
  - name: rg
    required: false
    reason: "Use when local source, diffs, logs, or traces need targeted evidence checks."
edits_allowed: false
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: adversarial_review
  required_fields:
    - review_criteria
    - valid_failures
    - weak_or_irrelevant_attacks
    - mitigations
    - residual_risk
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "The artifact is too vague to attack with criteria."
  - "Required evidence is inaccessible."
  - "The review would require editing instead of critique."
```

### Ordinary Implementation

```yaml
task_id: "implementation-1"
actor_id: "implementation-actor-a"
role: implementation_worker
phase: implementation
objective: "Implement the confirmed code change and run validation."
required_skills:
  - name: coding-project
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md
    required: true
    reason: "Existing repository code must be changed using language, project convention, and validation rules."
required_references: []
required_mcp: []
required_tools:
  - name: project test/build commands
    required: true
    reason: "Validation must use the target project's own tooling."
edits_allowed: true
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: implementation_report
  required_fields:
    - changed_files
    - artifact_version
    - review_risk_tags
    - validation_commands
    - validation_results
    - residual_risks
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "Required skills or validation tools are unavailable."
  - "The change requires files outside ownership."
  - "Architecture, schema, migration, auth, payment, deployment, or user-visible behavior risk appears outside the contract."
```

### TDD Implementation

```yaml
task_id: "implementation-1"
actor_id: "implementation-actor-a"
role: tdd_worker
phase: implementation
objective: "Complete one visible behavior or small module with TDD."
required_skills:
  - name: coding-tdd
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md
    required: true
    reason: "The user requested test-first or red-green-refactor; the TDD loop must be protected."
required_references: []
required_mcp: []
required_tools:
  - name: project test commands
    required: true
    reason: "The red-green-refactor loop needs executable tests."
edits_allowed: true
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: tdd_report
  required_fields:
    - failing_test
    - implementation_slice
    - artifact_version
    - review_risk_tags
    - green_validation
    - refactor_validation
    - final_entry_to_output_check
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "A failing test cannot be made red for the intended behavior."
  - "The implementation scope expands beyond one behavior slice."
  - "Required validation tooling is unavailable."
```

## Parallelization Rules

- `grilling` is serial only. It asks one question at a time.
- `diagnosing-problem` may run beside read-only evidence collection, but the main agent must merge the final problem frame and handoff.
- `exploring-project` may parallelize read-only exploration across different modules; it must not write files.
- `reviewing-code` may parallelize read-only review lanes across code and security packs; it must not write files.
- Mandatory review starts only after the reviewed implementation or fix reaches terminal state, and its actor must be independent of the actor that produced that version.
- `coding-project` may write in parallel only when ownership paths and logical subsystems do not overlap.
- `coding-tdd` may parallelize only independent functions or modules; shared APIs, schemas, generated artifacts, migrations, and final integration stay serial.

## Reconciliation Checklist

Before accepting specialist output, check:

- `skills_loaded` includes every `required: true` skill from the task contract.
- `references_loaded` includes every `required: true` reference from the task contract.
- `mcp_used` and `tools_used` include every `required: true` capability from the task contract.
- `skill_instructions_followed` is reported.
- Any `deviations` are justified and do not break the task goal.
- The output includes every `expected_output.required_fields` item.
- Write ownership was respected.
- Validation results come from commands, tests, builds, user acceptance, or checkable evidence.
- Terminal implementation and fix reports include `artifact_version` and `review_risk_tags`, and every matching mandatory trigger created a review task.
- Review actor identity differs from the actor that implemented or fixed the pinned target version.
- The review result target exactly matches the current artifact version; changed artifacts invalidate prior reviews.
- Blocking findings prevent acceptance, dispatch fix work, and every completed fix is followed by re-review.

If any check fails, do not accept the result directly; request completion, rerun, or verify independently.
