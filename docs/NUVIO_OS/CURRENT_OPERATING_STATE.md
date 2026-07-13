# Current Operating State

This file represents the current working state for agents. Current source code and git status override this document.

## Executive Status

| Area | Current state |
| --- | --- |
| Product maturity | Mostly built; release-readiness mode. |
| Backoffice | Central reusable product surface. |
| Public runtime | cms5 remains useful for current runtime/testing history. |
| Public template | Reference is the clean future public-site template candidate. |
| Deployment | Production-like deployment is a real blocker. |
| Nuvio own website/landing | Readiness/launch-critical. |
| First-client readiness | Active priority. |
| Broad backlog | Not active by default. |

## Launch-Critical Priorities

| Priority | Agent instruction |
| --- | --- |
| Coolify/base deployment | Use deployment docs and validate real provider state. |
| CMS snapshot restore | Controlled one-off restore only; include records and storage files. |
| Smoke testing | Prove critical demo flows and public/runtime behavior. |
| Nuvio own landing/site | Keep pricing cautious and do not connect scope to CMS unless requested. |
| Umami/analytics validation | Validate real deploy behavior; keep provider secrets server-side. |
| Security/client-role validation | Verify scoped endpoints and website access, not UI hiding. |
| First-client onboarding | Favor repeatable, safe, documented process over feature expansion. |

## Readiness Tasks

Readiness tasks are usually acceptable when scoped:

- deployment documentation;
- environment matrix clarification;
- smoke checklist refinement;
- regression matrix work;
- client-role/security checks;
- public endpoint validation;
- email/newsletter/booking E2E validation;
- reports/analytics confidence work;
- demo data/demo flow preparation;
- backup/restore rehearsal.

## Polish Tasks

Polish tasks are allowed only when local and low-risk:

- small UI layout fixes;
- copy clarity;
- display fallback cleanup;
- navigation/reference UI polish;
- small operational UX improvements.

Polish must not change backend behavior, endpoint contracts, auth, booking logic, newsletter send/lifecycle logic, or reports calculations unless explicitly scoped.

## Deferred / Post-First-Client

| Area | Current decision |
| --- | --- |
| Reviews / Google Places sync | Deferred/inactive for Nuvio 1.0. |
| Booking multi-capacity | Deferred. Current booking is not full capacity planning. |
| Advanced reports snapshots/history | Deferred. Current Reports is operational DTO/dashboard oriented. |
| Data exports | Deferred. Useful but not first-deploy blocker. |
| Self-host/Gitea/homelab migration | Deferred. Coolify path is current. |
| Final pricing model | Unknown / needs confirmation. Do not invent. |
| Advanced newsletter automation | Not current. Do not oversell. |

## Five Critical Demo Flows

| Flow | Success means |
| --- | --- |
| Website settings/setup | Identity, feature settings, SEO basics, preview origins, and public URLs behave safely. |
| CMS + SEO + public rendering | Pages, blocks, assets, SEO, preview, sitemap/robots, and public rendering work together. |
| Leads / Contact / WhatsApp | Public interest is captured, attributed, visible, scoped, and actionable. |
| Booking | Services, slots, appointment submit, status lifecycle, email behavior, and backoffice operations work safely. |
| Reports / Analytics / Health | Reports load useful scoped summaries, analytics setup state is clear, and health/deploy smoke passes. |

## Blockers vs Non-Blockers

| Type | Examples | Agent behavior |
| --- | --- | --- |
| Blocker | Deployment cannot run, restore breaks, public critical flow fails, client-role leaks data, secrets exposed. | Stop and fix or report clearly. |
| Readiness gap | Smoke checklist incomplete, docs unclear, tests missing for risky path. | Improve if scoped; otherwise report. |
| Polish | Spacing, wording, minor UI clarity. | Keep minimal and local. |
| Non-blocker | Deferred feature, advanced analytics, exports, multi-capacity, final pricing. | Do not implement unless explicitly revived. |

## Principle

Manual process is acceptable for first deployment when it is safe, documented, and repeatable enough.

Unsafe process is not acceptable, even if automated.

Examples:

- Manual snapshot restore is acceptable if backend state, target path, safety flags, and storage files are verified.
- Automatic restore on container startup is not acceptable.
- Manual smoke testing is acceptable if recorded.
- Hidden UI controls are not acceptable as the only security layer.