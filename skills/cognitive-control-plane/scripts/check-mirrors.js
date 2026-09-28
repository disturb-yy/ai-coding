#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");
const policy = require("../config/guard-policy.json");
const SKILL_DIR = path.resolve(process.env.CCP_SKILL_DIR || path.join(__dirname, ".."));
const MIRROR_ROOT = path.join(SKILL_DIR, "zh");

function canonicalFiles(root = SKILL_DIR) {
  const files = [path.join(root, "SKILL.md")];
  const refs = path.join(root, "references");
  if (fs.existsSync(refs)) {
    for (const entry of fs.readdirSync(refs)) {
      if (entry.endsWith(".md") && !entry.endsWith(".zh-CN.md")) files.push(path.join(refs, entry));
    }
  }
  return files;
}

function mirrorForCanonical(file, root = SKILL_DIR) {
  const relative = path.relative(root, file);
  if (relative === "SKILL.md") return path.join(root, "zh", "SKILL.zh-CN.md");
  if (path.dirname(relative) === "references" && relative.endsWith(".md")) {
    return path.join(root, "zh", "references", `${path.basename(relative, ".md")}.zh-CN.md`);
  }
  return null;
}

function checkMirrors(root = SKILL_DIR, files = canonicalFiles(root)) {
  const problems = [];
  const staleMs = Number(process.env.CCP_MIRROR_STALE_MS ?? policy.mirror_stale_ms);
  if (!Number.isFinite(staleMs) || staleMs < 0) return ["CCP_MIRROR_STALE_MS must be a non-negative number"];
  for (const canonical of files) {
    const mirror = mirrorForCanonical(canonical, root);
    if (!mirror) continue;
    if (!fs.existsSync(canonical)) { problems.push(`missing canonical: ${canonical}`); continue; }
    if (!fs.existsSync(mirror)) { problems.push(`missing mirror: ${mirror}`); continue; }
    if (fs.statSync(canonical).mtimeMs - fs.statSync(mirror).mtimeMs > staleMs) {
      problems.push(`mirror freshness marker is older than canonical: ${mirror}`);
    }
  }
  return problems;
}

if (require.main === module) {
  const problems = checkMirrors();
  if (problems.length) {
    process.stderr.write(problems.join("\n") + "\n");
    process.exitCode = 2;
  } else process.stdout.write("Mirror existence/freshness checks passed (translation quality is not checked).\n");
}

module.exports = { SKILL_DIR, MIRROR_ROOT, canonicalFiles, mirrorForCanonical, checkMirrors };
