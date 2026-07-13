# Booking Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on Booking services, public appointment submission, availability windows, exceptions/rules, auto-confirm behavior, reschedule/admin workflows, `.ics` emails, booking lead support, and booking-related Contacts records.

## 2. Current operating status

Done but needs regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines source order and safety rules. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms routing and stop conditions for this feature. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Booking has high logic and customer-facing risk. |
| 1 | Admin UI Contract | [../../NUVIO_ADMIN_UI_CONTRACT.md](../../NUVIO_ADMIN_UI_CONTRACT.md) | Defines scoped backoffice behavior. |
| 2 | Operating Manual Booking | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Booking.md` | Human guide to Booking behavior. |
| 2 | Security Hardening | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Security Hardening.md` | Public endpoint and client-role safety. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Validation and smoke checks. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |
| 2 | Emails and Templates Card | [EMAILS_TEMPLATES.md](EMAILS_TEMPLATES.md) | Booking emails and `.ics` are coupled. |

## 4. Likely code areas

- `ui/src/components/booking/PageBooking.svelte`
- `examples/base/nuvio_booking.go`
- `examples/base/nuvio_booking_public_test.go`
- Booking-related backend tests under `examples/base`.
- Public runtime booking route/helpers in the active public site repo only when the task targets public booking.
- Email/template helpers referenced by booking code. Inspect exact files before changing.

## 5. Decisions to preserve

- Booking is sensitive; why: small changes can create wrong slots, double bookings, or wrong customer emails; agent implication: prefer narrow fixes and regression checks.
- Multiple weekly availability windows are supported; why: businesses may have split schedules; agent implication: do not collapse availability to one window per day.
- Preserve `calendarBlockingMode`; why: it controls how external calendar conflicts block availability; agent implication: do not simplify conflict logic casually.
- Preserve service snapshots on appointments; why: historical appointments must not change when a service is edited; agent implication: do not replace snapshots with live service reads.
- Auto-confirm is settings-driven; why: public submissions should be pending unless settings explicitly auto-confirm; agent implication: do not force status from the frontend.
- Preserve status timestamps; why: operational history depends on them; agent implication: update status through existing helpers/endpoints.
- Admin reschedule flow and send-email checkbox are operational contracts; why: businesses rely on clear communication; agent implication: do not change payloads or save semantics casually.
- `.ics` emails are part of the visitor/business flow; why: calendar attachments affect real appointments; agent implication: do not modify `.ics` internals without explicit request and tests.
- Booking leads may be Contacts with `channel = booking`; why: Leads follow-up should support booking-origin contacts; agent implication: do not classify by visual label alone.
- Visitor self-manage tokens are skipped/deferred unless the roadmap changes; why: token security and UX need deliberate design; agent implication: do not invent self-manage links.

## 6. Allowed work now

- UI polish that does not change booking semantics.
- DTO validation or display fixes.
- Regression fixes for existing endpoints.
- Manual smoke documentation and focused tests.
- Reschedule modal polish if payloads and behavior stay unchanged.
- Booking lead display/follow-up fixes that use existing scoped endpoints.

## 7. Do not change unless explicitly requested

- Slot logic.
- Availability windows, exceptions, or rules.
- `calendarBlockingMode` behavior.
- Service snapshot semantics.
- Auto-confirm logic.
- Status timestamp semantics.
- Admin reschedule save semantics.
- Public booking endpoint payload contracts.
- Booking schema or migrations.
- `.ics` generation internals.
- Visitor self-manage tokens.

## 8. Common agent failure modes

- Treating a UI label issue as permission to rewrite slot generation.
- Forcing `confirmed` or `pending` from the public frontend instead of respecting backend settings.
- Removing split-day availability support.
- Editing live service data and accidentally changing historical appointment meaning.
- Changing reschedule payloads while doing visual polish.
- Forgetting to test email and `.ics` side effects.

## 9. Validation checklist

- Run `cd ui; npm run build` when UI changed.
- Run focused booking backend tests when backend booking code changed.
- Manually check service selection, date selection, slot selection, public submit, auto-confirm behavior, and admin reschedule.
- If emails changed, validate visitor/business email content and `.ics` attachment behavior.
- If Leads integration changed, confirm booking Contacts can be marked contacted without odd selection state.
- Check client-role assigned user behavior when scoped Booking access is relevant.

## 10. Reporting requirements

- Changed files.
- Whether UI-only, backend, tests, docs, or migrations changed.
- Explicit confirmation that slot logic/availability/exceptions/rules were not changed, if true.
- Auto-confirm/status behavior impact.
- `.ics` and email impact.
- Public endpoint payload impact.
- Validation results and manual smoke scope.
