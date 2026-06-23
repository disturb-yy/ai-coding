# CodeMap — 面向 AI Agent 的项目知识层

CodeMap 是一个**面向 AI Agent 的项目知识层**，为 Go 和 Java 项目构建结构化认知模型。

它不是：
- ❌ 上下文记忆（那是 Context Mode 的职责）
- ❌ 代码编辑器
- ❌ IDE 插件
- ❌ 调用图可视化工具

---

## 核心定位

```
AI Agent 需求
      ↓
认知层（get_feature_map / get_navigation_hints）
      ↓
导航层（入口文件、相关模块、风险区域）
      ↓
事实层（模块 / 依赖 / 路由 / 数据流 / 调用图）
      ↓
源代码
```

### 事实层（Fact Layer）

CodeMap 扫描源码，自动构建结构化知识图谱并存入 SQLite：

| 能力 | 说明 |
|------|------|
| 模块分析 | 识别项目模块、路径、导出类型/函数/接口 |
| 依赖分析 | 模块间导入依赖关系 |
| 路由分析 | HTTP 路由识别（Go） |
| 数据流分析 | 跨模块调用流 |
| 调用图分析 | 函数级调用关系 |
| 影响分析 | 修改某函数会影响哪些模块 |

### 认知层（Cognitive Layer）

在事实层之上，CodeMap 自动推导业务认知：

| 能力 | MCP 工具 | 说明 |
|------|---------|------|
| 功能地图 | `get_feature_map` | 从数据流/路由/模块自动合成业务功能列表 |
| 导航提示 | `get_navigation_hints` | 每个功能的入口文件、相关模块、风险区域 |

认知层**无需 LLM**，完全基于规则从事实层数据自动生成。

---

## 与其他系统的关系

```
CodeMap = 项目知识（架构、模块、依赖、数据流）
    ↓ 提供结构化上下文
Context Mode = 会话记忆（改了什么、做了什么决策）
    ↓ 提供历史上下文
AI Agent = 结合两者，高效理解和修改代码
```


---

## 架构

```
源代码
    ↓
分析器层（Go AST 分析器 / Java 分析器）
    ↓
知识模型（Project → Module → Route → Flow → CallEdge）
    ↓
SQLite 数据库（唯一真实来源）
    ↓                    ↓
Markdown 导出          MCP 服务（10 个工具 + 6 个资源）
    ↓
AI Agent 消费
```

### 项目结构

```
codemap/
├── cmd/codemap/          # CLI 入口
├── internal/
│   ├── analyzer/         # 分析器接口 + Go/Java 实现
│   ├── model/            # 核心领域模型
│   ├── storage/          # 存储接口 + SQLite 实现
│   ├── generator/        # Markdown 文档生成
│   └── mcp/              # MCP 协议服务
```

---

## 快速开始

```bash
# 构建
go build -o codemap ./cmd/codemap/

# 索引项目
codemap -project /path/to/your-project

# 启动 MCP 服务
codemap -project /path/to/your-project --serve
```

详细安装和配置说明请查看 [INSTALL.md](./INSTALL.md)。

---

## 支持的 MCP 工具

| 工具 | 说明 |
|------|------|
| `get_project_info` | 项目元信息 |
| `list_modules` | 所有模块详情 |
| `search_module` | 按名称搜索模块 |
| `related_modules` | 模块依赖/被依赖关系 |
| `search_route` | HTTP 路由搜索 |
| `search_flow` | 数据/调用流搜索 |
| `call_graph` | 模块调用图 |
| `impact_analysis` | 函数影响分析 |
| `get_feature_map` | 业务功能地图 |
| `get_navigation_hints` | 代码导航提示 |
| `find_change_points` | 决策层：根据需求推断候选模块、文件、路由、流程、风险和下一步动作 |

---

## 语言支持

| 语言 | 状态 | 能力 |
|------|------|------|
| Go | ✅ 完整 | 模块、依赖、路由、数据流、调用图、影响分析、功能地图、导航提示 |
| Java | ✅ 基础 | 模块、依赖、导出类型/方法/接口 |

---

## 设计文档

- [DESIGN.md](./feature/DESIGN.md) — v1 设计规范
- [DESIGNV2.md](./feature/DESIGNV2.md) — v2 认知层设计
- [DESIGNV3.md](./feature/DESIGNV3.md) — v3 决策层 / find_change_points 设计
