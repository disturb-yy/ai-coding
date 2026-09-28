#!/usr/bin/env node
"use strict";

const { inspect } = require("./guard-core.js");

async function main() {
  let input = "";
  for await (const chunk of process.stdin) input += chunk;
  let payload;
  try { payload = JSON.parse(input || "{}"); } catch { return; }
  const event = payload.hook_event_name || payload.hookEventName || payload.hook_event || payload.hookEvent || payload.event || process.argv[2];
  const tool = payload.tool_name || payload.toolName || payload.name || payload.tool || "";
  const args = payload.tool_input || payload.toolInput || payload.input || payload.arguments || payload.args || payload.parameters || {};
  const result = inspect(event, tool, args, payload.cwd || process.cwd());
  if (result.message) process.stderr.write(result.message + "\n");
  process.exitCode = result.block ? 2 : 0;
}

if (require.main === module) main().catch((error) => {
  process.stderr.write(`CCP maintenance hook unavailable: ${error.message}\n`);
  process.exitCode = 0;
});

module.exports = { main };
