---
okf_version: "0.2"
---

# Cognitive Control Plane 角色目录

本目录是角色文档的正式维护位置。原 `~/tool/skills/role` 仅保留兼容符号链接；从兼容位置读取文档时，先解析真实路径再解析相对引用。

仅在需要角色职责或已有角色交接契约时读取对应文件。普通会话以 [SKILL.md](../SKILL.md) 和 [skill routing](../references/skill-orchestration.md) 为准；角色名称不创建新 agent，输出字段按消费者实际契约选取，角色转换不暂停已授权任务。宿主模式的严格要求只在该模式显式启用时适用。

# 共享标准

* [角色交接标准](handoff-standard.md) - 在角色之间交接任务、证据、工具路由与失败状态的统一合同。

* [记忆生成者](memory-generator.md) - 从已完成的会话与任务中提炼可验证、可复用且有明确生命周期的长期记忆。
* [需求访谈者](requirements-interviewer.md) - 将模糊需求澄清为可执行目标、验收标准与下一步路由。
* [问题建模者](problem-framer.md) - 将模糊症状或问题界定为可验证的分析框架与行动交接。
* [代码库探索者](codebase-explorer.md) - 在安全编辑前定位经验证的代码入口、流程、改动点与测试路径。
* [对抗式审查者](adversarial-reviewer.md) - 以明确标准攻击具体工件并给出缓解措施与残余风险。
* [代码审查者](code-reviewer.md) - 对固定代码变更进行证据驱动、按严重性排序的质量与安全审查。
* [实施执行者](implementation-worker.md) - 在既有仓库中实施已确认的变更并提供验证证据。
* [测试驱动执行者](tdd-worker.md) - 通过红绿重构循环实施行为，并保留可验证的测试证据。
