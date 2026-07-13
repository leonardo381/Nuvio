# Nuvio OS

## What Nuvio OS Is

Nuvio OS is an agent operating layer for Nuvio work. It helps AI implementation agents route tasks, choose source docs, avoid unsafe changes, classify launch readiness, and report work consistently.

It is designed for:

- Codex agents.
- Hermes agents.
- Future AI implementation agents.
- Humans supervising agent work who need a quick routing map.

## What Nuvio OS Is Not

- Not human product documentation.
- Not a product strategy rewrite.
- Not a replacement for current source code.
- Not a replacement for repo contracts or tests.
- Not permission to deploy, restore, migrate, or change dangerous areas.
- Not a backlog revival mechanism.

## Source-of-Truth Hierarchy

1. Current source code and current git status.
2. Repo docs/contracts.
3. Nuvio OS routing and safety docs.
4. Obsidian/human operating notes.
5. Old chat/history/backlog.

If sources conflict, do not guess. Mark `Unknown / needs confirmation` and report the conflict.

## Quick Start For Agents

1. Confirm target repo and allowed files.
2. Run or inspect current git status when allowed.
3. Read [CORE.md](CORE.md).
4. Read [TASK_ROUTER.md](TASK_ROUTER.md).
5. Check [DANGER_ZONES.md](DANGER_ZONES.md) before touching auth, public endpoints, booking, newsletter, reports, env, restore, migrations, or client-role paths.
6. If the request matches a recurring task, use [task_packs/README.md](task_packs/README.md) first.
7. Use [OS_NAVIGATION.md](OS_NAVIGATION.md) to jump to the right card for one-off or unusual tasks.
8. Use [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) before running or reporting checks.
9. Use [REPORTING_FORMATS.md](REPORTING_FORMATS.md) for the final response.

## If You Only Read 5 Files

| Order | File | Why |
| --- | --- | --- |
| 1 | [CORE.md](CORE.md) | Compressed source order, current mode, demo flows, and global rules. |
| 2 | [TASK_ROUTER.md](TASK_ROUTER.md) | Routes task type to docs, dangers, validation, and report format. |
| 3 | [DANGER_ZONES.md](DANGER_ZONES.md) | Identifies areas that can leak data, break deployment, or cause destructive changes. |
| 4 | [OS_NAVIGATION.md](OS_NAVIGATION.md) | One-jump map to feature, operations, and launch docs. |
| 5 | [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Defines what must be checked or honestly skipped. |

Use [REPORTING_FORMATS.md](REPORTING_FORMATS.md) before writing the final answer.

## Core Pack

| File | Use when |
| --- | --- |
| [CORE.md](CORE.md) | Always. Mandatory compressed context. |
| [SOURCE_OF_TRUTH.md](SOURCE_OF_TRUTH.md) | Sources conflict or multiple repos are involved. |
| [CURRENT_OPERATING_STATE.md](CURRENT_OPERATING_STATE.md) | Classifying launch/readiness/polish/deferred work. |
| [CANONICAL_DECISIONS.md](CANONICAL_DECISIONS.md) | Preserving architecture/product decisions. |
| [TASK_ROUTER.md](TASK_ROUTER.md) | Selecting docs, forbidden areas, validation, and reports by task type. |
| [DANGER_ZONES.md](DANGER_ZONES.md) | Avoiding sensitive regressions. |
| [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) | Choosing checks and reporting skipped checks. |
| [REPORTING_FORMATS.md](REPORTING_FORMATS.md) | Standardizing agent output. |

## Agent Layers

| Layer | Entry point | Use when |
| --- | --- | --- |
| Feature Agent Cards | [features/README.md](features/README.md) | Work touches CMS, Assets, Leads, Booking, Newsletter, Reports, Public Runtime, Security, or Emails. |
| Operations Agent Cards | [operations/README.md](operations/README.md) | Work touches deployment, Docker/Compose, env/secrets, instances, snapshot/restore, backup/rollback, git/release, smoke validation, Umami, or public runtime deployment. |
| Launch / Readiness Layer | [launch/README.md](launch/README.md) | Work touches demo readiness, first-client readiness, base deployment readiness, blockers vs polish, demo flow, or agent handoff. |
| Recurring Task Packs | [task_packs/README.md](task_packs/README.md) | Work matches a common recurring task such as first deployment, five-flow regression, client-role smoke, Booking E2E, Newsletter lifecycle, env/secrets, snapshot/restore, demo flow, or first-client onboarding. |
| Agent Process Layer | [agent_processes/README.md](agent_processes/README.md) | Work is a repeatable multi-stage process with gates, artifacts, and strict stage boundaries, such as Website Factory. |
| One-jump navigation | [OS_NAVIGATION.md](OS_NAVIGATION.md) | You know the task type but need the right docs fast. |
| OS QA checklist | [OS_QA_CHECKLIST.md](OS_QA_CHECKLIST.md) | You are checking whether Nuvio OS itself is navigable and safe. |

## Task Classification Rule

Every task must be classified before action:

| Class | Meaning |
| --- | --- |
| Launch-critical | Needed for deploy/demo/first-client readiness. |
| Readiness | Improves safety, repeatability, smoke tests, docs, or deployment confidence. |
| Polish | UI/copy/small workflow improvement that should stay local. |
| Enhancement | Useful later but not required for first client. |
| Deferred | Parked until explicit revival. |
| Unsafe/unknown | Requires audit or confirmation before implementation. |

Use [launch/READINESS_DECISION_MATRIX.md](launch/READINESS_DECISION_MATRIX.md) and [launch/LAUNCH_BLOCKERS_VS_POLISH.md](launch/LAUNCH_BLOCKERS_VS_POLISH.md) for launch/readiness decisions.

## Non-Negotiable Rules

- Do not revive old backlog automatically.
- Do not treat polish as a launch blocker.
- Preserve current architecture and do not introduce new patterns without audit.
- Do not add raw PocketBase writes for client-role product flows.
- Do not rely on UI hiding as security.
- Do not expose secrets in `VITE_*` or browser/client code.
- Do not run destructive restore/reset/migration/deploy paths casually.
- Do not make Reference or real sites depend on `cms5` or `Srcs` at runtime.
- Do not claim success without verification.

## Audits

- [audits/2026-06-17_SOURCE_INVENTORY.md](audits/2026-06-17_SOURCE_INVENTORY.md)
- [audits/2026-06-17_CONFLICT_STALENESS_AUDIT.md](audits/2026-06-17_CONFLICT_STALENESS_AUDIT.md)
- [audits/2026-06-17_CORE_PACK_QA.md](audits/2026-06-17_CORE_PACK_QA.md)

## Final Reporting

Use [REPORTING_FORMATS.md](REPORTING_FORMATS.md). For docs-only phases, report:

- files read;
- files created/modified;
- source docs used;
- verification performed;
- unknowns/conflicts;
- confirmation no product code/config/env/build files changed;
- git status summary;
- next recommended phase.



