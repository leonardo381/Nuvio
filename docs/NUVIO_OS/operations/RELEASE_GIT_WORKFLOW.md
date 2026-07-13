# Release and Git Workflow

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Operations Cards](README.md)

## Purpose for agents

Help agents keep git/release work clean, scope-bound, and deployment-safe, especially for docs-only phases and tag-based releases.

## Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Git/status checklist and global rules. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes release and docs-only tasks. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Area-specific validation requirements. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Changed-files reporting rules. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Stop conditions before risky changes. |
| 2 | Obsidian Git Workflow | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Git Workflow.md` | Human release flow. |
| 2 | Obsidian Commands Cheat Sheet | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Commands Cheat Sheet.md` | Command examples; verify before use. |
| 2 | Deployment Quick Guide | [../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md) | Version/tag as instance decision. |

## When to use this card

- Release planning.
- Tag/version readiness.
- Docs-only scope checks.
- Git status/diff hygiene.
- Preparing final reports for changed files.

## Current operating model

- Obsidian notes state daily work happens on `dev`; releases go through `main` and tags.
- Clients are instances, not branches.
- Deployments should use `main` or approved tags, not `dev`.
- Docs-only Nuvio OS phases should only modify `docs/NUVIO_OS` unless explicitly requested.
- Current branch and exact release tag readiness must be checked from git when release work is requested.

## Agent permissions

### Agents may

- Run read-only git status/diff/log commands.
- Report changed files and unexpected changes.
- Create docs-only release checklists.
- Recommend validation based on touched areas.
- Stop and ask/report if unrelated dirty files are present in touched paths.

### Agents must not

- Do not amend commits or rewrite history unless explicitly requested.
- Do not run destructive git commands unless explicitly requested and approved.
- Do not tag or push unless explicitly requested.
- Do not commit real `.env` files.
- Do not expose secrets in `VITE_*` variables.
- Do not use wildcard localhost-style CORS in production.
- Do not share writable `pb_data` across real clients.
- Do not run destructive restore/migration/backup commands unless explicitly requested.
- Do not assume local success equals deployment readiness.
- Do not change deployment architecture without explicit approval.

## Inputs required before acting

- Target repo.
- Whether task is docs-only, implementation, release, or hotfix.
- Target branch/tag policy.
- Expected changed files.
- Validation commands requested or required.
- Whether commit/tag/push is explicitly requested.
- Deployment target, if release is tied to deployment.

## Standard workflow

1. Confirm target repo and allowed files.
2. Run `git status --short --untracked-files=all` when allowed.
3. Read relevant source docs/contracts.
4. Make only requested scoped changes.
5. Verify changed files remain in scope.
6. Run validation only when required by phase and touched area.
7. Report changed files, validation, skipped checks, and non-touched areas.
8. Do not commit/tag/push unless explicitly requested.

## Validation checklist

- [ ] `git status --short --untracked-files=all` reviewed.
- [ ] Changed files are within scope.
- [ ] No `.env`, `pb_data`, database, storage, or log files are tracked.
- [ ] Product validation was run only when requested/required.
- [ ] Docs-only phases changed only docs paths.
- [ ] Release tags/branches are not assumed without git confirmation.
- [ ] Final report includes files read, files changed, validation, risks, and non-actions.

## Common failure modes

- Editing outside allowed scope because the file was open in the IDE.
- Running formatters across the repo during a scoped task.
- Committing generated/runtime files.
- Treating old root README/PocketBase docs as current Nuvio truth.
- Tagging work before validation.
- Ignoring dirty unrelated files.

## Reporting format

- Target repo.
- Current git status summary.
- Changed files.
- Whether docs-only/product code/config/env changed.
- Validation run or skipped.
- Any unrelated dirty files observed.
- Explicit confirmation no commit/tag/push was performed unless requested.

## Related Nuvio OS docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Deployment and Coolify](DEPLOYMENT_COOLIFY.md)
- [Backup and Rollback](BACKUP_ROLLBACK.md)
- [Security Feature Card](../features/SECURITY_CLIENT_ROLE.md)
