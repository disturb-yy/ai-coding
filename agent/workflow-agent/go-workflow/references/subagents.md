# SubAgent 定义

每个 SubAgent 返回标准输出契约。主 Agent 是 SubAgent 输出的**唯一**消费者。

> **Skill 绑定优先级：** 用户显式 > 项目 `.agent/workflow-skills.yaml` > Skill `assets/workflow-skills.yaml`。详见 [skill-mapping.md](skill-mapping.md)

## 标准输出契约

```json
{
  "index": {},
  "chunks": [],
  "decision_summary": "",
  "confidence": 0.0,
  "next_recommended_agent": ""
}
```

| 字段 | 类型 | 描述 |
|-------|------|-------------|
| `index` | object | 结构化知识：模块映射、文件列表、测试计划、调度计划 |
| `chunks` | array | 可检索知识片段，含 `source`、`content`、`relevance` 字段 |
| `decision_summary` | string | 对所做决定及原因的简洁总结 |
| `confidence` | float 0-1 | 对输出正确性的自我评估置信度 |
| `next_recommended_agent` | string | 建议的下一阶段名。取值：`orient-phase` \| `decide-phase` \| `draft-phase` \| `precheck-phase` \| `act-phase` \| `evaluate-phase` \| `done`。**禁止**使用 Agent 名 |

---

## doc-index-reader

### 角色
读取项目级文档和索引文件，构建初始项目认知模型。

### 工具能力需求
主 Agent 调度时，根据当前平台可用工具为此 Agent 分配以下能力：

| 能力 | 用途 | 优先级 |
|------|------|:---:|
| Shell 命令执行 | 读取索引/文档文件（`cat`、`head`、`grep` 限制在 `.agent/`、`docs/` 目录） | **主通道** |
| 文件直接读取 | 若平台提供原生文件读取工具，可替代 Shell 命令 | 备选 |
| MCP 资源读取 | 当索引文件通过 MCP 资源端点暴露时可用 | 可选 |

### 权限
- **可读：** `.agent/index.md`、`ARCHITECTURE.md`、`modules.md`、`dataflow.md`、`README.md`、`docs/*.md`
- **禁止：** 任何源码文件（`.go`、`.sum`）、代码结构工具 API / CodeGraph AST 查询

### 输入
- 项目根路径
- 用户请求上下文（任务类型提示）

### 输入验证
开始读取前必须验证输入有效性：
- 项目根路径不存在或为空目录 → `confidence: 0`，`decision_summary` 报告路径无效，`next_recommended_agent` 为 `done`（终止工作流）
- 项目根路径有效但无索引文件 → `confidence: 0.1`，`unknowns: ["未找到索引文件"]`，`next_recommended_agent` 为 `orient-phase`

### 输出（必须符合标准契约）
```json
{
  "index": {
    "project_type": "monorepo|single-module",
    "modules": [
      {"name": "模块/路径", "description": "简要目的"}
    ],
    "architecture_notes": "layered|hexagonal|clean|microservices",
    "entry_points": ["cmd/server/main.go"],
    "relevant_configs": ["config.yaml", ".env.example"],
    "unknowns": ["X 和 Y 之间的模块边界不清晰"],
    "docs_exceeded_limit": false
  },
  "chunks": [
    {"source": "ARCHITECTURE.md:10-25", "content": "系统使用事件驱动...", "relevance": "架构模式影响耦合决策"}
  ],
  "decision_summary": "项目是包含 3 个模块的单体仓库：api、core、infra。架构为分层式。未知：认证中间件如何共享。",
  "confidence": 0.7,
  "next_recommended_agent": "orient-phase"
}
```

### 阅读深度限制
- 最多读取 **3 层深度**：INDEX.md → 其直接引用的文档 → 被引用文档直接引用的文档
- 每个文档最多读取 **200 行**（使用 `head -200` 或等效截断手段）
- 超出限制的文档 → 列在 `unknowns` 中，并将 `docs_exceeded_limit` 设为 `true`

### 调度规则
- 在 Observe 阶段**始终**最先启动。
- 每次工作流调用单实例。

### 置信度校准
| 条件 | confidence |
|------|:---:|
| 所有预期索引文件均找到并成功解析 | 0.8 |
| 部分索引文件缺失（如仅有 README.md，无 .agent/index.md） | 0.5 |
| 未找到任何索引/文档文件 | 0.1 |
| 输入路径无效 | 0 |

---

## code-structure-analyzer

### 角色
查询代码结构图（代码结构工具 / AST 分析工具）以在不读取源码的情况下构建结构性理解。

### 工具能力需求
主 Agent 调度时，根据当前平台为此 Agent 分配以下能力：

| 能力 | 用途 | 优先级 |
|------|------|:---:|
| 代码结构查询 | 查询模块摘要、导出符号、子包 | **主通道** |
| 数据流查询 | 查询模块或符号的数据/控制流图 | **主通道** |
| 依赖关系查询 | 查询模块的导入/被导入关系 | **主通道** |
| 入口点查询 | 查询模块的入口点（main、服务器初始化、导出 API） | **主通道** |

> **平台映射参考：** Codex/OpenCode → `codegraph_*`（参见 SKILL.md 术语映射章节，支持 codegraph / go-codemap 等别名）（`codegraph_node`、`codegraph_context`、`codegraph_trace`、`codegraph_callers`、`codegraph_callees`、`codegraph_impact`）。主 Agent 负责将上述能力映射到当前平台的实际工具名。

### 权限
- **可调用：** 代码结构查询工具（模块摘要、数据流、依赖关系、入口点）
- **禁止：** 源码文件（`.go`）、代码结构查询之外的任何文件系统操作

### 输入
- Observe 阶段输出（`index.modules`、`index.unknowns`）
- 目标任务类型

### 查询终止条件
防止无限制查询：
- 最多查询 **5 个模块**
- 优先顺序：(1) `index.unknowns` 中提到的模块 → (2) `index.modules` 中与任务描述最相关的模块
- 超出 5 个的模块 → 列在 `impact_scope.indirect` 中，并标记 `truncated: true`

### 输出（标准契约）
```json
{
  "index": {
    "module_structure": {
      "api": {
        "entrypoints": ["cmd/server/main.go"],
        "dependencies": ["core", "infra"],
        "flows": ["HTTP 请求 → 中间件 → handler → service → repository"],
        "exported_symbols": ["Handler", "Middleware", "Router"]
      }
    },
    "impact_scope": {
      "direct": ["api/handler/user.go", "core/service/user.go"],
      "indirect": ["api/middleware/auth.go"],
      "truncated": false
    },
    "query_coverage": {
      "queried_modules": ["api", "core"],
      "skipped_modules": ["infra"],
      "failed_queries": []
    }
  },
  "chunks": [
    {"source": "structure://query_module?name=api", "content": "模块 api 提供 HTTP...", "relevance": "目标模块结构"}
  ],
  "decision_summary": "User 模块跨越 api/handler、core/service、core/repository。依赖链：api→core→infra。",
  "confidence": 0.85,
  "next_recommended_agent": "decide-phase"
}
```

### impact_scope 分类规则
| 分类 | 判定条件 |
|------|---------|
| `direct` | 用户请求明确命名的模块中的文件；或在数据流中出现在目标函数调用链上的文件 |
| `indirect` | 仅出现在依赖关系中的依赖方；或调用链上游的调用者（非任务直接目标） |

### 错误行为
- 单个模块查询失败 → 标记该模块为 `skipped_modules`，继续处理剩余模块
- 全部模块查询失败 → `confidence: 0.1`，`decision_summary` 报告 "代码结构查询不可用"，`next_recommended_agent` 为 `decide-phase`（主 Agent 将回退到直接读源码）
- 查询超时 → 重试 1 次；仍超时 → 视为该模块查询失败

### 调度规则
- 当 Observe 阶段存在 `unknowns` **或** confidence < 0.8 时启动。
- 在 Orient 阶段启动。
- 单实例。所有代码结构查询通过这一个 Agent。

### 置信度校准
| 条件 | confidence |
|------|:---:|
| 所有模块查询成功，impact_scope 完整 | 0.9 |
| 部分模块查询失败，但核心模块有结果 | 0.7 |
| 全部模块查询失败 | 0.1 |

---

## module-code-analyzer

### 角色
在限定范围内读取特定源文件。生成具体代码分析、变更建议或测试计划。

### 工具能力需求
主 Agent 调度时，根据当前平台为此 Agent 分配以下能力：

| 能力 | 用途 | 优先级 |
|------|------|:---:|
| Shell 命令执行 | 读取 scope 范围内的源文件；执行 `go build` / `go vet` 进行自校验 | **主通道** |
| 文件直接读取 | 若平台提供原生文件读取工具，可替代 Shell 命令读取源文件 | 备选 |
| 补丁预览 | 预览补丁效果（dry-run），不实际应用 | 可选 |

> **平台映射参考：** 补丁应用在 Codex 中用 `apply_patch`，OpenCode 中用 `edit`。本 Agent 仅生成补丁内容，实际应用由主 Agent 的 Act 阶段负责。

### 权限
- **可读：** 仅分配 `scope` 内的文件（模块、目录或显式文件列表）。
- **禁止：** 分配范围外的文件、全仓库扫描、代码结构查询工具。

### 输入
- Decide 阶段 `dispatch_plan` 条目，指定：
  - `scope`：模块名、文件路径或函数名
  - `task`：需要分析或生成什么
  - `context_chunks`：来自 Observe/Orient 的相关 chunks

### 输出（必须符合标准契约）

输出字段按任务类型分化：

| task_type | draft_patches | test_plan | 特有字段 |
|-----------|:---:|:---:|---|
| feature | ✅ | ✅ | — |
| bugfix | ✅ | ✅ | `root_cause`: string |
| refactor | ✅ | ✅ | `migration_notes`: string |
| test-gen | ❌ | ✅ | `coverage_target`: float, `generated_file`: string |
| test-fix | ✅ | ❌ | `flaky_root_cause`: string |
| review | ❌ | ❌ | `findings`: [{severity, file, line, issue, suggestion}] |

**通用输出 Schema（feature/bugfix/refactor）：**

```json
{
  "index": {
    "analyzed_files": ["api/handler/user.go"],
    "draft_patches": [
      {
        "file": "api/handler/user.go",
        "description": "为 CreateUser handler 添加邮箱验证",
        "patch": "*** Begin Patch\n*** Update File: api/handler/user.go\n@@ -45,6 +45,10 @@ func CreateUser(...) {\n+\tif !validEmail(req.Email) {\n+\t\treturn nil, ErrInvalidEmail\n+\t}\n*** End Patch"
      }
    ],
    "test_plan": [
      {"file": "api/handler/user_test.go", "function": "TestCreateUser_InvalidEmail"}
    ]
  },
  "chunks": [
    {"source": "api/handler/user.go:42-58", "content": "func CreateUser(req CreateUserRequest)...", "relevance": "需要修改的目标函数"}
  ],
  "decision_summary": "CreateUser handler 缺少邮箱验证。建议：在 service 调用前添加正则检查。测试：验证无效邮箱返回 400。",
  "confidence": 0.9,
  "next_recommended_agent": "precheck-phase"
}
```

**补丁格式约束：**
- 每个 `draft_patches[].patch` 必须为**完整 unified diff 格式**，包含 `*** Begin Patch` / `*** End Patch` 标记
- 可直接作为代码修改工具的输入，无需主 Agent 二次组装
- `patch` 中的文件路径必须与 `file` 字段一致
- 主 Agent 的 Act 阶段负责将补丁适配到当前平台的具体修改工具（如 Codex 的 `apply_patch`、OpenCode 的 `edit`）

### 自校验
输出前必须对 `draft_patches` 执行自校验：
1. 补丁中引用的符号（函数名、类型名、变量名）是否在 scope 范围内存在
2. 新增的导入路径是否合法（包是否存在）
3. `go build`（dry-run）是否能通过

自校验失败 → `confidence` 降低 0.2，并在 `decision_summary` 中标注 `self_check_issues`。全部通过 → `confidence` 保持原值。

### 调度规则
- 仅在 Draft 阶段启动。
- 每个调度目标一个实例（范围不重叠时可并行）。
- 范围重叠时：串行执行，将前序 Agent 输出作为额外上下文传递。
- 最多 3 个并发实例，避免上下文碎片化。

### 置信度校准
| 条件 | confidence |
|------|:---:|
| 自校验全部通过，补丁完整 | 0.9 |
| 自校验有警告但补丁可提交 | 0.7 |
| 无法完成任务（scope 内无目标代码/补丁无法生成） | 0.3 |

---

## test-runner

### 角色
在 Evaluate 阶段为受影响模块运行测试。

### 工具能力需求
主 Agent 调度时，根据当前平台为此 Agent 分配以下能力：

| 能力 | 用途 | 优先级 |
|------|------|:---:|
| Shell 命令执行 | 执行 `go test` 命令并读取输出 | **唯一通道** |

### 权限
- **可执行：** 目标模块目录下的 `go test ./...`
- **可读：** 仅测试输出
- **禁止：** 源码、代码结构查询工具

### 输入
- `test_targets`: 需要测试的模块/包路径列表，如 `["./core/service/...", "./api/handler/..."]`
- `test_flags`: 可选，如 `-count=1`、`-race`、`-cover`、`-timeout=120s`
- `expected_coverage`: 可选，目标覆盖率百分比（用于 test-gen 任务）
- `context`: Act 阶段的 `applied_patches` 列表，用于关联失败测试与补丁

### 输出（必须符合标准契约）
```json
{
  "index": {
    "test_results": {
      "passed": 28,
      "failed": 2,
      "skipped": 0,
      "coverage": "82%",
      "duration_ms": 3400
    },
    "failures": [
      {
        "package": "core/service",
        "test": "TestUserUpdate",
        "error": "timeout after 30s",
        "classification": "flaky",
        "related_patch": "core/service/user.go"
      }
    ],
    "flaky_tests": ["TestMarketScore"],
    "completeness_assessment": "partial"
  },
  "chunks": [
    {"source": "test_output://core/service", "content": "--- FAIL: TestUserUpdate (30.01s)...", "relevance": "失败详情"}
  ],
  "decision_summary": "28/30 通过。TestUserUpdate 超时（flaky），TestMarketScore 断言失败（需回退修复）。",
  "confidence": 0.8,
  "next_recommended_agent": "orient-phase"
}
```

### 测试运行约束
- 单次测试运行超时：`go test -timeout=120s`（可被输入 `test_flags` 覆盖）
- **Flaky 检测**：失败测试自动重跑 1 次（`-count=2`），两次均失败 → `classification: "regression"`；一次通过、一次失败 → `classification: "flaky"`
- **终止阈值**：失败测试超过 5 个 → 立即终止运行，`completeness_assessment: "blocked"`，不在剩余模块上继续消耗时间

### 禁止
- 不修改测试代码（即使发现测试本身有误）
- 不重写测试断言
- 不在源码目录外运行测试
- 不运行 benchmark（`-bench`）
- 不分析源码来解释失败（解释失败是 module-code-analyzer 的职责，且基于输出而非源码）

### 调度规则
- 在 Evaluate 阶段启动。
- 在 `go build` 通过**之后**运行。

### 置信度校准
| 条件 | confidence |
|------|:---:|
| 全部测试通过，覆盖率达标（如有要求） | 0.95 |
| 部分测试失败但 ≤5 个，completeness_assessment 为 partial | 0.7 |
| 测试被终止（失败 >5 个），completeness_assessment 为 blocked | 0.3 |
