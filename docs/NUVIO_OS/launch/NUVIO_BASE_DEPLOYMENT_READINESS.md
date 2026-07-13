# Nuvio Base Deployment Readiness

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Launch Layer](README.md)

## Purpose

Define readiness for the first production-like demo/staging Nuvio Base instance.

This document is for agents preparing or auditing deployment readiness. It does not authorize deployment, restore, migration, or destructive commands.

## Read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Current deployment/readiness mode. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Deployment and env routing. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Restore, secrets, CORS, and deployment risks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Deployment smoke checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Deployment report format. |
| 2 | Deployment Coolify | [../operations/DEPLOYMENT_COOLIFY.md](../operations/DEPLOYMENT_COOLIFY.md) | Coolify/base service mapping. |
| 2 | Docker Compose | [../operations/DOCKER_COMPOSE.md](../operations/DOCKER_COMPOSE.md) | Local Compose baseline. |
| 2 | Instance Bootstrap | [../operations/INSTANCE_BOOTSTRAP.md](../operations/INSTANCE_BOOTSTRAP.md) | Instance decisions. |
| 2 | Env and Secrets | [../operations/ENV_SECRETS.md](../operations/ENV_SECRETS.md) | Env boundaries. |
| 2 | Backup and Rollback | [../operations/BACKUP_ROLLBACK.md](../operations/BACKUP_ROLLBACK.md) | Backup proof. |
| 2 | Public Runtime Deployment | [../operations/PUBLIC_RUNTIME_DEPLOYMENT.md](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md) | Public runtime boundary. |
| 2 | Coolify Plan | [../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md) | Current repo deployment plan. |

## Current confirmed model

| Area | Confirmed from docs |
| --- | --- |
| Backoffice/backend service | Container port `8090`; health path `/api/health`. |
| Public runtime service | Container port `3000`; public health currently route `/` unless later dedicated endpoint exists. |
| Data volume | Backoffice owns persistent `/app/pb_data`. |
| Public runtime | Separate service/app from backoffice/backend. |
| Env | `VITE_*` values are browser-exposed and build-time relevant. Secrets must remain server-side. |
| Snapshot restore | Controlled one-off step; not startup automation. |
| First deployment target | Coolify/VPS-style plan exists, but real execution readiness is Unknown / needs confirmation. |

## Unknown / needs confirmation

- Coolify account/payment/provider readiness.
- Real domains/DNS/TLS status.
- Final backend/backoffice tag for deployment.
- Final public runtime tag for deployment.
- Whether public runtime for first base deploy is still `cms5` or another chosen site repo.
- Real backup target and retention.
- Restore mechanism inside Coolify or host/container maintenance flow.
- Umami provider/site ID status.
- Resend sender/domain status.

## Readiness checklist

### Instance decisions

- [ ] Instance slug chosen.
- [ ] Public URL chosen.
- [ ] Admin/backoffice URL chosen.
- [ ] Backend/API URL chosen, or same-origin admin/API model confirmed.
- [ ] `pb_data` volume/path chosen.
- [ ] Storage path chosen.
- [ ] Backup path/bucket chosen.
- [ ] CMS snapshot name chosen.
- [ ] Code version/tag chosen for backoffice/backend.
- [ ] Code version/tag chosen for public runtime.
- [ ] Enabled features listed.

### Backend/backoffice

- [ ] Backoffice image/service deploys from approved tag or `main`.
- [ ] Runtime command/path matches current Docker plan.
- [ ] `/app/pb_data` is mounted only to backoffice service.
- [ ] `/api/health` returns healthy.
- [ ] Backoffice login works.
- [ ] Migrations are applied through approved deployment/start behavior.
- [ ] `NUVIO_ALLOW_DEV_RESET` is not present on running service.

### Backoffice UI build/env

- [ ] `VITE_PB_BACKEND_URL` set as build-time/browser value.
- [ ] `VITE_PUBLIC_SITE_BASE_URL` set as build-time/browser value.
- [ ] No secrets in `VITE_*`.
- [ ] Browser API URL matches deployed backend/admin model.

### Public runtime

- [ ] Public runtime repo selected and confirmed.
- [ ] Public runtime image/service deploys from approved tag or `main`.
- [ ] `VITE_PB_URL` set as build-time/browser value.
- [ ] `VITE_NUVIO_BACKEND_URL` set as build-time/browser value.
- [ ] `VITE_PUBLIC_SITE_BASE_URL` set as build-time/browser value.
- [ ] `VITE_CMS_PREVIEW_PARENT_ORIGIN` set exactly.
- [ ] `NUVIO_BACKEND_URL` or documented server-only fallback set for server helpers.
- [ ] Public homepage returns HTTP 200.

### CORS/CSP/frame

- [ ] `NUVIO_CORS_ALLOWED_ORIGINS` uses exact admin/public origins.
- [ ] `NUVIO_CMS_PREVIEW_FRAME_SRC` allows exact public runtime origin.
- [ ] Public runtime preview parent origin allows exact admin/backoffice origin.
- [ ] Browser console has no unexpected CORS/CSP/frame errors.

### Storage/snapshot

- [ ] CMS snapshot includes records and storage files.
- [ ] Restore target `pb_data` path/volume confirmed.
- [ ] Backend stopped for restore when required.
- [ ] Native file fields render after restore.
- [ ] Assets render in backoffice and public runtime.

### Providers

- [ ] Resend configured only if email is enabled.
- [ ] Resend key server-side only.
- [ ] Newsletter/contact/booking email links use public URL.
- [ ] Umami configured only if analytics is enabled.
- [ ] Umami credentials server-side only.
- [ ] Reports show configured analytics data or honest unavailable state.

### Backup

- [ ] Initial backup exists after successful smoke.
- [ ] Storage files are included with database backup.
- [ ] Restore rehearsal is completed or scheduled before higher-risk use.
- [ ] Deployment metadata recorded privately.

## Minimum deployment smoke

- [ ] Backend `/api/health` healthy.
- [ ] Backoffice login works.
- [ ] CMS dashboard loads.
- [ ] Public runtime loads.
- [ ] CMS preview iframe works.
- [ ] Assets render.
- [ ] Public page renders.
- [ ] Contact/WhatsApp work if enabled.
- [ ] Booking works if enabled.
- [ ] Newsletter lifecycle works if enabled.
- [ ] Reports load.
- [ ] CORS/CSP/frame console is clean.
- [ ] Initial backup exists.

## Do not proceed to first-client handoff if

- Deployment cannot be reproduced from recorded tags/versions.
- `pb_data` or storage is shared with another real client.
- Secrets are in browser env or docs.
- Public endpoints leak internals or PII.
- Restore has broken assets/storage references.
- No backup exists after smoke testing.
- Client-role scoping has not been smoke-tested.

## Related docs

- [README](README.md)
- [First Client Readiness](FIRST_CLIENT_READINESS.md)
- [Demo Flow Runbook](DEMO_FLOW_RUNBOOK.md)
- [Launch Blockers vs Polish](LAUNCH_BLOCKERS_VS_POLISH.md)
- [Deployment and Coolify](../operations/DEPLOYMENT_COOLIFY.md)
- [Env and Secrets](../operations/ENV_SECRETS.md)
- [Snapshot and Restore](../operations/SNAPSHOT_RESTORE.md)
- [Backup and Rollback](../operations/BACKUP_ROLLBACK.md)
- [Public Runtime](../features/PUBLIC_RUNTIME.md)
- [Security and Client Role](../features/SECURITY_CLIENT_ROLE.md)
