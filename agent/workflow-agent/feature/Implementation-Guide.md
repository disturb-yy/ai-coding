
# Workflow-Agent Implementation Guide

## 目标

实现一个跨平台 AI Workflow 系统，支持：

- Codex
- OpenCode

核心能力：

- 多阶段软件开发流程编排
- SubAgent / Skill 调度
- Artifact 驱动通信
- Codemap + Understand 优先分析代码库
- 最小上下文原则

---

# 📁 目录结构要求

请严格创建以下结构：

```
agent/workflow-agent/
├── AGENT.md
├── workflow/
│   └── WORKFLOW.md
├── codex/
│   └── AGENTS.md
└── opencode/
    └── workflow-agent.md
```

---

# 1️⃣ workflow/WORKFLOW.md（核心规则）

## 要求

实现一个**平台无关 workflow 规范文件**，必须包含：

### 必须内容

- Workflow Lifecycle
  - Created → Planning → Running → Verifying → Completed
  - Failure: Running → Failed

- Workflow Phases（必须顺序执行）
  1. Requirement Analysis
  2. Project Understanding
  3. Solution Design
  4. Implementation
  5. Testing
  6. Verification
  7. Summary

- Artifact 机制（必须 JSON 结构）
  - requirement_analysis
  - project_understanding
  - solution_design
  - implementation_result
  - test_result
  - verification_result
  - workflow_summary

- Context Rules（必须）
  - 禁止加载整个仓库
  - 优先 Codemap → Understand → Source Code
  - 每阶段只允许最小上下文

- Verification Rules
  - build 必须通过
  - test 必须通过
  - acceptance criteria 必须满足

- Fix Loop
  - 最大重试 3 次
  - failure → root cause → fix → verify

---

# 2️⃣ AGENT.md（OpenCode / Codex 通用说明）

## 要求

实现一个**极简 orchestrator agent**

必须做到：

### 职责

- 读取 workflow/WORKFLOW.md
- 控制流程执行
- 选择 skill（不实现 skill）
- 控制上下文大小
- 负责最终总结

### 禁止

- 不写业务代码
- 不写测试代码
- 不做代码分析实现细节

### 必须包含

- Workflow execution order
- Tool priority rule:
  1. Codemap
  2. Understand
  3. Source Code

- Artifact-driven execution
- Completion rules

---

# 3️⃣ codex/AGENTS.md（Codex适配层）

## 要求

这是 Codex 的入口控制文件

必须：

- 引用 workflow/WORKFLOW.md
- 定义 workflow 执行顺序
- 强调 artifact 驱动
- 强调最小上下文原则

## 禁止

- 不定义 Agent 行为细节
- 不重复 WORKFLOW.md 内容

## 内容结构

- Role: workflow controller
- Responsibilities
- Workflow execution order
- Tool priority
- Completion rules

---

# 4️⃣ opencode/workflow-agent.md（OpenCode适配层）

## 要求

定义 OpenCode agent：

### 必须包含

- name: workflow-agent
- description: orchestrates workflow execution
- reference: workflow/WORKFLOW.md

### 职责

- orchestration
- sub-agent delegation
- artifact coordination
- verification

### 必须规则

- follow workflow strictly
- delegate to skills
- never implement business logic directly

---

# ⚠️ 关键设计约束（非常重要）

## 1. Workflow 必须唯一

workflow/WORKFLOW.md 是：

> 单一事实源（Single Source of Truth）

---

## 2. Agent 不允许包含 workflow 逻辑

Agent 只能：

- 调度 workflow
- 不实现 workflow

---

## 3. Codex / OpenCode 只做适配

不要在两个系统重复定义 workflow。

---

## 4. Artifact 是跨系统通信协议

所有阶段必须输出 JSON Artifact：

```json
{
  "artifact_type": "",
  "phase": "",
  "content": {}
}
```

---

## 5. 不要设计 Skill（下一阶段再做）

当前任务只做：

- workflow
- agent
- adapter

---

# 🎯 验收标准

完成后必须满足：

### 结构正确

- 文件齐全
- 路径正确

### 逻辑正确

- workflow 可独立阅读
- agent 不包含业务逻辑
- codex/opencode 无重复 workflow

### 架构正确

- workflow = core
- agent = orchestrator
- adapter = platform glue

---

# 🚀 输出要求

请生成完整文件内容，并确保：

- markdown 可直接使用
- 不省略关键字段
- 不添加多余设计
- 保持简洁可执行
```

---

# 如果你下一步要做什么（建议）

下一阶段你应该让另一个AI做这个：

> 👉 生成 Skill 层（requirement-analysis / project-understanding）

因为你现在已经完成：

```text
Workflow（规则）
Agent（调度器）
Adapter（平台层）
```

下一步才是：

```text
Execution能力层（Skills）
```

如果你需要，:contentReference[oaicite:0]{index=0}。