#!/usr/bin/env node
"use strict";

const fs = require("fs");

const [hooksJson, hookScript] = process.argv.slice(2);
if (!hooksJson || !hookScript) {
  process.stderr.write("Usage: register-codex-hook.js <hooks.json> <hook-script>\n");
  process.exit(1);
}

const command = `"${process.execPath}" "${hookScript}"`;
const doc = JSON.parse(fs.readFileSync(hooksJson, "utf8") || "{}");
doc.hooks ||= {};

function ensureHook(event) {
  doc.hooks[event] ||= [];
  const exists = doc.hooks[event].some((entry) =>
    JSON.stringify(entry).includes("cognitive-control-plane-guard.js")
  );
  if (exists) return;
  doc.hooks[event].unshift({
    matcher: "",
    hooks: [
      {
        type: "command",
        command,
        timeout: 10,
      },
    ],
  });
}

ensureHook("PreToolUse");
ensureHook("PostToolUse");

fs.writeFileSync(hooksJson, `${JSON.stringify(doc, null, 2)}\n`);
