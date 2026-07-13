# Website Settings and SEO Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on website settings save/load behavior, hidden/admin-only keys, SEO fields, canonical/social/noindex behavior, sitemap/robots implications, and public rendering metadata.

## 2. Current operating status

Done but needs regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Establishes source order and boundaries. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms routing for settings, SEO, and public runtime tasks. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Settings and SEO affect public output. |
| 1 | Website Settings and SEO Contract | [../../NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md](../../NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md) | Canonical contract for settings and SEO fields. |
| 1 | Admin UI Contract | [../../NUVIO_ADMIN_UI_CONTRACT.md](../../NUVIO_ADMIN_UI_CONTRACT.md) | Defines scoped admin behavior. |
| 2 | Operating Manual CMS | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\CMS.md` | Explains website settings in the CMS workflow. |
| 2 | Operating Manual Public Runtime | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Public Runtime.md` | Explains public SEO rendering concerns. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Defines validation expectations. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- `ui/src/components/cms/PageCms.svelte`
- `examples/base/nuvio_cms_backoffice.go`
- `pb_migrations/*` only when schema is explicitly requested.
- Public runtime SEO, sitemap, robots, and metadata files in the target public runtime repo. Inspect before changing.
- Reports SEO tab only when the task explicitly mentions reporting.

## 5. Decisions to preserve

- Preserve unknown settings keys; why: existing deployments may hold legacy or future settings; agent implication: merge instead of replacing settings objects.
- Keep hidden/admin-only keys protected; why: clients should not see or mutate internal controls; agent implication: maintain filtering and role-sensitive UI.
- Keep SEO fields in the established model; why: reports and public runtime expect those fields; agent implication: do not move SEO fields into arbitrary nested settings.
- Treat canonical, social image, noindex, sitemap, and robots as public contract outputs; why: they influence how the site appears to crawlers and social previews; agent implication: validate generated public metadata.
- Do not remove legacy compatibility keys casually; why: public runtime and existing data may still read them; agent implication: audit consumers first.

## 6. Allowed work now

- UI display or spacing fixes for settings panels.
- Save/load bug fixes that preserve unknown keys.
- SEO metadata bug fixes based on existing fields.
- Tests for settings round trips and scoped access.
- Documentation clarifications.

## 7. Do not change unless explicitly requested

- Website settings data model.
- Legacy settings keys.
- Canonical/noindex/social/sitemap/robots field meaning.
- Public runtime SEO output contract.
- Migrations for settings/SEO fields.
- Client-role access rules.

## 8. Common agent failure modes

- Saving a partial form and dropping hidden settings.
- Replacing settings instead of merging them.
- Changing public SEO output while only intending to polish the admin UI.
- Adding SEO checks that the DTO does not support.
- Treating noindex/sitemap/robots as purely visual settings.

## 9. Validation checklist

- Run `cd ui; npm run build` when UI changed.
- Run focused backend tests when settings endpoints or DTOs changed.
- Manually save website settings and confirm hidden/unknown keys survive.
- Manually check page title, description, canonical, robots, sitemap, and social image output if public runtime SEO changed.
- Confirm client-role visibility remains scoped and safe.

## 10. Reporting requirements

- Changed files.
- Whether settings model, SEO output, UI, backend, tests, or docs changed.
- Unknown key preservation result.
- Hidden/admin-only key handling.
- Public SEO outputs checked.
- Validation command results or reason not run.
- Remaining unknowns.
