# Nuvio OS QA Checklist

## Purpose

Lightweight checklist for validating Nuvio OS itself. Use this after adding or changing Core, Feature, Operations, Launch, routing, or navigation docs.

This checklist validates the documentation operating layer only. It does not validate product behavior.

## Read first

- [CORE.md](CORE.md)
- [TASK_ROUTER.md](TASK_ROUTER.md)
- [DANGER_ZONES.md](DANGER_ZONES.md)
- [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md)
- [REPORTING_FORMATS.md](REPORTING_FORMATS.md)
- [OS_NAVIGATION.md](OS_NAVIGATION.md)

## File Existence

- [ ] [README.md](README.md) exists.
- [ ] [OS_NAVIGATION.md](OS_NAVIGATION.md) exists.
- [ ] [OS_QA_CHECKLIST.md](OS_QA_CHECKLIST.md) exists.
- [ ] [CORE.md](CORE.md) exists.
- [ ] [TASK_ROUTER.md](TASK_ROUTER.md) exists.
- [ ] [DANGER_ZONES.md](DANGER_ZONES.md) exists.
- [ ] [VALIDATION_MATRIX.md](VALIDATION_MATRIX.md) exists.
- [ ] [REPORTING_FORMATS.md](REPORTING_FORMATS.md) exists.
- [ ] [features/README.md](features/README.md) exists.
- [ ] [operations/README.md](operations/README.md) exists.
- [ ] [launch/README.md](launch/README.md) exists.

## Routing Coverage

- [ ] Every major task type has a routing path in [TASK_ROUTER.md](TASK_ROUTER.md) or [OS_NAVIGATION.md](OS_NAVIGATION.md).
- [ ] CMS, Assets, Website Settings/SEO, Leads, Booking, Newsletter, Reports, Public Runtime, Security, and Emails route to Feature Agent Cards.
- [ ] Deployment, Docker/Compose, Instance Bootstrap, Env/Secrets, Snapshot/Restore, Backup/Rollback, Release/Git, Smoke Validation, Umami, and Public Runtime Deployment route to Operations Agent Cards.
- [ ] Demo, first-client readiness, base deployment readiness, blocker/polish classification, and handoff route to Launch docs.
- [ ] High-risk areas route to [DANGER_ZONES.md](DANGER_ZONES.md).

## High-Risk Feature Coverage

- [ ] Booking has a Feature Agent Card.
- [ ] Newsletter has a Feature Agent Card.
- [ ] Security/client-role has a Feature Agent Card.
- [ ] Public runtime has a Feature Agent Card.
- [ ] Reports/analytics/health has a Feature Agent Card.
- [ ] Emails/templates has a Feature Agent Card.
- [ ] Assets/storage has a Feature Agent Card.

## Operations Coverage

- [ ] Deployment/Coolify has an Operations Agent Card.
- [ ] Docker/Compose has an Operations Agent Card.
- [ ] Env/Secrets has an Operations Agent Card.
- [ ] Snapshot/Restore has an Operations Agent Card.
- [ ] Backup/Rollback has an Operations Agent Card.
- [ ] Public Runtime Deployment has an Operations Agent Card.
- [ ] Umami Analytics Operations has an Operations Agent Card.
- [ ] Smoke Validation/Troubleshooting has an Operations Agent Card.

## Launch / Demo / First Client Coverage

- [ ] First-client readiness links to both Feature and Operations cards.
- [ ] Nuvio Base deployment readiness links to deployment, env, backup, public runtime, and security docs.
- [ ] Demo runbook covers the five critical demo flows.
- [ ] Launch blockers vs polish distinguishes launch-critical, readiness, polish, enhancement, and deferred.
- [ ] Agent handoff checklist requires files read, files changed, source docs used, commands run, validation, unknowns, risks, and next phase.

## Source-of-Truth Safety

- [ ] No doc claims current implementation status without source support or `Unknown / needs confirmation`.
- [ ] No doc overrides current source code or git status.
- [ ] No doc treats old backlog as current by default.
- [ ] No advanced/deferred scope is promoted into the launch path.
- [ ] No doc treats polish as a launch blocker.
- [ ] No doc presents manual process as unsafe; manual is acceptable when controlled and verified.

## Link Sanity

Manual pass for obvious relative links:

- [ ] Top-level README links to Core, Task Router, Danger Zones, Validation Matrix, Reporting Formats, Features, Operations, Launch, OS Navigation, and OS QA Checklist.
- [ ] Feature README links back to OS home and navigation.
- [ ] Operations README links back to OS home and navigation.
- [ ] Launch README links back to OS home and navigation.
- [ ] Navigation guide links to relevant feature, operations, and launch docs.
- [ ] Task Router links to relevant feature, operations, launch, validation, and reporting docs.
- [ ] Every Feature, Operations, and Launch card has a breadcrumb path back to OS home, OS Navigation, and its layer README.
- [ ] Every Feature, Operations, and Launch card links to Core, Task Router, Validation Matrix, and Reporting Formats either directly or through its required source table.

## Scope Hygiene

- [ ] Documentation-only phases only modify allowed docs paths.
- [ ] Product code is not changed during Nuvio OS documentation phases.
- [ ] Configs, env files, migrations, package files, and build files are not changed during Nuvio OS documentation phases.
- [ ] No builds/tests are run unless explicitly requested.
- [ ] `git status --short --untracked-files=all` is reported.

## Agent Final Report Checklist

- [ ] Files created.
- [ ] Files modified.
- [ ] Source docs inspected.
- [ ] Navigation/routing improvements summarized.
- [ ] Unknowns or inconsistencies listed.
- [ ] Confirmation no product code/config/env/build files changed.
- [ ] Git status summary.
- [ ] Recommended next phase.

