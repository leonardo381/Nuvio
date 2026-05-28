package main

import (
	"database/sql"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const backofficeWebsitesTestSuperuserAuthToken = "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhdXRoIiwiY29sbGVjdGlvbklkIjoicGJjXzMxNDI2MzU4MjMiLCJleHAiOjI1MjQ2MDQ0NjEsInJlZnJlc2hhYmxlIjp0cnVlfQ.UXgO3j-0BumcugrFjbd7j0M4MQvbrLggLlcu_YNGjoY"

func TestNuvioBackofficeWebsitesEndpoint(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin superuser receives all websites",
			Method: http.MethodGet,
			URL:    "/api/nuvio/backoffice/websites",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBackofficeWebsitesRoute(t, app, e)
				websiteIDs := seedNuvioBackofficeWebsiteSelectorRecords(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, websiteIDs)
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"slug":"alpha-site"`,
				`"slug":"beta-site"`,
				`"slug":"gamma-site"`,
			},
			NotExpectedContent: []string{
				`"settings":`,
				`"apiUrl"`,
			},
		},
		{
			Name:   "client superuser receives only assigned websites",
			Method: http.MethodGet,
			URL:    "/api/nuvio/backoffice/websites",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBackofficeWebsitesRoute(t, app, e)
				websiteIDs := seedNuvioBackofficeWebsiteSelectorRecords(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{websiteIDs["beta-site"]})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"slug":"beta-site"`,
			},
			NotExpectedContent: []string{
				`"slug":"alpha-site"`,
				`"slug":"gamma-site"`,
				`"settings":`,
				`"apiUrl"`,
			},
		},
		{
			Name:   "client superuser with multiple websiteAccess receives exactly assigned websites",
			Method: http.MethodGet,
			URL:    "/api/nuvio/backoffice/websites",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBackofficeWebsitesRoute(t, app, e)
				websiteIDs := seedNuvioBackofficeWebsiteSelectorRecords(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
					websiteIDs["beta-site"],
					websiteIDs["gamma-site"],
				})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"slug":"beta-site"`,
				`"slug":"gamma-site"`,
			},
			NotExpectedContent: []string{
				`"slug":"alpha-site"`,
				`"settings":`,
				`"apiUrl"`,
			},
		},
		{
			Name:   "client superuser with no website access receives empty list",
			Method: http.MethodGet,
			URL:    "/api/nuvio/backoffice/websites",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBackofficeWebsitesRoute(t, app, e)
				seedNuvioBackofficeWebsiteSelectorRecords(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{})
			},
			ExpectedStatus:  200,
			ExpectedContent: []string{`[]`},
		},
		{
			Name:   "missing role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/backoffice/websites",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBackofficeWebsitesRoute(t, app, e)
				websiteIDs := seedNuvioBackofficeWebsiteSelectorRecords(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "", websiteIDs)
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role superuser denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/backoffice/websites",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBackofficeWebsitesRoute(t, app, e)
				websiteIDs := seedNuvioBackofficeWebsiteSelectorRecords(t, app)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, "manager", websiteIDs)
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated denied",
			Method: http.MethodGet,
			URL:    "/api/nuvio/backoffice/websites",
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioBackofficeWebsitesRoute(t, app, e)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestListNuvioBackofficeWebsitesForAuthUsesFreshWebsiteAccessFromDB(t *testing.T) {
	t.Parallel()

	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer app.Cleanup()

	websiteIDs := seedNuvioBackofficeWebsiteSelectorRecords(t, app)
	setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{
		websiteIDs["beta-site"],
		websiteIDs["gamma-site"],
	})

	authRecord, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, "test@example.com")
	if err != nil {
		t.Fatalf("failed to load test superuser: %v", err)
	}

	staleAuthRecord := authRecord.Clone()
	staleAuthRecord.Set("websiteAccess", []string{})

	records, err := listNuvioBackofficeWebsitesForAuth(app, staleAuthRecord)
	if err != nil {
		t.Fatalf("listNuvioBackofficeWebsitesForAuth failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 scoped websites, got %d", len(records))
	}

	gotIDs := map[string]struct{}{}
	for _, record := range records {
		gotIDs[record.Id] = struct{}{}
	}

	for _, expectedID := range []string{websiteIDs["beta-site"], websiteIDs["gamma-site"]} {
		if _, ok := gotIDs[expectedID]; !ok {
			t.Fatalf("expected scoped website id %q to be present in response", expectedID)
		}
	}
}

func TestNormalizeNuvioBackofficeWebsiteAccessIDs(t *testing.T) {
	t.Parallel()

	got := normalizeNuvioBackofficeWebsiteAccessIDs(
		[]string{"alpha", " beta ", "", "alpha"},
		[]any{
			"gamma",
			map[string]any{"id": "delta"},
			map[string]any{"recordId": "epsilon"},
			map[string]any{"websiteId": "zeta"},
			nil,
		},
		map[string]any{"id": "theta"},
	)

	expected := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "theta"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("unexpected normalized ids; expected %v, got %v", expected, got)
	}
}

func setupNuvioBackofficeWebsitesRoute(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()
	_ = app
	registerNuvioBackofficeWebsitesRoutes(e)
}

func seedNuvioBackofficeWebsiteSelectorRecords(t testing.TB, app *tests.TestApp) map[string]string {
	t.Helper()

	websitesCollection := ensureNuvioBackofficeWebsitesCollection(t, app)

	alphaID := upsertNuvioBackofficeTestWebsiteRecord(t, app, websitesCollection, "alpha-site", "Alpha Website", "active")
	betaID := upsertNuvioBackofficeTestWebsiteRecord(t, app, websitesCollection, "beta-site", "Beta Website", "active")
	gammaID := upsertNuvioBackofficeTestWebsiteRecord(t, app, websitesCollection, "gamma-site", "Gamma Website", "inactive")

	return map[string]string{
		"alpha-site": alphaID,
		"beta-site":  betaID,
		"gamma-site": gammaID,
	}
}

func ensureNuvioBackofficeWebsitesCollection(t testing.TB, app *tests.TestApp) *core.Collection {
	t.Helper()

	collection, err := app.FindCollectionByNameOrId(nuvioWebsitesCollectionID)
	if err == nil {
		return collection
	}

	collection = core.NewBaseCollection("Websites", nuvioWebsitesCollectionID)
	collection.Fields.Add(&core.TextField{Name: "name"})
	collection.Fields.Add(&core.TextField{Name: "title"})
	collection.Fields.Add(&core.TextField{Name: "slug"})
	collection.Fields.Add(&core.TextField{Name: "domain"})
	collection.Fields.Add(&core.TextField{Name: "status"})
	collection.Fields.Add(&core.JSONField{Name: "settings"})

	if saveErr := app.Save(collection); saveErr != nil {
		t.Fatalf("failed to create Websites collection: %v", saveErr)
	}

	return collection
}

func upsertNuvioBackofficeTestWebsiteRecord(
	t testing.TB,
	app *tests.TestApp,
	collection *core.Collection,
	slug string,
	name string,
	status string,
) string {
	t.Helper()

	record, err := app.FindFirstRecordByFilter(collection, "slug={:slug}", dbx.Params{"slug": slug})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("failed to lookup website %s: %v", slug, err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		record = core.NewRecord(collection)
	}

	record.Set("name", name)
	record.Set("title", name)
	record.Set("slug", slug)
	record.Set("domain", slug+".example.test")
	record.Set("status", status)
	record.Set("settings", map[string]any{
		"reports": map[string]any{
			"analytics": map[string]any{
				"apiUrl": "https://attacker-controlled.example.test/api",
			},
		},
	})

	if saveErr := app.Save(record); saveErr != nil {
		t.Fatalf("failed to save website %s: %v", slug, saveErr)
	}

	return record.Id
}

func ensureNuvioBackofficeSuperuserRoleField(t testing.TB, app *tests.TestApp) {
	t.Helper()

	superusersCollection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatalf("failed to load superusers collection: %v", err)
	}

	updated := false
	if superusersCollection.Fields.GetByName("role") == nil {
		superusersCollection.Fields.Add(&core.SelectField{
			Name:      "role",
			MaxSelect: 1,
			Values:    []string{apis.SuperuserRoleAdmin, apis.SuperuserRoleClient},
		})
		updated = true
	}

	if superusersCollection.Fields.GetByName("websiteAccess") == nil {
		superusersCollection.Fields.Add(&core.RelationField{
			Name:         "websiteAccess",
			CollectionId: nuvioWebsitesCollectionID,
			MaxSelect:    100,
		})
		updated = true
	}

	if updated {
		if err := app.Save(superusersCollection); err != nil {
			t.Fatalf("failed to update superusers collection fields: %v", err)
		}
	}
}

func setNuvioBackofficeSuperuserRoleAndAccess(t testing.TB, app *tests.TestApp, role string, websiteAccess any) {
	t.Helper()

	ensureNuvioBackofficeSuperuserRoleField(t, app)

	superuser, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, "test@example.com")
	if err != nil {
		t.Fatalf("failed to load test superuser: %v", err)
	}

	superuser.Set("role", role)

	switch typed := websiteAccess.(type) {
	case []string:
		superuser.Set("websiteAccess", typed)
	case map[string]string:
		access := make([]string, 0, len(typed))
		for _, id := range typed {
			access = append(access, id)
		}
		superuser.Set("websiteAccess", access)
	default:
		superuser.Set("websiteAccess", []string{})
	}

	if role != "" && role != apis.SuperuserRoleAdmin && role != apis.SuperuserRoleClient {
		err = app.SaveNoValidate(superuser)
	} else {
		err = app.Save(superuser)
	}

	if err != nil {
		t.Fatalf("failed to save test superuser role/access: %v", err)
	}
}
