# Cognitive Control Plane 评测包

这是 `skills/cognitive-control-plane` 的配套评测包中文版本。

本包评估的是该 skill 作为**控制平面路由器**的行为，而不是通用回答质量。核心问题是：它是否能正确改变下一步行动，同时不把自己变成流程仪式或实现工人。

## 本包测量什么

评测分三层：

1. **静态包完整性**
   - 被引用的文件存在
   - canonical/mirror 维护规则存在
   - 机器可读的 skill 路由图结构一致
   - 随包 guard 仍然强制预期的镜像策略

2. **行为路由**
   - activation 精度：只在过程控制有意义时介入
   - Tiny / Small / Large 分类
   - first-unsatisfied-surface 路由
   - stop-routing 行为
   - 编排、所有权、skill 路由和验证门

3. **语义质量**
   - 选中的控制面是否真的改善了下一步
   - assistant 是否避免在应该路由时直接解决任务
   - critique 是否使用显式标准
   - 最终 handoff 是否无需重新解释即可使用
   - response 是否避免了不必要的流程仪式

## 设计原则

本包遵循 eval-first 循环：

```text
cases -> run -> auto score -> judge/human review -> failure taxonomy
      -> root-cause diagnosis -> change -> regression run -> keep/revert
```

不要把每个失败都当成 prompt 失败。先分类为路由规则缺陷、case 歧义、缺失上下文、运行时 instrumentation 缺口、evaluator 缺陷或架构上限。

## 目录布局

```text
eval.zh-CN/
├── README.md
├── eval-design.md
├── taxonomy.yaml
├── rubric.yaml
├── cases/
│   ├── 01-activation-classification.yaml
│   ├── 02-surface-routing.yaml
│   ├── 03-orchestration-skill-routing.yaml
│   └── 04-negative-controls-maintenance.yaml
├── prompts/
│   ├── execution-wrapper.md
│   └── judge.md
└── meta-eval/
    └── judge-golden.yaml
```

脚本、schema、requirements 和示例结果仍复用英文包 `eval/` 下的原件。这个中文包只本地化人读内容和 case prompt；机器字段、枚举值、trace key、rubric id 和 skill 名称保持英文。

## 推荐安装位置

把本目录放在：

```text
skills/cognitive-control-plane/eval.zh-CN/
```

脚本仍从原英文包运行，并显式指向中文 cases 或中文 wrapper。

## 快速开始

安装唯一运行依赖：

```bash
python -m pip install -r eval/requirements.txt
```

运行静态检查：

```bash
python eval/scripts/static_checks.py \
  --skill-dir skills/cognitive-control-plane
```

为每个中文 case 生成一个 prompt：

```bash
python eval/scripts/build_case_prompts.py \
  --cases eval.zh-CN/cases \
  --wrapper eval.zh-CN/prompts/execution-wrapper.md \
  --out eval.zh-CN/.generated/prompts
```

用你要评估的模型或运行时执行这些 prompt。结果文件每行保存一个 JSON 对象，结构匹配 `eval/schemas/result.schema.json`。

为运行结果打分：

```bash
python eval/scripts/score.py \
  --cases eval.zh-CN/cases \
  --results eval/results/run-001.jsonl \
  --out-json eval/results/run-001-report.json \
  --out-md eval/results/run-001-report.md
```

## 证据等级

路由 trace 的可信度取决于来源。

使用可获得的最强证据：

1. `runtime_trace`: 从真实 skill/tool/hook 事件捕获
2. `executor_trace`: 由 agent runner 或编排框架捕获
3. `self_report`: 模型在 eval wrapper 下输出
4. `human`: 从 transcript 手动复原

不要让 self-reported trace 覆盖可观察行为。模型可能声称使用了 Context control，但可见地启动了 Adversarial review。

## 结果格式

每个 case result 应包含：

```json
{
  "case_id": "ACP-001",
  "evidence_source": "runtime_trace",
  "trace": {
    "activated": false,
    "classification": "Tiny",
    "active_surface": "none",
    "references_read": [],
    "orchestration_used": false,
    "required_skills": [],
    "next_action": "direct_answer",
    "asked_user_question": false,
    "strict_schema_during_exploration": false,
    "stopped_routing": true
  },
  "response": "..."
}
```

可选 judge 输出：

```json
{
  "judge": {
    "scores": {
      "materially_improves_next_action": 2,
      "thin_router_behavior": 2,
      "phase_appropriate_output": 2,
      "usable_handoff": 2,
      "anti_ceremony": 2
    },
    "flags": [],
    "notes": ""
  }
}
```

## 基线门槛

建议的首轮基线：

- 没有 hard-fail case
- auto score >= 90%
- activation false-positive rate <= 10%
- Large-miss rate = 0%
- surface-order violations = 0
- ownership-conflict violations = 0
- judge average >= 1.6 / 2.0
- 没有 evaluator meta-eval 回归

至少完成三次真实运行之前，不要冻结这些阈值。

## 迭代纪律

每次变更都记录：

```yaml
change_id: CH-001
hypothesis: ""
failure_modes_targeted: []
cases_expected_to_change: []
cases_expected_not_to_change: []
before_run: ""
after_run: ""
decision: keep | revert | inconclusive
regressions: []
```

每次变更尽量只保留一个 hypothesis。多个规则一起变化时，把它们记录为一个 change set，避免丢失归因。
