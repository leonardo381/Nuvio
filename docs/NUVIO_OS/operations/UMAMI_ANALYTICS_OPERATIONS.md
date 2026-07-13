# Umami Analytics Operations

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents validate and document Umami analytics operations safely, especially provider env, traffic reporting, public tracking, conversion events, and no-PII analytics rules.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Source order and demo-flow priorities. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes analytics and reports tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Analytics/Umami smoke checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Required reporting. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | PII and secrets risks. |
| 2 | Reports Feature Card | [../features/REPORTS_ANALYTICS_HEALTH.md](../features/REPORTS_ANALYTICS_HEALTH.md) | Reports and analytics decisions. |
| 2 | Deployment Env Matrix | [../../NUVIO_DEPLOYMENT_ENV_MATRIX.md](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md) | Umami env variables and secret handling. |
| 2 | Deployment Quick Guide | [../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md) | Optional analytics provider setup. |
| 2 | Obsidian Deployment Quick Guide | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Deployment Quick Guide.md` | Human operations context. |

## When to use this card

- Umami provider setup audit.
- Reports traffic analytics troubleshooting.
- Public tracking validation.
- CTA/conversion event validation.
- No-PII analytics review.

## Current operating model

- Repo env matrix documents Umami as optional analytics provider configuration.
- Server-side Umami variables include `NUVIO_UMAMI_API_URL`, `NUVIO_UMAMI_API_KEY`, `NUVIO_UMAMI_USERNAME`, and `NUVIO_UMAMI_PASSWORD`.
- Umami credentials are secrets and must not appear in browser env, UI, docs, or logs.
- Reports should be client-friendly and not a raw provider dashboard.
- Umami is for traffic analytics; Nuvio custom/business events cover leads, bookings, newsletter, WhatsApp, and other business actions where implemented.
- Public tracking implementation details must be confirmed in the active public runtime repo before changing anything.

## Agent permissions

### Agents may

- Audit analytics env and docs.
- Prepare Umami smoke checklists.
- Validate that Reports shows configured/unavailable states honestly.
- Confirm no PII is sent to analytics.
- Update docs if explicitly requested.

### Agents must not

- Do not add new analytics scope without explicit request.
- Do not expose Umami API keys/usernames/passwords to browser code.
- Do not send names, emails, phones, messages, tokens, or other PII to analytics.
- Do not commit real `.env` files.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not run destructive restore/migration/backup commands unless explicitly requested.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Whether analytics is enabled for the instance.
- Public domain to track.
- Umami API URL.
- Umami auth mode: API key or username/password.
- Umami website/site identifier, if used by Reports/settings.
- Public runtime where tracking script/events are implemented.
- CTA/conversion event names, if already defined.
- Reports expected behavior when analytics is unavailable.

## Standard workflow

1. Confirm analytics is enabled and in scope.
2. Read env matrix and Reports feature card.
3. Confirm secrets are server-side only.
4. Confirm public tracking code exists before validating it.
5. Validate script/pageviews only in the target public runtime.
6. Validate CTA/conversion events only if implementation already exists.
7. Confirm Reports uses current DTO/provider behavior without fake insights.
8. Report unknowns instead of inventing tracking architecture.

## Validation checklist

- [ ] Umami provider env exists only server-side.
- [ ] No Umami secrets in `VITE_*`, UI, browser bundle, docs, or logs.
- [ ] Public tracking script loads if implementation exists.
- [ ] Pageviews arrive for the correct public domain if analytics is enabled.
- [ ] CTA/conversion events arrive only if already implemented.
- [ ] No PII is sent in event names, URLs, properties, or logs.
- [ ] Reports traffic tab shows useful configured data or a clear setup/unavailable state.
- [ ] Analytics unavailable state is not presented as fake zero-performance insight.

## Common failure modes

- Adding analytics events that include form field values.
- Putting Umami API credentials in browser-exposed env.
- Assuming tracking exists in every public runtime repo.
- Treating Umami traffic as lead/booking truth.
- Embedding raw provider dashboards instead of Nuvio summaries.
- Using the wrong public domain/site ID.

## Reporting format

- Analytics task and target environment.
- Source docs read.
- Env variables reviewed.
- Public tracking implementation status.
- Reports behavior status.
- PII/secrets exposure result.
- Checks run/skipped and why.
- Unknowns needing operator/provider confirmation.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Reports Feature Card](../features/REPORTS_ANALYTICS_HEALTH.md)
- [Env and Secrets](ENV_SECRETS.md)
- [Smoke Validation and Troubleshooting](SMOKE_VALIDATION_TROUBLESHOOTING.md)
- [Public Runtime Deployment](PUBLIC_RUNTIME_DEPLOYMENT.md)
