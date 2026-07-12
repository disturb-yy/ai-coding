import { formatReview, reviewProject } from "../../lib/review-lint.mjs";

const WRITE_TOOLS = /^(bash|apply_patch|edit|write|patch|multiedit|notebookedit)$/i;

export default async function reviewLintPlugin({ directory, worktree }) {
  const root = worktree || directory || process.cwd();
  return {
    "tool.execute.after": async (input) => {
      if (!WRITE_TOOLS.test(input.tool)) return;
      const result = reviewProject({ cwd: root });
      if (result.violations.length) throw new Error(formatReview(result));
    },
  };
}
