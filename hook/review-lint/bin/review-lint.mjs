#!/usr/bin/env node
import { formatReview, reviewProject } from "../lib/review-lint.mjs";

const args = process.argv.slice(2);
let cwd = process.cwd();
let policyPath;
let json = false;

for (let index = 0; index < args.length; index += 1) {
  const arg = args[index];
  if (arg === "--cwd") cwd = args[++index];
  else if (arg === "--policy") policyPath = args[++index];
  else if (arg === "--json") json = true;
  else if (arg === "--help" || arg === "-h") {
    process.stdout.write("Usage: review-lint [--cwd DIR] [--policy FILE] [--json]\n");
    process.exit(0);
  } else {
    process.stderr.write(`Unknown argument: ${arg}\n`);
    process.exit(2);
  }
}

try {
  const result = reviewProject({ cwd, policyPath });
  process.stdout.write(`${json ? JSON.stringify(result, null, 2) : formatReview(result)}\n`);
  process.exit(result.violations.length ? 1 : 0);
} catch (error) {
  process.stderr.write(`review-lint configuration error: ${error.message}\n`);
  process.exit(2);
}
