package main

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestNuvioMarketingSeedFixtureContract(t *testing.T) {
	fixture, err := loadNuvioMarketingSeedFixture("")
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}
	if err := validateNuvioMarketingSeedFixture(fixture); err != nil {
		t.Fatalf("fixture contract failed: %v", err)
	}

	if fixture.Website.Slug != "nuvio" {
		t.Fatalf("expected website slug nuvio, got %q", fixture.Website.Slug)
	}
	if len(fixture.Components) != 17 {
		t.Fatalf("expected 17 component definitions, got %d", len(fixture.Components))
	}

	expectedPreviewRoutes := map[string]string{"home": "/", "services": "/services", "pricing": "/pricing", "contact": "/contact"}
	previewRoutes := nuvioMarketingSeedMapValue(fixture.Website.Settings["previewRoutes"])
	for pageSlug, expectedPath := range expectedPreviewRoutes {
		if previewRoutes[pageSlug] != expectedPath {
			t.Fatalf("expected preview route %s=%q, got %q", pageSlug, expectedPath, previewRoutes[pageSlug])
		}
	}

	i18n := nuvioMarketingSeedMapValue(fixture.Website.Settings["i18n"])
	if i18n["enabled"] != true {
		t.Fatalf("expected fixture i18n to be enabled, got %#v", i18n["enabled"])
	}
	if i18n["defaultLanguage"] != "pt-PT" {
		t.Fatalf("expected fixture default language pt-PT, got %#v", i18n["defaultLanguage"])
	}
	languages, ok := i18n["languages"].([]any)
	if !ok || len(languages) != 2 {
		t.Fatalf("expected fixture i18n languages pt-PT/en, got %#v", i18n["languages"])
	}
	for index, expectedCode := range []string{"pt-PT", "en"} {
		language := nuvioMarketingSeedMapValue(languages[index])
		if language["code"] != expectedCode {
			t.Fatalf("expected fixture language %d to be %q, got %#v", index, expectedCode, language["code"])
		}
	}

	expectedBlocks := map[string]int{"home": 9, "services": 6, "pricing": 4, "contact": 1}
	for _, page := range fixture.Pages {
		if expectedBlocks[page.Slug] != len(page.Blocks) {
			t.Fatalf("expected %d blocks for %s, got %d", expectedBlocks[page.Slug], page.Slug, len(page.Blocks))
		}
		seoTranslations := nuvioMarketingSeedMapValue(page.SeoTranslations)
		englishSeo := nuvioMarketingSeedMapValue(seoTranslations["en"])
		if fmt.Sprint(englishSeo["title"]) == "" || fmt.Sprint(englishSeo["description"]) == "" {
			t.Fatalf("expected English SEO translation for %s, got %#v", page.Slug, englishSeo)
		}
		for _, block := range page.Blocks {
			englishProps := nuvioMarketingSeedMapValue(block.Translations["en"])
			if len(englishProps) == 0 {
				t.Fatalf("expected English block translation for %s.%s", page.Slug, block.Slot)
			}
		}
	}

	homePage := findNuvioMarketingSeedPage(fixture, "home")
	if homePage == nil {
		t.Fatal("expected home fixture page")
	}
	homeHero := findNuvioMarketingSeedBlock(*homePage, "nuvio-home-hero")
	if homeHero == nil {
		t.Fatal("expected home hero fixture block")
	}
	if fmt.Sprint(homeHero.Props["headingPrefix"]) != "Seja encontrado, transmita confian\u00e7a e transforme visitantes em" {
		t.Fatalf("expected PT home hero props, got %#v", homeHero.Props["headingPrefix"])
	}
	englishHomeHero := nuvioMarketingSeedMapValue(homeHero.Translations["en"])
	if fmt.Sprint(englishHomeHero["headingPrefix"]) != "Get found, look professional, and turn visitors into" {
		t.Fatalf("expected EN home hero translation, got %#v", englishHomeHero["headingPrefix"])
	}
}

func TestNuvioMarketingSeedRejectsUnsafeTrustedMarkup(t *testing.T) {
	fixture, err := loadNuvioMarketingSeedFixture("")
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	servicesPage := findNuvioMarketingSeedPage(fixture, "services")
	if servicesPage == nil {
		t.Fatal("expected services page fixture")
	}
	heroBlock := findNuvioMarketingSeedBlock(*servicesPage, "nuvio-services-hero")
	if heroBlock == nil {
		t.Fatal("expected services hero block")
	}
	heroBlock.Props["trustedSvgIllustration"] = `<svg onload="alert(1)"><path d="M0 0"></path></svg>`

	if err := validateNuvioMarketingSeedFixture(fixture); err == nil {
		t.Fatal("expected unsafe trusted markup fixture to be rejected")
	}
}
func TestNuvioMarketingSeedRejectsUnsafeTrustedIconMarkup(t *testing.T) {
	fixture, err := loadNuvioMarketingSeedFixture("")
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	homePage := findNuvioMarketingSeedPage(fixture, "home")
	if homePage == nil {
		t.Fatal("expected home page fixture")
	}
	block := findNuvioMarketingSeedBlock(*homePage, "nuvio-reassurance-rail")
	if block == nil {
		t.Fatal("expected reassurance rail block")
	}
	items, ok := block.Props["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatal("expected reassurance rail items")
	}
	item := nuvioMarketingSeedMapValue(items[0])
	item["trustedIconSvg"] = `<svg><script>alert(1)</script></svg>`

	if err := validateNuvioMarketingSeedFixture(fixture); err == nil {
		t.Fatal("expected unsafe trusted icon markup fixture to be rejected")
	}
}

func TestNuvioMarketingSeedAllowsSafeRichTextAndRejectsUnsafeRichText(t *testing.T) {
	fixture, err := loadNuvioMarketingSeedFixture("")
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	pricingPage := findNuvioMarketingSeedPage(fixture, "pricing")
	if pricingPage == nil {
		t.Fatal("expected pricing page fixture")
	}
	faqBlock := findNuvioMarketingSeedBlock(*pricingPage, "nuvio-pricing-faq")
	if faqBlock == nil {
		t.Fatal("expected pricing FAQ block")
	}
	items, ok := faqBlock.Props["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatal("expected FAQ items")
	}
	item := nuvioMarketingSeedMapValue(items[0])
	item["answer"] = `<p><strong>Bold</strong> and <em>italic</em></p>`
	if err := validateNuvioMarketingSeedFixture(fixture); err != nil {
		t.Fatalf("expected safe rich text answer to pass: %v", err)
	}
	item["answer"] = `<p onclick="alert(1)">Bad</p>`
	if err := validateNuvioMarketingSeedFixture(fixture); err == nil {
		t.Fatal("expected unsafe rich text answer to be rejected")
	}
}

func TestNuvioMarketingSeedAppliesIdempotentlyToCleanCMSCollections(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer app.Cleanup()

	ensureNuvioMarketingSeedTestCollections(t, app)
	fixture, err := loadNuvioMarketingSeedFixture("")
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}

	stats, err := applyNuvioMarketingSeedFixture(app, fixture, true)
	if err != nil {
		t.Fatalf("failed to apply fixture: %v", err)
	}
	if stats.Created["Websites"] != 1 || stats.Created["Pages"] != 4 || stats.Created["Components"] != 17 || stats.Created["Blocks"] != 20 {
		t.Fatalf("unexpected first apply stats: %#v", stats.Created)
	}

	stats, err = applyNuvioMarketingSeedFixture(app, fixture, true)
	if err != nil {
		t.Fatalf("failed to reapply fixture: %v", err)
	}
	if stats.Created["Websites"] != 0 || stats.Created["Pages"] != 0 || stats.Created["Components"] != 0 || stats.Created["Blocks"] != 0 {
		t.Fatalf("expected second apply to create no duplicate records, got: %#v", stats.Created)
	}

	website, err := app.FindFirstRecordByFilter(nuvioWebsitesCollectionID, "slug={:slug}", dbx.Params{"slug": "nuvio"})
	if err != nil {
		t.Fatalf("expected seeded Nuvio website: %v", err)
	}
	settings := nuvioMarketingSeedMapValue(website.Get("settings"))
	seededI18n := nuvioMarketingSeedMapValue(settings["i18n"])
	if seededI18n["enabled"] != true || seededI18n["defaultLanguage"] != "pt-PT" {
		t.Fatalf("expected seeded website i18n pt-PT/en settings, got %#v", seededI18n)
	}
	seededLanguages, ok := seededI18n["languages"].([]any)
	if !ok || len(seededLanguages) != 2 {
		t.Fatalf("expected seeded language list, got %#v", seededI18n["languages"])
	}
	previewRoutes := nuvioMarketingSeedMapValue(settings["previewRoutes"])
	for pageSlug, expectedPath := range map[string]string{"home": "/", "services": "/services", "pricing": "/pricing", "contact": "/contact"} {
		if previewRoutes[pageSlug] != expectedPath {
			t.Fatalf("expected seeded preview route %s=%q, got %q", pageSlug, expectedPath, previewRoutes[pageSlug])
		}
	}

	for _, pageSlug := range []string{"home", "services", "pricing", "contact"} {
		page, err := findNuvioPublicPageBySlug(app, mustNuvioMarketingSeedWebsiteID(t, app), pageSlug)
		if err != nil {
			t.Fatalf("expected seeded page %s: %v", pageSlug, err)
		}
		blocks, err := findNuvioPublicBlocksByPageID(app, page.Id)
		if err != nil {
			t.Fatalf("expected seeded blocks for %s: %v", pageSlug, err)
		}
		if len(blocks) == 0 {
			t.Fatalf("expected public blocks for %s", pageSlug)
		}
		seoTranslations := nuvioMarketingSeedMapValue(page.Get("seo_translations"))
		if englishSeo := nuvioMarketingSeedMapValue(seoTranslations["en"]); fmt.Sprint(englishSeo["title"]) == "" {
			t.Fatalf("expected seeded English SEO translation for %s, got %#v", pageSlug, seoTranslations)
		}
		for _, block := range blocks {
			translations := nuvioMarketingSeedMapValue(block.Get("translations"))
			if len(nuvioMarketingSeedMapValue(translations["en"])) == 0 {
				t.Fatalf("expected seeded English translation for %s block %s", pageSlug, block.GetString("slot"))
			}
		}
	}
}

func TestNuvioMarketingSeedPreservesExistingEnglishContentAsTranslations(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer app.Cleanup()

	ensureNuvioMarketingSeedTestCollections(t, app)
	fixture, err := loadNuvioMarketingSeedFixture("")
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}
	if _, err := applyNuvioMarketingSeedFixture(app, fixture, true); err != nil {
		t.Fatalf("failed to apply fixture: %v", err)
	}

	websiteID := mustNuvioMarketingSeedWebsiteID(t, app)
	homePage, err := findNuvioPublicPageBySlug(app, websiteID, "home")
	if err != nil {
		t.Fatalf("expected home page: %v", err)
	}
	homePage.Set("seo_title", "Edited English SEO title")
	homePage.Set("seo_description", "Edited English SEO description")
	homePage.Set("seo_translations", map[string]any{})
	if err := app.Save(homePage); err != nil {
		t.Fatalf("failed to save edited SEO: %v", err)
	}

	homeFixturePage := findNuvioMarketingSeedPage(fixture, "home")
	if homeFixturePage == nil {
		t.Fatal("expected home fixture page")
	}
	homeFixtureHero := findNuvioMarketingSeedBlock(*homeFixturePage, "nuvio-home-hero")
	if homeFixtureHero == nil {
		t.Fatal("expected home hero fixture block")
	}

	blocksCollection, err := app.FindCachedCollectionByNameOrId(nuvioBlocksCollectionID)
	if err != nil {
		t.Fatalf("failed to resolve blocks collection: %v", err)
	}
	heroRecord, err := app.FindFirstRecordByFilter(blocksCollection, "page={:page} && slot={:slot}", dbx.Params{"page": homePage.Id, "slot": homeFixtureHero.Slot})
	if err != nil {
		t.Fatalf("expected home hero block: %v", err)
	}
	editedEnglishProps := cloneNuvioMarketingSeedMap(nuvioMarketingSeedMapValue(homeFixtureHero.Translations["en"]))
	editedEnglishProps["headingPrefix"] = "Edited English CMS headline"
	heroRecord.Set("props", editedEnglishProps)
	heroRecord.Set("translations", map[string]any{})
	if err := app.Save(heroRecord); err != nil {
		t.Fatalf("failed to save edited hero props: %v", err)
	}

	if _, err := applyNuvioMarketingSeedFixture(app, fixture, true); err != nil {
		t.Fatalf("failed to reapply fixture: %v", err)
	}

	reseededPage, err := findNuvioPublicPageBySlug(app, websiteID, "home")
	if err != nil {
		t.Fatalf("expected reseeded home page: %v", err)
	}
	if got := reseededPage.GetString("seo_title"); got != homeFixturePage.SeoTitle {
		t.Fatalf("expected PT SEO title restored to %q, got %q", homeFixturePage.SeoTitle, got)
	}
	englishSeo := nuvioMarketingSeedMapValue(nuvioMarketingSeedMapValue(reseededPage.Get("seo_translations"))["en"])
	if englishSeo["title"] != "Edited English SEO title" || englishSeo["description"] != "Edited English SEO description" {
		t.Fatalf("expected existing English SEO preserved, got %#v", englishSeo)
	}

	reseededHero, err := app.FindFirstRecordByFilter(blocksCollection, "page={:page} && slot={:slot}", dbx.Params{"page": reseededPage.Id, "slot": homeFixtureHero.Slot})
	if err != nil {
		t.Fatalf("expected reseeded home hero block: %v", err)
	}
	props := nuvioMarketingSeedMapValue(reseededHero.Get("props"))
	if props["headingPrefix"] != homeFixtureHero.Props["headingPrefix"] {
		t.Fatalf("expected PT props restored, got %#v", props["headingPrefix"])
	}
	englishTranslation := nuvioMarketingSeedMapValue(nuvioMarketingSeedMapValue(reseededHero.Get("translations"))["en"])
	if englishTranslation["headingPrefix"] != "Edited English CMS headline" {
		t.Fatalf("expected edited English props preserved in translations.en, got %#v", englishTranslation["headingPrefix"])
	}
}

func TestNuvioMarketingSeedPublicContentEndpoints(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "seeded home page renders through public content endpoint",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=nuvio&pageSlug=home",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioMarketingSeedPublicContentScenario(t, app, e)
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"slug":"nuvio"`, `"slug":"home"`, `nuvio-customer-website-concept`, `"translations":{"en"`},
		},
		{
			Name:   "seeded services page renders through public content endpoint",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=nuvio&pageSlug=services",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioMarketingSeedPublicContentScenario(t, app, e)
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"slug":"services"`, `nuvio-services-mapping`, `"translations":{"en"`},
		},
		{
			Name:   "seeded pricing page renders comparison and foundation source block",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=nuvio&pageSlug=pricing",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioMarketingSeedPublicContentScenario(t, app, e)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"slug":"pricing"`,
				`nuvio-pricing-plans`,
				`"comparison"`,
				`"foundation"`,
				"\u20ac590",
			},
		},
		{
			Name:   "seeded contact page renders presentation props without mechanics",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=nuvio&pageSlug=contact",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioMarketingSeedPublicContentScenario(t, app, e)
			},
			ExpectedStatus:     200,
			ExpectedContent:    []string{`"slug":"contact"`, `nuvio-contact-request`, `"context"`, `"form"`},
			NotExpectedContent: []string{`"endpoint"`, `"validation"`},
		},
		{
			Name:   "unknown seeded page remains not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=nuvio&pageSlug=unknown-page",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioMarketingSeedPublicContentScenario(t, app, e)
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"message":"Page not found."`},
		},
		{
			Name:   "unknown seeded website remains not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=unknown-site&pageSlug=home",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioMarketingSeedPublicContentScenario(t, app, e)
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"message":"Website not found."`},
		},
		{
			Name:   "seeded sitemap data includes Nuvio pages",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/sitemap-data",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioMarketingSeedPublicContentScenario(t, app, e)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"websiteSlug":"nuvio"`,
				`"slug":"home"`,
				`"slug":"services"`,
				`"slug":"pricing"`,
				`"slug":"contact"`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioMarketingSeedPublicContentScenario(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	setupNuvioPublicContentRoutesForTest(t, app, e)
	ensureNuvioMarketingSeedTestCollections(t, app)
	fixture, err := loadNuvioMarketingSeedFixture("")
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}
	if _, err := applyNuvioMarketingSeedFixture(app, fixture, true); err != nil {
		t.Fatalf("failed to apply fixture: %v", err)
	}
}

func ensureNuvioMarketingSeedTestCollections(t testing.TB, app *tests.TestApp) {
	t.Helper()
	ensureNuvioCMSBackofficeCollection(t, app, "Websites", nuvioWebsitesCollectionID, []core.Field{
		&core.TextField{Name: "name"},
		&core.TextField{Name: "title"},
		&core.TextField{Name: "slug"},
		&core.TextField{Name: "domain"},
		&core.BoolField{Name: "enabled"},
		&core.BoolField{Name: "active"},
		&core.BoolField{Name: "published"},
		&core.BoolField{Name: "visible"},
		&core.BoolField{Name: "private"},
		&core.TextField{Name: "status"},
		&core.JSONField{Name: "settings"},
	})
	ensureNuvioCMSBackofficeCollection(t, app, "Pages", nuvioPagesCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "slug"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "title"},
		&core.TextField{Name: "seo_title"},
		&core.TextField{Name: "seo_description"},
		&core.BoolField{Name: "published"},
		&core.BoolField{Name: "enabled"},
		&core.BoolField{Name: "active"},
		&core.BoolField{Name: "visible"},
		&core.BoolField{Name: "private"},
		&core.TextField{Name: "status"},
		&core.BoolField{Name: "seo_noindex"},
		&core.BoolField{Name: "seo_exclude_from_sitemap"},
		&core.JSONField{Name: "seo_translations"},
	})
	ensureNuvioCMSBackofficeCollection(t, app, "Blocks", nuvioBlocksCollectionID, []core.Field{
		&core.TextField{Name: "page"},
		&core.TextField{Name: "component"},
		&core.TextField{Name: "component_key"},
		&core.TextField{Name: "slot"},
		&core.TextField{Name: "title"},
		&core.NumberField{Name: "displayOrder"},
		&core.JSONField{Name: "props"},
		&core.JSONField{Name: "translations"},
		&core.BoolField{Name: "enabled"},
		&core.BoolField{Name: "visible"},
		&core.BoolField{Name: "private"},
		&core.TextField{Name: "status"},
	})
	ensureNuvioCMSBackofficeCollection(t, app, "Components", nuvioComponentsCollectionID, []core.Field{
		&core.TextField{Name: "key"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "gallery"},
		&core.JSONField{Name: "schema"},
	})
}

func mustNuvioMarketingSeedWebsiteID(t testing.TB, app *tests.TestApp) string {
	t.Helper()
	website, err := findNuvioPublicWebsiteBySlugOrDomain(app, "nuvio")
	if err != nil {
		t.Fatalf("failed to load seeded website: %v", err)
	}
	return website.Id
}
