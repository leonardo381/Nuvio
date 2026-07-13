# Reports Umami Health Task Pack

## Purpose
Use this task pack for Reports dashboard validation, Umami analytics setup, health endpoint checks, traffic confidence, operational DTO summaries, or analytics provider readiness.

## Task classification
- readiness
- regression
- operations
- launch-critical

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Reports, Analytics, and Health](../features/REPORTS_ANALYTICS_HEALTH.md)
- [Umami Analytics Operations](../operations/UMAMI_ANALYTICS_OPERATIONS.md)
- [Environment and Secrets](../operations/ENV_SECRETS.md)
- [Deployment and Coolify](../operations/DEPLOYMENT_COOLIFY.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)

## Optional source docs
- [Deployment Env Matrix](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md)
- [Coolify Base Deployment Plan](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md)
- Obsidian Reports and Security Hardening docs.
- Current Reports backend/UI source and tests.
- Current Umami/provider configuration if supplied.

## Preconditions
- Target website and Reports period are known.
- Analytics provider configured/unconfigured state is known or marked unknown.
- Whether Umami scope means public tracking/events, Reports provider API, or both is known.
- Provider credentials are available only in server-side secret store.
- Task is audit/regression unless implementation is explicitly requested.
- Live health URL is known if health smoke is requested.

## Source-of-truth rules
1. Current Reports source, DTOs, provider config, and git status win.
2. Deployment env docs win for variable names.
3. Nuvio OS Reports/Umami/Env docs define guardrails.
4. Obsidian docs are context only.
5. Provider/live health state must be observed before claiming success.

## Allowed work
- Audit Reports/Umami readiness.
- Validate configured and unconfigured analytics states when safe.
- Check health endpoint behavior if deployment smoke is in scope.
- Improve UI copy/status only if scoped and backed by DTO/provider state.

## Forbidden work
- Do not expose Umami API keys, usernames, or passwords.
- Do not invent traffic, rankings, or analytics insights.
- Do not show fake metrics as real.
- Do not add unsupported report fields or snapshots.
- Do not put provider secrets in `VITE_*`.

## Danger zones
- PII in analytics events.
- Provider secret exposure.
- Fake analytics confidence.
- Reports DTO fields missing but claimed.
- Health check confused with full app readiness.

## Execution outline
1. Confirm target website, environment, and provider state.
2. Read Reports, Umami, Env, Deployment, and Validation docs.
3. Inspect current Reports DTO/source if implementation is requested.
4. Check Reports dashboard behavior and analytics setup state.
5. Check health endpoint only if environment is available.
6. Report unavailable provider/live checks explicitly.

## Validation checklist
### Doc validation
- Data sources and provider state are described.
- Unsupported metrics are marked deferred or unknown.
- No provider secrets appear in docs/report.

### Code/build/test validation, if future implementation applies
- If backend Reports/provider code changes, run relevant backend Reports tests.
- If UI changes, run UI build/check.
- If public tracking changes, validate target public runtime.
- If deployment health is in scope, check `/api/health` only in approved environment.

### Manual smoke validation
- Reports dashboard loads.
- Traffic tab shows configured or clear unavailable/setup state.
- No provider credentials in browser responses/env.
- Health endpoint returns expected result if deployed.
- Public tracking/events are verified only if tracking is implemented and enabled.
- Client-role assigned user sees scoped Reports data if applicable.

### User confirmation needed
- Target website/period.
- Umami provider configured status.
- Safe access to server-side logs/config.
- Live environment URL.
- Whether tracking is enabled on public runtime.

## Expected report format
- Files read.
- Files changed.
- Reports/Umami/health checks performed.
- Data-source limitations.
- Unknowns and risks.
- Validation run/skipped.
- Confirmation no fake metrics/secrets.
- Next recommended step.

## Stop conditions
- Provider credentials are needed but unavailable.
- Requested metric is not backed by current DTO/provider data.
- Testing live analytics would expose PII.
- Health/deploy state is required but no environment is specified.

