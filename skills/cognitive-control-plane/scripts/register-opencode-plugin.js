#!/usr/bin/env node
"use strict";
const fs = require("node:fs");
const path = require("node:path");

function register(doc, pluginPath) {
  doc.plugin ||= [];
  if (!Array.isArray(doc.plugin)) throw new Error("OpenCode plugin must be an array");
  const target = path.resolve(pluginPath);
  const isCcp = (value) => typeof value === "string" && /(?:^|\/)cognitive-control-plane-guard\/?$/.test(value);
  let found = false;
  doc.plugin = doc.plugin.flatMap((entry) => {
    if (!isCcp(entry) && entry !== target) return [entry];
    if (found) return [];
    found = true;
    return [target];
  });
  if (!found) doc.plugin.push(target);
  return doc;
}

if (require.main === module) {
  const [file, plugin] = process.argv.slice(2);
  if (!file || !plugin) throw new Error("Usage: register-opencode-plugin.js <opencode.json> <plugin-path>");
  fs.writeFileSync(file, JSON.stringify(register(JSON.parse(fs.readFileSync(file, "utf8")), plugin), null, 2) + "\n");
}
module.exports = { register };
