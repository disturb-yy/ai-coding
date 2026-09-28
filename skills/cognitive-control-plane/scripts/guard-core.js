"use strict";

const fs = require("node:fs");
const path = require("node:path");
const policy = require("../config/guard-policy.json");
const { SKILL_DIR, checkMirrors, mirrorForCanonical } = require("./check-mirrors.js");
const READ_TOOLS = /^(read|read_file|open|grep|search|view|view_image)$/i;
const WRITE_TOOLS = /^(write|write_file|edit|apply_patch|notebookedit)$/i;
const SHELL_TOOLS = /^(bash|shell|exec_command)$/i;
const READ_COMMANDS = new Set(["cat", "sed", "awk", "grep", "rg", "head", "tail", "less", "more", "nl", "bat"]);

function realPath(file) {
  try { return fs.realpathSync(file); } catch { return path.resolve(file); }
}

function inside(file, dir) {
  return file === dir || file.startsWith(dir + path.sep);
}

function protectedTarget(file, cwd, root) {
  const target = realPath(path.resolve(cwd, file));
  const protectedFiles = policy.protected_files.map((item) => path.join(root, item));
  const protectedDirs = policy.protected_directories.map((item) => path.join(root, item));
  return protectedFiles.some((item) => target === item || inside(item, target)) ||
    protectedDirs.some((item) => inside(target, item) || inside(item, target));
}

function structuredPaths(args) {
  return [args.path, args.file_path, args.filePath, args.filename].filter((value) => typeof value === "string");
}

// 仅识别简单命令的显式路径；这不是 shell 解析器或系统安全边界。
function shellReads(command) {
  if (typeof command !== "string") return [];
  const tokens = command.match(/"[^"]*"|'[^']*'|[^\s]+/g) || [];
  const words = tokens.map((word) => word.replace(/^(["'])(.*)\1$/, "$2"));
  const program = path.basename(words[0] || "");
  if (!READ_COMMANDS.has(program)) return [];
  if (program === "rg" && words.includes("--files")) return [];
  let patternPending = ["rg", "grep", "sed", "awk"].includes(program);
  const takesValue = new Set(["-e", "--regexp", "-f", "--file", "-g", "--glob", "--iglob", "-t", "--type", "-T", "--type-not", "-A", "-B", "-C", "--context", "--max-count", "--encoding"]);
  const files = [];
  for (let index = 1; index < words.length; index++) {
    const word = words[index];
    if (["-e", "--regexp", "-f", "--file"].includes(word)) patternPending = false;
    if (takesValue.has(word)) { index++; continue; }
    if (word.startsWith("-") || /[|;&<>`$]/.test(word)) continue;
    if (patternPending) { patternPending = false; continue; }
    files.push(word);
  }
  return files;
}

function touchedCanonicals(tool, args, cwd, root) {
  if (!WRITE_TOOLS.test(tool)) return [];
  const files = structuredPaths(args);
  const patch = typeof args === "string" ? args : args.patch || args.input || "";
  if (tool === "apply_patch" && typeof patch === "string") {
    for (const match of patch.matchAll(/^\*\*\* (?:Add File|Update File|Delete File|Move to): (.+)$/gm)) files.push(match[1]);
  }
  return [...new Set(files.map((file) => realPath(path.resolve(cwd, file))))]
    .filter((file) => inside(file, root) && mirrorForCanonical(file, root));
}

function inspect(event, tool, args = {}, cwd = process.cwd(), skillDir = SKILL_DIR) {
  const root = realPath(skillDir);
  const name = String(tool).split(".").pop();
  cwd = args?.workdir || args?.cwd || cwd;
  const paths = READ_TOOLS.test(name) ? structuredPaths(args) :
    SHELL_TOOLS.test(name) ? shellReads(args.cmd || args.command) : [];
  if (event === "PreToolUse" && paths.some((file) => protectedTarget(file, cwd, root))) {
    return { block: true, message: "CCP: read the canonical SKILL.md/references instead of human-only README.md or zh/ mirrors." };
  }
  if (event === "PostToolUse") {
    const files = touchedCanonicals(name, args, cwd, root);
    if (files.length) {
      let problems;
      try { problems = checkMirrors(root, files); }
      catch (error) { return { block: false, message: `CCP mirror check deferred: ${error.message}` }; }
      if (problems.length) return { block: false, message: "CCP maintenance reminder: update Chinese mirrors before finishing the batch.\n" + problems.join("\n") };
    }
  }
  return { block: false, message: "" };
}

module.exports = { inspect };
