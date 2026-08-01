# Reporting Formats

Use these templates when reporting back to Leonardo/Hermes. Keep reports concise, concrete, and honest.

Every non-trivial report must include:

- goal;
- files read;
- files changed;
- decisions used;
- tests/builds run;
- manual checks;
- risks;
- blockers or unknowns;
- next recommended step;
- explicit what I did not change.

## 1. Standard Implementation Report

```text
Goal:
- <what was requested>

Files read:
- <paths>

Files changed:
- <paths>

Decisions used:
- <canonical decisions/contracts applied>

What changed:
- <behavior/UI/docs summary>

What I did not change:
- <explicit non-touched areas>

Tests/builds run:
- <commands and results>

Manual checks:
- <manual smoke/visual checks or not run with reason>

Risks:
- <remaining risk or none>

Blockers/unknowns:
- <items or none>

Next recommended step:
- <one scoped next step, if any>
```

## 2. Audit-Only Report

```text
Goal:
- <audit purpose>

Files read / sources inspected:
- <paths/docs/files>

Files changed:
- <usually audit file only>

Decisions used:
- <source-of-truth order or audit criteria>

Findings:
- <top findings>

Canonical decisions or recommendations:
- <decisions/recommendations>

Tests/builds run:
- <usually none for docs-only audit; explain if none>

Manual checks:
- <file existence/link/path sanity/git status checks>

Risks:
- <remaining risk or none>

Blockers/unknowns:
- <items>

What I did not change:
- Product code unchanged.
- Existing docs unchanged unless explicitly requested.

Next recommended step:
- <phase name and scope>
```

## 3. Regression Report

```text
Goal:
- <regression area>

Files read:
- <paths>

Files changed:
- <paths or none>

Decisions used:
- <contracts/source-of-truth/danger-zone rules>

Baseline checked:
- <current code/docs/tests>

Findings:
- Critical:
- High:
- Medium:
- Low:

No findings:
- <state explicitly if no findings>

Tests/builds run:
- <commands and results>

Manual checks:
- <manual smoke/role/visual checks>

Risks:
- <untested areas>

Blockers/unknowns:
- <items or none>

What I did not change:
- <confirm audit-only unless fixes were requested>

Next recommended step:
- <if any>
```

## 4. Deployment Report

```text
Goal:
- <deployment task>

Deployment target/context:
- <local/Coolify/staging/production-like>

Files read / docs used:
- <deployment docs/runbooks/source files>

Files changed:
- <paths or none>

Decisions used:
- <VITE/server-only/secrets/origins/volumes/snapshot decisions>

Env/build/runtime decisions:
- <VITE/server-only/secrets/origins/volumes>

Tests/builds run:
- <compose/build/health commands and results>

Manual checks:
- Backend health:
- Backoffice login:
- Public runtime:
- CMS preview:
- Assets:
- Contact/newsletter/booking/reports as applicable:

Risks:
- <remaining deploy risk>

Blockers/unknowns:
- <DNS, secrets, restore mechanism, backup target, etc.>

What I did not change:
- <no secrets, no code, no restore automation, etc.>

Next recommended step:
- <safe next deploy step>
```

## 5. Security Review Report

```text
Goal:
- <security/client-role/public endpoint area>

Files read / docs inspected:
- <paths>

Files changed:
- <paths or none>

Decisions used:
- <scoped endpoint, websiteAccess, public endpoint, logging/redaction rules>

Threat/risk focus:
- <auth, websiteAccess, raw PB, tokens, PII, CORS/CSP, provider secrets>

Findings:
- Critical:
- High:
- Medium:
- Low:

Required fixes:
- <if any>

Tests/builds run:
- <backend/UI commands and results>

Manual checks:
- <client-role/public endpoint/log checks>

Risks:
- <remaining risk>

Blockers/unknowns:
- <unknowns>

What I did not change:
- <schema/backend/UI/etc.>

Next recommended step:
- <if any>
```

## 6. Documentation Update Report

```text
Goal:
- <doc phase>

Files read / source docs used:
- <paths>

Files created:
- <paths>

Files changed:
- <paths>

Decisions used:
- <source hierarchy/canonical decisions>

Key decisions encoded:
- <bullets>

Tests/builds run:
- <usually none for docs-only; explain if none>

Manual checks:
- <file existence, links/path plausibility, git status>

Risks:
- <remaining doc/source risk>

Blockers/unknowns:
- <items>

Stale docs not modified:
- <paths/reasons>

What I did not change:
- Product code unchanged.
- Existing docs unchanged unless requested.

Next recommended phase:
- <phase>
```

## 7. Blocked / Unknown Report

```text
Goal:
- <requested outcome>

Files read:
- <paths/docs/code>

Files changed:
- <paths or none>

Decisions used:
- <source hierarchy/danger-zone rules>

Blocking condition:
- <what prevents safe progress>

What is known:
- <facts>

Unknown / needs confirmation:
- <specific questions or missing inputs>

Safe alternatives attempted:
- <if any>

Tests/builds run:
- <commands or not run with reason>

Manual checks:
- <checks or not run with reason>

Risks:
- <risk if proceeding without answer>

Blockers:
- <explicit blocker summary>

What I did not change:
- <confirm no unsafe changes>

User decision needed:
- <specific decision/input>

Next recommended step:
- <smallest safe unblock>
```

## Extra Reporting Rules

- If a task touches a danger zone, explicitly name which danger-zone rules were applied.
- If validation is skipped, explain why and state remaining risk.
- If no files changed, say so explicitly.
- If old docs conflict with current code, say which source won.
- If pricing or tier context is needed, read the central Nuvio OS `TIERS_AND_PRICING.md`; mark `Unknown / needs confirmation` only for decisions not covered there.
