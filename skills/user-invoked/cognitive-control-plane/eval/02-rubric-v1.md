# Rubric v1 — cognitive-control-plane

12 项评判标准。6 Auto（可在 hook/脚本中检查）+ 6 Human（需人工判断）。

每项：field checked / check / mode / fail action

---

## Auto（6 项，可程序化检查）

### R-A1 — 任务分类先于路由
- **Check**: 在路由或做实质性工作之前，任务大小（Tiny/Small/Large）已被显式声明
- **Mode**: Auto（检测输出中是否出现 "Tiny" / "Small" / "Large" 分类声明，且出现在第一个控制面引用之前）
- **Fail action**: `quality_flag: no_classification_before_routing`

### R-A2 — Small 条件已检查
- **Check**: 当分类为 Small 时，输出中至少引用了 Small 的核心条件（scope known / no risk / single ownership / local edit），或显式声明所有条件满足
- **Mode**: Auto（检测是否出现 "Small" + 条件关键词或显式确认）
- **Fail action**: `quality_flag: small_without_condition_check`

### R-A3 — Large 实现触发了 Guard
- **Check**: 当任务被分类为 Large 且后续出现代码编辑时，输出中必须先出现 delegation contract 或显式的 direct-implementation exception 声明
- **Mode**: Auto（检测 Large 声明后第一个 Edit/Write 工具调用之前，是否有 task contract YAML 或 "direct implementation" 声明）
- **Fail action**: `hard_fail: large_implementation_without_guard`

### R-A4 — Task Contract 必需字段完整
- **Check**: 每次 delegation 的 task contract 包含 `objective`, `scope`/`ownership`, `expected_output`, `stop_if`（Small）或 `role`, `phase`, `objective`, `ownership`, `expected_output`, `stop_if`（Large）
- **Mode**: Auto（YAML 字段存在性检查）
- **Fail action**: `quality_flag: incomplete_task_contract`

### R-A5 — 控制面已显式命名
- **Check**: 当走 Large 路由时，选中的控制面（Context / Epistemic / Adversarial / Output / orchestration state）被显式命名
- **Mode**: Auto（检测是否出现控制面名称）
- **Fail action**: `quality_flag: surface_not_named`

### R-A6 — 控制面顺序正确
- **Check**: 当激活多个控制面时，顺序为 Context → Epistemic → Adversarial → Output（最早未满足的优先）
- **Mode**: Auto（检测控制面出现的相对顺序；只在对同一任务使用多个面时检查）
- **Fail action**: `quality_flag: surface_order_violation`

---

## Human（6 项，需人工判断）

### R-H1 — 任务大小分类正确
- **Check**: 给定输入的实际复杂度和风险，Tiny/Small/Large 分类是否恰当？
- **Mode**: Human
- **Fail action**: `quality_flag: misclassified_task_size`
- **特别注意**: 假 Small（应该 Large 但分了 Small）比假 Large 更严重

### R-H2 — 第一个控制面选择正确
- **Check**: 在四个控制面中，第一个激活的面是否是当前任务真正的瓶颈？
  - 目标/边界不清 → Context
  - 假设/证据不足 → Epistemic
  - 具体方案需攻击 → Adversarial
  - 准备交付/实现 → Output
- **Mode**: Human
- **Fail action**: `quality_flag: wrong_first_surface`

### R-H3 — 无仪式蔓延
- **Check**: 对 Tiny/Small 任务，没有出现完整的四阶段流程、task board、delegation contract 等大型仪式
- **Mode**: Human
- **Fail action**: `quality_flag: ceremony_creep`

### R-H4 — 知道何时停止路由
- **Check**: 路由在合适的时间点停止，交棒给具体实现/交付。没有在已明确该做什么后继续"再想想还有什么控制面"
- **Mode**: Human
- **Fail action**: `quality_flag: routing_forever`

### R-H5 — Task Contract 完整且自包含
- **Check**: 如果委托了 worker，task contract 是否足够完整，让 specialist 不需要猜测角色、范围、权限、输出格式、停止条件？
- **Mode**: Human
- **Fail action**: `quality_flag: contract_incomplete_for_specialist`

### R-H6 — 编排者没有偷当实现者
- **Check**: 对于 Large 任务，主 agent（作为 control plane）是否只做了编排和验证，而没有默默地自己做大量实现工作？
- **Mode**: Human
- **Fail action**: `quality_flag: orchestrator_as_implementer`

---

## Rubric 总结

| Mode | Items | 说明 |
|------|-------|------|
| Auto | R-A1 ~ R-A6 | 可在 PostToolUse hook 或独立脚本中检查 |
| Human | R-H1 ~ R-H6 | 需人工对照 Golden Set 场景评分 |

| Severity | Items |
|----------|-------|
| hard_fail | R-A3（Large 实现无 guard——最严重的失败模式） |
| quality_flag | 其余 11 项 |
