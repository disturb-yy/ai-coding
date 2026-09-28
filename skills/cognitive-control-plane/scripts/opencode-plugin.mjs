import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const { inspect } = require("./guard-core.js");

export default async function cognitiveControlPlaneGuard(context = {}) {
  const cwd = context.directory || context.worktree || process.cwd();
  return {
    "tool.execute.before": async (input, output) => {
      const result = inspect("PreToolUse", input.tool, output.args || {}, cwd);
      if (result.block) throw new Error(result.message);
    },
    "tool.execute.after": async (input) => {
      const result = inspect("PostToolUse", input.tool, input.args || {}, cwd);
      if (result.message) process.stderr.write(result.message + "\n");
    },
  };
}
