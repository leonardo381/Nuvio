# Nuvio Coolify Base Deployment Plan

## Purpose

This plan maps the proven local Docker Compose base instance to Coolify for the first Nuvio Base deployment.

This is a deployment plan only. It does not deploy anything, change application behavior, change env variable names, or automate CMS snapshot restore. Use `docs/NUVIO_DEPLOYMENT_QUICK_GUIDE.md` for the short operator guide and `docs/NUVIO_DEPLOYMENT_ENV_MATRIX.md` for the full env reference.

No real domains, secrets, or client values belong in this document.

## D3 Artifact Summary

| Area | Current D3 behavior | Coolify implication |
| --- | --- | --- |
| Backoffice service | `nuvio-backoffice`, Dockerfile in backend repo, container port `8090`. | Create one Coolify service from the backend/backoffice repo. |
| Public runtime service | `nuvio-public`, Dockerfile in cms5 repo, container port `3000`. | Create one Coolify service from the cms5 repo. |
| Backoffice build args | `VITE_PB_BACKEND_URL`, `VITE_PUBLIC_SITE_BASE_URL`. | Must be configured as build-time variables in Coolify. |
| Public build args | `VITE_PB_URL`, `VITE_NUVIO_BACKEND_URL`, `VITE_PUBLIC_SITE_BASE_URL`, `VITE_CMS_PREVIEW_PARENT_ORIGIN`. | Must be configured as build-time variables in Coolify. |
| Backoffice runtime env | `PB_URL`, `NUVIO_PUBLIC_BASE_URL`, `NUVIO_CORS_ALLOWED_ORIGINS`, `NUVIO_CMS_PREVIEW_FRAME_SRC`, optional provider secrets. | Configure as runtime env variables/secrets on the backoffice service. |
| Public runtime env | `NUVIO_BACKEND_URL`, `PB_URL`, plus browser-exposed `VITE_*` values. | Configure server-only backend URL at runtime; keep `VITE_*` available at build time. |
| Backoffice volume | Named local compose volume mounted at `/app/pb_data`. | Create a persistent Coolify volume mounted at `/app/pb_data` on the backoffice service only. |
| Backoffice healthcheck | `/api/health` on port `8090`. | Use `/api/health` as health path if Coolify supports it. |
| Public healthcheck | D3 verifies HTTP `200` on `/`. | Use `/` or a future dedicated public health path. |
| Local-only assumptions | `localhost`, host port overrides, local env files. | Replace with exact Coolify domains/origins and Coolify-managed env/secrets. |
| Snapshot restore | Manual one-off step after backend volume exists. | Do not run restore in container startup. Use a controlled shell/job step. |

## Recommended Coolify Project Layout

Project name:

```text
Nuvio Base
```

Recommended services:

| Coolify service | Source repo | Branch/tag | Build context | Dockerfile | Container port | Public domain | Healthcheck |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `nuvio-base-backoffice` | Backend/backoffice repo | `<approved-backoffice-tag>` | repo root | `Dockerfile` | `8090` | `https://admin-base.example.com` | `/api/health` |
| `nuvio-base-public` | cms5 public runtime repo | `<approved-public-runtime-tag>` | repo root | `Dockerfile` | `3000` | `https://base.example.com` | `/` |

Persistent volume:

| Volume | Mount target | Service | Purpose |
| --- | --- | --- | --- |
| `nuvio-base-pb-data` | `/app/pb_data` | `nuvio-base-backoffice` | PocketBase database, runtime files, and storage. |

Optional future storage:

| Storage | Purpose | Notes |
| --- | --- | --- |
| Backup destination | Database/storage backup copies. | Can be host path, object storage, or platform backup integration. |
| Snapshot staging path | CMS snapshot import/export files. | Keep separate from committed repo files and runtime app code. |

## Domain And Origin Model

For the first Coolify deployment, prefer the simplest safe two-domain setup:

```text
Public runtime:
https://base.example.com

Admin/backoffice/API:
https://admin-base.example.com
```

Why this is the preferred first step:

- The backend and PocketBase admin already live in one service.
- The backoffice UI and API are same-origin, which reduces CORS complexity.
- The public runtime remains separate, so CMS preview and public traffic still exercise cross-origin policy correctly.
- It avoids introducing a third API domain before there is a product need.

Optional later split-domain setup:

```text
Public runtime:
https://base.example.com

Admin/backoffice:
https://admin-base.example.com

Backend/API:
https://api-base.example.com
```

Use split API/admin only if routing, branding, or platform constraints require it. If split later, update every API-facing `VITE_*`, `PB_URL`, `NUVIO_BACKEND_URL`, CORS origin, and preview setting together.

## Coolify Env And Build-Arg Mapping

### Backoffice Service: `nuvio-base-backoffice`

Build-time variables for the backoffice UI bundle:

| Name | Example placeholder | Browser-exposed | Notes |
| --- | --- | --- | --- |
| `VITE_PB_BACKEND_URL` | `https://admin-base.example.com` | Yes | Backoffice browser API URL. |
| `VITE_PUBLIC_SITE_BASE_URL` | `https://base.example.com` | Yes | CMS preview/open-site base URL. |

Runtime variables for the backend process:

| Name | Example placeholder | Secret | Required for first deploy | Notes |
| --- | --- | --- | --- | --- |
| `PB_URL` | `https://admin-base.example.com` | No | Yes | Backend/API public URL. |
| `NUVIO_PUBLIC_BASE_URL` | `https://base.example.com` | No | Yes | Public URL for lifecycle links. |
| `NUVIO_CORS_ALLOWED_ORIGINS` | `https://admin-base.example.com https://base.example.com` | No | Yes | Exact origins only. No wildcard for production-like deploys. |
| `NUVIO_CMS_PREVIEW_FRAME_SRC` | `https://base.example.com` | No | Yes | Public runtime origin the backoffice may iframe. |
| `NUVIO_RESEND_API_KEY` | `<secret-resend-api-key>` | Yes | Only if email is enabled | Prefer `NUVIO_*` names. |
| `NUVIO_RESEND_FROM` | `Nuvio <hello@example.com>` | No/instance-specific | Only if email is enabled | Use a verified sender/domain. |
| `NUVIO_UMAMI_API_URL` | `https://analytics.example.com/api` | No/instance-specific | Only if analytics is enabled | Provider API base URL. |
| `NUVIO_UMAMI_API_KEY` | `<secret-umami-api-key>` | Yes | Only if analytics is enabled | Use either API key or username/password auth. |
| `NUVIO_UMAMI_USERNAME` | `<secret-umami-username>` | Yes | Optional | Do not expose to browser. |
| `NUVIO_UMAMI_PASSWORD` | `<secret-umami-password>` | Yes | Optional | Do not expose to browser. |
| `NUVIO_GOOGLE_PLACES_API_KEY` | `<secret-google-places-key>` | Yes | Only if Google Places is enabled | Keep server-side only. |

Do not set `NUVIO_ALLOW_DEV_RESET` on the running service. Use it only for a controlled one-off snapshot restore command if that restore path is chosen.

### Public Runtime Service: `nuvio-base-public`

Build-time variables for the browser bundle:

| Name | Example placeholder | Browser-exposed | Notes |
| --- | --- | --- | --- |
| `VITE_PB_URL` | `https://admin-base.example.com` | Yes | Public runtime browser/API URL. |
| `VITE_NUVIO_BACKEND_URL` | `https://admin-base.example.com` | Yes | Public form/booking helper backend URL. |
| `VITE_PUBLIC_SITE_BASE_URL` | `https://base.example.com` | Yes | Canonical public site URL. |
| `VITE_CMS_PREVIEW_PARENT_ORIGIN` | `https://admin-base.example.com` | Yes | Admin origin allowed for CMS preview framing/postMessage. |

Runtime variables for the Node server:

| Name | Recommended first-deploy value | Alternative if Coolify service DNS is confirmed | Secret | Notes |
| --- | --- | --- | --- | --- |
| `NUVIO_BACKEND_URL` | `https://admin-base.example.com` | `http://nuvio-base-backoffice:8090` | No | Prefer public URL first to avoid internal DNS ambiguity. Internal DNS is fine later if tested. |
| `PB_URL` | `https://admin-base.example.com` | `http://nuvio-base-backoffice:8090` | No | Server-side fallback. Prefer `NUVIO_BACKEND_URL` as the primary variable. |
| `NUVIO_WEBSITE_DEBUG` | unset or `false` | unset or `false` | Operational | Do not enable in production unless actively diagnosing. |

Important Coolify rule:

`VITE_*` values must be available at image build time, not only at runtime. If Coolify separates build arguments from runtime environment, set them in both places only when needed, but never put secrets in `VITE_*`.

## Coolify Service Setup Steps

1. Create Coolify project `Nuvio Base`.
2. Add `nuvio-base-backoffice` from the backend/backoffice repo.
3. Select the approved branch/tag, for example `<approved-backoffice-tag>`.
4. Use Dockerfile build from repo root.
5. Set container port `8090`.
6. Attach domain `https://admin-base.example.com`.
7. Add persistent volume `nuvio-base-pb-data` mounted at `/app/pb_data`.
8. Configure backoffice build args and runtime env vars.
9. Deploy the backoffice service once.
10. Confirm `https://admin-base.example.com/api/health` returns healthy.
11. Add `nuvio-base-public` from the cms5 public runtime repo.
12. Select the matching approved branch/tag, for example `<approved-public-runtime-tag>`.
13. Use Dockerfile build from repo root.
14. Set container port `3000`.
15. Attach domain `https://base.example.com`.
16. Configure public runtime build args and runtime env vars.
17. Deploy the public runtime service.
18. Confirm `https://base.example.com` returns HTTP 200.

## CMS Snapshot Restore Workflow In Coolify

CMS snapshot restore should remain a controlled one-off bootstrap step.

Do not run restore automatically in the container `CMD`, startup script, or healthcheck.

Recommended first bootstrap flow:

1. Deploy `nuvio-base-backoffice` once so the `/app/pb_data` volume exists and migrations have run.
2. Confirm `/api/health` is healthy.
3. Stop the backoffice service.
4. Make the intended CMS snapshot available to the restore command. The snapshot must include records and storage files.
5. Run the approved restore tool from the same backend code version against the mounted `pb_data` volume.
6. Use a website ID guard when available.
7. Confirm restore output reports records and storage files restored.
8. Restart the backoffice service.
9. Deploy or restart the public runtime.
10. Smoke test CMS, preview, public pages, and assets.

Expected restore command shape if using the current CMS QA snapshot tool:

```powershell
$env:NUVIO_ALLOW_DEV_RESET="1"
go run ./tools/dev/restore_cms_qa_snapshot.go --name <snapshot-name> --websiteId <expected-website-id> --backendStopped --confirm RESTORE_CMS_QA_SNAPSHOT
```

Operational notes:

- Do not leave `NUVIO_ALLOW_DEV_RESET=1` on the running Coolify service.
- The current restore tool expects the backend to be stopped for write mode.
- The restore tool writes a safety backup and restore log under the snapshot workspace.
- Restore must include physical storage files for native file fields such as `Websites.logo`, `Websites.seoImage`, `Assets.file`, `Pages.seo_social_image`, and `Blocks.image`.
- If Coolify shell access can run the command inside a compatible container with the volume mounted, use that path.
- If Coolify shell access cannot safely mount the volume, use a temporary one-off container or host-level maintenance flow that mounts the same `nuvio-base-pb-data` volume.
- Confirm the exact one-off restore mechanism before the first real deployment.

## Backup Strategy V1

Minimal backup policy for the first base/staging instance:

| Item | What to back up | Frequency | Retention | Notes |
| --- | --- | --- | --- | --- |
| PocketBase data | Entire `/app/pb_data` volume | Daily for staging, more often for production | Start with 7 daily and 4 weekly | Includes SQLite database and storage if mounted together. |
| Storage files | `/app/pb_data/storage` | Same cadence as database | Same retention as database | Must stay consistent with database file references. |
| CMS snapshots | Approved CMS snapshot folders/manifests | When snapshot changes | Keep versioned | Store outside git if they contain runtime files. |
| Deployment metadata | Version/tag, env decisions, snapshot name | On each deploy | Permanent deployment record | Use private tracker/secrets manager, not docs. |

Before handoff, run one restore rehearsal in staging/QA. A backup that has never been restored is just a hopeful zip file wearing a tiny crown.

## Coolify Smoke Checklist

Run this after the first deploy and after snapshot restore:

| Area | Check |
| --- | --- |
| Backend health | `https://admin-base.example.com/api/health` returns healthy. |
| Backoffice login | Superuser/admin login works. |
| CMS dashboard | Website, pages, blocks, components, and assets load. |
| CMS preview | Preview iframe loads the public runtime and refreshes after edits. |
| Public runtime | `https://base.example.com` returns HTTP 200. |
| Public page render | Homepage and at least one inner page render correctly. |
| Assets | Logo, social image, page SEO image, block images, and Assets files render. |
| Contact form | Submit works if enabled and does not show CORS/CSP errors. |
| Newsletter | Subscribe/confirm/unsubscribe and campaign preview work if enabled. |
| Booking | Services, slots, appointment submit, and reschedule work if enabled. |
| Reports | Reports dashboard loads for the deployed website. |
| SEO | Title, description, canonical, robots, sitemap, and social image render as expected. |
| Browser console | No unexpected CORS, CSP, frame, hydration, or mixed-content errors. |
| Logs | No raw tokens, secrets, or visitor PII appear in logs. |

## Decisions And Blockers Before Clicking Deploy

Resolve these before the first Coolify deployment:

| Decision/blocker | Needed answer |
| --- | --- |
| Domains/DNS | Are `https://admin-base.example.com` and `https://base.example.com` created and pointed at Coolify? |
| TLS | Will Coolify issue/manage certificates for both domains? |
| Branch/tag | Which backend/backoffice tag and cms5 tag are approved for the base instance? |
| Env values | Are exact public/admin URLs, CORS origins, and preview origins confirmed? |
| Secrets | Are Resend, Umami, and Google Places secrets available only if those features are enabled? |
| Volume | Is `nuvio-base-pb-data` created and mounted only to the backoffice service? |
| Snapshot | Which CMS snapshot name and website ID should be restored? |
| Restore mechanism | Will restore run through Coolify shell, a one-off container, or host-level maintenance? |
| Backup target | Where will `/app/pb_data` and storage backups be stored? |
| Initial smoke owner | Who will run and record the smoke checklist? |

## Recommended First Deploy Order

1. Confirm domains and DNS.
2. Choose immutable code tags for both repos.
3. Create Coolify project and services.
4. Configure backoffice build args, runtime env, volume, domain, and healthcheck.
5. Deploy backoffice once.
6. Confirm backend health.
7. Stop backoffice and restore CMS snapshot through a controlled one-off flow.
8. Restart backoffice and verify CMS data/assets.
9. Configure public runtime build args, runtime env, domain, and healthcheck.
10. Deploy public runtime.
11. Run the Coolify smoke checklist.
12. Create initial backup.
13. Record deployment metadata privately.

## Follow-Up After First Base Deploy

- Decide whether to keep public runtime server calls on the public admin/API URL or switch to internal Coolify service DNS after testing.
- Add a dedicated production-safe snapshot/bootstrap runbook if the current dev/QA snapshot tool becomes awkward in Coolify.
- Add a formal backup job once the first manual backup/restore rehearsal is proven.
- Consider adding a public runtime health endpoint if `/` is too broad for health checks.
- Document the final production service names, tags, and backup locations in a private deployment record.