# Five Flow Regression Task Pack

## Purpose
Use this task pack for broad release-readiness regression over the five critical Nuvio flows: website setup, CMS/SEO/public rendering, Leads/Contact/WhatsApp, Booking, and Reports/Analytics/Health.

## Task classification
- regression
- readiness
- launch-critical

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Demo Flow Runbook](../launch/DEMO_FLOW_RUNBOOK.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)
- [CMS](../features/CMS.md)
- [Website Settings and SEO](../features/WEBSITE_SETTINGS_SEO.md)
- [Assets](../features/ASSETS.md)
- [Leads, Contact, and WhatsApp](../features/LEADS_CONTACT_WHATSAPP.md)
- [Booking](../features/BOOKING.md)
- [Newsletter](../features/NEWSLETTER.md)
- [Reports, Analytics, and Health](../features/REPORTS_ANALYTICS_HEALTH.md)
- [Public Runtime](../features/PUBLIC_RUNTIME.md)

## Optional source docs
- Obsidian Backoffice 1.0 Status and Current Roadmap.
- Obsidian feature docs for CMS, Booking, Leads, Newsletter, Reports, Public Runtime, and Security Hardening.
- Current source/tests for each touched module.
- Deployment smoke checklist if testing a deployed instance.

## Preconditions
- Target environment and website are known.
- Admin and client-role test users are available or skipped with reason.
- Public runtime URL and backoffice URL are known.
- Enabled features are known.
- Task is audit/regression unless implementation is explicitly requested.

## Source-of-truth rules
1. Current source, tests, runtime behavior, and git status win.
2. Repo contracts define implementation boundaries.
3. Nuvio OS docs define regression scope and guardrails.
4. Obsidian docs provide context only.
5. Skipped checks must be reported as skipped, not passed.

## Allowed work
- Run or plan a no-change regression audit.
- Create a regression checklist or report.
- Identify blockers vs readiness gaps vs polish.
- Run validation commands only if explicitly allowed and setup is available.

## Forbidden work
- Do not fix unrelated issues during regression unless user scopes a fix.
- Do not invent pass/fail results without running or observing checks.
- Do not use fake provider/analytics/email success.
- Do not treat polish as launch blocker.
- Do not touch runtime data or snapshots unless explicitly approved.

## Danger zones
- Broad scope can hide skipped checks.
- Public endpoints can expose PII/tokens if logs are inspected carelessly.
- Booking and newsletter have customer-facing side effects.
- Reports can show fake confidence if provider state is unknown.

## Execution outline
1. Confirm target instance/environment and enabled features.
2. Read the demo runbook and validation matrix.
3. Map each of the five flows to feature cards and manual smoke checks.
4. Execute only approved checks.
5. Record pass/fail/skipped with reasons.
6. Classify issues as blocker, readiness gap, polish, enhancement, or deferred.
7. Recommend the smallest next phase.

## Validation checklist
### Doc validation
- Regression checklist exists in the report.
- Skipped checks include reason and residual risk.
- No unsupported current-state claims are made.

### Code/build/test validation, if future implementation applies
- Run documented UI/backend/public checks only for touched areas or requested regression scope.
- Do not run broad test suites or builds unless requested and safe.
- Report unavailable scripts or failing setup honestly.

### Manual smoke validation
- Website setup/settings flow.
- CMS edit, preview, public rendering, assets, and SEO.
- Public contact and WhatsApp if enabled.
- Public booking and backoffice appointment handling if enabled.
- Reports/analytics/health state.
- Client-role access if applicable.

### User confirmation needed
- Target environment and website.
- Available users/credentials.
- Enabled features.
- Whether provider side effects are safe to trigger.
- Whether failures should be fixed now or only reported.

## Expected report format
- Files/sources read.
- Flows checked.
- Pass/fail/skipped table.
- Blockers/readiness gaps/polish.
- Unknowns and risks.
- Validation performed.
- Confirmation no product changes unless explicitly scoped.
- Next recommended phase.

## Stop conditions
- No target environment is specified.
- Checks require credentials or provider access not available.
- A check would send real email or mutate customer data without approval.
- Regression reveals a blocker outside requested scope.
