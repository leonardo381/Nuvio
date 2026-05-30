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
	nuvioBookingPublicWebsiteID       = "pubbooksite0001"
	nuvioBookingPublicWebsiteSlug     = "public-booking-site"
	nuvioBookingPublicSecondarySiteID = "pubbooksite0002"
	nuvioBookingPublicServiceID       = "pubbooksvc00001"
	nuvioBookingPublicForeignService  = "pubbooksvc00002"
)

func TestNuvioBookingPublicServicesEndpoint(t *testing.T) {
	t.Parallel()

	slotDate := nextNuvioBookingPublicDateForWeekday(time.Monday, 8)

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid services request with websiteSlug succeeds",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/services?websiteSlug=" + nuvioBookingPublicWebsiteSlug,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"websiteId":"` + nuvioBookingPublicWebsiteID + `"`,
				`"id":"` + nuvioBookingPublicServiceID + `"`,
				`"name":"Public Booking Service"`,
			},
		},
		{
			Name:   "missing website context fails safely",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/services",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Missing website context."`},
		},
		{
			Name:   "disabled booking feature fails safely",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/services?websiteId=" + nuvioBookingPublicWebsiteID,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, false, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Booking feature is unavailable for this website."`},
		},
		{
			Name:   "unknown services query key is rejected",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/services?websiteId=" + nuvioBookingPublicWebsiteID + "&debug=true",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Field \"debug\" is not allowed in this endpoint."`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioBookingPublicSlotsEndpoint(t *testing.T) {
	t.Parallel()

	slotDate := nextNuvioBookingPublicDateForWeekday(time.Monday, 8)
	outsideWindowDate := nextNuvioBookingPublicDateForWeekday(time.Monday, 60)

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid slots request succeeds",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/slots?websiteSlug=" + nuvioBookingPublicWebsiteSlug + "&serviceId=" + nuvioBookingPublicServiceID + "&date=" + slotDate,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"serviceId":"` + nuvioBookingPublicServiceID + `"`,
				`"date":"` + slotDate + `"`,
				`"slots":[`,
			},
		},
		{
			Name:   "missing serviceId fails safely",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/slots?websiteId=" + nuvioBookingPublicWebsiteID + "&date=" + slotDate,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Missing serviceId."`},
		},
		{
			Name:   "invalid date format fails safely",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/slots?websiteId=" + nuvioBookingPublicWebsiteID + "&serviceId=" + nuvioBookingPublicServiceID + "&date=01-06-2026",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Date must use YYYY-MM-DD format."`},
		},
		{
			Name:   "date outside booking window fails safely",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/slots?websiteId=" + nuvioBookingPublicWebsiteID + "&serviceId=" + nuvioBookingPublicServiceID + "&date=" + outsideWindowDate,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"This date is outside the booking window."`},
		},
		{
			Name:   "service from another website fails safely",
			Method: http.MethodGet,
			URL:    "/api/nuvio/booking/slots?websiteId=" + nuvioBookingPublicWebsiteID + "&serviceId=" + nuvioBookingPublicForeignService + "&date=" + slotDate,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"message":"Service not found."`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioBookingPublicAppointmentsEndpoint(t *testing.T) {
	t.Parallel()

	slotDate := nextNuvioBookingPublicDateForWeekday(time.Monday, 8)
	tooLongName := strings.Repeat("n", nuvioBookingPublicNameMaxLen+1)

	validPayload := `{
		"websiteSlug":"` + nuvioBookingPublicWebsiteSlug + `",
		"serviceId":"` + nuvioBookingPublicServiceID + `",
		"date":"` + slotDate + `",
		"time":"09:00",
		"name":"Alice Visitor",
		"email":"alice.booking@example.test",
		"phone":"+351900000001",
		"notes":"Please ring the bell."
	}`

	scenarios := []tests.ApiScenario{
		{
			Name:   "valid appointment submit succeeds",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/appointments",
			Body:   strings.NewReader(validPayload),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"ok":true`,
				`"appointmentId":"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				appointment := mustFindNuvioBookingPublicAppointmentByEmail(t, app, "alice.booking@example.test")
				assertNuvioLeadsDashboardRecordFieldString(t, appointment, "website", nuvioBookingPublicWebsiteID)
				assertNuvioLeadsDashboardRecordFieldString(t, appointment, "service", nuvioBookingPublicServiceID)
				assertNuvioLeadsDashboardRecordFieldString(t, appointment, "date", slotDate)
				assertNuvioLeadsDashboardRecordFieldString(t, appointment, "time", "09:00")

				contact := mustFindNuvioPublicContactRecordByEmail(t, app, "alice.booking@example.test")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "website", nuvioBookingPublicWebsiteID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "new")
			},
		},
		{
			Name:   "missing website context fails safely",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/appointments",
			Body: strings.NewReader(`{
				"serviceId":"` + nuvioBookingPublicServiceID + `",
				"date":"` + slotDate + `",
				"time":"09:00",
				"name":"Alice Visitor",
				"email":"alice.booking@example.test"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Missing website context."`},
		},
		{
			Name:   "invalid email fails safely",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/appointments",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingPublicWebsiteID + `",
				"serviceId":"` + nuvioBookingPublicServiceID + `",
				"date":"` + slotDate + `",
				"time":"09:00",
				"name":"Alice Visitor",
				"email":"alice"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"A valid email is required."`},
		},
		{
			Name:   "overlong name fails safely",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/appointments",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingPublicWebsiteID + `",
				"serviceId":"` + nuvioBookingPublicServiceID + `",
				"date":"` + slotDate + `",
				"time":"09:00",
				"name":"` + tooLongName + `",
				"email":"alice.booking@example.test"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"message":"Name is too long. Maximum 160 characters."`},
		},
		{
			Name:   "unknown appointment field is rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/appointments",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingPublicWebsiteID + `",
				"serviceId":"` + nuvioBookingPublicServiceID + `",
				"date":"` + slotDate + `",
				"time":"09:00",
				"name":"Alice Visitor",
				"email":"alice.booking@example.test",
				"hacker":"1"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"Field \"hacker\" is not allowed in this endpoint."`},
		},
		{
			Name:   "slot is revalidated at submit time",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/appointments",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingPublicWebsiteID + `",
				"serviceId":"` + nuvioBookingPublicServiceID + `",
				"date":"` + slotDate + `",
				"time":"09:00",
				"name":"Bob Visitor",
				"email":"bob.booking@example.test"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
				seedNuvioBookingPublicAppointment(
					t,
					app,
					"pbkexistappt001",
					nuvioBookingPublicWebsiteID,
					nuvioBookingPublicServiceID,
					slotDate,
					"09:00",
					"pending",
					"existing@example.test",
				)
			},
			ExpectedStatus:  409,
			ExpectedContent: []string{`"error":"This time is no longer available."`},
		},
		{
			Name:   "duplicate recent submit is blocked",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/appointments",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingPublicWebsiteID + `",
				"serviceId":"` + nuvioBookingPublicServiceID + `",
				"date":"` + slotDate + `",
				"time":"09:00",
				"name":"Alice Visitor",
				"email":"alice.booking@example.test"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingPublicRoutes(t, app, e)
				seedNuvioBookingPublicData(t, app, slotDate, true, true)
				seedNuvioBookingPublicAppointment(
					t,
					app,
					"pbkexistappt002",
					nuvioBookingPublicWebsiteID,
					nuvioBookingPublicServiceID,
					slotDate,
					"09:00",
					"pending",
					"alice.booking@example.test",
				)
			},
			ExpectedStatus:  409,
			ExpectedContent: []string{`"error":"A booking request for this slot was already submitted recently."`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioBookingPublicRoutes(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioBookingRoutes(e)
}

func seedNuvioBookingPublicData(
	t testing.TB,
	app *tests.TestApp,
	slotDate string,
	bookingFeatureAvailable bool,
	bookingEnabled bool,
) {
	t.Helper()

	websitesCollection := ensureNuvioBackofficeWebsitesCollection(t, app)
	servicesCollection := ensureNuvioBookingBackofficeDashboardCollection(
		t,
		app,
		"BookingServices",
		nuvioBookingServicesCollectionID,
		[]core.Field{
			&core.TextField{Name: "website"},
			&core.TextField{Name: "name"},
			&core.TextField{Name: "description"},
			&core.NumberField{Name: "durationMinutes"},
			&core.NumberField{Name: "displayOrder"},
			&core.BoolField{Name: "active"},
		},
	)
	availabilityCollection := ensureNuvioBookingBackofficeDashboardCollection(
		t,
		app,
		"BookingAvailability",
		nuvioBookingAvailabilityCollectionID,
		[]core.Field{
			&core.TextField{Name: "website"},
			&core.TextField{Name: "service"},
			&core.TextField{Name: "dayOfWeek"},
			&core.TextField{Name: "startTime"},
			&core.TextField{Name: "endTime"},
			&core.BoolField{Name: "active"},
			&core.NumberField{Name: "capacity"},
		},
	)
	_ = ensureNuvioBookingBackofficeDashboardCollection(
		t,
		app,
		"BookingExceptions",
		nuvioBookingExceptionsCollectionID,
		[]core.Field{
			&core.TextField{Name: "website"},
			&core.TextField{Name: "service"},
			&core.TextField{Name: "date"},
			&core.TextField{Name: "type"},
			&core.TextField{Name: "startTime"},
			&core.TextField{Name: "endTime"},
			&core.BoolField{Name: "active"},
		},
	)
	_ = ensureNuvioBookingBackofficeDashboardCollection(
		t,
		app,
		"Appointments",
		nuvioAppointmentsCollectionID,
		[]core.Field{
			&core.TextField{Name: "website"},
			&core.TextField{Name: "service"},
			&core.TextField{Name: "name"},
			&core.TextField{Name: "email"},
			&core.TextField{Name: "phone"},
			&core.TextField{Name: "date"},
			&core.TextField{Name: "time"},
			&core.TextField{Name: "notes"},
			&core.TextField{Name: "status"},
			&core.TextField{Name: "serviceNameSnapshot"},
			&core.NumberField{Name: "serviceDurationMinutesSnapshot"},
			&core.TextField{Name: "serviceDescriptionSnapshot"},
		},
	)
	ensureNuvioLeadsDashboardContactsCollection(t, app)

	upsertNuvioBookingBackofficeWebsiteRecord(
		t,
		app,
		websitesCollection,
		nuvioBookingPublicWebsiteID,
		nuvioBookingPublicWebsiteSlug,
		"Public Booking Site",
		map[string]any{
			"featureFlags": map[string]any{
				"booking": bookingFeatureAvailable,
			},
			"booking": map[string]any{
				"enabled": bookingEnabled,
				"rules": map[string]any{
					"minNoticeHours":    0,
					"bookingWindowDays": 30,
					"bufferMinutes":     0,
				},
			},
		},
	)
	upsertNuvioBookingBackofficeWebsiteRecord(
		t,
		app,
		websitesCollection,
		nuvioBookingPublicSecondarySiteID,
		"public-booking-secondary",
		"Public Booking Secondary",
		map[string]any{
			"featureFlags": map[string]any{
				"booking": true,
			},
			"booking": map[string]any{
				"enabled": true,
				"rules": map[string]any{
					"minNoticeHours":    0,
					"bookingWindowDays": 30,
					"bufferMinutes":     0,
				},
			},
		},
	)

	upsertNuvioBookingBackofficeRecord(
		t,
		app,
		servicesCollection,
		nuvioBookingPublicServiceID,
		map[string]any{
			"website":         nuvioBookingPublicWebsiteID,
			"name":            "Public Booking Service",
			"description":     "Public booking service description",
			"durationMinutes": 30,
			"displayOrder":    1,
			"active":          true,
		},
	)
	upsertNuvioBookingBackofficeRecord(
		t,
		app,
		servicesCollection,
		nuvioBookingPublicForeignService,
		map[string]any{
			"website":         nuvioBookingPublicSecondarySiteID,
			"name":            "Foreign Booking Service",
			"description":     "Foreign booking service description",
			"durationMinutes": 30,
			"displayOrder":    2,
			"active":          true,
		},
	)

	dayOfWeek, err := dateToNuvioBookingDayOfWeek(slotDate)
	if err != nil {
		t.Fatalf("failed to derive booking day of week: %v", err)
	}

	upsertNuvioBookingBackofficeRecord(
		t,
		app,
		availabilityCollection,
		"pubbookavl00001",
		map[string]any{
			"website":   nuvioBookingPublicWebsiteID,
			"service":   nuvioBookingPublicServiceID,
			"dayOfWeek": dayOfWeek,
			"startTime": "09:00",
			"endTime":   "12:00",
			"active":    true,
			"capacity":  1,
		},
	)
}

func seedNuvioBookingPublicAppointment(
	t testing.TB,
	app *tests.TestApp,
	recordID string,
	websiteID string,
	serviceID string,
	dateValue string,
	timeValue string,
	status string,
	email string,
) {
	t.Helper()

	appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
		t,
		app,
		"Appointments",
		nuvioAppointmentsCollectionID,
		[]core.Field{
			&core.TextField{Name: "website"},
			&core.TextField{Name: "service"},
			&core.TextField{Name: "name"},
			&core.TextField{Name: "email"},
			&core.TextField{Name: "phone"},
			&core.TextField{Name: "date"},
			&core.TextField{Name: "time"},
			&core.TextField{Name: "notes"},
			&core.TextField{Name: "status"},
		},
	)

	upsertNuvioBookingBackofficeRecord(
		t,
		app,
		appointmentsCollection,
		recordID,
		map[string]any{
			"website": websiteID,
			"service": serviceID,
			"name":    "Existing Visitor",
			"email":   email,
			"phone":   "+351900000099",
			"date":    dateValue,
			"time":    timeValue,
			"notes":   "Existing booking",
			"status":  status,
		},
	)
}

func mustFindNuvioBookingPublicAppointmentByEmail(t testing.TB, app *tests.TestApp, email string) *core.Record {
	t.Helper()

	appointmentsCollection, err := app.FindCollectionByNameOrId(nuvioAppointmentsCollectionID)
	if err != nil {
		t.Fatalf("failed to load appointments collection: %v", err)
	}

	record, err := app.FindFirstRecordByFilter(
		appointmentsCollection,
		"email={:email}",
		dbx.Params{"email": email},
	)
	if err != nil {
		t.Fatalf("failed to load appointment by email: %v", err)
	}

	return record
}

func nextNuvioBookingPublicDateForWeekday(target time.Weekday, minDaysAhead int) string {
	location := getNuvioBookingLocation()
	start := time.Now().In(location)
	if minDaysAhead > 0 {
		start = start.AddDate(0, 0, minDaysAhead)
	}

	for offset := 0; offset < 31; offset++ {
		candidate := start.AddDate(0, 0, offset)
		if candidate.Weekday() == target {
			return candidate.Format("2006-01-02")
		}
	}

	return start.Format("2006-01-02")
}
