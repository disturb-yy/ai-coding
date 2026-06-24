# Database Diagnosis Reference

Use this reference for SQL, migration, transaction, locking, index, ORM, data
correctness, and query performance problems. Diagnose only; do not modify data
or schema.

Collect query text, schema shape, migration history, database engine/version,
transaction boundaries, isolation level, relevant ORM-generated SQL, row counts,
and sanitized sample data.

Prefer read-only inspection:

- Explain plans.
- Schema inspection.
- Read-only SELECT probes.
- Lock/wait diagnostics.
- Query timing with representative parameters.

Never run destructive statements, migrations, writes, or cleanup commands from
this skill. Put any proposed migration or data repair into handoff only.

