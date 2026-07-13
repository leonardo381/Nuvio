# Operations Agent Cards

## Purpose

Operations Agent Cards route agents working on deployment, env/secrets, Docker/Compose, Coolify, instances, snapshot/restore, backup/rollback, release/git workflow, smoke validation, Umami operations, and public runtime deployment.

Use this layer when the task can affect deployed state, data, storage, secrets, provider config, or launch safety.

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
| Task mentions deployment, Coolify, Docker, Compose, env, secrets, domains, CORS, CSP, backups, rollback, snapshot, restore, smoke, Umami, or public runtime deploy. | Open the matching Operations Agent Card. |
| Task includes a destructive or provider-facing operation. | Read [../DANGER_ZONES.md](../DANGER_ZONES.md) first. |
| Task affects a product feature after deployment. | Also open [../features/README.md](../features/README.md). |
| Task affects first-client/demo readiness. | Also open [../launch/README.md](../launch/README.md). |

## Cards

| Operations area | Card |
| --- | --- |
| Deployment and Coolify | [DEPLOYMENT_COOLIFY.md](DEPLOYMENT_COOLIFY.md) |
| Docker and Compose | [DOCKER_COMPOSE.md](DOCKER_COMPOSE.md) |
| Instance Bootstrap | [INSTANCE_BOOTSTRAP.md](INSTANCE_BOOTSTRAP.md) |
| Env and Secrets | [ENV_SECRETS.md](ENV_SECRETS.md) |
| Snapshot and Restore | [SNAPSHOT_RESTORE.md](SNAPSHOT_RESTORE.md) |
| Backup and Rollback | [BACKUP_ROLLBACK.md](BACKUP_ROLLBACK.md) |
| Release and Git Workflow | [RELEASE_GIT_WORKFLOW.md](RELEASE_GIT_WORKFLOW.md) |
| Smoke Validation and Troubleshooting | [SMOKE_VALIDATION_TROUBLESHOOTING.md](SMOKE_VALIDATION_TROUBLESHOOTING.md) |
| Umami Analytics Operations | [UMAMI_ANALYTICS_OPERATIONS.md](UMAMI_ANALYTICS_OPERATIONS.md) |
| Public Runtime Deployment | [PUBLIC_RUNTIME_DEPLOYMENT.md](PUBLIC_RUNTIME_DEPLOYMENT.md) |

## Source Priority Rules

1. Current source code and git status override docs.
2. Nuvio OS docs define agent routing and safety expectations.
3. Repo docs under `docs/` and `deploy/` override older Obsidian notes when they conflict.
4. Obsidian notes are context and operator guidance, but may be older.
5. cms5 and Reference docs are relevant only for public runtime/deployment boundaries.
6. If sources conflict, mark `Unknown / needs confirmation` instead of guessing.

## Global Operations Warnings

- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*`, `PUBLIC_*`, browser code, screenshots, docs, or logs.
- Do not use wildcard or localhost-style CORS/CSP/frame settings in production-like deployments.
- Do not share writable `pb_data` or storage across real clients.
- Do not run destructive restore, migration, backup, seed, deploy, or rollback commands unless explicitly requested.
- Do not assume local Compose success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Launch Readiness Categories

| Category | Meaning |
| --- | --- |
| Launch-critical | Blocks first real deployment or demo if wrong. |
| Readiness | Needed before handoff or first-client confidence. |
| Polish | Improves operator confidence but does not block core launch. |
| Enhancement | Useful later, not needed for first client. |
| Deferred | Intentionally out of scope until explicitly revived. |
