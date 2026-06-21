#!/bin/bash
set -euo pipefail

context_file="${1:-.agent/context/context.json}"

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required to verify context." >&2
  exit 127
fi

if [ ! -f "$context_file" ]; then
  echo "Error: context file not found: $context_file" >&2
  echo "Run scripts/build-context.sh first, or pass a context file path." >&2
  exit 1
fi

printf "modules: "
jq '.modules|length' "$context_file"

printf "flows: "
jq '.flows|length' "$context_file"

printf "routes: "
jq '.routes|length' "$context_file"
