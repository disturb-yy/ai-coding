# JavaScript And TypeScript Diagnosis Reference

Use this reference for Node, browser, React, TypeScript, package manager,
bundler, hydration, async UI, or test failures. Diagnose only; do not modify
files.

Collect runtime version, package manager, lockfile type, exact failing command,
browser console output when relevant, stack traces, and the smallest route,
component, script, or test that reproduces the symptom.

Prefer non-mutating commands such as:

```bash
node --version
npm test -- --runInBand
npm run build
npm run typecheck
pnpm test
pnpm run build
```

Avoid install, update, format, or lockfile-changing commands during diagnosis
unless the user explicitly asks for repair.

