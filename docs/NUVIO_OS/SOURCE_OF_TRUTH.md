# Source Of Truth

Use this file when docs conflict or when a task spans multiple repos.

## Known Local Paths

| Source | Path | Role |
| --- | --- | --- |
| Main backoffice repo | `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio` | Backoffice/backend implementation, contracts, deployment docs. |
| cms5 public runtime/test app | `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5` | Current public runtime/testing history and runtime code. |
| Reference public repo | `C:\Users\Leo\Documents\Nuvio\Sites\Reference` | Clean public-site reference/template direction. |
| Obsidian Nuvio manual | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio` | Human/product/operations context. |

## Source Hierarchy

| Rank | Source | Canonical for | Caution |
| --- | --- | --- | --- |
| 1 | Current source code, tests, migrations, Dockerfiles, package scripts, git status | Actual behavior and implementation truth. | Still inspect recent uncommitted changes before editing. |
| 2 | Main repo `AGENTS.md` and repo-local contracts | Backoffice/backend implementation rules. | Contracts can still lag code in edge cases. |
| 3 | Main repo deployment docs | Deployment/env/Coolify/bootstrap model. | Placeholder domains/secrets; real deploy state must be confirmed. |
| 4 | Reference repo docs | Clean public-site template direction. | Not necessarily canonical for deployed cms5 runtime. |
| 5 | cms5 docs/code notes | Runtime/testing history and current runtime behavior when code confirms it. | Root README/default and generator docs may be stale. |
| 6 | Obsidian Operating Manual | Product state, current phase, operations, feature context. | Human docs may lag current source. |
| 7 | Hermes/ChatGPT context | Consolidated context if provided. | No durable local source found in Phase 1. |
| 8 | Old backlog/roadmap | Historical ideas. | Never active unless confirmed. |

## Default Conflict Rule

If sources conflict:

1. Check current source code and git status.
2. Prefer repo-local contracts for implementation details.
3. Prefer Reference docs over cms5 docs for future public-site template decisions.
4. Use Obsidian for current product/operations context, not exact code behavior.
5. Mark unresolved conflicts as `Unknown / needs confirmation`.

## Docs To Read By Category

### Global

- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\AGENTS.md`
- `docs/NUVIO_OS/CORE.md`
- `docs/NUVIO_OS/TASK_ROUTER.md`
- `docs/NUVIO_OS/DANGER_ZONES.md`
- `docs/NUVIO_OS/audits/2026-06-17_SOURCE_INVENTORY.md`
- `docs/NUVIO_OS/audits/2026-06-17_CONFLICT_STALENESS_AUDIT.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\00 - Nuvio Index.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Current Roadmap.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Deferred Features.md`

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
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Coolify Plan.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Docker and Compose.md`

### Public Runtime / Reference

- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\AGENTS.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NUVIO_PUBLIC_SITE_ENV_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\TEMPLATE_BUILD.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\TEMPLATE_ADAPTER.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\GLOBAL_SOURCE_LIBRARY.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\02 - Technical Overview\Public Runtime cms5.md`

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
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Booking.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Newsletter.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Reports.md`
- Current source files and tests for the feature.

### Business / Landing

- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\TEMPLATE_ADAPTER.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\GLOBAL_SOURCE_LIBRARY.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Backoffice 1.0 Status.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Current Roadmap.md`
- Real target website repo docs/source.

## Stale/Default Docs Caution

- Main repo root `README.md` is mostly PocketBase-oriented.
- cms5 root `README.md` is default Svelte starter text.
- cms5 `src/layoutBuilder.md` and `src/templateAdapter.md` are generator/lab notes.
- Old roadmaps/backlogs are inactive unless confirmed.
- Pricing/business plan files were not found locally in Phase 1; treat pricing as unknown unless directly provided.