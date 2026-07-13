# Feature Agent Cards

## Purpose

Feature Agent Cards route agents working on Nuvio product modules. Use this layer when a task touches a feature area, UI module, public endpoint, or feature-specific behavior.

Do not treat these cards as permission to change dangerous areas. They point to source docs, decisions to preserve, validation checks, and reporting requirements.

## Navigate Back

| Destination | Link |
| --- | --- |
| Nuvio OS home | [../README.md](../README.md) |
| One-jump navigation | [../OS_NAVIGATION.md](../OS_NAVIGATION.md) |
| Core | [../CORE.md](../CORE.md) |
| Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) |
| Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) |
| Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) |
| Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) |

## When To Use This Layer

| Use this layer when | Then |
| --- | --- |
| Task names a module such as CMS, Leads, Booking, Newsletter, Reports, Public Runtime, Security, Emails. | Open the matching Feature Agent Card. |
| Task touches public endpoint behavior. | Read the feature card and [../DANGER_ZONES.md](../DANGER_ZONES.md). |
| Task touches deployment/env around a feature. | Also open [../operations/README.md](../operations/README.md). |
| Task affects demo/first-client readiness. | Also open [../launch/README.md](../launch/README.md). |

## Cards

| Feature | Card |
| --- | --- |
| CMS | [CMS.md](CMS.md) |
| Assets and Images | [ASSETS.md](ASSETS.md) |
| Website Settings and SEO | [WEBSITE_SETTINGS_SEO.md](WEBSITE_SETTINGS_SEO.md) |
| Leads, Contact Form, and WhatsApp | [LEADS_CONTACT_WHATSAPP.md](LEADS_CONTACT_WHATSAPP.md) |
| Booking | [BOOKING.md](BOOKING.md) |
| Newsletter | [NEWSLETTER.md](NEWSLETTER.md) |
| Reports, Analytics, and Health | [REPORTS_ANALYTICS_HEALTH.md](REPORTS_ANALYTICS_HEALTH.md) |
| Public Runtime | [PUBLIC_RUNTIME.md](PUBLIC_RUNTIME.md) |
| Security and Client Role | [SECURITY_CLIENT_ROLE.md](SECURITY_CLIENT_ROLE.md) |
| Emails and Templates | [EMAILS_TEMPLATES.md](EMAILS_TEMPLATES.md) |

## How To Use These Cards

1. Start with [../CORE.md](../CORE.md) and [../TASK_ROUTER.md](../TASK_ROUTER.md).
2. Open the card for the feature named by the user.
3. Read the card's source docs before touching code.
4. Inspect exact source/test files before changing anything.
5. Use [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) to choose checks.
6. Use [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) for the final answer.

## Feature Layer Warnings

- Do not add raw PocketBase writes for client-role product flows.
- Do not rely on UI hiding as security.
- Do not change public endpoint contracts casually.
- Do not change Booking slot/status/email behavior during UI polish.
- Do not expose newsletter tokens or provider secrets.
- Do not invent analytics insights not backed by current DTO/provider data.
- Do not connect unrelated modules during feature polish.

## Reporting Baseline

Every feature-task report should include:

- changed files;
- whether product code, docs, tests, migrations, or env changed;
- source docs read;
- decisions preserved;
- validation commands run or intentionally not run;
- manual smoke checklist;
- remaining unknowns or deferred work.
