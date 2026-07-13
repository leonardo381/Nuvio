# Emails and Templates E2E Task Pack

## Purpose
Use this task pack for Resend/provider setup, contact notifications, Booking emails, Newsletter lifecycle/campaign emails, templates, public base URLs, and email side-effect validation.

## Task classification
- launch-critical
- regression
- operations
- readiness

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Emails and Templates](../features/EMAILS_TEMPLATES.md)
- [Newsletter](../features/NEWSLETTER.md)
- [Booking](../features/BOOKING.md)
- [Leads, Contact, and WhatsApp](../features/LEADS_CONTACT_WHATSAPP.md)
- [Environment and Secrets](../operations/ENV_SECRETS.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)

## Optional source docs
- [Deployment Quick Guide](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md)
- [Deployment Env Matrix](../../NUVIO_DEPLOYMENT_ENV_MATRIX.md)
- Obsidian Newsletter, Booking, Contact Form and WhatsApp, and Security Hardening docs.
- Current email/provider backend source and tests.

## Preconditions
- Email provider is configured or explicitly unavailable.
- Safe sender and recipient are known.
- Public base URL is correct for lifecycle links.
- In-scope email families are known: contact, booking, newsletter lifecycle, or campaign.
- User approved any real email send side effect.

## Source-of-truth rules
1. Current email/provider source, tests, env, and git status win.
2. Deployment env docs define variable names and server/browser boundaries.
3. Nuvio OS Emails/Env/Danger docs define safety rules.
4. Obsidian docs provide behavior context.
5. Real provider state must be verified before claiming email success.

## Allowed work
- Audit email template and provider readiness.
- Validate safe test emails only when explicitly approved.
- Check public URL/link generation.
- Document missing provider/env setup.
- Make copy/template fixes if scoped and behavior is preserved.

## Forbidden work
- Do not send real emails without approval.
- Do not expose provider secrets in browser env, docs, screenshots, or reports.
- Do not paste real lifecycle tokens or private links.
- Do not change `.ics` generation or campaign send semantics unless explicitly scoped.
- Do not use placeholder domains in live email links.

## Danger zones
- Provider secret leakage.
- Wrong public base URL in emails.
- Token exposure in lifecycle links.
- Booking `.ics` side effects.
- Campaign blast risk.
- PII in email logs/reports.

## Execution outline
1. Confirm provider mode and safe recipient.
2. Read Emails, Env, and feature docs for affected email family.
3. Inspect current provider/template source if implementation is requested.
4. Validate env names and public base URLs.
5. Run safe email smoke only with approval.
6. Report skipped sends and residual risk.

## Validation checklist
### Doc validation
- Email families in scope are listed.
- Provider/env unknowns are marked.
- No real secrets or tokens appear in report.

### Code/build/test validation, if future implementation applies
- If backend email/template code changes, run relevant backend tests.
- If UI template editor changes, run UI build/check.
- If public route/lifecycle link changes, validate target public runtime.
- If `.ics` changes, validate attachment content safely.

### Manual smoke validation
- Contact notification if enabled.
- Booking visitor/business email and `.ics` if enabled.
- Newsletter confirm/unsubscribe links if enabled.
- Campaign preview/send only if explicitly approved.
- Logs checked for token/PII redaction when safe.

### User confirmation needed
- Provider configured status.
- Safe test sender/recipient.
- Approval for email send.
- Correct public/admin URLs.
- Which templates/features are enabled.

## Expected report format
- Files read.
- Files changed.
- Email families checked.
- Provider/env/link status.
- Unknowns and risks.
- Validation run/skipped.
- Confirmation no secrets/tokens leaked.
- Next recommended step.

## Stop conditions
- No safe recipient/provider approval is available.
- Testing would email real customers.
- A fix would change campaign send or `.ics` behavior outside scope.
- Public base URL is unknown.
