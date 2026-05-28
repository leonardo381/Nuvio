package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const (
	nuvioBookingBackofficeAlphaAvailabilityID = "avlalpha0000001"
	nuvioBookingBackofficeAlphaExceptionID    = "excalpha0000001"
)

func TestNuvioBookingBackofficeConfigEndpoints(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin can create service",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/services",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"name":"Config Service",
				"description":"Config test description",
				"durationMinutes":60,
				"displayOrder":25,
				"active":true
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID, nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"name":"Config Service"`,
				`"durationMinutes":60`,
			},
			NotExpectedContent: []string{
				`"settings":`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record, err := app.FindFirstRecordByFilter(
					nuvioBookingServicesCollectionID,
					"website={:website} && name={:name}",
					dbx.Params{
						"website": nuvioBookingDashboardAlphaWebsiteID,
						"name":    "Config Service",
					},
				)
				if err != nil {
					t.Fatalf("expected created service record: %v", err)
				}
				assertNuvioBookingBackofficeRecordFieldString(t, record, "website", nuvioBookingDashboardAlphaWebsiteID)
			},
		},
		{
			Name:   "client assigned can update service",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/services/" + nuvioBookingBackofficeAlphaServiceID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"description":"Client updated description",
				"durationMinutes":55,
				"active":false,
				"calendarBlockingMode":"website"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"description":"Client updated description"`,
				`"durationMinutes":55`,
				`"active":false`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeGenericRecord(t, app, nuvioBookingServicesCollectionID, nuvioBookingBackofficeAlphaServiceID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "website", nuvioBookingDashboardAlphaWebsiteID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "description", "Client updated description")
			},
		},
		{
			Name:   "client unassigned denied for service create",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/services",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"name":"Blocked Service",
				"durationMinutes":30
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "service update rejects arbitrary fields and website change",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/services/" + nuvioBookingBackofficeAlphaServiceID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"website":"` + nuvioBookingDashboardBetaWebsiteID + `",
				"name":"Attempted Hijack"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID, nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeGenericRecord(t, app, nuvioBookingServicesCollectionID, nuvioBookingBackofficeAlphaServiceID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "website", nuvioBookingDashboardAlphaWebsiteID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "name", "Alpha Service")
			},
		},
		{
			Name:   "admin can create availability",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/availability",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"service":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"dayOfWeek":"tue",
				"startTime":"09:00",
				"endTime":"11:00",
				"active":true
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"dayOfWeek":"tue"`,
				`"startTime":"09:00"`,
			},
		},
		{
			Name:   "availability create validates service website ownership",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/availability",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"service":"` + nuvioBookingBackofficeBetaServiceID + `",
				"dayOfWeek":"wed",
				"startTime":"09:00",
				"endTime":"11:00",
				"active":true
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID, nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "client assigned can update availability",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/availability/" + nuvioBookingBackofficeAlphaAvailabilityID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"startTime":"10:00",
				"endTime":"12:00",
				"active":false
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"startTime":"10:00"`,
				`"endTime":"12:00"`,
				`"active":false`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeGenericRecord(t, app, nuvioBookingAvailabilityCollectionID, nuvioBookingBackofficeAlphaAvailabilityID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "website", nuvioBookingDashboardAlphaWebsiteID)
			},
		},
		{
			Name:   "availability update validates time range",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/availability/" + nuvioBookingBackofficeAlphaAvailabilityID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"startTime":"14:00",
				"endTime":"12:00"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "admin can create exception",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/exceptions",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"service":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"date":"2099-02-10",
				"type":"customHours",
				"startTime":"14:00",
				"endTime":"16:00",
				"note":"Special opening",
				"active":true
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"type":"customHours"`,
				`"startTime":"14:00"`,
				`"endTime":"16:00"`,
			},
		},
		{
			Name:   "exception create validates type",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/exceptions",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"date":"2099-02-11",
				"type":"holiday",
				"active":true
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "exception update validates service website ownership",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/exceptions/" + nuvioBookingBackofficeAlphaExceptionID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"service":"` + nuvioBookingBackofficeBetaServiceID + `",
				"type":"closed"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID, nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "client assigned can update exception",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/exceptions/" + nuvioBookingBackofficeAlphaExceptionID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"type":"closed",
				"active":false,
				"note":"Client updated note"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"type":"closed"`,
				`"active":false`,
			},
		},
		{
			Name:   "admin can update booking rules and preserve hidden settings",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/settings/rules",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"rules":{
					"minNoticeHours":5,
					"bookingWindowDays":30,
					"bufferMinutes":15
				}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				injectNuvioBookingBackofficeHiddenSettings(t, app, nuvioBookingDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `"`,
				`"minNoticeHours":5`,
				`"bookingWindowDays":30`,
				`"bufferMinutes":15`,
			},
			NotExpectedContent: []string{
				`"settings":`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website := mustFindNuvioBookingBackofficeGenericRecord(t, app, nuvioWebsitesCollectionID, nuvioBookingDashboardAlphaWebsiteID)
				settings := parseNuvioSettingsObject(website.Get("settings"))

				bookingSettings, ok := toStringAnyMap(settings["booking"])
				if !ok {
					t.Fatalf("expected booking settings object")
				}
				if strings.TrimSpace(parseStringValue(bookingSettings["privateKey"])) != "keep-me" {
					t.Fatalf("expected hidden booking privateKey to be preserved")
				}

				newsletterSettings, ok := toStringAnyMap(settings["newsletter"])
				if !ok {
					t.Fatalf("expected newsletter settings object")
				}
				if enabled, ok := parseBoolValue(newsletterSettings["enabled"]); !ok || !enabled {
					t.Fatalf("expected newsletter.enabled to be preserved")
				}

				rulesSettings, ok := toStringAnyMap(bookingSettings["rules"])
				if !ok {
					t.Fatalf("expected booking rules object")
				}
				if parseNuvioNonNegativeInt(rulesSettings["minNoticeHours"], 0) != 5 {
					t.Fatalf("expected minNoticeHours to be updated")
				}
				if parseNuvioNonNegativeInt(rulesSettings["bookingWindowDays"], 0) != 30 {
					t.Fatalf("expected bookingWindowDays to be updated")
				}
				if parseNuvioNonNegativeInt(rulesSettings["bufferMinutes"], 0) != 15 {
					t.Fatalf("expected bufferMinutes to be updated")
				}
			},
		},
		{
			Name:   "rules endpoint validates numeric rule values",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/settings/rules",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"rules":{"minNoticeHours":"not-a-number"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "rules endpoint rejects unrelated settings overwrite",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/settings/rules",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"settings":{"booking":{"enabled":false}},
				"rules":{"minNoticeHours":1}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "client unassigned denied for rules endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/settings/rules",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"rules":{"minNoticeHours":2}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "missing role denied for config write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/services",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"name":"Role blocked",
				"durationMinutes":30
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role denied for config write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/services",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"name":"Role blocked",
				"durationMinutes":30
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied for config write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/services",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"name":"No auth",
				"durationMinutes":30
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeConfigRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioBookingBackofficeConfigRoutes(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioBookingRoutes(e)
}

func mustFindNuvioBookingBackofficeGenericRecord(
	t testing.TB,
	app *tests.TestApp,
	collectionID string,
	recordID string,
) *core.Record {
	t.Helper()

	record, err := app.FindRecordById(collectionID, recordID)
	if err != nil {
		t.Fatalf("failed to load record %s from %s: %v", recordID, collectionID, err)
	}
	return record
}

func injectNuvioBookingBackofficeHiddenSettings(t testing.TB, app *tests.TestApp, websiteID string) {
	t.Helper()

	website := mustFindNuvioBookingBackofficeGenericRecord(t, app, nuvioWebsitesCollectionID, websiteID)
	settings := parseNuvioSettingsObject(website.Get("settings"))

	bookingSettings, ok := toStringAnyMap(settings["booking"])
	if !ok {
		bookingSettings = map[string]any{}
	}
	bookingSettings["privateKey"] = "keep-me"
	settings["booking"] = bookingSettings

	settings["newsletter"] = map[string]any{
		"enabled": true,
		"legacy":  "preserve",
	}

	website.Set("settings", settings)
	if err := app.Save(website); err != nil {
		t.Fatalf("failed to inject hidden settings: %v", err)
	}
}
