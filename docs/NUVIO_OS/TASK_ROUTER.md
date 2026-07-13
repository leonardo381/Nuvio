# Task Router

## Purpose

Route agent tasks to the correct Nuvio OS docs, source contracts, validation checks, and report format.

Always start with [CORE.md](CORE.md), current git status, and the target repo `AGENTS.md` if present.

## Universal Routing Rules

- If the task touches auth, public endpoints, restore, env/secrets, booking, newsletter tokens, migrations, analytics credentials, deployment, or client-role data access, read [DANGER_ZONES.md](DANGER_ZONES.md) first.
- If the task is docs-only, still verify git status and final changed files.
- If a command is not documented or fails due to missing setup, report it instead of inventing a replacement silently.
- If the task revives a deferred feature, classify it before action and pause/report unless the user explicitly scoped it.
- Preserve current architecture and do not introduce new patterns without audit.
- Current source code and git status override documentation.
- If the task matches a recurring task pack, start with [task_packs/README.md](task_packs/README.md) before assembling docs manually.


## Recurring Task Packs

For common recurring work, prefer the matching task pack. Task packs do not replace Core, Danger Zones, Validation Matrix, Reporting Formats, or current source code.

| Recurring task | Task pack |
| --- | --- |
| Nuvio landing and Umami | [Landing and Umami](task_packs/LANDING_UMAMI_TASK_PACK.md) |
| First production-like deployment | [First Deployment](task_packs/FIRST_DEPLOYMENT_TASK_PACK.md) |
| Five-flow release regression | [Five Flow Regression](task_packs/FIVE_FLOW_REGRESSION_TASK_PACK.md) |
| Client-role security smoke | [Client Role Security Smoke](task_packs/CLIENT_ROLE_SECURITY_SMOKE_TASK_PACK.md) |
| CMS, SEO, and public rendering | [CMS SEO Public Rendering](task_packs/CMS_SEO_PUBLIC_RENDERING_TASK_PACK.md) |
| Leads, Contact, and WhatsApp E2E | [Leads Contact WhatsApp E2E](task_packs/LEADS_CONTACT_WHATSAPP_E2E_TASK_PACK.md) |
| Booking E2E regression | [Booking E2E Regression](task_packs/BOOKING_E2E_REGRESSION_TASK_PACK.md) |
| Newsletter lifecycle | [Newsletter Lifecycle](task_packs/NEWSLETTER_LIFECYCLE_TASK_PACK.md) |
| Emails and templates E2E | [Emails and Templates E2E](task_packs/EMAILS_TEMPLATES_E2E_TASK_PACK.md) |
| Reports, Umami, and health | [Reports Umami Health](task_packs/REPORTS_UMAMI_HEALTH_TASK_PACK.md) |
| Env/secrets review | [Environment and Secrets Review](task_packs/ENV_SECRETS_REVIEW_TASK_PACK.md) |
| Snapshot, restore, and backup | [Snapshot Restore Backup](task_packs/SNAPSHOT_RESTORE_BACKUP_TASK_PACK.md) |
| Demo flow and data | [Demo Flow and Data](task_packs/DEMO_FLOW_AND_DATA_TASK_PACK.md) |
| First-client onboarding | [First Client Onboarding](task_packs/FIRST_CLIENT_ONBOARDING_TASK_PACK.md) |
## Primary Router

| Task type | Classification default | Docs to read | What not to change | Validation required | Expected report |
| --- | --- | --- | --- | --- | --- |
| Any task | Depends | [CORE.md](CORE.md), [OS_NAVIGATION.md](OS_NAVIGATION.md), current git status, target `AGENTS.md`. | Unrelated files, stale backlog, broad refactors. | Status/diff review; task-specific checks. | Standard or audit report from [REPORTING_FORMATS.md](REPORTING_FORMATS.md). |
| UI polish | Polish | Main `AGENTS.md`, `docs/NUVIO_ADMIN_UI_CONTRACT.md`, target feature card. | New visual system, unrelated redesign, local custom primitives when shared primitives exist. | UI build/check for target app; manual visual checklist. | Standard implementation report. |
| SchemaForm/form change | Bugfix or readiness | `docs/NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md`, [features/CMS.md](features/CMS.md). | TinyMCE/file upload/parser behavior unless requested; broad SchemaForm rewrite. | UI build/check; manual save/load; backend tests if endpoint/schema touched. | Standard report with form behavior impact. |
| Website settings/SEO | Bugfix or readiness | `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md`, [features/WEBSITE_SETTINGS_SEO.md](features/WEBSITE_SETTINGS_SEO.md), [CMS SEO Public Rendering task pack](task_packs/CMS_SEO_PUBLIC_RENDERING_TASK_PACK.md). | Moving SEO into `websites.settings`, overwriting hidden keys, duplicate settings sources. | UI build; backend tests if endpoint touched; public SEO smoke if runtime touched. | Standard report with unknown-key and SEO output confirmation. |
| CMS/content editing | Bugfix/readiness/polish | Main `AGENTS.md`, [features/CMS.md](features/CMS.md), [features/ASSETS.md](features/ASSETS.md), [CMS SEO Public Rendering task pack](task_packs/CMS_SEO_PUBLIC_RENDERING_TASK_PACK.md). | Broad CMS redesign, raw PB writes, unsafe restore/reset, asset policy changes. | UI build; CMS save/preview/assets smoke; backend tests if endpoints touched. | Standard report with preview/i18n/asset notes. |
| Leads/contact/WhatsApp | Bugfix or readiness | [features/LEADS_CONTACT_WHATSAPP.md](features/LEADS_CONTACT_WHATSAPP.md), [features/SECURITY_CLIENT_ROLE.md](features/SECURITY_CLIENT_ROLE.md), [Leads Contact WhatsApp E2E task pack](task_packs/LEADS_CONTACT_WHATSAPP_E2E_TASK_PACK.md). | Raw PB writes, PII exposure, unsupported public payload fields, broken unified Leads model. | UI build if UI touched; backend Leads/contact tests if backend touched; public contact/WhatsApp smoke. | Standard or security report. |
| Booking | Launch-critical if behavior; polish if visual only | [features/BOOKING.md](features/BOOKING.md), [DANGER_ZONES.md](DANGER_ZONES.md), [Booking E2E Regression task pack](task_packs/BOOKING_E2E_REGRESSION_TASK_PACK.md). | Slot logic, status defaults, public payloads, `.ics`, notifications unless explicitly requested. | UI build if UI touched; backend booking tests if logic touched; public booking smoke. | Standard report with slot/status/email impact. |
| Newsletter | Launch-critical if lifecycle/send; polish if visual only | [features/NEWSLETTER.md](features/NEWSLETTER.md), [features/EMAILS_TEMPLATES.md](features/EMAILS_TEMPLATES.md), [Newsletter Lifecycle task pack](task_packs/NEWSLETTER_LIFECYCLE_TASK_PACK.md). | Token leaks, provider secrets in browser, send behavior during save/copy polish. | UI build; backend newsletter tests if backend touched; subscribe/confirm/unsubscribe smoke. | Standard report with token/send impact. |
| Reports/analytics/health | Readiness | [features/REPORTS_ANALYTICS_HEALTH.md](features/REPORTS_ANALYTICS_HEALTH.md), [operations/UMAMI_ANALYTICS_OPERATIONS.md](operations/UMAMI_ANALYTICS_OPERATIONS.md), [Reports Umami Health task pack](task_packs/REPORTS_UMAMI_HEALTH_TASK_PACK.md). | Fake metrics, provider secrets in browser, unsupported analytics claims. | UI build if UI touched; backend Reports tests if backend touched; Umami configured/unconfigured smoke. | Standard report with data-source and PII confirmation. |
| Public runtime/site template | Readiness or launch-critical | [features/PUBLIC_RUNTIME.md](features/PUBLIC_RUNTIME.md), [operations/PUBLIC_RUNTIME_DEPLOYMENT.md](operations/PUBLIC_RUNTIME_DEPLOYMENT.md). | Copying cms5 wholesale, exposing server env, adding deps without approval, backend/schema changes from public repo. | `npm run check`, `npm run lint`, `npm run build` where available; route smoke. | Standard report with target repo and env boundary. |
| Website Factory / client website build | Readiness or launch-critical | [agent_processes/processes/website_factory/WEBSITE_FACTORY_PROCESS.md](agent_processes/processes/website_factory/WEBSITE_FACTORY_PROCESS.md), [agent_processes/AGENT_PROCESS_STANDARD.md](agent_processes/AGENT_PROCESS_STANDARD.md), [features/PUBLIC_RUNTIME.md](features/PUBLIC_RUNTIME.md). | Mixing strategy/design/code/copy/integration/QA/deploy in one stage, modifying source blocks, using cms5/Reference/Srcs as runtime dependencies. | Stage artifact checks; target public-site check/lint/build only when implementation stage touches code; final deployment smoke if deploy stage is explicitly scoped. | Standard or process-stage report with artifacts, stage, validation, severity, and next stage. |
| Nuvio official landing / own website | Launch-critical if acquisition/demo path; otherwise readiness | [features/PUBLIC_RUNTIME.md](features/PUBLIC_RUNTIME.md), [launch/FIRST_CLIENT_READINESS.md](launch/FIRST_CLIENT_READINESS.md), [Landing and Umami task pack](task_packs/LANDING_UMAMI_TASK_PACK.md), real site source. | Final pricing invention, unsupported promises, editing Reference instead of real site, runtime dependency on `Srcs`. | Public site check/lint/build if code touched; manual copy/CTA route review. | Standard report with business-claim and route checks. |
| Production-like deployment | Launch-critical | [operations/DEPLOYMENT_COOLIFY.md](operations/DEPLOYMENT_COOLIFY.md), [launch/NUVIO_BASE_DEPLOYMENT_READINESS.md](launch/NUVIO_BASE_DEPLOYMENT_READINESS.md), [operations/ENV_SECRETS.md](operations/ENV_SECRETS.md), [First Deployment task pack](task_packs/FIRST_DEPLOYMENT_TASK_PACK.md). | Real secrets in docs, wildcard CORS, shared volumes, auto restore on startup, architecture changes. | Deployment smoke from [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md); health, preview, public runtime, backup checks. | Deployment report. |
| Coolify deployment | Launch-critical when executing; readiness when planning | [operations/DEPLOYMENT_COOLIFY.md](operations/DEPLOYMENT_COOLIFY.md), `docs/NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md`. | Assuming account/domain/provider readiness, exposing secrets, changing architecture. | Health checks and Coolify smoke if executing; docs/status check if planning. | Deployment report with unknowns. |
| Public runtime deployment | Launch-critical if first deploy; readiness otherwise | [operations/PUBLIC_RUNTIME_DEPLOYMENT.md](operations/PUBLIC_RUNTIME_DEPLOYMENT.md), [features/PUBLIC_RUNTIME.md](features/PUBLIC_RUNTIME.md), [First Deployment task pack](task_packs/FIRST_DEPLOYMENT_TASK_PACK.md). | Mixing cms5/Reference/real site roles, server env in browser, hardcoded localhost. | Public runtime build/route smoke for target repo; CORS/CSP/frame checks. | Deployment or standard report. |
| Umami analytics validation | Readiness; launch-critical only if demo/reporting promise depends on it | [operations/UMAMI_ANALYTICS_OPERATIONS.md](operations/UMAMI_ANALYTICS_OPERATIONS.md), [features/REPORTS_ANALYTICS_HEALTH.md](features/REPORTS_ANALYTICS_HEALTH.md), [Reports Umami Health task pack](task_packs/REPORTS_UMAMI_HEALTH_TASK_PACK.md). | PII in analytics, provider secrets in browser, fake analytics. | Umami configured/unconfigured smoke; Reports traffic state check. | Standard report with PII/secrets confirmation. |
| First-client readiness | Launch-critical | [launch/FIRST_CLIENT_READINESS.md](launch/FIRST_CLIENT_READINESS.md), [launch/LAUNCH_BLOCKERS_VS_POLISH.md](launch/LAUNCH_BLOCKERS_VS_POLISH.md), [First Client Onboarding task pack](task_packs/FIRST_CLIENT_ONBOARDING_TASK_PACK.md). | Blocking on deferred polish; ignoring security/deploy gaps. | Five-flow smoke, security/client-role smoke, deployment/backup checks. | Audit or readiness report. |
| Demo flow | Readiness | [launch/DEMO_FLOW_RUNBOOK.md](launch/DEMO_FLOW_RUNBOOK.md), [operations/SMOKE_VALIDATION_TROUBLESHOOTING.md](operations/SMOKE_VALIDATION_TROUBLESHOOTING.md), [Demo Flow and Data task pack](task_packs/DEMO_FLOW_AND_DATA_TASK_PACK.md). | Fake data presented as real, skipped smoke hidden as success. | Demo flow checklist; skipped checks reported. | Regression/demo report. |
| Security/client-role smoke | Launch-critical | [features/SECURITY_CLIENT_ROLE.md](features/SECURITY_CLIENT_ROLE.md), [operations/SMOKE_VALIDATION_TROUBLESHOOTING.md](operations/SMOKE_VALIDATION_TROUBLESHOOTING.md), [Client Role Security Smoke task pack](task_packs/CLIENT_ROLE_SECURITY_SMOKE_TASK_PACK.md). | UI-only security, raw PB writes, missing websiteAccess checks, token/PII logs. | Backend auth/security tests if code touched; manual admin/assigned/unassigned client smoke. | Security review report. |
| Email/newsletter/booking E2E validation | Launch-critical if enabled | [features/EMAILS_TEMPLATES.md](features/EMAILS_TEMPLATES.md), [features/NEWSLETTER.md](features/NEWSLETTER.md), [features/BOOKING.md](features/BOOKING.md), [operations/ENV_SECRETS.md](operations/ENV_SECRETS.md), [Emails and Templates E2E task pack](task_packs/EMAILS_TEMPLATES_E2E_TASK_PACK.md). | Provider secrets in `VITE_*`, token leakage, send semantics changes, `.ics` changes. | Email/newsletter smoke from [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md); booking email/`.ics` smoke if enabled. | Standard report with lifecycle/link/provider impact. |
| Backup/rollback | Launch-critical before handoff; readiness while planning | [operations/BACKUP_ROLLBACK.md](operations/BACKUP_ROLLBACK.md), [operations/SNAPSHOT_RESTORE.md](operations/SNAPSHOT_RESTORE.md), [Snapshot Restore Backup task pack](task_packs/SNAPSHOT_RESTORE_BACKUP_TASK_PACK.md). | Invented backup commands, destructive restore, DB/storage split, untested rollback. | Backup/restore proof or checklist; no destructive commands unless explicitly requested. | Deployment or audit report. |
| Snapshot/restore | Launch-critical for base bootstrap; unsafe if destructive | [operations/SNAPSHOT_RESTORE.md](operations/SNAPSHOT_RESTORE.md), [features/ASSETS.md](features/ASSETS.md), [DANGER_ZONES.md](DANGER_ZONES.md), [Snapshot Restore Backup task pack](task_packs/SNAPSHOT_RESTORE_BACKUP_TASK_PACK.md). | Auto restore on startup, wrong `pb_data`, missing storage files, leaving dev reset enabled. | Restore dry run/real output only if requested; records+storage verification. | Deployment/audit report with explicit action mode. |
| Env/secrets review | Launch-critical before deploy | [operations/ENV_SECRETS.md](operations/ENV_SECRETS.md), [Environment and Secrets Review task pack](task_packs/ENV_SECRETS_REVIEW_TASK_PACK.md), `docs/NUVIO_DEPLOYMENT_ENV_MATRIX.md`. | Real secrets in docs, secrets in `VITE_*`, wildcard CORS, localhost production origins. | Env review; no product build unless env code changed; deployment smoke if deployed. | Security or deployment report. |
| Agent handoff | Readiness | [launch/AGENT_HANDOFF_CHECKLIST.md](launch/AGENT_HANDOFF_CHECKLIST.md), [REPORTING_FORMATS.md](REPORTING_FORMATS.md). | Claiming success without verification, hiding skipped checks, losing unknowns. | Git status, file list, validation/skipped check summary. | Documentation/handoff report. |
| Regression/testing | Readiness | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md), relevant feature/operation card. | Destructive commands, assuming command names, hiding skipped checks. | Run documented relevant checks; report unavailable/failing checks honestly. | Regression report. |
| New feature request | Usually deferred/enhancement until classified | Current Roadmap, Deferred Features, [launch/READINESS_DECISION_MATRIX.md](launch/READINESS_DECISION_MATRIX.md). | Reviving deferred backlog, scope creep, broad architecture change. | Audit/classify first; validation depends on touched area. | Audit or standard report. |
| Refactor request | Unsafe/unknown until audited | Main `AGENTS.md`, relevant contracts, current source/tests. | Behavior drift, cross-module churn, public contract changes, unnecessary migrations. | Existing tests/builds for touched area; diff review; manual smoke if behavior risk exists. | Standard report with behavior compatibility. |

## Layer Entry Points

| Need | Go to |
| --- | --- |
| Feature/module work | [features/README.md](features/README.md) |
| Deployment/operations work | [operations/README.md](operations/README.md) |
| Demo/first-client/launch work | [launch/README.md](launch/README.md) |
| One-jump navigation | [OS_NAVIGATION.md](OS_NAVIGATION.md) |
| Nuvio OS quality check | [OS_QA_CHECKLIST.md](OS_QA_CHECKLIST.md) |

## Expected Reports

Use [REPORTING_FORMATS.md](REPORTING_FORMATS.md):

- Standard Implementation Report for code/docs implementation.
- Audit-Only Report for investigation phases.
- Regression Report for review/testing phases.
- Deployment Report for deploy/env/infra tasks.
- Security Review Report for auth/client/public endpoint/security tasks.
- Documentation Update Report for Nuvio OS docs phases.
- Blocked / Unknown Report when required inputs are missing.



