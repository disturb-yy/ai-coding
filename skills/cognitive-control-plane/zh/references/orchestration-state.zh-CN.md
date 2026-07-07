# 编排状态

当工作需要多个代理、后台任务、并行编码通道或阶段性对账时，使用编排状态。

这不是第五个控制面。它是让路由、委派、所有权、持久状态和反思保持一致的运行时层。

## Scheduler-First

编排者默认不是实现 worker。

工作开始前，决定编排者应该：

- 询问阻塞性问题
- 读取最小路由上下文
- 委派发现、调研、实现、设计、评审或媒体分析
- 直接运行最终验证
- 综合终态 specialist 输出

完成标准：编排者负责协调和验证；当委派能增加明确价值时，specialist 负责有边界的工作。

## 依赖图

委派前，识别：

- 现在可以运行的独立任务
- 必须等待的依赖任务
- 每个任务拥有的文件、文件夹或子系统
- 实现前所需的输出
- 最终回复前所需的验证

完成标准：没有任务在不知道自己独立、依赖或阻塞状态的情况下被委派。

## 任务契约

每个被委派任务必须自包含：

```yaml
objective: ""
constraints: []
required_skills:
  - name: ""
    source: available_skill # available_skill | file_path | repo_skill | none
    path: ""
    required: true
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
    - skill_instructions_followed
    - deviations
validation: []
what_not_to_do: []
```

完成标准：specialist 不需要猜测范围、权限、必需 skill、预期输出或验证方式。

## Skill 路由

当 specialist 应该使用某个 skill 时，在 `required_skills` 中声明；不要依赖隐含的自然语言提示。

规则：

- 当 skill 已经出现在当前环境的 skill 列表中时，使用 `available_skill`。
- 当 specialist 必须先读取某个具体 `SKILL.md` 路径时，使用 `file_path`。
- 当 skill 捆绑在目标仓库中时，使用 `repo_skill`。
- 只有不需要专门 skill 时才使用 `none`。
- 只有跳过该 skill 会实质改变结果时，才标记 `required: true`。
- 要求 specialist 回报 `skills_loaded`、是否遵循说明，以及任何偏离。

完成标准：每个依赖专门流程的委派任务都命名 skill、来源、为什么需要，以及编排者如何确认它被使用。

## 所有权边界

同一时间只有一个有写权限的 worker 可以拥有一个文件或子系统。

规则：

- 只有路径不重叠时，才允许并行写任务。
- 只读发现通常可以和大多数工作并行。
- 评审任务必须等待被评审工作达到终态。
- 修改共享组件的 UI 工作不得与这些组件上的实现工作重叠。
- 取消 writer 不是回滚；替换前先检查并对账部分变更。

完成标准：没有两个运行中的写任务能修改同一个文件、文件夹或逻辑子系统。

## 持久状态

把委派工作跟踪成一个小任务板：

```yaml
tasks:
  - id: ""
    specialist: ""
    objective: ""
    state: running # running | completed | error | cancelled | timed_out
    required_skills: []
    skills_confirmed: []
    ownership:
      files: []
      areas: []
    dependencies: []
    result: ""
```

对于长任务或高风险工作，把状态持久化到项目本地 markdown 文件或任务 artifact。包含：

- 目标
- 约束
- 阶段计划
- 任务板
- 决策轨迹
- 评审门
- 验证日志
- 未解决风险

完成标准：下一轮可以从显式状态恢复，而不是依赖分散的对话记忆。

## 对账

specialist 输出是输入，不是最终真相。

任务完成时：

1. 将结果与原始用户目标比较。
2. 检查与其他任务输出是否冲突。
3. 检查 required skills 是否已加载，以及偏离是否合理。
4. 决定接受、修订、拒绝，或派发后续工作。
5. 更新任务板。
6. 在下一次交接中保留有用决策。

完成标准：最终工作不依赖未经评审的 specialist 输出，也不依赖未经验证的 required-skill 使用。

## 保守反思

不要因为一次有趣运行就创建新的 skill、agent、command、rule 或 playbook。

使用这个阈值：

- 一个有用洞察 -> 保存到笔记或最终总结
- 2-3 次类似运行中反复摩擦 -> 建议项目指南或 CLAUDE.md
- 稳定重复工作流且触发条件清楚 -> 建议一个窄 skill
- 证据弱的猜测性改进 -> 什么都不创建

完成标准：过程改进有证据支持、保持最小，并放在能解决问题的最低持久层。

## 验证门

最终回复前：

- 所有必需任务都已终态
- 依赖任务已消费其等待的输出
- 文件所有权冲突已解决
- 相关检查已运行，或解释了跳过检查
- 残余风险已明确

完成标准：用户收到的是对账后的结果，而不是一堆 agent 报告。
