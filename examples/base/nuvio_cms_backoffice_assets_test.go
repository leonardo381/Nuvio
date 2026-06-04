package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

const (
	nuvioCMSAssetAlphaRecordID  = "cmsassetalpha01"
	nuvioCMSAssetBetaRecordID   = "cmsassetbeta001"
	nuvioCMSAssetLegacyRecordID = "cmsassetlegacy1"
)

func TestNuvioCMSBackofficeAssetEndpoints(t *testing.T) {
	t.Parallel()

	pngBody := makeNuvioCMSAssetTinyPNG(t)
	jpegBody := makeNuvioCMSAssetTinyJPEG(t)
	svgBody := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"><rect width="1" height="1"/></svg>`)
	textBody := []byte("not-an-image")
	oversizedBody := bytes.Repeat([]byte("a"), nuvioCMSBackofficeAssetMaxFileSizeBytes+1)
	expectedPNGChecksum := nuvioCMSAssetSHA256Hex(pngBody)
	expectedJPEGChecksum := nuvioCMSAssetSHA256Hex(jpegBody)

	adminPNGUploadBody, adminPNGContentType := buildNuvioCMSAssetUploadBody(
		t,
		map[string]string{"websiteId": nuvioCMSDashboardAlphaWebsiteID},
		map[string][]byte{"file:admin-upload.png": pngBody},
	)
	clientJPEGUploadBody, clientJPEGContentType := buildNuvioCMSAssetUploadBody(
		t,
		map[string]string{"websiteId": nuvioCMSDashboardAlphaWebsiteID},
		map[string][]byte{"file:client-upload.jpg": jpegBody},
	)
	unassignedUploadBody, unassignedUploadContentType := buildNuvioCMSAssetUploadBody(
		t,
		map[string]string{"websiteId": nuvioCMSDashboardBetaWebsiteID},
		map[string][]byte{"file:forbidden-upload.png": pngBody},
	)
	missingWebsiteBody, missingWebsiteContentType := buildNuvioCMSAssetUploadBody(
		t,
		map[string]string{},
		map[string][]byte{"file:no-website.png": pngBody},
	)
	missingFileBody, missingFileContentType := buildNuvioCMSAssetUploadBody(
		t,
		map[string]string{"websiteId": nuvioCMSDashboardAlphaWebsiteID},
		map[string][]byte{},
	)
	svgUploadBody, svgUploadContentType := buildNuvioCMSAssetUploadBody(
		t,
		map[string]string{"websiteId": nuvioCMSDashboardAlphaWebsiteID},
		map[string][]byte{"file:blocked.svg": svgBody},
	)
	invalidMimeUploadBody, invalidMimeUploadContentType := buildNuvioCMSAssetUploadBody(
		t,
		map[string]string{"websiteId": nuvioCMSDashboardAlphaWebsiteID},
		map[string][]byte{"file:invalid.txt": textBody},
	)
	oversizedUploadBody, oversizedUploadContentType := buildNuvioCMSAssetUploadBody(
		t,
		map[string]string{"websiteId": nuvioCMSDashboardAlphaWebsiteID},
		map[string][]byte{"file:too-large.png": oversizedBody},
	)

	scenarios := []tests.ApiScenario{
		{
			Name:   "admin can list all instance assets after website access check",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/assets?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
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
				`"id":"` + nuvioCMSAssetAlphaRecordID + `"`,
				`"website":"` + nuvioCMSDashboardAlphaWebsiteID + `"`,
				`"id":"` + nuvioCMSAssetBetaRecordID + `"`,
				`"website":"` + nuvioCMSDashboardBetaWebsiteID + `"`,
				`"id":"` + nuvioCMSAssetLegacyRecordID + `"`,
			},
			NotExpectedContent: []string{
				`"checksum"`,
				`"data":`,
			},
		},
		{
			Name:   "client can list all instance assets after assigned website access check",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/assets?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"id":"` + nuvioCMSAssetAlphaRecordID + `"`,
				`"id":"` + nuvioCMSAssetBetaRecordID + `"`,
				`"id":"` + nuvioCMSAssetLegacyRecordID + `"`,
			},
			NotExpectedContent: []string{`"checksum"`},
		},
		{
			Name:   "client cannot list assets for unassigned website",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/assets?websiteId=" + nuvioCMSDashboardBetaWebsiteID,
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "unauthenticated user cannot list assets",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/assets?websiteId=" + nuvioCMSDashboardAlphaWebsiteID,
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
			},
			ExpectedStatus:  401,
			ExpectedContent: []string{`"status":401`},
		},
		{
			Name:   "assets list missing websiteId returns bad request",
			Method: http.MethodGet,
			URL:    "/api/nuvio/cms/assets",
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "admin can upload valid png asset",
			Method: http.MethodPost,
			URL:    "/api/nuvio/cms/assets",
			Body:   bytes.NewReader(adminPNGUploadBody),
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
				"Content-Type":  adminPNGContentType,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
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
				`"website":"` + nuvioCMSDashboardAlphaWebsiteID + `"`,
				`"mimeType":"image/png"`,
				`"originalName":"admin-upload.png"`,
				`"collection":"Assets"`,
			},
			NotExpectedContent: []string{
				`"checksum"`,
				`"storage"`,
			},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				assetsCollection, err := app.FindCollectionByNameOrId(nuvioAssetsCollectionID)
				if err != nil {
					t.Fatalf("expected assets collection: %v", err)
				}
				record, err := app.FindFirstRecordByFilter(
					assetsCollection,
					`originalName={:name}`,
					map[string]any{"name": "admin-upload.png"},
				)
				if err != nil {
					t.Fatalf("expected uploaded asset record: %v", err)
				}

				websiteID := resolveNuvioPublicRelationID(record, "website", "site")
				if websiteID != nuvioCMSDashboardAlphaWebsiteID {
					t.Fatalf("expected asset website %q, got %q", nuvioCMSDashboardAlphaWebsiteID, websiteID)
				}
				if checksum := strings.TrimSpace(record.GetString("checksum")); checksum != expectedPNGChecksum {
					t.Fatalf("expected uploaded png checksum %q, got %q", expectedPNGChecksum, checksum)
				}
			},
		},
		{
			Name:   "client assigned can upload valid jpeg asset",
			Method: http.MethodPost,
			URL:    "/api/nuvio/cms/assets",
			Body:   bytes.NewReader(clientJPEGUploadBody),
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
				"Content-Type":  clientJPEGContentType,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus: 200,
			ExpectedContent: []string{
				`"state":"ok"`,
				`"website":"` + nuvioCMSDashboardAlphaWebsiteID + `"`,
				`"mimeType":"image/jpeg"`,
				`"originalName":"client-upload.jpg"`,
			},
			NotExpectedContent: []string{`"checksum"`},
			AfterTestFunc: func(t testing.TB, app *tests.TestApp, _ *http.Response) {
				assetsCollection, err := app.FindCollectionByNameOrId(nuvioAssetsCollectionID)
				if err != nil {
					t.Fatalf("expected assets collection: %v", err)
				}
				record, err := app.FindFirstRecordByFilter(
					assetsCollection,
					`originalName={:name}`,
					map[string]any{"name": "client-upload.jpg"},
				)
				if err != nil {
					t.Fatalf("expected uploaded client asset record: %v", err)
				}
				if checksum := strings.TrimSpace(record.GetString("checksum")); checksum != expectedJPEGChecksum {
					t.Fatalf("expected uploaded jpeg checksum %q, got %q", expectedJPEGChecksum, checksum)
				}
			},
		},
		{
			Name:   "client unassigned cannot upload asset",
			Method: http.MethodPost,
			URL:    "/api/nuvio/cms/assets",
			Body:   bytes.NewReader(unassignedUploadBody),
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
				"Content-Type":  unassignedUploadContentType,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleClient, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  403,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "upload missing websiteId returns bad request",
			Method: http.MethodPost,
			URL:    "/api/nuvio/cms/assets",
			Body:   bytes.NewReader(missingWebsiteBody),
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
				"Content-Type":  missingWebsiteContentType,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "upload missing file returns bad request",
			Method: http.MethodPost,
			URL:    "/api/nuvio/cms/assets",
			Body:   bytes.NewReader(missingFileBody),
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
				"Content-Type":  missingFileContentType,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "upload svg is rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/cms/assets",
			Body:   bytes.NewReader(svgUploadBody),
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
				"Content-Type":  svgUploadContentType,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "upload invalid mime is rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/cms/assets",
			Body:   bytes.NewReader(invalidMimeUploadBody),
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
				"Content-Type":  invalidMimeUploadContentType,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
		{
			Name:   "upload oversized file is rejected",
			Method: http.MethodPost,
			URL:    "/api/nuvio/cms/assets",
			Body:   bytes.NewReader(oversizedUploadBody),
			Headers: map[string]string{
				"Authorization": backofficeWebsitesTestSuperuserAuthToken,
				"Content-Type":  oversizedUploadContentType,
			},
			BeforeTestFunc: func(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
				setupNuvioCMSBackofficeAssetEndpointsData(t, app, e)
				setNuvioBackofficeSuperuserRoleAndAccess(t, app, apis.SuperuserRoleAdmin, []string{nuvioCMSDashboardAlphaWebsiteID})
			},
			ExpectedStatus:  400,
			ExpectedContent: []string{`"data":{}`},
		},
	}

	for _, scenario := range scenarios {
		scenario.Test(t)
	}
}

func setupNuvioCMSBackofficeAssetEndpointsData(t testing.TB, app *tests.TestApp, e *core.ServeEvent) {
	t.Helper()

	setupNuvioCMSBackofficeDashboardRoute(t, app, e)
	seedNuvioCMSBackofficeDashboardData(t, app)
	seedNuvioCMSBackofficeAssetRecords(t, app)
}

func seedNuvioCMSBackofficeAssetRecords(t testing.TB, app *tests.TestApp) {
	t.Helper()

	assetsCollection := ensureNuvioCMSBackofficeCollection(
		t,
		app,
		"Assets",
		nuvioAssetsCollectionID,
		[]core.Field{
			&core.FileField{
				Name:      "file",
				MaxSelect: 1,
				MaxSize:   nuvioCMSBackofficeAssetMaxFileSizeBytes,
				MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
			},
			&core.RelationField{
				Name:         "website",
				CollectionId: nuvioWebsitesCollectionID,
				MaxSelect:    1,
				MinSelect:    0,
			},
			&core.TextField{Name: "originalName"},
			&core.TextField{Name: "mimeType"},
			&core.NumberField{Name: "size"},
			&core.TextField{Name: "checksum"},
		},
	)

	alphaSeedFile := makeNuvioCMSSeedAssetFile(t, makeNuvioCMSAssetTinyPNG(t), "alpha-seeded.png")
	betaSeedFile := makeNuvioCMSSeedAssetFile(t, makeNuvioCMSAssetTinyPNG(t), "beta-seeded.png")
	legacySeedFile := makeNuvioCMSSeedAssetFile(t, makeNuvioCMSAssetTinyPNG(t), "legacy-seeded.png")

	upsertNuvioCMSBackofficeRecord(t, app, assetsCollection, nuvioCMSAssetAlphaRecordID, map[string]any{
		"file":         alphaSeedFile,
		"website":      []string{nuvioCMSDashboardAlphaWebsiteID},
		"originalName": "alpha-seeded.png",
		"mimeType":     "image/png",
		"size":         1024,
	})
	upsertNuvioCMSBackofficeRecord(t, app, assetsCollection, nuvioCMSAssetBetaRecordID, map[string]any{
		"file":         betaSeedFile,
		"website":      []string{nuvioCMSDashboardBetaWebsiteID},
		"originalName": "beta-seeded.png",
		"mimeType":     "image/png",
		"size":         2048,
	})
	upsertNuvioCMSBackofficeRecord(t, app, assetsCollection, nuvioCMSAssetLegacyRecordID, map[string]any{
		"file":         legacySeedFile,
		"originalName": "legacy-seeded.png",
		"mimeType":     "image/png",
		"size":         4096,
	})
}

func makeNuvioCMSAssetTinyPNG(t testing.TB) []byte {
	t.Helper()

	imageBuffer := new(bytes.Buffer)
	rect := image.Rect(0, 0, 1, 1)
	img := image.NewNRGBA(rect)
	img.Set(0, 0, color.NRGBA{R: 17, G: 52, B: 86, A: 255})

	if err := png.Encode(imageBuffer, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	return imageBuffer.Bytes()
}

func makeNuvioCMSAssetTinyJPEG(t testing.TB) []byte {
	t.Helper()

	imageBuffer := new(bytes.Buffer)
	rect := image.Rect(0, 0, 1, 1)
	img := image.NewNRGBA(rect)
	img.Set(0, 0, color.NRGBA{R: 32, G: 64, B: 96, A: 255})

	if err := jpeg.Encode(imageBuffer, img, &jpeg.Options{Quality: 75}); err != nil {
		t.Fatalf("failed to encode jpeg: %v", err)
	}

	return imageBuffer.Bytes()
}

func makeNuvioCMSSeedAssetFile(t testing.TB, content []byte, name string) *filesystem.File {
	t.Helper()

	file, err := filesystem.NewFileFromBytes(content, name)
	if err != nil {
		t.Fatalf("failed to create seed asset file %q: %v", name, err)
	}
	file.Name = name

	return file
}

func buildNuvioCMSAssetUploadBody(
	t testing.TB,
	fields map[string]string,
	files map[string][]byte,
) ([]byte, string) {
	t.Helper()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("failed to write multipart field %q: %v", key, err)
		}
	}

	for key, content := range files {
		parts := strings.SplitN(key, ":", 2)
		fieldName := strings.TrimSpace(parts[0])
		filename := fieldName
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			filename = strings.TrimSpace(parts[1])
		}

		partWriter, err := writer.CreateFormFile(fieldName, filename)
		if err != nil {
			t.Fatalf("failed to create multipart file field %q: %v", fieldName, err)
		}
		if _, err := io.Copy(partWriter, bytes.NewReader(content)); err != nil {
			t.Fatalf("failed to write multipart file %q: %v", fieldName, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	return body.Bytes(), writer.FormDataContentType()
}

func nuvioCMSAssetSHA256Hex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
