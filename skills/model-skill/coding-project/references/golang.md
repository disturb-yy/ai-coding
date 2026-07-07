# Go Coding Reference

Load this reference when editing Go code or when the project contains `go.mod`, `go.sum`, or relevant `*.go` files.

## Orientation Checklist

- Confirm the module path and Go version in `go.mod`.
- Use the module path for internal imports.
- Match existing package names, directory boundaries, and file organization.
- Check nearby tests and table-driven test patterns before adding new tests.
- Prefer standard library and project helpers before adding dependencies.
- Preserve generated files unless the project generator command is available and needed.

## Code Conventions

- Package names are lowercase, short, and without underscores.
- Keep interfaces small; accept interfaces at boundaries and return concrete structs when that matches local style.
- Put `context.Context` first in functions that accept it.
- Handle every error explicitly; avoid `_` for errors unless the surrounding code establishes a clear exception.
- Wrap errors with useful operation context using `%w` when returning to callers.
- Keep goroutine lifecycles explicit with `context`, `sync.WaitGroup`, `errgroup`, or owned channels.
- Group imports as standard library, third-party, then internal packages.
- Use all-caps for common initialisms in identifiers: `ID`, `URL`, `HTTP`, `API`.

## Testing

- Prefer table-driven tests with `t.Run` for multiple cases.
- Use the same package style as nearby tests (`package foo` or `package foo_test`).
- Use existing assertion libraries only when already present.
- Add focused tests for changed behavior, error paths, and boundary cases.
- Avoid broad integration tests when a narrow unit test covers the requested behavior.

## Validation Commands

Prefer commands documented by the project. When absent, use the smallest relevant subset:

```bash
go test ./target/... -count=1
go test ./... -count=1
go vet ./...
go build ./...
golangci-lint run
go mod tidy
```

Use `go mod tidy` only after dependency or import changes that legitimately affect `go.mod` or `go.sum`.


## Few-Shot Examples

### Add context-aware repository call

Input:

```text
Update the user service to pass request context into the repository call.
```

Expected behavior:

```text
Confirm the repository interface and nearby call sites first.
Keep `context.Context` as the first parameter.
Update the interface, implementation, call sites, and focused tests only.
Run the affected package tests before broader validation.
```

Example shape:

```diff
-func (r *Repo) FindUser(id string) (*User, error)
+func (r *Repo) FindUser(ctx context.Context, id string) (*User, error)
```

### Wrap lower-level errors with operation context

Input:

```text
Make this Go config loader return useful errors when reading the file fails.
```

Expected behavior:

```text
Preserve the existing exported API unless the task requires changing it.
Use `%w` so callers can still inspect the original error.
Do not log and return the same error unless that is already project style.
```

Example shape:

```diff
- return nil, err
+ return nil, fmt.Errorf("load config %s: %w", path, err)
```

### Preserve cancellation in concurrent work

Input:

```text
Make the Go batch processor stop promptly when the request context is cancelled.
```

Expected behavior:

```text
Inspect existing goroutine ownership and error propagation first.
Pass context through worker boundaries instead of creating a detached background context.
Use existing errgroup, WaitGroup, or channel patterns from the package.
Add a focused cancellation test when the behavior changes.
```

Example shape:

```diff
- go worker.Run(context.Background(), job)
+ group.Go(func() error {
+     return worker.Run(ctx, job)
+ })
```
