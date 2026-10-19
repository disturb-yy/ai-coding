---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: maintenance_guide
---
# Cognitive Control Plane

这个 skill 用来处理任务中的卡点：关键需求不清楚、结论缺证据、方案需要审视、反复尝试没有进展，或需要交给新会话继续。它先恢复可执行的下一步，再继续完成已授权工作。

## 使用方式

可以直接说：

- “用 cognitive-control-plane 看看为什么一直没进展，查清后继续修复。”
- “检查这份方案的关键漏洞，然后按已确认范围修改。”
- “整理一份新会话能直接接手的提示词，带上已完成工作和验证结果。”

目标和权限已经明确时直接执行。只有会改变结果的关键选择需要提问；用户只要求设计时，交付设计即可。换 skill、换阶段不会重新要求批准已经授权的工作。

普通流程集中在 [SKILL.md](SKILL.md)，只在需要时加载三个参考：

| 参考 | 读取时机 |
| --- | --- |
| [记忆](references/memory-routing.md) | 偏好、配置、习惯或连接信息会影响处理方式 |
| [协作](references/orchestration-state.md) | 确实要委派工作或协调多个执行者 |
| [独立审查](references/reviewer-enforcement.md) | 用户、仓库或已启用的宿主策略要求独立审查 |

已移除旧的 `role/` 角色模板和 `config/skill-orchestration-map.yaml` 路由表；专业流程从当前环境的 skill 目录选择。旧的上下文、证据、方案评审、输出与 skill 路由参考已合入入口。依赖旧文件路径的自定义调用方需改用上述入口和参考。

## 安装到 Codex

准备 Bash、Node.js 和完整的 `cognitive-control-plane` 源码目录。在该目录运行：

```bash
cd /path/to/cognitive-control-plane
export CCP_INSTALL_DIR="$HOME/.local/share/skills/cognitive-control-plane"
node scripts/install-cli.js --skill-only --install-dir "$CCP_INSTALL_DIR"
node "$CCP_INSTALL_DIR/scripts/setup-memory.js" --install
```

`--skill-only` 复制入口、三个参考、README 和记忆设置脚本，共 6 个文件，并在 `${CODEX_HOME:-$HOME/.codex}/skills` 和 `~/.agents/skills` 创建发现链接。第二条命令创建记忆索引，并在 Codex 全局 `AGENTS.md` 中加入索引入口；已有 `AGENTS.override.md` 时写入该文件。**两条命令都执行后，新开 Codex 会话。** 仅安装 skill 不会让未调用 skill 的会话自动读取记忆。

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

只有明确使用宿主工具时才需要这些文档：

- [适配器](adapters/README.md)：校验合同、生成交接和显式启动新进程。
- [工作项调度](references/work-item-scheduling.md)：宿主实际管理队列、租约和续接时使用。
- [追踪字段](references/trace-semantics.md)：调用方要求路由遥测时使用。
- [维护与验证](references/maintenance.md)：修改本 skill 或安装打包时使用。

这些链接指向完整源码中的可选文件；精简安装不包含它们。安装 Hook 或插件不会自动创建调度服务，也不代表真实宿主已经加载成功。

## 记忆和凭据文件

全局指令让 Codex 在配置、用户喜好、使用习惯或连接信息可能影响任务时先读 `index.md`，再按主题读取最少的记录；它不要求显式调用此 skill。具体记录须核对范围、来源和时效。写入持久记忆需要用户授权，并在同一次变更中更新索引；宿主限制仍然适用。

连接记录可以保存服务名、账号别名、登录目标和凭据文件路径。根索引只链接记录；路径不是读取文件或登录的授权。执行已授权的登录任务时，由可信工具在使用时读取凭据文件并只返回状态，避免在聊天、命令参数、日志或记忆中暴露账号密码。宿主不允许读取路径时，应遵守该边界并报告。
