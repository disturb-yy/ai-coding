#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");
const os = require("node:os");
const { pathToFileURL } = require("node:url");
const { register: codexRegister } = require("./register-codex-hook.js");
const { register: opencodeRegister } = require("./register-opencode-plugin.js");
const root = path.resolve(process.env.CCP_SKILL_DIR || path.join(__dirname, ".."));
const codexRoot = process.env.CODEX_HOME || path.join(os.homedir(), ".codex");
const configRoot = process.env.XDG_CONFIG_HOME || path.join(os.homedir(), ".config");
const locations = {
  hook: process.env.CCP_CODEX_HOOK || path.join(codexRoot, "hooks", "cognitive-control-plane-guard.js"),
  hooks: process.env.CCP_CODEX_HOOKS_JSON || path.join(codexRoot, "hooks.json"),
  plugin: process.env.CCP_OPENCODE_PLUGIN || path.join(configRoot, "opencode", "plugins", "cognitive-control-plane-guard"),
  opencode: process.env.CCP_OPENCODE_CONFIG || path.join(configRoot, "opencode", "opencode.json"),
};

function launchers() {
  const core = path.join(root, "scripts", "cognitive-control-plane-guard.js");
  const plugin = pathToFileURL(path.join(root, "scripts", "opencode-plugin.mjs")).href;
  return {
    [locations.hook]: `#!/usr/bin/env node\nrequire(${JSON.stringify(core)}).main().catch(error => { console.error(error.message); process.exitCode = 0; });\n`,
    [path.join(locations.plugin, "index.mjs")]: `export { default } from ${JSON.stringify(plugin)};\n`,
    [path.join(locations.plugin, "package.json")]: fs.readFileSync(path.join(root, "scripts", "opencode-plugin-package.json"), "utf8"),
  };
}

function readConfig(file) {
  return fs.existsSync(file) ? JSON.parse(fs.readFileSync(file, "utf8")) : {};
}

function writeChanged(file, content) {
  if (fs.existsSync(file) && fs.readFileSync(file, "utf8") === content) return;
  fs.mkdirSync(path.dirname(file), { recursive: true });
  if (fs.existsSync(file)) fs.copyFileSync(file, `${file}.bak-ccp-${Date.now()}`);
  const temporary = `${file}.ccp-${process.pid}.tmp`;
  fs.writeFileSync(temporary, content, { mode: fs.existsSync(file) ? fs.statSync(file).mode & 0o777 : 0o600 });
  fs.renameSync(temporary, file);
}

function desiredFiles(host) {
  const files = launchers();
  if (host === "opencode") delete files[locations.hook];
  if (host === "codex") for (const file of Object.keys(files)) if (file !== locations.hook) delete files[file];
  // 先解析全部选中配置；格式错误时不进行部分安装。
  if (host !== "opencode") files[locations.hooks] = JSON.stringify(codexRegister(readConfig(locations.hooks), locations.hook), null, 2) + "\n";
  if (host !== "codex") files[locations.opencode] = JSON.stringify(opencodeRegister(readConfig(locations.opencode), locations.plugin), null, 2) + "\n";
  return files;
}

function verify(files) {
  for (const [file, content] of Object.entries(files)) {
    if (!fs.existsSync(file) || fs.readFileSync(file, "utf8") !== content) throw new Error(`Missing or outdated registration/launcher: ${file}`);
  }
}

function main() {
  const [action, host = "all"] = process.argv.slice(2);
  if (!["install", "verify"].includes(action) || !["all", "codex", "opencode"].includes(host)) {
    throw new Error("Usage: host-install.js install|verify [all|codex|opencode]");
  }
  for (const file of ["scripts/guard-core.js", "scripts/opencode-plugin.mjs"]) {
    if (!fs.existsSync(path.join(root, file))) throw new Error(`Missing skill dependency: ${file}`);
  }
  const files = desiredFiles(host);
  if (action === "install") for (const [file, content] of Object.entries(files)) writeChanged(file, content);
  verify(files);
  console.log(`CCP ${action}: ${host} registration and launchers verified; source=${root}`);
  if (action === "install") console.log("Host hook enable/trust settings were preserved. Reload the host to refresh a cached plugin.");
}

try { main(); } catch (error) { console.error(error.message); process.exitCode = 1; }
