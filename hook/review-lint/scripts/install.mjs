#!/usr/bin/env node
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const SOURCE_ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const ALL_TARGETS = ["cac", "claude", "codex", "opencode"];

function quoteCommand(value) {
  return `"${value.replaceAll('"', '\\"')}"`;
}

function copyRuntime(destination) {
  fs.mkdirSync(destination, { recursive: true });
  for (const directory of ["adapters", "bin", "config", "lib"]) {
    fs.cpSync(path.join(SOURCE_ROOT, directory), path.join(destination, directory), {
      recursive: true,
      force: true,
    });
  }
  fs.copyFileSync(path.join(SOURCE_ROOT, "package.json"), path.join(destination, "package.json"));
}

function readJson(file) {
  if (!fs.existsSync(file)) return {};
  return JSON.parse(fs.readFileSync(file, "utf8") || "{}");
}

function backupAndWrite(file, content) {
  const previous = fs.existsSync(file) ? fs.readFileSync(file, "utf8") : null;
  if (previous === content) return false;
  fs.mkdirSync(path.dirname(file), { recursive: true });
  if (previous !== null) {
    const stamp = new Date().toISOString().replace(/[-:TZ.]/g, "");
    fs.copyFileSync(file, `${file}.bak-review-lint-${stamp}`);
  }
  fs.writeFileSync(file, content);
  return true;
}

function ensureCommandHook(doc, event, command, matcher) {
  doc.hooks ||= {};
  doc.hooks[event] ||= [];
  const marker = "/review-lint/adapters/shared/hook.mjs";
  if (doc.hooks[event].some((entry) => JSON.stringify(entry).replaceAll("\\", "/").includes(marker))) {
    return;
  }
  const entry = {
    hooks: [{ type: "command", command, timeout: 120 }],
  };
  if (matcher) entry.matcher = matcher;
  doc.hooks[event].push(entry);
}

function installJsonHooks(file, command, matcher) {
  const doc = readJson(file);
  ensureCommandHook(doc, "PostToolUse", command, matcher);
  ensureCommandHook(doc, "Stop", command);
  backupAndWrite(file, `${JSON.stringify(doc, null, 2)}\n`);
}

export function install({ home = os.homedir(), targets = ALL_TARGETS } = {}) {
  const installRoot = path.join(home, ".local", "share", "review-lint");
  copyRuntime(installRoot);
  const hook = path.join(installRoot, "adapters", "shared", "hook.mjs");
  const command = `${quoteCommand(process.execPath)} ${quoteCommand(hook)}`;

  if (targets.includes("codex")) {
    installJsonHooks(
      path.join(home, ".codex", "hooks.json"),
      command,
      "^(Bash|apply_patch|Edit|Write)$",
    );
  }
  if (targets.includes("cac")) {
    installJsonHooks(
      path.join(home, ".cac", "settings.json"),
      command,
      "^(Bash|Edit|Write|MultiEdit|NotebookEdit)$",
    );
  }
  if (targets.includes("claude")) {
    installJsonHooks(
      path.join(home, ".claude", "settings.json"),
      command,
      "^(Bash|Edit|Write|MultiEdit|NotebookEdit)$",
    );
  }
  if (targets.includes("opencode")) {
    const adapterUrl = pathToFileURL(
      path.join(installRoot, "adapters", "opencode", "review-lint.mjs"),
    ).href;
    const plugin = `export { default } from ${JSON.stringify(adapterUrl)};\n`;
    const pluginFile = path.join(home, ".config", "opencode", "plugins", "review-lint.mjs");
    backupAndWrite(pluginFile, plugin);

    const configFile = path.join(home, ".config", "opencode", "opencode.json");
    const config = readJson(configFile);
    config.plugin ||= [];
    if (!config.plugin.includes(pluginFile)) {
      config.plugin.push(pluginFile);
      backupAndWrite(configFile, `${JSON.stringify(config, null, 2)}\n`);
    }
  }

  return { installRoot, targets };
}

function main() {
  const args = process.argv.slice(2);
  let home = os.homedir();
  let targets = ALL_TARGETS;
  for (let index = 0; index < args.length; index += 1) {
    if (args[index] === "--home") home = path.resolve(args[++index]);
    else if (args[index] === "--targets") targets = args[++index].split(",").filter(Boolean);
    else if (args[index] === "--help" || args[index] === "-h") {
      process.stdout.write("Usage: node scripts/install.mjs [--home DIR] [--targets cac,claude,codex,opencode]\n");
      return;
    } else throw new Error(`unknown argument: ${args[index]}`);
  }
  const unknown = targets.filter((target) => !ALL_TARGETS.includes(target));
  if (unknown.length) throw new Error(`unknown target(s): ${unknown.join(", ")}`);
  const result = install({ home, targets });
  process.stdout.write(`Installed review-lint in ${result.installRoot}\n`);
  process.stdout.write(`Configured: ${result.targets.join(", ")}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  try {
    main();
  } catch (error) {
    process.stderr.write(`Install failed: ${error.message}\n`);
    process.exit(1);
  }
}
