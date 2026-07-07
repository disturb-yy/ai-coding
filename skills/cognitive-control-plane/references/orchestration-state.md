# Orchestration State

Use orchestration state when work requires multiple agents, background tasks, parallel coding lanes, or staged reconciliation.

This is not a fifth control plane. It is the runtime layer that keeps routing, delegation, ownership, persistence, and reflection coherent.

## Scheduler-First

The orchestrator is not the default implementation worker.

Before work begins, decide whether the orchestrator should:

- ask a blocking question
- read minimal routing context
- delegate discovery, research, implementation, design, review, or media analysis
- run final verification directly
- synthesize terminal specialist outputs

Completion criterion: the orchestrator owns coordination and verification; specialists own bounded work when delegation adds clear value.

## Dependency Graph

Before delegating, identify:

- independent tasks that can run now
- dependent tasks that must wait
- files, folders, or subsystems each task owns
- outputs required before implementation
- verification required before final response

Completion criterion: no task is delegated without knowing whether it is independent, dependent, or blocked.

## Task Contract

Every delegated task must be self-contained:

```yaml
objective: ""
constraints: []
required_skills:
  - name: ""
    source: available_skill # available_skill | file_path | repo_skill | none
    path: ""
    required: true
    reason: ""
search_scope: []
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
edits_allowed: false
expected_output:
  format: ""
  required_fields: []
  must_report:
    - skills_loaded
    - skill_instructions_followed
    - deviations
validation: []
what_not_to_do: []
```

Completion criterion: the specialist can work without guessing scope, permissions, required skills, expected output, or validation.

## Skill Routing

When a specialist should use a skill, declare it in `required_skills`; do not rely on an implicit natural-language hint.

Rules:

- Use `available_skill` when the skill is already visible in the current environment's skill list.
- Use `file_path` when the specialist must read a specific `SKILL.md` path before acting.
- Use `repo_skill` when the skill is bundled inside the target repository.
- Use `none` only when no specialized skill is needed.
- Mark `required: true` only when skipping the skill would materially change the result.
- Require the specialist to report `skills_loaded`, whether instructions were followed, and any deviations.

Completion criterion: every delegated task that depends on specialized procedure names the skill, its source, why it is needed, and how the orchestrator will confirm it was used.

## Ownership Boundaries

Only one write-capable worker may own a file or subsystem at a time.

Rules:

- Parallel write tasks are allowed only when paths do not overlap.
- Read-only discovery can run in parallel with most work.
- Review tasks must wait for the work they review to reach terminal state.
- UI work that changes shared components must not overlap with implementation work on those components.
- Cancelling a writer is not rollback; inspect and reconcile partial changes before replacement.

Completion criterion: no two running write tasks can modify the same file, folder, or logical subsystem.

## Persistent State

Track delegated work as a small job board:

```yaml
tasks:
  - id: ""
    specialist: ""
    objective: ""
    state: running # running | completed | error | cancelled | timed_out
    required_skills: []
    skills_confirmed: []
    ownership:
      files: []
      areas: []
    dependencies: []
    result: ""
```

For long or high-risk work, persist state to a project-local markdown file or task artifact. Include:

- goal
- constraints
- phase plan
- task board
- decision trace
- review gates
- validation log
- unresolved risks

Completion criterion: the next turn can resume from explicit state instead of scattered conversation memory.

## Reconciliation

Specialist outputs are inputs, not final truth.

When a task completes:

1. Compare the result against the original user goal.
2. Check conflicts with other task outputs.
3. Check whether required skills were loaded and whether deviations are justified.
4. Decide whether to accept, revise, reject, or dispatch follow-up work.
5. Update the task board.
6. Preserve useful decisions in the next handoff.

Completion criterion: final work does not rely on unreviewed specialist output or unverified required-skill use.

## Conservative Reflection

Do not create a new skill, agent, command, rule, or playbook from a single interesting run.

Use this threshold:

- one useful insight -> save to notes or final summary
- repeated friction across 2-3 similar runs -> suggest project guidance or CLAUDE.md
- stable repeated workflow with clear triggers -> suggest a narrow skill
- speculative improvement with weak evidence -> create nothing

Completion criterion: process improvements are evidence-backed, minimal, and placed at the lowest durable layer that solves the problem.

## Verification Gate

Before final response:

- all required tasks are terminal
- dependent work consumed the outputs it waited for
- file ownership conflicts are resolved
- relevant checks ran, or skipped checks are explained
- residual risks are explicit

Completion criterion: the user receives a reconciled outcome, not a pile of agent reports.
