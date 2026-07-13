# Agent Processes

## Purpose

The Agent Process Layer defines how agents execute repeatable Nuvio work.

Technical docs explain how Nuvio works. Agent Process docs explain how agents must work through multi-step workflows without mixing strategy, design, code, copywriting, integration, QA, and deployment in one pass.

## What This Layer Is

- A process-contract layer for repeatable agent work.
- A place for staged workflows, gates, artifacts, and definitions of done.
- A guardrail against broad, mixed-scope agent execution.

## What This Layer Is Not

- Not a replacement for feature, operations, launch, task-pack, validation, danger-zone, or reporting docs.
- Not permission to change code, deploy, restore, or touch runtime data.
- Not a place for product strategy brainstorming.

## When To Use This Layer

Use this layer when a task is a multi-stage workflow with required artifacts and gates, such as building a client or official website from a brief.

For one-off implementation, feature, operations, or launch tasks, start with the existing Nuvio OS router and task packs first.

## Entry Points

| File | Purpose |
| --- | --- |
| [Agent Process Registry](AGENT_PROCESS_REGISTRY.md) | Lists active and planned agent processes. |
| [Agent Process Standard](AGENT_PROCESS_STANDARD.md) | Defines the global process contract all agent processes must follow. |
| [Website Factory](processes/website_factory/README.md) | First active process: staged client/official website production. |

## Related Nuvio OS Docs

- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [OS Navigation](../OS_NAVIGATION.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
