# Agent Process Registry

## Purpose

This registry lists repeatable Nuvio agent processes. Active processes have a process directory and can be used by agents. Planned processes are reserved names and should not be invented ad hoc during implementation work.

## Active Processes

| Process | Status | Purpose | Path | Notes / when to use |
| --- | --- | --- | --- | --- |
| `WEBSITE_FACTORY` | active | Build client or official public websites from brief to deploy through strict staged gates. | [processes/website_factory](processes/website_factory/README.md) | Use when a website must move from client brief, source block selection, import, copy, visual adaptation, integration, QA, and handoff. |

## Planned / Future Processes

| Process | Status | Purpose | Current path | Notes / when to use |
| --- | --- | --- | --- | --- |
| `WEBSITE_AUDIT_AND_PROPOSAL` | planned | Audit an existing business/site and produce a scoped proposal. | Not created. | Use before Website Factory when the brief is unclear. |
| `FIRST_CLIENT_ONBOARDING` | planned | Guide first accompanied client setup and handoff. | Not created. | Should align with launch readiness docs. |
| `DEPLOYMENT` | planned | Run staged deployment planning/execution with gates. | Not created. | Should reference operations deployment docs and never bypass approvals. |
| `RELEASE_REGRESSION` | planned | Run release readiness and regression checks. | Not created. | Should align with validation matrix and release workflow. |
| `CONTENT_UPDATE` | planned | Process client/site content updates safely. | Not created. | Should separate intake, copy, CMS update, QA, and handoff. |
| `SUPPORT_REQUEST_HANDLING` | planned | Triage and resolve support requests. | Not created. | Should include severity, reproduction, fix, validation, and report gates. |
| `ANALYTICS_REVIEW` | planned | Review analytics and business signals. | Not created. | Should avoid fake insights and PII exposure. |
| `SEO_REVIEW` | planned | Review SEO fields, public output, and practical fixes. | Not created. | Should not invent unsupported technical checks. |
| `NEWSLETTER_CAMPAIGN_SETUP` | planned | Prepare newsletter campaigns safely. | Not created. | Should separate audience, copy, save, preview, approval, and send. |
| `BOOKING_SETUP` | planned | Configure booking services, availability, and public flow. | Not created. | Should separate settings, slots, emails, QA, and handoff. |
| `NUVIO_OS_DOC_UPDATE` | planned | Update Nuvio OS docs safely. | Not created. | Should preserve source hierarchy and avoid broad rewrites. |

## Registry Rules

- Do not create a process during implementation unless the user scopes a process-doc phase.
- Do not treat planned processes as active instructions.
- If a task matches an active process, use that process before improvising a workflow.
