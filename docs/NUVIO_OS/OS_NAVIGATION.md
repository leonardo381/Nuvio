# Nuvio OS Navigation

## Purpose

One-jump navigation for agents. Use this when you know the task shape and need the right Nuvio OS docs fast.

Start with [CORE.md](CORE.md), then use this guide to route.
## Recurring Task Packs First

If the request matches a recurring task, start with [Task Packs](task_packs/README.md) before using the one-jump tables below. Task packs are optimized for repeated jobs such as first deployment, five-flow regression, client-role smoke, Booking E2E, Newsletter lifecycle, env/secrets review, snapshot/restore, demo flow, and first-client onboarding.

Use this navigation guide when the task is one-off, feature-specific, or does not clearly match a task pack.
## I Need To Run An Agent Process

| Task | Read first | Then read | Validation | Danger zone |
| --- | --- | --- | --- | --- |
| Build a client/official website from brief to deploy | [Website Factory](agent_processes/processes/website_factory/README.md) | [Agent Process Standard](agent_processes/AGENT_PROCESS_STANDARD.md), [Public Runtime](features/PUBLIC_RUNTIME.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Source blocks, cms5 vs Reference, public runtime, secrets, deployment. |

## I Need To Work On A Feature / Module

| Task | Read first | Then read | Validation | Danger zone |
| --- | --- | --- | --- | --- |
| CMS/content editing | [features/CMS.md](features/CMS.md) | [features/ASSETS.md](features/ASSETS.md), [features/WEBSITE_SETTINGS_SEO.md](features/WEBSITE_SETTINGS_SEO.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | SchemaForm, raw PB writes, preview, assets. |
| Assets/images/media | [features/ASSETS.md](features/ASSETS.md) | [operations/SNAPSHOT_RESTORE.md](operations/SNAPSHOT_RESTORE.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Storage files, upload policy, SVG, restore. |
| Website settings/SEO | [features/WEBSITE_SETTINGS_SEO.md](features/WEBSITE_SETTINGS_SEO.md) | [features/PUBLIC_RUNTIME.md](features/PUBLIC_RUNTIME.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Hidden keys, public SEO output, legacy keys. |
| Leads/contact/WhatsApp | [features/LEADS_CONTACT_WHATSAPP.md](features/LEADS_CONTACT_WHATSAPP.md) | [features/SECURITY_CLIENT_ROLE.md](features/SECURITY_CLIENT_ROLE.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | PII, public endpoints, raw PB writes. |
| Booking | [features/BOOKING.md](features/BOOKING.md) | [features/EMAILS_TEMPLATES.md](features/EMAILS_TEMPLATES.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Slot logic, auto-confirm, public payloads, `.ics`. |
| Newsletter | [features/NEWSLETTER.md](features/NEWSLETTER.md) | [features/EMAILS_TEMPLATES.md](features/EMAILS_TEMPLATES.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Tokens, send behavior, provider secrets. |
| Reports/analytics/health | [features/REPORTS_ANALYTICS_HEALTH.md](features/REPORTS_ANALYTICS_HEALTH.md) | [operations/UMAMI_ANALYTICS_OPERATIONS.md](operations/UMAMI_ANALYTICS_OPERATIONS.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Fake analytics, PII, provider secrets. |
| Public runtime | [features/PUBLIC_RUNTIME.md](features/PUBLIC_RUNTIME.md) | [operations/PUBLIC_RUNTIME_DEPLOYMENT.md](operations/PUBLIC_RUNTIME_DEPLOYMENT.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | cms5 vs Reference, server/client env, CORS/CSP. |
| Security/client role | [features/SECURITY_CLIENT_ROLE.md](features/SECURITY_CLIENT_ROLE.md) | [DANGER_ZONES.md](DANGER_ZONES.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | UI-only security, websiteAccess, raw PB writes. |
| Emails/templates | [features/EMAILS_TEMPLATES.md](features/EMAILS_TEMPLATES.md) | [operations/ENV_SECRETS.md](operations/ENV_SECRETS.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Tokens, Resend, public links, `.ics`. |

## I Need To Work On Deployment / Operations

| Task | Read first | Then read | Validation | Danger zone |
| --- | --- | --- | --- | --- |
| Production-like deployment | [operations/DEPLOYMENT_COOLIFY.md](operations/DEPLOYMENT_COOLIFY.md) | [launch/NUVIO_BASE_DEPLOYMENT_READINESS.md](launch/NUVIO_BASE_DEPLOYMENT_READINESS.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Secrets, volumes, CORS/CSP, restore. |
| Coolify deployment | [operations/DEPLOYMENT_COOLIFY.md](operations/DEPLOYMENT_COOLIFY.md) | [operations/ENV_SECRETS.md](operations/ENV_SECRETS.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Provider/domain assumptions, build-time `VITE_*`. |
| Docker/Compose | [operations/DOCKER_COMPOSE.md](operations/DOCKER_COMPOSE.md) | [operations/PUBLIC_RUNTIME_DEPLOYMENT.md](operations/PUBLIC_RUNTIME_DEPLOYMENT.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Volumes, ports, env injection. |
| Instance bootstrap | [operations/INSTANCE_BOOTSTRAP.md](operations/INSTANCE_BOOTSTRAP.md) | [operations/SNAPSHOT_RESTORE.md](operations/SNAPSHOT_RESTORE.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Shared `pb_data`, wrong snapshot. |
| Env/secrets review | [operations/ENV_SECRETS.md](operations/ENV_SECRETS.md) | [features/SECURITY_CLIENT_ROLE.md](features/SECURITY_CLIENT_ROLE.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Secrets in browser env, wildcard origins. |
| Snapshot/restore | [operations/SNAPSHOT_RESTORE.md](operations/SNAPSHOT_RESTORE.md) | [features/ASSETS.md](features/ASSETS.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Destructive restore, storage mismatch. |
| Backup/rollback | [operations/BACKUP_ROLLBACK.md](operations/BACKUP_ROLLBACK.md) | [operations/SNAPSHOT_RESTORE.md](operations/SNAPSHOT_RESTORE.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Untested restore, database/storage split. |
| Release/git workflow | [operations/RELEASE_GIT_WORKFLOW.md](operations/RELEASE_GIT_WORKFLOW.md) | [REPORTING_FORMATS.md](REPORTING_FORMATS.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Dirty worktree, env/runtime files, untested tags. |
| Smoke/troubleshooting | [operations/SMOKE_VALIDATION_TROUBLESHOOTING.md](operations/SMOKE_VALIDATION_TROUBLESHOOTING.md) | [launch/DEMO_FLOW_RUNBOOK.md](launch/DEMO_FLOW_RUNBOOK.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Claiming success without verification. |
| Umami analytics operations | [operations/UMAMI_ANALYTICS_OPERATIONS.md](operations/UMAMI_ANALYTICS_OPERATIONS.md) | [features/REPORTS_ANALYTICS_HEALTH.md](features/REPORTS_ANALYTICS_HEALTH.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Provider secrets, PII, fake analytics. |
| Public runtime deployment | [operations/PUBLIC_RUNTIME_DEPLOYMENT.md](operations/PUBLIC_RUNTIME_DEPLOYMENT.md) | [features/PUBLIC_RUNTIME.md](features/PUBLIC_RUNTIME.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | CORS/CSP/frame, cms5/Reference boundary. |

## I Need To Prepare Demo / First Client / Launch

| Task | Read first | Then read | Validation | Danger zone |
| --- | --- | --- | --- | --- |
| First-client readiness | [launch/FIRST_CLIENT_READINESS.md](launch/FIRST_CLIENT_READINESS.md) | [launch/LAUNCH_BLOCKERS_VS_POLISH.md](launch/LAUNCH_BLOCKERS_VS_POLISH.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Unsafe manual process, unverified security. |
| Base deployment readiness | [launch/NUVIO_BASE_DEPLOYMENT_READINESS.md](launch/NUVIO_BASE_DEPLOYMENT_READINESS.md) | [operations/DEPLOYMENT_COOLIFY.md](operations/DEPLOYMENT_COOLIFY.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Assuming deploy is done. |
| Demo flow | [launch/DEMO_FLOW_RUNBOOK.md](launch/DEMO_FLOW_RUNBOOK.md) | Feature cards for each flow | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Fake demo data, skipped smoke. |
| Agent handoff | [launch/AGENT_HANDOFF_CHECKLIST.md](launch/AGENT_HANDOFF_CHECKLIST.md) | [REPORTING_FORMATS.md](REPORTING_FORMATS.md) | Git status and scoped verification | Missing unknowns or skipped checks. |
| Blocker vs polish | [launch/LAUNCH_BLOCKERS_VS_POLISH.md](launch/LAUNCH_BLOCKERS_VS_POLISH.md) | [launch/READINESS_DECISION_MATRIX.md](launch/READINESS_DECISION_MATRIX.md) | Source/status check | Turning polish into blocker. |

## I Need To Decide Whether Something Is Blocker Or Polish

| Task | Read first | Then read | Validation | Danger zone |
| --- | --- | --- | --- | --- |
| Classify new task | [launch/READINESS_DECISION_MATRIX.md](launch/READINESS_DECISION_MATRIX.md) | [CURRENT_OPERATING_STATE.md](CURRENT_OPERATING_STATE.md) | Report classification | Reviving old backlog. |
| Check current blockers | [launch/LAUNCH_BLOCKERS_VS_POLISH.md](launch/LAUNCH_BLOCKERS_VS_POLISH.md) | [launch/FIRST_CLIENT_READINESS.md](launch/FIRST_CLIENT_READINESS.md) | Source/status check | Inventing status. |
| Old backlog appears | [DANGER_ZONES.md](DANGER_ZONES.md) | [TASK_ROUTER.md](TASK_ROUTER.md) | Ask/report before action | Treating deferred as active. |

## I Need To Prepare A Codex Prompt

| Task | Read first | Then read | Validation | Danger zone |
| --- | --- | --- | --- | --- |
| Implementation prompt | [TASK_ROUTER.md](TASK_ROUTER.md) | Relevant feature/ops/launch card | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Missing forbidden scope. |
| Audit prompt | [REPORTING_FORMATS.md](REPORTING_FORMATS.md) | [DANGER_ZONES.md](DANGER_ZONES.md) | Status/diff only unless asked | Accidentally implementing. |
| Deployment prompt | [operations/DEPLOYMENT_COOLIFY.md](operations/DEPLOYMENT_COOLIFY.md) | [operations/ENV_SECRETS.md](operations/ENV_SECRETS.md) | Deployment smoke plan | Secrets/destructive actions. |
| Handoff prompt | [launch/AGENT_HANDOFF_CHECKLIST.md](launch/AGENT_HANDOFF_CHECKLIST.md) | [REPORTING_FORMATS.md](REPORTING_FORMATS.md) | Git status and file list | Claiming unverified success. |

## I Need To Touch High-Risk Areas

| Area | Read first | Then read | Validation | Danger zone |
| --- | --- | --- | --- | --- |
| Auth/client-role | [DANGER_ZONES.md](DANGER_ZONES.md) | [features/SECURITY_CLIENT_ROLE.md](features/SECURITY_CLIENT_ROLE.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | UI-only security. |
| Public endpoints | [DANGER_ZONES.md](DANGER_ZONES.md) | Feature card for endpoint | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | PII/tokens/errors. |
| Booking | [DANGER_ZONES.md](DANGER_ZONES.md) | [features/BOOKING.md](features/BOOKING.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Slot/status/email logic. |
| Newsletter tokens/send | [DANGER_ZONES.md](DANGER_ZONES.md) | [features/NEWSLETTER.md](features/NEWSLETTER.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Token leak/send behavior. |
| Reports/analytics | [DANGER_ZONES.md](DANGER_ZONES.md) | [features/REPORTS_ANALYTICS_HEALTH.md](features/REPORTS_ANALYTICS_HEALTH.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | PII/fake analytics. |
| Env/secrets | [DANGER_ZONES.md](DANGER_ZONES.md) | [operations/ENV_SECRETS.md](operations/ENV_SECRETS.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Secrets in browser/env/docs. |
| Restore/migrations | [DANGER_ZONES.md](DANGER_ZONES.md) | [operations/SNAPSHOT_RESTORE.md](operations/SNAPSHOT_RESTORE.md) | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Destructive data loss. |

## Output Rule

Every final answer should use [REPORTING_FORMATS.md](REPORTING_FORMATS.md) and explicitly state what was not changed.


