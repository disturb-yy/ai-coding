# CodeMap 集成规则

## 目录
- [API 参考](#api-参考)
- [查询顺序规则](#查询顺序规则)
- [CodeMap 不足够时](#codemap-不足够时)
- [缓存与去重](#缓存与去重)
- [错误处理](#错误处理)

---

## API 参考

### `query_module(name: string) → ModuleSummary`

返回 Go 模块的结构性概览。

```json
{
  "name": "core/service",
  "packages": ["service", "service/mocks"],
  "exported_symbols": [
    {"name": "UserService", "kind": "interface", "methods": ["Register", "Update", "Delete", "Get"]},
    {"name": "Register", "kind": "function", "signature": "func(ctx, req) (User, error)"}
  ],
  "sub_packages": ["mocks"]
}
```

**何时调用：** 对 `index.modules` 中的每个模块。Orient 阶段的第一个查询。

**Token 节省：** 替代读取一个包中所有 `.go` 文件（每个文件约 500-2000 Token）。

---

### `query_flow(module: string, symbol?: string) → FlowGraph`

返回模块或符号的数据/控制流图。

```json
{
  "symbol": "Register",
  "flow": [
    {"step": 1, "action": "验证输入", "calls": ["validator.Validate"]},
    {"step": 2, "action": "哈希密码", "calls": ["bcrypt.GenerateFromPassword"]},
    {"step": 3, "action": "插入用户", "calls": ["repo.Insert"]},
    {"step": 4, "action": "返回 User", "returns": "User"}
  ],
  "side_effects": ["repo.Insert"],
  "error_paths": ["步骤 1: 无效输入 → 400", "步骤 3: DB 错误 → 500"]
}
```

**何时调用：** 针对被修改的目标函数/方法。Orient 阶段的第二个查询。

**Token 节省：** 替代手动追踪调用链（约 1000-5000 Token）。

---

### `query_dependency(module: string, direction: "incoming"|"outgoing") → DependencyGraph`

```json
{
  "module": "core/service",
  "outgoing": ["core/repo", "infra/cache", "pkg/validator"],
  "incoming": ["api/handler", "cmd/worker"]
}
```

**何时调用：** 对 `outgoing`：始终调用，理解模块依赖什么。
对 `incoming`：当修改导出 API 接口时（破坏性变更风险评估）。

**Token 节省：** 替代 `go mod graph` 解析 + import 语句扫描。

---

### `query_entrypoint(module: string) → EntrypointMap`

```json
{
  "module": "api",
  "entrypoints": [
    {"file": "cmd/server/main.go", "function": "main", "description": "HTTP 服务器启动"},
    {"file": "api/router.go", "function": "NewRouter", "description": "路由注册"}
  ],
  "public_api": ["GET /users", "POST /users", "DELETE /users/:id"],
  "init_order": ["config.Load", "db.Connect", "cache.Warmup", "server.Start"]
}
```

**何时调用：** 针对被修改的模块。帮助理解变更在哪里集成。

---

## 查询顺序规则

按此顺序执行 CodeMap 查询以最大化每条 Token 的信息量：

| 优先级 | 查询 | 何时 |
|--------|------|------|
| 1 | `query_module` | 始终，对范围内的每个模块 |
| 2 | `query_entrypoint` | 对每个被修改的模块 |
| 3 | `query_flow` | 对目标函数/方法（如已知） |
| 4 | `query_dependency(outgoing)` | 对每个被修改的模块 |
| 5 | `query_dependency(incoming)` | 仅当修改导出 API 时 |

---

## CodeMap 不足够时

CodeMap 提供结构但非实现细节。以下情况需降级到 `module-code-analyzer`：

1. **需要具体代码变更** — CodeMap 显示*有什么*但无法提供补丁所需的行级细节。
2. **实现逻辑分析** — 需要读取实际函数体（边界情况、验证逻辑）。
3. **测试生成** — CodeMap 显示函数签名但无法提供覆盖率所需的内部分支。
4. **Bug 根因** — CodeMap 显示流程但无法显示导致 bug 的状态变更。
5. **代码审查** — 需要读取实际代码以检查安全/质量问题。

---

## 缓存与去重

- 在单次工作流调用中缓存每个模块的所有 CodeMap 响应。
- 若同一模块被多个 SubAgent 查询，重用缓存结果。
- **不要**对同一模块+查询组合重复查询 CodeMap。

---

## 错误处理

| 错误 | 操作 |
|--------|--------|
| 模块未找到 | 标记模块为未知，继续处理剩余模块 |
| 符号未找到 | 扩展到父级范围查询，重试 |
| 超时 | 重试一次；若仍失败，标记模块为需要深入源码 |
| 认证/权限错误 | 立即向用户报告；不要继续 |
