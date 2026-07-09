# Functionality Analysis

Use this checklist to decide whether the change does the right thing.

## Checks

- Required behavior from the issue, PRD, PR description, user prompt, or tests is missing, partial, or implemented on the wrong path.
- The new behavior works for the happy path but fails important edge cases: empty input, null/missing values, duplicates, time zones, concurrency, retries, cancellation, pagination, permissions, and partial failures.
- The change alters unrelated behavior, public API contracts, persistence shape, event formats, metrics, logging, or error semantics.
- The implementation updates one layer but misses another: route to service, service to storage, UI to API, migration to model, tests to fixtures.
- The tests do not cover the changed contract or assert implementation details instead of behavior.
- Rollback, migration, cache invalidation, background jobs, or idempotency requirements are missing when the change needs them.

## Evidence

Trace from entry point to implementation to persistence/integration and tests. Use CodeMap when available for call chains and impact, then verify with source and tests.

## Report

For each finding, state the expected behavior, actual behavior in the code, affected path, and a minimal fix direction.
