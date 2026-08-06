# DESIGN.md Format

Read this reference when creating or updating the shared design-system file.
It summarizes the `google-labs-code/design.md` alpha format; use the upstream
spec when a detail here is insufficient.

## Contract

`DESIGN.md` has two layers:

1. YAML front matter holds normative, machine-readable tokens.
2. Markdown explains visual intent and how to apply those tokens.

Use only reusable visual-system information here. Keep a page's task flow,
wireframe, data fields, and exceptions in its page specification.

## Token shape

```yaml
---
version: alpha
name: Product name
description: Optional visual-system summary
colors:
  primary: "#1D4ED8"
  neutral-0: "#FFFFFF"
typography:
  body-md:
    fontFamily: Inter
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.5
rounded:
  sm: 6px
spacing:
  sm: 8px
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.neutral-0}"
    rounded: "{rounded.sm}"
    padding: 8px
---
```

Supported top-level token groups are `version`, `name`, `description`,
`colors`, `typography`, `rounded`, `spacing`, and `components`. Reference a
defined token with `{path.to.token}`. Component rules may define
`backgroundColor`, `textColor`, `typography`, `rounded`, `padding`, `size`,
`height`, or `width`.

## Markdown body

Use only the sections relevant to the system, in this order:

1. `## Overview`
2. `## Colors`
3. `## Typography`
4. `## Layout`
5. `## Elevation & Depth`
6. `## Shapes`
7. `## Components`
8. `## Do's and Don'ts`

Describe the rationale that lets an agent make a consistent choice when a
specific token does not cover the case. State accessibility-relevant rules in
the applicable section, such as semantic status indicators that do not rely on
color alone.

## Validation

When permitted, run:

```bash
npx @google/design.md lint path/to/DESIGN.md
```

The linter checks token references, contrast for component foreground and
background pairs, token usage, typography presence, and canonical section
order. The format is alpha, so treat a version upgrade as a deliberate review
event.
