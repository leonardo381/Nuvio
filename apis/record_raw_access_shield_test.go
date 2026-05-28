package apis_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

const testSuperuserAuthToken = "eyJhbGciOiJIUzI1NiJ9.eyJpZCI6InN5d2JoZWNuaDQ2cmhtMCIsInR5cGUiOiJhdXRoIiwiY29sbGVjdGlvbklkIjoicGJjXzMxNDI2MzU4MjMiLCJleHAiOjI1MjQ2MDQ0NjEsInJlZnJlc2hhYmxlIjp0cnVlfQ.UXgO3j-0BumcugrFjbd7j0M4MQvbrLggLlcu_YNGjoY"

func setRawAccessTestSuperuserRole(t testing.TB, app *tests.TestApp, role string) {
	t.Helper()

	superusersCollection, err := app.FindCollectionByNameOrId(core.CollectionNameSuperusers)
	if err != nil {
		t.Fatalf("failed to load superusers collection: %v", err)
	}

	if superusersCollection.Fields.GetByName("role") == nil {
		superusersCollection.Fields.Add(&core.SelectField{
			Name:      "role",
			MaxSelect: 1,
			Values:    []string{apis.SuperuserRoleAdmin, apis.SuperuserRoleClient},
		})
		if err := app.Save(superusersCollection); err != nil {
			t.Fatalf("failed to add superuser role field: %v", err)
		}
	}

	superuser, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, "test@example.com")
	if err != nil {
		t.Fatalf("failed to load test superuser: %v", err)
	}

	superuser.Set("role", role)
	if role != "" && role != apis.SuperuserRoleAdmin && role != apis.SuperuserRoleClient {
		err = app.SaveNoValidate(superuser)
	} else {
		err = app.Save(superuser)
	}
	if err != nil {
		t.Fatalf("failed to save test superuser role: %v", err)
	}
}

func TestRawRecordCrudClientRoleShield(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin superuser can still list raw records",
			Method: http.MethodGet,
			URL:    "/api/collections/demo2/records",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "admin")
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"items":[{`,
				`"totalItems":3`,
			},
		},
		{
			Name:   "client-role superuser denied listing raw records",
			Method: http.MethodGet,
			URL:    "/api/collections/demo2/records",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "client")
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "missing role superuser denied listing raw records",
			Method: http.MethodGet,
			URL:    "/api/collections/demo2/records",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "")
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unknown role superuser denied listing raw records",
			Method: http.MethodGet,
			URL:    "/api/collections/demo2/records",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "manager")
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:           "unauthenticated public list behavior preserved",
			Method:         http.MethodGet,
			URL:            "/api/collections/demo2/records",
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"items":[{`,
				`"totalItems":3`,
			},
		},
		{
			Name:   "client-role superuser denied viewing raw record",
			Method: http.MethodGet,
			URL:    "/api/collections/demo2/records/0yxhwia2amd8gec",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "client")
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client-role superuser denied creating raw record",
			Method: http.MethodPost,
			URL:    "/api/collections/demo2/records",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"title":"raw_guard_create"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "client")
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client-role superuser denied updating raw record",
			Method: http.MethodPatch,
			URL:    "/api/collections/demo2/records/0yxhwia2amd8gec",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			Body: strings.NewReader(`{"title":"raw_guard_update"}`),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "client")
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "client-role superuser denied deleting raw record",
			Method: http.MethodDelete,
			URL:    "/api/collections/demo2/records/0yxhwia2amd8gec",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "client")
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestBatchRawRecordCrudClientRoleShield(t *testing.T) {
	t.Parallel()

	batchCreateBody := `{
		"requests": [
			{"method":"POST", "url":"/api/collections/demo2/records", "body": {"title": "batch_raw_guard"}}
		]
	}`

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin superuser can still perform batch raw create",
			Method: http.MethodPost,
			URL:    "/api/batch",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			Body: strings.NewReader(batchCreateBody),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "admin")
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"status":200`,
				`"title":"batch_raw_guard"`,
			},
		},
		{
			Name:   "client-role superuser denied batch raw create",
			Method: http.MethodPost,
			URL:    "/api/batch",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			Body: strings.NewReader(batchCreateBody),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "client")
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"batch_request_failed"`,
				`"The authorized record is not allowed to perform this action."`,
			},
		},
		{
			Name:   "missing role superuser denied batch raw create",
			Method: http.MethodPost,
			URL:    "/api/batch",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			Body: strings.NewReader(batchCreateBody),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "")
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"batch_request_failed"`,
				`"The authorized record is not allowed to perform this action."`,
			},
		},
		{
			Name:   "unknown role superuser denied batch raw create",
			Method: http.MethodPost,
			URL:    "/api/batch",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			Body: strings.NewReader(batchCreateBody),
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "manager")
			},
			ExpectedStatus: 400,
			ExpectedContent: []string{
				`"batch_request_failed"`,
				`"The authorized record is not allowed to perform this action."`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func TestRawRecordShieldDoesNotBlockAuthRefresh(t *testing.T) {
	t.Parallel()

	scenarios := []tests.ApiScenario{
		{
			Name:   "client-role superuser can still refresh auth token",
			Method: http.MethodPost,
			URL:    "/api/collections/_superusers/auth-refresh",
			Headers: map[string]string{
				"Authorization": testSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, _ *core.ServeEvent) {
				setRawAccessTestSuperuserRole(t, app, "client")
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"token":"`,
				`"record":{`,
			},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}
