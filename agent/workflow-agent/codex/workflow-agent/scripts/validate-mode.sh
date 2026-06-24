#!/bin/bash
# validate-mode.sh
set -euo pipefail

PHASE="${1:-}"
PLATFORM="${2:-${WORKFLOW_AGENT_PLATFORM:-codex}}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -z "$PHASE" ]; then
    echo "USAGE: validate-mode.sh <phase-name> [platform]"
    exit 2
fi

python3 "$SCRIPT_DIR/_check_mode.py" "$PHASE" "$PLATFORM"
