---
name: go-coding
description: >
  Implement, refactor, or fix Go code with a disciplined OODA-E workflow:
  observe project context, orient to Go conventions, decide a plan, draft
  before editing, precheck like a reviewer, act with tools, and validate
  in a self-correcting loop. Workflow-compatible skill for the
  Implementation phase. Use when writing or modifying Go code -- implementing
  features, fixing bugs, refactoring, writing tests, or updating Go module
  dependencies.
---

# Go Coding

## Purpose

Execute Go coding tasks with a disciplined OODA-E pipeline: observe project
context via go.mod and indexes, orient to Go idioms and project conventions,
decide a scoped plan, draft in a brain sandbox before touching files, precheck
against a quality gate, act with editing tools, and evaluate with
`go build` / `go vet` / `go test` / `golangci-lint` in a self-correcting loop
until clean. Produces an `implementation_result` artifact.

## Input

- **Task description**: Feature request, bug report, refactoring goal, or
  test spec (required).
- **Go project context**: `go.mod`, project indexes, source files (required).
- **Previous artifact** (optional): prior `solution_design` or
  `project_understanding` artifact to scope implementation.

## Output

Artifact envelope (returned as the primary output):

```json
{
  "artifact_type": "implementation_result",
  "created_at": "<ISO 8601 timestamp>",
  "phase": "Implementation",
  "content": {
    "modified_files": ["..."],
    "created_files": ["..."],
    "change_summary": ""
  }
}
```

Implementation output structured in XML phases: `<observe>`, `<orient>`,
`<decide>`, `<draft>`, `<precheck>`, `<act>`, `<evaluate>`.

## Responsibilities

- Scan project context: go.mod, indexes, source files, conventions docs.
- Align implementation with Go idioms and existing project patterns.
- Draft Go code in text before editing files (brain sandbox draft).
- Precheck drafts against a 5-point quality gate before touching files.
- Apply changes with editing tools; regenerate proto stubs when needed.
- Validate with go build/vet/test/golangci-lint in a self-correcting loop.
- Produce an `implementation_result` artifact listing changed files.

## Forbidden Actions

- Do NOT decide workflow phase transitions or modify lifecycle state.
- Do NOT call other skills directly or persist hidden state.
- Do NOT invent APIs, business rules, coding standards, or module boundaries.
- Do NOT hand-write patterns that exist in the standard library or project
  infrastructure (HTTP client, retries, logging, metrics, caching, rate
  limiting, auth, validation, connection pooling, graceful shutdown).
- Do NOT skip the draft or precheck phases.
- Do NOT leave validation failures unaddressed.

## Execution Rules

- Execute phases in order: Observe, Orient, Decide, Draft, Precheck, Act,
  Evaluate. Self-correct on validation failure by returning to the appropriate
  earlier phase.
- Use previous artifacts instead of chat history for context.
- Read only Go files required for the implementation scope.
- Prefer narrow, idiomatic changes over broad refactoring.
- Keep the skill usable by both Codex direct execution and OpenCode delegated
  execution.
- Produce structured artifact output, not free-form text.

---

Use this skill for Go coding tasks that need disciplined context gathering,
alignment with Go idioms and project conventions, draft-first implementation,
tool-based edits, and validation. The operating principle: observe the Go
project, orient to its module structure and coding patterns, decide a small
plan, draft before editing, precheck like a reviewer, act with tools, and
evaluate with `go build`, `go vet`, `go test`, and `golangci-lint` until
correct.

For the Go conventions checklist and toolchain command reference, see
`references/go-conventions.md`.

## Required Context Scan

Before designing or editing Go code, scan the project context:

1. User requirement and current repository state.
2. `go.mod` -- module path, Go version, direct and indirect dependencies.
3. Existing index files: `index.md`, module indexes, API indexes,
   infrastructure indexes.
4. Project README, architecture docs, and coding standards.
5. Business-specific docs when present.
6. Relevant Go source files, test files (`*_test.go`), and generated code.

Do not skip scanning when relevant files exist. If an expected file is
missing, state that and continue with the closest available source. Never
invent APIs, business rules, coding standards, or module boundaries.

## Output Contract

Structure visible reasoning in this exact XML order:

```xml
<observe>
</observe>
<orient>
</orient>
<decide>
</decide>
<draft>
</draft>
<precheck>
</precheck>
<act>
</act>
<evaluate>
</evaluate>
```

## Complete Example

**User request:** "Add pagination to the ListUsers RPC -- page and pageSize
int32 params, default page 1, max pageSize 100."

```xml
<observe>
Task: Add pagination params to ListUsers RPC in user service.
Affected: internal/service/user.go, api/proto/user.proto,
internal/service/user_test.go.

Scanned go.mod: module github.com/company/platform, Go 1.22.
Scanned index.md -> shared Pagination type in pkg/pagination/.
Scanned infrastructure index -> Pagination helper in pkg/pagination:
func BuildMeta(total int64, page, pageSize int32) Meta.

Existing patterns: service methods accept (ctx, req) -> (*Resp, error).
Test files use testify/assert with table-driven tests.
</observe>

<orient>
Naming: page, pageSize (proto snake_case -> Page, PageSize in Go).
Error handling: invalid params -> return nil, status.Error(codes.InvalidArgument, ...).
Module path: internal imports use github.com/company/platform/... prefix.
Contracts: response must include pagination metadata alongside data.
Infrastructure: use pkg/pagination.BuildMeta(), do not hand-write page math.
Scope: only ListUsers RPC -- proto message, service handler, and tests.
</orient>

<decide>
1. Add Page, PageSize fields to ListUsersRequest in api/proto/user.proto.
2. Regenerate proto: buf generate.
3. Update service handler: extract params, validate bounds, pass to repo.
4. Update repo List method: accept page/pageSize, return ([]User, int64, error).
5. Build pagination meta with pkg/pagination.BuildMeta().
6. Add table-driven tests: default values, boundary (pageSize=100), invalid (0, -1, 101).
7. Run: go build ./..., go vet ./..., go test ./internal/service/ -run TestListUsers,
   golangci-lint run.
</decide>

<draft>
// api/proto/user.proto -- add to ListUsersRequest
message ListUsersRequest {
  int32 page = 1;      // default 1
  int32 page_size = 2; // default 20, max 100
}

// internal/service/user.go -- updated handler
func (s *UserService) ListUsers(
    ctx context.Context, req *pb.ListUsersRequest,
) (*pb.ListUsersResponse, error) {
    page, pageSize := normalizePagination(req.Page, req.PageSize)
    if page < 1 || pageSize < 1 || pageSize > 100 {
        return nil, status.Error(codes.InvalidArgument, "invalid pagination")
    }
    users, total, err := s.repo.List(ctx, page, pageSize)
    if err != nil {
        return nil, fmt.Errorf("list users: %w", err)
    }
    return &pb.ListUsersResponse{
        Users:      users,
        Pagination: pagination.BuildMeta(total, page, pageSize),
    }, nil
}

func normalizePagination(page, pageSize int32) (int32, int32) {
    if page < 1 { page = 1 }
    if pageSize < 1 { pageSize = 20 }
    if pageSize > 100 { pageSize = 100 }
    return page, pageSize
}
</draft>

<precheck>
All imports (fmt, status, codes, pb, pagination) use correct module path v
Dependencies: pkg/pagination.BuildMeta() signature matches call v
Error wrapping: fmt.Errorf("list users: %w", err) per convention v
Context: ctx is first param v
Naming: normalizePagination is unexported camelCase v
Proto field numbers: 1, 2 -- no collisions with existing fields v
Blast radius: only ListUsers handler + proto + tests affected v
</precheck>

<act>
Applied changes to proto, service handler, and test files.
Regenerated proto with buf generate.
</act>

<evaluate>
go build ./... -> passed.
go vet ./... -> passed.
go test ./internal/service/ -run TestListUsers -> 8/8 passed (5 new cases).
golangci-lint run -> no new issues.
Static review: error wrapping, context propagation, naming all match conventions v
</evaluate>
```

## OODA-E Workflow

### `<observe>` Requirement And Project Reconnaissance

Summarize the requested code change in one sentence. Identify the affected Go
packages, files, interfaces, or RPC methods. Read `go.mod`, indexes, source
files, test files, and configuration. When an infrastructure index exists,
use it to locate public contracts before reading private implementation.

### `<orient>` Standards, Contracts, And Dependency Alignment

Align with Go project conventions: package naming, import grouping, error
handling (`%w` wrapping, sentinel errors), context propagation, interface
design, concurrency patterns, and testing style. Load
`references/go-conventions.md` for the full checklist.

Check whether the change hand-writes a pattern that exists in the standard
library or project infrastructure (HTTP client, retries, logging, metrics,
caching, rate limiting, auth middleware, validation, connection pooling,
graceful shutdown). If yes, use the existing capability -- do not hand-write
it.

Prefer narrow, idiomatic changes.

### `<decide>` Implementation Plan

List concrete steps: target Go files, imports, interfaces, structs, functions,
test files. Include validation commands. Keep scoped.

### `<draft>` Brain Sandbox Draft

Draft Go code in text before editing files. Cover: imports (grouped), types,
function signatures, error handling, context propagation, concurrency safety,
and test strategy. Do not call file-writing tools here.

### `<precheck>` Static Quality Gate

Inspect the draft:

- **Dependency health**: imports use correct module path. Referenced
  types/functions exist.
- **Type/contract health**: signatures match callers. Interface satisfaction
  correct. Proto contracts respected.
- **Syntax health**: no unused imports, shadowed variables, incorrect
  defer/close semantics.
- **Standards health**: error wrapping, context propagation, naming
  (exported/unexported), package conventions. See
  `references/go-conventions.md`.
- **Blast-radius health**: no changes to unrelated packages, no global
  state mutations.

### `<act>` Tool-Based File Write

Apply changes with file editing tools. Do not use Markdown-only unless asked.
Respect the working tree. If proto changes were made, run code generation.

### `<evaluate>` Dynamic Validation And Self-Correction

Validate with the Go toolchain. See `references/go-conventions.md` for the
full command reference and sequence.

Standard sequence: `go build ./...` -> `go vet ./...` -> `go test ./...
-count=1` -> `golangci-lint run`.

If validation fails, report the error and self-correct without waiting. Return
to the appropriate earlier phase, fix, and rerun until checks pass or a
genuine external blocker is reached.

---

## Index Awareness

When an `index.md` or OODA-style index exists, treat it as the navigation
and contract layer:

- Use it to find public APIs before reading implementation files.
- Keep implementation aligned with documented contracts.
- If updating a public API, update the relevant index when within scope.
- If the index appears stale, verify against source and mention the mismatch.

---

## Resources

### references/
- `go-conventions.md` -- Go naming, error handling, concurrency patterns,
  and toolchain command reference. Load during `<orient>` and `<evaluate>`
  phases.
