# Integration Checklist

## Purpose

Use this during Stage 9 to verify Nuvio integrations. This file lists process checks only. It does not replace Nuvio technical docs.

Read technical references first:

- [Public Runtime](../../../features/PUBLIC_RUNTIME.md)
- [Public Runtime Deployment](../../../operations/PUBLIC_RUNTIME_DEPLOYMENT.md)
- [Landing and Umami Task Pack](../../../task_packs/LANDING_UMAMI_TASK_PACK.md)
- [Validation Matrix](../../../VALIDATION_MATRIX.md)

## Global Integration Rules

- Do not guess endpoints or payloads.
- Do not expose server-only env values to browser/client code.
- Do not add unsupported backend fields.
- Do not connect disabled features.
- Do not modify backend schemas or routes as part of Website Factory unless a separate task explicitly scopes it.

## Contact Form

- Confirm form is needed.
- Confirm required fields.
- Confirm backend helper/endpoint contract.
- Confirm success/error states.
- Confirm source/page attribution if supported.
- Confirm no fake fields are sent.

## WhatsApp CTA

- Confirm WhatsApp is enabled and configured.
- Confirm destination/source tracking behavior.
- Confirm link is not shown if unsafe or unconfigured.
- Confirm copy sets expectation for direct contact.

## Booking

- Confirm booking is needed.
- Confirm service/slot flow exists and is configured.
- Confirm public booking submission behavior.
- Confirm no frontend-forced status unless contract requires it.
- Confirm confirmation/error states.

## Newsletter

- Confirm newsletter is enabled.
- Confirm subscribe/confirm/unsubscribe routes.
- Confirm copy does not imply campaigns are active if not configured.
- Confirm lifecycle links use correct public domain.

## SEO Basics

- Confirm title and description plan.
- Confirm canonical/public URL behavior.
- Confirm robots/noindex needs.
- Confirm sitemap needs.
- Confirm social image availability.
- Do not invent advanced SEO checks not supported by current data.

## Analytics / Umami

- Confirm analytics is enabled.
- Confirm event names and page tracking plan.
- Confirm no secrets are exposed.
- Confirm consent/legal requirements if applicable.
- Reference the Landing and Umami task pack for detail.

## Public Runtime Env

- Confirm public URL.
- Confirm backend/API URL.
- Confirm browser-safe vs server-only env split.
- Confirm no real secrets are committed.

## CMS Mapping

- Confirm which content must be editable.
- Confirm existing CMS fields/blocks support the map.
- Do not invent schema in this stage.

## Images / Assets

- Confirm assets are local or approved.
- Confirm alt text where meaningful.
- Confirm no runtime dependency on external source library paths.
- Confirm images do not expose private client data.

## Legal / Footer Links

- Confirm required links for the project scope.
- Confirm privacy/cookies/legal needs.
- Do not invent legal text.
