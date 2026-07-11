---
access:
  audience: model
  model_read: false
  model_write: true
  purpose: skill_reference
---
# 技能编排

委派依赖专用技能的工作时使用此参考。它定义从任务状态到所需技能的默认路由，以及专家必须收到的任务契约形态。

机器可读映射：[`../config/skill-orchestration-map.yaml`](../config/skill-orchestration-map.yaml)

## 技能注册表

| skill | role | source | path | use |
|---|---|---|---|---|
| `grilling` | requirements_interviewer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md` | 每次提出一个问题来澄清需求、方案、设计或问题陈述，直到可执行。 |
| `diagnosing-problem` | problem_framer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md` | 将含糊问题构造成可回答的陈述、假设、证据标准和交接。 |
| `exploring-project` | codebase_explorer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md` | 探索项目结构、行为路径、模块、路由、函数、测试和变更点。 |
| `reviewing-code` | code_reviewer | `available_skill` | `/home/jadon/projects/ai-coding/skills/user-invoked/reviewing-code/SKILL.md` | 审查代码变更、PR、分支、diff 和安全敏感实现制品中的语法、功能、标准及安全问题。 |
| `coding-project` | implementation_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md` | 实现常规代码变更、测试变更、验证、生成制品和语言感知的项目工作。 |
| `coding-tdd` | tdd_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md` | 执行测试优先、红-绿-重构、回归测试优先和按行为切片的实现。 |
| `adversarial-control` | adversarial_reviewer | `file_reference` | `/home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md` | 在接受前攻击具体方案、设计、架构、提示词、技能、实现方法或 agent 追踪。 |

`diagnosing-problem` 构造问题和交接。代码导航仍路由到 `exploring-project`。如果存在专门的根因或运行时故障技能，则从 `diagnosing-problem` 交接路由到它。

## 路由规则

选择第一个匹配的路由：

| signal | required skill | delegate | next action |
|---|---|---|---|
| 用户目标、验收标准、约束、业务规则、角色或交互不清楚，或用户要求澄清需求。 | `grilling` | 通常不委派 | 每次提出一个问题；如果仓库可以回答，就探索而不是提问。 |
| 用户询问原因、哪里有问题、如何定位问题或如何分析现象，且问题尚未构造。 | `diagnosing-problem` | 可选只读 | 生成构造后的问题、假设、证据标准和交接；不要用于澄清功能需求。 |
| 现有项目结构、入口点、调用链、路由、模块、函数、测试或变更点不清楚。 | `exploring-project` | 可选只读 | 选择并记录直接最小化探索或委派的只读报告；生成已验证流程、候选文件、风险和附近测试。 |
| 代码变更、PR、分支、diff、提交范围或安全敏感实现制品需要代码或安全审查。 | `reviewing-code` | 可选只读 | 生成验证矩阵以及按严重性排序的发现表，并包含证据、建议、跳过的检查和残余风险。 |
| 终态实现或修复制品属于安全敏感、跨模块、公共 API、模式/迁移、认证/权限或部署/回滚关键型。 | `reviewing-code` | 强制独立只读 | 预检审查者能力、固定制品版本，然后在接受前运行独立审查；仅当专用技能不可用时使用明确的独立只读回退。 |
| 具体方案、设计、架构、提示词、技能、实现方法、diff、PR 或 agent 运行需要批判。 | `adversarial-control` | 可选只读 | 生成基于标准的批判、有效失败、缓解措施和残余风险。 |
| 需求清楚，现有代码或测试需要常规编辑。 | `coding-project` | 可选的有边界写入 | 窄范围实现并验证。 |
| 用户要求 TDD、测试优先、红-绿-重构或先写回归测试。 | `coding-tdd` | 可选的有边界写入 | 对每个行为切片运行红 -> 绿 -> 重构。 |

## 默认流水线

```text
intake
  -> grilling                   # 需求不清楚
  -> diagnosing-problem         # 问题或现象尚未构造
  -> exploring-project          # 项目路径或变更点不清楚
  -> reviewing-code             # 代码、PR、diff、分支或安全审查
  -> adversarial-control        # 具体方案或 agent 输出需要攻击
  -> coding-tdd | coding-project # 根据测试优先要求选择
  -> verification
  -> ExecutionRun / VerifiedExperience / KnowledgeAsset
```

## 阶段关卡

### Grilling

在以下情况进入：

- 用户给出方向但没有验收标准。
- 业务规则、角色、边界情况或禁止行为可能改变实现。
- 用户正在讨论方案、设计或需求，而不是要求立即执行。
- 用户要求澄清需求，或表示规则、角色、交互或范围尚未想清楚。

当目标、验收标准、范围、约束和下一步行动清楚时退出。如果仓库可以回答该问题，则路由到 `exploring-project`，而不是询问用户。

### Diagnosing Problem

在以下情况进入：

- 输入是含糊、开放式、概念性、战略性、决策、问题定位或原因分析请求。
- 尚不清楚工作应转为代码探索、运行时诊断、外部研究、设计讨论还是直接回答。
- 另一个技能在行动前需要问题陈述、假设、证据标准或成功标准。

不要针对规则、角色、交互或验收标准不清楚的产品或功能需求进入此阶段；将其路由到 `grilling`。

当交接中包含构造后的问题、选定解释、排除的解释、关键假设、证据标准和下一步行动时退出。如果需要项目导航，则路由到 `exploring-project`。

### Exploring Project

在以下情况进入：

- 任务需要项目结构、路由、模块、调用链、函数、测试或变更点。
- 编码前不知道安全编辑边界。

当候选文件、函数、路由、测试、风险和关键证据已通过源代码、测试、Graphify、CodeMap 或定向搜索验证时退出。

### Code Review

在以下情况进入：

- 代码变更、PR、分支、diff、提交范围或实现制品在接受前需要审查。
- 用户要求代码审查、安全审查、PR 审查、diff 审查、分支审查或审查某个 ref 之后的变更。
- 语法、功能正确性、仓库标准或安全发现比实现更重要。

如果审查目标或 diff 基准不清楚，先使用上下文控制或 `exploring-project`。如果制品不是代码，或用户希望攻击方案/设计/提示词/agent 追踪，则路由到 `adversarial-control`。

对于强制的实现后审查，阅读 [`reviewer-enforcement.md`](reviewer-enforcement.md)。审查者 actor 必须不同于实现或修复固定版本的 actor。

当代码/安全发现有证据支持、按严重性排序、已去重、包含跳过的检查和残余风险，并且审查结果绑定到当前制品版本时退出。强制关卡只有在最新有效审查没有阻断性发现时才能退出到接受。

### Coding Project

在以下情况进入：

- 需求清楚。
- 现有仓库代码、测试、依赖、生成制品或实现文档需要编辑。
- 没有明确的测试优先要求。

当窄范围变更完成，且相关验证通过、因具体无关原因失败或因具体原因受阻时退出。

### Adversarial Review

在以下情况进入：

- 具体方案、设计、架构、提示词、技能、实现方法或 agent 运行需要批判。
- 用户要求审查、风险分析、预演失败分析、红队，或询问 agent 是否违反流程。
- 失败模式比产生新实现更重要。

当请求的批判依赖明确且未经验证的关键假设时，不要进入；先路由到认知控制。
不要对常规代码、PR、分支、diff、提交范围或安全审查使用对抗性审查；将其路由到 `reviewing-code`。

当批判命名其标准、区分有效失败与薄弱攻击、给出缓解措施并留下明确残余风险时退出。

### Coding TDD

在以下情况进入：

- 用户要求 TDD、测试优先、红-绿-重构或回归测试优先工作。
- 可在实现前切分出可见行为或小型模块。

当每个切片都经历失败测试 -> 最小实现 -> 变绿 -> 仅在绿色状态重构，并且已运行合并后的受影响测试或最终入口验证时退出。

## SubAgent 契约模式

### 需求澄清

```yaml
task_id: "requirements-1"
actor_id: "requirements-actor-a"
role: requirements_interviewer
phase: context
objective: "澄清用户需求，直到可实现。"
required_skills:
  - name: grilling
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md
    required: true
    reason: "需求仍影响实现路由，需要一次一个问题地澄清。"
required_references: []
required_mcp: []
required_tools: []
edits_allowed: false
expected_output:
  format: clarification_state
  required_fields:
    - clarified_goal
    - acceptance_criteria
    - constraints
    - open_questions
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "仓库或源制品能比用户更好地回答问题。"
```

### 问题构造

```yaml
task_id: "problem-frame-1"
actor_id: "problem-framer-a"
role: problem_framer
phase: context
objective: "构造问题并生成交接。"
required_skills:
  - name: diagnosing-problem
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md
    required: true
    reason: "任务在执行前需要选定解释、假设、证据标准和交接。"
required_references: []
required_mcp: []
required_tools: []
edits_allowed: false
expected_output:
  format: problem_framing
  required_fields:
    - framed_problem
    - selected_interpretation
    - rejected_interpretations
    - assumptions
    - evidence_standard
    - handoff
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "在问题框架被接受前，任务变成具体实现。"
```

### 项目探索

```yaml
task_id: "exploration-1"
actor_id: "explorer-a"
role: codebase_explorer
phase: context
objective: "定位入口点、调用链、候选变更点和附近测试。"
required_skills:
  - name: exploring-project
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md
    required: true
    reason: "编码前必须验证项目路径和安全编辑边界。"
required_references: []
required_mcp:
  - name: CodeMap or Graphify
    required: false
    reason: "可用时用于架构、调用链或跨区域导航。"
required_tools:
  - name: rg
    required: true
    reason: "需要快速源代码搜索来验证候选文件和流程。"
edits_allowed: false
expected_output:
  format: change_point_report
  required_fields:
    - target
    - relevant_files
    - flow
    - leads_checked
    - risks
    - next_change_location
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "在候选文件、流程、风险和测试验证前就要求实现。"
```

### 代码审查

```yaml
task_id: "review-1"
actor_id: "reviewer-actor-b"
role: code_reviewer
phase: review
objective: "在接受前审查代码变更、PR、分支、diff、提交范围或安全敏感实现制品。"
review_of_task_id: "implementation-1"
review_of_actor_id: "implementation-actor-a"
review_iteration: 1
supersedes_review_task_id: ""
review_fallback: none
review_target:
  kind: git_range
  base_sha: ""
  head_sha: ""
  diff_hash: ""
  stable_id: ""
required_skills:
  - name: reviewing-code
    source: available_skill
    path: /home/jadon/projects/ai-coding/skills/user-invoked/reviewing-code/SKILL.md
    required: true
    reason: "代码和安全审查需要语法、功能、标准及安全通道，并进行有证据支持的聚合。"
required_references:
  - name: reviewer-enforcement
    source: file_reference
    path: references/reviewer-enforcement.md
    required: true
    reason: "强制审查需要审查者独立性、制品新鲜度和最终关卡规则。"
required_mcp:
  - name: GitHub
    required: false
    reason: "当审查目标是 PR、审查线程、检查运行、关联 issue 或远程 diff 时使用。"
  - name: CodeMap or Graphify
    required: false
    reason: "当调用链、路由影响、相关测试或跨区域风险需要导航时使用。"
required_tools:
  - name: git diff or PR file list
    required: true
    reason: "接受发现前必须固定审查目标。"
  - name: rg
    required: false
    reason: "用于定向检查源代码、标准、测试和配置证据。"
edits_allowed: false
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: code_review_report
  required_fields:
    - verification_matrix
    - findings_table
    - findings
    - blocking_findings
    - non_blocking_findings
    - no_findings
    - skipped_or_blocked
    - residual_risk
    - review_summary
    - review_target
    - gate_decision
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "审查目标、diff 基准、PR、分支、提交范围或文件列表不可用。"
  - "审查目标可变或缺少稳定版本标识符。"
  - "分配的审查者 actor 与实现或修复目标版本的 actor 相同。"
  - "无法访问所需审查证据。"
  - "任务变成实现而非审查。"
```

如果 `reviewing-code` 不可用，不要悄然省略审查。仅当宿主能够启动不同的只读 actor 时，才可创建 `independent_read_only_reviewer` 契约。设置 `review_fallback: independent_read_only_reviewer`，从 `required_skills` 中省略不可用技能，保留必需的 `reviewer-enforcement` 参考，并在 `deviations` 下列出不可用技能。否则发出交接并保持关卡阻塞。

### 强制实现后审查

在每个实现和修复任务后应用此流水线：

```text
terminal implementation or fix
  -> assess delivered artifact risk
  -> mandatory trigger present?
       -> no: continue normal verification
       -> yes: create independent reviewing-code task
  -> pin artifact version
  -> reconcile review
       -> no blocking findings: clear gate for this exact version
       -> blocking findings: block final, dispatch fix
  -> fix completed
  -> invalidate prior review, pin new version, re-review
  -> repeat until cleared or explicitly terminated without acceptance
```

强制触发项为 `security_sensitive`、`cross_module_change`、`public_api_change`、`schema_change`、`migration`、`auth_or_permission_change` 和 `deployment_or_rollback_critical`。测试通过绝不会把强制审查转为可选审查。

对于每次审查迭代，强制执行：

```yaml
review_invariants:
  reviewer_independent: "review.actor_id != review.review_of_actor_id"
  reviewer_read_only: true
  target_immutable: true
  stale_review_clears_gate: false
  blocking_findings_allow_acceptance: false
  fix_requires_rereview: true
```

如果宿主无法启动独立审查者，发出交接并保持强制关卡阻塞。如果循环被明确终止，将结果报告为已终止且未接受，并包含未解决发现和残余风险。

对此强制实现后任务使用 `next_action: delegate_read_only`，因为必须实际启动不同的 actor。对于普通的用户请求代码审查，如果当前 actor 只是将控制权交给已安装技能，且独立委派本身不是要求，则保留 `next_action: route_skill`。

### 对抗性审查

```yaml
task_id: "adversarial-review-1"
actor_id: "adversarial-reviewer-a"
role: adversarial_reviewer
phase: review
objective: "在接受前攻击具体方案、设计、实现方法、diff、PR、技能、提示词或 agent 运行。"
required_skills:
  - name: none
    source: none
    path: ""
    required: false
    reason: "审查依赖控制面参考，而不是独立技能。"
required_references:
  - path: /home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md
    required: true
    reason: "专家必须使用基于标准的批判、红队区分和预演失败分析结构。"
required_mcp:
  - name: GitHub
    required: false
    reason: "当被审制品是 PR、issue、提交或审查线程时使用。"
required_tools:
  - name: rg
    required: false
    reason: "当本地源代码、diff、日志或追踪需要定向证据检查时使用。"
edits_allowed: false
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: adversarial_review
  required_fields:
    - review_criteria
    - valid_failures
    - weak_or_irrelevant_attacks
    - mitigations
    - residual_risk
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "制品过于含糊，无法基于标准进行攻击。"
  - "无法访问所需证据。"
  - "审查需要编辑而不是批判。"
```

### 常规实现

```yaml
task_id: "implementation-1"
actor_id: "implementation-actor-a"
role: implementation_worker
phase: implementation
objective: "实现已确认的代码变更并运行验证。"
required_skills:
  - name: coding-project
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md
    required: true
    reason: "必须根据语言、项目约定和验证规则更改现有仓库代码。"
required_references: []
required_mcp: []
required_tools:
  - name: project test/build commands
    required: true
    reason: "验证必须使用目标项目自己的工具。"
edits_allowed: true
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: implementation_report
  required_fields:
    - changed_files
    - artifact_version
    - review_risk_tags
    - validation_commands
    - validation_results
    - residual_risks
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "所需技能或验证工具不可用。"
  - "变更需要所有权范围之外的文件。"
  - "契约外出现架构、模式、迁移、认证、支付、部署或用户可见行为风险。"
```

### TDD 实现

```yaml
task_id: "implementation-1"
actor_id: "implementation-actor-a"
role: tdd_worker
phase: implementation
objective: "使用 TDD 完成一个可见行为或小型模块。"
required_skills:
  - name: coding-tdd
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md
    required: true
    reason: "用户要求测试优先或红-绿-重构；必须保护 TDD 循环。"
required_references: []
required_mcp: []
required_tools:
  - name: project test commands
    required: true
    reason: "红-绿-重构循环需要可执行测试。"
edits_allowed: true
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: tdd_report
  required_fields:
    - failing_test
    - implementation_slice
    - artifact_version
    - review_risk_tags
    - green_validation
    - refactor_validation
    - final_entry_to_output_check
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
stop_if:
  - "无法为预期行为使测试变红。"
  - "实现范围扩展到一个行为切片之外。"
  - "所需验证工具不可用。"
```

## 并行化规则

- `grilling` 只能串行。它一次提出一个问题。
- `diagnosing-problem` 可以与只读证据收集并行运行，但主 agent 必须合并最终问题框架和交接。
- `exploring-project` 可以在不同模块间并行化只读探索；它不得写入文件。
- `reviewing-code` 可以跨代码和安全包并行化只读审查通道；它不得写入文件。
- 强制审查仅在被审实现或修复达到终态后开始，且其 actor 必须独立于产生该版本的 actor。
- `coding-project` 仅当所有权路径和逻辑子系统不重叠时才可并行写入。
- `coding-tdd` 只能并行处理独立函数或模块；共享 API、模式、生成制品、迁移和最终集成保持串行。

## 协调统一检查清单

接受专家输出前，检查：

- `skills_loaded` 包含任务契约中每个 `required: true` 技能。
- `references_loaded` 包含任务契约中每个 `required: true` 参考。
- `mcp_used` 和 `tools_used` 包含任务契约中每个 `required: true` 能力。
- 已报告 `skill_instructions_followed`。
- 所有 `deviations` 都有合理解释，且不会破坏任务目标。
- 输出包含每个 `expected_output.required_fields` 项。
- 遵守了写入所有权。
- 验证结果来自命令、测试、构建、用户验收或可核查证据。
- 终态实现和修复报告包含 `artifact_version` 和 `review_risk_tags`，且每个匹配的强制触发项都创建了审查任务。
- 审查 actor 身份不同于实现或修复固定目标版本的 actor。
- 审查结果目标与当前制品版本完全匹配；制品变化会使先前审查失效。
- 阻断性发现会阻止接受并派发修复工作，并且每个已完成修复之后都进行重新审查。

如果任一检查失败，不要直接接受结果；请求补全、重新运行或独立验证。
