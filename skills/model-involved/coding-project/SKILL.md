---
name: coding-project
version: "1.0.0"
description: Use when an agent will edit an existing repository's source, tests, dependencies, generated artifacts, or implementation-facing project docs to implement features, fix bugs, refactor, update dependencies, add tests, or validate code changes. Use especially for language-aware Go or Java work. Do not use for read-only exploration, GitHub PR/CI triage, new-project scaffolding, product design without repo code edits, wiki/vault work, or skill/plugin authoring. Detect edited languages and load matching references before changing files.
---

# Coding Project

## Localization Maintenance

- When modifying this `SKILL.md`, update `SKILL.zh-CN.md` in the same change.
- `SKILL.zh-CN.md` is user-facing documentation only. Do not read or use it as model instructions, task context, or execution guidance.
- Treat this English `SKILL.md` and the referenced files under `references/` as the model-readable source of truth.

## Purpose

Run a tight coding loop in an existing project: observe the smallest useful context, load matching language references, decide a scoped plan, draft risky changes, precheck the plan, edit narrowly, and validate with the project toolchain.

## Role Contract

Act as the local [`implementation_worker`](role/implementation-worker.md) role copy. Read its linked
[handoff standard](../../role/handoff-standard.md) before editing. The role owns implementation
scope, ownership, stopping conditions, and final reporting; this skill owns the coding loop.
Finish with `changed_files`, `artifact_version`, `review_risk_tags`, `validation_commands`,
`validation_results`, and `residual_risks`. Preserve unrelated user changes and do not expand scope
without rerouting.

## Conceptual Space

```yaml
conceptual_space:
  target_region: "Source, test, dependency, generated-code, and project-doc changes inside an existing software repository, completed through a tight observe-orient-decide-draft-precheck-act-evaluate loop."
  deviation_region:
    - "Route read-only project discovery, architecture explanation, and implementation planning without edits to exploration or planning skills."
    - "Route GitHub PR review feedback, CI failures, issue management, and publishing to GitHub-specific workflows."
    - "Route new-repository scaffolding, product or UI design, wiki/vault maintenance, and skill/plugin authoring to their domain skills."
    - "Refuse broad cleanup, formatting churn, or new infrastructure unless it is required to satisfy the requested code change."
  priority_dimensions:
    - "Project-local conventions and loaded language references over generic preferences."
    - "Narrow, request-tied edits over opportunistic refactors."
    - "Validation evidence over unverified completion claims."
    - "Risk-aware drafting before touching public APIs, persistence, concurrency, security, migrations, or cross-module contracts."
  entry_conditions:
    - "The task requires editing code, tests, dependencies, generated artifacts, or implementation-facing project documentation in an existing repository."
    - "The task needs language-aware implementation or validation, especially in Go or Java."
  exit_conditions:
    - "Every edited file is tied to the request and follows local patterns."
    - "Relevant language references were loaded or no matching reference exists."
    - "The narrowest practical validation passed, failed for a shown unrelated reason, or is blocked by a concrete missing prerequisite."
  pre_output_check:
    - "Report changed files, validation commands, results, and any material skipped or substituted validation."
    - "Call out unresolved risk only when it affects the requested behavior or validation confidence."
  sedimentation:
    - "Leave durable project artifacts only when the request or behavior change requires them: source, tests, docs, manifests, lockfiles, generated output."
    - "Do not leave scratch plans, speculative TODOs, new conventions, or stale generated artifacts behind."
```

## Reference Route

Before editing code, detect the language from repository files and load only references that match files being changed:

| Editing target or signal | Load |
|--------------------------|------|
| Go source, `go.mod`, `go.sum`, or relevant `*.go` | `references/golang.md` |
| Go tests, `*_test.go`, or Go behavior that needs test coverage | `references/golang-ut.md` plus `references/golang.md` |
| Java source, `pom.xml`, `build.gradle`, `settings.gradle`, or relevant `*.java` | `references/java.md` |
| Java tests, `src/test/java`, or Java behavior that needs test coverage | `references/java-ut.md` plus `references/java.md` |

If several languages are edited, load each matching reference. If no matching reference exists, continue with project-local conventions and standard language knowledge. Do not invent a coding standard that is absent from the project and loaded references.

## Required Context Scan

Scan only the context needed for the requested change. The scan is complete when the affected files, local conventions, dependency/build surface, and likely validation commands are known.

1. User requirement and current repository state.
2. Project manifests, build files, dependency files, and lockfiles for the edited language.
3. Existing indexes, README files, architecture docs, and local coding standards when present.
4. Relevant source files, test files, generated code, and nearby examples.
5. Existing validation commands in CI config, package scripts, Makefiles, Gradle/Maven files, or project docs.

If an expected file is missing, state that briefly and continue from the closest available source.

## Execution Workflow

Use this sequence for implementation work:

| Step | Action | Completion criterion |
|------|--------|----------------------|
| **Observe** | Identify affected files, project structure, language, dependencies, and validation commands. | You can name the files or modules to inspect or change, and the first validation command to try. |
| **Orient** | Load the relevant language references and align with existing project patterns. | The loaded references match the edited language and test surface, or no matching reference exists. |
| **Decide** | Make a scoped plan with files to change and tests to run. | The plan excludes unrelated refactors and names the validation path. |
| **Draft** | Sketch non-trivial changes before editing APIs, data models, concurrency, persistence, security, migrations, or cross-module contracts. | Risky behavior, interfaces, data flow, and rollback or compatibility concerns are accounted for. |
| **Precheck** | Review the plan or draft for API fit, naming, error handling, dependency usage, test coverage, security, and blast radius. | No obvious mismatch remains between the planned change, local conventions, and loaded references. |
| **Act** | Edit files with narrow changes that follow local patterns. | Each edited file is tied to the request, and generated artifacts are changed only through project tools. |
| **Evaluate** | Run the most relevant validation commands. | Targeted validation passes, or any failure is fixed, shown unrelated, or reported as blocked with the concrete cause. |

For tiny mechanical edits, keep Draft and Precheck brief. For shared behavior, public APIs, migrations, or production-critical paths, make Draft and Precheck explicit before editing.

## Verification Checkpoints

Use fail-fast checkpoints instead of saving validation until the end. A checkpoint is complete only when its command has passed, failed for a known unrelated reason, or is blocked by a concrete missing prerequisite.

Resolve validation commands in this order:

1. Use commands documented in project docs, Makefiles, package scripts, CI config, Maven/Gradle files, or Go tooling wrappers.
2. If no project command exists, use the narrowest relevant command from the loaded language reference.
3. If the documented command is unavailable on the current platform, run the closest available equivalent and report the substitution.

| Change type | Checkpoint |
|-------------|------------|
| Dependency or import changes | Run the language package/build command; run dependency cleanup only when the project requires it or manifests changed. |
| Generated code or schema changes | Run the documented generator, inspect generated diffs, then edit dependent code. |
| Production logic changes | Run the narrowest relevant unit test first, then broaden to package/module validation. |
| Test-only changes | Run the targeted test with cache disabled or the project equivalent. |
| Cross-module/API changes | Run affected package/module tests and the broader build command when practical. |

If a checkpoint fails, stop the forward path, inspect the failure, fix the cause, and rerun the failed checkpoint before continuing.

## Security Precheck

Before editing security-sensitive code, check the relevant risk directly:

| Risk area | Required check |
|-----------|----------------|
| User input, request parameters, file paths, shell commands, or SQL | Validate inputs and use structured APIs, parameterized queries, or argument arrays instead of string concatenation. |
| Authentication, authorization, tenant boundaries, or permissions | Preserve existing access-control patterns and add regression coverage for allowed and denied paths when behavior changes. |
| Secrets, tokens, credentials, or personally identifiable data | Do not hardcode secrets, pass credentials on command lines, or add logs that expose sensitive values. |
| Network calls, retries, or timeouts | Follow project client patterns; use HTTPS and explicit timeouts when adding new outbound calls. |
| Migrations, destructive operations, or generated artifacts | Use documented project tools, inspect diffs, and require an explicit rollback or backup path when data can be changed or removed. |

## Implementation Rules

- Prefer project-local helpers, framework conventions, and existing abstractions over new infrastructure.
- Keep changes scoped to the user request; avoid unrelated refactors and formatting churn.
- Use structured parsers, compilers, generators, or framework tools when available instead of ad hoc text edits.
- Update tests when behavior changes or when the project already has relevant test coverage.
- Regenerate generated artifacts only with the project's documented generator commands.
- Do not hand-write common infrastructure already provided by the project, such as logging, metrics, retries, validation, authentication, database access, or dependency injection patterns.
- Keep each production-code function or method to at most 50 effective code lines. Exclude blank lines and comments; do not apply this limit to test code. Split an over-limit function by coherent responsibility rather than extracting arbitrary fragments.
- Do not leave validation failures unaddressed without explaining why they are unrelated or blocked.

## Examples

### Go feature with focused tests

Input:

```text
Add page and pageSize support to the Go ListUsers service and cover defaults and invalid input.
```

Expected behavior:

```text
Observe go.mod, service files, repository interface, and nearby *_test.go files.
Load references/golang.md and references/golang-ut.md.
Plan the smallest service/repository/test changes.
Draft the pagination handling before editing.
Run a targeted go test for the affected package, then broaden if practical.
```

Example shape:

```diff
- users, err := s.repo.List(ctx)
+ page, pageSize, err := normalizePage(req.Page, req.PageSize)
+ if err != nil {
+     return nil, err
+ }
+ users, err := s.repo.List(ctx, page, pageSize)
```

### Go test-only change with gomonkey

Input:

```text
Add unit tests for this Go function that calls time.Now and an external package function.
```

Expected behavior:

```text
Load references/golang.md and references/golang-ut.md.
Prefer a table-driven test.
Use gomonkey only if project-local injection or fakes are not available.
Clean up patches with t.Cleanup(patches.Reset).
Run gomonkey-dependent validation with go test -gcflags=all=-l ./... or the project-equivalent no-inline command.
```

### Java bug fix

Input:

```text
Fix the Java order status mapper when the upstream status is null and add a regression test.
```

Expected behavior:

```text
Observe pom.xml or build.gradle, mapper implementation, and nearby tests.
Load references/java.md and references/java-ut.md.
Use existing null-handling, assertion, and fixture patterns.
Run the targeted Maven or Gradle test, then the affected module test when practical.
```

### No matching language reference

Input:

```text
Fix a small Python CLI argument parsing bug in this mixed repository.
```

Expected behavior:

```text
No Python reference exists. Use project-local conventions, relevant files, and existing tests.
Do not invent a repository-wide Python standard.
Make the narrow fix and run the closest project-documented validation.
```

## Output

When reporting completion, include:

- Changed files and what changed.
- Validation commands run and pass/fail results.
- Any skipped or substituted validation with the concrete reason.
- Follow-up risks only when they are material to the requested task.
