#!/bin/bash
# delegate-skill.sh - Generate task prompt for delegated skill execution
set -euo pipefail

SKILL_NAME="${1:-}"
PHASE="${2:-}"
ARTIFACT_TYPE="${3:-}"
INPUT_FILE="${4:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_HOME="$(cd "$SCRIPT_DIR/.." && pwd)"
AGENT_HOME="${WORKFLOW_AGENT_HOME:-$DEFAULT_HOME}"
SKILLS_ROOT="${WORKFLOW_AGENT_SKILLS_DIR:-$AGENT_HOME/skills}"

if [ -z "$SKILL_NAME" ] || [ -z "$PHASE" ] || [ -z "$ARTIFACT_TYPE" ]; then
    echo "USAGE: delegate-skill.sh <skill-name> <phase> <artifact-type> [input-file]"
    exit 2
fi

SKILL_DIR="$SKILLS_ROOT/$SKILL_NAME"
SKILL_MD="$SKILL_DIR/SKILL.md"
if [ ! -f "$SKILL_MD" ]; then
    echo "ERROR: $SKILL_MD not found"
    exit 1
fi

if [ -n "$INPUT_FILE" ] && [ ! -f "$INPUT_FILE" ]; then
    echo "ERROR: input file not found: $INPUT_FILE"
    exit 1
fi

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
echo "6. The artifact_type MUST be: $ARTIFACT_TYPE"
echo "7. Do NOT decide phase transitions"
echo "8. Return ONLY the JSON artifact. No extra commentary."
