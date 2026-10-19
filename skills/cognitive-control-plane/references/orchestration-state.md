# Coordination

Read before actual delegation or concurrent work. A single agent changing skills or phases needs no coordination record. A continuation handoff uses the fields in `SKILL.md`.

## Dispatch

Delegate only within current authorization and available capabilities. Give each worker a bounded result with:

- the objective and checkable success condition;
- source artifacts, relevant evidence, and accepted decisions;
- writable paths or an explicit read-only scope;
- dependencies, constraints, and any indispensable skill or tool;
- expected artifact or findings, validation, and conditions requiring escalation.

Plain prose is enough unless a host requires a schema. Use the real worker/task identifier returned by the host; a prepared prompt or contract is not a started worker. If optional delegation is unavailable, proceed serially. If independent review is required, follow [review rules](reviewer-enforcement.md).

## Ownership

Parallelize independent work with disjoint write ownership. Files can still conflict through a shared API, schema, or generated output: establish the shared contract first or serialize the dependent work. Discovery overlapping edits must identify the version it inspected; final review waits for a stable target.

Before replacing a failed or cancelled writer, inspect its partial changes. Cancellation is not rollback. Keep one owner for each integration decision.

## Reconcile

The parent retains responsibility for the user outcome. Inspect returned artifacts and validation against the assignment, resolve conflicts, and check affected integration behavior. A worker's completion message or list of skills used is not sufficient evidence.

Track task, owner, dependency, status, and result only as needed for actual concurrency. Persist state when interruption or a consumer requires it. Finish with one coherent result and identify any unresolved dependency.
