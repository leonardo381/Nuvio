# Nuvio OS Core Pack QA

Date: 2026-06-17

## 1. Files reviewed

- `docs/NUVIO_OS/README.md`
- `docs/NUVIO_OS/CORE.md`
- `docs/NUVIO_OS/SOURCE_OF_TRUTH.md`
- `docs/NUVIO_OS/CURRENT_OPERATING_STATE.md`
- `docs/NUVIO_OS/CANONICAL_DECISIONS.md`
- `docs/NUVIO_OS/TASK_ROUTER.md`
- `docs/NUVIO_OS/DANGER_ZONES.md`
- `docs/NUVIO_OS/VALIDATION_MATRIX.md`
- `docs/NUVIO_OS/REPORTING_FORMATS.md`

Reference inputs checked:

- `docs/NUVIO_OS/audits/2026-06-17_SOURCE_INVENTORY.md`
- `docs/NUVIO_OS/audits/2026-06-17_CONFLICT_STALENESS_AUDIT.md`

## 2. Overall verdict

Pass with minor fixes.

The Core Pack is usable for fresh agents. It clearly communicates release-readiness mode, source-of-truth order, cms5 vs Reference boundaries, stale docs caution, task routing, danger zones, validation expectations, and reporting expectations.

The only hardening issue found was in `REPORTING_FORMATS.md`: several templates had the right intent but did not explicitly include every required reporting field. That file was patched.

## 3. Issues found

| Severity | File | Issue | Fix applied? | Notes |
| --- | --- | --- | --- | --- |
| Low | `docs/NUVIO_OS/REPORTING_FORMATS.md` | Some templates did not explicitly include all required fields: decisions used, tests/builds run, manual checks, risks, blockers/unknowns, and what was not changed. | Yes | Rewrote the templates to include the full required field set consistently. |
| Low | Core Pack overall | `CANONICAL_DECISIONS.md`, `TASK_ROUTER.md`, and `DANGER_ZONES.md` are table-heavy and intentionally dense. | No | Acceptable for agent use. Future Phase 4 feature cards can make module-specific routing easier. |
| Low | Core Pack overall | External Hermes/ChatGPT context and Plan B/business plan remain unknown because no durable local files were found in Phase 1/2. | No | Correctly marked as unknown; no local source to patch against. |

## 4. Fixes applied

Patched:

- `docs/NUVIO_OS/REPORTING_FORMATS.md`

Summary:

- Added a universal required-field list at the top.
- Updated every template to explicitly include:
  - goal;
  - files read;
  - files changed;
  - decisions used;
  - tests/builds run;
  - manual checks;
  - risks;
  - blockers/unknowns;
  - next recommended step;
  - what was not changed.
- Added extra reporting rules for danger-zone tasks, skipped validation, no-file-change reports, source conflicts, and unavailable business/Hermes context.

No product code was changed.

## 5. Remaining unknowns

- Durable Hermes/ChatGPT consolidated context files were not found locally.
- Plan B/business plan files were not found locally.
- Final pricing remains unknown and must not be invented.
- Exact Coolify/deployment state still needs confirmation during real deploy work.
- Feature-level validation commands remain task-specific and must be checked against current package scripts/tests at implementation time.

## 6. Readiness for Phase 4

The Core Pack is ready for Phase 4.

Phase 4 can safely create feature/operations cards because:

- source hierarchy is explicit;
- current operating mode is explicit;
- canonical decisions are centralized;
- task routing exists;
- danger zones are explicit;
- validation expectations are defined;
- reporting formats are now complete enough for repeatable agent output.

## 7. Recommended Phase 4 scope

Create feature and operations cards only. Do not rewrite the Core Pack unless a real contradiction is found.

Recommended feature cards:

- `docs/NUVIO_OS/features/CMS.md`
- `docs/NUVIO_OS/features/ASSETS.md`
- `docs/NUVIO_OS/features/WEBSITE_SETTINGS_SEO.md`
- `docs/NUVIO_OS/features/LEADS_CONTACT_WHATSAPP.md`
- `docs/NUVIO_OS/features/BOOKING.md`
- `docs/NUVIO_OS/features/NEWSLETTER.md`
- `docs/NUVIO_OS/features/REPORTS_ANALYTICS.md`
- `docs/NUVIO_OS/features/PUBLIC_RUNTIME.md`
- `docs/NUVIO_OS/features/SECURITY_CLIENT_ROLE.md`

Recommended operations cards:

- `docs/NUVIO_OS/operations/DEPLOYMENT_COOLIFY.md`
- `docs/NUVIO_OS/operations/INSTANCE_BOOTSTRAP.md`
- `docs/NUVIO_OS/operations/SNAPSHOT_RESTORE.md`
- `docs/NUVIO_OS/operations/RELEASE_READINESS.md`
- `docs/NUVIO_OS/operations/FIRST_CLIENT_READINESS.md`

Each card should be agent-focused and should include:

- canonical sources;
- current status;
- do-not-change rules;
- common task routing;
- validation requirements;
- danger zones;
- reporting notes.