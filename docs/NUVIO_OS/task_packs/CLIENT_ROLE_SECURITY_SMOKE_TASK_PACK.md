# Client Role Security Smoke Task Pack

## Purpose
Use this task pack when validating client-role access, website scoping, raw PocketBase avoidance, public endpoint safety, or permission-sensitive backoffice behavior.

## Task classification
- security
- regression
- launch-critical
- readiness

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Security and Client Role](../features/SECURITY_CLIENT_ROLE.md)
- [Smoke Validation and Troubleshooting](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)
- [CMS](../features/CMS.md)
- [Leads, Contact, and WhatsApp](../features/LEADS_CONTACT_WHATSAPP.md)
- [Booking](../features/BOOKING.md)
- [Newsletter](../features/NEWSLETTER.md)
- [Reports, Analytics, and Health](../features/REPORTS_ANALYTICS_HEALTH.md)

## Optional source docs
- Obsidian Security Hardening.
- Obsidian Instance Model.
- Current backend middleware, website access helpers, and endpoint tests.
- Current UI client-role source and route guards.
- Deployment/env docs for CORS/CSP if public origins are involved.

## Preconditions
- Client-role user and assigned website are known.
- Unassigned website or negative-test path is available.
- Task is smoke/audit or implementation is explicitly scoped.
- Allowed checks will not expose real client PII.
- Current source and git status can be inspected.

## Source-of-truth rules
1. Current auth/source/tests and git status win.
2. Repo contracts define security expectations.
3. Nuvio OS danger zones and security card define guardrails.
4. Obsidian Security Hardening is context, not proof.
5. UI behavior never proves backend authorization by itself.

## Allowed work
- Audit client-role UI and scoped endpoint behavior.
- Run or plan smoke checks for admin, assigned client, and unassigned client.
- Inspect network/API paths for raw collection writes when safe.
- Document missing tests or unavailable fixtures.

## Forbidden work
- Do not rely on UI hiding as security.
- Do not add raw PB client writes for client-role product flows.
- Do not widen collection rules or backend access casually.
- Do not paste credentials, tokens, or PII in reports.
- Do not change auth/session behavior unless explicitly requested.

## Danger zones
- Cross-website access.
- Raw `/api/collections/*` writes from client-role surfaces.
- PII or token leakage in logs, UI, network, or reports.
- UI-only filtering mistaken for backend authorization.
- Provider secrets visible in browser responses.
- Allowed client-role pages failing with raw PB 403s because they depend on raw collection APIs instead of scoped endpoints.

## Execution outline
1. Confirm user roles, website assignments, and test environment.
2. Read Security card plus affected feature cards.
3. Identify scoped endpoints used by the feature.
4. Check admin access, assigned client access, and unassigned client denial where possible.
5. Check network paths for raw PB writes where practical.
6. Report results without changing permission behavior unless scoped.

## Validation checklist
### Doc validation
- Sources and inspected feature areas are listed.
- Unknown fixture/credential gaps are marked.
- No real secrets or PII are included in the report.

### Code/build/test validation, if future implementation applies
- If code changes are approved, run relevant backend auth/security tests and UI build/check for touched UI.
- If no tests exist, report the gap and closest safe smoke.

### Manual smoke validation
- Admin can access intended data.
- Assigned client can access only assigned website data.
- Unassigned client cannot access other website data.
- No raw PB writes on client-role product paths.
- Allowed client-role pages load without raw PB 403s on supported scoped flows.
- Public endpoint errors are visitor-safe.

### User confirmation needed
- Client-role credentials and test website assignments.
- Whether negative tests can be performed.
- Approval before changing backend auth/rules.
- Approval before inspecting live logs containing visitor data.

## Expected report format
- Files read.
- Files changed.
- Users/scopes tested or unavailable.
- Findings by severity.
- What was verified.
- Unknowns and risks.
- Validation run/skipped.
- Confirmation no security behavior changed unless scoped.
- Next recommended step.

## Stop conditions
- No safe test user/website assignment is available.
- Testing would expose real PII or secrets.
- A proposed fix would rely only on UI hiding.
- A change would widen backend access rules without explicit approval.
- An allowed page requires raw collection access to function.


