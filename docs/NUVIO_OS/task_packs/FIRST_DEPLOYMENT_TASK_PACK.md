# First Deployment Task Pack

## Purpose
Use this task pack for the first production-like Nuvio Base deployment, Coolify readiness, or deployment planning that could affect domains, volumes, env, restore, public runtime, smoke tests, or backup proof.

## Task classification
- launch-critical
- operations
- readiness

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Deployment and Coolify](../operations/DEPLOYMENT_COOLIFY.md)
- [Docker and Compose](../operations/DOCKER_COMPOSE.md)
- [Instance Bootstrap](../operations/INSTANCE_BOOTSTRAP.md)
- [Environment and Secrets](../operations/ENV_SECRETS.md)
- [Snapshot and Restore](../operations/SNAPSHOT_RESTORE.md)
- [Backup and Rollback](../operations/BACKUP_ROLLBACK.md)
- [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md)
- [Nuvio Base Deployment Readiness](../launch/NUVIO_BASE_DEPLOYMENT_READINESS.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)

## Optional source docs
- [Deployment Quick Guide](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md)
- [Deployment Env Matrix](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md)
- [Instance Bootstrap Checklist](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md)
- [Coolify Base Deployment Plan](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md)
- [Local Compose README](../../../deploy/README.md)
- [Local Compose example](../../../deploy/docker-compose.base.example.yml)
- Obsidian Coolify Plan and Deployment Quick Guide.

## Preconditions
- Deployment target is known: local, staging, Coolify, production-like, or audit-only.
- Domains, DNS, TLS, provider account, and secret store status are known or marked unknown.
- Persistent volume path and backup target are known or marked unknown.
- Snapshot restore mode is known: not running, dry-run, or explicitly approved one-off.
- User has explicitly approved any live deploy, restore, or destructive action.

## Source-of-truth rules
1. Current Dockerfiles, compose files, source code, and git status win.
2. Repo deployment docs win for planned deployment model.
3. Nuvio OS docs classify risk and validation.
4. Obsidian docs are human context and may lag.
5. Provider dashboard/live state must be verified directly before execution.

## Allowed work
- Audit deployment readiness and missing decisions.
- Create or update deployment docs when scoped.
- Run safe config/status checks only if the future task allows commands.
- Prepare smoke checklist and handoff report.
- Mark live provider state as unknown when it is not directly verified.

## Forbidden work
- Do not deploy, start services, run Docker, run migrations, or restore data unless explicitly requested.
- Do not put real secrets in docs or git.
- Do not use wildcard production CORS or localhost production origins.
- Do not share `pb_data`, storage, or backup volumes between instances.
- Do not automate snapshot restore on container startup.

## Danger zones
- Destructive restore against the wrong `pb_data`.
- Storage/database mismatch after snapshot restore.
- Secrets in `VITE_*` or public env.
- Coolify/provider assumptions treated as verified facts.
- Skipping backup proof before handoff.

## Execution outline
1. Confirm task mode and target environment.
2. Read required deployment, env, restore, backup, and public runtime docs.
3. List known and unknown deployment decisions.
4. Check whether the task is planning-only or execution-approved.
5. If planning, produce a deploy checklist and stop before live actions.
6. If execution is approved later, use documented commands and report all outputs.
7. End with smoke, backup, and rollback readiness status.

## Validation checklist
### Doc validation
- All deployment docs referenced by the task exist.
- Unknown provider/live status is explicitly marked.
- Any docs changed remain within allowed scope.

### Code/build/test validation, if future implementation applies
- If execution is approved later, validate backend health, public runtime, preview, assets, contact/newsletter/booking if enabled, Reports, CORS/CSP, and backup.
- Run only documented commands for the target environment.
- Report unavailable commands or provider limitations honestly.

### Manual smoke validation
- Backend `/api/health`.
- Backoffice login.
- CMS and preview iframe.
- Public runtime.
- Assets.
- Public flows if enabled.
- Reports and backup proof.

### User confirmation needed
- Live deploy approval.
- Exact domains and DNS status.
- Secret values in private secret store.
- Snapshot name and restore mechanism.
- Backup target and retention.

## Expected report format
- Files read.
- Files changed.
- Deployment target/context.
- Env/build/runtime decisions without secret values.
- What was verified vs unknown.
- Risks and blockers.
- Confirmation no live deploy/restore occurred unless explicitly approved.
- Next recommended deployment step.

## Stop conditions
- User asks to deploy/restore without explicit target and approval.
- Real secrets would need to be printed or committed.
- Domain/provider/volume state is unknown but required for action.
- Current git/source status conflicts with deployment docs.
- Any destructive action would be needed without approval.
