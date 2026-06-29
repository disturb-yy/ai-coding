---
name: gen-index
description: 从 .codemap 文件、CodeMap 事实、目标源码读取和人工维护文档生成或更新 agent 面向项目索引。适用于创建 PROJECT_INDEX.md、NAVIGATION.md、CHANGE_GUIDE.md、AI 可读仓库索引、可选根目录 INDEX.md，或带澄清结论的项目索引文档。
---

# Gen Index

## 本地化维护

- 修改英文 `SKILL.md` 时，必须在同一次变更中同步更新 `SKILL.zh-CN.md`。
- 模型不得把 `SKILL.zh-CN.md` 作为操作指令或任务上下文读取；该文件只作为面向用户阅读的本地化副本。

## 目的

从 CodeMap 生成事实、`.codemap/` 文件、目标源码读取和人工维护的项目文档生成简洁的 agent 面向项目索引。只有在写入索引前缺少业务术语、功能边界、架构意图或稳定决策时，才使用 grilling/domain-modeling 进行澄清。

## 输入

按以下顺序优先使用输入：

1. CodeMap 生成文件或 MCP 事实，包括存在时的 `.codemap/` 目录文件。将 `.codemap/architecture/`、`.codemap/callgraph/`、`.codemap/flows/`、`.codemap/modules/` 和 `.codemap/routes/` 作为架构、调用链、功能流程、模块映射和路由入口的主要生成输入。
2. 针对关键行为、入口点、测试和文件路径验证进行目标源码读取。
3. 人工维护的项目文档：`README.md`、架构文档、ADR、术语表、贡献指南，以及其他承载项目意图的文档。
4. 现有生成索引文件，例如 `.agent/PROJECT_INDEX.md`、`.agent/NAVIGATION.md`、`.agent/CHANGE_GUIDE.md` 或 `INDEX.md`，只作为增量更新的历史草稿。不要把它们当作事实层。
5. 仅当无法安全推断项目含义时，才向用户澄清。

## 输出

- 主输出：`.agent/PROJECT_INDEX.md`。
- 配套输出：`.agent/NAVIGATION.md` 和 `.agent/CHANGE_GUIDE.md`。
- 可选输出：`.agent/GLOSSARY.md`、`.agent/adr/*.md`，以及在澄清产生稳定项目知识时生成的 `.agent/ARCHITECTURE.md` 或 `.agent/architecture/*.md`。
- 根目录 `INDEX.md`：仅当用户明确要求根目录人类阅读索引，或仓库已经把它作为约定时才生成。

## 工作流

1. 说明目标索引路径。默认使用 `.agent/PROJECT_INDEX.md`；只有在用户明确要求或仓库已有约定时，才使用根目录 `INDEX.md`。
2. 先读取 CodeMap 输出，包括存在时的 `.codemap/architecture/`、`.codemap/callgraph/`、`.codemap/flows/`、`.codemap/modules/` 和 `.codemap/routes/` 相关文件。不要从全仓源码扫描开始。
3. 读取目标源码文件和人工维护文档，用于验证行为并捕获意图。
4. 如果存在现有生成索引文件，只读取它们以便在增量更新中保留有用结构和用户补充说明。
5. 如果缺失信息是概念性的，而不是事实性的，运行 `/grilling` session，并使用 `/domain-modeling` 进行澄清。如果该路径不可用，直接提出聚焦澄清问题，并把稳定答案记录到 `.agent/GLOSSARY.md`、`.agent/adr/*.md` 或 `.agent/ARCHITECTURE.md`。
6. 把索引写成导航 artifact，而不是完整代码讲解。
7. 验证列出的路径存在，或明确标记为 `unknown`、`generated`、`external`、`planned`。结束前对仓库路径使用 `rg --files` 或 `test -e <path>` 验证。
8. 报告变更文件、假设，以及 CodeMap 重新生成后需要刷新的部分。

## 索引结构

除非项目上下文使某一节不相关，否则包含这些部分：

- Purpose：项目做什么、谁使用。
- System Map：主要区域、职责和起始文件。
- Core Capabilities：业务能力、入口点、主要模块和备注。
- Architecture：架构风格、运行单元、数据存储、集成和横切关注点。
- Navigation：常见任务和起始位置。
- Risk Areas：认证、支付、迁移、调度器、关键流程或其他高影响区域。
- Evidence：使用过的 CodeMap 文件/事实、文档和目标源码文件。
- Unknowns：需要确认或重新生成 CodeMap 的事实。

## 规则

- 优先描述业务能力，而不是原始目录列表。
- 用 CodeMap 事实描述结构，包括 `.codemap/architecture/`、`.codemap/callgraph/`、`.codemap/flows/`、`.codemap/modules/` 和 `.codemap/routes/`；用目标源码读取验证行为。
- 当用户维护的现有文档与生成事实冲突时，优先采用现有文档中的命名。
- 将现有生成索引文件视为历史输出，而不是权威输入。
- 使用 `unknown`、`not found in CodeMap` 或 `needs confirmation`，不要猜测。
- 除本文件明确描述的 `/grilling` 和 `/domain-modeling` 澄清路径外，不要隐式调用其他 skill。

## 示例

### 示例 1：全新 Agent 索引

输入：

```text
Generate agent indexes for this project from .codemap.
```

预期行为：

```text
读取 `.codemap/architecture/`、`.codemap/modules/`、`.codemap/routes/`、
`.codemap/flows/` 和 `.codemap/callgraph/`；用目标源码读取验证重要入口文件；
写入 `.agent/PROJECT_INDEX.md`、`.agent/NAVIGATION.md` 和
`.agent/CHANGE_GUIDE.md`。
```

### 示例 2：增量更新

输入：

```text
Update the project index after regenerating CodeMap.
```

预期行为：

```text
只把现有 `.agent/PROJECT_INDEX.md`、`.agent/NAVIGATION.md` 和
`.agent/CHANGE_GUIDE.md` 当作历史草稿。保留有用的用户补充说明，但在重写前
用 `.codemap/` 和源码文件验证结构与路径。
```

### 示例 3：概念缺口

输入：

```text
Generate indexes, but the module names do not explain the business features.
```

预期行为：

```text
使用 `/grilling` 和 `/domain-modeling` 澄清业务术语和功能边界。如果不可用，
直接提出聚焦问题。将稳定术语记录到 `.agent/GLOSSARY.md`，或将决策记录到
`.agent/adr/*.md`，然后生成索引。
```
