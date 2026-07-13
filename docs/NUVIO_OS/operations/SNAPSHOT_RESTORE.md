# Snapshot and Restore

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents choose and document the correct snapshot/restore flow while preventing accidental destructive restores, broken storage references, or misuse of dev tooling.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Global restore safety. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes destructive restore tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Restore smoke checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Audit/deployment reporting. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Stop conditions for destructive operations. |
| 2 | Coolify Base Plan | [../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md) | Coolify restore workflow. |
| 2 | Instance Bootstrap Checklist | [../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md) | CMS snapshot bootstrap checks. |
| 2 | Obsidian Snapshot and Restore | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Snapshot and Restore.md` | Snapshot types and safety rules. |
| 2 | Assets Feature Card | [../features/ASSETS.md](../features/ASSETS.md) | Native file/storage coverage risk. |

## When to use this card

- CMS snapshot create/restore planning.
- Full QA snapshot planning.
- Operational QA seed/snapshot review.
- Bootstrap restore checklist.
- Restore safety audit.

## Current operating model

- CMS snapshot is for CMS content/templates and CMS storage files.
- CMS snapshot excludes operational data such as contacts, WhatsApp interactions, appointments, subscribers, campaigns, and booking operational data.
- Full QA snapshot covers the whole local `pb_data` flow, based on Obsidian notes; current command details must be verified before use.
- Operational QA seed creates test operational data and must not be used against production data.
- Restore must not run automatically in container startup.
- Real restore should be a controlled one-off operation with backend stopped when write mode requires it.
- `NUVIO_ALLOW_DEV_RESET=1` is a dangerous dev/QA opt-in and must not remain on running services.

## Agent permissions

### Agents may

- Audit snapshot docs and manifests.
- Create restore checklists.
- Verify source docs mention records and storage coverage.
- Identify missing snapshot name, website ID, or restore target.
- Recommend dry-run validation when documented.

### Agents must not

- Do not run restore, reset, seed, migration, or destructive backup commands unless explicitly requested.
- Do not automate restore in container startup.
- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Snapshot type: CMS, full QA, operational QA seed, or real backup restore.
- Snapshot name/path.
- Target instance slug.
- Target `pb_data` path/volume.
- Target website ID guard, if available.
- Backend stopped/running requirement.
- Confirm token, if documented and explicitly approved.
- Whether operation is dry-run or real write.
- Backup path before restore.

## Standard workflow

1. Classify the snapshot/restore type.
2. Confirm this is audit/docs-only or explicitly approved execution.
3. Confirm target instance and `pb_data` path.
4. Confirm snapshot includes required records and storage files.
5. Confirm backend stopped requirement.
6. Confirm backup before destructive restore.
7. Use dry-run if supported and requested.
8. After restore, validate data, assets, preview, public pages, and critical flows.

## Validation checklist

- [ ] Snapshot type identified.
- [ ] Target instance and `pb_data` confirmed.
- [ ] Snapshot name and website ID confirmed.
- [ ] Backend stopped requirement confirmed.
- [ ] Pre-restore backup exists or is explicitly planned.
- [ ] Records and physical storage files restored together.
- [ ] Native file fields checked: `Websites.logo`, `Websites.seoImage`, `Assets.file`, `Pages.seo_social_image`, `Blocks.image`.
- [ ] CMS, assets, public runtime, preview, and SEO smoke checked after restore.
- [ ] `NUVIO_ALLOW_DEV_RESET` removed from running service env.

## Common failure modes

- Running restore against the wrong `pb_data` path.
- Restoring records without physical storage files.
- Using CMS snapshot as if it were an operational backup.
- Leaving destructive flags enabled after the operation.
- Restoring while backend is still writing to the database.
- Treating Obsidian command examples as current without checking repo tools.

## Reporting format

- Snapshot/restore type.
- Source docs read.
- Target instance/data path.
- Snapshot name and website ID.
- Whether action was audit-only, dry-run, or real write.
- Storage coverage result.
- Validation/smoke result.
- Explicit confirmation no destructive command was run unless requested.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Backup and Rollback](BACKUP_ROLLBACK.md)
- [Instance Bootstrap](INSTANCE_BOOTSTRAP.md)
- [Assets Feature Card](../features/ASSETS.md)
- [Security Feature Card](../features/SECURITY_CLIENT_ROLE.md)
