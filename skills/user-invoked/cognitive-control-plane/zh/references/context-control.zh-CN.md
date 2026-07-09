---
access:
  audience: human
  model_read: false
  model_write: true
  purpose: zh_mirror
---
# 上下文控制

当缺失或噪声上下文会导致跑偏时，使用上下文控制。

## 收集

只捕获会改变下一步动作的信息：

```yaml
goal: ""
current_state: ""
constraints: []
attempts: []
evidence: []
blocker: ""
expected_output: ""
allowed_change: []
forbidden_change: []
done_definition: ""
```

## 范围锁定

对于代码或项目工作，定义最小可行边界：

```yaml
target: []
allowed_change: []
forbidden_change: []
verification: []
```

完成标准：下一位执行者能判断要改什么、不要改什么，以及什么算完成。

## 状态转换

当对话切换阶段、上下文被污染，或反复纠正显示已经跑偏时使用。

持久化：

- 当前目标
- 已同意的约束
- 关键决策及理由
- 未解决问题
- 下一步动作

完成标准：新阶段可以开始，而不依赖陈旧或散落的对话历史。
