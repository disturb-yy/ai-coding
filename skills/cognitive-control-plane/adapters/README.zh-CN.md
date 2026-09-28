# 工作项调度与新会话启动：中文使用说明

这套机制把“需要完成的事情”和“处理它的一次 AI 会话”分开：

- **work item（工作项）**：持久化的业务对象。可以是 `issue`、`request`、`transaction` 或已拆分的 `ticket`。
- **run（运行）**：处理同一工作项的一次会话尝试。一次 run 结束并不等于工作项结束。
- **Scheduler（调度器）**：读取持久化状态，判断依赖、租约、预算和终态，决定是否可以开始下一次 run。
- **Adapter（适配器）**：仅在调度器允许时，以新会话启动 Codex 或 OpenCode。

因此，不能把“当前会话的聊天历史”当成续跑状态。续跑时必须创建新的
run，并把 checkpoint 显式传入新会话。

## 前置条件

1. 已安装 Node.js、Codex CLI 或 OpenCode CLI。
2. 宿主系统已有持久化存储；可以是数据库、问题单系统或队列。脚本只根据
   `state.json` 计算决策，不替代存储、抢租约或写回状态。
3. 每次可执行 run 都先生成符合
   [`contract.schema.json`](contract.schema.json) 的 `contract.json`。工作项为
   `transaction` 时，必须有非空的 `work_item.idempotency_key`。

## 最小工作流

```text
领取或读取工作项
  -> 持久化系统检查依赖并获取租约
  -> work-item-scheduler.js 给出决策
  -> work-item-loop.js 校验决策与契约
  -> ccp-adapter.js 启动一个新会话
  -> Runner 写 checkpoint / 验证证据 / 结论
  -> 持久化系统更新状态，再进入下一轮
```

调度器只会在 `dispatch` 或安全的 `continue` 时允许启动会话。`wait`、
`checkpoint`、`verify`、`close` 和 `wait_for_human` 都不会启动新的 Worker。

## 1. 准备初次派发的状态与契约

下面示例使用一个尚未开始的 issue。将文件存放在工作区外的运行状态目录，
不要把真实租约、审批信息或密钥提交进代码库。

`state.json`：

```json
{
  "now": "2026-08-02T09:00:00Z",
  "work_item": {
    "id": "INC-200",
    "kind": "issue",
    "dependencies": []
  }
}
```

`contract.json`：

```json
{
  "ccp_version": 1,
  "next_action": "delegate_write",
  "task": {
    "task_id": "INC-200-R01",
    "actor_id": "worker-1",
    "role": "work_item_runner",
    "phase": "work_item",
    "objective": "定位并解决 INC-200，并记录验证证据。",
    "work_item": {
      "id": "INC-200",
      "kind": "issue",
      "objective": "定位并解决 INC-200。",
      "acceptance_criteria": ["记录可验证的终态或结论。"],
      "dependencies": [],
      "authorization": []
    },
    "run": {
      "id": "INC-200-R01",
      "attempt": 1,
      "lease_id": "由持久化系统生成的租约 ID",
      "lease_expires_at": "2026-08-02T09:30:00Z",
      "resume_checkpoint_ref": "",
      "budget": {
        "checkpoint_at_fraction": 0.4,
        "handoff_at_fraction": 0.45,
        "hard_stop_at_fraction": 0.5
      }
    },
    "constraints": [],
    "required_skills": [],
    "required_references": [],
    "required_mcp": [],
    "required_tools": [],
    "ownership": {
      "writable_paths": [],
      "read_only_paths": [],
      "forbidden_paths": []
    },
    "edits_allowed": true,
    "expected_output": {
      "format": "json",
      "required_fields": [],
      "must_report": []
    },
    "validation": [],
    "stop_if": []
  }
}
```

## 2. 先只计算调度决策

```bash
node scripts/work-item-scheduler.js --dry-run state.json
```

初次运行应得到 `action: "dispatch"`。调度器的主要决策如下：

| 决策 | 含义 | 宿主下一步 |
| --- | --- | --- |
| `dispatch` | 工作项可以首次运行 | 生成/确认 run 契约，启动新会话 |
| `checkpoint` | 当前 run 到达 40% 或 50% 且未保存状态 | 让当前 Worker 写 checkpoint |
| `continue` | 旧 run 已 checkpointed 或 expired，可以续同一工作项 | 创建下一 attempt 的新 run 后启动 |
| `verify` | 声称 resolved 但缺验证证据 | 先运行验收检查 |
| `close` | 已有充分验证或结论证据 | 关闭工作项，再选择下一工作项 |
| `wait` | 依赖未完成、租约仍有效或旧 run 仍在运行 | 不启动会话 |
| `wait_for_human` | blocked、escalated 或事务缺幂等键 | 等待人工或外部条件 |

## 3. 预演与启动 Codex

默认行为是预演，不会启动进程：

```bash
node scripts/work-item-loop.js \
  --platform codex \
  --workspace /absolute/path/to/repository \
  state.json contract.json
```

确认 JSON 输出中 `status` 为 `dry_run`、`scheduling.action` 为 `dispatch` 或
`continue` 后，才显式启动：

```bash
node scripts/work-item-loop.js \
  --platform codex \
  --workspace /absolute/path/to/repository \
  --sandbox workspace-write \
  --approval on-request \
  --execute \
  state.json contract.json
```

适配器实际使用 `codex exec --json -C ... -s ... -a ...`。返回 `started` 和本地
`pid` 只表示进程已启动；`work_id` / `run_id` 是本系统的可移植标识，不能当作
Codex 原生 session ID。

## 4. 启动 OpenCode

将平台替换为 `opencode` 即可：

```bash
node scripts/work-item-loop.js \
  --platform opencode \
  --workspace /absolute/path/to/repository \
  --execute \
  state.json contract.json
```

适配器使用 `opencode run --format json --dir ...`。两种平台都只创建新会话，
不会自动使用 `resume`、`continue` 或跳过权限的危险参数。

## 5. 在 50% session 预算内续跑

建议策略：40% 写 checkpoint，45% 只完成当前原子操作、验证或交接，50% 结束
当前 run。若旧 run 已写入 checkpoint 并结束，状态可类似：

```json
{
  "now": "2026-08-02T09:20:00Z",
  "work_item": { "id": "INC-200", "kind": "issue", "dependencies": [] },
  "run": {
    "id": "INC-200-R01",
    "attempt": 1,
    "state": "checkpointed",
    "usage_fraction": 0.46,
    "checkpoint": { "written": true }
  }
}
```

此时新契约必须把 run 改为 `INC-200-R02`、`attempt: 2`，并设置：

```json
"resume_checkpoint_ref": "checkpoints/INC-200-R01.json"
```

如果旧 run 仍是 `running`，即使已有 checkpoint，调度器也会返回 `wait`；这样
不会出现两个会话同时处理同一工作项。

## Programmatic Tool Calling（PTC）放在哪里

PTC 仅属于一次 run 内部。例如：批量读取 30 个服务状态、过滤异常项、合并结果
或机械校验 JSON。它的结果应是小型结构化数据，再由宿主保存和调度。

下列事项不能放进 PTC 循环：持久状态、租约转换、启动/停止会话、审批、外部写入
以及需要新模型判断的语义决策。原因是 PTC 运行环境不承担这些持久化和权限职责。
详见 [OpenAI Programmatic Tool Calling 指南](https://developers.openai.com/api/docs/guides/tools-programmatic-tool-calling)。

## 验证本地改动

```bash
node eval/tests/work-item-scheduler.test.js
node eval/tests/work-item-loop.test.js
node eval/tests/ccp-adapter-launch.test.js
python3 eval/scripts/static_checks.py --skill-dir .
node scripts/check-mirrors.js
```

模型型 eval 位于 `eval/cases/06-work-item-scheduler.yaml`，当前共有 62 个运行时
案例。运行完整基线会调用模型，应在允许相应外部成本后再执行。
