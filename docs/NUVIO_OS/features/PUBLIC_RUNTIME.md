# Public Runtime Agent Card

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Feature Cards](README.md)

## 1. Purpose

Guide agents working on public website runtime behavior, separate public site repos, cms5 legacy/lab context, the clean Reference template, sitemap/robots, public rendering, forms, booking/newsletter flows, environment boundaries, CORS, CSP, and preview frame behavior.

## 2. Current operating status

Needs polish + regression.

## 3. Source docs to read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Defines repo boundaries and source order. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Confirms routing and stop conditions for this feature. |
| 1 | Source of Truth | [../SOURCE_OF_TRUTH.md](../SOURCE_OF_TRUTH.md) | Clarifies which repo/docs are canonical. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Public runtime touches public endpoints and frame/CORS rules. |
| 2 | Operating Manual Public Runtime | `C:\Users\Leo\Documents\Documentation\Nuvio\Nuvio\04 - Features\Public Runtime.md` | Human guide to public rendering behavior. |
| 2 | Public Site Reference Contract | `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NUVIO_PUBLIC_SITE_REFERENCE_CONTRACT.md` | Defines the clean template contract. |
| 2 | Public Site Env Contract | `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\NUVIO_PUBLIC_SITE_ENV_CONTRACT.md` | Defines server-only vs browser-safe env. |
| 2 | Template Build | `C:\Users\Leo\Documents\Nuvio\Sites\Reference\docs\TEMPLATE_BUILD.md` | Explains template build and adaptation boundaries. |
| 2 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Selects checks. |
| 2 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Defines required final response structure. |

## 4. Likely code areas

- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\src\lib\nuvio\*` for the clean public site template.
- `C:\Users\Leo\Documents\Nuvio\Sites\Reference\src\routes\*` for reference routes.
- Real public site repos under `C:\Users\Leo\Documents\Nuvio\Sites\*` only when explicitly targeted.
- `C:\Users\Leo\Documents\Nuvio\tmpGallery\cms5` as lab/dev/history, not the canonical starter.
- Backend public DTO endpoints in the Nuvio backoffice repo when endpoint contracts are explicitly targeted.

## 5. Decisions to preserve

- Public sites are separate apps/repos/instances; why: each real site can have custom design and deployment lifecycle; agent implication: do not treat public runtime as one shared monolith.
- cms5 is test/dev/history; why: it contains useful learning but is not the clean starter; agent implication: do not turn cms5 into the canonical template.
- Reference is the clean starter/template; why: it documents safe integration boundaries; agent implication: update Reference only for reusable template concerns.
- Public websites can be custom, but Nuvio integration should stay predictable; why: backend contracts and env boundaries must remain understandable; agent implication: keep `src/lib/nuvio` contracts boring and explicit.
- Sitemap, robots, SEO, and public rendering are public contracts; why: crawlers and visitors consume them; agent implication: validate outputs, not just compile.
- Public contact, newsletter, and booking flows should call backend through server-safe boundaries where required; why: tokens/secrets/server-only env must not leak; agent implication: do not import private env into browser code.
- CORS, CSP, and preview frame settings are deployment-sensitive; why: preview can fail even when pages build; agent implication: check exact origins.

## 6. Allowed work now

- Template documentation and integration-boundary polish.
- Static marketing pages in real site repos when explicitly requested.
- Public route UI fixes that preserve backend contracts.
- Server/client boundary fixes.
- Sitemap/robots/SEO fixes based on existing fields.
- Validation and smoke docs.

## 7. Do not change unless explicitly requested

- Do not make cms5 the canonical starter.
- Do not make Reference depend on cms5 or global source libraries at runtime.
- Do not expose server-only env or tokens to browser code.
- Do not guess backend endpoints.
- Do not change backend public endpoint contracts from a public-site task.
- Do not bulk-copy static source libraries into template repos.
- Do not connect static marketing pages to CMS unless requested.

## 8. Common agent failure modes

- Copying cms5 code into Reference without adapting the contract.
- Treating `VITE_*` or `PUBLIC_*` variables as safe for secrets.
- Hardcoding local URLs into public runtime code.
- Breaking CMS preview by changing frame origins or postMessage behavior.
- Changing backend routes because a public page needs copy/layout work.
- Making a site depend on files outside its repo at runtime.

## 9. Validation checklist

- For Reference or SvelteKit public sites, run `npm run check`, `npm run lint`, and `npm run build` if scripts exist.
- For cms5, inspect scripts before running checks; do not assume it matches Reference.
- Manually check `/`, dynamic page routes, contact, booking, newsletter, `sitemap.xml`, `robots.txt`, and preview behavior when touched.
- Confirm no server-only env is imported by browser-safe modules.
- Confirm CORS/CSP/frame origins are exact and placeholder-safe in docs.

## 10. Reporting requirements

- Target repo and path.
- Changed files.
- Whether work changed Reference, a real site repo, cms5, or backend.
- Server/client env boundary confirmation.
- Public endpoint contract impact.
- Sitemap/robots/SEO/preview checks if touched.
- Validation results.
