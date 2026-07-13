# Environment and Secrets

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents audit and document env/secrets boundaries for backend, backoffice UI, public runtime, Resend, Umami, Google Places, CORS, CSP, and preview frame settings.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Secret and browser-env safety rules. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes env/secrets tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Env-related smoke checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Required reporting. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Secrets and public endpoint risks. |
| 2 | Deployment Env Matrix | [../../NUVIO_DEPLOYMENT_ENV_MATRIX.md](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md) | Full env reference. |
| 2 | Deployment Quick Guide | [../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md) | Minimum required env. |
| 2 | Coolify Base Plan | [../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md) | Build-time vs runtime mapping. |
| 2 | Emails Feature Card | [../features/EMAILS_TEMPLATES.md](../features/EMAILS_TEMPLATES.md) | Resend and public URL email risks. |
| 2 | Reports Feature Card | [../features/REPORTS_ANALYTICS_HEALTH.md](../features/REPORTS_ANALYTICS_HEALTH.md) | Umami and analytics secret risks. |

## When to use this card

- Env audit or docs update.
- Secret exposure review.
- CORS/CSP/frame policy setup.
- Resend or Umami configuration review.
- Public runtime build/runtime env mapping.

## Current operating model

- Backend minimum env includes `PB_URL`, `NUVIO_PUBLIC_BASE_URL`, `NUVIO_CORS_ALLOWED_ORIGINS`, and `NUVIO_CMS_PREVIEW_FRAME_SRC`.
- Backoffice UI browser env includes `VITE_PB_BACKEND_URL` and `VITE_PUBLIC_SITE_BASE_URL`.
- Public runtime uses browser-exposed `VITE_*` values and server-only fallbacks such as `NUVIO_BACKEND_URL`.
- `VITE_*` values are browser-exposed and must never contain secrets.
- Resend and Umami credentials are server-side secrets.
- `NUVIO_ALLOW_DEV_RESET`, `PB_SUPERUSER_EMAIL`, and `PB_SUPERUSER_PASSWORD` are dev/QA tooling variables, not normal production service env.
- Exact production origins are required; wildcard/localhost production CORS is not acceptable.

## Agent permissions

### Agents may

- Audit env docs and examples for placeholder safety.
- Update docs when env guidance changes.
- Flag missing or ambiguous env variables.
- Verify browser-exposed vs server-only boundaries.
- Prepare provider setup checklists.

### Agents must not

- Do not create or commit real `.env` files.
- Do not print real secret values.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not run destructive restore/migration/backup commands unless explicitly requested.
- Do not assume local success equals deployment readiness.
- Do not change env variable names without explicit approval.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Target app: backend, backoffice UI, public runtime, or provider setup.
- Public URL.
- Admin/backoffice URL.
- Backend/API URL.
- Whether API/admin are same-origin or split.
- CORS allowed origins.
- CMS preview frame source.
- Public runtime preview parent origin.
- Enabled providers: Resend, Umami, Google Places.
- Deployment target secret-store behavior.

## Standard workflow

1. Read the env matrix and target app docs.
2. Classify each variable as server-only, browser-exposed, dev/QA-only, or optional provider.
3. Confirm placeholders only in committed examples/docs.
4. Confirm `VITE_*` contains no secrets.
5. Confirm public base URL values match email, SEO, robots, sitemap, and preview needs.
6. Confirm CORS/CSP/frame origins are exact.
7. Report missing values and do not invent real domains or secrets.

## Validation checklist

- [ ] No real `.env` files added or modified.
- [ ] No real secrets in docs, code, logs, screenshots, or examples.
- [ ] No secrets in `VITE_*` or other browser-exposed variables.
- [ ] Backend env includes exact CORS and preview frame origins.
- [ ] Public runtime env includes exact preview parent origin.
- [ ] Resend configured only if email is enabled.
- [ ] Umami configured only if analytics is enabled.
- [ ] `NUVIO_ALLOW_DEV_RESET` absent from running production-like service env.

## Common failure modes

- Treating `VITE_*` as private because it is in an env file.
- Leaving local URLs in production-like configuration.
- Mixing public runtime URL with backend/API URL in emails.
- Adding provider secrets to docs for convenience.
- Enabling dev reset flags outside a controlled one-off local/QA operation.
- Forgetting build-time `VITE_*` requirements in deployment providers.

## Reporting format

- Env area audited.
- Source docs read.
- Variables reviewed.
- Server-only vs browser-exposed classification.
- Missing inputs or unknowns.
- Secrets exposure result.
- CORS/CSP/frame result.
- Explicit confirmation no real secrets were added.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Deployment and Coolify](DEPLOYMENT_COOLIFY.md)
- [Public Runtime Deployment](PUBLIC_RUNTIME_DEPLOYMENT.md)
- [Emails Feature Card](../features/EMAILS_TEMPLATES.md)
- [Security Feature Card](../features/SECURITY_CLIENT_ROLE.md)
