# Nuvio OS Source Inventory Audit

Date: 2026-06-17

## 1. Purpose

This audit inventories the documentation sources that should inform a future agent-focused Nuvio OS.

Nuvio OS is not intended to be a human operating manual or a new product feature. It should later help AI agents working on Nuvio understand where truth lives, which documentation to read for a task, which work is current, which work is deferred, and how to avoid unsafe or stale changes.

This file is an audit report only. It does not create the final Nuvio OS, rewrite existing documentation, or declare final canonical decisions beyond the temporary source-of-truth order below.

## 2. Source-of-truth order

For future Nuvio OS work, the intended source-of-truth order is:

1. Current source code and current git status.
2. Repo-local docs and contracts.
3. Public runtime/reference repo docs.
4. Obsidian Nuvio docs.
5. Hermes/ChatGPT consolidated context.
6. Old notes/backlog only after validation.

Working rule for agents: when sources conflict, do not average them. Prefer the higher source in the list, then mark remaining uncertainty as `Unknown / needs confirmation`.

## 3. Repositories and sources inspected

### Main backoffice repo

| Field | Value |
| --- | --- |
| Path | `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio` |
| Exists? | Yes |
| Accessible? | Yes |
| Summary | PocketBase fork/custom Nuvio backoffice/backend repo with UI, scoped APIs, migrations, deployment docs, env docs, Docker/Compose docs, and the most important repo-local contracts. |
| Notable docs found | `AGENTS.md`, `docs/NUVIO_ADMIN_UI_CONTRACT.md`, `docs/NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md`, `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`, deployment/env/bootstrap/Coolify docs, `deploy/README.md`. |
| Missing/unknown | Root `README.md` is still mostly PocketBase upstream-oriented, so agents should not treat it as the primary Nuvio product guide. No single Nuvio OS source map exists yet. |

### cms5 public/template test app

| Field | Value |
| --- | --- |
| Path | `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5` |
| Exists? | Yes |
| Accessible? | Yes |
| Summary | SvelteKit public runtime/lab app. It contains current public runtime code and some older agent-oriented markdown notes under `src/`. It is valuable for public runtime behavior, but should not be treated as the clean starter template. |
| Notable docs found | `README.md`, `src/architectureReview.md`, `src/layoutBuilder.md`, `src/templateAdapter.md`, `.env.example`, Dockerfile/env comments. |
| Missing/unknown | Root `README.md` appears to be the default Svelte `sv` README, not a Nuvio-specific runtime guide. Some `src/*.md` guidance appears generator/lab-oriented and may be superseded by Reference docs for new public sites. |

### Reference public repo

| Field | Value |
| --- | --- |
| Path | `C:\Users\Leo\Documents\Nuvio\Sites\Reference` |
| Exists? | Yes |
| Accessible? | Yes |
| Summary | Clean Nuvio public website template/reference. It documents the current intended public-site integration boundary, server/client env separation, template extraction from cms5, CI/template hardening, and global source library usage. |
| Notable docs found | `AGENTS.md`, `README.md`, `docs/NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md`, `docs/NUVIO_PUBLIC_SITE_ENV_CONTRACT.md`, `docs/CMS5_PUBLIC_SITE_INTEGRATION_AUDIT.md`, `docs/CMS5_TO_TEMPLATE_EXTRACTION_PLAN.md`, `docs/TEMPLATE_REGRESSION_AUDIT.md`, `docs/TEMPLATE_BUILD.md`, `docs/TEMPLATE_ADAPTER.md`, `docs/GLOBAL_SOURCE_LIBRARY.md`. |
| Missing/unknown | Reference docs are strong for future public site repos, but not necessarily canonical for the deployed `cms5` public runtime unless current code confirms it. Analytics/Umami remains deferred in the template docs. |

### Obsidian Nuvio folder

| Field | Value |
| --- | --- |
| Path | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio` |
| Exists? | Yes |
| Accessible? | Yes |
| Summary | Human-facing Nuvio Operating Manual with product state, technical overview, operations, features, runbooks, troubleshooting, and archive folders. It is highly useful for current context but lower priority than current code and repo-local docs for implementation details. |
| Notable docs found | `00 - Nuvio Index.md`, Product State docs, Technical Overview docs, Operations docs, Feature docs, Runbooks, Troubleshooting, `99 - Archive/README.md`. |
| Missing/unknown | No final agent-focused Nuvio OS exists yet. Some Obsidian docs summarize repo docs and prior phases, so freshness must be checked against current code before implementation. |

### Hermes/ChatGPT consolidated context

| Field | Value |
| --- | --- |
| Exists? | Unknown / needs confirmation |
| Accessible? | Not inspected as a local file source in this audit |
| Summary | Potentially valuable as consolidation and task history, but not a direct implementation authority. |
| Notable docs found | None in the inspected local paths. |
| Missing/unknown | Need a future phase to identify whether durable Hermes/ChatGPT summaries exist as files and where they live. |

## 4. Documentation inventory table

| Source | Area | File/Note | Path | Purpose | Likely current? | Agent relevance | Risk of staleness | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Main repo | Agent instructions | `AGENTS.md` | `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\AGENTS.md` | Backoffice/backend agent rules, architecture assumptions, do-not-invent guidance, validation posture. | High | Critical | Medium | Must be read before implementation in main repo. Mentions Reviews in module lists; verify active/deferred state before Reviews work. |
| Main repo | Admin UI contract | `NUVIO_ADMIN_UI_CONTRACT.md` | `docs/NUVIO_ADMIN_UI_CONTRACT.md` | Shared UI primitives, operations layout, component creation boundaries. | High | Critical for UI work | Low | Strong source for backoffice UI consistency. |
| Main repo | Forms/UI contracts | `NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md` | `docs/NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md` | SchemaForm, form systems, file fields, TinyMCE, hardcoded forms. | High | Critical for CMS/forms work | Low | Prevents collapsing all forms into one abstraction. |
| Main repo | Website settings/SEO | `NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md` | `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md` | Website settings boundaries, SEO fields, runtime SEO, feature availability. | High | Critical for CMS/settings/SEO | Low/Medium | Mentions Reviews/Newsletter/Booking/Reports settings; check current feature state before changing Reviews. |
| Main repo | Deployment | `NUVIO_DEPLOYMENT_QUICK_GUIDE.md` | `docs/NUVIO_DEPLOYMENT_QUICK_GUIDE.md` | Human-practical env/deployment guide. | High | Critical for deployment | Low | Shorter guide; pair with env matrix for details. |
| Main repo | Deployment/env | `NUVIO_DEPLOYMENT_ENV_MATRIX.md` | `docs/NUVIO_DEPLOYMENT_ENV_MATRIX.md` | Full env variable reference and deployment knobs. | High | Critical for deployment/env work | Medium | Matrix is dense and may include optional/reserved aliases; verify code when changing env behavior. |
| Main repo | Instance bootstrap | `NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md` | `docs/NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md` | New instance checklist, domains, env, migrations, snapshot restore, smoke tests. | High | Critical for first-client/deployment work | Low | Good deployment operator checklist. |
| Main repo | Coolify/deployment | `NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md` | `docs/NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md` | Coolify mapping, service/domain/volume/env plan, snapshot restore, backup, smoke checklist. | High | Critical for Coolify work | Medium | Planning doc; Coolify deploy still pending per Obsidian. Needs confirmation at deploy time. |
| Main repo | Docker/local compose | `deploy/README.md` | `deploy/README.md` | Local/staging Compose run instructions and snapshot restore note. | High | High for Docker/Compose work | Low | Explains no automatic snapshot restore on startup. |
| Main repo | Security | `.github/SECURITY.md` | `.github/SECURITY.md` | Upstream/security reporting info. | Unknown | Low/Medium | Medium | Not Nuvio-specific enough to be a primary agent guide. |
| Main repo | Upstream README | `README.md` | `README.md` | PocketBase upstream overview/build/test basics. | Partly | Medium | High for Nuvio specifics | Useful for PocketBase base commands, but not Nuvio product truth. |
| Main repo | Changelog | `CHANGELOG*.md` | root changelog files | PocketBase/upstream release history. | Historical | Low | High for Nuvio specifics | Do not use as current Nuvio roadmap. |
| cms5 | Public runtime README | `README.md` | `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\README.md` | Default Svelte project commands. | Low/Medium | Medium | High | Useful for command shape only; not Nuvio-specific. |
| cms5 | Architecture review prompt | `architectureReview.md` | `src/architectureReview.md` | Senior-review checklist for cms5/gallery template work. | Medium | Medium | Medium/High | Agent prompt style; may be superseded by Reference contracts for new sites. |
| cms5 | Template/layout generator | `layoutBuilder.md` | `src/layoutBuilder.md` | Generator-oriented instructions for schema/import and page scaffolding. | Medium | Medium | High | Useful historical adapter guidance; risky if applied to current Reference or official site without validation. |
| cms5 | Template adapter | `templateAdapter.md` | `src/templateAdapter.md` | cms5 adapter/generator rules and component pattern. | Medium | Medium | High | Older/lab-oriented; Reference docs explicitly say not to copy cms5 wholesale. |
| cms5 | Env/deploy comments | `.env.example`, `Dockerfile` | cms5 root | Public runtime env/build assumptions. | Medium/High | High for runtime deploy | Medium | Confirm against current Dockerfile and code before using. |
| Reference | Agent rules | `AGENTS.md` | `C:\Users\Leo\Documents\Nuvio\Sites\Reference\AGENTS.md` | Public-site reference/template rules, boundaries, dependency constraints, source library rules. | High | Critical for Reference work | Low | Strong source for template repo behavior. |
| Reference | Template overview | `README.md` | Reference root | What Reference is/is not, docs map, env, CI, global source library, current structure. | High | High | Low | Good entry point for public site template work. |
| Reference | Public site contract | `NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md` | `docs/NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md` | Public-site responsibilities, backend endpoint rule, server-side request boundary, clean route pattern. | High | Critical for public-site work | Low | Must read before public site/template implementation. |
| Reference | Env contract | `NUVIO_PUBLIC_SITE_ENV_CONTRACT.md` | `docs/NUVIO_PUBLIC_SITE_ENV_CONTRACT.md` | Server-only vs `PUBLIC_*` environment rules. | High | Critical for public runtime/env work | Low | Prevents secret exposure in SvelteKit/browser code. |
| Reference | cms5 audit | `CMS5_PUBLIC_SITE_INTEGRATION_AUDIT.md` | `docs/CMS5_PUBLIC_SITE_INTEGRATION_AUDIT.md` | Audit of cms5 concepts useful for clean template extraction. | Medium/High | High for public runtime/template history | Medium | Audit says useful concepts should be extracted, not copied. |
| Reference | Extraction plan | `CMS5_TO_TEMPLATE_EXTRACTION_PLAN.md` | `docs/CMS5_TO_TEMPLATE_EXTRACTION_PLAN.md` | Phase plan/status for template extraction from cms5. | Medium/High | High for template roadmap | Medium | Includes phase status; verify current code before assuming incomplete work. |
| Reference | Regression audit | `TEMPLATE_REGRESSION_AUDIT.md` | `docs/TEMPLATE_REGRESSION_AUDIT.md` | Template feature/boundary audit after booking MVP. | High | High for template readiness | Low/Medium | Notes analytics/Umami and CI/hardening state at that time. |
| Reference | Template build | `TEMPLATE_BUILD.md` | `docs/TEMPLATE_BUILD.md` | How to create a real public site from Reference. | High | High for public-site build work | Low | Distinguishes Reference, cms5, real site repos, and Srcs. |
| Reference | Template adapter | `TEMPLATE_ADAPTER.md` | `docs/TEMPLATE_ADAPTER.md` | How to adapt static/custom UI into a real Nuvio-connected site. | High | High for landing/business site work | Low | Useful for official Nuvio website phases. |
| Reference | Global source library | `GLOBAL_SOURCE_LIBRARY.md` | `docs/GLOBAL_SOURCE_LIBRARY.md` | Rules for using `C:\Users\Leo\Documents\Nuvio\Srcs`. | High | High for visual/landing work | Low | Srcs is reference-only, not runtime dependency. |
| Reference | New-site checklist | `NEW_SITE_FROM_REFERENCE_CHECKLIST.md` | `docs/NEW_SITE_FROM_REFERENCE_CHECKLIST.md` | Checklist for creating a new repo/site from Reference. | High | High | Low | Good site bootstrap checklist. |
| Obsidian | Manual index | `00 - Nuvio Index.md` | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\00 - Nuvio Index.md` | Human-facing entry point and current status map. | High for context | High | Medium | Clearly says active front is deployment/smoke/backup, not new features. |
| Obsidian | Product state | `Backoffice 1.0 Status.md` | `01 - Product State/Backoffice 1.0 Status.md` | Feature-complete/beta-stable/sellable status and not-yet-online checklist. | High for context | Critical for task classification | Medium | Dated 2026-06-04; verify current code/status for exact details. |
| Obsidian | Roadmap | `Current Roadmap.md` | `01 - Product State/Current Roadmap.md` | Current phase: DevOps / instance deployment; next/later/do-not-start. | High for context | Critical for task classification | Medium | Strong warning against starting non-critical feature work. |
| Obsidian | Deferred work | `Deferred Features.md` | `01 - Product State/Deferred Features.md` | Parked features including Google Places/Reviews, exports, multi-capacity, advanced report history. | High for context | Critical for avoiding stale backlog | Low/Medium | Must be consulted before reviving old ideas. |
| Obsidian | Architecture | `Architecture Overview.md` | `02 - Technical Overview/Architecture Overview.md` | Two-service model, DTO flow, preview iframe, scoped endpoint concept, instance model. | High for context | Critical | Medium | Good architectural map; verify endpoint details in repo code/docs. |
| Obsidian | Backend overview | `Backoffice Backend.md` | `02 - Technical Overview/Backoffice Backend.md` | PocketBase fork/custom app, UI embed, pb_data, migrations, health endpoint, scoped endpoints. | High for context | High | Medium | Good orientation for backend tasks. |
| Obsidian | Data model | `Data Model Overview.md` | `02 - Technical Overview/Data Model Overview.md` | CMS vs operational collections/entities and snapshot ownership. | High for context | High | Medium | Good high-level model; schema/migrations still override. |
| Obsidian | Public runtime | `Public Runtime cms5.md` | `02 - Technical Overview/Public Runtime cms5.md` | cms5 role, rendering, public DTOs, forms, SEO, preview, env, Docker. | High for context | High for runtime work | Medium | Must be compared against current cms5 code. |
| Obsidian | Repos/branches | `Repositories and Branches.md` | `02 - Technical Overview/Repositories and Branches.md` | Repo identity, branch/tag rules, commands. | High for context | High | Medium | Tags/branch state must be checked in git before release work. |
| Obsidian | Commands | `Commands Cheat Sheet.md` | `03 - Operations/Commands Cheat Sheet.md` | Common PowerShell-friendly commands. | High for operations | High | Medium | Commands should be verified before running destructive operations. |
| Obsidian | Docker/Coolify | `Docker and Compose.md`, `Coolify Plan.md` | `03 - Operations/` | Docker runtime, Compose base, Coolify plan, build/runtime envs. | High for context | Critical for deploy | Medium | Coolify real deployment still pending. |
| Obsidian | Instance model | `Instance Model.md` | `03 - Operations/Instance Model.md` | Defines instance ownership and what code/data belongs per instance. | High | Critical | Low | Important for first-client work. |
| Obsidian | Snapshot/restore | `Snapshot and Restore.md` | `03 - Operations/Snapshot and Restore.md` | Snapshot types, CMS snapshot contents/exclusions, commands, safety. | High | Critical | Medium | Restore tooling/code must be checked before use. |
| Obsidian | Feature docs | `CMS.md`, `Assets and Images.md`, `Leads.md`, `Booking.md`, `Newsletter.md`, `Reports.md`, etc. | `04 - Features/` | Human-readable feature behavior, flows, manual tests, gotchas. | High for context | High for feature work | Medium | Useful but lower priority than code/contracts/tests. |
| Obsidian | Security hardening | `Security Hardening.md` | `04 - Features/Security Hardening.md` | Scoped endpoints, public DTO safety, XSS/content hardening, CORS/CSP/logging/env hygiene. | High for context | Critical for security/client-role work | Medium | Must pair with repo-local contracts and current tests. |
| Obsidian | Runbooks | `Create New Instance.md`, `Deploy Nuvio Base.md`, `Restore CMS Snapshot.md`, `Release New Version.md`, `Emergency Rollback.md` | `05 - Runbooks/` | Step-by-step operator workflows. | High for operations | High | Medium | Human-runbook level; commands and paths require final confirmation. |
| Obsidian | Troubleshooting | `Known Errors.md`, `Docker Issues.md`, `CMS Issues.md`, `Public Runtime Issues.md`, `Common Commands.md` | `06 - Troubleshooting/` | Diagnosis order and known issues. | High for troubleshooting | Medium/High | Medium | Use after reproducing issue; do not jump straight to broad fixes. |
| Obsidian | Archive policy | `99 - Archive/README.md` | `99 - Archive/README.md` | Rules for archived/stale docs. | High for doc hygiene | Medium | Low | Archive is not implementation truth. |

## 5. High-value docs for future agents

### Must read before any task

- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\AGENTS.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\00 - Nuvio Index.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Current Roadmap.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Deferred Features.md`
- Current repo `git status` and relevant source files/tests for the exact task.

### Must read for deployment

- `docs/NUVIO_DEPLOYMENT_QUICK_GUIDE.md`
- `docs/NUVIO_DEPLOYMENT_ENV_MATRIX.md`
- `docs/NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md`
- `docs/NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md`
- `deploy/README.md`
- Obsidian: `03 - Operations/Deployment Quick Guide.md`
- Obsidian: `03 - Operations/Coolify Plan.md`
- Obsidian: `05 - Runbooks/Deploy Nuvio Base.md`
- Obsidian: `05 - Runbooks/Create New Instance.md`

### Must read for security/client-role work

- `docs/NUVIO_ADMIN_UI_CONTRACT.md`
- `docs/NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md`
- `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`
- Obsidian: `04 - Features/Security Hardening.md`
- Obsidian: relevant feature doc, for example `Leads.md`, `Booking.md`, `Newsletter.md`, `CMS.md`.
- Current backend tests for scoped endpoints and raw PocketBase access rules.

### Must read for public runtime work

- Reference: `AGENTS.md`
- Reference: `docs/NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md`
- Reference: `docs/NUVIO_PUBLIC_SITE_ENV_CONTRACT.md`
- Reference: `docs/CMS5_PUBLIC_SITE_INTEGRATION_AUDIT.md`
- Reference: `docs/CMS5_TO_TEMPLATE_EXTRACTION_PLAN.md`
- Obsidian: `02 - Technical Overview/Public Runtime cms5.md`
- cms5 current code and `.env.example` / Dockerfile comments.

### Must read for feature work

- Main repo `AGENTS.md`
- `docs/NUVIO_ADMIN_UI_CONTRACT.md`
- `docs/NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md`
- `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`
- Obsidian `04 - Features/<feature>.md`
- Current source files and tests for the exact feature.
- `Current Roadmap.md` and `Deferred Features.md` before adding or reviving scope.

### Must read for business/landing work

- Reference: `docs/TEMPLATE_BUILD.md`
- Reference: `docs/TEMPLATE_ADAPTER.md`
- Reference: `docs/GLOBAL_SOURCE_LIBRARY.md`
- Reference: `docs/NEW_SITE_FROM_REFERENCE_CHECKLIST.md`
- Obsidian: `01 - Product State/Backoffice 1.0 Status.md`
- Obsidian: `01 - Product State/Current Roadmap.md`
- Real website repo docs/code, if the task targets `C:\Users\Leo\Documents\Nuvio\Sites\Nuvio`. Not inspected in this audit because it was not in the requested source list.

## 6. Obvious gaps

Observed or strongly implied gaps for a future Nuvio OS:

- No single agent source map exists yet.
- No task-routing doc exists that maps task types to required docs, source files, tests, and forbidden areas.
- No conflict/staleness index exists across main repo docs, Reference docs, cms5 lab docs, and Obsidian manual.
- No final agent reporting format exists for implementation phases, audits, deployment work, and security work.
- No single regression matrix exists that ties features to exact automated tests, manual smoke tests, and role/security checks.
- No current canonical list of “dangerous changes” exists in one place, though pieces exist across AGENTS, contracts, security docs, deployment docs, and runbooks.
- No central “feature do-not-touch / do-not-cross-module” guide exists; this is usually stated per prompt or per feature doc.
- No single public endpoint hardening matrix was found in repo-local docs, though Obsidian Security Hardening summarizes the concept.
- cms5 has useful lab/runtime guidance, but no polished Nuvio-specific README; its root README is still the default Svelte starter text.
- Main repo root README is PocketBase-oriented, so agents need a stronger Nuvio-specific entry point than root README.
- Hermes/ChatGPT consolidated context was not found as a local durable source in this audit.

## 7. Potential stale/conflict candidates

These are candidates for Phase 2. This audit does not deeply resolve them.

- Old roadmap/backlog language vs current release-readiness state. Obsidian says the current priority is Coolify/base deployment, snapshot restore, smoke testing, backup rehearsal, and first-client readiness.
- Reviews / Google Places. Main repo AGENTS and settings docs mention Reviews, but Obsidian Deferred Features says Google Places / Reviews sync is inactive and not required for launch.
- cms5 as runtime vs Reference as clean starter. cms5 is still the public runtime for the current Nuvio 1.0 flow, but Reference docs say cms5 is lab/dev history and not the clean template for future sites.
- Public runtime env naming and aliases. Deployment matrix lists multiple backend URL fallbacks and sender aliases; current code must decide what is truly active.
- Analytics/Umami. Deployment docs include Umami vars, current focus includes validation, but Reference template docs mark analytics implementation as deferred. Need separate classification by main deployment/runtime vs Reference template.
- Full self-service, billing, advanced CRM, and advanced newsletter workflows. If found in older backlog or external notes, they should not be treated as launch/current without confirmation.
- Reports history/snapshots. Current reports are DTO/front-end enriched; Obsidian Deferred Features parks advanced report history/snapshots.
- Booking multi-capacity. Obsidian parks it; current booking is single-capacity-oriented unless code says otherwise.
- Raw PocketBase/client-role assumptions. Current architecture emphasizes scoped endpoints and no raw PB writes for protected operational actions; old docs or code examples may not reflect this.
- Restore/reset tooling. Dev/QA reset env (`NUVIO_ALLOW_DEV_RESET`) is dangerous and must not be left enabled. Any old restore notes need careful comparison with current restore tooling and safety flags.
- Old generator docs in cms5 (`layoutBuilder.md`, `templateAdapter.md`) may conflict with Reference docs that reject wholesale cms5 copying and universal block adaptation as default.
- Root README files may be misleading: main repo root README is PocketBase-oriented and cms5 root README is default Svelte starter text.

## 8. Recommended next phase

Recommended next phase:

Phase 2 - Conflict/Staleness Audit and Canonical Source Decisions.

Phase 2 should compare:

- Main repo `AGENTS.md` and contracts vs Obsidian Operating Manual status/roadmap/deferred docs.
- Main repo deployment docs vs Obsidian deployment/runbook docs.
- cms5 public runtime docs/code comments vs Reference public-site template docs.
- Reviews/Google Places mentions across AGENTS/settings/env docs vs Deferred Features.
- Analytics/Umami deployment/runtime expectations vs Reference template deferred status.
- Snapshot/restore docs vs current snapshot/restore tool behavior and safety flags.
- Client-role/security docs vs current scoped endpoint tests and raw PocketBase shielding.
- Root README assumptions vs current Nuvio-specific docs.
- Any Hermes/ChatGPT summary files if a durable local source can be identified.

Phase 2 output should not rewrite the whole docs set. It should produce a conflict index, a canonical-doc decision table, and a first draft of an agent task-routing map.