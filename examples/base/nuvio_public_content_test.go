package main

import (
	"net/http"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestNuvioPublicContentEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid public content request returns sanitized website page and blocks",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=alpha-cms&pageSlug=home",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"website":{`, `"page":{`, `"blocks":[`, `"slug":"alpha-cms"`},
		},
		{
			Name:   "invalid website slug returns safe bad request",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=bad/slug&pageSlug=home",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Invalid website slug."`},
		},
		{
			Name:   "invalid page slug returns safe bad request",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=alpha-cms&pageSlug=bad/slug",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Invalid page slug."`},
		},
		{
			Name:   "unknown content query key is rejected",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=alpha-cms&pageSlug=home&extra=1",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Query parameter \"extra\" is not allowed."`},
		},
		{
			Name:   "missing website returns safe not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=missing-site&pageSlug=home",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"message":"Website not found."`},
		},
		{
			Name:   "missing page returns safe not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=alpha-cms&pageSlug=missing-page",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"message":"Page not found."`},
		},
		{
			Name:   "unknown sitemap query key is rejected",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/sitemap-data?foo=bar",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Query parameter \"foo\" is not allowed."`},
		},
		{
			Name:   "sitemap data returns public-safe website and page fields",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/sitemap-data",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"websites":[`, `"pages":[`, `"websiteSlug":"alpha-cms"`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioPublicContentDTOExcludesSensitiveFields(t *testing.T) {
	t.Parallel()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer app.Cleanup()

	seedNuvioCMSBackofficeDashboardData(t, app)

	websiteRecord, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
	if err != nil {
		t.Fatalf("failed to load website record: %v", err)
	}

	websiteDTO := buildNuvioPublicWebsiteDTO(websiteRecord)
	settings := mustNuvioPublicMapValue(t, websiteDTO["settings"], "website.settings")

	if _, hasProviderInternals := settings["providerInternals"]; hasProviderInternals {
		t.Fatalf("public website settings unexpectedly expose providerInternals")
	}

	if contactForm, ok := toStringAnyMap(settings["contactForm"]); ok {
		if _, hasEmailNotifications := contactForm["emailNotifications"]; hasEmailNotifications {
			t.Fatalf("public contactForm settings unexpectedly expose emailNotifications")
		}
	}

	if reports, ok := toStringAnyMap(settings["reports"]); ok {
		if analytics, ok := toStringAnyMap(reports["analytics"]); ok {
			if _, hasAPIURL := analytics["apiUrl"]; hasAPIURL {
				t.Fatalf("public reports analytics unexpectedly expose apiUrl")
			}
			if _, hasAPIURLSnake := analytics["api_url"]; hasAPIURLSnake {
				t.Fatalf("public reports analytics unexpectedly expose api_url")
			}
		}
	}

	blockRecords, err := findNuvioPublicBlocksByPageID(app, nuvioCMSDashboardAlphaPageID)
	if err != nil {
		t.Fatalf("failed to load public blocks: %v", err)
	}

	blockDTOs := buildNuvioPublicBlocksDTO(blockRecords)
	if len(blockDTOs) == 0 {
		t.Fatalf("expected public block dto entries")
	}

	for _, blockDTO := range blockDTOs {
		if _, hasSchema := blockDTO["schema"]; hasSchema {
			t.Fatalf("public block dto unexpectedly exposes schema")
		}
		if _, hasProviderConfig := blockDTO["providerConfig"]; hasProviderConfig {
			t.Fatalf("public block dto unexpectedly exposes providerConfig")
		}
		if _, hasInternalNotes := blockDTO["internalNotes"]; hasInternalNotes {
			t.Fatalf("public block dto unexpectedly exposes internalNotes")
		}

		expandValue, hasExpand := toStringAnyMap(blockDTO["expand"])
		if !hasExpand {
			continue
		}

		componentValue, hasComponent := toStringAnyMap(expandValue["component"])
		if !hasComponent {
			continue
		}

		if _, hasSchema := componentValue["schema"]; hasSchema {
			t.Fatalf("public block component unexpectedly exposes schema")
		}
		if _, hasProviderConfig := componentValue["providerConfig"]; hasProviderConfig {
			t.Fatalf("public block component unexpectedly exposes providerConfig")
		}
		if _, hasInternalNotes := componentValue["internalNotes"]; hasInternalNotes {
			t.Fatalf("public block component unexpectedly exposes internalNotes")
		}
	}
}

func TestNuvioPublicSitemapDTOExposesMinimalShape(t *testing.T) {
	t.Parallel()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer app.Cleanup()

	seedNuvioCMSBackofficeDashboardData(t, app)

	websiteRecord, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
	if err != nil {
		t.Fatalf("failed to load website record: %v", err)
	}

	websiteDTO := buildNuvioPublicWebsiteSitemapDTO(websiteRecord)
	for _, deniedKey := range []string{"id", "status", "enabled", "active", "published"} {
		if _, exists := websiteDTO[deniedKey]; exists {
			t.Fatalf("public sitemap website dto unexpectedly exposes %s", deniedKey)
		}
	}

	pageRecord, err := app.FindRecordById(nuvioPagesCollectionID, nuvioCMSDashboardAlphaPageID)
	if err != nil {
		t.Fatalf("failed to load page record: %v", err)
	}

	pageDTO := buildNuvioPublicSitemapPageDTO(pageRecord, "alpha-cms")
	if pageDTO["websiteSlug"] != "alpha-cms" {
		t.Fatalf("expected websiteSlug=alpha-cms, got %#v", pageDTO["websiteSlug"])
	}

	for _, deniedKey := range []string{
		"id",
		"website",
		"websiteId",
		"status",
		"enabled",
		"active",
		"published",
		"created",
		"seo_noindex",
		"seo_exclude_from_sitemap",
	} {
		if _, exists := pageDTO[deniedKey]; exists {
			t.Fatalf("public sitemap page dto unexpectedly exposes %s", deniedKey)
		}
	}
}

func setupNuvioPublicContentRoutesForTest(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioPublicContentRoutes(e)
}

func mustNuvioPublicMapValue(t testing.TB, value any, label string) map[string]any {
	t.Helper()

	parsed, ok := toStringAnyMap(value)
	if !ok {
		t.Fatalf("expected %s to be an object map, got %T", label, value)
	}

	return parsed
}
