# Emails and Templates Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on contact form notifications, WhatsApp notifications, booking visitor/business emails, `.ics` attachments, newsletter confirmation/unsubscribe/campaign emails, Resend configuration, public base URLs, and template safety.

## 2. Current operating status

Needs polish + regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines source order and boundaries. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms routing and stop conditions for this feature. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Emails include tokens, secrets, and customer-facing content. |
| 2 | Contact Form and WhatsApp | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Contact Form and WhatsApp.md` | Contact/WhatsApp notification behavior. |
| 2 | Booking | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Booking.md` | Booking visitor/business emails and `.ics` context. |
| 2 | Newsletter | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Newsletter.md` | Subscribe, confirm, unsubscribe, and campaign lifecycle. |
| 2 | Deployment Quick Guide | [../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md) | Required public base URL and Resend env guidance. |
| 2 | Deployment Env Matrix | [../../NUVIO_DEPLOYMENT_ENV_MATRIX.md](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md) | Full email/provider env reference. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Selects tests and manual checks. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- Contact/WhatsApp notification backend files under `examples/base`. Inspect exact files before changing.
- Booking backend email/template helpers referenced by `examples/base/nuvio_booking.go`.
- Newsletter backend email/template/campaign files under `examples/base`. Inspect exact files before changing.
- Deployment env docs when provider variables or public URL guidance changes.
- Public runtime lifecycle routes only when email links or target routes are explicitly in scope.

## 5. Decisions to preserve

- Contact form notifications are operational alerts; why: businesses rely on them to respond; agent implication: preserve recipient, subject, and lead context behavior unless requested.
- WhatsApp notification behavior is separate from WhatsApp redirects/tracking; why: tracking and alerts have different risks; agent implication: inspect which flow is being changed.
- Booking has visitor and business emails; why: both sides need different content; agent implication: do not collapse templates into one generic message.
- `.ics` attachments are part of booking email behavior; why: calendar clients consume structured data; agent implication: do not edit `.ics` internals without explicit tests.
- Newsletter confirmation/unsubscribe/campaign emails include token/hash safety; why: links are public and consent-related; agent implication: avoid logging or exposing tokens.
- Resend env is server-side; why: provider keys are secrets; agent implication: do not put Resend keys in `VITE_*`, `PUBLIC_*`, UI code, or docs.
- `NUVIO_PUBLIC_BASE_URL` matters for public links; why: wrong base URLs break email actions; agent implication: validate generated links after env changes.
- Do not invent Reports email flow; why: it is not part of the current operating contract; agent implication: avoid adding copy or code that promises report emails.

## 6. Allowed work now

- Copy/template polish that preserves variables and behavior.
- Bug fixes for already-supported email flows.
- Resend/env documentation fixes.
- Tests for email dispatch/link generation when backend email code changes.
- Manual smoke guidance for contact, booking, and newsletter lifecycle emails.

## 7. Do not change unless explicitly requested

- Provider env variable names.
- Token/hash generation or validation.
- `.ics` generation internals.
- Newsletter campaign send semantics.
- Contact/WhatsApp/booking public endpoint contracts.
- Reports email flows.
- Secrets handling model.

## 8. Common agent failure modes

- Sending or testing against real recipients unintentionally.
- Logging tokens, unsubscribe links, API keys, or visitor PII.
- Breaking links by using the admin/backend URL instead of `NUVIO_PUBLIC_BASE_URL`.
- Editing `.ics` output as plain text without understanding calendar format requirements.
- Adding new template variables that are not populated by backend DTOs.
- Documenting provider aliases as preferred when `NUVIO_*` names should be used for new deployments.

## 9. Validation checklist

- Run focused backend tests for the email flow when backend code changes.
- For UI-only template selection/copy, run the relevant UI build if UI changed.
- Manually test contact form notification if contact email behavior changed.
- Manually test booking visitor/business emails and `.ics` if booking email behavior changed.
- Manually test newsletter confirm/unsubscribe/campaign preview/send only in a controlled environment.
- Confirm provider secrets are server-only and no real secrets are committed.
- Confirm public links use the intended public site URL.

## 10. Reporting requirements

- Changed files.
- Email flow affected: contact, WhatsApp, booking, newsletter, or docs only.
- Provider/env impact.
- Public link/base URL behavior.
- Token/hash or `.ics` impact.
- Confirmation that Reports email flow was not invented.
- Validation results and manual smoke scope.
