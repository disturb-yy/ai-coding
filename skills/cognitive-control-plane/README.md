---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: maintenance_guide
---
# Cognitive Control Plane

帮助 agent 找到当前真正阻碍任务推进的环节，并继续完成已授权工作。运行规则以 `SKILL.md` 和按需加载的英文 `references/` 为准。

普通会话无需安装 hook、运行调度器或输出固定 YAML。Tiny/Small/Large 仅在影响执行或调用方需要 trace 时分类；Context、Epistemic、Adversarial、Output 按实际瓶颈选择，没有固定四阶段流程。专业 skill 与 worker 分配是不同决定，缺少可选 worker 时可直接执行。

## 文件与依赖

- `SKILL.md`、`references/`：普通会话规则。
- `config/guard-policy.json`、`scripts/guard-core.js`：两个宿主共用的维护策略与实现。
- `scripts/cognitive-control-plane-guard.js`、`scripts/opencode-plugin.mjs`：宿主入口。
- `config/skill-orchestration-map.yaml`、`adapters/`、调度脚本：显式宿主模式的可选能力。
- `eval/cases-v2/`：当前会话路由回归用例；`eval/cases/` 是保留的旧宿主规则用例。
- `zh/`：供人查看的中文镜像，英文是唯一正式来源。

脚本使用 Node.js 标准库；评估另需 Python 和 `eval/requirements.txt`。专业 skills 从当前环境解析，不复制到本目录。虚拟环境、历史评估结果和生成提示词不是 skill 运行依赖。

## 安装与检查

要求 Bash 和 Node.js；安装界面使用 Node.js 标准库，不需额外依赖。默认源码目录为 `/home/jadon/projects/ai-coding/skills/cognitive-control-plane`，其他位置可用 `CCP_SKILL_DIR` 指定。

### 1. 运行安装界面

交互式安装：

```bash
export CCP_SKILL_DIR="$HOME/projects/ai-coding/skills/cognitive-control-plane"
node "$CCP_SKILL_DIR/scripts/install-cli.js"
```

菜单可选：只安装 skill 发现链接、同时安装 Codex hook、同时安装 OpenCode 插件，或安装全部；之后可输入自定义安装目录。留空会直接链接源码目录。指定目录时安装器先复制 skill，再把发现链接和所选宿主集成指向该安装副本。

非交互／自动化安装可使用 `--skill-only`、`--codex`、`--opencode` 或 `--all`；`--help` 显示用法。例如：

```bash
node "$CCP_SKILL_DIR/scripts/install-cli.js" --all \
  --install-dir "$HOME/.local/share/skills/cognitive-control-plane"
```

安装副本不包含 `.git` 元数据。指定目录不存在时自动创建；目录已有此 skill 时先备份旧安装再替换；目录非空且不包含此 skill 时会停止，避免覆盖其他文件。

安装器将 skill 链接到 `$CODEX_HOME/skills`（默认 `~/.codex/skills`）和 `~/.agents/skills`。原目标若是旧符号链接，会先以 `.bak-ccp-<时间戳>` 保存再更新；若目标是普通文件或目录，则停止，不覆盖它。自定义 `CODEX_HOME` 时，运行 Codex 和安装 hook 都需使用同一个值。

### 2. 手动安装宿主 Hook / 插件

普通 skill 会话不要求安装 hook。需要单独安装宿主集成时，也可直接使用底层脚本：

```bash
bash "$CCP_SKILL_DIR/scripts/install-guards.sh" all
bash "$CCP_SKILL_DIR/scripts/verify-install.sh" all
```

将 `all` 换成 `codex` 或 `opencode` 可只处理一个宿主。Codex 注册项写入 `$CODEX_HOME/hooks.json` 并安装薄 hook 启动器；OpenCode 插件和注册项按 `$XDG_CONFIG_HOME`（默认 `~/.config`）写入。安装器会先备份被覆盖的文件，并保留现有 hook 启用／信任配置。若宿主缓存了插件，安装后重启或重新加载宿主。

源码搬迁后重新运行 `install-cli.js --all`（或选择需要的宿主）以更新发现链接和启动器中的源码路径；可同时使用 `--install-dir PATH` 将其复制到稳定安装目录。

### 3. 检查安装与中文镜像

```bash
node "$CCP_SKILL_DIR/scripts/check-mirrors.js"
```

此检查确认镜像存在且时间戳新于英文源文件，不验证翻译质量。运行模型评估另需 Python 与 `eval/requirements.txt`，普通安装不需要。

安装器核对宿主注册项和启动器；缺失或过期的中文镜像会给出警告，不阻塞宿主安装。镜像状态可用上面的 `check-mirrors.js` 单独检查；宿主集成用 `verify-install.sh all` 核对。

安装器也支持通过 `CCP_SKILL_DIR`、`CCP_CODEX_HOOK`、`CCP_CODEX_HOOKS_JSON`、`CCP_OPENCODE_PLUGIN`、`CCP_OPENCODE_CONFIG` 覆盖默认源路径或安装目标。

Hook 不判断任务复杂度，也不强制多 agent。PreToolUse 只对可识别的受保护正文读取给出限制，允许列文件名；它不是通用 shell 安全沙箱。PostToolUse 仅对明确写入的英文正式文件提醒更新镜像，不将正常编辑中间状态判成工具失败。完成整批维护后执行严格镜像检查；时间新鲜度不等于翻译准确性。
