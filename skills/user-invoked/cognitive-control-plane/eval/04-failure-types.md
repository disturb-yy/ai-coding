# Failure Types — cognitive-control-plane

把"这次路由不好"拆成具体可修复的问题类别。

每种 Failure Type：表现 / 后果 / 修复方向 / 来源

---

## FT-01 — 假 Tiny (Fake Tiny)

- **表现**: 应该走控制面分析的任务被直接回答
- **后果**: 模糊需求被当作明确指令执行；隐藏假设未被暴露
- **修复方向**: 降低 Tiny 的触发阈值；任何含"优化/改进/设计/架构"关键词的请求自动升级
- **来源**: SKILL.md §Work Classification Gate — Tiny 定义太宽

## FT-02 — 假 Small (Fake Small)

- **表现**: 被分类为 Small 直接执行，但实际有跨模块影响、安全风险或未明确的依赖
- **后果**: 引入 regression、打破其他模块、遗漏安全审查
- **修复方向**: 加强 Small 的条件检查 prompt；增加"如果有任何不确定，升级"的权重
- **来源**: SKILL.md §Small — 条件列表不够明确，或模型忽略了 stop_if

## FT-03 — 假 Large (Fake Large)

- **表现**: 简单任务走了完整的 Large 流程（分类→选面→读 reference→task contract→编排）
- **后果**: 浪费 token 和时间；过度仪式感让用户困惑
- **修复方向**: 加强 SKILL.md §Do Not 第一条（"Do not turn every task into a full four-stage ceremony"）
- **来源**: SKILL.md 入口条件太敏感

## FT-04 — 路由选错面 (Wrong Surface)

- **表现**: 选的控制面不是真正的瓶颈
  - 本应 Context（需求不清）却选了 Adversarial（攻击方案）
  - 本应 Epistemic（假设有问题）却直接 Output（生成交付物）
- **后果**: 在错误的方向上深入，浪费精力且不解决问题
- **修复方向**: 加强四面的区分描述；增加"为什么选 X 而不是 Y"的自我检查
- **来源**: SKILL.md §Route — 四面描述不够差异化

## FT-05 — 跳过门禁 (Gate Skip)

- **表现**: 没有显式的任务分类声明就直接开始工作
- **后果**: 后续行为无法评估——不知道模型认为自己是在做 Small 还是 Large
- **修复方向**: 在 SKILL.md §Operating Steps 第一步加强；考虑在 hook 中检测
- **来源**: SKILL.md 步骤执行不严格

## FT-06 — 永不交棒 (Routing Forever)

- **表现**: 在路由阶段循环：选了面 A → 觉得不够 → 选面 B → 又加 orchestration → 还在想...
- **后果**: 用户等不到实际输出
- **修复方向**: 明确 exit condition；增加"路由最多 N 步后必须交棒"的规则
- **来源**: SKILL.md §exit_conditions 不够强制

## FT-07 — 偷当实现者 (Orchestrator-as-Implementer)

- **表现**: 主 agent 分类为 Large，但跳过 delegation contract，自己默默实现
- **后果**: 没有 ownership boundary、没有 review gate、没有 stop condition——和没用 Skill 一样
- **修复方向**: 加强 Implementation Guard 的语言；在 hook 中检测 Large→Write 的跳转
- **来源**: SKILL.md §Implementation Guard — 约束不够强

## FT-08 — 合同不全 (Incomplete Contract)

- **表现**: Task contract 缺少关键字段（如没有 stop_if、没有 ownership、没有 expected_output）
- **后果**: Worker 可能越界修改、不知道何时停止、输出格式不匹配
- **修复方向**: 在 SKILL.md §Task Contract 中标注必需字段；考虑模板强制
- **来源**: SKILL.md §Task Contract

## FT-09 — 幽灵实现 (Phantom Implementation)

- **表现**: 没有 delegation contract，也没有显式的 direct-implementation exception 声明，直接做了 Large 实现
- **后果**: 最严重的失败——完全绕过了 control plane
- **修复方向**: 这是 R-A3 hard_fail 的场景。修复 Implementation Guard 的措辞和 hook 检测
- **来源**: SKILL.md §Implementation Guard 被忽略

---

## Failure Types 诊断流程

```
观察到"路由不好"
  ↓
1. 有没有显式任务分类？ → 没有 → FT-05 跳过门禁
  ↓ 有
2. 分类对吗？ → Tiny/Small/Large？
  ↓
  错了 → FT-01/02/03（假分类）
  ↓ 对
3. 是 Large 吗？→ 是 → 有 Guard 吗？→ 没有 → FT-07/09
  ↓ 有         ↓ 不是
4. 控制面选对了吗？ → 不对 → FT-04
  ↓ 对
5. 是否该交棒了？ → 还在路由 → FT-06
  ↓ 交棒了
6. Contract 完整吗？ → 不完整 → FT-08
  ↓
✅ 通过
```
