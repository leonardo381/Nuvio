# Phase 5C — Agent Task-Pack Dry Run

## Executive verdict

Usable with minor improvements before patching; ready for real agent use after the targeted patches listed below.

The task-pack layer now gives future agents a clear route for recurring Nuvio work. The strongest pattern is consistent: each pack points back to Core, Danger Zones, current source/git status, validation expectations, reporting format, and explicit stop conditions. The dry run found no missing primary pack for the eight simulated prompts.

The main risk is not missing documentation. It is overclaiming live/provider/deployment facts that cannot be known from docs alone. The packs correctly require `Unknown / needs confirmation` for live Coolify, DNS, credentials, provider, test-user, restore, and smoke-test state.

## Dry-run matrix

### 1. Landing + Umami

| Field | Dry-run result |
| --- | --- |
| Simulated task | Build the Nuvio public landing v1 and wire Umami tracking. |
| Primary task pack | [Landing and Umami](../task_packs/LANDING_UMAMI_TASK_PACK.md) |
| Required docs | [CORE](../CORE.md), [DANGER_ZONES](../DANGER_ZONES.md), [VALIDATION_MATRIX](../VALIDATION_MATRIX.md), [REPORTING_FORMATS](../REPORTING_FORMATS.md), [Public Runtime](../features/PUBLIC_RUNTIME.md), [Reports Analytics Health](../features/REPORTS_ANALYTICS_HEALTH.md), [Umami Analytics Operations](../operations/UMAMI_ANALYTICS_OPERATIONS.md), current target site repo source/git status. |
| Optional docs | [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md), [First Client Readiness](../launch/FIRST_CLIENT_READINESS.md), real-site docs, provider configuration if supplied. |
| Docs explicitly not needed | Booking, Newsletter, Snapshot/Restore, and client-role packs unless landing work touches those flows. |
| Routing clarity | Clear. |
| Constraints clarity | Strong. |
| Validation clarity | Adequate. Public site build/check/lint and Umami smoke depend on target repo/provider state. |
| Report clarity | Strong. |
| Stop conditions clarity | Strong. |
| Likely agent failure modes | Editing Reference instead of the real target site, inventing pricing/guarantees, exposing provider secrets in browser env, claiming Umami works without provider confirmation. |
| Patches made | None. The pack already separates public tracking, provider status, and business unknowns well enough. |

### 2. First deployment

| Field | Dry-run result |
| --- | --- |
| Simulated task | Prepare the first production-like Nuvio base deployment. |
| Primary task pack | [First Deployment](../task_packs/FIRST_DEPLOYMENT_TASK_PACK.md) |
| Required docs | [CORE](../CORE.md), [DANGER_ZONES](../DANGER_ZONES.md), [VALIDATION_MATRIX](../VALIDATION_MATRIX.md), [REPORTING_FORMATS](../REPORTING_FORMATS.md), [Deployment Coolify](../operations/DEPLOYMENT_COOLIFY.md), [Docker Compose](../operations/DOCKER_COMPOSE.md), [Env Secrets](../operations/ENV_SECRETS.md), [Instance Bootstrap](../operations/INSTANCE_BOOTSTRAP.md), [Snapshot Restore](../operations/SNAPSHOT_RESTORE.md), [Backup Rollback](../operations/BACKUP_ROLLBACK.md), [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md), [Nuvio Base Deployment Readiness](../launch/NUVIO_BASE_DEPLOYMENT_READINESS.md). |
| Optional docs | Existing repo deployment quick guide, env matrix, Coolify plan, deployment README, live provider notes if supplied. |
| Docs explicitly not needed | Feature implementation cards except as smoke references for enabled flows. |
| Routing clarity | Clear. |
| Constraints clarity | Strong. |
| Validation clarity | Strong for planning; live validation requires explicit deployment scope and provider access. |
| Report clarity | Strong. |
| Stop conditions clarity | Strong. |
| Likely agent failure modes | Treating docs as proof that DNS/Coolify/secrets are ready, running Docker/deploy/restore without approval, assuming restore mechanism, sharing `pb_data`/storage, putting secrets in `VITE_*`. |
| Patches made | None. The pack is explicit about live provider state and destructive action approval. |

### 3. Client-role security smoke

| Field | Dry-run result |
| --- | --- |
| Simulated task | Validate that a client-role user can only access assigned websites and does not hit raw PB 403s on allowed pages. |
| Primary task pack | [Client Role Security Smoke](../task_packs/CLIENT_ROLE_SECURITY_SMOKE_TASK_PACK.md) |
| Required docs | [CORE](../CORE.md), [DANGER_ZONES](../DANGER_ZONES.md), [VALIDATION_MATRIX](../VALIDATION_MATRIX.md), [REPORTING_FORMATS](../REPORTING_FORMATS.md), [Security Client Role](../features/SECURITY_CLIENT_ROLE.md), affected feature cards for pages being smoked, current auth/UI/backend source and git status. |
| Optional docs | [Smoke Validation Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md), Obsidian security docs, current endpoint tests if code behavior is inspected. |
| Docs explicitly not needed | Deployment, snapshot, landing, and provider docs unless the smoke runs against a deployed instance or provider-backed feature. |
| Routing clarity | Clear. |
| Constraints clarity | Strong after patch. |
| Validation clarity | Strong after patch. |
| Report clarity | Strong. |
| Stop conditions clarity | Strong after patch. |
| Likely agent failure modes | Mistaking UI hiding for security, widening backend rules to avoid a raw PB 403, printing test credentials/PII, skipping unassigned-website negative checks. |
| Patches made | Added explicit raw PB 403/allowed-page warnings, validation, and stop condition. |

### 4. Booking E2E regression

| Field | Dry-run result |
| --- | --- |
| Simulated task | Run Booking public-to-backoffice E2E regression and identify blockers. |
| Primary task pack | [Booking E2E Regression](../task_packs/BOOKING_E2E_REGRESSION_TASK_PACK.md) |
| Required docs | [CORE](../CORE.md), [DANGER_ZONES](../DANGER_ZONES.md), [VALIDATION_MATRIX](../VALIDATION_MATRIX.md), [REPORTING_FORMATS](../REPORTING_FORMATS.md), [Booking](../features/BOOKING.md), [Emails Templates](../features/EMAILS_TEMPLATES.md), current Booking backend/UI/public route source and git status. |
| Optional docs | [Env Secrets](../operations/ENV_SECRETS.md) if email/provider behavior is in scope, [Client Role Security Smoke](../task_packs/CLIENT_ROLE_SECURITY_SMOKE_TASK_PACK.md) if client-role Booking is checked. |
| Docs explicitly not needed | Reports, Newsletter, Landing, and Snapshot/Restore unless regression crosses those areas. |
| Routing clarity | Clear. |
| Constraints clarity | Strong. |
| Validation clarity | Strong. |
| Report clarity | Strong. |
| Stop conditions clarity | Strong. |
| Likely agent failure modes | Sending real emails/calendar invites, forcing status from frontend, changing slot/status behavior during validation, overclaiming auto-confirm behavior without settings. |
| Patches made | None. The pack already calls out status, email, `.ics`, settings, and client-role risk. |

### 5. Newsletter lifecycle

| Field | Dry-run result |
| --- | --- |
| Simulated task | Validate newsletter subscribe, confirm, unsubscribe, and campaign readiness. |
| Primary task pack | [Newsletter Lifecycle](../task_packs/NEWSLETTER_LIFECYCLE_TASK_PACK.md) |
| Required docs | [CORE](../CORE.md), [DANGER_ZONES](../DANGER_ZONES.md), [VALIDATION_MATRIX](../VALIDATION_MATRIX.md), [REPORTING_FORMATS](../REPORTING_FORMATS.md), [Newsletter](../features/NEWSLETTER.md), [Emails Templates](../features/EMAILS_TEMPLATES.md), [Env Secrets](../operations/ENV_SECRETS.md), current Newsletter backend/UI/public route source and git status. |
| Optional docs | [Client Role Security Smoke](../task_packs/CLIENT_ROLE_SECURITY_SMOKE_TASK_PACK.md) if scoped Newsletter access is checked, provider notes if supplied. |
| Docs explicitly not needed | Booking, Reports, Snapshot/Restore, and landing docs unless campaign readiness crosses provider/deploy/site copy. |
| Routing clarity | Clear. |
| Constraints clarity | Strong. |
| Validation clarity | Strong. |
| Report clarity | Strong. |
| Stop conditions clarity | Strong. |
| Likely agent failure modes | Printing lifecycle tokens, sending real campaigns, treating draft/save readiness as send readiness, exposing provider secrets, changing send behavior during polish. |
| Patches made | None. The pack already covers lifecycle tokens, provider side effects, campaigns, and client-role readiness. |

### 6. CMS + SEO + public rendering

| Field | Dry-run result |
| --- | --- |
| Simulated task | Validate CMS content, translated SEO, preview, sitemap, robots, and public rendering. |
| Primary task pack | [CMS SEO Public Rendering](../task_packs/CMS_SEO_PUBLIC_RENDERING_TASK_PACK.md) |
| Required docs | [CORE](../CORE.md), [DANGER_ZONES](../DANGER_ZONES.md), [VALIDATION_MATRIX](../VALIDATION_MATRIX.md), [REPORTING_FORMATS](../REPORTING_FORMATS.md), [CMS](../features/CMS.md), [Assets](../features/ASSETS.md), [Website Settings SEO](../features/WEBSITE_SETTINGS_SEO.md), [Public Runtime](../features/PUBLIC_RUNTIME.md), website settings/SEO contract, current CMS/backend/public runtime source and git status. |
| Optional docs | Public runtime deployment docs if deployed runtime smoke is in scope, snapshot docs if restored content/assets are part of the validation. |
| Docs explicitly not needed | Booking, Newsletter, Reports, and deployment docs unless public runtime/deploy behavior is being validated live. |
| Routing clarity | Clear after patch. |
| Constraints clarity | Strong. |
| Validation clarity | Strong after patch. |
| Report clarity | Strong. |
| Stop conditions clarity | Strong. |
| Likely agent failure modes | Claiming translated SEO/i18n works without observing runtime output, adding unavailable sitemap/canonical checks, moving SEO fields into settings, missing native file/storage issues. |
| Patches made | Added explicit translated SEO/i18n wording to purpose, danger zones, and manual smoke validation. |

### 7. Reports + Umami + Health

| Field | Dry-run result |
| --- | --- |
| Simulated task | Validate reports, Umami tracking/events, and health check for demo readiness. |
| Primary task pack | [Reports Umami Health](../task_packs/REPORTS_UMAMI_HEALTH_TASK_PACK.md) |
| Required docs | [CORE](../CORE.md), [DANGER_ZONES](../DANGER_ZONES.md), [VALIDATION_MATRIX](../VALIDATION_MATRIX.md), [REPORTING_FORMATS](../REPORTING_FORMATS.md), [Reports Analytics Health](../features/REPORTS_ANALYTICS_HEALTH.md), [Umami Analytics Operations](../operations/UMAMI_ANALYTICS_OPERATIONS.md), [Env Secrets](../operations/ENV_SECRETS.md), current Reports/provider source and git status. |
| Optional docs | Deployment/Coolify docs if health is checked on a live deployment, public runtime docs if public tracking events are part of scope. |
| Docs explicitly not needed | Booking, Newsletter, Snapshot/Restore, and CMS docs unless those flows provide demo data for Reports. |
| Routing clarity | Clear after patch. |
| Constraints clarity | Strong. |
| Validation clarity | Strong after patch. |
| Report clarity | Strong. |
| Stop conditions clarity | Strong. |
| Likely agent failure modes | Confusing public Umami event tracking with Reports provider API checks, claiming health means full app readiness, exposing provider credentials, inventing metrics not backed by DTO/provider data. |
| Patches made | Added explicit public tracking/events vs Reports provider API scope and manual smoke wording. |

### 8. Snapshot / restore / backup

| Field | Dry-run result |
| --- | --- |
| Simulated task | Prepare snapshot/restore/backup validation for a demo/staging instance. |
| Primary task pack | [Snapshot Restore Backup](../task_packs/SNAPSHOT_RESTORE_BACKUP_TASK_PACK.md) |
| Required docs | [CORE](../CORE.md), [DANGER_ZONES](../DANGER_ZONES.md), [VALIDATION_MATRIX](../VALIDATION_MATRIX.md), [REPORTING_FORMATS](../REPORTING_FORMATS.md), [Snapshot Restore](../operations/SNAPSHOT_RESTORE.md), [Backup Rollback](../operations/BACKUP_ROLLBACK.md), [Assets](../features/ASSETS.md), current snapshot/restore tool source and git status if execution or implementation is requested. |
| Optional docs | [Deployment Coolify](../operations/DEPLOYMENT_COOLIFY.md), [Docker Compose](../operations/DOCKER_COMPOSE.md), Obsidian Snapshot/Restore and Coolify docs, manifest/storage output if supplied. |
| Docs explicitly not needed | Landing, Newsletter, Reports, and Booking implementation docs unless post-restore smoke covers those flows. |
| Routing clarity | Clear. |
| Constraints clarity | Strong. |
| Validation clarity | Strong. |
| Report clarity | Strong. |
| Stop conditions clarity | Strong. |
| Likely agent failure modes | Running restore in a docs-only phase, touching `pb_data`, assuming backup exists, restoring records without storage files, leaving dev reset enabled, confusing CMS snapshot with operational backup. |
| Patches made | None. The pack already has explicit action modes, storage-file warnings, and destructive stop conditions. |

## Cross-cutting issues

- Live/provider/deployment facts remain outside documentation. Agents must not claim DNS, Coolify, health, email, Umami, backup, restore, or smoke success unless directly verified in the task context.
- Several packs intentionally require current source and git status. This is correct, but future agents may still over-trust docs if they skip source inspection.
- Provider and public-runtime analytics language can be easy to blur. The Reports patch now makes public tracking/events vs provider API scope explicit.
- Security/client-role tasks need to prevent both security leaks and false negatives from allowed pages using forbidden raw collection APIs. The client-role patch now makes raw PB 403s on allowed scoped flows explicit.
- CMS public rendering tasks include many adjacent concerns. The CMS patch now makes translated SEO/i18n explicit so agents do not validate only the default-language path.

## Patches made

- `docs/NUVIO_OS/task_packs/CLIENT_ROLE_SECURITY_SMOKE_TASK_PACK.md`: added explicit raw PB 403 allowed-page danger, validation, and stop-condition wording.
- `docs/NUVIO_OS/task_packs/CMS_SEO_PUBLIC_RENDERING_TASK_PACK.md`: added translated SEO/i18n wording to purpose, danger zones, and runtime validation.
- `docs/NUVIO_OS/task_packs/REPORTS_UMAMI_HEALTH_TASK_PACK.md`: added public tracking/events vs Reports provider API scope and manual smoke validation wording.
- `docs/NUVIO_OS/audits/2026-06-18_AGENT_TASK_PACK_DRY_RUN.md`: added this dry-run audit report.

## Remaining unknowns

- Real Coolify/DNS/TLS/provider account state.
- Real env values, secret-store state, and whether optional providers are enabled.
- Actual client-role test users, assigned websites, and safe fixture data.
- Real deployment health URLs and public runtime URLs.
- Approved snapshot name, website ID, target `pb_data`, and backup target.
- Whether public Umami tracking/events are currently implemented/enabled in the real site being tested.
- Which target repo is in scope for future public landing/site work.

## Recommendation

Start using one real task pack for actual work, preferably a low-risk audit or readiness task first. The task-pack layer is ready enough for real agent use, and the best next validation will come from using one pack end-to-end on live work.

A later Phase 5D OS final hardening pass is still useful after the first real use, but it does not need to block task-pack adoption.

