---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---

# Skill 编排

当委派工作依赖专门 skills 时，使用这个 reference。它定义从任务状态到 required skill、reference、MCP 和 tool 的默认路由，以及 specialist 必须收到的任务契约形状。

机器可读映射：[`skill-orchestration-map.yaml`](skill-orchestration-map.yaml)

## Skill 注册表

| skill | role | source | path | use |
|---|---|---|---|---|
| `grilling` | requirements_interviewer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md` | 逐个问题澄清需求、计划、设计或问题陈述，直到可执行。 |
| `diagnosing-problem` | problem_framer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md` | 把模糊问题框定为可回答陈述、假设、证据标准和 handoff。 |
| `exploring-project` | codebase_explorer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md` | 探索项目结构、行为路径、模块、路由、函数、测试和变更点。 |
| `coding-project` | implementation_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md` | 实现普通代码变更、测试变更、验证、生成产物和语言感知的项目工作。 |
| `coding-tdd` | tdd_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md` | 执行 test-first、red-green-refactor、regression-test-first 和行为切片实现。 |
| `adversarial-control` | adversarial_reviewer | `file_reference` | `/home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md` | 在接受前攻击具体计划、设计、架构、提示词、skill、实现方案或 agent trace。 |

`diagnosing-problem` 负责问题框定和 handoff。代码导航仍路由到 `exploring-project`。如果存在专门的根因或运行时失败 skill，则从 `diagnosing-problem` handoff 路由过去。

## 路由规则

选择第一个匹配路由：

| signal | required skill | delegate | next action |
|---|---|---|---|
| 用户目标、验收标准、约束或业务规则不清。 | `grilling` | 通常否 | 一次问一个问题；如果仓库能回答，就探索而不是问用户。 |
| 用户询问为什么、哪里错了、如何定位问题或如何分析现象，且问题尚未被框定。 | `diagnosing-problem` | 可选只读 | 产出问题框定、假设、证据标准和 handoff。 |
| 现有项目结构、入口点、调用链、路由、模块、函数、测试或变更点不清。 | `exploring-project` | 可选只读 | 产出已验证流程、候选文件、风险和附近测试。 |
| 具体计划、设计、架构、提示词、skill、实现方案、diff、PR 或 agent run 需要批判。 | `adversarial-control` | 可选只读 | 产出基于标准的批判、有效失败、缓解措施和残余风险。 |
| 需求清楚，且现有代码或测试需要普通编辑。 | `coding-project` | 可选有边界写入 | 窄范围实现并验证。 |
| 用户要求 TDD、test-first、red-green-refactor 或先写回归测试。 | `coding-tdd` | 可选有边界写入 | 对每个行为切片执行 red -> green -> refactor。 |

## 默认管线

```text
intake
  -> grilling                   # requirements unclear
  -> diagnosing-problem          # problem or phenomenon not framed
  -> exploring-project           # project path or change points unclear
  -> adversarial-control          # concrete plan or agent output needs attack
  -> coding-tdd | coding-project # choose by test-first requirement
  -> verification
  -> ExecutionRun / VerifiedExperience / KnowledgeAsset
```

## 阶段门

### Grilling

进入条件：

- 用户给了方向但没有验收标准。
- 业务规则、角色、边界情况或禁止行为会改变实现。
- 用户在讨论计划、设计或需求，而不是要求立即执行。

退出条件：目标、验收标准、范围、约束和下一步行动清楚。如果仓库能回答问题，路由到 `exploring-project` 而不是问用户。

### Diagnosing Problem

进入条件：

- 输入是模糊、开放、概念、战略、决策、问题定位或原因分析请求。
- 尚不清楚工作应该进入代码探索、运行时诊断、外部调研、设计讨论还是直接回答。
- 另一个 skill 在行动前需要问题陈述、假设、证据标准或成功标准。

退出条件：handoff 包含框定问题、选定解释、被拒解释、承重假设、证据标准和下一步行动。如果需要项目导航，路由到 `exploring-project`。

### Exploring Project

进入条件：

- 任务需要项目结构、路由、模块、调用链、函数、测试或变更点。
- 编码前安全编辑边界未知。

退出条件：候选文件、函数、路由、测试、风险和关键证据已由源码、测试、Graphify、CodeMap 或定向搜索验证。

### Adversarial Review

进入条件：

- 具体计划、设计、架构、提示词、skill、实现方案、diff、PR 或 agent run 需要批判。
- 用户要求 review、风险分析、pre-mortem、red-team，或询问 agent 是否违背流程。
- 失败模式比产出新实现更重要。

退出条件：批判命名了标准，分离有效失败和弱攻击，给出缓解措施，并留下明确残余风险。

### Coding Project

进入条件：

- 需求清楚。
- 现有仓库代码、测试、依赖、生成产物或实现文档需要编辑。
- 没有显式 test-first 要求。

退出条件：窄范围变更完成，相关验证通过、因具体无关原因失败，或因具体原因被阻塞。

### Coding TDD

进入条件：

- 用户要求 TDD、test-first、red-green-refactor 或 regression-test-first。
- 一个可见行为或小模块可以先切片再实现。

退出条件：每个切片都完成 failing test -> minimal implementation -> green -> green-only refactor，且受影响测试或最终入口验证已运行。

## SubAgent 契约模式

### 需求澄清

```yaml
role: requirements_interviewer
phase: context
objective: "Clarify the user requirement until it is implementable."
required_skills:
  - name: grilling
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md
    required: true
    reason: "Requirements still affect the implementation route and need one-question-at-a-time clarification."
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
  - "The repository or source artifacts can answer the question better than the user."
```

### 问题框定

```yaml
role: problem_framer
phase: context
objective: "Frame the problem and produce a handoff."
required_skills:
  - name: diagnosing-problem
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md
    required: true
    reason: "The task needs selected interpretation, assumptions, evidence standard, and handoff before execution."
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
  - "The task becomes concrete implementation before the problem frame is accepted."
```

### 项目探索

```yaml
role: codebase_explorer
phase: context
objective: "Locate entry points, call chain, candidate change points, and nearby tests."
required_skills:
  - name: exploring-project
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md
    required: true
    reason: "The project path and safe edit boundary must be verified before coding."
required_references: []
required_mcp:
  - name: CodeMap or Graphify
    required: false
    reason: "Use when available for architecture, call-chain, or cross-area navigation."
required_tools:
  - name: rg
    required: true
    reason: "Fast source search is required to verify candidate files and flows."
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
  - "Implementation is requested before candidate files, flows, risks, and tests are verified."
```

### 对抗评审

```yaml
role: adversarial_reviewer
phase: review
objective: "Attack the concrete plan, design, implementation approach, diff, PR, skill, prompt, or agent run before acceptance."
required_skills:
  - name: none
    source: none
    path: ""
    required: false
    reason: "The review depends on a control-surface reference rather than a standalone skill."
required_references:
  - path: /home/jadon/tool/ai-coding/skills/user-skill/cognitive-control-plane/references/adversarial-control.md
    required: true
    reason: "The specialist must use criterion-based critique, red-team separation, and pre-mortem structure."
required_mcp:
  - name: GitHub
    required: false
    reason: "Use when the artifact under review is a PR, issue, commit, or review thread."
required_tools:
  - name: rg
    required: false
    reason: "Use when local source, diffs, logs, or traces need targeted evidence checks."
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
  - "The artifact is too vague to attack with criteria."
  - "Required evidence is inaccessible."
  - "The review would require editing instead of critique."
```

### 普通实现

```yaml
role: implementation_worker
phase: implementation
objective: "Implement the confirmed code change and run validation."
required_skills:
  - name: coding-project
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md
    required: true
    reason: "Existing repository code must be changed using language, project convention, and validation rules."
required_references: []
required_mcp: []
required_tools:
  - name: project test/build commands
    required: true
    reason: "Validation must use the target project's own tooling."
edits_allowed: true
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
expected_output:
  format: implementation_report
  required_fields:
    - changed_files
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
  - "Required skills or validation tools are unavailable."
  - "The change requires files outside ownership."
  - "Architecture, schema, migration, auth, payment, deployment, or user-visible behavior risk appears outside the contract."
```

### TDD 实现

```yaml
role: tdd_worker
phase: implementation
objective: "Complete one visible behavior or small module with TDD."
required_skills:
  - name: coding-tdd
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md
    required: true
    reason: "The user requested test-first or red-green-refactor; the TDD loop must be protected."
required_references: []
required_mcp: []
required_tools:
  - name: project test commands
    required: true
    reason: "The red-green-refactor loop needs executable tests."
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
  - "A failing test cannot be made red for the intended behavior."
  - "The implementation scope expands beyond one behavior slice."
  - "Required validation tooling is unavailable."
```

## 并行规则

- `grilling` 只能串行。它一次问一个问题。
- `diagnosing-problem` 可以和只读证据收集并行，但主 agent 必须合并最终问题框定和 handoff。
- `exploring-project` 可以跨不同模块并行只读探索；它不得写文件。
- `adversarial-control` 是只读评审，必须等被评审产物达到可评审状态。
- `coding-project` 只有在所有权路径和逻辑子系统不重叠时才可以并行写。
- `coding-tdd` 只能并行独立函数或模块；共享 API、schema、生成产物、migration 和最终集成保持串行。

## 协调检查清单

接受 specialist 输出前，检查：

- `skills_loaded` 包含任务契约中每个 `required: true` 的 skill。
- `references_loaded` 包含任务契约中每个 `required: true` 的 reference。
- `mcp_used` 和 `tools_used` 包含任务契约中每个 `required: true` 的能力。
- `skill_instructions_followed` 已报告。
- 任何 `deviations` 都有理由，且不破坏任务目标。
- 输出包含每个 `expected_output.required_fields` 项。
- 写所有权已被遵守。
- 验证结果来自命令、测试、构建、用户验收或可检查证据。

如果任何检查失败，不要直接接受结果；要求补全、重跑，或独立验证。
