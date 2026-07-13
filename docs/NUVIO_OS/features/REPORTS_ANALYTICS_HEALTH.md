# Reports, Analytics, and Health Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on Reports UI, analytics summaries, Umami traffic data, Nuvio custom events, business health cards, dashboard DTOs, empty states, and health-check/deployment validation.

## 2. Current operating status

Needs polish + regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines source order and boundaries. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Routes Reports vs analytics vs deployment health tasks. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Analytics can expose PII or mislead operators. |
| 2 | Operating Manual Reports | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Reports.md` | Human guide to Reports behavior. |
| 2 | Deployment Env Matrix | [../../NUVIO_DEPLOYMENT_ENV_MATRIX.md](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md) | Umami and health-related env reference. |
| 2 | Coolify Plan | [../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md](../../NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md) | Health checks and deployment smoke context. |
| 2 | Current Operating State | [../CURRENT_OPERATING_STATE.md](../CURRENT_OPERATING_STATE.md) | Current readiness and known launch posture. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Defines checks and smoke tests. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- `ui/src/components/reports/PageReports.svelte`
- Reports backend/dashboard endpoint files: inspect `examples/base` before changing.
- Analytics provider helpers for Umami: inspect exact files before changing.
- Deployment health endpoint, likely `/api/health`, only when task targets health/deployment.

## 5. Decisions to preserve

- Reports should be client-friendly; why: clients need practical business status, not raw analytics dashboards; agent implication: use clear summaries and avoid provider jargon.
- Umami is for traffic analytics; why: it tracks visits and pages; agent implication: do not treat it as the source for all business actions.
- Nuvio custom events cover business actions; why: leads, bookings, newsletter, and WhatsApp have product-specific meaning; agent implication: keep business metrics grounded in Nuvio DTOs/events.
- No PII in analytics; why: privacy and operational safety; agent implication: never send names, emails, phones, messages, or tokens into analytics events.
- Health checks are deployment validation, not product analytics; why: they prove service availability; agent implication: keep `/api/health` simple and safe.
- Reports need confidence in empty/demo states; why: launch instances may have sparse data; agent implication: avoid fake insights and label empty states honestly.
- Raw analytics dashboard direction is not desired; why: Nuvio should interpret the data for small businesses; agent implication: do not embed provider dashboards as the main Reports UX.

## 6. Allowed work now

- UI layout polish and spacing normalization.
- Frontend-only calculations from existing DTO fields.
- Empty-state and helper-copy improvements.
- Backend DTO fixes only when a real contract gap is proven.
- Tests for scoped Reports endpoints when backend changes are made.
- Health-check documentation and smoke checklist updates.

## 7. Do not change unless explicitly requested

- Reports endpoint contracts.
- Report calculations beyond the requested scope.
- Analytics provider credentials/env names.
- PII handling rules.
- Raw provider dashboard embedding as a substitute for Nuvio Reports.
- Public runtime tracking behavior.
- CMS, Leads, Newsletter, Booking logic while doing Reports polish.

## 8. Common agent failure modes

- Inventing insights not supported by the DTO.
- Treating current-state SEO/page data as period-filtered without proof.
- Sending PII to analytics for convenience.
- Adding chart libraries for simple report rows.
- Changing backend endpoints to fix a spacing/layout issue.
- Over-polishing Reports into a separate visual system that no longer matches Nuvio.

## 9. Validation checklist

- Run `cd ui; npm run build` when Reports UI changed.
- Run focused Reports backend tests when backend code changes. Inspect exact test names first.
- Manually switch every Reports tab.
- Manually change website/period filters and refresh.
- Confirm empty states do not claim fake insights.
- Confirm no PII is exposed in analytics, browser logs, or report summaries.
- Check `/api/health` only when health/deployment behavior changed.

## 10. Reporting requirements

- Changed files.
- Whether UI-only, frontend calculation, backend DTO, analytics, docs, or health behavior changed.
- Data source for every new metric or section.
- PII handling confirmation.
- Empty/demo-state behavior.
- Validation results and manual tab smoke scope.
