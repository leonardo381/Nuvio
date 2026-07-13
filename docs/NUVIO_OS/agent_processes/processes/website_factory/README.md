# Website Factory

## Purpose

`WEBSITE_FACTORY` is the standard Nuvio agent process for building client or official public websites with minimum friction.

It uses approved source blocks, staged agent workflows, required artifacts, and explicit gates so agents do not mix strategy, design, code, copywriting, integration, QA, and deploy in the same step.

## Use This Process When

- Building a new client website.
- Building or revising an official Nuvio public website.
- Moving from brief to sitemap, page blueprint, block selection, import, assembly, copy, visual adaptation, integration, QA, deploy, and handoff.

## Non-Negotiable Rules

- Read [Website Factory Process](WEBSITE_FACTORY_PROCESS.md) before executing.
- The source block library is immutable.
- Raw block import is a separate stage.
- Do not modify source blocks.
- Do not adapt copied blocks until the correct later stage.
- Do not make real sites depend on cms5, Reference, or source libraries at runtime.

## Local Docs

| File | Purpose |
| --- | --- |
| [Website Factory Process](WEBSITE_FACTORY_PROCESS.md) | Canonical stage-gate process skeleton. |
| [Block Library Rules](BLOCK_LIBRARY_RULES.md) | Source block immutability and raw import rules. |
| [Artifacts](artifacts/README.md) | Expected Website Factory artifacts. |

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
- [Danger Zones](../../../DANGER_ZONES.md)
- [Validation Matrix](../../../VALIDATION_MATRIX.md)
- [Reporting Formats](../../../REPORTING_FORMATS.md)

## Current Status

This is a skeleton process. WF2 should expand the stage details and artifact templates.
