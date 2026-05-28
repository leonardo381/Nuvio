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
	nuvioLeadsDashboardAlphaWebsiteID = "leadsalpha00001"
	nuvioLeadsDashboardBetaWebsiteID  = "leadsbeta000002"
	nuvioLeadsDashboardGammaWebsiteID = "leadsgamma00001"

	nuvioLeadsDashboardAlphaContactRecordID  = "leadctalpha0001"
	nuvioLeadsDashboardAlphaWhatsappRecordID = "leadwhalpha0001"
)

func TestNuvioLeadsDashboardEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin superuser receives scoped leads datasets",
			Method: http.MethodGet,
			URL:    "/api/nuvio/leads/dashboard?websiteId=" + nuvioLeadsDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioLeadsDashboardAlphaWebsiteID, nuvioLeadsDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioLeadsDashboardAlphaWebsiteID + `"`,
				`"datasets":{"contacts":[`,
				`"whatsapp":[`,
				`"capabilities":{"contacts":`,
				`"allowedStatus":["new","read","archived"]`,
				`"name":"Alpha Contact"`,
				`"name":"Alpha WhatsApp"`,
			},
			NotExpectedContent: []string{
				`"name":"Beta Contact"`,
				`"name":"Beta WhatsApp"`,
				`"settings":`,
				`"expand":`,
				`"apiUrl"`,
			},
		},
		{
			Name:   "admin receives empty dataset keys when website has no records",
			Method: http.MethodGet,
			URL:    "/api/nuvio/leads/dashboard?websiteId=" + nuvioLeadsDashboardGammaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioLeadsDashboardAlphaWebsiteID, nuvioLeadsDashboardBetaWebsiteID, nuvioLeadsDashboardGammaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"websiteId":"` + nuvioLeadsDashboardGammaWebsiteID + `"`,
				`"contacts":[]`,
				`"whatsapp":[]`,
			},
		},
		{
			Name:   "client superuser receives only assigned website data",
			Method: http.MethodGet,
			URL:    "/api/nuvio/leads/dashboard?websiteId=" + nuvioLeadsDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioLeadsDashboardAlphaWebsiteID + `"`,
				`"name":"Alpha Contact"`,
			},
			NotExpectedContent: []string{
				`"name":"Beta Contact"`,
				`"name":"Beta WhatsApp"`,
			},
		},
		{
			Name:   "client superuser denied for unassigned website",
			Method: http.MethodGet,
			URL:    "/api/nuvio/leads/dashboard?websiteId=" + nuvioLeadsDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioLeadsDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client superuser with no website access denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/leads/dashboard?websiteId=" + nuvioLeadsDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "missing role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/leads/dashboard?websiteId=" + nuvioLeadsDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/leads/dashboard?websiteId=" + nuvioLeadsDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/leads/dashboard?websiteId=" + nuvioLeadsDashboardAlphaWebsiteID,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioLeadsWriteEndpoints(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin can update contact status",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"read"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"status":"read"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "read")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "notes", "Alpha notes")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "email", "alpha.contact@example.test")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "message", "Alpha contact message")
			},
		},
		{
			Name:   "client with assigned website can update contact status",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"archived"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"status":"archived"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "archived")
			},
		},
		{
			Name:   "client with unassigned website cannot update contact status",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"read"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioLeadsDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "new")
			},
		},
		{
			Name:   "invalid contact status rejected",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"pending"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "new")
			},
		},
		{
			Name:   "admin can update whatsapp status",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/whatsapp/" + nuvioLeadsDashboardAlphaWhatsappRecordID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"read"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"status":"read"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				whatsapp := mustFindNuvioLeadsDashboardRecord(t, app, nuvioWhatsappCollectionID, nuvioLeadsDashboardAlphaWhatsappRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "status", "read")
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "notes", "Alpha WhatsApp notes")
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "email", "alpha.whatsapp@example.test")
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "message", "Alpha WhatsApp message")
			},
		},
		{
			Name:   "client with unassigned website cannot update whatsapp status",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/whatsapp/" + nuvioLeadsDashboardAlphaWhatsappRecordID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"archived"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioLeadsDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				whatsapp := mustFindNuvioLeadsDashboardRecord(t, app, nuvioWhatsappCollectionID, nuvioLeadsDashboardAlphaWhatsappRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "status", "new")
			},
		},
		{
			Name:   "admin can update contact follow-up",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"notes":"Updated Alpha notes","lastContactedAt":"2026-06-01T09:15:00Z"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"notes":"Updated Alpha notes"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "notes", "Updated Alpha notes")
				assertNuvioLeadsDashboardRecordContains(t, contact, "lastContactedAt", "2026-06-01")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "new")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "email", "alpha.contact@example.test")
			},
		},
		{
			Name:   "client with assigned website can update contact follow-up",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"notes":"Client follow-up note"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"notes":"Client follow-up note"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "notes", "Client follow-up note")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "new")
			},
		},
		{
			Name:   "client with unassigned website cannot update contact follow-up",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"notes":"Blocked update"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioLeadsDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "notes", "Alpha notes")
			},
		},
		{
			Name:   "invalid contact follow-up datetime rejected",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"lastContactedAt":"invalid-date"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "new")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "notes", "Alpha notes")
			},
		},
		{
			Name:   "contact follow-up endpoint rejects arbitrary fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"notes":"Should fail","status":"archived"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				contact := mustFindNuvioLeadsDashboardRecord(t, app, nuvioContactsCollectionID, nuvioLeadsDashboardAlphaContactRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "status", "new")
				assertNuvioLeadsDashboardRecordFieldString(t, contact, "notes", "Alpha notes")
			},
		},
		{
			Name:   "admin can update whatsapp follow-up",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/whatsapp/" + nuvioLeadsDashboardAlphaWhatsappRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"notes":"Updated WhatsApp notes","lastContactedAt":"2026-06-02T15:45:00Z"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"notes":"Updated WhatsApp notes"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				whatsapp := mustFindNuvioLeadsDashboardRecord(t, app, nuvioWhatsappCollectionID, nuvioLeadsDashboardAlphaWhatsappRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "notes", "Updated WhatsApp notes")
				assertNuvioLeadsDashboardRecordContains(t, whatsapp, "lastContactedAt", "2026-06-02")
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "status", "new")
			},
		},
		{
			Name:   "client with unassigned website cannot update whatsapp follow-up",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/whatsapp/" + nuvioLeadsDashboardAlphaWhatsappRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"notes":"Blocked WhatsApp update"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioLeadsDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				whatsapp := mustFindNuvioLeadsDashboardRecord(t, app, nuvioWhatsappCollectionID, nuvioLeadsDashboardAlphaWhatsappRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "notes", "Alpha WhatsApp notes")
			},
		},
		{
			Name:   "invalid whatsapp follow-up datetime rejected",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/whatsapp/" + nuvioLeadsDashboardAlphaWhatsappRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"lastContactedAt":"invalid-date"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, res *http.Response) {
				whatsapp := mustFindNuvioLeadsDashboardRecord(t, app, nuvioWhatsappCollectionID, nuvioLeadsDashboardAlphaWhatsappRecordID)
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "status", "new")
				assertNuvioLeadsDashboardRecordFieldString(t, whatsapp, "notes", "Alpha WhatsApp notes")
			},
		},
		{
			Name:   "missing role superuser denied for write endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/status",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"status":"read"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role superuser denied for write endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/whatsapp/" + nuvioLeadsDashboardAlphaWhatsappRecordID + "/follow-up",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"notes":"No permission"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioLeadsDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied for write endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/leads/contacts/" + nuvioLeadsDashboardAlphaContactRecordID + "/status",
			Body:   strings.NewReader(`{"status":"read"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioLeadsDashboardRoute(t, app, e)
				seedNuvioLeadsDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioLeadsDashboardRoute(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioLeadsDashboardRoutes(e)
}

func seedNuvioLeadsDashboardData(t testing.TB, app *tests.TestApp) {
	t.Helper()

	websitesCollection := ensureNuvioBackofficeWebsitesCollection(t, app)
	contactsCollection := ensureNuvioLeadsDashboardContactsCollection(t, app)
	whatsappCollection := ensureNuvioLeadsDashboardWhatsappCollection(t, app)

	upsertNuvioReportsDashboardWebsite(t, app, websitesCollection, nuvioLeadsDashboardAlphaWebsiteID, "alpha-leads", "Alpha Leads")
	upsertNuvioReportsDashboardWebsite(t, app, websitesCollection, nuvioLeadsDashboardBetaWebsiteID, "beta-leads", "Beta Leads")
	upsertNuvioReportsDashboardWebsite(t, app, websitesCollection, nuvioLeadsDashboardGammaWebsiteID, "gamma-leads", "Gamma Leads")

	upsertNuvioLeadsDashboardRecord(t, app, contactsCollection, "leadctalpha0001", map[string]any{
		"website":         nuvioLeadsDashboardAlphaWebsiteID,
		"channel":         "contact",
		"status":          "new",
		"name":            "Alpha Contact",
		"email":           "alpha.contact@example.test",
		"phone":           "+111111",
		"subject":         "Alpha contact subject",
		"message":         "Alpha contact message",
		"page":            "/alpha-contact",
		"source":          "hero-form",
		"notes":           "Alpha notes",
		"lastContactedAt": "2026-05-28T10:00:00Z",
	})
	upsertNuvioLeadsDashboardRecord(t, app, contactsCollection, "leadctbeta00002", map[string]any{
		"website":         nuvioLeadsDashboardBetaWebsiteID,
		"channel":         "contact",
		"status":          "new",
		"name":            "Beta Contact",
		"email":           "beta.contact@example.test",
		"phone":           "+222222",
		"subject":         "Beta contact subject",
		"message":         "Beta contact message",
		"page":            "/beta-contact",
		"source":          "footer-form",
		"notes":           "Beta notes",
		"lastContactedAt": "2026-05-27T10:00:00Z",
	})

	upsertNuvioLeadsDashboardRecord(t, app, whatsappCollection, "leadwhalpha0001", map[string]any{
		"website":         nuvioLeadsDashboardAlphaWebsiteID,
		"status":          "new",
		"source":          "floating_button",
		"page":            "/alpha-whatsapp",
		"name":            "Alpha WhatsApp",
		"email":           "alpha.whatsapp@example.test",
		"phone":           "+333333",
		"message":         "Alpha WhatsApp message",
		"defaultMessage":  "Hello from Alpha",
		"notes":           "Alpha WhatsApp notes",
		"lastContactedAt": "2026-05-26T10:00:00Z",
	})
	upsertNuvioLeadsDashboardRecord(t, app, whatsappCollection, "leadwhbeta00002", map[string]any{
		"website":         nuvioLeadsDashboardBetaWebsiteID,
		"status":          "new",
		"source":          "header-cta",
		"page":            "/beta-whatsapp",
		"name":            "Beta WhatsApp",
		"email":           "beta.whatsapp@example.test",
		"phone":           "+444444",
		"message":         "Beta WhatsApp message",
		"defaultMessage":  "Hello from Beta",
		"notes":           "Beta WhatsApp notes",
		"lastContactedAt": "2026-05-25T10:00:00Z",
	})
}

func ensureNuvioLeadsDashboardContactsCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioLeadsDashboardCollection(t, app, "Contacts", nuvioContactsCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "channel"},
		&core.SelectField{Name: "status", Values: []string{"new", "read", "archived"}},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "email"},
		&core.TextField{Name: "phone"},
		&core.TextField{Name: "subject"},
		&core.TextField{Name: "message"},
		&core.TextField{Name: "page"},
		&core.TextField{Name: "source"},
		&core.TextField{Name: "notes"},
		&core.DateField{Name: "lastContactedAt"},
	})
}

func ensureNuvioLeadsDashboardWhatsappCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioLeadsDashboardCollection(t, app, "WhatsAppInteractions", nuvioWhatsappCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.SelectField{Name: "status", Values: []string{"new", "read", "archived"}},
		&core.TextField{Name: "source"},
		&core.TextField{Name: "page"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "email"},
		&core.TextField{Name: "phone"},
		&core.TextField{Name: "message"},
		&core.TextField{Name: "defaultMessage"},
		&core.TextField{Name: "notes"},
		&core.DateField{Name: "lastContactedAt"},
	})
}

func ensureNuvioLeadsDashboardCollection(
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

func upsertNuvioLeadsDashboardRecord(
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

func mustFindNuvioLeadsDashboardRecord(
	t testing.TB,
	app *tests.TestApp,
	collectionID string,
	recordID string,
) *core.Record {
	t.Helper()

	record, err := app.FindRecordById(collectionID, recordID)
	if err != nil {
		t.Fatalf("failed to load %s record %s: %v", collectionID, recordID, err)
	}

	return record
}

func assertNuvioLeadsDashboardRecordFieldString(
	t testing.TB,
	record *core.Record,
	fieldName string,
	expected string,
) {
	t.Helper()

	actual := strings.TrimSpace(record.GetString(fieldName))
	if actual != expected {
		t.Fatalf("unexpected %s value: got %q want %q", fieldName, actual, expected)
	}
}

func assertNuvioLeadsDashboardRecordContains(
	t testing.TB,
	record *core.Record,
	fieldName string,
	contains string,
) {
	t.Helper()

	value := strings.TrimSpace(record.GetString(fieldName))
	if !strings.Contains(value, contains) {
		t.Fatalf("expected %s to contain %q, got %q", fieldName, contains, value)
	}
}
