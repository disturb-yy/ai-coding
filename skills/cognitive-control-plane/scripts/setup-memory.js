#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");

const START = "<!-- cognitive-control-plane:memory:start -->";
const END = "<!-- cognitive-control-plane:memory:end -->";

function expandHome(value) {
  return value === "~" ? os.homedir() : value.startsWith("~/") ? path.join(os.homedir(), value.slice(2)) : value;
}

function parseArgs(argv) {
  const options = {
    memoryDir: path.join(os.homedir(), ".local", "share", "cognitive-control-plane", "memory"),
    codexHome: process.env.CODEX_HOME || path.join(os.homedir(), ".codex"),
    install: false,
  };
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === "--help" || arg === "-h") return { help: true };
    if (arg === "--install") { options.install = true; continue; }
    if ((arg === "--memory-dir" || arg === "--codex-home") && argv[i + 1]) {
      options[arg === "--memory-dir" ? "memoryDir" : "codexHome"] = argv[++i];
      continue;
    }
    throw new Error(`Unknown or incomplete option: ${arg}`);
  }
  options.memoryDir = path.resolve(expandHome(options.memoryDir));
  options.codexHome = path.resolve(expandHome(options.codexHome));
  if (/[\r\n`]/.test(options.memoryDir)) throw new Error("Memory directory cannot contain a newline or backtick");
  return options;
}

function activeInstructions(codexHome) {
  const override = path.join(codexHome, "AGENTS.override.md");
  return fs.existsSync(override) ? override : path.join(codexHome, "AGENTS.md");
}

function memoryBlock(memoryDir) {
  const index = path.join(memoryDir, "index.md");
  return `${START}
## Task memory

- Memory index: \`${index}\`. On tasks where configuration, user preferences, usage habits, or connection details could affect the work, read this index before choosing an approach, even if no memory skill was invoked.
- Follow only relevant index entries. Read the smallest linked record needed; check its scope, date, and source against current evidence. Treat missing or stale records as unknown.
- Keep secrets, tokens, and full connection strings out of memory. Write or change persistent memory only with the user's authorization and within host rules; then update the index in the same change.
- A connection record may point to a credential file, but its path is only a locator. For an authorized login, verify the intended service and permitted file access, then use a trusted login tool that reads the file at the point of use and returns status only. Never print credential values into tool output, chat, logs, or memory; if secure use is unavailable, report the boundary.
${END}`;
}

function initialIndex() {
  return `# Memory index

Use this file to find stable facts that affect future tasks. Keep each entry short and link to one focused record. Do not store secrets or full connection strings.

## Routes

| Topic | Scope | Record | Read when |
| --- | --- | --- | --- |

No records yet. Add a row when the first record is created. Typical topics are user preferences, usage habits, configuration, and connection metadata.

## Record rules

- Each record states its scope, source, last verified date, and review date when facts may drift.
- Update or retire an existing record before adding a duplicate. Update this index in the same change.
- The root index links to focused records; keep any credential file path only in a scoped connection record. A path does not authorize reading the file or logging in.
- Connection records may name a service, account alias, or configuration location; never store credential values, tokens, or full connection strings.
`;
}

function replaceBlock(existing, block) {
  const start = existing.indexOf(START);
  const end = existing.indexOf(END);
  if ((start === -1) !== (end === -1)) throw new Error("Existing memory instruction block has only one marker");
  if (start === -1) return existing.trimEnd() + (existing.trim() ? "\n\n" : "") + block + "\n";
  if (end < start || existing.indexOf(START, start + START.length) !== -1 || existing.indexOf(END, end + END.length) !== -1) {
    throw new Error("Existing memory instruction block is malformed or duplicated");
  }
  return existing.slice(0, start) + block + existing.slice(end + END.length);
}

function assertRegularOrMissing(file) {
  try {
    const stat = fs.lstatSync(file);
    if (!stat.isFile()) throw new Error(`Refusing to overwrite a non-file: ${file}`);
  } catch (error) {
    if (error.code !== "ENOENT") throw error;
  }
}

function install(options) {
  const instructions = activeInstructions(options.codexHome);
  const index = path.join(options.memoryDir, "index.md");
  assertRegularOrMissing(index);
  assertRegularOrMissing(instructions);
  if (!fs.existsSync(index) && fs.existsSync(options.memoryDir) && fs.readdirSync(options.memoryDir).length) {
    throw new Error(`Memory directory already has files but no index.md; index those records before installation: ${options.memoryDir}`);
  }
  const oldInstructions = fs.existsSync(instructions) ? fs.readFileSync(instructions, "utf8") : "";
  const nextInstructions = replaceBlock(oldInstructions, memoryBlock(options.memoryDir));
  if (!options.install) {
    process.stdout.write(`Preview only. Memory index: ${index}\nGlobal instructions: ${instructions}\n\n${memoryBlock(options.memoryDir)}\n`);
    return;
  }
  fs.mkdirSync(options.memoryDir, { recursive: true, mode: 0o700 });
  fs.mkdirSync(options.codexHome, { recursive: true });
  if (!fs.existsSync(index)) fs.writeFileSync(index, initialIndex(), { flag: "wx", mode: 0o600 });
  if (nextInstructions !== oldInstructions) {
    if (oldInstructions) {
      const backup = `${instructions}.bak-ccp-memory-${Date.now()}`;
      fs.copyFileSync(instructions, backup);
      process.stdout.write(`Backed up global instructions: ${backup}\n`);
    }
    const temporary = `${instructions}.ccp-${process.pid}.tmp`;
    fs.writeFileSync(temporary, nextInstructions, { mode: 0o600 });
    fs.renameSync(temporary, instructions);
  }
  process.stdout.write(`Memory index: ${index}\nGlobal instructions: ${instructions}\nStart a new Codex session to load the global instructions.\n`);
}

function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.help) {
    process.stdout.write("Usage: node scripts/setup-memory.js [--memory-dir PATH] [--codex-home PATH] [--install]\nWithout --install, preview the changes.\n");
    return;
  }
  install(options);
}

if (require.main === module) {
  try { main(); } catch (error) { process.stderr.write(`${error.message}\n`); process.exitCode = 1; }
}

module.exports = { activeInstructions, initialIndex, memoryBlock, parseArgs, replaceBlock };
