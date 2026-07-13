# Demo Flow and Data Task Pack

## Purpose
Use this task pack when preparing or validating the Nuvio demo narrative, demo data, five critical flows, first-client presentation path, or sales/demo handoff.

## Task classification
- sales/demo
- readiness
- regression

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Demo Flow Runbook](../launch/DEMO_FLOW_RUNBOOK.md)
- [First Client Readiness](../launch/FIRST_CLIENT_READINESS.md)
- [Launch Blockers vs Polish](../launch/LAUNCH_BLOCKERS_VS_POLISH.md)
- [Readiness Decision Matrix](../launch/READINESS_DECISION_MATRIX.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)
- [CMS](../features/CMS.md)
- [Leads, Contact, and WhatsApp](../features/LEADS_CONTACT_WHATSAPP.md)
- [Booking](../features/BOOKING.md)
- [Reports, Analytics, and Health](../features/REPORTS_ANALYTICS_HEALTH.md)

## Optional source docs
- Obsidian Backoffice 1.0 Status, Current Roadmap, and feature docs.
- Current demo website/CMS data source if provided.
- Deployment docs if demo runs against deployed environment.
- Nuvio official landing/source if demo starts from request-review path.

## Preconditions
- Demo target is known: local, staging, deployed base, or screenshots only.
- Demo data source is known and safe.
- Features enabled for demo are known.
- Known failures/polish items are classified.
- Task is demo prep/audit unless implementation is explicitly requested.

## Source-of-truth rules
1. Current deployed/source state wins.
2. Launch docs define demo flow and readiness classification.
3. Feature cards define module-specific safety.
4. Obsidian docs provide product narrative context.
5. Demo claims must match verified behavior or be marked unknown.

## Allowed work
- Prepare a demo flow checklist.
- Classify blockers vs polish.
- Use safe demo data and avoid real visitor/customer PII.
- Recommend demo substitutions for unavailable live components.
- Document skipped checks honestly.

## Forbidden work
- Do not fake live provider success.
- Do not use real customer data without approval.
- Do not present deferred features as available.
- Do not block demo on minor polish.
- Do not alter production data just to make a demo look good.

## Danger zones
- Fake data presented as real.
- Skipped smoke hidden as success.
- Demo account showing wrong website/client data.
- Provider/email side effects during demo.
- Nuvio landing/request-review path claimed before verified.

## Execution outline
1. Confirm demo environment, audience, and allowed data.
2. Read demo/readiness docs and relevant feature cards.
3. Map the demo narrative to five critical flows.
4. Identify blockers, readiness gaps, polish, and substitutions.
5. Prepare final demo checklist and handoff.
6. Stop before mutating data unless explicitly scoped.

## Validation checklist
### Doc validation
- Demo flow and data sources are documented.
- Known unknowns and substitutions are listed.
- No real customer PII is included.

### Code/build/test validation, if future implementation applies
- If future implementation applies, run feature-specific validation for any changed flow.
- If demo environment is live, run smoke checks instead of relying on docs.
- Do not run builds/tests in docs-only demo planning.

### Manual smoke validation
- Website settings/setup.
- CMS + SEO + public rendering.
- Leads/contact/WhatsApp.
- Booking.
- Reports/analytics/health.
- Optional landing/request-review path.
- Client-role view if part of demo.

### User confirmation needed
- Demo target environment.
- Demo data/snapshot.
- Audience and required flows.
- Whether provider side effects are allowed.
- Known acceptable imperfections.

## Expected report format
- Files read.
- Files changed.
- Demo flows prepared/validated.
- Known blockers/polish.
- Unknowns and risks.
- Validation run/skipped.
- Confirmation no production data changed.
- Next recommended step.

## Stop conditions
- Demo target or data source is unclear.
- Demo would require real customer PII.
- A critical flow is broken and user expects go/no-go.
- Provider/email side effects are required but not approved.
