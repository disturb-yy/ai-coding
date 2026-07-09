# Go Unit Test Reference

Load this reference when adding or modifying Go unit tests, working in `*_test.go`, or when implementation changes require focused test coverage.

## Table of Contents

- Testing Scope
- Table-Driven Tests
- gomonkey Usage
- Mock Boundaries
- Naming and Assertions
- State Cleanup
- Validation Commands
- Few-Shot Examples

## Testing Scope

- Prefer the smallest module, package, function, or method boundary that proves the changed behavior.
- Cover the behavior introduced or changed by the task; avoid broad package-wide or end-to-end coverage unless the requested behavior crosses those boundaries.
- Add regression tests for bug fixes and boundary tests for changed validation or branching logic.
- Do not chase coverage percentage by testing unrelated paths.
- Keep tests deterministic and independent of execution order.

## Table-Driven Tests

- Prefer table-driven tests for multiple inputs, branches, validation cases, and error cases.
- Use `t.Run(tc.name, func(t *testing.T) { ... })` with descriptive case names.
- Include expected outputs, expected errors, and mock setup in each case when it keeps the test readable.
- Keep case structs local to the test unless several tests genuinely share the same shape.
- Avoid table-driven tests for a single simple case when a direct test is clearer.

## gomonkey Usage

- Prefer interface mocks, dependency injection, fakes, or local test doubles when the code already supports them.
- Use `gomonkey` only as a fallback for legacy code, tightly coupled code, global functions, third-party functions, constructors, time, environment, package-level state, or behavior that is difficult to inject safely.
- Always reset patches immediately after creating them; prefer `t.Cleanup(patches.Reset)` in tests, or use `defer patches.Reset()` only when no `testing.T` cleanup is available.
- Keep patches scoped to the smallest test or subtest that needs them.
- Do not use `t.Parallel()` in tests or subtests that use gomonkey. gomonkey patches are process-wide and not thread-safe.
- Prevent compiler inlining from invalidating patches. Any test run that depends on gomonkey must use no-inline flags, preferably `go test -gcflags=all=-l ./...` for the final run or the project-equivalent command.
- Avoid patching unrelated behavior just to make a wide test pass; narrow the tested module instead.

## Mock Boundaries

- Mock external systems, slow dependencies, nondeterministic behavior, and cross-module collaborators.
- Do not mock the function under test or simple data transformations that should be asserted directly.
- Reuse existing mocks, fakes, fixtures, builders, and helper assertions when present.
- Keep mock expectations close to the test case that needs them.
- Assert observable behavior first; assert call counts only when the interaction is part of the contract.

## Naming and Assertions

- Name tests as `TestFunctionName` or `TestType_Method` following nearby project style.
- Use standard `testing` helpers or the project's existing assertion library; do not add a new assertion dependency for one test.
- Call `t.Helper()` inside reusable test helpers.
- Prefer explicit assertions for important fields over deep equality on large structs when only a few fields matter.
- For errors, assert both presence and meaningful classification or message content according to project conventions.

## State Cleanup

- Use `t.Cleanup` for temporary files, environment variables, global state, clocks, and test-local resources.
- Use `t.Setenv` instead of manual environment mutation when available.
- Avoid shared mutable package-level state across tests. If unavoidable, restore it in `t.Cleanup`.
- Use `t.TempDir` for filesystem tests.
- Avoid `t.Parallel` when tests patch globals, mutate shared state, or use process-wide resources.

## Validation Commands

Prefer the narrowest useful command while iterating, then run the relevant broader command when practical:

```bash
go test ./path/to/package -run TestName -count=1
go test -gcflags=all=-l ./path/to/package -run TestName -count=1
go test ./path/to/package -count=1
go test ./... -count=1
go test -gcflags=all=-l ./...
```

If gomonkey is used, account for inlining. Follow the project's existing no-inline test command when present; otherwise run gomonkey-dependent validation with `go test -gcflags=all=-l ./...` before finishing. Use `-gcflags=all='-N -l'` only when the project already does so or when diagnosing optimization-sensitive patch failures.


## Few-Shot Examples

### Complete table-driven test

Input:

```text
Add unit tests for an integer Divide function covering normal division and division by zero.
```

Code under test:

```go
package calculator

import "errors"

func Divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }

    return a / b, nil
}
```

Expected test:

```go
package calculator

import "testing"

func TestDivide(t *testing.T) {
    tests := []struct {
        name      string
        a         int
        b         int
        want      int
        wantError bool
    }{
        {
            name:      "positive numbers",
            a:         10,
            b:         2,
            want:      5,
            wantError: false,
        },
        {
            name:      "negative dividend",
            a:         -10,
            b:         2,
            want:      -5,
            wantError: false,
        },
        {
            name:      "zero dividend",
            a:         0,
            b:         3,
            want:      0,
            wantError: false,
        },
        {
            name:      "division by zero",
            a:         10,
            b:         0,
            want:      0,
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)

            if tt.wantError {
                if err == nil {
                    t.Fatalf("expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            if got != tt.want {
                t.Fatalf("Divide(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

### Local fake instead of patching

Input:

```text
Test an OrderService that depends on a repository interface.
```

Expected behavior:

```text
Prefer a small local fake when the dependency is already an interface.
Keep fake behavior explicit in the test case.
Do not add gomonkey or a new mocking dependency for this case.
```

Example shape:

```go
type fakeOrderRepo struct {
    order Order
    err   error
}

func (f fakeOrderRepo) FindByID(ctx context.Context, id int64) (Order, error) {
    return f.order, f.err
}

func TestOrderService_Get(t *testing.T) {
    svc := NewOrderService(fakeOrderRepo{
        order: Order{ID: 42, Status: "paid"},
    })

    got, err := svc.Get(context.Background(), 42)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if got.Status != "paid" {
        t.Fatalf("status = %q, want %q", got.Status, "paid")
    }
}
```

### Table-driven gomonkey patches

Input:

```text
Test CanPay with user status and balance package functions patched per case.
```

Expected behavior:

```text
Keep gomonkey setup inside each table case through setupMock.
Use chain-style ApplyFunc calls when patching several functions for the same case.
Use t.Cleanup(patches.Reset) immediately after creating patches.
Assert unexpected collaborator calls with t.Fatalf inside the patch function.
Do not use t.Parallel with gomonkey patches.
Run targeted tests with the project's no-inline command, or add -gcflags=all=-l.
```

Example shape:

```go
package order

import (
    "errors"
    "testing"

    "github.com/agiledragon/gomonkey/v2"
)

func TestCanPay(t *testing.T) {
    tests := []struct {
        name string

        userID     int64
        orderPrice int64

        setupMock func(t *testing.T, patches *gomonkey.Patches)

        want    bool
        wantErr bool
    }{
        {
            name:       "normal user with enough balance",
            userID:     1001,
            orderPrice: 100,
            setupMock: func(t *testing.T, patches *gomonkey.Patches) {
                patches.
                    ApplyFunc(GetUserStatus, func(userID int64) (string, error) {
                        if userID != 1001 {
                            t.Fatalf("GetUserStatus userID = %d, want %d", userID, 1001)
                        }

                        return "normal", nil
                    }).
                    ApplyFunc(GetUserBalance, func(userID int64) (int64, error) {
                        if userID != 1001 {
                            t.Fatalf("GetUserBalance userID = %d, want %d", userID, 1001)
                        }

                        return 200, nil
                    })
            },
            want:    true,
            wantErr: false,
        },
        {
            name:       "normal user with insufficient balance",
            userID:     1002,
            orderPrice: 300,
            setupMock: func(t *testing.T, patches *gomonkey.Patches) {
                patches.
                    ApplyFunc(GetUserStatus, func(userID int64) (string, error) {
                        return "normal", nil
                    }).
                    ApplyFunc(GetUserBalance, func(userID int64) (int64, error) {
                        return 100, nil
                    })
            },
            want:    false,
            wantErr: false,
        },
        {
            name:       "frozen user cannot pay",
            userID:     1003,
            orderPrice: 100,
            setupMock: func(t *testing.T, patches *gomonkey.Patches) {
                patches.
                    ApplyFunc(GetUserStatus, func(userID int64) (string, error) {
                        return "frozen", nil
                    }).
                    ApplyFunc(GetUserBalance, func(userID int64) (int64, error) {
                        t.Fatalf("GetUserBalance should not be called")
                        return 0, nil
                    })
            },
            want:    false,
            wantErr: false,
        },
        {
            name:       "query user status failed",
            userID:     1004,
            orderPrice: 100,
            setupMock: func(t *testing.T, patches *gomonkey.Patches) {
                patches.
                    ApplyFunc(GetUserStatus, func(userID int64) (string, error) {
                        return "", errors.New("status db error")
                    }).
                    ApplyFunc(GetUserBalance, func(userID int64) (int64, error) {
                        t.Fatalf("GetUserBalance should not be called")
                        return 0, nil
                    })
            },
            want:    false,
            wantErr: true,
        },
        {
            name:       "query user balance failed",
            userID:     1005,
            orderPrice: 100,
            setupMock: func(t *testing.T, patches *gomonkey.Patches) {
                patches.
                    ApplyFunc(GetUserStatus, func(userID int64) (string, error) {
                        return "normal", nil
                    }).
                    ApplyFunc(GetUserBalance, func(userID int64) (int64, error) {
                        return 0, errors.New("balance db error")
                    })
            },
            want:    false,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            patches := gomonkey.NewPatches()
            t.Cleanup(patches.Reset)

            if tt.setupMock != nil {
                tt.setupMock(t, patches)
            }

            got, err := CanPay(tt.userID, tt.orderPrice)

            if tt.wantErr {
                if err == nil {
                    t.Fatalf("expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }

            if got != tt.want {
                t.Fatalf("CanPay(%d, %d) = %v, want %v",
                    tt.userID,
                    tt.orderPrice,
                    got,
                    tt.want,
                )
            }
        })
    }
}
```
