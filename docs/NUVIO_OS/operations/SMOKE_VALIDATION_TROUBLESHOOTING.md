# Smoke Validation and Troubleshooting

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents design and report smoke validation for launch/demo/first-client readiness without confusing product tests, deployment checks, and troubleshooting steps.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Five critical demo flows and task rules. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes validation tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Canonical smoke checklist. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Regression/deployment report structure. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Stop conditions and sensitive flows. |
| 2 | Coolify Base Plan | [../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md) | Deployment smoke checklist. |
| 2 | Instance Bootstrap Checklist | [../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md) | Full smoke table. |
| 2 | Obsidian Commands Cheat Sheet | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Commands Cheat Sheet.md` | Command categories; verify current scripts first. |

## When to use this card

- After deployment or restore.
- Before demo or first-client handoff.
- Regression validation planning.
- Troubleshooting CORS/CSP/preview/email/storage/report issues.
- Audit-only smoke checklist creation.

## Current operating model

- Nuvio OS defines five critical demo flows: website settings/setup; CMS + SEO + public rendering; Leads/Contact/WhatsApp; Booking; Reports/Analytics/Health.
- Deployment smoke includes backend health, backoffice login, CMS dashboard, public runtime, preview iframe, assets, contact/newsletter/booking flows if enabled, reports, and browser console CORS/CSP/frame checks.
- Docs-only changes do not require product builds unless links/scripts are changed.
- Exact commands depend on the touched repo and must be confirmed before running.

## Agent permissions

### Agents may

- Build smoke validation plans.
- Run read-only checks and documented validation commands when requested by the phase.
- Report skipped checks with residual risk.
- Categorize failures by likely area.
- Update troubleshooting docs if explicitly requested.

### Agents must not

- Do not run destructive restore/migration/backup commands unless explicitly requested.
- Do not change code while doing audit-only validation.
- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Target environment: local, staging, Coolify, production-like.
- Public URL.
- Admin/backoffice URL.
- Backend/API URL.
- Enabled features.
- Test user role: admin, client-role assigned, client-role unassigned, public visitor.
- Whether email/Resend is enabled.
- Whether analytics/Umami is enabled.
- Whether CMS snapshot/restore occurred.
- Validation commands allowed by the phase.

## Standard workflow

1. Classify validation scope: deployment, regression, feature, security, or smoke-only.
2. Read feature cards for touched flows.
3. Select checks from the Validation Matrix.
4. Confirm environment and credentials are available without exposing secrets.
5. Run allowed checks only.
6. Record pass/fail/skipped status per check.
7. Categorize failures: env, CORS/CSP, storage, auth, public runtime, provider, data, or product bug.
8. Report residual risk and next safe action.

## Validation checklist

- [ ] Website settings/setup loads and saves where relevant.
- [ ] CMS pages, blocks, assets, SEO fields, and preview work.
- [ ] Public homepage and at least one inner page render.
- [ ] Sitemap and robots return expected content.
- [ ] Contact form creates lead and does not leak PII.
- [ ] WhatsApp flow works if enabled.
- [ ] Booking services, slots, appointment submit, and reschedule work if enabled.
- [ ] Newsletter subscribe/confirm/unsubscribe/campaign preview work if enabled.
- [ ] Reports load and analytics unavailable states are clear if Umami is not configured.
- [ ] `/api/health` returns healthy.
- [ ] Assets render after restore.
- [ ] Browser console has no unexpected CORS/CSP/frame/mixed-content errors.
- [ ] Emails use correct public base URLs if email is enabled.
- [ ] Initial backup exists or is explicitly missing before handoff.

## Common failure modes

- Reporting only build success instead of smoke results.
- Testing as admin only and missing client-role behavior.
- Missing preview iframe failures caused by frame origins.
- Treating analytics unavailable as a product bug without checking provider env.
- Forgetting assets/storage after snapshot restore.
- Logging tokens or PII while troubleshooting public flows.

## Reporting format

- Validation goal and environment.
- Source docs/cards used.
- Checks run, passed, failed, skipped.
- For skipped checks: why skipped and remaining risk.
- Failure categories and likely next safe action.
- Confirmation no destructive commands were run unless requested.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Deployment and Coolify](DEPLOYMENT_COOLIFY.md)
- [Public Runtime Deployment](PUBLIC_RUNTIME_DEPLOYMENT.md)
- [Reports Feature Card](../features/REPORTS_ANALYTICS_HEALTH.md)
- [Booking Feature Card](../features/BOOKING.md)
- [Leads Feature Card](../features/LEADS_CONTACT_WHATSAPP.md)
