---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Reviewer Enforcement

Use when deciding whether consequential implementation needs independent review, or when enforcing a review requirement from the user, repository, or configured host policy.

## Decide Whether Review Is Required

Identify the source and scope of the requirement. Security or authorization changes, tenant isolation, destructive migrations, public compatibility changes, and deployment/rollback behavior merit independent review when failure has material consequences. Assess actual behavior and impact; crossing two files or modules alone does not establish this risk.

A user, repository, or explicit host policy may require stricter review. The bundled host profile retains its mandatory risk tags in [the machine map](../config/skill-orchestration-map.yaml); load that map only when the host profile applies. This skill adds no authority above user instructions or host restrictions.

When independent review is required, check reviewer availability early. Prefer `reviewing-code` for code. If it is unavailable, a distinct read-only reviewer may use this reference and the project standards; record the substitution. If no independent reviewer can run, complete authorized implementation and checks that do not depend on review, then report the artifact as awaiting independent review. Do not label self-checking as independent review or claim a required gate cleared.

## Independence and Version Binding

A reviewer must differ from every actor that implemented or fixed the reviewed changes. Track contributors across the artifact lineage; checking only the latest fixer can accidentally approve an original implementer as reviewer. A different role or task id for the same actor is insufficient. Reviewers remain read-only; the implementer may fix accepted findings.

Pin the target with `base_sha`, `head_sha`, and `diff_hash` for a Git range, or a content digest/immutable `stable_id` for other artifacts. Include uncommitted content in the digest when it is in scope. A branch name or mutable path alone cannot identify reviewed content.

Before consuming the result, compare the final artifact with that identity. A change invalidates the claim that the new version is reviewed. The reviewer may examine the delta plus affected behavior instead of repeating unrelated checks, but must record a new decision for the new version.

## Findings and Completion

Keep verification evidence separate from defects. Findings identify location, severity, evidence, consequence, and recommended correction. Use tables only when the consumer requires them or they improve comparison; an empty review needs no empty matrix.

Reconcile unsupported claims against evidence. A blocking finding demonstrates a material correctness, security, data integrity, compatibility, or operational defect; a preference is not a blocking defect. Fix accepted findings within authorized scope, validate, then re-review the changed target when review is required. Passing tests complements independent review rather than proving it occurred.

Bound the loop by progress and the applicable budget. Continue while fixes or evidence make progress. If the same blocking finding returns after an attempted fix, investigate its cause before another iteration. If no useful action remains without unavailable authority/evidence, report the blocker and current artifact instead of spawning repeated reviews. Honor explicit termination; it is not acceptance.

For a required gate, record the policy source, contributor and reviewer identities, target version, blocking findings, validation, and decision (`pending`, `running`, `blocked`, `cleared`, `invalidated`, or `terminated`). Persist this state only when coordination or a host requires it. Clear the gate only for the exact independently reviewed version with no unresolved blocking findings.
