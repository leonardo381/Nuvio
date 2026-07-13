# Website Factory QA Checklist

## Purpose

Use this during Stage 11 QA and Stage 12 final polish. QA records findings; it does not silently fix or redesign.

## Severity Model

| Severity | Meaning | Process behavior |
| --- | --- | --- |
| P0 | Blocks safe launch, deploy, handoff, or a core public flow. | Stop or fix before proceeding. |
| P1 | Important defect or missing requirement. | Fix in current scope if allowed; otherwise record and gate. |
| P2 | Polish, refinement, or minor quality issue. | Defer unless the active stage includes polish. |
| deferred | Out of current stage or intentionally later. | Record; do not implement casually. |

## QA Scope

- implemented pages and routes;
- CTAs and internal links;
- responsive behavior;
- forms and integrations;
- SEO basics;
- analytics events if enabled;
- accessibility basics;
- performance basics;
- broken links;
- browser/device smoke;
- deploy smoke when applicable;
- no-secret checks;
- source block immutability.

## Responsive Checks

- Desktop, tablet-ish, and mobile widths.
- No clipped text.
- Cards stack correctly.
- CTAs remain tappable.
- Mockups remain readable or simplify gracefully.

## CTA Checks

- Primary CTA label is consistent.
- CTA destination is correct.
- Secondary CTA does not distract.
- Final CTA exists where expected.

## Forms Checks

- Form renders.
- Required fields are clear.
- Success and error states work.
- Payload uses confirmed helper/contract only.
- No secrets or private values are exposed.

## WhatsApp Checks

- CTA appears only when configured/safe.
- Link or redirect works.
- Tracking/source behavior matches contract if enabled.
- No raw phone/config internals are exposed unnecessarily.

## Booking Checks

- Services load if booking is enabled.
- Slot selection works.
- Submit behavior matches contract.
- Status behavior is backend/config controlled.
- Errors are clear.

## Newsletter Checks

- Subscribe route/form works if enabled.
- Confirm/unsubscribe lifecycle routes work if in scope.
- Copy does not overstate newsletter activity.

## SEO Basics

- Title and meta description are appropriate.
- Canonical behavior is correct.
- Robots/noindex behavior is correct.
- Sitemap behavior is correct if present.
- Social image/OG basics are present if in scope.

## Analytics Events

- Page tracking works if enabled.
- Events use approved names.
- No PII is sent.
- No provider secret is browser-exposed.

## Accessibility Basics

- Semantic headings are ordered reasonably.
- Links/buttons have clear text.
- Form labels are present.
- Focus state is visible.
- Images have meaningful alt or empty decorative alt.

## Performance Basics

- No unnecessary large images.
- No source-library runtime references.
- No avoidable dependency bloat.
- Page loads without console errors.

## Broken Link Checks

- Main nav links.
- CTA links.
- Footer/legal links.
- Sitemap/robots where applicable.
- External links open correctly and use placeholders only in docs.

## Browser / Device Smoke

- At least one Chromium-based browser.
- Mobile viewport.
- Any project-specific browser/device requirement from the brief.

## Deploy Smoke

- Public URL loads.
- Main pages load.
- Forms/CTAs behave as expected.
- Assets load.
- No mixed-content/CORS/frame errors.

## No-Secret Checks

- No real secrets in docs.
- No private tokens in artifacts.
- No `.env` committed.
- Browser-exposed env values are intentionally public.

## Source Block Immutability Checks

- Source library files unchanged.
- Only copied target files changed.
- Stage 5 import log confirms no forbidden changes.
