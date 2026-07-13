# Booking E2E Regression Task Pack

## Purpose
Use this task pack for Booking public submit, services, slots, availability, exceptions, appointment status lifecycle, reschedule, emails, `.ics`, archive, notes, or Booking client-role validation.

## Task classification
- launch-critical
- regression
- readiness

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Booking](../features/BOOKING.md)
- [Emails and Templates](../features/EMAILS_TEMPLATES.md)
- [Security and Client Role](../features/SECURITY_CLIENT_ROLE.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)

## Optional source docs
- Obsidian Booking and Security Hardening docs.
- Current Booking backend source/tests.
- Current Booking UI source.
- Public site booking route/source if a public site is involved.
- Deployment/env docs if email/provider behavior is in scope.

## Preconditions
- Target website, service, date, and safe test slot are known.
- Booking settings such as auto-confirm are known or marked unknown.
- Email sending and calendar side effects are approved or disabled.
- Task mode is regression/audit unless implementation is explicitly requested.
- Client-role credentials exist if scoped Booking access is checked.

## Source-of-truth rules
1. Current Booking source, tests, settings, and git status win.
2. Backend contract and scoped endpoints win over frontend assumptions.
3. Nuvio OS Booking and Danger Zone docs define risk.
4. Obsidian docs provide behavior context only.
5. Setting-dependent pending/confirmed behavior must be verified, not guessed.

## Allowed work
- Run a Booking regression audit or smoke plan.
- Make UI-only Booking polish if explicitly scoped.
- Fix tightly scoped Booking bugs only with matching validation.
- Document setting-dependent pending/confirmed behavior instead of forcing frontend status.

## Forbidden work
- Do not casually change slot logic, availability, exceptions, or rules.
- Do not change public payload contracts without backend confirmation.
- Do not force appointment status from the frontend unless backend contract requires it.
- Do not alter `.ics`, notifications, or email templates unless scoped.
- Do not add Booking multi-capacity during regression.

## Danger zones
- Slot availability and timezone/date logic.
- Auto-confirm and status timestamp behavior.
- Duplicate/conflict handling.
- Email and `.ics` side effects.
- Visitor PII in logs/reports.
- Client-role website access.

## Execution outline
1. Confirm target website/service/date and side-effect permissions.
2. Read Booking, Emails, Security, and Validation docs.
3. Inspect current Booking tests/source before implementation.
4. Validate public services/slots/submit.
5. Validate backoffice appointment status/reschedule/archive/notes if in scope.
6. Validate client-role scope if required.
7. Report setting-dependent behavior and unknowns.

## Validation checklist
### Doc validation
- Booking expectations and settings assumptions are stated.
- Skipped side-effect checks are explained.
- No fake pass is claimed for email/calendar behavior.

### Code/build/test validation, if future implementation applies
- If backend Booking logic changes, run relevant Booking tests.
- If Booking UI changes, run UI build/check.
- If public site changes, run target public app validation.
- If email/.ics changes, validate with safe provider/test recipient or report skipped.

### Manual smoke validation
- Public services load.
- Slots load for known service/date.
- Appointment submit creates expected status based on settings.
- Backoffice appointment operations work.
- Emails/.ics behave if enabled and approved.
- Client-role access is scoped.

### User confirmation needed
- Safe test service/date/slot.
- Auto-confirm/settings expectations.
- Email/calendar side-effect approval.
- Client-role fixtures.
- Approval before backend/schema changes.

## Expected report format
- Files read.
- Files changed.
- Booking flows checked.
- Status/slot/email impacts.
- Unknowns and risks.
- Validation run/skipped.
- Confirmation no unrelated modules changed.
- Next recommended step.

## Stop conditions
- Required test data/settings are unknown.
- Testing would send real customer emails without approval.
- A fix would require schema/migration or slot logic changes outside scope.
- Public payload behavior is unclear and would be guessed.
