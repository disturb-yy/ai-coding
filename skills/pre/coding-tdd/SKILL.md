---
name: coding-tdd
description: "Run test-first code changes in existing projects by combining TDD with coding-project. Use for red-green-refactor, regression fixes, integration-style tests, endpoint-to-response workflows, and requests to implement one small function or module at a time. Do not use for TDD explanation only, read-only test review, or general coding without a test-first requirement."
---

# Coding TDD

## Localization Maintenance

- When modifying this `SKILL.md`, update `SKILL.zh-CN.md` in the same change.
- `SKILL.zh-CN.md` is user-facing documentation only. Do not read or use it as model instructions, task context, or execution guidance.
- Treat this English `SKILL.md` as the model-readable source of truth.

## Purpose

Run a test-first coding workflow that composes `/tdd` and `/coding-project`.

- TDD owns the test cases, behavior slicing, red-green-refactor loop, and end-to-end verification shape.
- `coding-project` owns language-aware code edits, project conventions, dependency usage, security prechecks, and validation commands.
- Keep modules minimal. Implement one small module or one function per cycle unless the project structure requires an inseparable change.

## Required Composition

Use this order for every coding task:

1. Load `/coding-project` to observe the repository, detect language, load matching language and unit-test references, and identify validation commands.
2. Apply `/tdd` discipline to choose the smallest externally observable behavior to test first.
3. Write one focused failing test for that behavior before implementation.
4. Use `/coding-project` to implement only the smallest function or module needed to pass that test.
5. Run the narrowest relevant test command and confirm GREEN.
6. Refactor only while GREEN, then rerun the same test command.
7. Repeat for the next independent behavior or function.
8. After all modules are complete, write and run an end-to-end command that verifies the full entry-to-output path.

If any validation fails, stop forward progress, inspect the cause, fix it, and rerun the failed checkpoint before continuing.

## Behavior Slicing

Prefer vertical behavior slices over broad implementation plans.

```text
Request A -> function B -> function C -> function D -> Response E
```

Test in this shape:

| Scope | Test responsibility | Dependency rule |
|-------|---------------------|-----------------|
| Request A behavior | Verify the public request/API behavior and expected response contract. | Mock downstream dependencies that are outside this slice. |
| Function B | Verify B's behavior through its public interface. | Mock B's dependencies, such as C or external clients. |
| Function C | Verify C's behavior through its public interface. | Mock C's dependencies. |
| Function D | Verify D's behavior through its public interface. | Mock D's dependencies. |
| End-to-end path | Verify Request A produces Response E in the running app or closest project-supported environment. | Prefer real wiring; mock only unavailable external systems according to project conventions. |

Do not write all tests first and then all code. Write one test, implement one small target, validate, then continue.

## Parallel SubAgent Rule

Use subAgents only when subagent tooling is available and work units are independent. If no subagent tool is available, process the same units serially with the same boundaries.

| Situation | Action |
|-----------|--------|
| Functions have no direct dependency on each other and their tests can run independently | Start one subAgent per function or module when tooling is available. Each subAgent must use `/coding-project`, implement only its assigned target, and run its narrow validation command. |
| Function B depends on C's interface or behavior | Do not parallelize B and C until the contract is clear and stable. |
| Shared files, migrations, generated code, or public API contracts are involved | Keep work serial unless the subAgent boundaries are explicit and merge conflicts are unlikely. |
| Security-sensitive code is involved | Run `/coding-project` security precheck before assigning subAgents. |

Each subAgent must receive:

```text
Use /coding-project. Implement only <function/module>. The TDD test already defines the expected behavior. Keep the change minimal, follow project conventions, and run <targeted validation command>.
```

After subAgents finish, inspect their diffs together, resolve integration issues, and run the combined affected tests before end-to-end validation. Do not delegate shared API design, schema changes, or final integration decisions.

## Verification Checkpoints

Use fail-fast checkpoints:

1. After writing each test, run the narrow test and confirm it fails for the expected reason.
2. After implementing the target function or module, rerun the same test and confirm it passes.
3. After refactoring, rerun the same test.
4. After combining independent modules, run the affected package/module tests.
5. After all modules are complete, run an entry-to-output command.

For HTTP APIs, write the final `curl` command in a copyable form:

```bash
curl -i -X POST "$BASE_URL/path" \
  -H "Content-Type: application/json" \
  -d '{"example":"value"}'
```

Adjust method, headers, auth, and payload to match the project. Use documented local server commands or test environment setup from the repository before running `curl`.

For non-HTTP workflows, use the closest project-supported entry command:

```bash
<project command> <input-or-fixture>
```

State the expected observable output and verify it from stdout, generated files, database-visible behavior, or the project-supported inspection command.

## Decision Table

| User request | Use this skill? | Action |
|--------------|-----------------|--------|
| "Build this feature test-first" | Yes | Use `/coding-project`, then start the TDD loop. |
| "Fix this bug with a regression test first" | Yes | Write the failing regression test, then implement the smallest fix. |
| "Add integration tests around this endpoint" | Yes | Start with request-level behavior, then test and implement internal functions as needed. |
| "Explain TDD" | No | Answer conceptually or use `/tdd` if available for guidance only. |
| "Just implement this change" | No, unless the user asks for test-first | Use `/coding-project`. |
| "Review my test plan" | No | Review without editing unless asked to implement. |

## Examples

### Endpoint Feature

Input:

```text
Add POST /orders/quote test-first. It should return a quote total and reject empty carts.
```

Expected behavior:

```text
Use /coding-project to inspect routes, handlers, services, test patterns, and validation commands.
Write one failing request-level test for the valid quote path.
Implement only the handler/service function needed for that test.
Run the targeted test to GREEN.
Add the empty-cart test, implement the smallest validation change, and rerun.
Finish with a curl command that verifies POST /orders/quote returns the expected response.
```

### Independent Functions

Input:

```text
Request A calls normalizeCustomer, calculateDiscount, and formatResponse. They do not depend on each other. Build this TDD.
```

Expected behavior:

```text
Write focused tests for each public behavior with dependencies mocked.
If the functions are independent and touch separate files or stable contracts, assign one subAgent per function.
Each subAgent uses /coding-project, implements only its function, and runs the targeted test.
After merging, run the affected module tests and the request-level test.
Finish with the request-to-response curl command.
```

### Regression Fix

Input:

```text
Fix the null status bug red-green-refactor style.
```

Expected behavior:

```text
Use /coding-project to locate the failing path and test command.
Write one regression test that fails because null status is mishandled.
Implement the smallest function-level fix.
Run the targeted test to GREEN, refactor only if needed, and rerun.
Run the broader affected tests if the fix touches shared status mapping.
```

### Blocked Test Environment

Input:

```text
Add this endpoint with TDD, but there is no documented test command.
```

Expected behavior:

```text
Use /coding-project to search project docs, CI, package scripts, Makefiles, and build files for validation commands.
If no command exists, infer the narrowest standard command for the detected language and state the assumption.
If the test framework is missing or cannot run, report the blocker before implementation unless a small local test harness is appropriate for the project.
```
