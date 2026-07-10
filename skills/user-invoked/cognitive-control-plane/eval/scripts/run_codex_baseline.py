#!/usr/bin/env python3
from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Any

try:
    import yaml
except ImportError as exc:
    raise SystemExit(
        "PyYAML is required. Install with: python -m pip install -r eval/requirements.txt"
    ) from exc


def load_cases(case_dir: Path) -> list[dict[str, Any]]:
    cases: list[dict[str, Any]] = []
    seen: set[str] = set()
    for path in sorted(case_dir.glob("*.yaml")):
        doc = yaml.safe_load(path.read_text(encoding="utf-8")) or {}
        for case in doc.get("cases", []):
            case_id = case["id"]
            if case_id in seen:
                raise ValueError(f"Duplicate case id: {case_id}")
            seen.add(case_id)
            if not case.get("static_only"):
                cases.append(case)
    return cases


def select_cases(
    cases: list[dict[str, Any]],
    ids: list[str],
    categories: list[str],
    limit: int | None,
) -> list[dict[str, Any]]:
    selected = cases
    if ids:
        wanted = set(ids)
        selected = [c for c in selected if c["id"] in wanted]
        missing = sorted(wanted - {c["id"] for c in selected})
        if missing:
            raise ValueError(f"Unknown case ids: {', '.join(missing)}")
    if categories:
        wanted_categories = set(categories)
        selected = [c for c in selected if c.get("category") in wanted_categories]
    if limit is not None:
        selected = selected[:limit]
    return selected


def ensure_workspace(workspace: Path, skill_dir: Path) -> None:
    workspace.mkdir(parents=True, exist_ok=True)
    if not (workspace / ".git").exists():
        subprocess.run(["git", "init", "-q"], cwd=workspace, check=True)

    skill_parent = workspace / ".agents" / "skills"
    skill_parent.mkdir(parents=True, exist_ok=True)
    link = skill_parent / "cognitive-control-plane"

    if link.is_symlink() or link.exists():
        if link.is_symlink() and link.resolve() == skill_dir.resolve():
            return
        if link.is_dir() and not link.is_symlink():
            shutil.rmtree(link)
        else:
            link.unlink()

    link.symlink_to(skill_dir.resolve(), target_is_directory=True)


def codex_version() -> str:
    try:
        proc = subprocess.run(
            ["codex", "--version"], text=True, capture_output=True, check=False
        )
        return (proc.stdout or proc.stderr).strip()
    except FileNotFoundError:
        return "codex-not-found"


def extract_codex_error(raw_path: Path, stderr: str) -> str:
    stderr = (stderr or "").strip()
    if stderr:
        return stderr.splitlines()[-1][:1000]

    try:
        lines = [line.strip() for line in raw_path.read_text(encoding="utf-8").splitlines()]
    except OSError as exc:
        return f"no stderr and raw output unreadable: {exc}"

    for line in reversed(lines):
        if not line:
            continue
        try:
            item = json.loads(line)
        except json.JSONDecodeError:
            return line[:1000]
        for path in (
            ("error", "message"),
            ("message",),
            ("item", "error"),
            ("item", "aggregated_output"),
        ):
            cur: Any = item
            for part in path:
                if not isinstance(cur, dict) or part not in cur:
                    cur = None
                    break
                cur = cur[part]
            if cur:
                return str(cur).splitlines()[-1][:1000]
        return json.dumps(item, ensure_ascii=False)[:1000]

    return "codex failed without stderr or raw json output"


def fallback_result(case_id: str, error: str) -> dict[str, Any]:
    return {
        "case_id": case_id,
        "evidence_source": "self_report",
        "trace": {
            "activated": False,
            "classification": "Tiny",
            "active_surface": "none",
            "surfaces_used": [],
            "references_read": [],
            "orchestration_used": False,
            "dependency_graph_created": False,
            "persistent_state_used": False,
            "required_skills": [],
            "next_action": "direct_answer",
            "asked_user_question": False,
            "strict_schema_during_exploration": False,
            "stopped_routing": False,
            "task_contract_complete": False,
            "ownership_conflict": False,
            "behaviors": [],
        },
        "response": f"EVAL_EXECUTION_ERROR: {error}",
        "runner_error": error,
    }


def parse_final_json(path: Path, case_id: str) -> dict[str, Any]:
    raw = path.read_text(encoding="utf-8").strip()
    try:
        data = json.loads(raw)
    except json.JSONDecodeError as exc:
        raise ValueError(f"Final output is not valid JSON: {exc}") from exc

    if data.get("case_id") != case_id:
        raise ValueError(
            f"case_id mismatch: expected {case_id!r}, got {data.get('case_id')!r}"
        )
    data["evidence_source"] = "self_report"
    return data


def load_raw_events(path: Path) -> list[dict[str, Any]]:
    events: list[dict[str, Any]] = []
    if not path.exists():
        return events

    for line in path.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        try:
            events.append(json.loads(line))
        except json.JSONDecodeError as exc:
            events.append({"type": "malformed_jsonl", "error": str(exc), "line": line})
    return events


def normalize_raw_event(event: dict[str, Any]) -> dict[str, Any] | None:
    raw_type = str(event.get("type") or "unknown")
    item = event.get("item")

    if isinstance(item, dict):
        item_type = str(item.get("type") or raw_type)
        if item_type == "agent_message":
            return {
                "role": "assistant",
                "kind": "message",
                "raw_type": raw_type,
                "content": str(item.get("text") or ""),
                "metadata": {"item_id": item.get("id")},
            }
        if item_type == "command_execution":
            return {
                "role": "tool",
                "kind": "command_execution",
                "raw_type": raw_type,
                "content": str(item.get("aggregated_output") or ""),
                "metadata": {
                    "item_id": item.get("id"),
                    "command": item.get("command"),
                    "exit_code": item.get("exit_code"),
                    "status": item.get("status"),
                },
            }
        return {
            "role": "system",
            "kind": item_type,
            "raw_type": raw_type,
            "content": json.dumps(item, ensure_ascii=False),
            "metadata": {"item_id": item.get("id")},
        }

    if raw_type in {"thread.started", "turn.started", "turn.completed"}:
        return {
            "role": "system",
            "kind": raw_type.replace(".", "_"),
            "raw_type": raw_type,
            "content": "",
            "metadata": {k: v for k, v in event.items() if k != "type"},
        }

    if raw_type == "malformed_jsonl":
        return {
            "role": "system",
            "kind": "malformed_jsonl",
            "raw_type": raw_type,
            "content": str(event.get("line") or ""),
            "metadata": {"error": event.get("error")},
        }

    return {
        "role": "system",
        "kind": raw_type.replace(".", "_"),
        "raw_type": raw_type,
        "content": json.dumps(event, ensure_ascii=False),
        "metadata": {},
    }


def record_model_conversation(
    *,
    case_id: str,
    prompt: str,
    raw_path: Path,
    final_path: Path,
    conversation_path: Path,
    status: str,
    runner_error: str | None,
) -> dict[str, Any]:
    raw_events = load_raw_events(raw_path)
    messages = [
        {
            "role": "user",
            "kind": "eval_prompt",
            "raw_type": "runner.prompt",
            "content": prompt,
            "metadata": {"case_id": case_id},
        }
    ]
    for event in raw_events:
        message = normalize_raw_event(event)
        if message is not None:
            messages.append(message)

    payload = {
        "schema_version": "1.0",
        "case_id": case_id,
        "executor": "codex-cli",
        "created_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "status": status,
        "runner_error": runner_error,
        "prompt": {
            "role": "user",
            "content": prompt,
        },
        "raw_event_count": len(raw_events),
        "artifacts": {
            "raw_events": str(raw_path),
            "final_result": str(final_path),
        },
        "messages": messages,
    }
    conversation_path.write_text(
        json.dumps(payload, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8",
    )
    return payload


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Run the cognitive-control-plane eval pack with Codex CLI."
    )
    parser.add_argument("--skill-dir", required=True)
    parser.add_argument("--eval-dir")
    parser.add_argument("--workspace")
    parser.add_argument("--run-id")
    parser.add_argument("--model")
    parser.add_argument(
        "--sandbox",
        choices=["read-only", "workspace-write", "danger-full-access"],
        default="read-only",
        help="Sandbox mode passed to codex exec. Default preserves the isolated baseline design.",
    )
    parser.add_argument("--case", action="append", default=[])
    parser.add_argument("--category", action="append", default=[])
    parser.add_argument("--limit", type=int)
    parser.add_argument("--timeout", type=int, default=600)
    parser.add_argument(
        "--use-user-config",
        action="store_true",
        help="Keep ~/.codex/config.toml and execpolicy rules. Default is isolated.",
    )
    parser.add_argument(
        "--stop-on-error",
        action="store_true",
        help="Stop immediately when one Codex case fails.",
    )
    args = parser.parse_args()

    if shutil.which("codex") is None:
        raise SystemExit("codex command not found in PATH")

    skill_dir = Path(args.skill_dir).resolve()
    eval_dir = (
        Path(args.eval_dir).resolve()
        if args.eval_dir
        else (skill_dir / "eval").resolve()
    )
    case_dir = eval_dir / "cases"
    wrapper_path = eval_dir / "prompts" / "codex-baseline-wrapper.md"
    schema_path = eval_dir / "schemas" / "result.schema.json"
    score_script = eval_dir / "scripts" / "score.py"
    static_script = eval_dir / "scripts" / "static_checks.py"

    for required in [case_dir, wrapper_path, schema_path, score_script, static_script]:
        if not required.exists():
            raise SystemExit(f"Missing eval asset: {required}")

    all_cases = load_cases(case_dir)
    selected = select_cases(
        all_cases,
        ids=args.case,
        categories=args.category,
        limit=args.limit,
    )
    if not selected:
        raise SystemExit("No cases selected")

    run_id = args.run_id or dt.datetime.now().strftime("baseline-%Y%m%d-%H%M%S")
    results_root = eval_dir / "results" / run_id
    raw_dir = results_root / "raw"
    final_dir = results_root / "final"
    conversation_dir = results_root / "conversations"
    raw_dir.mkdir(parents=True, exist_ok=True)
    final_dir.mkdir(parents=True, exist_ok=True)
    conversation_dir.mkdir(parents=True, exist_ok=True)

    workspace = (
        Path(args.workspace).resolve()
        if args.workspace
        else (eval_dir / ".codex-baseline-workspace").resolve()
    )
    ensure_workspace(workspace, skill_dir)

    wrapper = wrapper_path.read_text(encoding="utf-8")
    combined_path = results_root / "results.jsonl"
    metadata_path = results_root / "run.json"
    report_json = results_root / "report.json"
    report_md = results_root / "report.md"
    static_json = results_root / "static-checks.json"

    static_proc = subprocess.run(
        [
            sys.executable,
            str(static_script),
            "--skill-dir",
            str(skill_dir),
            "--json-out",
            str(static_json),
        ],
        cwd=workspace,
        text=True,
        capture_output=True,
        check=False,
    )
    print(static_proc.stdout, end="")
    if static_proc.returncode != 0:
        print(static_proc.stderr, file=sys.stderr, end="")
        print(
            "Static checks failed. Runtime baseline will still run so semantic failures stay visible.",
            file=sys.stderr,
        )

    started_at = dt.datetime.now(dt.timezone.utc).isoformat()
    metadata = {
        "run_id": run_id,
        "skill_version": "git-working-tree",
        "executor": {
            "type": "codex-cli",
            "version": codex_version(),
            "model": args.model or "config-default",
            "isolated_user_config": not args.use_user_config,
            "sandbox": args.sandbox,
            "evidence_source": "self_report",
        },
        "case_count": len(selected),
        "selected_case_ids": [c["id"] for c in selected],
        "started_at": started_at,
        "finished_at": None,
        "workspace": str(workspace),
        "skill_dir": str(skill_dir),
        "artifacts": {
            "raw_events_dir": str(raw_dir),
            "final_results_dir": str(final_dir),
            "model_conversations_dir": str(conversation_dir),
            "results_jsonl": str(combined_path),
            "report_json": str(report_json),
            "report_md": str(report_md),
            "static_checks_json": str(static_json),
        },
    }
    metadata_path.write_text(
        json.dumps(metadata, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8",
    )

    original_home = Path.home()
    isolated_home = workspace / ".eval-home"
    isolated_home.mkdir(parents=True, exist_ok=True)

    env = os.environ.copy()
    if not args.use_user_config:
        env.setdefault("CODEX_HOME", str(original_home / ".codex"))
        env["HOME"] = str(isolated_home)

    results: list[dict[str, Any]] = []
    error_count = 0

    for index, case in enumerate(selected, start=1):
        case_id = case["id"]
        print(f"[{index}/{len(selected)}] {case_id} — {case.get('title', '')}")
        should_stop = False

        prompt = (
            wrapper.replace("{{CASE_ID}}", case_id)
            .replace("{{CASE_PROMPT}}", case["prompt"])
        )

        raw_path = raw_dir / f"{case_id}.jsonl"
        final_path = final_dir / f"{case_id}.json"
        conversation_path = conversation_dir / f"{case_id}.json"
        case_status = "completed"
        case_error: str | None = None

        # `--ask-for-approval` is a global Codex flag. Some CLI versions
        # reject it after the `exec` subcommand, so keep it before `exec`.
        cmd = [
            "codex",
            "--ask-for-approval",
            "never",
            "exec",
            "--ephemeral",
            "--sandbox",
            args.sandbox,
            "--cd",
            str(workspace),
            "--json",
            "--output-schema",
            str(schema_path),
            "--output-last-message",
            str(final_path),
        ]
        if not args.use_user_config:
            cmd.extend(["--ignore-user-config", "--ignore-rules"])
        if args.model:
            cmd.extend(["--model", args.model])
        cmd.append("-")

        try:
            with raw_path.open("w", encoding="utf-8") as raw_file:
                proc = subprocess.run(
                    cmd,
                    input=prompt,
                    text=True,
                    stdout=raw_file,
                    stderr=subprocess.PIPE,
                    cwd=workspace,
                    env=env,
                    timeout=args.timeout,
                    check=False,
                )

            if proc.returncode != 0:
                detail = extract_codex_error(raw_path, proc.stderr)
                raise RuntimeError(f"codex exit {proc.returncode}: {detail}")

            result = parse_final_json(final_path, case_id)
        except Exception as exc:
            error_count += 1
            case_status = "error"
            case_error = str(exc)
            result = fallback_result(case_id, str(exc))
            final_path.write_text(
                json.dumps(result, indent=2, ensure_ascii=False) + "\n",
                encoding="utf-8",
            )
            print(f"  ERROR: {exc}", file=sys.stderr)
            if args.stop_on_error:
                should_stop = True

        record_model_conversation(
            case_id=case_id,
            prompt=prompt,
            raw_path=raw_path,
            final_path=final_path,
            conversation_path=conversation_path,
            status=case_status,
            runner_error=case_error,
        )
        results.append(result)
        if should_stop:
            break

    combined_path.write_text(
        "\n".join(json.dumps(item, ensure_ascii=False) for item in results) + "\n",
        encoding="utf-8",
    )

    metadata["finished_at"] = dt.datetime.now(dt.timezone.utc).isoformat()
    metadata["runtime_error_count"] = error_count
    metadata["result_count"] = len(results)
    metadata["conversation_count"] = len(list(conversation_dir.glob("*.json")))
    metadata_path.write_text(
        json.dumps(metadata, indent=2, ensure_ascii=False) + "\n",
        encoding="utf-8",
    )

    score_proc = subprocess.run(
        [
            sys.executable,
            str(score_script),
            "--cases",
            str(case_dir),
            "--results",
            str(combined_path),
            "--out-json",
            str(report_json),
            "--out-md",
            str(report_md),
            *sum((["--case", case["id"]] for case in selected), []),
        ],
        cwd=workspace,
        text=True,
        capture_output=True,
        check=False,
    )

    print(score_proc.stdout, end="")
    if score_proc.stderr:
        print(score_proc.stderr, file=sys.stderr, end="")

    print()
    print(f"Run directory: {results_root}")
    print(f"Report:        {report_md}")
    print(f"Raw events:    {raw_dir}")
    print(f"Conversations: {conversation_dir}")
    print(f"Final results: {combined_path}")
    print()
    print(
        "Baseline evidence level: self_report. Raw Codex JSONL events are retained "
        "for a later runtime-trace extractor."
    )

    return score_proc.returncode


if __name__ == "__main__":
    raise SystemExit(main())
