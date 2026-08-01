# Danger Zones

Read this before touching sensitive areas. These rules are intentionally explicit and redundant.

| Zone | Why dangerous | What not to do | Allowed safe work | Required validation |
| --- | --- | --- | --- | --- |
| Global product scope | Nuvio is in release-readiness mode. | Do not revive broad backlog or deferred features silently. | Classify task and keep scope narrow. | Report classification and changed files. |
| Auth/client-role/websiteAccess | Can leak client data or create false security. | Do not rely on UI hiding; do not bypass websiteAccess; do not add raw PB writes. | UI polish that preserves backend checks; scoped endpoint fixes. | Backend auth/security tests; manual client-role smoke. |
| Raw PocketBase access | Raw collection writes can bypass scoped product rules. | Do not reintroduce raw PB writes for client-role product flows. | Use existing scoped endpoints; add scoped endpoint only when proven missing and allowed. | Endpoint tests and network/manual check. |
| Public endpoints | Visitor flows touch untrusted input and logs. | Do not leak stack traces, provider details, tokens, or PII. Do not accept arbitrary payloads. | Validation, bounded fields, clean error messages, redaction. | Public endpoint tests/smoke; log review if relevant. |
| Booking | Affects slots, appointment status, notifications, and trust. | Do not casually change slot logic, status defaults, public payloads, or email side effects. | UI-only polish; copy changes; tightly scoped bugfixes with tests. | Booking backend tests if logic touched; public booking E2E/manual smoke. |
| Newsletter tokens | Confirm/unsubscribe and send flows can leak tokens or email data. | Do not expose lifecycle tokens in UI/logs; do not put provider secrets in browser env; do not mix save and send behavior. | Copy/UI polish; scoped lifecycle/server fixes. | Newsletter tests; subscribe/confirm/unsubscribe smoke; log/token check. |
| Reports/analytics/PII | Reports can expose provider secrets or unsupported claims. | Do not show fake metrics; do not expose Umami credentials; do not claim unavailable analytics. | Clear setup states; DTO-backed summaries. | Reports build/tests; Umami configured/unconfigured smoke. |
| Settings/SEO | Wrong boundaries break runtime SEO and settings persistence. | Do not move SEO fields into `websites.settings`; do not overwrite hidden settings keys; do not show raw JSON. | UI clarity and scoped save fixes that preserve hidden keys. | Settings save/load; public SEO smoke if relevant. |
| SchemaForm/TinyMCE/file fields | Value sync and upload behavior are fragile. | Do not touch TinyMCE, file upload behavior, or dispatcher/parser casually. | Local UI polish with no behavior changes; targeted fixes with tests/manual checks. | UI build; manual form/file save; backend tests if endpoint touched. |
| Deployment/secrets | Env mistakes can leak secrets or break deployed URLs. | Do not put secrets in docs, `VITE_*`, or browser/client code. Do not use wildcard CORS casually. | Placeholder docs; exact origin config; server-only secret handling. | Deployment smoke; env review; health checks. |
| Snapshot/restore/reset | Restore can overwrite content or break storage references. | Do not auto-run restore on startup; do not leave `NUVIO_ALLOW_DEV_RESET=1`; do not restore without target confirmation. | Controlled one-off restore with safety flags and backups. | Restore dry run/real output; records+storage verification; backup check. |
| Migrations/schema | Can affect data and compatibility. | Do not add migrations for UI-only issues; do not rename/remove fields casually. | Non-destructive migrations when proven needed. | Backend tests; migration review; manual smoke if data path affected. |
| cms5 vs Reference | Wrong repo can contaminate the clean template or runtime. | Do not copy cms5 wholesale into Reference; do not make Reference depend on cms5. | Use cms5 for runtime history; use Reference for clean site contracts. | Check changed repo/files; public site build if touched. |
| Srcs source library | Source assets are not runtime dependencies. | Do not bulk-copy Srcs; do not runtime-import from Srcs; do not add UI libraries because a source uses them. | Copy/adapt selected assets into real site repo only. | Public site build; changed-file review. |
| Stale backlog | Old notes may describe parked ideas as desirable. | Do not treat old roadmap/backlog as active. | Mention as deferred or unknown. | Check Current Roadmap and Deferred Features. |
| Pricing/business claims | Pricing is now defined in the central Nuvio OS `TIERS_AND_PRICING.md`; do not invent new prices, discounts, remaining founder spots, or guarantees. | Do not invent exact prices or guarantees. | Use central pricing/tier authority; mark unknowns only when a newer explicit pricing decision is missing. | Manual copy review. |

## Stop Conditions

Stop and ask or report `Unknown / needs confirmation` when:

- requested work conflicts with current source-of-truth order;
- a change would require backend/schema work outside the allowed scope;
- a public/client-role path cannot be secured by existing endpoints;
- a destructive restore/reset/migration is needed but not explicitly approved;
- pricing/business facts are needed but no canonical source exists;
- docs and code disagree in a way that affects data, auth, or public behavior.
