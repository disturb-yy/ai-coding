# Model-invoked skills

本目录的技能面向模型自动路由：当用户请求符合 `SKILL.md` 中的描述和触发条件时，agent 应加载相应技能。用户也可以直接点名技能。下面的说明用于选型；执行时必须以各技能的 `SKILL.md` 和其按需引用的文件为准。

## 一览

| 技能 | 解决什么问题 | 何时使用 | 核心做法与产出 |
| --- | --- | --- | --- |
| [agent-reach](agent-reach/SKILL.md) | 从互联网和支持的平台获取信息 | 用户要求调研、搜索、查找，提到受支持社交平台或给出 URL | 先检查可用后端，再按 search、social、career、dev、web 或 video 路由抓取；产出带来源的检索材料，而非代写发布内容。 |
| [coding-project](coding-project/SKILL.md) | 在已有仓库中安全实现代码变更 | 要编辑源码、测试、依赖、生成物或实现相关文档 | 先扫描上下文和语言引用，再按 Observe → Orient → Decide → Draft → Precheck → Act → Evaluate 完成狭窄修改与验证。 |
| [coding-tdd](coding-tdd/SKILL.md) | 用测试先行实现可观察行为 | 用户要求 TDD、回归修复或可切成小片的集成行为 | 以一条外部可观察的行为为单位，严格走 Red → Green；所有切片完成后先审查，再做保持绿色的最终重构。 |
| [diagnosing-problem](diagnosing-problem/SKILL.md) | 把含糊问题收敛为可回答的定义 | 需要澄清解释、暴露假设、比较理解方式、设定证据标准或形成交接 | 依次 framing、gate、gather、answer/handoff；明确选择的解释、假设、证据强度、限制和下一步。 |
| [exploring-project](exploring-project/SKILL.md) | 在改动前或解释前定位代码路径 | 要查看结构、追踪功能/路由/模块/函数，或计划范围明确的代码变更 | 优先以 Graphify 看跨区结构、CodeMap 找候选，再用 `rg`、源码和测试验证；输出已核实的流程和下一修改点。 |
| [grilling](grilling/SKILL.md) | 逐项拷问方案或设计中的未决问题 | 用户说要 grill、压力测试方案，或需要深入对齐 | 每次只提出一个问题并给出推荐答案；能从代码库得到的答案先探索，不把可验证事实变成用户问答。 |
| [okf-frontmatter](okf-frontmatter/SKILL.md) | 维护可检索、不过度复制 schema 的 Markdown 知识库 | 要添加/校验 OKF frontmatter、找文档归属或 schema、创建 ADR/章节 | 用 YAML frontmatter 管理文档元数据，schema 指向权威代码；先精确 `rg`，模糊时再运行 `find`、`schema`、`index`、`lint` 或 `new` 工具。 |
| [reviewing-code](reviewing-code/SKILL.md) | 对代码或实现产物做循证审查 | 审查 PR、提交、分支、diff、文件，或明确要求安全审查 | 定义审查目标、选择代码/安全审查包、分离审查通道并核验证据；交付验证矩阵、分级问题和残余风险。 |

## 常见组合

1. **不清楚要解决什么**：先用 `diagnosing-problem` 界定问题，再按交接结果进入探索、研究或实现。
2. **要改代码但还不知道位置**：用 `exploring-project` 找到已验证的调用路径和测试，随后交给 `coding-project`。
3. **要测试驱动实现**：`coding-tdd` 以 `coding-project` 的项目约定、语言引用和验证路径为基础；完成所有 GREEN 切片后调用 `reviewing-code`，然后才重构。
4. **外部资料会影响结论**：使用 `agent-reach` 获取材料；不要将搜索摘要直接当作已验证的领域事实。
5. **文档也是交付物**：仅需仓库说明时遵循项目文档约定；当要用 OKF 元数据、按意图检索或解析 schema 指针时使用 `okf-frontmatter`。

## 边界提醒

- `coding-project` 不用于纯阅读、GitHub PR 流程、全新项目脚手架、产品设计或技能创作。
- `exploring-project` 的导航图、CodeMap 命中和搜索结果都是线索，关键行为仍须由源码或测试确认。
- `reviewing-code` 需要具体审查对象；没有代码证据的架构讨论不应伪装成审查结论。
- `agent-reach` 负责获取互联网信息，不负责代用户发帖、评论、点赞或将资料加工为最终报告。
