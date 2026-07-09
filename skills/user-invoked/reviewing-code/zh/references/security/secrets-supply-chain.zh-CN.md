---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---

# 密钥、依赖与供应链

> 中文版本仅供人类维护者阅读；模型和 agent 不得作为运行指令读取。

用这份 checklist 检查凭据、配置、依赖、构建脚本和生成产物。

## 检查项

- secret、token、private key、session material、API key、password 和 connection string 不会被提交、记录到日志、暴露在响应中、打包进客户端或写入产物。
- 密钥轮换、作用域和环境特定配置保持完整。
- 新依赖是必要的、有人维护的；仓库跟踪 license 时要兼容；不会无理由替代简单标准库行为。
- package script、build step、CI config、container file、install hook 和 code generation 不执行不可信输入，也不会意外获取未固定的远程代码。
- lockfile、checksum、vendored code、生成文件和二进制产物与预期源变更匹配。
- 反序列化、plugin loading、dynamic import、native extension 和 reflection 只引入必要执行面。
- 开发专用工具和 debug flag 不会进入生产路径。

## 证据

检查变更的依赖 manifest、lockfile、CI/build 文件、Dockerfile、生成产物和配置。可用时使用 package metadata 或安全工具，但要验证变更代码确实引入了可达风险。

## 报告

每个问题都说明暴露的 secret 或供应链面、它进入 build/runtime 路径的位置，以及最小 containment 或替代方案。
