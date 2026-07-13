# Task Packs

## Purpose

Task packs are the preferred entrypoint for recurring Nuvio agent tasks. They give agents a compact bundle of required reads, source docs, preconditions, danger zones, validation, report expectations, and stop conditions for common work.

Task packs do not replace [Core](../CORE.md), [Danger Zones](../DANGER_ZONES.md), [Validation Matrix](../VALIDATION_MATRIX.md), [Reporting Formats](../REPORTING_FORMATS.md), or current source code. Current source code and git status still win over every document.

## When To Use Task Packs

Use a task pack when the request matches one of the recurring task shapes below. Use [OS Navigation](../OS_NAVIGATION.md) when the task is feature-specific or unusual and no task pack clearly fits. Use [Task Router](../TASK_ROUTER.md) when you need to classify a new or ambiguous task.

## Task Pack Index

| Task pack | Classification | Use when |
| --- | --- | --- |
| [Landing and Umami](LANDING_UMAMI_TASK_PACK.md) | launch-critical, readiness, sales/demo | Nuvio landing, real public site work, public tracking, or Umami boundary questions. |
| [First Deployment](FIRST_DEPLOYMENT_TASK_PACK.md) | launch-critical, operations, readiness | First production-like deploy, Coolify readiness, domains, volumes, env, restore, smoke, backup. |
| [Five Flow Regression](FIVE_FLOW_REGRESSION_TASK_PACK.md) | regression, readiness, launch-critical | Broad release-readiness smoke across the five critical Nuvio flows. |
| [Client Role Security Smoke](CLIENT_ROLE_SECURITY_SMOKE_TASK_PACK.md) | security, regression, launch-critical, readiness | Client-role access, website scoping, raw PB avoidance, public endpoint safety. |
| [CMS SEO Public Rendering](CMS_SEO_PUBLIC_RENDERING_TASK_PACK.md) | regression, readiness, launch-critical | CMS editing, settings, SEO, preview, assets, public DTO rendering, sitemap/robots. |
| [Leads Contact WhatsApp E2E](LEADS_CONTACT_WHATSAPP_E2E_TASK_PACK.md) | regression, readiness, launch-critical | Public contact, WhatsApp tracking, Leads dashboard, attribution, client-role Leads scope. |
| [Booking E2E Regression](BOOKING_E2E_REGRESSION_TASK_PACK.md) | launch-critical, regression, readiness | Booking public submit, slots, status lifecycle, reschedule, emails, `.ics`, client-role Booking. |
| [Newsletter Lifecycle](NEWSLETTER_LIFECYCLE_TASK_PACK.md) | launch-critical, regression, readiness | Subscribe, confirm, unsubscribe, groups, campaigns, tokens, client-role Newsletter. |
| [Emails and Templates E2E](EMAILS_TEMPLATES_E2E_TASK_PACK.md) | launch-critical, regression, operations, readiness | Resend/provider setup, contact/booking/newsletter emails, templates, public base URLs. |
| [Reports Umami Health](REPORTS_UMAMI_HEALTH_TASK_PACK.md) | readiness, regression, operations, launch-critical | Reports dashboard, Umami analytics setup, traffic confidence, health checks. |
| [Environment and Secrets Review](ENV_SECRETS_REVIEW_TASK_PACK.md) | security, operations, launch-critical, readiness | Env vars, browser/server boundaries, CORS/CSP, provider secrets, deployment env groups. |
| [Snapshot Restore Backup](SNAPSHOT_RESTORE_BACKUP_TASK_PACK.md) | launch-critical, operations, security, readiness | CMS snapshot restore, backup/rollback readiness, storage-file coverage, restore rehearsal. |
| [Demo Flow and Data](DEMO_FLOW_AND_DATA_TASK_PACK.md) | sales/demo, readiness, regression | Demo narrative, demo data, five critical flows, sales/demo handoff. |
| [First Client Onboarding](FIRST_CLIENT_ONBOARDING_TASK_PACK.md) | launch-critical, operations, sales/demo, readiness | First accompanied client readiness, onboarding, handoff, go/no-go checks. |

## Task Packs vs OS Navigation

| Need | Use |
| --- | --- |
| The request matches a recurring task exactly or nearly exactly. | Start with the relevant task pack. |
| The request is a one-off feature/module task. | Use [OS Navigation](../OS_NAVIGATION.md) and the relevant Feature/Operations/Launch card. |
| The request is ambiguous or may revive old backlog. | Use [Task Router](../TASK_ROUTER.md) and [Readiness Decision Matrix](../launch/READINESS_DECISION_MATRIX.md). |
| The request touches auth, public endpoints, env/secrets, restore, Booking, Newsletter tokens, analytics, deployment, migrations, or client-role data. | Read [Danger Zones](../DANGER_ZONES.md) before acting, even if the task pack also links it. |

## Rules

- Task packs are routing aids, not proof of current implementation.
- Do not skip [Core](../CORE.md), [Danger Zones](../DANGER_ZONES.md), [Validation Matrix](../VALIDATION_MATRIX.md), or [Reporting Formats](../REPORTING_FORMATS.md).
- Do not let task packs override current source code, current git status, repo contracts, or explicit user constraints.
- If a task pack requires live provider/deployment/client data and it is unavailable, report `Unknown / needs confirmation`.
- If the task becomes destructive or would expose secrets/PII, stop and ask for explicit approval.
