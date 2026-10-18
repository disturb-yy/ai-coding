---
access:
  audience: model
  model_read: true
  model_write: true
  purpose: skill_reference
---
# Maintenance

Use this reference only when modifying the Cognitive Control Plane skill or its bundled guards.

## Bundled Scripts

Use scripts only for installation and validation, not for semantic task classification.

- Run `node scripts/host-install.js verify [all|codex|opencode]` after optional host installation. Checks compare actual registration and launcher contents; host activation still needs a host reload when plugins are cached.
- Run `node scripts/host-install.js install [all|codex|opencode]` to install thin launchers manually. Implementations stay under this skill's `scripts/`; host directories contain registration, launchers, and recovery backups only. Existing host enable/trust settings are preserved.
- The optional guard protects the human-facing `README.md` from model instruction reads; it never scans document contents for access metadata. File listing is allowed. Simple shell checks are a routing aid, not a complete shell parser or a security boundary.
- Core conversational use requires only `SKILL.md` and selected references. Host adapters and scheduling are optional. Keep generated results and virtual environments out of distribution packages; preserve historical evidence separately when cleaning up.
