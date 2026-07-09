#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");

const DEFAULT_SKILL_DIR = path.resolve(__dirname, "..");
const SKILL_DIR = path.resolve(process.env.CCP_SKILL_DIR || DEFAULT_SKILL_DIR);
const MIRROR_ROOT = path.join(SKILL_DIR, "zh");
const STALE_MS = Number(process.env.CCP_MIRROR_STALE_MS || "2000");
const ACCESS_SCAN_DIRS = ["README.md", "SKILL.md", "references", "config", "zh"];

function mirrorForCanonical(canonicalPath) {
  const relative = path.relative(SKILL_DIR, canonicalPath);
  if (relative === "SKILL.md") {
    return path.join(MIRROR_ROOT, "SKILL.zh-CN.md");
  }
  if (relative.startsWith(`references${path.sep}`)) {
    const base = path.basename(relative, ".md");
    return path.join(MIRROR_ROOT, "references", `${base}.zh-CN.md`);
  }
  return null;
}

function canonicalFiles() {
  const files = [path.join(SKILL_DIR, "SKILL.md")];
  const refs = path.join(SKILL_DIR, "references");
  try {
    for (const entry of fs.readdirSync(refs)) {
      if (!entry.endsWith(".md")) continue;
      if (entry.endsWith(".zh-CN.md")) continue;
      files.push(path.join(refs, entry));
    }
  } catch {
    // Missing references are reported by normal skill validation.
  }
  return files;
}

function checkMirrors() {
  const problems = [];
  for (const canonical of canonicalFiles()) {
    const mirror = mirrorForCanonical(canonical);
    if (!mirror) continue;

    let cstat;
    let mstat;
    try {
      cstat = fs.statSync(canonical);
    } catch {
      problems.push(`missing canonical: ${canonical}`);
      continue;
    }
    try {
      mstat = fs.statSync(mirror);
    } catch {
      problems.push(`missing mirror for ${canonical}: ${mirror}`);
      continue;
    }
    if (cstat.mtimeMs - mstat.mtimeMs > STALE_MS) {
      problems.push(`mirror older than canonical: ${mirror}`);
    }
  }
  return problems;
}

function parseAccessValue(value) {
  const normalized = String(value || "").trim().replace(/^["']|["']$/g, "").toLowerCase();
  if (normalized === "false") return false;
  if (normalized === "true") return true;
  return undefined;
}

function accessFromYamlBlock(text) {
  const lines = String(text || "").split(/\r?\n/);
  let accessIndent = null;
  const access = {};
  for (const line of lines) {
    if (!line.trim()) continue;
    const accessMatch = line.match(/^(\s*)access:\s*$/);
    if (accessMatch) {
      accessIndent = accessMatch[1].length;
      continue;
    }
    if (accessIndent === null) continue;
    const indent = line.match(/^\s*/)[0].length;
    if (indent <= accessIndent) {
      accessIndent = null;
      continue;
    }
    const match = line.match(/^\s*([A-Za-z0-9_-]+):\s*(.+?)\s*$/);
    if (!match) continue;
    const value = parseAccessValue(match[2]);
    access[match[1]] = value === undefined ? match[2].trim() : value;
  }
  return access;
}

function frontmatterBlock(text) {
  const normalized = String(text || "").replace(/^\uFEFF/, "");
  if (!normalized.startsWith("---\n") && !normalized.startsWith("---\r\n")) return null;
  const match = normalized.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/);
  return match ? match[1] : null;
}

function accessPolicy(filePath) {
  let text;
  try {
    text = fs.readFileSync(filePath, "utf8");
  } catch {
    return {};
  }
  const fm = frontmatterBlock(text);
  if (fm !== null) return accessFromYamlBlock(fm);
  if (filePath.endsWith(".yaml") || filePath.endsWith(".yml")) return accessFromYamlBlock(text);
  return {};
}

function walkFiles(root, out = []) {
  let stat;
  try {
    stat = fs.statSync(root);
  } catch {
    return out;
  }
  if (stat.isFile()) {
    out.push(root);
    return out;
  }
  if (!stat.isDirectory()) return out;
  for (const entry of fs.readdirSync(root)) {
    if (entry === ".git" || entry === "node_modules") continue;
    walkFiles(path.join(root, entry), out);
  }
  return out;
}

function accessFiles() {
  const files = [];
  for (const entry of ACCESS_SCAN_DIRS) {
    walkFiles(path.join(SKILL_DIR, entry), files);
  }
  return files.filter((file) => /\.(md|yaml|yml)$/i.test(file));
}

function protectedReadFiles() {
  return accessFiles().filter((file) => accessPolicy(file).model_read === false);
}

function main() {
  const problems = checkMirrors();
  if (problems.length) {
    process.stderr.write(
      [
        "Cognitive Control Plane mirror check failed.",
        "Canonical English files must have synchronized Chinese mirrors under zh/.",
        ...problems.map((problem) => `- ${problem}`),
      ].join("\n") + "\n"
    );
    process.exit(2);
  }
  process.stdout.write("Cognitive Control Plane mirror check passed.\n");
}

if (require.main === module) {
  main();
}

module.exports = {
  SKILL_DIR,
  MIRROR_ROOT,
  canonicalFiles,
  mirrorForCanonical,
  checkMirrors,
  accessPolicy,
  protectedReadFiles,
};
