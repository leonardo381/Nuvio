# Launch and Readiness Agent Layer

## Purpose

This layer connects Feature Agent Cards and Operations Agent Cards into a practical execution path for demo readiness, first-client readiness, base deployment readiness, blocker classification, agent handoff, and launch validation.

Use it when the question is not just "can this feature work?" but "does this move Nuvio closer to being deployed, demonstrated, sold, or delivered to a first client?"

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

## Launch-readiness principle

Does this move Nuvio closer to being deployed, demonstrated, sold, or delivered to a first client?

If the answer is no, classify the work as polish, enhancement, deferred, or distraction before acting.

## When To Use This Layer

| Use this layer when | Then |
| --- | --- |
| Task mentions demo, first client, launch, Nuvio Base, deployment readiness, blocker, polish, handoff, or readiness. | Open the relevant Launch doc. |
| Task touches a critical demo flow. | Also open the relevant Feature card. |
| Task touches deployment/env/backup/public runtime. | Also open the relevant Operations card. |
| Task revives old backlog. | Classify it with [READINESS_DECISION_MATRIX.md](READINESS_DECISION_MATRIX.md) before action. |

## Launch Docs

| Doc | Use when |
| --- | --- |
| [FIRST_CLIENT_READINESS.md](FIRST_CLIENT_READINESS.md) | Deciding whether Nuvio is safe enough for a first accompanied client. |
| [NUVIO_BASE_DEPLOYMENT_READINESS.md](NUVIO_BASE_DEPLOYMENT_READINESS.md) | Preparing the first production-like demo/staging instance. |
| [DEMO_FLOW_RUNBOOK.md](DEMO_FLOW_RUNBOOK.md) | Running or rehearsing the five critical demo flows. |
| [AGENT_HANDOFF_CHECKLIST.md](AGENT_HANDOFF_CHECKLIST.md) | Passing work between agents without losing context or overstating success. |
| [READINESS_DECISION_MATRIX.md](READINESS_DECISION_MATRIX.md) | Classifying work as launch-critical, readiness, polish, enhancement, deferred, or distraction. |
| [LAUNCH_BLOCKERS_VS_POLISH.md](LAUNCH_BLOCKERS_VS_POLISH.md) | Separating known blockers/readiness gaps from acceptable imperfections. |

## Source Status Snapshot

| Area | Status from current docs |
| --- | --- |
| Backoffice product | Feature-complete / beta-stable per Operating Manual product-state docs. |
| Current mode | Release-readiness and production-like deployment proof. |
| Real Coolify deploy | Not proven complete; Unknown / needs confirmation before claiming done. |
| Nuvio Base online | Not proven complete; Unknown / needs confirmation before claiming done. |
| Nuvio website/landing published | Not proven complete; Unknown / needs confirmation before claiming done. |
| Backup automation | Not proven complete; manual backup/restore proof may be acceptable first. |
| First-client readiness | Depends on deployment smoke, security/client-role smoke, lifecycle checks, demo data, and onboarding path. |

## Five Critical Demo Flows

1. Website setup / website settings.
2. CMS + SEO + public rendering.
3. Leads / Contact / WhatsApp.
4. Booking.
5. Reports / Analytics / Health.

## Related Operations Cards

- [Deployment and Coolify](../operations/DEPLOYMENT_COOLIFY.md)
- [Docker and Compose](../operations/DOCKER_COMPOSE.md)
- [Instance Bootstrap](../operations/INSTANCE_BOOTSTRAP.md)
- [Env and Secrets](../operations/ENV_SECRETS.md)
- [Snapshot and Restore](../operations/SNAPSHOT_RESTORE.md)
- [Backup and Rollback](../operations/BACKUP_ROLLBACK.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)
- [Umami Analytics Operations](../operations/UMAMI_ANALYTICS_OPERATIONS.md)
- [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md)

## Related Feature Cards

- [CMS](../features/CMS.md)
- [Website Settings and SEO](../features/WEBSITE_SETTINGS_SEO.md)
- [Leads, Contact Form, and WhatsApp](../features/LEADS_CONTACT_WHATSAPP.md)
- [Booking](../features/BOOKING.md)
- [Newsletter](../features/NEWSLETTER.md)
- [Reports, Analytics, and Health](../features/REPORTS_ANALYTICS_HEALTH.md)
- [Public Runtime](../features/PUBLIC_RUNTIME.md)
- [Security and Client Role](../features/SECURITY_CLIENT_ROLE.md)
- [Emails and Templates](../features/EMAILS_TEMPLATES.md)

## Launch Rules

- Nuvio is mostly built but not yet fully proven in production-like, demo-ready, first-client-ready conditions.
- Manual process is acceptable for the first client. Unsafe process is not.
- Do not revive old backlog items automatically.
- Do not treat polish as a launch blocker.
- Do not create new product scope to avoid deployment work.
- Do not claim success without verification.
- Do not run destructive restore, migration, backup, or deploy commands unless explicitly requested.
- Do not expose secrets in docs, logs, screenshots, `VITE_*`, or browser code.
