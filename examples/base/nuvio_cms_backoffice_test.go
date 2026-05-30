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
	nuvioCMSDashboardAlphaWebsiteID = "cmsalpha0000001"
	nuvioCMSDashboardBetaWebsiteID  = "cmsbeta00000002"
	nuvioCMSDashboardGammaWebsiteID = "cmsgamma0000003"

	nuvioCMSDashboardAlphaPageID       = "cmspagealpha001"
	nuvioCMSDashboardAlphaSecondPageID = "cmspagealpha002"
	nuvioCMSDashboardBetaPageID        = "cmspagebeta0001"

	nuvioCMSDashboardAlphaBlockID      = "cmsblockalpha01"
	nuvioCMSDashboardAlphaOtherBlockID = "cmsblockalpha02"
	nuvioCMSDashboardBetaBlockID       = "cmsblockbeta001"

	nuvioCMSDashboardHeroComponentID = "cmscomponent001"
	nuvioCMSDashboardFaqComponentID  = "cmscomponent002"
)

func TestNuvioCMSBackofficeDashboardEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin superuser receives scoped cms dashboard data",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{
						nuvioCMSDashboardAlphaWebsiteID,
						nuvioCMSDashboardBetaWebsiteID,
					},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioCMSDashboardAlphaWebsiteID + `"`,
				`"website":{"id":"` + nuvioCMSDashboardAlphaWebsiteID + `"`,
				`"pages":[`,
				`"page":{"id":"` + nuvioCMSDashboardAlphaPageID + `"`,
				`"blocks":[`,
				`"components":[`,
				`"canUseFileFields":false`,
				`"id":"` + nuvioCMSDashboardAlphaBlockID + `"`,
				`"id":"` + nuvioCMSDashboardHeroComponentID + `"`,
			},
			NotExpectedContent: []string{
				`"id":"` + nuvioCMSDashboardBetaPageID + `"`,
				`"id":"` + nuvioCMSDashboardBetaBlockID + `"`,
				`"id":"` + nuvioCMSDashboardAlphaOtherBlockID + `"`,
				`"apiUrl"`,
				`"smsGatewayToken"`,
				`"secret component metadata"`,
				`"providerConfig"`,
			},
		},
		{
			Name:   "client superuser receives only assigned website cms data",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleClient,
					[]string{nuvioCMSDashboardAlphaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioCMSDashboardAlphaWebsiteID + `"`,
				`"id":"` + nuvioCMSDashboardAlphaPageID + `"`,
				`"id":"` + nuvioCMSDashboardAlphaBlockID + `"`,
			},
			NotExpectedContent: []string{
				`"id":"` + nuvioCMSDashboardBetaPageID + `"`,
				`"id":"` + nuvioCMSDashboardBetaBlockID + `"`,
				`"siteId":"alpha-umami-site"`,
				`"scriptUrl"`,
				`"apiUrl"`,
			},
		},
		{
			Name:   "client superuser denied for unassigned website",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleClient,
					[]string{nuvioCMSDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client superuser with no website access denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "missing role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "foreign page id denied with not found",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID + "&pageId=" + nuvioCMSDashboardBetaPageID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioCMSDashboardAlphaWebsiteID, nuvioCMSDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "explicit page id returns only matching page blocks",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/dashboard?websiteId=" + nuvioCMSDashboardAlphaWebsiteID + "&pageId=" + nuvioCMSDashboardAlphaSecondPageID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioCMSDashboardAlphaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"page":{"id":"` + nuvioCMSDashboardAlphaSecondPageID + `"`,
				`"id":"` + nuvioCMSDashboardAlphaOtherBlockID + `"`,
			},
			NotExpectedContent: []string{
				`"id":"` + nuvioCMSDashboardAlphaBlockID + `"`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioCMSBackofficeWebsiteEndpoints(t *testing.T) {
	t.Parallel()

	overlongCMSSettingsMessage := strings.Repeat("m", nuvioCMSBackofficeSettingsMessageMaxLen+1)
	overlongCMSTemplateSubject := strings.Repeat("s", nuvioCMSBackofficeSettingsTemplateSubjectMaxLen+1)
	overlongCMSTemplateText := strings.Repeat("t", nuvioCMSBackofficeSettingsTemplateTextMaxLen+1)
	overlongCMSI18NLabel := strings.Repeat("l", nuvioCMSBackofficeI18NLanguageLabelMaxLen+1)

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin can update website identity and global seo",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seoTitle":"Updated Alpha SEO title",
				"businessPhone":"+351933000111"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioCMSDashboardAlphaWebsiteID, nuvioCMSDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioCMSDashboardAlphaWebsiteID + `"`,
				`"seoTitle":"Updated Alpha SEO title"`,
			},
			NotExpectedContent: []string{
				`"apiUrl"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(website, []string{"seoTitle", "seo_title"})) != "Updated Alpha SEO title" {
					t.Fatalf("expected seoTitle to be updated")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(website, []string{"business_phone", "businessPhone"})) != "+351933000111" {
					t.Fatalf("expected business_phone to be updated")
				}
			},
		},
		{
			Name:   "client assigned can update allowed identity fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seoDescription":"Client updated SEO description"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"seoDescription":"Client updated SEO description"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(website, []string{"seoDescription", "seo_description"})) != "Client updated SEO description" {
					t.Fatalf("expected seoDescription to be updated")
				}
			},
		},
		{
			Name:   "identity endpoint accepts valid canonical domain",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_canonical_domain":"https://alpha.example/base"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"seo_canonical_domain":"https://alpha.example/base"`,
			},
		},
		{
			Name:   "identity endpoint rejects javascript canonical domain",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_canonical_domain":"javascript:alert(1)"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "identity endpoint accepts valid business social profiles",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"business_social_profiles":"https://facebook.com/alpha,https://instagram.com/alpha"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(website, []string{"business_social_profiles", "businessSocialProfiles"})) == "" {
					t.Fatalf("expected business_social_profiles to be updated")
				}
			},
		},
		{
			Name:   "identity endpoint rejects invalid business social profile entry",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"business_social_profiles":"https://facebook.com/alpha,javascript:alert(1)"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "identity endpoint accepts valid business email",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"business_email":"seo-team@alpha.example"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"business_email":"seo-team@alpha.example"`,
			},
		},
		{
			Name:   "identity endpoint rejects invalid business email",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"business_email":"not-an-email"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "identity endpoint rejects overlong seo title",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seoTitle":"` + strings.Repeat("x", 301) + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "client unassigned denied for identity endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seoTitle":"Blocked"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "identity endpoint rejects settings and feature flags fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"settings":{"booking":{"enabled":false}},
				"featureFlags":{"booking":false}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "identity endpoint rejects unknown field",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"ownerEmail":"should-not-work"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "identity endpoint defers logo and seo image file fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/identity",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"logo":"new-logo.webp"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "admin can update settings and preserve hidden keys",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"settings":{
					"featureFlags":{"booking":false},
					"contactForm":{"enabled":false,"confirmationMessage":"Updated confirmation"},
					"reports":{"analytics":{"enabled":true,"provider":"umami","scriptEnabled":false,"siteId":"updated-site","scriptUrl":"https://analytics.updated.example/script.js","events":{"scrollDepth":false}}},
					"booking":{"rules":{"minNoticeHours":4}}
				}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				injectNuvioCMSBackofficeHiddenSettings(t, app, nuvioCMSDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"websiteId":"` + nuvioCMSDashboardAlphaWebsiteID + `"`,
				`"minNoticeHours":4`,
				`"siteId":"updated-site"`,
				`"scriptUrl":"https://analytics.updated.example/script.js"`,
			},
			NotExpectedContent: []string{
				`"apiUrl"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				settings := parseNuvioSettingsObject(website.Get("settings"))
				featureFlags, ok := toStringAnyMap(settings["featureFlags"])
				if !ok {
					t.Fatalf("expected featureFlags settings")
				}
				if value, ok := parseBoolValue(featureFlags["booking"]); !ok || value {
					t.Fatalf("expected featureFlags.booking to be false")
				}

				contactForm, ok := toStringAnyMap(settings["contactForm"])
				if !ok {
					t.Fatalf("expected contactForm settings")
				}
				if value, ok := parseBoolValue(contactForm["enabled"]); !ok || value {
					t.Fatalf("expected contactForm.enabled to be false")
				}
				if strings.TrimSpace(parseStringValue(contactForm["confirmationMessage"])) != "Updated confirmation" {
					t.Fatalf("expected contactForm.confirmationMessage to be updated")
				}

				booking, ok := toStringAnyMap(settings["booking"])
				if !ok {
					t.Fatalf("expected booking settings")
				}
				if strings.TrimSpace(parseStringValue(booking["privateKey"])) != "keep-hidden" {
					t.Fatalf("expected hidden booking.privateKey to be preserved")
				}
				rules, ok := toStringAnyMap(booking["rules"])
				if !ok {
					t.Fatalf("expected booking rules")
				}
				if parseNuvioNonNegativeInt(rules["minNoticeHours"], 0) != 4 {
					t.Fatalf("expected minNoticeHours to be updated")
				}

				newsletter, ok := toStringAnyMap(settings["newsletter"])
				if !ok {
					t.Fatalf("expected newsletter settings")
				}
				if strings.TrimSpace(parseStringValue(newsletter["legacyHidden"])) != "keep-hidden" {
					t.Fatalf("expected hidden newsletter setting to be preserved")
				}

				reports, ok := toStringAnyMap(settings["reports"])
				if !ok {
					t.Fatalf("expected reports settings")
				}
				analytics, ok := toStringAnyMap(reports["analytics"])
				if !ok {
					t.Fatalf("expected reports.analytics settings")
				}
				if strings.TrimSpace(parseStringValue(analytics["apiUrl"])) != "https://analytics.alpha.example/api" {
					t.Fatalf("expected reports.analytics.apiUrl to be preserved")
				}
				if strings.TrimSpace(parseStringValue(analytics["scriptUrl"])) != "https://analytics.updated.example/script.js" {
					t.Fatalf("expected reports.analytics.scriptUrl to be updated")
				}
			},
		},
		{
			Name:   "settings endpoint accepts valid contact form confirmation message",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{"confirmationMessage":"Thanks for reaching out. We will contact you shortly."}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"confirmationMessage":"Thanks for reaching out. We will contact you shortly."`,
			},
		},
		{
			Name:   "settings endpoint rejects overlong contact form confirmation message",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"contactForm":{"confirmationMessage":"` + overlongCMSSettingsMessage + `"}}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "settings endpoint accepts valid whatsapp default message",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"whatsapp":{"defaultMessage":"Hello, I would like more information about this service."}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"defaultMessage":"Hello, I would like more information about this service."`,
			},
		},
		{
			Name:   "settings endpoint rejects overlong whatsapp default message",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"whatsapp":{"defaultMessage":"` + overlongCMSSettingsMessage + `"}}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "settings endpoint accepts valid newsletter confirmation template text",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"newsletter":{"lifecycle":{"confirmationTemplate":{"subject":"Please confirm your subscription","introText":"Click the button below to confirm.","footerText":"If you did not subscribe, you can ignore this email."}}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"subject":"Please confirm your subscription"`,
			},
		},
		{
			Name:   "settings endpoint rejects overlong newsletter confirmation template subject",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"newsletter":{"lifecycle":{"confirmationTemplate":{"subject":"` + overlongCMSTemplateSubject + `"}}}}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "settings endpoint rejects overlong newsletter confirmation template body",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"newsletter":{"lifecycle":{"confirmationTemplate":{"introText":"` + overlongCMSTemplateText + `"}}}}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "settings endpoint accepts valid booking visitor template text",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"booking":{"visitorEmails":{"confirmationTemplate":{"subject":"Your booking is confirmed","introText":"Thank you for your booking.","footerText":"Reply to this email if you need help."}}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"subject":"Your booking is confirmed"`,
			},
		},
		{
			Name:   "settings endpoint rejects overlong booking visitor template text",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"booking":{"visitorEmails":{"requestTemplate":{"footerText":"` + overlongCMSTemplateText + `"}}}}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "settings endpoint rejects invalid i18n language code",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"i18n":{"languages":[{"code":"pt<script>","label":"Portuguese"}]}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "settings endpoint rejects overlong i18n language label",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"i18n":{"languages":[{"code":"en","label":"` + overlongCMSI18NLabel + `"}]}}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "client assigned can update client-safe settings groups",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{"enabled":false},
				"booking":{"rules":{"bufferMinutes":25}},
				"i18n":{"enabled":true,"languages":[{"code":"en","label":"English"},{"code":"pt","label":"Portuguese"}]}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"bufferMinutes":25`,
				`"languages":[{"code":"en","label":"English"},{"code":"pt","label":"Portuguese"}]`,
			},
			NotExpectedContent: []string{
				`"apiUrl"`,
				`"siteId":"alpha-umami-site"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				settings := parseNuvioSettingsObject(website.Get("settings"))
				contactForm, ok := toStringAnyMap(settings["contactForm"])
				if !ok {
					t.Fatalf("expected contactForm settings")
				}
				if value, ok := parseBoolValue(contactForm["enabled"]); !ok || value {
					t.Fatalf("expected contactForm.enabled to be false")
				}

				booking, ok := toStringAnyMap(settings["booking"])
				if !ok {
					t.Fatalf("expected booking settings")
				}
				rules, ok := toStringAnyMap(booking["rules"])
				if !ok {
					t.Fatalf("expected booking rules")
				}
				if parseNuvioNonNegativeInt(rules["bufferMinutes"], 0) != 25 {
					t.Fatalf("expected booking.rules.bufferMinutes to be updated")
				}
			},
		},
		{
			Name:   "client unassigned denied for settings endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{"enabled":false}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client cannot update feature flags through settings endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"featureFlags":{"booking":false}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "client cannot update reports analytics technical config",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"reports":{"analytics":{"siteId":"hijack-site","scriptEnabled":false}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "settings endpoint rejects reports analytics apiUrl mutation",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"reports":{"analytics":{"apiUrl":"https://evil.example/api"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "admin can update notification recipients",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{"emailNotifications":{"to":["hijack@example.test"],"cc":["cc-hijack@example.test"]}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"hijack@example.test"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				settings := parseNuvioSettingsObject(website.Get("settings"))
				contactForm, ok := toStringAnyMap(settings["contactForm"])
				if !ok {
					t.Fatalf("expected contactForm settings")
				}
				emailNotifications, ok := toStringAnyMap(contactForm["emailNotifications"])
				if !ok {
					t.Fatalf("expected contactForm.emailNotifications settings")
				}
				toRecipients := parseNuvioRecipientIDs(emailNotifications["to"])
				if len(toRecipients) != 1 || toRecipients[0] != "hijack@example.test" {
					t.Fatalf("expected contactForm.emailNotifications.to to be updated")
				}
				ccRecipients := parseNuvioRecipientIDs(emailNotifications["cc"])
				if len(ccRecipients) != 1 || ccRecipients[0] != "cc-hijack@example.test" {
					t.Fatalf("expected contactForm.emailNotifications.cc to be updated")
				}
			},
		},
		{
			Name:   "admin can update settings using flat compatibility payloads",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{
					"to":["flat-contact@example.test"],
					"cc":["flat-contact-cc@example.test"],
					"template":{"subject":"Flat contact template subject"}
				},
				"whatsapp":{
					"to":["flat-whatsapp@example.test"],
					"cc":["flat-whatsapp-cc@example.test"],
					"template":{"subject":"Flat whatsapp template subject"}
				},
				"newsletter":{
					"confirmationTemplate":{"enabled":true,"subject":"Flat newsletter template subject"}
				},
				"booking":{
					"to":["flat-booking@example.test"],
					"cc":["flat-booking-cc@example.test"],
					"businessTemplate":{"subject":"Flat booking business template subject"},
					"requestTemplate":{"subject":"Flat request template subject"},
					"confirmationTemplate":{"subject":"Flat confirmation template subject"},
					"rescheduleTemplate":{"subject":"Flat reschedule template subject"},
					"minNoticeHours":7,
					"bookingWindowDays":21,
					"bufferMinutes":10,
					"calendarBlockingMode":"website"
				}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				settings := parseNuvioSettingsObject(website.Get("settings"))

				contactForm, ok := toStringAnyMap(settings["contactForm"])
				if !ok {
					t.Fatalf("expected contactForm settings")
				}
				contactNotifications, ok := toStringAnyMap(contactForm["emailNotifications"])
				if !ok {
					t.Fatalf("expected contactForm.emailNotifications settings")
				}
				contactTo := parseNuvioRecipientIDs(contactNotifications["to"])
				if len(contactTo) != 1 || contactTo[0] != "flat-contact@example.test" {
					t.Fatalf("expected contactForm flat to recipient to be mapped")
				}

				whatsapp, ok := toStringAnyMap(settings["whatsapp"])
				if !ok {
					t.Fatalf("expected whatsapp settings")
				}
				whatsappNotifications, ok := toStringAnyMap(whatsapp["emailNotifications"])
				if !ok {
					t.Fatalf("expected whatsapp.emailNotifications settings")
				}
				whatsappTo := parseNuvioRecipientIDs(whatsappNotifications["to"])
				if len(whatsappTo) != 1 || whatsappTo[0] != "flat-whatsapp@example.test" {
					t.Fatalf("expected whatsapp flat to recipient to be mapped")
				}

				newsletter, ok := toStringAnyMap(settings["newsletter"])
				if !ok {
					t.Fatalf("expected newsletter settings")
				}
				lifecycle, ok := toStringAnyMap(newsletter["lifecycle"])
				if !ok {
					t.Fatalf("expected newsletter.lifecycle settings")
				}
				confirmationTemplate, ok := toStringAnyMap(lifecycle["confirmationTemplate"])
				if !ok {
					t.Fatalf("expected newsletter.lifecycle.confirmationTemplate settings")
				}
				if strings.TrimSpace(parseStringValue(confirmationTemplate["subject"])) != "Flat newsletter template subject" {
					t.Fatalf("expected newsletter flat confirmationTemplate to be mapped")
				}

				booking, ok := toStringAnyMap(settings["booking"])
				if !ok {
					t.Fatalf("expected booking settings")
				}
				bookingNotifications, ok := toStringAnyMap(booking["emailNotifications"])
				if !ok {
					t.Fatalf("expected booking.emailNotifications settings")
				}
				bookingTo := parseNuvioRecipientIDs(bookingNotifications["to"])
				if len(bookingTo) != 1 || bookingTo[0] != "flat-booking@example.test" {
					t.Fatalf("expected booking flat to recipient to be mapped")
				}
				businessTemplate, ok := toStringAnyMap(bookingNotifications["businessTemplate"])
				if !ok {
					t.Fatalf("expected booking email businessTemplate settings")
				}
				if strings.TrimSpace(parseStringValue(businessTemplate["subject"])) != "Flat booking business template subject" {
					t.Fatalf("expected booking flat businessTemplate to be mapped")
				}

				visitorEmails, ok := toStringAnyMap(booking["visitorEmails"])
				if !ok {
					t.Fatalf("expected booking.visitorEmails settings")
				}
				requestTemplate, ok := toStringAnyMap(visitorEmails["requestTemplate"])
				if !ok || strings.TrimSpace(parseStringValue(requestTemplate["subject"])) != "Flat request template subject" {
					t.Fatalf("expected booking flat requestTemplate to be mapped")
				}
				confirmationVisitorTemplate, ok := toStringAnyMap(visitorEmails["confirmationTemplate"])
				if !ok || strings.TrimSpace(parseStringValue(confirmationVisitorTemplate["subject"])) != "Flat confirmation template subject" {
					t.Fatalf("expected booking flat confirmationTemplate to be mapped")
				}
				rescheduleTemplate, ok := toStringAnyMap(visitorEmails["rescheduleTemplate"])
				if !ok || strings.TrimSpace(parseStringValue(rescheduleTemplate["subject"])) != "Flat reschedule template subject" {
					t.Fatalf("expected booking flat rescheduleTemplate to be mapped")
				}

				rules, ok := toStringAnyMap(booking["rules"])
				if !ok {
					t.Fatalf("expected booking.rules settings")
				}
				if parseNuvioNonNegativeInt(rules["minNoticeHours"], 0) != 7 {
					t.Fatalf("expected booking flat minNoticeHours to be mapped")
				}
				if parseNuvioNonNegativeInt(rules["bookingWindowDays"], 0) != 21 {
					t.Fatalf("expected booking flat bookingWindowDays to be mapped")
				}
				if parseNuvioNonNegativeInt(rules["bufferMinutes"], 0) != 10 {
					t.Fatalf("expected booking flat bufferMinutes to be mapped")
				}
				if strings.TrimSpace(parseStringValue(rules["calendarBlockingMode"])) != "website" {
					t.Fatalf("expected booking flat calendarBlockingMode to be mapped")
				}
			},
		},
		{
			Name:   "client can update contact form notification recipients and preserve hidden keys",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{"emailNotifications":{"to":["client-contact@example.test"],"cc":["client-contact-cc@example.test"]}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				injectNuvioCMSBackofficeHiddenSettings(t, app, nuvioCMSDashboardAlphaWebsiteID)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"client-contact@example.test"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				settings := parseNuvioSettingsObject(website.Get("settings"))
				contactForm, ok := toStringAnyMap(settings["contactForm"])
				if !ok {
					t.Fatalf("expected contactForm settings")
				}
				emailNotifications, ok := toStringAnyMap(contactForm["emailNotifications"])
				if !ok {
					t.Fatalf("expected contactForm.emailNotifications settings")
				}
				toRecipients := parseNuvioRecipientIDs(emailNotifications["to"])
				if len(toRecipients) != 1 || toRecipients[0] != "client-contact@example.test" {
					t.Fatalf("expected contactForm.emailNotifications.to to be updated for client")
				}
				ccRecipients := parseNuvioRecipientIDs(emailNotifications["cc"])
				if len(ccRecipients) != 1 || ccRecipients[0] != "client-contact-cc@example.test" {
					t.Fatalf("expected contactForm.emailNotifications.cc to be updated for client")
				}

				booking, ok := toStringAnyMap(settings["booking"])
				if !ok {
					t.Fatalf("expected booking settings")
				}
				if strings.TrimSpace(parseStringValue(booking["privateKey"])) != "keep-hidden" {
					t.Fatalf("expected hidden booking.privateKey to be preserved")
				}
			},
		},
		{
			Name:   "client can update whatsapp notification recipients",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"whatsapp":{"emailNotifications":{"to":["client-whatsapp@example.test"],"cc":["client-whatsapp-cc@example.test"]}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"client-whatsapp@example.test"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				settings := parseNuvioSettingsObject(website.Get("settings"))
				whatsapp, ok := toStringAnyMap(settings["whatsapp"])
				if !ok {
					t.Fatalf("expected whatsapp settings")
				}
				emailNotifications, ok := toStringAnyMap(whatsapp["emailNotifications"])
				if !ok {
					t.Fatalf("expected whatsapp.emailNotifications settings")
				}
				toRecipients := parseNuvioRecipientIDs(emailNotifications["to"])
				if len(toRecipients) != 1 || toRecipients[0] != "client-whatsapp@example.test" {
					t.Fatalf("expected whatsapp.emailNotifications.to to be updated for client")
				}
				ccRecipients := parseNuvioRecipientIDs(emailNotifications["cc"])
				if len(ccRecipients) != 1 || ccRecipients[0] != "client-whatsapp-cc@example.test" {
					t.Fatalf("expected whatsapp.emailNotifications.cc to be updated for client")
				}
			},
		},
		{
			Name:   "client can update booking notification recipients",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"booking":{"emailNotifications":{"to":["client-booking@example.test"],"cc":["client-booking-cc@example.test"]}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"client-booking@example.test"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				website, err := app.FindRecordById(nuvioWebsitesCollectionID, nuvioCMSDashboardAlphaWebsiteID)
				if err != nil {
					t.Fatalf("expected website to exist: %v", err)
				}

				settings := parseNuvioSettingsObject(website.Get("settings"))
				booking, ok := toStringAnyMap(settings["booking"])
				if !ok {
					t.Fatalf("expected booking settings")
				}
				emailNotifications, ok := toStringAnyMap(booking["emailNotifications"])
				if !ok {
					t.Fatalf("expected booking.emailNotifications settings")
				}
				toRecipients := parseNuvioRecipientIDs(emailNotifications["to"])
				if len(toRecipients) != 1 || toRecipients[0] != "client-booking@example.test" {
					t.Fatalf("expected booking.emailNotifications.to to be updated for client")
				}
				ccRecipients := parseNuvioRecipientIDs(emailNotifications["cc"])
				if len(ccRecipients) != 1 || ccRecipients[0] != "client-booking-cc@example.test" {
					t.Fatalf("expected booking.emailNotifications.cc to be updated for client")
				}
			},
		},
		{
			Name:   "settings endpoint rejects invalid notification recipients payload shape",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{"emailNotifications":{"to":"invalid-shape@example.test"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "settings endpoint rejects arbitrary keys",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"smsGateway":{"token":"should-not-work"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "missing role denied for cms website write endpoints",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{"enabled":false}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role denied for cms website write endpoints",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"contactForm":{"enabled":false}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied for cms website write endpoints",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/websites/" + nuvioCMSDashboardAlphaWebsiteID + "/settings",
			Body: strings.NewReader(`{
				"contactForm":{"enabled":false}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioCMSBackofficePageSEOEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin can update page seo fields and preserve non seo fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_title":"Updated Alpha Home SEO title",
				"seo_description":"Updated Alpha Home SEO description",
				"seo_social_image":"updated-alpha-home-seo.webp",
				"seo_canonical_url":"https://alpha.example/updated-home",
				"seo_noindex":true,
				"seo_exclude_from_sitemap":true,
				"seo_focus_keyword":"updated alpha keyword",
				"seo_translations":{
					"en":{"title":"Updated Home","description":"Updated EN description"},
					"pt":{"title":"Inicio atualizado","description":"Descricao atualizada"}
				}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioCMSDashboardAlphaWebsiteID, nuvioCMSDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"page":{"id":"` + nuvioCMSDashboardAlphaPageID + `"`,
				`"seo_title":"Updated Alpha Home SEO title"`,
				`"seo_description":"Updated Alpha Home SEO description"`,
				`"seo_social_image":"updated-alpha-home-seo.webp"`,
				`"seo_canonical_url":"https://alpha.example/updated-home"`,
				`"seo_noindex":true`,
				`"seo_exclude_from_sitemap":true`,
				`"seo_focus_keyword":"updated alpha keyword"`,
			},
			NotExpectedContent: []string{
				`"settings":`,
				`"blocks":`,
				`"components":`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				pageRecord, err := app.FindRecordById(nuvioPagesCollectionID, nuvioCMSDashboardAlphaPageID)
				if err != nil {
					t.Fatalf("failed to load page record: %v", err)
				}

				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(pageRecord, []string{"seo_title", "seoTitle"})) != "Updated Alpha Home SEO title" {
					t.Fatalf("expected seo_title to be updated")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(pageRecord, []string{"seo_social_image", "seoSocialImage"})) != "updated-alpha-home-seo.webp" {
					t.Fatalf("expected seo_social_image to be updated")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(pageRecord, []string{"slug"})) != "home" {
					t.Fatalf("expected slug to be preserved")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(pageRecord, []string{"title"})) != "A Home" {
					t.Fatalf("expected title to be preserved")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(pageRecord, []string{"status"})) != "published" {
					t.Fatalf("expected status to be preserved")
				}
				if strings.TrimSpace(resolveNuvioPublicRelationID(pageRecord, "website", "site")) != nuvioCMSDashboardAlphaWebsiteID {
					t.Fatalf("expected website relation to be preserved")
				}
			},
		},
		{
			Name:   "client assigned can update page seo using camelCase aliases",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaSecondPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seoTitle":"Client updated contact title",
				"seoDescription":"Client updated contact description",
				"seoCanonicalUrl":"https://alpha.example/contact-updated",
				"seoNoindex":false,
				"seoExcludeFromSitemap":false,
				"seoFocusKeyword":"client keyword",
				"seoTranslations":{"en":{"title":"Client EN title","description":"Client EN description"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"page":{"id":"` + nuvioCMSDashboardAlphaSecondPageID + `"`,
				`"seo_title":"Client updated contact title"`,
				`"seo_description":"Client updated contact description"`,
				`"seo_canonical_url":"https://alpha.example/contact-updated"`,
				`"seo_focus_keyword":"client keyword"`,
			},
		},
		{
			Name:   "page seo endpoint accepts site relative canonical url",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_canonical_url":"/site/alpha/home"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"seo_canonical_url":"/site/alpha/home"`,
			},
		},
		{
			Name:   "page seo endpoint rejects javascript canonical url",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_canonical_url":"javascript:alert(1)"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "page seo endpoint accepts absolute social image url",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_social_image":"https://cdn.alpha.example/seo/home.webp"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"seo_social_image":"https://cdn.alpha.example/seo/home.webp"`,
			},
		},
		{
			Name:   "page seo endpoint rejects unsafe social image string",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_social_image":"javascript:alert(1)"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "page seo endpoint rejects overlong seo title",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_title":"` + strings.Repeat("t", 301) + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "page seo endpoint rejects overlong seo description",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_description":"` + strings.Repeat("d", 1001) + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "page seo endpoint rejects overlong seo focus keyword",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_focus_keyword":"` + strings.Repeat("k", 256) + `"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "client unassigned denied for page seo endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_title":"Blocked"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "page seo endpoint returns 404 for unknown page",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/unknown-page-id/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_title":"Missing page"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "page seo endpoint rejects arbitrary fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"customMeta":"not-allowed"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "page seo endpoint rejects website slug status title mutation",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"website":"` + nuvioCMSDashboardBetaWebsiteID + `",
				"slug":"hijacked",
				"status":"draft",
				"title":"Hijacked title",
				"seo_title":"Valid SEO title"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID, nuvioCMSDashboardBetaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				pageRecord, err := app.FindRecordById(nuvioPagesCollectionID, nuvioCMSDashboardAlphaPageID)
				if err != nil {
					t.Fatalf("failed to load page record: %v", err)
				}

				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(pageRecord, []string{"slug"})) != "home" {
					t.Fatalf("expected slug to remain unchanged")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(pageRecord, []string{"status"})) != "published" {
					t.Fatalf("expected status to remain unchanged")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(pageRecord, []string{"title"})) != "A Home" {
					t.Fatalf("expected title to remain unchanged")
				}
			},
		},
		{
			Name:   "page seo endpoint rejects social image file-like payload",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_social_image":{"filename":"hero.webp"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "page seo endpoint rejects invalid seo translations shape",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_translations":"invalid-shape"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "missing role denied for page seo endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_title":"Missing role"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role denied for page seo endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"seo_title":"Unknown role"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied for page seo endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/pages/" + nuvioCMSDashboardAlphaPageID + "/seo",
			Body: strings.NewReader(`{
				"seo_title":"No auth"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestNuvioCMSBackofficeBlockEndpoint(t *testing.T) {
	t.Parallel()

	overlongBlockURLValue := "https://example.test/" + strings.Repeat("a", nuvioCMSBackofficeURLMaxLen)
	overlongBlockTextValue := strings.Repeat("a", nuvioCMSBackofficeBlockStringMaxLen+1)

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin can update block props and translations and preserve non content fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"title":"Updated Hero Title","cta":{"label":"Book now"}},
				"translations":{"en":{"title":"Updated Hero EN"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioCMSDashboardAlphaWebsiteID, nuvioCMSDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"block":{"id":"` + nuvioCMSDashboardAlphaBlockID + `"`,
				`"title":"Updated Hero Title"`,
				`"label":"Book now"`,
			},
			NotExpectedContent: []string{
				`"settings":`,
				`"components":`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				blockRecord, err := app.FindRecordById(nuvioBlocksCollectionID, nuvioCMSDashboardAlphaBlockID)
				if err != nil {
					t.Fatalf("failed to load block record: %v", err)
				}

				props, ok := toStringAnyMap(normalizeNuvioPublicJSONValue(blockRecord.Get("props")))
				if !ok {
					t.Fatalf("expected block props map")
				}
				if strings.TrimSpace(parseStringValue(props["title"])) != "Updated Hero Title" {
					t.Fatalf("expected block props title to be updated")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(blockRecord, []string{"component_key", "componentKey"})) != "hero" {
					t.Fatalf("expected component key to remain unchanged")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(blockRecord, []string{"status"})) != "active" {
					t.Fatalf("expected block status to remain unchanged")
				}
				if parseNuvioCMSDashboardIntByAliases(blockRecord, []string{"displayOrder", "display_order", "order"}, -1) != 1 {
					t.Fatalf("expected block order to remain unchanged")
				}
			},
		},
		{
			Name:   "client assigned can update block props and translations array",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaOtherBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"items":[{"q":"When?","a":"Today"}]},
				"translations":[{"lang":"en","title":"FAQ"}]
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"block":{"id":"` + nuvioCMSDashboardAlphaOtherBlockID + `"`,
			},
		},
		{
			Name:   "client unassigned denied for block endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"title":"Blocked update"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardBetaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "block endpoint returns 404 for unknown block",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/unknown-block-id",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"title":"Missing block"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "block endpoint returns not found when block page is missing",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"title":"Will fail"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})

				blockRecord, err := app.FindRecordById(nuvioBlocksCollectionID, nuvioCMSDashboardAlphaBlockID)
				if err != nil {
					t.Fatalf("failed to load block: %v", err)
				}
				blockRecord.Set("page", "missing-page-id")
				if saveErr := app.Save(blockRecord); saveErr != nil {
					t.Fatalf("failed to corrupt block page relation: %v", saveErr)
				}
			},
			ExpectedStatus:  404,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "block endpoint rejects arbitrary fields",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"custom":"not-allowed"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint rejects page component order and status mutation",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"page":"` + nuvioCMSDashboardBetaPageID + `",
				"component":"` + nuvioCMSDashboardFaqComponentID + `",
				"component_key":"faq",
				"displayOrder":99,
				"order":99,
				"status":"draft",
				"visible":false,
				"props":{"title":"Should not apply"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(
					t,
					app,
					apis.SuperuserRoleAdmin,
					[]string{nuvioCMSDashboardAlphaWebsiteID, nuvioCMSDashboardBetaWebsiteID},
				)
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				blockRecord, err := app.FindRecordById(nuvioBlocksCollectionID, nuvioCMSDashboardAlphaBlockID)
				if err != nil {
					t.Fatalf("failed to load block: %v", err)
				}

				if strings.TrimSpace(resolveNuvioPublicRelationID(blockRecord, "page")) != nuvioCMSDashboardAlphaPageID {
					t.Fatalf("expected block page to remain unchanged")
				}
				if strings.TrimSpace(resolveNuvioPublicRelationID(blockRecord, "component")) != nuvioCMSDashboardHeroComponentID {
					t.Fatalf("expected block component to remain unchanged")
				}
				if parseNuvioCMSDashboardIntByAliases(blockRecord, []string{"displayOrder", "display_order", "order"}, -1) != 1 {
					t.Fatalf("expected block order to remain unchanged")
				}
				if strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(blockRecord, []string{"status"})) != "active" {
					t.Fatalf("expected block status to remain unchanged")
				}
			},
		},
		{
			Name:   "block endpoint rejects invalid props shape",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":"invalid"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint rejects invalid translations shape",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"translations":"invalid"
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint rejects file like payload in props",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"image":{"name":"hero.png","size":12345,"type":"image/png"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint accepts safe nested url like props and rich text",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{
					"description":"Contains http://example.test inside plain text.",
					"cta":{"href":"https://example.test/book","linkUrl":"/contact","actionUrl":"mailto:hello@example.test","profileUrl":"tel:+351999000111"},
					"hero":{"imageUrl":"hero-banner.webp","backgroundImageUrl":"/assets/bg.webp","embedUrl":"https://www.youtube.com/embed/abc123"}
				},
				"translations":{"en":{"buttonUrl":"https://example.test/en/book"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
			},
		},
		{
			Name:   "block endpoint rejects javascript href value",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"cta":{"href":"javascript:alert(1)"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint rejects data image url value",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"hero":{"imageUrl":"data:text/html;base64,SGVsbG8="}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint rejects protocol relative url value",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"cta":{"url":"//evil.example.test/path"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint rejects overlong url like value",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"props":{"ctaUrl":"` + overlongBlockURLValue + `"}}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint rejects overlong normal content string",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"props":{"content":"` + overlongBlockTextValue + `"}}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "block endpoint accepts non url field containing http text",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"description":"Documentation is at http://example.test/docs and should remain plain text."}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
			},
		},
		{
			Name:   "block endpoint rejects unsafe url like value inside translations",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"translations":{"en":{"ctaUrl":"javascript:alert(1)"}}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"data":{}`,
			},
		},
		{
			Name:   "missing role denied for block endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"title":"Missing role"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role denied for block endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			Body: strings.NewReader(`{
				"props":{"title":"Unknown role"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied for block endpoint",
			Method: http.MethodPatch,
			URL:    "/api/nuvio/cms/blocks/" + nuvioCMSDashboardAlphaBlockID,
			Body: strings.NewReader(`{
				"props":{"title":"No auth"}
			}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeDashboardRoute(t, app, e)
				seedNuvioCMSBackofficeDashboardData(t, app)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioCMSBackofficeDashboardRoute(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioCMSBackofficeRoutes(e)
}

func injectNuvioCMSBackofficeHiddenSettings(t testing.TB, app *tests.TestApp, websiteID string) {
	t.Helper()

	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		t.Fatalf("failed to load website %s: %v", websiteID, err)
	}

	settings := parseNuvioSettingsObject(website.Get("settings"))

	bookingSettings, ok := toStringAnyMap(settings["booking"])
	if !ok {
		bookingSettings = map[string]any{}
	}
	bookingSettings["privateKey"] = "keep-hidden"
	settings["booking"] = bookingSettings

	newsletterSettings, ok := toStringAnyMap(settings["newsletter"])
	if !ok {
		newsletterSettings = map[string]any{}
	}
	newsletterSettings["legacyHidden"] = "keep-hidden"
	settings["newsletter"] = newsletterSettings

	settings["adminOnlyHidden"] = map[string]any{
		"secret": "preserve-me",
	}

	website.Set("settings", settings)
	if saveErr := app.Save(website); saveErr != nil {
		t.Fatalf("failed to save hidden settings: %v", saveErr)
	}
}

func seedNuvioCMSBackofficeDashboardData(t testing.TB, app *tests.TestApp) {
	t.Helper()

	websitesCollection := ensureNuvioCMSBackofficeCollection(t, app, "Websites", nuvioWebsitesCollectionID, []core.Field{
		&core.TextField{Name: "name"},
		&core.TextField{Name: "title"},
		&core.TextField{Name: "slug"},
		&core.TextField{Name: "domain"},
		&core.TextField{Name: "logo"},
		&core.TextField{Name: "seoTitle"},
		&core.TextField{Name: "seoDescription"},
		&core.TextField{Name: "seoImage"},
		&core.TextField{Name: "seo_title_template"},
		&core.TextField{Name: "seo_title_separator"},
		&core.TextField{Name: "seo_canonical_domain"},
		&core.TextField{Name: "business_name"},
		&core.TextField{Name: "business_type"},
		&core.TextField{Name: "business_primary_category"},
		&core.TextField{Name: "business_phone"},
		&core.TextField{Name: "business_email"},
		&core.TextField{Name: "business_address"},
		&core.TextField{Name: "business_city"},
		&core.TextField{Name: "business_postal_code"},
		&core.TextField{Name: "business_country"},
		&core.TextField{Name: "business_service_area"},
		&core.TextField{Name: "business_opening_hours"},
		&core.TextField{Name: "business_google_place_id"},
		&core.TextField{Name: "business_social_profiles"},
		&core.TextField{Name: "business_price_range"},
		&core.JSONField{Name: "settings"},
	})
	pagesCollection := ensureNuvioCMSBackofficeCollection(t, app, "Pages", nuvioPagesCollectionID, []core.Field{
		&core.TextField{Name: "website"},
		&core.TextField{Name: "title"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "slug"},
		&core.TextField{Name: "path"},
		&core.TextField{Name: "url"},
		&core.TextField{Name: "status"},
		&core.BoolField{Name: "published"},
		&core.BoolField{Name: "visible"},
		&core.TextField{Name: "seo_title"},
		&core.TextField{Name: "seo_description"},
		&core.TextField{Name: "seo_social_image"},
		&core.TextField{Name: "seo_canonical_url"},
		&core.BoolField{Name: "seo_noindex"},
		&core.BoolField{Name: "seo_exclude_from_sitemap"},
		&core.TextField{Name: "seo_focus_keyword"},
		&core.JSONField{Name: "seo_translations"},
	})
	blocksCollection := ensureNuvioCMSBackofficeCollection(t, app, "Blocks", nuvioBlocksCollectionID, []core.Field{
		&core.TextField{Name: "page"},
		&core.TextField{Name: "website"},
		&core.TextField{Name: "component"},
		&core.TextField{Name: "component_key"},
		&core.TextField{Name: "variant"},
		&core.TextField{Name: "slot"},
		&core.NumberField{Name: "displayOrder"},
		&core.NumberField{Name: "order"},
		&core.JSONField{Name: "props"},
		&core.JSONField{Name: "translations"},
		&core.BoolField{Name: "enabled"},
		&core.BoolField{Name: "visible"},
		&core.TextField{Name: "status"},
	})
	componentsCollection := ensureNuvioCMSBackofficeCollection(t, app, "Components", nuvioComponentsCollectionID, []core.Field{
		&core.TextField{Name: "key"},
		&core.TextField{Name: "component_key"},
		&core.TextField{Name: "name"},
		&core.TextField{Name: "title"},
		&core.TextField{Name: "label"},
		&core.TextField{Name: "category"},
		&core.TextField{Name: "group"},
		&core.TextField{Name: "variant"},
		&core.TextField{Name: "defaultVariant"},
		&core.JSONField{Name: "schema"},
		&core.TextField{Name: "internalNotes"},
		&core.JSONField{Name: "providerConfig"},
	})

	upsertNuvioCMSBackofficeWebsiteRecord(t, app, websitesCollection, nuvioCMSDashboardAlphaWebsiteID, "alpha-cms", "Alpha CMS", map[string]any{
		"featureFlags": map[string]any{
			"whatsapp":    true,
			"contactForm": true,
			"newsletter":  true,
			"booking":     true,
			"reports":     true,
			"i18n":        true,
			"reviews":     true,
		},
		"contactForm": map[string]any{
			"enabled":             true,
			"confirmationMessage": "Thanks for contacting us",
			"fields": map[string]any{
				"phone": true,
			},
			"emailNotifications": map[string]any{
				"enabled": true,
				"to":      []any{"owner-alpha@example.test"},
				"cc":      []any{"cc-alpha@example.test"},
				"template": map[string]any{
					"enabled":            true,
					"subject":            "New contact lead",
					"introText":          "A new lead was received.",
					"includeLeadDetails": true,
					"footerText":         "Regards",
				},
			},
		},
		"whatsapp": map[string]any{
			"enabled":            true,
			"phone":              "+351900000000",
			"defaultMessage":     "Hello from WhatsApp",
			"showFloatingButton": true,
			"emailNotifications": map[string]any{
				"enabled": true,
				"to":      []any{"whatsapp-alpha@example.test"},
				"cc":      []any{"cc-whatsapp-alpha@example.test"},
				"template": map[string]any{
					"enabled":            true,
					"subject":            "New WhatsApp interaction",
					"introText":          "A WhatsApp interaction was tracked.",
					"includeLeadDetails": true,
					"footerText":         "Regards",
				},
			},
		},
		"newsletter": map[string]any{
			"doubleOptIn": true,
			"lifecycle": map[string]any{
				"confirmationTemplate": map[string]any{
					"enabled":    true,
					"subject":    "Confirm your subscription",
					"introText":  "Please confirm.",
					"footerText": "Team Alpha",
				},
			},
		},
		"booking": map[string]any{
			"enabled":          true,
			"confirmationMode": "request",
			"emailNotifications": map[string]any{
				"enabled": true,
				"to":      []any{"booking-alpha@example.test"},
				"cc":      []any{"booking-cc-alpha@example.test"},
				"businessTemplate": map[string]any{
					"enabled":                   true,
					"subject":                   "Booking request",
					"introText":                 "New booking request received.",
					"includeAppointmentDetails": true,
					"footerText":                "Regards",
				},
			},
			"visitorEmails": map[string]any{
				"requestTemplate": map[string]any{
					"enabled":    true,
					"subject":    "Request received",
					"introText":  "We received your request",
					"footerText": "Regards",
				},
				"confirmationTemplate": map[string]any{
					"enabled":    true,
					"subject":    "Booking confirmed",
					"introText":  "Your booking is confirmed",
					"footerText": "Regards",
				},
			},
			"rules": map[string]any{
				"minNoticeHours":       2,
				"bookingWindowDays":    30,
				"bufferMinutes":        15,
				"calendarBlockingMode": "service",
			},
		},
		"reports": map[string]any{
			"analytics": map[string]any{
				"enabled":       true,
				"provider":      "umami",
				"siteId":        "alpha-umami-site",
				"scriptEnabled": true,
				"scriptUrl":     "https://analytics.alpha.example/script.js",
				"apiUrl":        "https://analytics.alpha.example/api",
				"events": map[string]any{
					"scrollDepth": true,
				},
			},
		},
		"i18n": map[string]any{
			"enabled":         true,
			"defaultLanguage": "pt",
			"languages":       []any{"pt", "en"},
		},
		"providerInternals": map[string]any{
			"smsGatewayToken": "should-not-leak",
		},
	})
	upsertNuvioCMSBackofficeWebsiteRecord(t, app, websitesCollection, nuvioCMSDashboardBetaWebsiteID, "beta-cms", "Beta CMS", map[string]any{
		"featureFlags": map[string]any{
			"whatsapp": true,
			"reports":  true,
		},
	})
	upsertNuvioCMSBackofficeWebsiteRecord(t, app, websitesCollection, nuvioCMSDashboardGammaWebsiteID, "gamma-cms", "Gamma CMS", map[string]any{})

	upsertNuvioCMSBackofficeRecord(t, app, pagesCollection, nuvioCMSDashboardAlphaPageID, map[string]any{
		"website":                  nuvioCMSDashboardAlphaWebsiteID,
		"title":                    "A Home",
		"name":                     "Home",
		"slug":                     "home",
		"path":                     "/",
		"url":                      "/",
		"status":                   "published",
		"published":                true,
		"visible":                  true,
		"seo_title":                "Alpha Home SEO title",
		"seo_description":          "Alpha Home SEO description",
		"seo_social_image":         "alpha-home-seo.webp",
		"seo_canonical_url":        "https://alpha.example/home",
		"seo_noindex":              false,
		"seo_exclude_from_sitemap": false,
		"seo_focus_keyword":        "alpha keyword",
		"seo_translations": map[string]any{
			"pt": map[string]any{
				"title":       "Pagina inicial",
				"description": "Descricao em portugues",
			},
		},
	})
	upsertNuvioCMSBackofficeRecord(t, app, pagesCollection, nuvioCMSDashboardAlphaSecondPageID, map[string]any{
		"website":                  nuvioCMSDashboardAlphaWebsiteID,
		"title":                    "Z Contact",
		"name":                     "Contact",
		"slug":                     "contact",
		"path":                     "/contact",
		"url":                      "/contact",
		"status":                   "published",
		"published":                true,
		"visible":                  true,
		"seo_title":                "Contact SEO",
		"seo_description":          "Contact SEO description",
		"seo_social_image":         "alpha-contact-seo.webp",
		"seo_canonical_url":        "https://alpha.example/contact",
		"seo_noindex":              false,
		"seo_exclude_from_sitemap": false,
		"seo_focus_keyword":        "contact keyword",
		"seo_translations":         map[string]any{},
	})
	upsertNuvioCMSBackofficeRecord(t, app, pagesCollection, nuvioCMSDashboardBetaPageID, map[string]any{
		"website":                  nuvioCMSDashboardBetaWebsiteID,
		"title":                    "Beta Home",
		"name":                     "Beta Home",
		"slug":                     "beta-home",
		"path":                     "/beta-home",
		"url":                      "/beta-home",
		"status":                   "published",
		"published":                true,
		"visible":                  true,
		"seo_title":                "Beta SEO title",
		"seo_description":          "Beta SEO description",
		"seo_social_image":         "beta-seo.webp",
		"seo_canonical_url":        "https://beta.example/home",
		"seo_noindex":              false,
		"seo_exclude_from_sitemap": false,
		"seo_focus_keyword":        "beta keyword",
		"seo_translations":         map[string]any{},
	})

	upsertNuvioCMSBackofficeRecord(t, app, componentsCollection, nuvioCMSDashboardHeroComponentID, map[string]any{
		"key":            "hero",
		"component_key":  "hero",
		"name":           "Hero Section",
		"title":          "Hero",
		"label":          "Hero",
		"category":       "layout",
		"group":          "home",
		"variant":        "default",
		"defaultVariant": "default",
		"schema": map[string]any{
			"fields": []any{
				map[string]any{
					"key":   "title",
					"type":  "text",
					"label": "Title",
				},
				map[string]any{
					"key":   "image",
					"type":  "file",
					"label": "Hero image",
				},
			},
		},
		"internalNotes": "secret component metadata",
		"providerConfig": map[string]any{
			"apiKey": "hidden-api-key",
		},
	})
	upsertNuvioCMSBackofficeRecord(t, app, componentsCollection, nuvioCMSDashboardFaqComponentID, map[string]any{
		"key":            "faq",
		"component_key":  "faq",
		"name":           "FAQ Section",
		"title":          "FAQ",
		"label":          "FAQ",
		"category":       "content",
		"group":          "generic",
		"variant":        "default",
		"defaultVariant": "default",
		"schema": map[string]any{
			"fields": []any{
				map[string]any{
					"key":   "items",
					"type":  "array",
					"label": "FAQ items",
				},
			},
		},
	})

	upsertNuvioCMSBackofficeRecord(t, app, blocksCollection, nuvioCMSDashboardAlphaBlockID, map[string]any{
		"page":          nuvioCMSDashboardAlphaPageID,
		"website":       nuvioCMSDashboardAlphaWebsiteID,
		"component":     nuvioCMSDashboardHeroComponentID,
		"component_key": "hero",
		"variant":       "default",
		"slot":          "main",
		"displayOrder":  1,
		"order":         1,
		"props": map[string]any{
			"title": "Alpha Hero Title",
		},
		"translations": map[string]any{
			"pt": map[string]any{
				"title": "Titulo hero",
			},
		},
		"enabled": true,
		"visible": true,
		"status":  "active",
	})
	upsertNuvioCMSBackofficeRecord(t, app, blocksCollection, nuvioCMSDashboardAlphaOtherBlockID, map[string]any{
		"page":          nuvioCMSDashboardAlphaSecondPageID,
		"website":       nuvioCMSDashboardAlphaWebsiteID,
		"component":     nuvioCMSDashboardFaqComponentID,
		"component_key": "faq",
		"variant":       "default",
		"slot":          "main",
		"displayOrder":  2,
		"order":         2,
		"props": map[string]any{
			"items": []any{},
		},
		"translations": map[string]any{},
		"enabled":      true,
		"visible":      true,
		"status":       "active",
	})
	upsertNuvioCMSBackofficeRecord(t, app, blocksCollection, nuvioCMSDashboardBetaBlockID, map[string]any{
		"page":          nuvioCMSDashboardBetaPageID,
		"website":       nuvioCMSDashboardBetaWebsiteID,
		"component":     nuvioCMSDashboardHeroComponentID,
		"component_key": "hero",
		"variant":       "default",
		"slot":          "main",
		"displayOrder":  1,
		"order":         1,
		"props": map[string]any{
			"title": "Beta Hero Title",
		},
		"translations": map[string]any{},
		"enabled":      true,
		"visible":      true,
		"status":       "active",
	})
}

func ensureNuvioCMSBackofficeCollection(
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

func upsertNuvioCMSBackofficeWebsiteRecord(
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
	record.Set("logo", slug+"-logo.webp")
	record.Set("seoTitle", name+" SEO title")
	record.Set("seoDescription", name+" SEO description")
	record.Set("seoImage", slug+"-seo.webp")
	record.Set("seo_title_template", "{{pageName}} | {{siteName}}")
	record.Set("seo_title_separator", "|")
	record.Set("seo_canonical_domain", slug+".example.test")
	record.Set("business_name", name+" Business")
	record.Set("business_type", "LocalBusiness")
	record.Set("business_primary_category", "Services")
	record.Set("business_phone", "+351999000111")
	record.Set("business_email", slug+"@example.test")
	record.Set("business_address", "Main Street 1")
	record.Set("business_city", "Lisbon")
	record.Set("business_postal_code", "1000-100")
	record.Set("business_country", "PT")
	record.Set("business_service_area", "Lisbon")
	record.Set("business_opening_hours", "Mon-Fri 09:00-18:00")
	record.Set("business_google_place_id", "place_"+slug)
	record.Set("business_social_profiles", "https://instagram.com/"+slug)
	record.Set("business_price_range", "$$")
	record.Set("settings", settings)

	if saveErr := app.Save(record); saveErr != nil {
		t.Fatalf("failed to save website %s: %v", websiteID, saveErr)
	}
}

func upsertNuvioCMSBackofficeRecord(
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
