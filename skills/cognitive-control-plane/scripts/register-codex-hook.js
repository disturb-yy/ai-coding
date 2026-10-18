#!/usr/bin/env node
"use strict";
const fs = require("node:fs");

function register(doc, hookScript) {
  const quote = (value) => "'" + value.replaceAll("'", "'\\''") + "'";
  const command = `${quote(process.execPath)} ${quote(hookScript)}`;
  doc.hooks ||= {};
  const entries = doc.hooks.PreToolUse ||= [];
  if (!Array.isArray(entries)) throw new Error("PreToolUse must be an array");
  let found = false;
  for (const entry of entries) {
    for (const hook of entry.hooks || []) {
      if (String(hook.command || "").includes("cognitive-control-plane-guard.js")) {
        // 保留注册位置和宿主已有的启用/信任状态。
        const legacy = `${JSON.stringify(process.execPath)} ${JSON.stringify(hookScript)}`;
        if (hook.command !== legacy) hook.command = command;
        found = true;
      }
    }
  }
  if (!found) entries.push({ matcher: "", hooks: [{ type: "command", command, timeout: 10 }] });
  if (Array.isArray(doc.hooks.PostToolUse)) {
    doc.hooks.PostToolUse = doc.hooks.PostToolUse.flatMap((entry) => {
      if (!Array.isArray(entry.hooks)) return [entry];
      const hooks = entry.hooks.filter((hook) => !String(hook.command || "").includes("cognitive-control-plane-guard.js"));
      return hooks.length ? [{ ...entry, hooks }] : [];
    });
    if (!doc.hooks.PostToolUse.length) delete doc.hooks.PostToolUse;
  }
  return doc;
}

if (require.main === module) {
  const [file, script] = process.argv.slice(2);
  if (!file || !script) throw new Error("Usage: register-codex-hook.js <hooks.json> <hook-script>");
  fs.writeFileSync(file, JSON.stringify(register(JSON.parse(fs.readFileSync(file, "utf8")), script), null, 2) + "\n");
}
module.exports = { register };
