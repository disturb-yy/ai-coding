#!/bin/bash
# check-artifact.sh - Validate workflow artifact JSON
ARTIFACT_FILE="$1"
EXPECTED_TYPE="${2:-}"
if [ ! -f "$ARTIFACT_FILE" ]; then echo "FAIL: file not found: $ARTIFACT_FILE"; exit 1; fi
python3 -c "
import json, sys
with open('$ARTIFACT_FILE') as f:
    data = json.load(f)
errors = []
for field in ['artifact_type', 'created_at', 'phase', 'content']:
    if field not in data:
        errors.append(f'Missing: {field}')
actual = data.get('artifact_type', '')
if '$EXPECTED_TYPE' and actual != '$EXPECTED_TYPE':
    errors.append(f'Type mismatch: expected=$EXPECTED_TYPE got={actual}')
if not data.get('content'):
    errors.append('Content is empty')
if errors:
    for e in errors:
        print(f'FAIL: {e}')
    sys.exit(1)
print(f'PASS: valid {actual} artifact')
" || exit 1
