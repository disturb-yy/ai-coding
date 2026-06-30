---
name: coding-project
description: Coordinate implementation, refactoring, bug fixes, tests, dependency updates, and validation in an existing software project with a language-aware workflow, primarily for Go and Java. Use when Codex will edit project code or tests. Do not use for read-only project exploration, GitHub PR or CI triage, new-project scaffolding, or skill/plugin authoring. Detect the language from repository files and load matching references when available.
---

# Coding Project

## Localization Maintenance

- When modifying this `SKILL.md`, update `SKILL.zh-CN.md` in the same change.
- `SKILL.zh-CN.md` is user-facing documentation only. Do not read or use it as model instructions, task context, or execution guidance.
- Treat this English `SKILL.md` and the referenced files under `references/` as the model-readable source of truth.

## Purpose

Execute coding tasks in existing projects with a disciplined, language-aware workflow: observe project context, select the matching language reference, decide a small plan, draft the risky parts before editing, precheck against project and language conventions, act with tools, and evaluate with the project toolchain.

## Language Reference Selection

Before editing code, detect the primary language and load only the matching reference file when it exists:

| Signals | Reference |
|---------|-----------|
| `go.mod`, `go.sum`, `*.go` | `references/golang.md` |
| Go unit tests, `*_test.go`, test-only changes | `references/golang-ut.md` |
| `pom.xml`, `build.gradle`, `settings.gradle`, `*.java` | `references/java.md` |
| Java unit tests, `src/test/java`, test-only changes | `references/java-ut.md` |

If several languages are involved, load the reference for each language being edited. When adding or modifying tests, load the matching unit-test reference in addition to the language reference. If no matching reference exists, continue with project-local conventions and standard language knowledge. Do not invent a coding standard that is not present in either the project or the loaded reference.

## Required Context Scan

Scan only the context needed for the requested change:

1. User requirement and current repository state.
2. Project manifests, build files, dependency files, and lockfiles for the edited language.
3. Existing indexes, README files, architecture docs, and local coding standards when present.
4. Relevant source files, test files, generated code, and nearby examples.
5. Existing validation commands in CI config, package scripts, Makefiles, Gradle/Maven files, or project docs.

State missing expected files briefly and continue with the closest available source.

## Execution Workflow

Use this sequence for implementation work:

1. **Observe**: Identify affected files, project structure, language, dependencies, and validation commands.
2. **Orient**: Load the relevant language reference and align with existing project patterns.
3. **Decide**: Make a scoped plan with files to change and tests to run.
4. **Draft**: Sketch the non-trivial code changes before editing when the change affects APIs, data models, concurrency, persistence, security, or cross-module contracts.
5. **Precheck**: Review the plan or draft for API fit, naming, error handling, dependency usage, test coverage, and blast radius.
6. **Act**: Edit files with narrow changes that follow local patterns.
7. **Evaluate**: Run the most relevant validation commands. On failure, inspect the cause, adjust the implementation, and rerun targeted validation.

For tiny mechanical edits, keep the draft and precheck concise. For shared behavior, public APIs, migrations, or production-critical paths, make the draft and precheck explicit.

## Verification Checkpoints

Use fail-fast checkpoints instead of saving validation until the end:

Resolve validation commands in this order:

1. Use commands documented in project docs, Makefiles, package scripts, CI config, Maven/Gradle files, or Go tooling wrappers.
2. If no project command exists, use the narrowest relevant command from the loaded language reference.
3. If the documented command is unavailable on the current platform, run the closest available equivalent and report the substitution.

| Change type | Checkpoint |
|-------------|------------|
| Dependency or import changes | Run the language package/build command and dependency cleanup command only if the project requires it. |
| Generated code or schema changes | Run the documented generator, then inspect generated diffs before editing dependent code. |
| Production logic changes | Run the narrowest relevant unit test first, then broaden to the package/module test. |
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
- Validation commands run and their results.
- Any skipped validation with the concrete reason.
- Follow-up risks only when they are material to the requested task.
