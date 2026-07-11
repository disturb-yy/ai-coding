#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
from collections import Counter, defaultdict
from pathlib import Path
from typing import Any

try:
    import yaml
except ImportError as exc:
    raise SystemExit("PyYAML is required. Install with: python -m pip install -r eval/requirements.txt") from exc


def load_cases(case_dir: Path) -> dict[str, dict]:
    cases: dict[str, dict] = {}
    for path in sorted(case_dir.glob("*.yaml")):
        doc = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
        for case in doc.get("cases", []):
            case_id = case["id"]
            if case_id in cases:
                raise ValueError(f"Duplicate case id: {case_id}")
            cases[case_id] = case
    return cases


def load_results(path: Path) -> dict[str, dict]:
    results: dict[str, dict] = {}
    for idx, line in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
        if not line.strip():
            continue
        item = json.loads(line)
        case_id = item["case_id"]
        if case_id in results:
            raise ValueError(f"Duplicate result for {case_id} on line {idx}")
        results[case_id] = item
    return results


def get_path(obj: Any, dotted: str) -> Any:
    cur = obj
    for part in dotted.split("."):
        if not isinstance(cur, dict) or part not in cur:
            return None
        cur = cur[part]
    return cur


def reference_path_matches(actual: Any, expected: Any) -> bool:
    actual_text = str(actual).replace("\\", "/")
    expected_text = str(expected).replace("\\", "/")
    return actual_text == expected_text or actual_text.endswith(f"/{expected_text}")


def contains_all(actual: Any, expected: Any, path: str | None = None) -> bool:
    if not isinstance(actual, list):
        return False
    if path == "trace.references_read":
        return all(
            any(reference_path_matches(item, value) for item in actual)
            for value in expected
        )
    return all(v in actual for v in expected)


def evaluate(actual: Any, op: str, expected: Any, path: str | None = None) -> bool:
    if op == "eq":
        return actual == expected
    if op == "neq":
        return actual != expected
    if op == "in":
        return actual in expected
    if op == "not_in":
        return actual not in expected
    if op == "contains_all":
        return contains_all(actual, expected, path)
    if op == "excludes_all":
        return isinstance(actual, list) and all(v not in actual for v in expected)
    if op == "count_max":
        return isinstance(actual, list) and len(actual) <= int(expected)
    raise ValueError(f"Unsupported op: {op}")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--cases", required=True)
    parser.add_argument("--results", required=True)
    parser.add_argument("--out-json")
    parser.add_argument("--out-md")
    parser.add_argument(
        "--case",
        action="append",
        default=[],
        help="Score only the selected case id. Repeat for multiple cases.",
    )
    args = parser.parse_args()

    cases = load_cases(Path(args.cases))
    if args.case:
        selected_ids = set(args.case)
        unknown = sorted(selected_ids - set(cases))
        if unknown:
            raise SystemExit(f"Unknown case ids: {', '.join(unknown)}")
        cases = {case_id: case for case_id, case in cases.items() if case_id in selected_ids}
    results = load_results(Path(args.results))

    assertion_total = 0
    assertion_pass = 0
    hard_failures = []
    failures = []
    missing_results = []
    by_rubric = defaultdict(lambda: {"pass": 0, "fail": 0})
    by_category = defaultdict(lambda: {"pass": 0, "fail": 0})
    failure_modes = Counter()

    for case_id, case in cases.items():
        if case.get("static_only"):
            continue
        result = results.get(case_id)
        if result is None:
            missing_results.append(case_id)
            continue

        case_failed = False
        for assertion in case.get("auto_assertions", []):
            assertion_total += 1
            actual = get_path(result, assertion["path"])
            ok = evaluate(actual, assertion["op"], assertion.get("value"), assertion["path"])
            rubric = assertion.get("rubric") or "unmapped"
            if ok:
                assertion_pass += 1
                by_rubric[rubric]["pass"] += 1
            else:
                case_failed = True
                by_rubric[rubric]["fail"] += 1
                failure = {
                    "case_id": case_id,
                    "title": case.get("title"),
                    "category": case.get("category"),
                    "rubric": rubric,
                    "path": assertion["path"],
                    "op": assertion["op"],
                    "expected": assertion.get("value"),
                    "actual": actual,
                    "hard_fail": bool(assertion.get("hard_fail")),
                }
                failures.append(failure)
                if failure["hard_fail"]:
                    hard_failures.append(failure)

        if case_failed:
            by_category[case.get("category","unknown")]["fail"] += 1
        else:
            by_category[case.get("category","unknown")]["pass"] += 1

        judge = result.get("judge") or {}
        for flag in judge.get("flags", []):
            failure_modes[flag] += 1
        if judge.get("failure_mode"):
            failure_modes[judge["failure_mode"]] += 1

    auto_score = assertion_pass / assertion_total if assertion_total else 0.0
    report = {
        "case_count": len([c for c in cases.values() if not c.get("static_only")]),
        "result_count": len(results),
        "missing_result_count": len(missing_results),
        "missing_results": missing_results,
        "assertion_total": assertion_total,
        "assertion_pass": assertion_pass,
        "auto_score": round(auto_score, 4),
        "hard_fail_count": len(hard_failures),
        "hard_failures": hard_failures,
        "failure_count": len(failures),
        "failures": failures,
        "by_rubric": dict(by_rubric),
        "by_category": dict(by_category),
        "judge_failure_modes": dict(failure_modes),
    }

    payload = json.dumps(report, indent=2, ensure_ascii=False)
    print(payload)

    if args.out_json:
        Path(args.out_json).write_text(payload + "\n", encoding="utf-8")

    if args.out_md:
        lines = [
            "# Cognitive Control Plane Eval Report",
            "",
            f"- Auto score: **{auto_score:.1%}**",
            f"- Assertions: **{assertion_pass}/{assertion_total}**",
            f"- Hard fails: **{len(hard_failures)}**",
            f"- Missing results: **{len(missing_results)}**",
            "",
            "## Failures",
            "",
        ]
        if not failures:
            lines.append("No automatic assertion failures.")
        else:
            lines.append("| Case | Rubric | Assertion | Expected | Actual | Hard fail |")
            lines.append("|---|---|---|---|---|---|")
            for f in failures:
                lines.append(
                    f"| {f['case_id']} | {f['rubric']} | `{f['path']} {f['op']}` | "
                    f"`{json.dumps(f['expected'], ensure_ascii=False)}` | "
                    f"`{json.dumps(f['actual'], ensure_ascii=False)}` | "
                    f"{'YES' if f['hard_fail'] else 'no'} |"
                )
        lines.extend(["", "## By rubric", "", "| Rubric | Pass | Fail |", "|---|---:|---:|"])
        for rid, counts in sorted(by_rubric.items()):
            lines.append(f"| {rid} | {counts['pass']} | {counts['fail']} |")
        Path(args.out_md).write_text("\n".join(lines) + "\n", encoding="utf-8")

    return 0 if not hard_failures and not missing_results else 1


if __name__ == "__main__":
    raise SystemExit(main())
