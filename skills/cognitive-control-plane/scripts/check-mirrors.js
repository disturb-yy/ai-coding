#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");

const DEFAULT_SKILL_DIR = path.resolve(__dirname, "..");
const SKILL_DIR = path.resolve(process.env.CCP_SKILL_DIR || DEFAULT_SKILL_DIR);
const MIRROR_ROOT = path.join(SKILL_DIR, "zh");
const STALE_MS = Number(process.env.CCP_MIRROR_STALE_MS || "2000");

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
};
