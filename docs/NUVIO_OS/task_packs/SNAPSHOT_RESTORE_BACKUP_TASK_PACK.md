# Snapshot Restore Backup Task Pack

## Purpose
Use this task pack for CMS snapshot restore, backup/rollback readiness, storage-file verification, native file field coverage, restore rehearsal, or destructive reset planning.

## Task classification
- launch-critical
- operations
- security
- readiness

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Snapshot and Restore](../operations/SNAPSHOT_RESTORE.md)
- [Backup and Rollback](../operations/BACKUP_ROLLBACK.md)
- [Assets](../features/ASSETS.md)
- [Instance Bootstrap](../operations/INSTANCE_BOOTSTRAP.md)
- [Deployment and Coolify](../operations/DEPLOYMENT_COOLIFY.md)
- [Nuvio Base Deployment Readiness](../launch/NUVIO_BASE_DEPLOYMENT_READINESS.md)

## Optional source docs
- [Instance Bootstrap Checklist](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md)
- [Coolify Base Deployment Plan](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md)
- Obsidian Snapshot and Restore, Docker and Compose, and Coolify Plan docs.
- Current snapshot/restore tool source if implementation or execution is requested.

## Preconditions
- Action mode is known: audit-only, dry-run, restore rehearsal, or approved real restore.
- Target `pb_data` path/volume is known.
- Snapshot name/source and expected website ID are known.
- Backend stopped/online requirements are known.
- Backup target exists or is marked unknown.
- User explicitly approved destructive restore/reset if applicable.

## Source-of-truth rules
1. Current snapshot/restore tool source, manifest, target filesystem, and git status win.
2. Repo deployment/restore docs define intended workflow.
3. Nuvio OS Restore/Backup/Assets docs define guardrails.
4. Obsidian docs are context only.
5. Runtime data state must be verified directly before action.

## Allowed work
- Audit snapshot/backup/restore readiness.
- Verify docs and manifest/storage requirements.
- Run only safe listing/status checks unless destructive action is explicitly approved.
- Recommend restore rehearsal steps without executing them in docs-only phases.

## Forbidden work
- Do not restore automatically on container startup.
- Do not run destructive restore/reset without explicit approval.
- Do not touch `pb_data`, storage, snapshots, or runtime data in docs-only phases.
- Do not restore records without required storage files.
- Do not leave dev reset safety flags enabled after a controlled operation.

## Danger zones
- Wrong target volume/path.
- Missing storage files for native file fields.
- Database/storage mismatch.
- Restore overwriting important content.
- Untested backup.
- Confusing CMS snapshots with full operational backups.

## Execution outline
1. Confirm action mode and target environment.
2. Read Snapshot, Backup, Assets, Instance, and Deployment docs.
3. List required inputs and unknowns.
4. For audit-only tasks, stop at readiness/reporting.
5. For future approved restore, require backup, backend state, snapshot manifest, storage verification, and post-restore smoke.
6. Report exactly what was and was not touched.

## Validation checklist
### Doc validation
- Action mode is explicitly stated.
- Required restore inputs and unknowns are listed.
- Docs changed only inside allowed scope.
- No runtime data was touched in docs-only tasks.

### Code/build/test validation, if future implementation applies
- If a future restore is approved, run only documented restore/dry-run commands and capture output.
- After approved restore, verify records plus storage files and run smoke checks.
- If backup code changes, run relevant tool tests.

### Manual smoke validation
- CMS pages/blocks/assets render after restore.
- Native file fields resolve: website logo/SEO image, assets, page SEO image, block images.
- Public runtime renders restored content.
- Backoffice CMS loads.
- Backup exists before restore and restore log is recorded.

### User confirmation needed
- Explicit approval for restore/reset.
- Target `pb_data`/volume.
- Snapshot name and expected website ID.
- Backend state requirement.
- Backup target and retention.

## Expected report format
- Files read.
- Files changed.
- Action mode.
- Snapshot/backup inputs known/unknown.
- What was verified.
- Risks and stop conditions.
- Confirmation no runtime data touched unless explicitly approved.
- Next recommended step.

## Stop conditions
- Action would be destructive and approval is missing.
- Target volume/path or snapshot identity is unclear.
- Required storage files are missing or unverified.
- No backup exists before real restore.
- The backend running/stopped requirement is unknown.
