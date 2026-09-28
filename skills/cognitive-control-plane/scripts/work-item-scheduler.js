#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");

const WORK_ITEM_KINDS = new Set(["issue", "request", "transaction", "ticket"]);
const TERMINAL_STATES = new Set(["resolved", "concluded", "duplicate", "blocked", "escalated", "cancelled"]);
const ACTIONS = new Set(["dispatch", "wait", "checkpoint", "continue", "verify", "close", "wait_for_human"]);
const CONTINUABLE_RUN_STATES = new Set(["checkpointed", "expired"]);
const DEFAULT_BUDGET = Object.freeze({
  checkpoint_at_fraction: 0.40,
  handoff_at_fraction: 0.45,
  hard_stop_at_fraction: 0.50,
});

function usage() {
  return "Usage: work-item-scheduler.js [--dry-run] STATE.json";
}

function parseArgs(argv) {
  const args = { dryRun: false, statePath: "" };
  for (const arg of argv) {
    if (arg === "--dry-run") args.dryRun = true;
    else if (!args.statePath) args.statePath = arg;
    else throw new Error(usage());
  }
  if (!args.statePath) throw new Error(usage());
  return args;
}

function isObject(value) {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function list(value) {
  return Array.isArray(value) ? value : [];
}

function nonEmptyString(value) {
  return typeof value === "string" && value.length > 0;
}

function errorPlan(message) {
  return {
    ok: false,
    action: "wait_for_human",
    reason: message,
    work_item_id: "",
    run_id: "",
    work_item_terminal_state: "none",
    dry_run: true,
  };
}

function validateState(state) {
  if (!isObject(state)) return "state must be an object";
  const item = state.work_item;
  if (!isObject(item)) return "work_item must be an object";
  if (!nonEmptyString(item.id)) return "work_item.id must be a non-empty string";
  if (!WORK_ITEM_KINDS.has(item.kind)) return "work_item.kind must be issue, request, transaction, or ticket";
  if (state.run !== undefined && !isObject(state.run)) return "run must be an object when present";
  return "";
}

function dependencyBlocked(dependencies) {
  return list(dependencies).some((dependency) => {
    if (!isObject(dependency)) return true;
    return dependency.state !== "resolved" && dependency.state !== "concluded" && dependency.state !== "duplicate";
  });
}

function hasEvidence(value) {
  return list(value).some(nonEmptyString);
}

function leaseIsLive(lease, now) {
  if (!isObject(lease) || !nonEmptyString(lease.expires_at)) return false;
  const expiry = Date.parse(lease.expires_at);
  const current = Date.parse(now);
  return Number.isFinite(expiry) && Number.isFinite(current) && expiry > current;
}

function budgetFor(state) {
  const supplied = isObject(state.run) && isObject(state.run.budget) ? state.run.budget : {};
  return { ...DEFAULT_BUDGET, ...supplied };
}

function plan(state, options = {}) {
  const validationError = validateState(state);
  if (validationError) return errorPlan(validationError);

  const item = state.work_item;
  const run = isObject(state.run) ? state.run : {};
  const now = nonEmptyString(state.now) ? state.now : new Date().toISOString();
  const requestedTerminalState = state.requested_terminal_state || item.terminal_state || "none";
  const base = {
    ok: true,
    work_item_id: item.id,
    run_id: typeof run.id === "string" ? run.id : "",
    work_item_kind: item.kind,
    work_item_terminal_state: TERMINAL_STATES.has(requestedTerminalState) ? requestedTerminalState : "none",
    source_run_id: typeof run.id === "string" ? run.id : "",
    expected_run_attempt: Number.isInteger(run.attempt) && run.attempt > 0 ? run.attempt : 1,
    requires_fresh_session: false,
    dry_run: Boolean(options.dryRun),
  };

  if (requestedTerminalState === "blocked" || requestedTerminalState === "escalated") {
    return { ...base, action: "wait_for_human", reason: "terminal_state_stops_automatic_retry" };
  }
  if (requestedTerminalState === "resolved") {
    if (run.validation && run.validation.passed === true && hasEvidence(run.validation.evidence_refs)) {
      return { ...base, action: "close", reason: "resolved_validation_evidence_present" };
    }
    return { ...base, work_item_terminal_state: "none", action: "verify", reason: "resolved_requires_validation_evidence" };
  }
  if (requestedTerminalState === "concluded") {
    if (hasEvidence(run.conclusion_evidence_refs)) {
      return { ...base, action: "close", reason: "concluded_evidence_present" };
    }
    return { ...base, work_item_terminal_state: "none", action: "wait", reason: "concluded_requires_evidence" };
  }
  if (requestedTerminalState === "duplicate" || requestedTerminalState === "cancelled") {
    return { ...base, action: "close", reason: "non_execution_terminal_state" };
  }

  if (dependencyBlocked(item.dependencies)) {
    return { ...base, action: "wait", reason: "dependency_not_terminal" };
  }
  if (leaseIsLive(item.lease, now)) {
    return { ...base, action: "wait", reason: "live_lease_exists" };
  }
  if (item.kind === "transaction" && !nonEmptyString(item.idempotency_key)) {
    return { ...base, action: "wait_for_human", reason: "transaction_requires_idempotency_key" };
  }

  const budget = budgetFor(state);
  const usage = typeof run.usage_fraction === "number" ? run.usage_fraction : 0;
  const checkpointWritten = Boolean(run.checkpoint && run.checkpoint.written);
  if (usage >= budget.hard_stop_at_fraction) {
    if (!checkpointWritten) return { ...base, action: "checkpoint", reason: "hard_stop_requires_checkpoint" };
    if (!CONTINUABLE_RUN_STATES.has(run.state)) {
      return { ...base, action: "wait", reason: "previous_run_must_end_before_continuation" };
    }
    return {
      ...base,
      action: "continue",
      reason: "hard_stop_continue_same_work_item",
      expected_run_attempt: base.expected_run_attempt + 1,
      requires_fresh_session: true,
    };
  }
  if (usage >= budget.handoff_at_fraction) {
    if (!checkpointWritten) return { ...base, action: "checkpoint", reason: "handoff_requires_checkpoint" };
    if (!CONTINUABLE_RUN_STATES.has(run.state)) {
      return { ...base, action: "wait", reason: "previous_run_must_end_before_continuation" };
    }
    return {
      ...base,
      action: "continue",
      reason: "handoff_continue_same_work_item",
      expected_run_attempt: base.expected_run_attempt + 1,
      requires_fresh_session: true,
    };
  }
  if (usage >= budget.checkpoint_at_fraction && !checkpointWritten) {
    return { ...base, action: "checkpoint", reason: "checkpoint_threshold_reached" };
  }
  return {
    ...base,
    action: "dispatch",
    reason: item.lease ? "expired_lease_recovered" : "ready_for_dispatch",
    requires_fresh_session: true,
  };
}

function main() {
  try {
    const args = parseArgs(process.argv.slice(2));
    const state = JSON.parse(fs.readFileSync(path.resolve(args.statePath), "utf8"));
    const output = plan(state, { dryRun: args.dryRun });
    if (!ACTIONS.has(output.action)) throw new Error("scheduler returned an invalid action");
    process.stdout.write(`${JSON.stringify(output)}\n`);
    process.exit(output.ok ? 0 : 2);
  } catch (error) {
    process.stdout.write(`${JSON.stringify(errorPlan(error.message))}\n`);
    process.exit(2);
  }
}

if (require.main === module) main();

module.exports = { DEFAULT_BUDGET, plan };
