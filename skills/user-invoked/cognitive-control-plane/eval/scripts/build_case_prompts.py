#!/usr/bin/env python3
from __future__ import annotations

import argparse
from pathlib import Path

try:
    import yaml
except ImportError as exc:
    raise SystemExit("PyYAML is required. Install with: python -m pip install -r eval/requirements.txt") from exc


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--cases", required=True)
    parser.add_argument("--wrapper", required=True)
    parser.add_argument("--out", required=True)
    args = parser.parse_args()

    case_dir = Path(args.cases)
    wrapper = Path(args.wrapper).read_text(encoding="utf-8")
    out = Path(args.out)
    out.mkdir(parents=True, exist_ok=True)

    count = 0
    for case_file in sorted(case_dir.glob("*.yaml")):
        doc = yaml.safe_load(case_file.read_text(encoding="utf-8")) or {}
        for case in doc.get("cases", []):
            if case.get("static_only"):
                continue
            text = wrapper.replace("{{CASE_ID}}", case["id"]).replace("{{CASE_PROMPT}}", case["prompt"])
            (out / f"{case['id']}.md").write_text(text, encoding="utf-8")
            count += 1

    print(f"Generated {count} case prompts in {out}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
