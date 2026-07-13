# Nuvio OS Core

Read this before working on Nuvio.

## Project Identity

Nuvio is a proprietary website/CMS product for small businesses. It combines:

- a PocketBase-based backoffice/backend;
- scoped Nuvio APIs;
- CMS/page/block/assets editing;
- public website rendering;
- leads/contact/WhatsApp;
- booking;
- newsletter;
- reports/analytics;
- deployment and snapshot tooling.

## Current Operating Mode

Nuvio is in release-readiness mode, not broad backlog mode.

Current focus:

- production-like deployment;
- Coolify/base instance readiness;
- CMS snapshot restore;
- smoke testing;
- own Nuvio website/landing;
- Umami analytics validation;
- security/client-role validation;
- public endpoint hardening;
- email/newsletter/booking E2E validation;
- reports confidence;
- demo data and demo flow;
- first-client onboarding.

## Architecture Summary

| Layer | Canonical role |
| --- | --- |
| Backoffice/backend | Central reusable PocketBase fork/custom app with embedded admin UI, scoped Nuvio APIs, migrations, and `pb_data`. |
| Public runtime `cms5` | Current public runtime/test app and runtime behavior reference. Useful, but not the clean starter template. |
| Reference repo | Clean public-site reference/template for future public websites. Canonical for new site starter direction. |
| Real public websites | Separate apps/repos/instances created from Reference or aligned with its contracts. |
| Instance | Per-client/deployment env, domains, `pb_data`, storage, backups, and CMS snapshot. |

## Source-of-Truth Order

1. Current source code and current git status.
2. Main repo-local docs/contracts.
3. Main repo deployment docs.
4. Reference public-site docs for template direction.
5. cms5 docs/code notes for runtime/testing history.
6. Obsidian Nuvio Operating Manual for product/operations context.
7. External Hermes/ChatGPT context only if explicitly provided.
8. Old backlog/roadmap notes only after validation.

## Global Do-Not-Change Rules

- Do not revive old backlog as active work.
- Do not invent architecture when current code supports the task.
- Do not change unrelated modules.
- Do not add raw PocketBase writes for client-role product flows.
- Do not rely on UI hiding as security.
- Do not expose secrets in `VITE_*` or browser/client code.
- Do not make Reference or real sites depend on `cms5` or `Srcs` at runtime.
- Do not run destructive restore/reset paths casually.
- Do not invent final pricing.
- Do not treat default/root READMEs as Nuvio truth without verification.

## Default Task Classification

Classify every task before changing files:

| Class | Meaning | Default action |
| --- | --- | --- |
| Launch-critical | Needed for deploy/demo/first-client readiness. | Prioritize and validate carefully. |
| Readiness | Improves safety, repeatability, smoke tests, docs, or deployment confidence. | Usually acceptable if scoped. |
| Polish | UI/copy/small workflow improvement. | Keep local and avoid logic changes. |
| Bugfix | Current behavior contradicts contract or expected product behavior. | Inspect source/tests, fix smallest safe path. |
| Deferred | Parked until after base deployment or explicit revival. | Do not implement unless user explicitly scopes it. |
| Unsafe/unknown | Could affect auth, data, public endpoints, restore, booking, secrets, migrations, or cross-module behavior. | Audit first or ask for confirmation. |

## Five Critical Demo Flows

1. Website settings/setup.
2. CMS + SEO + public rendering.
3. Leads / Contact / WhatsApp.
4. Booking.
5. Reports / Analytics / Health.

These flows matter more than broad feature expansion.

## Before Changing Anything Checklist

- [ ] Confirm target repo and allowed files.
- [ ] Check current git status.
- [ ] Read `AGENTS.md` for the target repo if present.
- [ ] Read task-specific contracts via `TASK_ROUTER.md`.
- [ ] Confirm whether the task is launch-critical, readiness, polish, bugfix, deferred, or unsafe/unknown.
- [ ] Inspect current source and tests before trusting docs.
- [ ] Identify danger zones before editing.
- [ ] Decide validation commands before editing.

## After Changing Anything Checklist

- [ ] Verify changed files are within scope.
- [ ] Run required validation or clearly report why not run.
- [ ] Check git status/diff.
- [ ] Confirm no unrelated files changed.
- [ ] Report files read, files changed, decisions used, validation, risks, and what was not changed.
- [ ] If blocked or uncertain, mark `Unknown / needs confirmation` explicitly.