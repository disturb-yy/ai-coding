---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---

# 规范分析

> 中文版本仅供人类维护者阅读；模型和 agent 不得作为运行指令读取。

用这份 checklist 对照仓库约定检查变更。

## 来源顺序

1. 显式仓库规则：`CONTRIBUTING`、`CODING_STANDARDS`、docs、package README、架构说明、lint config、formatter config、type config。
2. 附近文件和测试中的本地模式。
3. 只有缺少仓库证据时，才使用通用可维护性启发。

仓库规则覆盖通用启发。附近可工作的例子强于全局偏好。

## 检查项

- 命名、文件位置、模块边界、分层和依赖方向符合附近代码。
- 错误处理、日志、重试、指标、事务和清理遵循既有模式。
- 测试使用本地 fixture、helper、命名和断言风格。
- 公共接口、schema、迁移和生成产物遵循仓库既有更新路径。
- 变更避免重复逻辑、投机抽象、feature envy、shotgun edits、middle-man wrapper 和重复条件级联，除非仓库模式要求。

## 报告

区分硬性仓库规则违反和判断项。硬违反要引用规则文件或附近先例。通用设计异味标记为判断项。
