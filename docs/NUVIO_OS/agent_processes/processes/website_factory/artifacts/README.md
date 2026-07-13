# Website Factory Artifacts

## Purpose

This folder stores reusable Website Factory artifact templates.

Generated per-client/per-site artifacts should usually live in the target website repo under:

```text
docs/website_factory/
```

If content is private or client-sensitive, store it in the approved private client folder and reference it without secrets.

Do not place client secrets, credentials, private tokens, or sensitive personal data in Website Factory artifacts.

## Stage To Template Map

| Stage | Template | Purpose |
| --- | --- | --- |
| Stage 0 | [SITE_BRIEF_TEMPLATE.md](SITE_BRIEF_TEMPLATE.md) | Capture the initial business and website brief. |
| Stage 1 | [WEBSITE_AUDIT_TEMPLATE.md](WEBSITE_AUDIT_TEMPLATE.md) | Audit current site/business/content status. |
| Stage 2 | [SITEMAP_TEMPLATE.md](SITEMAP_TEMPLATE.md) | Define planned pages, routes, priorities, and CTA roles. |
| Stage 3 | [PAGE_BLUEPRINTS_TEMPLATE.md](PAGE_BLUEPRINTS_TEMPLATE.md) | Blueprint each page section-by-section. |
| Stage 4 | [BLOCK_SELECTION_TEMPLATE.md](BLOCK_SELECTION_TEMPLATE.md) | Record exact selected source blocks. |
| Stage 5 | [BLOCK_IMPORT_LOG_TEMPLATE.md](BLOCK_IMPORT_LOG_TEMPLATE.md) | Log copied blocks and minimal raw import fixes. |
| Stage 7 | [COPY_DECK_TEMPLATE.md](COPY_DECK_TEMPLATE.md) | Record final copy by page and section. |
| Stage 8 | [VISUAL_ADAPTATION_NOTES_TEMPLATE.md](VISUAL_ADAPTATION_NOTES_TEMPLATE.md) | Record brand and visual adaptation decisions. |
| Stage 9 | [INTEGRATION_CHECKLIST_TEMPLATE.md](INTEGRATION_CHECKLIST_TEMPLATE.md) | Track enabled Nuvio integrations. |
| Stage 10 | [CMS_CONTENT_MAP_TEMPLATE.md](CMS_CONTENT_MAP_TEMPLATE.md) | Map static website content to CMS/editability. |
| Stage 11 | [WEBSITE_QA_MATRIX_TEMPLATE.md](WEBSITE_QA_MATRIX_TEMPLATE.md) | Record website QA checks and severities. |
| Stage 12 | [FINAL_REVIEW_TEMPLATE.md](FINAL_REVIEW_TEMPLATE.md) | Record final readiness and known limitations. |
| Stage 13 | [DEPLOYMENT_RECORD_TEMPLATE.md](DEPLOYMENT_RECORD_TEMPLATE.md) | Record deployment facts without secrets. |
| Stage 12 | [DEFERRED_LIST_TEMPLATE.md](DEFERRED_LIST_TEMPLATE.md) | Track deferred items intentionally left out of the current pass. |
| Stage 13 | [CLIENT_HANDOFF_TEMPLATE.md](CLIENT_HANDOFF_TEMPLATE.md) | Prepare client/operator handoff notes. |

## Artifact Rules

- Use the template for the active stage.
- Do not skip required status, open questions, or decision log fields.
- Mark unknowns as `Unknown / needs confirmation`.
- Do not use artifacts as a place to store secrets.
- Keep generated artifacts with the target site unless a private client folder is required.

## Related Docs

- [Website Factory Process](../WEBSITE_FACTORY_PROCESS.md)
- [Stage Gates](../STAGE_GATES.md)
- [Block Library Rules](../BLOCK_LIBRARY_RULES.md)
- [QA Checklist](../QA_CHECKLIST.md)
