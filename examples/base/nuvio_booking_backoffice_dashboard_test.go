package main

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const (
	nuvioBookingDashboardAlphaWebsiteID = "bookalpha000001"
	nuvioBookingDashboardBetaWebsiteID  = "bookbeta0000002"
	nuvioBookingDashboardGammaWebsiteID = "bookgamma000003"
)

func TestNuvioBookingBackofficeDashboardEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin superuser receives scoped booking datasets",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/backoffice/dashboard?websiteId=" + nuvioBookingDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeDashboardRoute(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioBookingDashboardAlphaWebsiteID, nuvioBookingDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `"`,
				`"displayName":"Alpha Booking"`,
				`"businessNotificationsReady":true`,
				`"usingContactFormFallback":true`,
				`"services":[`,
				`"availability":[`,
				`"exceptions":[`,
				`"appointments":[`,
				`"name":"Alpha Service"`,
				`"name":"Alpha Visitor"`,
			},
			NotExpectedContent: []string{
				`"name":"Beta Service"`,
				`"name":"Beta Visitor"`,
				`"settings":`,
				`"manageToken"`,
				`"providerPayload"`,
				`"icsPayload"`,
			},
		},
		{
			Name:   "admin receives empty dataset keys when website has no records",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/backoffice/dashboard?websiteId=" + nuvioBookingDashboardGammaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeDashboardRoute(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioBookingDashboardAlphaWebsiteID, nuvioBookingDashboardBetaWebsiteID, nuvioBookingDashboardGammaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"websiteId":"` + nuvioBookingDashboardGammaWebsiteID + `"`,
				`"services":[]`,
				`"availability":[]`,
				`"exceptions":[]`,
				`"appointments":[]`,
			},
		},
		{
			Name:   "client superuser receives only assigned website booking data",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/backoffice/dashboard?websiteId=" + nuvioBookingDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeDashboardRoute(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `"`,
				`"name":"Alpha Service"`,
			},
			NotExpectedContent: []string{
				`"name":"Beta Service"`,
				`"name":"Beta Visitor"`,
				`"manageToken"`,
			},
		},
		{
			Name:   "client superuser denied for unassigned website",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/backoffice/dashboard?websiteId=" + nuvioBookingDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeDashboardRoute(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client superuser with no websiteAccess denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/backoffice/dashboard?websiteId=" + nuvioBookingDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeDashboardRoute(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "missing role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/backoffice/dashboard?websiteId=" + nuvioBookingDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeDashboardRoute(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/backoffice/dashboard?websiteId=" + nuvioBookingDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeDashboardRoute(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/backoffice/dashboard?websiteId=" + nuvioBookingDashboardAlphaWebsiteID,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeDashboardRoute(t, app, e)
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

func setupNuvioBookingBackofficeDashboardRoute(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioBookingRoutes(e)
}

func seedNuvioBookingBackofficeDashboardData(t testing.TB, app *tests.TestApp) {
	t.Helper()

	websitesCollection := ensureNuvioBackofficeWebsitesCollection(t, app)
	servicesCollection := ensureNuvioBookingBackofficeDashboardCollection(t, app, "BookingServices", nuvioBookingServicesCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "description"},
		&core.NumberField{Name: "durationMinutes"},
		&core.NumberField{Name: "displayOrder"},
		&core.BoolField{Name: "active"},
		&core.NumberField{Name: "bufferBefore"},
		&core.NumberField{Name: "bufferAfter"},
		&core.NumberField{Name: "price"},
		&core.TextField{Name: "calendarBlockingMode"},
		&core.BoolField{Name: "autoConfirm"},
	})
	availabilityCollection := ensureNuvioBookingBackofficeDashboardCollection(t, app, "BookingAvailability", nuvioBookingAvailabilityCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "service"},
		&core.TextField{Name: "dayOfWeek"},
		&core.TextField{Name: "startTime"},
		&core.TextField{Name: "endTime"},
		&core.BoolField{Name: "active"},
		&core.NumberField{Name: "capacity"},
	})
	exceptionsCollection := ensureNuvioBookingBackofficeDashboardCollection(t, app, "BookingExceptions", nuvioBookingExceptionsCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "service"},
		&core.TextField{Name: "date"},
		&core.TextField{Name: "type"},
		&core.TextField{Name: "startTime"},
		&core.TextField{Name: "endTime"},
		&core.TextField{Name: "reason"},
		&core.TextField{Name: "note"},
		&core.BoolField{Name: "active"},
	})
	appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(t, app, "Appointments", nuvioAppointmentsCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "service"},
		&core.TextField{Name: "status"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "email"},
		&core.TextField{Name: "phone"},
		&core.TextField{Name: "date"},
		&core.TextField{Name: "time"},
		&core.TextField{Name: "notes"},
		&core.TextField{Name: "message"},
		&core.TextField{Name: "internalNotes"},
		&core.TextField{Name: "archivedAt"},
		&core.TextField{Name: "confirmedAt"},
		&core.TextField{Name: "cancelledAt"},
		&core.TextField{Name: "rescheduledAt"},
		&core.TextField{Name: "serviceNameSnapshot"},
		&core.NumberField{Name: "serviceDurationMinutesSnapshot"},
		&core.TextField{Name: "serviceDescriptionSnapshot"},
		&core.TextField{Name: "manageToken"},
		&core.JSONField{Name: "providerPayload"},
		&core.TextField{Name: "icsPayload"},
	})

	upsertNuvioBookingBackofficeWebsiteRecord(t, app, websitesCollection, nuvioBookingDashboardAlphaWebsiteID, "alpha-booking", "Alpha Booking", map[string]any{
		"featureFlags": map[string]any{
			"booking": true,
		},
		"booking": map[string]any{
			"enabled":          true,
			"confirmationMode": "request",
			"rules": map[string]any{
				"minNoticeHours":       2,
				"bookingWindowDays":    21,
				"bufferMinutes":        10,
				"calendarBlockingMode": "service",
			},
			"emailNotifications": map[string]any{
				"enabled": true,
				"to":      []any{},
			},
		},
		"contactForm": map[string]any{
			"emailNotifications": map[string]any{
				"enabled": true,
				"to":      []any{"alpha-team@example.test"},
			},
		},
	})
	upsertNuvioBookingBackofficeWebsiteRecord(t, app, websitesCollection, nuvioBookingDashboardBetaWebsiteID, "beta-booking", "Beta Booking", map[string]any{
		"featureFlags": map[string]any{
			"booking": true,
		},
		"booking": map[string]any{
			"enabled": true,
		},
	})
	upsertNuvioBookingBackofficeWebsiteRecord(t, app, websitesCollection, nuvioBookingDashboardGammaWebsiteID, "gamma-booking", "Gamma Booking", map[string]any{
		"featureFlags": map[string]any{
			"booking": true,
		},
		"booking": map[string]any{
			"enabled": true,
		},
	})

	upsertNuvioBookingBackofficeRecord(t, app, servicesCollection, "svcalpha0000001", map[string]any{
		"website":              nuvioBookingDashboardAlphaWebsiteID,
		"name":                 "Alpha Service",
		"description":          "Alpha service description",
		"durationMinutes":      45,
		"displayOrder":         10,
		"active":               true,
		"bufferBefore":         5,
		"bufferAfter":          10,
		"price":                99.5,
		"calendarBlockingMode": "service",
		"autoConfirm":          false,
	})
	upsertNuvioBookingBackofficeRecord(t, app, servicesCollection, "svcbeta00000002", map[string]any{
		"website":         nuvioBookingDashboardBetaWebsiteID,
		"name":            "Beta Service",
		"description":     "Beta service description",
		"durationMinutes": 30,
		"displayOrder":    50,
		"active":          true,
	})

	upsertNuvioBookingBackofficeRecord(t, app, availabilityCollection, "avlalpha0000001", map[string]any{
		"website":   nuvioBookingDashboardAlphaWebsiteID,
		"service":   "svcalpha0000001",
		"dayOfWeek": "mon",
		"startTime": "09:00",
		"endTime":   "12:00",
		"active":    true,
		"capacity":  1,
	})
	upsertNuvioBookingBackofficeRecord(t, app, availabilityCollection, "avlbeta00000002", map[string]any{
		"website":   nuvioBookingDashboardBetaWebsiteID,
		"service":   "svcbeta00000002",
		"dayOfWeek": "tue",
		"startTime": "10:00",
		"endTime":   "13:00",
		"active":    true,
		"capacity":  1,
	})

	upsertNuvioBookingBackofficeRecord(t, app, exceptionsCollection, "excalpha0000001", map[string]any{
		"website":   nuvioBookingDashboardAlphaWebsiteID,
		"service":   "svcalpha0000001",
		"date":      "2026-06-01",
		"type":      "closed",
		"startTime": "",
		"endTime":   "",
		"reason":    "Holiday",
		"note":      "Office closed",
		"active":    true,
	})
	upsertNuvioBookingBackofficeRecord(t, app, exceptionsCollection, "excbeta00000002", map[string]any{
		"website":   nuvioBookingDashboardBetaWebsiteID,
		"service":   "svcbeta00000002",
		"date":      "2026-06-02",
		"type":      "customHours",
		"startTime": "14:00",
		"endTime":   "16:00",
		"reason":    "Special event",
		"note":      "Limited opening",
		"active":    true,
	})

	upsertNuvioBookingBackofficeRecord(t, app, appointmentsCollection, "aptalpha0000001", map[string]any{
		"website":                        nuvioBookingDashboardAlphaWebsiteID,
		"service":                        "svcalpha0000001",
		"status":                         "confirmed",
		"name":                           "Alpha Visitor",
		"email":                          "alpha.visitor@example.test",
		"phone":                          "+351111111111",
		"date":                           "2026-06-03",
		"time":                           "10:00",
		"notes":                          "Alpha note",
		"message":                        "Alpha message",
		"internalNotes":                  "Internal alpha note",
		"archivedAt":                     "",
		"confirmedAt":                    "2026-06-01T10:00:00Z",
		"cancelledAt":                    "",
		"rescheduledAt":                  "",
		"serviceNameSnapshot":            "Alpha Service",
		"serviceDurationMinutesSnapshot": 45,
		"serviceDescriptionSnapshot":     "Alpha service description",
		"manageToken":                    "sensitive-token-value",
		"providerPayload":                map[string]any{"provider": "internal"},
		"icsPayload":                     "BEGIN:VCALENDAR\n...",
	})
	upsertNuvioBookingBackofficeRecord(t, app, appointmentsCollection, "aptbeta00000002", map[string]any{
		"website":                        nuvioBookingDashboardBetaWebsiteID,
		"service":                        "svcbeta00000002",
		"status":                         "pending",
		"name":                           "Beta Visitor",
		"email":                          "beta.visitor@example.test",
		"phone":                          "+351222222222",
		"date":                           "2026-06-04",
		"time":                           "11:00",
		"notes":                          "Beta note",
		"message":                        "Beta message",
		"internalNotes":                  "Internal beta note",
		"serviceNameSnapshot":            "Beta Service",
		"serviceDurationMinutesSnapshot": 30,
		"serviceDescriptionSnapshot":     "Beta service description",
	})
}

func ensureNuvioBookingBackofficeDashboardCollection(
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

func upsertNuvioBookingBackofficeWebsiteRecord(
	t testing.TB,
	app *tests.TestApp,
	collection *core.Collection,
	websiteID string,
	slug string,
	name string,
	settings map[string]any,
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
	record.Set("settings", settings)

	if saveErr := app.Save(record); saveErr != nil {
		t.Fatalf("failed to save website %s: %v", websiteID, saveErr)
	}
}

func upsertNuvioBookingBackofficeRecord(
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
