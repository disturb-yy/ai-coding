---
name: gen-index
description: 通过交叉使用 Graphify、CodeMap、目标 rg/源码读取和人工维护文档，为现有仓库生成或刷新 agent map。适用于创建 PROJECT_INDEX.md、NAVIGATION.md、CHANGE_GUIDE.md、FEATURES.md、AI 可读仓库索引、根目录 INDEX.md、仓库导航、功能目录或有证据支撑的项目索引文档。
---

# Gen Index

## 本地化维护

- 修改英文 `SKILL.md` 时，必须在同一次变更中同步更新 `SKILL.zh-CN.md`。
- 模型不得把 `SKILL.zh-CN.md` 作为操作指令或任务上下文读取；该文件只作为面向用户阅读的本地化副本。

## 目的

为现有仓库生成简洁的 agent map。该 map 必须帮助后续 agent 快速回答三个问题：项目做什么，重要行为从哪里开始，以及每个结论由什么证据支撑。使用交叉取证：Graphify 负责架构关系，CodeMap 负责代码导航，目标 `rg`/源码读取负责事实验证。

## 概念空间

```yaml
conceptual_space:
  target_region: "有证据支撑、面向 agent 的仓库地图，通过交叉使用图谱、代码地图、源码和文档证据，说明项目目的、重要行为入口、导航路径、变更触点、风险区域和新鲜度限制。"
  deviation_region:
    - "通用架构长文、README 重写、入门教程或详尽源码清单。"
    - "Graphify 生成、CodeMap 生成，或超出读取既有产物/查询结果作为证据的图谱/代码地图分析工作。"
    - "为没有稳定子项目、package、service、module 或 workflow 边界的普通目录生成目录级指南。"
    - "生成根目录 INDEX.md，除非用户明确要求，或仓库已经把它作为约定。"
  priority_dimensions:
    - "证据先于推断：主要结论必须引用 Graphify、CodeMap、源码、文档，或明确 unknown。"
    - "按层选择工具：Graphify 用于架构，CodeMap 用于 route/module/flow/call chain，`rg` 加源码读取用于验证。"
    - "导航先于完整：优化后续 agent 从哪里开始，而不是记录每个文件。"
    - "业务能力先于目录形状：在事实支持时描述用户可见或领域层面的能力。"
    - "刷新时保留先于重写：保留有用的用户补充说明，但必须验证或标记。"
  entry_conditions:
    - "用户要求 AI 可读仓库索引、agent map、项目索引、导航文件、变更指南、功能目录或根目录 INDEX.md。"
    - "任务是 Graphify、CodeMap、源码或文档变化后刷新生成的 `.agent/` 索引文件。"
    - "代码库在后续 agent 工作前需要有证据支撑的定向 artifact。"
  exit_conditions:
    - "目标文件和配套文件已按输出规则写入，或有意跳过。"
    - "每个列出的仓库路径都存在，或标记为 `unknown`、`generated`、`external`、`planned`。"
    - "主要项目结论引用 Graphify 输出、CodeMap 输出、源码、文档，或显式 unknown/needs-confirmation 标记。"
    - "最终报告列出变更文件、假设、证据限制和 stale-when 刷新触发条件。"
  pre_output_check:
    - "检查 `.agent/FEATURES.md` 是否聚焦业务能力，`.agent/NAVIGATION.md` 是否聚焦从哪里开始。"
    - "检查现有生成文件是否只作为草稿，而不是权威来源。"
    - "检查缺失的 Graphify 或 CodeMap 事实没有被全仓源码扫描猜测静默替代。"
    - "检查没有在范围规则之外创建根目录或目录级索引。"
  sedimentation:
    - "当澄清出的稳定术语会影响后续索引时，写入 `.agent/GLOSSARY.md`；稳定决策写入 `.agent/adr/*.md`。"
    - "把过期、未验证或被反驳的历史索引内容移动到 Unknowns 或删除，不要继续当作事实携带。"
    - "记录新鲜度限制，让后续 Graphify、CodeMap、route、schema 或架构变化有明确刷新触发条件。"
```

## 证据阶梯

使用这条证据阶梯。草稿中的每个主要项目结论都能引用 Graphify 输出、CodeMap 输出、源码、文档，或被明确标记为 unknown 时，输入阶段才算完成。

| 优先级 | 输入 | 用途 | 完成标准 |
| --- | --- | --- | --- |
| 1 | Graphify 产物，尤其是 `graphify-out/graph.json`、`graphify-out/GRAPH_REPORT.md` 和 `graphify-out/.graphify_analysis.json` | 架构、社区、重要节点、跨区域关系、功能流程和推荐导航问题 | 存在时已读取相关 `graphify-out/` 文件。 |
| 2 | CodeMap MCP 结果或产物，尤其是 `.codemap/INDEX.md`、routes、modules、flows、callgraph 和 impact 文件 | 代码入口点、模块边界、路由、调用链、变更触点和受影响测试 | 存在时已检查相关 CodeMap 查询或 `.codemap/` 文件；缺失或过期已说明。 |
| 3 | 目标 `rg` 和源码读取 | 关键行为、入口点、测试和路径验证 | 已在源码中检查列出的仓库路径，或标记为 `unknown`、`generated`、`external`、`planned`。 |
| 4 | 人工维护文档 | 项目目的、领域语言、架构意图、贡献约定、稳定决策 | 存在时已检查 README、架构文档、ADR、术语表和贡献文档。 |
| 5 | 现有生成索引文件 | 增量结构和用户补充说明 | 既有 `.agent/*.md` 或 `INDEX.md` 内容只作为草稿，不作为事实。 |
| 6 | 用户澄清 | 无法安全推断的业务术语、功能边界、架构意图或稳定决策 | 只提出会阻塞准确索引的聚焦问题；把稳定答案记录到生成文件中。 |

## 工具角色

| 工具 | 用于 | 不用于 | 上下文规则 |
| --- | --- | --- | --- |
| Graphify | 架构关系、god nodes、社区、跨文档链接、意外连接和高层能力簇 | 源码行为或路由 handler 的最终证明 | 在打开大范围源码前，查询或读取最小相关图谱/报告片段。 |
| CodeMap | 模块、路由、flow、call graph、impact analysis 和变更触点 | 设计意图、业务词汇，或代码结构中不存在的结论 | 当 `.codemap/` 或 MCP server 可用时，优先于宽泛源码读取使用。 |
| `rg` / 源码读取 | 路径存在性、精确符号、路由注册、测试、配置和最终证据 | 通过全仓扫描推断架构 | 使用来自 Graphify/CodeMap 线索的目标 pattern；避免全仓 dump。 |

## 输出

- 主输出：`.agent/PROJECT_INDEX.md`。
- 配套输出：`.agent/NAVIGATION.md`、`.agent/CHANGE_GUIDE.md` 和 `.agent/FEATURES.md`。
- 可选输出：`.agent/GLOSSARY.md`、`.agent/adr/*.md`，以及在澄清产生稳定项目知识时生成的 `.agent/ARCHITECTURE.md` 或 `.agent/architecture/*.md`。
- 根目录 `INDEX.md`：仅当用户明确要求根目录人类阅读索引，或仓库已经把它作为约定时才生成。

## 工作流

| 步骤 | 动作 | 完成标准 |
| --- | --- | --- |
| **Target** | 说明目标索引路径。默认使用 `.agent/PROJECT_INDEX.md`；只有在用户明确要求或仓库已有约定时，才使用根目录 `INDEX.md`。 | 写入前已明确目标文件和任何配套文件。 |
| **Map** | 先读取 Graphify 输出以确认项目形状。不要从全仓源码扫描开始。 | 可用的相关 `graphify-out/` 文件已检查，或已说明缺失。 |
| **Navigate** | 可用时使用 CodeMap MCP 结果或 `.codemap/` 产物确认代码形状。 | 在打开大量源码前，已列出候选模块、路由、flow、调用链、测试和不确定点。 |
| **Verify** | 使用目标 `rg`、源码文件和人工维护文档验证行为并捕获意图。 | 草稿中的重要入口点、测试和高风险路径有源码或文档证据。 |
| **Preserve** | 增量更新时，只读取现有生成索引文件以保留有用结构和用户补充说明。 | 被保留的说明已验证、明确标为未验证，或移动到 Unknowns。 |
| **Clarify** | 如果缺口是概念性的而不是事实性的，直接提出聚焦问题。 | 阻塞性的术语、功能边界、架构意图或决策缺口已回答，或记录为 `needs confirmation`。 |
| **Write** | 写成导航 artifact，而不是完整代码讲解。 | 文件使用稳定小节名称、紧凑 bullet/table，并符合下面的结构。 |
| **Check** | 使用 `rg --files` 或 `test -e <path>` 验证列出的仓库路径；可用时用 CodeMap/源码验证主要 flow。 | 每个列出的路径都存在，或标记为 `unknown`、`generated`、`external`、`planned`；主要 flow 已引用证据或标记为 `not found`。 |
| **Report** | 总结变更文件、假设、证据限制和刷新需求。 | 最终回复列出修改文件，以及 Graphify 输出变化后需要刷新的部分。 |

## 索引结构

除非项目上下文使某一节不相关，否则包含这些部分：

- Purpose：项目做什么、谁使用。
- System Map：主要区域、职责和起始文件。
- Core Capabilities：业务能力、入口点、主要模块和备注。
- Architecture：架构风格、运行单元、数据存储、集成和横切关注点。
- Navigation：常见任务和起始位置。
- Risk Areas：认证、支付、迁移、调度器、关键流程或其他高影响区域。
- Freshness：仅在可从文件元数据、Graphify 输出、CodeMap 输出或用户上下文确认时记录生成日期；记录使用过的 Graphify/CodeMap 来源路径，以及何时视为过期。如果日期未知，写 `unknown`。
- Evidence：使用过的 Graphify 产物、CodeMap 产物/结果、文档、`rg` pattern 和目标源码文件。
- Unknowns：需要确认或刷新 Graphify 输出的事实。

## 配套文件结构

事实可用时，用以下稳定字段编写配套文件。字段缺失时使用 `unknown` 或 `None found in Graphify`，不要直接省略。不要为了重复根目录总览而创建配套文件。

## 范围与放置位置

根目录 `.agent/` 文件作为项目地图。只有当 package、module、app、service、library 或其他有明确边界的子树需要稳定的本地导航时，才生成目录级 `.agent/` 文件：

- 默认使用目录级 `.agent/NAVIGATION.md` 作为子树指南。覆盖该目录的职责、入口点、关键文件、相关测试、常见流程、相邻模块和本地注意事项。
- 只有当该子树承载明确的业务能力或用户工作流时，才添加目录级 `.agent/FEATURES.md`。
- 只有当该子树的变更有重复出现的触点、测试命令或风险模式时，才添加目录级 `.agent/CHANGE_GUIDE.md`。
- 只有当该子树实际上是独立子项目时，才添加目录级 `.agent/PROJECT_INDEX.md`，例如 monorepo 中的 app、service、package 或 library。

普通目录不需要每种指南都具备。优先维护一份准确的 `NAVIGATION.md`，不要制造多份过期或重复文件。不要把根目录项目总览复制到目录级指南里。

### `.agent/NAVIGATION.md`

每个功能或工作流包含：

```text
Feature:
Start From:
Related Modules:
Related Routes:
Related Flows:
Tests:
Risk:
Source Evidence:
CodeMap Evidence:
```

### `.agent/CHANGE_GUIDE.md`

按常见变更类型组织触点：

```text
Change Type:
Touch:
Typical Flow:
Tests:
Risk:
Evidence:
```

包含项目内常见的变更类型，例如新增 API、修改持久化、添加任务、添加事件、修改认证或修改关键集成。

### `.agent/FEATURES.md`

每个业务能力包含：

```text
Feature:
Description:
Entry Points:
Modules:
Routes:
Flows:
Tests:
Unknowns:
Evidence:
```

### 最小示例

```text
# Feature Navigation

Feature: User Login
Start From: src/routes/login.ts
Related Modules: auth, users
Related Routes: POST /login
Related Flows: Login request -> AuthService -> UserRepository
Tests: tests/auth/login.test.ts
Risk: Auth behavior and session creation
Source Evidence: graphify-out/graph.json; src/routes/login.ts
CodeMap Evidence: .codemap/routes/index.md

# Feature Catalog

Feature: User Login
Description: Authenticates a user and creates a session.
Entry Points: src/routes/login.ts
Modules: auth, users
Routes: POST /login
Flows: Login request -> AuthService -> UserRepository
Tests: tests/auth/login.test.ts
Unknowns: unknown
Evidence: graphify-out/graph.json; .codemap/routes/index.md; src/routes/login.ts
```

## 规则

- 优先描述业务能力，而不是原始目录列表。
- 用 Graphify 产物描述架构结构，用 CodeMap 做代码导航，用目标源码读取验证行为。
- 当用户维护的现有文档与生成事实冲突时，优先采用现有文档中的命名。
- 将现有生成索引文件视为历史输出，而不是权威输入。
- 使用 `unknown`、`not found in Graphify`、`not found in CodeMap` 或 `needs confirmation`，不要猜测。
- 让 `.agent/FEATURES.md` 聚焦业务能力，让 `.agent/NAVIGATION.md` 聚焦从哪里开始阅读或修改代码。
- 生成索引要保持紧凑：足够导航，不写成完整源码清单。

## 示例

### 示例 1：全新 Agent 索引

输入：

```text
Generate agent indexes for this project from Graphify output.
```

预期行为：

```text
读取 `graphify-out/graph.json`、`graphify-out/GRAPH_REPORT.md`、
`graphify-out/.graphify_analysis.json`，以及可用的 `.codemap/` route/module/flow
产物；用目标 `rg` 和源码读取验证重要入口文件；
写入 `.agent/PROJECT_INDEX.md`、`.agent/NAVIGATION.md`、
`.agent/CHANGE_GUIDE.md` 和 `.agent/FEATURES.md`。
```

### 示例 2：增量更新

输入：

```text
Update the project index after Graphify output changed.
```

预期行为：

```text
只把现有 `.agent/PROJECT_INDEX.md`、`.agent/NAVIGATION.md`、
`.agent/CHANGE_GUIDE.md` 和 `.agent/FEATURES.md` 当作历史草稿。
保留有用的用户补充说明，但在重写前用 Graphify 输出、CodeMap 输出和源码文件验证结构与路径。
```

### 示例 3：概念缺口

输入：

```text
Generate indexes, but the module names do not explain the business features.
```

预期行为：

```text
直接提出聚焦问题，澄清业务术语和功能边界。将稳定术语记录到
`.agent/GLOSSARY.md`，或将决策记录到 `.agent/adr/*.md`，然后生成索引。
```

### 示例 4：没有 Graphify 输出

输入：

```text
Create an agent-facing repository index, but this project has no graphify-out folder.
```

预期行为：

```text
说明 Graphify 输出不存在。存在 CodeMap 时先使用 CodeMap，然后使用 README、
架构文档、manifest 和目标源码读取作为证据。把缺失的架构事实标记为
`not found in Graphify`，把缺失的 route/callgraph/flow 事实标记为
`not found in CodeMap`，不要通过全仓源码扫描来推断。
```

### 示例 5：Graphify、CodeMap 和 rg 共同使用

输入：

```text
Generate agent indexes for this Go service; it has graphify-out and .codemap.
```

预期行为：

```text
使用 Graphify 识别核心概念和跨区域关系。使用 CodeMap 的 routes、modules、
flows 和 call graph 命名候选入口点和变更触点。写入 `.agent/PROJECT_INDEX.md`、
`.agent/NAVIGATION.md`、`.agent/CHANGE_GUIDE.md` 和 `.agent/FEATURES.md`
之前，用目标 `rg` 和源码读取验证路径与主要 flow。
```
