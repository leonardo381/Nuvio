# Nuvio Instance Bootstrap Checklist

Use this checklist when provisioning a new Nuvio instance for Nuvio's own website, a client website, staging, or QA.

No real values belong in this document. Fill the deployment record in a private deployment tracker or secrets manager, not in git.

## 1. Choose Client Identity

- Choose the client display name.
- Choose the instance slug.
- Choose naming conventions for deployment services, volumes, backups, and CMS snapshots.
- Confirm whether this is production, staging, QA, or local development.

## 2. Choose Domains

Define exact origins before configuring CORS/CSP/frame policy.

| Purpose | Example placeholder |
| --- | --- |
| Public site domain | `https://www.client.example.com` |
| Admin/backoffice domain | `https://admin.client.example.com` |
| Backend/API domain | `https://api.client.example.com` |
| Analytics domain, if self-hosted | `https://analytics.client.example.com` |

If backend and admin share one origin, record that explicitly.

## 3. Provision Deployment or Service

- Create the backend/backoffice service from the approved Nuvio code version.
- Create the public runtime service from the matching template/runtime code version.
- Configure build commands and start commands according to the deployment target.
- Ensure each client has isolated persistent data and storage.
- Confirm the deployment platform's secret store is available before adding credentials.

## 4. Create Persistent Storage

Each instance needs isolated storage:

- `pb_data` for PocketBase database and runtime data.
- `pb_data/storage` or equivalent storage/files volume.
- Backup destination.
- CMS snapshot directory or object storage location.
- Optional logs location, if the host does not handle logs centrally.

Do not share writable `pb_data` or storage volumes across clients.

## 5. Configure Backend Env Variables

Set variables in the hosting provider secret/config system. Do not commit real `.env` files.

Minimum production-oriented backend variables:

| Variable | Placeholder |
| --- | --- |
| `PB_URL` | `https://api.client.example.com` |
| `NUVIO_PUBLIC_BASE_URL` | `https://www.client.example.com` |
| `NUVIO_CORS_ALLOWED_ORIGINS` | `https://admin.client.example.com https://www.client.example.com` |
| `NUVIO_CMS_PREVIEW_FRAME_SRC` | `https://www.client.example.com` |
| `NUVIO_RESEND_API_KEY` | `<secret-resend-api-key>` |
| `NUVIO_RESEND_FROM` | `Nuvio <hello@client.example.com>` |
| `NUVIO_UMAMI_API_URL` | `https://umami.example.com/api` |
| `NUVIO_UMAMI_API_KEY` | `<secret-umami-api-key>` |
| `NUVIO_GOOGLE_PLACES_API_KEY` | `<secret-google-places-key>` |

Only set optional provider variables for features the instance actually uses.

## 6. Configure Backoffice UI Env Variables

Every `VITE_*` variable is exposed to the browser. Never place secrets here.

| Variable | Placeholder |
| --- | --- |
| `VITE_PB_BACKEND_URL` | `https://api.client.example.com` |
| `VITE_PUBLIC_SITE_BASE_URL` | `https://www.client.example.com` |
| `VITE_PB_VERSION` | `<code-version-or-release-label>` |

Optional documentation link variables may keep the safe defaults from `ui/.env.example`.

## 7. Configure Public Runtime Env Variables

Every `VITE_*` variable is exposed to the browser. Keep server-only fallbacks separate.

| Variable | Placeholder |
| --- | --- |
| `VITE_PB_URL` | `https://api.client.example.com` |
| `VITE_NUVIO_BACKEND_URL` | `https://api.client.example.com` |
| `VITE_CMS_PREVIEW_PARENT_ORIGIN` | `https://admin.client.example.com` |
| `VITE_PUBLIC_SITE_BASE_URL` | `https://www.client.example.com` |
| `NUVIO_BACKEND_URL` | `https://api.client.example.com` |

Set `NUVIO_WEBSITE_DEBUG=false` or leave it unset for production.

## 8. Apply Migrations

- Confirm the deployed code version.
- Confirm the target `pb_data` path is correct.
- Run migrations using the approved deployment command for the environment.
- Confirm migrations complete successfully.
- Do not manually edit PocketBase migration history.

## 9. Restore CMS Base Snapshot

- Select the intended CMS snapshot.
- Confirm the snapshot belongs to the intended code/template version.
- Confirm the snapshot includes records and storage files.
- Restore into the target instance only after confirming the instance slug and data path.
- Verify native file fields after restore:
  - `Websites.logo`
  - `Websites.seoImage`
  - `Assets.file`
  - `Pages.seo_social_image`
  - `Blocks.image`

Restore must not silently leave broken file references.

## 10. Configure Website Settings

In the CMS/backoffice, configure:

- Identity and global SEO.
- Logo and social image assets.
- Page SEO defaults and page-specific SEO fields.
- Contact form settings and notification recipients.
- WhatsApp settings.
- Newsletter lifecycle templates.
- Booking services, availability, and visitor email templates.
- Reports/analytics settings.
- Internationalization languages, if needed.

Keep SEO fields on website/page top-level fields. Do not move SEO data into `websites.settings`.

## 11. Configure Email Provider

- Verify sender domain.
- Configure `NUVIO_RESEND_API_KEY`.
- Configure `NUVIO_RESEND_FROM` or the chosen fallback sender variable.
- Send a controlled test email.
- Verify newsletter lifecycle links use the public site domain.
- Verify contact/booking notification emails use the expected sender and recipients.

## 12. Configure Analytics Provider

- Configure Umami API URL and credentials.
- Confirm the website/site identifier in Nuvio Reports settings.
- Confirm reports load without exposing provider credentials to the browser.
- Confirm the public site tracking script is configured only if the client uses analytics.

## 13. Configure CORS, CSP, and Preview Origins

Backend/backoffice:

- Set `NUVIO_CORS_ALLOWED_ORIGINS` to exact admin and public origins.
- Set `NUVIO_CMS_PREVIEW_FRAME_SRC` to exact public runtime origins that the admin may iframe.
- Confirm production does not depend on wildcard CORS.

Public runtime:

- Set `VITE_CMS_PREVIEW_PARENT_ORIGIN` to exact admin/backoffice origin(s).
- Confirm public runtime is not frameable by arbitrary origins in production.
- Confirm CMS preview works after deploy.

## 14. Run Smoke Tests

Run focused smoke tests after deployment:

| Area | Check |
| --- | --- |
| Admin login | Superuser/admin login works. |
| CMS dashboard | Website, pages, and blocks load. |
| CMS preview | Preview iframe loads, refreshes, and can focus edited blocks. |
| Public page | Public homepage and at least one inner page render. |
| Assets | Existing images render, scoped upload works if enabled for the role. |
| Contact form | Valid public submit creates a lead and sends notifications if enabled. |
| WhatsApp tracking | Public tracking endpoint works without logging PII. |
| Newsletter | Subscribe, confirm, unsubscribe, and campaign preview behave safely. |
| Booking | Services, slots, appointment submit, and reschedule flow work. |
| Reports | Dashboard loads with the configured website and period. |
| SEO | Title, description, canonical, robots, sitemap, and social image render as expected. |
| CORS/CSP | Browser console has no unexpected CORS/CSP/frame errors. |

## 15. Create Initial Backup

- Create a database backup after successful smoke tests.
- Create or verify storage/files backup.
- Store the CMS snapshot identifier used for bootstrap.
- Record backup location and retention policy.
- Test restore procedure in staging/QA before relying on it for production.

## 16. Record Deployed Code Version

Record:

- Backend/backoffice commit or tag.
- Public runtime commit or tag.
- CMS snapshot name/version.
- Migration status.
- Deployment date.
- Operator.

## Deployment Record Template

Use this as a private deployment tracker template. Do not commit filled real values.

```text
Client:
Instance slug:
Environment:
Code version:
Backend/backoffice version:
Public runtime version:
Backend URL:
Public URL:
Admin URL:
pb_data volume/path:
Storage path:
Backup path:
CMS snapshot used:
Email provider:
Email sender:
Analytics provider:
Analytics site ID:
CORS allowed origins:
Preview frame src:
Preview parent origin:
Deployment date:
Operator:
Initial backup created:
Smoke tests completed:
Notes:
```

## Final Safety Check

Before handing the instance over:

- Real `.env` files are not committed.
- `.env.example` files contain placeholders only.
- No `VITE_*` value contains a secret.
- Backend secrets exist only in the secret store.
- `pb_data`, storage, and backups are isolated for this instance.
- CORS and frame origins are exact production origins.
- CMS snapshot restore included physical storage files.
- Initial backup exists.
- The deployed code version is recorded.
