---
name: problem-diagnosis
description: Diagnose bugs, failures, regressions, unexpected behavior, crashes, hangs, build/test failures, performance problems, and ambiguous technical issues without modifying code. Use when the user asks to investigate, debug, locate the problem, explain why something fails, find a root cause, create reproduction steps, or prepare a handoff for a later fix. Produces reusable diagnosis artifacts and forbids direct code changes.
---

# Problem Diagnosis

## Purpose

Investigate problems without applying fixes. Preserve enough evidence, context,
and reasoning for a later model or skill to continue from the artifact instead
of rediscovering the same facts.

## Hard Boundaries

- Do not modify source files, generated files, lockfiles, config files, or tests.
- Do not call `apply_patch` or use shell writes to change project files.
- Do not format, tidy, migrate, regenerate, commit, revert, or delete files.
- Do not produce a fix artifact such as `diagnosis_fix`.
- Do not change workflow lifecycle state.
- Do not present a guess as a root cause. Mark uncertainty explicitly.

Allowed actions:

- Read files, logs, traces, configs, and git history.
- Run non-mutating diagnostic commands.
- Run build/test/reproduction commands when they do not intentionally rewrite
  project files.
- Create workflow diagnosis artifacts and logs when the active workflow requires
  persistence.

## Workflow

Execute these steps in order. Skip a step only when the artifact records the
reason.

1. Clarify the problem.
2. Identify the likely domain or runtime.
3. Load only the relevant reference file.
4. Build or document a reproduction path.
5. Collect evidence.
6. Inspect the smallest relevant code or system surface.
7. Generate and test falsifiable hypotheses.
8. Record root cause status and handoff.

## Domain Routing

Identify the domain from file extensions, stack traces, package manifests,
commands, error formats, logs, and user wording.

Load at most the relevant references:

- Go runtime, Go tests, `go build`, `go.mod`, goroutine, race, panic, pprof:
  read `references/go.md`.
- Python exceptions, pytest, packaging, async, import, virtualenv issues:
  read `references/python.md`.
- JavaScript or TypeScript runtime, Node, browser, React, npm/pnpm/yarn,
  bundler, hydration, async UI issues: read `references/javascript.md`.
- SQL, migrations, transactions, indexes, ORM behavior, data correctness:
  read `references/database.md`.
- HTTP, DNS, TLS, proxies, timeouts, retries, queues, distributed calls:
  read `references/network.md`.

If no domain is clear, continue with the generic workflow and record
`domain: "unknown"`.

## Reproduction Discipline

Prefer a single command or short step list that demonstrates the exact symptom.
Record whether the problem is:

- `reproduced`: observed locally with evidence.
- `partial`: related failure observed but not the exact symptom.
- `not_reproduced`: attempted and did not occur.
- `unknown`: not attempted or insufficient information.

When the problem is not reproduced, still record attempted commands, outputs,
environment facts, missing inputs, and next probes.

## Hypothesis Rules

Before naming a root cause, produce 3-5 ranked hypotheses unless the evidence
already proves a single cause. Each hypothesis must include:

- The suspected mechanism.
- Evidence for and against it.
- A command, file, log, or observation that would falsify it.
- Confidence from `0.0` to `1.0`.

## Output Artifact

Return a JSON artifact envelope. Use the workflow artifact type when the current
phase requires one:

- `problem_understanding` for Problem Understanding.
- `root_cause_analysis` for Root Cause Analysis.
- `regression_check` for Regression Check when verifying without modifying code.
- `problem_diagnosis` for standalone diagnosis outside a workflow phase.

For standalone or combined diagnosis, use this content shape:

```json
{
  "artifact_type": "problem_diagnosis",
  "created_at": "<ISO 8601 timestamp>",
  "phase": "Problem Diagnosis",
  "content": {
    "domain": "go|python|javascript|database|network|unknown",
    "problem_statement": "",
    "observed_behavior": "",
    "expected_behavior": "",
    "environment": {},
    "reproduction": {
      "status": "reproduced|partial|not_reproduced|unknown",
      "commands": [],
      "steps": [],
      "evidence": []
    },
    "important_facts": [],
    "files_examined": [],
    "call_chain_or_data_flow": [],
    "hypotheses": [
      {
        "summary": "",
        "confidence": 0.0,
        "evidence_for": [],
        "evidence_against": [],
        "falsification_probe": "",
        "result": "confirmed|rejected|untested|inconclusive"
      }
    ],
    "root_cause": {
      "status": "identified|suspected|not_found",
      "summary": "",
      "location": "",
      "confidence": 0.0,
      "evidence": []
    },
    "handoff": {
      "recommended_next_skill": "code-implementation|go-coding|regression-check|unknown",
      "suggested_fix_direction": "",
      "validation_command": "",
      "open_questions": [],
      "blocked_reason": ""
    }
  }
}
```

## Handoff Requirements

Always include handoff information, even when no root cause is found. The next
model must know:

- What has already been tried.
- Which facts are trustworthy.
- Which hypotheses were rejected.
- What remains unknown.
- Which command or scenario should validate a future fix.
