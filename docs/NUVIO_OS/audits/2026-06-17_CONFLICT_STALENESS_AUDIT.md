# Nuvio OS Conflict/Staleness Audit

Date: 2026-06-17

## 1. Purpose

This audit compares the documentation sources discovered in Phase 1 and classifies what future agents should treat as canonical, useful-but-partial, stale/superseded, or unknown.

The goal is to prepare for an agent-focused Nuvio OS without creating the final OS files yet. Future Codex/Hermes/AI agents should be able to use this audit to avoid stale backlog, preserve architecture decisions, route tasks to the right docs, and distinguish launch-critical readiness work from polish or deferred product ideas.

This report does not change product code, rewrite existing docs, or implement Nuvio OS. It creates a decision layer over the existing documentation landscape.

## 2. Inputs reviewed

### Main backoffice repo

- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\docs\NUVIO_OS\audits\2026-06-17_SOURCE_INVENTORY.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\AGENTS.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\docs\NUVIO_ADMIN_UI_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\docs\NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\docs\NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\docs\NUVIO_DEPLOYMENT_QUICK_GUIDE.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\docs\NUVIO_DEPLOYMENT_ENV_MATRIX.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\docs\NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\docs\NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\deploy\README.md`
- `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\README.md`

### Reference public-site repo

- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\AGENTS.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\README.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NUVIO_PUBLIC_SITE_ENV_CONTRACT.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\CMS5_PUBLIC_SITE_INTEGRATION_AUDIT.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\CMS5_TO_TEMPLATE_EXTRACTION_PLAN.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\TEMPLATE_REGRESSION_AUDIT.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\TEMPLATE_BUILD.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\TEMPLATE_ADAPTER.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\GLOBAL_SOURCE_LIBRARY.md`
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NEW_SITE_FROM_REFERENCE_CHECKLIST.md`

### cms5 public runtime/test app

- `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\README.md`
- `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\src\architectureReview.md`
- `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\src\layoutBuilder.md`
- `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\src\templateAdapter.md`
- `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\.env.example`
- `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\Dockerfile`

### Obsidian Operating Manual

- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\00 - Nuvio Index.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Backoffice 1.0 Status.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Current Roadmap.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Deferred Features.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\02 - Technical Overview\Architecture Overview.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\02 - Technical Overview\Backoffice Backend.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\02 - Technical Overview\Data Model Overview.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\02 - Technical Overview\Public Runtime cms5.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\02 - Technical Overview\Repositories and Branches.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Commands Cheat Sheet.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Coolify Plan.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Deployment Quick Guide.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Docker and Compose.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Git Workflow.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Instance Model.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Snapshot and Restore.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Booking.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\CMS.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Contact Form and WhatsApp.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Leads.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Newsletter.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Public Runtime.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Reports.md`
- `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Security Hardening.md`

### Not found as durable local inputs

- Hermes/ChatGPT consolidated context files were not found locally in Phase 1.
- Plan B / business plan files were not found in the inspected source list. Treat Plan B as external/unknown unless later provided as a file.

## 3. Canonical source hierarchy

1. **Implementation truth**: current source code, tests, migrations, package scripts, Dockerfiles, and current git status. This overrides every document.
2. **Repo-local implementation contracts**: main repo `AGENTS.md` and contract docs under `docs/`. These are canonical for backoffice/backend implementation details.
3. **Repo-local deployment docs**: deployment quick guide, env matrix, bootstrap checklist, Coolify plan, and deploy README. These are canonical for deployment planning, but final values must be verified against code and provider state.
4. **Reference public-site docs**: canonical for clean future public-site template direction. Reference overrides cms5 for new public website starter decisions.
5. **cms5 docs/code notes**: useful for current public runtime/testing history. Not canonical for clean template direction.
6. **Obsidian Operating Manual**: valuable for product state, operations, roadmap, deferred features, runbooks, and feature overview context. Not authoritative over repo code/contracts for implementation detail.
7. **Hermes/ChatGPT context**: useful only when explicitly provided or stored as a file. Not found locally in Phase 1.
8. **Old roadmap/backlog notes**: never active by default. Use only after confirmation against current code, current docs, and user intent.
## 4. Conflict/staleness matrix

| Topic | Sources compared | Conflict or risk | Canonical decision | Agent instruction | Status |
| --- | --- | --- | --- | --- | --- |
| Overall project state: broad backlog vs release-readiness | Obsidian Index, Current Roadmap, Backoffice 1.0 Status; main repo docs | Old backlog energy can tempt agents into feature expansion. | Nuvio is in release-readiness mode: deployment, demo, first-client readiness, validation, and Nuvio website/landing work are current. | Do not start broad new product work unless explicitly requested. Classify new requests as readiness, launch-critical, polish, deferred, or unsafe. | Canonical |
| Backoffice architecture | Main `AGENTS.md`; Obsidian Architecture Overview and Backoffice Backend; current code | Obsidian is high-level and may omit implementation detail. | Backoffice/backend is a PocketBase fork/custom app with embedded UI, scoped Nuvio APIs, migrations, `pb_data`, and per-instance data. | For implementation, inspect code/tests first. Use Obsidian for orientation only. | Canonical |
| Admin UI standards | Main `AGENTS.md`; `NUVIO_ADMIN_UI_CONTRACT.md`; Obsidian feature docs | Per-feature prompts may ask for local polish and risk new visual systems. | Shared Nuvio UI rhythm is canonical: operations headers/tabs, summary pills, labels, buttons, OverlayPanel, Form v2 rhythm. | Do not invent feature-specific visual systems. Preserve established patterns unless explicitly tasked. | Canonical |
| SchemaForm/forms behavior | `NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md`; main `AGENTS.md`; CMS docs | Risk of forcing all forms into SchemaForm. | Multiple form systems are valid. SchemaForm is canonical for dynamic block/component editing; workflow forms may remain hardcoded. | Do not unify all forms. Treat file fields and TinyMCE as high-risk. | Canonical |
| Website settings + SEO model | `NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`; Obsidian CMS/Public Runtime/Security docs | Risk of moving SEO into `websites.settings` or exposing technical settings. | Top-level website/page SEO fields are canonical. `websites.settings` stores feature configuration and must preserve hidden keys. | Do not duplicate SEO/settings sources. Do not show raw JSON or technical keys to clients. | Canonical |
| Deployment/Coolify/env model | Main deployment quick guide, env matrix, Coolify plan, bootstrap checklist, deploy README; Obsidian operations docs | Docs are planning/placeholder-heavy; Coolify real deploy is pending. | Canonical deployment model is two services, isolated `pb_data`, exact origins, build-time `VITE_*`, server-only secrets, controlled snapshot restore. | Verify actual Dockerfiles/code before deploy. Never put secrets in `VITE_*`. Do not auto-restore snapshots on startup. | Canonical |
| Instance model | Obsidian Instance Model; deployment docs; architecture docs | Agents may confuse clients with branches/repos. | Clients are instances/deployments, not branches or shared writable data. Each instance owns env, domains, `pb_data`, storage, backups, and CMS snapshot. | Do not share writable `pb_data` or storage across clients. Keep code shared and data isolated. | Canonical |
| Snapshot/restore | Main Coolify plan, deploy README, bootstrap checklist; Obsidian Snapshot and Restore, runbooks | Restore tooling is powerful and dangerous; docs are operational, not proof of current command behavior. | CMS snapshot restore is a controlled one-off flow. It must include records and physical storage files. `NUVIO_ALLOW_DEV_RESET` is dangerous/dev-only. | Confirm restore tool flags/current code before use. Stop backend if required. Do not leave reset env enabled. | Canonical |
| Public runtime direction | Obsidian Public Runtime cms5/Public Runtime; Reference contract/build docs; cms5 docs | cms5 is current runtime history, but Reference is clean future template. | Use cms5 for current Nuvio 1.0 runtime behavior; use Reference for future public site template direction. | Do not copy cms5 wholesale into Reference or real sites. Verify current cms5 code for deployed runtime tasks. | Useful but partial |
| cms5 role | cms5 README and `src/*.md`; Reference docs; Obsidian runtime docs | cms5 docs include default Svelte README and lab/generator notes. | cms5 is a public runtime/test/dev app and source of historical contracts, not the canonical starter template. | Treat cms5 docs with caution. Use current code for runtime behavior and Reference docs for new site architecture. | Stale/superseded |
| Reference repo role | Reference `AGENTS.md`, README, public-site contracts, template build/adapter docs | Risk of turning Reference into a real website or overloading it with client visuals. | Reference is the clean public-site template candidate. Real websites must be separate repos/apps created from it. | Do not put client-specific content, real domains, secrets, or final Nuvio official site work into Reference. | Canonical |
| Real client public websites | Reference Template Build/Adapter, New Site checklist, Global Source Library | Risk of runtime dependency on Reference/cms5/Srcs. | Real public websites can be custom but should preserve Nuvio integration boundaries. They own their UI, copy, routes, and env. | Copy/adapt selected assets into the real site repo. Do not runtime-import from `Srcs` or cms5. | Canonical |
| Client-role/auth/websiteAccess | Main `AGENTS.md`; Obsidian Security/feature docs; current scoped endpoint tests | UI hiding can be mistaken for security. | Auth/tenant model is admin/client role plus `websiteAccess`. Security must be enforced in scoped endpoints/backend checks. | Do not rely on UI-only restrictions. For client-role changes, inspect backend auth tests and scoped endpoints. | Canonical |
| Raw PocketBase vs scoped endpoints | Main AGENTS; Security Hardening; feature docs; current code/tests | Old PB examples or default PocketBase docs may imply raw record access. | Protected operational flows should use scoped custom endpoints; do not reintroduce raw PB writes for client-role product actions. | If a UI needs data mutation, find or add scoped endpoint only when proven missing. Do not call raw PB collections casually. | Canonical |
| CMS/current editing model | Main AGENTS, SchemaForm contract, Website Settings/SEO contract, Obsidian CMS/Assets docs | Risk of moving data boundaries or broad restore/reset changes. | CMS editing uses existing page/block/settings/save flows, SchemaForm for dynamic props, scoped assets, SEO fields at top-level. | Do not redesign CMS. Preserve hidden settings keys, assets policy, preview iframe, and SEO field boundaries. | Canonical |
| Leads vs Contact Form/WhatsApp | Obsidian Leads and Contact Form/WhatsApp; current code/tests; main AGENTS | Risk of treating contact and WhatsApp as unrelated or overbuilding CRM. | Contact Form and WhatsApp are unified conceptually as Leads. Leads are operational opportunity capture, not a full CRM. | Use scoped Leads endpoints and current DTO fields. Keep UI/client copy human-readable. | Canonical |
| Booking sensitivity | Obsidian Booking docs; current booking code/tests; main AGENTS | Booking touches public forms, slots, appointments, emails, and status logic. | Booking is sensitive scheduling workflow; not a casual UI-only area unless task is explicitly UI polish. | Avoid changing slot/status/email payloads without tests. Treat multi-capacity as deferred. | Canonical |
| Newsletter scope | Obsidian Newsletter docs; current code; deployment/email docs | Risk of overselling newsletter as advanced automation or exposing provider secrets. | Newsletter supports subscribers, groups, campaigns, public lifecycle, and server-side send. It is not advanced CRM/automation. | Keep provider secrets server-side. Keep save/send semantics separate. Do not promise open/click analytics unless implemented. | Canonical |
| Reports/analytics/Umami | Obsidian Reports docs; deployment env docs; Reference template docs | Main deployment needs Umami validation, while Reference template analytics is deferred. | Reports are client-friendly operational dashboards. Umami is current expected analytics provider for Reports/traffic, but Reference template analytics implementation remains deferred. | Do not show fake analytics. Keep provider credentials server-side. Separate deployed runtime needs from template roadmap. | Conflicting / needs confirmation |
| Reviews/Google Places | Main AGENTS/settings docs mention Reviews; deployment env docs list Google Places; Obsidian Deferred Features says deferred/inactive. | Mentioned in some current docs, but not launch-critical. | Reviews/Google Places is deferred/inactive for Nuvio 1.0 unless user explicitly revives it. | Do not start Reviews/Google Places work before base deployment unless directly requested and scoped. | Canonical |
| Security hardening/public endpoints | Main contracts, Obsidian Security Hardening, current tests/code | Some docs summarize; source code/tests prove actual behavior. | Scoped endpoints, public DTO safety, validation, CORS/CSP/frame policy, token/PII log redaction, and safe errors are mandatory. | For any public endpoint/security work, inspect tests and current middleware. Do not expose secrets or raw provider details. | Canonical |
| Email/templates | Deployment docs, Obsidian Booking/Newsletter/Contact docs, current backend code | Feature docs are high-level; provider behavior is code-dependent. | Email is server-side/provider-backed. Resend/server env values are optional per enabled feature. Newsletter lifecycle and booking/contact notifications must avoid exposing secrets/tokens. | Do not put email secrets in `VITE_*`. Validate lifecycle URLs and token/log redaction. | Useful but partial |
| First-client readiness | Obsidian Index/Roadmap/Status; deployment docs | Agents may drift into feature polish instead of deploy readiness. | First-client readiness means deploy path, restored CMS snapshot, smoke tests, demo data/flows, backup rehearsal, and version records. | Prioritize readiness blockers over deferred enhancements. | Canonical |
| Pricing/business positioning | User-provided context; Reference/real-site work; no durable Plan B file found | Pricing is not final and business plan files were not found. | Pricing is a positioning/conversion source, not an implementation checklist. Exact prices are unknown unless explicitly provided. | Do not invent final prices. Use cautious copy and mark business assumptions as needing confirmation. | Unknown / needs confirmation |
| Deferred features | Obsidian Deferred Features and Current Roadmap; Phase 1 audit | Old docs may describe deferred work as desirable. | Deferred means parked, not active. Examples: Google Places/Reviews sync, data exports, booking multi-capacity, advanced report history/snapshots, self-host migration. | Do not mix deferred items into readiness/deploy tasks. Create a small explicit phase prompt if revived. | Canonical |
| Root READMEs/default starter docs | Main repo README, cms5 README, Reference README | Main/cms5 root READMEs are not Nuvio-specific enough. | Reference README is useful and current for Reference. Main and cms5 root READMEs are cautionary/default/upstream docs. | Use main/cms5 root READMEs for basic commands only, never product state or architecture truth. | Stale/superseded |

## 5. Canonical decisions for Nuvio OS

Future agents must treat these as current unless explicitly overridden by current source code, current git status, or a direct user instruction.

- Nuvio is mostly built and is in release-readiness mode, not broad early backlog mode.
- Current operating front: deployment, demo, first-client readiness, validation, Nuvio website/landing, analytics validation, regression testing, public endpoint hardening, email/newsletter/booking E2E checks, reports confidence, and backup/restore rehearsal.
- Five critical demo flows are Website settings/setup, CMS + SEO + public rendering, Leads / Contact / WhatsApp, Booking, and Reports / Analytics / Health.
- Backoffice/backend is central, reusable, and instance-scoped through data/env/domains rather than per-client code forks.
- Clients are instances/deployments, not branches or shared repos.
- Every instance owns its env, `pb_data`, storage, domains, backups, and CMS snapshot/bootstrap data.
- Public websites should be separate apps/repos/instances.
- cms5 is useful for current public runtime/testing history, but it is not the canonical starter for new public websites.
- Reference is the clean public-site reference/template candidate.
- Client-role access must be backed by scoped custom endpoints and website access checks, not just hidden UI.
- Protected client-role product flows must not reintroduce raw PocketBase writes.
- Auth/tenant model is admin/client role plus `websiteAccess`.
- Website Settings and SEO must preserve data boundaries: SEO identity fields are top-level, feature config belongs in `websites.settings`, and hidden keys must be preserved.
- SchemaForm is not the only form system. Do not force workflow forms into SchemaForm.
- CMS changes should preserve existing save flows, preview behavior, asset policy, and SEO field boundaries.
- Contact Form and WhatsApp are unified conceptually as Leads.
- Booking is sensitive. Do not casually change slot logic, appointment status logic, public payloads, or booking email behavior.
- Newsletter supports subscribers, groups, campaigns, public lifecycle, and scoped send; do not oversell it as advanced automation.
- Reports should be client-friendly operational dashboards, not raw analytics warehouses or fake insight generators.
- Umami analytics needs real deploy validation; provider secrets must remain server-side.
- Reviews / Google Places is deferred/inactive for Nuvio 1.0.
- Pricing is not final. Do not invent exact commercial pricing unless a current business source is provided.
- `NUVIO_ALLOW_DEV_RESET` is dangerous/dev-only and must not stay enabled on running services.
- Snapshot restore must be controlled and one-off. Do not run restore automatically on container startup.
- CMS snapshots must include records and physical storage files.
- `VITE_*` values are browser-exposed and must never contain secrets.
- `PUBLIC_*` values in SvelteKit public sites may be browser-exposed; non-public env must stay server-only.
- Root README files are not enough to understand Nuvio. Agents must use Nuvio-specific docs/contracts.
## 6. Docs to treat as authoritative

### Always read / global agent context

| Path | Why it matters | Task types |
| --- | --- | --- |
| `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\AGENTS.md` | Primary backoffice/backend agent rules, architecture assumptions, and safety posture. | Any main repo task, feature work, UI work, backend work. |
| `docs/NUVIO_OS/audits/2026-06-17_SOURCE_INVENTORY.md` | Source inventory and first-pass staleness risks. | Nuvio OS work, doc-routing work. |
| `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\00 - Nuvio Index.md` | Current human operating map and status. | Any broad Nuvio task, planning, triage. |
| `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Current Roadmap.md` | Declares current phase and do-not-start areas. | Any task that might expand scope. |
| `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Deferred Features.md` | Prevents reviving parked work accidentally. | New feature requests, roadmap, backlog triage. |

### UI/forms/settings

| Path | Why it matters | Task types |
| --- | --- | --- |
| `docs/NUVIO_ADMIN_UI_CONTRACT.md` | Canonical shared admin UI rhythm and component rules. | UI polish, layout normalization, feature pages. |
| `docs/NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md` | Canonical form-system boundaries. | SchemaForm, CMS props, workflow forms, TinyMCE, file fields. |
| `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md` | Canonical settings/SEO data boundaries and UI rules. | Website settings, SEO, public runtime SEO, CMS settings. |

### Deployment/operations

| Path | Why it matters | Task types |
| --- | --- | --- |
| `docs/NUVIO_DEPLOYMENT_QUICK_GUIDE.md` | Practical required/optional env guide. | New instance, deployment, env setup. |
| `docs/NUVIO_DEPLOYMENT_ENV_MATRIX.md` | Full variable reference. | Env audit, provider configuration, deployment debugging. |
| `docs/NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md` | Full new instance checklist. | First-client setup, staging, QA instance bootstrap. |
| `docs/NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md` | Current Coolify mapping and deploy plan. | Coolify deployment, domains, volumes, build args, smoke checks. |
| `deploy/README.md` | Local Compose workflow and snapshot restore note. | Local/staging compose validation. |
| Obsidian `03 - Operations/*` | Human runbooks and command context. | Operations, release, rollback, troubleshooting. |

### Public runtime/reference

| Path | Why it matters | Task types |
| --- | --- | --- |
| `C:\Users\Leo\Documents\Nuvio\Sites\Reference\AGENTS.md` | Canonical Reference repo rules. | Reference/template work. |
| `Reference/docs/NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md` | Public-site module responsibilities and request boundaries. | Public site template, real site integration. |
| `Reference/docs/NUVIO_PUBLIC_SITE_ENV_CONTRACT.md` | Server-only vs browser-safe env contract. | Public site env/config, SvelteKit boundary checks. |
| `Reference/docs/TEMPLATE_BUILD.md` | Build process and repo role separation. | New real site from Reference. |
| `Reference/docs/TEMPLATE_ADAPTER.md` | Static/custom UI adaptation rules. | Nuvio landing/business website work. |
| `Reference/docs/GLOBAL_SOURCE_LIBRARY.md` | Rules for `Srcs` usage. | Landing pages, asset/template adaptation. |
| Obsidian `02 - Technical Overview/Public Runtime cms5.md` | Current cms5 runtime orientation. | cms5 runtime/deployment/public DTO tasks. |

### Security/permissions

| Path | Why it matters | Task types |
| --- | --- | --- |
| Obsidian `04 - Features/Security Hardening.md` | Security posture, scoped endpoint and public endpoint checklist. | Security, client-role, public endpoints, CORS/CSP. |
| Main `AGENTS.md` | Warns about UI-only security, hidden controls, and current code truth. | Permission/UI/security changes. |
| Feature-specific Obsidian docs | Explain client-role behavior and scoped operations per module. | Leads, Booking, Newsletter, Reports, CMS. |
| Current backend tests/source | Actual security implementation truth. | Any permission-affecting code. |

### Feature work

| Path | Why it matters | Task types |
| --- | --- | --- |
| Obsidian `04 - Features/CMS.md` | CMS editing model, preview, SEO, save flows. | CMS. |
| Obsidian `04 - Features/Assets and Images.md` | Scoped assets and native file fields. | Assets/uploads/images. |
| Obsidian `04 - Features/Leads.md` | Leads list/detail/bulk/invite behavior. | Leads. |
| Obsidian `04 - Features/Contact Form and WhatsApp.md` | Contact/WhatsApp public and Leads relationship. | Contact/WhatsApp. |
| Obsidian `04 - Features/Booking.md` | Booking workflow, public flow, backoffice operations. | Booking. |
| Obsidian `04 - Features/Newsletter.md` | Newsletter subscribers/groups/campaigns/lifecycle. | Newsletter. |
| Obsidian `04 - Features/Reports.md` | Reports DTO, analytics, tab behavior, limits. | Reports/analytics. |
| Current feature source/tests | Actual behavior and validation. | Every implementation task. |

### Business/landing/positioning

| Path | Why it matters | Task types |
| --- | --- | --- |
| Obsidian `01 - Product State/Backoffice 1.0 Status.md` | Product state and readiness framing. | Nuvio landing, sales/readiness. |
| Obsidian `01 - Product State/Current Roadmap.md` | Current priority and not-now work. | Positioning and scope triage. |
| Reference `docs/TEMPLATE_ADAPTER.md` | How to adapt static/custom UI into a real Nuvio site. | Official website/landing pages. |
| Reference `docs/GLOBAL_SOURCE_LIBRARY.md` | Safe use of reusable visual source library. | Marketing site UI adaptation. |
| Real Nuvio website repo docs/source | Actual official site implementation truth. | Official website tasks. |

## 7. Docs to treat with caution

| Path | Caution reason | How agents should use it |
| --- | --- | --- |
| `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\README.md` | Mostly PocketBase upstream README, not Nuvio-specific product truth. | Use only for generic PocketBase/build context after checking Nuvio docs and code. |
| `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\CHANGELOG*.md` | Upstream/historical release notes, not Nuvio roadmap. | Do not use for current product priorities. |
| `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\README.md` | Default Svelte `sv` README. | Use only for basic Svelte commands if needed; not product architecture. |
| `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\src\layoutBuilder.md` | Generator/lab instructions; may conflict with clean Reference direction. | Historical only; do not apply to Reference or real sites without explicit task. |
| `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\src\templateAdapter.md` | Older cms5 adapter/generator guidance. | Use for historical pattern investigation; Reference docs override for new sites. |
| `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5\src\architectureReview.md` | Strong review prompt but not a current source map. | Useful for review mindset; not authoritative for implementation. |
| `Reference/docs/CMS5_TO_TEMPLATE_EXTRACTION_PLAN.md` | Phase plan/status can become stale as template implementation advances. | Pair with current Reference code and `TEMPLATE_REGRESSION_AUDIT.md`. |
| `Reference/docs/CMS5_PUBLIC_SITE_INTEGRATION_AUDIT.md` | Audit of cms5, not final implementation contract. | Use for understanding extracted concepts; do not copy code wholesale. |
| Obsidian docs with dated status | Human context may lag current code/git state. | Use for context and task classification, then verify in repo. |
| Any old backlog/roadmap notes outside active manual | Old ideas may not be launch-critical. | Treat as inactive unless current docs/code/user request confirms. |
| External Plan B/business plan notes | Not found locally in Phase 1. Pricing/business decisions unknown. | Use only if explicitly provided; do not infer final prices. |

## 8. Docs or areas needing update later

Do not update these in this phase. These are future doc hygiene candidates.

- Main repo root `README.md` should eventually get a Nuvio-specific entry section or point agents/operators to Nuvio docs.
- cms5 root `README.md` should eventually explain its Nuvio public runtime/test role instead of default Svelte starter text.
- A final `docs/NUVIO_OS/` core pack should be created after this audit.
- A task-routing document should map every common task type to docs, source files, forbidden areas, and validation commands.
- A regression matrix should map features to automated tests and manual smoke checks.
- A dangerous-changes index should centralize high-risk areas: auth, data, public endpoints, booking, newsletter tokens, restore, migrations, secrets.
- Reviews/Google Places mentions should eventually be annotated as deferred/inactive where they appear in active implementation docs, if not already clear.
- Analytics/Umami docs should clarify the split between deployed runtime/reporting needs and Reference template deferred implementation.
- Deployment docs should be refreshed after the first real Coolify deploy, because current Coolify docs are still plans/placeholders.
- Snapshot/restore docs should be refreshed after the first real restore rehearsal.
- Business/pricing/landing positioning needs a durable canonical source if it will guide agents.
- Hermes/ChatGPT summaries, if used, need a known storage location and lower-priority source classification.
## 9. First draft: agent task-routing map

| Task type | Required docs to read first | Optional docs | Forbidden/risky areas | Required validation |
| --- | --- | --- | --- | --- |
| Any task | Current `git status`; repo `AGENTS.md`; Phase 1 inventory; this audit; current source files for target area. | Obsidian Index/Roadmap/Deferred. | Acting from stale docs, broad refactors, touching unrelated modules. | At minimum status/diff review. Run task-specific checks. |
| UI polish | Main `AGENTS.md`; `NUVIO_ADMIN_UI_CONTRACT.md`; target component/source. | Obsidian feature doc. | New visual system, unrelated redesign, custom modals/drawers when shared primitives exist. | UI build/check command for touched app; visual/manual checklist if applicable. |
| SchemaForm/form change | `NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md`; main `AGENTS.md`; target form files. | CMS feature doc. | TinyMCE, file upload behavior, forcing all forms into SchemaForm. | UI build/check; targeted manual form save/load test; backend tests if endpoint/schema touched. |
| Website settings/SEO | `NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`; CMS/Public Runtime docs; current settings/SEO source. | Obsidian CMS, Public Runtime, Security. | Moving SEO into `websites.settings`, overwriting hidden keys, duplicating settings sources. | UI build; backend tests if settings endpoints touched; manual SEO/runtime smoke. |
| CMS/content editing | Main AGENTS; Admin UI contract; SchemaForm contract; Website Settings/SEO contract; Obsidian CMS/Assets. | Snapshot and Restore docs. | Raw PB writes, broad CMS redesign, unsafe restore/reset, file upload policy changes. | UI build; targeted CMS save/preview/assets manual test; backend tests if scoped endpoints touched. |
| Leads/contact/WhatsApp | Obsidian Leads and Contact Form/WhatsApp; Security Hardening; current Leads/contact source/tests. | Reports doc for DTO impacts. | Raw PB writes, exposing PII, changing public payloads casually, breaking unified Leads concept. | UI build for Leads UI; backend Leads/contact tests if backend touched; public contact/WhatsApp smoke. |
| Booking | Obsidian Booking; Security Hardening; current booking backend/UI/source/tests. | Deployment/email docs for notifications. | Slot logic, appointment status/defaults, public booking payload, email side effects, multi-capacity. | UI build if UI touched; backend booking tests if backend touched; public booking E2E/manual smoke. |
| Newsletter | Obsidian Newsletter; Security Hardening; current newsletter source/tests. | Email/provider docs. | Lifecycle token leaks, provider secrets in browser, changing send behavior during save polish, overselling automation. | UI build; backend newsletter tests if backend touched; subscribe/confirm/unsubscribe smoke. |
| Reports/analytics/Umami | Obsidian Reports; deployment env docs; Security Hardening; current reports source/tests. | Public runtime docs for tracking/runtime. | Fake analytics, exposing provider credentials, claiming unavailable metrics, advanced snapshots/history. | UI build; backend reports tests if backend touched; Umami configured/unconfigured smoke. |
| Public runtime/site template | Reference AGENTS; Reference public-site/env contracts; Template Build/Adapter; current target repo code. | cms5 audit for history; Obsidian Public Runtime cms5 for runtime context. | Copying cms5 wholesale, exposing server env, adding dependencies without approval, backend/schema changes from public repo. | `npm run check`, `npm run lint`, `npm run build` where available; route smoke. |
| Deployment/Coolify/env | Deployment quick guide; env matrix; bootstrap checklist; Coolify plan; deploy README; Obsidian operations/runbooks. | Docker and Compose, Troubleshooting. | Real secrets in docs, wildcard CORS, missing build-time `VITE_*`, auto snapshot restore on startup, shared volumes. | `docker compose config/build/up` as scoped; health checks; smoke checklist; backup/restore rehearsal. |
| Security/client-role/permissions | Main AGENTS; Security Hardening; relevant feature doc; current backend middleware/endpoint tests. | Deployment CORS/CSP docs. | UI-only security, raw PB writes, missing websiteAccess checks, token/PII logs. | Backend auth/security tests; UI build if UI touched; manual client-role smoke. |
| Email/templates | Deployment env docs; Obsidian Newsletter/Booking/Contact docs; current backend email code. | Public runtime lifecycle docs. | Provider secrets in `VITE_*`, token leakage, changing send semantics without tests. | Backend email-related tests if present; lifecycle/notification smoke with safe env. |
| Nuvio landing/business positioning | Reference Template Adapter/Build/Global Source Library; Obsidian Product State/Roadmap; real site repo docs/source. | External business plan only if provided. | Final pricing invention, unsupported promises, editing Reference instead of real site, runtime dependency on Srcs. | Public site `check/lint/build`; manual route review. |
| Regression/testing | Obsidian Commands Cheat Sheet; feature docs; package scripts; current tests. | Troubleshooting docs. | Running destructive commands, assuming command names, skipping failing tests silently. | Run documented relevant commands; report unavailable/failing commands honestly. |
| New feature request | Current Roadmap; Deferred Features; main AGENTS; relevant feature docs/source. | Product state docs. | Reviving deferred backlog, broad architecture changes, hidden scope creep. | Start with classification/audit if scope is unclear; validation depends on touched area. |
| Refactor request | Main AGENTS; relevant contracts; current code/tests. | Architecture docs. | Refactor without behavior tests, cross-module churn, changing public contracts, data migrations without need. | Existing tests/builds for touched area; diff review; manual smoke if behavior risk exists. |

## 10. Recommended Nuvio OS structure

Proposed final structure for `docs/NUVIO_OS/`:

```text
docs/NUVIO_OS/
  README.md
  CORE.md
  SOURCE_OF_TRUTH.md
  TASK_ROUTER.md
  CURRENT_OPERATING_STATE.md
  DANGER_ZONES.md
  VALIDATION_MATRIX.md
  REPORTING_FORMATS.md
  CANONICAL_DECISIONS.md
  docs-map/
    MAIN_REPO_DOC_MAP.md
    PUBLIC_RUNTIME_DOC_MAP.md
    REFERENCE_DOC_MAP.md
    OBSIDIAN_DOC_MAP.md
  features/
    CMS.md
    ASSETS.md
    LEADS_CONTACT_WHATSAPP.md
    BOOKING.md
    NEWSLETTER.md
    REPORTS_ANALYTICS.md
    WEBSITE_SETTINGS_SEO.md
    PUBLIC_RUNTIME.md
    SECURITY_CLIENT_ROLE.md
  operations/
    DEPLOYMENT_COOLIFY.md
    INSTANCE_BOOTSTRAP.md
    SNAPSHOT_RESTORE.md
    RELEASE_READINESS.md
    FIRST_CLIENT_READINESS.md
  templates/
    IMPLEMENTATION_REPORT_TEMPLATE.md
    AUDIT_REPORT_TEMPLATE.md
    REVIEW_REPORT_TEMPLATE.md
    DEPLOYMENT_REPORT_TEMPLATE.md
    BUGFIX_PROMPT_TEMPLATE.md
    FEATURE_PHASE_PROMPT_TEMPLATE.md
  audits/
    2026-06-17_SOURCE_INVENTORY.md
    2026-06-17_CONFLICT_STALENESS_AUDIT.md
```

Purpose of key files:

- `README.md`: agent entry point and shortest safe operating path.
- `CORE.md`: compact agent doctrine: what Nuvio is, current phase, what not to do.
- `SOURCE_OF_TRUTH.md`: canonical hierarchy and conflict resolution rules.
- `TASK_ROUTER.md`: route each task type to docs/source/tests/forbidden areas.
- `CURRENT_OPERATING_STATE.md`: release-readiness snapshot and current critical paths.
- `DANGER_ZONES.md`: auth, data, public endpoints, booking, newsletter tokens, restore, migrations, secrets.
- `VALIDATION_MATRIX.md`: commands/tests/manual checks by module.
- `REPORTING_FORMATS.md`: required final report shapes for implementation, audit, review, deploy, bugfix.
- `CANONICAL_DECISIONS.md`: durable decisions extracted from audits and current docs.
- `features/`: agent-focused feature cards, not full human docs.
- `operations/`: agent-focused deployment/runbook routing.
- `templates/`: prompt/report scaffolds for repeatable phase work.
- `audits/`: historical audit evidence, not final operating docs.

## 11. Recommended next phase

Recommended next phase:

Phase 3 - Create Nuvio OS Core Pack.

Phase 3 should create only the core agent operating files, not all feature guides yet.

Recommended files to create:

- `docs/NUVIO_OS/README.md`
- `docs/NUVIO_OS/CORE.md`
- `docs/NUVIO_OS/SOURCE_OF_TRUTH.md`
- `docs/NUVIO_OS/CURRENT_OPERATING_STATE.md`
- `docs/NUVIO_OS/CANONICAL_DECISIONS.md`
- `docs/NUVIO_OS/TASK_ROUTER.md`
- `docs/NUVIO_OS/DANGER_ZONES.md`
- `docs/NUVIO_OS/VALIDATION_MATRIX.md`
- `docs/NUVIO_OS/REPORTING_FORMATS.md`

Inputs Phase 3 should use:

- `docs/NUVIO_OS/audits/2026-06-17_SOURCE_INVENTORY.md`
- `docs/NUVIO_OS/audits/2026-06-17_CONFLICT_STALENESS_AUDIT.md`
- Main repo `AGENTS.md`
- Main repo contract docs
- Main repo deployment docs
- Reference public-site docs
- Obsidian Index, Current Roadmap, Deferred Features, Product State, Operations docs, Security Hardening

Phase 3 should not create feature-specific deep guides yet. It should create a compact operating layer that future agents can read before any task, then Phase 4 can add module-level feature cards and validation maps.