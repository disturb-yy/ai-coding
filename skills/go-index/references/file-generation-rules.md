# 文件生成规则

每份生成文档的详细规则。起草输出文件时参照此文档。

---

## .agent/index.md

**作用**：项目导航入口 —— Agent 读取的第一份文档。

**篇幅目标**：≤200 行。

**必填章节**：

```markdown
# 项目概述

一段话描述：项目做什么、服务谁、主要领域。

# 项目类型

- 单体 / 微服务 / 库 / CLI 工具
- 部署方式：二进制、容器、Lambda 等

# 技术栈

列出关键技术及版本（来自 go.mod）：
- Go 版本
- 关键框架和库（HTTP 路由、ORM、消息队列等）
- 数据库 / 缓存 / 外部服务

# 入口点

表格格式：

| 入口 | 路径 | 描述 |

每个入口：cmd 二进制路径、HTTP 路由根路径或 gRPC 服务名。

# 核心模块

表格格式：

| 模块 | 路径 | 职责 |

仅列出顶层模块（5–10 条）。

# 重要文档

相对项目根目录的关键文档列表：
- docs/ARCHITECTURE.md
- docs/modules.md
- docs/dataflow.md
- 任何值得注意的已有项目文档

# 推荐阅读顺序

编号列表：
1. .agent/index.md（本文件）
2. docs/ARCHITECTURE.md
3. docs/modules.md
4. docs/dataflow.md
5. 源代码（从入口点开始）
```

**规则**：
- 所有信息必须从实际项目文件中提取（go.mod、目录结构、源代码）
- 禁止编造模块名称或职责
- 如果某章节为空，标注"未识别到"而非省略

---

## docs/ARCHITECTURE.md

**作用**：系统级架构描述。

**必填章节**：

```markdown
# 架构概览

一段话描述整体架构风格（分层、六边形、CQRS 等）。

# 分层设计

描述每个架构层：

## Handler/Controller 层
- 做什么（请求解析、校验、响应格式化）
- 位置（目录路径）
- 调用什么（服务层）

## Service/Business 层
- 做什么（业务逻辑、编排）
- 位置
- 调用什么（仓库层、外部服务）

## Repository/Data 层
- 做什么（数据访问、持久化）
- 位置
- 依赖什么（数据库驱动、ORM）

## 附加分层（如适用）
- Gateway/Client 层（外部 API 调用）
- MQ/Event 层（消息队列、事件总线）
- Cache 层（Redis、内存缓存）

# 运行时组件

列出运行时基础设施：
- 数据库
- 消息队列
- 缓存
- 调用的外部服务
- 定时任务 / cron

# 依赖规则

描述依赖方向规则（例如"Handler 依赖 Service，Service 依赖 Repository，禁止反向依赖"）。

注明依赖注入方式。

# 入口点

列出所有入口点及其类型：
- HTTP 路由（方法 + 路径模式）
- gRPC 服务
- CLI 命令
- 消息队列消费者
- 定时任务
```

**规则**：
- 描述模式，而非实现细节
- 禁止粘贴源代码
- 禁止列出单个函数签名
- 使用目录路径锚定描述

---

## docs/modules.md

**作用**：模块边界文档。

**对每个模块**，生成：

```markdown
## 模块：<名称>

- **路径**：`<相对项目根目录的路径>`
- **职责**：2–4 条要点，描述该模块负责什么
- **依赖**：
  - 内部：依赖的其他项目模块
  - 外部：第三方包（仅关键项）
- **关键文件**：3–6 个最重要文件，附一行描述
- **公开接口**：
  - 其他模块使用的导出类型/函数
  - 接口契约（仅当定义了接口时）
```

**模块识别规则**：
- 模块是 `internal/`、`pkg/` 或顶层领域目录下的一个目录
- 将小型工具目录归入父模块
- 跳过纯测试目录和 `mocks/` 目录
- 仅包含有业务逻辑的模块；跳过纯配置/常量目录

**规则**：
- 禁止粘贴源代码
- 禁止列出所有函数 —— 仅公开接口
- 每个模块条目控制在 15–30 行
- 如果模块没有公开接口，标注"仅内部使用的模块"

---

## docs/dataflow.md

**作用**：核心业务流程文档。

**对每个流程**，生成：

```markdown
## 流程：<流程名>

- **入口点**：handler 名称 + 文件路径
- **执行路径**：

HandlerName (file.go)
↓
ServiceName.Method()
↓
RepositoryName.Method()

- **依赖**：涉及的数据库、外部 API、消息队列
- **输出**：流程产生的结果（HTTP 响应、数据库行、事件等）
- **错误路径**：关键失败模式（可选，仅在非显而易见时包含）
```

**流程识别规则**：
- 流程是完整的请求生命周期：入口 → 处理 → 输出
- 通过从 handler 追踪 service 调用来识别流程
- 仅包含代表独立业务操作的流程
- 当多个 CRUD 操作共享相同路径时，归入单个流程
- 如初始化/启动流程非平凡，请包含

**规则**：
- 每个流程控制在 5–15 行
- 使用 ↓ 箭头表示调用链
- 禁止粘贴源代码
- 仅展示关键路径；省略日志、指标、中间件
- 至少包含 3–5 个最重要的业务流程

---

## Token 优化

从源代码提取信息时：
- 提取：名称、路径、依赖、调用关系、职责描述
- 跳过：函数体、变量声明、注释、错误处理样板代码
- 对每个文件，仅提取导出符号及其关系
- 使用 `rg` 配合模式匹配高效查找调用关系

## 跨文档一致性

写入文件前验证：
1. `modules.md` 中的每个模块在 `ARCHITECTURE.md` 分层中均有出现
2. `dataflow.md` 中的每个流程入口点在 `index.md` 入口点中均有出现
3. 模块依赖方向与架构依赖规则一致
4. 所有文档间路径引用一致

---

## 完整输出示例

以下为"用户模块"的完整示例，展示理想的深度和篇幅：

### modules.md 完整条目示例

```markdown
## 模块：用户管理

- **路径**：`internal/user/`
- **职责**：
  - 用户注册、登录、信息管理
  - 密码加密与验证（bcrypt）
  - 会话令牌签发与校验
- **依赖**：
  - 内部：internal/cache/（会话缓存）、internal/repo/（数据持久化）
  - 外部：golang.org/x/crypto/bcrypt、github.com/golang-jwt/jwt/v5
- **关键文件**：
  - service.go：注册/登录业务逻辑编排
  - handler.go：HTTP 请求处理与校验
  - model.go：User 结构体定义
- **公开接口**：
  - `func NewService(repo UserRepo, cache SessionCache) *Service`
  - `func (s *Service) Register(ctx context.Context, req RegisterRequest) (*User, error)`
  - `func (s *Service) Login(ctx context.Context, req LoginRequest) (*Session, error)`
  - `UserRepo` 接口：定义数据访问契约
```

### dataflow.md 完整条目示例

```markdown
## 流程：用户注册

- **入口点**：POST /api/v1/users → handler/user.go UserHandler.Register()
- **执行路径**：

UserHandler.Register() (handler/user.go)
↓  校验请求参数、反序列化 JSON
UserService.Register() (service/user.go)
↓  检查用户名唯一性、bcrypt 加密密码
UserRepo.Insert() (repo/user.go)
↓  INSERT INTO users
Cache.PutSession() (cache/redis.go)
↓  生成 JWT、写入 Redis

- **依赖**：MySQL（users 表）、Redis（会话缓存）
- **输出**：201 Created + 用户 JSON（含 ID 和令牌）
- **错误路径**：
  - 用户名已存在 → 409 Conflict
  - 密码格式不符 → 422 Unprocessable Entity
  - 数据库写入失败 → 500 Internal Server Error
```
