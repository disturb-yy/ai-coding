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

- Run `node scripts/check-mirrors.js` once after completing a canonical/mirror edit batch. It checks existence and timestamp freshness using filesystem metadata; it does not prove translation accuracy.
- Runtime hooks only remind about mirror updates for explicit canonical writes. They do not fail completed tool calls because the translation is temporarily pending. Shell writes are covered by the explicit end-of-batch check.
- Run `bash scripts/verify-install.sh [all|codex|opencode]` after installation. Installation checks compare actual registration and launcher contents; host activation still needs a host reload when plugins are cached.
- Run `bash scripts/install-guards.sh [all|codex|opencode]` to install thin launchers. Implementations and policy stay under this skill's `scripts/` and `config/`; host directories contain registration, launchers, and recovery backups only. Existing host enable/trust settings are preserved.
- Guard read targets are declared in `config/guard-policy.json`; the hook never scans document contents for access metadata. File listing is allowed. Simple shell checks are a routing aid, not a complete shell parser or a security boundary.
- Core conversational use requires only `SKILL.md` and selected references. Host adapters, scheduling, and evaluation are optional. Keep generated results and virtual environments out of distribution packages; preserve historical evidence separately when cleaning up.
