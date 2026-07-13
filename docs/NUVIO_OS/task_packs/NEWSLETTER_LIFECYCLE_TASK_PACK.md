# Newsletter Lifecycle Task Pack

## Purpose
Use this task pack for Newsletter subscriber lifecycle, subscribe, confirm, unsubscribe, groups, campaigns, invites, campaign save/send boundaries, or newsletter client-role validation.

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
- [Newsletter](../features/NEWSLETTER.md)
- [Emails and Templates](../features/EMAILS_TEMPLATES.md)
- [Environment and Secrets](../operations/ENV_SECRETS.md)
- [Security and Client Role](../features/SECURITY_CLIENT_ROLE.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)

## Optional source docs
- Obsidian Newsletter and Security Hardening docs.
- Current newsletter backend source/tests.
- Current Newsletter UI source.
- Public newsletter routes/source if a public site is involved.
- Deployment/env docs for Resend and public base URLs.

## Preconditions
- Target website and safe test email are known.
- Email provider behavior is approved or disabled.
- Lifecycle token handling can be tested safely without exposing tokens.
- Task mode is audit/regression unless implementation is explicitly requested.
- Client-role credentials exist if scoped Newsletter access is checked.

## Source-of-truth rules
1. Current newsletter source, tests, provider configuration, and git status win.
2. Backend lifecycle endpoints win over UI assumptions.
3. Nuvio OS Newsletter/Emails/Env docs define safety rules.
4. Obsidian docs provide product context.
5. Token/send behavior must be observed or marked unknown.

## Allowed work
- Audit subscriber lifecycle and backoffice behavior.
- Validate subscribe/confirm/unsubscribe with safe test data if approved.
- Fix UI/save-flow polish only when behavior is preserved.
- Document provider/env gaps clearly.

## Forbidden work
- Do not expose lifecycle tokens in logs, reports, screenshots, or browser state.
- Do not put Resend or provider secrets in browser env.
- Do not change campaign send behavior during save/copy/group polish.
- Do not claim open/click automation unless implemented.
- Do not send real campaigns unless explicitly approved.

## Danger zones
- Token leakage.
- Provider side effects.
- Campaign send semantics.
- Group recipient selection and overlapping subscribers.
- Client-role website access.
- Public base URL used in lifecycle links.

## Execution outline
1. Confirm target website and safe test email.
2. Read Newsletter, Emails, Env, and Security docs.
3. Inspect current newsletter source/tests before implementation.
4. Validate subscribe/confirm/unsubscribe when approved.
5. Validate backoffice subscribers/groups/campaigns if in scope.
6. Check client-role scope if required.
7. Report provider/token unknowns.

## Validation checklist
### Doc validation
- Token handling is described without leaking token values.
- Provider/env assumptions are explicit.
- No unsupported automation claims are added.

### Code/build/test validation, if future implementation applies
- If backend newsletter changes, run relevant newsletter tests.
- If UI changes, run UI build/check.
- If public routes change, run target public app validation.
- If send/invite behavior changes, perform safe provider smoke only with approval.

### Manual smoke validation
- Subscribe creates or updates subscriber lifecycle.
- Confirm link works safely.
- Unsubscribe link works safely.
- Campaign save/update/duplicate/delete still works if touched.
- Send/invite restrictions remain unchanged unless scoped.
- Client-role access is scoped.

### User confirmation needed
- Safe test recipient.
- Provider/send approval.
- Public base URL.
- Client-role fixtures.
- Approval before changing send/lifecycle/backend behavior.

## Expected report format
- Files read.
- Files changed.
- Lifecycle and send boundaries checked.
- Unknowns and risks.
- Validation run/skipped.
- Confirmation no token/secrets leak.
- Next recommended step.

## Stop conditions
- Safe test email or provider status is unavailable.
- A task would send real campaign email without explicit approval.
- Token values would need to be printed.
- A save-flow polish task would alter send behavior.
