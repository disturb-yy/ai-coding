---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---

# 语法分析

> 中文版本仅供人类维护者阅读；模型和 agent 不得作为运行指令读取。

用这份 checklist 查找应在行为审查前发现的缺陷。

## 检查项

- 变更文件存在解析或编译失败。
- 类型错误、缺失 import、错误 export、未定义符号、破坏的泛型和无效 annotation。
- caller 与 callee 的 API 签名不匹配。
- JSON、YAML、TOML、SQL、GraphQL、模板、迁移和生成 manifest 中的配置/schema 语法问题。
- 语法合法但语言工具会标记的 async、资源或生命周期误用。
- 测试文件不再能编译、import、发现或运行。

## 证据

优先使用项目原生命令：lint、typecheck、compile、test discovery 或包级验证。命令不可用时，读取最小的 caller/callee 或 config/schema 对来证明问题。

## 报告

只报告 diff 中可见或由 diff 直接造成的问题。包含精确符号、import、签名、配置 key 或命令失败。
