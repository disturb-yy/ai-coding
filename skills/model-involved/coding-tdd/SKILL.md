---
name: coding-tdd
description: "Run test-first changes in an existing project. Use for red-green development, regression fixes, and integration behavior that can be delivered as small observable slices."
---

# Coding TDD

## Localization Maintenance

- When modifying this `SKILL.md`, update `SKILL.zh-CN.md` in the same change.
- `SKILL.zh-CN.md` is user-facing documentation only. Do not read or use it as model instructions, task context, or execution guidance.
- Treat this English `SKILL.md` as the model-readable source of truth.

## Purpose

Use `/coding-project` for project conventions, safe edits, and validation commands. Use this skill to drive the change in **red → green** cycles, then refactor once after the requested behavior is complete and verified.

Each cycle is one vertical, externally observable behavior: one test, one minimal implementation, one GREEN result. A test is a tracer bullet, not a speculative test plan.

## Start

1. Load `/coding-project` and complete its context scan for the affected code and test surface. Read `CONTEXT.md` and local ADRs when they exist.
   - Complete when the first public seam, its project vocabulary, and the narrowest test command are known.
2. Name the smallest observable behavior and its public seam: an API response, command output, public function result, event, or user-visible state.
   - Complete when the behavior has an independently derived expected result. If several seams are plausible and the choice materially changes scope, confirm the seam with the user.
3. Split the request into an ordered list of thin vertical slices. Start with the smallest useful success or regression case.
   - Complete when the next slice can be expressed as one focused test. Keep later slices as brief names; do not write their tests yet.

## Red → Green Loop

Repeat this loop for every slice. Finish the current row before beginning the next.

| State | Action | Completion criterion |
| --- | --- | --- |
| **Red** | Write one test through the agreed public seam. Run the narrow test command. | It fails for the expected missing or incorrect behavior, not because of a broken harness or unrelated failure. |
| **Green** | Change only the production code needed to make that same test pass. Run the same command again. | The test passes. No behavior for a later slice was added. |
| **Next slice** | Use what the GREEN result revealed to choose the next smallest behavior. | The new slice has its own focused RED test; the previous slice remains green. |

When a checkpoint is not RED or GREEN as expected, investigate and restore that checkpoint before advancing.

## Test Design

A useful test specifies behavior through its public seam and can survive an internal rewrite. Derive expected values from a specification, fixture, worked example, or known-good literal—not by restating the implementation's algorithm.

Keep each test focused on one outcome. Use project-native fakes or mocks only at dependencies outside the slice. For end-to-end behavior, prefer real project wiring and replace only unavailable external systems according to local conventions.

Choose vertical slices over horizontal layers:

```text
request → observable response
```

Drive the request behavior first when it is the clearest seam. Add separate function or module tests only when their public interfaces contain independently important behavior. This keeps tests tied to capabilities rather than internal collaboration.

## Review Before Refactor

After every requested slice is GREEN and the affected test scope passes, load `/reviewing-code` and review the completed diff before changing its structure. Treat the completed implementation, its tests, and the validation evidence as the review target.

Resolve material findings before refactoring. When a fix changes observable behavior, return to the red → green loop with its own test; when it only corrects the completed implementation, rerun the affected tests to restore GREEN. Record accepted residual risks instead of disguising them as refactoring work.

## Final Refactor

Enter this stage only after the `/reviewing-code` review is complete and material findings are resolved. Refactoring is a separate final pass, not a step inside the red → green loop.

1. Review the completed change for duplication, unclear names, and structure that obscures the tested behavior.
2. Make the smallest refactor that improves the design without changing the observable contracts.
3. Run the affected test suite after each refactor. Restore GREEN before another refactor or before reporting completion.
4. Run the broader package/module validation required by `/coding-project`; run an entry-to-output check when the project supports one and the requested behavior crosses that boundary.

The skill is complete when every requested behavior has a witnessed RED test and GREEN result, the pre-refactor review is complete, the final refactor pass is GREEN, and validation evidence covers the affected scope.

## Boundaries

| Request | Use this skill? | Action |
| --- | --- | --- |
| Build a feature test-first | Yes | Find the first observable slice and start RED. |
| Fix a bug with a regression test | Yes | Make the regression RED, then make it GREEN with the smallest fix. |
| Add integration behavior around an endpoint | Yes | Start from the request/response seam and iterate in small slices. |
| Explain or review TDD without code changes | No | Answer or review without entering the loop. |
| Implement without a test-first requirement | No | Use `/coding-project`. |
