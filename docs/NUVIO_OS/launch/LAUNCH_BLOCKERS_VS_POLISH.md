# Launch Blockers vs Polish

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Launch Layer](README.md)

## Purpose

List current launch-critical blockers, readiness gaps, polish, enhancements, and deferred work so agents do not block first-client progress with non-critical improvements or ignore unsafe gaps.

This file summarizes source docs; it must not invent status. If current source code or git status disagrees with docs, current source wins.

## Read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Current operating mode and classification. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes blockers by area. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Unsafe launch areas. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Required checks. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Reporting requirements. |
| 2 | Backoffice 1.0 Status | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Backoffice 1.0 Status.md` | Done/not-done status. |
| 2 | Current Roadmap | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Current Roadmap.md` | Current phase and next steps. |
| 2 | Deferred Features | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Deferred Features.md` | Explicit deferred work. |
| 2 | Operations Cards | [../operations/README.md](../operations/README.md) | Deployment and smoke readiness. |
| 2 | Feature Cards | [../features/README.md](../features/README.md) | Feature-level launch risk. |

## Current known launch-critical blockers / must-prove items

| Item | Status | Why it matters | Route |
| --- | --- | --- | --- |
| Production-like deploy | Not proven complete / needs confirmation | Product is sellable once deployment path is proven. | [Deployment](../operations/DEPLOYMENT_COOLIFY.md) |
| Nuvio Base online | Not proven complete / needs confirmation | Base instance proves deploy, restore, public runtime, smoke, and backup path. | [Base Deployment Readiness](NUVIO_BASE_DEPLOYMENT_READINESS.md) |
| CMS snapshot restore into deployed instance | Needs proof in target environment | Base CMS content and storage must restore together. | [Snapshot Restore](../operations/SNAPSHOT_RESTORE.md) |
| Deployment smoke tests | Needs proof in target environment | Local readiness is not deployment readiness. | [Smoke Validation](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md) |
| Security/client-role final smoke | Needs proof before first client | Prevents accidental data/access leaks. | [Security](../features/SECURITY_CLIENT_ROLE.md) |
| Public endpoint hardening minimum | Done in docs, needs final smoke | Contact/WhatsApp/newsletter/booking are public surfaces. | [Security](../features/SECURITY_CLIENT_ROLE.md) |
| Emails/newsletter/booking lifecycle validation | Needs environment-specific proof if enabled | Broken links/emails damage trust quickly. | [Emails](../features/EMAILS_TEMPLATES.md) |
| Initial backup and restore rehearsal | Backup needed before handoff; rehearsal may be staged | Customer data requires recovery confidence. | [Backup Rollback](../operations/BACKUP_ROLLBACK.md) |
| Own Nuvio landing/request review path | Not proven published / needs confirmation | Needed if using Nuvio site as acquisition/demo path. | [Public Runtime](../features/PUBLIC_RUNTIME.md) |

## Readiness gaps

| Item | Status | Why not always a blocker | Route |
| --- | --- | --- | --- |
| Reports confidence with sparse/demo data | Needs smoke and honest empty states | Can proceed if clear unavailable/setup states are shown. | [Reports](../features/REPORTS_ANALYTICS_HEALTH.md) |
| Umami analytics readiness | Optional unless demo/reporting depends on it | Reports can show setup/unavailable state honestly. | [Umami](../operations/UMAMI_ANALYTICS_OPERATIONS.md) |
| Demo data/demo website | Needed for persuasive demo | Manual prep is acceptable if safe and not fake. | [Demo Runbook](DEMO_FLOW_RUNBOOK.md) |
| First-client onboarding process | Manual process acceptable | Automation is not required for accompanied first client. | [First Client Readiness](FIRST_CLIENT_READINESS.md) |
| Deployment metadata/private record | Needed before handoff | Can be manual in private tracker. | [Instance Bootstrap](../operations/INSTANCE_BOOTSTRAP.md) |
| Backup automation | Not done per product state docs | Manual backup proof can be acceptable early; unsafe/no backup is not. | [Backup Rollback](../operations/BACKUP_ROLLBACK.md) |

## Polish, not blockers

| Item | Why polish |
| --- | --- |
| Fine UI spacing/alignment issues that do not break flows | Improve perceived quality but do not block launch if workflows are safe. |
| Additional Reports visual refinements | Useful after demo-critical data confidence is proven. |
| Newsletter selection micro-polish after behavior is correct | Does not block first client if save/send/lifecycle are safe. |
| Booking modal label polish after slot/reschedule behavior is safe | Does not block launch if booking logic and emails work. |
| Copy refinements that do not affect core promise or legal/trust posture | Can continue iteratively. |

## Enhancements / post-first-client

| Item | Source status | Why not first-client blocker |
| --- | --- | --- |
| Data exports | Later/deferred | Useful, not required for first deploy. |
| Booking multi-capacity | Later/deferred | Current booking is single-capacity. |
| Advanced reports snapshots/history | Later/deferred | Current reports can be operational/current-state. |
| Advanced newsletter automation | Deferred | Lifecycle and campaign basics are enough for first client. |
| Deeper rate limiting/monitoring | Important after deploy | Not a replacement for current public endpoint hardening smoke. |
| Client onboarding automation | Not done | Manual accompanied onboarding is acceptable. |
| Self-host/Gitea/homelab migration | Do-not-start now | Coolify path is current deployment target. |

## Deferred / do not revive without explicit request

- Google Places / Reviews sync.
- Custom newsletter confirm/unsubscribe pages.
- Data exports.
- Booking multi-capacity.
- Advanced reports snapshots/history.
- Embed host allowlisting beyond current baseline.
- Remaining public runtime sanitization beyond current prioritized hardening.
- Gitea/self-host migration.
- Mini server/homelab.
- Billing/self-service portal unless explicitly scoped later.

## Acceptable imperfections for controlled demo / first client

- Manual onboarding by operator.
- Manual backup process if backup exists and restore path is understood.
- Manual demo data setup.
- Clear analytics unavailable/setup state when Umami is not configured.
- Sparse reports, as long as no fake insights are shown.
- Some UI polish gaps, as long as five critical flows are safe and understandable.
- Limited automation around newsletter, reports, and onboarding.

## Not acceptable

- No isolated `pb_data`/storage for real client instance.
- No backup after smoke and before handoff.
- Secrets in browser env or docs.
- Wildcard/local CORS in production-like deploy.
- Public endpoints leaking internals or PII.
- Client-role scoping unverified.
- Broken CMS restore storage references.
- Broken contact/booking/newsletter lifecycle links while those features are enabled.
- Claiming Nuvio Base or Nuvio site is online without verification.

## How to update this file

Agents may update this file only when:

- source docs change;
- current source/gstatus proves a status change;
- a launch-readiness task explicitly asks for it;
- an unknown is confirmed.

Always report:

- status source;
- changed classification;
- validation behind the change;
- whether it changes first-client readiness.

## Related docs

- [README](README.md)
- [First Client Readiness](FIRST_CLIENT_READINESS.md)
- [Nuvio Base Deployment Readiness](NUVIO_BASE_DEPLOYMENT_READINESS.md)
- [Readiness Decision Matrix](READINESS_DECISION_MATRIX.md)
- [Smoke Validation](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)
- [Deployment](../operations/DEPLOYMENT_COOLIFY.md)
- [Env Secrets](../operations/ENV_SECRETS.md)
- [Public Runtime](../features/PUBLIC_RUNTIME.md)
- [Security Client Role](../features/SECURITY_CLIENT_ROLE.md)
- [Reports Analytics Health](../features/REPORTS_ANALYTICS_HEALTH.md)
