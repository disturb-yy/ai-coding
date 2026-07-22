# execute-sop

以受控、可交接的方式处理问题单。它先分析问题，再由用户批准 ticket DAG；用户执行 ticket，模型维护状态和交接信息。运行规则以 [SKILL.md](SKILL.md) 为准；本页说明使用方式。

## 适用场景

- 一个问题需要拆为多个有依赖关系的处理 ticket。
- 希望限制模型的操作边界，避免从问题描述直接进入代码修改、提交或部署。
- 需要在不同模型、不同会话或多人之间持续记录当前处理阶段、证据和下一步。

默认不会执行 ticket，也不会修改业务源码、创建实现分支、提交或推送。

## 发起方式

```text
$execute-sop

处理以下问题单，并将 case file 创建在 .scratch/order-refund-timeout/：
退款服务偶发超时，客户看见失败但支付渠道可能已成功扣款。
先分析证据和风险，提出 ticket DAG；不要创建 ticket，等我批准后再写入。
```

首次调用会分析问题并将状态设为 `awaiting-ticket-approval`。批准示例：

```text
批准当前 ticket DAG。请创建 tickets；我会自行执行 frontier 中的 ticket。
```

## Case file 结构

默认位置为 `.scratch/<problem-slug>/`：

```text
.scratch/<problem-slug>/
├── STATE.md
└── issues/
    ├── 01-<ticket-slug>.md
    └── 02-<ticket-slug>.md
```

- `STATE.md`：唯一的状态索引，保存阶段、frontier、ticket 状态、证据指针、阻塞项和交接信息。
- `issues/*.md`：每张 ticket 的完整执行目标、阻塞边和验收条件。

不要在 `STATE.md` 重复 ticket 的详细操作；它只负责回答“当前在哪里、谁该做什么、下一步是什么”。

## 阶段与职责

| 阶段 | 模型做什么 | 用户做什么 |
|---|---|---|
| `analyze` | 分析事实、假设、风险和 ticket 草案 | 补充决定性信息 |
| `awaiting-ticket-approval` | 说明 DAG 和 frontier | 批准、合并、拆分或拒绝 tickets |
| `ticketed-awaiting-user` | 更新状态并说明可执行 frontier | 执行未被阻塞的 ticket，并提供证据 |
| `awaiting-verification` | 复核用户提供的证据或按请求进行只读检查 | 补充证据或确认结果 |
| `blocked` / `resolved` | 记录阻塞或总结关闭状态 | 清除阻塞或明确重开 |

## 继续处理与交接

后续会话直接提供 case file 路径：

```text
$execute-sop

继续处理 .scratch/order-refund-timeout/。
我已完成“确认支付渠道最终状态”，证据在 issues/01-confirm-payment-state.md；
请更新状态和新的 frontier。
```

模型会先读取 `STATE.md`，然后只执行该阶段允许的状态维护动作。每次交付都会说明 case file 路径、当前阶段、frontier、阻塞项及用户需要做的下一步。
