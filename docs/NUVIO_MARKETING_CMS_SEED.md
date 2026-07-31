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
