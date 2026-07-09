---
name: reviewing-code
description: "审查代码变更、PR、分支、diff 或实现产物，覆盖语法、功能正确性、仓库规范和安全问题。"
metadata:
  access:
    audience: human
    model_read: false
    model_write: true
    purpose: zh_mirror
---

# 代码审查

> 中文版本仅供人类维护者阅读。模型和 agent 不得把 `zh/` 文件作为运行指令或任务上下文读取；英文文件是唯一 canonical 模型指令。

## 本地化维护

英文文件是 canonical、面向模型的指令。修改此 skill 时，必须在同一次变更中同步更新 `zh/` 下对应的中文镜像，供人类维护者阅读。

## 概念空间

```yaml
conceptual_space:
  target_region: "需要在接受前审查的代码变更或具体实现产物。"
  deviation_region:
    - "没有代码或 diff 证据的一般架构批判。"
    - "检视意见被接受后的实现工作。"
    - "只有 lint 或 SAST 工具输出、没有人工审查。"
  priority_dimensions:
    - "先找实质缺陷，再处理风格偏好。"
    - "区分证据和猜测。"
    - "各审查 lane 在汇总前保持独立。"
    - "用 diff、源码、测试或引用规范验证主张。"
  entry_conditions:
    - "用户要求代码审查、PR 审查、分支审查、diff 审查、安全审查，或 review since 某个 ref。"
    - "具体计划或实现有代码、diff、commit、PR 或文件可检查。"
  exit_conditions:
    - "每个选中的审查 lane 都报告了发现，或明确说明没有发现。"
    - "发现已去重、按严重度排序，并绑定到文件/行、hunk、测试、spec 或规范证据。"
    - "残余风险和跳过的 lane 已明确。"
```

## 工作流

1. 框定审查目标。识别输入是 PR、分支/ref 范围、工作区 diff、指定文件还是粘贴代码。如果需要 diff base 但用户未提供，先询问；否则使用最小可用目标。
   完成标准：审查目标、diff 命令或文件列表，以及用户给出的验收标准已明确。
2. 盘点证据。收集变更文件、commit、附近测试、仓库规范、可用的 spec/issue/PRD，以及安全相关面：认证、输入处理、密钥、持久化、网络调用和依赖变更。
   完成标准：每个选中的 lane 都有需要的文件和证据，或缺失证据已标记为约束。
3. 选择审查包。用 `references/code/` 做语法、功能和规范审查。当代码涉及安全敏感逻辑、依赖/配置变更、外部输入、凭据、权限、数据访问、网络调用，或用户要求安全审查时，用 `references/security/`。
   完成标准：已命名选中的 reference 文件夹，未选中的文件夹有原因。
4. 委派审查 lane。当 subagent 或 subtask 可用时，为每个选中的 reference 文件夹创建一个独立只读 subtask；每个 subtask 必须先读取该文件夹内所有文件再审查。如果文件夹包含互相独立的大型 checklist，再按文件拆分。没有 subagent 时，用分离笔记串行执行同样的 lane。
   完成标准：每个选中文件夹都有一个完成的 lane report，或每个拆分文件都有一个完成报告。
5. 验证发现。拒绝仅基于模糊怀疑的发现。每个实质问题都要通过相关源码、diff hunk、测试、配置、规范或 spec 确认。可用且相关时运行本地工具；工具失败与审查发现分开报告。
   完成标准：每个保留发现都有具体证据和可信失败模式。
6. 汇总。去重重叠发现，保留 lane 来源，按严重度排序，输出可执行检视意见和残余风险。
   完成标准：用户收到一个审查报告，而不是一堆原始 subtask 报告。

## Reference 包

- 读取 `references/code/` 做代码审查 lane：语法分析、功能分析和规范分析。
- 读取 `references/security/` 做安全审查 lane：认证与访问控制、输入与数据处理、密钥/依赖/供应链。

Reference 文件夹就是审查包。包级 subtask 必须在产出发现前读取该文件夹下每个文件。只有当文件夹过大或文件映射到独立 specialist 时，才按文件拆分。

## 工具角色

| 层级 | 优先使用 | 适合 | 必须用什么验证 |
| --- | --- | --- | --- |
| Diff 和所有权 | `git diff`、PR 文件、变更文件列表 | 审查范围、触及文件、hunk、commit | 源码读取和附近测试 |
| 代码导航和影响 | 可用时用 CodeMap MCP | 调用链、路由、函数影响、相关测试 | 源码、测试和 diff hunk |
| 精确证据 | `rg`、源码读取、测试、linter、typechecker | 符号、规范、失败行为、语法/类型问题 | 最小相关源码或测试集合 |
| 外部 PR 上下文 | 可用时用 GitHub MCP | PR 描述、review thread、关联 issue、checks | 本地 diff/source 或已获取 PR 产物 |

工具输出是证据，不是审查本身。linter 或 scanner 可以支撑发现，但最终报告必须解释问题为什么重要。

## 审查 Lane 契约

```yaml
role: code_review_lane
phase: review
objective: ""
review_pack: "references/code or references/security"
required_references:
  - "selected reference folder 中的每个文件"
target:
  diff_command: ""
  files: []
  specs_or_standards: []
edits_allowed: false
expected_output:
  format: lane_report
  required_fields:
    - lane
    - evidence_checked
    - findings
    - no_findings_statement
    - skipped_checks
    - residual_risk
stop_if:
  - "目标 diff 或文件不可用。"
  - "Required reference 文件无法读取。"
  - "发现需要在没有源码、测试、spec 或规范证据的情况下猜测。"
```

## 证据规则

- 可用时引用文件路径和行号。只有 PR 产物时，引用 hunk、文件和 commit 或 PR reference。
- 每个问题都要绑定到观察到的行为、破坏的契约、违反的规范、不安全数据流或可信利用路径。
- 区分硬失败和判断项。语法、类型、破坏测试、数据泄露和 auth bypass 在被证明时是硬失败；风格和设计异味是判断项，除非仓库规范规定为强制。
- 仓库既有规范优先于通用风格规则。如果仓库明确认可某模式，压制对该模式的通用反对意见。
- 不报告生成文件、vendored code、lockfile churn 或纯格式变更，除非它们造成真实缺陷或用户要求。
- 审查报告中不要建议重写，除非当前代码有具体缺陷或实质风险。

## 严重度

- `Critical`：可利用安全问题、数据丢失、认证绕过、密钥暴露、破坏性迁移风险或生产事故路径。
- `High`：很可能的功能失败、用户可见行为破坏、权限/数据边界违反，或缺少必要行为。
- `Medium`：边界情况 bug、近期维护成本风险、验证不完整、不安全默认值，或会导致缺陷的规范违反。
- `Low`：局部清晰度、小型规范漂移、弱测试或低风险清理。

## 输出形状

先给发现，摘要放后面。

```text
Findings:
- Severity - file:line - title
  Evidence: ...
  Impact: ...
  Recommendation: ...
  Lane: code/functionality/security/etc.

No Findings:
- lane: what was checked

Skipped Or Blocked:
- ...

Residual Risk:
- ...

Review Summary:
- target reviewed
- references loaded
- tools used
```

如果没有问题，要明确说明，并仍然列出审查 lane、证据和测试/工具缺口。
