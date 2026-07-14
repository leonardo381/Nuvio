# Website Factory Process

## Purpose

`WEBSITE_FACTORY` is the canonical Nuvio agent process for taking a client or official public website from brief to deploy through strict staged execution.

The process exists to stop agents from designing, coding, copywriting, integrating, QAing, and deploying in the same pass. It is an operating contract, not a technical architecture document.

## Scope

Use this process for:

- new client websites;
- official Nuvio website iterations;
- websites created from Nuvio public templates;
- static-to-Nuvio adaptation work;
- staged source block selection, import, copy, brand, integration, QA, and handoff.

Do not use this process to define backend schemas, PocketBase rules, deployment infrastructure, product feature behavior, or Nuvio OS global policy. Reference the technical Nuvio OS docs for those.

## Non-Negotiable Rules

- Agents MUST NOT mix stages.
- Agents MUST NOT modify source block library files.
- Agents MUST NOT copy source blocks before Stage 5.
- Agents MUST NOT adapt copied blocks during Stage 5 raw import.
- Agents MUST NOT guess backend endpoints, payloads, env names, CMS schemas, or deployment behavior.
- Agents MUST stop when a required input is missing or a requested action belongs to another stage.
- Agents MUST produce the required artifact for the active stage.
- Agents MUST report skipped validation honestly.
- Agents MUST avoid secrets, credentials, tokens, private client data, and unnecessary PII in Website Factory artifacts.
- Agents MUST stop for Hermes/Leonardo approval at explicit approval gates.

## Source Block Immutability Rule Summary

The source block library is immutable.

A source block is a read-only source asset used for selection. A copied block is a target-site copy that can be adapted only in the correct later stage.

Agents may:

- inspect source blocks read-only;
- select approved source blocks;
- copy approved blocks into a target website during Stage 5;
- adapt copied blocks only in later approved stages.

Agents must never:

- edit source block files;
- bulk-copy an entire source library;
- make a site depend on source-library files at runtime;
- replace exact block adaptation with a new invented design.

See [Block Library Rules](BLOCK_LIBRARY_RULES.md).

## Stage-Gate Model

Each stage has:

- entry inputs;
- required docs;
- allowed actions;
- forbidden actions;
- expected agent behavior;
- required artifacts;
- definition of done;
- stop conditions;
- common mistakes;
- an example prompt.

The next stage may start only when the required artifact exists and the current stage exit gate is satisfied or explicitly marked `Unknown / needs confirmation`.

## Approval Gates

Agents MUST stop for Hermes/Leonardo approval before:

- accepting the sitemap/page plan as the build plan;
- importing selected blocks into a target repo;
- starting Nuvio integration work;
- deploy/handoff;
- any source-block-related exception;
- any schema, backend, public runtime, env, config, or deployment change suggestion.

If approval is missing, record `Unknown / needs confirmation` and stop. Do not treat silence as approval.

## Reduced Website Factory Mode

Simple websites may use Reduced Website Factory Mode, but the stage model still applies.

Reduced mode MAY:

- make artifacts shorter;
- combine brief content into concise tables;
- mark non-applicable sections as `Not applicable`;
- keep copy, visual, integration, and QA notes compact.

Reduced mode MUST NOT:

- skip gates;
- skip raw block import logging;
- skip QA;
- skip deploy/handoff records;
- skip `DEFERRED_LIST.md` when anything is intentionally left out;
- hide unknowns or approvals.

Any compressed detail must be explicitly documented in the active artifact.

## Artifact Model

Generated per-client/per-site artifacts should usually live in the target website repo under:

```text
docs/website_factory/
```

If an artifact contains private or client-sensitive information, store it in the approved private client folder and reference it without secrets.

Do not place client secrets, credentials, private tokens, or sensitive personal data in Website Factory artifacts.

Artifact templates live in [artifacts](artifacts/README.md).

## Severity Model

| Severity | Meaning | Process behavior |
| --- | --- | --- |
| P0 | Blocks safe launch, deploy, handoff, or a core public flow. | Stop or fix before proceeding. |
| P1 | Important defect or missing requirement. | Fix in current scope if allowed; otherwise record and gate. |
| P2 | Polish, refinement, or minor quality issue. | Defer unless the active stage includes polish. |
| deferred | Out of current stage or intentionally later. | Record; do not implement casually. |

## Stage Overview

| Stage | Name | Goal | Required artifact |
| --- | --- | --- | --- |
| 0 | Intake | Capture business, site, client, operational, and launch constraints before any creative or technical work starts. | SITE_BRIEF.md |
| 1 | Audit / strategic analysis | Analyze the current business/site/content/positioning and identify practical problems, opportunities, and risks. | WEBSITE_AUDIT.md |
| 2 | Sitemap / page plan | Define pages, routes, page purposes, priority, dependencies, and CTA roles. | SITEMAP.md |
| 3 | Page blueprint | Define each page section-by-section before block selection or coding. | PAGE_BLUEPRINTS.md |
| 4 | Block selection | Select approved source blocks that best match the page blueprint without modifying or importing them. | BLOCK_SELECTION.md |
| 5 | Raw block import | Copy selected blocks into the target site as-is and make only minimal import/build fixes required to compile. | BLOCK_IMPORT_LOG.md |
| 6 | Page assembly | Assemble imported copied blocks into target routes/pages according to the blueprint while preserving selected block structure. | PAGE_BLUEPRINTS.md / BLOCK_IMPORT_LOG.md updates |
| 7 | Copywriting pass | Replace placeholder/source copy with site-specific copy while preserving approved structure. | COPY_DECK.md |
| 8 | Brand / visual adaptation pass | Adapt copied blocks to the site brand, visual hierarchy, spacing, imagery, and assets without changing strategy. | VISUAL_ADAPTATION_NOTES.md |
| 9 | Nuvio integration pass | Integrate applicable Nuvio flows and public runtime needs using existing technical contracts. | INTEGRATION_CHECKLIST.md |
| 10 | CMS/content mapping | Map static content to Nuvio CMS fields/blocks where needed and identify what remains static. | CMS_CONTENT_MAP.md |
| 11 | QA pass | Validate the website using the severity model and produce a clear issue matrix. | WEBSITE_QA_MATRIX.md |
| 12 | Final polish | Fix approved P0/P1 issues and cheap P2 polish without reopening strategy or design. | FINAL_REVIEW.md and DEFERRED_LIST.md |
| 13 | Deploy / handoff | Deploy through the approved environment/process, record facts, and prepare a clear client/operator handoff. | DEPLOYMENT_RECORD.md and CLIENT_HANDOFF.md |

## Full Expanded Stages 0-13


### Stage 0 - Intake

#### Goal

Capture business, site, client, operational, and launch constraints before any creative or technical work starts.

#### Inputs

- User request or sales notes
- Business name and type
- Known goals and constraints
- Any existing site/social/profile links
- Known integrations or must-have flows

#### Required docs to read

- README.md
- WEBSITE_FACTORY_PROCESS.md
- artifacts/SITE_BRIEF_TEMPLATE.md
- ../../../CORE.md
- ../../../DANGER_ZONES.md

#### Allowed actions

- Ask only concrete missing intake questions
- Record unknowns as `Unknown / needs confirmation`
- Identify constraints, audiences, CTAs, integrations, and available content
- Create or update `SITE_BRIEF.md`

#### Forbidden actions

- Designing layouts
- Selecting source blocks
- Writing final copy
- Coding
- Changing routes or integrations
- Promising scope not confirmed by the brief

#### Expected agent behavior

- Stay in discovery mode
- Prefer specific fields over open-ended brainstorming
- Flag missing inputs that block later stages
- Do not solve the website yet

#### Required output artifacts

- `SITE_BRIEF.md`

#### Definition of done

- `SITE_BRIEF.md` exists
- Goals, audience, pages, CTA, brand inputs, integrations, unknowns, and constraints are captured
- No design or implementation decisions are locked beyond confirmed constraints

#### Stop conditions

- Business identity is unknown
- Primary CTA is unknown and cannot be inferred safely
- The user asks for implementation during intake
- Sensitive client data would need to be stored in repo docs

#### Common mistakes

- Turning intake into a design proposal
- Filling unknowns with assumptions
- Selecting blocks before the site purpose is clear

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 0 - Intake only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Capture business, site, client, operational, and launch constraints before any creative or technical work starts.
- Produce or update: SITE_BRIEF.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 1 - Audit / strategic analysis

#### Goal

Analyze the current business/site/content/positioning and identify practical problems, opportunities, and risks.

#### Inputs

- `SITE_BRIEF.md`
- Existing site or profile links when available
- Current content/assets
- Known business goals

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- artifacts/WEBSITE_AUDIT_TEMPLATE.md
- ../../../features/PUBLIC_RUNTIME.md
- ../../../launch/FIRST_CLIENT_READINESS.md

#### Allowed actions

- Assess current positioning
- Identify conversion, content, SEO basics, and trust gaps
- Identify opportunities and risks
- Create or update `WEBSITE_AUDIT.md`

#### Forbidden actions

- Locking the sitemap
- Selecting blocks
- Writing final copy
- Implementing fixes
- Making unsupported SEO or analytics claims

#### Expected agent behavior

- Be evidence-based
- Separate confirmed issues from hypotheses
- Preserve business reality over generic SaaS advice
- Record what needs client confirmation

#### Required output artifacts

- `WEBSITE_AUDIT.md`

#### Definition of done

- Audit contains strengths, problems, content gaps, conversion issues, SEO basics, technical risks, opportunities, and recommended direction
- No sitemap is locked

#### Stop conditions

- No meaningful business or current-state information exists
- The audit requires private credentials
- The user asks to skip directly to implementation without accepting unknowns

#### Common mistakes

- Calling opinions facts
- Inventing traffic or ranking data
- Over-auditing beyond first useful direction

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 1 - Audit / strategic analysis only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Analyze the current business/site/content/positioning and identify practical problems, opportunities, and risks.
- Produce or update: WEBSITE_AUDIT.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 2 - Sitemap / page plan

#### Goal

Define pages, routes, page purposes, priority, dependencies, and CTA roles.

#### Inputs

- `SITE_BRIEF.md`
- `WEBSITE_AUDIT.md`
- Confirmed business goals
- Known launch scope

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- PAGE_PATTERNS.md
- artifacts/SITEMAP_TEMPLATE.md

#### Allowed actions

- Propose pages and routes
- Assign page purpose and primary CTA
- Mark priority and dependencies
- Create or update `SITEMAP.md`

#### Forbidden actions

- Selecting blocks
- Copying blocks
- Writing final section copy
- Creating routes in code
- Expanding scope without approval

#### Expected agent behavior

- Keep first launch scope lean
- Make every page justify its role
- Record deferred pages separately
- Prefer fewer strong pages over many weak pages

#### Required output artifacts

- `SITEMAP.md`

#### Definition of done

- Every planned page has route, purpose, CTA, priority, dependencies, and status
- Deferred pages are explicitly marked
- Hermes/Leonardo approval is recorded before treating the sitemap as accepted

#### Stop conditions

- Primary launch pages cannot be agreed
- A required page depends on missing legal/commercial content
- Sitemap would imply unsupported integrations

#### Common mistakes

- Planning pages because competitors have them
- Creating route sprawl
- Mixing page plan with block selection

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 2 - Sitemap / page plan only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Define pages, routes, page purposes, priority, dependencies, and CTA roles.
- Produce or update: SITEMAP.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 3 - Page blueprint

#### Goal

Define each page section-by-section before block selection or coding.

#### Inputs

- `SITEMAP.md`
- `SITE_BRIEF.md`
- `WEBSITE_AUDIT.md`

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- PAGE_PATTERNS.md
- VISUAL_WEIGHT_RULES.md
- artifacts/PAGE_BLUEPRINTS_TEMPLATE.md

#### Allowed actions

- Define page sections
- Define section goals, content needs, CTA role, and visual role
- Suggest block type categories without choosing exact source files
- Create or update `PAGE_BLUEPRINTS.md`

#### Forbidden actions

- Copying source blocks
- Selecting exact source block files
- Writing final copy
- Changing sitemap without logging the reason
- Implementing routes

#### Expected agent behavior

- Think in intent and structure
- Keep sections necessary and ordered
- Mark content dependencies
- Use page patterns without turning them into rigid templates

#### Required output artifacts

- `PAGE_BLUEPRINTS.md`

#### Definition of done

- Each page has ordered sections with goal, content needed, CTA role, visual role, suggested block type, and notes

#### Stop conditions

- A page purpose is unclear
- Required content is unavailable and affects structure
- Blueprint would require a feature not planned for launch

#### Common mistakes

- Blueprinting from a favorite block instead of page intent
- Skipping mobile implications
- Making every page equally heavy

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 3 - Page blueprint only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Define each page section-by-section before block selection or coding.
- Produce or update: PAGE_BLUEPRINTS.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 4 - Block selection

#### Goal

Select approved source blocks that best match the page blueprint without modifying or importing them.

#### Inputs

- `PAGE_BLUEPRINTS.md`
- Approved source block library path
- Known target stack constraints

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- BLOCK_LIBRARY_RULES.md
- VISUAL_WEIGHT_RULES.md
- artifacts/BLOCK_SELECTION_TEMPLATE.md

#### Allowed actions

- Inspect source block files read-only
- Select exact source blocks
- Record source path, target page/section, rationale, fit notes, risks, and dependencies
- Create or update `BLOCK_SELECTION.md`

#### Forbidden actions

- Editing source blocks
- Copying blocks into the target repo
- Adapting copy or brand
- Replacing blocks with new designs
- Selecting blocks not compatible with target constraints without recording risk

#### Expected agent behavior

- Prefer fit to novelty
- Choose blocks that can survive later copy and brand passes
- Record risks early
- Stop if the source library is insufficient

#### Required output artifacts

- `BLOCK_SELECTION.md`

#### Definition of done

- Every selected block has source path, target page/section, reason, fit notes, risks, and dependencies
- Source files remain untouched

#### Stop conditions

- No approved block matches a critical section
- A block requires dependencies not approved for the target site
- A source file would need modification

#### Common mistakes

- Treating source blocks as inspiration only
- Selecting too many blocks
- Ignoring dependency or asset needs

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 4 - Block selection only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Select approved source blocks that best match the page blueprint without modifying or importing them.
- Produce or update: BLOCK_SELECTION.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 5 - Raw block import

#### Goal

Copy selected blocks into the target site as-is and make only minimal import/build fixes required to compile.

#### Inputs

- `BLOCK_SELECTION.md`
- Target repo path
- Approved source blocks
- Target build/lint commands

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- BLOCK_LIBRARY_RULES.md
- artifacts/BLOCK_IMPORT_LOG_TEMPLATE.md
- ../../../VALIDATION_MATRIX.md

#### Allowed actions

- Copy selected blocks
- Copy required unchanged assets
- Fix import paths, file paths, dependency paths
- Make minimal syntax or lint formatting fixes required to compile
- Log every minimal fix in `BLOCK_IMPORT_LOG.md`

#### Forbidden actions

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

#### Expected agent behavior

- Be boring and literal
- Log all deviations
- Stop when tempted to improve
- Validate only enough to prove import/build status

#### Required output artifacts

- `BLOCK_IMPORT_LOG.md`

#### Definition of done

- Copied blocks exist in target repo
- Only allowed minimal fixes were made
- `BLOCK_IMPORT_LOG.md` records copied blocks, paths, fixes, issues, and forbidden-change confirmation
- Hermes/Leonardo approval to import selected blocks is recorded

#### Stop conditions

- A block cannot compile without redesign
- An unapproved dependency is required
- A source asset is missing
- A requested change belongs to a later stage

#### Common mistakes

- Doing copywriting during import
- Tidying component architecture
- Changing spacing because it looks nicer
- Forgetting to log tiny syntax fixes

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 5 - Raw block import only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Copy selected blocks into the target site as-is and make only minimal import/build fixes required to compile.
- Produce or update: BLOCK_IMPORT_LOG.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 6 - Page assembly

#### Goal

Assemble imported copied blocks into target routes/pages according to the blueprint while preserving selected block structure.

#### Inputs

- Imported blocks
- `PAGE_BLUEPRINTS.md`
- `BLOCK_IMPORT_LOG.md`
- `SITEMAP.md`

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- PAGE_PATTERNS.md
- BLOCK_LIBRARY_RULES.md

#### Allowed actions

- Place imported blocks into target pages/routes
- Wire static anchors and internal links required by the blueprint
- Perform minimal composition adjustments required to make blocks coexist
- Update blueprint/import log when assembly exposes a mismatch

#### Forbidden actions

- Final copywriting
- Brand redesign
- CMS integration
- Major layout redesign
- Source block modification
- Adding new pages not in sitemap

#### Expected agent behavior

- Preserve block fidelity
- Prefer explicit notes over silent restructuring
- Only adjust what assembly requires
- Keep future passes unblocked

#### Required output artifacts

- `PAGE_BLUEPRINTS.md / BLOCK_IMPORT_LOG.md updates`

#### Definition of done

- Pages exist or are updated according to sitemap
- Imported blocks are assembled in blueprint order
- Any deviations are logged
- No copy/brand/integration pass was smuggled in

#### Stop conditions

- Blueprint and selected blocks conflict materially
- Assembly requires redesign
- A missing route or dependency changes launch scope

#### Common mistakes

- Treating assembly as final design
- Collapsing sections without approval
- Mixing forms or CMS work into static assembly

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 6 - Page assembly only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Assemble imported copied blocks into target routes/pages according to the blueprint while preserving selected block structure.
- Produce or update: PAGE_BLUEPRINTS.md / BLOCK_IMPORT_LOG.md updates.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 7 - Copywriting pass

#### Goal

Replace placeholder/source copy with site-specific copy while preserving approved structure.

#### Inputs

- Assembled pages
- `SITE_BRIEF.md`
- `WEBSITE_AUDIT.md`
- `PAGE_BLUEPRINTS.md`
- Client content and tone requirements

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- COPYWRITING_GUIDE.md
- artifacts/COPY_DECK_TEMPLATE.md

#### Allowed actions

- Write page/section headlines, body copy, CTAs, microcopy, and SEO draft copy if applicable
- Remove unsupported claims
- Record unresolved copy questions
- Create or update `COPY_DECK.md`

#### Forbidden actions

- Redesigning layout
- Changing visual hierarchy except to flag a blueprint issue
- Adding unsupported promises
- Adding new integrations
- Reopening sitemap without approval

#### Expected agent behavior

- Use concrete business outcomes
- Avoid hype and guaranteed results
- Keep Portuguese-first or target-language copy adaptable
- Preserve approved structure unless it clearly fails the copy

#### Required output artifacts

- `COPY_DECK.md`

#### Definition of done

- Copy deck mirrors implemented page/section structure
- CTAs are clear
- Unsupported claims are removed
- Open questions are recorded

#### Stop conditions

- Required facts are unknown
- Legal/commercial claims need client approval
- Copy reveals a blueprint problem that would require redesign

#### Common mistakes

- Writing enterprise SaaS fluff
- Overpromising lead volume
- Changing layout to fit copy instead of flagging the issue

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 7 - Copywriting pass only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Replace placeholder/source copy with site-specific copy while preserving approved structure.
- Produce or update: COPY_DECK.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 8 - Brand / visual adaptation pass

#### Goal

Adapt copied blocks to the site brand, visual hierarchy, spacing, imagery, and assets without changing strategy.

#### Inputs

- Assembled pages with copy
- `VISUAL_ADAPTATION_NOTES.md` if existing
- Brand inputs/assets
- `PAGE_BLUEPRINTS.md`

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- VISUAL_WEIGHT_RULES.md
- BLOCK_LIBRARY_RULES.md
- artifacts/VISUAL_ADAPTATION_NOTES_TEMPLATE.md

#### Allowed actions

- Apply approved colors, typography choices, spacing adjustments, imagery, logos, icons, and visual consistency fixes
- Replace placeholders with approved assets
- Log deviations from source blocks
- Create or update `VISUAL_ADAPTATION_NOTES.md`

#### Forbidden actions

- Changing page strategy
- Adding sections without approval
- Changing copy positioning materially
- Adding dependencies casually
- Modifying source blocks

#### Expected agent behavior

- Make the copied blocks feel like one site
- Keep visual weight intentional
- Avoid decoration that hides CTAs
- Log risky deviations

#### Required output artifacts

- `VISUAL_ADAPTATION_NOTES.md`

#### Definition of done

- Visual adaptation notes record brand inputs, changes, deviations, and risks
- Pages are visually coherent across desktop and mobile
- No strategy changes were introduced

#### Stop conditions

- Brand inputs are missing for a required decision
- The selected block cannot support brand adaptation without redesign
- Assets are unlicensed or unavailable

#### Common mistakes

- Turning polish into redesign
- Making every section visually loud
- Ignoring mobile density

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 8 - Brand / visual adaptation pass only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Adapt copied blocks to the site brand, visual hierarchy, spacing, imagery, and assets without changing strategy.
- Produce or update: VISUAL_ADAPTATION_NOTES.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 9 - Nuvio integration pass

#### Goal

Integrate applicable Nuvio flows and public runtime needs using existing technical contracts.

#### Inputs

- Assembled/adapted pages
- `SITEMAP.md`
- `PAGE_BLUEPRINTS.md`
- Confirmed integrations needed
- Environment/config requirements

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- INTEGRATION_CHECKLIST.md
- ../../../features/PUBLIC_RUNTIME.md
- ../../../operations/PUBLIC_RUNTIME_DEPLOYMENT.md
- ../../../task_packs/LANDING_UMAMI_TASK_PACK.md
- artifacts/INTEGRATION_CHECKLIST_TEMPLATE.md

#### Allowed actions

- Integrate confirmed contact, WhatsApp, booking, newsletter, SEO basics, analytics/Umami, public runtime env, CMS mapping needs, and assets according to technical docs
- Record status and notes
- Create or update `INTEGRATION_CHECKLIST.md`

#### Forbidden actions

- Guessing backend endpoints
- Changing backend contracts
- Inventing CMS architecture
- Adding unsupported fields
- Exposing secrets or server-only env values
- Reworking visual design

#### Expected agent behavior

- Reference technical docs instead of duplicating implementation details
- Keep integration limited to confirmed flows
- Treat missing contract details as stop conditions

#### Required output artifacts

- `INTEGRATION_CHECKLIST.md`

#### Definition of done

- Integration checklist records each applicable flow, status, notes, and validation needs
- No unsupported contracts were invented
- Hermes/Leonardo approval to start integration work is recorded

#### Stop conditions

- Endpoint or payload contract is unknown
- Required env/secrets are missing
- Integration would expose private values
- Backend changes would be required but not scoped

#### Common mistakes

- Adding fake form fields
- Mixing analytics events with copy changes
- Treating public runtime env as browser-safe without checking

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 9 - Nuvio integration pass only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Integrate applicable Nuvio flows and public runtime needs using existing technical contracts.
- Produce or update: INTEGRATION_CHECKLIST.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 10 - CMS/content mapping

#### Goal

Map static content to Nuvio CMS fields/blocks where needed and identify what remains static.

#### Inputs

- Final-ish static pages
- `PAGE_BLUEPRINTS.md`
- `COPY_DECK.md`
- Known CMS model/fields
- Client editability needs

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- INTEGRATION_CHECKLIST.md
- ../../../features/PUBLIC_RUNTIME.md
- artifacts/CMS_CONTENT_MAP_TEMPLATE.md

#### Allowed actions

- Map content to existing CMS fields/blocks
- Mark editable vs static content
- Define fallbacks and ownership
- Create or update `CMS_CONTENT_MAP.md`

#### Forbidden actions

- Inventing CMS architecture
- Adding schema/migrations
- Changing backend model
- Moving SEO/settings data without contract
- Making everything editable by default

#### Expected agent behavior

- Use existing Nuvio CMS concepts
- Keep editability intentional
- Record gaps instead of forcing bad mappings

#### Required output artifacts

- `CMS_CONTENT_MAP.md`

#### Definition of done

- Each relevant content item has CMS/static status, fallback, owner, and notes
- Unmapped items are explicit

#### Stop conditions

- CMS field/block does not exist for a required editable item
- Mapping would violate Nuvio SEO/settings data boundaries
- Client editability requirements conflict with current CMS

#### Common mistakes

- Mapping decorative UI into CMS unnecessarily
- Inventing block schema in process docs
- Forgetting fallbacks

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 10 - CMS/content mapping only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Map static content to Nuvio CMS fields/blocks where needed and identify what remains static.
- Produce or update: CMS_CONTENT_MAP.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 11 - QA pass

#### Goal

Validate the website using the severity model and produce a clear issue matrix.

#### Inputs

- Built site or preview environment
- Current artifacts
- Validation commands/checks
- Target browser/device expectations

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- QA_CHECKLIST.md
- ../../../VALIDATION_MATRIX.md
- ../../../REPORTING_FORMATS.md
- artifacts/WEBSITE_QA_MATRIX_TEMPLATE.md

#### Allowed actions

- Run scoped checks
- Classify issues as P0/P1/P2/deferred
- Record expected vs actual
- Create or update `WEBSITE_QA_MATRIX.md`

#### Forbidden actions

- Fixing issues before logging them unless the task specifically combines QA and fix
- Reopening strategy/design
- Deploying
- Ignoring failed validations
- Hiding skipped checks

#### Expected agent behavior

- Be factual and reproducible
- Classify severity conservatively
- Separate QA findings from fixes
- Report skipped validation honestly

#### Required output artifacts

- `WEBSITE_QA_MATRIX.md`

#### Definition of done

- QA matrix records checks, results, severity, status, owner, and notes
- P0/P1/deferred items are clear

#### Stop conditions

- Required validation environment is unavailable
- A P0 blocks safe continuation
- Secrets/PII appear in output or logs

#### Common mistakes

- Calling visual preference a P0
- Fixing while testing without recording
- Skipping mobile checks

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 11 - QA pass only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Validate the website using the severity model and produce a clear issue matrix.
- Produce or update: WEBSITE_QA_MATRIX.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 12 - Final polish

#### Goal

Fix approved P0/P1 issues and cheap P2 polish without reopening strategy or design.

#### Inputs

- `WEBSITE_QA_MATRIX.md`
- Approved fix list
- Current site
- Known launch/defer decisions

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- QA_CHECKLIST.md
- artifacts/FINAL_REVIEW_TEMPLATE.md
- artifacts/DEFERRED_LIST_TEMPLATE.md

#### Allowed actions

- Fix approved P0/P1
- Apply cheap P2 polish that does not alter scope
- Record final readiness and deferred items
- Create or update final review and deferred list

#### Forbidden actions

- Rewriting the site
- Adding new sections/features
- Changing strategy
- Adding unapproved dependencies
- Starting deploy without final gate

#### Expected agent behavior

- Be surgical
- Respect QA severity
- Document known limitations
- Keep deferred work visible

#### Required output artifacts

- `FINAL_REVIEW.md and DEFERRED_LIST.md`

#### Definition of done

- Approved P0/P1 are fixed or explicitly blocked
- Final review has readiness verdict
- Deferred list is current

#### Stop conditions

- A P0 remains unresolved
- A requested polish item is actually a redesign
- Fixing would require a new process stage

#### Common mistakes

- Using polish as a redesign loophole
- Hiding deferred work
- Fixing low-value P2 while P1 remains

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 12 - Final polish only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Fix approved P0/P1 issues and cheap P2 polish without reopening strategy or design.
- Produce or update: FINAL_REVIEW.md and DEFERRED_LIST.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


### Stage 13 - Deploy / handoff

#### Goal

Deploy through the approved environment/process, record facts, and prepare a clear client/operator handoff.

#### Inputs

- Final approved site
- `FINAL_REVIEW.md`
- Deployment environment
- Approved env values without secrets in docs
- Rollback plan

#### Required docs to read

- WEBSITE_FACTORY_PROCESS.md
- ../../../operations/PUBLIC_RUNTIME_DEPLOYMENT.md
- ../../../launch/AGENT_HANDOFF_CHECKLIST.md
- ../../../VALIDATION_MATRIX.md
- artifacts/DEPLOYMENT_RECORD_TEMPLATE.md
- artifacts/CLIENT_HANDOFF_TEMPLATE.md

#### Allowed actions

- Deploy only if explicitly instructed
- Record environment, URL, branch/commit, build command, deploy method, smoke result, rollback notes
- Prepare handoff instructions

#### Forbidden actions

- Deploying without approval
- Committing secrets
- Changing env names
- Skipping smoke checks
- Editing source blocks or reopening build stages

#### Expected agent behavior

- Follow target deployment runbook
- Record exact facts
- Keep client handoff concise
- Escalate blockers before changing infrastructure

#### Required output artifacts

- `DEPLOYMENT_RECORD.md and CLIENT_HANDOFF.md`

#### Definition of done

- Deployment record and client handoff exist
- Hermes/Leonardo deploy/handoff approval is recorded
- Smoke checks are recorded
- Rollback notes exist
- Known limitations are visible

#### Stop conditions

- Deployment approval is missing
- Env/secrets are missing or unsafe
- P0 launch blocker remains
- Rollback path is unknown

#### Common mistakes

- Treating deploy as a build command only
- Leaving no record of commit/env/deploy method
- Promising unsupported maintenance

#### Example prompt

```text
Before changing anything, follow:
- You are executing Website Factory Stage 13 - Deploy / handoff only.
- Read the Website Factory Process, Stage Gates, Block Library Rules when relevant, and the artifact template for this stage.
- Goal: Deploy through the approved environment/process, record facts, and prepare a clear client/operator handoff.
- Produce or update: DEPLOYMENT_RECORD.md and CLIENT_HANDOFF.md.
- Do not perform work from any other Website Factory stage.
- Stop if required inputs are missing, source blocks would be modified, secrets/PII would be stored, or the work requires a forbidden action.
- Report files inspected, files changed, artifact produced, validation performed or skipped, P0/P1/P2/deferred findings, and next safe stage.
```


## Global Stop Conditions

Stop and report `Unknown / needs confirmation` when:

- required inputs are missing;
- required Hermes/Leonardo approval is missing;
- source-of-truth docs conflict;
- the requested action belongs to another Website Factory stage;
- a source block would need to be edited;
- the target site would need unapproved dependencies;
- an endpoint, env variable, CMS field, or deployment behavior is unknown;
- secrets, credentials, private tokens, or unnecessary PII would be recorded;
- validation required by the gate cannot be performed.

## Handoff Rules

- Every handoff MUST state the current stage and next safe stage.
- Every handoff MUST list artifacts created or updated.
- Every handoff MUST list P0/P1/P2/deferred items.
- Every handoff MUST say whether source blocks were untouched.
- Every handoff MUST say whether any stage boundary was intentionally stopped.
- Do not hand off a site as ready when P0 issues remain.

## Final Release Gate

A Website Factory site is release-ready only when:

- Stage 0-13 required artifacts exist or are explicitly marked not applicable;
- required approval gates are recorded;
- source block immutability is confirmed;
- final QA has no unresolved P0;
- P1 issues are fixed or accepted with owner/date;
- deployment record exists;
- rollback or revert notes exist;
- client/operator handoff exists;
- no secrets are committed;
- integration checks match the actual enabled features.
