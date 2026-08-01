# Nuvio Marketing CMS Seed

## Purpose

This seed recreates the controlled CMS records for the Nuvio marketing website (`nuvio`) without committing `pb_data`.

It covers:

- website `nuvio`;
- pages `home`, `services`, `pricing`, and `contact`;
- the `nuvio-*` component definitions required by final-site strict mode;
- all strict-required blocks for the four root marketing pages.

The seed is intentionally narrow. It does not create arbitrary page-builder layouts, raw HTML fields, editor-controlled classes, secrets, or production data.

## Source Of Truth

The fixture is generated from the final public site marketing defaults:

```powershell
cd C:\Users\Leo\Documents\Nuvio\Sites\Nuvio-CalmEditorialV2
npm run export:marketing-cms-fixture -- --out C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\examples\base\fixtures\nuvio_marketing_cms_fixture.json
```

The exporter validates the pricing lock, rejects raw-HTML-like content, and keeps Contact mechanics out of CMS props.
## Preview Route Metadata

Backoffice preview routes are data-driven.

Preferred model:

- page-level safe preview path fields, when present;
- otherwise `website.settings.previewRoutes[pageSlug]`;
- otherwise fallback `/site/<websiteSlug>/<pageSlug>`.

The Nuvio marketing fixture/exporter records these root marketing preview routes in `website.settings.previewRoutes`:

| Page slug | Preview path |
| --- | --- |
| `home` | `/` |
| `services` | `/services` |
| `pricing` | `/pricing` |
| `contact` | `/contact` |

The backoffice preview iframe should preserve `cmsPreview=1` when opening or focusing a page preview.

Do not reintroduce hardcoded `websiteSlug === "nuvio"` routing in `PageCms.svelte`. Normal client websites should continue to preview through `/site/<websiteSlug>/<pageSlug>` unless their website/page records explicitly provide safe preview route metadata.

## Dry Run

Dry-run mode is the default and does not write records:

```powershell
cd C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio
go run ./examples/base --dir pb_data --migrationsDir pb_migrations seed-nuvio-marketing
```

Use a staging/local `--dir` path. Do not run this against production data without an explicit deployment plan and backup.

## Apply

Apply mode creates or updates only fixture-owned Nuvio marketing records:

```powershell
cd C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio
go run ./examples/base --dir pb_data --migrationsDir pb_migrations seed-nuvio-marketing --apply
```

For a fresh smoke DB, use a temporary directory instead of `pb_data`:

```powershell
$seedDir = Join-Path $env:TEMP ('nuvio-seed-' + [guid]::NewGuid().ToString('N'))
go run ./examples/base --dir $seedDir --migrationsDir pb_migrations seed-nuvio-marketing --apply
```

## Safety Behavior

- Natural keys are website slug, page slug, component key, and page block slot.
- The fixture marks owned records with `nuvio-marketing-cms-fixture`.
- Existing unowned `nuvio` website records are treated as conflicts and are not overwritten.
- Existing unowned matching component keys are treated as conflicts and are not overwritten.
- The seed does not delete records, reset databases, or modify unrelated websites.
- `pb_data` remains local runtime data and must not be committed.

## Ordering Caveat

If the `Blocks` collection has a `displayOrder` or `order` field, the seed writes deterministic order values.

If neither field exists, public block order relies on deterministic creation order. A future schema migration can add an explicit display-order field if needed.

## Final Site Strict Smoke

After seeding local/staging data, run the final public site in strict CMS mode:

```powershell
$env:NUVIO_API_URL = 'http://127.0.0.1:8090'
$env:NUVIO_ASSET_BASE_URL = 'http://127.0.0.1:8090'
$env:NUVIO_WEBSITE_ID = 'nuvio'
$env:NUVIO_MARKETING_CMS_ENABLED = 'true'
$env:NUVIO_MARKETING_CMS_STRICT = 'true'
$env:NUVIO_MARKETING_WEBSITE_ID = 'nuvio'
node build
```

Expected root routes:

- `/` -> 200
- `/services` -> 200
- `/pricing` -> 200
- `/contact` -> 200
- `/home` -> 404

Strict mode proves the root marketing pages can render from the seeded CMS records instead of silently falling back to code defaults.

## Temporary CMS Admin For Local/Staging Smoke

`superuser upsert` intentionally keeps normal PocketBase behavior unless a trusted operator asks for a Nuvio role explicitly. For local or staging authoring smoke, create the account and grant the admin role in one CLI-only command:

```powershell
go run ./examples/base --dir $seedDir --migrationsDir pb_migrations superuser upsert phase46-smoke@example.test <temporary-password> --role admin
```

This is for trusted operator access only. It does not create a browser/public bootstrap endpoint, does not make every superuser an admin automatically, and still relies on the runtime CMS middleware requiring `role=admin` or `role=client` plus `websiteAccess`. If the backend is behind Coolify or another proxy and you trust proxy headers for rate limiting or logs, configure trusted proxy handling only for that controlled proxy path; do not expose the backend directly with untrusted forwarded headers.
