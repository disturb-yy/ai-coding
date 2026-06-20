---
name: go-index
description: >-
  Analyze Go repositories and generate agent-consumable cognitive layer documentation
  including project index, architecture, module boundaries, and dataflow.
  Use when setting up a Go project for multi-agent or CodeMap consumption,
  generating .agent/index.md and docs/ for a Go codebase, building structured project
  context for downstream Workflow Agents or MCP tools, or when asked to document the
  architecture or generate project docs for a Go repo.
---

# Go 仓库文档索引器

分析 Go 项目，生成四份结构化文档，为下游 Agent 构建认知层。

## 输出文件

| 文件 | 作用 |
|---|---|
| `.agent/index.md` | 项目导航入口 |
| `docs/ARCHITECTURE.md` | 系统架构描述 |
| `docs/modules.md` | 模块边界与接口 |
| `docs/dataflow.md` | 核心业务流程文档 |

## 工作流（OODA-E）

严格遵循此顺序，不得跳过任何阶段。

### 1. Observe（观察）—— 快速结构扫描


**前置验证**：确认 `go.mod` 存在于项目根目录。若不存在，报告错误并退出——这不是一个 Go 项目。
读取以下文件构建项目概览：

- `README.md`（如果存在）
- `go.mod`（模块名、Go 版本、依赖项）
- 顶层目录列表（`ls` 或 `tree -L 2`）
- 列出 `cmd/`、`internal/`、`pkg/` 下的目录

识别：
- 项目类型（CLI、HTTP 服务、gRPC 服务、库、单体应用）
- 通过 go.mod 依赖判断技术栈
- 通过 `cmd/` 子目录识别入口点
- 通过 `internal/` 和 `pkg/` 识别核心领域目录

**产出**：项目范围的心理模型。此阶段不写任何文件。

### 2. Orient（定位）—— 构建代码地图

深入源码结构：

- 对每个模块目录，列出 `.go` 文件及其导出符号
- 在每个模块目录下使用 `rg "func [A-Z]" --go` 和 `rg "type [A-Z]" --go`
- 追踪项目内 import 路径，绘制内部依赖关系
- 识别架构模式：handler→service→repository 或其他模式

将每个目录按角色分类：
- **Handler/Controller**：HTTP/gRPC 处理器、中间件、路由
- **Service**：业务逻辑、编排
- **Repository/DAO**：数据访问、数据库查询
- **Client/Gateway**：外部 API 调用
- **MQ**：消息队列生产者/消费者
- **Cache**：缓存层

**产出**：内部依赖图和模块角色映射。

### 3. Decide（决策）—— 确定生成方案

基于观察结果：

- 列出哪些目录符合模块条件（跳过测试数据、mock、纯配置目录）
- 通过追踪 handler→service→repository 链路，识别 3–7 个最重要的业务流程
- 判断是否需要合并模块（小型工具目录）
- 决定在 index.md 中引用哪些已有文档

**产出**：待记录的具体模块与流程清单。如有歧义请与用户确认。

### 4. Draft（起草）—— 生成文档内容

在内存中生成全部四份文档。使用 `references/file-generation-rules.md` 获取每份文件的确切格式和章节要求。

起草前先阅读 `references/file-generation-rules.md`。

关键起草规则：
- 仅从源码提取名称、路径、依赖、职责描述
- 禁止粘贴源代码或函数体
- `.agent/index.md` 控制在 200 行以内
- `ARCHITECTURE.md` 聚焦模式而非实现
- `modules.md` 中每个模块条目控制在 15–30 行
- `dataflow.md` 中每个流程控制在 5–15 行
- 数据流中使用 `↓` 箭头表示调用链

### 5. Precheck（预检）—— 写入前验证

落盘前验证：

1. 草案 `modules.md` 中的每个模块在 `ARCHITECTURE.md` 分层中均有对应条目
2. 草案 `dataflow.md` 中的每个流程入口点在 `index.md` 入口点列表中均存在
3. 模块依赖方向与 `ARCHITECTURE.md` 中声明的依赖规则一致
4. 所有路径引用一致且相对于项目根目录
5. 没有任何源代码被意外复制到文档中
6. 没有遗漏模块、入口点或关键流程

如有任何检查未通过，返回 Draft 阶段修复后再继续。

### 6. Act（执行）—— 写入文件

创建目录并写入文件：

```
mkdir -p .agent docs
```

使用 `apply_patch` 或直接文件创建写入：
1. `.agent/index.md`
2. `docs/ARCHITECTURE.md`
3. `docs/modules.md`
4. `docs/dataflow.md`

### 7. Evaluate（评估）—— 验证可用性

写入后，验证 Agent 能否仅凭这些文档理解项目：

- 仅靠 `index.md` 能否定位项目入口点？
- 从 `modules.md` 能否找到特定模块及其职责？
- 从 `dataflow.md` 能否追踪业务流程？
- `ARCHITECTURE.md` 是否准确描述了新代码应添加在哪里？

如有任何答案为否，指出缺失并提供修复。

## 快速开始

当被要求为 Go 项目索引时：

1. 确认项目根目录
2. 执行上述 OODA-E 工作流
3. 汇报生成的文件及简要概述

## 资源

### references/file-generation-rules.md

每份输出文档的详细格式规范。起草任何文档内容前请先阅读。包含全部四份输出文件的必填章节、篇幅目标和字段定义。

### references/examples.md

三个 Few-Shot 示例，覆盖典型项目（正常情况）、小型项目（边界情况）、非 Go 项目（错误情况）。当不确定输出深度或格式时，先阅读示例再起草。
