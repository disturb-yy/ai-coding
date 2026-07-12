#!/usr/bin/env node
import { changedGoJavaFiles, formatReview, reviewProject } from "../../lib/review-lint.mjs";

const WRITE_TOOLS = /^(Bash|apply_patch|Edit|Write|MultiEdit|NotebookEdit)$/i;

async function readInput() {
  let input = "";
  for await (const chunk of process.stdin) input += chunk;
  return input.trim() ? JSON.parse(input) : {};
}

try {
  const input = await readInput();
  const event = input.hook_event_name || input.hookEventName;
  const tool = input.tool_name || input.toolName;
  if (event === "PostToolUse" && tool && !WRITE_TOOLS.test(tool)) {
    process.stdout.write("{}\n");
    process.exit(0);
  }

  const cwd = input.cwd || process.cwd();
  const changedFiles = changedGoJavaFiles(input, cwd);

  const result = reviewProject({ cwd, files: changedFiles });
  if (!result.violations.length) {
    process.stdout.write("{}\n");
    process.exit(0);
  }

  const scoped = changedFiles.length > 0
    ? ` (scoped to ${result.checked.length} changed file(s))`
    : "";
  process.stdout.write(`${JSON.stringify({ decision: "block", reason: formatReview(result) + scoped })}\n`);
} catch (error) {
  process.stdout.write(`${JSON.stringify({
    decision: "block",
    reason: `review-lint failed to run: ${error.message}`,
  })}\n`);
}
