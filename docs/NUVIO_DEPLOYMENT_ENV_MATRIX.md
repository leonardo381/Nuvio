# Nuvio Deployment Environment Matrix

## Overview

This document lists the environment variables and deployment knobs needed to run a Nuvio instance. Nuvio currently uses one PocketBase/backoffice instance per client, with separate runtime data, storage, domains, backups, and CMS snapshot/bootstrap data per instance.

Use this matrix when provisioning:

- Nuvio's own website.
- A new client website instance.
- A staging or QA copy of a client instance.

## Rules

- Never commit real `.env` files.
- Never paste real secret values into docs, tickets, commits, screenshots, or logs.
- Store secrets in the deployment provider secret store or another secrets manager.
- Each client instance must have its own `pb_data`, storage files, backups, domains, origins, and CMS snapshot/bootstrap data.
- `VITE_*` variables are exposed to the browser bundle. Never put secrets in `VITE_*`.
- Prefer `NUVIO_*` names for Nuvio-specific server variables.
- Keep production CORS, CSP, and preview frame origins explicit. Do not rely on localhost or wildcard behavior.
- Treat API keys, service account passwords, email provider keys, analytics credentials, and optional encryption keys as secrets.

## Backend/Server Env Variables

These variables belong to the backend/backoffice repo at `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio`.

| Variable | Runtime side | Secret level | Required | Scope | Example placeholder | Used for | Referenced in |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `PB_URL` | Server-only/tools | Public or sensitive URL | Optional | Per instance | `https://api.client.example.com` | Backend/API base URL for QA tools and public runtime fallback. | `.env.example`, `tools/dev/seed_operational_qa_data.go`, public runtime server helpers |
| `PB_SUPERUSER_EMAIL` | Server-only/tools | Secret | Required for QA/dev tooling and bootstrap automation | Per instance | `admin@client.example.com` | Superuser login for internal QA seed/bootstrap tools. | `.env.example`, `tools/dev/seed_operational_qa_data.go` |
| `PB_SUPERUSER_PASSWORD` | Server-only/tools | Secret | Required for QA/dev tooling and bootstrap automation | Per instance | `<secret-superuser-password>` | Superuser password for internal QA seed/bootstrap tools. | `.env.example`, `tools/dev/seed_operational_qa_data.go` |
| `PB_SERVICE_EMAIL` | Server-only | Secret | Optional/reserved | Per instance | `service@client.example.com` | Service auth account placeholder. No direct code reference found in this audit. | `.env.example` |
| `PB_SERVICE_PASSWORD` | Server-only | Secret | Optional/reserved | Per instance | `<secret-service-password>` | Service auth password placeholder. No direct code reference found in this audit. | `.env.example` |
| `NUVIO_PUBLIC_BASE_URL` | Server-only | Public URL | Required if newsletter lifecycle emails are enabled | Per instance | `https://www.client.example.com` | Visitor-facing base URL for newsletter confirmation/unsubscribe links. | `.env.example`, `examples/base/nuvio_newsletter.go` |
| `NUVIO_CORS_ALLOWED_ORIGINS` | Server-only | Public origins | Required for production | Per instance | `https://admin.client.example.com https://www.client.example.com` | Explicit CORS origin allowlist. | `.env.example`, `apis/serve.go`, `cmd/serve.go` |
| `NUVIO_CMS_PREVIEW_FRAME_SRC` | Server-only | Public origins | Required when CMS preview uses a separate public runtime origin | Per instance | `https://www.client.example.com` | Backoffice CSP `frame-src` and `child-src` preview allowlist. | `.env.example`, `apis/serve.go` |
| `NUVIO_CONTACT_SUBMIT_RATE_LIMIT_MAX` | Server-only | Operational | Optional | Per instance | `5` | Maximum unauthenticated public contact submissions per contact-submit route + client IP within the configured window. | `.env.example`, `examples/base/nuvio_leads_notifications.go` |
| `NUVIO_CONTACT_SUBMIT_RATE_LIMIT_WINDOW_SECONDS` | Server-only | Operational | Optional | Per instance | `60` | Contact submit rate-limit window in seconds. | `.env.example`, `examples/base/nuvio_leads_notifications.go` |
| `NUVIO_RESEND_API_KEY` | Server-only | Secret | Required if Resend email sending is enabled | Per instance or provider account | `<secret-resend-api-key>` | Primary Resend API key. | `.env.example`, `examples/base/nuvio_email.go` |
| `RESEND_API_KEY` | Server-only | Secret | Optional fallback | Per instance or provider account | `<secret-resend-api-key>` | Resend API key fallback. Prefer `NUVIO_RESEND_API_KEY`. | `.env.example`, `examples/base/nuvio_email.go` |
| `NUVIO_RESEND_FROM` | Server-only | Public/sensitive email | Optional | Per instance | `Nuvio <hello@client.example.com>` | Primary email sender identity. | `.env.example`, `examples/base/nuvio_email.go` |
| `RESEND_FROM` | Server-only | Public/sensitive email | Optional fallback | Per instance | `Nuvio <hello@client.example.com>` | Email sender fallback. Prefer `NUVIO_RESEND_FROM`. | `.env.example`, `examples/base/nuvio_email.go` |
| `NUVIO_RESEND_FROM_EMAIL` | Server-only | Public/sensitive email | Optional fallback | Per instance | `hello@client.example.com` | Email sender fallback. | `.env.example`, `examples/base/nuvio_email.go` |
| `RESEND_FROM_EMAIL` | Server-only | Public/sensitive email | Optional fallback | Per instance | `hello@client.example.com` | Email sender fallback. Prefer `NUVIO_RESEND_FROM` or `NUVIO_RESEND_FROM_EMAIL`. | `.env.example`, `examples/base/nuvio_email.go` |
| `NUVIO_UMAMI_API_URL` | Server-only | Public/sensitive URL | Required if Reports analytics integration is enabled | Per analytics instance | `https://umami.example.com/api` | Umami API base URL. | `.env.example`, `examples/base/nuvio_reports.go` |
| `NUVIO_UMAMI_API_KEY` | Server-only | Secret | Required for API-key Umami auth | Per analytics account | `<secret-umami-api-key>` | Umami API key. | `.env.example`, `examples/base/nuvio_reports.go` |
| `NUVIO_UMAMI_USERNAME` | Server-only | Secret | Optional, depending on Umami auth mode | Per analytics account | `<secret-umami-username>` | Umami username auth. | `.env.example`, `examples/base/nuvio_reports.go` |
| `NUVIO_UMAMI_PASSWORD` | Server-only | Secret | Optional, depending on Umami auth mode | Per analytics account | `<secret-umami-password>` | Umami password auth. | `.env.example`, `examples/base/nuvio_reports.go` |
| `NUVIO_GOOGLE_PLACES_API_KEY` | Server-only | Secret | Required if Google Places reviews sync is enabled | Per Google project/client | `<secret-google-places-key>` | Primary Google Places API key. | `.env.example`, `examples/base/nuvio_reviews.go` |
| `GOOGLE_PLACES_API_KEY` | Server-only | Secret | Optional fallback | Per Google project/client | `<secret-google-places-key>` | Google Places API key fallback. Prefer `NUVIO_GOOGLE_PLACES_API_KEY`. | `.env.example`, `examples/base/nuvio_reviews.go` |
| `NUVIO_QA_BASE_URL` | Server-only/tools | Public/sensitive URL | Dev/QA only | Per QA instance | `https://qa-api.client.example.com` | QA tool base URL override. | `.env.example`, `tools/dev/seed_operational_qa_data.go` |
| `NUVIO_ALLOW_DEV_RESET` | Server-only/tools | Sensitive safety flag | Dev/QA only | Per local/QA run | `0` | Explicit opt-in for destructive QA snapshot/reset tooling. Use `1` only for controlled local/QA flows. | `.env.example`, `tools/dev/*qa*`, `tools/dev/cmsqasnapshot/cms_qa_snapshot.go` |

## Backoffice UI Env Variables

These variables belong to `ui` in the backend/backoffice repo. They are browser-exposed because `ui/vite.config.js` uses `envPrefix: ['VITE_']`.

| Variable | Runtime side | Secret level | Required | Scope | Example placeholder | Used for | Referenced in |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `VITE_PB_BACKEND_URL` | Browser-exposed | Public URL | Required | Per instance | `https://api.client.example.com` | PocketBase/API base URL for the backoffice UI and scoped asset calls. | `ui/.env.example`, `ui/src/utils/ApiClient.js`, CMS, newsletter, record confirmation flows, `InputFile.svelte` |
| `VITE_PUBLIC_SITE_BASE_URL` | Browser-exposed | Public URL | Optional but recommended for CMS preview links | Per instance | `https://www.client.example.com` | Preferred public site base URL for CMS preview. | `ui/src/components/cms/PageCms.svelte` |
| `VITE_SITE_BASE_URL` | Browser-exposed | Public URL | Optional legacy fallback | Per instance | `https://www.client.example.com` | Legacy/fallback public site base URL for CMS preview. Prefer `VITE_PUBLIC_SITE_BASE_URL`. | `ui/src/components/cms/PageCms.svelte` |
| `VITE_PB_DOCS_URL` | Browser-exposed | Public URL | Optional | Global | `https://pocketbase.io/docs` | Documentation links in admin UI. | `ui/.env.example`, `PageWrapper.svelte`, `PageCrons.svelte` |
| `VITE_PB_RELEASES` | Browser-exposed | Public URL | Optional | Global | `https://github.com/pocketbase/pocketbase/releases` | Release link in admin UI. | `ui/.env.example`, `PageWrapper.svelte` |
| `VITE_PB_VERSION` | Browser-exposed | Public text | Optional | Code version/build | `dev` | Displayed PocketBase/Nuvio base version. | `ui/.env.example`, `PageWrapper.svelte` |
| `VITE_PB_JS_SDK_URL` | Browser-exposed | Public URL | Optional | Global | `https://github.com/pocketbase/js-sdk` | JS SDK documentation link. | `ui/.env.example`, `SdkTabs.svelte` |
| `VITE_PB_DART_SDK_URL` | Browser-exposed | Public URL | Optional | Global | `https://github.com/pocketbase/dart-sdk` | Dart SDK documentation link. | `ui/.env.example`, `SdkTabs.svelte` |
| `VITE_PB_RULES_SYNTAX_DOCS` | Browser-exposed | Public URL | Optional | Global | `https://pocketbase.io/docs/api-rules-and-filters` | API rule syntax documentation link. | `ui/.env.example`, `CollectionRulesTab.svelte` |
| `VITE_PB_OAUTH2_EXAMPLE` | Browser-exposed | Public URL | Optional | Global | `https://github.com/pocketbase/pocketbase/discussions` | OAuth2 example link. | `ui/.env.example`, `AuthWithOAuth2Docs.svelte` |
| `VITE_PB_FILE_UPLOAD_DOCS` | Browser-exposed | Public URL | Optional | Global | `https://pocketbase.io/docs/files-handling` | File upload docs links. | `ui/.env.example`, collection docs |
| `VITE_PB_MFA_DOCS` | Browser-exposed | Public URL | Optional | Global | `https://pocketbase.io/docs/authentication` | MFA documentation link. | `ui/.env.example`, `MFAAccordion.svelte` |
| `VITE_PB_PROTECTED_FILE_DOCS` | Browser-exposed | Public URL | Optional | Global | `https://pocketbase.io/docs/files-handling` | Protected file docs link. | `ui/.env.example`, `SchemaFieldFile.svelte` |

## Public Runtime Env Variables

These variables belong to the public/template runtime repo at `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5`.

| Variable | Runtime side | Secret level | Required | Scope | Example placeholder | Used for | Referenced in |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `VITE_PB_URL` | Browser-exposed and server-side Vite env | Public URL | Required | Per instance | `https://api.client.example.com` | PocketBase/API base URL for public content, server routes, and browser helpers. | `.env.example`, `src/hooks.server.js`, `src/lib/server/pb.js`, booking/newsletter helpers |
| `VITE_NUVIO_BACKEND_URL` | Browser-exposed | Public URL | Optional | Per instance | `https://api.client.example.com` | Explicit backend URL for public booking/contact helpers. Falls back to `VITE_PB_URL`. | `.env.example`, `src/lib/utils/booking.js` |
| `VITE_CMS_PREVIEW_PARENT_ORIGIN` | Browser-exposed | Public origin(s) | Required for production cross-origin CMS preview | Per instance | `https://admin.client.example.com` | Public runtime frame-ancestor allowlist and preview postMessage target origin. | `.env.example`, `src/hooks.server.js`, `SitePageRenderer.svelte` |
| `VITE_PUBLIC_SITE_BASE_URL` | Browser-exposed | Public URL | Recommended | Per instance | `https://www.client.example.com` | Canonical public site base for robots and SEO helpers. | `.env.example`, `robots.txt/+server.js`, `src/lib/server/seo.js` |
| `NUVIO_BACKEND_URL` | Server-only | Public/sensitive URL | Optional but recommended server fallback | Per instance | `https://api.client.example.com` | Server-side backend URL fallback for newsletter lifecycle and notification proxy helpers. | `.env.example`, `newsletter-lifecycle.js`, `nuvio-notifications.js` |
| `PB_URL` | Server-only | Public/sensitive URL | Optional server fallback | Per instance | `https://api.client.example.com` | Server-side backend URL fallback. Prefer `NUVIO_BACKEND_URL` in public runtime server env. | `.env.example`, `newsletter-lifecycle.js`, `nuvio-notifications.js` |
| `NUVIO_WEBSITE_DEBUG` | Server-only | Sensitive operational flag | Optional/dev only | Per environment | `false` | Enables public runtime backend notification debug logging. Do not enable in production unless actively diagnosing. | `src/lib/server/nuvio-notifications.js` |

## Third-Party Service Variables

| Service | Variables | Secret handling | Notes |
| --- | --- | --- | --- |
| Resend | `NUVIO_RESEND_API_KEY`, `RESEND_API_KEY`, `NUVIO_RESEND_FROM`, `RESEND_FROM`, `NUVIO_RESEND_FROM_EMAIL`, `RESEND_FROM_EMAIL` | API keys are secret. Sender emails are not secret but are instance-specific. | Prefer `NUVIO_RESEND_*` names. Configure verified sender/domain before smoke tests. |
| Umami | `NUVIO_UMAMI_API_URL`, `NUVIO_UMAMI_API_KEY`, `NUVIO_UMAMI_USERNAME`, `NUVIO_UMAMI_PASSWORD` | API key, username, and password are secret. API URL may be public but is instance/provider-specific. | Use either API key auth or username/password according to the deployed Umami setup. |
| Google Places | `NUVIO_GOOGLE_PLACES_API_KEY`, `GOOGLE_PLACES_API_KEY` | API key is secret. | Prefer `NUVIO_GOOGLE_PLACES_API_KEY`. Restrict the key in Google Cloud where possible. |

## CORS, CSP, and Preview Frame Variables

| Variable | App | Required | Production guidance |
| --- | --- | --- | --- |
| `NUVIO_CORS_ALLOWED_ORIGINS` | Backend/server | Yes | Set exact admin and public origins. Do not use wildcard origins in production. |
| `NUVIO_CMS_PREVIEW_FRAME_SRC` | Backend/server | If CMS preview frames public runtime cross-origin | Set exact public runtime origin(s) that the backoffice may iframe. |
| PocketBase trusted proxy headers | Backend/server | If backend runs behind Coolify/reverse proxy and is not directly reachable | Configure trusted proxy headers so `RequestEvent.RealIP()` keys contact submit rate limits by visitor IP instead of the proxy IP. Do not trust forwarded headers if the app is directly reachable. |
| `VITE_CMS_PREVIEW_PARENT_ORIGIN` | Public runtime | If CMS preview frames public runtime cross-origin | Set exact backoffice/admin origin(s) allowed to frame the public runtime. |

## QA and Dev-Only Variables

| Variable | App | Secret level | Notes |
| --- | --- | --- | --- |
| `NUVIO_ALLOW_DEV_RESET` | Backend/dev tools | Sensitive safety flag | Enables destructive QA reset/snapshot flows when set to `1`. Never enable casually in staging or production. |
| `NUVIO_QA_BASE_URL` | Backend/dev tools | Public/sensitive URL | Optional base URL for QA seed tools. |
| `PB_SUPERUSER_EMAIL` | Backend/dev tools | Secret | Used by QA seed/bootstrap tools. |
| `PB_SUPERUSER_PASSWORD` | Backend/dev tools | Secret | Used by QA seed/bootstrap tools. |

## PocketBase Operational Variables

These are inherited PocketBase-level environment variables found in code or docs. They are not Nuvio client feature settings, but they may matter for production operations.

| Variable | Secret level | Required | Used for | Referenced in |
| --- | --- | --- | --- | --- |
| `PB_THUMBS_MAX_WORKERS` | Public operational setting | Optional | Thumbnail worker concurrency. | `apis/file.go` |
| `PB_THUMBS_MAX_WAIT` | Public operational setting | Optional | Thumbnail generation wait limit. | `apis/file.go` |
| `PB_FILES_DELETE_MAX_WORKERS` | Public operational setting | Optional | File deletion worker concurrency. | `core/base.go` |
| `PB_ID_TOKEN_LEEWAY` | Public operational setting | Optional | OIDC token leeway in seconds. | `tools/auth/oidc.go`, changelog |
| custom `--encryptionEnv` target, for example `PB_ENCRYPTION_KEY` | Secret | Optional, required only if encrypted app settings are configured | Optional app settings encryption key. The env variable name is selected via the `--encryptionEnv` flag/config, not hardcoded. | `pocketbase.go`, `core/settings_*` |

## Per-Client Environment Checklist

For every client instance, define:

- Public site origin, for example `https://www.client.example.com`.
- Admin/backoffice origin, for example `https://admin.client.example.com`.
- Backend/API origin, for example `https://api.client.example.com`, if separate.
- Persistent `pb_data` volume/path.
- Persistent storage/files path.
- Backup destination and retention.
- CMS snapshot/bootstrap source.
- Email provider configuration.
- Analytics provider configuration.
- CORS origins.
- CSP/frame preview origins.
- Public runtime base URLs.

## Missing or Unknown Variables to Confirm

The audit found these documentation/template gaps:

- Backoffice UI code references `VITE_PUBLIC_SITE_BASE_URL`, but `ui/.env.example` does not currently document it.
- Backoffice UI code references legacy `VITE_SITE_BASE_URL`, but `ui/.env.example` does not currently document it.
- Public runtime code references `NUVIO_WEBSITE_DEBUG`, but `cms5/.env.example` does not currently document it.
- Backend `.env.example` does not document optional PocketBase operational variables such as `PB_THUMBS_MAX_WORKERS`, `PB_THUMBS_MAX_WAIT`, `PB_FILES_DELETE_MAX_WORKERS`, or `PB_ID_TOKEN_LEEWAY`.
- `PB_SERVICE_EMAIL` and `PB_SERVICE_PASSWORD` are present in backend `.env.example`, but no direct code reference was found in this audit. Keep them documented as reserved/service-auth placeholders until their final usage is confirmed.
- Resend sender variables have several aliases: `NUVIO_RESEND_FROM`, `RESEND_FROM`, `NUVIO_RESEND_FROM_EMAIL`, and `RESEND_FROM_EMAIL`. Prefer the `NUVIO_*` names in new deployments.
- Public runtime server helpers accept `NUVIO_BACKEND_URL`, `PB_URL`, and `VITE_PB_URL` as backend URL fallbacks. Prefer `NUVIO_BACKEND_URL` for server-only configuration and `VITE_PB_URL` only where browser exposure is intentional.
