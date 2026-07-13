# Agent Process Standard

## Purpose

This standard defines the global structure every Nuvio agent process must follow.

Agent processes are strict operational contracts. They must reduce mixed-scope agent behavior and make handoff, validation, and stop conditions explicit.

## Global Rule

Agents must not mix strategy, design, coding, copywriting, integration, QA, and deploy in the same stage unless the active process explicitly allows it.

## Required Process Sections

Every process must define:

- process purpose;
- stages;
- gates;
- required inputs;
- required docs;
- allowed actions;
- forbidden actions;
- expected agent behavior;
- required artifacts;
- definition of done;
- stop conditions;
- severity rules;
- reporting expectations;
- example prompt requirement.

## Stage Gate Standard

Every stage must answer:

- What is the goal?
- What inputs are required?
- What actions are allowed?
- What actions are forbidden?
- What should the agent produce?
- What proves the stage is done?
- What mistakes commonly happen?
- What prompt should start the stage?

An agent should not enter the next stage until required artifacts and gate checks are complete or explicitly marked `Unknown / needs confirmation`.

## Severity Model

| Severity | Meaning | Agent behavior |
| --- | --- | --- |
| `P0` | Blocks safe launch, deploy, client handoff, or core flow. | Stop or fix before moving forward. |
| `P1` | Important defect or missing requirement, but not an immediate launch blocker if work is accompanied. | Fix in the current process if in scope, otherwise document. |
| `P2` | Polish, refinement, or minor quality issue. | Defer unless the current stage explicitly includes polish. |
| `deferred` | Out of current process scope or intentionally later. | Record in deferred artifact; do not implement casually. |

## Required References

Do not duplicate global safety and validation rules. Process docs must link to:

- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)

## Reporting Standard

Every process must define the final report shape or reference the appropriate Nuvio OS reporting format.

At minimum, reports must include:

- files read;
- files changed;
- stage completed;
- artifacts produced or updated;
- validation performed or skipped;
- P0/P1/P2/deferred findings;
- unknowns;
- next safe stage.

## Stop Conditions

Agents must stop or report `Unknown / needs confirmation` when:

- required inputs are missing;
- a stage would require forbidden actions;
- source-of-truth docs conflict;
- implementation would cross into another stage;
- validation cannot be performed but is required for the gate;
- secrets, PII, runtime data, source blocks, or deployment state could be affected unexpectedly.
