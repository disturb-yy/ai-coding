# OODA-E 阶段规范

以下每个阶段定义了该工作流阶段的契约。阶段按顺序执行，由入口/出口条件控制。

---

---

## 阶段 0：初始化（Initialization）

### 目标
在任何分析开始前，建立日志基础设施并确定执行路径（SubAgent 或回退本地）。

### 输入
- 技能激活信号
- 项目根路径
- 当前可用的工具列表（检查 spawn_agent 是否可用）

### 输出
```json
{
  "index": {
    "log_file": ".agent/workflow-logs/<ISO8601-ts>-diagnose.log",
    "execution_mode": "subagent | fallback_local | spawn_ok_wait_failed",
    "available_tools": ["spawn_agent 可用/不可用"]
  },
  "chunks": [],
  "decision_summary": "日志目录已初始化。spawn_agent 状态：<可用/不可用>。执行模式：<subagent/fallback_local>。",
  "confidence": 0.0,
  "next_recommended_agent": "doc-index-reader | observe-phase（本地回退）"
}
```

### 启动条件
- 技能激活后**立即**执行。在任何文件读取或分析之前。

### 退出条件（必须全部满足）
- 日志目录 `.agent/workflow-logs/` 已创建（若不存在则 `mkdir -p`）。
- 首条日志已写入：`[go-workflow] phase=Observe prev=start confidence=0.0`。
- 已检测 `spawn_agent`（`multi_agent_v1_spawn_agent`）工具是否在可用工具列表中。
- 若 spawn_agent 可用：已启动 `doc-index-reader` SubAgent。
- 若 spawn_agent 不可用：已写入 `spawn_failed` + `fallback_local` 两条日志，然后开始主 Agent 本地读取索引。
- **已写入日志**: 启动协议中的两条日志（spawn 尝试结果）。

### 源码读取
**禁止。**

### 代码结构工具访问
**禁止。**

### SubAgent 启动
仅尝试 `doc-index-reader`。若失败则回退本地。

### 日志写入示例
```bash
# 步骤 0
mkdir -p .agent/workflow-logs/

# 步骤 1
echo "[go-workflow] phase=Observe prev=start confidence=0.0" >> .agent/workflow-logs/$(date -u +%Y-%m-%dT%H%M%SZ)-diagnose.log

# 步骤 2（若 spawn_agent 不可用）
echo "[go-workflow] agent=doc-index-reader event=spawn_failed phase=Observe reason=unsupported" >> <日志文件>
echo "[go-workflow] agent=fallback event=fallback_local phase=Observe" >> <日志文件>
```

## 阶段 1：Observe（观察）

### 目标
从人工维护的索引文件构建项目认知模型。在不接触源码的情况下，从架构层面理解系统。

### 输入
- 用户请求（任务描述、模块提示、错误信息）
- 项目根路径
- 可用索引文件：`.agent/index.md`、`ARCHITECTURE.md`、`modules.md`、`dataflow.md`、`README.md`、`docs/*`

### 输出
```json
{
  "index": {
    "project_type": "monorepo|single-module",
    "modules": ["<模块名>"],
    "architecture_notes": "<关键架构模式>",
    "relevant_docs": ["<路径>"],
    "unknowns": ["<空白/未知项>"]
  },
  "chunks": [
    {"source": "<文件:行号>", "content": "<相关摘录>", "relevance": "<为何相关>"}
  ],
  "decision_summary": "从索引文件中学到了什么；仍有哪些未知",
  "confidence": 0.0-1.0,
  "next_recommended_agent": "code-structure-analyzer | module-code-analyzer | orient-phase"
}
```

### 启动条件
- 任何工作流触发（feature、bugfix、refactor、test-gen、test-fix、review）。
- 每次调用必须是**第一个**阶段。

### 退出条件（必须全部满足）
- 至少已读取 `.agent/index.md` **或** `README.md`。
- 已确定 `index.project_type`。
- `index.modules` 已填充 >=1 个模块名。
- `decision_summary` 非空。
- 已设置 `next_recommended_agent`。
- **已写入日志**: `[go-workflow] phase=Observe prev=start confidence=<value>` 通过 exec_command 追加到日志文件
  ```bash
  echo "[go-workflow] phase=Observe prev=start confidence=<value>" >> <日志文件>
  ```

### 源码读取
**禁止。** 此阶段只读索引/文档文件。

### 代码结构工具访问
**禁止。** 代码结构工具保留给 Orient 阶段。

### SubAgent 启动
**必须。** 必须启动 `doc-index-reader` 作为主要读取者。绝不由主 Agent 直接读取索引文件（除非 Phase 0 已判定回退本地）。

**回退路径：** 若 Phase 0 中 spawn_agent 不可用，主 Agent 回退执行，但仍禁止读取 .go 源码。

---

## 阶段 2：Orient（定向）

### 目标
通过查询代码结构工具填补 Observe 中识别的空白。构建模块、数据流、依赖关系和入口点的结构性理解。

### 输入
- Observe 阶段 `index`（尤其是 `unknowns` 和 `modules`）
- 代码结构工具 API 端点：`query_module`、`query_flow`、`query_dependency`、`query_entrypoint`

### 输出
```json
{
  "index": {
    "<继承自 Observe，并补充>": "",
    "module_structure": {"<模块>": {"entrypoints": [], "dependencies": [], "flows": []}},
    "impact_scope": ["<受影响的模块/文件>"]
  },
  "chunks": [
    {"source": "structure://query_module?name=X", "content": "<响应>", "relevance": "<为何相关>"}
  ],
  "decision_summary": "代码结构工具揭示了什么；对结构理解的置信度",
  "confidence": 0.0-1.0,
  "next_recommended_agent": "decide-phase | module-code-analyzer"
}
```

### 启动条件
- Observe 阶段已完成且 `confidence < 0.8` **或** `unknowns` 数组非空。
- 若 Observe 达到 `confidence >= 0.8` **且** `unknowns` 为空，直接跳到 Decide。

### 退出条件（必须全部满足）
- 对 `index.modules` 中的每个模块：已调用 `query_module`。
- 对目标模块：已调用 `query_entrypoint`。
- 对跨模块任务：已调用 `query_dependency`。
- 已填充 `impact_scope`。
- `decision_summary` 包含是否需要深入源码的评估。
- **已写入日志**: `[go-workflow] phase=Orient prev=Observe confidence=<value>` 通过 exec_command 追加
  ```bash
  echo "[go-workflow] phase=Orient prev=Observe confidence=<value>" >> <日志文件>
  ```

### 源码读取
**禁止。** 代码结构工具是此阶段的信息上限。

### 代码结构工具访问
**必须。** 必须启动 `code-structure-analyzer` SubAgent 来调用 代码结构工具 API。

### SubAgent 启动
**必须。** 必须启动 `code-structure-analyzer`。绝不由主 Agent 直接调用 代码结构工具 API。

---

## 阶段 3：Decide（决策）

### 目标
综合 Observe + Orient 的输出为可操作计划。确定任务类型、受影响模块、所需 SubAgent 和调度策略。

### 输入
- Observe 阶段输出（索引模型）
- Orient 阶段输出（结构分析）
- 用户请求

### 输出
```json
{
  "index": {
    "task_type": "feature|bugfix|refactor|test-gen|test-fix|review|diagnose",
    "target_modules": ["<模块>"],
    "target_files": ["<文件>"],
    "dispatch_plan": [
      {
        "agent": "<subagent-name>",
        "scope": "<文件或模块>",
        "task": "<描述>",
        "skills": ["<skill-name>"],
        "skills_source": "user|project_config|skill_asset|hardcoded"
      }
    ],
    "estimated_complexity": "low|medium|high"
  },
  "chunks": [],
  "decision_summary": "最终计划：做什么、调度哪些 Agent、预期输出",
  "confidence": 0.0-1.0,
  "next_recommended_agent": "draft-phase"
}
```

### 启动条件
- Orient 阶段已完成（或以 confidence >= 0.8 跳过）。
- 已填充 `impact_scope`。

### 退出条件
- 已确定 `task_type`。
  - `diagnose` 类型：dispatch_plan 可为空数组（跳过 Draft/Precheck/Act，直接 Evaluate）。
  - 其他类型：dispatch_plan 含 >=1 个 Agent。
- 已列出 `target_modules` 和 `target_files`。
- `dispatch_plan` 已分配 >=1 个 Agent（diagnose 和审查任务除外，可能不需要代码变更 Agent）。
- dispatch plan 的 `confidence >= 0.7`。
- **已写入日志**: `[go-workflow] phase=Decide task_type=<type> dispatch_count=<N>` 通过 exec_command 追加
  ```bash
  echo "[go-workflow] phase=Decide task_type=<type> dispatch_count=<N> strategy=<strategy>" >> <日志文件>
  ```

### 源码读取
**禁止。** 决策仅基于索引 + 代码结构工具数据。

### 代码结构工具访问
**禁止。** 已在 Orient 阶段消费。

### SubAgent 启动
**禁止。** 此阶段仅分析；不启动 SubAgent。

---

## 阶段 4：Draft（起草）

### 目标
读取目标源码并生成具体代码变更（或测试计划）作为补丁。这是允许读取源码的第一个阶段。

### 输入
- Decide 阶段 `dispatch_plan`
- Orient 阶段 `module_structure` 和 `impact_scope`

### 输出
```json
{
  "index": {
    "draft_patches": [
      {"file": "<路径>", "description": "<变更内容>", "patch_preview": "<diff>"}
    ],
    "test_plan": ["<测试文件>", "<测试函数>"]
  },
  "chunks": [
    {"source": "<文件:行号>", "content": "<证明变更的源码摘录>", "relevance": "<为何相关>"}
  ],
  "decision_summary": "建议哪些变更、为何建议、用什么测试验证",
  "confidence": 0.0-1.0,
  "next_recommended_agent": "precheck-phase"
}
```

### 启动条件
- Decide 阶段已完成。
- `dispatch_plan` 非空。

### 退出条件
- `dispatch_plan` 中的每个 Agent 已返回结果。
- `draft_patches` 有 >=1 个条目（审查任务可以为空但需说明理由）。
- 每个补丁有描述和预览。
- 对 test-gen/test-fix：已填充 `test_plan`。
- **已写入日志**: `[go-workflow] phase=Draft` 含所有 SubAgent spawn/return + decision_summary，通过 exec_command 追加
  ```bash
  echo "[go-workflow] phase=Draft prev=Decide confidence=<value>" >> <日志文件>
  ```

### 源码读取
**允许，但有约束：**
- 只能通过 `module-code-analyzer` SubAgent 进行。
- 每个 SubAgent 仅读取其分配的 `target_module`、`target_file` 或 `target_function`。
- 禁止全仓库扫描。
- 主 Agent 绝不直接读取源码。

### 代码结构工具访问
**禁止**（除非 SubAgent 在起草过程中发现新模块需要新鲜的 代码结构工具上下文——此时需在输出中标记）。

### SubAgent 启动
**必须。** 每个调度目标启动一个 `module-code-analyzer`。对于并行安全的任务，可并行启动多个。

---

## 阶段 5：Precheck（预检）

### 目标
在应用前验证草案变更。运行静态分析，检查一致性，验证无冲突。

### 输入
- Draft 阶段 `draft_patches`
- Draft 阶段 `test_plan`

### 输出
```json
{
  "index": {
    "precheck_results": [
      {"check": "build", "status": "pass|fail", "detail": "<输出>"},
      {"check": "vet", "status": "pass|fail", "detail": "<输出>"},
      {"check": "conflict", "status": "pass|fail", "detail": "<输出>"}
    ],
    "blockers": ["<问题描述>"]
  },
  "chunks": [],
  "decision_summary": "预检结果：可以应用或需要修订",
  "confidence": 0.0-1.0,
  "next_recommended_agent": "act-phase | draft-phase"
}
```

### 启动条件
- Draft 阶段已完成。
- `draft_patches` 非空。

### 退出条件
- `precheck_results` 至少有 `build` 和 `vet` 条目。
- 若任何检查失败：`blockers` 非空且 `next_recommended_agent` 指向 Draft。
- 若全部通过：`next_recommended_agent` 为 `act-phase`。
- **已写入日志**: `[go-workflow] phase=Precheck build=<pass/fail> vet=<pass/fail>` 通过 exec_command 追加
  ```bash
  echo "[go-workflow] phase=Precheck build=pass vet=pass conflicts=0" >> <日志文件>
  ```

### 源码读取
**禁止**（除非阻塞项需要重新读取——则通过 Draft 阶段路由）。

### 代码结构工具访问
**禁止。**

### SubAgent 启动
**禁止。** Precheck 本地运行：`go build`、`go vet`、补丁冲突检测。

---

## 阶段 6：Act（执行）

### 目标
将已验证的补丁应用到代码库。执行文件修改。

### 输入
- Precheck 阶段结果（必须全部通过）
- Draft 阶段 `draft_patches`

### 输出
```json
{
  "index": {
    "applied_patches": ["<文件>"],
    "patch_results": [
      {"file": "<路径>", "status": "applied|failed", "detail": "<失败原因>"}
    ]
  },
  "chunks": [],
  "decision_summary": "已应用什么、失败什么、下一步",
  "confidence": 0.0-1.0,
  "next_recommended_agent": "evaluate-phase"
}
```

### 启动条件
- Precheck 阶段已完成且**所有**检查通过。
- 所有 `blockers` 已解决。

### 退出条件
- 所有来自 `draft_patches` 的补丁已应用或有合理理由显式跳过。
- `applied_patches` 非空。
- **已写入日志**: `[go-workflow] phase=Act file=<path> status=<applied/failed/skipped>` 每条补丁一条，通过 exec_command 追加
  ```bash
  echo "[go-workflow] phase=Act file=domain/strategy/http_handler.go status=applied" >> <日志文件>
  ```

### 源码读取
**禁止。** 源码已在 Draft 阶段读取。

### 代码结构工具访问
**禁止。**

### SubAgent 启动
**禁止。** 主 Agent 使用 `apply_patch` 工具应用补丁。每个文件独立一个补丁。

---

## 阶段 7：Evaluate（评估）

### 目标
验证已应用的变更：运行测试、检查构建、评估完成度。确定任务是完成还是需要迭代。

### 输入
- Act 阶段 `applied_patches`
- Draft 阶段 `test_plan`

### 输出
```json
{
  "index": {
    "evaluation_results": [
      {"check": "build", "status": "pass|fail", "detail": "<输出>"},
      {"check": "tests", "status": "pass|fail|partial", "detail": "<输出>"},
      {"check": "completeness", "status": "done|needs_iteration", "detail": "<剩余差距>"}
    ],
    "iteration_count": 0,
    "log_analysis": {
      "workflow_quality": {"score": "optimal|normal|degraded", "notes": []},
      "info_gain": {"index_delta": 0.0, "structure_delta": 0.0, "assessment": ""},
      "token_efficiency": {"draft_tokens_total": 0, "assessment": ""},
      "risk_signals": [],
      "recommendations": []
    },
    "visual_summary": {
      "mermaid": "<Mermaid 流程图源码>",
      "confidence_chart": "<ASCII 置信度爬升曲线>",
      "token_flow": "<Token 流向图>",
      "agent_sequence": "<SubAgent 交互时序>",
      "dashboard": "<关键指标仪表盘>",
      "mcp_comparison": "<MCP vs 无MCP 对比表>"
    }
  },
  "chunks": [],
  "decision_summary": "最终评估：任务完成或需要新一轮循环。log_analysis 和 visual_summary 规范见 references/log-analysis.md",
  "confidence": 0.0-1.0,
  "next_recommended_agent": "done | orient-phase"
}
```

### 启动条件
- Act 阶段已完成。

### 退出条件
- 已在受影响模块上运行 `go build ./...`。
- 已运行相关测试（`go test`）。
- 若 `completeness` 为 `needs_iteration`：回退到 Orient 阶段（最多 3 次总迭代）。
- 若 `completeness` 为 `done`：工作流结束。
- **已写入日志**: `[go-workflow] phase=Evaluate completeness=<done/needs_iteration>` 通过 exec_command 追加
  ```bash
  echo "[go-workflow] phase=Evaluate build=pass tests=<N>/<total> completeness=done iterations=<N>" >> <日志文件>
  ```
- **已输出 log_analysis JSON** + **6 种可视化图表** (Mermaid 流程图、置信度曲线、Token 流向、交互时序、仪表盘、MCP 对比)

### 源码读取
**允许（有限）：** 仅读取测试输出/构建错误。绝不用于全新分析。

### 代码结构工具访问
**首次评估禁止。** 仅在迭代次数 >=2 且需要重新 Orient 时允许。

### SubAgent 启动
**必须。** 启动 `test-runner` SubAgent 执行测试。回退时主 Agent 本地执行 `go test`。

### 可视化输出
Evaluate 阶段结束时，根据 `output_detail` 配置输出不同级别的可视化：

输出级别由配置项 `output_detail` 控制，详见 [log-analysis.md](log-analysis.md#可视化输出)。
1. **Mermaid 阶段流程图** — 展示实际执行的阶段路径
2. **诊断报告** — 问题清单（含严重程度 P0-P3、文件路径、建议）
3. **关键指标仪表盘** — 测试通过率、覆盖率、构建状态

生成规则详见 [log-analysis.md](log-analysis.md#可视化输出)。
