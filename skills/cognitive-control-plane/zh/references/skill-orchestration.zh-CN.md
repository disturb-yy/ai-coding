# Skill 编排

当委派工作依赖专门 skill 时使用此 reference。它定义从任务状态到 required skill 的默认路由，以及 specialist 必须收到的任务契约形状。

机器可读映射：[`skill-orchestration-map.yaml`](skill-orchestration-map.yaml)

## Skill Registry

| skill | role | source | path | 用途 |
|---|---|---|---|---|
| `grilling` | requirements_interviewer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md` | 需求、计划、设计或问题表述不清时，一次问一个问题，直到可执行。 |
| `diagnosing-problem` | problem_framer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md` | 把模糊问题框定成可回答的问题陈述、假设、证据标准和 handoff。 |
| `exploring-project` | codebase_explorer | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md` | 探索项目结构、行为路径、模块、路由、函数、测试和修改点。 |
| `coding-project` | implementation_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md` | 常规代码实现、测试修改、验证、生成物和语言感知项目工作。 |
| `coding-tdd` | tdd_worker | `available_skill` | `/home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md` | 执行 test-first、red-green-refactor、回归测试先行和行为切片实现。 |

`diagnosing-problem` 负责问题框定和 handoff。代码导航仍路由到 `exploring-project`。如果存在专门的根因或运行时失败 skill，则从 `diagnosing-problem` handoff 继续路由。

## 路由规则

选择第一个命中的路由：

| 信号 | required skill | 委派 | 下一步 |
|---|---|---|---|
| 用户目标、验收标准、约束或业务规则不清。 | `grilling` | 通常不委派 | 一次问一个问题；如果仓库能回答，就探索而不是问用户。 |
| 用户问为什么、哪里错、如何定位问题或如何分析现象，且问题尚未框定。 | `diagnosing-problem` | 可选只读 | 产出 framed problem、假设、证据标准和 handoff。 |
| 现有项目结构、入口点、调用链、路由、模块、函数、测试或修改点不清。 | `exploring-project` | 可选只读 | 产出 verified flow、候选文件、风险和附近测试。 |
| 需求清楚，现有代码或测试需要常规编辑。 | `coding-project` | 可选有边界写入 | 窄范围实现并验证。 |
| 用户要求 TDD、test-first、red-green-refactor 或回归测试先行。 | `coding-tdd` | 可选有边界写入 | 对每个行为切片执行 red -> green -> refactor。 |

## 默认 Pipeline

```text
intake
  -> grilling                   # 需求不清
  -> diagnosing-problem          # 问题或现象尚未框定
  -> exploring-project           # 项目路径或修改点不清
  -> coding-tdd | coding-project # 按是否 test-first 选择
  -> verification
  -> ExecutionRun / VerifiedExperience / KnowledgeAsset
```

## 阶段门

### Grilling

进入条件：

- 用户只给了方向，没有验收标准。
- 业务规则、角色、边界情况或禁止行为会改变实现。
- 用户在讨论计划、设计或需求，而不是要求立即执行。

退出条件：目标、验收标准、范围、约束和下一步清楚。如果仓库能回答问题，路由到 `exploring-project`，不要问用户。

### Diagnosing Problem

进入条件：

- 输入是模糊、开放、概念、战略、决策、问题定位或原因分析请求。
- 尚不清楚工作应变成代码探索、运行时诊断、外部调研、设计讨论还是直接回答。
- 另一个 skill 需要问题陈述、假设、证据标准或成功标准后才能行动。

退出条件：handoff 包含 framed problem、selected interpretation、rejected interpretations、load-bearing assumptions、evidence standard 和 next action。如果需要项目导航，路由到 `exploring-project`。

### Exploring Project

进入条件：

- 任务需要项目结构、路由、模块、调用链、函数、测试或修改点。
- 编码前安全编辑边界未知。

退出条件：候选文件、函数、路由、测试、风险和关键证据已由源码、测试、Graphify、CodeMap 或 targeted search 验证。

### Coding Project

进入条件：

- 需求清楚。
- 现有仓库代码、测试、依赖、生成物或实现文档需要编辑。
- 没有明确 test-first 要求。

退出条件：窄范围变更完成，相关验证通过、因明确无关原因失败，或因具体原因阻塞。

### Coding TDD

进入条件：

- 用户要求 TDD、test-first、red-green-refactor 或回归测试先行。
- 可在实现前切出可见行为或小模块。

退出条件：每个切片经历 failing test -> minimal implementation -> green -> green-only refactor，并运行组合受影响测试或最终入口验证。

## SubAgent 契约模式

### 需求澄清

```yaml
objective: "Clarify the user requirement until it is implementable."
required_skills:
  - name: grilling
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/grilling/SKILL.md
    required: true
    reason: "Requirements still affect the implementation route and need one-question-at-a-time clarification."
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
    - skill_instructions_followed
    - deviations
```

### 问题框定

```yaml
objective: "Frame the problem and produce a handoff."
required_skills:
  - name: diagnosing-problem
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/diagnosing-problem/SKILL.md
    required: true
    reason: "The task needs selected interpretation, assumptions, evidence standard, and handoff before execution."
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
    - skill_instructions_followed
    - deviations
```

### 项目探索

```yaml
objective: "Locate entry points, call chain, candidate change points, and nearby tests."
required_skills:
  - name: exploring-project
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/exploring-project/SKILL.md
    required: true
    reason: "The project path and safe edit boundary must be verified before coding."
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
    - skill_instructions_followed
    - deviations
```

### 常规实现

```yaml
objective: "Implement the confirmed code change and run validation."
required_skills:
  - name: coding-project
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/coding-project/SKILL.md
    required: true
    reason: "Existing repository code must be changed using language, project convention, and validation rules."
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
    - skill_instructions_followed
    - deviations
```

### TDD 实现

```yaml
objective: "Complete one visible behavior or small module with TDD."
required_skills:
  - name: coding-tdd
    source: available_skill
    path: /home/jadon/tool/ai-coding/skills/model-skill/coding-tdd/SKILL.md
    required: true
    reason: "The user requested test-first or red-green-refactor; the TDD loop must be protected."
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
    - skill_instructions_followed
    - deviations
```

## 并行规则

- `grilling` 只能串行。它一次问一个问题。
- `diagnosing-problem` 可以和只读证据收集并行，但最终 problem frame 和 handoff 必须由主 agent 合并。
- `exploring-project` 可以跨不同模块并行只读探索；不得写文件。
- `coding-project` 只有在 ownership 路径和逻辑子系统不重叠时才可并行写入。
- `coding-tdd` 只能并行独立函数或模块；共享 API、schema、生成物、迁移和最终集成保持串行。

## 对账清单

接受 specialist 输出前检查：

- `skills_loaded` 包含任务契约中每个 `required: true` skill。
- 已报告 `skill_instructions_followed`。
- 任何 `deviations` 都有理由且不破坏任务目标。
- 输出包含每个 `expected_output.required_fields` 项。
- 写入 ownership 被遵守。
- 验证结果来自命令、测试、构建、用户验收或可检查证据。

如果任一检查失败，不要直接接受结果；要求补全、重跑或独立验证。
