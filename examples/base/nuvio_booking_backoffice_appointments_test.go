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

const (
	nuvioBookingBackofficeAlphaServiceID     = "svcalpha0000001"
	nuvioBookingBackofficeBetaServiceID      = "svcbeta00000002"
	nuvioBookingBackofficeAlphaAppointmentID = "aptalpha0000001"
	nuvioBookingBackofficeFutureDateA        = "2099-01-05"
	nuvioBookingBackofficeFutureDateB        = "2099-01-12"
)

func TestNuvioBookingBackofficeAppointmentEndpoints(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin can create booking appointment",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"serviceId":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"date":"` + nuvioBookingBackofficeFutureDateA + `",
				"time":"09:00",
				"name":"Backoffice Alpha Visitor",
				"email":"alpha.backoffice@example.test",
				"phone":"+351910000001",
				"notes":"Manual note",
				"internalNotes":"Internal note",
				"status":"confirmed"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				extendNuvioBookingDashboardWebsiteWindowForTest(t, app, nuvioBookingDashboardAlphaWebsiteID)
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
				`"name":"Backoffice Alpha Visitor"`,
				`"status":"confirmed"`,
			},
			NotExpectedContent: []string{
				`"manageToken"`,
				`"providerPayload"`,
				`"icsPayload"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				records, err := app.FindRecordsByFilter(
					nuvioAppointmentsCollectionID,
					"website={:website} && email={:email}",
					"",
					5,
					0,
					dbx.Params{
						"website": nuvioBookingDashboardAlphaWebsiteID,
						"email":   "alpha.backoffice@example.test",
					},
				)
				if err != nil {
					t.Fatalf("failed to find created appointment: %v", err)
				}
				if len(records) == 0 {
					t.Fatalf("expected created appointment record")
				}
			},
		},
		{
			Name:   "client assigned can create booking appointment",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"service":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"date":"` + nuvioBookingBackofficeFutureDateA + `",
				"time":"09:45",
				"name":"Client Alpha Visitor",
				"email":"client.alpha@example.test",
				"phone":"+351910000002",
				"notes":"Client note",
				"status":"pending"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				extendNuvioBookingDashboardWebsiteWindowForTest(t, app, nuvioBookingDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"name":"Client Alpha Visitor"`,
			},
		},
		{
			Name:   "client unassigned cannot create appointment",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"serviceId":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"date":"` + nuvioBookingBackofficeFutureDateA + `",
				"time":"09:00",
				"name":"Blocked Visitor",
				"email":"blocked@example.test"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				extendNuvioBookingDashboardWebsiteWindowForTest(t, app, nuvioBookingDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "create validates service belongs to website",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioBookingDashboardAlphaWebsiteID + `",
				"serviceId":"` + nuvioBookingBackofficeBetaServiceID + `",
				"date":"` + nuvioBookingBackofficeFutureDateA + `",
				"time":"09:00",
				"name":"Wrong Service",
				"email":"wrong.service@example.test"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				extendNuvioBookingDashboardWebsiteWindowForTest(t, app, nuvioBookingDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioBookingDashboardAlphaWebsiteID, nuvioBookingDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus: 404,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "admin can update appointment status",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"cancelled"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
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
				`"status":"cancelled"`,
			},
			NotExpectedContent: []string{
				`"manageToken"`,
				`"providerPayload"`,
				`"icsPayload"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeAppointmentRecord(t, app, nuvioBookingBackofficeAlphaAppointmentID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "status", "cancelled")
			},
		},
		{
			Name:   "client assigned can update appointment status",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"confirmed"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"status":"confirmed"`,
			},
		},
		{
			Name:   "client assigned can request status confirmation email",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"confirmed","sendEmail":true}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
					t,
					app,
					"Appointments",
					nuvioAppointmentsCollectionID,
					[]core.Field{
						&core.TextField{Name: "website"},
					},
				)
				upsertNuvioBookingBackofficeAppointmentRecord(t, app, appointmentsCollection, nuvioBookingBackofficeAlphaAppointmentID, map[string]any{
					"status":      "pending",
					"confirmedAt": "",
					"email":       "",
				})
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"status":"confirmed"`,
				`"warning":"Appointment confirmed, but customer email is missing."`,
			},
		},
		{
			Name:   "client unassigned cannot update appointment status",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"confirmed"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "status endpoint rejects invalid status",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"done"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "status endpoint rejects arbitrary field patching",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"confirmed","name":"Hacker"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeAppointmentRecord(t, app, nuvioBookingBackofficeAlphaAppointmentID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "name", "Alpha Visitor")
				assertNuvioBookingBackofficeRecordFieldString(t, record, "status", "confirmed")
			},
		},
		{
			Name:   "status endpoint accepts sendEmail and warns when email is missing",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"confirmed","sendEmail":true}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
					t,
					app,
					"Appointments",
					nuvioAppointmentsCollectionID,
					[]core.Field{
						&core.TextField{Name: "website"},
					},
				)
				upsertNuvioBookingBackofficeAppointmentRecord(t, app, appointmentsCollection, nuvioBookingBackofficeAlphaAppointmentID, map[string]any{
					"status":      "pending",
					"confirmedAt": "",
					"email":       "",
				})
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"status":"confirmed"`,
				`"warning":"Appointment confirmed, but customer email is missing."`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeAppointmentRecord(t, app, nuvioBookingBackofficeAlphaAppointmentID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "status", "confirmed")
			},
		},
		{
			Name:   "admin can trigger calendar action with sendEmail",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/calendar",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"sendEmail":true}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
					t,
					app,
					"Appointments",
					nuvioAppointmentsCollectionID,
					[]core.Field{
						&core.TextField{Name: "website"},
					},
				)
				upsertNuvioBookingBackofficeAppointmentRecord(t, app, appointmentsCollection, nuvioBookingBackofficeAlphaAppointmentID, map[string]any{
					"status": "confirmed",
					"email":  "",
				})
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"emailSent":false`,
				`"warning":"Appointment confirmed, but customer email is missing."`,
			},
		},
		{
			Name:   "client assigned can trigger calendar action with sendEmail",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/calendar",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"sendEmail":true}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
					t,
					app,
					"Appointments",
					nuvioAppointmentsCollectionID,
					[]core.Field{
						&core.TextField{Name: "website"},
					},
				)
				upsertNuvioBookingBackofficeAppointmentRecord(t, app, appointmentsCollection, nuvioBookingBackofficeAlphaAppointmentID, map[string]any{
					"status": "confirmed",
					"email":  "",
				})
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"emailSent":false`,
				`"warning":"Appointment confirmed, but customer email is missing."`,
			},
		},
		{
			Name:   "calendar action defaults sendEmail to true when omitted",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/calendar",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
					t,
					app,
					"Appointments",
					nuvioAppointmentsCollectionID,
					[]core.Field{
						&core.TextField{Name: "website"},
					},
				)
				upsertNuvioBookingBackofficeAppointmentRecord(t, app, appointmentsCollection, nuvioBookingBackofficeAlphaAppointmentID, map[string]any{
					"status": "confirmed",
					"email":  "",
				})
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"emailSent":false`,
				`"warning":"Appointment confirmed, but customer email is missing."`,
			},
		},
		{
			Name:   "calendar action skips email when sendEmail is false",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/calendar",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"sendEmail":false}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
					t,
					app,
					"Appointments",
					nuvioAppointmentsCollectionID,
					[]core.Field{
						&core.TextField{Name: "website"},
					},
				)
				upsertNuvioBookingBackofficeAppointmentRecord(t, app, appointmentsCollection, nuvioBookingBackofficeAlphaAppointmentID, map[string]any{
					"status": "confirmed",
					"email":  "",
				})
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"emailSent":false`,
			},
			NotExpectedContent: []string{
				`"warning":"Appointment confirmed, but customer email is missing."`,
			},
		},
		{
			Name:   "client unassigned cannot trigger calendar action",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/calendar",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"sendEmail":true}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "admin can reschedule appointment",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/reschedule",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"serviceId":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"date":"` + nuvioBookingBackofficeFutureDateB + `",
				"time":"09:45"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				extendNuvioBookingDashboardWebsiteWindowForTest(t, app, nuvioBookingDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"date":"` + nuvioBookingBackofficeFutureDateB + `"`,
				`"time":"09:45"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeAppointmentRecord(t, app, nuvioBookingBackofficeAlphaAppointmentID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "date", nuvioBookingBackofficeFutureDateB)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "time", "09:45")
			},
		},
		{
			Name:   "client assigned can reschedule appointment",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/reschedule",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"service":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"date":"` + nuvioBookingBackofficeFutureDateB + `",
				"time":"10:30"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				extendNuvioBookingDashboardWebsiteWindowForTest(t, app, nuvioBookingDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"date":"` + nuvioBookingBackofficeFutureDateB + `"`,
				`"time":"10:30"`,
			},
		},
		{
			Name:   "reschedule defaults sendEmail to true when omitted",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/reschedule",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"serviceId":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"date":"` + nuvioBookingBackofficeFutureDateB + `",
				"time":"11:15"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				extendNuvioBookingDashboardWebsiteWindowForTest(t, app, nuvioBookingDashboardAlphaWebsiteID)
				appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
					t,
					app,
					"Appointments",
					nuvioAppointmentsCollectionID,
					[]core.Field{
						&core.TextField{Name: "website"},
					},
				)
				upsertNuvioBookingBackofficeAppointmentRecord(t, app, appointmentsCollection, nuvioBookingBackofficeAlphaAppointmentID, map[string]any{
					"email": "",
				})
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"date":"` + nuvioBookingBackofficeFutureDateB + `"`,
				`"time":"11:15"`,
				`"warning":"Appointment rescheduled, but customer email is missing."`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeAppointmentRecord(t, app, nuvioBookingBackofficeAlphaAppointmentID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "date", nuvioBookingBackofficeFutureDateB)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "time", "11:15")
			},
		},
		{
			Name:   "reschedule validates date and time",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/reschedule",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"serviceId":"` + nuvioBookingBackofficeAlphaServiceID + `",
				"date":"invalid-date",
				"time":"25:00"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "reschedule validates service website ownership",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/reschedule",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"serviceId":"` + nuvioBookingBackofficeBetaServiceID + `",
				"date":"` + nuvioBookingBackofficeFutureDateB + `",
				"time":"09:00"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				extendNuvioBookingDashboardWebsiteWindowForTest(t, app, nuvioBookingDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID, nuvioBookingDashboardBetaWebsiteID})
			},
			ExpectedStatus: 404,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "internal notes endpoint updates only internal notes",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/internal-notes",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"internalNotes":"Updated internal notes"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"internalNotes":"Updated internal notes"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeAppointmentRecord(t, app, nuvioBookingBackofficeAlphaAppointmentID)
				assertNuvioBookingBackofficeRecordFieldString(t, record, "internalNotes", "Updated internal notes")
				assertNuvioBookingBackofficeRecordFieldString(t, record, "status", "confirmed")
			},
		},
		{
			Name:   "archive endpoint sets archivedAt",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/archive",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"archived":true}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"archivedAt":"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeAppointmentRecord(t, app, nuvioBookingBackofficeAlphaAppointmentID)
				archivedAt := strings.TrimSpace(record.GetString("archivedAt"))
				if archivedAt == "" {
					t.Fatalf("expected archivedAt to be set")
				}
			},
		},
		{
			Name:   "archive endpoint clears archivedAt",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/archive",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"archived":false}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				appointmentsCollection := ensureNuvioBookingBackofficeDashboardCollection(
					t,
					app,
					"Appointments",
					nuvioAppointmentsCollectionID,
					[]core.Field{
						&core.TextField{Name: "website"},
					},
				)
				upsertNuvioBookingBackofficeAppointmentRecord(t, app, appointmentsCollection, nuvioBookingBackofficeAlphaAppointmentID, map[string]any{
					"archivedAt": "2026-06-05T10:00:00Z",
				})
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioBookingBackofficeAppointmentRecord(t, app, nuvioBookingBackofficeAlphaAppointmentID)
				if strings.TrimSpace(record.GetString("archivedAt")) != "" {
					t.Fatalf("expected archivedAt to be cleared")
				}
			},
		},
		{
			Name:   "missing role denied for write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"pending"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "unknown role denied for write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"pending"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioBookingDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "unauthenticated denied for write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/booking/backoffice/appointments/" + nuvioBookingBackofficeAlphaAppointmentID + "/status",
			Body:   strings.NewReader(`{"status":"pending"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBookingBackofficeAppointmentRoutes(t, app, e)
				seedNuvioBookingBackofficeDashboardData(t, app)
			},
			ExpectedStatus: 401,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioBookingBackofficeAppointmentRoutes(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioBookingRoutes(e)
}

func extendNuvioBookingDashboardWebsiteWindowForTest(t testing.TB, app *tests.TestApp, websiteID string) {
	t.Helper()

	record, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		t.Fatalf("failed to load website %s: %v", websiteID, err)
	}

	settings := parseNuvioSettingsObject(record.Get("settings"))
	bookingSettings, ok := toStringAnyMap(settings["booking"])
	if !ok {
		bookingSettings = map[string]any{}
	}
	rulesSettings, ok := toStringAnyMap(bookingSettings["rules"])
	if !ok {
		rulesSettings = map[string]any{}
	}
	rulesSettings["bookingWindowDays"] = 50000
	rulesSettings["minNoticeHours"] = 0
	bookingSettings["rules"] = rulesSettings
	settings["booking"] = bookingSettings
	record.Set("settings", settings)

	if err := app.Save(record); err != nil {
		t.Fatalf("failed to update booking rules for website %s: %v", websiteID, err)
	}
}

func mustFindNuvioBookingBackofficeAppointmentRecord(
	t testing.TB,
	app *tests.TestApp,
	recordID string,
) *core.Record {
	t.Helper()

	record, err := app.FindRecordById(nuvioAppointmentsCollectionID, recordID)
	if err != nil {
		t.Fatalf("failed to find appointment %s: %v", recordID, err)
	}
	return record
}

func assertNuvioBookingBackofficeRecordFieldString(
	t testing.TB,
	record *core.Record,
	fieldName string,
	expected string,
) {
	t.Helper()

	value := strings.TrimSpace(record.GetString(fieldName))
	if value != expected {
		t.Fatalf("unexpected %s value: got %q want %q", fieldName, value, expected)
	}
}

func upsertNuvioBookingBackofficeAppointmentRecord(
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

	for fieldName, value := range values {
		record.Set(fieldName, value)
	}

	if saveErr := app.Save(record); saveErr != nil {
		t.Fatalf("failed to save %s record %s: %v", collection.Name, recordID, saveErr)
	}
}
