#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const readline = require("node:readline");
const { spawnSync } = require("node:child_process");

const sourceDir = fs.realpathSync(path.resolve(process.env.CCP_SKILL_DIR || path.join(__dirname, "..")));
const homeDir = os.homedir();
const codexHome = path.resolve(process.env.CODEX_HOME || path.join(homeDir, ".codex"));
const skillLinkDirs = [path.join(codexHome, "skills"), path.join(homeDir, ".agents", "skills")];

function skillLinks(installDir) {
  return skillLinkDirs.map(dir => path.join(dir, "cognitive-control-plane"));
}

function usage() {
  console.log(`Cognitive Control Plane 安装器

用法：
  node scripts/install-cli.js             打开交互菜单
  node scripts/install-cli.js --skill-only 仅注册 skill 发现链接
  node scripts/install-cli.js --codex      注册链接并安装 Codex hook
  node scripts/install-cli.js --opencode   注册链接并安装 OpenCode 插件
  node scripts/install-cli.js --all        安装链接及全部宿主集成
  node scripts/install-cli.js --all --install-dir PATH 复制到指定目录后安装
  node scripts/install-cli.js --help       显示帮助

源码目录：${sourceDir}
CODEX_HOME：${codexHome}`);
}

function parseArgs(args) {
  if (args.length === 0) return { choice: null, installDir: null };
  if (args.length === 1 && (args[0] === "--help" || args[0] === "-h")) return { choice: "help", installDir: null };
  const choices = new Map([
    ["--skill-only", "skill-only"],
    ["--codex", "codex"],
    ["--opencode", "opencode"],
    ["--all", "all"],
  ]);
  let choice = null;
  let installDir = null;
  for (let i = 0; i < args.length; i += 1) {
    const arg = args[i];
    if (choices.has(arg)) {
      if (choice !== null) throw new Error("一次只能选择一种安装模式。");
      choice = choices.get(arg);
      continue;
    }
    if (arg === "--install-dir") {
      if (installDir !== null || !args[i + 1] || args[i + 1].startsWith("--")) {
        throw new Error("--install-dir 后必须提供一个目录路径。");
      }
      installDir = args[++i];
      continue;
    }
    throw new Error(`未知选项：${arg}（使用 --help 查看用法）`);
  }
  if (choice === null && installDir !== null) throw new Error("--install-dir 需要同时指定一种安装模式，例如 --all。");
  return { choice, installDir };
}

function promptChoice() {
  return new Promise(resolve => {
    const input = readline.createInterface({ input: process.stdin, output: process.stdout });
    console.log(`\nCognitive Control Plane 安装器\n源码目录：${sourceDir}\n`);
    console.log("  1) 只注册 skill（Codex + Agent Skills）");
    console.log("  2) 注册 skill + Codex hook");
    console.log("  3) 注册 skill + OpenCode 插件");
    console.log("  4) 注册 skill + Codex 和 OpenCode 集成");
    console.log("  0) 退出\n");
    input.question("请选择 [1-4，默认 1]：", answer => {
      const value = answer.trim() || "1";
      const choices = { "1": "skill-only", "2": "codex", "3": "opencode", "4": "all", "0": "cancel" };
      const choice = choices[value] || "invalid";
      if (choice === "cancel" || choice === "invalid") {
        input.close();
        resolve({ choice, installDir: null });
        return;
      }
      input.question("安装目录 [回车直接链接源码]：", directory => {
        input.close();
        resolve({ choice, installDir: directory.trim() || null });
      });
    });
  });
}

function expandHome(value) {
  if (value === "~") return homeDir;
  if (value.startsWith("~/")) return path.join(homeDir, value.slice(2));
  return value;
}

function resolveInstallDir(value) {
  const requested = path.resolve(expandHome(value));
  if (requested === sourceDir) return sourceDir;
  if (requested.startsWith(`${sourceDir}${path.sep}`) || sourceDir.startsWith(`${requested}${path.sep}`)) {
    throw new Error("安装目录不能位于源码目录内部，也不能是源码目录的父目录。");
  }
  let stat;
  try { stat = fs.lstatSync(requested); } catch (error) {
    if (error.code === "ENOENT") return requested;
    throw error;
  }
  if (stat.isSymbolicLink()) {
    if (fs.realpathSync(requested) === sourceDir) return sourceDir;
    throw new Error(`指定的安装目录本身是符号链接，请填写其实际目录：${requested}`);
  }
  if (!stat.isDirectory()) throw new Error(`安装目录已存在且不是目录：${requested}`);
  if (fs.realpathSync(requested) === sourceDir) return sourceDir;
  const entries = fs.readdirSync(requested);
  if (entries.length > 0) {
    const manifest = path.join(requested, "SKILL.md");
    const isThisSkill = fs.existsSync(manifest) && /^name:\s*cognitive-control-plane\s*$/m.test(fs.readFileSync(manifest, "utf8"));
    if (!isThisSkill) throw new Error(`目标目录非空且不是此 skill 的安装目录，为避免覆盖内容已停止：${requested}`);
  }
  return requested;
}

function uniqueSibling(target, suffix) {
  let candidate = `${target}.${suffix}-${Date.now()}`;
  let serial = 1;
  while (fs.existsSync(candidate)) candidate = `${target}.${suffix}-${Date.now()}-${serial++}`;
  return candidate;
}

const defaultFiles = [
  "SKILL.md",
  "README.md",
  "references/context-control.md",
  "references/epistemic-control.md",
  "references/adversarial-control.md",
  "references/output-control.md",
  "references/memory-routing.md",
  "references/skill-orchestration.md",
  "references/orchestration-state.md",
  "references/reviewer-enforcement.md",
  "scripts/setup-memory.js",
];

function installCopy(target, choice) {
  fs.mkdirSync(path.dirname(target), { recursive: true });
  const stage = uniqueSibling(target, `stage-${process.pid}`);
  const targetExists = fs.existsSync(target);
  let backup = null;
  try {
    if (choice === "skill-only") {
      fs.mkdirSync(stage, { recursive: true });
      for (const relative of defaultFiles) {
        const destination = path.join(stage, relative);
        fs.mkdirSync(path.dirname(destination), { recursive: true });
        fs.copyFileSync(path.join(sourceDir, relative), destination);
      }
    } else {
      fs.cpSync(sourceDir, stage, {
        recursive: true,
        force: true,
        filter: item => path.basename(item) !== ".git",
      });
    }
    if (targetExists) {
      backup = uniqueSibling(target, "bak-ccp");
      fs.renameSync(target, backup);
    }
    fs.renameSync(stage, target);
  } catch (error) {
    if (fs.existsSync(stage)) fs.rmSync(stage, { recursive: true, force: true });
    if (backup && fs.existsSync(backup) && !fs.existsSync(target)) fs.renameSync(backup, target);
    throw error;
  }
  console.log(`已复制安装到：${target}${backup ? `（旧安装备份：${backup}）` : ""}`);
}

function planLinks(installDir) {
  if (!fs.existsSync(path.join(sourceDir, "SKILL.md"))) {
    throw new Error(`源码目录中未找到 SKILL.md：${sourceDir}`);
  }
  for (const target of skillLinks(installDir)) {
    if (!fs.existsSync(path.dirname(target))) continue;
    let stat;
    try { stat = fs.lstatSync(target); } catch (error) {
      if (error.code === "ENOENT") continue;
      throw error;
    }
    if (!stat.isSymbolicLink()) {
      throw new Error(`目标路径不是符号链接，为避免覆盖已停止：${target}`);
    }
  }
}

function installLinks(installDir) {
  for (const target of skillLinks(installDir)) {
    fs.mkdirSync(path.dirname(target), { recursive: true });
    let current;
    try { current = fs.readlinkSync(target); } catch (error) {
      if (error.code !== "ENOENT") throw error;
    }
    if (current && path.resolve(path.dirname(target), current) === installDir) {
      console.log(`链接已指向当前源码：${target}`);
      continue;
    }

    let backup;
    if (current !== undefined) {
      backup = `${target}.bak-ccp-${Date.now()}`;
      fs.renameSync(target, backup);
    }
    try {
      fs.symlinkSync(installDir, target, "dir");
      console.log(`已注册：${target} -> ${installDir}${backup ? `（旧链接备份：${backup}）` : ""}`);
    } catch (error) {
      if (backup) fs.renameSync(backup, target);
      throw error;
    }
  }
}

function installHost(host, installDir) {
  const installer = path.join(installDir, "scripts", "host-install.js");
  const result = spawnSync(process.execPath, [installer, "install", host], {
    stdio: "inherit",
    env: { ...process.env, CCP_SKILL_DIR: installDir, CODEX_HOME: codexHome },
  });
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(`宿主集成安装失败（${host}，退出码 ${result.status ?? "未知"}）。`);
}

async function main() {
  let options = parseArgs(process.argv.slice(2));
  if (options.choice === "help") return usage();
  if (options.choice === null) {
    if (!process.stdin.isTTY || !process.stdout.isTTY) {
      throw new Error("交互模式需要终端；非交互安装请指定 --skill-only、--codex、--opencode 或 --all。");
    }
    options = await promptChoice();
  }
  if (options.choice === "cancel") {
    console.log("已取消安装。");
    return;
  }
  if (options.choice === "invalid") throw new Error("选项无效，请选择菜单中的编号。");

  if (options.installDir === null && process.stdin.isTTY && process.stdout.isTTY && process.argv.length > 2) {
    const input = readline.createInterface({ input: process.stdin, output: process.stdout });
    options.installDir = await new Promise(resolve => input.question("安装目录 [回车直接链接源码]：", answer => {
      input.close();
      resolve(answer.trim() || null);
    }));
  }

  const targetDir = options.installDir === null ? sourceDir : resolveInstallDir(options.installDir);
  planLinks(targetDir);
  if (targetDir !== sourceDir) installCopy(targetDir, options.choice);
  const host = { codex: "codex", opencode: "opencode", all: "all" }[options.choice];
  if (host) installHost(host, targetDir);
  installLinks(targetDir);
  console.log(`\n安装完成，skill 目录：${targetDir}`);
  console.log("如果宿主缓存了 skill 或插件列表，请重新打开会话或重载宿主。");
}

main().catch(error => {
  console.error(`安装失败：${error.message}`);
  process.exitCode = 1;
});
