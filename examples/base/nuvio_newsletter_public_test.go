package main

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const (
	nuvioPublicNewsletterWebsiteID = "pubnewswebsite1"
	nuvioPublicNewsletterSlug      = "public-news-site"
)

func TestNuvioNewsletterPublicSubscribeEndpoint(t *testing.T) {
	t.Parallel()

	tooLongName := strings.Repeat("n", nuvioNewsletterMaxNameLen+1)
	tooLongEmail := strings.Repeat("a", nuvioNewsletterPublicEmailMaxLen+1)

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid newsletter subscribe with website slug succeeds",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/subscribe",
			Body: strings.NewReader(`{
				"websiteSlug":"` + nuvioPublicNewsletterSlug + `",
				"email":"alice@example.test",
				"name":"Alice Example"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, false)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"ok":true`,
				`"status":"active"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := findNuvioNewsletterBackofficeSubscriberByWebsiteEmail(
					t,
					app,
					nuvioPublicNewsletterWebsiteID,
					"alice@example.test",
				)
				assertNuvioNewsletterBackofficeRecordField(t, record, "status", nuvioNewsletterStatusActive)
			},
		},
		{
			Name:   "missing website context returns clean validation error",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/subscribe",
			Body: strings.NewReader(`{
				"email":"alice@example.test",
				"name":"Alice Example"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, false)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Missing website context."`},
		},
		{
			Name:   "invalid email is rejected cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/subscribe",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNewsletterWebsiteID + `",
				"email":"alice",
				"name":"Alice Example"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, false)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"A valid email is required."`},
		},
		{
			Name:   "overlong email is rejected cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/subscribe",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNewsletterWebsiteID + `",
				"email":"` + tooLongEmail + `",
				"name":"Alice Example"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, false)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Email is too long. Maximum 320 characters."`},
		},
		{
			Name:   "overlong name is rejected cleanly",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/subscribe",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNewsletterWebsiteID + `",
				"email":"alice@example.test",
				"name":"` + tooLongName + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, false)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Name is too long. Maximum 200 characters."`},
		},
		{
			Name:   "unknown subscribe payload field is rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/subscribe",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNewsletterWebsiteID + `",
				"email":"alice@example.test",
				"name":"Alice Example",
				"hacker":"1"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, false)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Field \"hacker\" is not allowed in this endpoint."`},
		},
		{
			Name:   "duplicate subscribe does not create duplicate records",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/subscribe",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNewsletterWebsiteID + `",
				"email":"ALICE@EXAMPLE.TEST",
				"name":"Alice Updated"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, false)
				subscribers := ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
				upsertNuvioNewsletterBackofficeRecord(t, app, subscribers, "pubnewssub00001", map[string]any{
					"website":              nuvioPublicNewsletterWebsiteID,
					"email":                "alice@example.test",
					"name":                 "Alice Original",
					"status":               nuvioNewsletterStatusPending,
					"unsubscribeTokenHash": "seeded-unsubscribe-hash",
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"ok":true`,
				`"status":"active"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				count := countNuvioNewsletterSubscribersByWebsiteEmail(
					t,
					app,
					nuvioPublicNewsletterWebsiteID,
					"alice@example.test",
				)
				if count != 1 {
					t.Fatalf("expected exactly one subscriber after duplicate subscribe, got %d", count)
				}
			},
		},
		{
			Name:   "disabled newsletter feature returns clean response",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/subscribe",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioPublicNewsletterWebsiteID + `",
				"email":"alice@example.test",
				"name":"Alice Example"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, false, false)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Newsletter is unavailable for this website."`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioNewsletterPublicLifecycleEndpoints(t *testing.T) {
	t.Parallel()

	validConfirmToken := "confirm-token-valid-01"
	expiredConfirmToken := "confirm-token-expired-01"
	validUnsubscribeToken := "unsubscribe-token-valid-01"
	alreadyUnsubscribedToken := "unsubscribe-token-already-01"

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid confirm token confirms subscriber",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/confirm?format=json&token=" + validConfirmToken,
			Headers: map[string]string{
				"Accept": "application/json",
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, true)
				subscribers := ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
				upsertNuvioNewsletterBackofficeRecord(t, app, subscribers, "pubnscfrmv00001", map[string]any{
					"website":                    nuvioPublicNewsletterWebsiteID,
					"email":                      "confirm-valid@example.test",
					"name":                       "Confirm Valid",
					"status":                     nuvioNewsletterStatusPending,
					"confirmationTokenHash":      hashNuvioNewsletterToken(validConfirmToken),
					"confirmationTokenExpiresAt": time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339),
					"unsubscribeTokenHash":       "unsubscribe-hash-confirm-valid",
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"ok":true`,
				`"status":"active"`,
				`"message":"Subscription confirmed."`,
			},
			NotExpectedContent: []string{
				"confirmationTokenHash",
				"unsubscribeTokenHash",
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := findNuvioNewsletterBackofficeSubscriberByWebsiteEmail(
					t,
					app,
					nuvioPublicNewsletterWebsiteID,
					"confirm-valid@example.test",
				)
				assertNuvioNewsletterBackofficeRecordField(t, record, "status", nuvioNewsletterStatusActive)
				assertNuvioNewsletterBackofficeRecordField(t, record, "confirmationTokenHash", "")
				assertNuvioNewsletterBackofficeRecordField(t, record, "confirmationTokenExpiresAt", "")
			},
		},
		{
			Name:   "invalid confirm token returns safe response",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/confirm?format=json&token=missing-confirm-token",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, true)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Invalid or expired confirmation link."`},
			NotExpectedContent: []string{
				"confirmationTokenHash",
				"unsubscribeTokenHash",
			},
		},
		{
			Name:   "expired confirm token returns safe response",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/confirm?format=json&token=" + expiredConfirmToken,
			Headers: map[string]string{
				"Accept": "application/json",
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, true)
				subscribers := ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
				upsertNuvioNewsletterBackofficeRecord(t, app, subscribers, "pubnscfrme00001", map[string]any{
					"website":                    nuvioPublicNewsletterWebsiteID,
					"email":                      "confirm-expired@example.test",
					"name":                       "Confirm Expired",
					"status":                     nuvioNewsletterStatusPending,
					"confirmationTokenHash":      hashNuvioNewsletterToken(expiredConfirmToken),
					"confirmationTokenExpiresAt": time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339),
					"unsubscribeTokenHash":       "unsubscribe-hash-confirm-expired",
				})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Invalid or expired confirmation link."`},
			NotExpectedContent: []string{
				"confirmationTokenHash",
				"unsubscribeTokenHash",
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := findNuvioNewsletterBackofficeSubscriberByWebsiteEmail(
					t,
					app,
					nuvioPublicNewsletterWebsiteID,
					"confirm-expired@example.test",
				)
				assertNuvioNewsletterBackofficeRecordField(t, record, "confirmationTokenHash", "")
				assertNuvioNewsletterBackofficeRecordField(t, record, "confirmationTokenExpiresAt", "")
			},
		},
		{
			Name:   "valid unsubscribe token unsubscribes subscriber",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/unsubscribe?format=json&token=" + validUnsubscribeToken,
			Headers: map[string]string{
				"Accept": "application/json",
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, true)
				subscribers := ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
				upsertNuvioNewsletterBackofficeRecord(t, app, subscribers, "pubnsunsbv00001", map[string]any{
					"website":              nuvioPublicNewsletterWebsiteID,
					"email":                "unsubscribe-valid@example.test",
					"name":                 "Unsubscribe Valid",
					"status":               nuvioNewsletterStatusActive,
					"unsubscribeTokenHash": hashNuvioNewsletterToken(validUnsubscribeToken),
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"ok":true`,
				`"status":"unsubscribed"`,
				`"message":"You have been unsubscribed."`,
			},
			NotExpectedContent: []string{
				"confirmationTokenHash",
				"unsubscribeTokenHash",
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := findNuvioNewsletterBackofficeSubscriberByWebsiteEmail(
					t,
					app,
					nuvioPublicNewsletterWebsiteID,
					"unsubscribe-valid@example.test",
				)
				assertNuvioNewsletterBackofficeRecordField(t, record, "status", nuvioNewsletterStatusUnsubscribed)
			},
		},
		{
			Name:   "invalid unsubscribe token returns safe response",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/unsubscribe?format=json&token=missing-unsubscribe-token",
			Headers: map[string]string{
				"Accept": "application/json",
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, true)
				ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Invalid unsubscribe link."`},
			NotExpectedContent: []string{
				"confirmationTokenHash",
				"unsubscribeTokenHash",
			},
		},
		{
			Name:   "unsubscribe is idempotent for already unsubscribed subscriber",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/unsubscribe?format=json&token=" + alreadyUnsubscribedToken,
			Headers: map[string]string{
				"Accept": "application/json",
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioPublicNewsletterRoutes(t, app, e)
				seedNuvioPublicNewsletterWebsite(t, app, true, true)
				subscribers := ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
				upsertNuvioNewsletterBackofficeRecord(t, app, subscribers, "pubnsunsba00001", map[string]any{
					"website":              nuvioPublicNewsletterWebsiteID,
					"email":                "unsubscribe-already@example.test",
					"name":                 "Unsubscribe Already",
					"status":               nuvioNewsletterStatusUnsubscribed,
					"unsubscribeTokenHash": hashNuvioNewsletterToken(alreadyUnsubscribedToken),
					"unsubscribedAt":       time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"ok":true`,
				`"status":"unsubscribed"`,
				`"alreadyUnsubscribed":true`,
			},
			NotExpectedContent: []string{
				"confirmationTokenHash",
				"unsubscribeTokenHash",
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioPublicNewsletterRoutes(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioNewsletterRoutes(e)
}

func seedNuvioPublicNewsletterWebsite(
	t testing.TB,
	app *tests.TestApp,
	newsletterAvailable bool,
	doubleOptIn bool,
) {
	t.Helper()

	websitesCollection := ensureNuvioBackofficeWebsitesCollection(t, app)
	record := core.NewRecord(websitesCollection)
	record.Id = nuvioPublicNewsletterWebsiteID
	record.Set("name", "Public Newsletter Site")
	record.Set("title", "Public Newsletter Site")
	record.Set("slug", nuvioPublicNewsletterSlug)
	record.Set("domain", "newsletter.example.test")
	record.Set("status", "active")
	record.Set("settings", map[string]any{
		"featureFlags": map[string]any{
			"newsletter": newsletterAvailable,
		},
		"newsletter": map[string]any{
			"enabled":     true,
			"doubleOptIn": doubleOptIn,
		},
	})

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to seed website: %v", err)
	}
}

func countNuvioNewsletterSubscribersByWebsiteEmail(
	t testing.TB,
	app *tests.TestApp,
	websiteID string,
	email string,
) int {
	t.Helper()

	subscribersCollection := ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
	records, err := app.FindRecordsByFilter(
		subscribersCollection,
		"website={:website}",
		"",
		nuvioNewsletterMaxSubscriberScan,
		0,
		dbx.Params{"website": websiteID},
	)
	if err != nil {
		t.Fatalf("failed to load subscribers for website %s: %v", websiteID, err)
	}

	normalizedTargetEmail, ok := normalizeNuvioEmail(email)
	if !ok {
		t.Fatalf("invalid email used in test assertion: %s", email)
	}

	matches := 0
	for _, record := range records {
		normalizedEmail, emailOk := normalizeNuvioEmail(record.GetString("email"))
		if emailOk && normalizedEmail == normalizedTargetEmail {
			matches++
		}
	}

	return matches
}
