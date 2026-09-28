#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");
const { launch, validateContract } = require("./ccp-adapter.js");
const { plan } = require("./work-item-scheduler.js");

function usage() {
  return "Usage: work-item-loop.js --platform codex|opencode --workspace WORKSPACE [--sandbox MODE] [--approval POLICY] [--executable PATH] [--execute] STATE.json CONTRACT.json";
}

function parseArgs(argv) {
  const args = {
    _: [],
    approval: "on-request",
    execute: false,
    executable: "",
    platform: "",
    sandbox: "workspace-write",
    workspace: "",
  };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (["--platform", "--workspace", "--sandbox", "--approval", "--executable"].includes(arg)) {
      const value = argv[++index];
      if (!value || value.startsWith("--")) throw new Error(`${arg} requires a value`);
      args[arg.slice(2)] = value;
    } else if (arg === "--execute") {
      args.execute = true;
    } else {
      args._.push(arg);
    }
  }
  if (args._.length !== 2 || !args.platform || !args.workspace) throw new Error(usage());
  return args;
}

function readJson(file) {
  return JSON.parse(fs.readFileSync(path.resolve(file), "utf8"));
}

function invalid(reason, scheduling) {
  return { status: "invalid", reason, scheduling: scheduling || null };
}

function validateDispatchContract(state, contract, scheduling) {
  const errors = validateContract(contract);
  if (errors.length) return errors.join("; ");
  const task = contract.task;
  if (task.phase !== "work_item") return "dispatch contract must use work_item phase";
  if (task.work_item.id !== state.work_item.id || task.work_item.kind !== state.work_item.kind) {
    return "dispatch contract must match the scheduled work item id and kind";
  }
  if (task.run.attempt !== scheduling.expected_run_attempt) {
    return "dispatch contract run attempt must match the scheduler expected attempt";
  }
  if (scheduling.action === "continue") {
    if (!scheduling.source_run_id || task.run.id === scheduling.source_run_id) {
      return "continuation must use a new run id instead of native resume";
    }
    if (!task.run.resume_checkpoint_ref) {
      return "continuation requires a durable resume_checkpoint_ref";
    }
  }
  return "";
}

function runLoop(state, contract, args) {
  const scheduling = plan(state, { dryRun: !args.execute });
  if (!scheduling.ok) return invalid(scheduling.reason, scheduling);
  if (!["dispatch", "continue"].includes(scheduling.action)) {
    return { status: "not_dispatched", scheduling };
  }
  if (!scheduling.requires_fresh_session) {
    return invalid("scheduler dispatch lacked a fresh-session requirement", scheduling);
  }
  const mismatch = validateDispatchContract(state, contract, scheduling);
  if (mismatch) return invalid(mismatch, scheduling);
  const launched = launch(contract, args);
  return { status: launched.status, scheduling, launch: launched };
}

function main() {
  try {
    const args = parseArgs(process.argv.slice(2));
    const result = runLoop(readJson(args._[0]), readJson(args._[1]), args);
    process.stdout.write(`${JSON.stringify(result, null, 2)}\n`);
    process.exit(["dry_run", "started", "not_dispatched"].includes(result.status) ? 0 : 2);
  } catch (error) {
    process.stdout.write(`${JSON.stringify(invalid(error.message))}\n`);
    process.exit(2);
  }
}

if (require.main === module) main();

module.exports = { parseArgs, runLoop, validateDispatchContract };
