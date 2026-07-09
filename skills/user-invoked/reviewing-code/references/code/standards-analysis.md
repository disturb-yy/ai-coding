# Standards Analysis

Use this checklist to compare the change against repo conventions.

## Source Order

1. Explicit repo rules: `CONTRIBUTING`, `CODING_STANDARDS`, docs, package READMEs, architecture notes, lint configs, formatter configs, type configs.
2. Local patterns in nearby files and tests.
3. Generic maintainability heuristics only when repo-specific evidence is missing.

Repo rules override generic heuristics. Nearby working examples are stronger than global preferences.

## Checks

- Naming, file placement, module boundaries, layering, and dependency direction match nearby code.
- Error handling, logging, retries, metrics, transactions, and cleanup follow established patterns.
- Tests use local fixtures, helpers, naming, and assertion style.
- Public interfaces, schemas, migrations, and generated artifacts follow the repo's existing update path.
- The change avoids duplicated logic, speculative abstraction, feature envy, shotgun edits, middle-man wrappers, and repeated condition cascades unless a repo pattern requires them.

## Report

Separate hard repo-rule violations from judgment calls. Cite the rule file or nearby precedent for hard violations. Label generic design smells as judgment calls.
