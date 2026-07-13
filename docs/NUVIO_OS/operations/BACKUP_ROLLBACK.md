# Backup and Rollback

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents reason about backup readiness, rollback safety, and restore proof before first demo/client use without inventing backup commands or running destructive operations.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Global safety and source order. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes backup/rollback tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Deployment and restore smoke checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Deployment/rollback report structure. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Destructive operation stop conditions. |
| 2 | Coolify Base Plan | [../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md) | Backup Strategy V1 and smoke checklist. |
| 2 | Instance Bootstrap Checklist | [../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md) | Initial backup and restore rehearsal requirements. |
| 2 | Obsidian Snapshot and Restore | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Snapshot and Restore.md` | Snapshot vs real backup distinction. |
| 2 | Obsidian Git Workflow | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Git Workflow.md` | Release rollback context. |

## When to use this card

- Backup plan audit.
- Initial backup checklist before first client/demo.
- Rollback plan creation.
- Restore rehearsal planning.
- Incident or emergency rollback documentation.

## Current operating model

- Repo docs recommend backing up the entire `/app/pb_data` volume.
- Storage files under `/app/pb_data/storage` must remain consistent with database references.
- CMS snapshots should be versioned and stored outside git if they contain runtime files.
- Deployment metadata belongs in a private deployment tracker or secrets manager, not public docs.
- A restore rehearsal is recommended before relying on backups for production.
- Exact backup provider, command, schedule, and retention for a real deployment: Unknown / needs confirmation.

## Agent permissions

### Agents may

- Draft backup/rollback checklists.
- Audit whether backup target, retention, and restore rehearsal are documented.
- Identify missing rollback inputs.
- Report restore proof gaps.
- Update docs when explicitly requested.

### Agents must not

- Do not run backup, restore, rollback, reset, or delete commands unless explicitly requested.
- Do not invent backup commands or provider-specific steps.
- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Instance slug and environment.
- Backup target/path/provider.
- Backup frequency and retention.
- `pb_data` path/volume.
- Storage path if separate.
- CMS snapshot location.
- Code version/tag to roll back to.
- Deployment target rollback mechanism.
- Restore rehearsal target and owner.
- RPO/RTO expectations if known.

## Standard workflow

1. Confirm whether the task is planning, audit, or execution.
2. Identify data that must be backed up: database, storage, snapshots, deployment metadata.
3. Confirm backup destination and retention.
4. Confirm restore rehearsal plan.
5. Confirm rollback code version/tag and deployment mechanism.
6. Confirm smoke checks after restore/rollback.
7. If execution is requested, stop for explicit approval before destructive actions.

## Validation checklist

- [ ] Backup target exists or is documented as missing.
- [ ] `pb_data` and storage are included.
- [ ] CMS snapshots are stored/versioned outside runtime code.
- [ ] Deployment metadata is recorded privately.
- [ ] Restore rehearsal is planned or completed.
- [ ] Rollback version/tag is known.
- [ ] Post-rollback smoke checks are listed.
- [ ] No backup/restore command was invented.

## Common failure modes

- Treating CMS snapshots as full operational backups.
- Backing up database without storage files.
- Keeping backups on the same disposable volume as runtime data.
- Not testing restore until an emergency.
- Rolling back code without checking migrations/data compatibility.
- Losing the snapshot/version metadata needed to reproduce a known good state.

## Reporting format

- Backup/rollback objective.
- Source docs read.
- Data included/excluded.
- Backup target, frequency, and retention status.
- Restore rehearsal status.
- Rollback version/tag status.
- Unknowns and blockers.
- Explicit confirmation no destructive command was run unless requested.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Snapshot and Restore](SNAPSHOT_RESTORE.md)
- [Deployment and Coolify](DEPLOYMENT_COOLIFY.md)
- [Release and Git Workflow](RELEASE_GIT_WORKFLOW.md)
- [Security Feature Card](../features/SECURITY_CLIENT_ROLE.md)
