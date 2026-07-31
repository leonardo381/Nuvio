# Leads, Contact Form, and WhatsApp Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on the unified Leads experience, public contact form submissions, WhatsApp tracking/contact flows, lead attribution/context, notifications, and client-safe scoped lead management.

## 2. Current operating status

Done but needs regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines project rules and source order. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms routing and stop conditions for this feature. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Leads touch public endpoints and client data. |
| 1 | Admin UI Contract | [../../NUVIO_ADMIN_UI_CONTRACT.md](../../NUVIO_ADMIN_UI_CONTRACT.md) | Defines scoped endpoint and client-role requirements. |
| 2 | Operating Manual Leads | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Leads.md` | Human guide to the Leads module. |
| 2 | Contact Form and WhatsApp | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Contact Form and WhatsApp.md` | Explains public contact and WhatsApp flows. |
| 2 | Security Hardening | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Security Hardening.md` | Explains public endpoint and scoped access expectations. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Selects validation commands. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- `ui/src/components/leads/PageLeads.svelte`
- `examples/base/nuvio_leads_notifications.go`
- `examples/base/nuvio_leads_dashboard.go`
- `examples/base/nuvio_leads_notifications_test.go`
- `examples/base/nuvio_leads_dashboard_test.go`
- Public website contact helpers in the active public site repo only when the task targets public form behavior.
- WhatsApp backend/public route files: inspect exact names before changing.

## 5. Decisions to preserve

- Preserve the unified Leads model; why: contact form, WhatsApp, and booking-related contacts share one operator workflow; agent implication: do not split flows into unrelated dashboards.
- Keep `settings.contactForm` compatibility; why: existing public sites and settings depend on it; agent implication: do not replace with only a new nested model.
- Keep `settings.whatsapp` compatibility; why: WhatsApp public links and notifications may read it; agent implication: do not prematurely migrate to `settings.leads.channels`.
- Harden public endpoints separately from admin endpoints; why: public submissions are unauthenticated threat surfaces; agent implication: validate input, rate-limit public submits, and avoid exposing internal data.
- Notification templates are part of the flow; why: submissions often trigger business emails; agent implication: do not change notification payloads casually.
- Attribution and context should be human-readable; why: clients need useful origin data; agent implication: filter placeholders and technical keys in UI display.

## 6. Allowed work now

- UI polish for Leads list/detail/bulk actions.
- Display fallback fixes for origin, attribution, and context.
- Scoped endpoint bug fixes.
- Public contact/WhatsApp validation fixes based on existing fields.
- Tests for public submit, dashboard DTO, follow-up, read/archive, and client scoping.
- Notification template fixes that preserve existing behavior.

## 7. Do not change unless explicitly requested

- Do not migrate settings to `settings.leads.channels` prematurely.
- Do not add unsupported public payload fields.
- Do not weaken public endpoint validation.
- Do not expose PII in analytics or logs. The public contact submit rate limiter must key by the logical contact-submit route + `RequestEvent.RealIP()` and return a generic 429 before saving records or sending notifications.
- Do not change Booking, Newsletter, Reports, CMS, or public runtime routing as part of a Leads polish task.
- Do not reintroduce raw PB writes from the UI.

## 8. Common agent failure modes

- Treating visual source labels as storage contracts.
- Showing raw technical values such as `reference_contact` to clients.
- Joining placeholder context parts into strings like `N/A · /contact`.
- Fixing public form behavior by adding unsupported fields instead of checking backend accepted fields.
- Forgetting booking leads may be contact records with `channel = booking`.
- Changing notification behavior without validating existing tests.

## 9. Validation checklist

- Run `cd ui; npm run build` when Leads UI changed.
- Run focused Leads/contact backend tests when backend code changed, including the public contact submit rate-limit regression.
- Manually submit a public contact form and confirm a lead appears.
- Manually test WhatsApp tracking/contact flow if touched.
- Check client-role assigned user can see only scoped leads.
- Confirm no partial placeholder values appear in Origin, Attribution, or Context.

## 10. Reporting requirements

- Changed files.
- Whether UI-only or backend/public endpoint changed.
- Public payload fields accepted and persisted.
- Compatibility with `settings.contactForm` and `settings.whatsapp`.
- Notification behavior impact.
- Client-role/scoped access validation.
- Validation results.
