# Nuvio SchemaForm and Forms Contract

Nuvio has multiple active form systems:

1. PocketBase record forms
2. Nuvio SchemaForm dynamic forms
3. Hardcoded product/workflow forms

Do not force all forms into one logic system.

The goal is shared visual language, not shared behavior.

## Canonical dynamic Nuvio form

The active dynamic Nuvio form system is:

- `ui/src/components/base/nuvio/schema/SchemaForm.svelte`

SchemaForm is used for:

- editing `blocks.props`
- editing role-scoped `websites.settings`
- dynamic schema-driven forms

## SchemaForm rules

Do not change these unless explicitly requested:

- schema parsing
- recursive rendering
- nested path handling
- `propsChange` / `change` dispatch contracts
- saved payload shape
- value serialization
- array/object update behavior
- file upload value shape
- TinyMCE value sync

## High-risk components

These require audit before changes:

- `InputArray.svelte`
- `InputObject.svelte`
- `InputFile.svelte`
- `InputTextArea.svelte` / TinyMCE
- `SchemaForm.svelte` dispatcher/parser

## Form v2 visual direction

Form visual language should follow Form v2:

- consistent labels
- consistent helper text
- consistent error text
- consistent required markers
- shared button classes
- theme variables instead of hardcoded palettes
- light nested surfaces
- fields remain the focus
- containers provide structure without becoming heavy cards

## Hardcoded forms

Hardcoded forms are allowed and should remain hardcoded when they are workflow-specific.

Examples:

- Newsletter campaign forms
- CMS SEO forms
- CMS identity/global SEO forms

They should visually align with Form v2, but they do not need to use SchemaForm.

## File fields

Do not change file upload behavior casually.

For `InputFile`:

- preserve upload logic
- preserve checksum/dedupe behavior
- preserve asset picker behavior
- preserve emitted value shape
- preserve remove semantics

Visual polish is allowed only if behavior is preserved.

## TinyMCE

Do not touch `InputTextArea.svelte` / TinyMCE unless explicitly requested.

TinyMCE has lifecycle/value-sync risk.