# Eval 设计

## 1. 被测系统

被测系统是 `cognitive-control-plane` skill。

该 skill 被评估为一个路由器，它有五个可观察职责：

1. 判断控制平面介入是否有用
2. 将工作分类为 Tiny、Small 或 Large
3. 选择第一个会实质改变下一步行动的控制面
4. 只在需要运行时协调时使用编排状态
5. 当应该开始直接执行或交付时停止路由并 hand off

该 eval 刻意**不**评分隐藏推理。

## 2. 轴

### 轴 A：Activation

```text
required  -> 不激活会实质损害下一步行动
forbidden -> 激活没有价值，只会制造仪式
optional  -> 直接处理或轻量控制步骤都可以接受
```

### 轴 B：工作分类

```text
Tiny  -> 没有实质性 worker/task skill 会增加价值
Small -> 所有 Small 条件都被证明为真
Large -> 一个 Large 信号就足够；不确定性升级
```

### 轴 C：Active surface

```text
none
context
epistemic
adversarial
output
```

当多个 surface 同时适用时，期望标签是最早尚未满足的那个。

### 轴 D：运行时编排

```text
required
forbidden
optional
```

编排不是第五个 surface。它单独评估。

### 轴 E：Handoff

最终下一步行动应是：

```text
direct_answer
direct_execute
ask_blocking_question
route_skill
delegate_read_only
delegate_write
verify
deliver
```

### 轴 F：反模式

eval 跟踪：

- all-surfaces ceremony
- 未证明所有条件就选择 Small
- 漏掉 Large 信号
- assumptions 之前就 critique
- exploration 阶段使用刚性 schema
- 向用户询问仓库可回答的事实
- 实现或交付应开始后仍继续路由循环
- 隐式 skill 依赖
- 写入所有权重叠
- 未审阅 specialist output
- mirror reads
- stale mirrors

## 3. Case 组成

golden set 有意混合：

- positive controls：必须介入
- negative controls：不得介入
- adversarial controls：措辞相似但路由不同
- order controls：多个 surface 适用；只有最早未满足者胜出
- orchestration controls：依赖和所有权很重要
- maintenance controls：hooks 和 mirror policy 在语义路由之外测试

不要只评估“是否介入”。一个总是激活的路由器看起来安全，但实际没有用。

## 4. 自动检查 vs judge 检查

### 自动检查

对以下内容使用 trace assertions：

- classification
- active surface
- orchestration on/off
- required skill selection
- next action
- reference-loading bounds
- stop-routing state
- ownership conflicts
- required validation gates

### Judge 或 human

对以下内容使用语义判断：

- 介入是否实质改善了下一步行动
- router 是否保持轻薄
- evidence/assumption 处理是否充分
- critique 是否基于标准
- 最终交付物是否可用
- response 是否变成流程仪式

## 5. Hard-fail 策略

默认 hard-fail：

- 将 Large-risk task 分类为 Small
- 在重叠或依赖工作中漏掉 required orchestration
- 允许写入所有权重叠
- 当任务依赖 specialized skill 时跳过它
- 把未审阅 specialist output 当成最终事实
- 在 required verification 前 finalize
- 读取 Chinese mirror
- 修改 canonical 文件后留下 stale/missing mirror

默认 quality flags：

- 不必要的额外 reference read
- 轻微过度结构化
- 过度冗长
- 非关键 over-delegation
- handoff 薄弱但仍可用

## 6. 已知歧义策略

case 可以标记 `ambiguous: true`。

Ambiguous cases：

- 不计入 hard pass/fail
- 仍出现在 disagreement reports 中
- 用来审查 case、taxonomy 或 skill 是否需要修订

失败运行后绝不要静默改写 golden expectation。记录原因。

## 7. 回归策略

只有满足以下条件才接受变更：

- 目标失败得到改善
- 没有 hard-fail 回归
- negative controls 保持稳定
- meta-eval 仍通过

保留失败尝试和 revert 后的报告。
