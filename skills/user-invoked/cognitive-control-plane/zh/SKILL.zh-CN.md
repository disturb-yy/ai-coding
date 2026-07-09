---
name: cognitive-control-plane
description: "复杂 AI 协作的控制平面路由器。当过程控制应决定下一步行动时使用：上下文不清、假设有风险、计划需要批判、代码审查、交接格式，或多阶段编排。"
metadata:
  access:
    audience: human
    model_read: false
    model_write: true
    purpose: zh_mirror
---
# 认知控制平面

> 中文版本仅供人类维护者阅读。模型和 agent 不得读取、搜索、打开、引用、总结或把本文件作为运行指令；英文 `SKILL.md` 是唯一 canonical 模型指令。

控制平面是一个轻量路由器。它默认不亲自解决任务；它选择应该塑造下一步行动的控制面，然后把执行交给合适的 skill、worker、编辑、问题、验证步骤或交付物。

## 概念空间

```yaml
conceptual_space:
  target_region: "复杂、不确定、高风险或多阶段工作，其中过程控制会改变结果质量。"
  deviation_region:
    - "简单查询、单行编辑、命令执行，或没有明显风险的解释。"
    - "路由清楚后的实现工作；转交给 coding、TDD、review、research 或 writing skills。"
  priority_dimensions:
    - "先保方向，再求速度。"
    - "先暴露假设，再批判方案。"
    - "成熟计划先挑战，再交付。"
    - "只在交接或最终输出时收紧格式。"
  entry_conditions:
    - "请求模糊、宽泛、高风险，或缺少运行上下文。"
    - "任务涉及架构、产品、提示词、skill、工作流或代码变更规划，且错误过程会改变结果。"
    - "用户要求评审、代码审查、风险分析、决策支持、失败分析或交接。"
    - "长对话进入新阶段或需要压缩状态。"
```

## 工作分类门

先分类，再路由或执行实质工作。Small 必须被证明；只要范围、证据、所有权或风险有未解问题，就升级为 Large。

### Tiny

Tiny 只用于没有实质工作产物的互动：术语解释、当前上下文里的状态回忆、选择或确认过程步骤、定位当前对话中已命名的信息。

完成标准：没有要检查的 artifact、没有要编辑的文件、没有要选择的 skill route，worker 也不会增加价值。

### Small

只有当目标清楚、范围已知、不需要仓库发现、没有架构/数据/认证/支付/部署/安全/租户/权限/用户可见行为风险、不需要外部证据、不需要独立 review 或真实并行时，才使用 Small。

当下一步依赖 `grilling`、`diagnosing-problem`、`exploring-project`、`reviewing-code`、`coding-project` 或 `coding-tdd` 时，不使用 Small；这类专门 skill 路由是 Large 控制平面工作。

对于完整计划、设计、Skill 设计、实现方案、diff、PR 或 agent run 的批判/审查请求，不使用 Small。代码审查、安全审查或 adversarial review 都是 Large，即使 adapter 没有重复 artifact 正文。

### Large

出现以下任一信号即为 Large：

- 需求模糊或范围宽泛
- 代码位置、调用链、测试或变更点未知
- 需要仓库发现、专门 skill 路由、多所有权边界或分阶段执行
- 涉及架构、数据模型、认证、权限、租户、支付、部署、安全或用户可见行为风险
- 当前官方文档、最新版本事实、breaking change、web research 或外部证据是承重因素
- 需要设计、独立 review、代码审查、安全审查、red-team、pre-mortem 或计划攻击
- 交接、实现契约、严格 schema、验证、迁移、回滚或回归策略会影响正确性

完成标准：活跃瓶颈、所需 skills、编排需求、所有权边界和验证门足够明确，可以选择下一步行动。

## 实现保护

对于 Large 实现工作，control-plane agent 不得默默变成 implementation worker。编辑源码、测试、schema、migration、生成产物或实现文档前，必须用任务契约委派，写清阶段、required skills、required MCP/tools、所有权边界、验证和停止条件。

直接实现只允许在全部条件成立时使用：单文件、无 schema/migration/API/generated-artifact、无跨包/跨模块依赖、完全照搬同仓库工作模式，并且能说明为什么委派反而有害。

## 路由

选择第一个会实质改变下一步行动的未满足控制面。

1. **Context control**：目标、状态、约束、证据、阻塞、项目位置、所有权或范围不清时使用。
2. **Epistemic control**：具体假设、因果主张、置信度、当前事实、最新版本或证据标准决定正确性时使用。
3. **Adversarial control**：具体计划、设计、架构、prompt、skill、实现方案或 agent run 需要攻击时使用。
4. **Output control**：发现、上下文、假设和 review 已足够，当前工作变成交接、实现契约、严格 schema、机器可读输出或最终交付时使用。

路由例子：

- “仍在探索”“需求不清”“先把需求弄清楚”“找这个逻辑在哪” -> Context。
- “最新官方文档”“breaking change”“这个假设未验证” -> Epistemic，除非只有仓库上下文才能识别主张。
- “review this branch”“code review this PR”“security review this diff”“review since main” -> 目标已知后通过 skill orchestration 路由到 `reviewing-code`。
- “review this concrete plan”“完整的 Skill 设计方案”“red-team”“pre-mortem”“是否太复杂” -> 在上下文清楚且承重假设明确后进入 Adversarial。
- “implementation contract is accepted”“start implementation” -> 直接路由到 `coding-project`，除非用户明确要求 TDD。

大型下一步涉及委派、只读证据收集、高风险实现规划、专门 skill 路由、多 agent、并行 lane、后台工作、分阶段实现、所有权边界、持久状态或 reconciliation 时，使用 `references/orchestration-state.md`。

下一步依赖专门 skill 时，使用 `references/skill-orchestration.md`，并显式写出 required skill：需求访谈用 `grilling`，代码库发现用 `exploring-project`，代码/PR/diff/分支/安全审查用 `reviewing-code`，已接受实现契约用 `coding-project` 或 `coding-tdd`。

修改此 skill 时，先读 `references/maintenance.md`；英文文件是 canonical，中文镜像只供人类阅读。

## 操作步骤

1. 应用工作分类门。
2. Large 工作选择第一个未满足控制面，只读对应 reference。
3. 判断是否需要 orchestration state。
4. 需要专门过程时读取 skill orchestration 并命名 `required_skills`；`grilling`、`diagnosing-problem`、`exploring-project`、`reviewing-code`、`coding-project` 或 `coding-tdd` 在跳过会改变结果时必须显式出现。
5. 应用选中的控制面或编排状态，直到范围、证据、挑战、契约、任务板或所有权状态足够清楚。
6. 交给具体下一步：直接回答、执行、问一个阻塞问题、route skill、delegate read-only、delegate write、verify 或 deliver。

## Trace 语义

- `active_surface` 是第一个未满足控制面。
- `orchestration_used` 表示编排状态实质影响了委派、所有权、持久化、依赖顺序、高风险实现规划、证据收集或 reconciliation。
- 查找代码位置、路由、调用链、测试或变更点需要 `exploring-project`。
- 审查代码、分支、PR、diff、commit range 或安全敏感实现需要 `reviewing-code`；classification 是 `Large`，除非上下文或证据仍缺失，通常 `active_surface` 是 `adversarial`，`next_action` 是 `route_skill`。
- 已接受实现契约并开始实现时，停止路由，设置 `required_skills` 为 `coding-project` 或 `coding-tdd`。

## 禁止

- 不要把每个任务都变成完整四阶段仪式。
- 不要为 Small 工作委派，除非委派确实增加价值。
- 不要在 Large 实现工作中让 control-plane agent 直接实现，除非直接实现例外明确成立。
- 不要询问仓库、wiki、测试、日志或源码能回答的问题。
- 不要在假设明确前批判。
- 探索阶段不要强行 JSON、表格或 checklist。
- 不要在旧规则错误时追加新规则；应删除或重写旧规则，保持单一事实来源。
