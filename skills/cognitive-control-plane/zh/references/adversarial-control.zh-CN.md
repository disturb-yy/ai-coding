---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---

# 对抗控制

当方案已经具体到可以被攻击时，使用对抗控制。

## 基于标准的批判

先选择标准，再开始批判：

```yaml
review_dimensions:
  - "它是否解决真实问题？"
  - "它是否重复已有能力？"
  - "它是否增加不必要复杂度？"
  - "收益是否可验证？"
  - "是否存在更简单选项？"
  - "维护成本是否可接受？"
```

完成标准：批评绑定到明确标准，而不是态度。

## 红队评审

只攻击方案，然后把有效失败点和噪声分开：

```yaml
attack_report:
  valid_failures: []
  weak_or_irrelevant_attacks: []
  mitigations: []
  residual_risk: []
```

完成标准：方案被修改、增加缓解措施，或显式留下残余风险。

## 事前尸检

提示词：

```text
假设这个方案上线六个月后失败了。
列出三个最可能的失败原因、被忽略的早期预警信号，以及每个原因最低成本的预防步骤。
```

完成标准：每个主要失败模式都有早期预警信号和预防步骤。
