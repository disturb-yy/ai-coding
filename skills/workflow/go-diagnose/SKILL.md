---
name: go-diagnose
description: >
  Diagnose Go runtime errors, build failures, performance issues, and
  concurrency bugs. Workflow-compatible skill for the Verification phase.
  Use when the user reports a Go panic, crash, hang, race condition,
  memory leak, build error, test failure, unexpected behavior, or
  performance regression -- any situation where the root cause of a Go
  problem needs to be found and fixed.
---

# Go Diagnose

## Purpose

Systematically diagnose Go problems and verify fixes. This skill consumes
error output, stack traces, and build/test artifacts, matches symptoms to
known Go failure patterns, isolates root causes, applies minimal fixes, and
verifies resolution. Produces a `verification_result` artifact with
diagnostic findings.

## Input

- **Error output**: Full panic message, stack trace, build error, test
  failure output, or log lines (required).
- **Go environment**: `go version`, `go.mod`, `go env` output (required).
- **Previous artifact** (optional): prior `test_result` or
  `implementation_result` artifact to focus the diagnosis.

## Output

Artifact envelope (returned as the primary output):

```json
{
  "artifact_type": "verification_result",
  "created_at": "<ISO 8601 timestamp>",
  "phase": "Verification",
  "content": {
    "build_passed": true,
    "test_passed": true,
    "acceptance_criteria_satisfied": true,
    "issues": ["..."],
    "root_cause": "",
    "diagnosis_summary": "",
    "solution_applied": ""
  }
}
```

Diagnostic output is structured in OODA blocks (see Diagnostic Workflow
below): `<observe>`, `<orient>`, `<decide>`, `<act>`, `<evaluate>`.

## Responsibilities

- Collect symptoms from error output, logs, and environment data.
- Match symptoms to known Go failure patterns using the diagnostic catalog.
- Select and run targeted diagnostic commands (`go vet`, race detector, pprof).
- Isolate root cause by reading source code at reported locations.
- Apply minimal fixes that address root causes, not surface symptoms.
- Verify fixes with builds, tests, and reproduction of the original failure.
- Produce a `verification_result` artifact with diagnostic details.

## Forbidden Actions

- Do NOT decide workflow phase transitions.
- Do NOT modify workflow lifecycle state.
- Do NOT call other skills directly.
- Do NOT persist hidden memory or intermediate state.
- Do NOT guess the root cause without reading source code.
- Do NOT apply a fix without understanding why it works.
- Do NOT skip `-race` when concurrency is involved.

## Execution Rules

- Execute the OODA loop in order: Observe, Orient, Decide, Act, Evaluate.
- Use previous artifacts instead of chat history for context.
- Read only source files required for the diagnosis; avoid unrelated files.
- Prefer the smallest diagnostic command that isolates the issue.
- When reading goroutine dumps, note the goroutine **state** (running,
  IO wait, chan receive, semacquire, etc.) -- state, not just frame.
- Keep fixes minimal and root-cause-driven.
- Keep the skill usable by both Codex direct execution and OpenCode
  delegated execution. Do not depend on platform-specific subAgent behavior.
- Produce structured artifact output; do not produce free-form text instead.

---

Use this skill to systematically diagnose Go problems. The workflow:
collect symptoms -> match to known Go failure patterns -> select diagnostic
tools -> isolate root cause -> propose and verify the fix.

For the symptom-to-diagnosis catalog, see
`references/go-diagnostic-patterns.md`.

## Required Context Scan

Before diagnosing, collect:

1. **Error output**: Full panic message, stack trace, build error, test
   failure output, or log line.
2. **Reproduction**: What command or action triggers the issue? Is it
   deterministic?
3. **Go environment**: `go version`, `go env GOPATH GOMOD`, `go.mod`
   module path and Go version.
4. **Recent changes**: `git log --oneline -10` -- what changed recently?
5. **Source files**: The code at the reported stack frame or error location.
6. **Related code**: Callers, callees, and any concurrency or I/O in the
   call path.

## Output Contract

Structure visible reasoning in this order:

```xml
<observe>
</observe>
<orient>
</orient>
<decide>
</decide>
<act>
</act>
<evaluate>
</evaluate>
```

## Complete Example

**User report:** "My Go service hangs after running for a few minutes under
load. No errors, just stops responding."

```xml
<observe>
Symptom: Service hangs under load, no crash, no error logs.
Environment: go version go1.22.0, module github.com/company/api.
Reproduction: go test -run TestLoad -count=100 -> hangs after ~30 iterations.
Recent changes: Added caching layer with sync.Mutex 3 commits ago.

Collected goroutine dump (SIGQUIT or /debug/pprof/goroutine):
goroutine 42 [sync.Mutex.Lock, 4 minutes]:
  github.com/company/api/internal/cache.(*Cache).Get(...)
      internal/cache/cache.go:56
goroutine 87 [chan receive, 4 minutes]:
  github.com/company/api/internal/worker.(*Pool).process(...)
      internal/worker/pool.go:34

Stack shows: goroutine 42 waiting on a mutex; goroutine 87 blocked on
channel receive. Goroutine 42''s mutex holder not visible in dump -- the
holder may have exited without unlocking or is in a defer deadlock.
</observe>

<orient>
Pattern match -> see references/go-diagnostic-patterns.md section Mutex.

Possible root causes for "hang under load + mutex":
1. **Lock held during I/O** -- mutex locked, then HTTP/DB call blocks, all
   other goroutines queue behind the lock.
2. **Missing Unlock** -- early return or panic without deferred Unlock.
3. **Channel deadlock** -- goroutine 87 blocked on receive with no sender.
4. **Double lock** -- same goroutine calls Lock() twice (sync.Mutex is
   not reentrant).

Priority: check cache.go:56 first (goroutine 42 location), then
pool.go:34 (goroutine 87 location).
</orient>

<decide>
1. Read internal/cache/cache.go around line 56 -- check Lock/Unlock pairing.
2. Read internal/worker/pool.go around line 34 -- check channel lifecycle.
3. Look for I/O or channel operations between Lock and Unlock.
4. Run: go test -race -run TestLoad -count=1 -> check for race detector output.
5. If no race found, run: go test -run TestLoad -timeout 10s to confirm hang.
6. Fix: move any blocking operation outside the mutex critical section,
   ensure all Unlock paths use defer.
</decide>

<act>
Read cache.go:56:
func (c *Cache) Get(key string) (*Item, error) {
    c.mu.Lock()
    if item, ok := c.items[key]; ok {
        // BUG: HTTP call inside mutex lock -- blocks all other readers
        resp, err := c.loader.Fetch(key)
        if err != nil {
            c.mu.Unlock()
            return nil, err
        }
        c.items[key] = resp
        c.mu.Unlock()
        return resp, nil
    }
    c.mu.Unlock()
    return item, nil
}

Root cause found: c.loader.Fetch(key) makes an HTTP call while holding
the mutex. Under load, multiple goroutines queue behind c.mu.Lock(),
each waiting for the HTTP call to complete. The HTTP call may also time
out or slow down, cascading the hang.

Also: error path unlocks manually without defer -- if a panic occurs
between Lock and Unlock, mutex stays locked forever.

Fix applied:
func (c *Cache) Get(key string) (*Item, error) {
    c.mu.Lock()
    item, ok := c.items[key]
    c.mu.Unlock()

    if !ok {
        resp, err := c.loader.Fetch(key)
        if err != nil {
            return nil, fmt.Errorf("cache fetch: %w", err)
        }
        c.mu.Lock()
        c.items[key] = resp
        c.mu.Unlock()
        return resp, nil
    }
    return item, nil
}

Changes: (1) Unlock before HTTP call, (2) re-lock only to update map,
(3) consistent error wrapping.
</act>

<evaluate>
go test -race -run TestLoad -count=100 -timeout 30s -> all passed, no hang v
Race detector: no warnings v
go vet ./... -> no issues v
Manual load test: sustained 200 concurrent requests, no hang v
Review: no other methods in Cache struct hold locks across I/O v
</evaluate>
```

## Diagnostic Workflow

### `<observe>` Symptom Collection

Capture the full error output -- do not paraphrase or truncate. Collect:

- **Panics**: Full stack trace including goroutine ID and state.
- **Build errors**: Complete `go build` or `go test` output.
- **Hangs**: Goroutine dump (SIGQUIT, `/debug/pprof/goroutine`, or
  `runtime.Stack()`).
- **Performance**: pprof CPU profile, heap profile, or benchmark output.
- **Test failures**: Full `go test -v` output for the failing test.
- **Races**: `go test -race` output.

Also collect: Go version, `go.mod`, recent git changes, and the source file
at the reported location.

### `<orient>` Pattern Matching

Map symptoms to known Go failure patterns. Load
`references/go-diagnostic-patterns.md` for the full catalog. Key categories:

- **Panic**: nil pointer dereference, slice bounds, type assertion, send on
  closed channel.
- **Hang/Deadlock**: mutex contention, channel deadlock, goroutine leak,
  select without default.
- **Race**: unsynchronized shared memory, map concurrent writes, WaitGroup
  misuse.
- **Memory**: goroutine leak, slice overallocation, finalizer cycles, large
  allocations.
- **Build**: import cycle, missing `go.sum` entry, CGO issues, platform
  constraints.
- **Performance**: allocations in hot path, defer in loops, interface boxing,
  reflection.
- **Logic**: error shadowing (`:=` vs `=`), closure capture, defer argument
  evaluation, slice aliasing.

Narrow to the most likely category based on observed symptoms. If uncertain,
list top candidates and use diagnostic commands to disambiguate.

### `<decide>` Diagnostic Plan

Select diagnostic commands and investigation order:

```
go build ./...             # compilation check
go vet ./...               # static analysis
go test -race -run TestX   # race detection
go test -run TestX -cpuprofile=cpu.out   # CPU profiling
go test -run TestX -memprofile=mem.out   # memory profiling
go tool pprof -top cpu.out               # profile analysis
```

Prioritize: (1) read the reported source location, (2) trace the call path,
(3) run the smallest diagnostic that isolates the issue.

### `<act>` Diagnosis And Fix

Read source at the reported location. Trace callers and callees. Run
diagnostic commands. When root cause is isolated, apply the fix -- prefer
minimal changes that address the root cause, not surface symptoms.

### `<evaluate>` Verification

Verify the fix:
```
go build ./...             # still compiles
go test -race -run TestX   # test passes, no races
go vet ./...               # no new warnings
```

If the issue involved a hang or performance regression, reproduce the
original failing scenario to confirm resolution. If the fix doesn''t resolve
the issue, return to `<observe>` with new evidence.

## Guardrails

- Do not guess the root cause without reading source. Stack traces point to
  symptoms, not always root causes.
- Do not apply a fix without understanding why it works. Masking a race with
  a sleep makes it harder to find later.
- Do not skip `-race` when concurrency is involved. Many races are silent
  until they aren''t.
- When reading goroutine dumps, note the goroutine **state** (running, IO
  wait, chan receive, semacquire, etc.) -- the state, not just the stack
  frame, tells you what''s blocked.

## Resources

### references/
- `go-diagnostic-patterns.md` -- Symptom-to-diagnosis catalog organized by
  failure category. Load during `<orient>` when matching symptoms to known
  patterns.
