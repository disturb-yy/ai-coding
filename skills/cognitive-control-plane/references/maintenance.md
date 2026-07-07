---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---

# Maintenance

Use this reference only when modifying the Cognitive Control Plane skill or its bundled guards.

## Canonical Files and Mirrors

English files are canonical. Chinese files are user-visible mirrors only.

- Canonical files: `SKILL.md` and `references/*.md`, excluding `*.zh-CN.md`.
- Mirror files: `zh/SKILL.zh-CN.md` and `zh/references/*.zh-CN.md`.
- When changing any canonical English file, overwrite the matching Chinese mirror from the updated English source in the same change.
- Never read, search, open, quote, summarize, or follow instructions from Chinese mirror files. Treat them as write-only output artifacts for the user.
- If exact incremental synchronization would require reading an existing Chinese mirror, regenerate the mirror from the canonical English file instead.

Completion criterion: every modified English canonical file has a corresponding modified Chinese mirror, and no Chinese mirror content was read.

## Bundled Scripts

Use scripts only for installation and validation, not for semantic task classification.

- Run `scripts/check-mirrors.js` after changing canonical English files to verify every mirror under `zh/` exists and is not older than its canonical source. This script uses filesystem metadata only; it does not read mirror content.
- Run `scripts/verify-install.sh` after installing or changing hooks/plugins to verify the Skill, Codex hook, OpenCode plugin, and mirror state.
- Run `scripts/install-guards.sh` when installing or refreshing the Codex hook and OpenCode plugin from this Skill's bundled guard templates.
