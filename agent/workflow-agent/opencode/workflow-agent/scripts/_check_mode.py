#!/usr/bin/env python3
import json
import os
import sys
from pathlib import Path

if len(sys.argv) < 2:
    print("USAGE: _check_mode.py <phase-name> [platform]")
    sys.exit(2)

phase = sys.argv[1]
platform = sys.argv[2] if len(sys.argv) > 2 else os.environ.get("WORKFLOW_AGENT_PLATFORM", "opencode")
script_dir = Path(__file__).resolve().parent
default_home = script_dir.parent
agent_home = Path(os.environ.get("WORKFLOW_AGENT_HOME", default_home)).expanduser()
registry_path = Path(
    os.environ.get("WORKFLOW_AGENT_REGISTRY", agent_home / "config" / "workflow-skills.json")
).expanduser()

if not registry_path.is_file():
    print(f"ERROR: registry not found: {registry_path}")
    sys.exit(2)

with registry_path.open(encoding="utf-8") as f:
    registry = json.load(f)

for s in registry.get("skills", []):
    if s.get("phase") != phase or not s.get("enabled", True):
        continue

    mode = s.get("execution", {}).get(platform, "direct")
    skill = s.get("skill", "unknown")
    atype = s.get("artifact_type", "unknown")
    print(f"Phase: {phase} | Platform: {platform} | Skill: {skill} | Artifact: {atype} | Mode: {mode}")
    if mode == "delegated":
        print(">>> DELEGATED - MUST use task() tool <<<")
        sys.exit(1)

    print("Direct mode")
    sys.exit(0)

print(f"PHASE_NOT_FOUND: {phase} (direct assumed)")
sys.exit(0)
