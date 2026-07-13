# Agent Prompts

## Purpose

Use these prompts to start one Website Factory stage at a time. Do not combine stages.


## Stage 0 - Intake

```text
Before changing anything, follow:
- You are executing Website Factory Stage 0 - Intake only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Capture business, site, client, operational, and launch constraints before any creative or technical work starts.
- Produce or update: SITE_BRIEF.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Ask only concrete missing intake questions
- Record unknowns as `Unknown / needs confirmation`
- Identify constraints, audiences, CTAs, integrations, and available content
- Create or update `SITE_BRIEF.md`

Forbidden actions:
- Designing layouts
- Selecting source blocks
- Writing final copy
- Coding
- Changing routes or integrations
- Promising scope not confirmed by the brief

Expected artifact:
- SITE_BRIEF.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 1 - Audit / strategic analysis

```text
Before changing anything, follow:
- You are executing Website Factory Stage 1 - Audit / strategic analysis only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Analyze the current business/site/content/positioning and identify practical problems, opportunities, and risks.
- Produce or update: WEBSITE_AUDIT.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Assess current positioning
- Identify conversion, content, SEO basics, and trust gaps
- Identify opportunities and risks
- Create or update `WEBSITE_AUDIT.md`

Forbidden actions:
- Locking the sitemap
- Selecting blocks
- Writing final copy
- Implementing fixes
- Making unsupported SEO or analytics claims

Expected artifact:
- WEBSITE_AUDIT.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 2 - Sitemap / page plan

```text
Before changing anything, follow:
- You are executing Website Factory Stage 2 - Sitemap / page plan only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Define pages, routes, page purposes, priority, dependencies, and CTA roles.
- Produce or update: SITEMAP.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Propose pages and routes
- Assign page purpose and primary CTA
- Mark priority and dependencies
- Create or update `SITEMAP.md`

Forbidden actions:
- Selecting blocks
- Copying blocks
- Writing final section copy
- Creating routes in code
- Expanding scope without approval

Expected artifact:
- SITEMAP.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 3 - Page blueprint

```text
Before changing anything, follow:
- You are executing Website Factory Stage 3 - Page blueprint only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Define each page section-by-section before block selection or coding.
- Produce or update: PAGE_BLUEPRINTS.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Define page sections
- Define section goals, content needs, CTA role, and visual role
- Suggest block type categories without choosing exact source files
- Create or update `PAGE_BLUEPRINTS.md`

Forbidden actions:
- Copying source blocks
- Selecting exact source block files
- Writing final copy
- Changing sitemap without logging the reason
- Implementing routes

Expected artifact:
- PAGE_BLUEPRINTS.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 4 - Block selection

```text
Before changing anything, follow:
- You are executing Website Factory Stage 4 - Block selection only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Select approved source blocks that best match the page blueprint without modifying or importing them.
- Produce or update: BLOCK_SELECTION.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Inspect source block files read-only
- Select exact source blocks
- Record source path, target page/section, rationale, fit notes, risks, and dependencies
- Create or update `BLOCK_SELECTION.md`

Forbidden actions:
- Editing source blocks
- Copying blocks into the target repo
- Adapting copy or brand
- Replacing blocks with new designs
- Selecting blocks not compatible with target constraints without recording risk

Expected artifact:
- BLOCK_SELECTION.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 5 - Raw block import

```text
Before changing anything, follow:
- You are executing Website Factory Stage 5 - Raw block import only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Copy selected blocks into the target site as-is and make only minimal import/build fixes required to compile.
- Produce or update: BLOCK_IMPORT_LOG.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Copy selected blocks
- Copy required unchanged assets
- Fix import paths, file paths, dependency paths
- Make minimal syntax or lint formatting fixes required to compile
- Log every minimal fix in `BLOCK_IMPORT_LOG.md`

Forbidden actions:
- Changing copy/text
- Changing visual hierarchy
- Redesigning layout
- Merging or splitting blocks
- Abstracting components
- Renaming design concepts
- Adding CMS, forms, analytics, or integrations
- Changing responsive behavior intentionally
- Architecture cleanup
- Replacing the block with a new implementation

Expected artifact:
- BLOCK_IMPORT_LOG.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 6 - Page assembly

```text
Before changing anything, follow:
- You are executing Website Factory Stage 6 - Page assembly only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Assemble imported copied blocks into target routes/pages according to the blueprint while preserving selected block structure.
- Produce or update: PAGE_BLUEPRINTS.md / BLOCK_IMPORT_LOG.md updates.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Place imported blocks into target pages/routes
- Wire static anchors and internal links required by the blueprint
- Perform minimal composition adjustments required to make blocks coexist
- Update blueprint/import log when assembly exposes a mismatch

Forbidden actions:
- Final copywriting
- Brand redesign
- CMS integration
- Major layout redesign
- Source block modification
- Adding new pages not in sitemap

Expected artifact:
- PAGE_BLUEPRINTS.md / BLOCK_IMPORT_LOG.md updates

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 7 - Copywriting pass

```text
Before changing anything, follow:
- You are executing Website Factory Stage 7 - Copywriting pass only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Replace placeholder/source copy with site-specific copy while preserving approved structure.
- Produce or update: COPY_DECK.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Write page/section headlines, body copy, CTAs, microcopy, and SEO draft copy if applicable
- Remove unsupported claims
- Record unresolved copy questions
- Create or update `COPY_DECK.md`

Forbidden actions:
- Redesigning layout
- Changing visual hierarchy except to flag a blueprint issue
- Adding unsupported promises
- Adding new integrations
- Reopening sitemap without approval

Expected artifact:
- COPY_DECK.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 8 - Brand / visual adaptation pass

```text
Before changing anything, follow:
- You are executing Website Factory Stage 8 - Brand / visual adaptation pass only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Adapt copied blocks to the site brand, visual hierarchy, spacing, imagery, and assets without changing strategy.
- Produce or update: VISUAL_ADAPTATION_NOTES.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Apply approved colors, typography choices, spacing adjustments, imagery, logos, icons, and visual consistency fixes
- Replace placeholders with approved assets
- Log deviations from source blocks
- Create or update `VISUAL_ADAPTATION_NOTES.md`

Forbidden actions:
- Changing page strategy
- Adding sections without approval
- Changing copy positioning materially
- Adding dependencies casually
- Modifying source blocks

Expected artifact:
- VISUAL_ADAPTATION_NOTES.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 9 - Nuvio integration pass

```text
Before changing anything, follow:
- You are executing Website Factory Stage 9 - Nuvio integration pass only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Integrate applicable Nuvio flows and public runtime needs using existing technical contracts.
- Produce or update: INTEGRATION_CHECKLIST.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Integrate confirmed contact, WhatsApp, booking, newsletter, SEO basics, analytics/Umami, public runtime env, CMS mapping needs, and assets according to technical docs
- Record status and notes
- Create or update `INTEGRATION_CHECKLIST.md`

Forbidden actions:
- Guessing backend endpoints
- Changing backend contracts
- Inventing CMS architecture
- Adding unsupported fields
- Exposing secrets or server-only env values
- Reworking visual design

Expected artifact:
- INTEGRATION_CHECKLIST.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 10 - CMS/content mapping

```text
Before changing anything, follow:
- You are executing Website Factory Stage 10 - CMS/content mapping only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Map static content to Nuvio CMS fields/blocks where needed and identify what remains static.
- Produce or update: CMS_CONTENT_MAP.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Map content to existing CMS fields/blocks
- Mark editable vs static content
- Define fallbacks and ownership
- Create or update `CMS_CONTENT_MAP.md`

Forbidden actions:
- Inventing CMS architecture
- Adding schema/migrations
- Changing backend model
- Moving SEO/settings data without contract
- Making everything editable by default

Expected artifact:
- CMS_CONTENT_MAP.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 11 - QA pass

```text
Before changing anything, follow:
- You are executing Website Factory Stage 11 - QA pass only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Validate the website using the severity model and produce a clear issue matrix.
- Produce or update: WEBSITE_QA_MATRIX.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Run scoped checks
- Classify issues as P0/P1/P2/deferred
- Record expected vs actual
- Create or update `WEBSITE_QA_MATRIX.md`

Forbidden actions:
- Fixing issues before logging them unless the task specifically combines QA and fix
- Reopening strategy/design
- Deploying
- Ignoring failed validations
- Hiding skipped checks

Expected artifact:
- WEBSITE_QA_MATRIX.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 12 - Final polish

```text
Before changing anything, follow:
- You are executing Website Factory Stage 12 - Final polish only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Fix approved P0/P1 issues and cheap P2 polish without reopening strategy or design.
- Produce or update: FINAL_REVIEW.md and DEFERRED_LIST.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Fix approved P0/P1
- Apply cheap P2 polish that does not alter scope
- Record final readiness and deferred items
- Create or update final review and deferred list

Forbidden actions:
- Rewriting the site
- Adding new sections/features
- Changing strategy
- Adding unapproved dependencies
- Starting deploy without final gate

Expected artifact:
- FINAL_REVIEW.md and DEFERRED_LIST.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```


## Stage 13 - Deploy / handoff

```text
Before changing anything, follow:
- You are executing Website Factory Stage 13 - Deploy / handoff only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Deploy through the approved environment/process, record facts, and prepare a clear client/operator handoff.
- Produce or update: DEPLOYMENT_RECORD.md and CLIENT_HANDOFF.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.

Allowed actions:
- Deploy only if explicitly instructed
- Record environment, URL, branch/commit, build command, deploy method, smoke result, rollback notes
- Prepare handoff instructions

Forbidden actions:
- Deploying without approval
- Committing secrets
- Changing env names
- Skipping smoke checks
- Editing source blocks or reopening build stages

Expected artifact:
- DEPLOYMENT_RECORD.md and CLIENT_HANDOFF.md

Report format:
- Files inspected
- Files changed
- Artifact produced or updated
- Validation performed or skipped
- P0/P1/P2/deferred findings
- Stop conditions encountered
- Next safe stage
```
