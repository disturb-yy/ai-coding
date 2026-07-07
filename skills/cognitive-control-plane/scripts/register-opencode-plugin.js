#!/usr/bin/env node
"use strict";

const fs = require("fs");
const path = require("path");

const [configPath, pluginPath] = process.argv.slice(2);
if (!configPath || !pluginPath) {
  process.stderr.write("Usage: register-opencode-plugin.js <opencode.json> <plugin-path>\n");
  process.exit(1);
}

const resolvedPlugin = path.resolve(pluginPath);
const doc = JSON.parse(fs.readFileSync(configPath, "utf8") || "{}");
doc.plugin ||= [];
if (!Array.isArray(doc.plugin)) {
  process.stderr.write("Expected opencode.json plugin field to be an array.\n");
  process.exit(1);
}

if (!doc.plugin.includes(resolvedPlugin)) {
  doc.plugin.push(resolvedPlugin);
}

fs.writeFileSync(configPath, `${JSON.stringify(doc, null, 2)}\n`);
