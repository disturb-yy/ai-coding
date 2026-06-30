---
name: coding-tdd
description: "在现有项目中结合 TDD 和 coding-project 执行测试先行的代码变更。适用于 red-green-refactor、回归修复、集成式测试、从入口到输出的工作流，以及要求一次只实现一个小函数或小模块的请求。不用于仅解释 TDD、只读测试评审，或没有测试先行要求的一般编码任务。"
---

# Coding TDD

> 本文件是 `SKILL.md` 的中文描述副本，仅供用户阅读和发布展示使用。
> 严禁模型读取本文件作为操作指令、任务上下文或执行依据。
> 修改英文版本 `SKILL.md` 时，必须在同一次变更中同步更新本中文版本。

## 本地化维护

- 修改英文 `SKILL.md` 时，必须同步更新 `SKILL.zh-CN.md`。
- `SKILL.zh-CN.md` 只面向用户可见，不作为模型可读的执行说明。
- 模型可读的权威来源是英文 `SKILL.md`。

## 目标

运行一个测试先行的编码工作流，组合 `/tdd` 和 `/coding-project`。

- TDD 负责测试用例、行为切片、red-green-refactor 循环和端到端验证形态。
- `coding-project` 负责语言感知的代码编辑、项目约定、依赖使用、安全预检和验证命令。
- 保持模块最小化。每个循环只实现一个小模块或一个函数，除非项目结构要求不可拆分的变更。

## 必要组合流程

每个编码任务都按以下顺序执行：

1. 加载 `/coding-project`，观察仓库、识别语言、加载匹配的语言和单元测试规范，并确定验证命令。
2. 应用 `/tdd` 纪律，选择最小的外部可观察行为优先测试。
3. 在实现前，为该行为写一个聚焦的失败测试。
4. 使用 `/coding-project` 只实现让该测试通过所需的最小函数或模块。
5. 运行最窄相关测试命令，并确认 GREEN。
6. 只在 GREEN 状态下重构，然后重新运行同一个测试命令。
7. 对下一个独立行为或函数重复上述循环。
8. 所有模块完成后，编写并运行端到端命令，验证完整的入口到输出路径。

如果任何验证失败，停止继续推进，检查原因、修复问题，并重新运行失败的检查点。

## 行为切片

优先使用垂直行为切片，而不是宽泛实现计划。

```text
Request A -> function B -> function C -> function D -> Response E
```

测试按以下形态组织：

| 范围 | 测试职责 | 依赖规则 |
| --- | --- | --- |
| Request A 行为 | 验证公开请求/API 行为和预期响应契约。 | mock 本切片之外的下游依赖。 |
| Function B | 通过公开接口验证 B 的行为。 | mock B 的依赖，例如 C 或外部 client。 |
| Function C | 通过公开接口验证 C 的行为。 | mock C 的依赖。 |
| Function D | 通过公开接口验证 D 的行为。 | mock D 的依赖。 |
| 端到端路径 | 验证 Request A 在运行中的应用或最接近的项目支持环境中产生 Response E。 | 优先使用真实 wiring；只按项目约定 mock 不可用的外部系统。 |

不要先写完所有测试再写所有实现。一次写一个测试、实现一个小目标、验证通过后再继续。

## 并行 SubAgent 规则

只有当 subagent 工具可用，并且工作单元相互独立时，才使用 subAgent。如果没有 subagent 工具，则按相同边界串行处理。

| 场景 | 动作 |
| --- | --- |
| 函数之间没有直接依赖，且测试可独立运行 | 工具可用时，每个函数或模块启动一个 subAgent。每个 subAgent 必须使用 `/coding-project`，只实现分配目标，并运行自己的窄验证命令。 |
| Function B 依赖 C 的接口或行为 | 在契约清晰稳定前，不并行 B 和 C。 |
| 涉及共享文件、迁移、生成代码或公开 API 契约 | 保持串行，除非 subAgent 边界明确且不容易产生合并冲突。 |
| 涉及安全敏感代码 | 分配 subAgent 前先执行 `/coding-project` 安全预检。 |

每个 subAgent 必须收到：

```text
Use /coding-project. Implement only <function/module>. The TDD test already defines the expected behavior. Keep the change minimal, follow project conventions, and run <targeted validation command>.
```

subAgent 完成后，统一检查它们的 diff，解决集成问题，并在端到端验证前运行组合后的受影响测试。不要把共享 API 设计、schema 变更或最终集成决策委托给 subAgent。

## 验证检查点

使用快速失败检查点：

1. 每写完一个测试后，运行窄测试，并确认它因预期原因失败。
2. 实现目标函数或模块后，重新运行同一测试并确认通过。
3. 重构后，重新运行同一测试。
4. 合并独立模块后，运行受影响 package/module 测试。
5. 所有模块完成后，运行入口到输出命令。

对于 HTTP API，写出可直接复制的最终 `curl` 命令：

```bash
curl -i -X POST "$BASE_URL/path" \
  -H "Content-Type: application/json" \
  -d '{"example":"value"}'
```

根据项目调整 method、headers、auth 和 payload。运行 `curl` 前，使用仓库记录的本地服务启动命令或测试环境设置。

对于非 HTTP 工作流，使用最接近的项目支持入口命令：

```bash
<project command> <input-or-fixture>
```

说明预期可观察输出，并通过 stdout、生成文件、数据库可见行为或项目支持的检查命令验证。

## 决策表

| 用户请求 | 使用本 skill? | 动作 |
| --- | --- | --- |
| “测试先行构建这个功能” | 是 | 使用 `/coding-project`，然后开始 TDD 循环。 |
| “先写回归测试再修这个 bug” | 是 | 写失败的回归测试，再实现最小修复。 |
| “围绕这个 endpoint 增加集成测试” | 是 | 从请求级行为开始，再按需要测试和实现内部函数。 |
| “解释 TDD” | 否 | 做概念性回答，或仅在可用时用 `/tdd` 辅助指导。 |
| “直接实现这个变更” | 否，除非用户要求测试先行 | 使用 `/coding-project`。 |
| “评审我的测试计划” | 否 | 只评审，不编辑，除非用户要求实现。 |

## 示例

### Endpoint 功能

输入：

```text
Add POST /orders/quote test-first. It should return a quote total and reject empty carts.
```

期望行为：

```text
Use /coding-project to inspect routes, handlers, services, test patterns, and validation commands.
Write one failing request-level test for the valid quote path.
Implement only the handler/service function needed for that test.
Run the targeted test to GREEN.
Add the empty-cart test, implement the smallest validation change, and rerun.
Finish with a curl command that verifies POST /orders/quote returns the expected response.
```

### 独立函数

输入：

```text
Request A calls normalizeCustomer, calculateDiscount, and formatResponse. They do not depend on each other. Build this TDD.
```

期望行为：

```text
Write focused tests for each public behavior with dependencies mocked.
If the functions are independent and touch separate files or stable contracts, assign one subAgent per function.
Each subAgent uses /coding-project, implements only its function, and runs the targeted test.
After merging, run the affected module tests and the request-level test.
Finish with the request-to-response curl command.
```

### 回归修复

输入：

```text
Fix the null status bug red-green-refactor style.
```

期望行为：

```text
Use /coding-project to locate the failing path and test command.
Write one regression test that fails because null status is mishandled.
Implement the smallest function-level fix.
Run the targeted test to GREEN, refactor only if needed, and rerun.
Run the broader affected tests if the fix touches shared status mapping.
```

### 测试环境阻塞

输入：

```text
Add this endpoint with TDD, but there is no documented test command.
```

期望行为：

```text
Use /coding-project to search project docs, CI, package scripts, Makefiles, and build files for validation commands.
If no command exists, infer the narrowest standard command for the detected language and state the assumption.
If the test framework is missing or cannot run, report the blocker before implementation unless a small local test harness is appropriate for the project.
```
