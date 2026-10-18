"use strict";

const fs = require("node:fs");
const path = require("node:path");
const SKILL_DIR = path.resolve(process.env.CCP_SKILL_DIR || path.join(__dirname, ".."));
const PROTECTED_FILES = ["README.md"];
const READ_TOOLS = /^(read|read_file|open|grep|search|view|view_image)$/i;
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
  const protectedFiles = PROTECTED_FILES.map((item) => path.join(root, item));
  return protectedFiles.some((item) => target === item || inside(item, target));
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

function inspect(event, tool, args = {}, cwd = process.cwd(), skillDir = SKILL_DIR) {
  const root = realPath(skillDir);
  const name = String(tool).split(".").pop();
  cwd = args?.workdir || args?.cwd || cwd;
  const paths = READ_TOOLS.test(name) ? structuredPaths(args) :
    SHELL_TOOLS.test(name) ? shellReads(args.cmd || args.command) : [];
  if (event === "PreToolUse" && paths.some((file) => protectedTarget(file, cwd, root))) {
    return { block: true, message: "CCP: read SKILL.md/references for model instructions instead of the human-facing README.md." };
  }
  return { block: false, message: "" };
}

module.exports = { inspect };
