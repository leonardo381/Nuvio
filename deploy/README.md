# Nuvio Base Local Docker Compose

This folder contains a local/staging-oriented Docker Compose template for one Nuvio base instance:

- `nuvio-backoffice`: the PocketBase-backed Nuvio backend/backoffice service on port `8090`.
- `nuvio-public`: the cms5 public runtime service on port `3000`.
- `nuvio_base_pb_data`: an isolated named volume mounted at `/app/pb_data`.

The files here are examples and placeholders. Do not commit real secrets or production `.env` files.

## 1. Prepare Local Env Files

From this `deploy` directory, copy the example env files:

```powershell
Copy-Item env.backend.local.example env.backend.local
Copy-Item env.public.local.example env.public.local
```

The default example values target local Docker:

- Backend/API: `http://localhost:8090`
- Backoffice: `http://localhost:8090/_/`
- Public runtime: `http://localhost:3000`

`VITE_*` variables are browser-exposed and are also wired as Docker build args in the compose file. Never place secrets in `VITE_*` variables.
If one of the default host ports is already in use, set explicit local overrides before building and running. Keep the browser URL and host port aligned:

```powershell
$env:NUVIO_PUBLIC_PORT="3001"
$env:NUVIO_PUBLIC_BROWSER_URL="http://localhost:3001"
docker compose -f deploy/docker-compose.base.example.yml build nuvio-public
docker compose -f deploy/docker-compose.base.example.yml up
```

## 2. Build

From the backend repo root:

```powershell
docker compose -f deploy/docker-compose.base.example.yml build
```

## 3. Run

```powershell
docker compose -f deploy/docker-compose.base.example.yml up
```

Or run detached:

```powershell
docker compose -f deploy/docker-compose.base.example.yml up -d
```

Then check:

- Backend health: `http://localhost:8090/api/health`
- Backoffice: `http://localhost:8090/_/`
- Public runtime: `http://localhost:3000`

## 4. Stop

```powershell
docker compose -f deploy/docker-compose.base.example.yml down
```

The named `nuvio_base_pb_data` volume is intentionally persistent. Remove it only when you intentionally want to destroy the local instance data.

## CMS Snapshot Restore Workflow

CMS snapshot restore is a controlled one-off operator step, not a container startup action.

Recommended flow:

1. Stop the compose stack.
2. Restore the snapshot against the mounted `pb_data` volume/path for this instance.
3. Confirm records and storage files restored together.
4. Start the compose stack again.
5. Smoke test the CMS, public site, assets, and preview iframe.

Do not run automatic snapshot restore from the compose startup command.