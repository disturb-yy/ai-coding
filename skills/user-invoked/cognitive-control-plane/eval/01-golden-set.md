# Golden Set — cognitive-control-plane

7 个固定测试场景，覆盖 Tiny/Small/Large 分类 + 四个控制面 + 边界升级。

每次修改 Skill 后，在新对话中跑完这 7 个场景，对照 Rubric 评分。

---

## S-01 — Tiny: 简单信息查询

**输入**：
```
这个项目有几个目录？
```

**期望行为**：
- 分类为 Tiny
- 直接回答，不触发任何控制面
- 不出现 "Context control" / "Epistemic control" 等路由语言
- 不创建 task contract

**测试目标**：确认 Skill 不会对简单问题过度仪式化

---

## S-02 — Small: 单文件明确修改

**输入**：
```
README.md 第三行有个拼写错误，"implemntation" 应该是 "implementation"，帮我修一下。
```

**期望行为**：
- 分类为 Small
- 确认 6 个 Small 条件（或至少不触发 Large 升级）
- 直接执行修改
- 不创建 task contract（Small delegation 除外）

**测试目标**：确认 Small 分类正确，不升级为 Large

---

## S-03 — Large-Context: 模糊需求

**输入**：
```
帮我优化一下这个系统。
```

**期望行为**：
- 识别需求模糊 → 升级 Large（一个 Large signal 足够）
- 选择 Context control（目标/状态/约束/边界不明确）
- 不直接开始实现
- 可能追问或读取项目结构来澄清上下文

**测试目标**：模糊需求不跳过 Context control 直接实现

---

## S-04 — Large-Epistemic: 隐藏假设

**输入**：
```
因为数据库查询太慢了，帮我给所有 API 加一个 Redis 缓存层。
```

**期望行为**：
- 识别隐藏假设（"数据库慢"是未经验证的诊断，不是事实）
- 选择 Epistemic control（假设/证据/因果链需检查）
- 追问证据或建议先 profiling，而非直接接受"加缓存"的方案
- 不直接开始实现 Redis 缓存层

**测试目标**：不对未经验证的假设直接行动

---

## S-05 — Large-Adversarial: 需要攻击的成熟计划

**输入**：
```
我设计了一个微服务架构方案，服务 A 通过 RabbitMQ 发消息给服务 B，
B 处理后写入 PostgreSQL 再通知服务 C。
帮我开始实现服务 A 和 B。
```

**期望行为**：
- 识别这是一个需要审查的成熟设计
- 选择 Adversarial control（具体计划需要攻击）
- 先做 red-team/pre-mortem 分析（消息丢失？B 挂了怎么办？）
- 不直接开始实现

**测试目标**：不对未审查的设计直接实现

---

## S-06 — Large-Orchestration: 多模块编排

**输入**：
```
我需要给用户系统添加双因素认证功能。
包括：数据库 schema 变更（user 表加字段）、
API 新增 /auth/2fa/setup 和 /auth/2fa/verify 端点、
前端加设置页面和登录时的验证码输入框、
以及单元测试和集成测试。
```

**期望行为**：
- 分类为 Large（多模块、多所有权边界、数据模型变更）
- 使用 orchestration state
- 创建 task contract，分配 ownership boundaries
- 编排多 worker 或分阶段执行
- 不自己默默实现全部

**测试目标**：多模块任务正确使用编排，ownership 边界明确

---

## S-07 — 升级测试：从 Small 到 Large

**输入**：
```
把 utils/helpers.js 里的 formatDate 函数改成用 dayjs 库。
```

然后在对话中途（第二轮）追加：
```
对了，这个改动会影响到 user-profile、dashboard、reports、admin-panel
这四个页面，它们都调用了 formatDate。
```

**期望行为**：
- 第一轮可能分类为 Small（单文件、明确范围）
- 第二轮识别 scope 扩展 → 升级为 Large
- stop_if 条件触发
- 重新评估：可能需要 orchestration（多文件）或至少 Adversarial（regression 风险）

**测试目标**：不确定时升级，而不是坚持初始分类
