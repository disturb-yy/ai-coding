#!/usr/bin/env python3
from __future__ import annotations

import argparse
import importlib.util
import json
import os
import re
import shutil
import subprocess
import tempfile
from pathlib import Path

try:
    import yaml
except ImportError as exc:
    raise SystemExit("PyYAML is required. Install with: python -m pip install -r eval/requirements.txt") from exc


def check(condition: bool, check_id: str, message: str, failures: list[dict]) -> None:
    if not condition:
        failures.append({"check_id": check_id, "message": message})


def list_contains_all(actual: object, expected: set[str]) -> tuple[bool, list[str]]:
    if not isinstance(actual, list):
        return False, sorted(expected)
    missing = sorted(expected - set(actual))
    return not missing, missing


def strict_schema_problems(node: object, path: str = "$") -> list[str]:
    problems: list[str] = []
    if isinstance(node, dict):
        if node.get("type") == "object":
            properties = node.get("properties") or {}
            required = set(node.get("required") or [])
            property_names = set(properties.keys())
            if node.get("additionalProperties") is not False:
                problems.append(f"{path}: additionalProperties must be false")
            missing_required = sorted(property_names - required)
            extra_required = sorted(required - property_names)
            if missing_required:
                problems.append(f"{path}: properties not listed in required: {missing_required}")
            if extra_required:
                problems.append(f"{path}: required names missing from properties: {extra_required}")
        for key, value in node.items():
            problems.extend(strict_schema_problems(value, f"{path}.{key}"))
    elif isinstance(node, list):
        for index, value in enumerate(node):
            problems.extend(strict_schema_problems(value, f"{path}[{index}]"))
    return problems


def import_python_module(path: Path, module_name: str):
    spec = importlib.util.spec_from_file_location(module_name, path)
    if spec is None or spec.loader is None:
        raise ImportError(f"Cannot import module from {path}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def check_conversation_recorder(runner_path: Path) -> tuple[bool, str]:
    module = import_python_module(runner_path, "ccp_eval_runner_static_check")
    if not hasattr(module, "record_model_conversation"):
        return False, "record_model_conversation helper is missing"

    with tempfile.TemporaryDirectory() as tmp:
        root = Path(tmp)
        raw_path = root / "ACP-TEST.jsonl"
        final_path = root / "ACP-TEST.json"
        conversation_path = root / "ACP-TEST-conversation.json"
        raw_path.write_text(
            "\n".join(
                [
                    json.dumps({"type": "thread.started", "thread_id": "t1"}),
                    json.dumps({
                        "type": "item.completed",
                        "item": {
                            "id": "m1",
                            "type": "agent_message",
                            "text": "assistant reply",
                        },
                    }),
                    json.dumps({
                        "type": "item.completed",
                        "item": {
                            "id": "c1",
                            "type": "command_execution",
                            "command": "echo ok",
                            "aggregated_output": "ok\n",
                            "exit_code": 0,
                            "status": "completed",
                        },
                    }),
                ]
            )
            + "\n",
            encoding="utf-8",
        )
        final_path.write_text('{"case_id":"ACP-TEST"}\n', encoding="utf-8")

        payload = module.record_model_conversation(
            case_id="ACP-TEST",
            prompt="user prompt",
            raw_path=raw_path,
            final_path=final_path,
            conversation_path=conversation_path,
            status="completed",
            runner_error=None,
        )
        stored = json.loads(conversation_path.read_text(encoding="utf-8"))
        roles = [message.get("role") for message in stored.get("messages", [])]
        kinds = [message.get("kind") for message in stored.get("messages", [])]
        if payload.get("raw_event_count") != 3 or stored.get("raw_event_count") != 3:
            return False, "conversation raw_event_count did not match raw JSONL events"
        if roles[:1] != ["user"] or "assistant" not in roles or "tool" not in roles:
            return False, f"conversation roles are incomplete: {roles}"
        if "eval_prompt" not in kinds or "message" not in kinds or "command_execution" not in kinds:
            return False, f"conversation kinds are incomplete: {kinds}"
        artifacts = stored.get("artifacts") or {}
        if not artifacts.get("raw_events") or not artifacts.get("final_result"):
            return False, "conversation artifacts do not link raw and final outputs"

    return True, ""


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--skill-dir", required=True)
    parser.add_argument("--json-out")
    args = parser.parse_args()

    skill = Path(args.skill_dir).resolve()
    failures: list[dict] = []
    passes: list[str] = []

    skill_md = skill / "SKILL.md"
    check(skill_md.exists(), "S-01", f"Missing {skill_md}", failures)
    if not skill_md.exists():
        return 2

    text = skill_md.read_text(encoding="utf-8")

    # Referenced local markdown files must exist.
    links = re.findall(r"\]\((references/[^)#]+\.md)\)", text)
    for rel in sorted(set(links)):
        target = skill / rel
        check(target.exists(), "S-02", f"Referenced file does not exist: {rel}", failures)

    # Core route references.
    required_refs = {
        "references/context-control.md",
        "references/epistemic-control.md",
        "references/adversarial-control.md",
        "references/output-control.md",
        "references/orchestration-state.md",
        "references/skill-orchestration.md",
        "references/maintenance.md",
    }
    for rel in required_refs:
        check((skill / rel).exists(), "S-03", f"Missing core reference: {rel}", failures)

    # Machine map consistency.
    map_path = skill / "config" / "skill-orchestration-map.yaml"
    check(map_path.exists(), "S-04", "Missing skill-orchestration-map.yaml", failures)
    if map_path.exists():
        data = yaml.safe_load(map_path.read_text(encoding="utf-8")) or {}
        registry = set((data.get("skill_registry") or {}).keys())
        route_skills = {r.get("skill") for r in (data.get("routing_order") or []) if r.get("skill")}
        check(route_skills.issubset(registry), "S-05", f"routing_order contains unregistered skills: {sorted(route_skills - registry)}", failures)
        expected = {"grilling", "diagnosing-problem", "exploring-project", "reviewing-code", "coding-project", "coding-tdd"}
        check(expected.issubset(registry), "S-06", f"Missing expected skill registry entries: {sorted(expected - registry)}", failures)
        path_entries = []
        def collect_paths(node: object, prefix: str = "$") -> None:
            if isinstance(node, dict):
                for key, value in node.items():
                    current = f"{prefix}.{key}"
                    if key == "path":
                        path_entries.append(current)
                    collect_paths(value, current)
            elif isinstance(node, list):
                for index, value in enumerate(node):
                    collect_paths(value, f"{prefix}[{index}]")
        collect_paths(data)
        check(not path_entries, "S-15", f"Config map must not hardcode path fields: {path_entries}", failures)

        contract_schema = data.get("contract_schema") or {}
        schema_must_report = set(contract_schema.get("must_report") or [])
        defaults = data.get("subagent_contract_defaults") or {}
        defaults_must_report_ok, defaults_must_report_missing = list_contains_all(
            defaults.get("must_report"),
            schema_must_report,
        )
        check(
            isinstance(defaults, dict) and defaults_must_report_ok,
            "S-16",
            "subagent_contract_defaults.must_report must include contract_schema.must_report: "
            f"{defaults_must_report_missing}",
            failures,
        )

        default_stop_if = {
            "required_skill_unavailable",
            "required_reference_unavailable",
            "required_mcp_or_tool_unavailable",
            "scope_unclear",
            "ownership_conflict",
            "validation_blocked",
        }
        stop_if_ok, stop_if_missing = list_contains_all(defaults.get("stop_if"), default_stop_if)
        check(
            stop_if_ok,
            "S-17",
            f"subagent_contract_defaults.stop_if missing required stop conditions: {stop_if_missing}",
            failures,
        )

        memory_policy = defaults.get("memory_policy") or {}
        event_fields_ok, event_fields_missing = list_contains_all(
            memory_policy.get("event_log_fields"),
            {"timestamp", "actor", "task_id", "event_type", "summary", "next_action"},
        )
        event_types_ok, event_types_missing = list_contains_all(
            memory_policy.get("event_types"),
            {"started", "blocked", "completed", "decision", "validation", "handoff"},
        )
        check(
            bool(memory_policy.get("persist_event")) and event_fields_ok and event_types_ok,
            "S-18",
            "subagent_contract_defaults.memory_policy must persist resumable event logs; "
            f"missing fields={event_fields_missing}, missing types={event_types_missing}",
            failures,
        )

        parallelization = data.get("parallelization") or {}
        missing_parallelization = sorted(registry - set(parallelization.keys()))
        check(
            not missing_parallelization,
            "S-19",
            f"parallelization is missing registry entries: {missing_parallelization}",
            failures,
        )
        check(
            parallelization.get("coding-project") == "write_parallel_only_with_non_overlapping_ownership",
            "S-20",
            "coding-project parallelization must preserve non-overlapping write ownership.",
            failures,
        )
        check(
            parallelization.get("coding-tdd") == "parallel_only_for_independent_functions_or_modules",
            "S-21",
            "coding-tdd parallelization must limit parallel work to independent functions or modules.",
            failures,
        )

        reconciliation_checks_ok, reconciliation_missing = list_contains_all(
            data.get("reconciliation_checks"),
            {
                "required_skills_loaded",
                "required_references_loaded",
                "required_mcp_and_tools_used",
                "skill_instructions_followed_reported",
                "deviations_justified",
                "expected_fields_present",
                "ownership_respected",
                "validation_evidence_present",
            },
        )
        check(
            reconciliation_checks_ok,
            "S-22",
            f"reconciliation_checks missing required acceptance checks: {reconciliation_missing}",
            failures,
        )

        adapter_contract = data.get("adapter_contract") or {}
        schema_ref = adapter_contract.get("schema_ref")
        check(
            isinstance(schema_ref, str) and (skill / schema_ref).exists(),
            "S-23",
            f"adapter_contract.schema_ref is missing or does not exist: {schema_ref}",
            failures,
        )
        adapter_actions_ok, adapter_actions_missing = list_contains_all(
            adapter_contract.get("required_for_next_actions"),
            {"delegate_read_only", "delegate_write"},
        )
        check(
            adapter_actions_ok,
            "S-24",
            f"adapter_contract.required_for_next_actions missing delegate actions: {adapter_actions_missing}",
            failures,
        )
        delegation_requirements_ok, delegation_requirements_missing = list_contains_all(
            adapter_contract.get("real_delegation_requires"),
            {
                "adapter_result.status_started",
                "adapter_result.task_id_present",
                "adapter_result.subagent_started_true",
            },
        )
        check(
            delegation_requirements_ok,
            "S-25",
            f"adapter_contract.real_delegation_requires missing requirements: {delegation_requirements_missing}",
            failures,
        )
        platform_adapters = adapter_contract.get("platform_adapters") or {}
        for platform in ["opencode", "codex", "claude-code"]:
            rel = platform_adapters.get(platform)
            check(
                isinstance(rel, str) and (skill / rel).exists(),
                "S-26",
                f"adapter_contract.platform_adapters.{platform} missing or does not exist: {rel}",
                failures,
            )

    adapter_cli = skill / "scripts" / "ccp-adapter.js"
    check(adapter_cli.exists(), "S-27", "Missing scripts/ccp-adapter.js", failures)
    adapter_schema = skill / "adapters" / "contract.schema.json"
    check(adapter_schema.exists(), "S-28", "Missing adapters/contract.schema.json", failures)
    for rel in ["README.md", "opencode.md", "codex.md", "claude-code.md"]:
        check((skill / "adapters" / rel).exists(), "S-29", f"Missing adapter guide: adapters/{rel}", failures)
    if adapter_schema.exists():
        try:
            json.loads(adapter_schema.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            failures.append({
                "check_id": "S-28",
                "message": f"Adapter contract schema is not valid JSON: {exc}",
            })

    # Maintenance policy must preserve canonical/mirror invariants.
    maintenance = skill / "references" / "maintenance.md"
    if maintenance.exists():
        mt = maintenance.read_text(encoding="utf-8")
        check("Never read, search, open" in mt and "Chinese mirror" in mt, "S-07", "Mirror read prohibition is missing or weakened.", failures)
        check("overwrite the matching Chinese mirror" in mt, "S-08", "Canonical-to-mirror synchronization rule is missing.", failures)

    # Guard behavior checks. Avoid implementation-name matching because the
    # guard may be refactored without changing behavior.
    guard = skill / "scripts" / "cognitive-control-plane-guard.js"
    check(guard.exists(), "S-09", "Missing cognitive-control-plane-guard.js", failures)
    if guard.exists() and shutil.which("node"):
        guard_env = dict(os.environ)
        guard_env["CCP_SKILL_DIR"] = str(skill)

        mirror_payload = json.dumps({
            "hook_event_name": "PreToolUse",
            "tool_name": "read",
            "tool_input": {"path": str(skill / "zh" / "SKILL.zh-CN.md")},
        })
        blocked = subprocess.run(
            ["node", str(guard)],
            input=mirror_payload,
            text=True,
            capture_output=True,
            cwd=skill,
            env=guard_env,
        )
        check(
            blocked.returncode == 2,
            "S-10",
            "PreToolUse guard did not block reading a Chinese mirror.",
            failures,
        )

        canonical_payload = json.dumps({
            "hook_event_name": "PreToolUse",
            "tool_name": "read",
            "tool_input": {"path": str(skill / "SKILL.md")},
        })
        allowed = subprocess.run(
            ["node", str(guard)],
            input=canonical_payload,
            text=True,
            capture_output=True,
            cwd=skill,
            env=guard_env,
        )
        check(
            allowed.returncode == 0,
            "S-11",
            "PreToolUse guard incorrectly blocked reading a canonical file.",
            failures,
        )
    elif guard.exists():
        failures.append({
            "check_id": "S-10",
            "message": "Node.js is unavailable; cannot behavior-test the bundled guard.",
        })

    # Run bundled mirror checker when Node is available.
    mirror_checker = skill / "scripts" / "check-mirrors.js"
    if mirror_checker.exists() and shutil.which("node"):
        proc = subprocess.run(
            ["node", str(mirror_checker)],
            cwd=skill,
            text=True,
            capture_output=True,
        )
        check(proc.returncode == 0, "S-12", (proc.stderr or proc.stdout).strip() or "Mirror checker failed.", failures)
    elif not mirror_checker.exists():
        failures.append({"check_id": "S-12", "message": "Missing scripts/check-mirrors.js"})

    # Codex Structured Outputs preflight. Every object must be closed and
    # every declared property must be listed in `required`.
    result_schema = skill / "eval" / "schemas" / "result.schema.json"
    check(result_schema.exists(), "S-13", f"Missing {result_schema}", failures)
    if result_schema.exists():
        try:
            schema_doc = json.loads(result_schema.read_text(encoding="utf-8"))
            schema_problems = strict_schema_problems(schema_doc)
            check(
                not schema_problems,
                "S-14",
                "Strict output schema is invalid: " + "; ".join(schema_problems),
                failures,
            )
        except json.JSONDecodeError as exc:
            failures.append({
                "check_id": "S-14",
                "message": f"Result schema is not valid JSON: {exc}",
            })

    conversation_schema = skill / "eval" / "schemas" / "conversation.schema.json"
    check(
        conversation_schema.exists(),
        "S-30",
        f"Missing {conversation_schema}",
        failures,
    )
    if conversation_schema.exists():
        try:
            json.loads(conversation_schema.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            failures.append({
                "check_id": "S-30",
                "message": f"Conversation schema is not valid JSON: {exc}",
            })

    runner = skill / "eval" / "scripts" / "run_codex_baseline.py"
    check(runner.exists(), "S-31", f"Missing {runner}", failures)
    if runner.exists():
        runner_text = runner.read_text(encoding="utf-8")
        check(
            "conversation_dir" in runner_text
            and "model_conversations_dir" in runner_text
            and "record_model_conversation" in runner_text,
            "S-32",
            "Codex baseline runner must persist model conversation artifacts.",
            failures,
        )
        try:
            recorder_ok, recorder_message = check_conversation_recorder(runner)
        except Exception as exc:
            recorder_ok, recorder_message = False, str(exc)
        check(
            recorder_ok,
            "S-33",
            f"Conversation recorder behavior check failed: {recorder_message}",
            failures,
        )

    run_record_schema = skill / "eval" / "schemas" / "run-record.schema.json"
    check(run_record_schema.exists(), "S-34", f"Missing {run_record_schema}", failures)
    if run_record_schema.exists():
        try:
            run_schema_doc = json.loads(run_record_schema.read_text(encoding="utf-8"))
            artifact_props = (
                (run_schema_doc.get("properties") or {})
                .get("artifacts", {})
                .get("properties", {})
            )
            check(
                "model_conversations_dir" in artifact_props,
                "S-35",
                "Run record schema must expose artifacts.model_conversations_dir.",
                failures,
            )
        except json.JSONDecodeError as exc:
            failures.append({
                "check_id": "S-34",
                "message": f"Run record schema is not valid JSON: {exc}",
            })

    report = {
        "skill_dir": str(skill),
        "passed": len(failures) == 0,
        "failure_count": len(failures),
        "failures": failures,
    }

    payload = json.dumps(report, indent=2, ensure_ascii=False)
    print(payload)
    if args.json_out:
        Path(args.json_out).write_text(payload + "\n", encoding="utf-8")
    return 0 if not failures else 1


if __name__ == "__main__":
    raise SystemExit(main())
