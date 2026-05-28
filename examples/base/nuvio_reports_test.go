package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestResolveTrustedNuvioUmamiAPIURL(t *testing.T) {
	t.Run("missing env returns provider_unconfigured", func(t *testing.T) {
		t.Setenv("NUVIO_UMAMI_API_URL", "")

		resolvedURL, err := resolveTrustedNuvioUmamiAPIURL()
		if err == nil {
			t.Fatal("expected error for missing NUVIO_UMAMI_API_URL, got nil")
		}
		if resolvedURL != "" {
			t.Fatalf("expected empty url on error, got %q", resolvedURL)
		}

		stateErr, ok := unwrapNuvioReportsTrafficStateError(err)
		if !ok {
			t.Fatalf("expected nuvioReportsTrafficStateError, got %T", err)
		}
		if stateErr.State != "provider_unconfigured" {
			t.Fatalf("expected state provider_unconfigured, got %q", stateErr.State)
		}
	})

	t.Run("invalid env returns safe configuration error", func(t *testing.T) {
		t.Setenv("NUVIO_UMAMI_API_URL", "javascript:alert(1)")

		resolvedURL, err := resolveTrustedNuvioUmamiAPIURL()
		if err == nil {
			t.Fatal("expected error for invalid NUVIO_UMAMI_API_URL, got nil")
		}
		if resolvedURL != "" {
			t.Fatalf("expected empty url on error, got %q", resolvedURL)
		}

		stateErr, ok := unwrapNuvioReportsTrafficStateError(err)
		if !ok {
			t.Fatalf("expected nuvioReportsTrafficStateError, got %T", err)
		}
		if stateErr.State != "provider_unconfigured" {
			t.Fatalf("expected state provider_unconfigured, got %q", stateErr.State)
		}
		if !strings.Contains(strings.ToLower(stateErr.Message), "misconfigured") {
			t.Fatalf("expected safe misconfigured message, got %q", stateErr.Message)
		}
	})

	t.Run("valid env is normalized", func(t *testing.T) {
		t.Setenv("NUVIO_UMAMI_API_URL", "https://umami.example.com/api/")

		resolvedURL, err := resolveTrustedNuvioUmamiAPIURL()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resolvedURL != "https://umami.example.com/api" {
			t.Fatalf("expected normalized url, got %q", resolvedURL)
		}
	})
}

func TestLoadNuvioUmamiConfig(t *testing.T) {
	t.Run("uses trusted env url and api key auth", func(t *testing.T) {
		t.Setenv("NUVIO_UMAMI_API_URL", "https://umami.example.com/api")
		t.Setenv("NUVIO_UMAMI_API_KEY", "test-key")
		t.Setenv("NUVIO_UMAMI_USERNAME", "")
		t.Setenv("NUVIO_UMAMI_PASSWORD", "")

		config, err := loadNuvioUmamiConfig()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if config.APIBaseURL != "https://umami.example.com/api" {
			t.Fatalf("expected APIBaseURL to use trusted env URL, got %q", config.APIBaseURL)
		}
		if config.RequestBaseURL != "https://umami.example.com/api" {
			t.Fatalf("expected RequestBaseURL to use trusted env URL, got %q", config.RequestBaseURL)
		}
		if config.APIKey != "test-key" {
			t.Fatalf("expected API key to be loaded from env, got %q", config.APIKey)
		}
	})

	t.Run("missing env returns provider_unconfigured", func(t *testing.T) {
		t.Setenv("NUVIO_UMAMI_API_URL", "")
		t.Setenv("NUVIO_UMAMI_API_KEY", "test-key")
		t.Setenv("NUVIO_UMAMI_USERNAME", "")
		t.Setenv("NUVIO_UMAMI_PASSWORD", "")

		_, err := loadNuvioUmamiConfig()
		if err == nil {
			t.Fatal("expected error for missing NUVIO_UMAMI_API_URL, got nil")
		}

		stateErr, ok := unwrapNuvioReportsTrafficStateError(err)
		if !ok {
			t.Fatalf("expected nuvioReportsTrafficStateError, got %T", err)
		}
		if stateErr.State != "provider_unconfigured" {
			t.Fatalf("expected state provider_unconfigured, got %q", stateErr.State)
		}
	})
}

const (
	nuvioReportsDashboardAlphaWebsiteID = "alphawebsite001"
	nuvioReportsDashboardBetaWebsiteID  = "betawebsite0002"
	nuvioReportsDashboardGammaWebsiteID = "gammawebsite003"
)

func TestNuvioReportsDashboardEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin superuser receives scoped dashboard datasets",
			Method: http.MethodGet,
			URL:    "/api/nuvio/reports/dashboard?websiteId=" + nuvioReportsDashboardAlphaWebsiteID + "&period=thisMonth",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioReportsDashboardRoute(t, app, e)
				seedNuvioReportsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioReportsDashboardAlphaWebsiteID, nuvioReportsDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioReportsDashboardAlphaWebsiteID + `"`,
				`"reports":true`,
				`"contacts":[`,
				`"whatsapp":[`,
				`"appointments":[`,
				`"bookingServices":[`,
				`"subscribers":[`,
				`"campaigns":[`,
				`"pages":[`,
				`"name":"Alpha Contact"`,
			},
			NotExpectedContent: []string{
				`"name":"Beta Contact"`,
				`"settings":`,
				`"apiUrl"`,
			},
		},
		{
			Name:   "admin receives empty dataset keys when website has no records",
			Method: http.MethodGet,
			URL:    "/api/nuvio/reports/dashboard?websiteId=" + nuvioReportsDashboardGammaWebsiteID + "&period=allTime",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioReportsDashboardRoute(t, app, e)
				seedNuvioReportsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioReportsDashboardAlphaWebsiteID, nuvioReportsDashboardBetaWebsiteID, nuvioReportsDashboardGammaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"websiteId":"` + nuvioReportsDashboardGammaWebsiteID + `"`,
				`"contacts":[]`,
				`"whatsapp":[]`,
				`"appointments":[]`,
				`"bookingServices":[]`,
				`"subscribers":[]`,
				`"campaigns":[]`,
				`"pages":[]`,
			},
		},
		{
			Name:   "client superuser receives only assigned website dashboard",
			Method: http.MethodGet,
			URL:    "/api/nuvio/reports/dashboard?websiteId=" + nuvioReportsDashboardAlphaWebsiteID + "&period=thisMonth",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioReportsDashboardRoute(t, app, e)
				seedNuvioReportsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioReportsDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioReportsDashboardAlphaWebsiteID + `"`,
				`"name":"Alpha Contact"`,
			},
			NotExpectedContent: []string{
				`"name":"Beta Contact"`,
				`"apiUrl"`,
			},
		},
		{
			Name:   "client superuser denied for unassigned website",
			Method: http.MethodGet,
			URL:    "/api/nuvio/reports/dashboard?websiteId=" + nuvioReportsDashboardAlphaWebsiteID + "&period=thisMonth",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioReportsDashboardRoute(t, app, e)
				seedNuvioReportsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioReportsDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client superuser with no website access denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/reports/dashboard?websiteId=" + nuvioReportsDashboardAlphaWebsiteID + "&period=thisMonth",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioReportsDashboardRoute(t, app, e)
				seedNuvioReportsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "missing role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/reports/dashboard?websiteId=" + nuvioReportsDashboardAlphaWebsiteID + "&period=thisMonth",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioReportsDashboardRoute(t, app, e)
				seedNuvioReportsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioReportsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/reports/dashboard?websiteId=" + nuvioReportsDashboardAlphaWebsiteID + "&period=thisMonth",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioReportsDashboardRoute(t, app, e)
				seedNuvioReportsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioReportsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/reports/dashboard?websiteId=" + nuvioReportsDashboardAlphaWebsiteID + "&period=thisMonth",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioReportsDashboardRoute(t, app, e)
				seedNuvioReportsDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioReportsDashboardRoute(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioReportsRoutes(e)
}

func seedNuvioReportsDashboardData(t testing.TB, app *tests.TestApp) {
	t.Helper()

	websitesCollection := ensureNuvioBackofficeWebsitesCollection(t, app)
	contactsCollection := ensureNuvioReportsDashboardContactsCollection(t, app)
	whatsappCollection := ensureNuvioReportsDashboardWhatsappCollection(t, app)
	appointmentsCollection := ensureNuvioReportsDashboardAppointmentsCollection(t, app)
	bookingServicesCollection := ensureNuvioReportsDashboardBookingServicesCollection(t, app)
	subscribersCollection := ensureNuvioReportsDashboardSubscribersCollection(t, app)
	campaignsCollection := ensureNuvioReportsDashboardCampaignsCollection(t, app)
	pagesCollection := ensureNuvioReportsDashboardPagesCollection(t, app)

	upsertNuvioReportsDashboardWebsite(t, app, websitesCollection, nuvioReportsDashboardAlphaWebsiteID, "alpha-site", "Alpha Website")
	upsertNuvioReportsDashboardWebsite(t, app, websitesCollection, nuvioReportsDashboardBetaWebsiteID, "beta-site", "Beta Website")
	upsertNuvioReportsDashboardWebsite(t, app, websitesCollection, nuvioReportsDashboardGammaWebsiteID, "gamma-site", "Gamma Website")

	upsertNuvioReportsDashboardRecord(t, app, contactsCollection, "alphacontact001", map[string]any{
		"website": nuvioReportsDashboardAlphaWebsiteID,
		"channel": "contact",
		"status":  "new",
		"name":    "Alpha Contact",
		"email":   "alpha@example.test",
		"phone":   "+111111111",
		"subject": "Alpha lead subject",
		"message": "Alpha lead message",
	})
	upsertNuvioReportsDashboardRecord(t, app, contactsCollection, "betacontact0001", map[string]any{
		"website": nuvioReportsDashboardBetaWebsiteID,
		"channel": "contact",
		"status":  "new",
		"name":    "Beta Contact",
		"email":   "beta@example.test",
		"phone":   "+222222222",
		"subject": "Beta lead subject",
		"message": "Beta lead message",
	})

	upsertNuvioReportsDashboardRecord(t, app, whatsappCollection, "alphawhatsapp01", map[string]any{
		"website":        nuvioReportsDashboardAlphaWebsiteID,
		"status":         "new",
		"source":         "floating_button",
		"name":           "Alpha WhatsApp",
		"email":          "alpha-wa@example.test",
		"phone":          "+111111111",
		"message":        "Alpha WhatsApp message",
		"defaultMessage": "Default Alpha message",
	})

	upsertNuvioReportsDashboardRecord(t, app, appointmentsCollection, "alphaappoint001", map[string]any{
		"website": nuvioReportsDashboardAlphaWebsiteID,
		"service": "alphaservice001",
		"status":  "pending",
		"name":    "Alpha Appointment",
		"email":   "alpha-booking@example.test",
		"phone":   "+111111111",
		"date":    "2026-05-20",
		"time":    "10:00",
	})

	upsertNuvioReportsDashboardRecord(t, app, bookingServicesCollection, "alphaservice001", map[string]any{
		"website": nuvioReportsDashboardAlphaWebsiteID,
		"name":    "Alpha Service",
	})

	upsertNuvioReportsDashboardRecord(t, app, subscribersCollection, "alphasubscr0001", map[string]any{
		"website": nuvioReportsDashboardAlphaWebsiteID,
		"email":   "subscriber-alpha@example.test",
		"status":  "active",
	})

	upsertNuvioReportsDashboardRecord(t, app, campaignsCollection, "alphacampaig001", map[string]any{
		"website":         nuvioReportsDashboardAlphaWebsiteID,
		"subject":         "Alpha Campaign",
		"status":          "sent",
		"recipientsCount": 10,
		"sentAt":          "2026-05-15T10:00:00Z",
	})

	upsertNuvioReportsDashboardRecord(t, app, pagesCollection, "alphapage000001", map[string]any{
		"website":          nuvioReportsDashboardAlphaWebsiteID,
		"title":            "Alpha Home",
		"name":             "Alpha Home",
		"slug":             "home",
		"path":             "/",
		"url":              "/",
		"seo_title":        "Alpha SEO Title",
		"seo_description":  "Alpha SEO Description",
		"seo_social_image": "alpha-social.jpg",
		"seo_noindex":      false,
	})
}

func ensureNuvioReportsDashboardContactsCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioReportsDashboardCollection(t, app, "Contacts", nuvioContactsCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "channel"},
		&core.TextField{Name: "status"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "email"},
		&core.TextField{Name: "phone"},
		&core.TextField{Name: "subject"},
		&core.TextField{Name: "message"},
	})
}

func ensureNuvioReportsDashboardWhatsappCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioReportsDashboardCollection(t, app, "WhatsAppInteractions", nuvioWhatsappCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "status"},
		&core.TextField{Name: "source"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "email"},
		&core.TextField{Name: "phone"},
		&core.TextField{Name: "message"},
		&core.TextField{Name: "defaultMessage"},
	})
}

func ensureNuvioReportsDashboardAppointmentsCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioReportsDashboardCollection(t, app, "Appointments", nuvioAppointmentsCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "service"},
		&core.TextField{Name: "status"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "email"},
		&core.TextField{Name: "phone"},
		&core.TextField{Name: "date"},
		&core.TextField{Name: "time"},
	})
}

func ensureNuvioReportsDashboardBookingServicesCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioReportsDashboardCollection(t, app, "BookingServices", nuvioBookingServicesCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "name"},
	})
}

func ensureNuvioReportsDashboardSubscribersCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioReportsDashboardCollection(t, app, "Subscribers", nuvioSubscribersCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "email"},
		&core.TextField{Name: "status"},
	})
}

func ensureNuvioReportsDashboardCampaignsCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioReportsDashboardCollection(t, app, "Campaigns", nuvioCampaignsCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "subject"},
		&core.TextField{Name: "status"},
		&core.NumberField{Name: "recipientsCount"},
		&core.TextField{Name: "sentAt"},
	})
}

func ensureNuvioReportsDashboardPagesCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioReportsDashboardCollection(t, app, "Pages", nuvioPagesCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "title"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "slug"},
		&core.TextField{Name: "path"},
		&core.TextField{Name: "url"},
		&core.TextField{Name: "seo_title"},
		&core.TextField{Name: "seo_description"},
		&core.TextField{Name: "seo_social_image"},
		&core.BoolField{Name: "seo_noindex"},
	})
}

func ensureNuvioReportsDashboardCollection(
	t testing.TB,
	app *tests.TestApp,
	collectionName string,
	collectionID string,
	fields []core.Field,
) *core.Collection {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId(collectionID)
	if err == nil {
		return collection
	}

	collection = core.NewBaseCollection(collectionName, collectionID)
	for _, field := range fields {
		collection.Fields.Add(field)
	}

	if saveErr := app.Save(collection); saveErr != nil {
		t.Fatalf("failed to create %s collection: %v", collectionName, saveErr)
	}

	return collection
}

func upsertNuvioReportsDashboardWebsite(
	t testing.TB,
	app *tests.TestApp,
	collection *core.Collection,
	websiteID string,
	slug string,
	name string,
) {
	t.Helper()

	record, err := app.FindFirstRecordByFilter(collection, "id={:id}", dbx.Params{"id": websiteID})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("failed to lookup website %s: %v", websiteID, err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		record = core.NewRecord(collection)
		record.Id = websiteID
	}

	record.Set("name", name)
	record.Set("title", name)
	record.Set("slug", slug)
	record.Set("domain", slug+".example.test")
	record.Set("status", "active")
	record.Set("seoTitle", name+" SEO")
	record.Set("seoDescription", name+" description")
	record.Set("seoImage", "social.jpg")
	record.Set("settings", map[string]any{
		"featureFlags": map[string]any{
			"reports": true,
		},
		"reports": map[string]any{
			"analytics": map[string]any{
				"enabled":  true,
				"provider": "umami",
				"siteId":   "site-" + slug,
				"apiUrl":   "https://attacker.example.test/api",
			},
		},
	})

	if saveErr := app.Save(record); saveErr != nil {
		t.Fatalf("failed to save website %s: %v", websiteID, saveErr)
	}
}

func upsertNuvioReportsDashboardRecord(
	t testing.TB,
	app *tests.TestApp,
	collection *core.Collection,
	recordID string,
	values map[string]any,
) {
	t.Helper()

	record, err := app.FindFirstRecordByFilter(collection, "id={:id}", dbx.Params{"id": recordID})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("failed to lookup %s record %s: %v", collection.Name, recordID, err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		record = core.NewRecord(collection)
		record.Id = recordID
	}

	for field, value := range values {
		record.Set(field, value)
	}

	if saveErr := app.Save(record); saveErr != nil {
		t.Fatalf("failed to save %s record %s: %v", collection.Name, recordID, saveErr)
	}
}
