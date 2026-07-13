# Readiness Decision Matrix

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Launch Layer](README.md)

## Purpose

Help agents classify tasks before acting, especially when old backlog, polish requests, deployment readiness, or launch-critical gaps compete for attention.

## Read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines default task classification. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes task type to source docs. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Stop conditions and deferred-scope risk. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Validation by class/touched area. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Classification/report requirements. |
| 2 | Current Roadmap | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Current Roadmap.md` | Current phase and do-not-start list. |
| 2 | Deferred Features | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\01 - Product State\Deferred Features.md` | Known deferred work. |
| 2 | Feature Cards | [../features/README.md](../features/README.md) | Feature-specific risk. |
| 2 | Operations Cards | [../operations/README.md](../operations/README.md) | Operations risk. |

## Classification table

| Class | Meaning | Default action | Examples |
| --- | --- | --- | --- |
| Launch-critical | Blocks production-like deploy, demo, first-client safety, or core trust. | Prioritize, validate carefully, report risk. | Deployment health, public endpoint hardening minimum, client-role smoke, backup proof, contact/booking/newsletter lifecycle. |
| Readiness | Improves repeatability, demo confidence, handoff, documentation, or smoke coverage. | Usually do if scoped. | Demo runbook, env checklist, reports empty-state confidence, Umami smoke checklist, onboarding checklist. |
| Polish | Improves UI/copy/comfort but does not block core delivery. | Keep local; do not expand logic. | Booking modal label spacing, Reports card alignment, newsletter chip visual state. |
| Enhancement | Useful product improvement after first deployment path is stable. | Defer unless explicitly requested. | Data exports, booking multi-capacity, deeper monitoring/rate limiting. |
| Deferred/post-first-client | Known parked feature. | Do not implement unless explicitly revived. | Google Places/Reviews sync, advanced reports snapshots, custom newsletter lifecycle pages, self-host/Gitea migration. |
| Distraction | Does not move Nuvio toward deploy/demo/sale/delivery. | Stop or ask for scope. | Broad redesign, generic SaaS refactor, new UI library, unrelated architecture rewrite. |

## Decision questions

1. Does this affect one of the five critical demo flows?
2. Does this affect deployment, env, secrets, data, restore, CORS/CSP, public endpoints, emails, or analytics?
3. Does it help Nuvio become deployed, demonstrated, sold, or delivered to a first client?
4. Is it a proven gap from current source/docs, or old backlog resurfacing?
5. Can first client proceed manually without this?
6. Is the process safe even if manual?

## Nuvio-specific examples

| Task | Classification | Reason | Route |
| --- | --- | --- | --- |
| Real production-like deploy | Launch-critical | Nuvio is sellable once deployment path is proven. | [Deployment](../operations/DEPLOYMENT_COOLIFY.md) |
| Backend `/api/health` deployment smoke | Launch-critical | Required to prove service availability. | [Smoke Validation](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md) |
| Public endpoint hardening minimum | Launch-critical | Public forms are internet-facing. | [Security](../features/SECURITY_CLIENT_ROLE.md) |
| Client-role final smoke | Launch-critical | Prevents data/access leaks. | [Security](../features/SECURITY_CLIENT_ROLE.md) |
| Umami tracking setup | Readiness or launch-critical | Readiness if optional; launch-critical if demo/reporting promise depends on it. | [Umami Operations](../operations/UMAMI_ANALYTICS_OPERATIONS.md) |
| Reports empty states | Readiness | Prevents fake confidence in sparse demo data. | [Reports](../features/REPORTS_ANALYTICS_HEALTH.md) |
| Booking UI polish | Polish | Does not block launch if behavior is safe. | [Booking](../features/BOOKING.md) |
| Own landing/request review CTA | Launch-critical or readiness | Needed for selling/demo if Nuvio site is the acquisition path. | [Public Runtime](../features/PUBLIC_RUNTIME.md) |
| Demo data | Readiness | Helps demo; manual setup is acceptable if safe. | [Demo Flow Runbook](DEMO_FLOW_RUNBOOK.md) |
| Reviews/Google Places | Deferred/post-first-client | Explicitly deferred/inactive in current docs. | Do not start without explicit request. |
| Billing/self-service | Deferred/post-first-client | Not required for accompanied first client. | Do not start without explicit request. |
| Advanced newsletter automation | Deferred/post-first-client | Current newsletter lifecycle/campaign behavior is enough for first deploy. | [Newsletter](../features/NEWSLETTER.md) |
| Static visual redesign while deploy is blocked | Distraction | Does not prove deployment/smoke/security. | Stop or re-scope. |

## Old backlog rule

Old backlog must be classified before action.

If a note or request revives old backlog, agents must:

- check [../CORE.md](../CORE.md) source order;
- check current roadmap/deferred docs;
- mark launch relevance;
- ask/report if it conflicts with current release-readiness mode;
- avoid implementation unless the user explicitly scopes it.

## Reporting requirements

Always report:

- classification chosen;
- why it was chosen;
- linked source docs/cards;
- whether it affects launch/demo/first-client readiness;
- validation needed;
- whether it is safe to defer.

## Related docs

- [README](README.md)
- [First Client Readiness](FIRST_CLIENT_READINESS.md)
- [Launch Blockers vs Polish](LAUNCH_BLOCKERS_VS_POLISH.md)
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
