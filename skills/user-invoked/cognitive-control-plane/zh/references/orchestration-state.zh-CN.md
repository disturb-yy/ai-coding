---
access:
  audience: model
  model_read: false
  model_write: true
purpose: skill_reference
---

## 工作项 Scheduler

Scheduler 是宿主侧的持久工作控制过程，而不是 specialist skill 或 Runner。将
`issue`、`request`、`transaction` 和 `ticket` 标准化为工作项；一个 `run` 只是
同一工作项的一次 session 尝试。

启动 run 前，Scheduler 必须确认依赖已满足、取得唯一有效租约、解决重叠写入所有权、
分配单次预算，并带上先前 checkpoint 创建自包含契约。状态为：

```text
工作项活动态：ready -> leased -> running -> validating
工作项终态：resolved | concluded | duplicate | blocked | escalated | cancelled
run：scheduled | leased | running | checkpointed | completed | expired | cancelled
```

40% token 时 checkpoint；45% 后停止扩张范围；50% 强制结束本次 run。中断、
checkpoint 或租约过期时，Scheduler 为同一工作项创建下一 attempt，而非复制工作项。
`resolved` 需要验证，`concluded` 需要证据，`blocked`/`escalated` 在外部条件改变前
不得自动重试；写入型 `transaction` 还需要幂等键。
# 编排状态

当工作需要多个 agent、后台任务、并行编码通道或分阶段协调统一时，使用编排状态。

这不是第五个控制平面。它是保持路由、委派、所有权、持久化和反思一致的运行时层。

## 调度优先

编排者默认不是实现工作者。

工作开始前，决定编排者是否应：

- 提出阻断性问题
- 阅读最少量的路由上下文
- 委派发现、研究、实现、设计、审查或媒体分析
- 直接运行最终验证
- 综合终态专家输出

对于 Large 实现工作，编排者必须委派实现。仅当实现防护中的每个条件都满足时，才允许直接实现；否则必须委派。

完成标准：编排者负责协调和验证；当委派具有明确价值时，专家负责有边界的工作。

## 只读探索策略

对于 Large 仓库发现，在广泛阅读源代码前决定由编排者运行**直接最小化探索**，还是委派一个或多个**只读探索**任务。`exploring-project` 命名所需流程；它不会自动要求使用 subagent。

仅对小型依赖证据链使用直接最小化探索。记录有边界的搜索范围、检查过的来源，以及为何单独工作者或并行通道不会实质改善下一项决策。当搜索跨越独立区域、能从独立证据报告中获益，或范围足够大以至于并行通道能降低不确定性时，委派只读探索。不要声称已路由到 `exploring-project`，却悄然跳过其流程。

完成标准：任务板或追踪记录 `direct_minimal_exploration` 或 `delegated_read_only_exploration`、证据范围和选择理由。

## 依赖图

委派前，识别：

- 现在可以运行的独立任务
- 必须等待的依赖任务
- 每个任务拥有的文件、文件夹或子系统
- 实现前所需的输出
- 最终响应前所需的验证

完成标准：在不知道任务是独立、依赖还是被阻塞时，不委派该任务。

## 任务契约

每个委派任务都必须自包含：

```yaml
task_id: ""
actor_id: ""
role: ""
phase: context | design | implementation | review | verification
objective: ""
review_of_task_id: "" # review 阶段必需
review_of_actor_id: "" # review 阶段必需
review_iteration: 0
supersedes_review_task_id: ""
review_target: # review 阶段必需
  kind: none # none | git_range | stable_artifact
  base_sha: ""
  head_sha: ""
  diff_hash: ""
  stable_id: ""
constraints: []
required_skills:
  - name: ""
    source: available_skill # available_skill | file_path | repo_skill | none
    path: ""
    required: true
    reason: ""
required_references:
  - path: ""
    required: false
    reason: ""
required_mcp:
  - name: ""
    required: false
    reason: ""
required_tools:
  - name: ""
    required: false
    reason: ""
search_scope: []
ownership:
  writable_paths: []
  read_only_paths: []
  forbidden_paths: []
edits_allowed: false
expected_output:
  format: ""
  required_fields: []
  must_report:
    - skills_loaded
    - references_loaded
    - mcp_used
    - tools_used
    - skill_instructions_followed
    - deviations
validation: []
stop_if: []
```

完成标准：专家无需猜测任务或 actor 身份、角色、阶段、范围、权限、所需技能、所需参考、所需 MCP/工具、审查谱系和目标、预期输出、停止条件或验证即可工作。

## 技能路由

当专家应使用某项技能时，在 `required_skills` 中声明；不要依赖隐含的自然语言提示。默认技能路由映射见 [`skill-orchestration.md`](skill-orchestration.md)。

规则：

- 当技能已在当前环境的技能列表中可见时，使用 `available_skill`。
- 当专家必须在行动前阅读特定 `SKILL.md` 路径时，使用 `file_path`。
- 当技能捆绑在目标仓库内时，使用 `repo_skill`。
- 仅当不需要专用技能时使用 `none`。
- 仅当跳过该技能会实质改变结果时，才标记 `required: true`。
- 要求专家报告 `skills_loaded`、是否遵循了指令以及任何偏离。

完成标准：每个依赖专用流程的委派任务都命名技能、其来源、需要它的原因，以及编排者如何确认它已被使用。

## 参考路由

当专家依赖控制面文件、设计指南、项目指南、ADR、模式文档或其他非技能参考时，在 `required_references` 中声明。

规则：

- 对必须读取但不是独立技能的文件使用 `required_references`。
- 当跳过该参考会实质改变结果时，标记 `required: true`。
- 要求专家报告 `references_loaded` 和偏离。
- 如果必需参考不可用，专家必须停止，而不是根据记忆重建。

完成标准：每个依赖非技能参考材料的委派任务都命名文件、需要它的原因，以及编排者如何确认它已被使用。

## MCP 和工具路由

当专家依赖连接器、MCP 服务器、代码导航工具、搜索工具、浏览器、构建工具或框架命令时，在 `required_mcp` 或 `required_tools` 中声明。

规则：

- 对 GitHub、CodeMap、Graphify、网页读取器、问题跟踪器或数据库工具等具名 MCP/连接器使用 `required_mcp`。
- 对 shell 命令、语言工具、测试运行器、格式化器、包管理器、浏览器或本地 CLI 使用 `required_tools`。
- 当跳过该能力会实质改变正确性时，标记 `required: true`。
- 要求专家报告 `mcp_used`、`tools_used` 和偏离。
- 如果必需能力不可用，专家必须停止，而不是替换为未经批准的路径。

完成标准：每个依赖外部能力的委派任务都命名该能力、需要它的原因，以及编排者如何确认它已被使用。

## 所有权边界

同一时间只能有一个可写工作者拥有某个文件或子系统。

规则：

- 仅当路径不重叠时，才允许并行写入任务。
- 如果用户要求可写工作者重叠，应拒绝并行写入、串行化任务，并在委派前解决冲突。
- 只读发现可以与大多数工作并行运行。
- 审查任务必须等待其所审查的工作达到终态。
- 审查 actor 必须不同于实现或修复被审版本的 actor。新的任务 ID 或角色名称不会让同一 actor 具有独立性。
- 审查工作者必须保持只读，也不得拥有针对其发现的修复任务。
- 修改共享组件的 UI 工作不得与这些组件的实现工作重叠。
- 取消写入者不等于回滚；在替换前检查并协调部分变更。

完成标准：没有两个运行中的写入任务可以修改同一文件、文件夹或逻辑子系统，并且不存在被接受的自我审查。

## 审查者强制循环

每个实现或修复任务变为终态后，阅读 [`reviewer-enforcement.md`](reviewer-enforcement.md)，并针对以下强制审查标签评估实际制品：

- `security_sensitive`
- `cross_module_change`
- `public_api_change`
- `schema_change`
- `migration`
- `auth_or_permission_change`
- `deployment_or_rollback_critical`

只要任一标签适用，就预检 `reviewing-code` 和独立审查者 actor。可用时，创建依赖的只读 `reviewing-code` 任务。如果该技能不可用，仅当宿主能够启动一个不同的只读 actor，使其加载 [`reviewer-enforcement.md`](reviewer-enforcement.md) 并将不可用技能报告为偏离时，才使用 `independent_read_only_reviewer` 回退。否则保持审查关卡阻塞并发出交接；绝不要用自我审查替代。记录不同的实现者和审查者 `actor_id` 值，并使用 `base_sha`、`head_sha` 和 `diff_hash` 固定 `review_target`，或使用等效的不可变 `stable_id`。

推进关卡前协调统一发现：

- 当前固定版本没有阻断性发现 -> 将审查关卡标记为 `cleared`
- 存在阻断性发现 -> 将其标记为 `blocked`，阻止最终接受，并派发依赖的修复任务
- 修复完成或发生任何其他制品变化 -> 将先前审查标记为 `invalidated`，计算新版本，并创建新的独立审查任务
- 明确终止循环 -> 将其标记为 `terminated`，并报告包含残余风险的未接受结果

重复执行，直到最新制品版本经过独立审查且没有阻断性发现。测试和验证补充审查；它们不能免除强制审查。

完成标准：接受前，当前制品版本的每个强制审查关卡都已放行；或者运行明确以已终止且未接受的状态结束。

## 持久状态

将委派工作作为小型任务板进行跟踪：

```yaml
tasks:
  - id: ""
    actor_id: ""
    specialist: ""
    phase: ""
    objective: ""
    state: pending # pending | running | completed | error | cancelled | timed_out
    required_skills: []
    skills_confirmed: []
    required_references: []
    references_confirmed: []
    required_mcp: []
    mcp_confirmed: []
    required_tools: []
    tools_confirmed: []
    ownership:
      files: []
      areas: []
    dependencies: []
    risk_tags: []
    review_required: false
    review_of_task_id: ""
    review_of_actor_id: ""
    review_iteration: 0
    review_status: not_required # not_required | pending | running | blocked | cleared | invalidated | terminated
    review_target:
      kind: none # none | git_range | stable_artifact
      base_sha: ""
      head_sha: ""
      diff_hash: ""
      stable_id: ""
    blocking_finding_ids: []
    supersedes_review_task_id: ""
    result: ""
event_log:
  - timestamp: ""
    actor: "" # orchestrator | specialist role
    task_id: ""
    event_type: started # started | blocked | completed | decision | validation | handoff
    summary: ""
    evidence_refs: []
    artifact_refs: []
    next_action: ""
```

对于长时间或高风险工作，将状态持久化到项目本地 Markdown 文件或任务制品。包括：

- 目标
- 约束
- 阶段计划
- 任务板
- 决策追踪
- 审查关卡
- 事件日志
- 验证日志
- 未解决风险

完成标准：下一轮可以从明确状态恢复，而不是依赖散落的对话记忆。

## 协调统一

专家输出是输入，而不是最终事实。

任务完成时：

1. 将结果与原始用户目标比较。
2. 检查与其他任务输出的冲突。
3. 检查是否加载了所需技能，以及偏离是否合理。
4. 检查是否加载了所需参考，以及偏离是否合理。
5. 检查是否使用了所需 MCP/工具，以及偏离是否合理。
6. 对于实现和修复结果，根据交付制品评估强制审查风险。
7. 对于审查结果，在使用发现前确认 actor 独立性和精确的制品版本匹配。
8. 如果仍有阻断性发现，则阻止接受并派发修复；如果修复改变了制品，则使先前审查失效并派发重新审查。
9. 决定接受、修订、拒绝、派发后续工作，还是终止且不接受。
10. 更新任务板和审查关卡。
11. 在下一次交接中保留有用决策。

完成标准：最终工作不依赖未经审查的专家输出、自我审查、过期审查、未解决的阻断性发现、未经验证的所需技能使用、未经验证的所需参考使用或未经验证的所需能力使用。

## 保守反思

不要根据一次有趣的运行创建新的技能、agent、命令、规则或操作手册。

使用以下阈值：

- 一个有用洞见 -> 保存到笔记或最终摘要
- 在 2-3 次相似运行中反复出现摩擦 -> 建议项目指南或 CLAUDE.md
- 具有明确触发条件的稳定重复工作流 -> 建议一个窄技能
- 证据薄弱的推测性改进 -> 什么也不创建

完成标准：流程改进有证据支持、最小化，并放置在能够解决问题的最低持久层。

## 验证关卡

最终响应前：

- 所有必需任务均处于终态
- 依赖工作已使用其等待的输出
- 所需技能、参考、MCP 和工具已确认，或偏离已被明确接受
- 文件所有权冲突已解决
- 每个实现制品都已按照强制审查触发项接受评估
- 每个强制审查者 actor 都不同于实现或修复被审版本的 actor
- 最新制品版本与有效的固定审查目标完全匹配
- 当前制品的强制审查关卡为 `cleared`；历史上被取代的审查记录可以保留为 `invalidated`
- 每个阻断性发现的修复之后都进行了重新审查，且最新有效审查没有阻断性发现
- 已运行相关检查，或已解释跳过的检查
- 残余风险已明确

完成标准：用户收到的是协调统一后的结果，而不是一堆 agent 报告。
