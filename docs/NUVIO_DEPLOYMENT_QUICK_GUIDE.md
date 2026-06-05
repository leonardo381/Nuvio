# Nuvio Deployment Quick Guide

## Purpose

This is the practical day-to-day guide for creating a new Nuvio instance.

Use this when you need to deploy a client or Nuvio-owned website without reading every edge case. The full reference remains `docs/NUVIO_DEPLOYMENT_ENV_MATRIX.md`.

No real secrets or real client domains belong in this file. Use placeholders in docs and store real values in the deployment provider or secrets manager.

## The 10 Decisions Needed For Every Instance

Before setting env vars, decide:

1. Instance/client slug, for example `client-slug`.
2. Public site URL, for example `https://www.client.example.com`.
3. Admin/backoffice URL, for example `https://admin.client.example.com`.
4. Backend/API URL, for example `https://api.client.example.com`.
5. `pb_data` volume/path.
6. Storage/files path.
7. Backup path or bucket.
8. CMS snapshot name.
9. Code version/tag.
10. Enabled features, for example CMS, Contact, WhatsApp, Newsletter, Booking, Reports, Reviews.

## Minimum Required Env Vars

### Backend

| Variable | Plain-language purpose |
| --- | --- |
| `PB_URL` | The backend/API URL other tools and runtimes can use. |
| `NUVIO_PUBLIC_BASE_URL` | The public website URL used for visitor-facing links, especially newsletter lifecycle links. |
| `NUVIO_CORS_ALLOWED_ORIGINS` | The exact origins allowed to call the backend, usually admin and public site URLs. |
| `NUVIO_CMS_PREVIEW_FRAME_SRC` | The public site origin the backoffice may embed in the CMS preview iframe. |

### Backoffice UI

Every `VITE_*` variable is exposed to the browser. Never put secrets here.

| Variable | Plain-language purpose |
| --- | --- |
| `VITE_PB_BACKEND_URL` | The API URL the backoffice UI uses. |
| `VITE_PUBLIC_SITE_BASE_URL` | The public website URL used by CMS preview/open-site links. |

### Public Runtime

Every `VITE_*` variable is exposed to the browser. Never put secrets here.

| Variable | Plain-language purpose |
| --- | --- |
| `VITE_PB_URL` | The backend/API URL used by the public runtime. |
| `VITE_NUVIO_BACKEND_URL` | The backend/API URL used by public forms and booking helpers. Usually same as `VITE_PB_URL`. |
| `VITE_PUBLIC_SITE_BASE_URL` | The public website URL used for SEO, robots, and canonical helpers. |
| `VITE_CMS_PREVIEW_PARENT_ORIGIN` | The admin/backoffice origin allowed to frame the public runtime for CMS preview. |
| `NUVIO_BACKEND_URL` | Server-only backend/API URL fallback for public runtime server helpers. |

## Optional Feature Env Vars

Only configure these when the corresponding feature is enabled for the instance.

### Email / Resend

| Variable | Plain-language purpose |
| --- | --- |
| `NUVIO_RESEND_API_KEY` | Secret API key for sending email through Resend. |
| `NUVIO_RESEND_FROM` | Sender identity used for outgoing emails. |

### Analytics / Umami

| Variable | Plain-language purpose |
| --- | --- |
| `NUVIO_UMAMI_API_URL` | Umami API URL. |
| `NUVIO_UMAMI_API_KEY` | Secret Umami API key, if using API-key auth. |
| `NUVIO_UMAMI_USERNAME` | Secret Umami username, if using username/password auth. |
| `NUVIO_UMAMI_PASSWORD` | Secret Umami password, if using username/password auth. |

### Google Places

| Variable | Plain-language purpose |
| --- | --- |
| `NUVIO_GOOGLE_PLACES_API_KEY` | Secret Google Places API key for reviews/place lookup features. |

## Dev/QA Only

Do not configure these for normal production deployments unless there is a very deliberate operational reason.

| Variable | Plain-language purpose |
| --- | --- |
| `NUVIO_ALLOW_DEV_RESET` | Enables destructive QA/dev reset or snapshot tooling when set to `1`. Never enable this in production casually. |
| `NUVIO_QA_BASE_URL` | QA tool base URL override. |
| `PB_SUPERUSER_EMAIL` | Superuser email for internal QA/dev automation. |
| `PB_SUPERUSER_PASSWORD` | Superuser password for internal QA/dev automation. |

## Ignore For First Deployment Unless Needed

These are valid variables, but most first deployments do not need them.

- `PB_THUMBS_MAX_WORKERS`
- `PB_THUMBS_MAX_WAIT`
- `PB_FILES_DELETE_MAX_WORKERS`
- `PB_ID_TOKEN_LEEWAY`
- `PB_SERVICE_EMAIL`
- `PB_SERVICE_PASSWORD`
- `VITE_PB_*` documentation link variables such as `VITE_PB_DOCS_URL`, `VITE_PB_RELEASES`, and SDK/docs links.

## Simple Deployment Example

Placeholder decisions:

```text
Public:
https://www.client.example.com

Admin:
https://admin.client.example.com

API:
https://api.client.example.com
```

Minimum backend env:

```env
PB_URL=https://api.client.example.com
NUVIO_PUBLIC_BASE_URL=https://www.client.example.com
NUVIO_CORS_ALLOWED_ORIGINS=https://admin.client.example.com https://www.client.example.com
NUVIO_CMS_PREVIEW_FRAME_SRC=https://www.client.example.com
```

Minimum backoffice UI env:

```env
VITE_PB_BACKEND_URL=https://api.client.example.com
VITE_PUBLIC_SITE_BASE_URL=https://www.client.example.com
```

Minimum public runtime env:

```env
VITE_PB_URL=https://api.client.example.com
VITE_NUVIO_BACKEND_URL=https://api.client.example.com
VITE_PUBLIC_SITE_BASE_URL=https://www.client.example.com
VITE_CMS_PREVIEW_PARENT_ORIGIN=https://admin.client.example.com
NUVIO_BACKEND_URL=https://api.client.example.com
```

## New Instance Checklist

- Choose domains.
- Create isolated `pb_data`, storage, and backup volumes/paths.
- Set minimum env vars.
- Add optional feature env vars only for enabled features.
- Deploy backend/backoffice.
- Deploy public runtime.
- Apply migrations.
- Restore CMS snapshot.
- Configure website settings in CMS.
- Run smoke tests.
- Create initial backup.
- Record code version/tag and CMS snapshot name.

## Reference

For the complete variable reference, edge cases, aliases, and notes, use:

- `docs/NUVIO_DEPLOYMENT_ENV_MATRIX.md`
- `docs/NUVIO_INSTANCE_BOOTSTRAP_CHECKLIST.md`
