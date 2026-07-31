package main

import (
	"net/http"
	"testing"

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

	expectedBlocks := map[string]int{"home": 9, "services": 6, "pricing": 4, "contact": 1}
	for _, page := range fixture.Pages {
		if expectedBlocks[page.Slug] != len(page.Blocks) {
			t.Fatalf("expected %d blocks for %s, got %d", expectedBlocks[page.Slug], page.Slug, len(page.Blocks))
		}
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
			ExpectedContent: []string{`"slug":"nuvio"`, `"slug":"home"`, `nuvio-customer-website-concept`},
		},
		{
			Name:   "seeded services page renders through public content endpoint",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=nuvio&pageSlug=services",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioMarketingSeedPublicContentScenario(t, app, e)
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"slug":"services"`, `nuvio-services-mapping`},
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
