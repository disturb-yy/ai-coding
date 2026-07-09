---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---
# 输出控制

当工作准备进入交接、实现或机器消费时，使用输出控制。

## 阶段门

交付前不要强制严格 schema：

```text
Discovery -> 自由探索
Synthesis -> 结构化判断
Delivery -> 严格契约
```

完成标准：输出格式匹配当前阶段。

## 交付契约

选择下游消费者真正需要的最小契约：

```yaml
deliverable_type: ""
consumer: ""
required_fields: []
forbidden_content: []
validation: []
```

常见契约：

- 实施计划：文件、改动、测试、风险
- ADR：决策、背景、选项、后果
- 评审：发现、严重程度、文件引用、开放问题
- 交接：目标、状态、决策、约束、下一步
- 机器输出：schema、必填字段、验证规则

完成标准：下游消费者不需要重新解释即可使用输出。
