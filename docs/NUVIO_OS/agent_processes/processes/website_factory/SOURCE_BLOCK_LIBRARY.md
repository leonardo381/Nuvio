# Source Block Library

## Purpose

This file standardizes where Website Factory agents find approved source blocks and how they reference them during Stage 4 Block Selection.

It does not authorize source edits, block imports, implementation work, or runtime dependencies.

## Canonical Source Block Library

Canonical Website Factory source block library:

```text
C:\Users\Leo\Documents\Nuvio\Srcs\html
```

This path contains static HTML source blocks organized by category, including sections, elements, page examples, heroes, contact sections, pricing, FAQs, CTAs, footers, stats, testimonials, and related marketing blocks.

Broader global source/assets library:

```text
C:\Users\Leo\Documents\Nuvio\Srcs
```

`Srcs` may contain assets, screenshots, downloads, archives, and non-HTML framework examples. For Website Factory block selection, use `Srcs\html` first.

## Non-Canonical Candidates

| Path | Status | Rule |
| --- | --- | --- |
| `C:\Users\Leo\Documents\Nuvio\Srcs\react` | Reference only | Do not use as canonical for Website Factory unless Hermes/Leonardo explicitly approves a React-source adaptation. |
| `C:\Users\Leo\Documents\Nuvio\Srcs\vue` | Reference only | Do not use as canonical for Website Factory unless Hermes/Leonardo explicitly approves a Vue-source adaptation. |
| `C:\Users\Leo\Documents\Nuvio\Sites\Reference` | Clean public-site technical template | Do not use as a source block library. Use it for public-site contracts and starter architecture. |
| `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5` | Lab/dev/runtime history | Do not use as canonical source blocks unless a task explicitly targets cms5. |
| Real site repos under `C:\Users\Leo\Documents\Nuvio\Sites\*` | Site-specific implementations | Do not mine them for source blocks unless explicitly approved. |

## Fallback Rule If Path Is Missing

If `C:\Users\Leo\Documents\Nuvio\Srcs\html` does not exist or cannot be read:

- do not invent source paths;
- do not select from cms5 automatically;
- do not select from Reference automatically;
- record `Unknown / needs source block library path`;
- stop before Stage 4 final approval;
- ask Hermes/Leonardo for the approved source block library path.

Agents may recommend block types in Stage 3 or early Stage 4, but Stage 4 cannot be complete without approved source block references or an explicit Hermes/Leonardo exception.

## Approved Block Definition

An approved source block is:

- a specific source file or folder under the canonical source block library;
- selected during Stage 4;
- recorded in `BLOCK_SELECTION.md`;
- approved or accepted by Hermes/Leonardo before Stage 5 import;
- treated as immutable source material.

An approved source block is not:

- a vague visual idea;
- a copied target-site component;
- a cms5/Reference implementation detail;
- a block modified in-place to make it easier to import.

## Source Block vs Copied Block

| Term | Meaning | Mutable? |
| --- | --- | --- |
| Source block | Original source file/folder under `Srcs\html`. | No. Read-only. |
| Approved source block | Source block selected and recorded in Stage 4. | No. Read-only. |
| Copied block | Target-site copy created during Stage 5 raw import. | Yes, but only under the active stage rules. |

## Allowed Agent Actions

During Stage 4 agents may:

- inspect `Srcs\html` read-only;
- list candidate block paths;
- open small source files only as needed to understand fit;
- record exact source paths in `BLOCK_SELECTION.md`;
- record fit, risks, dependencies, and reasons;
- recommend block types when exact paths cannot be selected yet;
- stop when approval or source path is missing.

## Forbidden Agent Actions

Agents must not:

- modify source block files;
- rename source block files;
- reformat source block files;
- delete or reorganize source block folders;
- copy/import blocks during Stage 4;
- bulk-copy source folders;
- use `Srcs` as a runtime dependency;
- make a target site import files from `Srcs`;
- use cms5 as canonical source without explicit approval;
- use Reference as a source block library;
- modify source blocks to make import easier.

## Recording Source Blocks In `BLOCK_SELECTION.md`

Every selected block entry should include:

- canonical source library path;
- exact source path;
- source category, such as `sections/heroes` or `sections/contact-sections`;
- target page and section;
- reason selected;
- fit notes or score;
- dependencies, such as images, icons, JS behavior, or framework assumptions;
- risks;
- confirmation that source files were inspected read-only;
- note that Stage 5 import approval is still required.

If exact source path is unknown, use:

```text
Unknown / needs source block library path
```

Do not fabricate plausible paths.

## Recording Raw Import In `BLOCK_IMPORT_LOG.md`

During Stage 5, record:

- approved source path;
- copied target path;
- import date;
- minimal import/build fixes;
- copied assets, if any;
- confirmation that source files remain unchanged;
- confirmation that copy/text, visual hierarchy, layout, integration, refactor, and component cleanup were not changed during raw import.

If import requires fixes, fixes happen only in the copied target files and are logged in `BLOCK_IMPORT_LOG.md`.

## If A Needed Block Is Missing

If no canonical block fits a required section:

1. Do not force a poor fit.
2. Record the missing block need in `BLOCK_SELECTION.md`.
3. Recommend a block type, not an invented exact path.
4. Mark the section as blocked or needs Hermes/Leonardo decision.
5. Options are:
   - select a different approved source block;
   - simplify the blueprint;
   - approve a custom target-site section in a later implementation stage;
   - add a new source block to the library only in a separate, explicit source-library curation task.

## If A Block Seems Broken

If a source block appears broken, incomplete, or incompatible:

- do not edit the source block;
- record the issue and source path;
- choose another block if possible;
- ask for Hermes/Leonardo approval before using it;
- if approved for import, make only copied-target fixes during Stage 5 and log them.

## If Multiple Block Candidates Exist

Choose the block that best matches:

- page blueprint section goal;
- visual role;
- target stack;
- expected adaptation cost;
- mobile behavior;
- dependency risk;
- copy/brand flexibility.

Record alternatives only when they are genuinely close contenders. Do not turn block selection into a broad design exploration.

## Source Block Immutability Reminder

The canonical source block library is read-only source material. A target site receives copied blocks only during Stage 5. Source files must remain untouched before, during, and after import.

## Stop Conditions

Stop when:

- `Srcs\html` is missing or unreadable;
- the needed source block path is unknown;
- source block selection would require editing source files;
- the best candidate requires unapproved dependencies;
- cms5, Reference, React, Vue, or a real site repo would be used without explicit approval;
- a block import is requested during Stage 4;
- Hermes/Leonardo approval is missing before Stage 4 final approval or Stage 5 import.

## Approval Requirements

Hermes/Leonardo approval is required before:

- treating Stage 4 block selection as final;
- importing blocks into a target repo;
- using non-canonical candidates such as `Srcs\react`, `Srcs\vue`, cms5, Reference, or a real site repo;
- creating or curating new source blocks;
- making any source-block-related exception.
