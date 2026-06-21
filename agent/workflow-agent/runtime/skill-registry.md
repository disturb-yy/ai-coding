# Skill Registry

## Purpose

The skill registry tells workflow-agent which workflow-compatible skill to use for each phase and artifact type.

Converted skills created with `workflow-skill-gen` are not active until they are registered.

## Default Path

```text
config/workflow-skills.json
```

Use `config/workflow-skills.example.json` as the template when creating a project-specific registry.

## Registry Contract

```json
{
  "version": 1,
  "skills": [
    {
      "phase": "Implementation",
      "artifact_type": "implementation_result",
      "skill": "coding",
      "path": "skills/coding",
      "enabled": true,
      "priority": 100,
      "execution": {
        "codex": "direct",
        "opencode": "delegated"
      },
      "config": {}
    }
  ]
}
```

## Selection Rules

- Match by `phase` and `artifact_type`.
- Ignore entries with `enabled: false`.
- Prefer the highest `priority` when multiple entries match.
- Use `execution.codex` in Codex mode.
- Use `execution.opencode` in OpenCode mode.
- Fall back to `direct` execution when delegated execution is unavailable.
- Reject a skill if its output artifact type does not match the registry entry.

## Adding A Converted Skill

After using `workflow-skill-gen`, add an entry:

```json
{
  "phase": "Implementation",
  "artifact_type": "implementation_result",
  "skill": "my-coding-skill",
  "path": "skills/my-coding-skill",
  "enabled": true,
  "priority": 100,
  "execution": {
    "codex": "direct",
    "opencode": "delegated"
  },
  "config": {
    "mode": "implement"
  }
}
```

For a UT generation skill:

```json
{
  "phase": "Testing",
  "artifact_type": "test_result",
  "skill": "my-ut-skill",
  "path": "skills/my-ut-skill",
  "enabled": true,
  "priority": 100,
  "execution": {
    "codex": "direct",
    "opencode": "delegated"
  },
  "config": {
    "mode": "generate",
    "run_tests": true
  }
}
```

## Validation Rules

- `path` must contain `SKILL.md`.
- `path` should contain `config.schema.json`.
- `path` should contain `example.json`.
- `config` must validate against the skill's `config.schema.json`.
- Skill output must use the workflow artifact envelope.
