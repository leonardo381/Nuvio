# First Client Onboarding Task Pack

## Purpose
Use this task pack when preparing the first accompanied client, first-client readiness, onboarding runbook, handoff, base deployment go/no-go, or client-facing safety checklist.

## Task classification
- launch-critical
- operations
- sales/demo
- readiness

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [First Client Readiness](../launch/FIRST_CLIENT_READINESS.md)
- [Nuvio Base Deployment Readiness](../launch/NUVIO_BASE_DEPLOYMENT_READINESS.md)
- [Launch Blockers vs Polish](../launch/LAUNCH_BLOCKERS_VS_POLISH.md)
- [Readiness Decision Matrix](../launch/READINESS_DECISION_MATRIX.md)
- [Agent Handoff Checklist](../launch/AGENT_HANDOFF_CHECKLIST.md)
- [Deployment and Coolify](../operations/DEPLOYMENT_COOLIFY.md)
- [Environment and Secrets](../operations/ENV_SECRETS.md)
- [Backup and Rollback](../operations/BACKUP_ROLLBACK.md)
- [Demo Flow Runbook](../launch/DEMO_FLOW_RUNBOOK.md)

## Optional source docs
- [Deployment Quick Guide](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md)
- [Instance Bootstrap Checklist](../../NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md)
- [Coolify Base Deployment Plan](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md)
- Obsidian Backoffice 1.0 Status, Current Roadmap, Deployment Quick Guide, Coolify Plan, and feature docs.
- Private deployment tracker or secrets manager for real values, never docs.

## Preconditions
- Client/instance identity and scope are known.
- Deployment status is verified or marked unknown.
- Domains/env/secrets/backup/snapshot status are known or marked unknown.
- Client-role user and assigned website model are understood.
- Manual substitute is accepted for any unavailable landing/onboarding automation.
- User has approved any live deployment or restore action if requested.

## Source-of-truth rules
1. Current deploy/source/client setup and git status win.
2. Repo deployment docs define the repeatable instance model.
3. Launch docs define readiness and go/no-go classification.
4. Obsidian docs provide current product state context.
5. Private deployment records/secrets are required for real values and must not be copied into docs.

## Allowed work
- Prepare onboarding readiness checklist.
- Classify blockers vs acceptable imperfections.
- Use docs to plan client handoff and smoke validation.
- Recommend next step without inventing deployment/client status.
- Update docs only when scoped.

## Forbidden work
- Do not claim Nuvio Base is online unless verified.
- Do not commit real client secrets/domains.
- Do not start deferred feature work as onboarding.
- Do not treat polish as blocker.
- Do not onboard without backup/restore/security smoke clarity.

## Danger zones
- First client gets access to wrong website data.
- No backup or restore rehearsal.
- Provider secrets missing or exposed.
- Deployment status assumed from docs.
- Nuvio landing/request-review path unavailable but claimed.
- Deferred features promised to client.

## Execution outline
1. Confirm client/onboarding scope and current deployment status.
2. Read launch readiness, deployment, env, backup, and demo docs.
3. List must-haves, readiness gaps, polish, and deferred items.
4. Check whether each first-client requirement is verified, unknown, or blocked.
5. Prepare handoff report and next action.
6. Stop before live deploy/restore without explicit approval.

## Validation checklist
### Doc validation
- Readiness table uses verified/unknown/blocked labels.
- No real client secrets or domains are included.
- Deferred items remain deferred unless explicitly revived.

### Code/build/test validation, if future implementation applies
- If future execution applies, run deployment smoke, client-role smoke, feature smoke, and backup proof according to Validation Matrix.
- If code/docs are changed, run only scoped validation for touched area.
- Do not run builds/tests in docs-only onboarding planning.

### Manual smoke validation
- Backend health.
- Backoffice login and CMS.
- Public runtime.
- CMS preview.
- Assets.
- Contact/newsletter/booking if enabled.
- Reports.
- Client-role assigned user scope.
- Backup exists.
- Handoff report is complete.

### User confirmation needed
- Client identity/instance slug.
- Domains/env/secrets.
- Snapshot name and backup target.
- Client-role user setup.
- Live deploy/restore approval.
- Known acceptable imperfections.

## Expected report format
- Files read.
- Files changed.
- Readiness status by category.
- Unknowns and blockers.
- Risks.
- Validation run/skipped.
- Confirmation no live deploy/client data changes unless approved.
- Next recommended step.

## Stop conditions
- Client/deployment identity is unclear.
- Live deployment or restore is needed but not approved.
- Security/client-role smoke cannot be performed but is required for go/no-go.
- Secrets/domains would need to be printed.
- User asks to promise unavailable/deferred features.
