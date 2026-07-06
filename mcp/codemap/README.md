# CodeMap

**中文文档** | [English](#english)

CodeMap 是一个面向 AI Agent 的项目知识层。它把 Go / Java 项目分析成可查询的结构化知识：模块、依赖、HTTP 路由、调用流、调用图、影响范围、功能地图和代码导航提示，并通过 MCP 暴露给 Codex、OpenCode、Cursor、Claude Desktop 等支持 MCP 的客户端。

它的目标不是替代 IDE，也不是替代全文搜索，而是让模型在改代码前先获得稳定、低噪声、可验证的项目结构上下文。

## 目录

- [核心能力](#核心能力)
- [适用场景](#适用场景)
- [架构](#架构)
- [安装](#安装)
- [快速开始](#快速开始)
- [单项目模式](#单项目模式)
- [Workspace 模式](#workspace-模式)
- [MCP 配置](#mcp-配置)
- [MCP 工具](#mcp-工具)
- [MCP 资源](#mcp-资源)
- [使用示例](#使用示例)
- [和其他方案对比](#和其他方案对比)
- [语言支持](#语言支持)
- [常见问题](#常见问题)
- [English](#english)

## 核心能力

| 能力 | 说明 |
|------|------|
| 项目索引 | 扫描源码并生成 `.codemap/codemap.db`，SQLite 是唯一事实来源 |
| 模块分析 | 识别模块路径、导出类型、导出函数、导出方法、关键接口 |
| 依赖分析 | 分析项目内部模块依赖 |
| 路由分析 | 识别 Go HTTP 路由，包括常见 router 和链式路由写法 |
| 包装 handler 识别 | 支持 `To(Filter(registerApi.Login))` 这类包装调用，尽量解析到真实 handler |
| 调用流分析 | 生成跨模块调用 / 数据流线索 |
| 调用图 | 记录函数级调用边，用于 caller / callee 查询 |
| 影响分析 | 根据函数名查找潜在调用方 |
| 功能地图 | 从模块、路由、flow 中推导业务 feature |
| 导航提示 | 给出某个 feature 的入口文件、相关模块、相关 flow、风险区域 |
| Workspace 查询 | 在 `/projects` 这类多子仓目录中，按 `auth`、`login` 等子项目路由查询 |
| Markdown 导出 | 生成 `.codemap/INDEX.md`、模块文档、架构文档、路由文档、flow 文档、调用图文档 |
| MCP 服务 | 通过 stdio MCP 暴露工具和资源，供 AI Agent 调用 |

## 适用场景

CodeMap 适合以下工作流：

- **让模型快速理解项目结构**：先问模块、依赖、路由、功能入口，再打开源码。
- **改需求前做导航**：用 `find_change_points` 找候选模块、候选文件、相关路由和风险。
- **排查接口逻辑**：搜索 `/api/v1/auth/login` 这类路由，定位 handler 和所属模块。
- **评估影响面**：改某个函数或模块前，用 `impact_analysis` / `call_graph` 看调用关系。
- **多服务 workspace**：在 `/projects` 下启动 OpenCode，但查询时明确让模型看 `auth`、`login` 等子服务。
- **团队共享项目地图**：提交或分发 `.codemap/` 中的 Markdown 文档，帮助 onboarding。

它不适合做这些事：

- 替代 AST / LSP 级别的精确重构。
- 替代测试、编译、类型检查。
- 在没有重新索引的情况下保证实时反映最新源码。
- 自动证明跨服务调用链。Workspace 模式能帮模型切换子项目，但跨服务 HTTP client 到目标 route 的自动硬链接还没有实现。

## 架构

```text
source code
  -> analyzer layer (Go AST / Java scanner)
  -> project model (Project, Module, Route, Flow, CallEdge)
  -> SQLite database (.codemap/codemap.db)
  -> Markdown generator + MCP server
  -> AI Agent
```

目录结构：

```text
codemap/
  cmd/codemap/                 CLI entrypoint
  internal/analyzer/            language analyzers
  internal/analyzer/golang/     Go analyzer
  internal/analyzer/java/       Java analyzer
  internal/cognitive/           feature map, navigation hints, change points
  internal/generator/markdown/  Markdown export
  internal/mcp/                 MCP tools and resources
  internal/model/               project data model
  internal/storage/             storage interface
  internal/storage/sqlite/      SQLite implementation
  internal/workspace/           multi-project workspace registry
```

## 安装

### 前置条件

- Go 1.23 或更高版本。
- 可选：`sqlite3` CLI，用于直接检查 `.codemap/codemap.db`。
- 一个支持 MCP stdio transport 的客户端，例如 OpenCode、Codex、Cursor、Claude Desktop、VS Code MCP extension。

### 从源码构建

```bash
git clone https://github.com/disturb-yy/codemap.git
cd codemap
go build -o codemap ./cmd/codemap/
```

放到 PATH：

```bash
sudo mv codemap /usr/local/bin/
```

或只给当前用户使用：

```bash
mkdir -p ~/.local/bin
mv codemap ~/.local/bin/
```

确认安装：

```bash
codemap -h
```

### 本仓库本地构建

如果你就在这个仓库里开发：

```bash
go build -o bin/codemap ./cmd/codemap/
./bin/codemap -h
```

### 让 OpenCode / Codex 代理直接安装

如果你已经在 OpenCode、Codex 或其他 coding agent 里打开了本仓库，可以直接把下面的指令发给 agent，让它代为构建、复制二进制、更新 MCP 配置。

OpenCode 单项目安装：

```text
请在当前 codemap 仓库中执行：
1. go build -o /tmp/codemap ./cmd/codemap/
2. mkdir -p ~/tool/ai-coding/mcp
3. cp /tmp/codemap ~/tool/ai-coding/mcp/codemap
4. chmod 755 ~/tool/ai-coding/mcp/codemap
5. 更新 OpenCode 配置，让 codemap MCP 使用：
   command: ["~/tool/ai-coding/mcp/codemap", "--serve"]
不要添加 -project，让它使用 OpenCode 启动时的当前目录。
```

OpenCode workspace 安装：

```text
请在当前 codemap 仓库中执行：
1. go build -o /tmp/codemap ./cmd/codemap/
2. mkdir -p ~/tool/ai-coding/mcp
3. cp /tmp/codemap ~/tool/ai-coding/mcp/codemap
4. chmod 755 ~/tool/ai-coding/mcp/codemap
5. 更新 OpenCode 配置，让 codemap MCP 使用：
   command: ["~/tool/ai-coding/mcp/codemap", "--serve", "--workspace"]
不要添加 -project，让 workspace 根使用 OpenCode 启动时的当前目录。
```

如果你的 OpenCode 使用全局配置，通常修改：

```text
~/.config/opencode/opencode.json
```

配置形态通常是：

```json
{
  "mcp": {
    "codemap": {
      "command": [
        "/home/you/tool/ai-coding/mcp/codemap",
        "--serve",
        "--workspace"
      ],
      "enabled": true,
      "type": "local"
    }
  }
}
```

Codex 安装：

```text
请在当前 codemap 仓库中执行：
1. go build -o /tmp/codemap ./cmd/codemap/
2. mkdir -p ~/tool/ai-coding/mcp
3. cp /tmp/codemap ~/tool/ai-coding/mcp/codemap
4. chmod 755 ~/tool/ai-coding/mcp/codemap
5. 更新 ~/.codex/config.toml，添加或更新：
   [mcp_servers.codemap]
   command = "/home/you/tool/ai-coding/mcp/codemap"
   args = ["--serve"]
```

Codex workspace 模式把最后一行改为：

```toml
args = ["--serve", "--workspace"]
```

安装后需要重启 OpenCode / Codex session。`--serve` 只读取已有索引，首次使用前仍需要先对目标项目运行：

```bash
codemap -project /path/to/project
```

## 快速开始

### 1. 给项目建立索引

```bash
codemap -project /path/to/your-project
```

CodeMap 会自动检测语言：

| 条件 | 语言 |
|------|------|
| 存在 `go.mod` | Go |
| 存在 `pom.xml`、`build.gradle` 或 `settings.gradle` | Java |
| 存在 `src/main/java` | Java |
| 都不存在 | 默认按 Go 处理 |

索引后会生成：

```text
your-project/
  .codemap/
    codemap.db
    INDEX.md
    modules/
    architecture/
    routes/
    flows/
    callgraph/
```

### 2. 启动 MCP 服务

```bash
codemap -project /path/to/your-project --serve
```

MCP 服务只读取已有 `.codemap/codemap.db`，不会自动重新分析源码。源码改动后需要重新运行：

```bash
codemap -project /path/to/your-project
```

### 3. 在模型里提问

示例：

```text
查询这个项目有哪些模块。
搜索 /api/v1/auth/login 路由。
如果我要加短信登录，先用 find_change_points 判断应该看哪些模块和文件。
```

## 单项目模式

单项目模式适合一个目录就是一个服务 / 仓库的场景：

```bash
codemap -project /projects/auth
codemap -project /projects/auth --serve
```

OpenCode 配置：

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/projects/auth", "--serve"]
    }
  }
}
```

如果你在项目根目录启动 OpenCode，也可以省略 `-project`：

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["--serve"]
    }
  }
}
```

`-project` 默认是 `.`。CodeMap 会把普通单项目路径向上归一到最近的项目根，例如 `go.mod`、Java build 文件、`.git` 或 `.codemap/codemap.db` 所在目录。

## Workspace 模式

Workspace 模式适合 `/projects` 这种目录：

```text
/projects/
  auth/
  login/
  billing/
  other-service/
```

先分别给子项目建立索引：

```bash
codemap -project /projects/auth
codemap -project /projects/login
codemap -project /projects/billing
```

然后在 workspace 根启动 MCP：

```bash
codemap -project /projects --serve --workspace
```

OpenCode 配置：

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/projects", "--serve", "--workspace"]
    }
  }
}
```

Workspace 模式会扫描 `/projects` 的一级子目录，识别以下 marker：

```text
.codemap/codemap.db
go.mod
pom.xml
build.gradle
settings.gradle
.git
```

查询工具会增加可选 `project` 参数。例如：

```json
{"project": "auth", "query": "/api/v1/auth/login"}
```

如果 query 文本里包含子项目名，例如 `auth /login`，CodeMap 会尝试自动推断 `project=auth`。如果有多个子项目且无法推断，会返回需要指定 project 的错误，并提示可选项目列表。

注意：

- `/projects` 只作为 workspace 入口，不需要对 `/projects` 本身运行索引。
- 每个子项目仍然需要自己的 `.codemap/codemap.db`。
- 代码变动后，哪个子项目变了就重新索引哪个子项目。
- Workspace 模式负责子项目选择，不自动建立跨服务调用链。

## MCP 配置

### OpenCode

单项目：

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/absolute/path/to/project", "--serve"]
    }
  }
}
```

Workspace：

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/projects", "--serve", "--workspace"]
    }
  }
}
```

### Codex / Codex++

```toml
[mcp_servers.codemap]
command = "/usr/local/bin/codemap"
args = ["-project", "/absolute/path/to/project", "--serve"]
```

### Cursor

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "${workspaceFolder}", "--serve"]
    }
  }
}
```

### Claude Desktop

macOS:

```text
~/Library/Application Support/Claude/claude_desktop_config.json
```

Linux:

```text
~/.config/Claude/claude_desktop_config.json
```

配置：

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/absolute/path/to/project", "--serve"]
    }
  }
}
```

### VS Code MCP extension

```json
{
  "servers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "${workspaceFolder}", "--serve"]
    }
  }
}
```

## MCP 工具

单项目模式提供以下工具：

| 工具 | 参数 | 说明 |
|------|------|------|
| `get_project_info` | `{}` | 项目名、根路径、模块数 |
| `list_modules` | `{}` | 列出所有模块及其依赖、导出类型、导出函数、导出方法、接口 |
| `search_module` | `{"module":"auth"}` | 按模块名搜索 |
| `related_modules` | `{"module":"auth"}` | 查询某模块依赖谁、被谁依赖 |
| `search_route` | `{"query":"/login"}` | 搜索 HTTP 路由、handler、模块 |
| `search_flow` | `{"query":"login"}` | 搜索调用流 / 数据流 |
| `call_graph` | `{"module":"internal/auth"}` | 查询某模块调用了哪些函数 |
| `impact_analysis` | `{"function":"Login"}` | 查询哪些函数调用了目标函数 |
| `get_feature_map` | `{}` | 获取业务功能地图 |
| `get_navigation_hints` | `{}` | 获取功能入口文件、相关模块、风险区域 |
| `find_change_points` | `{"requirement":"Add SMS login","top_k":5}` | 根据需求推断候选模块、候选文件、相关路由、相关 flow、风险和下一步动作 |

Workspace 模式额外提供：

| 工具 | 参数 | 说明 |
|------|------|------|
| `list_projects` | `{}` | 列出 workspace 子项目 |

Workspace 模式下，上面的项目查询工具都支持可选 `project`：

```json
{"project":"auth","query":"/login"}
```

或：

```json
{"project":"login","requirement":"Add password reset flow"}
```

## MCP 资源

单项目模式还注册以下资源：

| URI | 内容 |
|-----|------|
| `codemap://modules` | 所有模块 JSON |
| `codemap://module/{name}` | 单模块 JSON |
| `codemap://modules-doc/{name}` | `.codemap/modules/{name}.md` |
| `codemap://architecture/overview` | 架构概览 Markdown |
| `codemap://architecture/dependencies` | 依赖图 Markdown |
| `codemap://routes/{name}` | 路由文档 |
| `codemap://flows/{name}` | flow 文档 |
| `codemap://callgraph/{name}` | 调用图文档 |

Workspace 模式当前以 tools 为主，不注册跨项目资源模板。

## 使用示例

### 架构理解

```text
读取项目模块列表和架构概览，给我一个 5 分钟 onboarding。
```

### 路由定位

```text
搜索 auth 服务里的 /api/v1/auth/login 路由，告诉我 handler 和相关模块。
```

Workspace 模式下模型应调用：

```json
{"project":"auth","query":"/api/v1/auth/login"}
```

### 需求变更规划

```text
我要在 login 增加密码重置流程。先用 find_change_points 判断要看哪些模块和文件。
```

### 影响分析

```text
如果修改 Login 函数，可能影响哪些调用方？
```

### 多服务查询

```text
查询 login 服务里调用 auth 的登录接口附近代码，然后再查 auth 对应接口入口。
```

当前 CodeMap 能稳定帮助模型切换 `login` 和 `auth` 两个子项目查询；跨服务调用链仍需要模型根据 client 代码、URL、配置 key 自己判断。

## 手动测试 MCP

不接 IDE 也可以测试 MCP 工具：

```bash
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"cli","version":"1"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_modules","arguments":{}}}\n' \
  | codemap -project . --serve 2>/dev/null \
  | tail -1 | python3 -m json.tool
```

Workspace 示例：

```bash
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"cli","version":"1"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_route","arguments":{"project":"auth","query":"/login"}}}\n' \
  | codemap -project /projects --serve --workspace 2>/dev/null \
  | tail -1 | python3 -m json.tool
```

## 直接查询 SQLite

```bash
sqlite3 .codemap/codemap.db "SELECT name, path FROM module ORDER BY path;"
sqlite3 .codemap/codemap.db "SELECT method, path, handler, module FROM route ORDER BY path;"
sqlite3 .codemap/codemap.db "SELECT caller_module, caller_func, callee_module, callee_func FROM call_edge LIMIT 20;"
```

## 和其他方案对比

| 方案 | 优点 | 局限 | CodeMap 的定位 |
|------|------|------|----------------|
| `grep` / `rg` | 快、精确、无索引成本 | 只返回文本命中，不理解模块、依赖、路由、调用关系 | CodeMap 给模型结构化入口，`rg` 仍适合最终源码验证 |
| IDE / LSP | 类型感知强，跳转精确 | 面向人类交互，不一定适合 MCP 自动查询和跨文档汇总 | CodeMap 把常用结构查询封装成工具 |
| 传统调用图工具 | 调用关系更专注 | 往往缺少路由、业务 feature、导航提示 | CodeMap 同时覆盖 route / module / flow / feature |
| 向量 RAG | 适合语义搜索和长文本召回 | 结果可能不稳定，难表达依赖和调用边 | CodeMap 是结构化事实层，可和 RAG 互补 |
| 手写项目文档 | 语义清晰，适合 onboarding | 容易过期，需要人工维护 | CodeMap 从源码再生成结构化文档 |
| Context / memory 系统 | 记录会话历史和决策 | 不等于当前代码结构事实 | CodeMap 提供当前项目事实，memory 提供历史上下文 |
| OpenAPI / Swagger | API 契约清楚 | 只覆盖暴露接口，不覆盖模块内部依赖和调用 | CodeMap 能从源码侧补足实现结构 |

## 语言支持

| 语言 | 状态 | 能力 |
|------|------|------|
| Go | 完整 | 模块、依赖、导出符号、HTTP 路由、flow、调用图、影响分析、功能地图、导航提示 |
| Java | 基础 | 模块、依赖、导出类型、方法、接口 |
| Python | 计划中 | 暂未实现 |

Go / Java 分析差异：

| 维度 | Go | Java |
|------|----|------|
| 模块识别 | 基于目录和 `go.mod` import path | 基于 package / src 目录 |
| 依赖解析 | Go import path 匹配 module path | 文件系统目录验证 |
| 路由提取 | 支持 | 暂不支持 |
| flow | 支持 | 暂不支持 |
| call graph | 支持 | 暂不支持 |
| 类型提取 | `ast.TypeSpec` | `public class` / `public interface` / `public enum` |
| 方法提取 | `Receiver.Method` | `ClassName.method` |

## 常见问题

### 需要每次手动更新索引吗？

是。`--serve` 只读取已有 DB，不会自动重新分析源码。代码改动后需要重新运行：

```bash
codemap -project /path/to/project
```

### 没有 `.codemap/codemap.db` 会怎样？

单项目模式下，`--serve` 会启动失败并提示先运行索引。

Workspace 模式下，`list_projects` 可以看到子项目，但未索引项目会显示 `indexed:false`；查询该项目会返回需要先运行：

```bash
codemap -project /projects/auth
```

### OpenCode 开多个 web session 会怎样？

stdio MCP 通常是一个 client session 启一个进程。当前 CodeMap 会用 `.codemap/server.lock` 避免同一项目重复运行有效服务，但后启动的 session 可能无法共享第一个 stdio 进程。

Workspace 模式解决的是“在 `/projects` 启动时按子项目查询”的问题，不是 daemon/proxy 共享问题。未来可以做 per-project daemon + stdio proxy。

### 为什么不自动打通跨服务调用链？

跨服务调用可能经过配置、网关、服务发现、SDK、反向代理、OpenAPI client 或 RPC。错误硬链接比没有链接更危险。当前推荐让模型先查来源服务 client 代码，再切到目标服务 route 查询。后续可以做候选式 `search_workspace` 或 `trace_external_call`。

### `.codemap/` 应该提交到 Git 吗？

默认会写入 `.gitignore`。如果团队希望共享索引文档，可以只提交 Markdown，也可以提交 DB；但 DB 可能较大，且需要和源码版本保持一致。

### 出现 `another codemap server is already running` 怎么办？

说明 `.codemap/server.lock` 指向的进程仍被认为存活。确认没有正在使用后可以删除：

```bash
rm .codemap/server.lock
```

### 出现 `unsupported call` 怎么办？

部分 MCP 客户端对 `tools/call` 支持不完整。可以优先尝试 resources/read，或用手动 MCP 测试命令确认服务本身是否正常。

## 设计文档

- [feature/DESIGN.md](./feature/DESIGN.md)
- [feature/DESIGNV2.md](./feature/DESIGNV2.md)
- [feature/DESIGNV3.md](./feature/DESIGNV3.md)

---

<a id="english"></a>

# CodeMap

[中文文档](#codemap) | **English**

CodeMap is a project knowledge layer for AI agents. It analyzes Go and Java projects into structured knowledge: modules, dependencies, HTTP routes, call flows, call graphs, impact relationships, feature maps, and navigation hints. It then exposes that knowledge through MCP for clients such as Codex, OpenCode, Cursor, Claude Desktop, and other MCP-compatible tools.

CodeMap is not an IDE replacement and not a plain text search replacement. Its job is to give the model stable, low-noise, verifiable project structure before it edits code.

## Table of Contents

- [Core Capabilities](#core-capabilities)
- [Use Cases](#use-cases)
- [Architecture](#architecture-1)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Single Project Mode](#single-project-mode)
- [Workspace Mode](#workspace-mode)
- [MCP Configuration](#mcp-configuration)
- [MCP Tools](#mcp-tools)
- [MCP Resources](#mcp-resources)
- [Examples](#examples)
- [Comparison](#comparison)
- [Language Support](#language-support)
- [FAQ](#faq)

## Core Capabilities

| Capability | Description |
|------------|-------------|
| Project indexing | Scans source code and writes `.codemap/codemap.db`; SQLite is the source of truth |
| Module analysis | Extracts module paths, exported types, exported functions, exported methods, and key interfaces |
| Dependency analysis | Captures internal module dependencies |
| Route analysis | Detects Go HTTP routes, including common routers and chained route declarations |
| Wrapped handler detection | Handles patterns such as `To(Filter(registerApi.Login))` and tries to resolve the real handler |
| Flow analysis | Produces cross-module call/data flow hints |
| Call graph | Stores function-level call edges for caller/callee queries |
| Impact analysis | Finds potential callers for a target function |
| Feature map | Infers business features from modules, routes, and flows |
| Navigation hints | Suggests entry files, related modules, related flows, and risk areas |
| Workspace routing | In a multi-repo directory such as `/projects`, routes queries to child projects like `auth` or `login` |
| Markdown export | Generates `.codemap/INDEX.md`, module docs, architecture docs, route docs, flow docs, and call graph docs |
| MCP server | Exposes tools and resources over stdio MCP |

## Use Cases

CodeMap is useful when you want an AI agent to:

- Understand the module and dependency structure before opening files.
- Plan a change with `find_change_points`.
- Locate a route such as `/api/v1/auth/login` and identify its handler.
- Estimate the blast radius of changing a function or module.
- Work inside a multi-service workspace and switch between child projects intentionally.
- Generate project maps that help onboarding and review.

CodeMap is not meant to:

- Replace type checking, tests, or compilation.
- Replace exact IDE/LSP refactoring.
- Reflect source changes without re-indexing.
- Prove cross-service call chains automatically.

## Architecture

```text
source code
  -> analyzer layer (Go AST / Java scanner)
  -> project model (Project, Module, Route, Flow, CallEdge)
  -> SQLite database (.codemap/codemap.db)
  -> Markdown generator + MCP server
  -> AI Agent
```

Repository layout:

```text
codemap/
  cmd/codemap/                 CLI entrypoint
  internal/analyzer/            language analyzers
  internal/analyzer/golang/     Go analyzer
  internal/analyzer/java/       Java analyzer
  internal/cognitive/           feature map, navigation hints, change points
  internal/generator/markdown/  Markdown export
  internal/mcp/                 MCP tools and resources
  internal/model/               project data model
  internal/storage/             storage interface
  internal/storage/sqlite/      SQLite implementation
  internal/workspace/           multi-project workspace registry
```

## Installation

### Requirements

- Go 1.23 or newer.
- Optional: the `sqlite3` CLI for inspecting `.codemap/codemap.db`.
- An MCP stdio client such as OpenCode, Codex, Cursor, Claude Desktop, or the VS Code MCP extension.

### Build from Source

```bash
git clone https://github.com/disturb-yy/codemap.git
cd codemap
go build -o codemap ./cmd/codemap/
```

Install globally:

```bash
sudo mv codemap /usr/local/bin/
```

Or install for the current user:

```bash
mkdir -p ~/.local/bin
mv codemap ~/.local/bin/
```

Verify:

```bash
codemap -h
```

### Agent-Assisted Install for OpenCode / Codex

If this repository is already open in OpenCode, Codex, or another coding agent, you can ask the agent to build CodeMap, copy the binary, and update MCP configuration for you.

OpenCode single-project install:

```text
In the current codemap repository:
1. Run: go build -o /tmp/codemap ./cmd/codemap/
2. Run: mkdir -p ~/tool/ai-coding/mcp
3. Run: cp /tmp/codemap ~/tool/ai-coding/mcp/codemap
4. Run: chmod 755 ~/tool/ai-coding/mcp/codemap
5. Update OpenCode so the codemap MCP command is:
   ["~/tool/ai-coding/mcp/codemap", "--serve"]
Do not add -project; let CodeMap use the directory where OpenCode starts.
```

OpenCode workspace install:

```text
In the current codemap repository:
1. Run: go build -o /tmp/codemap ./cmd/codemap/
2. Run: mkdir -p ~/tool/ai-coding/mcp
3. Run: cp /tmp/codemap ~/tool/ai-coding/mcp/codemap
4. Run: chmod 755 ~/tool/ai-coding/mcp/codemap
5. Update OpenCode so the codemap MCP command is:
   ["~/tool/ai-coding/mcp/codemap", "--serve", "--workspace"]
Do not add -project; let the workspace root be the directory where OpenCode starts.
```

For global OpenCode config, the file is commonly:

```text
~/.config/opencode/opencode.json
```

Typical shape:

```json
{
  "mcp": {
    "codemap": {
      "command": [
        "/home/you/tool/ai-coding/mcp/codemap",
        "--serve",
        "--workspace"
      ],
      "enabled": true,
      "type": "local"
    }
  }
}
```

Codex install:

```text
In the current codemap repository:
1. Run: go build -o /tmp/codemap ./cmd/codemap/
2. Run: mkdir -p ~/tool/ai-coding/mcp
3. Run: cp /tmp/codemap ~/tool/ai-coding/mcp/codemap
4. Run: chmod 755 ~/tool/ai-coding/mcp/codemap
5. Update ~/.codex/config.toml with:
   [mcp_servers.codemap]
   command = "/home/you/tool/ai-coding/mcp/codemap"
   args = ["--serve"]
```

For Codex workspace mode, use:

```toml
args = ["--serve", "--workspace"]
```

Restart the OpenCode / Codex session after changing MCP configuration. `--serve` reads an existing index; before first use, index the target project:

```bash
codemap -project /path/to/project
```

## Quick Start

### 1. Index a Project

```bash
codemap -project /path/to/your-project
```

Language detection:

| Condition | Language |
|-----------|----------|
| `go.mod` exists | Go |
| `pom.xml`, `build.gradle`, or `settings.gradle` exists | Java |
| `src/main/java` exists | Java |
| None of the above | Go by default |

Generated output:

```text
your-project/
  .codemap/
    codemap.db
    INDEX.md
    modules/
    architecture/
    routes/
    flows/
    callgraph/
```

### 2. Start the MCP Server

```bash
codemap -project /path/to/your-project --serve
```

The MCP server reads the existing database. It does not automatically re-index source code. After code changes, run:

```bash
codemap -project /path/to/your-project
```

## Single Project Mode

Use this mode when one directory is one project or service:

```bash
codemap -project /projects/auth
codemap -project /projects/auth --serve
```

OpenCode:

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/projects/auth", "--serve"]
    }
  }
}
```

If OpenCode starts in the project root, you may omit `-project`:

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["--serve"]
    }
  }
}
```

`-project` defaults to `.`. In single-project mode, CodeMap normalizes subdirectories back to the nearest project root marker such as `go.mod`, a Java build file, `.git`, or `.codemap/codemap.db`.

## Workspace Mode

Use workspace mode for a directory like this:

```text
/projects/
  auth/
  login/
  billing/
  other-service/
```

Index each child project first:

```bash
codemap -project /projects/auth
codemap -project /projects/login
codemap -project /projects/billing
```

Start the workspace MCP server:

```bash
codemap -project /projects --serve --workspace
```

OpenCode:

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/projects", "--serve", "--workspace"]
    }
  }
}
```

Workspace mode scans first-level child directories and recognizes:

```text
.codemap/codemap.db
go.mod
pom.xml
build.gradle
settings.gradle
.git
```

Workspace tools accept an optional `project` argument:

```json
{"project": "auth", "query": "/api/v1/auth/login"}
```

If the query text contains a child project name, CodeMap tries to infer the project. If multiple projects exist and no project can be inferred, CodeMap returns an error with the available project list.

Important details:

- You do not need to index `/projects` itself.
- Each child project needs its own `.codemap/codemap.db`.
- Re-index only the child project that changed.
- Workspace mode selects projects; it does not automatically prove cross-service call chains.

## MCP Configuration

### OpenCode

Single project:

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/absolute/path/to/project", "--serve"]
    }
  }
}
```

Workspace:

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/projects", "--serve", "--workspace"]
    }
  }
}
```

### Codex / Codex++

```toml
[mcp_servers.codemap]
command = "/usr/local/bin/codemap"
args = ["-project", "/absolute/path/to/project", "--serve"]
```

### Cursor

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "${workspaceFolder}", "--serve"]
    }
  }
}
```

### Claude Desktop

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/absolute/path/to/project", "--serve"]
    }
  }
}
```

### VS Code MCP Extension

```json
{
  "servers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "${workspaceFolder}", "--serve"]
    }
  }
}
```

## MCP Tools

Single-project tools:

| Tool | Arguments | Description |
|------|-----------|-------------|
| `get_project_info` | `{}` | Project name, root path, module count |
| `list_modules` | `{}` | List all modules and exported symbols |
| `search_module` | `{"module":"auth"}` | Search modules by name |
| `related_modules` | `{"module":"auth"}` | Find dependencies and dependents |
| `search_route` | `{"query":"/login"}` | Search HTTP routes, handlers, or modules |
| `search_flow` | `{"query":"login"}` | Search call/data flows |
| `call_graph` | `{"module":"internal/auth"}` | List callees for a module |
| `impact_analysis` | `{"function":"Login"}` | Find callers of a function |
| `get_feature_map` | `{}` | Get inferred business features |
| `get_navigation_hints` | `{}` | Get entry files, related modules, and risks |
| `find_change_points` | `{"requirement":"Add SMS login","top_k":5}` | Plan a change by returning candidate modules, files, routes, flows, risks, and next actions |

Workspace mode adds:

| Tool | Arguments | Description |
|------|-----------|-------------|
| `list_projects` | `{}` | List child projects in the workspace |

In workspace mode, project query tools also accept `project`:

```json
{"project":"auth","query":"/login"}
```

## MCP Resources

Single-project mode registers:

| URI | Content |
|-----|---------|
| `codemap://modules` | JSON list of all modules |
| `codemap://module/{name}` | JSON detail for one module |
| `codemap://modules-doc/{name}` | Markdown module doc |
| `codemap://architecture/overview` | Architecture overview |
| `codemap://architecture/dependencies` | Dependency graph |
| `codemap://routes/{name}` | Route doc |
| `codemap://flows/{name}` | Flow doc |
| `codemap://callgraph/{name}` | Call graph doc |

Workspace mode currently focuses on tools and does not register cross-project resource templates.

## Examples

Architecture onboarding:

```text
Read the module list and architecture overview, then give me a five-minute tour.
```

Route lookup:

```text
Search the auth service for /api/v1/auth/login and tell me the handler and related module.
```

Change planning:

```text
For login, use find_change_points for "Add password reset flow".
```

Impact analysis:

```text
If I modify Login, which callers may be affected?
```

Multi-service workflow:

```text
Inspect where login calls the auth login API, then inspect the corresponding auth route.
```

CodeMap can help the agent query both projects accurately. The agent still needs to inspect client code, URLs, and config keys to reason about the cross-service call.

## Manual MCP Testing

```bash
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"cli","version":"1"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_modules","arguments":{}}}\n' \
  | codemap -project . --serve 2>/dev/null \
  | tail -1 | python3 -m json.tool
```

Workspace example:

```bash
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"cli","version":"1"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_route","arguments":{"project":"auth","query":"/login"}}}\n' \
  | codemap -project /projects --serve --workspace 2>/dev/null \
  | tail -1 | python3 -m json.tool
```

## Direct SQLite Queries

```bash
sqlite3 .codemap/codemap.db "SELECT name, path FROM module ORDER BY path;"
sqlite3 .codemap/codemap.db "SELECT method, path, handler, module FROM route ORDER BY path;"
sqlite3 .codemap/codemap.db "SELECT caller_module, caller_func, callee_module, callee_func FROM call_edge LIMIT 20;"
```

## Comparison

| Alternative | Strengths | Limitations | CodeMap's Role |
|-------------|-----------|-------------|----------------|
| `grep` / `rg` | Fast and exact text search | No module/dependency/route/call structure | Provides structured entry points; use `rg` for final source verification |
| IDE / LSP | Strong type-aware navigation | Human-oriented and not always easy to expose through MCP | Packages common structure queries as agent tools |
| Call graph tools | Good at call relationships | Usually do not cover routes, features, or navigation hints | Combines routes, modules, flows, features, and call graph |
| Vector RAG | Good semantic retrieval | Can be noisy and weak at graph relationships | Provides structured facts that complement RAG |
| Hand-written docs | Clear when maintained | Drift over time | Regenerates project maps from source |
| Context / memory systems | Preserve conversation history and decisions | Not a current source structure fact layer | CodeMap provides current project facts |
| OpenAPI / Swagger | Clear API contracts | Does not explain internal implementation structure | Adds implementation-side module and handler context |

## Language Support

| Language | Status | Capabilities |
|----------|--------|--------------|
| Go | Full | Modules, dependencies, exported symbols, HTTP routes, flows, call graph, impact analysis, feature map, navigation hints |
| Java | Basic | Modules, dependencies, exported types, methods, interfaces |
| Python | Planned | Not implemented |

Go / Java differences:

| Dimension | Go | Java |
|-----------|----|------|
| Module detection | Directory and `go.mod` import path | Package / source tree |
| Dependency resolution | Import path matching | Filesystem directory verification |
| Route extraction | Supported | Not supported yet |
| Flow extraction | Supported | Not supported yet |
| Call graph | Supported | Not supported yet |
| Type extraction | `ast.TypeSpec` | `public class` / `public interface` / `public enum` |
| Method extraction | `Receiver.Method` | `ClassName.method` |

## FAQ

### Do I need to update the index manually?

Yes. `--serve` reads the existing DB and does not re-analyze source code. Re-run:

```bash
codemap -project /path/to/project
```

### What happens if `.codemap/codemap.db` does not exist?

Single-project `--serve` fails and asks you to run indexing first.

In workspace mode, `list_projects` can still show child projects, but unindexed projects show `indexed:false`. Querying an unindexed project returns an error telling you to run:

```bash
codemap -project /projects/auth
```

### What happens with multiple OpenCode web sessions?

stdio MCP commonly starts one process per client session. CodeMap uses `.codemap/server.lock` to avoid multiple active servers for the same project root, but later sessions may not share the first stdio process.

Workspace mode solves project selection inside a multi-repo root. It is not yet a daemon/proxy sharing layer.

### Why not automatically link cross-service calls?

Cross-service calls may go through config, gateways, service discovery, SDKs, reverse proxies, generated OpenAPI clients, or RPC. A wrong hard link is worse than no link. Today, the recommended flow is: inspect the source service client code, then query the target service route. Future work may add candidate-based workspace search or external call tracing.

### Should `.codemap/` be committed?

By default CodeMap adds `.codemap/` to `.gitignore`. Teams may choose to commit generated Markdown or the DB, but the artifacts should be kept in sync with source revisions.

### How do I fix `another codemap server is already running`?

After confirming no CodeMap server is active:

```bash
rm .codemap/server.lock
```

## Design Docs

- [feature/DESIGN.md](./feature/DESIGN.md)
- [feature/DESIGNV2.md](./feature/DESIGNV2.md)
- [feature/DESIGNV3.md](./feature/DESIGNV3.md)
