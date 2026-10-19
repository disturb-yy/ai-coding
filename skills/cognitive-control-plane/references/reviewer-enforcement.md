# Independent review

Use when the user, repository, or an explicitly enabled host policy requires an independent review. Identify the requirement and its scope. Consequential changes may justify recommending review; this skill does not create an automatic gate from file count or a risk label.

Check reviewer availability early. When required independence is unavailable, complete authorized preparation and validation, then report the artifact as awaiting review. Self-checking is useful validation but does not clear an independent-review requirement.

## Bind review to the artifact

The reviewer must differ from every actor that implemented or fixed the reviewed changes and must inspect the target read-only. A role switch for the same actor does not establish independence.

Identify the actual content: base/head commits plus a diff digest for a Git range, or a content digest/immutable artifact id. Include uncommitted files when in scope. A mutable path or branch name alone is insufficient.

Compare the delivered artifact with the reviewed identity. Changed content needs a new decision; review may focus on the delta and affected behavior. A review of an earlier version does not establish acceptance of a later one.

## Resolve findings

Findings need location, evidence, consequence, and a proposed correction. Separate defects from tool failures and unverified suspicions. A blocking finding demonstrates a material correctness, security, data-integrity, compatibility, or operational defect; style preferences alone do not block acceptance.

Fix accepted findings within authorized scope, validate, and obtain the required review of the changed target. If a finding recurs, investigate the failed correction before retrying. Stop the loop when an external input is indispensable or the applicable budget/termination requires it, and report the pending gate accurately.

For a required gate, retain its source, contributor/reviewer identities, artifact version, unresolved blocking findings, validation, and decision. Clear it only for the independently reviewed version with no unresolved blocking findings. Persist a gate record only when coordination or the host needs it.
