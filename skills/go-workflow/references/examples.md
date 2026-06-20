# Few-Shot 示例

每个示例追踪一种任务类型的完整 OODA-E 工作流。

---

---

## 示例 0：诊断 / 定位问题（diagnose）

**用户请求：** "定位问题" 或 "go test ./domain/strategy/... 失败"

### Phase 0：初始化



### Observe（本地回退）



### Orient（跳过 — confidence >= 0.8 且 unknowns 为空）



### Decide



### Draft / Precheck / Act（全部跳过）



### Evaluate



**关键点：**
- 启动协议强制执行，确保日志从 Phase 0 开始记录。
- spawn_agent 不可用时回退本地，但必须写 spawn_failed + fallback_local 日志。
- diagnose 类型跳过了 Draft/Precheck/Act，直接输出诊断报告。
- 即使无代码变更，Evaluate 阶段仍输出完整的 log_analysis JSON。

## 示例 1：功能开发（最常见）

**用户请求：** "新用户注册时添加邮件通知。"

### Observe
启动 `doc-index-reader`。
- 读取 `README.md` → 了解到项目是 "user-service" 单体仓库。
- 读取 `ARCHITECTURE.md` → 了解到使用事件驱动模式，有消息代理。
- 读取 `modules.md` → 模块：`api`、`core`、`infra`、`notification`。
- 输出：`confidence: 0.75`、`unknowns: ["notification 模块 API 接口"]`、`next: code-structure-analyzer`

### Orient
启动 `code-structure-analyzer`。
- `query_module("notification")` → 导出：`EmailSender`、`SMSTemplate`。
- `query_entrypoint("core")` → 用户注册流程：`core/service/user.go:Register()`。
- `query_flow("core", "Register")` → `Register → repo.Insert →（无通知步骤）`。
- 输出：`confidence: 0.9`、`impact_scope: {direct: ["core/service/user.go", "notification/email.go"]}`、`next: decide-phase`

### Decide
- `task_type: "feature"`
- `target_modules: ["core", "notification"]`
- `dispatch_plan:` 2 个 Agent：
  1. `module-code-analyzer` scope=`core/service/user.go` — 在 repo.Insert 之后添加通知调用
  2. `module-code-analyzer` scope=`notification/email.go` — 添加 NewUserEmail 模板
- 输出：`next: draft-phase`

### Draft
并行启动 2 个 `module-code-analyzer`：
- Agent 1：读取 `user.go`，建议补丁添加 `notif.EmailSender.Send(ctx, NewUserEmail{...})`。
- Agent 2：读取 `email.go`，建议新的 `NewUserEmail` 模板函数。
- 合并结果：2 个补丁，各 1 条测试计划。

### Precheck
- `go build ./core/... ./notification/...` → 通过
- `go vet ./...` → 通过
- 未检测到补丁冲突。
- 输出：`next: act-phase`

### Act
通过 `apply_patch` 应用两个补丁。

### Evaluate
- `go build ./...` → 通过
- `go test ./core/... ./notification/...` → 通过
- 完成度检查："新用户注册现在触发邮件通知。测试覆盖正常路径和模板格式化。"
- 输出：`completeness: done`，工作流结束。

---

## 示例 2：Bug 修复

**用户请求：** "修复：删除用户时数据库中残留孤儿会话记录。"

### Observe
`doc-index-reader` → `modules: ["api", "core", "infra"]`、`unknowns: ["会话清理流程"]`。

### Orient
`code-structure-analyzer`：
- `query_flow("core", "DeleteUser")` → `DeleteUser → repo.DeleteUser →（无会话清理）`。
- `query_module("infra")` → 包含带 `DeleteByUserID` 的 `SessionRepo`。
- 输出：`impact_scope: {direct: ["core/service/user.go"], indirect: ["infra/repo/session.go"]}`。

### Decide
- `task_type: "bugfix"`
- `dispatch_plan:` 1 个 Agent → `core/service/user.go` 范围，任务：添加 `sessionRepo.DeleteByUserID` 调用。

### Draft → Precheck → Act → Evaluate
标准流程。单个补丁，预检通过，应用成功，测试通过。

---

## 示例 3：重构

**用户请求：** "将认证逻辑从 api/handler 提取到专用的认证中间件包。"

### Observe → Orient
`code-structure-analyzer` `query_flow("api", "Auth")` 映射所有 handler 中的认证触点。

### Decide
- `dispatch_plan:` 3 个 Agent：
  1. `api/middleware/` — 创建新的 `auth.go`
  2. `api/handler/` — 从所有 handler 中移除内联认证
  3. `api/router.go` — 接入新中间件

### Draft（因范围重叠需串行）
Agent 1 先运行 → 创建中间件。
Agent 2 使用 Agent 1 的输出运行 → 剥离 handler。
Agent 3 使用 Agent 1+2 的输出运行 → 重接路由。

### Precheck
`go build` 捕获缺失导入 → 回退到 Draft 一次 → 修复 → 通过。

### Act → Evaluate
所有补丁已应用，测试通过。

---

## 示例 4：单元测试生成

**用户请求：** "为 core/service/user.go 生成测试，达到 80% 覆盖率。"

### Observe → Orient
`code-structure-analyzer` `query_entrypoint("core")` + `query_module("core")` → 映射所有导出函数。

### Decide
- `task_type: "test-gen"`
- `dispatch_plan:` 1 个 Agent → `core/service/user.go` 范围，任务：为 `Register`、`Update`、`Delete`、`Get` 生成表驱动测试。

### Draft
`module-code-analyzer` 读取 `user.go`，分析函数签名、错误路径、依赖关系。生成带 mock 设置的 `user_test.go`。

### Precheck
`go vet` → 通过（测试文件可编译）。`go test -run=TestRegister` → 样例运行通过。

### Act
应用测试文件补丁。

### Evaluate
`go test -cover ./core/service/` → 82% 覆盖率。`completeness: done`。

---

## 示例 5：代码审查

**用户请求：** "审查 api/handler/ 中的最近变更，检查安全问题。"

### Observe → Orient
`doc-index-reader` + `code-structure-analyzer` → 映射 handler 结构、依赖流。

### Decide
- `task_type: "review"`
- `dispatch_plan:` 1 个 Agent → scope=`api/handler/`，任务：安全审计（输入验证、认证检查、SQL 注入、错误暴露）。

### Draft
`module-code-analyzer` 读取所有 handler 文件。生成审查报告（无补丁），标记：
- `user.go:67` — email 字段缺少输入净化
- `admin.go:23` — DeleteUser 端点缺少角色检查

### Precheck（跳过 — 审查任务，无补丁可应用）

### Act（跳过 — 无补丁）

### Evaluate
审查发现报告给用户。`completeness: done`。

---

## 示例 6：单元测试修复

**用户请求：** "TestUserUpdate 是 flaky 测试 — 修复它。"

### Observe → Orient
`code-structure-analyzer` `query_flow("core", "UpdateUser")` → 发现使用 `time.Now()` 的时间依赖逻辑。

### Decide
- `task_type: "test-fix"`
- `dispatch_plan:` 1 个 Agent → scope=`core/service/user_test.go:TestUserUpdate`，任务：找到并修复 flakiness。

### Draft
`module-code-analyzer` 读取 `user_test.go`，识别 `time.Now()` 未使用注入时钟。建议：添加 `clock` 接口，在测试中注入 mock。

### Precheck → Act → Evaluate
补丁已应用。`go test -count=10 -run=TestUserUpdate` → 10/10 通过。完成。

---

## Before/After Diff 示例

### Diff 1: Bug 修复（示例 2 对应）

`domain/notify/dispatcher.go` — 补全 CooldownTracker.Record 调用：

```diff
 func (d *Dispatcher) Dispatch(ctx context.Context, events []AlertEvent) {
 	validEvents, skippedEvents := d.applyCooldown(events)
 
 	for _, e := range skippedEvents {
 		d.recordSkipped(e)
 	}
 
+	// 记录有效事件的冷却时间（防止短时间内重复推送）
+	for _, e := range validEvents {
+		if d.cooldown != nil {
+			d.cooldown.Record(e.Symbol, e.Type)
+		}
+	}
+
 	if len(validEvents) == 0 {
 		return
 	}
```

`domain/notify/notify_test.go` — 修正 HistoryBuffer 查询顺序断言：

```diff
-	if records[0].Status != StatusSuccess {
-		t.Error("first record should be success")
+	if records[0].Status != StatusFailed {
+		t.Error("first record (newest) should be failed")
 	}
-	if records[1].Status != StatusFailed {
-		t.Error("second record should be failed")
+	if records[1].Status != StatusSuccess {
+		t.Error("second record (oldest) should be success")
 	}
```

### Diff 2: 功能开发（示例 1 对应）

`core/service/user.go` — 在用户注册后添加通知调用：

```diff
 func (s *UserService) Register(ctx context.Context, req RegisterRequest) (*User, error) {
 	user, err := s.repo.Insert(ctx, req)
 	if err != nil {
 		return nil, fmt.Errorf("insert user: %w", err)
 	}
 
+	// 异步发送欢迎邮件
+	go func() {
+		if err := s.notifier.SendWelcome(context.Background(), user.Email); err != nil {
+			slog.Warn("send welcome email failed", "user_id", user.ID, "error", err)
+		}
+	}()
+
 	return user, nil
 }
```

### Diff 3: 重构（示例 3 对应）

`api/middleware/auth.go` — 新建认证中间件（从 handler 中提取）：

```go
// 新增文件: api/middleware/auth.go
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/myapp/domain/auth"
)

// AuthRequired 验证请求中的 JWT Token，未通过返回 401。
func AuthRequired(authenticator auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		principal, err := authenticator.ValidateBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("principal", principal)
		c.Next()
	}
}
```

老 handler 剥离后：

```diff
 func (h *Handler) GetUsers(c *gin.Context) {
-	token := c.GetHeader("Authorization")
-	if token == "" {
-		c.JSON(401, gin.H{"error": "missing token"})
-		return
-	}
-	claims, err := h.auth.ValidateBearerToken(token)
-	if err != nil {
-		c.JSON(401, gin.H{"error": "invalid token"})
-		return
-	}
-
 	users, err := h.service.List()
 	// ...
 }
```
