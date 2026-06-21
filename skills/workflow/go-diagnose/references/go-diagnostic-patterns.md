# Go Diagnostic Patterns

Symptom → diagnosis → fix catalog. Match observed symptoms to the closest pattern, then use the diagnostic commands to confirm.

## Panic Patterns

### Nil Pointer Dereference

**Symptom**: `panic: runtime error: invalid memory address or nil pointer dereference`

**Common causes**:
- Function returns `(*T, error)` — caller uses pointer before checking error.
- Map lookup without ok check: `m[key].Field` when key missing.
- Interface holding nil concrete type (not nil interface).
- Uninitialized struct field used before assignment.

**Diagnostic commands**:
```
go vet ./...        # may catch some nil deref patterns
go test -race -run TestX
```

**Fix pattern**: Check error/ok first. Use `if err != nil { return ... }` before using the value.

### Slice Bounds Out of Range

**Symptom**: `panic: runtime error: index out of range [X] with length Y`

**Common causes**:
- Off-by-one in loop: `for i := 0; i <= len(s); i++`.
- Empty slice indexing: `s[0]` on nil or empty slice.
- Slice expression exceeding capacity: `s[start:end]` where end > cap(s).

**Fix pattern**: Guard with `if len(s) > 0`, use `for i := range s`, or validate bounds before indexing.

### Send on Closed Channel

**Symptom**: `panic: send on closed channel`

**Common causes**:
- Multiple goroutines closing the same channel.
- Sender closes channel while other senders are still active.
- `defer close(ch)` in a function that's called multiple times.

**Fix pattern**: Use `sync.Once` to close. Or restructure so only the sender closes, never the receiver. Consider using `context.Context` for cancellation instead of closing channels.

### Concurrent Map Writes

**Symptom**: `fatal error: concurrent map writes` or `concurrent map read and map write`

**Common causes**:
- Multiple goroutines writing to the same map without synchronization.
- Reading and writing to a map concurrently.

**Diagnostic**: `go test -race` — the race detector catches this.

**Fix pattern**: Use `sync.RWMutex` for read-heavy maps, `sync.Mutex` for write-heavy, or `sync.Map` for specific access patterns (keys are static, entries written once).

## Hang / Deadlock Patterns

### Mutex Contention

**Symptom**: Goroutine dump shows many goroutines in `[sync.Mutex.Lock]` state.

**Common causes**:
- I/O or blocking call inside mutex critical section.
- Missing `Unlock()` — early return, panic without defer.
- Double lock: calling `Lock()` twice in same goroutine (`sync.Mutex` is not reentrant).

**Diagnostic**:
```
# Get goroutine dump
curl http://localhost:6060/debug/pprof/goroutine?debug=2
# Or send SIGQUIT to get dump on stderr
```

**Fix pattern**: 
- `defer mu.Unlock()` immediately after `mu.Lock()`.
- Move blocking operations outside the critical section.
- Keep critical sections short.

### Channel Deadlock

**Symptom**: `fatal error: all goroutines are asleep - deadlock!`

**Common causes**:
- Unbuffered channel send with no receiver.
- Channel receive with no sender.
- Select without default, no case ready.
- Goroutine leak: goroutine blocked on channel, never cleaned up.

**Diagnostic**:
```
go test -run TestX -timeout 10s    # confirm hang
go test -race -run TestX           # check for related races
```

**Fix pattern**: Ensure sends and receives are paired. Use buffered channels or select with default/timeout for optional operations. Track goroutine lifecycle with `sync.WaitGroup`.

### Goroutine Leak

**Symptom**: Memory grows over time. `runtime.NumGoroutine()` increases monotonically.

**Common causes**:
- Goroutine started in a loop without termination condition.
- Channel that never closes, goroutine blocks on receive forever.
- `context.Context` not passed or ignored, goroutine can't be cancelled.

**Diagnostic**:
```
go test -run TestX -memprofile=mem.out
go tool pprof -top mem.out            # check goroutine count
curl http://localhost:6060/debug/pprof/goroutine?debug=1  # goroutine count by state
```

**Fix pattern**: Pass `context.Context` to goroutines. Close channels to signal completion. Use `sync.WaitGroup` to track goroutine lifecycle. Avoid spawning goroutines in loops without bounds.

## Race Condition Patterns

### Unsynchronized Shared Variable

**Symptom**: Intermittent test failures. `go test -race` reports `DATA RACE`.

**Common causes**:
- Shared global or struct field accessed from multiple goroutines without synchronization.
- `WaitGroup.Add()` called concurrently with `Wait()` — must call Add before Wait.
- Closure captures loop variable (Go < 1.22).

**Diagnostic**: `go test -race -run TestX -count=5` (run multiple times; races are probabilistic).

**Fix pattern**: Use `sync.Mutex` or `atomic` for simple counters. For Go < 1.22, copy loop variable: `i := i` before closure.

### HTTP Handler Shared State

**Symptom**: Race detected in HTTP handler. Handler accesses shared map or slice.

**Common causes**: HTTP handlers are concurrent. A package-level map or a struct field accessed from multiple handlers races.

**Fix pattern**: Protect shared state with mutex, or make state per-request via `context.Context`.

## Build Error Patterns

### Import Cycle

**Symptom**: `import cycle not allowed`

**Common causes**: Package A imports B, B imports A (direct or transitive).

**Fix pattern**: Extract shared types/interfaces into a third package. Use interfaces to break the dependency: A defines the interface, B implements it. Or restructure into unidirectional dependency flow.

### Missing go.sum Entry

**Symptom**: `missing go.sum entry for module providing package ...`

**Fix**: Run `go mod tidy` to regenerate go.sum.

### CGO / Build Constraint Issues

**Symptom**: `build constraints exclude all Go files in ...` or CGO linking errors.

**Common causes**: Platform-specific files (`_linux.go`, `_windows.go`) with no fallback. Missing C library for CGO.

**Fix pattern**: Provide platform-independent stub when needed. Document CGO dependencies.

## Performance Patterns

### Allocation in Hot Path

**Symptom**: Benchmark shows high `allocs/op`. Profiling shows time spent in GC.

**Diagnostic**:
```
go test -bench=. -benchmem
go test -run TestX -memprofile=mem.out
go tool pprof -top mem.out
```

**Fix pattern**: Reuse buffers with `sync.Pool`. Preallocate slices with `make([]T, 0, capacity)`. Avoid `fmt.Sprintf` in hot paths — use `strings.Builder` or `strconv`.

### Defer in Loop

**Symptom**: Resources accumulate until function returns.

**Common cause**: `defer f.Close()` inside a loop — defers execute at function return, not at loop iteration end.

**Fix pattern**: Wrap loop body in anonymous function: `for ... { func() { defer ... }() }`. Or call Close directly without defer.

### Large Goroutine Count

**Symptom**: Thousands of goroutines, many idle. High memory usage.

**Common cause**: Goroutine-per-request model without bounds. No worker pool.

**Fix pattern**: Use worker pool pattern. Limit concurrency with buffered channel as semaphore: `sem := make(chan struct{}, maxConcurrency)`.

## Logic Bug Patterns

### Error Shadowing

**Symptom**: Error returned is nil, but inner error occurred.

**Common cause**: `:=` in inner scope shadows outer `err`. Example:
```go
var result string
if result, err := fetch(); err != nil {  // new 'err', doesn't set outer
    return err  // returns nil
}
```

**Fix pattern**: Declare `err` before the block: `var err error; var result string; result, err = fetch()`.

### Closure Loop Variable Capture

**Symptom**: All goroutines/closures see the same (last) value.

**Common cause** (Go < 1.22): Loop variable reused across iterations. Closures capture the variable, not its value.

**Fix pattern**: Go 1.22+ fixes this. For older: `i := i` at loop body start. Or pass as parameter.

### Slice Aliasing

**Symptom**: Modifying a slice unexpectedly changes another slice.

**Common cause**: Subslice shares underlying array: `b := a[1:3]`. Appending to `b` may modify `a`'s elements (if capacity allows).

**Fix pattern**: Use `copy(dst, src)` for independent slices. Or full slice expression `a[low:high:max]` to limit capacity.

### Nil Interface vs Nil Concrete Type

**Symptom**: `if err != nil` is true, but `err` prints as nil.

**Common cause**: Interface holds a nil concrete pointer: `var p *MyError = nil; var err error = p; err != nil // true!`

**Fix pattern**: Return `nil` directly, not a nil pointer: `return nil`. Don't assign typed nil to interface.
