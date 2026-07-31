package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const (
	nuvioPublicNotificationsWebsiteID = "pubntfwebsite01"
	nuvioPublicNotificationsSlug      = "public-notify-site"
)

func TestNuvioContactPublicSubmitEndpoint(t *testing.T) {
	t.Parallel()

	tooLongName := strings.Repeat("n", nuvioPublicContactNameMaxLen+1)
	tooLongPhone := strings.Repeat("1", nuvioPublicContactPhoneMaxLen+1)
	tooLongSubject := strings.Repeat("s", nuvioPublicContactSubjectMaxLen+1)
	tooLongMessage := strings.Repeat("m", nuvioPublicContactMessageMaxLen+1)
	tooLongSource := strings.Repeat("x", nuvioPublicContactSourceMaxLen+1)
	tooLongPage := strings.Repeat("p", nuvioPublicContactPageMaxLen+1)

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid contact submit with website slug succeeds",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteSlug":"` + nuvioPublicNotificationsSlug + `",
				"name":"Alice Example",
				"email":"alice@example.test",
				"phone":"+351900000001",
				"subject":"Need info",
				"message":"Please contact me.",
				"source":"feature_contact",
				"page":"/site/public-notify-site/home"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
				ensureNuvioLeadsDashboardWhatsappCollection(t, app)
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"ok":true`, `"confirmationMessage":"Thank you for contacting us. We'll reply soon."`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				record := mustFindNuvioPublicContactRecordByEmail(t, app, "alice@example.test")
				assertNuvioLeadsDashboardRecordFieldString(t, record, "website", nuvioPublicNotificationsWebsiteID)
				assertNuvioLeadsDashboardRecordFieldString(t, record, "status", "new")
				assertNuvioLeadsDashboardRecordFieldString(t, record, "name", "Alice Example")
				assertNuvioLeadsDashboardRecordFieldString(t, record, "subject", "Need info")
				assertNuvioLeadsDashboardRecordFieldString(t, record, "source", "feature_contact")
				assertNuvioLeadsDashboardRecordFieldString(t, record, "page", "/site/public-notify-site/home")
			},
		},
		{
			Name:   "missing website context returns clean validation error",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"name":"Alice Example",
				"email":"alice@example.test",
				"message":"Please contact me."
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Missing website context."`},
		},
		{
			Name:   "invalid email is rejected cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"Alice Example",
				"email":"alice",
				"message":"Please contact me."
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"A valid email is required."`},
		},
		{
			Name:   "unknown contact payload field is rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"Alice Example",
				"email":"alice@example.test",
				"message":"Please contact me.",
				"hacker":"1"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Field \"hacker\" is not allowed in this endpoint."`},
		},
		{
			Name:   "contact overlong name fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"` + tooLongName + `",
				"email":"alice@example.test",
				"message":"Please contact me."
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Name is too long. Maximum 160 characters."`},
		},
		{
			Name:   "contact overlong phone fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"Alice Example",
				"email":"alice@example.test",
				"phone":"` + tooLongPhone + `",
				"message":"Please contact me."
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Phone is too long. Maximum 80 characters."`},
		},
		{
			Name:   "contact overlong subject fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"Alice Example",
				"email":"alice@example.test",
				"subject":"` + tooLongSubject + `",
				"message":"Please contact me."
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Subject is too long. Maximum 200 characters."`},
		},
		{
			Name:   "contact overlong message fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"Alice Example",
				"email":"alice@example.test",
				"message":"` + tooLongMessage + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Message is too long. Maximum 4000 characters."`},
		},
		{
			Name:   "contact overlong source fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"Alice Example",
				"email":"alice@example.test",
				"message":"Please contact me.",
				"source":"` + tooLongSource + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Source is too long. Maximum 120 characters."`},
		},
		{
			Name:   "contact overlong page fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"Alice Example",
				"email":"alice@example.test",
				"message":"Please contact me.",
				"page":"` + tooLongPage + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Page is too long. Maximum 200 characters."`},
		},
		{
			Name:   "contact disabled feature returns clean response",
			Method: http.MethodPost,
			URL:    "/api/nuvio/contact/submit",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"name":"Alice Example",
				"email":"alice@example.test",
				"message":"Please contact me."
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, false, true)
				ensureNuvioLeadsDashboardContactsCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Contact Form feature is unavailable for this website."`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioContactPublicSubmitRateLimit(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer app.Cleanup()

	baseRouter, err := apis.NewRouter(app)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}

	serveEvent := &core.ServeEvent{
		App:    app,
		Router: baseRouter,
	}

	serveErr := app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
		registerNuvioLeadsRoutes(e)
		registerNuvioPublicContentRoutes(e)
		seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
		seedNuvioCMSBackofficeDashboardData(t, app)
		ensureNuvioLeadsDashboardContactsCollection(t, app)
		setNuvioContactSubmitRateLimitConfigForTest(app, nuvioContactSubmitRateLimitConfig{
			MaxRequests: 2,
			Window:      time.Minute,
		})
		if config := resolveNuvioContactSubmitRateLimitConfig(app); config.MaxRequests != 2 {
			t.Fatalf("expected test rate-limit config max=2, got %d", config.MaxRequests)
		}

		mux, err := e.Router.BuildMux()
		if err != nil {
			t.Fatalf("failed to build router mux: %v", err)
		}

		for i := 1; i <= 2; i++ {
			res := submitNuvioContactForRateLimitTest(t, mux, fmt.Sprintf("allowed-%d@example.test", i), "198.51.100.44")
			if res.Code != http.StatusOK {
				t.Fatalf("expected allowed request %d to return 200, got %d: %s", i, res.Code, res.Body.String())
			}
		}

		if count := countNuvioPublicContactRecordsForWebsite(t, app); count != 2 {
			t.Fatalf("expected 2 saved contact records before limit, got %d", count)
		}

		limited := submitNuvioContactForRateLimitTest(
			t,
			mux,
			"limited@example.test",
			"198.51.100.44",
		)
		if limited.Code != http.StatusTooManyRequests {
			t.Fatalf("expected third request to return 429, got %d: %s", limited.Code, limited.Body.String())
		}
		if retryAfter := limited.Header().Get("Retry-After"); retryAfter == "" {
			t.Fatalf("expected Retry-After header on rate-limited response")
		}
		if !strings.Contains(limited.Body.String(), nuvioContactSubmitRateLimitMessage) {
			t.Fatalf("expected generic rate-limit response message, got %s", limited.Body.String())
		}

		legacyLimited := submitNuvioContactForRateLimitPathTest(
			t,
			mux,
			"/api/nuvio/leads/contact/submit",
			"legacy-limited@example.test",
			"198.51.100.44",
		)
		if legacyLimited.Code != http.StatusTooManyRequests {
			t.Fatalf("expected compatibility contact route to share the rate limit, got %d: %s", legacyLimited.Code, legacyLimited.Body.String())
		}

		if count := countNuvioPublicContactRecordsForWebsite(t, app); count != 2 {
			t.Fatalf("expected rate-limited request not to create a contact record, got %d records", count)
		}

		contentRes := httptest.NewRecorder()
		contentReq := httptest.NewRequest(
			http.MethodGet,
			"/api/nuvio/public/content?websiteSlug=alpha-cms&pageSlug=home",
			nil,
		)
		mux.ServeHTTP(contentRes, contentReq)
		if contentRes.Code != http.StatusOK {
			t.Fatalf("expected unrelated public content route to remain available, got %d: %s", contentRes.Code, contentRes.Body.String())
		}

		return nil
	})
	if serveErr != nil {
		t.Fatalf("failed to trigger serve hook: %v", serveErr)
	}
}

func TestNuvioWhatsappPublicInteractionEndpoint(t *testing.T) {
	t.Parallel()

	tooLongSource := strings.Repeat("x", nuvioPublicWhatsappSourceMaxLen+1)
	tooLongPage := strings.Repeat("p", nuvioPublicWhatsappPageMaxLen+1)

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid whatsapp interaction succeeds",
			Method: http.MethodPost,
			URL:    "/api/nuvio/whatsapp/interactions",
			Body: strings.NewReader(`{
				"websiteSlug":"` + nuvioPublicNotificationsSlug + `",
				"source":"floating_button",
				"page":"/site/public-notify-site/home"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardWhatsappCollection(t, app)
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`"ok":true`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				record := mustFindNuvioPublicWhatsappRecordByPage(t, app, "/site/public-notify-site/home")
				assertNuvioLeadsDashboardRecordFieldString(t, record, "website", nuvioPublicNotificationsWebsiteID)
				assertNuvioLeadsDashboardRecordFieldString(t, record, "source", "floating_button")
				assertNuvioLeadsDashboardRecordFieldString(t, record, "status", "new")
			},
		},
		{
			Name:   "empty whatsapp tracking payload fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/whatsapp/interactions",
			Body:   strings.NewReader(`{"websiteId":"` + nuvioPublicNotificationsWebsiteID + `"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardWhatsappCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"At least source or page is required."`},
		},
		{
			Name:   "whatsapp overlong source fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/whatsapp/interactions",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"source":"` + tooLongSource + `",
				"page":"/site/public-notify-site/home"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardWhatsappCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Source is too long. Maximum 120 characters."`},
		},
		{
			Name:   "whatsapp overlong page fails cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/whatsapp/interactions",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"source":"floating_button",
				"page":"` + tooLongPage + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardWhatsappCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Page is too long. Maximum 200 characters."`},
		},
		{
			Name:   "unknown whatsapp payload field is rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/whatsapp/interactions",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"source":"floating_button",
				"page":"/site/public-notify-site/home",
				"message":"ignore"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, true)
				ensureNuvioLeadsDashboardWhatsappCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Field \"message\" is not allowed in this endpoint."`},
		},
		{
			Name:   "whatsapp disabled feature returns clean response",
			Method: http.MethodPost,
			URL:    "/api/nuvio/whatsapp/interactions",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNotificationsWebsiteID + `",
				"source":"floating_button",
				"page":"/site/public-notify-site/home"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicLeadsNotificationsRoutes(t, app, e)
				seedNuvioPublicLeadsNotificationsWebsite(t, app, true, false)
				ensureNuvioLeadsDashboardWhatsappCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"WhatsApp feature is unavailable for this website."`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setNuvioContactSubmitRateLimitConfigForTest(
	app core.App,
	config nuvioContactSubmitRateLimitConfig,
) {
	app.Store().Set(nuvioContactSubmitRateLimitConfigStoreKey, config)
	app.Store().Remove(nuvioContactSubmitRateLimitStoreKey)
}

func submitNuvioContactForRateLimitTest(
	t testing.TB,
	mux http.Handler,
	email string,
	forwardedFor string,
) *httptest.ResponseRecorder {
	t.Helper()

	return submitNuvioContactForRateLimitPathTest(
		t,
		mux,
		"/api/nuvio/contact/submit",
		email,
		forwardedFor,
	)
}

func submitNuvioContactForRateLimitPathTest(
	t testing.TB,
	mux http.Handler,
	path string,
	email string,
	forwardedFor string,
) *httptest.ResponseRecorder {
	t.Helper()

	body := strings.NewReader(`{
		"websiteSlug":"` + nuvioPublicNotificationsSlug + `",
		"name":"Rate Limit Test",
		"email":"` + email + `",
		"message":"Please contact me."
	}`)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, path, body)
	request.Header.Set("content-type", "application/json")
	request.RemoteAddr = "203.0.113.10:1234"
	if forwardedFor != "" {
		request.Header.Set("X-Forwarded-For", forwardedFor)
	}

	mux.ServeHTTP(recorder, request)
	return recorder
}

func countNuvioPublicContactRecordsForWebsite(t testing.TB, app core.App) int {
	t.Helper()

	contactsCollection, err := app.FindCollectionByNameOrId(nuvioContactsCollectionID)
	if err != nil {
		t.Fatalf("failed to load contacts collection: %v", err)
	}

	records, err := app.FindRecordsByFilter(
		contactsCollection,
		"website={:website}",
		"",
		50,
		0,
		dbx.Params{"website": nuvioPublicNotificationsWebsiteID},
	)
	if err != nil {
		t.Fatalf("failed to count contact records: %v", err)
	}

	return len(records)
}
func setupNuvioPublicLeadsNotificationsRoutes(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioLeadsRoutes(e)
}

func seedNuvioPublicLeadsNotificationsWebsite(
	t testing.TB,
	app *tests.TestApp,
	contactAvailable bool,
	whatsappAvailable bool,
) {
	t.Helper()

	websitesCollection := ensureNuvioBackofficeWebsitesCollection(t, app)
	website := core.NewRecord(websitesCollection)
	website.Id = nuvioPublicNotificationsWebsiteID
	website.Set("name", "Public Notify Site")
	website.Set("title", "Public Notify Site")
	website.Set("slug", nuvioPublicNotificationsSlug)
	website.Set("domain", "notify.example.test")
	website.Set("status", "active")
	website.Set("settings", map[string]any{
		"featureFlags": map[string]any{
			"contactForm": contactAvailable,
			"whatsapp":    whatsappAvailable,
		},
		"contactForm": map[string]any{
			"enabled": true,
			"fields": map[string]any{
				"phone": true,
			},
			"emailNotifications": map[string]any{
				"enabled": false,
				"to":      []string{},
				"cc":      []string{},
			},
		},
		"whatsapp": map[string]any{
			"enabled": true,
			"phone":   "+351900000999",
			"emailNotifications": map[string]any{
				"enabled": false,
				"to":      []string{},
				"cc":      []string{},
			},
		},
	})

	if saveErr := app.Save(website); saveErr != nil {
		t.Fatalf("failed to save seeded website: %v", saveErr)
	}
}

func mustFindNuvioPublicContactRecordByEmail(t testing.TB, app *tests.TestApp, email string) *core.Record {
	t.Helper()

	contactsCollection, err := app.FindCollectionByNameOrId(nuvioContactsCollectionID)
	if err != nil {
		t.Fatalf("failed to load contacts collection: %v", err)
	}

	record, err := app.FindFirstRecordByFilter(contactsCollection, "email={:email}", dbx.Params{"email": email})
	if err != nil {
		t.Fatalf("failed to load contact record by email: %v", err)
	}

	return record
}

func mustFindNuvioPublicWhatsappRecordByPage(t testing.TB, app *tests.TestApp, page string) *core.Record {
	t.Helper()

	whatsappCollection, err := app.FindCollectionByNameOrId(nuvioWhatsappCollectionID)
	if err != nil {
		t.Fatalf("failed to load whatsapp collection: %v", err)
	}

	record, err := app.FindFirstRecordByFilter(whatsappCollection, "page={:page}", dbx.Params{"page": page})
	if err != nil {
		t.Fatalf("failed to load whatsapp record by page: %v", err)
	}

	return record
}
