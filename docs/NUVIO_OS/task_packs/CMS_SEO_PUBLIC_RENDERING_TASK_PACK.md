# CMS SEO Public Rendering Task Pack

## Purpose
Use this task pack when validating CMS editing, website settings, page SEO, assets, public DTO rendering, preview iframe, sitemap, robots, translated SEO/i18n rendering, or public SEO output.

## Task classification
- regression
- readiness
- launch-critical

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [CMS](../features/CMS.md)
- [Website Settings and SEO](../features/WEBSITE_SETTINGS_SEO.md)
- [Assets](../features/ASSETS.md)
- [Public Runtime](../features/PUBLIC_RUNTIME.md)
- [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)

## Optional source docs
- [Admin UI Contract](../../NUVIO_ADMIN_UI_CONTRACT.md)
- [SchemaForm and Forms Contract](../../NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md)
- [Website Settings and SEO Contract](../../NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md)
- Obsidian CMS, Assets and Images, Public Runtime, and Security Hardening docs.
- Current CMS/public runtime source and tests.

## Preconditions
- Target website and public runtime are known.
- CMS preview origin and public site URL are known.
- Task scope is audit/regression or implementation is explicitly requested.
- Any page/asset data mutation is safe for the target environment.
- Current source and git status can be checked before changes.

## Source-of-truth rules
1. Current CMS/public runtime source, migrations, tests, and git status win.
2. Repo contracts win for data boundaries and forms behavior.
3. Nuvio OS cards route risk and validation.
4. Obsidian docs are human context.
5. Runtime output must be observed before claiming SEO/rendering success.

## Allowed work
- Audit CMS/settings/SEO/public rendering behavior.
- Make visual or copy fixes if explicitly scoped.
- Preserve SchemaForm payloads and hidden settings keys.
- Validate public output only when the future task permits smoke checks.

## Forbidden work
- Do not move SEO fields into `websites.settings`.
- Do not overwrite hidden settings keys.
- Do not change SchemaForm parser, TinyMCE, file upload, or asset behavior casually.
- Do not change public runtime contracts from an admin UI task.
- Do not add unavailable SEO checks or fake sitemap/canonical claims.

## Danger zones
- SchemaForm nested value behavior.
- Asset file/storage references and native file fields.
- Preview iframe CORS/CSP/frame origins.
- SEO fields split between page, website, settings, and translated/i18n output.
- Public runtime rendering of unsafe HTML or URLs.

## Execution outline
1. Confirm target environment and whether data edits are allowed.
2. Read CMS, SEO, Assets, Public Runtime, and contracts.
3. Inspect current CMS/public runtime implementation if code changes are requested.
4. Check save/preview/public rendering boundaries.
5. Separate blockers from polish.
6. Report unknown public runtime/deployment state.

## Validation checklist
### Doc validation
- Docs/contracts consulted are listed.
- Unknown SEO/runtime fields are marked.
- No unsupported public SEO claim is introduced.

### Code/build/test validation, if future implementation applies
- If implementation applies, run UI build/check and relevant backend/public tests.
- If public runtime is touched, run target public app check/lint/build where available.
- If endpoint/schema changes occur, run relevant backend tests.

### Manual smoke validation
- CMS page edit/save.
- Preview iframe loads and refreshes.
- Public page renders.
- Assets render.
- Title/meta/social/robots/sitemap behavior if runtime is in scope.
- Translated page SEO and public rendering if i18n is enabled.
- Client-role CMS access if applicable.

### User confirmation needed
- Target website/page.
- Permission to edit CMS data.
- Whether public runtime is cms5, Reference-derived site, or real site repo.
- Whether runtime SEO output is in scope.

## Expected report format
- Files read.
- Files changed.
- CMS/SEO/public rendering checks performed.
- Unknowns and skipped checks.
- Risks.
- Confirmation data boundaries were preserved.
- Next recommended step.

## Stop conditions
- Task requires changing SchemaForm/TinyMCE/file upload without explicit scope.
- Public runtime target is unclear.
- SEO field availability is unknown but needed for a claim.
- A fix would overwrite hidden settings or move SEO data.

