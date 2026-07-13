# Docker and Compose

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents work safely with local Docker/Compose documentation and readiness checks without changing runtime architecture or assuming Compose is production.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Global rules. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes Docker/deployment tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Compose and deployment validation expectations. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Required report structure. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Volume/restore/env safety. |
| 2 | Local Compose README | [../../../deploy/README.md](../../../deploy/README.md) | Local Compose workflow. |
| 2 | Compose Example | [../../../deploy/docker-compose.base.example.yml](../../../deploy/docker-compose.base.example.yml) | Confirmed service and volume definition. |
| 2 | Backend Env Example | [../../../deploy/env.backend.local.example](../../../deploy/env.backend.local.example) | Local backend env placeholders. |
| 2 | Public Env Example | [../../../deploy/env.public.local.example](../../../deploy/env.public.local.example) | Local public runtime env placeholders. |
| 2 | Obsidian Docker and Compose | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Docker and Compose.md` | Human local operations context. |

## When to use this card

- Local base instance Compose changes or audits.
- Dockerfile readiness review.
- Volume/env/port mapping checks.
- Compose-to-Coolify comparison.
- Local smoke validation planning.

## Current operating model

- Local Compose file is `deploy/docker-compose.base.example.yml`.
- Local services are `nuvio-backoffice` and `nuvio-public`.
- Backoffice maps container port `8090` and mounts named volume `nuvio_base_pb_data` at `/app/pb_data`.
- Public runtime maps container port `3000` and depends on backoffice health.
- Compose uses build args for browser-exposed `VITE_*` values.
- Public runtime server-side calls inside Docker use `http://nuvio-backoffice:8090`.
- Local URLs are `http://localhost:8090`, `http://localhost:8090/_/`, and `http://localhost:3000`.
- Compose is local/staging-oriented example infrastructure, not proof of production readiness.

## Agent permissions

### Agents may

- Audit Compose docs and service mappings.
- Update Compose documentation when explicitly requested.
- Validate Compose config if requested by the phase.
- Create smoke checklists.
- Mark runtime/container behavior as unknown if not confirmed by current files.

### Agents must not

- Do not run containers, build images, or alter Dockerfiles unless explicitly requested.
- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not run destructive restore/migration/backup commands unless explicitly requested.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Whether task is docs-only, audit, or implementation.
- Target compose file.
- Desired local ports.
- Backend/public runtime paths.
- Env file names and placeholders.
- Whether Docker is available and allowed to run.
- Snapshot restore expectations.
- Any provider-specific deployment constraints.

## Standard workflow

1. Inspect `deploy/README.md` and Compose file.
2. Confirm service names, ports, build contexts, env files, and volumes.
3. Confirm `VITE_*` build args are placeholders and contain no secrets.
4. Confirm server-only env values are not browser-exposed.
5. Confirm `pb_data` volume is isolated.
6. If execution is requested, run only documented safe validation commands.
7. Report whether Compose config/local smoke was run or skipped.

## Validation checklist

- [ ] Compose file parses if validation is requested.
- [ ] Backoffice service exposes `8090`.
- [ ] Public service exposes `3000`.
- [ ] Backoffice volume maps to `/app/pb_data`.
- [ ] Healthcheck uses `/api/health`.
- [ ] Public runtime depends on healthy backoffice.
- [ ] Local `.env` examples contain placeholders only.
- [ ] Snapshot restore is documented as a stopped-service one-off step.

## Common failure modes

- Editing Compose as if it were Coolify production config.
- Putting secrets in `VITE_*` build args.
- Using localhost inside one container to reach another container.
- Mounting the wrong host path or sharing a volume between instances.
- Running restore while the backend is still writing to `pb_data`.
- Treating `docker compose up` success as full application smoke success.

## Reporting format

- Compose files inspected.
- Services, ports, volumes, env files, and build args confirmed.
- Any commands run and results.
- Any commands skipped and why.
- Any unknown Docker/provider assumptions.
- Explicit no-secrets/no-real-env confirmation.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Deployment and Coolify](DEPLOYMENT_COOLIFY.md)
- [Env and Secrets](ENV_SECRETS.md)
- [Snapshot and Restore](SNAPSHOT_RESTORE.md)
- [Public Runtime Deployment](PUBLIC_RUNTIME_DEPLOYMENT.md)
