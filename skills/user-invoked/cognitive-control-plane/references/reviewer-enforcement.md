---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Reviewer Enforcement

Use this reference after implementation when risk determines whether independent review is mandatory, or when review findings, fixes, re-review, and artifact freshness must gate acceptance.

## Mandatory Review Policy

Create an independent `review` task after an implementation or fix task reaches a terminal state when the delivered change has one or more of these risk tags:

- `security_sensitive`
- `cross_module_change`
- `public_api_change`
- `schema_change`
- `migration`
- `auth_or_permission_change`
- `deployment_or_rollback_critical`

Assess risk from the actual delivered artifact, not only from the original request. A worker report that reveals a new trigger activates the policy. Passing implementation tests does not waive review.

The mandatory review task must:

- depend on the implementation or fix task it reviews
- use `phase: review` and `edits_allowed: false`
- require `reviewing-code` when that capability is available
- name the implementation task and actor
- bind a stable artifact version
- report blocking and non-blocking findings separately

Preflight `reviewing-code` and an independent reviewer actor before starting the task. If `reviewing-code` is unavailable, the only acceptable fallback is an `independent_read_only_reviewer`: a distinct host-launched read-only actor whose contract requires this reference, records `review_fallback: independent_read_only_reviewer`, and reports the unavailable skill as a deviation. If no such actor can start, leave the review gate blocked and hand off or terminate unaccepted. Do not downgrade to self-review.

Low-risk implementation may still receive optional review. Explicit user instructions may add review triggers but may not silently remove a mandatory trigger. If the user explicitly terminates the loop, report a terminated, unaccepted outcome with residual risk; do not represent it as reviewed acceptance.

Completion criterion: every terminal implementation artifact is risk-assessed, and every matching high-risk artifact has a pending or terminal independent review task.

## Review Report Format

Use a compact review report with two separate tables. Do not merge successful checks and defects into one undifferentiated list.

1. **Verification matrix**: `status`, `review lane/check`, `evidence`, and `result or limitation`. Include syntax, behavior, standards, security, tests, and skipped/blocked lanes as applicable.
2. **Findings table**: `id`, `severity`, `location`, `evidence`, `recommendation`, and `disposition`. State `no findings` explicitly when the table is empty.

Then give a short gate decision that names the immutable `review_target`, blocking-finding ids, residual risk, and whether the current version is `cleared`, `blocked`, or `terminated`. Tables make the final status scannable; keep investigation rationale in prose only when it materially explains a finding or limitation.

Completion criterion: a consumer can distinguish passed verification, skipped checks, blocking findings, non-blocking findings, and the gate decision without inferring them from prose.

## Reviewer Independence

Identify workers by stable `actor_id`, not by role label or prompt. For every review iteration:

```text
review.actor_id != reviewed_implementation.actor_id
```

The same actor may not implement, switch role, and review its own work. A distinct task id with the same actor id is not independent. The reviewer is read-only and must not also own a fix task for its findings. A reviewer may re-review a later version if that reviewer did not implement or fix that version.

If no independent reviewer can be started, keep the review gate blocked and emit a handoff or explicit termination state. Do not downgrade mandatory review to self-review.

Completion criterion: every accepted review records distinct implementation and reviewer actor ids.

## Artifact Version Binding

Pin each review to exactly one stable artifact version. Prefer this Git identity:

```yaml
review_target:
  kind: git_range
  base_sha: ""
  head_sha: ""
  diff_hash: ""
  stable_id: ""
```

Require `base_sha`, `head_sha`, and `diff_hash` for a Git diff. When Git identity is not available, use `kind: stable_artifact` and a non-empty `stable_id` such as an immutable build digest, generated bundle checksum, migration set digest, or content hash. A branch name, mutable file path, PR number without a head SHA, or prose such as "latest changes" is not stable enough.

Record the same target in the review task contract and review result. Before consuming the result, recompute or retrieve the current artifact identity and compare it with the reviewed identity.

Any artifact change after review invalidates the prior review, including a finding fix, cleanup edit, generated-file refresh, conflict resolution, rebase that changes content, or manual edit. Mark the old review `invalidated`; never carry its cleared gate forward to the new version.

Completion criterion: the accepted review target exactly equals the final delivered artifact version.

## Blocking Findings Gate

Use repository review severity rules when available. Otherwise treat findings as blocking when they show a material correctness, security, data integrity, permission, migration, public-API compatibility, deployment, or rollback defect that must be fixed before delivery.

When a reviewer reports one or more blocking findings:

1. Set the review gate to `blocked`.
2. Prevent final acceptance and delivery claims.
3. Dispatch a write-capable fix task that depends on the review task.
4. Include the accepted blocking findings and pinned reviewed version in the fix contract.
5. Keep non-blocking findings visible without silently expanding fix scope.

The orchestrator may reject an unsupported reviewer claim during reconciliation, but it must record the evidence-based rejection. An unresolved blocking finding is still blocking.

Completion criterion: no final acceptance path exists while an unresolved blocking finding applies to the current artifact lineage.

## Fix and Re-review Loop

After a fix task completes:

1. Compute the new artifact version.
2. Invalidate every prior review whose target differs from the new version.
3. Create a new independent review task for the new version.
4. Require the new reviewer actor to differ from the actor that implemented or fixed that version.
5. Reconcile the new findings.
6. Repeat the fix and re-review cycle while blocking findings remain.

The loop terminates only when:

- the latest artifact version has a valid independent review with no blocking findings; or
- the user or controlling policy explicitly terminates the loop.

Termination is not acceptance. Record the reason, unresolved findings, last reviewed version, current version, and residual risk.

Completion criterion: the latest version is independently cleared, or the run ends explicitly as terminated and unaccepted.

## Review Gate State Machine

```text
implementation_completed
  -> risk_assessed
  -> review_required? ---- no ----> verification
         |
        yes
         v
  review_pending -> review_running -> findings_reconciled
                                           |
                         no blocking ------+-----> cleared
                                           |
                         blocking ----------> fix_pending
                                                   |
                                                   v
                                              fix_completed
                                                   |
                                                   v
                                      prior_review_invalidated
                                                   |
                                                   v
                                             review_pending
```

At any point, an artifact-version change moves the completed review record to `invalidated` and creates or returns the current mandatory gate to `review_pending`. Retain historical invalidated review records as lineage evidence; only the gate for the current artifact version must become `cleared`. Explicit termination moves the current gate to `terminated`, which cannot satisfy final acceptance.

## Required Review State

Persist at least:

```yaml
review_gate:
  required: true
  risk_tags: []
  status: pending # pending | running | blocked | cleared | invalidated | terminated
  artifact_producer_task_id: "" # implementation or latest fix task
  artifact_producer_actor_id: "" # actor that produced the pinned version
  review_task_id: ""
  reviewer_actor_id: ""
  review_iteration: 1
  review_target:
    kind: git_range # git_range | stable_artifact
    base_sha: ""
    head_sha: ""
    diff_hash: ""
    stable_id: ""
  blocking_finding_ids: []
  supersedes_review_task_id: ""
  termination_reason: ""
```

Record transitions in the orchestration event log with evidence and artifact references.

## Final Acceptance Conditions

Before final acceptance, require all of the following:

- every implementation artifact was risk-assessed
- every mandatory review task used an actor independent of the reviewed implementation or fix actor
- the latest artifact version has a valid review target match
- the latest valid review has no unresolved blocking findings
- every blocking finding that was fixed was followed by re-review
- the mandatory review gate for the current artifact is `cleared`; historical superseded review records may remain `invalidated`

Completion criterion: acceptance refers to the independently reviewed final artifact, not an earlier version or an incomplete loop.
