# Secrets, Dependencies, And Supply Chain

Use this checklist for credentials, configuration, dependencies, build scripts, and generated artifacts.

## Checks

- Secrets, tokens, private keys, session material, API keys, passwords, and connection strings are not committed, logged, exposed in responses, bundled into clients, or written to artifacts.
- Secret rotation, scoping, and environment-specific configuration remain intact.
- New dependencies are necessary, maintained, license-compatible when the repo tracks licenses, and not replacing simple standard-library behavior without reason.
- Package scripts, build steps, CI config, container files, install hooks, and code generation do not execute untrusted input or fetch unpinned remote code unexpectedly.
- Lockfiles, checksums, vendored code, generated files, and binary artifacts match the intended source change.
- Deserialization, plugin loading, dynamic imports, native extensions, and reflection introduce only necessary execution surfaces.
- Development-only tools and debug flags do not ship into production paths.

## Evidence

Inspect changed dependency manifests, lockfiles, CI/build files, Dockerfiles, generated artifacts, and configuration. Use package metadata or security tooling when available, but verify that the changed code actually introduces the reachable risk.

## Report

For each issue, state the exposed secret or supply-chain surface, where it enters the build/runtime path, and the smallest containment or replacement.
