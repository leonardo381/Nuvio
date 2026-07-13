# Website Factory

## Purpose

`WEBSITE_FACTORY` is the Nuvio agent process for building client or official public websites from brief to deploy through strict staged gates.

Start with [Website Factory Process](WEBSITE_FACTORY_PROCESS.md). Do not improvise a workflow before reading the canonical process.

## Quick Start

1. Read [Website Factory Process](WEBSITE_FACTORY_PROCESS.md).
2. Confirm the active stage.
3. Read [Stage Gates](STAGE_GATES.md).
4. Read the stage-specific supporting doc and artifact template.
5. Produce only the artifact required for the active stage.
6. Stop before crossing into the next stage.

## Non-Negotiable Rules

- Source blocks are immutable.
- Raw block import is its own stage.
- No mixed-stage work.
- No source block edits.
- No runtime dependency on source libraries.
- No secrets or sensitive client data in repo artifacts.
- No guessed Nuvio endpoints, env variables, CMS fields, or deploy behavior.

## Local Process Docs

| File | Purpose |
| --- | --- |
| [Website Factory Process](WEBSITE_FACTORY_PROCESS.md) | Canonical stage 0-13 process contract. |
| [Stage Gates](STAGE_GATES.md) | Compact entry/exit gate checklist for every stage. |
| [Block Library Rules](BLOCK_LIBRARY_RULES.md) | Immutable source library and raw import rules. |
| [Page Patterns](PAGE_PATTERNS.md) | Home, Services, and Contact page patterns. |
| [Visual Weight Rules](VISUAL_WEIGHT_RULES.md) | Hierarchy, density, section rhythm, and visual restraint. |
| [Copywriting Guide](COPYWRITING_GUIDE.md) | Copywriting pass rules and tone constraints. |
| [Integration Checklist](INTEGRATION_CHECKLIST.md) | Nuvio integration checks and technical doc references. |
| [QA Checklist](QA_CHECKLIST.md) | QA scope, severity rules, and smoke checks. |
| [Agent Prompts](AGENT_PROMPTS.md) | Stage-scoped prompt templates for agents. |
| [Artifacts](artifacts/README.md) | Artifact templates and stage-to-template map. |

## Parent Process Docs

- [Agent Processes](../../README.md)
- [Agent Process Registry](../../AGENT_PROCESS_REGISTRY.md)
- [Agent Process Standard](../../AGENT_PROCESS_STANDARD.md)

## Relevant Nuvio OS Docs

- [Public Runtime](../../../features/PUBLIC_RUNTIME.md)
- [Public Runtime Deployment](../../../operations/PUBLIC_RUNTIME_DEPLOYMENT.md)
- [Landing and Umami Task Pack](../../../task_packs/LANDING_UMAMI_TASK_PACK.md)
- [First Client Readiness](../../../launch/FIRST_CLIENT_READINESS.md)
- [Demo Flow Runbook](../../../launch/DEMO_FLOW_RUNBOOK.md)
- [Agent Handoff Checklist](../../../launch/AGENT_HANDOFF_CHECKLIST.md)
- [Danger Zones](../../../DANGER_ZONES.md)
- [Validation Matrix](../../../VALIDATION_MATRIX.md)
- [Reporting Formats](../../../REPORTING_FORMATS.md)

## Artifact Location Rule

Generated per-site artifacts usually belong in the target website repo:

```text
docs/website_factory/
```

Private/client-sensitive artifacts belong in the approved private client folder. Do not commit secrets, credentials, private tokens, or sensitive personal data.

## Current Status

WF2 expanded this process into an operational standard with full stage gates, supporting docs, prompts, and artifact templates.
