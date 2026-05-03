# Nuvio Website Settings and SEO Contract

Website Settings and SEO must keep clear data boundaries.

## Data boundaries

Top-level website fields store identity and SEO data.

Examples:

- `logo`
- `seoTitle`
- `seoDescription`
- `seoImage`
- `seo_title_template`
- `seo_title_separator`
- `seo_canonical_domain`
- Local Business SEO fields

`websites.settings` stores feature configuration.

Examples:

- `whatsapp`
- `contactForm`
- `newsletter`
- `i18n`
- feature-specific settings

Do not move SEO fields into `websites.settings`.

Do not duplicate settings sources.

## Feature availability

Feature availability is controlled by admin/superuser configuration.

Client CMS must not expose feature availability controls.

Legacy `settings.<feature>.enabled` fields may exist and must be preserved, but they should not be shown as client-facing feature availability controls.

Client feature visibility should come from the admin-controlled availability source, such as `featureFlags`.

## Hidden key preservation

When saving `websites.settings`:

- preserve `featureFlags`
- preserve legacy hidden keys
- preserve admin-only groups
- preserve non-active feature groups
- do not overwrite settings with only visible fields
- do not remove hidden `enabled` keys

## Website Settings UI

Website Settings should be task-based.

Recommended structure:

- `Identity & SEO`
- `Features`

Inside `Features`, show feature group tabs such as:

- WhatsApp
- Contact form
- Reviews
- Newsletter
- Booking
- Reports
- Internationalization

If a feature is available but has no client-configurable fields, show a friendly empty state.

Do not show raw JSON or technical keys.

## SEO rules

SEO is a core Nuvio value proposition.

Do not treat SEO as just title and description.

SEO areas:

- Page SEO
- Global SEO defaults
- Social sharing
- Advanced indexing
- Local Business SEO
- Structured data
- Future i18n/hreflang

## Page SEO fields

Page-level SEO fields should stay on the `pages` collection.

Examples:

- `seo_title`
- `seo_description`
- `seo_social_image`
- `seo_canonical_url`
- `seo_noindex`
- `seo_exclude_from_sitemap`
- `seo_focus_keyword`

`seo_focus_keyword` is internal/helper-only.

Do not render it as meta keywords.

Do not add meta keywords.

## Runtime SEO

Do not implement runtime SEO during admin UI tasks unless explicitly requested.

Runtime SEO includes:

- `<title>`
- meta description
- OG tags
- Twitter tags
- canonical
- robots/noindex
- sitemap
- JSON-LD
- LocalBusiness schema
- hreflang

## SEO UI

SEO UI should include:

- live search preview
- fallback explanations
- length hints
- non-blocking warnings
- clear advanced sections
- visual-only checks that do not block saving

Do not change save behavior while doing visual SEO work.