# Third-party and reusable skills

本目录包含引入的第三方技能或可跨工作流复用的技能。各目录可能自带安装文档、脚本、引用材料或上游许可证；使用与维护时优先遵循该技能本身的 `SKILL.md` 和配套 README。

## 一览

| 技能 | 解决什么问题 | 适用场景 | 使用要点 |
| --- | --- | --- | --- |
| [agent-reach](agent-reach/SKILL.md) | 路由多平台互联网检索能力 | 需要网页、社交媒体、招聘、代码、视频、播客、RSS 或行情信息 | 先运行 `agent-reach doctor --json` 确认每个平台当前后端；按平台和任务类型读取对应 `references/`，只做信息获取，不执行社交写操作。 |
| [domain-modeling](domain-modeling/SKILL.md) | 构建和维护项目领域模型 | 要统一领域术语、澄清歧义、记录架构决策，或其他技能需要稳定的领域语言 | 挑战术语表、用具体场景检验措辞、与代码交叉验证；将已确认概念更新到 `CONTEXT.md`，仅在必要时创建 ADR。 |
| [graphify](graphify/SKILL.md) | 把代码、文档及多种内容转成可查询的持久知识图谱 | 需要了解跨文件关系、架构社区、影响路径，或已有 `graphify-out/` 要查询 | 检查安装和扫描范围，抽取实体/关系并构建图谱；图谱是导航和分析层，重要行为仍需用源码、测试或原始资料验证。 |
| [grill-me](grill-me/SKILL.md) | 发起严格的一问一答式方案访谈 | 希望在实现前把计划或设计问透 | 它委托给 `grilling`；每次一个问题、等待回答、附带推荐，直到关键决策已对齐。 |
| [grill-with-docs](grill-with-docs/SKILL.md) | 在方案访谈中同时沉淀领域文档 | 设计或架构讨论需要同步形成术语表、ADR 等长期资料 | 将 `grilling` 和 `domain-modeling` 组合，避免关键结论只存在聊天记录中。 |
| [grilling](grilling/SKILL.md) | 提供逐项追问的基础访谈机制 | 用户明确要求 grill、压力测试计划或设计 | 一次只问一个问题；能通过代码库探索回答时先探索，避免无谓地向用户追问事实。 |
| [okf-frontmatter](okf-frontmatter/SKILL.md) | 以 Open Knowledge Format 管理和检索 Markdown 文档 | 要维护 YAML frontmatter、找到某个概念由哪份文档负责、解析其 schema 指针，或创建 ADR/章节 | 保持 frontmatter 为元数据单一事实来源，以 `schema_source` 指回代码；精确命中优先 `rg`，歧义时用内置脚本排序或解析。 |
| [to-prd](to-prd/SKILL.md) | 将已有对话直接综合为 PRD 并发布到项目 issue tracker | 讨论已基本完成，希望不再访谈而直接沉淀产品需求 | 从现有会话提炼问题、方案、用户故事、实现/测试决定和范围外事项，再写入 issue tracker。 |

## 关系与取舍

- `grill-me` 是 `grilling` 的显式入口；`grill-with-docs` 在相同访谈基础上增加领域建模与文档沉淀。
- `graphify` 偏向全局关系和架构问题；查找单一符号、路由或实现事实时，应在图谱引导后使用定向搜索和源码验证。
- `agent-reach` 负责外部材料获取；`to-prd` 负责把已讨论的需求整理成可跟踪的产品需求。两者不能相互替代。
- `okf-frontmatter` 维护文档的结构和检索入口；`domain-modeling` 维护项目语言和决策。需要长期文档时，两者可先后使用。

## 重名技能

`agent-reach`、`grilling`、`okf-frontmatter` 等名称也可能在其他分类目录出现。它们是不同的目录来源或注册路径，不应仅凭名称同时加载。应根据当前 agent 的实际技能发现目录、安装来源和项目约定选择唯一实例。
