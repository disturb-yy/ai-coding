# Syntax Analysis

Use this checklist to find defects that should be caught before behavior review.

## Checks

- Parse or compile failures in changed files.
- Type errors, missing imports, wrong exports, undefined symbols, broken generics, and invalid annotations.
- API signature mismatches between caller and callee.
- Config/schema syntax problems in JSON, YAML, TOML, SQL, GraphQL, templates, migrations, and generated manifests.
- Async, resource, or lifecycle constructs that are syntactically valid but misused in a way the language tooling would flag.
- Test files that no longer compile, import, discover, or run.

## Evidence

Prefer project-native commands when available: lint, typecheck, compile, test discovery, or package-specific validation. If commands are unavailable, read the smallest caller/callee or config/schema pair needed to prove the issue.

## Report

Report only issues that are visible in the diff or directly caused by it. Include the exact symbol, import, signature, config key, or command failure.
