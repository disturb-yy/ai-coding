# 执行规则

从 SKILL.md 提取的源码读取、Token 优化、失败恢复规则。

---

## 源码读取规则

```
                  能否读源码？
                  ┌──────────────┐
主 Agent          │    永不      │
doc-index-reader  │    永不      │
code-structure-analyzer  │    永不      │
module-code-analyzer │ 仅限范围   │
test-runner       │    永不      │
```

### 范围读取

`module-code-analyzer` **仅读取**与其调度范围匹配的文件：
- **模块范围：** `module/...` 内的所有 `.go` 文件
- **文件范围：** 单个 `.go` 文件
- **函数范围：** 包含该函数的文件 + 直接调用者/被调用者（最多 3 个文件）

### 反模式（禁止）

- 在整个仓库执行 `rg` 或 `grep`
- 在 Observe 阶段范围外读取 `go.mod`、`go.sum`
- 仅一个函数相关时读取整个包的所有文件
- 分析生产代码时读取测试文件（除非任务是测试相关）
- 跨 SubAgent 重复读取同一文件

---

## Token 优化规则

1. **始终索引优先。** 人工维护的文档成本 < 代码结构工具查询 < 源码读取。
2. **一个 Agent 一个职责。** 不让一个 Agent 同时做索引读取和源码分析。
3. **仅限范围读取。** `module-code-analyzer` 只读所需内容，不多读。
4. **缓存 代码结构工具结果。** 跨 SubAgent 去重查询。
5. **并行化不重叠的范围。** 读取不同模块的两个 Agent 并发运行。
6. **主 Agent 保持精简。** 主 Agent 上下文仅包含结构化数据，永不包含原始源码。
7. **使用 chunks，不用完整文件。** SubAgent 返回相关摘录（`chunks`），而非整个文件。
8. **可能时跳过阶段。** 若 Observe 达到高置信度，跳过 Orient。

---

## 失败恢复规则

| 失败情况 | 阶段 | 恢复策略 |
|---------|-------|----------|
| doc-index-reader 返回空 | Observe | 直接启动 code-structure-analyzer；设置低置信度 |
| 代码结构工具查询超时 | Orient | 重试一次；若持续，标记模块为需要深入源码 |
| module-code-analyzer 生成无效补丁 | Draft | 用更具体的范围 + 错误详情重新启动；最多 2 次重试 |
| Precheck（go build）失败 | Precheck | 带构建错误返回 Draft；仅针对性重新分析 |
| 补丁冲突 | Act | 返回 Draft；仅重新分析冲突文件 |
| 测试失败 | Evaluate | 返回 Orient；重新评估影响范围；最多 3 次工作流迭代 |
| spawn_agent 不可用 | 任意 | 回退到主 Agent 本地执行（见 SKILL.md 回退策略） |
| 所有迭代已耗尽 | Evaluate | 向用户报告：已完成项、剩余项、失败原因 |

### 迭代限制

最多 3 次完整工作流迭代（Observe→Evaluate）。第 3 次迭代时即使未完成也报告结果（partial log_analysis）。绝不无限循环。

---

## 并发策略

| 场景 | 策略 |
|----------|----------|
| 不重叠的范围 | 并行调度（最多 3） |
| 重叠的范围 | 串行；将前序输出作为上下文传递 |
| 单一目标 | 单个 Agent |
