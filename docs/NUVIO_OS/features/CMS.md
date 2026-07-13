# CMS Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on CMS page/content editing, SchemaForm-driven edits, i18n language tabs, page SEO interaction, preview behavior, and public rendering safety.

## 2. Current operating status

Done but needs regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Establishes operating rules and source order. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms CMS task routing. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | CMS touches schema, preview, assets, and public runtime risk. |
| 1 | Admin UI Contract | [../../NUVIO_ADMIN_UI_CONTRACT.md](../../NUVIO_ADMIN_UI_CONTRACT.md) | Defines scoped admin/backoffice behavior. |
| 1 | SchemaForm Contract | [../../NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md](../../NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md) | Preserves SchemaForm architecture and form rules. |
| 1 | Website Settings and SEO Contract | [../../NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md](../../NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md) | Defines page SEO and website settings boundaries. |
| 2 | Operating Manual CMS | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\CMS.md` | Human operating guide for CMS behavior. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Selects build/test/manual smoke checks. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- `ui/src/components/cms/PageCms.svelte`
- `ui/src/components/base/nuvio/schema/SchemaForm.svelte`
- `ui/src/components/base/nuvio/schema/*`
- `examples/base/nuvio_cms_backoffice.go`
- `examples/base/nuvio_cms_backoffice_test.go`
- `pb_migrations/*` only when schema work is explicitly requested.
- Public runtime preview/rendering files only after confirming the target public runtime repo.

## 5. Decisions to preserve

- Use scoped CMS endpoints; why: client-role safety depends on website scoping; agent implication: do not reintroduce raw PocketBase writes from the UI.
- Keep SchemaForm as the editing backbone; why: many CMS forms share this contract; agent implication: do not rewrite the form architecture for a local bug.
- Preserve unknown settings and schema keys; why: settings may contain legacy or future-safe data; agent implication: merge carefully instead of replacing entire objects.
- Treat i18n language tabs as high regression risk; why: one tab can accidentally overwrite another language; agent implication: test language switching and saves.
- Keep page SEO fields aligned with the SEO contract; why: public rendering and reports read those fields; agent implication: do not move SEO into arbitrary settings blobs.
- Preview is a contract between backoffice and public runtime; why: CORS, frame policy, and postMessage can break silently; agent implication: validate preview manually when touched.

## 6. Allowed work now

- UI polish that preserves behavior.
- Regression fixes in CMS save/load flows.
- Scoped endpoint fixes when a contract gap is proven.
- SchemaForm bug fixes that keep existing data shape.
- Tests for CMS scoped endpoint behavior.
- Documentation clarifications.

## 7. Do not change unless explicitly requested

- Raw PocketBase collection writes from UI code.
- SchemaForm architecture or global field semantics.
- CMS public DTO contracts.
- Page SEO storage model.
- Website settings model or legacy keys.
- Migrations, restore behavior, or asset storage behavior.
- Public runtime rendering behavior when the task is only backoffice UI.

## 8. Common agent failure modes

- Fixing one CMS form by bypassing SchemaForm.
- Dropping unknown settings keys during save.
- Treating language-tab state as local-only and overwriting another locale.
- Reintroducing raw PB writes because they seem faster.
- Changing preview URLs, frame policy, or public DTO shape without testing preview.
- Mixing CMS work with Reports, Leads, Newsletter, or Booking polish.

## 9. Validation checklist

- Run `cd ui; npm run build` when UI changed.
- Run focused CMS backend tests when backend CMS code changed. If exact test names are unclear, inspect current tests before running.
- Manually check CMS dashboard load, page edit, block edit, language tab switching, save, preview refresh, and public page render.
- Check page SEO fields still render in public runtime if SEO or page metadata changed.
- Confirm no raw PB UI calls were introduced.

## 10. Reporting requirements

- Changed files.
- Whether UI, backend, tests, docs, or migrations changed.
- Source docs read.
- Whether raw PB writes were avoided.
- Whether SchemaForm architecture was preserved.
- i18n, preview, and SEO regression checks performed.
- Validation command results or reason not run.
- Remaining unknowns.
