# Deployment and Coolify

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents plan or audit Coolify/VPS-style deployment work without assuming deployment has been approved or executed.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Global source order and deployment safety. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes deployment and stop conditions. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Deployment smoke checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Deployment report format. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Destructive and secret-handling warnings. |
| 2 | Coolify Base Plan | [../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md) | Current planned Coolify mapping. |
| 2 | Deployment Quick Guide | [../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md) | Practical deployment checklist. |
| 2 | Instance Bootstrap Checklist | [../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md) | Instance bootstrap order. |
| 2 | Obsidian Coolify Plan | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Coolify Plan.md` | Human plan context. |

## When to use this card

- Coolify project/service planning.
- Domain/origin mapping.
- Deployment readiness audits.
- Pre-deploy, post-deploy, and rollback planning.
- Mapping local Compose assumptions to deployment provider assumptions.

## Current operating model

- Repo docs describe a planned first Nuvio Base deployment using two services: backoffice/backend on port `8090` and public runtime on port `3000`.
- Repo docs recommend one persistent volume mounted at `/app/pb_data` for the backoffice service.
- Repo docs prefer a simple two-domain first deployment: public runtime plus admin/backoffice/API.
- Backoffice health endpoint is `/api/health`.
- `VITE_*` values must be available at image build time where used by browser bundles.
- Snapshot restore is a controlled one-off step, not container startup behavior.
- Real Coolify account/payment/domain readiness: Unknown / needs confirmation.

## Agent permissions

### Agents may

- Audit deployment docs and compose/Coolify mappings.
- Create or update docs/runbooks when explicitly requested.
- Produce pre-deploy, deploy, post-deploy, and rollback checklists.
- Mark unknown provider/domain/secret details for operator confirmation.
- Propose smoke validation steps.

### Agents must not

- Do not deploy, restart, stop, migrate, restore, or rollback services unless explicitly requested.
- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not run destructive restore/migration/backup commands unless explicitly requested.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Instance slug.
- Public URL.
- Admin/backoffice URL.
- Backend/API URL, if separate.
- Deployment target and provider account readiness.
- Branch/tag for backend/backoffice.
- Branch/tag for public runtime.
- Volume name/path for `pb_data`.
- Backup path or backup provider.
- CMS snapshot name and website ID.
- Enabled features.
- Resend/Umami status.
- CORS/CSP/frame requirements.

## Standard workflow

1. Read current git status and target deployment docs.
2. Confirm deployment target and whether this is plan-only or execution.
3. Confirm domains, ports, volume, env, and build-time variables.
4. Confirm snapshot restore mechanism before any restore work.
5. Confirm backup target before production-like handoff.
6. Prepare deployment smoke checklist.
7. If execution is requested, ask for any missing irreversible inputs before acting.
8. Report unknowns instead of filling gaps.

## Validation checklist

- [ ] Backend `/api/health` path identified.
- [ ] Public runtime health target identified.
- [ ] Domains and origins are exact placeholders or confirmed real values.
- [ ] `VITE_*` build-time variables are separated from server-only secrets.
- [ ] `/app/pb_data` volume is isolated per instance.
- [ ] Snapshot restore is not automated on startup.
- [ ] Backup target is defined before handoff.
- [ ] Smoke checklist includes CMS, public runtime, preview, assets, contact/newsletter/booking if enabled, reports, CORS/CSP.

## Common failure modes

- Treating the Coolify plan as proof that deployment already happened.
- Forgetting that browser bundles need `VITE_*` at build time.
- Using internal service DNS before it has been tested in Coolify.
- Leaving `NUVIO_ALLOW_DEV_RESET=1` on a running service.
- Sharing `pb_data` or storage between client instances.
- Skipping backup planning until after customer data exists.

## Reporting format

- Deployment task and target environment.
- Source docs read.
- Services, domains, ports, volumes, and env/build args reviewed.
- Inputs still missing.
- Validation/smoke checklist prepared or run.
- Explicit confirmation that no secrets were printed.
- Explicit confirmation that no deploy/destructive command was run unless requested.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Public Runtime Feature Card](../features/PUBLIC_RUNTIME.md)
- [Security and Client Role Feature Card](../features/SECURITY_CLIENT_ROLE.md)
- [Env and Secrets](ENV_SECRETS.md)
- [Instance Bootstrap](INSTANCE_BOOTSTRAP.md)
- [Smoke Validation and Troubleshooting](SMOKE_VALIDATION_TROUBLESHOOTING.md)
