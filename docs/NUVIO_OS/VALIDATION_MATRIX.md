# Validation Matrix

Use this before reporting work. Do not invent commands. If a command is unavailable or cannot run in the environment, report that clearly.

## General Rules

- Validation must match the touched area.
- Build/test commands are not optional if the task asks for them and the environment supports them.
- For docs-only changes, product builds are not required unless links/scripts are changed.
- If backend/schema/auth/public endpoint behavior changes, run relevant backend tests.
- If UI code changes, run the relevant UI build/check/lint command where available.
- If validation is skipped, report why and what risk remains.

## Repo Commands Found In Audited Docs/History

| Repo/area | Command | Use when | Notes |
| --- | --- | --- | --- |
| Main backend examples | `$env:GOCACHE="$env:TEMP\nuvio-go-cache"`; `go test ./examples/base -run <TestName>` | Backend feature tests in `examples/base`. | Exact test name depends on feature. |
| Main APIs | `$env:GOCACHE="$env:TEMP\nuvio-go-cache"`; `go test ./apis -run <TestName>` | API/middleware/security tests. | `TestDoesNotExist` has been used as a no-op compile-style check in prior phases. |
| Main backoffice UI | `cd ui`; `npm run build` | Backoffice UI changes. | Check `ui/package.json` for other scripts before assuming lint/check. |
| cms5 public runtime | `npm run build` | cms5 runtime changes. | Other scripts must be checked in `package.json`. |
| Reference public template | `npm run check`; `npm run lint`; `npm run build` | Reference/template changes. | Lint may expose repo formatting state; report honestly. |
| Local Compose | `docker compose -f deploy/docker-compose.base.example.yml config` | Compose syntax validation. | Use actual file path and env setup. |
| Local Compose build/run | `docker compose -f deploy/docker-compose.base.example.yml build`; `docker compose -f deploy/docker-compose.base.example.yml up` | Local base instance validation. | Only run when Docker is available and task scope allows. |
| Backend health | `GET /api/health` | Deployment/runtime smoke. | URL depends on environment. |
| Public runtime health | `GET /` | Public runtime smoke. | A future dedicated health endpoint may replace this. |

## Main Backoffice/Backend Validation

| Change type | Required validation |
| --- | --- |
| Backend endpoint behavior | Relevant `go test ./examples/base -run <TestName>` and/or `go test ./apis -run <TestName>`. |
| Auth/client-role/security | Relevant backend tests plus manual client-role smoke. |
| Migrations/schema | Backend tests for affected feature; migration review; manual data smoke if possible. |
| CMS/assets/settings endpoints | Relevant examples/base tests; UI build if UI touched. |
| Docs-only under main repo | No product build required unless docs include executable scripts that must be checked. |

## Main UI Validation

| Change type | Required validation |
| --- | --- |
| Backoffice Svelte/CSS UI | `cd ui`; `npm run build`. |
| Shared UI primitives | Build plus manual check of affected pages. |
| Form/TinyMCE/file input work | Build plus manual form save/load/upload checks. |
| Reports/Leads/Booking/Newsletter UI | Build plus feature-specific manual smoke. |

## cms5 Validation

| Change type | Required validation |
| --- | --- |
| Runtime code | Check `package.json`, then run available build/check scripts. At minimum `npm run build` if present. |
| Public routes/sitemap/robots | Build plus route smoke for `/`, relevant page, `/sitemap.xml`, `/robots.txt`. |
| Env/Docker | Build or Docker validation when scoped. |
| Docs-only | No product build unless requested. |

## Reference Validation

| Change type | Required validation |
| --- | --- |
| Source code/routes/helpers | `npm run check`; `npm run lint`; `npm run build`. |
| Docs-only | Usually no build required unless requested. |
| Env/config boundary | Check server-only/public imports and run check/build. |
| Template CI/hardening | `npm ci` if safe, then check/lint/build. |

## Five Critical Demo Flow Smoke Checks

| Flow | Manual smoke checks |
| --- | --- |
| Website settings/setup | Open backoffice settings; confirm identity/SEO/settings save; confirm hidden settings keys preserved; confirm preview origins. |
| CMS + SEO + public rendering | Load CMS pages/blocks/assets; edit/save; preview iframe; public homepage and inner page render; title/description/social image/sitemap/robots work. |
| Leads / Contact / WhatsApp | Submit contact form; open Leads; confirm attribution/context; trigger WhatsApp redirect/tracking; verify no raw PII logs. |
| Booking | Load services/slots; submit appointment; confirm expected pending/confirmed behavior from settings; reschedule/status/archive in backoffice; check emails if enabled. |
| Reports / Analytics / Health | Load Reports dashboard; switch tabs; confirm configured/unconfigured analytics state; verify `/api/health`; check no provider secrets exposed. |

## Security / Client-Role Smoke

- Log in as assigned client-role user.
- Confirm only assigned website data appears.
- Confirm protected actions use scoped endpoints.
- Confirm no raw PocketBase writes for protected client-role product flows.
- Confirm public endpoints return clean visitor errors.
- Confirm logs redact tokens and sensitive query values.
- Confirm production-like CORS/CSP/frame origins are exact, not wildcard.

## Deployment Smoke

- Backend `/api/health` returns healthy.
- Backoffice login works.
- CMS dashboard loads.
- Public runtime returns HTTP 200.
- CMS preview iframe loads public runtime.
- Assets render after restore.
- Contact/newsletter/booking flows work if enabled.
- Reports load for deployed website.
- Browser console has no unexpected CORS/CSP/frame/mixed-content errors.
- Initial backup exists and restore path is documented.

## Email / Newsletter Smoke

- Resend/env values are server-side only.
- Public subscribe works.
- Confirm link works without leaking token in logs.
- Unsubscribe link works without leaking token in logs.
- Campaign preview is sanitized.
- Campaign send behavior is not changed by save/copy UI work.

## Analytics / Umami Smoke

- Umami provider settings are configured server-side.
- Reports traffic tab shows useful configured state or clear unavailable/setup state.
- No Umami API keys/usernames/passwords appear in browser env, UI, or logs.
- Public tracking behavior is checked only if implementation exists for the target runtime.

## Reporting Skipped Checks

When a check is skipped, report:

- command/check name;
- why it was skipped;
- what risk remains;
- closest safe alternative run, if any;
- whether user/deployment input is needed.