#!/usr/bin/env node
import { formatReview, reviewProject } from "../../lib/review-lint.mjs";

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

  const result = reviewProject({ cwd: input.cwd || process.cwd() });
  if (!result.violations.length) {
    process.stdout.write("{}\n");
    process.exit(0);
  }

  process.stdout.write(`${JSON.stringify({ decision: "block", reason: formatReview(result) })}\n`);
} catch (error) {
  process.stdout.write(`${JSON.stringify({
    decision: "block",
    reason: `review-lint failed to run: ${error.message}`,
  })}\n`);
}
