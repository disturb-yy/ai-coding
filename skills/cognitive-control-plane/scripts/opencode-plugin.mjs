import { createRequire } from "module";

const require = createRequire(import.meta.url);
const { checkMirrors, SKILL_DIR } = require("./check-mirrors.cjs");

const READ_COMMAND_RE = /\b(cat|sed|awk|grep|rg|head|tail|less|more|nl|bat|open|xdg-open|code|vim|nvim|nano|emacs)\b/i;
const READ_TOOL_RE = /^(read|open|grep|glob|find|view|webfetch|websearch)$/i;
const WRITE_TOOL_RE = /^(write|edit|apply_patch|notebookedit)$/i;

function collectStrings(value, out = []) {
  if (typeof value === "string") {
    out.push(value);
    return out;
  }
  if (Array.isArray(value)) {
    for (const item of value) collectStrings(item, out);
    return out;
  }
  if (value && typeof value === "object") {
    for (const item of Object.values(value)) collectStrings(item, out);
  }
  return out;
}

function mentionsMirror(text) {
  if (typeof text !== "string") return false;
  const normalized = text.replace(/\\/g, "/");
  const skillRoot = SKILL_DIR.replace(/\\/g, "/");
  return normalized.includes(`${skillRoot}/zh/`) || normalized.includes(".zh-CN.md");
}

function isReadLikeTool(name) {
  if (!name) return false;
  const normalized = String(name).replace(/^functions\./, "");
  if (WRITE_TOOL_RE.test(normalized)) return false;
  if (READ_TOOL_RE.test(normalized)) return true;
  return /read|open|grep|search|find|view|fetch/i.test(normalized);
}

function mirrorReadError() {
  return new Error("Blocked: Chinese mirror files under zh/ are write-only user artifacts. Do not read/search/open them.");
}

function assertNoMirrorRead(tool, args) {
  const strings = collectStrings(args);
  if (!strings.some(mentionsMirror)) return;
  if (isReadLikeTool(tool)) throw mirrorReadError();
  if (strings.some((s) => mentionsMirror(s) && READ_COMMAND_RE.test(s))) {
    throw mirrorReadError();
  }
}

function assertMirrorsCurrent() {
  const problems = checkMirrors();
  if (problems.length) {
    throw new Error([
      "Cognitive Control Plane mirror check failed.",
      "Canonical English files must have synchronized Chinese mirrors under zh/.",
      ...problems.map((p) => `- ${p}`),
    ].join("\n"));
  }
}

export default async function cognitiveControlPlaneGuard() {
  return {
    "command.execute.before": async (input) => {
      assertNoMirrorRead("shell", { command: input.command, arguments: input.arguments });
    },
    "tool.execute.before": async (input, output) => {
      assertNoMirrorRead(input.tool, output.args);
    },
    "tool.execute.after": async () => {
      assertMirrorsCurrent();
    },
  };
}
