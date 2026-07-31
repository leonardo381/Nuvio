package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	pbcmd "github.com/pocketbase/pocketbase/cmd"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestNuvioSuperuserRoleBootstrapCommandEnablesCMSRoutes(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("failed to create test app: %v", err)
	}
	defer app.Cleanup()

	baseRouter, err := apis.NewRouter(app)
	if err != nil {
		t.Fatalf("failed to create router: %v", err)
	}
	serveEvent := &core.ServeEvent{App: app, Router: baseRouter}
	setupNuvioCMSBackofficeDashboardRoute(t, app, serveEvent)
	seedNuvioCMSBackofficeDashboardData(t, app)
	ensureNuvioBackofficeSuperuserRoleField(t, app)

	mux, err := serveEvent.Router.BuildMux()
	if err != nil {
		t.Fatalf("failed to build router: %v", err)
	}

	superuser, err := app.FindAuthRecordByEmail(core.CollectionNameSuperusers, "test@example.com")
	if err != nil {
		t.Fatalf("failed to load test superuser: %v", err)
	}
	superuser.Set("role", "")
	if err := app.Save(superuser); err != nil {
		t.Fatalf("failed to clear superuser role: %v", err)
	}

	emptyRoleToken, err := superuser.NewAuthToken()
	if err != nil {
		t.Fatalf("failed to create empty-role auth token: %v", err)
	}
	emptyStatus, emptyBody := performNuvioRoleBootstrapRequest(
		t,
		mux,
		http.MethodGet,
		"/api/nuvio/cms/dashboard?websiteId="+nuvioCMSDashboardAlphaWebsiteID,
		emptyRoleToken,
		"",
	)
	if emptyStatus != http.StatusForbidden {
		t.Fatalf("expected empty-role superuser dashboard request to be 403, got %d: %s", emptyStatus, emptyBody)
	}

	command := pbcmd.NewSuperuserCommand(app)
	command.SetArgs([]string{"upsert", "test@example.com", "1234567890!", "--role", apis.SuperuserRoleAdmin})
	if err := command.Execute(); err != nil {
		t.Fatalf("failed to run admin-role bootstrap command: %v", err)
	}

	superuser, err = app.FindAuthRecordByEmail(core.CollectionNameSuperusers, "test@example.com")
	if err != nil {
		t.Fatalf("failed to reload test superuser: %v", err)
	}
	if role := superuser.GetString("role"); role != apis.SuperuserRoleAdmin {
		t.Fatalf("expected CLI command to set admin role, got %q", role)
	}

	adminToken, err := superuser.NewAuthToken()
	if err != nil {
		t.Fatalf("failed to create admin auth token: %v", err)
	}
	status, body := performNuvioRoleBootstrapRequest(
		t,
		mux,
		http.MethodGet,
		"/api/nuvio/cms/dashboard?websiteId="+nuvioCMSDashboardAlphaWebsiteID,
		adminToken,
		"",
	)
	if status != http.StatusOK {
		t.Fatalf("expected admin-role superuser dashboard request to be 200, got %d: %s", status, body)
	}
	if !strings.Contains(body, `"state":"ok"`) || !strings.Contains(body, `"blocks"`) {
		t.Fatalf("expected CMS dashboard response to include state and blocks, got: %s", body)
	}

	patchStatus, patchBody := performNuvioRoleBootstrapRequest(
		t,
		mux,
		http.MethodPatch,
		"/api/nuvio/cms/blocks/"+nuvioCMSDashboardAlphaBlockID,
		adminToken,
		`{"props":{"title":"CLI admin role changed title"}}`,
	)
	if patchStatus != http.StatusOK {
		t.Fatalf("expected admin-role superuser block update to be 200, got %d: %s", patchStatus, patchBody)
	}
	if !strings.Contains(patchBody, `"title":"CLI admin role changed title"`) {
		t.Fatalf("expected block update response to include changed title, got: %s", patchBody)
	}

	blockRecord, err := app.FindRecordById(nuvioBlocksCollectionID, nuvioCMSDashboardAlphaBlockID)
	if err != nil {
		t.Fatalf("failed to load updated block: %v", err)
	}
	props, ok := toStringAnyMap(normalizeNuvioPublicJSONValue(blockRecord.Get("props")))
	if !ok {
		t.Fatalf("expected block props map")
	}
	if got := strings.TrimSpace(parseStringValue(props["title"])); got != "CLI admin role changed title" {
		t.Fatalf("expected block title to be updated through official CMS route, got %q", got)
	}
}

func performNuvioRoleBootstrapRequest(
	t testing.TB,
	mux http.Handler,
	method string,
	url string,
	token string,
	body string,
) (int, string) {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, url, reader)
	req.Header.Set("content-type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)

	return recorder.Code, recorder.Body.String()
}
