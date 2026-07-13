# Block Library Rules

## Purpose

Define hard rules for using source blocks in the Website Factory process.

## Non-Negotiable Principle

The source block library is immutable.

Agents must never modify source blocks.

## Agents May Only

- Select source blocks.
- Copy approved source blocks into a target website/app.
- Adapt copied blocks only in the correct later stage.

## Raw Block Import Stage

Raw block import must be its own stage.

During raw block import:

- copy blocks as-is;
- no copywriting;
- no styling adaptation;
- no layout redesign;
- no Nuvio/CMS integration;
- no refactor;
- no component cleanup;
- no "while I am here" improvements.

Only minimal import/build fixes are allowed, and they must be logged in `BLOCK_IMPORT_LOG.md`.

## Allowed Raw Import Fixes

- Syntax conversion needed for the target framework.
- Import path fixes needed for compilation.
- Replacing unavailable external icons with minimal placeholders.
- Removing runtime references to source-library paths.
- Minimal accessibility fixes required to keep valid markup.

## Forbidden Raw Import Changes

- Rewriting layout structure.
- Changing section order.
- Replacing copy beyond mechanical placeholder mapping.
- Restyling colors, spacing, shadows, or typography.
- Creating reusable abstractions.
- Integrating Nuvio data, CMS fields, forms, or analytics.
- Adding dependencies because a source block used them.

## Copied Block Adaptation Rules

Copied blocks may be adapted only in later stages:

- Page assembly may arrange already imported blocks.
- Copywriting pass may replace copy.
- Brand / visual adaptation pass may adjust visual style.
- Nuvio integration pass may connect approved data and flows.
- CMS/content mapping may map content into Nuvio structures.

## Stop Conditions

Stop and report `Unknown / needs confirmation` when:

- the source block path is unclear;
- a block requires an unapproved dependency;
- import would require redesign;
- source files would need to be modified;
- copied code would depend on source-library runtime paths;
- build fixes become visual adaptation or refactor.

## Report Requirements

Raw import reports must include:

- source block path;
- target file path;
- exact copied block identity;
- minimal fixes made;
- anything intentionally not changed;
- `BLOCK_IMPORT_LOG.md` update status.
