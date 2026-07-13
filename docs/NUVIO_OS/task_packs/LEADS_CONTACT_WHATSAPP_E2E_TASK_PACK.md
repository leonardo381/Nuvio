# Leads Contact WhatsApp E2E Task Pack

## Purpose
Use this task pack for public contact form, WhatsApp tracking, Leads dashboard, attribution/context, follow-up status, client-role Leads access, or related end-to-end validation.

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
- [Leads, Contact, and WhatsApp](../features/LEADS_CONTACT_WHATSAPP.md)
- [Security and Client Role](../features/SECURITY_CLIENT_ROLE.md)
- [Emails and Templates](../features/EMAILS_TEMPLATES.md)
- [Reports, Analytics, and Health](../features/REPORTS_ANALYTICS_HEALTH.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)

## Optional source docs
- Obsidian Leads, Contact Form and WhatsApp, Reports, and Security Hardening docs.
- Current contact/WhatsApp backend source and tests.
- Current Leads UI source and dashboard DTO tests.
- Public runtime/contact route source if a public site is involved.

## Preconditions
- Target website/public site is known.
- Allowed test contact data is safe and not real visitor PII.
- Email/notification side effects are known or disabled.
- Client-role test user is available if scoped access is being checked.
- Task mode is audit/regression unless implementation is explicitly requested.

## Source-of-truth rules
1. Current backend/UI/public route source and git status win.
2. Repo contracts win for scoped endpoint and client-role behavior.
3. Nuvio OS cards define danger zones and validation.
4. Obsidian docs are context for product behavior.
5. Do not invent backend fields when DTO/schema support is unknown.

## Allowed work
- Audit or validate public contact and WhatsApp flows.
- Check attribution/context display and persistence.
- Check client-role Leads scope where applicable.
- Make display-only fixes if explicitly scoped and low-risk.
- Document backend DTO/schema gaps instead of inventing fields.

## Forbidden work
- Do not add unsupported public payload fields.
- Do not expose PII in logs or reports.
- Do not reintroduce raw PB writes.
- Do not change notification behavior unless explicitly scoped.
- Do not merge appointments as Leads unless current behavior requires it.

## Danger zones
- Public endpoint trust and validation.
- PII in request bodies, logs, screenshots, or reports.
- Context/source/page attribution drift.
- Raw PB writes in client-role paths.
- Notification/email side effects.

## Execution outline
1. Confirm target site, website ID, and safe test data.
2. Read Leads, Security, Emails, and Reports cards.
3. Inspect current endpoint/DTO/UI source if implementation is requested.
4. Validate contact submit, attribution/context, Leads visibility, and client-role scope if allowed.
5. Report any unavailable smoke or provider step.

## Validation checklist
### Doc validation
- Payload fields and accepted fields are documented if audited.
- Any schema/DTO uncertainty is marked unknown.
- No real PII is included.

### Code/build/test validation, if future implementation applies
- If backend changes are approved, run relevant Leads/contact tests.
- If UI changes are approved, run UI build/check.
- If public route changes are approved, run target public app validation.

### Manual smoke validation
- Public contact form submit creates a lead.
- WhatsApp tracking works if enabled.
- Backoffice Leads shows useful attribution/context.
- Client-role assigned user sees only assigned website leads.
- Reports lead summaries still load if in scope.

### User confirmation needed
- Safe test data and target website.
- Whether notifications may be sent.
- Client-role test credentials.
- Approval before backend/schema changes.

## Expected report format
- Files read.
- Files changed.
- Payload/DTO/display behavior verified.
- PII handling.
- Unknowns and risks.
- Validation run/skipped.
- Confirmation no unrelated Booking/Newsletter/Reports changes.
- Next recommended step.

## Stop conditions
- Real visitor PII would be exposed.
- The backend contract for a field is unclear and implementation would guess.
- Notification side effects cannot be safely tested.
- Client-role fixtures are unavailable for a required security conclusion.
