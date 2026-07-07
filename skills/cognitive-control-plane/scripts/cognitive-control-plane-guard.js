#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");

const SKILL_DIR = path.resolve(process.env.CCP_SKILL_DIR || "/home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane");
const { checkMirrors, protectedReadFiles } = require(path.join(SKILL_DIR, "scripts", "check-mirrors.js"));
const READ_COMMAND_RE = /\b(cat|sed|awk|grep|rg|head|tail|less|more|nl|bat|open|xdg-open|code|vim|nvim|nano|emacs)\b/i;
const READ_TOOL_RE = /^(read|open|grep|glob|find|view|view_image|webfetch|websearch)$/i;
const WRITE_TOOL_RE = /^(write|edit|apply_patch|notebookedit)$/i;

function readStdin() {
  try {
    return fs.readFileSync(0, "utf8");
  } catch {
    return "";
  }
}

function parseJson(text) {
  if (!text || !text.trim()) return {};
  try {
    return JSON.parse(text);
  } catch {
    return {};
  }
}

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

function eventName(payload) {
  return String(
    payload.hook_event_name ||
    payload.hookEventName ||
    payload.hook_event ||
    payload.hookEvent ||
    payload.event ||
    process.argv[2] ||
    ""
  );
}

function toolName(payload) {
  return String(
    payload.tool_name ||
    payload.toolName ||
    payload.name ||
    payload.tool ||
    ""
  );
}

function toolInput(payload) {
  return (
    payload.tool_input ||
    payload.toolInput ||
    payload.input ||
    payload.arguments ||
    payload.args ||
    payload.parameters ||
    payload
  );
}

function normalizedProtectedReadFiles() {
  return protectedReadFiles().map((file) => file.replace(/\\/g, "/"));
}

function commandTokens(text) {
  if (typeof text !== "string") return false;
  return text
    .replace(/["'`]/g, " ")
    .split(/\s+/)
    .filter(Boolean)
    .map((token) => token.replace(/\\/g, "/"));
}

function mentionsProtectedReadTarget(text) {
  if (typeof text !== "string") return false;
  const normalized = text.replace(/\\/g, "/");
  const protectedFiles = normalizedProtectedReadFiles();
  if (protectedFiles.some((file) => normalized.includes(file))) return true;
  for (const token of commandTokens(text)) {
    const clean = token.replace(/[,:;)\]}]+$/g, "");
    if (protectedFiles.some((file) => file === clean || file.startsWith(`${clean.replace(/\/+$/, "")}/`))) {
      return true;
    }
  }
  return false;
}

function isReadLikeTool(name) {
  if (!name) return false;
  const normalized = String(name).replace(/^functions\./, "");
  if (WRITE_TOOL_RE.test(normalized)) return false;
  if (READ_TOOL_RE.test(normalized)) return true;
  return /read|open|grep|search|find|view|fetch/i.test(normalized);
}

function shouldBlockPreTool(payload) {
  const name = toolName(payload);
  const input = toolInput(payload);
  const strings = collectStrings(input);

  if (strings.some(mentionsProtectedReadTarget)) {
    const commandish = strings.find((s) => mentionsProtectedReadTarget(s) && READ_COMMAND_RE.test(s));
    if (commandish) {
      return "Blocked: target file metadata sets access.model_read=false. Models must not read/search/open it.";
    }

    if (isReadLikeTool(name)) {
      return "Blocked: read/search/open tool attempted to access a file with access.model_read=false.";
    }
  }

  return null;
}

function main() {
  const payload = parseJson(readStdin());
  const event = eventName(payload);

  if (event === "PreToolUse") {
    const reason = shouldBlockPreTool(payload);
    if (reason) {
      process.stderr.write(`${reason}\n`);
      process.exit(2);
    }
    process.exit(0);
  }

  if (event === "PostToolUse") {
    const problems = checkMirrors();
    if (problems.length) {
      process.stderr.write(
        [
          "Cognitive Control Plane mirror check failed.",
          "Canonical English files must have synchronized Chinese mirrors under zh/.",
          ...problems.map((p) => `- ${p}`),
        ].join("\n") + "\n"
      );
      process.exit(2);
    }
    process.exit(0);
  }

  process.exit(0);
}

main();
