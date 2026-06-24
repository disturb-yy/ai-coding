#!/bin/bash
# check-artifact.sh - Validate workflow artifact JSON
set -euo pipefail

ARTIFACT_FILE="${1:-}"
EXPECTED_TYPE="${2:-}"

if [ -z "$ARTIFACT_FILE" ]; then
    echo "USAGE: check-artifact.sh <artifact-file> [expected-artifact-type]"
    exit 2
fi

if [ ! -f "$ARTIFACT_FILE" ]; then
    echo "FAIL: file not found: $ARTIFACT_FILE"
    exit 1
fi

python3 - "$ARTIFACT_FILE" "$EXPECTED_TYPE" <<'PY'
import json
import sys
from pathlib import Path

artifact_file = Path(sys.argv[1])
expected_type = sys.argv[2]

try:
    with artifact_file.open(encoding="utf-8-sig") as f:
        data = json.load(f)
except json.JSONDecodeError as exc:
    print(f"FAIL: invalid JSON: {exc}")
    sys.exit(1)

errors = []
for field in ["artifact_type", "created_at", "phase", "content"]:
    if field not in data:
        errors.append(f"Missing: {field}")

actual = data.get("artifact_type", "")
if expected_type and actual != expected_type:
    errors.append(f"Type mismatch: expected={expected_type} got={actual}")

if not isinstance(data.get("content"), dict) or not data.get("content"):
    errors.append("Content must be a non-empty object")

if errors:
    for error in errors:
        print(f"FAIL: {error}")
    sys.exit(1)

print(f"PASS: valid {actual} artifact")
PY
