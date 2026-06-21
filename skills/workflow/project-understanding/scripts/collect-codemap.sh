#!/bin/bash
set -euo pipefail

if ! command -v mcp-call >/dev/null 2>&1; then
  echo "Error: optional CLI helper requires mcp-call." >&2
  echo "In OpenCode or other MCP-native agents, call the CodeMap MCP tools directly instead of running this script." >&2
  exit 127
fi

mkdir -p .agent/context

mcp-call get_project_info \
  > .agent/context/project.json

mcp-call list_modules \
  > .agent/context/modules.json

mcp-call search_route \
  '{"query":""}' \
  > .agent/context/routes.json

mcp-call search_flow \
  '{"query":""}' \
  > .agent/context/flows.json
