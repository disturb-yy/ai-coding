#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");

const PLATFORMS = new Set(["opencode", "codex", "claude-code", "unknown"]);
const NEXT_ACTIONS = new Set(["delegate_read_only", "delegate_write", "route_skill", "verify", "deliver"]);
const PHASES = new Set(["context", "design", "implementation", "review", "verification"]);

function usage() {
  return [
    "Usage:",
    "  scripts/ccp-adapter.js detect [--platform NAME]",
    "  scripts/ccp-adapter.js validate CONTRACT.json",
    "  scripts/ccp-adapter.js render [--platform NAME] [--task-id ID] CONTRACT.json",
  ].join("\n");
}

function parseArgs(argv) {
  const args = { _: [] };
  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === "--platform") {
      args.platform = argv[++i];
    } else if (arg === "--task-id") {
      args.taskId = argv[++i];
    } else {
      args._.push(arg);
    }
  }
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

  if (contract.next_action === "delegate_write" && task.edits_allowed !== true) {
    errors.push("delegate_write requires $.task.edits_allowed=true");
  }
  if (contract.next_action === "delegate_read_only" && task.edits_allowed !== false) {
    errors.push("delegate_read_only requires $.task.edits_allowed=false");
  }

  return errors;
}

function render(contract, platform, taskId) {
  const launchable = Boolean(taskId);
  return {
    status: launchable ? "started" : "handoff",
    platform,
    task_id: taskId || "",
    subagent_started: launchable,
    reason: launchable
      ? "Host adapter reported a native task id."
      : "No native subagent task id was provided; preserve this as a handoff contract and stop before delegated implementation.",
    contract,
  };
}

function main() {
  const args = parseArgs(process.argv.slice(2));
  const command = args._[0];
  if (!command) fail(usage(), 1);

  if (command === "detect") {
    const platform = detectPlatform(args.platform);
    process.stdout.write(`${JSON.stringify(capabilities(platform), null, 2)}\n`);
    return;
  }

  const file = args._[1];
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
    process.stdout.write(`${JSON.stringify(render(contract, platform, args.taskId || ""), null, 2)}\n`);
    return;
  }

  fail(usage(), 1);
}

main();
