# Demo Flow Runbook

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Launch Layer](README.md)

## Purpose

Define a practical demo narrative for Nuvio using the five critical demo flows.

This runbook is for agents preparing, rehearsing, or reporting a demo. It should not create new product scope or hide unverified behavior.

## Read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines the five critical demo flows. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes feature and launch tasks. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Public endpoint, booking, email, and analytics risks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Demo smoke checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Demo/validation reporting. |
| 2 | Smoke Validation | [../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md) | Detailed smoke checklist. |
| 2 | First Client Readiness | [FIRST_CLIENT_READINESS.md](FIRST_CLIENT_READINESS.md) | Must-have vs acceptable imperfection. |

## Demo narrative

Nuvio helps small businesses stop losing digital opportunities by connecting a professional website to the operations that happen after visitor interest appears: content, contacts, WhatsApp, booking, newsletter, SEO, and reports.

Do not promise guaranteed lead volume. Show capture, management, and visibility infrastructure.

## Optional landing path

If the Nuvio own landing is in scope, the demo can start with:

1. Open Nuvio website / landing.
2. Show the core positioning.
3. Click `Request a website review` or equivalent CTA.
4. Continue into contact/lead capture flow.

Status of published Nuvio official website: Unknown / needs confirmation.

## Flow 1: Website setup / website settings

| Item | Detail |
| --- | --- |
| Required data | Website record, identity settings, public URL, admin URL, SEO/social defaults. |
| Expected path | Open backoffice, select website, configure identity/settings/SEO, save, confirm values persist. |
| Validation | Settings save/load works; hidden/unknown settings keys are preserved; public URLs are correct. |
| Likely failure points | Wrong website selected, hidden settings overwritten, preview URL mismatch, public base URL wrong. |
| Feature docs | [Website Settings and SEO](../features/WEBSITE_SETTINGS_SEO.md), [CMS](../features/CMS.md) |
| Operations docs | [Env and Secrets](../operations/ENV_SECRETS.md), [Instance Bootstrap](../operations/INSTANCE_BOOTSTRAP.md) |

Checklist:

- [ ] Website identity visible.
- [ ] SEO title/description/social image configured.
- [ ] Settings save and reload.
- [ ] Public URL values match deployment target.
- [ ] No hidden/admin-only settings exposed to client-role.

## Flow 2: CMS + SEO + public rendering

| Item | Detail |
| --- | --- |
| Required data | CMS snapshot or demo pages, blocks, assets, SEO fields, public runtime. |
| Expected path | Edit page/block, preview, open public page, inspect SEO/sitemap/robots basics. |
| Validation | Backoffice edit persists; preview iframe loads; public page renders; assets render; SEO output is sane. |
| Likely failure points | Missing storage files, broken preview frame origins, wrong public runtime env, unsafe SEO URL fields. |
| Feature docs | [CMS](../features/CMS.md), [ASSETS](../features/ASSETS.md), [Public Runtime](../features/PUBLIC_RUNTIME.md) |
| Operations docs | [Snapshot and Restore](../operations/SNAPSHOT_RESTORE.md), [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md) |

Checklist:

- [ ] CMS pages and blocks load.
- [ ] Edit/save one safe content field.
- [ ] Preview iframe loads and refreshes.
- [ ] Public homepage and one inner page render.
- [ ] `sitemap.xml` and `robots.txt` return expected content if implemented.
- [ ] Assets render after restore.

## Flow 3: Leads / Contact / WhatsApp

| Item | Detail |
| --- | --- |
| Required data | Contact form settings, WhatsApp settings if enabled, public runtime route, Leads access. |
| Expected path | Submit contact form, open Leads, inspect origin/context/status, test WhatsApp CTA/tracking if enabled. |
| Validation | Lead appears, attribution/context useful, no PII/token leakage, client-role access scoped. |
| Likely failure points | Contact source/page missing, `N/A` placeholder display, public endpoint validation error, raw PII logs. |
| Feature docs | [Leads Contact WhatsApp](../features/LEADS_CONTACT_WHATSAPP.md), [Security](../features/SECURITY_CLIENT_ROLE.md) |
| Operations docs | [Smoke Validation](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md), [Env and Secrets](../operations/ENV_SECRETS.md) |

Checklist:

- [ ] Public contact form submits.
- [ ] Lead appears in backoffice.
- [ ] Context/Origin is human-readable.
- [ ] WhatsApp CTA opens immediately if enabled.
- [ ] Tracking does not block visitor flow.
- [ ] Logs do not expose visitor message, phone, email, tokens, or stack traces.

## Flow 4: Booking

| Item | Detail |
| --- | --- |
| Required data | Booking service, availability, exception state if needed, public booking route, email settings if enabled. |
| Expected path | Load services, select date/slot, submit appointment, review appointment in backoffice, reschedule/status smoke if relevant. |
| Validation | Slots are correct, appointment status follows settings, emails/`.ics` work if enabled, booking-origin lead behavior is understood. |
| Likely failure points | Slot logic regression, auto-confirm mismatch, service snapshot confusion, email link/base URL issue, `.ics` regression. |
| Feature docs | [Booking](../features/BOOKING.md), [Emails and Templates](../features/EMAILS_TEMPLATES.md) |
| Operations docs | [Smoke Validation](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md), [Env and Secrets](../operations/ENV_SECRETS.md) |

Checklist:

- [ ] Services load publicly.
- [ ] Slots load for expected date/service.
- [ ] Appointment submits.
- [ ] Status is pending unless auto-confirm is explicitly configured.
- [ ] Backoffice appointment appears.
- [ ] Reschedule/status workflow works if included in demo.
- [ ] Visitor/business emails and `.ics` checked if enabled.

## Flow 5: Reports / Analytics / Health

| Item | Detail |
| --- | --- |
| Required data | Website selection, period, demo operational data, Umami config if analytics enabled, backend health. |
| Expected path | Open Reports, switch tabs, inspect overview, traffic, leads, booking, newsletter, SEO, history, then check health. |
| Validation | Reports load without crashing; analytics shows configured data or clear unavailable state; no provider secrets exposed; `/api/health` healthy. |
| Likely failure points | Fake zero analytics, provider secret exposure, empty-state confusion, broken period/website filter, PII in report summaries. |
| Feature docs | [Reports Analytics Health](../features/REPORTS_ANALYTICS_HEALTH.md) |
| Operations docs | [Umami Analytics Operations](../operations/UMAMI_ANALYTICS_OPERATIONS.md), [Smoke Validation](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md) |

Checklist:

- [ ] Reports overview loads.
- [ ] All tabs open.
- [ ] Website and period controls behave.
- [ ] Traffic tab is honest if Umami is not configured.
- [ ] Umami smoke performed if analytics is configured.
- [ ] `/api/health` returns healthy.
- [ ] No provider credentials or PII visible in browser/logs.

## Demo data rules

- Demo data may be manually prepared.
- Demo data must not include real customer PII unless explicitly approved and protected.
- Operational QA seed should not be used against production data.
- CMS snapshot should bootstrap content, not operational activity.
- If data is missing, say so; do not fake success.

## Reporting requirements

Agents must report:

- Demo environment and URLs used.
- Which flows were run.
- Which flows were skipped and why.
- Any failures and likely category: env, CORS/CSP, storage, auth, public endpoint, provider, data, product bug.
- Whether Umami analytics was configured or unavailable.
- Whether emails were actually sent or only reviewed.
- Next safest action before demo/client handoff.
