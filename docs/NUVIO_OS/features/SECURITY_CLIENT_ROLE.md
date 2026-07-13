# Security and Client Role Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on admin/client-role behavior, website access scoping, scoped endpoints, public endpoint threat boundaries, direct unassigned access denial, restore restrictions, raw PocketBase risks, and secrets hygiene.

## 2. Current operating status

Done but needs regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines source order and project safety. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms routing and stop conditions for this feature. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Security-sensitive areas and stop conditions. |
| 1 | Admin UI Contract | [../../NUVIO_ADMIN_UI_CONTRACT.md](../../NUVIO_ADMIN_UI_CONTRACT.md) | Defines scoped admin UI and endpoint expectations. |
| 2 | Operating Manual Security Hardening | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Security Hardening.md` | Human security baseline and do-not-regress list. |
| 2 | Instance Model | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\03 - Operations\Instance Model.md` | Explains clients as instances. |
| 2 | Deployment Quick Guide | [../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md](../../NUVIO_DEPLOYMENT_QUICK_GUIDE.md) | Env and secrets handling basics. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Defines client-role smoke tests and checks. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- `apis/*` for reusable API helpers and auth utilities. Inspect exact files before changing.
- `examples/base/*` scoped endpoints and tests.
- `ui/src/*` only when UI role visibility or scoped client behavior is targeted.
- `pb_migrations/*` only when schema/rules work is explicitly requested.
- Deployment docs/env files when secrets hygiene or instance setup is targeted.

## 5. Decisions to preserve

- Admin and client-role are different trust levels; why: clients must not get global or unassigned access; agent implication: always validate role-specific behavior.
- Website access is scoped by assignment; why: one backend can contain multiple websites/instances; agent implication: use `RequireWebsiteAccessById` or equivalent scoped checks where applicable.
- Direct unassigned access must be denied; why: unassigned clients should not infer or mutate data; agent implication: test denied paths, not only happy paths.
- Scoped endpoints are preferred over raw PB access; why: scoped endpoints enforce business authorization and DTO shaping; agent implication: do not reintroduce raw PB UI writes.
- Public endpoints have a separate threat model; why: they are unauthenticated and internet-facing; agent implication: validate input and avoid leaking internals.
- Restore/reset actions are destructive; why: they can overwrite website content; agent implication: restrict, label, and confirm before implementation.
- Secrets belong server-side only; why: browser env and logs can leak; agent implication: never put provider secrets in `VITE_*`, `PUBLIC_*`, UI code, or docs.

## 6. Allowed work now

- Scoped endpoint bug fixes.
- Client-role visibility and access fixes.
- UI hiding/disabled-state polish that mirrors backend restrictions.
- Public endpoint validation hardening.
- Tests for assigned/unassigned client behavior.
- Documentation and runbook clarifications.

## 7. Do not change unless explicitly requested

- PocketBase rules or migrations.
- Auth model or role semantics.
- Raw collection access from UI.
- Restore/reset permissions.
- Public endpoint payload contracts.
- Env variable names or secret handling model.
- Cross-feature permissions while doing a single feature polish task.

## 8. Common agent failure modes

- Hiding a UI button but leaving the backend endpoint open.
- Adding a backend endpoint without checking client-role and website access.
- Testing only admin behavior and missing client-role regressions.
- Treating public endpoints like authenticated admin endpoints.
- Committing real env values or examples that look like real secrets.
- Using raw PB calls for a quick client-side fix.

## 9. Validation checklist

- Run focused backend tests for the affected endpoint or feature.
- Run `cd ui; npm run build` when UI role behavior changed.
- Manually test admin user, assigned client-role user, and unassigned client-role denial when access rules changed.
- Confirm public endpoints still reject invalid or unsafe input.
- Confirm secrets are not in source, docs, browser env, or logs.
- Confirm restore/reset actions are not newly exposed to client-role unless explicitly designed.

## 10. Reporting requirements

- Changed files.
- Auth/scoping checks preserved or changed.
- Admin/client/unassigned behavior.
- Public endpoint threat-model impact.
- Raw PB access confirmation.
- Secrets/env exposure confirmation.
- Validation results.
