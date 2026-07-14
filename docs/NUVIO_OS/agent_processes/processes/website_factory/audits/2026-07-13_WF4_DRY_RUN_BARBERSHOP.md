# WF4 Dry Run - Barbearia Exemplo

## Dry Run Summary

- Verdict: Successful dry run for Website Factory stages 0-4 in Reduced Website Factory Mode.
- Process usable? yes
- Reduced Mode usable? yes
- Human/Hermes intervention points:
  - Approve the sitemap/page plan before treating it as accepted.
  - Provide or approve the source block library path before exact block selection.
  - Approve block import before any future Stage 5 work.
  - Confirm whether WhatsApp/manual booking is the only v1 booking flow.

This dry run did not build a website, import blocks, inspect source blocks, edit implementation files, or proceed beyond Stage 4.

## Simulated SITE_BRIEF

| Field | Value |
| --- | --- |
| Business name | Barbearia Exemplo |
| Business type | Local barbershop / grooming studio |
| Location | Porto, Portugal |
| Target audience | Men aged 20-55 looking for haircuts, beard trims, grooming, and easy booking. |
| Primary goals | Professional online presence; WhatsApp contacts; booking encouragement; clear services; visible hours/location; fast simple site. |
| Primary CTA | Marcar pelo WhatsApp |
| Secondary CTA | Ver servicos |
| Brand direction | Clean, premium but approachable; dark/neutral palette; no luxury overkill; mobile-first. |
| Pages needed | Home, Services, Contact |
| Integrations needed | WhatsApp CTA/contact path. Booking is manual/WhatsApp for v1. |
| Explicitly out of scope | Online payment, e-commerce, newsletter, blog, advanced CMS mapping, source block import. |
| Content availability | Unknown / needs photos, service list, prices if public, opening hours, address, WhatsApp number, social links. |
| Status | Reduced-mode simulated brief complete enough for Stage 1. |

Open questions:

- Exact WhatsApp number and preferred prefilled message.
- Exact opening hours.
- Exact address and parking/landmark notes.
- Final service list and whether prices should be public.
- Whether the business has approved photos or needs placeholder-free visual direction.

## Simulated WEBSITE_AUDIT

| Area | Finding |
| --- | --- |
| Current site/status | Unknown / no current website provided. Treat as first site or replacement unknown. |
| Strengths | Clear local service; simple audience; strong direct CTA; services are easy to explain. |
| Problems | No confirmed assets, hours, address, or service list. Booking flow is manual, so expectations must be clear. |
| Content gaps | Photos, service menu, opening hours, address, WhatsApp number, barber/team details, trust signals, social links. |
| Conversion issues | Visitors need immediate service clarity, fast WhatsApp path, and confidence that the shop is real/local. |
| SEO basics | Local search intent matters: Porto, barbearia, cortes, barba, grooming. Do not overbuild SEO for v1. |
| Technical risks | WhatsApp link cannot be validated without number; exact booking behavior must not imply online slot booking. |
| Opportunities | Mobile-first single CTA; services overview; location/hours; trust-focused visuals; simple contact page. |
| Recommended direction | Build a lean three-page site with strong Home, clear Services, and Contact page optimized for WhatsApp/manual booking. |

Stage 1 result:

- No sitemap locked yet during audit.
- No copywriting, block selection, or implementation performed.

## Simulated SITEMAP

| Page | Route | Purpose | Primary CTA | Secondary CTA | Priority | Dependencies | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Home | `/` | Explain the barbershop quickly, establish trust, route visitors to WhatsApp or services. | Marcar pelo WhatsApp | Ver servicos | P0 | WhatsApp number, service categories, hours/location summary, approved imagery direction. | Proposed / needs Hermes approval |
| Services | `/services` | Show haircut, beard, and grooming services clearly with simple expectations. | Marcar pelo WhatsApp | Contactos | P0 | Final service list, optional prices/duration, service descriptions. | Proposed / needs Hermes approval |
| Contact | `/contact` | Make WhatsApp/manual booking, hours, location, and next step obvious. | Marcar pelo WhatsApp | Ver servicos | P0 | WhatsApp number, address, opening hours, map/link if approved. | Proposed / needs Hermes approval |

Deferred:

| Item | Reason | Revisit trigger |
| --- | --- | --- |
| Blog | Explicitly out of scope for v1. | SEO/content phase requested later. |
| Newsletter | Explicitly out of scope for v1. | Business asks for retention campaigns. |
| Online payment/e-commerce | Explicitly out of scope. | Business decides to sell products or deposits online. |
| Online booking slots | v1 uses WhatsApp/manual booking. | Business wants structured booking module. |

Gate note:

- The process correctly requires Hermes/Leonardo approval before treating this sitemap as accepted.

## Simulated PAGE_BLUEPRINTS

### Home

| Section | Goal | Content needed | CTA role | Visual role | Suggested block type | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Hero | Explain the shop and primary action within seconds. | Business name, short promise, location, CTA labels. | Primary WhatsApp CTA; secondary Services anchor/link. | Strong dark/premium opening, mobile-first. | Hero with CTA and image/mockup area. | Do not over-luxury the tone. |
| Trust / quick facts | Show practical confidence markers. | Porto location, opening hours summary, service highlights. | Support CTA decision. | Compact badges/cards. | Stats/facts strip. | Keep factual only. |
| Services preview | Show core categories. | Haircut, beard trim, grooming combos. | Link to Services. | 3-card grid. | Service cards. | No detailed pricing unless confirmed. |
| How booking works | Clarify manual WhatsApp flow. | Step 1 choose service, Step 2 message on WhatsApp, Step 3 confirm time. | Reduce friction. | Simple steps. | Process steps block. | Must not imply online slot booking. |
| Location/hours preview | Make physical visit practical. | Address, hours, map/link if approved. | Secondary route to Contact. | Split contact/location block. | Contact/location preview. | Use placeholders until confirmed. |
| Final CTA | Convert. | Short reassurance and WhatsApp CTA. | Primary conversion. | CTA band. | Final CTA block. | Keep direct. |

### Services

| Section | Goal | Content needed | CTA role | Visual role | Suggested block type | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Services hero | Position services simply. | Headline, brief service promise. | WhatsApp CTA. | Clean header. | Page hero. | No generic salon language. |
| Service categories | Show what can be booked. | Final list of services and optional prices/durations. | Encourage inquiry. | Cards/list. | Pricing/service cards. | Prices unknown; mark as optional. |
| What to expect | Reduce uncertainty. | Appointment/manual booking expectations, walk-in policy if any. | Support contact. | Checklist or steps. | Checklist block. | Must be factual. |
| Grooming note | Add premium approachable feel. | Quality/hygiene/style note. | Soft trust builder. | Split text/image. | Feature split. | Needs approved imagery or no photo. |
| CTA | Convert. | WhatsApp booking CTA. | Primary conversion. | CTA band. | Final CTA block. | Link to Contact optional. |

### Contact

| Section | Goal | Content needed | CTA role | Visual role | Suggested block type | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Contact hero | Make next step obvious. | WhatsApp/manual booking message. | WhatsApp CTA. | Direct page header. | Contact hero. | Avoid adding form unless requested. |
| Contact options | Show WhatsApp, phone/social if available. | WhatsApp number, Instagram, phone if public. | Primary contact. | Contact cards. | Contact option cards. | Only show confirmed channels. |
| Hours/location | Make visit easy. | Address, hours, map link. | Practical support. | Location details. | Location/hours block. | Needs exact data. |
| What to send | Improve WhatsApp message quality. | Service, preferred day/time, name. | Reduce back-and-forth. | Checklist. | Checklist block. | Good for manual booking. |
| Final reassurance | Lower friction. | No online payment, manual confirmation, response expectation if known. | Final CTA. | Compact reassurance. | CTA/reassurance band. | Response time unknown. |

Stage 3 result:

- Blueprint stayed section-level.
- No source blocks selected.
- No final copy written.
- No routes or implementation created.

## Simulated BLOCK_SELECTION

Source block library path:

```text
Unknown / needs source block library path
```

Because the approved source block library path is unknown in this dry run, exact source paths were not invented. The selection below records required block types and how exact block choices should be completed once source blocks are available.

| Target page/section | Needed block type | Selection status | Source path | Fit notes | Risks / dependencies |
| --- | --- | --- | --- | --- | --- |
| Home / Hero | Dark premium hero with CTA and visual area | Pending source selection | Unknown / needs source block library path | Should support local business positioning and WhatsApp CTA. | Needs approved visual/photo or neutral CSS mockup. |
| Home / Trust quick facts | Compact stats/facts strip | Pending source selection | Unknown / needs source block library path | Should show Porto, hours summary, and services quickly. | Needs confirmed hours/address. |
| Home / Services preview | 3-card service grid | Pending source selection | Unknown / needs source block library path | Should map cleanly to haircut/beard/grooming categories. | Avoid over-featured SaaS cards. |
| Home / How booking works | 3-step process block | Pending source selection | Unknown / needs source block library path | Should explain WhatsApp/manual booking. | Must not imply online scheduling. |
| Home / Location preview | Split contact/location block | Pending source selection | Unknown / needs source block library path | Should make location/hours visible. | Needs exact address/hours. |
| Services / Service categories | Service/pricing cards or structured list | Pending source selection | Unknown / needs source block library path | Should support services with optional prices/durations. | Prices may be unavailable. |
| Services / What to expect | Checklist block | Pending source selection | Unknown / needs source block library path | Should reduce uncertainty. | Needs confirmed policies. |
| Contact / Contact options | Contact cards + CTA | Pending source selection | Unknown / needs source block library path | Should make WhatsApp primary. | Requires WhatsApp number. |
| Contact / Hours/location | Location/hours block | Pending source selection | Unknown / needs source block library path | Should support mobile-first contact behavior. | Needs exact address/hours. |
| Contact / What to send | Checklist block | Pending source selection | Unknown / needs source block library path | Should improve WhatsApp inquiry quality. | Needs preferred booking instructions. |

Source block safety:

| Check | Result |
| --- | --- |
| Source blocks inspected read-only | Not applicable; no source path provided. |
| Source files unchanged | yes |
| Import not started | yes |
| Exact source paths invented | no |
| Hermes/Leonardo approval needed before import | yes |

Stage 4 result:

- The process made it clear that exact block selection requires an approved source block library path.
- Block import did not start.
- No source block mutation risk occurred.

## Stage Gate Observations

| Stage | Gate clear? | Artifact practical? | Stage mixing risk? | Notes |
|---|---|---|---|---|
| 0 Intake | yes | yes | low | Reduced mode made the brief compact and usable. |
| 1 Audit | yes | yes | low | Clear enough to analyze without locking sitemap. |
| 2 Sitemap | yes | yes | low | Approval gate is clear; sitemap remains proposed. |
| 3 Page blueprint | yes | yes | medium-low | Blueprinting can tempt copywriting, but docs forbid final copy and exact block selection clearly. |
| 4 Block selection | yes | partial | low | Practical if source library path exists. Without it, the process correctly stops exact selection and records placeholders. |

## Process Gaps Found

- No blocker found.
- Real Stage 4 execution needs a standard way to identify the approved source block library path for a given website project.
- Reduced Mode is clear enough for a simple site and avoids artifact bloat.
- The process successfully prevented source-block import and stage mixing during the dry run.

## Patches Made

None.

The dry run did not reveal a blocker requiring process doc changes.

## Recommendation

Needs source block library path standardization before real use.

The Website Factory docs are usable for stages 0-4, but real block selection should not begin until the approved source block library path is provided in the task prompt, project brief, or a standard Nuvio OS source-library reference.

## Git Status

Command requested:

```powershell
git status --short --untracked-files=all
```

Expected WF4 delta:

```text
?? docs/NUVIO_OS/agent_processes/processes/website_factory/audits/2026-07-13_WF4_DRY_RUN_BARBERSHOP.md
```

Note:

The working tree already contained WF3 documentation modifications before this dry run. WF4 did not modify implementation/source/config/env/deploy files and did not patch Website Factory process docs.
