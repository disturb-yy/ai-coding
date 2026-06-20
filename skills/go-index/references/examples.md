# Few-Shot 示例

起草文档前参考以下示例，把握输出深度和格式。

---

## 示例 1：典型 HTTP 服务项目（正常情况）

**输入项目结构**：
```
myapp/
├── go.mod                          # module github.com/acme/myapp, Go 1.22
├── cmd/
│   └── server/
│       └── main.go                 # HTTP 服务入口
├── internal/
│   ├── handler/
│   │   ├── user.go                 # POST /users, GET /users/:id
│   │   └── order.go                # POST /orders, GET /orders/:id
│   ├── service/
│   │   ├── user.go                 # 用户业务逻辑
│   │   └── order.go                # 订单业务逻辑
│   └── repo/
│       ├── user.go                 # users 表 CRUD
│       └── order.go                # orders 表 CRUD
├── pkg/
│   └── middleware/
│       └── auth.go                 # JWT 认证中间件
└── go.sum
```

**关键 go.mod 依赖**：`github.com/gin-gonic/gin`, `gorm.io/gorm`, `github.com/golang-jwt/jwt/v5`

**预期输出片段**：

`.agent/index.md` 入口点：
```markdown
| 入口 | 路径 | 描述 |
|------|------|------|
| HTTP Server | cmd/server/main.go | Gin HTTP 服务，监听 :8080，提供用户和订单 REST API |
```

`docs/ARCHITECTURE.md` 分层：
```markdown
## Handler/Controller 层
- 位置：internal/handler/
- 职责：HTTP 请求解析、参数校验、响应格式化
- 调用：internal/service/

## Service/Business 层
- 位置：internal/service/
- 职责：用户注册/登录逻辑、订单创建/查询编排
- 调用：internal/repo/

## Repository/Data 层
- 位置：internal/repo/
- 职责：数据库 CRUD 操作
- 依赖：gorm.io/gorm（MySQL 驱动）
```

`docs/dataflow.md` 流程示例：
```markdown
## 流程：用户注册

- **入口点**：POST /users → handler/user.go Register()
- **执行路径**：

handler/user.go Register()
↓
service/user.go CreateUser()
↓
repo/user.go Insert()

- **依赖**：MySQL（users 表）
- **输出**：201 Created + 用户 JSON，或 409 Conflict
```

---

## 示例 2：极简 CLI 工具（边界情况）

**输入项目结构**：
```
tinycli/
├── go.mod              # module github.com/jane/tinycli, Go 1.21
├── main.go             # 唯一入口，flag 解析 + 核心逻辑
└── go.sum
```

**预期输出**：该项目仅一个文件，无 internal/pkg 分层。所有四份文档仍然产出，但模块列表和数据流会极短。

`.agent/index.md` 核心模块：
```markdown
| 模块 | 路径 | 职责 |
|------|------|------|
| CLI 入口 | main.go | 命令行参数解析与核心逻辑（单文件项目） |
```

`docs/ARCHITECTURE.md`：
```markdown
# 架构概览
单文件 CLI 工具，无分层架构。所有逻辑集中在 main.go 中。

# 分层设计
未识别到分层结构——项目为单文件。

# 入口点
| CLI 命令 | tinycli | 通过 flag 包解析命令行参数 |
```

`docs/modules.md`：
```markdown
## 模块：CLI 核心

- **路径**：`.`
- **职责**：
  - 解析命令行参数
  - 执行核心逻辑
- **依赖**：
  - 内部：无其他模块
  - 外部：标准库 flag、fmt、os
- **关键文件**：
  - main.go：唯一源文件，包含全部逻辑
- **公开接口**：未识别到导出接口（仅 main 包）
```

`docs/dataflow.md`：
```markdown
## 流程：CLI 执行

- **入口点**：main.go main()
- **执行路径**：

main()
↓
flag.Parse() → 核心逻辑

- **依赖**：无外部依赖
- **输出**：标准输出文本或文件
```

---

## 示例 3：非 Go 项目（错误情况）

**输入项目结构**：
```
random-project/
├── package.json
├── src/
│   └── index.js
└── README.md
```

**预期行为**：
- Observe 阶段检测到 `go.mod` 不存在
- **立即报告错误并退出**，不生成任何文档

```text
错误：未在项目根目录找到 go.mod 文件。这不是一个 Go 项目。
请确认工作目录是否正确，或切换到包含 go.mod 的 Go 项目目录。
```
