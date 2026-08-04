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


def check_adapter_review_contract(adapter_cli: Path) -> tuple[bool, str]:
    requirement = {"name": "reviewing-code", "source": "available_skill", "required": True, "reason": "review"}
    task = {
        "task_id": "review-1",
        "actor_id": "reviewer-b",
        "role": "code_reviewer",
        "phase": "review",
        "objective": "Review the pinned artifact.",
        "review_of_task_id": "implementation-1",
        "review_of_actor_id": "implementer-a",
        "review_iteration": 1,
        "supersedes_review_task_id": "",
        "review_fallback": "none",
        "review_target": {
            "kind": "git_range",
            "base_sha": "111",
            "head_sha": "222",
            "diff_hash": "sha256:abc",
        },
        "constraints": [],
        "required_skills": [requirement],
        "required_references": [],
        "required_mcp": [],
        "required_tools": [],
        "ownership": {"writable_paths": [], "read_only_paths": ["."], "forbidden_paths": []},
        "edits_allowed": False,
        "expected_output": {"format": "code_review_report", "required_fields": [], "must_report": []},
        "validation": [],
        "stop_if": [],
    }
    valid_contract = {"ccp_version": 4, "next_action": "delegate_read_only", "task": task}

    with tempfile.TemporaryDirectory() as tmp:
        path = Path(tmp) / "contract.json"
        path.write_text(json.dumps(valid_contract), encoding="utf-8")
        valid = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if valid.returncode != 0:
            return False, f"valid review contract was rejected: {valid.stdout or valid.stderr}"

        self_review = json.loads(json.dumps(valid_contract))
        self_review["task"]["actor_id"] = "implementer-a"
        path.write_text(json.dumps(self_review), encoding="utf-8")
        rejected = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if rejected.returncode == 0:
            return False, "self-review contract was accepted"

        stale_target = json.loads(json.dumps(valid_contract))
        del stale_target["task"]["review_target"]["diff_hash"]
        path.write_text(json.dumps(stale_target), encoding="utf-8")
        rejected = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if rejected.returncode == 0:
            return False, "unversioned git review target was accepted"

        fallback = json.loads(json.dumps(valid_contract))
        fallback["task"]["review_fallback"] = "independent_read_only_reviewer"
        fallback["task"]["required_skills"] = []
        fallback["task"]["required_references"] = [{
            "name": "reviewer-enforcement",
            "source": "file_reference",
            "required": True,
            "reason": "fallback review policy",
        }]
        path.write_text(json.dumps(fallback), encoding="utf-8")
        accepted = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if accepted.returncode != 0:
            return False, f"valid independent reviewer fallback was rejected: {accepted.stdout or accepted.stderr}"

        fallback["task"]["required_references"] = []
        path.write_text(json.dumps(fallback), encoding="utf-8")
        rejected = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if rejected.returncode == 0:
            return False, "fallback without reviewer-enforcement reference was accepted"

    return True, ""


def check_work_item_trace_contract(result_schema_path: Path, runner_path: Path) -> tuple[bool, str]:
    schema = json.loads(result_schema_path.read_text(encoding="utf-8"))
    trace = (schema.get("properties") or {}).get("trace") or {}
    properties = trace.get("properties") or {}
    required = set(trace.get("required") or [])
    run_states = set((properties.get("run_state") or {}).get("enum") or [])
    expected_run_states = {"none", "leased", "running", "checkpointed", "continued", "completed"}
    expected_terminal_states = {
        "none",
        "resolved",
        "concluded",
        "duplicate",
        "blocked",
        "escalated",
        "cancelled",
    }

    if run_states != expected_run_states:
        return False, f"run_state must be exactly {sorted(expected_run_states)}, got {sorted(run_states)}"
    if {"blocked", "escalated"} & run_states:
        return False, "run_state must not include work-item terminal states"
    terminal_states = set((properties.get("work_item_terminal_state") or {}).get("enum") or [])
    if terminal_states != expected_terminal_states:
        return False, f"work_item_terminal_state must be exactly {sorted(expected_terminal_states)}"
    if "transaction_idempotency_key_present" not in required:
        return False, "transaction_idempotency_key_present must be required in the trace"
    if (properties.get("transaction_idempotency_key_present") or {}).get("type") != "boolean":
        return False, "transaction_idempotency_key_present must be boolean"

    module = import_python_module(runner_path, "ccp_eval_runner_trace_check")
    fallback = module.fallback_result("ACP-TRACE", "synthetic error")
    default_trace = fallback.get("trace") or {}
    expected_defaults = {
        "work_item_kind": "none",
        "run_state": "none",
        "lease_acquired": False,
        "checkpoint_written": False,
        "transaction_idempotency_key_present": False,
        "scheduler_action": "none",
        "work_item_terminal_state": "none",
    }
    if any(default_trace.get(name) != value for name, value in expected_defaults.items()):
        return False, "fallback trace must use the documented work-item defaults"
    return True, ""


def check_adapter_work_item_schema(schema_path: Path) -> tuple[bool, str]:
    schema = json.loads(schema_path.read_text(encoding="utf-8"))
    work_item = ((schema.get("$defs") or {}).get("workItemContext") or {})
    key_schema = (work_item.get("properties") or {}).get("idempotency_key") or {}
    if key_schema.get("type") != "string" or key_schema.get("minLength") != 1:
        return False, "workItemContext.idempotency_key must be a non-empty string"

    for rule in work_item.get("allOf") or []:
        condition = (((rule.get("if") or {}).get("properties") or {}).get("kind") or {})
        transaction_rule = condition.get("const") == "transaction"
        requires_kind = "kind" in ((rule.get("if") or {}).get("required") or [])
        requires_key = "idempotency_key" in ((rule.get("then") or {}).get("required") or [])
        forbids_key = "idempotency_key" in ((((rule.get("else") or {}).get("not") or {}).get("required") or []))
        if transaction_rule and requires_kind and requires_key and forbids_key:
            return True, ""
    return False, "workItemContext must require idempotency_key iff kind is transaction"


def check_adapter_work_item_contract(adapter_cli: Path) -> tuple[bool, str]:
    task = {
        "task_id": "work-item-run-1",
        "actor_id": "runner-a",
        "role": "work_item_runner",
        "phase": "work_item",
        "objective": "Resolve the accepted work item.",
        "work_item": {
            "id": "INC-1",
            "kind": "issue",
            "objective": "Resolve the incident.",
            "acceptance_criteria": ["A verified terminal outcome is recorded."],
            "dependencies": [],
            "authorization": ["read_repository"],
        },
        "run": {
            "id": "INC-1-R01",
            "attempt": 1,
            "lease_id": "lease-1",
            "lease_expires_at": "2026-08-01T00:30:00Z",
            "resume_checkpoint_ref": "",
            "budget": {
                "checkpoint_at_fraction": 0.40,
                "handoff_at_fraction": 0.45,
                "hard_stop_at_fraction": 0.50,
            },
        },
        "constraints": [],
        "required_skills": [],
        "required_references": [],
        "required_mcp": [],
        "required_tools": [],
        "ownership": {"writable_paths": [], "read_only_paths": ["."], "forbidden_paths": []},
        "edits_allowed": False,
        "expected_output": {"format": "work_item_run_report", "required_fields": [], "must_report": []},
        "validation": [],
        "stop_if": [],
    }
    valid_contract = {"ccp_version": 4, "next_action": "delegate_read_only", "task": task}

    with tempfile.TemporaryDirectory() as tmp:
        path = Path(tmp) / "contract.json"
        path.write_text(json.dumps(valid_contract), encoding="utf-8")
        valid = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if valid.returncode != 0:
            return False, f"valid work-item contract was rejected: {valid.stdout or valid.stderr}"

        invalid_budget = json.loads(json.dumps(valid_contract))
        invalid_budget["task"]["run"]["budget"]["hard_stop_at_fraction"] = 0.51
        path.write_text(json.dumps(invalid_budget), encoding="utf-8")
        rejected = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if rejected.returncode == 0:
            return False, "work-item contract above the 50% hard stop was accepted"

        invalid_order = json.loads(json.dumps(valid_contract))
        invalid_order["task"]["run"]["budget"]["handoff_at_fraction"] = 0.40
        path.write_text(json.dumps(invalid_order), encoding="utf-8")
        rejected = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if rejected.returncode == 0:
            return False, "work-item contract with unordered budget thresholds was accepted"

        invalid_kind = json.loads(json.dumps(valid_contract))
        invalid_kind["task"]["work_item"]["kind"] = "alert"
        path.write_text(json.dumps(invalid_kind), encoding="utf-8")
        rejected = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if rejected.returncode == 0:
            return False, "work-item contract with an unsupported kind was accepted"

        transaction_without_key = json.loads(json.dumps(valid_contract))
        transaction_without_key["task"]["work_item"]["kind"] = "transaction"
        path.write_text(json.dumps(transaction_without_key), encoding="utf-8")
        rejected = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if rejected.returncode == 0:
            return False, "transaction work-item contract without idempotency_key was accepted"

        transaction_with_key = json.loads(json.dumps(transaction_without_key))
        transaction_with_key["task"]["work_item"]["idempotency_key"] = "payment-INC-1-v1"
        path.write_text(json.dumps(transaction_with_key), encoding="utf-8")
        accepted = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if accepted.returncode != 0:
            return False, f"transaction work-item contract with idempotency_key was rejected: {accepted.stdout or accepted.stderr}"

        non_transaction_with_key = json.loads(json.dumps(valid_contract))
        non_transaction_with_key["task"]["work_item"]["idempotency_key"] = "must-not-exist"
        path.write_text(json.dumps(non_transaction_with_key), encoding="utf-8")
        rejected = subprocess.run(["node", str(adapter_cli), "validate", str(path)], text=True, capture_output=True)
        if rejected.returncode == 0:
            return False, "non-transaction work-item contract with idempotency_key was accepted"

    return True, ""


def check_work_item_loop(loop_test: Path) -> tuple[bool, str]:
    if not loop_test.exists():
        return False, "work-item loop regression test is missing"
    executed = subprocess.run(["node", str(loop_test)], text=True, capture_output=True)
    if executed.returncode != 0:
        return False, executed.stdout or executed.stderr or "work-item loop regression test failed"
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
        "references/reviewer-enforcement.md",
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

        scheduler = data.get("work_item_scheduler_policy") or {}
        accepted_kinds_ok, accepted_kinds_missing = list_contains_all(
            scheduler.get("accepted_kinds"),
            {"issue", "request", "transaction", "ticket"},
        )
        terminal_states_ok, terminal_states_missing = list_contains_all(
            scheduler.get("terminal_states"),
            {"resolved", "concluded", "duplicate", "blocked", "escalated", "cancelled"},
        )
        budget = scheduler.get("budget") or {}
        continuation = scheduler.get("continuation") or {}
        programmatic_tool_calling = scheduler.get("programmatic_tool_calling") or {}
        terminal_gates = scheduler.get("terminal_gates") or {}
        stop_retry_ok, stop_retry_missing = list_contains_all(
            terminal_gates.get("stop_auto_retry_for"),
            {"blocked", "escalated"},
        )
        scheduler_owns_ok, scheduler_owns_missing = list_contains_all(
            scheduler.get("scheduler_owns"),
            {
                "intake_normalization",
                "dependency_gating",
                "lease_acquisition_and_expiry_recovery",
                "budget_enforcement",
                "session_continuation",
                "terminal_state_transition",
            },
        )
        check(
            accepted_kinds_ok
            and terminal_states_ok
            and scheduler.get("work_item") == "persistent_business_unit"
            and scheduler.get("run") == "one_session_processing_attempt"
            and scheduler_owns_ok
            and (scheduler.get("lease") or {}).get("required") is True
            and (scheduler.get("dependencies") or {}).get("required_before_lease") is True
            and budget.get("checkpoint_at_fraction") == 0.40
            and budget.get("handoff_at_fraction") == 0.45
            and budget.get("hard_stop_at_fraction") == 0.50
            and continuation.get("checkpoint_required_before_hard_stop") is True
            and continuation.get("resume_same_work_item") is True
            and set(continuation.get("predecessor_must_be") or []) == {"checkpointed", "expired"}
            and continuation.get("requires_fresh_session") is True
            and continuation.get("native_resume_allowed") is False
            and continuation.get("requires_new_run_id") is True
            and continuation.get("requires_higher_attempt") is True
            and continuation.get("requires_checkpoint_ref") is True
            and set(programmatic_tool_calling.get("allowed_for") or []) == {
                "bounded_read_tool_batches", "filtering", "joining", "aggregation", "mechanical_validation"
            }
            and set(programmatic_tool_calling.get("host_only") or []) == {
                "durable_state", "lease_transition", "session_launch", "approval", "external_write", "semantic_judgment"
            }
            and terminal_gates.get("resolved_requires_validation") is True
            and terminal_gates.get("concluded_requires_evidence") is True
            and stop_retry_ok
            and (scheduler.get("transactions") or {}).get("idempotency_required") is True,
            "S-46",
            "work_item_scheduler_policy is incomplete: "
            f"kinds={accepted_kinds_missing}, terminal_states={terminal_states_missing}, "
            f"scheduler_owns={scheduler_owns_missing}, stop_retry={stop_retry_missing}, "
            "fresh-session and programmatic-tool-calling policy must be complete",
            failures,
        )

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

        review_policy = data.get("reviewer_enforcement_policy") or {}
        risk_triggers_ok, risk_triggers_missing = list_contains_all(
            review_policy.get("risk_triggers"),
            {
                "security_sensitive",
                "cross_module_change",
                "public_api_change",
                "schema_change",
                "migration",
                "auth_or_permission_change",
                "deployment_or_rollback_critical",
            },
        )
        check(
            bool(review_policy.get("mandatory_after_terminal_implementation")) and risk_triggers_ok,
            "S-36",
            f"reviewer_enforcement_policy is missing mandatory risk triggers: {risk_triggers_missing}",
            failures,
        )
        independence = review_policy.get("reviewer_independence") or {}
        check(
            independence.get("compare_field") == "actor_id"
            and bool(independence.get("role_switch_is_not_independence"))
            and independence.get("reviewer_may_fix_own_findings") is False,
            "S-37",
            "Reviewer independence must compare actor_id, reject role-switch self-review, and keep reviewers out of fixes.",
            failures,
        )
        artifact_version = review_policy.get("artifact_version") or {}
        accepted_kinds = artifact_version.get("accepted_kinds") or {}
        git_fields_ok, git_fields_missing = list_contains_all(
            (accepted_kinds.get("git_range") or {}).get("required_fields"),
            {"base_sha", "head_sha", "diff_hash"},
        )
        stable_fields_ok, stable_fields_missing = list_contains_all(
            (accepted_kinds.get("stable_artifact") or {}).get("required_fields"),
            {"stable_id"},
        )
        check(
            git_fields_ok and stable_fields_ok and bool(artifact_version.get("invalidate_review_when_target_changes")),
            "S-38",
            "Artifact pinning must require base/head/diff hash or stable_id and invalidate stale review; "
            f"missing git={git_fields_missing}, stable={stable_fields_missing}",
            failures,
        )
        blocking = review_policy.get("blocking_findings") or {}
        check(
            bool(blocking.get("prevent_final_acceptance"))
            and bool(blocking.get("dispatch_fix_task"))
            and blocking.get("next_action") == "delegate_write",
            "S-39",
            "Blocking findings must prevent final acceptance and dispatch a write-capable fix task.",
            failures,
        )
        rereview = review_policy.get("rereview") or {}
        continue_ok, continue_missing = list_contains_all(
            rereview.get("continue_until"),
            {"latest_version_has_no_blocking_findings", "explicitly_terminated_unaccepted"},
        )
        check(
            bool(rereview.get("required_after_fix"))
            and bool(rereview.get("invalidate_prior_review"))
            and continue_ok,
            "S-40",
            f"Re-review loop is incomplete; missing terminal states: {continue_missing}",
            failures,
        )
        final_requirements_ok, final_requirements_missing = list_contains_all(
            review_policy.get("final_acceptance_requires"),
            {
                "implementation_risk_assessed",
                "mandatory_review_task_terminal",
                "reviewer_actor_independent",
                "reviewed_target_matches_current_artifact",
                "no_unresolved_blocking_findings",
                "every_fix_followed_by_rereview",
                "review_gate_cleared",
            },
        )
        check(
            final_requirements_ok,
            "S-41",
            f"Final acceptance reviewer gate is missing requirements: {final_requirements_missing}",
            failures,
        )
        fallback = review_policy.get("reviewing_code_unavailable") or {}
        check(
            bool(review_policy.get("reviewer_capability_preflight"))
            and fallback.get("allowed_fallback") == "independent_read_only_reviewer"
            and bool(fallback.get("requires_distinct_host_started_actor"))
            and bool(fallback.get("requires_reviewer_enforcement_reference"))
            and bool(fallback.get("must_report_unavailable_skill_as_deviation"))
            and fallback.get("otherwise") == "block_gate_and_emit_handoff",
            "S-44",
            "Reviewer fallback must preflight capability, require an independent host actor and reviewer-enforcement reference, then block when unavailable.",
            failures,
        )
        report_format = review_policy.get("review_report_format") or {}
        verification_ok, verification_missing = list_contains_all(
            report_format.get("verification_matrix_columns"),
            {"status", "review_lane_or_check", "evidence", "result_or_limitation"},
        )
        findings_ok, findings_missing = list_contains_all(
            report_format.get("findings_table_columns"),
            {"id", "severity", "location", "evidence", "recommendation", "disposition"},
        )
        check(
            verification_ok and findings_ok,
            "S-45",
            "Review report format is missing verification or findings table columns: "
            f"verification={verification_missing}, findings={findings_missing}",
            failures,
        )
        contract_patterns = data.get("contract_patterns") or {}
        missing_pattern_identity = sorted(
            name for name, pattern in contract_patterns.items()
            if not isinstance(pattern, dict) or "task_id" not in pattern or "actor_id" not in pattern
        )
        check(
            not missing_pattern_identity,
            "S-43",
            f"Contract patterns missing task_id/actor_id identity fields: {missing_pattern_identity}",
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
        try:
            work_item_schema_ok, work_item_schema_message = check_adapter_work_item_schema(adapter_schema)
        except Exception as exc:
            work_item_schema_ok, work_item_schema_message = False, str(exc)
        check(
            work_item_schema_ok,
            "S-49",
            f"Adapter work-item schema check failed: {work_item_schema_message}",
            failures,
        )
    if adapter_cli.exists() and shutil.which("node"):
        try:
            adapter_review_ok, adapter_review_message = check_adapter_review_contract(adapter_cli)
        except Exception as exc:
            adapter_review_ok, adapter_review_message = False, str(exc)
        check(
            adapter_review_ok,
            "S-42",
            f"Adapter reviewer-enforcement behavior check failed: {adapter_review_message}",
            failures,
        )
        try:
            adapter_work_item_ok, adapter_work_item_message = check_adapter_work_item_contract(adapter_cli)
        except Exception as exc:
            adapter_work_item_ok, adapter_work_item_message = False, str(exc)
        check(
            adapter_work_item_ok,
            "S-47",
            f"Adapter work-item scheduling behavior check failed: {adapter_work_item_message}",
            failures,
        )
        try:
            loop_ok, loop_message = check_work_item_loop(skill / "eval" / "tests" / "work-item-loop.test.js")
        except Exception as exc:
            loop_ok, loop_message = False, str(exc)
        check(
            loop_ok,
            "S-50",
            f"Fresh-session loop behavior check failed: {loop_message}",
            failures,
        )

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

    if result_schema.exists() and runner.exists():
        try:
            trace_contract_ok, trace_contract_message = check_work_item_trace_contract(result_schema, runner)
        except Exception as exc:
            trace_contract_ok, trace_contract_message = False, str(exc)
        check(
            trace_contract_ok,
            "S-48",
            f"Work-item trace contract check failed: {trace_contract_message}",
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
