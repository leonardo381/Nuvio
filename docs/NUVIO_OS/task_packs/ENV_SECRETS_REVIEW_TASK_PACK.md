# Environment and Secrets Review Task Pack

## Purpose
Use this task pack for env var review, browser/server boundary checks, CORS/CSP/preview origins, provider secrets, deployment env groups, or `.env.example` documentation.

## Task classification
- security
- operations
- launch-critical
- readiness

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Environment and Secrets](../operations/ENV_SECRETS.md)
- [Security and Client Role](../features/SECURITY_CLIENT_ROLE.md)
- [Deployment and Coolify](../operations/DEPLOYMENT_COOLIFY.md)
- [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md)
- [Nuvio Base Deployment Readiness](../launch/NUVIO_BASE_DEPLOYMENT_READINESS.md)

## Optional source docs
- [Deployment Quick Guide](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md)
- [Deployment Env Matrix](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md)
- [Instance Bootstrap Checklist](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md)
- [Coolify Base Deployment Plan](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md)
- Reference public-site env contract if public site/template env is in scope.
- Obsidian Deployment Quick Guide and Security Hardening docs.

## Preconditions
- Target app/repo is known: backend, backoffice UI, cms5, Reference, real public site, or deployment provider.
- Task is docs-only audit or implementation is explicitly requested.
- No real secrets need to be printed.
- Target environment and domains are known or marked unknown.
- Browser-exposed `VITE_*` or `PUBLIC_*` rules are understood for the target app.

## Source-of-truth rules
1. Current env-loading code, examples, deployment files, and git status win.
2. Repo deployment/env docs define intended variable names.
3. Nuvio OS Env/Security docs define guardrails.
4. Obsidian docs are context only.
5. Real provider/deployment settings must be verified outside docs.

## Allowed work
- Audit env docs and examples.
- Classify variables as server-only, browser-exposed, optional provider, or dev/QA-only.
- Update docs/examples with placeholders only when scoped.
- Recommend provider/env changes without printing real values.

## Forbidden work
- Do not create or commit real `.env` files.
- Do not paste real secrets, API keys, passwords, tokens, or private domains.
- Do not put secrets in `VITE_*` or public browser env.
- Do not use wildcard production CORS or localhost production origins.
- Do not leave `NUVIO_ALLOW_DEV_RESET` enabled outside controlled dev/QA restore flows.

## Danger zones
- Browser-exposed env leaks.
- Provider credentials in UI bundle or docs.
- Wildcard CORS/CSP/frame origins.
- Build-time vs runtime env confusion.
- Committed `.env` files.
- Dev reset safety flag in production-like env.

## Execution outline
1. Confirm target app/environment and docs-only vs implementation mode.
2. Read Env, Security, Deployment, Public Runtime, and relevant app docs.
3. Inventory variables by scope without printing real values.
4. Check `VITE_*`/`PUBLIC_*` exposure rules.
5. Check CORS/CSP/preview origin consistency.
6. Report required/optional/dev-only/unknown variables.

## Validation checklist
### Doc validation
- All reviewed vars are categorized.
- Placeholders only; no real secrets.
- Unknown env/provider facts are marked.
- No docs outside allowed scope changed unless explicitly scoped.

### Code/build/test validation, if future implementation applies
- If env code changes are approved, run target app validation and inspect browser exposure where safe.
- If deployment env changes are executed later, run deployment smoke checks.
- Do not run builds/tests in docs-only env review.

### Manual smoke validation
- Backoffice can reach backend with expected URL.
- Public runtime can reach backend with server/browser URLs.
- Preview iframe origins align.
- Provider features show configured or unavailable state.
- No browser env includes secrets.

### User confirmation needed
- Exact domains/origins.
- Enabled providers/features.
- Whether build args and runtime env are separate in provider.
- Approval before changing env code or deployment settings.

## Expected report format
- Files read.
- Files changed.
- Variables reviewed and categorized.
- Unknowns and risks.
- Validation run/skipped.
- Confirmation no real secrets included.
- Next recommended step.

## Stop conditions
- A real secret would need to be shown.
- Target environment/domains are unknown but required.
- A change would alter runtime behavior without approval.
- A deployment provider setting must be changed but approval is missing.
