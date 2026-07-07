---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---

# Skill Orchestration

Use this reference when delegating work that depends on specialized skills. It defines the default route from task state to required skill and the task-contract shape specialists must receive.

Machine-readable map: [`skill-orchestration-map.yaml`](skill-orchestration-map.yaml)

## Skill Registry

| skill | role | source | path | use |
|---|---|---|---|---|
| `grilling` | requirements_interviewer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md` | Clarify requirements, plans, designs, or problem statements one question at a time until executable. |
| `diagnosing-problem` | problem_framer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md` | Frame ambiguous problems into answerable statements, assumptions, evidence standards, and handoffs. |
| `exploring-project` | codebase_explorer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md` | Explore project structure, behavior paths, modules, routes, functions, tests, and change points. |
| `coding-project` | implementation_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md` | Implement ordinary code changes, test changes, validation, generated artifacts, and language-aware project work. |
| `coding-tdd` | tdd_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md` | Execute test-first, red-green-refactor, regression-test-first, and behavior-sliced implementation. |
| `adversarial-control` | adversarial_reviewer | `file_reference` | `/home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md` | Attack concrete plans, designs, architecture, prompts, skills, implementation approaches, or agent traces before acceptance. |

`diagnosing-problem` frames the problem and handoff. Code navigation still routes to `exploring-project`. If a dedicated root-cause or runtime failure skill exists, route to it from the `diagnosing-problem` handoff.

## Routing Rules

Pick the first matching route:

| signal | required skill | delegate | next action |
|---|---|---|---|
| User goal, acceptance criteria, constraints, or business rules are unclear. | `grilling` | usually no | Ask one question at a time; if the repository can answer, explore instead of asking. |
| User asks why, what is wrong, how to locate a problem, or how to analyze a phenomenon, and the problem is not framed yet. | `diagnosing-problem` | optional read-only | Produce framed problem, assumptions, evidence standard, and handoff. |
| Existing project structure, entry point, call chain, route, module, function, tests, or change points are unclear. | `exploring-project` | optional read-only | Produce verified flow, candidate files, risks, and nearby tests. |
| A concrete plan, design, architecture, prompt, skill, implementation approach, diff, PR, or agent run needs critique. | `adversarial-control` | optional read-only | Produce criterion-based critique, valid failures, mitigations, and residual risks. |
| Requirements are clear and existing code or tests need ordinary edits. | `coding-project` | optional bounded write | Implement narrowly and validate. |
| User asks for TDD, test-first, red-green-refactor, or regression test first. | `coding-tdd` | optional bounded write | Run red -> green -> refactor for each behavior slice. |

## Default Pipeline

```text
intake
  -> grilling                   # requirements unclear
  -> diagnosing-problem          # problem or phenomenon not framed
  -> exploring-project           # project path or change points unclear
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

Exit when goal, acceptance criteria, scope, constraints, and next action are clear. If the repository can answer the question, route to `exploring-project` instead of asking the user.

### Diagnosing Problem

Enter when:

- The input is an ambiguous, open-ended, conceptual, strategic, decision, problem-location, or cause-analysis request.
- It is not clear whether the work should become code exploration, runtime diagnosis, external research, design discussion, or a direct answer.
- Another skill needs a problem statement, assumptions, evidence standard, or success criteria before acting.

Exit when the handoff has a framed problem, selected interpretation, rejected interpretations, load-bearing assumptions, evidence standard, and next action. If project navigation is required, route to `exploring-project`.

### Exploring Project

Enter when:

- The task needs project structure, routes, modules, call chains, functions, tests, or change points.
- Safe edit boundaries are unknown before coding.

Exit when candidate files, functions, routes, tests, risks, and key evidence are verified by source, tests, Graphify, CodeMap, or targeted search.

### Coding Project

Enter when:

- Requirements are clear.
- Existing repository code, tests, dependencies, generated artifacts, or implementation docs need edits.
- There is no explicit test-first requirement.

Exit when narrow changes are complete and relevant validation passed, failed for a concrete unrelated reason, or is blocked with a concrete cause.

### Adversarial Review

Enter when:

- A concrete plan, design, architecture, prompt, skill, implementation approach, diff, PR, or agent run needs critique.
- The user asks for review, risk analysis, pre-mortem, red-team, or asks whether an agent violated process.
- Failure modes matter more than producing new implementation.

Exit when the critique names its criteria, separates valid failures from weak attacks, gives mitigations, and leaves explicit residual risk.

### Coding TDD

Enter when:

- The user asks for TDD, test-first, red-green-refactor, or regression-test-first work.
- A visible behavior or small module can be sliced before implementation.

Exit when every slice has gone failing test -> minimal implementation -> green -> green-only refactor, and combined affected tests or final entry validation have run.

## SubAgent Contract Patterns

### Requirements Clarification

```yaml
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

### Adversarial Review

```yaml
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

If any check fails, do not accept the result directly; request completion, rerun, or verify independently.
