---
access:
  audience: model
  model_read: false
  model_write: true
  purpose: skill_reference
---
# 审查者强制机制

在实现后，当风险决定是否必须进行独立审查，或审查发现、修复、重新审查和制品新鲜度必须作为接受关卡时，使用此参考。

## 强制审查策略

当实现或修复任务达到终态，且交付变更具有以下一个或多个风险标签时，创建独立的 `review` 任务：

- `security_sensitive`
- `cross_module_change`
- `public_api_change`
- `schema_change`
- `migration`
- `auth_or_permission_change`
- `deployment_or_rollback_critical`

根据实际交付制品评估风险，而不仅是原始请求。工作者报告披露的新触发项会激活该策略。实现测试通过不能免除审查。

强制审查任务必须：

- 依赖它所审查的实现或修复任务
- 使用 `phase: review` 和 `edits_allowed: false`
- 当该能力可用时要求 `reviewing-code`
- 命名实现任务及其 actor
- 绑定稳定的制品版本
- 分别报告阻断性发现和非阻断性发现

启动任务前预检 `reviewing-code` 和独立审查者 actor。如果 `reviewing-code` 不可用，唯一可接受的回退是 `independent_read_only_reviewer`：由宿主启动的不同只读 actor，其契约要求此参考、记录 `review_fallback: independent_read_only_reviewer`，并将不可用技能报告为偏离。如果无法启动此类 actor，则保持审查关卡阻塞，并交接或以未接受状态终止。不要降级为自我审查。

低风险实现仍可接受可选审查。用户明确指令可以增加审查触发项，但不得悄然移除强制触发项。如果用户明确终止循环，应报告已终止且未接受的结果以及残余风险；不要将其表示为经过审查的接受结果。

完成标准：每个终态实现制品都已接受风险评估，且每个匹配的高风险制品都有待处理或已终止的独立审查任务。

## 审查报告格式

使用包含两个独立表格的紧凑审查报告。不要把成功检查和缺陷合并为一个无差别列表。

1. **验证矩阵**：`status`、`review lane/check`、`evidence` 和 `result or limitation`。视情况包含语法、行为、标准、安全、测试以及跳过/受阻的通道。
2. **发现表**：`id`、`severity`、`location`、`evidence`、`recommendation` 和 `disposition`。当表为空时，明确声明 `no findings`。

然后给出简短的关卡决策，命名不可变的 `review_target`、阻断性发现 ID、残余风险，以及当前版本是 `cleared`、`blocked` 还是 `terminated`。表格让最终状态便于浏览；仅当调查理由能实质解释某项发现或限制时，才用正文说明。

完成标准：使用者无需从正文推断，即可区分已通过验证、跳过的检查、阻断性发现、非阻断性发现和关卡决策。

## 审查者独立性

使用稳定的 `actor_id` 标识工作者，而不是角色标签或提示词。对于每次审查迭代：

```text
review.actor_id != reviewed_implementation.actor_id
```

同一 actor 不得先实现、再切换角色并审查自己的工作。具有相同 actor ID 的不同任务 ID 不具独立性。审查者是只读的，也不得负责针对其发现的修复任务。如果审查者没有实现或修复后续版本，则可以重新审查该后续版本。

如果无法启动独立审查者，则保持审查关卡阻塞，并发出交接或明确终止状态。不要把强制审查降级为自我审查。

完成标准：每个被接受的审查都记录不同的实现者和审查者 actor ID。

## 制品版本绑定

将每次审查固定到恰好一个稳定制品版本。优先使用以下 Git 身份：

```yaml
review_target:
  kind: git_range
  base_sha: ""
  head_sha: ""
  diff_hash: ""
  stable_id: ""
```

对于 Git diff，要求提供 `base_sha`、`head_sha` 和 `diff_hash`。当 Git 身份不可用时，使用 `kind: stable_artifact` 和非空的 `stable_id`，例如不可变构建摘要、生成包校验和、迁移集摘要或内容哈希。分支名称、可变文件路径、没有 head SHA 的 PR 编号或“最新变更”等描述性文字都不够稳定。

在审查任务契约和审查结果中记录相同目标。在使用结果前，重新计算或检索当前制品身份，并将其与已审查身份比较。

审查后任何制品变更都会使先前审查失效，包括发现修复、清理编辑、生成文件刷新、冲突解决、改变内容的 rebase 或手工编辑。将旧审查标记为 `invalidated`；绝不要把其已放行关卡延续到新版本。

完成标准：被接受的审查目标与最终交付制品版本完全相同。

## 阻断性发现关卡

在有仓库审查严重性规则时使用该规则。否则，当发现表明存在交付前必须修复的重大正确性、安全、数据完整性、权限、迁移、公共 API 兼容性、部署或回滚缺陷时，将其视为阻断性发现。

当审查者报告一个或多个阻断性发现时：

1. 将审查关卡设为 `blocked`。
2. 阻止最终接受和交付声明。
3. 派发依赖该审查任务且可写的修复任务。
4. 在修复契约中包含已接受的阻断性发现和固定的被审版本。
5. 保持非阻断性发现可见，不要悄然扩大修复范围。

编排者可以在协调统一期间拒绝缺乏依据的审查者主张，但必须记录基于证据的拒绝理由。未解决的阻断性发现仍具有阻断性。

完成标准：只要当前制品谱系仍受未解决的阻断性发现影响，就不存在最终接受路径。

## 修复与重新审查循环

修复任务完成后：

1. 计算新的制品版本。
2. 使目标不同于新版本的所有先前审查失效。
3. 为新版本创建新的独立审查任务。
4. 要求新审查者 actor 不同于实现或修复该版本的 actor。
5. 协调统一新的发现。
6. 只要仍有阻断性发现，就重复修复和重新审查循环。

循环仅在以下情况终止：

- 最新制品版本具有有效的独立审查，且没有阻断性发现；或
- 用户或控制策略明确终止循环。

终止不等于接受。记录原因、未解决发现、最后被审版本、当前版本和残余风险。

完成标准：最新版本已被独立放行，或运行明确以已终止且未接受的状态结束。

## 审查关卡状态机

```text
implementation_completed
  -> risk_assessed
  -> review_required? ---- no ----> verification
         |
        yes
         v
  review_pending -> review_running -> findings_reconciled
                                           |
                         no blocking ------+-----> cleared
                                           |
                         blocking ----------> fix_pending
                                                   |
                                                   v
                                              fix_completed
                                                   |
                                                   v
                                      prior_review_invalidated
                                                   |
                                                   v
                                             review_pending
```

在任何时点，制品版本变化都会把已完成的审查记录移到 `invalidated`，并创建当前强制关卡或使其返回 `review_pending`。保留历史上已失效的审查记录作为谱系证据；只有当前制品版本的关卡必须变为 `cleared`。明确终止会把当前关卡移到 `terminated`，它不能满足最终接受条件。

## 必需的审查状态

至少持久化以下内容：

```yaml
review_gate:
  required: true
  risk_tags: []
  status: pending # pending | running | blocked | cleared | invalidated | terminated
  artifact_producer_task_id: "" # 实现任务或最新修复任务
  artifact_producer_actor_id: "" # 生成固定版本的 actor
  review_task_id: ""
  reviewer_actor_id: ""
  review_iteration: 1
  review_target:
    kind: git_range # git_range | stable_artifact
    base_sha: ""
    head_sha: ""
    diff_hash: ""
    stable_id: ""
  blocking_finding_ids: []
  supersedes_review_task_id: ""
  termination_reason: ""
```

在编排事件日志中记录状态转换，并附上证据和制品引用。

## 最终接受条件

在最终接受前，要求以下所有条件均满足：

- 每个实现制品都经过风险评估
- 每个强制审查任务所使用的 actor 都独立于被审实现或修复的 actor
- 最新制品版本具有有效且匹配的审查目标
- 最新有效审查没有未解决的阻断性发现
- 每个已修复的阻断性发现之后都进行了重新审查
- 当前制品的强制审查关卡为 `cleared`；历史上被取代的审查记录可以保留为 `invalidated`

完成标准：接受指向经过独立审查的最终制品，而不是更早版本或不完整循环。
