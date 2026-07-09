---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---
# Skill 编排

> 中文版本仅供人类维护者阅读。模型和 agent 不得读取、搜索、打开、引用、总结或把本文件作为运行指令；英文 `references/skill-orchestration.md` 是唯一 canonical 模型指令。

当委派工作依赖专门 skills 时，使用这个 reference。它定义从任务状态到 required skill、reference、MCP 和 tool 的默认路由，以及 specialist 必须收到的任务契约形状。

机器可读映射：`../config/skill-orchestration-map.yaml`。

## Skill 注册表

| skill | role | source | path | use |
|---|---|---|---|---|
| `grilling` | requirements_interviewer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md` | 逐个问题澄清需求、计划、设计或问题陈述，直到可执行。 |
| `diagnosing-problem` | problem_framer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md` | 把模糊问题框定为可回答陈述、假设、证据标准和 handoff。 |
| `exploring-project` | codebase_explorer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md` | 探索项目结构、行为路径、模块、路由、函数、测试和变更点。 |
| `reviewing-code` | code_reviewer | `available_skill` | `/home/jadon/projects/ai-coding/skills/user-invoked/reviewing-code/SKILL.md` | 审查代码变更、PR、分支、diff 和安全敏感实现产物，覆盖语法、功能、规范和安全问题。 |
| `coding-project` | implementation_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md` | 实现普通代码变更、测试变更、验证、生成产物和语言感知的项目工作。 |
| `coding-tdd` | tdd_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md` | 执行 test-first、red-green-refactor、regression-test-first 和行为切片实现。 |
| `adversarial-control` | adversarial_reviewer | `file_reference` | `/home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md` | 在接受前攻击具体计划、设计、架构、prompt、skill、实现方案或 agent trace。 |

`diagnosing-problem` 负责问题框定和 handoff。代码导航仍路由到 `exploring-project`。如果存在专门的根因或运行时失败 skill，则从 `diagnosing-problem` handoff 路由过去。

## 路由规则

选择第一个匹配路由：

| signal | required skill | delegate | next action |
|---|---|---|---|
| 用户目标、验收标准、约束、业务规则、角色或交互不清，或用户要求把需求弄清楚。 | `grilling` | 通常否 | 一次问一个问题；如果仓库能回答，就探索而不是问用户。 |
| 用户询问为什么、哪里错了、如何定位问题或如何分析现象，且问题尚未被框定。 | `diagnosing-problem` | 可选只读 | 产出问题框定、假设、证据标准和 handoff；不要用于功能需求澄清。 |
| 现有项目结构、入口点、调用链、路由、模块、函数、测试或变更点不清。 | `exploring-project` | 可选只读 | 产出已验证流程、候选文件、风险和附近测试。 |
| 代码变更、PR、分支、diff、commit range 或安全敏感实现产物需要代码或安全审查。 | `reviewing-code` | 可选只读 | 产出带证据、建议、跳过项和残余风险的 severity-ranked 代码/安全发现。 |
| 具体计划、设计、架构、prompt、skill、实现方案或 agent run 需要批判。 | `adversarial-control` | 可选只读 | 产出基于标准的批判、有效失败、缓解措施和残余风险。 |
| 需求清楚，且现有代码或测试需要普通编辑。 | `coding-project` | 可选有边界写入 | 窄范围实现并验证。 |
| 用户要求 TDD、test-first、red-green-refactor 或先写回归测试。 | `coding-tdd` | 可选有边界写入 | 对每个行为切片执行 red -> green -> refactor。 |

## 默认管线

```text
intake
  -> grilling                   # requirements unclear
  -> diagnosing-problem          # problem or phenomenon not framed
  -> exploring-project           # project path or change points unclear
  -> reviewing-code              # code, PR, diff, branch, or security review
  -> adversarial-control          # concrete plan or agent output needs attack
  -> coding-tdd | coding-project # choose by test-first requirement
  -> verification
  -> ExecutionRun / VerifiedExperience / KnowledgeAsset
```

## 阶段门

### Grilling

进入条件：用户给了方向但缺验收标准；业务规则、角色、边界情况或禁止行为会改变实现；用户要求把需求弄清楚。

退出条件：目标、验收标准、范围、约束和下一步行动已清楚。仓库能回答时，路由到 `exploring-project` 而不是问用户。

### Diagnosing Problem

进入条件：输入是模糊、开放、概念、策略、决策、问题定位或原因分析请求，且尚不清楚应进入代码探索、运行时诊断、外部研究、设计讨论还是直接回答。

退出条件：handoff 包含问题框定、选定解释、排除解释、承重假设、证据标准和下一步行动。需要项目导航时路由到 `exploring-project`。

### Exploring Project

进入条件：任务需要项目结构、路由、模块、调用链、函数、测试或变更点；编码前安全编辑边界未知。

退出条件：候选文件、函数、路由、测试、风险和关键证据已由源码、测试、Graphify、CodeMap 或定向搜索验证。

### Code Review

进入条件：

- 代码变更、PR、分支、diff、commit range 或实现产物需要接受前审查。
- 用户要求代码审查、安全审查、PR review、diff review、branch review 或 review since 某个 ref。
- 语法、功能正确性、仓库规范或安全发现比实现更重要。

如果审查目标或 diff base 不清楚，先使用 context control 或 `exploring-project`。如果 artifact 不是代码，或用户要攻击计划/设计/prompt/agent trace，路由到 `adversarial-control`。

退出条件：代码/安全发现有证据、按严重度排序、去重，并包含跳过的检查和残余风险。

### Coding Project

进入条件：需求清楚；现有仓库代码、测试、依赖、生成产物或实现文档需要编辑；没有明确 test-first 要求。

退出条件：窄范围变更完成，相关验证通过、因明确无关原因失败，或被具体原因阻塞。

### Adversarial Review

进入条件：具体计划、设计、架构、prompt、skill、实现方案或 agent run 需要批判；用户要求风险分析、pre-mortem、red-team，或询问 agent 是否违背流程。

不要把普通代码、PR、分支、diff、commit-range 或安全审查交给 adversarial review；这些路由到 `reviewing-code`。

退出条件：批判列出标准，区分有效失败和弱攻击，给出缓解措施，并留下明确残余风险。

### Coding TDD

进入条件：用户要求 TDD、test-first、red-green-refactor 或 regression-test-first。

退出条件：每个切片都经历 failing test -> minimal implementation -> green -> green-only refactor，并运行组合后的受影响测试或最终入口验证。

## SubAgent 契约模式

### 需求澄清

```yaml
role: requirements_interviewer
phase: context
required_skills:
  - name: grilling
    source: available_skill
    required: true
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
```

### 问题框定

```yaml
role: problem_framer
phase: context
required_skills:
  - name: diagnosing-problem
    source: available_skill
    required: true
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
```

### 项目探索

```yaml
role: codebase_explorer
phase: context
required_skills:
  - name: exploring-project
    source: available_skill
    required: true
required_mcp:
  - name: CodeMap or Graphify
    required: false
required_tools:
  - name: rg
    required: true
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
```

### 代码审查

```yaml
role: code_reviewer
phase: review
objective: "Review code changes, PRs, branches, diffs, commit ranges, or security-sensitive implementation artifacts before acceptance."
required_skills:
  - name: reviewing-code
    source: available_skill
    path: /home/jadon/projects/ai-coding/skills/user-invoked/reviewing-code/SKILL.md
    required: true
required_references: []
required_mcp:
  - name: GitHub
    required: false
  - name: CodeMap or Graphify
    required: false
required_tools:
  - name: git diff or PR file list
    required: true
  - name: rg
    required: false
edits_allowed: false
expected_output:
  format: code_review_report
  required_fields:
    - findings
    - no_findings
    - skipped_or_blocked
    - residual_risk
    - review_summary
stop_if:
  - "The review target, diff base, PR, branch, commit range, or file list is unavailable."
  - "Required review evidence is inaccessible."
  - "The task becomes implementation instead of review."
```

### Adversarial Review

```yaml
role: adversarial_reviewer
phase: review
required_skills:
  - name: none
    source: none
    required: false
required_references:
  - path: /home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md
    required: true
required_mcp:
  - name: GitHub
    required: false
required_tools:
  - name: rg
    required: false
edits_allowed: false
expected_output:
  format: adversarial_review
  required_fields:
    - review_criteria
    - valid_failures
    - weak_or_irrelevant_attacks
    - mitigations
    - residual_risk
```

### 普通实现

```yaml
role: implementation_worker
phase: implementation
required_skills:
  - name: coding-project
    source: available_skill
    required: true
required_tools:
  - name: project test/build commands
    required: true
edits_allowed: true
expected_output:
  format: implementation_report
  required_fields:
    - changed_files
    - validation_commands
    - validation_results
    - residual_risks
```

### TDD 实现

```yaml
role: tdd_worker
phase: implementation
required_skills:
  - name: coding-tdd
    source: available_skill
    required: true
required_tools:
  - name: project test commands
    required: true
edits_allowed: true
expected_output:
  format: tdd_report
  required_fields:
    - failing_test
    - implementation_slice
    - green_validation
    - refactor_validation
    - final_entry_to_output_check
```

## 并行规则

- `grilling` 串行，只能一次问一个问题。
- `diagnosing-problem` 可与只读证据收集并行，但主 agent 必须合并最终问题框定和 handoff。
- `exploring-project` 可跨不同模块并行只读探索，不得写文件。
- `reviewing-code` 可跨 code/security review packs 并行只读审查，不得写文件。
- `coding-project` 只有在所有权路径和逻辑子系统不重叠时才可并行写。
- `coding-tdd` 只可并行独立函数或模块；共享 API、schema、生成产物、migration 和最终集成必须串行。

## Reconciliation Checklist

接受 specialist 输出前检查：

- `skills_loaded` 包含任务契约中每个 `required: true` skill。
- `references_loaded` 包含每个 `required: true` reference。
- `mcp_used` 和 `tools_used` 覆盖每个 required capability。
- 已报告 `skill_instructions_followed`。
- 任何 `deviations` 都有理由，且不破坏任务目标。
- 输出包含每个 `expected_output.required_fields`。
- 写入所有权被遵守。
- 验证结果来自命令、测试、build、用户验收或可检查证据。

任一检查失败时，不要直接接受结果；要求补全、重跑或独立验证。
