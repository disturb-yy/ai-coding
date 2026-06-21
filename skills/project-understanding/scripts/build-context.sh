#!/bin/bash
set -euo pipefail

context_dir="${1:-.agent/context}"
output_file="${2:-$context_dir/context.json}"

if ! command -v jq >/dev/null 2>&1; then
  echo "Error: jq is required to build the merged context." >&2
  exit 127
fi

jq -s '
{
  project:.[0],
  modules:.[1],
  routes:.[2],
  flows:.[3]
}
' \
"$context_dir/project.json" \
"$context_dir/modules.json" \
"$context_dir/routes.json" \
"$context_dir/flows.json" \
> "$output_file"

echo "Wrote $output_file"
