---
name: memory-generator-memory-generator
description: Use the memory_generator role with memory-generator to deduplicate and safely persist verified, durable memory from completed work.
disable-model-invocation: true
---

# Memory Generator Memory Generator

Run the task as `memory_generator` through the `memory-generator` workflow.

## Workflow

1. Read [memory-generator.md](../../role/memory-generator.md), its linked
   [handoff standard](../../role/handoff-standard.md), and the current task's `AGENTS.md` local
   memory rules. Treat them as the role and storage contract.
2. Read and follow [memory-generator](../../model-involved/memory-generator/SKILL.md). Treat it as
   the memory selection, deduplication, lifecycle, and safe-persistence workflow.
3. Persist only user guidance or verified facts with durable future value. Check the same type and
   scope for an active equivalent record before writing; update it rather than adding a duplicate.
4. Use direct execution for every persistent write. Stop without writing when the source, scope,
   stability, storage permission, lifecycle, or sensitivity decision is unclear.

## Handoff

Report the role contract's required fields:

- `memory_records`
- `deduplication_decision`
- `retention_decision`
- `validation`
- `handoff`
