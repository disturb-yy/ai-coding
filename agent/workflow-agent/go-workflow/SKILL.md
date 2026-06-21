---
name: go-workflow
description: "面向 Go 项目的 OODA-E 工作流编排。协调多 Agent 完成功能开发、Bug 修复、重构、测试生成、代码审查和问题诊断。当任务涉及修改/创建/审查 Go 源码且项目含索引文件时触发。触发条件：Go 代码变更、go build/test 失败、重构请求、测试覆盖率目标、代码审查请求、定位问题。"
---

# Go Workflow

通过 7 阶段 OODA-E 循环编排多 Agent Go 开发。信息层级：**索引 → 代码结构图谱 → 源码**。
## 🔶 启动协议（必须在任何其他操作之前执行）
技能激活后，主 Agent **必须立即**按顺序执行以下 3 步：

```text
步骤 0: 创建日志目录
  → exec_command: mkdir -p .agent/workflow-logs/

步骤 1: 写入首条日志（Observe 阶段入口）
  → exec_command: echo "[go-workflow] phase=Observe prev=start confidence=0.0" >> .agent/workflow-logs/$(date -u +%Y-%m-%dT%H%M%SZ)-diagnose.log

步骤 2: 尝试启动 doc-index-reader SubAgent
  → spawn_agent(agent_type="explorer", message="读取项目索引文件...")
  → 成功 + wait_agent 可用：等待 SubAgent 返回 → 记录 agent=doc-index-reader event=return
  → spawn 成功但 wait_agent 不可用：记录 agent=doc-index-reader event=spawn_ok_wait_failed → 立即回退到主 Agent 本地读取（不等待）
  → spawn 失败：立即记录 agent=doc-index-reader event=spawn_failed → 回退到主 Agent 本地读取
```

**关键约束：**
- 步骤 0 和步骤 1 **必须**在读取任何项目文件之前完成。
- **跨平台：** `mkdir -p` 和 `echo ... >>` 用 `exec_command` 工具执行，工具自动适配平台 shell 差异。Windows 下若 `.agent/` 创建失败，回退到 `/tmp/go-workflow-logs/`。
- 步骤 1 中的 `task_type` 初始值填 `diagnose`，Decide 阶段再修正为实际类型。
- 日志文件路径使用 UTC ISO 8601 时间戳，格式：`YYYY-MM-DDTHHmmssZ-<task_type>.log`。
- 每条后续阶段日志用 `>>` 追加到同一文件。

---

## 术语映射

本技能中的「**代码结构图谱**」（code structure）是一个**通用概念层**，在不同平台有不同名称。**默认不绑定任何具体工具**，文档中统一使用通用名「代码结构工具」。
如需绑定到具体平台工具，在项目 `.agent/workflow-skills.yaml` 中配置：

| 通用概念 | 配置项 | 可用值 | 不配置时的行为 |
|---------|--------|--------|---------------|
| 代码结构分析工具 | `code_structure_tool` | `CodeMap`, `codegraph`, `go-codemap` | 使用通用名「代码结构工具」 |
| 结构分析 SubAgent | `code-structure-analyzer` | — | Agent 名称固定不变 |
| 结构分析配置项 | `skip_structure_analysis_for_single_file` | — | 配置键名固定不变 |

**配置方式**（`.agent/workflow-skills.yaml`）：

```yaml
terminology:
  code_structure_tool: "CodeMap"  # 设为 CodeMap / codegraph / go-codemap
```

设置后，技能文档中所有出现的「代码结构工具」替换为用户配置的平台名称。
**🔶 约束：** 主 Agent 在启动时必须检查 `code_structure_tool` 配置项，并将配置值注入到所有 SubAgent 的上下文中。若未配置或为空，**不替换**，继续使用通用名「代码结构工具」。

---

## 核心原则
- **主 Agent 永不读 Go 源码**。所有源码分析委托给 SubAgent。
- 绝不跳过层级。索引不足查代码结构图谱，结构不足读源码（仅限范围）。
- 所有 SubAgent 返回标准契约：`{index, chunks, decision_summary, confidence, next_recommended_agent}`。
- 每次阶段转换必须写日志到控制台 + `.agent/workflow-logs/<ts>-<task_type>.log`。

## 工作流总览

```text
Observe → Orient → Decide → Draft → Precheck → Act → Evaluate
                                     ←                        ↓
                                     └──（最多 3 次迭代）──────┘
```

## 阶段速查

| 阶段 | 目标 | 读源码 | 结构工具 | SubAgent | **退出条件（含日志检查点）** |
|------|------|:---:|:---:|------|------|
| Observe | 索引构建项目模型 | 否 | 否 | doc-index-reader | 见 phases.md |
| Orient | 代码结构图谱（scope + dependencies） | 否 | 是 | code-structure-analyzer | 见 phases.md |
| Decide | 任务类型 + 调度计划 | 否 | 否 | 无 | 见 phases.md |
| Draft | 分析源码 + 生成补丁 | 限范围 | 否 | module-code-analyzer | 见 phases.md |
| Precheck | 构建检查 + 冲突检测 | 否 | 否 | 无 | 见 phases.md |
| Act | 应用补丁 | 否 | 否 | 无 | 见 phases.md |
| Evaluate | 测试 + 评估完成度 | 有限 | 2+迭代 | test-runner | 见 phases.md |

---

## 调度决策树
```text
收到任务
├── 始终: doc-index-reader（Observe）
│   ├── Confidence < 0.8 或 unknown 非空 → code-structure-analyzer（Orient）
│   │   └── 记录日志: phase=Orient
│   ├── Confidence ≥ 0.8 且 unknown 为空 → 直接 Decide
│   │   └── 记录日志: skip_reason=confidence_high
│   └── [spawn_agent 不可用 或 wait_agent 不可用] → 记录 spawn_failed/spawn_ok_wait_failed 日志 → 主 Agent 本地执行
│       └── 两条日志：agent=<name> event=spawn_failed/spawn_ok_wait_failed → agent=fallback event=fallback_local
├── Decide: 确定 task_type + dispatch_plan
│   ├── 任务类型枚举: feature | bugfix | refactor | test-gen | test-fix | review | diagnose
│   │   └── diagnose: 仅定位问题、输出诊断报告，不做代码修改
│   ├── 单模块/单文件 → 1 个 module-code-analyzer
│   ├── 多模块/无重叠范围 → dispatch_count 个 module-code-analyzer（最多 3 并发）
│   └── 依赖变更 → 1 个 module-code-analyzer + 全部受影响的文件
├── Draft: module-code-analyzer × dispatch_count
│   ├── 成功 → draft_patches = N
│   └── 失败 → 记录原因 → 可选重启（范围更具体 + 错误详情）
├── Precheck
│   ├── 检查: go build, go vet
│   └── 失败 → 返回 Draft（带错误详情）
├── Act
│   └── 应用所有补丁
└── Evaluate
    ├── test-runner
    ├── 通过 → 记录日志: phase=Evaluate completeness=done
    └── 失败 → 回 Orient（最多 3 次迭代）
        └── 记录日志: iteration=N from=Evaluate to=Orient
```

## SubAgent 回退策略

子 Agent 可用时严格遵循调度决策树。**子 Agent 不可用时**（`spawn_agent` 返回 unsupported、模型 404，或 spawn 成功但 `wait_agent` 不可用）：

| 失效项 | 回退方案 | 日志格式 |
|--------|---------|---------|
| `doc-index-reader` | **主 Agent 本地读取** INDEX.md / README.md / specs/*.md（仍禁止读 .go 文件） | `agent=fallback event=fallback_local phase=Observe` |
| `code-structure-analyzer` | **主 Agent 本地调用** 代码结构工具（仍禁止读源码） | `agent=fallback event=fallback_local phase=Orient` |
| `module-code-analyzer` | **主 Agent 本地读取** 仅限 scope 范围的源码文件（遵循"仅限范围读取"规则） | `agent=fallback event=fallback_local phase=Draft scope=...` |
| `test-runner` | **主 Agent 本地执行** go test 并读取输出 | `agent=fallback event=fallback_local phase=Evaluate` |

**注意**: 回退模式下仍必须遵守所有阶段退出条件（含日志检查点 + Evaluate 的 log_analysis + 6 图表输出）。
### 🔶 回退日志写入机制

回退时必须写入**两条**日志（先 spawn_failed/spawn_ok_wait_failed，再 fallback_local）：

```bash
# 第一条：记录启动失败（依失败类型选择对应 event）
# spawn 工具不可用或模型 404：
echo "[go-workflow] agent=doc-index-reader event=spawn_failed phase=Observe reason=unsupported" >> .agent/workflow-logs/<日志文件>
# spawn 成功但 wait 不可用：
echo "[go-workflow] agent=doc-index-reader event=spawn_ok_wait_failed phase=Observe reason=wait_agent_unsupported" >> .agent/workflow-logs/<日志文件>

# 第二条：记录回退到本地
echo "[go-workflow] agent=fallback event=fallback_local phase=Observe" >> .agent/workflow-logs/<日志文件>
```

**强制规则：** 即使 spawn_agent 工具不可用（工具列表中无 `multi_agent_v1_spawn_agent`），也必须跳过 spawn 尝试，直接写入 `spawn_failed` + `fallback_local` 两条日志后再开始本地执行。spawn 成功但 wait 不可用时写入 `spawn_ok_wait_failed` + `fallback_local`。同时输出一条控制台消息说明回退原因。

## 内联示例

### Bug 修复

**任务**: `go test ./domain/notify/...` 失败 → 定位根因并修复
```
1. Observe: doc-index-reader → INDEX.md + specs/* → index 模块 7 个，识别 notify+strategy 域
   日志: [go-workflow] phase=Observe prev=start confidence=0.0

2. Orient: code-structure-analyzer → query_module(notify) → 定位 Dispatcher + HistoryBuffer 流
   日志: [go-workflow] phase=Orient prev=Observe confidence=0.75

3. Decide: task_type=bugfix, target=domain/notify, dispatch=1, strategy=single
   日志: [go-workflow] phase=Decide task_type=bugfix dispatch_count=1 strategy=single

4. Draft: module-code-analyzer → 读取 dispatcher.go + history.go → 2 处补丁：
   (a) CooldownTracker.Record 从未被调用 → 在 Dispatch() 中插入 Record 循环
   (b) HistoryBuffer.Query 返回降序，测试期望升序 → 修正测试断言

5. Precheck: go build ✓  go vet ✓  conflicts=0

6. Act: apply_patch × 2 files → applied

7. Evaluate: go test ./... ✓ → 输出 log_analysis + 6 图表 → 完成
```

### 问题诊断

**任务**: "定位问题" → 仅诊断不修改代码

更多示例（feature/refactor/test-gen/test-fix/review/diagnose）见 [examples.md](references/examples.md)。

## 迭代限制

最多 3 次完整循环（Observe→Evaluate）。第 3 次即使未完成也输出 **partial** log_analysis 并报告。

## 参考文件
| 文件 | 内容 | 何时加载 |
|------|------|---------|
| [phases.md](references/phases.md) | 7 阶段详细规范（含日志检查点） | 阶段细节不确定时 |
| [subagents.md](references/subagents.md) | 3 个 SubAgent 定义 + 标准契约 | 启动 SubAgent 前 |
| [logging.md](references/logging.md) | 10 类日志格式（kv 格式，含 spawn_failed/fallback_local/spawn_ok_wait_failed） | 阶段转换时 |
| [log-analysis.md](references/log-analysis.md) | log_analysis JSON schema + 6 种可视化图表模板 | Evaluate 阶段 |
| [code-structure-tool.md](references/code-structure-tool.md) | 代码结构工具 API 用法 + 查询顺序规则 | Orient 阶段 |
| [skill-mapping.md](references/skill-mapping.md) | SubAgent → Skill 绑定规则 | Decide 阶段 |
| [examples.md](references/examples.md) | 6 种任务类型完整走查 + Before/After diff | 同类任务时 |
| [execution-rules.md](references/execution-rules.md) | 源码读取规则 + Token 优化 + 失败恢复 | 首次 Draft 前 |