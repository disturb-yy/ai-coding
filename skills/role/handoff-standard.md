---
type: reference
title: 角色交接标准
description: 在 Cognitive Control Plane 角色之间交接任务、证据、工具路由与失败状态的统一合同。
tags: [cognitive-control-plane, handoff, tool-orchestration, programmatic-tool-calling]
intent: 角色交接合同与工具路由边界
status: stable
generated: { by: codex/gpt-5, at: 2026-08-01T01:37:30+08:00 }
sources:
  - id: orchestration-map
    resource: ../user-invoked/cognitive-control-plane/config/skill-orchestration-map.yaml
    title: Cognitive Control Plane skill orchestration map
  - id: openai-programmatic-tool-calling
    resource: https://developers.openai.com/api/docs/guides/tools-programmatic-tool-calling
    title: Programmatic Tool Calling | OpenAI API
---

# 角色交接标准

## 交接合同

每次角色交接都应给出紧凑、可验证的合同：

```yaml
handoff:
  from_role: ""
  to_role: ""
  phase: context | review | implementation
  objective: ""
  inputs:
    - resource: ""
      purpose: ""
  completed_work: []
  expected_result:
    fields: []
    evidence: []
  tool_route:
    mode: direct | programmatic | mixed
    eligible_tools: []
    direct_only_tools: []
  validation: []
  stop_conditions: []
  retry_policy:
    max_transient_retries: 0
    do_not_repeat_completed_or_side_effecting_calls: true
```

只交付下游需要的结构化结果与证据，不转发原始大段工具输出。若任一必需字段、来源或验证缺失，返回结构化 `blocked` 状态，并说明缺失项和安全的下一步。

## 工具路由

- 使用 **programmatic**：仅限控制流可预测的有界阶段，例如独立只读调用的并发、过滤、去重、连接、聚合或结构验证；必须指定可调用工具、输入/输出字段、结果形状、停止条件和重试上限。
- 使用 **direct**：单次调用、每一步需要新的语义判断、批准或写入敏感操作、最终引用与原生工件核验。
- 使用 **mixed**：只允许在合同中预先定义一次边界；programmatic 阶段先压缩中间结果，direct 阶段负责语义判断、授权与最终交付。
- 工具结果应是紧凑结构化数据；未知输出形状保持 direct，直到模型能够检查结果。
- 所有工具调用仍需在应用层验证参数和权限；高影响操作始终要求应用层审批。

## 程序化调用的运行时续接

仅当交接对象实现 Responses API 的 Programmatic Tool Calling 时适用：

- 明确 `allowed_callers`；只有列入 `eligible_tools` 的工具可由程序调用。
- 为可预测结果声明输入参数与 `output_schema`；程序只依据已声明字段处理结果。
- 客户端工具调用返回时保留原始 `call_id` 与 `caller`，以恢复正确的程序；持续处理直到出现最终消息。
- 返回完整程序运行链所需的响应项；无状态续接必须按原顺序保留程序、推理、函数调用及其输出。

## 完成与审计

交接完成前确认：目标已满足、预期字段与证据齐全、验证已通过或已明确阻塞、工具路由符合合同、未重复有副作用操作。记录实际工具、重试次数、跳过检查和残余风险；这些记录是后续角色的事实输入，而非再次推测的对象。
