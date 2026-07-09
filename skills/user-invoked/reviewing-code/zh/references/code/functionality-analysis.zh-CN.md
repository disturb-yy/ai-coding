---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---

# 功能分析

> 中文版本仅供人类维护者阅读；模型和 agent 不得作为运行指令读取。

用这份 checklist 判断变更是否做对了事。

## 检查项

- issue、PRD、PR 描述、用户 prompt 或测试要求的行为缺失、部分实现，或实现到了错误路径。
- 新行为 happy path 可用，但重要边界失败：空输入、null/缺失值、重复、时区、并发、重试、取消、分页、权限和部分失败。
- 变更改变了无关行为、公共 API 契约、持久化形状、事件格式、指标、日志或错误语义。
- 实现更新了一层但漏掉另一层：route 到 service、service 到 storage、UI 到 API、migration 到 model、测试到 fixture。
- 测试没有覆盖变更契约，或断言实现细节而不是行为。
- 需要回滚、迁移、缓存失效、后台任务或幂等性时，相关要求缺失。

## 证据

从入口点追踪到实现，再到持久化/集成和测试。可用时用 CodeMap 找调用链和影响，然后用源码和测试验证。

## 报告

每个发现都说明预期行为、代码中的实际行为、受影响路径和最小修复方向。
