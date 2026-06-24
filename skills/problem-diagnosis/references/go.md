# Go Diagnosis Reference

Use this reference only after the generic problem-diagnosis workflow identifies
the issue as Go-related. Treat all guidance here as diagnostic only. Do not
modify code from this skill.

## Evidence To Collect

- Full panic, build, vet, race, benchmark, or test output.
- `go version`.
- `go env GOPATH GOMOD GOOS GOARCH CGO_ENABLED`.
- `go.mod` and relevant `go.sum` symptoms.
- The smallest `go test`, `go build`, CLI, HTTP, or benchmark command that
  demonstrates the symptom.
- Stack frame files, callers, callees, goroutine states, and shared data paths.

## Useful Non-Mutating Commands

```bash
go test ./...
go test -run TestName -count=1 -v
go test -race -run TestName -count=1
go build ./...
go vet ./...
go test -bench=. -benchmem
go test -run TestName -memprofile=mem.out
go tool pprof -top mem.out
```

Avoid `go mod tidy` during diagnosis because it modifies files. Recommend it in
the handoff only when missing module metadata is the likely issue.

## Symptom Patterns

| Symptom | Common diagnostic direction |
|---|---|
| `nil pointer dereference` | Check error handling, map lookup `ok`, nil concrete values in interfaces, uninitialized fields. |
| `index out of range` | Check off-by-one loops, empty slices, invalid slice bounds, data-dependent lengths. |
| `send on closed channel` | Find all closers and senders. Confirm ownership of channel close. |
| `concurrent map writes` | Run race detector. Trace all goroutines accessing the map. |
| deadlock or hang | Capture goroutine dump. Inspect channel waits, mutex waits, context cancellation, and goroutine lifecycle. |
| `DATA RACE` | Map reported read/write stacks to shared state and synchronization boundaries. |
| import cycle | Build package graph and identify the smallest dependency inversion. |
| missing `go.sum` | Confirm module/package named in error. Do not run `go mod tidy`; record as likely fix direction. |
| CGO or build constraints | Check `GOOS`, `GOARCH`, `CGO_ENABLED`, build tags, and platform-specific files. |
| high allocs or slow benchmark | Use benchmark and pprof output to identify hot allocation or CPU path. |

## Hypothesis Examples

- If a panic is caused by unchecked `(*T, error)`, then the stack frame should
  show pointer use before the related `err` check.
- If a hang is caused by channel deadlock, then a goroutine dump should show
  matching send/receive waits without a live counterpart.
- If a test is flaky due to shared state, then `go test -race -run TestName`
  should report a read/write stack touching the same variable.

## Handoff Notes

For later Go fixes, record:

- Exact file and line of the suspected root cause.
- The minimal reproduction command.
- Whether `-race` is required for validation.
- Any command that should remain green after the fix.
- Any generated profiles or dumps and where they were stored.

