---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---
# 编排状态

当工作需要多个代理、后台任务、并行编码通道或阶段性协调时，使用编排状态。

这不是第五个控制面。它是让路由、委派、所有权、持久化和反思保持一致的运行时层。

## 调度优先

编排者默认不是实现 worker。

工作开始前，决定编排者应该：

- 询问阻塞问题
- 读取最小路由上下文
- 委派发现、调研、实现、设计、评审或媒体分析
- 直接运行最终验证
- 综合已终止 specialist 的输出

对于 Large 实现工作，编排者默认不得写实现文件。只有在说明为什么直接实现比委派更安全之后，才可以这样做。

完成标准：编排者拥有协调和验证；当委派能增加明确价值时，specialist 拥有有边界的工作。

## 依赖图

委派前，识别：

- 现在可以运行的独立任务
- 必须等待的依赖任务
- 每个任务拥有的文件、文件夹或子系统
- 实现前需要的输出
- 最终响应前需要的验证

完成标准：没有任务在不知道自己是独立、依赖还是阻塞状态的情况下被委派。

## 任务契约

每个被委派的任务都必须自包含：

```yaml
role: ""
phase: context | design | implementation | review | verification
objective: ""
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

完成标准：specialist 不需要猜测角色、阶段、范围、权限、required skills、required references、required MCP/tools、预期输出、停止条件或验证。

## Skill 路由

当 specialist 应该使用某个 skill 时，在 `required_skills` 中声明；不要依赖隐含的自然语言提示。默认 skill 路由图见 [`skill-orchestration.md`](skill-orchestration.md)。

规则：

- 当 skill 已经在当前环境的 skill 列表中可见时，使用 `available_skill`。
- 当 specialist 必须先读取某个具体 `SKILL.md` 路径时，使用 `file_path`。
- 当 skill 捆绑在目标仓库中时，使用 `repo_skill`。
- 只有没有专门 skill 需要时，才使用 `none`。
- 只有跳过该 skill 会实质改变结果时，才标记 `required: true`。
- 要求 specialist 报告 `skills_loaded`、是否遵循了说明，以及任何偏离。

完成标准：每个依赖专门流程的委派任务，都命名 skill、来源、为什么需要，以及编排者如何确认它被使用。

## Reference 路由

当 specialist 依赖控制面文件、设计指南、项目指南、ADR、schema 文档或其他非 skill reference 时，在 `required_references` 中声明。

规则：

- 对必须读取但不是独立 skill 的文件使用 `required_references`。
- 当跳过该 reference 会实质改变结果时，标记 `required: true`。
- 要求 specialist 报告 `references_loaded` 和偏离。
- 如果 required reference 不可用，specialist 必须停止，而不是凭记忆重建。

完成标准：每个依赖非 skill reference 材料的委派任务，都命名文件、为什么需要，以及编排者如何确认它被使用。

## MCP 和工具路由

当 specialist 依赖 connector、MCP server、代码导航工具、搜索工具、浏览器、构建工具或框架命令时，在 `required_mcp` 或 `required_tools` 中声明。

规则：

- 对 GitHub、CodeMap、Graphify、web reader、issue tracker、数据库工具等命名 MCP/connector 使用 `required_mcp`。
- 对 shell 命令、语言工具、测试 runner、formatter、包管理器、浏览器或本地 CLI 使用 `required_tools`。
- 当跳过该能力会实质影响正确性时，标记 `required: true`。
- 要求 specialist 报告 `mcp_used`、`tools_used` 和偏离。
- 如果 required 能力不可用，specialist 必须停止，而不是替换成未经批准的路径。

完成标准：每个依赖外部能力的委派任务，都命名该能力、为什么需要，以及编排者如何确认它被使用。

## 所有权边界

同一时间只有一个可写 worker 可以拥有一个文件或子系统。

规则：

- 只有路径不重叠时，才允许并行写任务。
- 只读发现可以和大多数工作并行。
- 评审任务必须等待被评审的工作进入终止状态。
- 修改共享组件的 UI 工作不得与这些组件上的实现工作重叠。
- 取消一个 writer 不是回滚；替换前要检查和协调部分改动。

完成标准：没有两个运行中的写任务可以修改同一个文件、文件夹或逻辑子系统。

## 持久状态

用一个小任务板追踪委派工作：

```yaml
tasks:
  - id: ""
    specialist: ""
    phase: ""
    objective: ""
    state: running # running | completed | error | cancelled | timed_out
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

对于长任务或高风险工作，将状态持久化到项目本地 markdown 文件或任务产物。包含：

- 目标
- 约束
- 阶段计划
- 任务板
- 决策轨迹
- 评审门
- 验证日志
- 未解决风险

完成标准：下一轮可以从显式状态恢复，而不是依赖散落的对话记忆。

## 协调

specialist 输出是输入，不是最终真相。

任务完成后：

1. 对照原始用户目标比较结果。
2. 检查与其他任务输出的冲突。
3. 检查 required skills 是否已加载，以及偏离是否合理。
4. 检查 required references 是否已加载，以及偏离是否合理。
5. 检查 required MCP/tools 是否已使用，以及偏离是否合理。
6. 决定接受、修订、拒绝，或派发后续工作。
7. 更新任务板。
8. 在下一次 handoff 中保留有用决策。

完成标准：最终工作不依赖未评审的 specialist 输出、未验证的 required-skill 使用、未验证的 required-reference 使用，或未验证的 required-capability 使用。

## 保守反思

不要因为单次有趣运行就创建新的 skill、agent、command、rule 或 playbook。

使用这个阈值：

- 一个有用洞察 -> 保存到笔记或最终总结
- 2-3 次相似运行中的重复摩擦 -> 建议项目指导或 CLAUDE.md
- 稳定重复、触发清楚的工作流 -> 建议一个更窄的 skill
- 证据薄弱的推测性改进 -> 什么都不创建

完成标准：流程改进有证据支撑、最小化，并放在能解决问题的最低持久层。

## 验证门

最终响应前：

- 所有 required tasks 都已终止
- 依赖工作已消费它等待的输出
- required skills、references、MCP 和 tools 已确认，或偏离已被显式接受
- 文件所有权冲突已解决
- 相关检查已运行，或跳过的检查已解释
- 残余风险已明确

完成标准：用户收到的是协调后的结果，而不是一堆 agent 报告。
