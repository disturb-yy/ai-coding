# Eval Design

## 1. System under test

The system under test is the `cognitive-control-plane` skill.

The skill is evaluated as a router with five observable responsibilities:

1. decide whether control-plane intervention is useful
2. classify work as Tiny, Small, or Large
3. select the first control surface that materially changes the next action
4. use orchestration state only when runtime coordination is needed
5. stop routing and hand off when direct execution or delivery should begin

The eval deliberately does **not** score hidden reasoning.

## 2. Axes

### Axis A — Activation

```text
required  -> not activating materially harms the next action
forbidden -> activation adds no value and creates ceremony
optional  -> either direct handling or a light control step is acceptable
```

### Axis B — Work classification

```text
Tiny  -> no substantive worker/task skill adds value
Small -> every Small condition is proven true
Large -> one Large signal is enough; uncertainty upgrades
```

### Axis C — Active surface

```text
none
context
epistemic
adversarial
output
```

When several surfaces apply, the expected label is the earliest unsatisfied one.

### Axis D — Runtime orchestration

```text
required
forbidden
optional
```

Orchestration is not a fifth surface. It is evaluated separately.

### Axis E — Handoff

The final next action should be one of:

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

### Axis F — Anti-patterns

The eval tracks:

- all-surfaces ceremony
- Small chosen without proving all conditions
- Large signal missed
- critique before assumptions
- rigid schema during exploration
- asking the user for repository-answerable facts
- routing loops after implementation/delivery should start
- implicit skill dependency
- overlapping write ownership
- unreviewed specialist output
- mirror reads
- stale mirrors

## 3. Case composition

The golden set is intentionally mixed:

- positive controls: intervention must occur
- negative controls: intervention must not occur
- adversarial controls: similar wording but different route
- order controls: multiple surfaces apply; only earliest unsatisfied wins
- orchestration controls: dependencies and ownership matter
- reviewer-enforcement controls: mandatory triggers, actor independence, blocking gates, re-review, and artifact version freshness matter
- maintenance controls: hooks and mirror policy are tested outside semantic routing

Do not evaluate only "did it intervene?" A router that activates on every case can look safe while being useless.

## 4. Automatic vs judge checks

### Automatic

Use trace assertions for:

- classification
- active surface
- orchestration on/off
- required skill selection
- next action
- reference-loading bounds
- stop-routing state
- ownership conflicts
- required validation gates

### Judge or human

Use semantic judgment for:

- whether intervention materially improved the next action
- whether the router remained thin
- whether evidence/assumption handling was adequate
- whether critique was criterion-based
- whether the final deliverable was usable
- whether the response became ceremony

## 5. Hard-fail policy

Hard-fail by default:

- classify a Large-risk task as Small
- miss required orchestration for overlapping or dependent work
- allow overlapping write ownership
- skip a required specialized skill when the task depends on it
- accept unreviewed specialist output as final truth
- finalize before required verification
- read a Chinese mirror
- leave a modified canonical file with a stale/missing mirror

Quality flags by default:

- unnecessary extra reference read
- mild over-structuring
- excessive verbosity
- non-critical over-delegation
- weak but still usable handoff

## 6. Known ambiguity policy

A case may be tagged `ambiguous: true`.

Ambiguous cases:

- do not count toward hard pass/fail
- still appear in disagreement reports
- are reviewed to decide whether the case, taxonomy, or skill needs revision

Never silently rewrite a golden expectation after a failed run. Record the reason.

## 7. Regression policy

A change is accepted only when:

- targeted failures improve
- no hard-fail regression appears
- negative controls stay stable
- meta-eval still passes

Keep failed attempts and reverted reports.
