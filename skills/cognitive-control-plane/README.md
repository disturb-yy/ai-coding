---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: maintenance_guide
---
# Cognitive Control Plane

这个 skill 帮助 Codex 在任务范围、证据、方案或协作方式不清楚时找到下一步，并继续执行已授权工作。普通会话只需 `SKILL.md` 和按需读取的 `references/`；不需要 Hook、配置文件或调度器。

## 安装到 Codex

准备 Bash、Node.js 和完整的 `cognitive-control-plane` 源码目录。在该目录运行：

```bash
cd /path/to/cognitive-control-plane
export CCP_INSTALL_DIR="$HOME/.local/share/skills/cognitive-control-plane"
node scripts/install-cli.js --skill-only --install-dir "$CCP_INSTALL_DIR"
node "$CCP_INSTALL_DIR/scripts/setup-memory.js" --install
```

`--skill-only` 只复制普通会话需要的文件，并在 `${CODEX_HOME:-$HOME/.codex}/skills` 和 `~/.agents/skills` 创建发现链接。第二条命令创建记忆索引，并在 Codex 全局 `AGENTS.md` 中加入索引入口；已有 `AGENTS.override.md` 时写入该文件。**两条命令都执行后，新开 Codex 会话。** 仅安装 skill 不会让未调用 skill 的会话自动读取记忆。

默认记忆目录为 `~/.local/share/cognitive-control-plane/memory`。要指定位置，先预览再安装：

```bash
node "$CCP_INSTALL_DIR/scripts/setup-memory.js" --memory-dir "$HOME/my-ai-memory"
node "$CCP_INSTALL_DIR/scripts/setup-memory.js" --memory-dir "$HOME/my-ai-memory" --install
```

自定义 Codex 配置目录时，让安装命令和 Codex 使用同一个 `CODEX_HOME`；记忆命令也可显式传入 `--codex-home PATH`。安装器不会迁移旧记忆，也不会覆盖已有记录。目录已有文件但缺少 `index.md` 时，需先为这些记录建立索引。

### 检查结果

```bash
readlink "${CODEX_HOME:-$HOME/.codex}/skills/cognitive-control-plane"
readlink "$HOME/.agents/skills/cognitive-control-plane"
test -f "$HOME/.local/share/cognitive-control-plane/memory/index.md" && echo '记忆索引已创建'
grep -l 'cognitive-control-plane:memory:start' "${CODEX_HOME:-$HOME/.codex}/AGENTS.md" "${CODEX_HOME:-$HOME/.codex}/AGENTS.override.md" 2>/dev/null
```

前两行应指向 `$CCP_INSTALL_DIR`；第三行按实际记忆目录检查；最后一行应输出包含记忆标记的全局指令文件。新会话中可请 Codex 说明它读取哪个记忆索引，以检查入口已加载。空索引不会产生记忆内容。

### 更新和其他模式

从更新后的源码目录重跑安装命令，会备份旧安装目录并替换安装副本；随后重跑该副本的 `setup-memory.js --install`，更新全局入口。已有其他指令和记忆记录会保留。若发现链接目标是普通文件或目录，安装器会停止，避免覆盖。

源码中仍保留可选的 Codex Hook、OpenCode 插件和工作项调度工具。需要宿主集成时，从**完整源码目录**运行 `node scripts/install-cli.js --codex --install-dir "$CCP_INSTALL_DIR"`；可将 `--codex` 换成 `--opencode` 或 `--all`，一次只选一种模式。这些模式会复制完整源码。`node scripts/install-cli.js` 打开交互菜单，`--help` 查看参数。宿主集成可在安装副本目录用 `node scripts/host-install.js verify codex` 检查（将 `codex` 换成所选模式），并在安装后重启宿主。OpenCode 的自动记忆读取需要另行配置等效全局指令。

## 记忆和凭据文件

全局指令让 Codex 在配置、用户喜好、使用习惯或连接信息可能影响任务时先读 `index.md`，再按主题读取最少的记录；它不要求显式调用此 skill。具体记录须核对范围、来源和时效。写入持久记忆需要用户授权，并在同一次变更中更新索引；宿主限制仍然适用。

连接记录可以保存服务名、账号别名、登录目标和凭据文件路径。根索引只链接记录；路径不是读取文件或登录的授权。执行已授权的登录任务时，由可信工具在使用时读取凭据文件并只返回状态，避免在聊天、命令参数、日志或记忆中暴露账号密码。宿主不允许读取路径时，应遵守该边界并报告。
