# Source Of Truth

Use this file when docs conflict or when a task spans multiple Nuvio repos.

## Known Local Paths

| Source | Path | Role |
| --- | --- | --- |
| Central documentation vault | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio` | Product model, offer/tier/pricing strategy, naming, prompt rules, lifecycle, offboarding, documentation authority. |
| Backend/backoffice repo | `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio` | CMS/backoffice implementation, endpoints, auth, schema contracts, seed/exporter behavior, middleware, tests, deployment docs. |
| Final public site repo | `C:\Users\Leo\Documents\Nuvio\Sites\Nuvio-CalmEditorialV2` | Current Nuvio public marketing site, root marketing CMS defaults, `/site`, `/site-preview`, sitemap, Contact mechanics, pricing guardrails. |
| Reference public repo | `C:\Users\Leo\Documents\Nuvio\Sites\Reference` | Clean reusable public-site starter/template conventions. |
| Official/older site repo | `C:\Users\Leo\Documents\Nuvio\Sites\Nuvio` | Supporting/historical business and website-factory material until migrated into the central vault. |
| cms5 public runtime/test app | `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5` | Runtime/testing history and legacy implementation reference when current code confirms it. |

## Source Hierarchy

| Rank | Source | Canonical for | Caution |
| --- | --- | --- | --- |
| 1 | Current source code, tests, migrations, Dockerfiles, package scripts, git status | Actual behavior and implementation truth for the target repo. | Always inspect recent uncommitted changes before editing. |
| 2 | Central vault P0 docs | Product model, tier/pricing strategy, naming, prompt rules, lifecycle, offboarding, documentation authority. | Does not override current code for runtime behavior. |
| 3 | Backend `AGENTS.md` and repo-local contracts | Backend/backoffice implementation rules and safety boundaries. | Contracts can lag code in edge cases. |
| 4 | Backend deployment docs | Deployment/env/Coolify/bootstrap model. | Placeholder domains/secrets; confirm real deploy state. |
| 5 | Final public site docs/source | Current Nuvio public-site implementation and public route behavior. | Do not infer backend schema rules from visual page docs. |
| 6 | Reference repo docs | Reusable template direction. | Not product/pricing/tier/naming authority. |
| 7 | Official/older site docs | Supporting business and workflow history until migrated. | Must not override central vault or current source. |
| 8 | Old backlog, historical logs, archived docs | Provenance and investigation. | Never active unless explicitly confirmed. |

## Default Conflict Rule

If sources conflict:

1. Check current source code and git status.
2. Prefer repo-local contracts for implementation details in that repo.
3. Prefer the central vault for product model, pricing, tiers, naming, prompts, lifecycle, offboarding, and documentation authority.
4. Prefer final-site source/docs for current public marketing route behavior.
5. Treat official/older site docs and historical logs as supporting evidence only.
6. Mark unresolved conflicts as `Unknown / needs confirmation`.

## Current Product And Pricing Authority

Current product/pricing authority lives in the central vault:

- `DOCUMENTATION_AUTHORITY.md`
- `PRODUCT_MODEL.md`
- `TIERS_AND_PRICING.md`
- `NAMING_POLICY.md`
- `PROMPT_RULES.md`
- `CLIENT_PROJECT_LIFECYCLE.md`
- `BACKOFFICE_BOUNDARY.md`
- `OFFBOARDING_AND_FALLBACK.md`

Current founder pricing is known and canonical there:

- Presença: €590 setup + €69/month.
- Crescimento: €990 setup + €99/month.
- Parceiro: €1,390 setup + €149/month.

All client websites are CMS/backoffice-connected regardless of tier. Tier differences are number of pages/sections, scope, customization depth, editable surface, revisions, support/reporting level, speed/priority, and ongoing evolution, not CMS vs no CMS.

## Backend / Backoffice Boundary

For NuvioCMS/backend/backoffice work related to mapped/editable sites, only touch schemas, component definitions, fixtures, seed data, and DB collections/records directly relevant to the mapped sites.

Do not modify established CMS/backoffice features, SchemaForm/editor UI, InputSelect, InputArray, PageCms, middleware, auth, pricing source code, Contact mechanics, renderer/layout/design, or unrelated modules unless there is a severe blocker and explicit approval.

NuvioCMS must not drift into a generic page builder. Public website layout/design remains code-owned; CMS edits are controlled props.

## Docs To Read By Category

### Global

- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\DOCUMENTATION_AUTHORITY.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\AGENTS.md`
- `docs/NUVIO_OS/CORE.md`
- `docs/NUVIO_OS/TASK_ROUTER.md`
- `docs/NUVIO_OS/DANGER_ZONES.md`

### Product / Pricing / Naming / Workflow

- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\PRODUCT_MODEL.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\TIERS_AND_PRICING.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\NAMING_POLICY.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\CLIENT_PROJECT_LIFECYCLE.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\OFFBOARDING_AND_FALLBACK.md`

### Prompt / Agent Rules

- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\PROMPT_RULES.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\BACKOFFICE_BOUNDARY.md`
- `AGENTS.md`

### UI / Forms / Settings

- `docs/NUVIO_ADMIN_UI_CONTRACT.md`
- `docs/NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md`
- `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`

### Deployment

- `docs/NUVIO_DEPLOYMENT_QUICK_GUIDE.md`
- `docs/NUVIO_DEPLOYMENT_ENV_MATRIX.md`
- `docs/NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md`
- `docs/NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md`
- `deploy/README.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Deployment Quick Guide.md`

### Public Runtime / Final Site / Reference

- `C:\Users\Leo\Documents\Nuvio\Sites\Nuvio-CalmEditorialV2\README.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Nuvio-CalmEditorialV2\docs\CMS_CONTENT_MAP.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Nuvio-CalmEditorialV2\docs\NUVIO_PUBLIC_SITE_ENV_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\AGENTS.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\GLOBAL_SOURCE_LIBRARY.md`

### Security / Permissions

- `docs/NUVIO_ADMIN_UI_CONTRACT.md`
- `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Security Hardening.md`
- Relevant feature docs and current backend tests/source.

### Feature Work

- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\CMS.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Assets and Images.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Leads.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Contact Form and WhatsApp.md`
- Current source files and tests for the feature.

## Stale/Default Docs Caution

- Main repo root `README.md` is mostly PocketBase-oriented.
- cms5 root `README.md` is default Svelte starter text.
- cms5 `src/layoutBuilder.md` and `src/templateAdapter.md` are generator/lab notes.
- Official/older site docs are supporting/historical unless central vault or a task explicitly references them.
- Old roadmaps/backlogs are inactive unless confirmed.
