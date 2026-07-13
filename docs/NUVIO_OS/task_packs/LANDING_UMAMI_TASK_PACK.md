# Landing and Umami Task Pack

## Purpose
Use this task pack when the task involves the Nuvio official landing site, a real public website repo, acquisition/demo copy, or Umami tracking/analytics behavior for a public site.

## Task classification
- launch-critical
- readiness
- sales/demo

## Required first reads
- [Core](../CORE.md)
- [Task Router](../TASK_ROUTER.md)
- [Danger Zones](../DANGER_ZONES.md)
- [Validation Matrix](../VALIDATION_MATRIX.md)
- [Reporting Formats](../REPORTING_FORMATS.md)
- [Source of Truth](../SOURCE_OF_TRUTH.md)
- [Public Runtime](../features/PUBLIC_RUNTIME.md)
- [Reports, Analytics, and Health](../features/REPORTS_ANALYTICS_HEALTH.md)
- [Umami Analytics Operations](../operations/UMAMI_ANALYTICS_OPERATIONS.md)
- [Public Runtime Deployment](../operations/PUBLIC_RUNTIME_DEPLOYMENT.md)
- [First Client Readiness](../launch/FIRST_CLIENT_READINESS.md)

## Optional source docs
- Reference repo public-site contract and env contract.
- Reference repo template build, template adapter, and global source library docs.
- Real target site repo `AGENTS.md`, README, source routes, and package scripts.
- Obsidian Product State and Current Roadmap.
- Current Umami/provider configuration, if supplied by the user or deployment environment.


## Target selection rule
- For the official Nuvio landing, inspect an existing separate official Nuvio site repo first if one exists. Prefer that repo for implementation when it is the intended public website.
- Use Reference only as the clean template/contract source unless the explicit task is to improve the reusable template.
- Use cms5 only as lab/dev/runtime history unless the explicit task targets cms5.
- Create a new site app only if no suitable official site repo exists or the user explicitly asks for a new app.
- Serve through an existing public runtime path only if current docs/source prove that path owns the official landing.

## Default minimal v1 scope
- Keep v1 publicable and narrow: hero, problem, solution, product preview or feature cards, how it works, request website review CTA, Umami pageview tracking, and a safe CTA click event if technically straightforward.
- Exclude pricing finalization, blog, multi-page SEO expansion, advanced animation, fake case studies, advanced lead pipeline, self-service signup, billing, marketplace/templates, and complex automation unless explicitly scoped.

## Preconditions
- Target repo/app is explicitly identified: Reference, cms5, Nuvio official site, or client site.
- Task mode is known: docs-only, static UI, implementation, or live validation.
- Umami scope is known: not used, public tracking, Reports provider validation, or both.
- Final pricing/business copy source is provided, or pricing remains cautious and non-final.
- Allowed files and validation commands are clear for the target repo.

## Source-of-truth rules
1. Current target repo source code and git status win.
2. Target repo `AGENTS.md` and repo docs win for implementation boundaries.
3. Nuvio OS docs route the work and define guardrails.
4. Reference docs guide clean template direction, not every real site choice.
5. Obsidian docs and business notes are context only unless current and explicit.

## Allowed work
- Audit landing/public-site source boundaries.
- Implement static public-site UI only when explicitly requested for the target repo.
- Use `Srcs` as source material only when the target repo permits it.
- Add browser-safe analytics IDs only if the contract confirms they are public.
- Document unknown business/pricing/provider facts instead of inventing them.

## Forbidden work
- Do not edit Reference when the target is a real site repo.
- Do not copy cms5 wholesale into a real site or Reference.
- Do not make any site depend on `Srcs` at runtime.
- Do not put provider API keys, usernames, passwords, or private tokens in browser env.
- Do not invent exact pricing, guaranteed lead promises, or unsupported analytics claims.
- Do not connect landing content to CMS unless explicitly scoped.

## Danger zones
- cms5 vs Reference vs real-site repo boundary.
- Umami events can expose PII if visitor details are tracked.
- Analytics provider secrets must stay server-side.
- Business positioning can drift into unsupported promises.
- Public route/CTA changes can break contact/review request flows.

## Execution outline
1. Confirm target repo and task mode.
2. Read required OS docs and target repo instructions.
3. Inspect current public-site routes, env docs, and package scripts before proposing commands.
4. Identify whether Umami work is public tracking, Reports provider setup, or documentation only.
5. Check copy for pricing, guarantees, and unsupported analytics claims.
6. If implementing, keep changes limited to the target site and run only target-approved validation.
7. Report unknowns and skipped checks.

## Validation checklist
### Doc validation
- Task pack and target docs are cited.
- Unknown business/provider facts are marked `Unknown / needs confirmation`.
- No runtime dependency on `Srcs` is introduced in docs.

### Code/build/test validation, if future implementation applies
- Run target public-site check/lint/build scripts when available.
- Verify no provider secrets are exposed to browser/client code.
- Validate CTA destinations and public route behavior.

### Manual smoke validation
- Open landing/home route.
- Open primary CTA route.
- Confirm request-review/contact path works or is clearly deferred.
- Confirm Umami behavior only if configured and safe to test.

### User confirmation needed
- Target repo and allowed files.
- Whether Umami should be enabled now.
- Whether any analytics ID is intentionally public.
- Canonical pricing/business copy, if exact pricing is requested.

## Expected report format
- Files read.
- Files changed.
- Target repo and why it was the correct target.
- What was verified for copy, routes, env, and analytics.
- Unknowns and risks.
- Confirmation no secrets or runtime dependencies on `Srcs` were added.
- Next recommended step.

## Stop conditions
- Target repo is unclear.
- The task asks for exact pricing without a canonical pricing source.
- The task would expose provider secrets in browser env.
- The task would edit Reference for a real site.
- Umami/provider status is required but unavailable.

