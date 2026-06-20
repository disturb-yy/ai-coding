# 日志规范

## 日志存储位置

所有 `[go-workflow]` 日志必须同时输出到：

| 目标 | 路径 | 说明 |
|------|------|------|
| 控制台 | stdout | 实时可见，kv 格式 |
| 文件 | `.agent/workflow-logs/<ISO8601-timestamp>-<task_type>.log` | 持久化存储，追加写入 |

**文件日志规则：**
- 日志目录自动创建：`mkdir -p .agent/workflow-logs/`；若项目无 `.agent/` 目录则使用 `/tmp/go-workflow-logs/`。
- 文件名格式：`2026-06-17T092130Z-feature.log`（ISO 8601 时间戳 + 任务类型）。
- 每条日志追加写入，不覆盖。
- Evaluate 阶段结束时在此文件中额外追加 `log_analysis` JSON 块。

**🔴 日志写入实现方式：**

所有日志必须通过 `exec_command` 工具写入，使用 `echo` 追加到日志文件。**禁止**仅输出到控制台而不写文件。

```bash
# 示例：Observe 阶段入口日志
exec_command: echo "[go-workflow] phase=Observe prev=start confidence=0.0" >> .agent/workflow-logs/<日志文件名>

# 示例：SubAgent spawn_failed + fallback_local
exec_command: echo "[go-workflow] agent=doc-index-reader event=spawn_failed phase=Observe reason=unsupported" >> .agent/workflow-logs/<日志文件名>
exec_command: echo "[go-workflow] agent=fallback event=fallback_local phase=Observe" >> .agent/workflow-logs/<日志文件名>

# 示例：Precheck 结果日志
exec_command: echo "[go-workflow] phase=Precheck build=pass vet=pass conflicts=0" >> .agent/workflow-logs/<日志文件名>

# 示例：Act 补丁应用日志
exec_command: echo "[go-workflow] phase=Act file=domain/strategy/http_handler.go status=applied" >> .agent/workflow-logs/<日志文件名>

# 示例：Evaluate 完成日志
exec_command: echo "[go-workflow] phase=Evaluate build=pass tests=28/28 completeness=done iterations=1" >> .agent/workflow-logs/<日志文件名>
```

**重要：** 日志文件名在启动协议步骤 1 中确定后，后续所有阶段日志必须使用**同一个文件名**追加。

---


每条日志统一前缀 `[go-workflow]`，kv 格式，可 grep 可结构化导入。

---



---

## 🔴 跨平台兼容性

日志写入使用 `exec_command` 工具执行 shell 命令，在不同平台存在差异。以下为兼容规范：

### 平台差异

| 平台 | Shell | mkdir | echo | 路径分隔符 | 特殊字符风险 |
|------|-------|-------|------|-----------|-------------|
| Linux | bash/sh | `mkdir -p` ✓ | `echo` ✓ | `/` | `"` `$` `` ` `` `!` |
| macOS | zsh/bash | `mkdir -p` ✓ | `echo` ✓ | `/` | `"` `$` `` ` `` `!` |
| Windows (bash) | bash (Git/MSYS2) | `mkdir -p` ✓ | `echo` ✓ | `/` | `"` `$` `` ` `` `!` |
| Windows (cmd) | cmd.exe | `md` | `echo` 行为不同 | `\` | `%` `^` `&` `<` `>` `|` |
| Windows (PowerShell) | pwsh | `New-Item -Force` | `Write-Output` | `\` | `$` `` ` `` `"` |

### 写入规范

1. **始终通过 `exec_command` 工具写入，不直接拼接 shell 命令字符串。**
   `exec_command` 工具自动处理底层 shell 差异和转义。

2. **日志内容中的特殊字符必须转义：**

   | 字符 | 转义为 | 说明 |
   |------|--------|------|
   | `"` | `\"` | 双引号破坏 echo 字符串边界 |
   | `$` | `\$` | shell 变量展开 |
   | `` ` `` | `\`` | 命令替换 |
   | `!` | `\!` | bash 历史展开 |
   | `\` | `\\` | 反斜杠自身 |

3. **日志目录创建：**
   - 优先：`mkdir -p .agent/workflow-logs/`
   - Windows cmd 回退：`if not exist .agent\workflow-logs md .agent\workflow-logs`
   - 最终回退：`/tmp/go-workflow-logs/`（所有平台通用）

4. **路径分隔符：** 日志文件路径统一使用正斜杠 `/`，`exec_command` 工具自动转换。

### 示例（安全写入）

```bash
# ❌ 危险：日志内容含特殊字符直接拼接
echo "[go-workflow] summary="CodeMap revealed: $symbol calls func()"" >> log.txt

# ✅ 安全：exec_command 独立传参，工具自动转义
exec_command: echo '[go-workflow] summary=CodeMap revealed calls chain' >> log.txt
```

### 摘要字段写入约束

`summary` 字段来自 SubAgent 输出，可能包含任意字符。
写入前必须经过**字符过滤**：移除或转义 `"`、`$`、`` ` ``、`!`、`
`。

## 🔴 必须记录

### 1. 阶段转换事件 `[log_level: always]`

每次进入新阶段时记录。

```
[go-workflow] phase=<阶段>  prev=<前序阶段>  confidence=<0.0-1.0>  [skip_reason=<跳过原因>]
```

| 字段 | 说明 |
|------|------|
| `phase` | 当前阶段名：`Observe\|Orient\|Decide\|Draft\|Precheck\|Act\|Evaluate` |
| `prev` | 前一个阶段名；Observe 阶段填 `start` |
| `confidence` | 进入该阶段时的累积置信度 |
| `skip_reason` | 仅当跳过 Orient 时填写，如 `skip_reason=confidence_high` |

**示例：**

```
[go-workflow] phase=Observe   prev=start      confidence=0.0
[go-workflow] phase=Orient    prev=Observe    confidence=0.75
[go-workflow] phase=Decide    prev=Orient     confidence=0.90
```

---

### 2. SubAgent 启动/结束 `[log_level: always]`

```
[go-workflow] agent=<名称>  event=spawn   phase=<阶段>  scope=<范围>  ts=<unix>
[go-workflow] agent=<名称>  event=return  phase=<阶段>  confidence=<0.0-1.0>
```

| 字段 | 说明 |
|------|------|
| `agent` | `doc-index-reader \| code-structure-analyzer \| module-code-analyzer \| test-runner` |
| `event` | `spawn` 启动 / `return` 返回 |
| `scope` | 仅 `module-code-analyzer` 填写，如 `scope=domain/strategy/http_handler.go` |
| `confidence` | SubAgent 返回的自我评估置信度 |

**示例：**

```
[go-workflow] agent=doc-index-reader       event=spawn   phase=Observe
[go-workflow] agent=doc-index-reader       event=return  phase=Observe   confidence=0.75
[go-workflow] agent=module-code-analyzer   event=spawn   phase=Draft     scope=domain/strategy/http_handler.go
[go-workflow] agent=module-code-analyzer   event=return  phase=Draft     confidence=0.90
```


### 2b. SubAgent 启动失败 / 回退到本地

子 Agent 不可用时（spawn_agent unsupported 或模型 404），记录回退事件。

**🔴 回退日志必须通过 exec_command 写入，不能仅输出到控制台：**

```bash
# spawn_failed 日志
echo "[go-workflow] agent=doc-index-reader event=spawn_failed phase=Observe reason=unsupported" >> .agent/workflow-logs/<日志文件>

# fallback_local 日志（紧跟其后，不可省略）
echo "[go-workflow] agent=fallback event=fallback_local phase=Observe" >> .agent/workflow-logs/<日志文件>
```

日志格式（控制台输出用）：

```
[go-workflow] agent=<名称>  event=spawn_failed  phase=<阶段>  reason=<原因>
[go-workflow] agent=fallback  event=fallback_local  phase=<阶段>  scope=<范围>
```

| 字段 | 说明 |
|------|------|
| `agent` | `doc-index-reader | code-structure-analyzer | module-code-analyzer | test-runner | fallback` |
| `event` | `spawn_failed` 启动失败 / `fallback_local` 回退到主 Agent 本地执行 |
| `reason` | 失败原因：`model_404` / `unsupported` / `timeout` |
| `scope` | 回退到本地执行时的范围（仅 module-code-analyzer 回退时填写） |

**示例：**

```
[go-workflow] agent=doc-index-reader       event=spawn_failed   phase=Observe   reason=model_404
[go-workflow] agent=fallback               event=fallback_local  phase=Observe
[go-workflow] agent=module-code-analyzer   event=spawn_failed   phase=Draft     reason=unsupported
[go-workflow] agent=fallback               event=fallback_local  phase=Draft     scope=domain/notify/history.go
```

**回退到本地时仍必须遵守所有阶段退出条件，包括日志检查点和 Evaluate 阶段的 log_analysis + 6 图表输出。**

---

### 3. SubAgent 决策摘要 `[log_level: always]`

SubAgent 返回后，记录标准契约中的 `decision_summary`。

```
[go-workflow] agent=<名称>  summary=<decision_summary 原文>
```

**格式约束：**
- `summary` 字段必须与 SubAgent 返回的 `decision_summary` 严格一致，不做截断或改写。
- 若 `decision_summary` 含换行或特殊字符，替换为空格。

**示例：**

```
[go-workflow] agent=doc-index-reader       summary="项目是包含 3 个模块的单体仓库：api、core、infra。架构为分层式。未知：认证中间件如何共享。"
[go-workflow] agent=module-code-analyzer   summary="CreateUser handler 缺少邮箱验证。建议：在 service 调用前添加正则检查。测试：验证无效邮箱返回 400。"
```

---

### 4. 错误与恢复 `[log_level: always]`

每次触发失败恢复规则时记录。

```
[go-workflow] phase=<阶段>  error=<错误类型>  recovery=<恢复动作>  [attempt=<N>]  [detail=<详情>]
```

**示例：**

```
[go-workflow] phase=Orient    error=structure_timeout    recovery=retry           attempt=1
[go-workflow] phase=Orient    error=structure_timeout    recovery=source_dive     attempt=2  detail="query_module(notification) 超时"
[go-workflow] phase=Draft     error=invalid_patch      recovery=relaunch        attempt=1  detail="parse error in http_handler.go:45"
[go-workflow] phase=Evaluate  error=test_failure       recovery=loop_to_orient  attempt=1  detail="TestUserUpdate flaky"
```

---

### 5. 补丁应用结果 `[log_level: always]`

Act 阶段每个文件应用后记录。

```
[go-workflow] phase=Act  file=<路径>  status=applied|failed|skipped  [detail=<原因>]
```

**示例：**

```
[go-workflow] phase=Act  file=domain/strategy/http_handler.go        status=applied
[go-workflow] phase=Act  file=domain/strategy/http_handler_test.go   status=applied
[go-workflow] phase=Act  file=api/router.go                          status=skipped  detail="审查任务，无代码变更"
```

---

## 🟡 重要记录

### 6. 置信度变化（阶段性） `[log_level: standard]`

每阶段结束时记录置信度增量。

```
[go-workflow] phase=<阶段>  confidence_before=<0.0-1.0>  confidence_after=<0.0-1.0>  delta=<±0.0-1.0>
```

**示例：**

```
[go-workflow] phase=Observe   confidence_before=0.0   confidence_after=0.75  delta=+0.75
[go-workflow] phase=Orient    confidence_before=0.75  confidence_after=0.90  delta=+0.15
[go-workflow] phase=Decide    confidence_before=0.90  confidence_after=0.85  delta=-0.05
```

当 `delta < 0` 时说明新阶段发现了降低置信度的信息，需特别关注。

---

### 7. 调度决策理由 `[log_level: standard]`

Decide 阶段记录为何选择当前调度策略。

```
[go-workflow] phase=Decide  task_type=<类型>  dispatch_count=<N>  strategy=parallel|serial|single  reason=<理由>
```

**示例：**

```
[go-workflow] phase=Decide  task_type=feature    dispatch_count=2  strategy=parallel  reason="scopes non-overlapping: core/service/user.go vs notification/email.go"
[go-workflow] phase=Decide  task_type=refactor   dispatch_count=3  strategy=serial    reason="scopes overlap: all touch api/handler/ auth logic"
[go-workflow] phase=Decide  task_type=bugfix     dispatch_count=1  strategy=single    reason="single file affected: core/service/user.go"
```

---

### 8. 上下文 Token 估算 `[log_level: standard]`

每次向 SubAgent 传递上下文时，粗估消耗的 Token。

```
[go-workflow] agent=<名称>  chunks_passed=<N>  tokens_estimated=<整数>
```

**估算公式（粗粒度）：**

- `index` 字段：按 1 Token / 4 字符（英文）或 1 Token / 1.5 字符（中文）估算。
- `chunks` 字段：逐条累加 `content` 字段的 Token 估算。
- 系统提示词、约束描述等框架文本：固定 +200 Token。

**示例：**

```
[go-workflow] agent=module-code-analyzer  chunks_passed=3  tokens_estimated=850
```

---

### 9. 预检 / 评估结果 `[log_level: always]`

```
[go-workflow] phase=Precheck   build=pass|fail  vet=pass|fail  conflicts=<N>  [detail=<失败详情>]
[go-workflow] phase=Evaluate   build=pass|fail  tests=<通过>/<总数>  completeness=done|needs_iteration  iterations=<N>
```

**示例：**

```
[go-workflow] phase=Precheck   build=pass  vet=pass  conflicts=0
[go-workflow] phase=Evaluate   build=pass  tests=28/28  completeness=done  iterations=1
[go-workflow] phase=Precheck   build=fail  vet=pass  conflicts=0  detail="undefined: foo.Bar in handler.go:42"
[go-workflow] phase=Evaluate   build=pass  tests=26/28  completeness=needs_iteration  iterations=2  detail="TestUserUpdate timeout, TestMarketScore flaky"
```

---

### 10. 迭代回退原因 `[log_level: standard]`

Evaluate 判定 `needs_iteration` 时记录。

```
[go-workflow] iteration=<N>  from=Evaluate  to=Orient|Draft  reason=<原因>
```

**示例：**

```
[go-workflow] iteration=1  from=Evaluate  to=Draft   reason="precheck build fail: missing import in handler.go"
[go-workflow] iteration=1  from=Evaluate  to=Orient  reason="test failure: TestUserUpdate flaky, re-assessing impact scope"
[go-workflow] iteration=2  from=Evaluate  to=Orient  reason="代码结构工具 revealed additional dependency: infra/cache"
[go-workflow] iteration=3  from=Evaluate  to=done    reason="max iterations reached, reporting partial results"
```

---

## 完整走查示例

一次 bugfix 任务（单 Agent、单次迭代、全部通过）的完整日志：

```
[go-workflow] phase=Observe   prev=start      confidence=0.0
[go-workflow] agent=doc-index-reader           event=spawn   phase=Observe
[go-workflow] agent=doc-index-reader           event=return  phase=Observe   confidence=0.75
[go-workflow] agent=doc-index-reader           summary="项目 stock-monitor 包含 7 个领域模块。architecture 为 DDF 分层。未知：strategy 模块参数校验范围。"
[go-workflow] phase=Observe   confidence_before=0.0   confidence_after=0.75  delta=+0.75
[go-workflow] phase=Orient    prev=Observe    confidence=0.75
[go-workflow] agent=code-structure-analyzer           event=spawn   phase=Orient
[go-workflow] agent=code-structure-analyzer           event=return  phase=Orient     confidence=0.90
[go-workflow] agent=code-structure-analyzer           summary="validateMACrossParams 位于 strategy/http_handler.go:362。调用链：CreateInstance → validateMACrossParams。仅 MA 交叉策略使用。"
[go-workflow] phase=Orient    confidence_before=0.75  confidence_after=0.90  delta=+0.15
[go-workflow] phase=Decide    prev=Orient     confidence=0.90
[go-workflow] phase=Decide    task_type=bugfix  dispatch_count=1  strategy=single  reason="single function affected"
[go-workflow] phase=Draft     prev=Decide     confidence=0.85
[go-workflow] agent=module-code-analyzer       event=spawn   phase=Draft     scope=domain/strategy/http_handler.go
[go-workflow] agent=module-code-analyzer       chunks_passed=2  tokens_estimated=600
[go-workflow] agent=module-code-analyzer       event=return  phase=Draft     confidence=0.90
[go-workflow] agent=module-code-analyzer       summary="validateMACrossParams 未校验缺失键和零值。建议 6 层逐级校验。测试覆盖 7 个单元用例 + 2 个 HTTP 集成用例。"
[go-workflow] phase=Precheck   prev=Draft      confidence=0.90
[go-workflow] phase=Precheck   build=pass  vet=pass  conflicts=0
[go-workflow] phase=Act        prev=Precheck   confidence=0.90
[go-workflow] phase=Act        file=domain/strategy/http_handler.go        status=applied
[go-workflow] phase=Act        file=domain/strategy/http_handler_test.go   status=applied
[go-workflow] phase=Evaluate   prev=Act        confidence=0.90
[go-workflow] phase=Evaluate   build=pass  tests=28/28  completeness=done  iterations=1
```
