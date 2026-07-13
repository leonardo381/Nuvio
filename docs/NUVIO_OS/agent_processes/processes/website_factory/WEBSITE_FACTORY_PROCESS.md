# Website Factory Process

## Purpose

The Website Factory Process is the canonical staged workflow for building client or official public websites with agent assistance.

It exists to keep strategy, design, source block handling, code, copywriting, integration, QA, polish, deploy, and handoff separated by gates.

## Non-Negotiable Rules

- Agents must not design, code, copywrite, integrate, and QA at the same time.
- Source blocks are immutable.
- Raw block import is a separate stage.
- Every stage must produce or update its required artifact.
- Agents must stop when a stage requires actions from another stage.
- Current source, target repo instructions, and Nuvio OS danger zones override this skeleton.

## Source Block Immutability Summary

The source block library is read-only. Agents may select and copy approved blocks, but must never modify source blocks. See [Block Library Rules](BLOCK_LIBRARY_RULES.md).

## Stage-Gate Model

Each stage has a goal, inputs, allowed actions, forbidden actions, required output artifacts, and a definition of done.

A stage is not complete until its artifact exists or is explicitly marked Unknown / needs confirmation with a reason.

## Artifact Rules

- Artifacts are process records, not optional notes.
- Artifacts should live with the target website project unless the user specifies another private client workspace.
- Do not create all artifact templates in this phase. See [Artifacts](artifacts/README.md).

## Severity Model Reference

Use the severity model from [Agent Process Standard](../../AGENT_PROCESS_STANDARD.md): P0, P1, P2, and deferred.

## Stage Overview

| Stage | Name | Primary artifact |
| --- | --- | --- || 0 | Intake | $(0 Intake Capture the brief and constraints. SITE_BRIEF.md[3]) |
| 1 | Audit / strategic analysis | $(1 Audit / strategic analysis Analyze business, audience, current site, and opportunities. WEBSITE_AUDIT.md[3]) |
| 2 | Sitemap / page plan | $(2 Sitemap / page plan Define pages, route purpose, and page priorities. SITEMAP.md[3]) |
| 3 | Page blueprint | $(3 Page blueprint Define section intent before selecting blocks. PAGE_BLUEPRINTS.md[3]) |
| 4 | Block selection | $(4 Block selection Select source blocks without importing or adapting them. BLOCK_SELECTION.md[3]) |
| 5 | Raw block import | $(5 Raw block import Copy selected blocks as-is with only minimal import/build fixes. BLOCK_IMPORT_LOG.md[3]) |
| 6 | Page assembly | $(6 Page assembly Assemble imported blocks into page routes without copy/brand/integration drift. PAGE_BLUEPRINTS.md, BLOCK_IMPORT_LOG.md[3]) |
| 7 | Copywriting pass | $(7 Copywriting pass Replace placeholder copy with approved page copy. COPY_DECK.md[3]) |
| 8 | Brand / visual adaptation pass | $(8 Brand / visual adaptation pass Apply approved visual and brand adaptation to copied blocks. VISUAL_ADAPTATION_NOTES.md[3]) |
| 9 | Nuvio integration pass | $(9 Nuvio integration pass Connect approved Nuvio forms, routes, analytics, and helpers. INTEGRATION_CHECKLIST.md[3]) |
| 10 | CMS/content mapping | $(10 CMS/content mapping Map static content to Nuvio CMS/content structures where useful. CMS_CONTENT_MAP.md[3]) |
| 11 | QA pass | $(11 QA pass Validate behavior, layout, accessibility, env boundaries, and launch risks. WEBSITE_QA_MATRIX.md[3]) |
| 12 | Final polish | $(12 Final polish Apply scoped final polish and record deferred work. FINAL_REVIEW.md, DEFERRED_LIST.md[3]) |
| 13 | Deploy / handoff | $(13 Deploy / handoff Prepare deploy record and client/operator handoff. DEPLOYMENT_RECORD.md, CLIENT_HANDOFF.md[3]) |

## Stage Details

### Stage 0 - Intake

#### Goal
Capture the brief and constraints. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(0 Intake Capture the brief and constraints. SITE_BRIEF.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 1 - Audit / strategic analysis

#### Goal
Analyze business, audience, current site, and opportunities. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(1 Audit / strategic analysis Analyze business, audience, current site, and opportunities. WEBSITE_AUDIT.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 2 - Sitemap / page plan

#### Goal
Define pages, route purpose, and page priorities. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(2 Sitemap / page plan Define pages, route purpose, and page priorities. SITEMAP.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 3 - Page blueprint

#### Goal
Define section intent before selecting blocks. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(3 Page blueprint Define section intent before selecting blocks. PAGE_BLUEPRINTS.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 4 - Block selection

#### Goal
Select source blocks without importing or adapting them. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(4 Block selection Select source blocks without importing or adapting them. BLOCK_SELECTION.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 5 - Raw block import

#### Goal
Copy selected blocks as-is with only minimal import/build fixes. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(5 Raw block import Copy selected blocks as-is with only minimal import/build fixes. BLOCK_IMPORT_LOG.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 6 - Page assembly

#### Goal
Assemble imported blocks into page routes without copy/brand/integration drift. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(6 Page assembly Assemble imported blocks into page routes without copy/brand/integration drift. PAGE_BLUEPRINTS.md, BLOCK_IMPORT_LOG.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 7 - Copywriting pass

#### Goal
Replace placeholder copy with approved page copy. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(7 Copywriting pass Replace placeholder copy with approved page copy. COPY_DECK.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 8 - Brand / visual adaptation pass

#### Goal
Apply approved visual and brand adaptation to copied blocks. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(8 Brand / visual adaptation pass Apply approved visual and brand adaptation to copied blocks. VISUAL_ADAPTATION_NOTES.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 9 - Nuvio integration pass

#### Goal
Connect approved Nuvio forms, routes, analytics, and helpers. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(9 Nuvio integration pass Connect approved Nuvio forms, routes, analytics, and helpers. INTEGRATION_CHECKLIST.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 10 - CMS/content mapping

#### Goal
Map static content to Nuvio CMS/content structures where useful. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(10 CMS/content mapping Map static content to Nuvio CMS/content structures where useful. CMS_CONTENT_MAP.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 11 - QA pass

#### Goal
Validate behavior, layout, accessibility, env boundaries, and launch risks. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(11 QA pass Validate behavior, layout, accessibility, env boundaries, and launch risks. WEBSITE_QA_MATRIX.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 12 - Final polish

#### Goal
Apply scoped final polish and record deferred work. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(12 Final polish Apply scoped final polish and record deferred work. FINAL_REVIEW.md, DEFERRED_LIST.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
### Stage 13 - Deploy / handoff

#### Goal
Prepare deploy record and client/operator handoff. WF2 will expand this section.

#### Inputs
- Prior stage artifact, if any.
- User-provided brief or confirmed process context.

#### Allowed actions
- Work only inside this stage's scope.
- Update the required stage artifact.

#### Forbidden actions
- Do not perform work from later stages.
- Do not touch implementation files unless this stage explicitly allows implementation.

#### Expected agent behavior
- Keep the stage narrow.
- Mark missing inputs as Unknown / needs confirmation.
- Stop rather than silently crossing into another stage.

#### Required output artifacts
- $(13 Deploy / handoff Prepare deploy record and client/operator handoff. DEPLOYMENT_RECORD.md, CLIENT_HANDOFF.md[3])

#### Definition of done
- Required artifact exists or is explicitly marked unavailable with a reason.
- Next stage inputs are clear.

#### Common mistakes
- Mixing this stage with copywriting, visual adaptation, integration, QA, or deploy.
- Treating assumptions as confirmed input.

#### Example prompt
WF2 will expand this section.
## Related Docs

- [Agent Process Standard](../../AGENT_PROCESS_STANDARD.md)
- [Block Library Rules](BLOCK_LIBRARY_RULES.md)
- [Artifacts](artifacts/README.md)
- [Public Runtime](../../../features/PUBLIC_RUNTIME.md)
- [Danger Zones](../../../DANGER_ZONES.md)
- [Validation Matrix](../../../VALIDATION_MATRIX.md)
- [Reporting Formats](../../../REPORTING_FORMATS.md)
