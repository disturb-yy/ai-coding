# User-invoked skills

本目录保存需要用户明确发起的技能。通常在请求中写 `$<技能名>` 即可，例如：

```text
$grill-ui-with-docs，为现有 React 项目设计预警中心页面。
```

这些技能把高控制度、较长流程或明确的人工确认点留给用户选择。运行时规则仍以各目录中的 `SKILL.md` 为准。

## 一览

| 技能 | 用途 | 怎么发起 | 主要产出/完成条件 |
| --- | --- | --- | --- |
| [cognitive-control-plane](cognitive-control-plane/SKILL.md) | 为复杂 AI 协作选择下一步控制面、委托方式和证据策略 | `$cognitive-control-plane`，用于上下文不清、假设风险、计划批评、跨阶段编排或交接 | 把工作分为 Tiny/Small/Large，选择 context、epistemic、adversarial 或 output 控制面；需要委托时形成可执行 contract，而不是假称已启动 worker。 |
| [explore-project](explore-project/SKILL.md) | 引导式地介绍项目 | `$explore-project` 后说明要了解的项目或问题 | 从代码库当前状态出发，给出循序渐进的项目介绍；适合“带我看懂这个项目”。 |
| [gen-index](gen-index/SKILL.md) | 生成或刷新 agent 可读的仓库索引 | `$gen-index`，并说明目标仓库或希望生成的索引 | 综合 Graphify、CodeMap、定向源码读取和人工文档，主要写 `.agent/PROJECT_INDEX.md`，按事实需要补充导航、变更指南或能力目录。 |
| [grill-me](grill-me/SKILL.md) | 单纯用追问打磨计划或设计 | `$grill-me` 后给出计划、设计或目标 | 进入一次 `grilling` 会话：每次只问一个会影响后续决策的问题，并提供建议答案。 |
| [grill-ui-with-docs](grill-ui-with-docs/SKILL.md) | 将模糊 UI 想法收敛为经确认的设计交接包 | `$grill-ui-with-docs`，描述页面/流程意图及现有项目约束 | 逐个确认设计决策，先交 Markdown/ASCII wireframe；用户批准后形成可实现的设计系统和交接文档。 |
| [grill-with-docs](grill-with-docs/SKILL.md) | 一边追问设计，一边沉淀领域文档 | `$grill-with-docs` 后给出待澄清的方案 | 结合 `grilling` 与 `domain-modeling`，将稳定结论写入术语表或 ADR，而不是只停留在对话中。 |
| [create-sop](create-sop/SKILL.md) | 将用户提供的流程转成可执行的 DAG 标准作业程序（SOP） | `$create-sop`，提供流程描述及需要的交付位置或格式 | 形成含节点、阻塞边、角色、决策门、例外升级、证据和完成条件的可干跑 SOP，并明确假设和待确认项。 |
| [execute-sop](execute-sop/SKILL.md) | 以阶段门和持久状态处理问题单 | `$execute-sop`，提供问题描述和可选的状态目录 | 先分析、等待用户批准，再建立用户执行的 DAG tickets；`STATE.md` 记录阶段、frontier、证据和交接状态。 |
| [reviewing-code](reviewing-code/SKILL.md) | 对具体代码变更做多通道、可追溯审查 | `$reviewing-code`，提供 PR、分支、diff、文件或比较基准 | 分开检查语法、功能、规范及按需安全面；每项发现都需有文件、hunk、测试或标准证据，并汇总为严重度排序的报告。 |
| [work-canvas](work-canvas/SKILL.md) | 将工作状态、决策或比较做成离线可打开的交互页面 | `$work-canvas`，说明是状态、决策还是比较场景 | 生成一个自包含 HTML 文件，带真实待确认事项、图例和来源脚注；不修改用户源码，也不依赖外网/CDN。 |
| [writing-great-skills](writing-great-skills/SKILL.md) | 编写或评估可预测、边界清晰的技能规则 | `$writing-great-skills`，用于创建或修改 skill 的描述、结构和引用层级 | 依据触发成本、信息层级、概念空间、拆分粒度和 no-op 剪枝原则，产出更可执行、易维护的 `SKILL.md` 设计。 |

## 选用建议

| 需要的结果 | 选择 |
| --- | --- |
| 先决定复杂任务应该由谁做、如何分阶段和如何验证 | `cognitive-control-plane` |
| 看懂现有项目的目录、核心概念和实现方式 | `explore-project` |
| 留下一份可供后续 agent 快速导航的仓库说明 | `gen-index` |
| 把产品、架构或方案的决策逐一问透 | `grill-me` |
| 同时要 UI wireframe、设计系统和实现交接 | `grill-ui-with-docs` |
| 将现有流程固化为可执行、可审计的 DAG SOP | `create-sop` |
| 让问题单经用户批准后拆为 tickets，并跨模型维护处理状态 | `execute-sop` |
| 不只做设计讨论，还要把领域语言和决策落为文档 | `grill-with-docs` |
| 要判断变更能否接受 | `reviewing-code` |
| 需要在浏览器里审阅进展、方案或多个选项 | `work-canvas` |
| 要创建或收敛一个 skill 的指令结构 | `writing-great-skills` |

## 使用注意

- `cognitive-control-plane` 的 README 是维护说明；模型运行时应读取其 `SKILL.md` 及必要的 `references/`，不要把维护 README 当成执行规则。
- `work-canvas` 输出是工作成果的可视化快照，不是要发布的产品 UI；产品页面应走相应的前端实现流程。
- `gen-index` 优先建立准确、可验证的导航，不用它替代产品 README 或逐文件百科。
- `grill-*` 系列会保留用户对关键决策的控制权；如果答案能由代码或现有资料直接验证，技能应先查证再提问。
