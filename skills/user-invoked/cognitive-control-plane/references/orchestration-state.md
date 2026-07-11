---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
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

For Large implementation work, the orchestrator must delegate implementation. Direct implementation is allowed only when every condition in the Implementation Guard is met; otherwise delegation is mandatory.

Completion criterion: the orchestrator owns coordination and verification; specialists own bounded work when delegation adds clear value.

## Read-only Exploration Strategy

For Large repository discovery, decide before broad source reading whether the orchestrator will run **direct minimal exploration** or delegate one or more **read-only exploration** tasks. `exploring-project` names the required procedure; it does not automatically require a subagent.

Use direct minimal exploration only for a small, dependent evidence chain. Record the bounded search scope, sources checked, and why a separate worker or parallel lane would not materially improve the next decision. Delegate read-only exploration when the search crosses independent areas, benefits from an independent evidence report, or has enough breadth that parallel lanes reduce uncertainty. Do not claim a route to `exploring-project` while silently skipping its procedure.

Completion criterion: the task board or trace records `direct_minimal_exploration` or `delegated_read_only_exploration`, the evidence scope, and the reason for the choice.

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
task_id: ""
actor_id: ""
role: ""
phase: context | design | implementation | review | verification
objective: ""
review_of_task_id: "" # required for review phase
review_of_actor_id: "" # required for review phase
review_iteration: 0
supersedes_review_task_id: ""
review_target: # required for review phase
  kind: none # none | git_range | stable_artifact
  base_sha: ""
  head_sha: ""
  diff_hash: ""
  stable_id: ""
constraints: []
required_skills:
  - name: ""
    source: available_skill # available_skill | file_path | repo_skill | none
    path: ""
    required: true
    reason: ""
required_references:
  - path: ""
    required: false
    reason: ""
required_mcp:
  - name: ""
    required: false
    reason: ""
required_tools:
  - name: ""
    required: false
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
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
validation: []
stop_if: []
```

Completion criterion: the specialist can work without guessing task or actor identity, role, phase, scope, permissions, required skills, required references, required MCP/tools, review lineage and target, expected output, stop conditions, or validation.

## Skill Routing

When a specialist should use a skill, declare it in `required_skills`; do not rely on an implicit natural-language hint. For the default skill routing map, read [`skill-orchestration.md`](skill-orchestration.md).

Rules:

- Use `available_skill` when the skill is already visible in the current environment's skill list.
- Use `file_path` when the specialist must read a specific `SKILL.md` path before acting.
- Use `repo_skill` when the skill is bundled inside the target repository.
- Use `none` only when no specialized skill is needed.
- Mark `required: true` only when skipping the skill would materially change the result.
- Require the specialist to report `skills_loaded`, whether instructions were followed, and any deviations.

Completion criterion: every delegated task that depends on specialized procedure names the skill, its source, why it is needed, and how the orchestrator will confirm it was used.

## Reference Routing

When a specialist depends on a control-surface file, design guide, project guide, ADR, schema document, or other non-skill reference, declare it in `required_references`.

Rules:

- Use `required_references` for files that must be read but are not standalone skills.
- Mark `required: true` when skipping the reference would materially change the result.
- Require the specialist to report `references_loaded` and deviations.
- If a required reference is unavailable, the specialist must stop instead of reconstructing it from memory.

Completion criterion: every delegated task that depends on non-skill reference material names the file, why it is needed, and how the orchestrator will confirm it was used.

## MCP and Tool Routing

When a specialist depends on a connector, MCP server, code navigation tool, search tool, browser, build tool, or framework command, declare it in `required_mcp` or `required_tools`.

Rules:

- Use `required_mcp` for named MCP/connectors such as GitHub, CodeMap, Graphify, web readers, issue trackers, or database tools.
- Use `required_tools` for shell commands, language tooling, test runners, formatters, package managers, browsers, or local CLIs.
- Mark `required: true` when skipping the capability would materially change correctness.
- Require the specialist to report `mcp_used`, `tools_used`, and deviations.
- If the required capability is unavailable, the specialist must stop instead of substituting an unapproved path.

Completion criterion: every delegated task that depends on external capability names the capability, why it is needed, and how the orchestrator will confirm it was used.

## Ownership Boundaries

Only one write-capable worker may own a file or subsystem at a time.

Rules:

- Parallel write tasks are allowed only when paths do not overlap.
- If a user requests overlapping write-capable workers, reject parallel writes, serialize the tasks, and resolve the conflict before delegation.
- Read-only discovery can run in parallel with most work.
- Review tasks must wait for the work they review to reach terminal state.
- A review actor must differ from the actor that implemented or fixed the reviewed version. A new task id or role name does not make the same actor independent.
- A review worker must stay read-only and must not own the fix task for its findings.
- UI work that changes shared components must not overlap with implementation work on those components.
- Cancelling a writer is not rollback; inspect and reconcile partial changes before replacement.

Completion criterion: no two running write tasks can modify the same file, folder, or logical subsystem, and no accepted review is self-review.

## Reviewer Enforcement Loop

After each implementation or fix task becomes terminal, read [`reviewer-enforcement.md`](reviewer-enforcement.md) and assess the actual artifact for these mandatory-review tags:

- `security_sensitive`
- `cross_module_change`
- `public_api_change`
- `schema_change`
- `migration`
- `auth_or_permission_change`
- `deployment_or_rollback_critical`

When any tag applies, preflight `reviewing-code` and an independent reviewer actor. When available, create a dependent read-only `reviewing-code` task. If the skill is unavailable, use the `independent_read_only_reviewer` fallback only when the host can start a distinct read-only actor that loads [`reviewer-enforcement.md`](reviewer-enforcement.md) and reports the unavailable skill as a deviation. Otherwise keep the review gate blocked and emit a handoff; never substitute self-review. Record distinct implementation and reviewer `actor_id` values and pin `review_target` with `base_sha`, `head_sha`, and `diff_hash`, or with an equivalent immutable `stable_id`.

Reconcile findings before advancing the gate:

- no blocking findings on the current pinned version -> mark the review gate `cleared`
- blocking findings -> mark it `blocked`, prevent final acceptance, and dispatch a dependent fix task
- fix completed or any other artifact change -> mark the previous review `invalidated`, compute the new version, and create a new independent review task
- explicit loop termination -> mark it `terminated` and report an unaccepted result with residual risk

Repeat until the latest artifact version is independently reviewed with no blocking findings. Tests and verification complement review; they do not waive mandatory review.

Completion criterion: every mandatory review gate is cleared for the current artifact version before acceptance, or the run ends explicitly as terminated and unaccepted.

## Persistent State

Track delegated work as a small job board:

```yaml
tasks:
  - id: ""
    actor_id: ""
    specialist: ""
    phase: ""
    objective: ""
    state: pending # pending | running | completed | error | cancelled | timed_out
    required_skills: []
    skills_confirmed: []
    required_references: []
    references_confirmed: []
    required_mcp: []
    mcp_confirmed: []
    required_tools: []
    tools_confirmed: []
    ownership:
      files: []
      areas: []
    dependencies: []
    risk_tags: []
    review_required: false
    review_of_task_id: ""
    review_of_actor_id: ""
    review_iteration: 0
    review_status: not_required # not_required | pending | running | blocked | cleared | invalidated | terminated
    review_target:
      kind: none # none | git_range | stable_artifact
      base_sha: ""
      head_sha: ""
      diff_hash: ""
      stable_id: ""
    blocking_finding_ids: []
    supersedes_review_task_id: ""
    result: ""
event_log:
  - timestamp: ""
    actor: "" # orchestrator | specialist role
    task_id: ""
    event_type: started # started | blocked | completed | decision | validation | handoff
    summary: ""
    evidence_refs: []
    artifact_refs: []
    next_action: ""
```

For long or high-risk work, persist state to a project-local markdown file or task artifact. Include:

- goal
- constraints
- phase plan
- task board
- decision trace
- review gates
- event log
- validation log
- unresolved risks

Completion criterion: the next turn can resume from explicit state instead of scattered conversation memory.

## Reconciliation

Specialist outputs are inputs, not final truth.

When a task completes:

1. Compare the result against the original user goal.
2. Check conflicts with other task outputs.
3. Check whether required skills were loaded and whether deviations are justified.
4. Check whether required references were loaded and whether deviations are justified.
5. Check whether required MCP/tools were used and whether deviations are justified.
6. For implementation and fix results, assess mandatory-review risk from the delivered artifact.
7. For review results, confirm actor independence and exact artifact-version match before consuming findings.
8. If blocking findings remain, block acceptance and dispatch a fix; if a fix changed the artifact, invalidate prior review and dispatch re-review.
9. Decide whether to accept, revise, reject, dispatch follow-up work, or terminate without acceptance.
10. Update the task board and review gate.
11. Preserve useful decisions in the next handoff.

Completion criterion: final work does not rely on unreviewed specialist output, self-review, stale review, unresolved blocking findings, unverified required-skill use, unverified required-reference use, or unverified required-capability use.

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
- required skills, references, MCP, and tools were confirmed or deviations were accepted explicitly
- file ownership conflicts are resolved
- every implementation artifact was assessed against mandatory review triggers
- every mandatory reviewer actor differs from the actor that implemented or fixed the reviewed version
- the latest artifact version exactly matches a valid pinned review target
- the mandatory review gate for the current artifact is `cleared`; historical superseded review records may remain `invalidated`
- every blocking-finding fix was followed by re-review and the latest valid review has no blocking findings
- relevant checks ran, or skipped checks are explained
- residual risks are explicit

Completion criterion: the user receives a reconciled outcome, not a pile of agent reports.
