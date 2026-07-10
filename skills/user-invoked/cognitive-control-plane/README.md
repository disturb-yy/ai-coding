---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: maintenance_guide
---
# Cognitive Control Plane Skill

> Human-only maintenance document. Models and agents must not read, search, open, summarize, quote, or use this file as runtime instruction. The executable skill contract is `SKILL.md` plus the linked files under `references/`. This README is for humans configuring and maintaining the skill.

## 中文使用说明

`cognitive-control-plane` 是一个平台无关的控制面 Skill：它不直接替代具体执行 agent，而是负责判断任务规模、选择控制面、生成标准委托契约，并把后续执行交给合适的 Skill、worker、adapter 或人工确认。

### 通用工作方式

1. 让主 agent 先读取 `SKILL.md`，并只按需要读取 `references/` 下的控制面文件。
2. 对复杂任务先分类为 `Tiny`、`Small` 或 `Large`。
3. 对 `Large` 任务选择第一个真正阻塞下一步的 surface：`context`、`epistemic`、`adversarial` 或 `output`。
4. 如果需要委托，生成 adapter contract；只有平台 adapter 返回 `status=started`、`task_id` 和 `subagent_started=true` 时，才算真正启动了 subagent。
5. 如果当前平台没有 subagent/task API，必须保留 handoff contract，并在实现前停止，不要假装已经委托。

### OpenCode

OpenCode 侧使用本仓库的 Skill 作为路由策略。当前 bundled OpenCode plugin 只做 guard：阻止读取受保护镜像文件，并检查中英文镜像状态；它本身不启动 subagent。

在 OpenCode 中需要真实 subagent 时，应由 OpenCode adapter 读取 `adapters/contract.schema.json`，把 contract 转换为 OpenCode 原生 task/subagent，并返回 task id。没有 task id 的结果只能视为 handoff。

### Codex

Codex 侧可以直接使用本 Skill 做控制面判断。`codex exec` 评测环境只验证 routing trace，不会启动 subagent。

如果 Codex 运行面暴露了 multi-agent/task 工具，adapter 应把 contract 作为 worker prompt 提交，并要求 worker 回报 `skills_loaded`、`tools_used`、`validation_results` 和 deviations。没有可用 task 工具时，应输出 handoff contract 并停止在实现前。

### Claude Code

Claude Code 侧通常通过 Task tool 启动 subagent。adapter 应把 `task.role` 映射为子任务角色，把 `objective`、`constraints`、`ownership`、`validation` 和 `stop_if` 合成为 Task prompt。

如果 Task tool 不可用，则不要继续执行 Large implementation；保留 contract 给用户或外层调度器。

### Adapter CLI

本仓库提供一个轻量 CLI 供本地验证：

```bash
node scripts/ccp-adapter.js detect --platform codex
node scripts/ccp-adapter.js validate contract.json
node scripts/ccp-adapter.js render --platform opencode contract.json
```

该 CLI 不会真正启动平台 subagent；它只做能力探测、contract 校验和确定性的 handoff/render。真实启动逻辑必须由平台 adapter 实现。

## Purpose

`cognitive-control-plane` is a routing skill for complex AI collaboration. It does not solve every task itself. It classifies work, selects the right control surface, and hands execution to a direct action, another skill, a bounded worker, a verification step, or a deliverable contract.

Use it when process control changes outcome quality:

- context is unclear
- assumptions or evidence are risky
- a plan, prompt, skill, architecture, diff, PR, or agent trace needs critique
- output must become a handoff, implementation contract, or machine-readable artifact
- work needs multiple agents, staged execution, persistent state, or ownership boundaries

## Runtime Files

Models should use these files only:

- `SKILL.md`: the runtime entry point and top-level routing rules
- `references/context-control.md`: context clarification
- `references/epistemic-control.md`: assumptions, evidence, confidence, and causality
- `references/adversarial-control.md`: plan critique, red-team review, and pre-mortem
- `references/output-control.md`: handoff and final output shaping
- `references/orchestration-state.md`: multi-agent, staged, persistent, or delegated work
- `references/skill-orchestration.md`: skill routing and task-contract patterns
- `config/skill-orchestration-map.yaml`: machine-readable routing map
- `config/skill-orchestration-map.example.yaml`: Chinese-commented example documenting every config field
- `references/maintenance.md`: required before changing canonical files or guards

Human-only files:

- `README.md`: this file
- `zh/**`: Chinese mirrors for user visibility only; models must not read them

## Operating Model

The skill uses this sequence:

1. Classify work as Tiny, Small, or Large.
2. For Large implementation work, apply the Implementation Guard before any edit.
3. Select one active control surface.
4. Load only the reference required by that surface.
5. Delegate with a full task contract when specialized skills, MCP, tools, or ownership boundaries matter.
6. Reconcile specialist output before final delivery.

Large implementation edits require one of these before touching source files, tests, schemas, migrations, generated artifacts, or implementation-facing docs:

- a visible delegation contract with phase, skills, references, MCP/tools, ownership, validation, and stop conditions
- an explicit direct-implementation exception explaining why direct work is safer than delegation

## SubAgent Contract Requirements

When delegating, include the fields below. The specialist should not infer them from prose.

```yaml
role: ""
phase: context | design | implementation | review | verification
objective: ""
constraints: []
required_skills: []
required_references: []
required_mcp: []
required_tools: []
search_scope: []
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
edits_allowed: false
expected_output:
  format: ""
  required_fields: []
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
validation: []
stop_if: []
```

Useful contract patterns are in `references/skill-orchestration.md`. The YAML equivalent is in `config/skill-orchestration-map.yaml`; see `config/skill-orchestration-map.example.yaml` for Chinese-commented field-by-field examples.

## Hook Behavior

The bundled guard protects two invariants:

- Chinese mirrors under `zh/` are write-only user artifacts. Models must not read, search, or open them.
- Root `README.md` is a human-only maintenance document. Models must not read, search, or open it.

The guard also runs a post-tool mirror freshness check. If an English canonical file is changed without updating its Chinese mirror, the hook fails the tool result.

Canonical files checked for mirrors:

- `SKILL.md`
- `references/*.md`

Mirrors:

- `zh/SKILL.zh-CN.md`
- `zh/references/*.zh-CN.md`

`README.md`, scripts, and YAML files are not canonical Markdown mirror sources.

## Install Or Refresh Hooks

From the skill root:

```bash
scripts/install-guards.sh
```

Default install locations:

```text
Skill source:
  /home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane

Codex hook:
  /home/jadon/.codex/hooks/cognitive-control-plane-guard.js

Codex hook config:
  /home/jadon/.codex/hooks.json

OpenCode plugin:
  /home/jadon/.config/opencode/plugins/cognitive-control-plane-guard

OpenCode config:
  /home/jadon/.config/opencode/opencode.json
```

The installer:

1. Copies `scripts/cognitive-control-plane-guard.js` into the Codex hook location.
2. Registers the Codex hook in `hooks.json`.
3. Copies the OpenCode plugin files into the OpenCode plugin directory.
4. Registers the OpenCode plugin in `opencode.json`.
5. Creates timestamped backups before modifying existing config files.

Restart OpenCode after changing plugin files or registration.

## Environment Overrides

Use these variables when installing or verifying a non-default copy:

```bash
export CCP_SKILL_DIR="/path/to/cognitive-control-plane"
export CCP_CODEX_HOOK="/path/to/cognitive-control-plane-guard.js"
export CCP_CODEX_HOOKS_JSON="/path/to/hooks.json"
export CCP_OPENCODE_PLUGIN="/path/to/cognitive-control-plane-guard"
export CCP_OPENCODE_CONFIG="/path/to/opencode.json"
```

Then run:

```bash
scripts/install-guards.sh
scripts/verify-install.sh
```

## Validation

Run these from the skill root after edits:

```bash
node scripts/check-mirrors.js
node --check scripts/cognitive-control-plane-guard.js
node --check scripts/register-codex-hook.js
node --check scripts/register-opencode-plugin.js
node --check scripts/check-mirrors.js
node --check /home/jadon/.codex/hooks/cognitive-control-plane-guard.js
scripts/verify-install.sh
```

If the skill validator is available:

```bash
python3 /home/jadon/tool/ai-coding/skills/.system/skill-creator/scripts/quick_validate.py .
```

For YAML routing map syntax:

```bash
python3 -c 'import yaml; yaml.safe_load(open("config/skill-orchestration-map.yaml")); yaml.safe_load(open("config/skill-orchestration-map.example.yaml")); print("YAML ok")'
```

## Updating The Skill

Before editing canonical behavior:

1. Read `references/maintenance.md`.
2. Edit English canonical files first.
3. Update matching Chinese mirrors as write-only output artifacts.
4. Run `node scripts/check-mirrors.js`.
5. If hook/plugin files changed, run `scripts/install-guards.sh` and `scripts/verify-install.sh`.
6. Copy the updated skill into the repository copy, if publishing from a local source tree.
7. Commit only the intended skill files.

## Troubleshooting

Mirror check fails:

- Update the matching file under `zh/`.
- Make sure the mirror timestamp is newer than the canonical file.
- Re-run `node scripts/check-mirrors.js`.

README read is blocked:

- This is expected for models and agents.
- Humans should use normal shell/editor access outside the model tool path.

Chinese mirror read is blocked:

- This is expected.
- Update mirrors by writing/regenerating them from the English canonical source.

OpenCode plugin does not appear active:

- Re-run `scripts/install-guards.sh`.
- Restart OpenCode.
- Run `scripts/verify-install.sh`.

Codex hook does not appear active:

- Check that `/home/jadon/.codex/hooks.json` references `cognitive-control-plane-guard.js`.
- Re-run `scripts/install-guards.sh`.
- Run `scripts/verify-install.sh`.
