# Instance Bootstrap

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents reason about creating or auditing a Nuvio instance without confusing client instances with branches, repos, or shared runtime data.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Source order and instance safety. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes bootstrap tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Smoke checks after bootstrap. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Report structure. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Data isolation and destructive path warnings. |
| 2 | Instance Bootstrap Checklist | [../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md) | Full new instance checklist. |
| 2 | Deployment Quick Guide | [../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md) | Ten required decisions and minimum env. |
| 2 | Deployment Env Matrix | [../../NUVIO_DEPLOYMENT_ENV_MATRIX.md](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md) | Env reference. |
| 2 | Obsidian Instance Model | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Instance Model.md` | Human mental model for instances. |

## When to use this card

- Creating a new base/client/staging instance plan.
- Auditing whether an instance has isolated data/storage/env.
- Preparing snapshot/bootstrap records.
- Creating deployment checklists.

## Current operating model

- A client is an instance/deployment, not a branch and not a repo.
- Code is shared and deployed from approved tags or `main`; `dev` is active development.
- Each instance owns its own env, `pb_data`, storage files, domains, backups, and CMS snapshot/bootstrap data.
- CMS snapshot can bootstrap base content; operational data should not be dragged into clean base instances.
- Google Places/Reviews is documented but deferred/inactive unless explicitly enabled.
- First real deployment/Nuvio Base online status: pending / needs confirmation from operator.

## Agent permissions

### Agents may

- Create docs/checklists for instance setup.
- Audit instance decision records for missing fields.
- Identify missing env/domain/backup/snapshot inputs.
- Prepare smoke checklist and handoff notes.

### Agents must not

- Do not create client branches or client repos as a default deployment strategy.
- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not run destructive restore/migration/backup commands unless explicitly requested.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Instance/client slug.
- Environment: production, staging, QA, or local.
- Public site URL.
- Admin/backoffice URL.
- Backend/API URL.
- `pb_data` volume/path.
- Storage path.
- Backup path/bucket and retention.
- CMS snapshot name and website ID.
- Code version/tag for backoffice/backend.
- Code version/tag for public runtime.
- Enabled features.
- Resend/Umami/Google Places status.
- CORS/CSP/preview origins.

## Standard workflow

1. Confirm the task is bootstrap planning or execution.
2. Collect the ten instance decisions from the Deployment Quick Guide.
3. Confirm data and storage isolation.
4. Confirm env values and secrets policy.
5. Confirm deployment target and code versions.
6. Confirm CMS snapshot source and storage coverage.
7. Confirm optional providers only for enabled features.
8. Prepare smoke tests and initial backup checklist.
9. Report missing inputs before any irreversible step.

## Validation checklist

- [ ] Instance slug recorded.
- [ ] Public/admin/API URLs defined.
- [ ] `pb_data` and storage are isolated.
- [ ] Backup target and retention defined.
- [ ] CMS snapshot name/version recorded.
- [ ] Code tags/versions recorded.
- [ ] Enabled features listed.
- [ ] Env values split into server-only and browser-exposed groups.
- [ ] Smoke tests planned.
- [ ] Initial backup planned before handoff.

## Common failure modes

- Creating branches per client.
- Reusing one writable volume across clients.
- Restoring operational QA data into a clean base instance.
- Forgetting public URL values used by emails, sitemap, robots, and preview.
- Enabling optional providers without secrets/domain verification.
- Treating a CMS snapshot as a full operational backup.

## Reporting format

- Instance goal and environment.
- Inputs collected.
- Missing inputs.
- Isolation model for env, data, storage, domains, backups.
- Snapshot/bootstrap source.
- Enabled features/providers.
- Smoke and backup plan.
- Explicit non-actions: no deploy/restore/migration unless requested.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Env and Secrets](ENV_SECRETS.md)
- [Snapshot and Restore](SNAPSHOT_RESTORE.md)
- [Backup and Rollback](BACKUP_ROLLBACK.md)
- [CMS Feature Card](../features/CMS.md)
