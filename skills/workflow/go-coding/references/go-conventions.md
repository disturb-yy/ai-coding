# Go Conventions & Toolchain Reference

Reference for the Go-specific orientation and toolchain commands. Load when orienting to a Go project or planning evaluation commands.

## Go-Specific Orientation Checklist

When orienting to the project, verify these Go-specific concerns:

- **Module path**: Confirm the module name in `go.mod`. All internal imports must use this prefix.
- **Package naming**: Lowercase, single-word, no underscores. Match existing package conventions.
- **Error handling**: `if err != nil` with wrapped errors (`fmt.Errorf("context: %w", err)`). Never ignore errors with `_`.
- **Context propagation**: Functions accepting `context.Context` as first parameter.
- **Interface segregation**: Small, focused interfaces. Accept interfaces, return structs.
- **Concurrency**: goroutines with explicit lifecycle management. Use `sync.WaitGroup`, `errgroup`, or channels with clear ownership.
- **Testing**: Table-driven tests using `t.Run()`. Test files in same package (`_test` suffix optional per project convention). Use `testify` if already a dependency.
- **Naming**: Exported identifiers start with uppercase. Acronyms are all-caps (`HTTP`, `URL`, `ID`). Unexported are camelCase.
- **Import grouping**: stdlib → third-party → internal, separated by blank lines.

## Go Toolchain Quick Reference

| Command | Purpose | When to run |
|---------|---------|-------------|
| `go build ./...` | Compile all packages | After any code change |
| `go vet ./...` | Static analysis for suspicious constructs | After any code change |
| `go test ./... -count=1` | Run all tests (no cache) | After logic changes |
| `go test ./pkg/... -run TestX` | Run specific tests | During targeted iteration |
| `golangci-lint run` | Comprehensive linting | Before considering work done |
| `go mod tidy` | Clean up go.mod/go.sum | After adding/removing imports |
| `buf generate` or `protoc` | Regenerate proto stubs | After .proto changes |

## Validation Sequence

Run in this order, stop on first failure:

1. `go build ./...` — compilation
2. `go vet ./...` — static analysis
3. `go test ./... -count=1` — tests (or `go test ./target/... -run TestX`)
4. `golangci-lint run` — lint (if configured in project)
