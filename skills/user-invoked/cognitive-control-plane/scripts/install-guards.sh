#!/usr/bin/env bash
set -euo pipefail

SKILL_DIR="${CCP_SKILL_DIR:-/home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane}"
CODEX_HOOK="${CCP_CODEX_HOOK:-/home/jadon/.codex/hooks/cognitive-control-plane-guard.js}"
CODEX_HOOKS_JSON="${CCP_CODEX_HOOKS_JSON:-/home/jadon/.codex/hooks.json}"
OPENCODE_PLUGIN="${CCP_OPENCODE_PLUGIN:-/home/jadon/.config/opencode/plugins/cognitive-control-plane-guard}"
OPENCODE_CONFIG="${CCP_OPENCODE_CONFIG:-/home/jadon/.config/opencode/opencode.json}"
STAMP="$(date +%Y%m%d%H%M%S)"

require_file() {
  if [ ! -f "$1" ]; then
    printf 'Missing required file: %s\n' "$1" >&2
    exit 1
  fi
}

require_file "$SKILL_DIR/scripts/check-mirrors.js"
require_file "$SKILL_DIR/scripts/cognitive-control-plane-guard.js"
require_file "$SKILL_DIR/scripts/opencode-plugin.mjs"
require_file "$SKILL_DIR/scripts/opencode-plugin-package.json"

mkdir -p "$(dirname "$CODEX_HOOK")"
cp "$SKILL_DIR/scripts/cognitive-control-plane-guard.js" "$CODEX_HOOK"
chmod +x "$CODEX_HOOK"

if [ -f "$CODEX_HOOKS_JSON" ]; then
  cp "$CODEX_HOOKS_JSON" "$CODEX_HOOKS_JSON.bak-cognitive-control-plane-$STAMP"
else
  mkdir -p "$(dirname "$CODEX_HOOKS_JSON")"
  printf '{"hooks":{}}\n' > "$CODEX_HOOKS_JSON"
fi

node "$SKILL_DIR/scripts/register-codex-hook.js" "$CODEX_HOOKS_JSON" "$CODEX_HOOK"

mkdir -p "$OPENCODE_PLUGIN"
cp "$SKILL_DIR/scripts/check-mirrors.js" "$OPENCODE_PLUGIN/check-mirrors.cjs"
cp "$SKILL_DIR/scripts/opencode-plugin.mjs" "$OPENCODE_PLUGIN/index.mjs"
cp "$SKILL_DIR/scripts/opencode-plugin-package.json" "$OPENCODE_PLUGIN/package.json"

if [ -f "$OPENCODE_CONFIG" ]; then
  cp "$OPENCODE_CONFIG" "$OPENCODE_CONFIG.bak-cognitive-control-plane-$STAMP"
else
  mkdir -p "$(dirname "$OPENCODE_CONFIG")"
  printf '{}\n' > "$OPENCODE_CONFIG"
fi

node "$SKILL_DIR/scripts/register-opencode-plugin.js" "$OPENCODE_CONFIG" "$OPENCODE_PLUGIN"

printf 'Installed Cognitive Control Plane guards.\n'
printf 'Restart OpenCode for plugin changes to take effect.\n'
