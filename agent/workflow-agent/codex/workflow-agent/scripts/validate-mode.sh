#!/bin/bash
# validate-mode.sh
PHASE="$1"
if [ -z "$PHASE" ]; then echo "USAGE: validate-mode.sh <phase-name>"; exit 2; fi
python3 ~/.config/opencode/workflow-agent/scripts/_check_mode.py "$PHASE"
exit $?
