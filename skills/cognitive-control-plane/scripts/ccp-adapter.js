#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const { spawn } = require("child_process");

const PLATFORMS = new Set(["opencode", "codex", "claude-code", "unknown"]);
const NEXT_ACTIONS = new Set(["delegate_read_only", "delegate_write", "route_skill", "verify", "deliver"]);
const PHASES = new Set(["context", "design", "implementation", "review", "verification", "work_item"]);
const WORK_ITEM_KINDS = new Set(["issue", "request", "transaction", "ticket"]);

function usage() {
  return [
    "Usage:",
    "  scripts/ccp-adapter.js detect [--platform NAME]",
    "  scripts/ccp-adapter.js validate CONTRACT.json",
    "  scripts/ccp-adapter.js render [--platform NAME] [--task-id ID] CONTRACT.json",
    "  scripts/ccp-adapter.js launch --platform codex|opencode --workspace WORKSPACE [--sandbox MODE] [--approval POLICY] [--execute] [--executable PATH] CONTRACT.json",
  ].join("\n");
}

function parseArgs(argv) {
  const args = { _: [], execute: false, workspace: "", sandbox: "workspace-write", approval: "on-request", executable: "", contractPath: "" };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (["--platform", "--task-id", "--workspace", "--sandbox", "--approval", "--executable", "--contract"].includes(arg)) {
      const value = argv[++i];
      if (!value || value.startsWith("--")) throw new Error(`${arg} requires a value`);
      if (arg === "--platform") args.platform = value;
      if (arg === "--task-id") args.taskId = value;
      if (arg === "--workspace") args.workspace = value;
      if (arg === "--sandbox") args.sandbox = value;
      if (arg === "--approval") args.approval = value;
      if (arg === "--executable") args.executable = value;
      if (arg === "--contract") args.contractPath = value;
    } else if (arg === "--execute") {
      args.execute = true;
    } else {
      args._.push(arg);
    }
  }
  const positionalContracts = args._.slice(1);
  if (args.contractPath && positionalContracts.length) throw new Error("contract must be supplied once");
  if (!args.contractPath && positionalContracts.length > 1) throw new Error("contract must be supplied once");
  if (!args.contractPath && positionalContracts.length === 1) args.contractPath = positionalContracts[0];
  return args;
}

function readJson(file) {
  return JSON.parse(fs.readFileSync(file, "utf8"));
}

function fail(message, code = 2) {
  process.stderr.write(`${message}\n`);
  process.exit(code);
}

function detectPlatform(explicit) {
  if (explicit) return PLATFORMS.has(explicit) ? explicit : "unknown";
  if (process.env.CCP_PLATFORM && PLATFORMS.has(process.env.CCP_PLATFORM)) return process.env.CCP_PLATFORM;
  if (process.env.OPENCODE_SESSION_ID || process.env.OPENCODE_CONFIG) return "opencode";
  if (process.env.CLAUDECODE || process.env.CLAUDE_CODE) return "claude-code";
  if (process.env.CODEX_HOME || process.env.CODEX_SANDBOX) return "codex";
  return "unknown";
}

function capabilities(platform) {
  return {
    platform,
    subagents: false,
    background_tasks: false,
    parallel_write_workers: false,
    skill_loading: platform !== "unknown",
    event_log: true,
    reason: "No host task API is exposed to this repository-level adapter. A host integration may override this result after it starts a native worker.",
  };
}

function requireObject(value, label, errors) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    errors.push(`${label} must be an object`);
    return false;
  }
  return true;
}

function requireArray(value, label, errors) {
  if (!Array.isArray(value)) {
    errors.push(`${label} must be an array`);
    return false;
  }
  return true;
}

function validateRequirement(item, label, errors) {
  if (!requireObject(item, label, errors)) return;
  for (const key of ["name", "source", "required", "reason"]) {
    if (!(key in item)) errors.push(`${label}.${key} is required`);
  }
  if (typeof item.required !== "boolean") errors.push(`${label}.required must be boolean`);
}

function validateReviewTarget(target, label, errors) {
  if (!requireObject(target, label, errors)) return;
  if (target.kind === "git_range") {
    for (const key of ["base_sha", "head_sha", "diff_hash"]) {
      if (typeof target[key] !== "string" || !target[key]) errors.push(`${label}.${key} must be a non-empty string for git_range`);
    }
    return;
  }
  if (target.kind === "stable_artifact") {
    if (typeof target.stable_id !== "string" || !target.stable_id) errors.push(`${label}.stable_id must be a non-empty string for stable_artifact`);
    return;
  }
  errors.push(`${label}.kind must be git_range or stable_artifact`);
}

function validateWorkItemContext(workItem, label, errors) {
  if (!requireObject(workItem, label, errors)) return;
  for (const key of ["id", "kind", "objective", "acceptance_criteria", "dependencies", "authorization"]) {
    if (!(key in workItem)) errors.push(`${label}.${key} is required`);
  }
  if (typeof workItem.id !== "string" || !workItem.id) errors.push(`${label}.id must be a non-empty string`);
  if (!WORK_ITEM_KINDS.has(workItem.kind)) errors.push(`${label}.kind must be issue, request, transaction, or ticket`);
  if (typeof workItem.objective !== "string" || !workItem.objective) errors.push(`${label}.objective must be a non-empty string`);
  for (const key of ["acceptance_criteria", "dependencies", "authorization"]) requireArray(workItem[key], `${label}.${key}`, errors);
  if (workItem.kind === "transaction" && !nonEmptyString(workItem.idempotency_key)) {
    errors.push(`${label}.idempotency_key must be a non-empty string for transaction work items`);
  }
  if (workItem.kind !== "transaction" && "idempotency_key" in workItem) {
    errors.push(`${label}.idempotency_key is allowed only for transaction work items`);
  }
}

function validateRunContext(run, label, errors) {
  if (!requireObject(run, label, errors)) return;
  for (const key of ["id", "attempt", "lease_id", "lease_expires_at", "resume_checkpoint_ref", "budget"]) {
    if (!(key in run)) errors.push(`${label}.${key} is required`);
  }
  if (typeof run.id !== "string" || !run.id) errors.push(`${label}.id must be a non-empty string`);
  if (!Number.isInteger(run.attempt) || run.attempt < 1) errors.push(`${label}.attempt must be a positive integer`);
  if (typeof run.lease_id !== "string" || !run.lease_id) errors.push(`${label}.lease_id must be a non-empty string`);
  if (typeof run.lease_expires_at !== "string" || !run.lease_expires_at) errors.push(`${label}.lease_expires_at must be a non-empty string`);
  if (typeof run.resume_checkpoint_ref !== "string") errors.push(`${label}.resume_checkpoint_ref must be a string`);
  if (!requireObject(run.budget, `${label}.budget`, errors)) return;
  const budget = run.budget;
  for (const key of ["checkpoint_at_fraction", "handoff_at_fraction", "hard_stop_at_fraction"]) {
    if (typeof budget[key] !== "number" || !Number.isFinite(budget[key]) || budget[key] <= 0) {
      errors.push(`${label}.budget.${key} must be a positive number`);
    }
  }
  if (typeof budget.hard_stop_at_fraction === "number" && budget.hard_stop_at_fraction > 0.5) {
    errors.push(`${label}.budget.hard_stop_at_fraction must not exceed 0.5`);
  }
  if (
    typeof budget.checkpoint_at_fraction === "number"
    && typeof budget.handoff_at_fraction === "number"
    && typeof budget.hard_stop_at_fraction === "number"
    && !(budget.checkpoint_at_fraction < budget.handoff_at_fraction && budget.handoff_at_fraction < budget.hard_stop_at_fraction)
  ) {
    errors.push(`${label}.budget must satisfy checkpoint_at_fraction < handoff_at_fraction < hard_stop_at_fraction`);
  }
}

function hasRequirement(items, name) {
  return Array.isArray(items) && items.some((item) => item && item.name === name);
}

function validateContract(contract) {
  const errors = [];
  if (!requireObject(contract, "$", errors)) return errors;
  if (!Number.isInteger(contract.ccp_version) || contract.ccp_version < 1) errors.push("$.ccp_version must be a positive integer");
  if (!NEXT_ACTIONS.has(contract.next_action)) errors.push("$.next_action is invalid");
  if (!requireObject(contract.task, "$.task", errors)) return errors;

  const task = contract.task;
  for (const key of [
    "task_id",
    "actor_id",
    "role",
    "phase",
    "objective",
    "constraints",
    "required_skills",
    "required_references",
    "required_mcp",
    "required_tools",
    "ownership",
    "edits_allowed",
    "expected_output",
    "validation",
    "stop_if",
  ]) {
    if (!(key in task)) errors.push(`$.task.${key} is required`);
  }
  if (typeof task.task_id !== "string" || !task.task_id) errors.push("$.task.task_id must be a non-empty string");
  if (typeof task.actor_id !== "string" || !task.actor_id) errors.push("$.task.actor_id must be a non-empty string");
  if (typeof task.role !== "string" || !task.role) errors.push("$.task.role must be a non-empty string");
  if (!PHASES.has(task.phase)) errors.push("$.task.phase is invalid");
  if (typeof task.objective !== "string" || !task.objective) errors.push("$.task.objective must be a non-empty string");
  if (typeof task.edits_allowed !== "boolean") errors.push("$.task.edits_allowed must be boolean");

  for (const key of ["constraints", "validation", "stop_if"]) requireArray(task[key], `$.task.${key}`, errors);
  for (const key of ["required_skills", "required_references", "required_mcp", "required_tools"]) {
    if (requireArray(task[key], `$.task.${key}`, errors)) {
      task[key].forEach((item, index) => validateRequirement(item, `$.task.${key}[${index}]`, errors));
    }
  }

  if (requireObject(task.ownership, "$.task.ownership", errors)) {
    for (const key of ["writable_paths", "read_only_paths", "forbidden_paths"]) requireArray(task.ownership[key], `$.task.ownership.${key}`, errors);
  }

  if (requireObject(task.expected_output, "$.task.expected_output", errors)) {
    if (typeof task.expected_output.format !== "string" || !task.expected_output.format) errors.push("$.task.expected_output.format must be a non-empty string");
    requireArray(task.expected_output.required_fields, "$.task.expected_output.required_fields", errors);
    requireArray(task.expected_output.must_report, "$.task.expected_output.must_report", errors);
  }

  if (task.phase === "review") {
    for (const key of ["review_of_task_id", "review_of_actor_id", "review_iteration", "supersedes_review_task_id", "review_fallback", "review_target"]) {
      if (!(key in task)) errors.push(`$.task.${key} is required for review phase`);
    }
    if (typeof task.review_of_task_id !== "string" || !task.review_of_task_id) errors.push("$.task.review_of_task_id must be a non-empty string for review phase");
    if (typeof task.review_of_actor_id !== "string" || !task.review_of_actor_id) errors.push("$.task.review_of_actor_id must be a non-empty string for review phase");
    if (!Number.isInteger(task.review_iteration) || task.review_iteration < 1) errors.push("$.task.review_iteration must be a positive integer for review phase");
    if (typeof task.supersedes_review_task_id !== "string") errors.push("$.task.supersedes_review_task_id must be a string for review phase");
    if (!["none", "independent_read_only_reviewer"].includes(task.review_fallback)) errors.push("$.task.review_fallback must be none or independent_read_only_reviewer for review phase");
    validateReviewTarget(task.review_target, "$.task.review_target", errors);
    if (task.actor_id && task.actor_id === task.review_of_actor_id) errors.push("review actor_id must differ from review_of_actor_id");
    if (task.edits_allowed !== false) errors.push("review phase requires $.task.edits_allowed=false");
    if (task.review_fallback === "none" && !hasRequirement(task.required_skills, "reviewing-code")) errors.push("standard review requires reviewing-code");
    if (task.review_fallback === "independent_read_only_reviewer") {
      if (hasRequirement(task.required_skills, "reviewing-code")) errors.push("fallback review must not claim unavailable reviewing-code");
      if (!hasRequirement(task.required_references, "reviewer-enforcement")) errors.push("fallback review requires reviewer-enforcement reference");
    }
  }

  if (task.phase === "work_item") {
    validateWorkItemContext(task.work_item, "$.task.work_item", errors);
    validateRunContext(task.run, "$.task.run", errors);
  }

  if (contract.next_action === "delegate_write" && task.edits_allowed !== true) {
    errors.push("delegate_write requires $.task.edits_allowed=true");
  }
  if (contract.next_action === "delegate_read_only" && task.edits_allowed !== false) {
    errors.push("delegate_read_only requires $.task.edits_allowed=false");
  }

  return errors;
}

function render(contract, platform) {
  return {
    status: "handoff",
    platform,
    subagent_started: false,
    reason: "Render only preserves a handoff contract. It never starts a native session; use launch --execute for an explicit launch.",
    contract,
  };
}

function nonEmptyString(value) {
  return typeof value === "string" && value.length > 0;
}

function launchIds(contract) {
  const task = contract.task;
  const workId = task.phase === "work_item" && task.work_item ? task.work_item.id : task.task_id;
  const runId = task.phase === "work_item" && task.run ? task.run.id : `${task.task_id}:run`;
  return { work_id: workId, run_id: runId };
}

function transactionIdempotencyError(contract) {
  const task = contract.task;
  if (task.phase !== "work_item" || !task.work_item || task.work_item.kind !== "transaction") return "";
  if (!nonEmptyString(task.work_item.idempotency_key)) {
    return "transaction work items require task.work_item.idempotency_key before launch";
  }
  return "";
}

function resolveExecutable(executable, platform) {
  const candidate = executable || platform;
  if (candidate.includes(path.sep)) {
    const absolute = path.resolve(candidate);
    try {
      fs.accessSync(absolute, fs.constants.X_OK);
      return absolute;
    } catch (_) {
      return "";
    }
  }
  for (const directory of (process.env.PATH || "").split(path.delimiter)) {
    if (!directory) continue;
    const absolute = path.join(directory, candidate);
    try {
      fs.accessSync(absolute, fs.constants.X_OK);
      return absolute;
    } catch (_) {
      // Continue searching PATH. A missing host client is an unavailable launch, not a false start.
    }
  }
  return "";
}

function launchPrompt(contract, ids) {
  return [
    "You are a fresh Cognitive Control Plane worker session.",
    `Work id: ${ids.work_id}`,
    `Run id: ${ids.run_id}`,
    "Follow this portable contract exactly. Do not claim native-session metadata that the adapter did not supply.",
    JSON.stringify(contract),
  ].join("\n\n");
}

function nativeCommand(platform, executable, workspace, sandbox, approval, prompt) {
  if (platform === "codex") {
    return { command: executable, args: ["exec", "--json", "-C", workspace, "-s", sandbox, "-a", approval, prompt] };
  }
  if (platform === "opencode") {
    return { command: executable, args: ["run", "--format", "json", "--dir", workspace, prompt] };
  }
  return null;
}

function launch(contract, args) {
  const platform = detectPlatform(args.platform);
  const ids = launchIds(contract);
  const base = { platform, ...ids, subagent_started: false };
  if (!new Set(["codex", "opencode"]).has(platform)) {
    return { status: "unavailable", ...base, reason: "launch supports only codex or opencode" };
  }
  if (!nonEmptyString(args.workspace)) {
    return { status: "invalid", ...base, reason: "launch requires --workspace" };
  }
  const workspace = path.resolve(args.workspace);
  try {
    if (!fs.statSync(workspace).isDirectory()) throw new Error("not a directory");
  } catch (_) {
    return { status: "invalid", ...base, reason: "workspace must be an existing directory" };
  }
  const idempotencyError = transactionIdempotencyError(contract);
  if (idempotencyError) return { status: "invalid", ...base, reason: idempotencyError };

  const executable = resolveExecutable(args.executable, platform);
  if (!executable) return { status: "unavailable", ...base, reason: `native ${platform} executable is not available` };
  const command = nativeCommand(platform, executable, workspace, args.sandbox, args.approval, launchPrompt(contract, ids));
  if (!args.execute) {
    return { status: "dry_run", ...base, dry_run: true, command: command.command, args: command.args.slice(0, -1) };
  }
  try {
    const child = spawn(command.command, command.args, { cwd: workspace, detached: true, stdio: "ignore" });
    if (!Number.isInteger(child.pid) || child.pid <= 0) throw new Error("native process did not provide a pid");
    child.unref();
    return { status: "started", ...base, subagent_started: true, pid: child.pid };
  } catch (error) {
    return { status: "unavailable", ...base, reason: `native ${platform} launch failed: ${error.message}` };
  }
}

function main() {
  let args;
  try {
    args = parseArgs(process.argv.slice(2));
  } catch (error) {
    fail(error.message, 1);
  }
  const command = args._[0];
  if (!command) fail(usage(), 1);

  if (command === "detect") {
    const platform = detectPlatform(args.platform);
    process.stdout.write(`${JSON.stringify(capabilities(platform), null, 2)}\n`);
    return;
  }

  const file = args.contractPath;
  if (!file) fail(usage(), 1);
  const contract = readJson(path.resolve(file));
  const errors = validateContract(contract);
  if (errors.length) {
    process.stdout.write(`${JSON.stringify({ valid: false, errors }, null, 2)}\n`);
    process.exit(2);
  }

  if (command === "validate") {
    process.stdout.write(`${JSON.stringify({ valid: true, errors: [] }, null, 2)}\n`);
    return;
  }

  if (command === "render") {
    const platform = detectPlatform(args.platform);
    process.stdout.write(`${JSON.stringify(render(contract, platform), null, 2)}\n`);
    return;
  }

  if (command === "launch") {
    const output = launch(contract, args);
    process.stdout.write(`${JSON.stringify(output, null, 2)}\n`);
    process.exit(output.status === "started" || output.status === "dry_run" ? 0 : 2);
  }

  fail(usage(), 1);
}

if (require.main === module) main();

module.exports = { launch, nativeCommand, parseArgs, render, validateContract };
