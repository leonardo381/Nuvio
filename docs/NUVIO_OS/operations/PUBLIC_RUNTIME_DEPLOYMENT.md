# Public Runtime Deployment

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents preserve the deployment boundary between the backoffice/backend and public runtime while handling public site URLs, build/runtime env, CORS, CSP, preview frames, and separate public website repos.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Public runtime boundary rules. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes public runtime and deployment tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Public runtime smoke checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Required reporting. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Env/CORS/public endpoint risks. |
| 2 | Public Runtime Feature Card | [../features/PUBLIC_RUNTIME.md](../features/PUBLIC_RUNTIME.md) | Public runtime decisions and repo boundaries. |
| 2 | Deployment Env Matrix | [../../NUVIO_DEPLOYMENT_ENV_MATRIX.md](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md) | Public runtime env variables. |
| 2 | Coolify Base Plan | [../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md) | Public runtime service mapping. |
| 2 | Reference Contract | `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md` | Clean template boundary. |
| 2 | Obsidian Public Runtime | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Public Runtime.md` | Human public runtime overview. |

## When to use this card

- Public runtime deploy/env audit.
- CORS/CSP/preview iframe troubleshooting.
- Mapping cms5, Reference, or real site repos to deployment.
- Sitemap/robots/SEO public output smoke.
- Public route deployment checks.

## Current operating model

- Backoffice/backend and public runtime are separate services in local Compose and planned Coolify mapping.
- cms5 is test/dev/history, not the canonical clean starter.
- Reference is the cleaner public website template/candidate for future sites.
- Real client/official public sites are separate apps/repos/instances created from the template approach.
- Public runtime build-time values include browser-exposed `VITE_*` variables.
- Public runtime server-side fallbacks include `NUVIO_BACKEND_URL` and `PB_URL` where documented.
- CMS preview requires exact backoffice/public origins on both sides.
- Public runtime deployment details differ by repo and must be confirmed before implementation.

## Agent permissions

### Agents may

- Audit public runtime deployment docs.
- Prepare env/CORS/CSP/frame smoke checklists.
- Update docs when explicitly requested.
- Validate static/public route behavior when requested.
- Mark repo-specific runtime assumptions as unknown until inspected.

### Agents must not

- Do not turn cms5 into the canonical starter.
- Do not make Reference or real sites depend on cms5 or `Srcs` at runtime.
- Do not expose server-only env or tokens to browser code.
- Do not guess backend endpoints.
- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not run destructive restore/migration/backup commands unless explicitly requested.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Target public runtime repo: cms5, Reference, Nuvio official site, or client site.
- Public site URL.
- Admin/backoffice URL.
- Backend/API URL.
- Public runtime build command and adapter.
- Browser-exposed `VITE_*` values.
- Server-only runtime env values.
- CMS preview parent/frame origins.
- Enabled public flows: contact, WhatsApp, newsletter, booking, analytics.
- Deployment target and healthcheck path.

## Standard workflow

1. Confirm target public runtime repo and allowed files.
2. Read Public Runtime feature card and env matrix.
3. Inspect repo scripts and adapter before assuming commands.
4. Confirm browser-exposed vs server-only env boundary.
5. Confirm public URL, backend URL, and preview origins.
6. Validate sitemap, robots, SEO, preview, and public flows if touched.
7. Confirm no runtime dependency on cms5 or external source libraries unless this is the cms5 repo itself.
8. Report unknowns rather than filling provider/domain gaps.

## Validation checklist

- [ ] Target public runtime repo confirmed.
- [ ] Build/runtime env separated.
- [ ] No secrets in browser env.
- [ ] Public URL and backend/API URL are correct for the deployment target.
- [ ] CMS preview iframe loads without CORS/CSP/frame errors.
- [ ] Public homepage and one inner page render.
- [ ] `sitemap.xml` and `robots.txt` return expected public URLs if implemented.
- [ ] Contact/newsletter/booking flows work if enabled.
- [ ] Analytics tracking loads only if configured and does not send PII.
- [ ] Public runtime healthcheck target is identified.

## Common failure modes

- Deploying cms5 assumptions into a Reference-derived real site.
- Putting server-only backend URL secrets into browser env.
- Forgetting `VITE_*` build-time requirements.
- Breaking preview through mismatched frame origins.
- Hardcoding localhost in production-like public runtime builds.
- Depending on files outside the site repo at runtime.

## Reporting format

- Target public runtime repo and environment.
- Source docs read.
- Env/build/runtime boundary result.
- Public/admin/backend URL mapping.
- CORS/CSP/frame result.
- Public route/SEO/sitemap/robots checks.
- Public flow checks run or skipped.
- Unknown deployment details.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Public Runtime Feature Card](../features/PUBLIC_RUNTIME.md)
- [Deployment and Coolify](DEPLOYMENT_COOLIFY.md)
- [Env and Secrets](ENV_SECRETS.md)
- [Umami Analytics Operations](UMAMI_ANALYTICS_OPERATIONS.md)
- [Smoke Validation and Troubleshooting](SMOKE_VALIDATION_TROUBLESHOOTING.md)
