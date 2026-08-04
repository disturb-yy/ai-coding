# Cognitive Control Plane Eval Pack

A companion evaluation pack for `skills/cognitive-control-plane`.

This pack evaluates the skill as a **control-plane router**, not as a general answer-quality benchmark. The core question is whether the skill changes the next action correctly without becoming a ceremony or an implementation worker.

## What this pack measures

Three layers are evaluated separately:

1. **Static package integrity**
   - referenced files exist
   - canonical/mirror maintenance rules are present
   - the machine-readable skill routing map is structurally consistent
   - bundled guards still enforce the intended mirror policy

2. **Behavioral routing**
   - activation precision: intervene only when process control matters
   - Tiny / Small / Large classification
   - first-unsatisfied-surface routing
   - stop-routing behavior
   - orchestration, ownership, skill-routing, and verification gates

3. **Semantic quality**
   - did the chosen control surface actually improve the next move?
   - did the assistant avoid solving the task when it should route?
   - did critique use explicit criteria?
   - is the final handoff usable without reinterpretation?
   - did the response avoid unnecessary ceremony?

## Design principles

The pack follows an eval-first loop:

```text
cases -> run -> auto score -> judge/human review -> failure taxonomy
      -> root-cause diagnosis -> change -> regression run -> keep/revert
```

Do not treat every failure as a prompt failure. Classify it first as a routing rule defect, case ambiguity, missing context, runtime instrumentation gap, evaluator defect, or architecture ceiling.

## Work-item scheduling regression set

`06-work-item-scheduler.yaml` protects the generic scheduler contract. A work
item is the durable unit accepted as an `issue`, `request`, `transaction`, or
`ticket`; a run is one session attempt. The cases require lease and dependency
gates, checkpoint/handoff/hard-stop thresholds at 40%/45%/50%, continuation of
the same work item in a fresh session, Programmatic Tool Calling boundaries,
and evidence-gated terminal states.

## Directory layout

```text
eval/
├── README.md
├── eval-design.md
├── taxonomy.yaml
├── rubric.yaml
├── requirements.txt
├── cases/
│   ├── 01-activation-classification.yaml
│   ├── 02-surface-routing.yaml
│   ├── 03-orchestration-skill-routing.yaml
│   ├── 04-negative-controls-maintenance.yaml
│   ├── 05-reviewer-enforcement.yaml
│   └── 06-work-item-scheduler.yaml
├── prompts/
│   ├── execution-wrapper.md
│   └── judge.md
├── schemas/
│   ├── result.schema.json
│   └── run-record.schema.json
├── scripts/
│   ├── static_checks.py
│   ├── score.py
│   └── build_case_prompts.py
├── meta-eval/
│   └── judge-golden.yaml
├── examples/
│   └── sample-results.jsonl
└── results/
```

## Recommended install location

Place this directory at:

```text
skills/cognitive-control-plane/eval/
```

The scripts assume the skill root is the parent directory of `eval/`, but you can override it with `--skill-dir`.

## Quick start

Install the only runtime dependency:

```bash
python -m pip install -r eval/requirements.txt
```

Run static checks:

```bash
python eval/scripts/static_checks.py \
  --skill-dir skills/cognitive-control-plane
```

Materialize one prompt per case:

```bash
python eval/scripts/build_case_prompts.py \
  --cases eval/cases \
  --wrapper eval/prompts/execution-wrapper.md \
  --out eval/.generated/prompts
```

Execute those prompts with the model/runtime you want to evaluate. Save one JSON object per line in a result file matching `schemas/result.schema.json`.

Score the run:

```bash
python eval/scripts/score.py \
  --cases eval/cases \
  --results eval/results/run-001.jsonl \
  --out-json eval/results/run-001-report.json \
  --out-md eval/results/run-001-report.md
```


## Run a baseline with Codex CLI

The recommended first run is an **isolated routing baseline**. The runner creates a small temporary Git workspace, exposes only this repository-scoped skill through `.agents/skills`, runs every case with `codex exec`, stores raw JSONL events, combines structured final results, and invokes the scorer.

First place this eval pack under the skill:

```text
skills/cognitive-control-plane/eval/
```

Verify Codex and choose a fixed model for reproducibility:

```bash
codex --version
codex debug models
```

Smoke-test 10 cases:

```bash
python skills/cognitive-control-plane/eval/scripts/run_codex_baseline.py \
  --skill-dir skills/cognitive-control-plane \
  --model <EXACT_MODEL_ID> \
  --limit 10 \
  --run-id baseline-smoke
```

Run the full 62 runtime cases:

```bash
python skills/cognitive-control-plane/eval/scripts/run_codex_baseline.py \
  --skill-dir skills/cognitive-control-plane \
  --model <EXACT_MODEL_ID> \
  --run-id baseline-v0
```

Run a single case while debugging:

```bash
python skills/cognitive-control-plane/eval/scripts/run_codex_baseline.py \
  --skill-dir skills/cognitive-control-plane \
  --model <EXACT_MODEL_ID> \
  --case ACP-107 \
  --run-id debug-ACP-107
```

The runner uses:

```text
codex
--ask-for-approval never
exec
--ephemeral
--sandbox read-only
--json
--output-schema ...
--output-last-message ...
```

By default it also uses `--ignore-user-config` and `--ignore-rules`, and isolates `$HOME` while preserving the real `CODEX_HOME` for authentication. This prevents unrelated user skills and local config from contaminating the skill baseline.

If your Codex setup depends on custom providers or user configuration, add:

```bash
--use-user-config
```

That produces an **environment baseline**, not a pure isolated skill baseline.

Outputs:

```text
eval/results/<run-id>/
├── run.json
├── static-checks.json
├── results.jsonl
├── report.json
├── report.md
├── raw/
│   └── ACP-xxx.jsonl
└── final/
    └── ACP-xxx.json
```

The first baseline uses `evidence_source=self_report`. Raw `codex exec --json` streams are retained so a later adapter can derive stronger `runtime_trace` evidence from actual events and hooks.


### Structured output schema compatibility

The Codex baseline schema is strict: every object sets `additionalProperties: false`, and every declared property is listed in `required`. The static preflight checks this before runtime cases start. This avoids losing an entire run to a response-format schema rejection.

### Recommended progression

```text
baseline-smoke (10 cases)
    -> fix harness only
baseline-v0 (62 cases)
    -> diagnose failures
one hypothesis / one change
    -> regression run on targeted + negative-control cases
full baseline-v1
```

Do not tune the skill against the first 10 smoke cases. The smoke run validates the harness, not the skill.

## Evidence levels

A routing trace is only as trustworthy as its source.

Use the strongest available evidence:

1. `runtime_trace`: captured from actual skill/tool/hook events
2. `executor_trace`: captured by an agent runner or orchestration framework
3. `self_report`: emitted by the model under the eval wrapper
4. `human`: manually reconstructed from the transcript

Never let a self-reported trace override observable behavior. A model can say it used Context control while visibly launching an Adversarial review.

## Result format

Each case result should contain:

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

Optional judge output may be added:

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

## Baseline gates

Recommended first baseline:

- no hard-fail cases
- auto score >= 90%
- activation false-positive rate <= 10%
- Large-miss rate = 0%
- surface-order violations = 0
- ownership-conflict violations = 0
- judge average >= 1.6 / 2.0
- no evaluator meta-eval regressions

Do not freeze these thresholds until at least three real runs exist.

## Iteration discipline

For every change, record:

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

Prefer one hypothesis per change. When several rules change together, record them as a change set so attribution is not lost.
