#!/bin/bash
# delegate-skill.sh - Generate task prompt for delegated skill execution
SKILL_NAME="$1"
PHASE="$2"
ARTIFACT_TYPE="$3"
INPUT_FILE="${4:-}"
SKILL_DIR="$HOME/.config/opencode/workflow-agent/skills/$SKILL_NAME"
SKILL_MD="$SKILL_DIR/SKILL.md"
if [ ! -f "$SKILL_MD" ]; then echo "ERROR: $SKILL_MD not found"; exit 1; fi
echo "=== TASK PROMPT FOR skill=$SKILL_NAME phase=$PHASE ==="
echo ""
cat "$SKILL_MD"
echo ""
echo "## Task Instructions"
echo "1. Read the skill definition above"
echo "2. Phase: $PHASE"
echo "3. Artifact type: $ARTIFACT_TYPE"
if [ -n "$INPUT_FILE" ] && [ -f "$INPUT_FILE" ]; then
    echo "4. Input (from previous artifact):"
    cat "$INPUT_FILE"
fi
echo "5. Return a JSON artifact with envelope: {artifact_type, created_at, phase, content}"
echo "6. Do NOT decide phase transitions"
