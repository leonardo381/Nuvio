# Phase 5A - Agent Usability & Source Coverage Audit

## 1. Executive verdict

The current Nuvio OS is usable by agents and good enough for real agent work, especially for scoped implementation, audit, deployment planning, and launch-readiness tasks.

It gives agents a clear entrypoint, a source hierarchy, task routing, feature/operations/launch cards, danger-zone warnings, validation expectations, and reporting formats. The strongest parts are the guardrails against unsafe architecture drift: raw PocketBase access, client-role security shortcuts, `VITE_*` secret leaks, Booking side effects, restore automation, and stale backlog revival.

The main weakness is not missing safety. It is task ergonomics. A careful agent can route correctly, but a rushed agent may still over-read broad source tables or miss the shortest useful path for a concrete real-world task like "landing v1 + Umami" or "first-client demo flow".

Top 3 risks:

1. Task routing is accurate but still broad. Agents may read too much or choose a nearby card instead of the best card bundle for real tasks.
2. Source coverage is strong, but Obsidian and older audit docs remain context sources that can lag code. Agents must still check current source/git status before implementation.
3. Validation guidance is good at the category level, but live/provider status and exact test commands often remain task-specific unknowns that agents must report honestly.

## 2. Scorecard

| Area | Score | Notes |
| --- | ---: | --- |
| Agent entrypoint | 4 | `README.md` explains what Nuvio OS is, what to read first, and what not to treat as truth. A new agent can start there safely. |
| Navigation/routing | 4 | `OS_NAVIGATION.md` and `TASK_ROUTER.md` cover major feature, operations, launch, and danger-zone routes. Concrete task packs would make this faster. |
| Source coverage | 4 | Repo contracts, deployment docs, deploy examples, Reference docs, cms5 context, and Obsidian docs are indexed. Live/current status must still be checked. |
| Feature cards | 4 | Cards are consistent, source-backed, and safety-oriented. They identify likely files/areas and forbidden changes well. |
| Operations cards | 4 | Deployment/env/restore/backup guidance is actionable and cautious. Real provider/DNS/secret state remains outside docs. |
| Launch layer | 4 | Launch docs connect product, operations, blockers, demo, and handoff well. They are useful for go/no-go thinking. |
| Guardrails | 5 | Strong repeated warnings for raw PB, Booking, secrets, restore, cms5/Reference boundaries, pricing, Reviews, and old backlog. |
| Validation guidance | 4 | Matrix and cards give good defaults. Exact commands still require source/package/test inspection per task. |
| Anti-staleness handling | 4 | Source hierarchy and canonical decisions are explicit. More task-level reminders to inspect current git/source would reduce mistakes. |
| Context efficiency | 3 | Useful redundancy helps safety, but agents may over-read. A task-pack layer would reduce context load without weakening guardrails. |

## 3. Source coverage matrix

| Source | Used by OS? | Current? | Risk | Recommended action |
| --- | --- | --- | --- | --- |
| `AGENTS.md` | Yes | Canonical for backoffice agent behavior | Low | Keep as mandatory read for backoffice tasks. |
| Current source code/tests/migrations/git status | Yes, via source hierarchy | Canonical when inspected | Medium | Keep reminding agents that docs do not override current source or dirty git state. |
| `docs/NUVIO_ADMIN_UI_CONTRACT.md` | Yes | Canonical for admin UI rhythm | Low | Keep direct links from UI/CMS/Leads/Booking/Newsletter/Security cards. |
| `docs/NUVIO_SCHEMAFORM_AND_FORMS_CONTRACT.md` | Yes | Canonical for SchemaForm/form boundaries | Low | Keep direct links from CMS/Assets/form-related routing. |
| `docs/NUVIO_WEBSITE_SETTINGS_AND_SEO_CONTRACT.md` | Yes | Canonical for settings/SEO data boundaries | Low | Keep direct links from CMS and Website Settings/SEO cards. |
| Deployment quick guide / env matrix / bootstrap checklist | Yes | Likely current, but deployment-specific | Medium | Continue pairing with current deployment files and real provider/env checks. |
| `docs/NUVIO_COOLIFY_BASE_DEPLOYMENT_PLAN.md` | Yes | Planning source; live status unknown | Medium | Treat as plan, not proof of deployment. Confirm Coolify/DNS/secrets before acting. |
| `deploy/README.md` and Compose examples | Yes | Likely current for local Compose | Low/Medium | Use for Docker/Compose tasks, but validate file paths and current Dockerfiles before changing. |
| Nuvio OS core docs | Yes | Current for agent routing | Low | Keep. Do not rewrite; evolve with small task-driven patches. |
| Feature cards | Yes | Current as agent guardrails | Low/Medium | Keep. Add only scoped updates when feature contracts change. |
| Operations cards | Yes | Current as operational guardrails | Low/Medium | Keep. Avoid embedding real secrets/provider assumptions. |
| Launch layer | Yes | Likely current readiness framing | Medium | Keep, but verify live status before claiming deploy/site/client readiness. |
| Obsidian Operating Manual | Yes | Possibly stale; human context | Medium | Use for product/operations context only. Current source/repo docs win on implementation facts. |
| Reference public site docs | Yes | Current for clean template direction | Medium | Use for future public-site template work, not automatically for cms5 runtime or real site repos. |
| cms5 docs/code | Yes, via source hierarchy | Current only when code confirms | Medium | Use for current runtime/testing history, not as canonical starter. |
| Nuvio official website repo docs/source | Indirectly via real target repo rule | Unknown / needs confirmation | Medium | For landing tasks, inspect actual target site repo before acting. Consider adding a task-pack reminder. |
| Business/pricing/Plan B source | Mentioned as unknown | Unknown / needs confirmation | Medium | Keep pricing cautious. Do not invent exact pricing or business claims. |
| Provider/live deployment state | Mentioned as unknown | Unknown / needs confirmation | High | Must be checked outside docs before deployment, smoke, Umami, backup, or first-client claims. |

## 4. Task dry-run results

| Task | Required docs | Missing/unclear info | Risk | OS sufficient? |
| --- | --- | --- | --- | --- |
| A. Build Nuvio landing v1 + Umami tracking | `README.md`, `CORE.md`, `TASK_ROUTER.md`, `SOURCE_OF_TRUTH.md`, `features/PUBLIC_RUNTIME.md`, `features/REPORTS_ANALYTICS_HEALTH.md`, `operations/UMAMI_ANALYTICS_OPERATIONS.md`, `operations/PUBLIC_RUNTIME_DEPLOYMENT.md`, `launch/FIRST_CLIENT_READINESS.md`, Reference template docs, real target site repo docs/source. | Exact target repo/current landing state; whether Umami belongs in this real site now; final pricing/business copy source. | Agents could edit Reference instead of real site, invent pricing, or expose analytics IDs/secrets incorrectly. | Partial. Routing exists, but a concrete landing/Umami task pack would reduce mistakes. |
| B. Run Booking E2E regression | `README.md`, `CORE.md`, `TASK_ROUTER.md`, `DANGER_ZONES.md`, `features/BOOKING.md`, `features/EMAILS_TEMPLATES.md`, `VALIDATION_MATRIX.md`, Obsidian Booking/Security docs, current booking tests/source. | Exact local environment availability; whether email provider is configured; current expected auto-confirm status per settings. | Booking touches slot/status/email trust. Running the wrong smoke or changing behavior during validation would be risky. | Yes. OS is sufficient if agent also inspects current code/tests. |
| C. Prepare first production-like deployment | `README.md`, `CORE.md`, `TASK_ROUTER.md`, `operations/DEPLOYMENT_COOLIFY.md`, `operations/DOCKER_COMPOSE.md`, `operations/ENV_SECRETS.md`, `operations/INSTANCE_BOOTSTRAP.md`, `operations/SNAPSHOT_RESTORE.md`, `operations/BACKUP_ROLLBACK.md`, `operations/PUBLIC_RUNTIME_DEPLOYMENT.md`, `launch/NUVIO_BASE_DEPLOYMENT_READINESS.md`, repo deployment docs, Obsidian Coolify/deployment docs. | Real Coolify account/payment/DNS/secrets/backup target/restore mechanism. | High if docs are mistaken for live deploy state. | Yes for planning/readiness. Partial for execution until live provider facts are confirmed. |
| D. Validate client-role security smoke | `README.md`, `CORE.md`, `TASK_ROUTER.md`, `DANGER_ZONES.md`, `features/SECURITY_CLIENT_ROLE.md`, affected feature cards, `VALIDATION_MATRIX.md`, Obsidian Security Hardening, current middleware/endpoint tests/source. | Which client-role fixture/user exists; assigned website(s); exact current smoke credentials. | UI-only security assumptions, raw PB writes, websiteAccess gaps, PII/token logs. | Yes. Strong guardrails and validation expectations. |
| E. Validate newsletter/email lifecycle | `README.md`, `CORE.md`, `TASK_ROUTER.md`, `DANGER_ZONES.md`, `features/NEWSLETTER.md`, `features/EMAILS_TEMPLATES.md`, `operations/ENV_SECRETS.md`, `VALIDATION_MATRIX.md`, deployment env docs, Obsidian Newsletter/Security docs, current newsletter tests/source. | Whether Resend is configured; safe test recipient/sender; token redaction evidence. | Provider side effects, lifecycle token leaks, changing send/campaign behavior while testing. | Yes for audit/smoke planning. Execution depends on provider/env facts. |
| F. Prepare first-client demo flow | `README.md`, `CORE.md`, `CURRENT_OPERATING_STATE.md`, `TASK_ROUTER.md`, `launch/DEMO_FLOW_RUNBOOK.md`, `launch/FIRST_CLIENT_READINESS.md`, `launch/LAUNCH_BLOCKERS_VS_POLISH.md`, `launch/READINESS_DECISION_MATRIX.md`, `operations/SMOKE_VALIDATION_TROUBLESHOOTING.md`, relevant feature cards, Obsidian Product State/Roadmap. | Current deployed instance status; demo data/snapshot; whether Nuvio landing/request-review path exists or manual substitute is chosen. | Agents may treat polish as blocker or skip a live smoke prerequisite. | Yes for planning. Partial for final go/no-go until live status is checked. |

## 5. High-priority gaps

1. No compact task-pack map for the most common real agent tasks.

   `TASK_ROUTER.md` routes by task type, and `OS_NAVIGATION.md` is a good jump table, but agents still need to assemble bundles for tasks like landing + Umami, first deployment, or demo preparation. This is the largest efficiency gap.

2. Live status remains intentionally unknown but could be missed by agents.

   Deployment, Nuvio Base online state, Nuvio official website publication, provider secrets, DNS, backups, and restore mechanism must be checked at task time. The OS says this in several places, but task-specific prompts should make the check unavoidable.

3. Public-site task boundaries are spread across Source of Truth, Public Runtime, Reference docs, cms5 notes, and real site repo instructions.

   This is accurate, but it is cognitively expensive. A future task-pack entry should spell out: Reference is template, cms5 is runtime/lab history, real site repos are separate, and `Srcs` is source material only.

## 6. Medium-priority improvements

- Add one short "common task packs" section or doc for 6-10 recurring tasks, with exact docs to read, docs not to over-read, forbidden changes, and validation defaults.
- Add a landing/public-site task pack that explicitly includes real target repo inspection, Reference docs, `Srcs` rules, pricing caution, and Umami secret boundaries.
- Add a deployment execution checklist that separates planning-doc status from live Coolify/provider status.
- Add a public endpoint hardening task pack that points to Security, Leads, Newsletter, Booking, Public Runtime, and current backend tests/source.
- Add a small "how to stop" rule for agents when docs say deployment is ready but provider state is unverified.
- Keep the existing repeated guardrails, but avoid copying long source tables into any new task-pack doc. Link to cards instead.

## 7. Things not to change

- Do not rewrite `CORE.md`. It is doing the right job as compressed context.
- Do not collapse `OS_NAVIGATION.md` and `TASK_ROUTER.md`. They overlap usefully: navigation is a jump map, router is a decision table.
- Do not remove repeated danger-zone reminders from feature and operations cards. For agents, this repetition prevents expensive mistakes.
- Do not shorten feature cards just for human aesthetics. They are structured well for implementation agents.
- Do not remove Obsidian references. They are useful context as long as the source hierarchy stays clear.
- Do not promote Obsidian or old backlog docs above repo contracts/current source.
- Do not turn Nuvio OS into product documentation. It should remain an agent operating layer.
- Do not add product-code validation commands to documentation-only phases unless explicitly requested.

## 8. Recommended Phase 5B

Recommended next phase: Phase 5B - Agent Task Packs and Dry-Run Acceptance Matrix.

Files to update:

- `docs/NUVIO_OS/TASK_ROUTER.md`
- `docs/NUVIO_OS/OS_NAVIGATION.md`
- `docs/NUVIO_OS/README.md`
- Optional new doc: `docs/NUVIO_OS/TASK_PACKS.md`
- Optional audit output: `docs/NUVIO_OS/audits/2026-06-18_AGENT_TASK_PACK_DRY_RUN.md`

Files not to touch:

- Product code.
- Config files.
- Env files.
- Migrations.
- Runtime data, `pb_data`, storage, snapshots, or deployment state.
- Existing feature/operations/launch cards unless a task-pack link is clearly needed.

Exact goals:

1. Create or add a compact task-pack map for recurring agent tasks:
   - Nuvio landing v1 + Umami.
   - Booking regression.
   - First production-like deployment.
   - Client-role security smoke.
   - Newsletter/email lifecycle validation.
   - First-client demo flow.
   - Snapshot/restore rehearsal.
   - Env/secrets audit.
   - CMS/SEO regression.
2. For each task pack, define:
   - required OS docs;
   - repo/Obsidian docs to consult;
   - what not to change;
   - required validation;
   - stop conditions;
   - final report format.
3. Link the task-pack doc from `README.md`, `OS_NAVIGATION.md`, and `TASK_ROUTER.md`.
4. Run a no-change dry run for the six Phase 5A simulated tasks and confirm the task packs route correctly.

Expected validation:

- Confirm new/updated docs exist.
- Run a relative Markdown link sanity check for `docs/NUVIO_OS`.
- Run `git status --short --untracked-files=all`.
- Confirm changes are limited to `docs/NUVIO_OS`.
- Do not run product builds/tests.
