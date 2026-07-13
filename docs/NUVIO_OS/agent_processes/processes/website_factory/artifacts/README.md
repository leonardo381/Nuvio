# Website Factory Artifacts

## Purpose

This file lists the expected Website Factory artifacts. Individual template files are intentionally not created in AP1/WF1.

WF2 should create or expand templates only when the process content is ready.

## Artifact Index

| Artifact | Purpose | Produced in stage | Required for v1 | Notes |
| --- | --- | --- | --- | --- |
| `SITE_BRIEF.md` | Captures client/business brief, constraints, goals, and known unknowns. | 0 - Intake | Yes | Starts the process. |
| `WEBSITE_AUDIT.md` | Records strategic analysis, current-site review, risks, and opportunity map. | 1 - Audit / strategic analysis | Yes | Do not mix with design. |
| `SITEMAP.md` | Defines pages, routes, priority, CTA role, and page purpose. | 2 - Sitemap / page plan | Yes | Keep page count controlled. |
| `PAGE_BLUEPRINTS.md` | Defines page section intent before block selection. | 3 - Page blueprint | Yes | No block importing here. |
| `BLOCK_SELECTION.md` | Records selected source blocks, source paths, rationale, and fit notes. | 4 - Block selection | Yes | Selection only. No import. |
| `BLOCK_IMPORT_LOG.md` | Logs raw copied blocks and minimal import/build fixes. | 5 - Raw block import | Yes | Required for every imported block. |
| `COPY_DECK.md` | Contains page and section copy after structure is stable. | 7 - Copywriting pass | Yes | No layout redesign here. |
| `VISUAL_ADAPTATION_NOTES.md` | Records brand and visual adaptation decisions. | 8 - Brand / visual adaptation pass | Yes | Applies only to copied target files. |
| `INTEGRATION_CHECKLIST.md` | Tracks Nuvio form, booking, newsletter, WhatsApp, SEO, analytics, and env integration. | 9 - Nuvio integration pass | Yes if integration is in scope | Static-only sites may mark items deferred. |
| `CMS_CONTENT_MAP.md` | Maps static site content to CMS/page/block/settings fields. | 10 - CMS/content mapping | Optional for static v1, required for CMS-connected v1 | Do not force CMS mapping prematurely. |
| `WEBSITE_QA_MATRIX.md` | Tracks QA checks, P0/P1/P2/deferred findings, and validation status. | 11 - QA pass | Yes | Must distinguish skipped checks. |
| `FINAL_REVIEW.md` | Records final review, accepted imperfections, and polish completion. | 12 - Final polish | Yes | Should reference deferred list. |
| `DEPLOYMENT_RECORD.md` | Records deploy target, version, env decisions, smoke checks, and rollback notes. | 13 - Deploy / handoff | Yes for deployed sites | No real secrets. |
| `DEFERRED_LIST.md` | Records out-of-scope or later work. | 12 - Final polish | Yes | Prevents hidden scope creep. |
| `CLIENT_HANDOFF.md` | Records operator/client handoff notes and next steps. | 13 - Deploy / handoff | Yes for handoff | Keep practical and non-technical where needed. |

## Artifact Rules

- Do not create filled artifacts with invented client facts.
- Do not include secrets, private tokens, real credentials, or private deployment values.
- Every stage report must say which artifact was produced or updated.
- If an artifact is skipped, explain why and what risk remains.

## Related Docs

- [Website Factory Process](../WEBSITE_FACTORY_PROCESS.md)
- [Block Library Rules](../BLOCK_LIBRARY_RULES.md)
- [Agent Process Standard](../../../AGENT_PROCESS_STANDARD.md)
