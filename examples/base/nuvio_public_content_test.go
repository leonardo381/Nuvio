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

func TestNuvioPublicContentRenderabilityFiltering(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "renderable page returns public content and skips non-renderable blocks",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=filter-cms&pageSlug=home",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioPublicContentRenderabilityData(t, app)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"slug":"filter-cms"`,
				`"slug":"home"`,
				`Renderable Hero`,
			},
			NotExpectedContent: []string{
				`Disabled Block Should Not Render`,
				`Private Block Should Not Render`,
				`Hidden Block Should Not Render`,
				`Draft Block Should Not Render`,
			},
		},
		{
			Name:   "unpublished page returns safe not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=filter-cms&pageSlug=unpublished-page",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioPublicContentRenderabilityData(t, app)
			},
			ExpectedStatus:     404,
			ExpectedContent:    []string{`"message":"Page not found."`},
			NotExpectedContent: []string{`Unpublished Page Should Not Render`},
		},
		{
			Name:   "draft page returns safe not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=filter-cms&pageSlug=draft-page",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioPublicContentRenderabilityData(t, app)
			},
			ExpectedStatus:     404,
			ExpectedContent:    []string{`"message":"Page not found."`},
			NotExpectedContent: []string{`Draft Page Should Not Render`},
		},
		{
			Name:   "private page returns safe not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=filter-cms&pageSlug=private-page",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioPublicContentRenderabilityData(t, app)
			},
			ExpectedStatus:     404,
			ExpectedContent:    []string{`"message":"Page not found."`},
			NotExpectedContent: []string{`Private Page Should Not Render`},
		},
		{
			Name:   "disabled website returns safe not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=disabled-cms&pageSlug=home",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioPublicContentRenderabilityData(t, app)
			},
			ExpectedStatus:     404,
			ExpectedContent:    []string{`"message":"Website not found."`},
			NotExpectedContent: []string{`Disabled CMS`},
		},
		{
			Name:   "private website returns safe not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/content?websiteSlug=private-cms&pageSlug=home",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioPublicContentRenderabilityData(t, app)
			},
			ExpectedStatus:     404,
			ExpectedContent:    []string{`"message":"Website not found."`},
			NotExpectedContent: []string{`Private CMS`},
		},
		{
			Name:   "sitemap data excludes non-renderable websites and pages",
			Method: http.MethodGet,
			URL:    "/api/nuvio/public/sitemap-data",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicContentRoutesForTest(t, app, e)
				seedNuvioPublicContentRenderabilityData(t, app)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"websiteSlug":"filter-cms"`,
				`"slug":"home"`,
			},
			NotExpectedContent: []string{
				`disabled-cms`,
				`private-cms`,
				`unpublished-page`,
				`draft-page`,
				`private-page`,
			},
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

func seedNuvioPublicContentRenderabilityData(t testing.TB, app *tests.TestApp) {
	t.Helper()

	websitesCollection := ensureNuvioCMSBackofficeCollection(t, app, "Websites", nuvioWebsitesCollectionID, []core.Field{
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
	pagesCollection := ensureNuvioCMSBackofficeCollection(t, app, "Pages", nuvioPagesCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "title"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "slug"},
		&core.TextField{Name: "status"},
		&core.BoolField{Name: "enabled"},
		&core.BoolField{Name: "active"},
		&core.BoolField{Name: "published"},
		&core.BoolField{Name: "visible"},
		&core.BoolField{Name: "private"},
		&core.BoolField{Name: "seo_noindex"},
		&core.BoolField{Name: "seo_exclude_from_sitemap"},
		&core.JSONField{Name: "seo_translations"},
	})
	blocksCollection := ensureNuvioCMSBackofficeCollection(t, app, "Blocks", nuvioBlocksCollectionID, []core.Field{
		&core.TextField{Name: "page"},
		&core.TextField{Name: "component"},
		&core.TextField{Name: "component_key"},
		&core.TextField{Name: "variant"},
		&core.TextField{Name: "slot"},
		&core.NumberField{Name: "displayOrder"},
		&core.JSONField{Name: "props"},
		&core.JSONField{Name: "translations"},
		&core.BoolField{Name: "enabled"},
		&core.BoolField{Name: "visible"},
		&core.BoolField{Name: "private"},
		&core.TextField{Name: "status"},
	})
	componentsCollection := ensureNuvioCMSBackofficeCollection(t, app, "Components", nuvioComponentsCollectionID, []core.Field{
		&core.TextField{Name: "key"},
		&core.TextField{Name: "component_key"},
		&core.TextField{Name: "name"},
	})

	upsertNuvioCMSBackofficeRecord(t, app, websitesCollection, "renderweb000001", map[string]any{
		"name":      "Filter CMS",
		"title":     "Filter CMS",
		"slug":      "filter-cms",
		"domain":    "filter-cms.example.test",
		"enabled":   true,
		"active":    true,
		"published": true,
		"visible":   true,
		"private":   false,
		"status":    "active",
		"settings":  map[string]any{},
	})
	upsertNuvioCMSBackofficeRecord(t, app, websitesCollection, "renderweb000002", map[string]any{
		"name":      "Disabled CMS",
		"title":     "Disabled CMS",
		"slug":      "disabled-cms",
		"domain":    "disabled-cms.example.test",
		"enabled":   false,
		"active":    true,
		"published": true,
		"visible":   true,
		"private":   false,
		"status":    "active",
		"settings":  map[string]any{},
	})
	upsertNuvioCMSBackofficeRecord(t, app, websitesCollection, "renderweb000003", map[string]any{
		"name":      "Private CMS",
		"title":     "Private CMS",
		"slug":      "private-cms",
		"domain":    "private-cms.example.test",
		"enabled":   true,
		"active":    true,
		"published": true,
		"visible":   true,
		"private":   true,
		"status":    "active",
		"settings":  map[string]any{},
	})

	pageDefaults := map[string]any{
		"website":                  "renderweb000001",
		"enabled":                  true,
		"active":                   true,
		"published":                true,
		"visible":                  true,
		"private":                  false,
		"status":                   "published",
		"seo_noindex":              false,
		"seo_exclude_from_sitemap": false,
		"seo_translations":         map[string]any{},
	}
	seedNuvioPublicContentRenderabilityPage(t, app, pagesCollection, "renderpage00001", "home", "Renderable Home", pageDefaults, nil)
	seedNuvioPublicContentRenderabilityPage(t, app, pagesCollection, "renderpage00002", "unpublished-page", "Unpublished Page Should Not Render", pageDefaults, map[string]any{"published": false, "status": "unpublished"})
	seedNuvioPublicContentRenderabilityPage(t, app, pagesCollection, "renderpage00003", "draft-page", "Draft Page Should Not Render", pageDefaults, map[string]any{"status": "draft"})
	seedNuvioPublicContentRenderabilityPage(t, app, pagesCollection, "renderpage00004", "private-page", "Private Page Should Not Render", pageDefaults, map[string]any{"private": true})
	seedNuvioPublicContentRenderabilityPage(t, app, pagesCollection, "renderpage00005", "disabled-page", "Disabled Page Should Not Render", pageDefaults, map[string]any{"enabled": false})

	upsertNuvioCMSBackofficeRecord(t, app, componentsCollection, "rendercomp00001", map[string]any{
		"key":           "hero",
		"component_key": "hero",
		"name":          "Hero Section",
	})

	blockDefaults := map[string]any{
		"page":          "renderpage00001",
		"component":     "rendercomp00001",
		"component_key": "hero",
		"variant":       "default",
		"slot":          "main",
		"enabled":       true,
		"visible":       true,
		"private":       false,
		"status":        "active",
		"translations":  map[string]any{},
	}
	seedNuvioPublicContentRenderabilityBlock(t, app, blocksCollection, "renderblock0001", blockDefaults, map[string]any{"displayOrder": 1, "props": map[string]any{"title": "Renderable Hero"}})
	seedNuvioPublicContentRenderabilityBlock(t, app, blocksCollection, "renderblock0002", blockDefaults, map[string]any{"displayOrder": 2, "enabled": false, "props": map[string]any{"title": "Disabled Block Should Not Render"}})
	seedNuvioPublicContentRenderabilityBlock(t, app, blocksCollection, "renderblock0003", blockDefaults, map[string]any{"displayOrder": 3, "private": true, "props": map[string]any{"title": "Private Block Should Not Render"}})
	seedNuvioPublicContentRenderabilityBlock(t, app, blocksCollection, "renderblock0004", blockDefaults, map[string]any{"displayOrder": 4, "visible": false, "props": map[string]any{"title": "Hidden Block Should Not Render"}})
	seedNuvioPublicContentRenderabilityBlock(t, app, blocksCollection, "renderblock0005", blockDefaults, map[string]any{"displayOrder": 5, "status": "draft", "props": map[string]any{"title": "Draft Block Should Not Render"}})
}

func seedNuvioPublicContentRenderabilityPage(t testing.TB, app *tests.TestApp, collection *core.Collection, id string, slug string, title string, defaults map[string]any, overrides map[string]any) {
	t.Helper()

	values := copyNuvioPublicContentTestValues(defaults)
	values["slug"] = slug
	values["title"] = title
	values["name"] = title
	for key, value := range overrides {
		values[key] = value
	}
	upsertNuvioCMSBackofficeRecord(t, app, collection, id, values)
}

func seedNuvioPublicContentRenderabilityBlock(t testing.TB, app *tests.TestApp, collection *core.Collection, id string, defaults map[string]any, overrides map[string]any) {
	t.Helper()

	values := copyNuvioPublicContentTestValues(defaults)
	for key, value := range overrides {
		values[key] = value
	}
	upsertNuvioCMSBackofficeRecord(t, app, collection, id, values)
}

func copyNuvioPublicContentTestValues(values map[string]any) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
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
