# Newsletter Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on newsletter subscribers, groups, subscribe/confirm/unsubscribe lifecycle, campaigns, selected recipients, Resend delivery, public links, and token/hash safety.

## 2. Current operating status

Done but needs regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines source order and boundaries. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms routing and stop conditions for this feature. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Newsletter includes tokens, emails, and public lifecycle routes. |
| 1 | Admin UI Contract | [../../NUVIO_ADMIN_UI_CONTRACT.md](../../NUVIO_ADMIN_UI_CONTRACT.md) | Defines scoped admin behavior. |
| 2 | Operating Manual Newsletter | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Newsletter.md` | Human guide to Newsletter behavior. |
| 2 | Deployment Quick Guide | [../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md) | Env guidance for public URLs and email provider setup. |
| 2 | Deployment Env Matrix | [../../NUVIO_DEPLOYMENT_ENV_MATRIX.md](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md) | Complete env reference for Resend and public base URLs. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Selects checks. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- `ui/src/components/newsletter/PageNewsletter.svelte`
- Newsletter backend endpoint files: inspect `examples/base` before changing.
- Public runtime newsletter routes/helpers in the active public site repo when lifecycle routes are targeted.
- Email provider/template helpers referenced by Newsletter code. Inspect exact files before changing.

## 5. Decisions to preserve

- Subscribe, confirm, and unsubscribe are lifecycle flows; why: each step affects consent and deliverability; agent implication: do not shortcut token/hash checks.
- Token/hash safety is a security boundary; why: unsubscribe and confirm links are public; agent implication: avoid logging tokens and avoid exposing secrets.
- `NUVIO_PUBLIC_BASE_URL` controls public lifecycle links; why: wrong values break email links; agent implication: validate env/link construction when changing email flows.
- Resend/env dependency is optional feature configuration; why: deployments may run without email provider enabled; agent implication: handle disabled/missing provider gracefully where current behavior expects it.
- Campaign send logic is sensitive; why: mistakes can send email to real subscribers; agent implication: do not change send behavior during UI polish.
- Advanced automation and reports email flows are not promised; why: current product should not oversell future automation; agent implication: do not document or implement unapproved flows.

## 6. Allowed work now

- UI polish for subscriber selection, groups, and campaign save flow.
- Reactivity fixes that preserve recipient logic.
- Copy improvements that keep behavior unchanged.
- Tests for subscribe/confirm/unsubscribe and campaign behavior when backend changes are explicit.
- Documentation clarifications around env and lifecycle links.

## 7. Do not change unless explicitly requested

- Campaign send logic.
- Token/hash algorithms or public lifecycle link contracts.
- Subscriber consent semantics.
- Resend provider behavior or env variable names.
- Public subscribe/confirm/unsubscribe endpoint contracts.
- Advanced automation or report-email features.
- Raw PB writes from UI code.

## 8. Common agent failure modes

- Using selected subscriber coverage as explicit group selection state.
- Treating a UI label change as a send-logic change.
- Exposing tokens or provider secrets in browser code/logs.
- Breaking confirm/unsubscribe links by changing base URL handling.
- Promise creep: adding copy about automation or analytics emails that do not exist.

## 9. Validation checklist

- Run `cd ui; npm run build` when UI changed.
- Run focused newsletter backend tests when backend lifecycle/campaign code changed.
- Manually check subscriber selection, group chips, selected recipient summary, save draft/update behavior, duplicate/delete, and send restrictions.
- Manually check subscribe, confirm, and unsubscribe routes if public lifecycle changed.
- Confirm `VITE_*` values do not contain secrets and provider secrets stay server-side.

## 10. Reporting requirements

- Changed files.
- Whether UI-only, backend, public runtime, tests, docs, or env docs changed.
- Campaign send logic confirmation.
- Token/hash/link behavior impact.
- Resend/env impact.
- Validation results.
- Deferred automation/report-email notes, if relevant.
