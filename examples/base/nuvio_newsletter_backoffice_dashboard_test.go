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
	nuvioNewsletterBackofficeAlphaWebsiteID = "nwsalpha0000001"
	nuvioNewsletterBackofficeBetaWebsiteID  = "nwsbeta00000002"
	nuvioNewsletterBackofficeGammaWebsiteID = "nwsgamma0000003"
	nuvioNewsletterBackofficeAlphaGroupID   = "nwsgrpalpha0001"
	nuvioNewsletterBackofficeBetaGroupID    = "nwsgrpbeta00001"
	nuvioNewsletterBackofficeAlphaSubID     = "nwvsubalpha0001"
	nuvioNewsletterBackofficeBetaSubID      = "nwvsubbeta00001"
	nuvioNewsletterBackofficeAlphaCampID    = "nwvcampalpha001"
	nuvioNewsletterBackofficeBetaCampID     = "nwvcampbeta0001"
)

func TestNuvioNewsletterBackofficeDashboardEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin superuser receives scoped newsletter datasets",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/backoffice/dashboard?websiteId=" + nuvioNewsletterBackofficeAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioNewsletterBackofficeAlphaWebsiteID, nuvioNewsletterBackofficeBetaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `"`,
				`"datasets":{"subscribers":[`,
				`"campaigns":[`,
				`"groups":[`,
				`"capabilities":{"subscribers":`,
				`"allowedStatus":["pending","active","unsubscribed"]`,
				`"allowedRecipientsType":["all","manual"]`,
				`"name":"Alpha Subscriber"`,
				`"subject":"Alpha Campaign"`,
				`"name":"Alpha Group"`,
			},
			NotExpectedContent: []string{
				`"name":"Beta Subscriber"`,
				`"subject":"Beta Campaign"`,
				`"name":"Beta Group"`,
				`"settings":`,
				`"expand":`,
				`"confirmationTokenHash"`,
				`"confirmationTokenExpiresAt"`,
				`"unsubscribeTokenHash"`,
				`"alpha-confirm-token-hash"`,
				`"alpha-unsubscribe-token-hash"`,
			},
		},
		{
			Name:   "admin receives empty dataset keys when website has no records",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/backoffice/dashboard?websiteId=" + nuvioNewsletterBackofficeGammaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{
						nuvioNewsletterBackofficeAlphaWebsiteID,
						nuvioNewsletterBackofficeBetaWebsiteID,
						nuvioNewsletterBackofficeGammaWebsiteID,
					},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"websiteId":"` + nuvioNewsletterBackofficeGammaWebsiteID + `"`,
				`"subscribers":[]`,
				`"campaigns":[]`,
				`"groups":[]`,
			},
		},
		{
			Name:   "client superuser receives only assigned website newsletter data",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/backoffice/dashboard?websiteId=" + nuvioNewsletterBackofficeAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioNewsletterBackofficeAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `"`,
				`"name":"Alpha Subscriber"`,
			},
			NotExpectedContent: []string{
				`"name":"Beta Subscriber"`,
				`"subject":"Beta Campaign"`,
				`"name":"Beta Group"`,
				`"confirmationTokenHash"`,
				`"confirmationTokenExpiresAt"`,
				`"unsubscribeTokenHash"`,
			},
		},
		{
			Name:   "client superuser denied for unassigned website",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/backoffice/dashboard?websiteId=" + nuvioNewsletterBackofficeAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioNewsletterBackofficeBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client superuser with no website access denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/backoffice/dashboard?websiteId=" + nuvioNewsletterBackofficeAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "missing role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/backoffice/dashboard?websiteId=" + nuvioNewsletterBackofficeAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioNewsletterBackofficeAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/backoffice/dashboard?websiteId=" + nuvioNewsletterBackofficeAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioNewsletterBackofficeAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/newsletter/backoffice/dashboard?websiteId=" + nuvioNewsletterBackofficeAlphaWebsiteID,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioNewsletterBackofficeWriteEndpoints(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin can create subscriber",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/subscribers",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"email":"new.alpha@example.test",
				"name":"New Alpha",
				"status":"active",
				"source":"manual_dashboard",
				"groups":["` + nuvioNewsletterBackofficeAlphaGroupID + `"]
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"subscriber":{"id":"`,
				`"email":"new.alpha@example.test"`,
				`"status":"active"`,
			},
			NotExpectedContent: []string{
				`"confirmationTokenHash"`,
				`"confirmationTokenExpiresAt"`,
				`"unsubscribeTokenHash"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := findNuvioNewsletterBackofficeSubscriberByWebsiteEmail(t, app, nuvioNewsletterBackofficeAlphaWebsiteID, "new.alpha@example.test")
				assertNuvioNewsletterBackofficeRecordField(t, record, "status", "active")
				assertNuvioNewsletterBackofficeRecordField(t, record, "name", "New Alpha")
			},
		},
		{
			Name:   "client assigned can create subscriber",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/subscribers",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"email":"client.alpha@example.test",
				"name":"Client Alpha",
				"status":"pending"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := findNuvioNewsletterBackofficeSubscriberByWebsiteEmail(t, app, nuvioNewsletterBackofficeAlphaWebsiteID, "client.alpha@example.test")
				assertNuvioNewsletterBackofficeRecordField(t, record, "status", "pending")
			},
		},
		{
			Name:   "client unassigned cannot create subscriber",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/subscribers",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"email":"blocked.alpha@example.test",
				"name":"Blocked Alpha",
				"status":"pending"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "client assigned can invite subscriber from leads flow",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/invite",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeBetaWebsiteID + `",
				"email":"beta-subscriber@example.test",
				"name":"Beta Subscriber",
				"source":"whatsapp"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"result":"already_active"`,
				`"status":"active"`,
			},
		},
		{
			Name:   "client unassigned cannot invite subscriber",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/invite",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeBetaWebsiteID + `",
				"email":"beta-subscriber@example.test",
				"name":"Beta Subscriber",
				"source":"contact"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "invalid subscriber email rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/subscribers",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"email":"invalid-email",
				"name":"Invalid Email"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "invalid subscriber status rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/subscribers",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"email":"status.invalid@example.test",
				"status":"invited"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "subscriber update allows only whitelisted fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/newsletter/backoffice/subscribers/" + nuvioNewsletterBackofficeAlphaSubID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"email":"updated.alpha@example.test",
				"name":"Updated Alpha",
				"status":"active",
				"source":"manual_dashboard",
				"groups":["` + nuvioNewsletterBackofficeAlphaGroupID + `"]
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioNewsletterBackofficeRecordByID(t, app, nuvioSubscribersCollectionID, nuvioNewsletterBackofficeAlphaSubID)
				assertNuvioNewsletterBackofficeRecordField(t, record, "email", "updated.alpha@example.test")
				assertNuvioNewsletterBackofficeRecordField(t, record, "status", "active")
				assertNuvioNewsletterBackofficeRecordField(t, record, "name", "Updated Alpha")
			},
		},
		{
			Name:   "subscriber update rejects lifecycle token fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/newsletter/backoffice/subscribers/" + nuvioNewsletterBackofficeAlphaSubID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"confirmationTokenHash":"attempted-overwrite"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				record := mustFindNuvioNewsletterBackofficeRecordByID(t, app, nuvioSubscribersCollectionID, nuvioNewsletterBackofficeAlphaSubID)
				assertNuvioNewsletterBackofficeRecordField(t, record, "confirmationTokenHash", "alpha-confirm-token-hash")
			},
		},
		{
			Name:   "subscriber delete enforces website access",
			Method: http.MethodDelete,
			URL:    "/api/nuvio/newsletter/backoffice/subscribers/" + nuvioNewsletterBackofficeAlphaSubID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				_ = mustFindNuvioNewsletterBackofficeRecordByID(t, app, nuvioSubscribersCollectionID, nuvioNewsletterBackofficeAlphaSubID)
			},
		},
		{
			Name:   "admin can create group",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/groups",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"name":"VIP Group"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
			},
		},
		{
			Name:   "group create unassigned denied",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/groups",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"name":"Blocked Group"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "group create empty name rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/groups",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"name":"  "
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "admin can create draft campaign",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/campaigns",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"subject":"Created Draft Campaign",
				"body":"<p>Created body</p>",
				"status":"draft",
				"recipientsType":"manual",
				"recipientsIds":["` + nuvioNewsletterBackofficeAlphaSubID + `"]
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"status":"draft"`,
			},
		},
		{
			Name:   "campaign create unassigned denied",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/campaigns",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"subject":"Blocked draft",
				"body":"<p>Blocked body</p>",
				"status":"draft",
				"recipientsType":"manual",
				"recipientsIds":["` + nuvioNewsletterBackofficeAlphaSubID + `"]
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "campaign update enforces website access",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/newsletter/backoffice/campaigns/" + nuvioNewsletterBackofficeAlphaCampID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"subject":"Blocked update"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "campaign update rejects arbitrary field",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/newsletter/backoffice/campaigns/" + nuvioNewsletterBackofficeAlphaCampID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"subject":"Safe subject","websiteId":"` + nuvioNewsletterBackofficeBetaWebsiteID + `"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "campaign delete enforces website access",
			Method: http.MethodDelete,
			URL:    "/api/nuvio/newsletter/backoffice/campaigns/" + nuvioNewsletterBackofficeAlphaCampID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "campaign duplicate creates new draft and clears sentAt",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/campaigns/" + nuvioNewsletterBackofficeBetaCampID + "/duplicate",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"status":"draft"`,
				`"subject":"Beta Campaign"`,
			},
			NotExpectedContent: []string{
				`"confirmationTokenHash"`,
				`"confirmationTokenExpiresAt"`,
				`"unsubscribeTokenHash"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				campaignsCollection := ensureNuvioNewsletterBackofficeCampaignsCollection(t, app)
				records, err := app.FindRecordsByFilter(
					campaignsCollection,
					`website={:website}`,
					"",
					10,
					0,
					dbx.Params{"website": nuvioNewsletterBackofficeBetaWebsiteID},
				)
				if err != nil {
					t.Fatalf("failed to list beta campaigns: %v", err)
				}
				if len(records) < 2 {
					t.Fatalf("expected duplicate campaign for beta website")
				}

				foundDraftDuplicate := false
				for _, record := range records {
					if record.Id == nuvioNewsletterBackofficeBetaCampID {
						continue
					}
					if strings.TrimSpace(record.GetString("subject")) == "Beta Campaign" && normalizeNewsletterStatusForTest(record.GetString("status")) == "draft" {
						if strings.TrimSpace(record.GetString("sentAt")) != "" {
							t.Fatalf("expected duplicated campaign sentAt to be empty")
						}
						foundDraftDuplicate = true
						break
					}
				}
				if !foundDraftDuplicate {
					t.Fatalf("failed to find duplicated beta draft campaign")
				}
			},
		},
		{
			Name:   "campaign recipients must belong to same website",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/campaigns",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"subject":"Cross Website Recipients",
				"body":"<p>Body</p>",
				"status":"draft",
				"recipientsType":"manual",
				"recipientsIds":["` + nuvioNewsletterBackofficeBetaSubID + `"]
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
					nuvioNewsletterBackofficeBetaWebsiteID,
				})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "unauthenticated denied for write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/groups",
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"name":"No Auth Group"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
			},
			ExpectedStatus: 401,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "missing role denied for write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/groups",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"name":"Missing Role Group"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "unknown role denied for write endpoint",
			Method: http.MethodPost,
			URL:    "/api/nuvio/newsletter/backoffice/campaigns",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"websiteId":"` + nuvioNewsletterBackofficeAlphaWebsiteID + `",
				"subject":"Unknown role",
				"body":"<p>Body</p>",
				"status":"draft"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioNewsletterBackofficeDashboardRoute(t, app, e)
				seedNuvioNewsletterBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{
					nuvioNewsletterBackofficeAlphaWebsiteID,
				})
			},
			ExpectedStatus: 403,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioNewsletterBackofficeDashboardRoute(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioNewsletterRoutes(e)
}

func seedNuvioNewsletterBackofficeDashboardData(t testing.TB, app *tests.TestApp) {
	t.Helper()

	websitesCollection := ensureNuvioBackofficeWebsitesCollection(t, app)
	subscribersCollection := ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
	campaignsCollection := ensureNuvioNewsletterBackofficeCampaignsCollection(t, app)
	groupsCollection := ensureNuvioNewsletterBackofficeGroupsCollection(t, app)

	upsertNuvioNewsletterBackofficeWebsite(t, app, websitesCollection, nuvioNewsletterBackofficeAlphaWebsiteID, "newsletter-alpha", "Newsletter Alpha")
	upsertNuvioNewsletterBackofficeWebsite(t, app, websitesCollection, nuvioNewsletterBackofficeBetaWebsiteID, "newsletter-beta", "Newsletter Beta")
	upsertNuvioNewsletterBackofficeWebsite(t, app, websitesCollection, nuvioNewsletterBackofficeGammaWebsiteID, "newsletter-gamma", "Newsletter Gamma")

	upsertNuvioNewsletterBackofficeRecord(t, app, groupsCollection, nuvioNewsletterBackofficeAlphaGroupID, map[string]any{
		"website": nuvioNewsletterBackofficeAlphaWebsiteID,
		"name":    "Alpha Group",
		"slug":    "alpha-group",
	})
	upsertNuvioNewsletterBackofficeRecord(t, app, groupsCollection, nuvioNewsletterBackofficeBetaGroupID, map[string]any{
		"website": nuvioNewsletterBackofficeBetaWebsiteID,
		"name":    "Beta Group",
		"slug":    "beta-group",
	})

	upsertNuvioNewsletterBackofficeRecord(t, app, subscribersCollection, nuvioNewsletterBackofficeAlphaSubID, map[string]any{
		"website":                    nuvioNewsletterBackofficeAlphaWebsiteID,
		"email":                      "alpha-subscriber@example.test",
		"name":                       "Alpha Subscriber",
		"status":                     "pending",
		"source":                     "manual_dashboard",
		"groups":                     []string{nuvioNewsletterBackofficeAlphaGroupID},
		"confirmedAt":                "2026-05-20T09:00:00Z",
		"unsubscribedAt":             "",
		"confirmationTokenHash":      "alpha-confirm-token-hash",
		"confirmationTokenExpiresAt": "2026-05-22T09:00:00Z",
		"unsubscribeTokenHash":       "alpha-unsubscribe-token-hash",
	})
	upsertNuvioNewsletterBackofficeRecord(t, app, subscribersCollection, nuvioNewsletterBackofficeBetaSubID, map[string]any{
		"website":                    nuvioNewsletterBackofficeBetaWebsiteID,
		"email":                      "beta-subscriber@example.test",
		"name":                       "Beta Subscriber",
		"status":                     "active",
		"source":                     "manual_dashboard",
		"groups":                     []string{nuvioNewsletterBackofficeBetaGroupID},
		"confirmedAt":                "2026-05-21T09:00:00Z",
		"unsubscribedAt":             "",
		"confirmationTokenHash":      "beta-confirm-token-hash",
		"confirmationTokenExpiresAt": "2026-05-23T09:00:00Z",
		"unsubscribeTokenHash":       "beta-unsubscribe-token-hash",
	})

	upsertNuvioNewsletterBackofficeRecord(t, app, campaignsCollection, nuvioNewsletterBackofficeAlphaCampID, map[string]any{
		"website":         nuvioNewsletterBackofficeAlphaWebsiteID,
		"subject":         "Alpha Campaign",
		"body":            "<p>Alpha campaign body</p>",
		"status":          "draft",
		"recipientsType":  "manual",
		"recipientsIds":   []string{nuvioNewsletterBackofficeAlphaSubID},
		"recipientsCount": 1,
		"sentAt":          "",
	})
	upsertNuvioNewsletterBackofficeRecord(t, app, campaignsCollection, nuvioNewsletterBackofficeBetaCampID, map[string]any{
		"website":         nuvioNewsletterBackofficeBetaWebsiteID,
		"subject":         "Beta Campaign",
		"body":            "<p>Beta campaign body</p>",
		"status":          "sent",
		"recipientsType":  "all",
		"recipientsIds":   []string{nuvioNewsletterBackofficeBetaSubID},
		"recipientsCount": 10,
		"sentAt":          "2026-05-25T09:00:00Z",
	})
}

func ensureNuvioNewsletterBackofficeSubscribersCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioNewsletterBackofficeCollection(t, app, "Subscribers", nuvioSubscribersCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "email"},
		&core.TextField{Name: "name"},
		&core.SelectField{Name: "status", MaxSelect: 1, Values: []string{"pending", "active", "unsubscribed"}},
		&core.TextField{Name: "source"},
		&core.JSONField{Name: "groups"},
		&core.TextField{Name: "confirmedAt"},
		&core.TextField{Name: "unsubscribedAt"},
		&core.TextField{Name: "confirmationTokenHash"},
		&core.TextField{Name: "confirmationTokenExpiresAt"},
		&core.TextField{Name: "unsubscribeTokenHash"},
	})
}

func ensureNuvioNewsletterBackofficeCampaignsCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioNewsletterBackofficeCollection(t, app, "Campaigns", nuvioCampaignsCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "subject"},
		&core.TextField{Name: "body"},
		&core.SelectField{Name: "status", MaxSelect: 1, Values: []string{"draft", "sent"}},
		&core.SelectField{Name: "recipientsType", MaxSelect: 1, Values: []string{"all", "manual"}},
		&core.JSONField{Name: "recipientsIds"},
		&core.NumberField{Name: "recipientsCount"},
		&core.TextField{Name: "sentAt"},
	})
}

func ensureNuvioNewsletterBackofficeGroupsCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()
	return ensureNuvioNewsletterBackofficeCollection(t, app, "SubscriberGroups", nuvioSubscriberGroupsCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "slug"},
	})
}

func ensureNuvioNewsletterBackofficeCollection(
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

func upsertNuvioNewsletterBackofficeWebsite(
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
	record.Set("settings", map[string]any{
		"newsletter": map[string]any{
			"doubleOptIn": true,
		},
		"reports": map[string]any{
			"analytics": map[string]any{
				"apiUrl": "https://attacker-controlled.example.test/api",
			},
		},
	})

	if saveErr := app.Save(record); saveErr != nil {
		t.Fatalf("failed to save website %s: %v", websiteID, saveErr)
	}
}

func upsertNuvioNewsletterBackofficeRecord(
	t testing.TB,
	app *tests.TestApp,
	collection *core.Collection,
	recordID string,
	values map[string]any,
) {
	t.Helper()

	record, err := app.FindFirstRecordByFilter(collection, "id={:id}", dbx.Params{"id": recordID})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("failed to lookup record %s in %s: %v", recordID, collection.Name, err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		record = core.NewRecord(collection)
		record.Id = recordID
	}

	for key, value := range values {
		record.Set(strings.TrimSpace(key), value)
	}

	if saveErr := app.Save(record); saveErr != nil {
		t.Fatalf("failed to save record %s in %s: %v", recordID, collection.Name, saveErr)
	}
}

func findNuvioNewsletterBackofficeSubscriberByWebsiteEmail(
	t testing.TB,
	app *tests.TestApp,
	websiteID string,
	email string,
) *core.Record {
	t.Helper()

	subscribersCollection := ensureNuvioNewsletterBackofficeSubscribersCollection(t, app)
	record, err := app.FindFirstRecordByFilter(
		subscribersCollection,
		"website={:website} && email={:email}",
		dbx.Params{
			"website": websiteID,
			"email":   email,
		},
	)
	if err != nil {
		t.Fatalf("failed to find subscriber by email %s: %v", email, err)
	}

	return record
}

func mustFindNuvioNewsletterBackofficeRecordByID(
	t testing.TB,
	app *tests.TestApp,
	collectionID string,
	recordID string,
) *core.Record {
	t.Helper()

	record, err := app.FindRecordById(collectionID, recordID)
	if err != nil {
		t.Fatalf("failed to find record %s in %s: %v", recordID, collectionID, err)
	}

	return record
}

func assertNuvioNewsletterBackofficeRecordField(
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

func normalizeNewsletterStatusForTest(rawStatus string) string {
	return strings.ToLower(strings.TrimSpace(rawStatus))
}
