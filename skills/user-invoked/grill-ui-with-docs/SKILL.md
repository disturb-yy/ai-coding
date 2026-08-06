---
name: grill-ui-with-docs
description: Interview a UI design one decision at a time, then turn an approved wireframe into a documented design-system handoff.
disable-model-invocation: true
---

# Grill UI with Docs

Design a page or flow through a **wireframe gate**: approval of a low-fidelity
wireframe is the sole condition that unlocks the design-system handoff.

## Scope

Produce design artefacts, not production frontend code. Use `/grilling` to run
the interview. Use `/domain-modeling` only when business terms, states, or
workflows need to be resolved. Use an available model-invoked UI/UX knowledge
skill for a visual, accessibility, chart, or interaction recommendation; make
the recommendation directly when none is available.

## 1. Ground the design

Inspect the repository before asking about facts it already contains: product
language, routes, existing screens, component library, design tokens, chart
library, responsive conventions, and data contracts.

Reuse established document locations. When the repository has no design-doc
convention, use `docs/design/DESIGN.md` for the shared system and
`docs/design/pages/<page>.md` for a page specification.

Complete this step when the known constraints, the design document paths, and
the unresolved decisions are explicit.

## 2. Grill the decisions

Walk the dependency tree one decision at a time. Supply a recommendation and
its rationale with every question. Prefer constrained choices that a
non-designer can approve.

Resolve decisions in this order when relevant:

1. user, primary task, and page scope;
2. information architecture and task flow;
3. page structure, information hierarchy, density, and data presentation;
4. loading, empty, error, permission, and destructive-action states;
5. responsive, keyboard, and reduced-motion behaviour;
6. visual direction and reusable component rules.

Record only accepted decisions that constrain the shared system or the page.
Put business vocabulary in `CONTEXT.md`; put durable architecture trade-offs in
an ADR; put visual and page decisions in the design artefacts.

Complete this step when every decision needed to draw the low-fidelity page is
accepted or explicitly marked as an assumption for approval.

## 3. Build the wireframe gate

Create or update the page specification with:

- the page goal and primary task;
- a Markdown/ASCII wireframe of every viewport-relevant region;
- component tree and information hierarchy;
- primary interaction flow;
- loading, empty, error, and permission states that apply;
- responsive behaviour and accessibility constraints;
- accepted decisions and remaining assumptions.

Present the wireframe and wait for explicit user approval. Revise it until the
user approves it. Completion is an unambiguous approval of the current
wireframe version.

## 4. Write the approved design system

After approval, read [the DESIGN.md format](references/design-md.md). Create
or update the shared `DESIGN.md` with reusable visual tokens and their
rationale. Keep the approved page-specific layout, flows, and exceptions in
its page specification.

Keep YAML token names stable and semantic. Keep the Markdown body in the
canonical section order. Represent new page-specific rules in prose unless
they are reusable system tokens or reusable component rules.

When the `@google/design.md` CLI is already available or the user authorizes
fetching it, lint the completed file. Resolve every lint error; report warnings
that require a design decision.

Complete this step when tokens resolve, prose explains their application, and
the page specification has no unexplained conflict with the shared system.

## 5. Hand off

Return a compact handoff containing the approved page, design-system path,
page-specification path, implementation constraints, and acceptance checks.
Mark implementation as allowed only after the wireframe gate and design-system
step both complete.
