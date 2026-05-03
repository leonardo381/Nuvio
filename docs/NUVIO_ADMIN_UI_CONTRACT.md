# Nuvio Admin UI Contract

The Nuvio admin/backoffice must feel like one product.

Do not invent a new visual system for each feature.

## Existing shared primitives

Use existing primitives first:

- `operations-head` for operation/page headers
- `operations-tabs` and `operations-tabs--nested` for tabs
- `summary-pill` for summary chips
- `label` and `label-sm` for badges/status/meta
- `btn`, `btn-outline`, `btn-sm`, `btn-loading` for buttons
- `OverlayPanel` for drawers, modals, and side panels
- Form v2 field rhythm for forms
- `FieldShell` / `SchemaForm` for dynamic Nuvio forms
- `panel`, `list`, and base surfaces for simple content areas

## Design decision order

1. Use existing primitives.
2. Create shared variants if existing primitives are close but insufficient.
3. Create reusable components if a repeated pattern exists.
4. Use local one-off styling only as a last resort.

## New components are allowed

New components or variants are allowed, but must be justified first.

Before creating a new component, class, or visual pattern, explain:

- which existing primitive was considered
- why it is insufficient
- whether the new thing should be a shared component, shared SCSS variant, or local layout-only class
- where else it could be reused
- how it fits the current Nuvio visual language

## Do not create

Do not create:

- underline-only tabs
- custom local button systems
- custom modal/drawer shells
- page-specific badge systems
- hardcoded color palettes
- heavy gray nested card systems
- random one-off form wrappers
- visual patterns that make one feature look like a separate app

## Local CSS rules

Local CSS is allowed for layout composition.

Local CSS is not allowed for redefining:

- buttons
- tabs
- badges
- drawers
- modals
- form primitives
- shared surfaces

## UI task rules

When improving UI:

- preserve existing behavior
- preserve save/data logic
- avoid broad rewrites
- prefer audit-first for shared systems
- keep changes small and reviewable
- do not redesign unrelated areas