# First Client Readiness

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Launch Layer](README.md)

## Purpose

Define what must be true before Nuvio is used with a first accompanied client.

This card helps agents avoid two opposite mistakes: blocking first client with non-critical polish, or moving forward with unsafe deployment/security/data practices.

## Read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines launch-critical mode and five demo flows. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes feature, ops, security, and deployment tasks. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Identifies unsafe launch paths. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Smoke and security checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Required agent handoff/report format. |
| 2 | Current Roadmap | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Current Roadmap.md` | Current phase and do-not-start items. |
| 2 | Backoffice 1.0 Status | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Backoffice 1.0 Status.md` | Product status and not-done items. |
| 2 | First deployment operations | [../operations/DEPLOYMENT_COOLIFY.md](../operations/DEPLOYMENT_COOLIFY.md) | Deployment target readiness. |
| 2 | Smoke validation | [../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md) | Critical flow checks. |

## Must be true before first accompanied client

### Product must-haves

- [ ] Website settings/setup can be configured and saved.
- [ ] CMS pages, blocks, assets, SEO, and preview work on a restored/demo site.
- [ ] Public rendering works for homepage and at least one inner page.
- [ ] Contact form creates a lead and has useful attribution/context.
- [ ] WhatsApp flow works if enabled and does not block visitor navigation.
- [ ] Booking service/slot/appointment flow works if enabled.
- [ ] Newsletter subscribe/confirm/unsubscribe lifecycle works if enabled.
- [ ] Reports load and show either real configured data or clear unavailable/setup state.
- [ ] `/api/health` works for the deployed backend.

### Security must-haves

- [ ] Assigned client-role user sees only assigned website data.
- [ ] Unassigned client-role access is denied.
- [ ] Protected flows use scoped endpoints, not raw PB writes.
- [ ] Public endpoints return clean visitor errors and do not leak internals.
- [ ] Logs do not expose tokens, API keys, visitor messages, emails, or phone numbers.
- [ ] CORS/CSP/frame origins are exact for production-like deploys.
- [ ] Secrets are server-side only; no secrets in `VITE_*` or public runtime browser code.

### Email/lifecycle must-haves

- [ ] Resend is configured only if email is enabled.
- [ ] Contact/booking notifications use expected sender/recipient if enabled.
- [ ] Newsletter confirmation/unsubscribe links use the public site URL.
- [ ] Booking visitor/business emails and `.ics` behavior are smoke-tested if enabled.
- [ ] No lifecycle token is logged or exposed.

### Operations must-haves

- [ ] Production-like deployment path is proven or clearly scoped as a supported manual setup.
- [ ] Instance has isolated env, `pb_data`, storage, domains, and backup path.
- [ ] CMS snapshot restore includes records and storage files.
- [ ] Initial backup exists before handoff.
- [ ] Restore rehearsal is completed or explicitly scheduled before higher-risk usage.
- [ ] Deployment metadata is recorded privately.

### Sales/onboarding must-haves

- [ ] Own Nuvio landing/request-review path exists or has a manual substitute.
- [ ] Demo website/data is ready enough to show the five critical flows.
- [ ] Operator can explain what Nuvio does without promising guaranteed lead volume.
- [ ] First-client onboarding can be accompanied manually.
- [ ] Known imperfections are documented before handoff.

## Acceptable imperfections for first accompanied client

| Imperfection | Why acceptable |
| --- | --- |
| Manual onboarding steps | First client can be accompanied; automation can come later. |
| Manual backup runbook | Acceptable if backup exists and restore path is understood. |
| Reports empty/unconfigured analytics states | Acceptable if honest and not fake. |
| UI polish gaps | Acceptable if core workflows are safe and understandable. |
| Limited newsletter automation | Advanced automation is deferred. |
| Single-capacity booking | Multi-capacity is deferred/post-first-client. |
| Manual demo data preparation | Fine for first demo if not unsafe. |

## Do not block first client with

- Google Places / Reviews sync.
- Data exports.
- Booking multi-capacity.
- Advanced reports history/snapshots.
- Advanced newsletter automation.
- Self-service billing/client portals.
- Perfect visual polish across every page.
- Full client onboarding automation.

## Required linked cards

| Area | Feature card | Operations card |
| --- | --- | --- |
| CMS/public rendering | [CMS](../features/CMS.md), [Public Runtime](../features/PUBLIC_RUNTIME.md) | [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md) |
| Leads/Contact/WhatsApp | [Leads](../features/LEADS_CONTACT_WHATSAPP.md) | [Smoke Validation](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md) |
| Booking | [Booking](../features/BOOKING.md) | [Emails](../features/EMAILS_TEMPLATES.md) |
| Newsletter | [Newsletter](../features/NEWSLETTER.md) | [Env and Secrets](../operations/ENV_SECRETS.md) |
| Reports/Analytics | [Reports](../features/REPORTS_ANALYTICS_HEALTH.md) | [Umami Analytics](../operations/UMAMI_ANALYTICS_OPERATIONS.md) |
| Security | [Security](../features/SECURITY_CLIENT_ROLE.md) | [Env and Secrets](../operations/ENV_SECRETS.md) |
| Deployment/backup | [Public Runtime](../features/PUBLIC_RUNTIME.md) | [Deployment](../operations/DEPLOYMENT_COOLIFY.md), [Backup](../operations/BACKUP_ROLLBACK.md) |

## Reporting requirements

Agents must report:

- Which must-haves are verified.
- Which must-haves are unverified.
- Which imperfections are acceptable.
- Any blocker classified as launch-critical.
- Commands run or intentionally not run.
- Manual smoke performed or not performed.
- Unknowns needing operator confirmation.
