---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---

# 认证与访问控制

> 中文版本仅供人类维护者阅读；模型和 agent 不得作为运行指令读取。

用这份 checklist 检查权限、租户、身份和受保护操作。

## 检查项

- 新增或变更的受保护路径需要认证。
- 授权检查发生在服务端或可信边界，而不是只在 UI 或客户端。
- 对象级访问控制能阻止读取或修改其他用户、租户、组织或项目的数据。
- admin、service-account、internal、feature-flag、debug 和 maintenance 路径不会暴露给普通用户。
- 安全决策不会在未验证时信任用户可控 identifier、header、cookie、query param、role name 或客户端状态。
- 新后台任务、webhook、queue、cron task 和 callback 保留授权与租户上下文。
- 错误信息和日志不会跨边界暴露受保护资源是否存在。

## 证据

从 request 或 job 入口追踪到 auth middleware、policy check、数据过滤和 storage query。用测试或既有 policy helper 做比较点。

## 报告

说明 actor、受保护资产、缺失或被绕过的检查，以及利用路径。当可能出现未授权数据访问或修改时标为 `Critical` 或 `High`。
