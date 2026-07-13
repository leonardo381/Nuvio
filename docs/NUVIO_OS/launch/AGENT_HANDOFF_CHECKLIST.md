# Agent Handoff Checklist

> Navigation: [Nuvio OS](../README.md) | [OS Navigation](../OS_NAVIGATION.md) | [Launch Layer](README.md)

## Purpose

Define how one agent should hand work to another without losing source context, overstating success, hiding skipped validation, or blurring documentation-only and implementation phases.

## Read first

| Priority | Source | Path | Why read it |
| --- | --- | --- | --- |
| 1 | Core | [../CORE.md](../CORE.md) | Before/after checklist and source order. |
| 1 | Task Router | [../TASK_ROUTER.md](../TASK_ROUTER.md) | Task routing and stop conditions. |
| 1 | Danger Zones | [../DANGER_ZONES.md](../DANGER_ZONES.md) | Sensitive areas needing explicit warnings. |
| 1 | Validation Matrix | [../VALIDATION_MATRIX.md](../VALIDATION_MATRIX.md) | Validation rules and skipped-check format. |
| 1 | Reporting Formats | [../REPORTING_FORMATS.md](../REPORTING_FORMATS.md) | Canonical report sections. |
| 2 | Release and Git Workflow | [../operations/RELEASE_GIT_WORKFLOW.md](../operations/RELEASE_GIT_WORKFLOW.md) | Git and docs-only scope hygiene. |
| 2 | Smoke Validation | [../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md) | Verification expectations. |

## Handoff principle

Do not claim success without verification.

A good handoff should let the next agent answer:

- What was requested?
- What sources were trusted?
- What changed?
- What was verified?
- What remains unknown?
- What must not be touched next?

## Required handoff format

| Field | Required content |
| --- | --- |
| Phase | Exact phase title and scope. |
| Goal | One sentence goal. |
| Files read | Source docs and code paths inspected. |
| Files changed | Exact paths changed. |
| Source docs used | Canonical docs/contracts/cards used. |
| Commands run | Exact commands and results. |
| Validation performed | Build/test/manual smoke/link/status checks. |
| Unknowns | Items marked `Unknown / needs confirmation`. |
| Risks | Remaining risk, especially launch/security/deployment risk. |
| Non-actions | What was intentionally not changed/run. |
| Next recommended phase | One scoped next step. |

## Documentation-only phase rules

- [ ] Confirm allowed docs folder before editing.
- [ ] Do not change product code.
- [ ] Do not change configs.
- [ ] Do not change env files.
- [ ] Do not run builds/tests unless explicitly requested.
- [ ] Verify requested docs exist.
- [ ] Verify required links exist.
- [ ] Run git status.
- [ ] Report if prior untracked docs make status scope noisy.

## Implementation phase rules

- [ ] Confirm target repo and allowed files.
- [ ] Inspect current source and tests before editing.
- [ ] Identify danger zones.
- [ ] Decide validation before editing.
- [ ] Make the smallest safe change.
- [ ] Run required validation for touched area.
- [ ] Check git status/diff.
- [ ] Report what was not changed.

## Launch/readiness handoff rules

- [ ] Classify work as launch-critical, readiness, polish, enhancement, deferred, or distraction.
- [ ] Link the relevant Feature Agent Card.
- [ ] Link the relevant Operations Agent Card.
- [ ] State whether this moves Nuvio closer to deployment/demo/first-client delivery.
- [ ] Do not treat polish as launch blocker.
- [ ] Do not revive deferred backlog without explicit request.

## Validation reporting rules

If validation was run, report:

- command;
- working directory;
- result;
- key failure lines if failed;
- whether failure is related to the current work.

If validation was skipped, report:

- check name;
- why skipped;
- remaining risk;
- closest safe alternative, if any;
- whether user/operator input is needed.

## Common bad handoffs

- `Done` without changed files.
- `Build passed` without command output context.
- `Should work` without smoke checks.
- `Only docs changed` without git status.
- `No blockers` while unknown deployment inputs remain.
- Hiding that a command failed or was not run.
- Treating previous agent assumptions as source truth.

## Related docs

- [README](README.md)
- [Readiness Decision Matrix](READINESS_DECISION_MATRIX.md)
- [Launch Blockers vs Polish](LAUNCH_BLOCKERS_VS_POLISH.md)
- [Release and Git Workflow](../operations/RELEASE_GIT_WORKFLOW.md)
- [Smoke Validation](../operations/SMOKE_VALIDATION_TROUBLESHOOTING.md)
- [Security and Client Role](../features/SECURITY_CLIENT_ROLE.md)
