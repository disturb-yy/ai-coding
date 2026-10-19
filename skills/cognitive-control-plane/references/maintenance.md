# Maintaining this skill

The conversational procedure lives in `SKILL.md`. Memory, coordination, and independent review are disclosed only for their respective branches. Keep specialist selection based on the current catalog; maintain no parallel skill registry or per-role prompt templates.

The optional host scripts and adapter schema are a separate runtime surface. They do not enforce the conversational workflow or grant permission to launch workers. Preserve their interfaces when changing only instructions. The README guard is an optional routing aid, not a security boundary or semantic validator.

## Verification

After changing the core or its packaging:

1. Run the available Skill Creator `quick_validate.py` on this directory and check local Markdown links.
2. Install with `scripts/install-cli.js --skill-only --install-dir PATH` in an isolated home. Check both discovery links, every packaged reference, and `scripts/setup-memory.js` preview/install behavior. Repeat installation to check backups and retained memory.
3. If changing optional runtime files, run relevant Node syntax and behavior checks. `scripts/host-install.js verify codex|opencode|all` checks registration after installation; it does not prove a host has loaded the integration.
4. Check the diff for unrelated edits and stale paths. Report source edits separately from installed-copy updates.

Use a child process with an isolated environment for installation checks; never point test installers at a real user profile. Global instructions and memory setup are explicit installation operations.

## Behavioral probes

Exercise these scenarios when a substantial rewrite changes routing. Inspect the resulting actions; matching headings or phrases cannot prove behavior.

| Request and context | Expected behavior |
| --- | --- |
| A clear, authorized edit | Execute and validate without a preliminary interview or coordination record. |
| An ambiguous product request; workspace answers some questions | Inspect those facts, ask only the remaining consequential choice, continue independent work. |
| User requests a design only | Deliver the design and respect the implementation boundary. |
| User corrects one requirement midway | Apply the correction while retaining accepted decisions and unfinished work. |
| A command fails twice without new evidence | Change the diagnostic check; inspect partial effects before another write. |
| A relevant memory record conflicts with live configuration | Verify current configuration and treat the old record as stale; no unsolicited memory write. |
| A requested handoff after partial implementation | Include actual artifacts, verification, authorization, remaining work, and the first next action. |
| A specialist or optional worker is unavailable | Continue through an adequate available method; identify only indispensable gaps. |
| Required independent review has no available reviewer | Complete preparation/checks and report pending review without claiming independence. |

Manual inspection of these cases is a walkthrough, not an executed agent evaluation. Keep that distinction in validation reports.
