#!/usr/bin/env bash
set -euo pipefail

SKILL_DIR="${CCP_SKILL_DIR:-/home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane}"
CODEX_HOOK="${CCP_CODEX_HOOK:-/home/jadon/.codex/hooks/cognitive-control-plane-guard.js}"
CODEX_HOOKS_JSON="${CCP_CODEX_HOOKS_JSON:-/home/jadon/.codex/hooks.json}"
OPENCODE_PLUGIN="${CCP_OPENCODE_PLUGIN:-/home/jadon/.config/opencode/plugins/cognitive-control-plane-guard}"
OPENCODE_CONFIG="${CCP_OPENCODE_CONFIG:-/home/jadon/.config/opencode/opencode.json}"

failures=0

check() {
  local label="$1"
  shift
  if "$@"; then
    printf 'ok - %s\n' "$label"
  else
    printf 'not ok - %s\n' "$label" >&2
    failures=$((failures + 1))
  fi
}

file_exists() {
  test -f "$1"
}

dir_exists() {
  test -d "$1"
}

json_contains() {
  local file="$1"
  local needle="$2"
  node -e 'const fs=require("fs"); const [file, needle]=process.argv.slice(1); const text=fs.readFileSync(file,"utf8"); process.exit(text.includes(needle) ? 0 : 1);' "$file" "$needle"
}

check "skill directory exists" dir_exists "$SKILL_DIR"
check "SKILL.md exists" file_exists "$SKILL_DIR/SKILL.md"
check "references directory exists" dir_exists "$SKILL_DIR/references"
check "zh mirror root exists" dir_exists "$SKILL_DIR/zh"
check "zh references mirror directory exists" dir_exists "$SKILL_DIR/zh/references"
check "mirror check script exists" file_exists "$SKILL_DIR/scripts/check-mirrors.js"
check "mirror check passes" node "$SKILL_DIR/scripts/check-mirrors.js"

check "Codex hook script exists" file_exists "$CODEX_HOOK"
check "Codex hook syntax" node --check "$CODEX_HOOK"
check "Codex Pre/Post hook registered" json_contains "$CODEX_HOOKS_JSON" "cognitive-control-plane-guard.js"
check "Codex hook blocks zh reads" bash -c "printf '%s' '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Bash\",\"tool_input\":{\"cmd\":\"rg test $SKILL_DIR/zh/SKILL.zh-CN.md\"}}' | '$CODEX_HOOK' >/dev/null 2>&1; test \"\$?\" -eq 2"
check "Codex hook allows zh writes" bash -c "printf '%s' '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Edit\",\"tool_input\":{\"file_path\":\"$SKILL_DIR/zh/SKILL.zh-CN.md\"}}' | '$CODEX_HOOK' >/dev/null 2>&1"
check "Codex hook mirror post-check" bash -c "printf '%s' '{\"hook_event_name\":\"PostToolUse\",\"tool_name\":\"Edit\",\"tool_input\":{}}' | '$CODEX_HOOK' >/dev/null 2>&1"

check "OpenCode plugin directory exists" dir_exists "$OPENCODE_PLUGIN"
check "OpenCode plugin script exists" file_exists "$OPENCODE_PLUGIN/index.mjs"
check "OpenCode plugin syntax" node --check "$OPENCODE_PLUGIN/index.mjs"
check "OpenCode plugin registered" json_contains "$OPENCODE_CONFIG" "$OPENCODE_PLUGIN"

if [ "$failures" -gt 0 ]; then
  printf '\n%d check(s) failed.\n' "$failures" >&2
  exit 1
fi

printf '\nAll Cognitive Control Plane install checks passed.\n'
