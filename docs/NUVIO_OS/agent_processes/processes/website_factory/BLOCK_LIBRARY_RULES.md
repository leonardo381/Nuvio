# Block Library Rules

## Purpose

This document defines how agents may use source blocks during Website Factory work.

For the canonical Website Factory source block path and Stage 4 selection source, read [Source Block Library](SOURCE_BLOCK_LIBRARY.md).

## Immutable Source Library Rule

The source block library is immutable.

Agents MUST NOT edit, rename, reformat, delete, normalize, or "clean up" source block files. Source files are references. They are not the target website.

## Definitions

| Term | Meaning | Mutable? |
| --- | --- | --- |
| Source block | A block, section, template, or asset inside the approved source library. | No. Read-only. |
| Approved source block | A source block selected in Stage 4 and recorded in `BLOCK_SELECTION.md`. | No. Read-only. |
| Copied block | A copy of an approved source block placed inside the target website repo during Stage 5. | Yes, but only in allowed later stages. |
| Target site | The real website repo/app receiving selected blocks. | Yes, within the active stage scope. |

## Selection Rules

- Select blocks only during Stage 4.
- Use the canonical source block library defined in [Source Block Library](SOURCE_BLOCK_LIBRARY.md).
- Select exact files or folders, not vague inspiration.
- Record source path, target page/section, reason, fit notes, risks, and dependencies.
- Prefer blocks that match the blueprint with minimal later adaptation.
- Do not select a block that requires unapproved dependencies unless the risk is recorded and approved.
- Do not select blocks from cms5, Reference, or any lab source unless that path is explicitly approved for the task.

## Raw Import Rules

Stage 5 raw block import MUST be boring and literal.

Hermes/Leonardo approval is required before importing blocks into the target repo.

Allowed:

- copy selected blocks as-is;
- copy required unchanged local assets;
- fix import paths;
- correct file paths;
- adjust dependency paths;
- apply minimal syntax fixes required to compile;
- apply minimal formatting required by target build/linter;
- rewire missing local asset paths if the source asset is copied unchanged;
- log every fix in `BLOCK_IMPORT_LOG.md`.

Forbidden:

- changing copy/text;
- changing visual hierarchy;
- redesigning layout;
- merging blocks;
- splitting blocks;
- abstracting components;
- renaming design concepts;
- adding CMS integration;
- adding forms/integrations;
- changing responsive behavior intentionally;
- cleaning up architecture;
- replacing block implementation with a new design;
- adding "while I am here" improvements.

## Copied-Block Adaptation Rules

Copied blocks may be adapted only in later stages:

| Stage | Allowed adaptation |
| --- | --- |
| Stage 6 | Assembly into pages/routes with minimal composition adjustments. |
| Stage 7 | Copy replacement while preserving approved structure. |
| Stage 8 | Brand, color, spacing, imagery, and visual consistency adaptation. |
| Stage 9 | Confirmed Nuvio integrations only. |
| Stage 10 | CMS/content mapping notes or implementation only if separately scoped. |
| Stage 12 | Approved final polish from QA findings. |

## Logging Requirements

`BLOCK_IMPORT_LOG.md` MUST record:

- copied block name;
- source path;
- target path;
- import date;
- minimal fixes made;
- build/import issues;
- confirmation that forbidden changes were not made.

If a copied block is later heavily adapted, log the reason in `VISUAL_ADAPTATION_NOTES.md` or the active stage artifact.

## Source Block Mutation Incident Response

If a source block is accidentally changed:

1. Stop immediately.
2. Do not continue adapting or importing.
3. Report the exact source path and change.
4. Restore only the accidental source-block change if safe and explicitly allowed by repo policy.
5. Record the incident in the stage report.
6. Resume only after the source library is confirmed clean.

## Stop Conditions

Stop when:

- Hermes/Leonardo approval is missing for block import or a source-block exception;

- a source block would need modification;
- import requires unapproved dependencies;
- a block cannot compile without redesign;
- a source asset is missing or licensing is unclear;
- the user asks for copy, brand, integration, or cleanup during Stage 5;
- exact source path cannot be identified.

## Allowed vs Forbidden Examples

| Scenario | Allowed? | Stage | Notes |
| --- | --- | --- | --- |
| Fix `../Button` import to `./Button` after copying a block. | Yes | 5 | Log it. |
| Convert `className` to `class` for Svelte compile. | Yes | 5 | Minimal syntax fix. |
| Change "Acme" to client name during raw import. | No | 7 | Copywriting pass only. |
| Replace a hero layout because another design looks better. | No | None | Requires new block selection or blueprint change. |
| Apply client colors after copy is approved. | Yes | 8 | Log visual adaptation. |
| Add contact form submit logic while importing a form block. | No | 9 | Integration pass only. |
