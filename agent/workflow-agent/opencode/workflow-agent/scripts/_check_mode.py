#!/usr/bin/env python3
import json, sys, os

if len(sys.argv) < 2:
    print("USAGE: _check_mode.py <phase-name>")
    sys.exit(2)

phase = sys.argv[1]
registry_path = os.path.expanduser("~/.config/opencode/workflow-agent/config/workflow-skills.json")

with open(registry_path) as f:
    registry = json.load(f)

for s in registry.get("skills", []):
    if s.get("phase") == phase:
        mode = s.get("execution", {}).get("opencode", "direct")
        skill = s.get("skill", "unknown")
        atype = s.get("artifact_type", "unknown")
        print(f"Phase: {phase} | Skill: {skill} | Artifact: {atype} | Mode: {mode}")
        if mode == "delegated":
            print(">>> DELEGATED - MUST use task() tool <<<")
            sys.exit(1)
        else:
            print("Direct mode")
            sys.exit(0)

print(f"PHASE_NOT_FOUND: {phase} (direct assumed)")
sys.exit(0)
