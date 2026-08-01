package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

const (
	nuvioComponentsCollectionID                     = "pbc_184785686"
	nuvioAssetsCollectionID                         = "pbc_1321337024"
	nuvioCMSDashboardMaxScanRecords                 = 5000
	nuvioCMSBackofficeAssetMaxFileSizeBytes         = 8 * 1024 * 1024
	nuvioCMSBackofficeURLMaxLen                     = 2048
	nuvioCMSBackofficeBlockStringMaxLen             = 10000
	nuvioCMSBackofficeSettingsMessageMaxLen         = 4000
	nuvioCMSBackofficeSettingsTemplateSubjectMaxLen = 200
	nuvioCMSBackofficeSettingsTemplateTextMaxLen    = 4000
	nuvioCMSBackofficeI18NLanguageCodeMaxLen        = 20
	nuvioCMSBackofficeI18NLanguageLabelMaxLen       = 80
)

var (
	nuvioCMSDashboardPagesCollectionAliases = []string{
		nuvioPagesCollectionID,
		"Pages",
		"pages",
	}
	nuvioCMSDashboardBlocksCollectionAliases = []string{
		nuvioBlocksCollectionID,
		"Blocks",
		"blocks",
	}
	nuvioCMSDashboardComponentsCollectionAliases = []string{
		nuvioComponentsCollectionID,
		"Components",
		"components",
		"Templates",
		"templates",
	}
	nuvioCMSDashboardAssetsCollectionAliases = []string{
		nuvioAssetsCollectionID,
		"Assets",
		"assets",
	}
	nuvioCMSBackofficeAssetAllowedMimeTypes = []string{
		"image/jpeg",
		"image/png",
		"image/webp",
		"image/gif",
	}
	nuvioCMSDashboardWebsiteFeatureFlagKeys = []string{
		"whatsapp",
		"contactForm",
		"reviews",
		"newsletter",
		"booking",
		"reports",
		"i18n",
	}
	nuvioCMSBackofficeIdentityAllowedPayloadKeys = map[string]struct{}{
		"name":                    {},
		"title":                   {},
		"displayname":             {},
		"slug":                    {},
		"domain":                  {},
		"seotitle":                {},
		"seodescription":          {},
		"seotitletemplate":        {},
		"seotitleseparator":       {},
		"seocanonicaldomain":      {},
		"businessname":            {},
		"businesstype":            {},
		"businessprimarycategory": {},
		"businessphone":           {},
		"businessemail":           {},
		"businessaddress":         {},
		"businesscity":            {},
		"businesspostalcode":      {},
		"businesscountry":         {},
		"businessservicearea":     {},
		"businessopeninghours":    {},
		"businessgoogleplaceid":   {},
		"businesssocialprofiles":  {},
		"businesspricerange":      {},
		"logo":                    {},
		"seoimage":                {},
		"seoimagecurrent":         {},
	}
	nuvioCMSBackofficeIdentityAdminOnlyPayloadKeys = map[string]struct{}{
		"domain": {},
	}
	nuvioCMSBackofficeIdentityDeferredFilePayloadKeys = map[string]struct{}{
		"seoimagecurrent": {},
	}
	nuvioCMSBackofficeSettingsAllowedTopLevelKeys = map[string]struct{}{
		"featureflags": {},
		"contactform":  {},
		"whatsapp":     {},
		"newsletter":   {},
		"booking":      {},
		"i18n":         {},
		"reports":      {},
	}
	nuvioCMSBackofficeSettingsClientDeniedTopLevelKeys = map[string]struct{}{
		"featureflags": {},
	}
	nuvioCMSBackofficeSettingsFeatureFlagAllowedKeys = map[string]struct{}{
		"whatsapp":    {},
		"contactform": {},
		"reviews":     {},
		"newsletter":  {},
		"booking":     {},
		"reports":     {},
		"i18n":        {},
	}
	nuvioCMSBackofficePageSEOAllowedPayloadKeys = map[string]struct{}{
		"seotitle":              {},
		"seodescription":        {},
		"seosocialimage":        {},
		"seocanonicalurl":       {},
		"seonoindex":            {},
		"seoexcludefromsitemap": {},
		"seofocuskeyword":       {},
		"seotranslations":       {},
	}
	nuvioCMSBackofficeBlockAllowedPayloadKeys = map[string]struct{}{
		"props":        {},
		"translations": {},
	}
	nuvioCMSBackofficeIdentityTextMaxByKey = map[string]int{
		"name":                    160,
		"title":                   160,
		"displayname":             160,
		"slug":                    160,
		"domain":                  255,
		"seotitle":                300,
		"seodescription":          1000,
		"seotitletemplate":        300,
		"seotitleseparator":       20,
		"seocanonicaldomain":      nuvioCMSBackofficeURLMaxLen,
		"businessname":            200,
		"businesstype":            120,
		"businessprimarycategory": 180,
		"businessphone":           40,
		"businessemail":           320,
		"businessaddress":         255,
		"businesscity":            120,
		"businesspostalcode":      40,
		"businesscountry":         120,
		"businessservicearea":     1500,
		"businessopeninghours":    2000,
		"businessgoogleplaceid":   255,
		"businesssocialprofiles":  4096,
		"businesspricerange":      80,
		"logo":                    nuvioCMSBackofficeURLMaxLen,
		"seoimage":                nuvioCMSBackofficeURLMaxLen,
	}
	nuvioCMSBackofficePageSEOTextMaxByKey = map[string]int{
		"seotitle":        300,
		"seodescription":  1000,
		"seofocuskeyword": 255,
	}
	nuvioCMSBackofficeBlockHrefLikePathSegments = map[string]struct{}{
		"url":        {},
		"href":       {},
		"link":       {},
		"linkurl":    {},
		"buttonurl":  {},
		"ctaurl":     {},
		"actionurl":  {},
		"socialurl":  {},
		"profileurl": {},
	}
	nuvioCMSBackofficeBlockAssetLikePathSegments = map[string]struct{}{
		"imageurl":           {},
		"src":                {},
		"image":              {},
		"backgroundimage":    {},
		"backgroundimageurl": {},
	}
	nuvioCMSBackofficeBlockEmbedLikePathSegments = map[string]struct{}{
		"videourl":  {},
		"embedurl":  {},
		"iframeurl": {},
	}
	nuvioCMSBackofficePhoneValuePattern       = regexp.MustCompile(`^[0-9+()./\-\s]+$`)
	nuvioCMSBackofficeI18NLanguageCodePattern = regexp.MustCompile(`^[a-z]{2,3}(?:-[a-z0-9]{2,8}){0,2}$`)
)

type nuvioCMSDashboardWebsiteDTO struct {
	ID                      string         `json:"id"`
	DisplayName             string         `json:"displayName"`
	Name                    string         `json:"name,omitempty"`
	Title                   string         `json:"title,omitempty"`
	Slug                    string         `json:"slug,omitempty"`
	Domain                  string         `json:"domain,omitempty"`
	Logo                    any            `json:"logo,omitempty"`
	SEOTitle                string         `json:"seoTitle,omitempty"`
	SEODescription          string         `json:"seoDescription,omitempty"`
	SEOImage                any            `json:"seoImage,omitempty"`
	SEOTitleTemplate        string         `json:"seo_title_template,omitempty"`
	SEOTitleSeparator       string         `json:"seo_title_separator,omitempty"`
	SEOCanonicalDomain      string         `json:"seo_canonical_domain,omitempty"`
	BusinessName            string         `json:"business_name,omitempty"`
	BusinessType            string         `json:"business_type,omitempty"`
	BusinessPrimaryCategory string         `json:"business_primary_category,omitempty"`
	BusinessPhone           string         `json:"business_phone,omitempty"`
	BusinessEmail           string         `json:"business_email,omitempty"`
	BusinessAddress         string         `json:"business_address,omitempty"`
	BusinessCity            string         `json:"business_city,omitempty"`
	BusinessPostalCode      string         `json:"business_postal_code,omitempty"`
	BusinessCountry         string         `json:"business_country,omitempty"`
	BusinessServiceArea     string         `json:"business_service_area,omitempty"`
	BusinessOpeningHours    string         `json:"business_opening_hours,omitempty"`
	BusinessGooglePlaceID   string         `json:"business_google_place_id,omitempty"`
	BusinessSocialProfiles  string         `json:"business_social_profiles,omitempty"`
	BusinessPriceRange      string         `json:"business_price_range,omitempty"`
	IdentitySEO             map[string]any `json:"identitySeo"`
	Settings                map[string]any `json:"settings"`
}

type nuvioCMSDashboardPageDTO struct {
	ID                    string `json:"id"`
	Website               string `json:"website"`
	Title                 string `json:"title"`
	Name                  string `json:"name"`
	Slug                  string `json:"slug"`
	Path                  string `json:"path"`
	URL                   string `json:"url"`
	Status                string `json:"status"`
	Published             bool   `json:"published"`
	Visible               bool   `json:"visible"`
	SEOTitle              string `json:"seo_title"`
	SEODescription        string `json:"seo_description"`
	SEOSocialImage        any    `json:"seo_social_image"`
	SEOCanonicalURL       string `json:"seo_canonical_url"`
	SEONoindex            bool   `json:"seo_noindex"`
	SEOExcludeFromSitemap bool   `json:"seo_exclude_from_sitemap"`
	SEOFocusKeyword       string `json:"seo_focus_keyword"`
	SEOTranslations       any    `json:"seo_translations"`
	Created               string `json:"created"`
	Updated               string `json:"updated"`
}

type nuvioCMSDashboardBlockDTO struct {
	ID           string `json:"id"`
	Page         string `json:"page"`
	Website      string `json:"website"`
	Component    string `json:"component"`
	ComponentKey string `json:"component_key"`
	Variant      string `json:"variant"`
	Slot         string `json:"slot"`
	DisplayOrder int    `json:"displayOrder"`
	Order        int    `json:"order"`
	Props        any    `json:"props"`
	Translations any    `json:"translations"`
	Enabled      bool   `json:"enabled"`
	Visible      bool   `json:"visible"`
	Status       string `json:"status"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

type nuvioCMSDashboardComponentDTO struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	ComponentKey   string `json:"component_key"`
	Name           string `json:"name"`
	Title          string `json:"title,omitempty"`
	Label          string `json:"label,omitempty"`
	Category       string `json:"category,omitempty"`
	Group          string `json:"group,omitempty"`
	Variant        string `json:"variant,omitempty"`
	DefaultVariant string `json:"defaultVariant,omitempty"`
	Schema         any    `json:"schema"`
}

type nuvioCMSDashboardCapabilities struct {
	CanEditWebsiteIdentitySEO bool `json:"canEditWebsiteIdentitySeo"`
	CanEditWebsiteSettings    bool `json:"canEditWebsiteSettings"`
	CanEditPageSEO            bool `json:"canEditPageSeo"`
	CanEditBlocks             bool `json:"canEditBlocks"`
	CanEditComponents         bool `json:"canEditComponents"`
	CanUseFileFields          bool `json:"canUseFileFields"`
}

type nuvioCMSDashboardResponse struct {
	State        string                          `json:"state"`
	WebsiteID    string                          `json:"websiteId"`
	Website      nuvioCMSDashboardWebsiteDTO     `json:"website"`
	Pages        []nuvioCMSDashboardPageDTO      `json:"pages"`
	Page         *nuvioCMSDashboardPageDTO       `json:"page"`
	Blocks       []nuvioCMSDashboardBlockDTO     `json:"blocks"`
	Components   []nuvioCMSDashboardComponentDTO `json:"components"`
	Capabilities nuvioCMSDashboardCapabilities   `json:"capabilities"`
}

type nuvioCMSBackofficeAssetFileRefDTO struct {
	RecordID   string `json:"recordId"`
	Filename   string `json:"filename"`
	Collection string `json:"collection"`
}

type nuvioCMSBackofficeAssetDTO struct {
	ID           string                             `json:"id"`
	Website      string                             `json:"website,omitempty"`
	Filename     string                             `json:"filename,omitempty"`
	OriginalName string                             `json:"originalName,omitempty"`
	MimeType     string                             `json:"mimeType,omitempty"`
	Size         int64                              `json:"size,omitempty"`
	Created      string                             `json:"created,omitempty"`
	Updated      string                             `json:"updated,omitempty"`
	Collection   string                             `json:"collection,omitempty"`
	File         *nuvioCMSBackofficeAssetFileRefDTO `json:"file,omitempty"`
}

type nuvioCMSBackofficeSEOAssetRef struct {
	Collection string
	RecordID   string
	Filename   string
}

// NUVIO CUSTOM START: Scoped CMS backoffice dashboard endpoint (A3.5.8B).
func registerNuvioCMSBackofficeRoutes(e *core.ServeEvent) {
	cmsGroup := e.Router.Group("/api/nuvio/cms").Bind(apis.RequireSuperuserAuth())

	cmsGroup.GET("/assets", func(e *core.RequestEvent) error {
		websiteID, err := resolveNuvioCMSBackofficeWebsiteIDFromRawInput(e, e.Request.URL.Query().Get("websiteId"))
		if err != nil {
			return err
		}

		assetsCollection, err := findNuvioCMSDashboardCollectionByAliases(e.App, nuvioCMSDashboardAssetsCollectionAliases)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms assets collection resolve failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Failed to load CMS assets.", nil)
		}

		records, err := findNuvioCMSDashboardRecordsByFilter(
			e.App,
			assetsCollection,
			"",
			nil,
			[]string{"-created", "-updated"},
		)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms assets list failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Failed to load CMS assets.", nil)
		}

		items := make([]nuvioCMSBackofficeAssetDTO, 0, len(records))
		for _, record := range records {
			// Access is scoped by the requested website above. Assets are a shared
			// media library inside a per-client Nuvio instance, so legacy/unassigned
			// and other website-linked assets remain selectable after access passes.
			items = append(items, buildNuvioCMSBackofficeAssetDTO(record, assetsCollection))
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":  "ok",
			"assets": items,
		})
	})

	cmsGroup.POST("/assets", func(e *core.RequestEvent) error {
		websiteID, err := resolveNuvioCMSBackofficeWebsiteIDFromRawInput(e, e.Request.FormValue("websiteId"))
		if err != nil {
			return err
		}

		if parseErr := e.Request.ParseMultipartForm(nuvioCMSBackofficeAssetMaxFileSizeBytes + 1024*1024); parseErr != nil {
			return e.BadRequestError("Invalid upload payload.", nil)
		}
		if err := validateNuvioCMSBackofficeAssetMultipartFields(e.Request.MultipartForm); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		uploadedFiles, fileErr := e.FindUploadedFiles("file")
		if fileErr != nil && !errors.Is(fileErr, http.ErrMissingFile) {
			e.App.Logger().Error(
				"NUVIO cms asset upload parse failed",
				"websiteId",
				websiteID,
				"error",
				fileErr.Error(),
			)
			return e.BadRequestError("Invalid upload payload.", nil)
		}
		if len(uploadedFiles) == 0 || uploadedFiles[0] == nil {
			return e.BadRequestError("Missing file.", nil)
		}
		if len(uploadedFiles) > 1 {
			return e.BadRequestError("Only one file is allowed per upload.", nil)
		}

		uploadedFile := uploadedFiles[0]
		detectedMIME, err := validateNuvioCMSBackofficeAssetUpload(uploadedFile)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		checksum, err := computeNuvioCMSBackofficeAssetChecksum(uploadedFile)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms asset checksum failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Invalid upload payload.", nil)
		}

		assetsCollection, err := findNuvioCMSDashboardCollectionByAliases(e.App, nuvioCMSDashboardAssetsCollectionAliases)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms assets collection resolve failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Failed to upload CMS asset.", nil)
		}

		fileFieldName := resolveNuvioCollectionFieldNameByAliases(assetsCollection, []string{"file"})
		if fileFieldName == "" {
			e.App.Logger().Error(
				"NUVIO cms assets file field missing",
				"websiteId",
				websiteID,
				"collection",
				strings.TrimSpace(assetsCollection.Name),
			)
			return e.InternalServerError("Failed to upload CMS asset.", nil)
		}

		assetRecord := core.NewRecord(assetsCollection)
		assetRecord.Set(fileFieldName, uploadedFile)
		setNuvioCMSBackofficeAssetWebsite(assetRecord, assetsCollection, websiteID)
		setNuvioCMSBackofficeAssetMetadata(assetRecord, assetsCollection, uploadedFile, detectedMIME, checksum)

		if saveErr := e.App.Save(assetRecord); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO cms asset upload failed",
				"websiteId",
				websiteID,
				"error",
				saveErr.Error(),
			)

			loweredSaveErr := strings.ToLower(saveErr.Error())
			if strings.Contains(loweredSaveErr, "validation") || strings.Contains(loweredSaveErr, "failed to upload") {
				return e.BadRequestError("Invalid upload payload.", nil)
			}

			return e.InternalServerError("Failed to upload CMS asset.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state": "ok",
			"asset": buildNuvioCMSBackofficeAssetDTO(assetRecord, assetsCollection),
		})
	})

	cmsGroup.GET("/dashboard", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		pageID := strings.TrimSpace(e.Request.URL.Query().Get("pageId"))

		websiteRecord, err := e.App.FindRecordById(nuvioWebsitesCollectionID, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}

			e.App.Logger().Error(
				"NUVIO cms dashboard website load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load CMS dashboard data.", nil)
		}

		pagesCollection, err := findNuvioCMSDashboardCollectionByAliases(e.App, nuvioCMSDashboardPagesCollectionAliases)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms dashboard pages collection resolve failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load CMS dashboard data.", nil)
		}
		blocksCollection, err := findNuvioCMSDashboardCollectionByAliases(e.App, nuvioCMSDashboardBlocksCollectionAliases)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms dashboard blocks collection resolve failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load CMS dashboard data.", nil)
		}
		componentsCollection, err := findNuvioCMSDashboardCollectionByAliases(e.App, nuvioCMSDashboardComponentsCollectionAliases)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms dashboard components collection resolve failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load CMS dashboard data.", nil)
		}

		pageRecords, pageDTOs, err := loadNuvioCMSDashboardPages(e.App, pagesCollection, websiteID)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms dashboard pages load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load CMS dashboard data.", nil)
		}

		selectedPageRecord, selectedPageDTO := resolveNuvioCMSDashboardSelectedPage(pageRecords, pageDTOs, pageID)
		if pageID != "" && selectedPageRecord == nil {
			return e.NotFoundError("Page not found.", nil)
		}

		blocksDTO, err := loadNuvioCMSDashboardBlocks(e.App, blocksCollection, websiteID, selectedPageRecord)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms dashboard blocks load failed",
				"websiteId",
				websiteID,
				"pageId",
				pageID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load CMS dashboard data.", nil)
		}

		componentsDTO, err := loadNuvioCMSDashboardComponents(e.App, componentsCollection)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO cms dashboard components load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load CMS dashboard data.", nil)
		}

		isAdmin := apis.IsAdminSuperuser(e.Auth)
		response := nuvioCMSDashboardResponse{
			State:      "ok",
			WebsiteID:  websiteID,
			Website:    buildNuvioCMSDashboardWebsiteDTO(websiteRecord, isAdmin),
			Pages:      pageDTOs,
			Page:       selectedPageDTO,
			Blocks:     blocksDTO,
			Components: componentsDTO,
			Capabilities: nuvioCMSDashboardCapabilities{
				CanEditWebsiteIdentitySEO: true,
				CanEditWebsiteSettings:    true,
				CanEditPageSEO:            true,
				CanEditBlocks:             true,
				CanEditComponents:         false,
				CanUseFileFields:          false,
			},
		}

		return e.JSON(http.StatusOK, response)
	})

	cmsGroup.PATCH("/websites/{id}/identity", func(e *core.RequestEvent) error {
		websiteRecord, err := resolveNuvioCMSBackofficeWebsiteWriteTarget(e)
		if err != nil {
			return err
		}

		payload, err := parseNuvioCMSBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if len(payload) == 0 {
			return e.BadRequestError("At least one identity field is required.", nil)
		}

		if err := applyNuvioCMSBackofficeIdentityPatch(e.App, websiteRecord, payload, apis.IsAdminSuperuser(e.Auth)); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(websiteRecord); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO cms website identity update failed",
				"websiteId",
				strings.TrimSpace(websiteRecord.Id),
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to update website identity.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":     "ok",
			"websiteId": strings.TrimSpace(websiteRecord.Id),
			"website":   buildNuvioCMSDashboardWebsiteDTO(websiteRecord, apis.IsAdminSuperuser(e.Auth)),
		})
	})

	cmsGroup.PATCH("/websites/{id}/settings", func(e *core.RequestEvent) error {
		websiteRecord, err := resolveNuvioCMSBackofficeWebsiteWriteTarget(e)
		if err != nil {
			return err
		}

		payload, err := parseNuvioCMSBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if len(payload) == 0 {
			return e.BadRequestError("At least one settings field is required.", nil)
		}

		settingsPatch, err := normalizeNuvioCMSBackofficeSettingsPatchPayload(payload)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		currentSettings := parseNuvioSettingsObject(websiteRecord.Get("settings"))
		nextSettings, err := mergeNuvioCMSBackofficeSettingsPatch(
			currentSettings,
			settingsPatch,
			apis.IsAdminSuperuser(e.Auth),
		)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteRecord.Set("settings", nextSettings)
		if saveErr := e.App.Save(websiteRecord); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO cms website settings update failed",
				"websiteId",
				strings.TrimSpace(websiteRecord.Id),
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to update website settings.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state":     "ok",
			"websiteId": strings.TrimSpace(websiteRecord.Id),
			"website":   buildNuvioCMSDashboardWebsiteDTO(websiteRecord, apis.IsAdminSuperuser(e.Auth)),
		})
	})

	cmsGroup.PATCH("/pages/{id}/seo", func(e *core.RequestEvent) error {
		pagesCollection, pageRecord, err := resolveNuvioCMSBackofficePageWriteTarget(e)
		if err != nil {
			return err
		}

		payload, err := parseNuvioCMSBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if len(payload) == 0 {
			return e.BadRequestError("At least one SEO field is required.", nil)
		}

		if err := applyNuvioCMSBackofficePageSEOPatch(e.App, pageRecord, pagesCollection, payload); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(pageRecord); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO cms page seo update failed",
				"pageId",
				strings.TrimSpace(pageRecord.Id),
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to update page SEO.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state": "ok",
			"page":  buildNuvioCMSDashboardPageDTO(pageRecord),
		})
	})

	cmsGroup.PATCH("/blocks/{id}", func(e *core.RequestEvent) error {
		blocksCollection, blockRecord, err := resolveNuvioCMSBackofficeBlockWriteTarget(e)
		if err != nil {
			return err
		}

		payload, err := parseNuvioCMSBackofficePayloadMap(e)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if len(payload) == 0 {
			return e.BadRequestError("At least one block content field is required.", nil)
		}

		if err := applyNuvioCMSBackofficeBlockPatch(blockRecord, blocksCollection, payload); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if saveErr := e.App.Save(blockRecord); saveErr != nil {
			e.App.Logger().Error(
				"NUVIO cms block update failed",
				"blockId",
				strings.TrimSpace(blockRecord.Id),
				"error",
				saveErr.Error(),
			)
			return e.BadRequestError("Failed to update block content.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"state": "ok",
			"block": buildNuvioCMSDashboardBlockDTO(blockRecord),
		})
	})
}

func parseNuvioCMSBackofficePayloadMap(e *core.RequestEvent) (map[string]any, error) {
	payload := map[string]any{}
	if err := e.BindBody(&payload); err != nil {
		return nil, fmt.Errorf("Invalid request payload")
	}

	return payload, nil
}

func normalizeNuvioCMSBackofficePayloadKey(rawKey string) string {
	normalized := strings.ToLower(strings.TrimSpace(rawKey))
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	return normalized
}

func resolveNuvioCMSBackofficeWebsiteWriteTarget(e *core.RequestEvent) (*core.Record, error) {
	websiteID := strings.TrimSpace(e.Request.PathValue("id"))
	if websiteID == "" {
		return nil, e.BadRequestError("Missing website id.", nil)
	}

	if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
		return nil, err
	}

	websiteRecord, err := e.App.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, e.NotFoundError("Website not found.", nil)
		}

		return nil, e.BadRequestError("Failed to load website.", nil)
	}

	return websiteRecord, nil
}

func resolveNuvioCMSBackofficeWebsiteIDFromRawInput(e *core.RequestEvent, rawWebsiteID string) (string, error) {
	websiteID := strings.TrimSpace(rawWebsiteID)
	if websiteID == "" {
		return "", e.BadRequestError("Missing websiteId.", nil)
	}

	if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
		return "", err
	}

	if _, err := e.App.FindRecordById(nuvioWebsitesCollectionID, websiteID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", e.NotFoundError("Website not found.", nil)
		}

		return "", e.InternalServerError("Failed to load website.", nil)
	}

	return websiteID, nil
}

func resolveNuvioCMSBackofficePageWriteTarget(e *core.RequestEvent) (*core.Collection, *core.Record, error) {
	pageID := strings.TrimSpace(e.Request.PathValue("id"))
	if pageID == "" {
		return nil, nil, e.BadRequestError("Missing page id.", nil)
	}

	pagesCollection, err := findNuvioCMSDashboardCollectionByAliases(e.App, nuvioCMSDashboardPagesCollectionAliases)
	if err != nil {
		return nil, nil, e.BadRequestError("Failed to resolve pages collection.", nil)
	}

	pageRecord, err := e.App.FindRecordById(pagesCollection.Id, pageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, e.NotFoundError("Page not found.", nil)
		}

		return nil, nil, e.BadRequestError("Failed to load page.", nil)
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(pageRecord, "website", "site"))
	if websiteID == "" {
		websiteFieldName := resolveNuvioCollectionFieldNameByAliases(pagesCollection, []string{"website", "site"})
		if websiteFieldName != "" {
			websiteID = strings.TrimSpace(parseStringValue(pageRecord.Get(websiteFieldName)))
		}
	}
	if websiteID == "" {
		return nil, nil, e.ForbiddenError("The authorized record is not allowed to perform this action.", nil)
	}

	if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
		return nil, nil, err
	}

	return pagesCollection, pageRecord, nil
}

func resolveNuvioCMSBackofficeBlockWriteTarget(e *core.RequestEvent) (*core.Collection, *core.Record, error) {
	blockID := strings.TrimSpace(e.Request.PathValue("id"))
	if blockID == "" {
		return nil, nil, e.BadRequestError("Missing block id.", nil)
	}

	blocksCollection, err := findNuvioCMSDashboardCollectionByAliases(e.App, nuvioCMSDashboardBlocksCollectionAliases)
	if err != nil {
		return nil, nil, e.BadRequestError("Failed to resolve blocks collection.", nil)
	}
	pagesCollection, err := findNuvioCMSDashboardCollectionByAliases(e.App, nuvioCMSDashboardPagesCollectionAliases)
	if err != nil {
		return nil, nil, e.BadRequestError("Failed to resolve pages collection.", nil)
	}

	blockRecord, err := e.App.FindRecordById(blocksCollection.Id, blockID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, e.NotFoundError("Block not found.", nil)
		}

		return nil, nil, e.BadRequestError("Failed to load block.", nil)
	}

	pageID := strings.TrimSpace(resolveNuvioPublicRelationID(blockRecord, "page"))
	if pageID == "" {
		pageFieldName := resolveNuvioCollectionFieldNameByAliases(blocksCollection, []string{"page"})
		if pageFieldName != "" {
			pageID = strings.TrimSpace(parseStringValue(blockRecord.Get(pageFieldName)))
		}
	}
	if pageID == "" {
		return nil, nil, e.NotFoundError("Block page not found.", nil)
	}

	pageRecord, err := e.App.FindRecordById(pagesCollection.Id, pageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, e.NotFoundError("Block page not found.", nil)
		}
		return nil, nil, e.BadRequestError("Failed to load block page.", nil)
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(pageRecord, "website", "site"))
	if websiteID == "" {
		websiteFieldName := resolveNuvioCollectionFieldNameByAliases(pagesCollection, []string{"website", "site"})
		if websiteFieldName != "" {
			websiteID = strings.TrimSpace(parseStringValue(pageRecord.Get(websiteFieldName)))
		}
	}
	if websiteID == "" {
		return nil, nil, e.NotFoundError("Block website not found.", nil)
	}

	if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
		return nil, nil, err
	}

	return blocksCollection, blockRecord, nil
}

func applyNuvioCMSBackofficeBlockPatch(
	blockRecord *core.Record,
	blocksCollection *core.Collection,
	payload map[string]any,
) error {
	if blockRecord == nil || blocksCollection == nil {
		return fmt.Errorf("Block not found")
	}

	updatedFields := 0
	for rawKey, rawValue := range payload {
		normalizedKey := normalizeNuvioCMSBackofficePayloadKey(rawKey)
		if normalizedKey == "" {
			return fmt.Errorf("Invalid payload field")
		}
		if _, ok := nuvioCMSBackofficeBlockAllowedPayloadKeys[normalizedKey]; !ok {
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}

		switch normalizedKey {
		case "props":
			propsValue, err := normalizeNuvioCMSBackofficeBlockPropsValue(rawValue)
			if err != nil {
				return err
			}
			if err := setNuvioCMSBackofficeBlockValueField(blockRecord, blocksCollection, []string{"props"}, propsValue); err != nil {
				return err
			}
			updatedFields++
		case "translations":
			translationsValue, err := normalizeNuvioCMSBackofficeBlockTranslationsValue(rawValue)
			if err != nil {
				return err
			}
			if err := setNuvioCMSBackofficeBlockValueField(blockRecord, blocksCollection, []string{"translations"}, translationsValue); err != nil {
				return err
			}
			updatedFields++
		}
	}

	if updatedFields == 0 {
		return fmt.Errorf("At least one block content field is required")
	}

	return nil
}

func setNuvioCMSBackofficeBlockValueField(
	blockRecord *core.Record,
	blocksCollection *core.Collection,
	aliases []string,
	value any,
) error {
	fieldName := resolveNuvioCollectionFieldNameByAliases(blocksCollection, aliases)
	if fieldName == "" {
		return fmt.Errorf("Field %q is not available for this blocks collection", aliases[0])
	}

	blockRecord.Set(fieldName, value)
	return nil
}

func normalizeNuvioCMSBackofficeBlockPropsValue(rawValue any) (map[string]any, error) {
	if rawValue == nil {
		return map[string]any{}, nil
	}

	propsMap, ok := toStringAnyMap(rawValue)
	if !ok {
		return nil, fmt.Errorf("props must be an object")
	}

	if containsNuvioCMSBackofficeFileLikePayload(propsMap) {
		return nil, fmt.Errorf("File upload payloads are not supported in this endpoint yet")
	}
	if err := validateNuvioCMSBackofficeBlockContentValue(propsMap, []string{"props"}); err != nil {
		return nil, err
	}

	return propsMap, nil
}

func normalizeNuvioCMSBackofficeBlockTranslationsValue(rawValue any) (any, error) {
	if rawValue == nil {
		return map[string]any{}, nil
	}

	if translationMap, ok := toStringAnyMap(rawValue); ok {
		if containsNuvioCMSBackofficeFileLikePayload(translationMap) {
			return nil, fmt.Errorf("File upload payloads are not supported in this endpoint yet")
		}
		if err := validateNuvioCMSBackofficeBlockContentValue(translationMap, []string{"translations"}); err != nil {
			return nil, err
		}
		return translationMap, nil
	}

	switch typed := rawValue.(type) {
	case []any:
		if containsNuvioCMSBackofficeFileLikePayload(typed) {
			return nil, fmt.Errorf("File upload payloads are not supported in this endpoint yet")
		}
		if err := validateNuvioCMSBackofficeBlockContentValue(typed, []string{"translations"}); err != nil {
			return nil, err
		}
		return typed, nil
	case string:
		if strings.TrimSpace(typed) == "" {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("translations must be an object or array")
	default:
		return nil, fmt.Errorf("translations must be an object or array")
	}
}

func validateNuvioCMSBackofficeBlockContentValue(value any, path []string) error {
	switch typed := value.(type) {
	case map[string]any:
		for rawKey, rawValue := range typed {
			normalizedKey := normalizeNuvioCMSBackofficePayloadKey(rawKey)
			if normalizedKey == "" {
				normalizedKey = strings.ToLower(strings.TrimSpace(rawKey))
			}

			childPath := appendNuvioCMSBackofficeBlockPathSegment(path, normalizedKey)
			if err := validateNuvioCMSBackofficeBlockContentValue(rawValue, childPath); err != nil {
				return err
			}
		}
	case []any:
		for index, item := range typed {
			childPath := appendNuvioCMSBackofficeBlockPathSegment(path, fmt.Sprintf("[%d]", index))
			if err := validateNuvioCMSBackofficeBlockContentValue(item, childPath); err != nil {
				return err
			}
		}
	case string:
		if err := validateNuvioCMSBackofficeBlockStringValue(path, typed); err != nil {
			return err
		}
	}

	return nil
}

func appendNuvioCMSBackofficeBlockPathSegment(path []string, segment string) []string {
	if segment == "" {
		return append([]string{}, path...)
	}

	nextPath := make([]string, 0, len(path)+1)
	nextPath = append(nextPath, path...)
	nextPath = append(nextPath, segment)
	return nextPath
}

func formatNuvioCMSBackofficeBlockPath(path []string) string {
	if len(path) == 0 {
		return "value"
	}

	var builder strings.Builder
	for _, segment := range path {
		if strings.TrimSpace(segment) == "" {
			continue
		}

		if strings.HasPrefix(segment, "[") && strings.HasSuffix(segment, "]") {
			builder.WriteString(segment)
			continue
		}

		if builder.Len() > 0 {
			builder.WriteString(".")
		}
		builder.WriteString(segment)
	}

	if builder.Len() == 0 {
		return "value"
	}
	return builder.String()
}

func validateNuvioCMSBackofficeBlockStringValue(path []string, rawValue string) error {
	if len([]rune(rawValue)) > nuvioCMSBackofficeBlockStringMaxLen {
		return fmt.Errorf("%s exceeds the maximum length of %d characters", formatNuvioCMSBackofficeBlockPath(path), nuvioCMSBackofficeBlockStringMaxLen)
	}

	trimmed := strings.TrimSpace(rawValue)
	if trimmed == "" {
		return nil
	}

	if len([]rune(trimmed)) > nuvioCMSBackofficeURLMaxLen && isNuvioCMSBackofficeBlockURLLikePath(path) {
		return fmt.Errorf("%s exceeds the maximum length of %d characters", formatNuvioCMSBackofficeBlockPath(path), nuvioCMSBackofficeURLMaxLen)
	}

	switch classifyNuvioCMSBackofficeBlockURLPath(path) {
	case "href":
		if err := validateNuvioCMSBackofficeBlockHrefLikeURL(trimmed); err != nil {
			return fmt.Errorf("%s must be a safe link URL", formatNuvioCMSBackofficeBlockPath(path))
		}
	case "asset":
		if err := validateNuvioCMSBackofficeBlockAssetLikeURL(trimmed); err != nil {
			return fmt.Errorf("%s must be a safe image/source URL", formatNuvioCMSBackofficeBlockPath(path))
		}
	case "embed":
		if err := validateNuvioCMSBackofficeBlockEmbedLikeURL(trimmed); err != nil {
			return fmt.Errorf("%s must be a safe embed URL", formatNuvioCMSBackofficeBlockPath(path))
		}
	}

	return nil
}

func isNuvioCMSBackofficeBlockURLLikePath(path []string) bool {
	return classifyNuvioCMSBackofficeBlockURLPath(path) != ""
}

func classifyNuvioCMSBackofficeBlockURLPath(path []string) string {
	for index := len(path) - 1; index >= 0; index-- {
		segment := strings.TrimSpace(path[index])
		if segment == "" || strings.HasPrefix(segment, "[") {
			continue
		}

		normalizedSegment := normalizeNuvioCMSBackofficePayloadKey(segment)
		if normalizedSegment == "" || normalizedSegment == "props" || normalizedSegment == "translations" {
			continue
		}

		if _, ok := nuvioCMSBackofficeBlockEmbedLikePathSegments[normalizedSegment]; ok {
			return "embed"
		}
		if _, ok := nuvioCMSBackofficeBlockAssetLikePathSegments[normalizedSegment]; ok {
			return "asset"
		}
		if _, ok := nuvioCMSBackofficeBlockHrefLikePathSegments[normalizedSegment]; ok {
			return "href"
		}

		if strings.HasSuffix(normalizedSegment, "url") {
			switch {
			case strings.Contains(normalizedSegment, "embed"),
				strings.Contains(normalizedSegment, "iframe"),
				strings.Contains(normalizedSegment, "video"):
				return "embed"
			case strings.Contains(normalizedSegment, "image"),
				strings.Contains(normalizedSegment, "background"),
				strings.Contains(normalizedSegment, "src"):
				return "asset"
			default:
				return "href"
			}
		}
	}

	return ""
}

func validateNuvioCMSBackofficeBlockHrefLikeURL(value string) error {
	if len(value) > nuvioCMSBackofficeURLMaxLen {
		return fmt.Errorf("value exceeds max URL length")
	}
	if hasNuvioCMSBackofficeBlockedURLScheme(value) || strings.HasPrefix(value, "//") || hasNuvioCMSBackofficeUnsafeURLChars(value) {
		return fmt.Errorf("invalid URL")
	}

	loweredValue := strings.ToLower(value)
	switch {
	case strings.HasPrefix(value, "/"),
		strings.HasPrefix(value, "#"),
		strings.HasPrefix(value, "?"):
		return nil
	case strings.HasPrefix(loweredValue, "mailto:"):
		return validateNuvioCMSBackofficeMailtoURL(value)
	case strings.HasPrefix(loweredValue, "tel:"):
		return validateNuvioCMSBackofficeTelURL(value)
	case strings.Contains(value, "://"):
		return validateNuvioCMSBackofficeAbsoluteHTTPURL(value)
	default:
		return fmt.Errorf("invalid URL")
	}
}

func validateNuvioCMSBackofficeMailtoURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return fmt.Errorf("invalid mailto URL")
	}
	if strings.ToLower(strings.TrimSpace(parsed.Scheme)) != "mailto" {
		return fmt.Errorf("invalid mailto URL")
	}

	address := strings.TrimSpace(parsed.Opaque)
	if address == "" {
		address = strings.TrimSpace(parsed.Path)
	}
	if address == "" || strings.ContainsAny(address, " \t\r\n<>\"'`") {
		return fmt.Errorf("invalid mailto URL")
	}
	return nil
}

func validateNuvioCMSBackofficeTelURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return fmt.Errorf("invalid tel URL")
	}
	if strings.ToLower(strings.TrimSpace(parsed.Scheme)) != "tel" {
		return fmt.Errorf("invalid tel URL")
	}

	number := strings.TrimSpace(parsed.Opaque)
	if number == "" {
		number = strings.TrimSpace(parsed.Path)
	}
	if number == "" || !nuvioCMSBackofficePhoneValuePattern.MatchString(number) {
		return fmt.Errorf("invalid tel URL")
	}

	return nil
}

func validateNuvioCMSBackofficeBlockAssetLikeURL(value string) error {
	if len(value) > nuvioCMSBackofficeURLMaxLen {
		return fmt.Errorf("value exceeds max URL length")
	}
	if hasNuvioCMSBackofficeBlockedURLScheme(value) || strings.HasPrefix(value, "//") || hasNuvioCMSBackofficeUnsafeURLChars(value) {
		return fmt.Errorf("invalid URL")
	}

	if strings.HasPrefix(value, "/") {
		return nil
	}
	if strings.Contains(value, "://") {
		return validateNuvioCMSBackofficeAbsoluteHTTPURL(value)
	}

	if strings.Contains(value, "\\") || strings.Contains(value, "..") || strings.Contains(value, ":") {
		return fmt.Errorf("invalid URL")
	}

	return nil
}

func validateNuvioCMSBackofficeBlockEmbedLikeURL(value string) error {
	if len(value) > nuvioCMSBackofficeURLMaxLen {
		return fmt.Errorf("value exceeds max URL length")
	}
	if hasNuvioCMSBackofficeBlockedURLScheme(value) || strings.HasPrefix(value, "//") || hasNuvioCMSBackofficeUnsafeURLChars(value) {
		return fmt.Errorf("invalid URL")
	}
	if !strings.Contains(value, "://") {
		return fmt.Errorf("invalid URL")
	}

	return validateNuvioCMSBackofficeAbsoluteHTTPURL(value)
}

func containsNuvioCMSBackofficeFileLikePayload(value any) bool {
	switch typed := value.(type) {
	case []any:
		for _, item := range typed {
			if containsNuvioCMSBackofficeFileLikePayload(item) {
				return true
			}
		}
	case map[string]any:
		if isNuvioCMSBackofficeFileLikeObject(typed) {
			return true
		}

		for _, entryValue := range typed {
			if containsNuvioCMSBackofficeFileLikePayload(entryValue) {
				return true
			}
		}
	}

	return false
}

func isNuvioCMSBackofficeFileLikeObject(value map[string]any) bool {
	if len(value) == 0 {
		return false
	}

	normalizedKeys := map[string]struct{}{}
	for key := range value {
		normalized := normalizeNuvioCMSBackofficePayloadKey(key)
		if normalized != "" {
			normalizedKeys[normalized] = struct{}{}
		}
	}

	_, hasName := normalizedKeys["name"]
	_, hasSize := normalizedKeys["size"]
	_, hasType := normalizedKeys["type"]
	_, hasLastModified := normalizedKeys["lastmodified"]
	_, hasWebkitRelativePath := normalizedKeys["webkitrelativepath"]
	_, hasOriginFileObj := normalizedKeys["originfileobj"]
	_, hasRawFile := normalizedKeys["rawfile"]
	_, hasTempFile := normalizedKeys["tempfile"]
	_, hasArrayBuffer := normalizedKeys["arraybuffer"]
	_, hasBlob := normalizedKeys["blob"]

	if hasLastModified || hasWebkitRelativePath || hasOriginFileObj || hasRawFile || hasTempFile || hasArrayBuffer || hasBlob {
		return true
	}
	if hasName && (hasSize || hasType) {
		return true
	}

	return false
}

func applyNuvioCMSBackofficePageSEOPatch(
	app core.App,
	pageRecord *core.Record,
	pagesCollection *core.Collection,
	payload map[string]any,
) error {
	if pageRecord == nil || pagesCollection == nil {
		return fmt.Errorf("Page not found")
	}

	updatedFields := 0
	for rawKey, rawValue := range payload {
		normalizedKey := normalizeNuvioCMSBackofficePayloadKey(rawKey)
		if normalizedKey == "" {
			return fmt.Errorf("Invalid payload field")
		}
		if _, ok := nuvioCMSBackofficePageSEOAllowedPayloadKeys[normalizedKey]; !ok {
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}

		switch normalizedKey {
		case "seotitle":
			stringValue := strings.TrimSpace(parseStringValue(rawValue))
			if err := validateNuvioCMSBackofficeLimitedTextField("seo_title", stringValue, nuvioCMSBackofficePageSEOTextMaxByKey["seotitle"]); err != nil {
				return err
			}
			if err := setNuvioCMSBackofficePageSEOStringField(pageRecord, pagesCollection, []string{"seo_title", "seoTitle"}, stringValue); err != nil {
				return err
			}
			updatedFields++
		case "seodescription":
			stringValue := strings.TrimSpace(parseStringValue(rawValue))
			if err := validateNuvioCMSBackofficeLimitedTextField("seo_description", stringValue, nuvioCMSBackofficePageSEOTextMaxByKey["seodescription"]); err != nil {
				return err
			}
			if err := setNuvioCMSBackofficePageSEOStringField(pageRecord, pagesCollection, []string{"seo_description", "seoDescription"}, stringValue); err != nil {
				return err
			}
			updatedFields++
		case "seosocialimage":
			if err := setNuvioCMSBackofficePageSEOSocialImageField(app, pageRecord, pagesCollection, rawValue); err != nil {
				return err
			}
			updatedFields++
		case "seocanonicalurl":
			stringValue := strings.TrimSpace(parseStringValue(rawValue))
			if err := validateNuvioCMSBackofficePageCanonicalURL(stringValue); err != nil {
				return err
			}
			if err := setNuvioCMSBackofficePageSEOStringField(pageRecord, pagesCollection, []string{"seo_canonical_url", "seoCanonicalUrl"}, stringValue); err != nil {
				return err
			}
			updatedFields++
		case "seonoindex":
			boolValue, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("seo_noindex must be a boolean")
			}
			if err := setNuvioCMSBackofficePageSEOValueField(pageRecord, pagesCollection, []string{"seo_noindex", "seoNoindex"}, boolValue); err != nil {
				return err
			}
			updatedFields++
		case "seoexcludefromsitemap":
			boolValue, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("seo_exclude_from_sitemap must be a boolean")
			}
			if err := setNuvioCMSBackofficePageSEOValueField(pageRecord, pagesCollection, []string{"seo_exclude_from_sitemap", "seoExcludeFromSitemap"}, boolValue); err != nil {
				return err
			}
			updatedFields++
		case "seofocuskeyword":
			stringValue := strings.TrimSpace(parseStringValue(rawValue))
			if err := validateNuvioCMSBackofficeLimitedTextField("seo_focus_keyword", stringValue, nuvioCMSBackofficePageSEOTextMaxByKey["seofocuskeyword"]); err != nil {
				return err
			}
			if err := setNuvioCMSBackofficePageSEOStringField(pageRecord, pagesCollection, []string{"seo_focus_keyword", "seoFocusKeyword"}, stringValue); err != nil {
				return err
			}
			updatedFields++
		case "seotranslations":
			normalizedTranslations, err := normalizeNuvioCMSBackofficeSEOTranslationsValue(rawValue)
			if err != nil {
				return err
			}
			if err := validateNuvioCMSBackofficeSEOTranslationsTextLimits(normalizedTranslations); err != nil {
				return err
			}
			if err := setNuvioCMSBackofficePageSEOValueField(pageRecord, pagesCollection, []string{"seo_translations", "seoTranslations"}, normalizedTranslations); err != nil {
				return err
			}
			updatedFields++
		}
	}

	if updatedFields == 0 {
		return fmt.Errorf("At least one SEO field is required")
	}

	return nil
}

func setNuvioCMSBackofficePageSEOStringField(
	pageRecord *core.Record,
	pagesCollection *core.Collection,
	aliases []string,
	value string,
) error {
	return setNuvioCMSBackofficePageSEOValueField(pageRecord, pagesCollection, aliases, value)
}

func setNuvioCMSBackofficePageSEOValueField(
	pageRecord *core.Record,
	pagesCollection *core.Collection,
	aliases []string,
	value any,
) error {
	fieldName := resolveNuvioCollectionFieldNameByAliases(pagesCollection, aliases)
	if fieldName == "" {
		return fmt.Errorf("Field %q is not available for this pages collection", aliases[0])
	}

	pageRecord.Set(fieldName, value)
	return nil
}

func setNuvioCMSBackofficePageSEOSocialImageField(
	app core.App,
	pageRecord *core.Record,
	pagesCollection *core.Collection,
	rawValue any,
) error {
	fieldName := resolveNuvioCollectionFieldNameByAliases(pagesCollection, []string{"seo_social_image", "seoSocialImage"})
	if fieldName == "" {
		return fmt.Errorf("Field %q is not available for this pages collection", "seo_social_image")
	}

	field := pagesCollection.Fields.GetByName(fieldName)

	if isNuvioCMSBackofficeEmptyAssetValue(rawValue) {
		pageRecord.Set(fieldName, "")
		return nil
	}

	if assetRef, ok, err := parseNuvioCMSBackofficeSEOAssetRef(rawValue); err != nil {
		return err
	} else if ok {
		if field != nil && field.Type() == core.FieldTypeFile {
			return setNuvioCMSBackofficeFileFieldFromAssetRef(app, pageRecord, fieldName, assetRef)
		}

		if _, err := resolveNuvioCMSBackofficeAssetRecordFromRef(app, assetRef); err != nil {
			return err
		}
		encoded, encodeErr := encodeNuvioCMSBackofficeSEOAssetRef(assetRef)
		if encodeErr != nil {
			return fmt.Errorf("seo_social_image expects a valid asset reference")
		}
		pageRecord.Set(fieldName, encoded)
		return nil
	}

	if field != nil && field.Type() == core.FieldTypeFile {
		return fmt.Errorf("seo_social_image expects an asset reference")
	}

	stringValue, ok := rawValue.(string)
	if !ok {
		return fmt.Errorf("seo_social_image expects a string value or asset reference")
	}
	stringValue = strings.TrimSpace(stringValue)
	if err := validateNuvioCMSBackofficePageSEOSocialImage(stringValue); err != nil {
		return err
	}
	pageRecord.Set(fieldName, stringValue)
	return nil
}

func setNuvioCMSBackofficeWebsiteSEOAssetField(
	app core.App,
	websiteRecord *core.Record,
	aliases []string,
	rawValue any,
	fieldLabel string,
) error {
	if websiteRecord == nil {
		return fmt.Errorf("Website not found")
	}

	collection := websiteRecord.Collection()
	if collection == nil {
		return fmt.Errorf("%s field is not available for this website collection", fieldLabel)
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, aliases)
	if fieldName == "" {
		return fmt.Errorf("%s field is not available for this website collection", fieldLabel)
	}

	field := collection.Fields.GetByName(fieldName)
	if isNuvioCMSBackofficeEmptyAssetValue(rawValue) {
		websiteRecord.Set(fieldName, "")
		return nil
	}

	if assetRef, ok, err := parseNuvioCMSBackofficeSEOAssetRef(rawValue); err != nil {
		return err
	} else if ok {
		if field != nil && field.Type() == core.FieldTypeFile {
			return setNuvioCMSBackofficeFileFieldFromAssetRef(app, websiteRecord, fieldName, assetRef)
		}

		if _, err := resolveNuvioCMSBackofficeAssetRecordFromRef(app, assetRef); err != nil {
			return err
		}
		encoded, encodeErr := encodeNuvioCMSBackofficeSEOAssetRef(assetRef)
		if encodeErr != nil {
			return fmt.Errorf("%s expects a valid asset reference", fieldLabel)
		}
		websiteRecord.Set(fieldName, encoded)
		return nil
	}

	if field != nil && field.Type() == core.FieldTypeFile {
		return fmt.Errorf("%s expects an asset reference", fieldLabel)
	}

	stringValue, ok := rawValue.(string)
	if !ok {
		return fmt.Errorf("%s expects a string value or asset reference", fieldLabel)
	}
	stringValue = strings.TrimSpace(stringValue)
	if err := validateNuvioCMSBackofficePageSEOSocialImage(stringValue); err != nil {
		return err
	}
	websiteRecord.Set(fieldName, stringValue)
	return nil
}

func isNuvioCMSBackofficeEmptyAssetValue(value any) bool {
	if value == nil {
		return true
	}
	if raw, ok := value.(string); ok {
		return strings.TrimSpace(raw) == ""
	}
	return false
}

func parseNuvioCMSBackofficeSEOAssetRef(value any) (nuvioCMSBackofficeSEOAssetRef, bool, error) {
	valueMap, ok := toStringAnyMap(value)
	if !ok {
		return nuvioCMSBackofficeSEOAssetRef{}, false, nil
	}

	if containsNuvioCMSBackofficeFileLikePayload(valueMap) {
		return nuvioCMSBackofficeSEOAssetRef{}, false, fmt.Errorf("SEO image expects an asset reference, not a file payload")
	}

	ref := nuvioCMSBackofficeSEOAssetRef{
		Collection: strings.TrimSpace(parseStringValue(valueMap["collection"])),
		RecordID:   strings.TrimSpace(parseStringValue(valueMap["recordId"])),
		Filename:   strings.TrimSpace(parseStringValue(valueMap["filename"])),
	}
	if ref.RecordID == "" {
		ref.RecordID = strings.TrimSpace(parseStringValue(valueMap["id"]))
	}
	if ref.Filename == "" {
		ref.Filename = strings.TrimSpace(parseStringValue(valueMap["file"]))
	}
	if ref.Collection == "" {
		ref.Collection = strings.TrimSpace(parseStringValue(valueMap["collectionName"]))
	}
	if ref.Collection == "" && toStringAnyMapValue(valueMap["file"]) != nil {
		fileMap := toStringAnyMapValue(valueMap["file"])
		ref.Collection = strings.TrimSpace(parseStringValue(fileMap["collection"]))
		ref.RecordID = firstNonEmptyNuvioCMSBackofficeString(ref.RecordID, parseStringValue(fileMap["recordId"]))
		ref.Filename = firstNonEmptyNuvioCMSBackofficeString(ref.Filename, parseStringValue(fileMap["filename"]))
	}

	if ref.Collection == "" {
		ref.Collection = "Assets"
	}

	if ref.RecordID == "" || ref.Filename == "" {
		return nuvioCMSBackofficeSEOAssetRef{}, false, fmt.Errorf("SEO image asset reference is incomplete")
	}

	return ref, true, nil
}

func toStringAnyMapValue(value any) map[string]any {
	result, _ := toStringAnyMap(value)
	return result
}

func firstNonEmptyNuvioCMSBackofficeString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func encodeNuvioCMSBackofficeSEOAssetRef(ref nuvioCMSBackofficeSEOAssetRef) (string, error) {
	payload := map[string]string{
		"collection": strings.TrimSpace(ref.Collection),
		"recordId":   strings.TrimSpace(ref.RecordID),
		"filename":   strings.TrimSpace(ref.Filename),
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func setNuvioCMSBackofficeFileFieldFromAssetRef(
	app core.App,
	targetRecord *core.Record,
	targetFieldName string,
	ref nuvioCMSBackofficeSEOAssetRef,
) error {
	if app == nil || targetRecord == nil {
		return fmt.Errorf("SEO image asset reference could not be resolved")
	}

	assetRecord, err := resolveNuvioCMSBackofficeAssetRecordFromRef(app, ref)
	if err != nil {
		return err
	}

	fsys, err := app.NewFilesystem()
	if err != nil {
		return fmt.Errorf("SEO image asset reference could not be resolved")
	}
	defer fsys.Close()

	reader, err := fsys.GetReader(assetRecord.BaseFilesPath() + "/" + strings.TrimSpace(ref.Filename))
	if err != nil {
		return fmt.Errorf("SEO image asset reference could not be loaded")
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("SEO image asset reference could not be loaded")
	}

	file, err := filesystem.NewFileFromBytes(content, strings.TrimSpace(ref.Filename))
	if err != nil {
		return fmt.Errorf("SEO image asset reference could not be loaded")
	}
	file.Name = strings.TrimSpace(ref.Filename)

	targetRecord.Set(targetFieldName, file)
	return nil
}

func resolveNuvioCMSBackofficeAssetRecordFromRef(
	app core.App,
	ref nuvioCMSBackofficeSEOAssetRef,
) (*core.Record, error) {
	if app == nil {
		return nil, fmt.Errorf("SEO image asset reference could not be resolved")
	}

	assetsCollection, err := findNuvioCMSDashboardCollectionByAliases(app, nuvioCMSDashboardAssetsCollectionAliases)
	if err != nil {
		return nil, fmt.Errorf("SEO image asset reference could not be resolved")
	}

	refCollection := strings.TrimSpace(ref.Collection)
	if refCollection != "" && refCollection != assetsCollection.Name && refCollection != assetsCollection.Id {
		return nil, fmt.Errorf("SEO image asset reference must point to Assets")
	}

	assetRecord, err := app.FindRecordById(assetsCollection.Id, strings.TrimSpace(ref.RecordID))
	if err != nil {
		return nil, fmt.Errorf("SEO image asset reference was not found")
	}

	fileFieldName := resolveNuvioCollectionFieldNameByAliases(assetsCollection, []string{"file"})
	if fileFieldName == "" {
		return nil, fmt.Errorf("SEO image asset reference could not be resolved")
	}

	assetFilename := toNuvioCMSDashboardSingleFileName(assetRecord.Get(fileFieldName))
	if assetFilename == "" || assetFilename != strings.TrimSpace(ref.Filename) {
		return nil, fmt.Errorf("SEO image asset reference does not match an existing asset file")
	}

	return assetRecord, nil
}

func normalizeNuvioCMSBackofficeSEOTranslationsValue(rawValue any) (any, error) {
	if rawValue == nil {
		return map[string]any{}, nil
	}

	if parsedMap, ok := toStringAnyMap(rawValue); ok {
		return parsedMap, nil
	}

	switch typed := rawValue.(type) {
	case []any:
		return typed, nil
	case string:
		if strings.TrimSpace(typed) == "" {
			return map[string]any{}, nil
		}
		return nil, fmt.Errorf("seo_translations must be an object or array")
	default:
		return nil, fmt.Errorf("seo_translations must be an object or array")
	}
}

func validateNuvioCMSBackofficeIdentityFieldValue(normalizedKey string, rawValue any) (string, error) {
	stringValue := strings.TrimSpace(parseStringValue(rawValue))

	if maxLen, exists := nuvioCMSBackofficeIdentityTextMaxByKey[normalizedKey]; exists {
		if err := validateNuvioCMSBackofficeLimitedTextField(normalizedKey, stringValue, maxLen); err != nil {
			return "", err
		}
	}

	switch normalizedKey {
	case "seocanonicaldomain":
		if err := validateNuvioCMSBackofficeCanonicalDomainValue(stringValue); err != nil {
			return "", err
		}
	case "logo", "seoimage":
		if containsNuvioCMSBackofficeFileLikePayload(rawValue) {
			return "", fmt.Errorf("%s expects a string value", normalizedKey)
		}
		if err := validateNuvioCMSBackofficePageSEOSocialImage(stringValue); err != nil {
			return "", err
		}
	case "businesssocialprofiles":
		if err := validateNuvioCMSBackofficeBusinessSocialProfilesValue(stringValue); err != nil {
			return "", err
		}
	case "businessemail":
		if err := validateNuvioCMSBackofficeBusinessEmailValue(stringValue); err != nil {
			return "", err
		}
	case "businessphone":
		if err := validateNuvioCMSBackofficeBusinessPhoneValue(stringValue); err != nil {
			return "", err
		}
	}

	return stringValue, nil
}

func validateNuvioCMSBackofficeLimitedTextField(fieldName string, value string, maxLen int) error {
	if maxLen <= 0 {
		return nil
	}

	if len([]rune(value)) > maxLen {
		return fmt.Errorf("%s exceeds the maximum length of %d characters", fieldName, maxLen)
	}

	return nil
}

func validateNuvioCMSBackofficeBusinessEmailValue(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	if _, ok := normalizeNuvioEmail(value); !ok {
		return fmt.Errorf("business_email must be a valid email")
	}

	return nil
}

func validateNuvioCMSBackofficeBusinessPhoneValue(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	if !nuvioCMSBackofficePhoneValuePattern.MatchString(trimmed) {
		return fmt.Errorf("business_phone contains invalid characters")
	}

	hasDigit := false
	for _, char := range trimmed {
		if char >= '0' && char <= '9' {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return fmt.Errorf("business_phone must contain at least one digit")
	}

	return nil
}

func validateNuvioCMSBackofficeBusinessSocialProfilesValue(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	entries := parseNuvioCMSBackofficeStringList(trimmed)
	for _, entry := range entries {
		if err := validateNuvioCMSBackofficeAbsoluteHTTPURL(entry); err != nil {
			return fmt.Errorf("business_social_profiles contains an invalid URL")
		}
	}

	return nil
}

func parseNuvioCMSBackofficeStringList(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}

	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		parsed := []any{}
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			values := make([]string, 0, len(parsed))
			for _, rawValue := range parsed {
				parsedValue := strings.TrimSpace(parseStringValue(rawValue))
				if parsedValue != "" {
					values = append(values, parsedValue)
				}
			}
			if len(values) > 0 {
				return values
			}
		}
	}

	pieces := strings.FieldsFunc(trimmed, func(char rune) bool {
		return char == '\n' || char == '\r' || char == ',' || char == ';'
	})
	values := make([]string, 0, len(pieces))
	for _, rawValue := range pieces {
		parsedValue := strings.TrimSpace(rawValue)
		if parsedValue != "" {
			values = append(values, parsedValue)
		}
	}
	if len(values) > 0 {
		return values
	}

	return []string{trimmed}
}

func validateNuvioCMSBackofficeCanonicalDomainValue(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	if len(trimmed) > nuvioCMSBackofficeURLMaxLen {
		return fmt.Errorf("seo_canonical_domain exceeds the maximum length of %d characters", nuvioCMSBackofficeURLMaxLen)
	}

	if hasNuvioCMSBackofficeBlockedURLScheme(trimmed) {
		return fmt.Errorf("seo_canonical_domain must be a valid http(s) URL or domain")
	}
	if strings.HasPrefix(trimmed, "//") {
		return fmt.Errorf("seo_canonical_domain must be a valid http(s) URL or domain")
	}
	if hasNuvioCMSBackofficeUnsafeURLChars(trimmed) {
		return fmt.Errorf("seo_canonical_domain must be a valid http(s) URL or domain")
	}

	if strings.Contains(trimmed, "://") {
		if err := validateNuvioCMSBackofficeAbsoluteHTTPURL(trimmed); err != nil {
			return fmt.Errorf("seo_canonical_domain must be a valid http(s) URL or domain")
		}
		return nil
	}

	if err := validateNuvioCMSBackofficeDomainBaseValue(trimmed); err != nil {
		return fmt.Errorf("seo_canonical_domain must be a valid http(s) URL or domain")
	}

	return nil
}

func validateNuvioCMSBackofficePageCanonicalURL(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	if len(trimmed) > nuvioCMSBackofficeURLMaxLen {
		return fmt.Errorf("seo_canonical_url exceeds the maximum length of %d characters", nuvioCMSBackofficeURLMaxLen)
	}
	if hasNuvioCMSBackofficeBlockedURLScheme(trimmed) {
		return fmt.Errorf("seo_canonical_url must be a valid http(s) URL or site-relative path")
	}
	if strings.HasPrefix(trimmed, "//") {
		return fmt.Errorf("seo_canonical_url must be a valid http(s) URL or site-relative path")
	}
	if hasNuvioCMSBackofficeUnsafeURLChars(trimmed) {
		return fmt.Errorf("seo_canonical_url must be a valid http(s) URL or site-relative path")
	}

	if strings.HasPrefix(trimmed, "/") {
		return nil
	}

	if err := validateNuvioCMSBackofficeAbsoluteHTTPURL(trimmed); err != nil {
		return fmt.Errorf("seo_canonical_url must be a valid http(s) URL or site-relative path")
	}

	return nil
}

func validateNuvioCMSBackofficePageSEOSocialImage(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	if len(trimmed) > nuvioCMSBackofficeURLMaxLen {
		return fmt.Errorf("seo_social_image exceeds the maximum length of %d characters", nuvioCMSBackofficeURLMaxLen)
	}
	if hasNuvioCMSBackofficeBlockedURLScheme(trimmed) {
		return fmt.Errorf("seo_social_image must be a safe URL, relative path, or file reference")
	}
	if strings.HasPrefix(trimmed, "//") {
		return fmt.Errorf("seo_social_image must be a safe URL, relative path, or file reference")
	}
	if hasNuvioCMSBackofficeUnsafeURLChars(trimmed) {
		return fmt.Errorf("seo_social_image must be a safe URL, relative path, or file reference")
	}

	if strings.HasPrefix(trimmed, "/") {
		return nil
	}

	if strings.Contains(trimmed, "://") {
		if err := validateNuvioCMSBackofficeAbsoluteHTTPURL(trimmed); err != nil {
			return fmt.Errorf("seo_social_image must be a safe URL, relative path, or file reference")
		}
		return nil
	}

	if strings.Contains(trimmed, "\\") || strings.Contains(trimmed, "..") || strings.Contains(trimmed, ":") {
		return fmt.Errorf("seo_social_image must be a safe URL, relative path, or file reference")
	}

	return nil
}

func validateNuvioCMSBackofficeSEOTranslationsTextLimits(value any) error {
	switch typed := value.(type) {
	case map[string]any:
		for rawKey, rawValue := range typed {
			normalizedKey := normalizeNuvioCMSBackofficePayloadKey(rawKey)
			switch normalizedKey {
			case "title", "seotitle":
				stringValue := strings.TrimSpace(parseStringValue(rawValue))
				if err := validateNuvioCMSBackofficeLimitedTextField("seo_translations.title", stringValue, nuvioCMSBackofficePageSEOTextMaxByKey["seotitle"]); err != nil {
					return err
				}
			case "description", "seodescription":
				stringValue := strings.TrimSpace(parseStringValue(rawValue))
				if err := validateNuvioCMSBackofficeLimitedTextField("seo_translations.description", stringValue, nuvioCMSBackofficePageSEOTextMaxByKey["seodescription"]); err != nil {
					return err
				}
			case "focuskeyword", "seofocuskeyword":
				stringValue := strings.TrimSpace(parseStringValue(rawValue))
				if err := validateNuvioCMSBackofficeLimitedTextField("seo_translations.focusKeyword", stringValue, nuvioCMSBackofficePageSEOTextMaxByKey["seofocuskeyword"]); err != nil {
					return err
				}
			}

			if err := validateNuvioCMSBackofficeSEOTranslationsTextLimits(rawValue); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range typed {
			if err := validateNuvioCMSBackofficeSEOTranslationsTextLimits(item); err != nil {
				return err
			}
		}
	}

	return nil
}

func hasNuvioCMSBackofficeBlockedURLScheme(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}

	return strings.HasPrefix(normalized, "javascript:") ||
		strings.HasPrefix(normalized, "data:") ||
		strings.HasPrefix(normalized, "vbscript:") ||
		strings.HasPrefix(normalized, "file:") ||
		strings.HasPrefix(normalized, "blob:")
}

func hasNuvioCMSBackofficeUnsafeURLChars(value string) bool {
	return strings.ContainsAny(value, " \t\r\n<>\"'`")
}

func validateNuvioCMSBackofficeAbsoluteHTTPURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return err
	}
	if parsed == nil {
		return fmt.Errorf("invalid URL")
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("invalid URL scheme")
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("invalid URL host")
	}
	if parsed.User != nil {
		return fmt.Errorf("invalid URL user info")
	}
	if hasNuvioCMSBackofficeUnsafeURLChars(parsed.Host) {
		return fmt.Errorf("invalid URL host")
	}

	return nil
}

func validateNuvioCMSBackofficeDomainBaseValue(value string) error {
	if value == "" {
		return nil
	}

	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "?") || strings.HasPrefix(value, "#") {
		return fmt.Errorf("invalid domain value")
	}
	if strings.Contains(value, "\\") {
		return fmt.Errorf("invalid domain value")
	}

	parsed, err := url.Parse("https://" + value)
	if err != nil {
		return err
	}
	if parsed == nil || strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("invalid domain value")
	}

	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return fmt.Errorf("invalid domain value")
	}

	normalizedHost := strings.ToLower(host)
	if normalizedHost != "localhost" && net.ParseIP(host) == nil && !strings.Contains(host, ".") {
		return fmt.Errorf("invalid domain value")
	}

	return nil
}

func applyNuvioCMSBackofficeIdentityPatch(app core.App, record *core.Record, payload map[string]any, isAdmin bool) error {
	if record == nil {
		return fmt.Errorf("Website not found")
	}

	updatedFieldCount := 0
	for rawKey, rawValue := range payload {
		normalizedKey := normalizeNuvioCMSBackofficePayloadKey(rawKey)
		if normalizedKey == "" {
			return fmt.Errorf("Invalid payload field")
		}

		if normalizedKey == "settings" || normalizedKey == "featureflags" || normalizedKey == "reports" {
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
		if _, deferredFileField := nuvioCMSBackofficeIdentityDeferredFilePayloadKeys[normalizedKey]; deferredFileField {
			return fmt.Errorf("File fields are not supported in this endpoint yet")
		}
		if _, allowed := nuvioCMSBackofficeIdentityAllowedPayloadKeys[normalizedKey]; !allowed {
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
		if _, adminOnly := nuvioCMSBackofficeIdentityAdminOnlyPayloadKeys[normalizedKey]; adminOnly && !isAdmin {
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}

		stringValue := ""
		if normalizedKey != "logo" && normalizedKey != "seoimage" {
			var err error
			stringValue, err = validateNuvioCMSBackofficeIdentityFieldValue(normalizedKey, rawValue)
			if err != nil {
				return err
			}
		}
		switch normalizedKey {
		case "name":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"name"}, stringValue) {
				updatedFieldCount++
			}
		case "title", "displayname":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"title", "name"}, stringValue) {
				updatedFieldCount++
			}
		case "slug":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"slug"}, stringValue) {
				updatedFieldCount++
			}
		case "domain":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"domain"}, stringValue) {
				updatedFieldCount++
			}
		case "seotitle":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"seoTitle", "seo_title"}, stringValue) {
				updatedFieldCount++
			}
		case "seodescription":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"seoDescription", "seo_description"}, stringValue) {
				updatedFieldCount++
			}
		case "logo":
			if err := setNuvioCMSBackofficeWebsiteSEOAssetField(app, record, []string{"logo"}, rawValue, "logo"); err != nil {
				return err
			}
			updatedFieldCount++
		case "seoimage":
			if err := setNuvioCMSBackofficeWebsiteSEOAssetField(app, record, []string{"seoImage", "seo_image"}, rawValue, "seoImage"); err != nil {
				return err
			}
			updatedFieldCount++
		case "seotitletemplate":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"seo_title_template", "seoTitleTemplate"}, stringValue) {
				updatedFieldCount++
			}
		case "seotitleseparator":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"seo_title_separator", "seoTitleSeparator"}, stringValue) {
				updatedFieldCount++
			}
		case "seocanonicaldomain":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"seo_canonical_domain", "seoCanonicalDomain"}, stringValue) {
				updatedFieldCount++
			}
		case "businessname":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_name", "businessName"}, stringValue) {
				updatedFieldCount++
			}
		case "businesstype":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_type", "businessType"}, stringValue) {
				updatedFieldCount++
			}
		case "businessprimarycategory":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_primary_category", "businessPrimaryCategory"}, stringValue) {
				updatedFieldCount++
			}
		case "businessphone":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_phone", "businessPhone"}, stringValue) {
				updatedFieldCount++
			}
		case "businessemail":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_email", "businessEmail"}, stringValue) {
				updatedFieldCount++
			}
		case "businessaddress":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_address", "businessAddress"}, stringValue) {
				updatedFieldCount++
			}
		case "businesscity":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_city", "businessCity"}, stringValue) {
				updatedFieldCount++
			}
		case "businesspostalcode":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_postal_code", "businessPostalCode"}, stringValue) {
				updatedFieldCount++
			}
		case "businesscountry":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_country", "businessCountry"}, stringValue) {
				updatedFieldCount++
			}
		case "businessservicearea":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_service_area", "businessServiceArea"}, stringValue) {
				updatedFieldCount++
			}
		case "businessopeninghours":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_opening_hours", "businessOpeningHours"}, stringValue) {
				updatedFieldCount++
			}
		case "businessgoogleplaceid":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_google_place_id", "businessGooglePlaceId"}, stringValue) {
				updatedFieldCount++
			}
		case "businesssocialprofiles":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_social_profiles", "businessSocialProfiles"}, stringValue) {
				updatedFieldCount++
			}
		case "businesspricerange":
			if setNuvioCMSBackofficeWebsiteStringField(record, []string{"business_price_range", "businessPriceRange"}, stringValue) {
				updatedFieldCount++
			}
		}
	}

	if updatedFieldCount == 0 {
		return fmt.Errorf("No editable identity fields are available for this website")
	}

	return nil
}

func setNuvioCMSBackofficeWebsiteStringField(record *core.Record, aliases []string, value string) bool {
	if record == nil {
		return false
	}

	collection := record.Collection()
	if collection == nil {
		return false
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, aliases)
	if fieldName == "" {
		return false
	}

	record.Set(fieldName, value)
	return true
}

func normalizeNuvioCMSBackofficeSettingsPatchPayload(payload map[string]any) (map[string]any, error) {
	settingsWrapperKey := ""
	var settingsWrapperValue any

	for key, value := range payload {
		if normalizeNuvioCMSBackofficePayloadKey(key) == "settings" {
			settingsWrapperKey = key
			settingsWrapperValue = value
			break
		}
	}

	if settingsWrapperKey == "" {
		return payload, nil
	}

	if len(payload) != 1 {
		return nil, fmt.Errorf("When settings is provided it must be the only top-level field")
	}

	settingsPatch, ok := toStringAnyMap(settingsWrapperValue)
	if !ok {
		return nil, fmt.Errorf("Settings must be an object")
	}
	if len(settingsPatch) == 0 {
		return nil, fmt.Errorf("At least one settings field is required")
	}

	return settingsPatch, nil
}

func mergeNuvioCMSBackofficeSettingsPatch(
	currentSettings map[string]any,
	patch map[string]any,
	isAdmin bool,
) (map[string]any, error) {
	if currentSettings == nil {
		currentSettings = map[string]any{}
	}

	for rawKey, rawValue := range patch {
		normalizedKey := normalizeNuvioCMSBackofficePayloadKey(rawKey)
		if normalizedKey == "" {
			return nil, fmt.Errorf("Invalid settings field")
		}

		if _, allowed := nuvioCMSBackofficeSettingsAllowedTopLevelKeys[normalizedKey]; !allowed {
			return nil, fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
		if !isAdmin {
			if _, denied := nuvioCMSBackofficeSettingsClientDeniedTopLevelKeys[normalizedKey]; denied {
				return nil, fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
			}
		}

		switch normalizedKey {
		case "featureflags":
			if err := applyNuvioCMSBackofficeFeatureFlagsPatch(currentSettings, rawValue, isAdmin); err != nil {
				return nil, err
			}
		case "contactform":
			if err := applyNuvioCMSBackofficeContactFormSettingsPatch(currentSettings, rawValue, isAdmin); err != nil {
				return nil, err
			}
		case "whatsapp":
			if err := applyNuvioCMSBackofficeWhatsAppSettingsPatch(currentSettings, rawValue, isAdmin); err != nil {
				return nil, err
			}
		case "newsletter":
			if err := applyNuvioCMSBackofficeNewsletterSettingsPatch(currentSettings, rawValue); err != nil {
				return nil, err
			}
		case "booking":
			if err := applyNuvioCMSBackofficeBookingSettingsPatch(currentSettings, rawValue, isAdmin); err != nil {
				return nil, err
			}
		case "i18n":
			if err := applyNuvioCMSBackofficeI18NSettingsPatch(currentSettings, rawValue); err != nil {
				return nil, err
			}
		case "reports":
			if err := applyNuvioCMSBackofficeReportsSettingsPatch(currentSettings, rawValue, isAdmin); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	return currentSettings, nil
}

func applyNuvioCMSBackofficeFeatureFlagsPatch(settings map[string]any, rawPatch any, isAdmin bool) error {
	if !isAdmin {
		return fmt.Errorf("Field \"featureFlags\" is not allowed in this endpoint")
	}

	featureFlagsPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("featureFlags must be an object")
	}
	if len(featureFlagsPatch) == 0 {
		return fmt.Errorf("featureFlags must include at least one field")
	}

	featureFlags := ensureNuvioCMSBackofficeChildMap(settings, "featureFlags")
	for rawKey, rawValue := range featureFlagsPatch {
		normalizedKey := normalizeNuvioCMSBackofficePayloadKey(rawKey)
		if _, allowed := nuvioCMSBackofficeSettingsFeatureFlagAllowedKeys[normalizedKey]; !allowed {
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}

		parsedValue, ok := parseBoolValue(rawValue)
		if !ok {
			return fmt.Errorf("featureFlags.%s must be a boolean", strings.TrimSpace(rawKey))
		}

		switch normalizedKey {
		case "contactform":
			featureFlags["contactForm"] = parsedValue
		default:
			featureFlags[normalizedKey] = parsedValue
		}
	}

	settings["featureFlags"] = featureFlags
	return nil
}

func applyNuvioCMSBackofficeContactFormSettingsPatch(settings map[string]any, rawPatch any, isAdmin bool) error {
	contactFormPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("contactForm must be an object")
	}
	if len(contactFormPatch) == 0 {
		return fmt.Errorf("contactForm must include at least one field")
	}

	contactFormSettings := ensureNuvioCMSBackofficeChildMap(settings, "contactForm")
	for rawKey, rawValue := range contactFormPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "enabled":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("contactForm.enabled must be a boolean")
			}
			contactFormSettings["enabled"] = value
		case "confirmationmessage":
			value, err := parseNuvioCMSBackofficeOptionalSettingString(rawValue, "contactForm.confirmationMessage", nuvioCMSBackofficeSettingsMessageMaxLen)
			if err != nil {
				return err
			}
			contactFormSettings["confirmationMessage"] = value
		case "fields":
			if err := applyNuvioCMSBackofficeContactFormFieldsPatch(contactFormSettings, rawValue); err != nil {
				return err
			}
		case "emailnotifications":
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(contactFormSettings, rawValue, "template", true, isAdmin); err != nil {
				return err
			}
		case "to", "cc", "template":
			// Backward-compatible flat payload support (maps to contactForm.emailNotifications.*).
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(contactFormSettings, map[string]any{rawKey: rawValue}, "template", true, isAdmin); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	settings["contactForm"] = contactFormSettings
	return nil
}

func applyNuvioCMSBackofficeContactFormFieldsPatch(contactFormSettings map[string]any, rawPatch any) error {
	fieldsPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("contactForm.fields must be an object")
	}

	fieldsSettings := ensureNuvioCMSBackofficeChildMap(contactFormSettings, "fields")
	for rawKey, rawValue := range fieldsPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "phone":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("contactForm.fields.phone must be a boolean")
			}
			fieldsSettings["phone"] = value
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	contactFormSettings["fields"] = fieldsSettings
	return nil
}

func applyNuvioCMSBackofficeWhatsAppSettingsPatch(settings map[string]any, rawPatch any, isAdmin bool) error {
	whatsAppPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("whatsapp must be an object")
	}
	if len(whatsAppPatch) == 0 {
		return fmt.Errorf("whatsapp must include at least one field")
	}

	whatsAppSettings := ensureNuvioCMSBackofficeChildMap(settings, "whatsapp")
	for rawKey, rawValue := range whatsAppPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "enabled":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("whatsapp.enabled must be a boolean")
			}
			whatsAppSettings["enabled"] = value
		case "phone":
			whatsAppSettings["phone"] = strings.TrimSpace(parseStringValue(rawValue))
		case "defaultmessage":
			value, err := parseNuvioCMSBackofficeOptionalSettingString(rawValue, "whatsapp.defaultMessage", nuvioCMSBackofficeSettingsMessageMaxLen)
			if err != nil {
				return err
			}
			whatsAppSettings["defaultMessage"] = value
		case "showfloatingbutton":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("whatsapp.showFloatingButton must be a boolean")
			}
			whatsAppSettings["showFloatingButton"] = value
		case "emailnotifications":
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(whatsAppSettings, rawValue, "template", true, isAdmin); err != nil {
				return err
			}
		case "to", "cc", "template":
			// Backward-compatible flat payload support (maps to whatsapp.emailNotifications.*).
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(whatsAppSettings, map[string]any{rawKey: rawValue}, "template", true, isAdmin); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	settings["whatsapp"] = whatsAppSettings
	return nil
}

func applyNuvioCMSBackofficeNewsletterSettingsPatch(settings map[string]any, rawPatch any) error {
	newsletterPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("newsletter must be an object")
	}
	if len(newsletterPatch) == 0 {
		return fmt.Errorf("newsletter must include at least one field")
	}

	newsletterSettings := ensureNuvioCMSBackofficeChildMap(settings, "newsletter")
	for rawKey, rawValue := range newsletterPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "doubleoptin":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("newsletter.doubleOptIn must be a boolean")
			}
			newsletterSettings["doubleOptIn"] = value
		case "lifecycle":
			lifecyclePatch, ok := toStringAnyMap(rawValue)
			if !ok {
				return fmt.Errorf("newsletter.lifecycle must be an object")
			}

			lifecycleSettings := ensureNuvioCMSBackofficeChildMap(newsletterSettings, "lifecycle")
			for lifecycleKey, lifecycleValue := range lifecyclePatch {
				switch normalizeNuvioCMSBackofficePayloadKey(lifecycleKey) {
				case "confirmationtemplate":
					templateSettings := ensureNuvioCMSBackofficeChildMap(lifecycleSettings, "confirmationTemplate")
					if err := applyNuvioCMSBackofficeTemplatePatch(templateSettings, lifecycleValue, false); err != nil {
						return fmt.Errorf("newsletter.lifecycle.confirmationTemplate: %w", err)
					}
					lifecycleSettings["confirmationTemplate"] = templateSettings
				default:
					return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(lifecycleKey))
				}
			}
			newsletterSettings["lifecycle"] = lifecycleSettings
		case "confirmationtemplate":
			// Backward-compatible flat payload support (maps to newsletter.lifecycle.confirmationTemplate).
			lifecycleSettings := ensureNuvioCMSBackofficeChildMap(newsletterSettings, "lifecycle")
			templateSettings := ensureNuvioCMSBackofficeChildMap(lifecycleSettings, "confirmationTemplate")
			if err := applyNuvioCMSBackofficeTemplatePatch(templateSettings, rawValue, false); err != nil {
				return fmt.Errorf("newsletter.lifecycle.confirmationTemplate: %w", err)
			}
			lifecycleSettings["confirmationTemplate"] = templateSettings
			newsletterSettings["lifecycle"] = lifecycleSettings
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	settings["newsletter"] = newsletterSettings
	return nil
}

func applyNuvioCMSBackofficeBookingSettingsPatch(settings map[string]any, rawPatch any, isAdmin bool) error {
	bookingPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("booking must be an object")
	}
	if len(bookingPatch) == 0 {
		return fmt.Errorf("booking must include at least one field")
	}

	bookingSettings := ensureNuvioCMSBackofficeChildMap(settings, "booking")
	for rawKey, rawValue := range bookingPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "enabled":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("booking.enabled must be a boolean")
			}
			bookingSettings["enabled"] = value
		case "confirmationmode":
			confirmationMode, err := parseNuvioCMSBackofficeBookingConfirmationMode(rawValue)
			if err != nil {
				return err
			}
			bookingSettings["confirmationMode"] = confirmationMode
		case "emailnotifications":
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(bookingSettings, rawValue, "businessTemplate", true, isAdmin); err != nil {
				return err
			}
		case "to", "cc", "businesstemplate":
			// Backward-compatible flat payload support (maps to booking.emailNotifications.*).
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(bookingSettings, map[string]any{rawKey: rawValue}, "businessTemplate", true, isAdmin); err != nil {
				return err
			}
		case "visitoremails":
			visitorEmailsPatch, ok := toStringAnyMap(rawValue)
			if !ok {
				return fmt.Errorf("booking.visitorEmails must be an object")
			}

			visitorEmailsSettings := ensureNuvioCMSBackofficeChildMap(bookingSettings, "visitorEmails")
			for visitorKey, visitorValue := range visitorEmailsPatch {
				var templateName string
				switch normalizeNuvioCMSBackofficePayloadKey(visitorKey) {
				case "requesttemplate":
					templateName = "requestTemplate"
				case "confirmationtemplate":
					templateName = "confirmationTemplate"
				case "rescheduletemplate":
					templateName = "rescheduleTemplate"
				default:
					return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(visitorKey))
				}

				templateSettings := ensureNuvioCMSBackofficeChildMap(visitorEmailsSettings, templateName)
				if err := applyNuvioCMSBackofficeTemplatePatch(templateSettings, visitorValue, false); err != nil {
					return fmt.Errorf("booking.visitorEmails.%s: %w", templateName, err)
				}
				visitorEmailsSettings[templateName] = templateSettings
			}
			bookingSettings["visitorEmails"] = visitorEmailsSettings
		case "requesttemplate", "confirmationtemplate", "rescheduletemplate":
			// Backward-compatible flat payload support (maps to booking.visitorEmails.*Template).
			visitorEmailsSettings := ensureNuvioCMSBackofficeChildMap(bookingSettings, "visitorEmails")
			var templateName string
			switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
			case "requesttemplate":
				templateName = "requestTemplate"
			case "confirmationtemplate":
				templateName = "confirmationTemplate"
			case "rescheduletemplate":
				templateName = "rescheduleTemplate"
			}
			templateSettings := ensureNuvioCMSBackofficeChildMap(visitorEmailsSettings, templateName)
			if err := applyNuvioCMSBackofficeTemplatePatch(templateSettings, rawValue, false); err != nil {
				return fmt.Errorf("booking.visitorEmails.%s: %w", templateName, err)
			}
			visitorEmailsSettings[templateName] = templateSettings
			bookingSettings["visitorEmails"] = visitorEmailsSettings
		case "rules":
			if err := applyNuvioCMSBackofficeBookingRulesPatch(bookingSettings, rawValue); err != nil {
				return err
			}
		case "minnoticehours", "bookingwindowdays", "bufferminutes", "calendarblockingmode":
			// Backward-compatible flat payload support (maps to booking.rules.*).
			if err := applyNuvioCMSBackofficeBookingRulesPatch(bookingSettings, map[string]any{rawKey: rawValue}); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	settings["booking"] = bookingSettings
	return nil
}

func applyNuvioCMSBackofficeBookingRulesPatch(bookingSettings map[string]any, rawPatch any) error {
	rulesPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("booking.rules must be an object")
	}
	if len(rulesPatch) == 0 {
		return fmt.Errorf("booking.rules must include at least one field")
	}

	rulesSettings := ensureNuvioCMSBackofficeChildMap(bookingSettings, "rules")
	for rawKey, rawValue := range rulesPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "minnoticehours":
			value, ok := parseNuvioBookingBackofficeNonNegativeInt(rawValue)
			if !ok {
				return fmt.Errorf("booking.rules.minNoticeHours must be a non-negative integer")
			}
			rulesSettings["minNoticeHours"] = value
		case "bookingwindowdays":
			value, ok := parseNuvioBookingBackofficeNonNegativeInt(rawValue)
			if !ok {
				return fmt.Errorf("booking.rules.bookingWindowDays must be a non-negative integer")
			}
			rulesSettings["bookingWindowDays"] = value
		case "bufferminutes":
			value, ok := parseNuvioBookingBackofficeNonNegativeInt(rawValue)
			if !ok {
				return fmt.Errorf("booking.rules.bufferMinutes must be a non-negative integer")
			}
			rulesSettings["bufferMinutes"] = value
		case "calendarblockingmode":
			mode, err := parseNuvioCMSBackofficeBookingCalendarBlockingMode(rawValue)
			if err != nil {
				return err
			}
			rulesSettings["calendarBlockingMode"] = mode
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	bookingSettings["rules"] = rulesSettings
	return nil
}

func applyNuvioCMSBackofficeI18NSettingsPatch(settings map[string]any, rawPatch any) error {
	i18nPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("i18n must be an object")
	}
	if len(i18nPatch) == 0 {
		return fmt.Errorf("i18n must include at least one field")
	}

	i18nSettings := ensureNuvioCMSBackofficeChildMap(settings, "i18n")
	for rawKey, rawValue := range i18nPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "enabled":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("i18n.enabled must be a boolean")
			}
			i18nSettings["enabled"] = value
		case "languages":
			languages, err := parseNuvioCMSBackofficeLanguages(rawValue)
			if err != nil {
				return err
			}
			i18nSettings["languages"] = languages
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	settings["i18n"] = i18nSettings
	return nil
}

func applyNuvioCMSBackofficeReportsSettingsPatch(settings map[string]any, rawPatch any, isAdmin bool) error {
	reportsPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("reports must be an object")
	}
	if len(reportsPatch) == 0 {
		return fmt.Errorf("reports must include at least one field")
	}

	reportsSettings := ensureNuvioCMSBackofficeChildMap(settings, "reports")
	for rawKey, rawValue := range reportsPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "analytics":
			if err := applyNuvioCMSBackofficeReportsAnalyticsPatch(reportsSettings, rawValue, isAdmin); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	settings["reports"] = reportsSettings
	return nil
}

func applyNuvioCMSBackofficeReportsAnalyticsPatch(reportsSettings map[string]any, rawPatch any, isAdmin bool) error {
	analyticsPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("reports.analytics must be an object")
	}
	if len(analyticsPatch) == 0 {
		return fmt.Errorf("reports.analytics must include at least one field")
	}

	analyticsSettings := ensureNuvioCMSBackofficeChildMap(reportsSettings, "analytics")
	for rawKey, rawValue := range analyticsPatch {
		normalizedKey := normalizeNuvioCMSBackofficePayloadKey(rawKey)
		switch normalizedKey {
		case "enabled":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("reports.analytics.enabled must be a boolean")
			}
			analyticsSettings["enabled"] = value
		case "provider":
			if !isAdmin {
				return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
			}
			provider := strings.ToLower(strings.TrimSpace(parseStringValue(rawValue)))
			if provider != "" && provider != "umami" {
				return fmt.Errorf("reports.analytics.provider must be umami or empty")
			}
			analyticsSettings["provider"] = provider
		case "scriptenabled":
			if !isAdmin {
				return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
			}
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("reports.analytics.scriptEnabled must be a boolean")
			}
			analyticsSettings["scriptEnabled"] = value
		case "siteid":
			if !isAdmin {
				return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
			}
			analyticsSettings["siteId"] = strings.TrimSpace(parseStringValue(rawValue))
		case "events":
			if !isAdmin {
				return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
			}

			eventsPatch, ok := toStringAnyMap(rawValue)
			if !ok {
				return fmt.Errorf("reports.analytics.events must be an object")
			}

			eventsSettings := ensureNuvioCMSBackofficeChildMap(analyticsSettings, "events")
			for eventsKey, eventsValue := range eventsPatch {
				switch normalizeNuvioCMSBackofficePayloadKey(eventsKey) {
				case "scrolldepth":
					value, ok := parseBoolValue(eventsValue)
					if !ok {
						return fmt.Errorf("reports.analytics.events.scrollDepth must be a boolean")
					}
					eventsSettings["scrollDepth"] = value
				default:
					return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(eventsKey))
				}
			}
			analyticsSettings["events"] = eventsSettings
		case "apiurl":
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		case "scripturl":
			if !isAdmin {
				return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
			}

			scriptURLValue := strings.TrimSpace(parseStringValue(rawValue))
			if scriptURLValue == "" {
				analyticsSettings["scriptUrl"] = ""
				continue
			}

			normalizedURL, err := normalizeNuvioAnalyticsURL(scriptURLValue)
			if err != nil {
				return fmt.Errorf("reports.analytics.scriptUrl must be a valid http(s) URL")
			}
			analyticsSettings["scriptUrl"] = normalizedURL
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	reportsSettings["analytics"] = analyticsSettings
	return nil
}

func applyNuvioCMSBackofficeEmailNotificationsPatch(
	parentSettings map[string]any,
	rawPatch any,
	templateKey string,
	allowDetailsFields bool,
	isAdmin bool,
) error {
	notificationsPatch, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("emailNotifications must be an object")
	}

	emailNotifications := ensureNuvioCMSBackofficeChildMap(parentSettings, "emailNotifications")
	for rawKey, rawValue := range notificationsPatch {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "enabled":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("emailNotifications.enabled must be a boolean")
			}
			emailNotifications["enabled"] = value
		case normalizeNuvioCMSBackofficePayloadKey(templateKey):
			templateSettings := ensureNuvioCMSBackofficeChildMap(emailNotifications, templateKey)
			if err := applyNuvioCMSBackofficeTemplatePatch(templateSettings, rawValue, allowDetailsFields); err != nil {
				return fmt.Errorf("emailNotifications.%s: %w", templateKey, err)
			}
			emailNotifications[templateKey] = templateSettings
		case "to", "cc":
			recipients, err := parseNuvioCMSBackofficeEmailRecipients(rawValue, "emailNotifications."+normalizeNuvioCMSBackofficePayloadKey(rawKey))
			if err != nil {
				return err
			}
			emailNotifications[normalizeNuvioCMSBackofficePayloadKey(rawKey)] = recipients
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	parentSettings["emailNotifications"] = emailNotifications
	return nil
}

func parseNuvioCMSBackofficeEmailRecipients(raw any, fieldPath string) ([]string, error) {
	rawItems := []any{}
	switch typed := raw.(type) {
	case []any:
		rawItems = typed
	case []string:
		for _, value := range typed {
			rawItems = append(rawItems, value)
		}
	default:
		return nil, fmt.Errorf("%s must be an array", fieldPath)
	}

	recipients := []string{}
	seen := map[string]struct{}{}
	for _, rawItem := range rawItems {
		rawRecipient := strings.TrimSpace(parseStringValue(rawItem))
		if rawRecipient == "" {
			continue
		}

		normalizedEmail, ok := normalizeNuvioEmail(rawRecipient)
		if !ok {
			return nil, fmt.Errorf("%s must contain valid email addresses", fieldPath)
		}

		if _, exists := seen[normalizedEmail]; exists {
			continue
		}
		seen[normalizedEmail] = struct{}{}
		recipients = append(recipients, normalizedEmail)
	}

	return recipients, nil
}

func applyNuvioCMSBackofficeTemplatePatch(
	templateSettings map[string]any,
	rawPatch any,
	allowDetailsFields bool,
) error {
	patchMap, ok := toStringAnyMap(rawPatch)
	if !ok {
		return fmt.Errorf("template must be an object")
	}

	for rawKey, rawValue := range patchMap {
		switch normalizeNuvioCMSBackofficePayloadKey(rawKey) {
		case "enabled":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("template.enabled must be a boolean")
			}
			templateSettings["enabled"] = value
		case "subject":
			value, err := parseNuvioCMSBackofficeOptionalSettingString(rawValue, "template.subject", nuvioCMSBackofficeSettingsTemplateSubjectMaxLen)
			if err != nil {
				return err
			}
			templateSettings["subject"] = value
		case "introtext":
			value, err := parseNuvioCMSBackofficeOptionalSettingString(rawValue, "template.introText", nuvioCMSBackofficeSettingsTemplateTextMaxLen)
			if err != nil {
				return err
			}
			templateSettings["introText"] = value
		case "footertext":
			value, err := parseNuvioCMSBackofficeOptionalSettingString(rawValue, "template.footerText", nuvioCMSBackofficeSettingsTemplateTextMaxLen)
			if err != nil {
				return err
			}
			templateSettings["footerText"] = value
		case "includeleaddetails":
			if !allowDetailsFields {
				return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
			}
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("template.includeLeadDetails must be a boolean")
			}
			templateSettings["includeLeadDetails"] = value
		case "includeappointmentdetails":
			if !allowDetailsFields {
				return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
			}
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("template.includeAppointmentDetails must be a boolean")
			}
			templateSettings["includeAppointmentDetails"] = value
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	return nil
}

func parseNuvioCMSBackofficeBookingConfirmationMode(raw any) (string, error) {
	normalized := strings.TrimSpace(parseStringValue(raw))
	if normalized == "" {
		return "", fmt.Errorf("booking.confirmationMode is required")
	}

	confirmationMode := normalizeNuvioBookingConfirmationMode(normalized)
	if !strings.EqualFold(normalized, nuvioBookingConfirmationModeRequest) &&
		!strings.EqualFold(normalized, nuvioBookingConfirmationModeAuto) {
		return "", fmt.Errorf("booking.confirmationMode must be request or autoConfirm")
	}

	return confirmationMode, nil
}

func parseNuvioCMSBackofficeBookingCalendarBlockingMode(raw any) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(parseStringValue(raw)))
	if normalized == "" {
		return "", fmt.Errorf("booking.rules.calendarBlockingMode is required")
	}

	switch normalized {
	case nuvioBookingBlockingModeService, nuvioBookingBlockingModeWebsite, nuvioBookingBlockingModeNone:
		return normalized, nil
	default:
		return "", fmt.Errorf("booking.rules.calendarBlockingMode must be service, website, or none")
	}
}

func parseNuvioCMSBackofficeLanguages(raw any) ([]map[string]string, error) {
	switch raw.(type) {
	case []any:
	default:
		return nil, fmt.Errorf("i18n.languages must be an array")
	}

	rawItems := normalizeNuvioPublicAnySlice(raw)
	languages := make([]map[string]string, 0, len(rawItems))
	for _, rawItem := range rawItems {
		item, ok := toStringAnyMap(rawItem)
		if !ok {
			return nil, fmt.Errorf("i18n.languages entries must be objects")
		}

		code, err := parseNuvioCMSBackofficeRequiredSettingString(item["code"], "i18n.languages[].code", nuvioCMSBackofficeI18NLanguageCodeMaxLen)
		if err != nil {
			return nil, err
		}
		if code == "" {
			return nil, fmt.Errorf("i18n.languages entries require code")
		}
		normalizedCode := normalizeNuvioCMSBackofficeLanguageCode(code)
		if normalizedCode == "" || !nuvioCMSBackofficeI18NLanguageCodePattern.MatchString(normalizedCode) {
			return nil, fmt.Errorf("i18n.languages entries require a valid code")
		}

		language := map[string]string{
			"code": normalizedCode,
		}
		label, err := parseNuvioCMSBackofficeOptionalSettingString(item["label"], "i18n.languages[].label", nuvioCMSBackofficeI18NLanguageLabelMaxLen)
		if err != nil {
			return nil, err
		}
		if label != "" {
			language["label"] = label
		}

		languages = append(languages, language)
	}

	return languages, nil
}

func parseNuvioCMSBackofficeOptionalSettingString(raw any, fieldPath string, maxLen int) (string, error) {
	switch typed := raw.(type) {
	case nil:
		return "", nil
	case string:
		value := strings.TrimSpace(typed)
		if err := validateNuvioCMSBackofficeLimitedTextField(fieldPath, value, maxLen); err != nil {
			return "", err
		}
		return value, nil
	default:
		return "", fmt.Errorf("%s must be a string", fieldPath)
	}
}

func parseNuvioCMSBackofficeRequiredSettingString(raw any, fieldPath string, maxLen int) (string, error) {
	value, err := parseNuvioCMSBackofficeOptionalSettingString(raw, fieldPath, maxLen)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", fmt.Errorf("%s is required", fieldPath)
	}
	return value, nil
}

func normalizeNuvioCMSBackofficeLanguageCode(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return normalized
}

func ensureNuvioCMSBackofficeChildMap(parent map[string]any, key string) map[string]any {
	if parent == nil {
		return map[string]any{}
	}

	existing, ok := toStringAnyMap(parent[key])
	if ok && existing != nil {
		return existing
	}

	return map[string]any{}
}

func findNuvioCMSDashboardCollectionByAliases(app core.App, aliases []string) (*core.Collection, error) {
	var lastErr error

	for _, alias := range aliases {
		normalizedAlias := strings.TrimSpace(alias)
		if normalizedAlias == "" {
			continue
		}

		collection, err := app.FindCachedCollectionByNameOrId(normalizedAlias)
		if err == nil {
			return collection, nil
		}

		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, sql.ErrNoRows
}

func findNuvioCMSDashboardRecordsByFilter(
	app core.App,
	collection *core.Collection,
	filterExpr string,
	params dbx.Params,
	sortCandidates []string,
) ([]*core.Record, error) {
	if collection == nil {
		return nil, fmt.Errorf("missing collection")
	}

	sanitizedSortCandidates := make([]string, 0, len(sortCandidates)+1)
	for _, sortCandidate := range sortCandidates {
		sortExpr := strings.TrimSpace(sortCandidate)
		if sortExpr != "" {
			sanitizedSortCandidates = append(sanitizedSortCandidates, sortExpr)
		}
	}
	sanitizedSortCandidates = append(sanitizedSortCandidates, "")

	var lastErr error
	for _, sortExpr := range sanitizedSortCandidates {
		records, err := app.FindRecordsByFilter(
			collection,
			filterExpr,
			sortExpr,
			nuvioCMSDashboardMaxScanRecords,
			0,
			params,
		)
		if err == nil {
			return records, nil
		}

		lastErr = err
		if sortExpr != "" && strings.Contains(strings.ToLower(err.Error()), "invalid sort field") {
			continue
		}

		return nil, err
	}

	return nil, lastErr
}

func loadNuvioCMSDashboardPages(
	app core.App,
	pagesCollection *core.Collection,
	websiteID string,
) ([]*core.Record, []nuvioCMSDashboardPageDTO, error) {
	if pagesCollection == nil {
		return []*core.Record{}, []nuvioCMSDashboardPageDTO{}, fmt.Errorf("missing pages collection")
	}

	websiteFieldName := resolveNuvioCollectionFieldNameByAliases(pagesCollection, []string{"website", "site"})
	if websiteFieldName == "" {
		return []*core.Record{}, []nuvioCMSDashboardPageDTO{}, fmt.Errorf("missing pages website relation field")
	}

	filterExpr := websiteFieldName + "={:websiteId}"
	records, err := findNuvioCMSDashboardRecordsByFilter(
		app,
		pagesCollection,
		filterExpr,
		dbx.Params{"websiteId": websiteID},
		[]string{
			"+title,+name,+slug,+created",
			"+name,+slug,+created",
			"+slug,+created",
			"+created",
		},
	)
	if err != nil {
		return []*core.Record{}, []nuvioCMSDashboardPageDTO{}, err
	}

	dtos := make([]nuvioCMSDashboardPageDTO, 0, len(records))
	for _, record := range records {
		dtos = append(dtos, buildNuvioCMSDashboardPageDTO(record))
	}

	return records, dtos, nil
}

func resolveNuvioCMSDashboardSelectedPage(
	pageRecords []*core.Record,
	pageDTOs []nuvioCMSDashboardPageDTO,
	pageID string,
) (*core.Record, *nuvioCMSDashboardPageDTO) {
	if len(pageRecords) == 0 || len(pageDTOs) == 0 {
		return nil, nil
	}

	if strings.TrimSpace(pageID) == "" {
		dto := pageDTOs[0]
		return pageRecords[0], &dto
	}

	for index, pageRecord := range pageRecords {
		if pageRecord == nil {
			continue
		}
		if strings.TrimSpace(pageRecord.Id) != strings.TrimSpace(pageID) {
			continue
		}

		dto := pageDTOs[index]
		return pageRecord, &dto
	}

	return nil, nil
}

func loadNuvioCMSDashboardBlocks(
	app core.App,
	blocksCollection *core.Collection,
	websiteID string,
	selectedPage *core.Record,
) ([]nuvioCMSDashboardBlockDTO, error) {
	if selectedPage == nil {
		return []nuvioCMSDashboardBlockDTO{}, nil
	}
	if blocksCollection == nil {
		return []nuvioCMSDashboardBlockDTO{}, fmt.Errorf("missing blocks collection")
	}

	pageFieldName := resolveNuvioCollectionFieldNameByAliases(blocksCollection, []string{"page"})
	if pageFieldName == "" {
		return []nuvioCMSDashboardBlockDTO{}, fmt.Errorf("missing blocks page relation field")
	}

	filterParts := []string{pageFieldName + "={:pageId}"}
	filterParams := dbx.Params{
		"pageId": strings.TrimSpace(selectedPage.Id),
	}

	if websiteFieldName := resolveNuvioCollectionFieldNameByAliases(blocksCollection, []string{"website", "site"}); websiteFieldName != "" {
		filterParts = append(filterParts, websiteFieldName+"={:websiteId}")
		filterParams["websiteId"] = websiteID
	}

	records, err := findNuvioCMSDashboardRecordsByFilter(
		app,
		blocksCollection,
		strings.Join(filterParts, " && "),
		filterParams,
		[]string{
			"+displayOrder,+order,+created",
			"+order,+created",
			"+created",
		},
	)
	if err != nil {
		return []nuvioCMSDashboardBlockDTO{}, err
	}

	dtos := make([]nuvioCMSDashboardBlockDTO, 0, len(records))
	for _, record := range records {
		dtos = append(dtos, buildNuvioCMSDashboardBlockDTO(record))
	}

	return dtos, nil
}

func loadNuvioCMSDashboardComponents(
	app core.App,
	componentsCollection *core.Collection,
) ([]nuvioCMSDashboardComponentDTO, error) {
	if componentsCollection == nil {
		return []nuvioCMSDashboardComponentDTO{}, fmt.Errorf("missing components collection")
	}

	records, err := findNuvioCMSDashboardRecordsByFilter(
		app,
		componentsCollection,
		"id != ''",
		dbx.Params{},
		[]string{
			"+name,+title,+key,+created",
			"+title,+key,+created",
			"+key,+created",
		},
	)
	if err != nil {
		return []nuvioCMSDashboardComponentDTO{}, err
	}

	dtos := make([]nuvioCMSDashboardComponentDTO, 0, len(records))
	for _, record := range records {
		dtos = append(dtos, buildNuvioCMSDashboardComponentDTO(record, componentsCollection))
	}

	return dtos, nil
}

func buildNuvioCMSDashboardWebsiteDTO(record *core.Record, isAdmin bool) nuvioCMSDashboardWebsiteDTO {
	dto := nuvioCMSDashboardWebsiteDTO{
		ID:                      "",
		DisplayName:             "",
		Name:                    "",
		Title:                   "",
		Slug:                    "",
		Domain:                  "",
		Logo:                    "",
		SEOTitle:                "",
		SEODescription:          "",
		SEOImage:                "",
		SEOTitleTemplate:        "",
		SEOTitleSeparator:       "",
		SEOCanonicalDomain:      "",
		BusinessName:            "",
		BusinessType:            "",
		BusinessPrimaryCategory: "",
		BusinessPhone:           "",
		BusinessEmail:           "",
		BusinessAddress:         "",
		BusinessCity:            "",
		BusinessPostalCode:      "",
		BusinessCountry:         "",
		BusinessServiceArea:     "",
		BusinessOpeningHours:    "",
		BusinessGooglePlaceID:   "",
		BusinessSocialProfiles:  "",
		BusinessPriceRange:      "",
		IdentitySEO:             map[string]any{},
		Settings:                map[string]any{},
	}

	if record == nil {
		return dto
	}

	dto.ID = strings.TrimSpace(record.Id)
	dto.Name = strings.TrimSpace(record.GetString("name"))
	dto.Title = strings.TrimSpace(record.GetString("title"))
	dto.Slug = strings.TrimSpace(record.GetString("slug"))
	dto.Domain = strings.TrimSpace(record.GetString("domain"))
	dto.DisplayName = resolveWebsiteDisplayName(record)

	dto.Logo = buildNuvioCMSDashboardSEOFileValue(record, []string{"logo"})
	dto.SEOTitle = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seoTitle", "seo_title"}))
	dto.SEODescription = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seoDescription", "seo_description"}))
	dto.SEOImage = buildNuvioCMSDashboardSEOFileValue(record, []string{"seoImage", "seo_image"})
	dto.SEOTitleTemplate = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seo_title_template", "seoTitleTemplate"}))
	dto.SEOTitleSeparator = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seo_title_separator", "seoTitleSeparator"}))
	dto.SEOCanonicalDomain = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seo_canonical_domain", "seoCanonicalDomain"}))
	dto.BusinessName = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_name", "businessName"}))
	dto.BusinessType = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_type", "businessType"}))
	dto.BusinessPrimaryCategory = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_primary_category", "businessPrimaryCategory"}))
	dto.BusinessPhone = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_phone", "businessPhone"}))
	dto.BusinessEmail = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_email", "businessEmail"}))
	dto.BusinessAddress = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_address", "businessAddress"}))
	dto.BusinessCity = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_city", "businessCity"}))
	dto.BusinessPostalCode = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_postal_code", "businessPostalCode"}))
	dto.BusinessCountry = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_country", "businessCountry"}))
	dto.BusinessServiceArea = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_service_area", "businessServiceArea"}))
	dto.BusinessOpeningHours = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_opening_hours", "businessOpeningHours"}))
	dto.BusinessGooglePlaceID = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_google_place_id", "businessGooglePlaceId"}))
	dto.BusinessSocialProfiles = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_social_profiles", "businessSocialProfiles"}))
	dto.BusinessPriceRange = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"business_price_range", "businessPriceRange"}))
	dto.IdentitySEO = buildNuvioCMSDashboardWebsiteIdentitySEODTO(dto)
	dto.Settings = buildNuvioCMSDashboardWebsiteSettingsDTO(record.Get("settings"), isAdmin)

	return dto
}

func buildNuvioCMSDashboardWebsiteIdentitySEODTO(websiteDTO nuvioCMSDashboardWebsiteDTO) map[string]any {
	return map[string]any{
		"logo":                      websiteDTO.Logo,
		"seoTitle":                  websiteDTO.SEOTitle,
		"seoDescription":            websiteDTO.SEODescription,
		"seoImage":                  websiteDTO.SEOImage,
		"seo_title_template":        websiteDTO.SEOTitleTemplate,
		"seo_title_separator":       websiteDTO.SEOTitleSeparator,
		"seo_canonical_domain":      websiteDTO.SEOCanonicalDomain,
		"business_name":             websiteDTO.BusinessName,
		"business_type":             websiteDTO.BusinessType,
		"business_primary_category": websiteDTO.BusinessPrimaryCategory,
		"business_phone":            websiteDTO.BusinessPhone,
		"business_email":            websiteDTO.BusinessEmail,
		"business_address":          websiteDTO.BusinessAddress,
		"business_city":             websiteDTO.BusinessCity,
		"business_postal_code":      websiteDTO.BusinessPostalCode,
		"business_country":          websiteDTO.BusinessCountry,
		"business_service_area":     websiteDTO.BusinessServiceArea,
		"business_opening_hours":    websiteDTO.BusinessOpeningHours,
		"business_google_place_id":  websiteDTO.BusinessGooglePlaceID,
		"business_social_profiles":  websiteDTO.BusinessSocialProfiles,
		"business_price_range":      websiteDTO.BusinessPriceRange,
	}
}

func buildNuvioCMSDashboardPageDTO(record *core.Record) nuvioCMSDashboardPageDTO {
	dto := nuvioCMSDashboardPageDTO{
		ID:                    "",
		Website:               "",
		Title:                 "",
		Name:                  "",
		Slug:                  "",
		Path:                  "",
		URL:                   "",
		Status:                "",
		Published:             false,
		Visible:               false,
		SEOTitle:              "",
		SEODescription:        "",
		SEOSocialImage:        "",
		SEOCanonicalURL:       "",
		SEONoindex:            false,
		SEOExcludeFromSitemap: false,
		SEOFocusKeyword:       "",
		SEOTranslations:       map[string]any{},
		Created:               "",
		Updated:               "",
	}

	if record == nil {
		return dto
	}

	publishedValue := false
	if value, ok := parseBoolValue(readNuvioCMSDashboardValueByAliases(record, []string{"published", "enabled", "active"})); ok {
		publishedValue = value
	}
	visibleValue := false
	if value, ok := parseBoolValue(readNuvioCMSDashboardValueByAliases(record, []string{"visible", "enabled", "active"})); ok {
		visibleValue = value
	}
	seoNoindex := false
	if value, ok := parseBoolValue(readNuvioCMSDashboardValueByAliases(record, []string{"seo_noindex", "seoNoindex"})); ok {
		seoNoindex = value
	}
	seoExcludeFromSitemap := false
	if value, ok := parseBoolValue(readNuvioCMSDashboardValueByAliases(record, []string{"seo_exclude_from_sitemap", "seoExcludeFromSitemap"})); ok {
		seoExcludeFromSitemap = value
	}

	dto.ID = strings.TrimSpace(record.Id)
	dto.Website = resolveNuvioPublicRelationID(record, "website", "site")
	dto.Title = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"title"}))
	dto.Name = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"name"}))
	dto.Slug = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"slug"}))
	dto.Path = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"path"}))
	dto.URL = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"url"}))
	dto.Status = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"status"}))
	dto.Published = publishedValue
	dto.Visible = visibleValue
	dto.SEOTitle = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seo_title", "seoTitle"}))
	dto.SEODescription = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seo_description", "seoDescription"}))
	dto.SEOSocialImage = buildNuvioCMSDashboardSEOFileValue(record, []string{"seo_social_image", "seoSocialImage"})
	dto.SEOCanonicalURL = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seo_canonical_url", "seoCanonicalUrl"}))
	dto.SEONoindex = seoNoindex
	dto.SEOExcludeFromSitemap = seoExcludeFromSitemap
	dto.SEOFocusKeyword = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seo_focus_keyword", "seoFocusKeyword"}))
	dto.SEOTranslations = normalizeNuvioPublicJSONValue(readNuvioCMSDashboardValueByAliases(record, []string{"seo_translations", "seoTranslations"}))
	dto.Created = strings.TrimSpace(record.GetString("created"))
	dto.Updated = strings.TrimSpace(record.GetString("updated"))

	return dto
}

func buildNuvioCMSDashboardBlockDTO(record *core.Record) nuvioCMSDashboardBlockDTO {
	dto := nuvioCMSDashboardBlockDTO{
		ID:           "",
		Page:         "",
		Website:      "",
		Component:    "",
		ComponentKey: "",
		Variant:      "",
		Slot:         "",
		DisplayOrder: 0,
		Order:        0,
		Props:        map[string]any{},
		Translations: map[string]any{},
		Enabled:      false,
		Visible:      false,
		Status:       "",
		Created:      "",
		Updated:      "",
	}

	if record == nil {
		return dto
	}

	enabledValue := false
	if value, ok := parseBoolValue(readNuvioCMSDashboardValueByAliases(record, []string{"enabled", "active"})); ok {
		enabledValue = value
	}
	visibleValue := false
	if value, ok := parseBoolValue(readNuvioCMSDashboardValueByAliases(record, []string{"visible", "enabled", "active"})); ok {
		visibleValue = value
	}

	dto.ID = strings.TrimSpace(record.Id)
	dto.Page = resolveNuvioPublicRelationID(record, "page")
	dto.Website = resolveNuvioPublicRelationID(record, "website", "site")
	dto.Component = resolveNuvioPublicRelationID(record, "component")
	dto.ComponentKey = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"component_key", "componentKey", "key"}))
	dto.Variant = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"variant"}))
	dto.Slot = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"slot", "region"}))
	dto.DisplayOrder = parseNuvioCMSDashboardIntByAliases(record, []string{"displayOrder", "display_order", "order"}, 0)
	dto.Order = parseNuvioCMSDashboardIntByAliases(record, []string{"order", "displayOrder", "display_order"}, dto.DisplayOrder)
	dto.Props = normalizeNuvioPublicJSONValue(readNuvioCMSDashboardValueByAliases(record, []string{"props"}))
	dto.Translations = normalizeNuvioPublicJSONValue(readNuvioCMSDashboardValueByAliases(record, []string{"translations"}))
	dto.Enabled = enabledValue
	dto.Visible = visibleValue
	dto.Status = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"status"}))
	dto.Created = strings.TrimSpace(record.GetString("created"))
	dto.Updated = strings.TrimSpace(record.GetString("updated"))

	return dto
}

func buildNuvioCMSDashboardComponentDTO(
	record *core.Record,
	collection *core.Collection,
) nuvioCMSDashboardComponentDTO {
	dto := nuvioCMSDashboardComponentDTO{
		ID:             "",
		Key:            "",
		ComponentKey:   "",
		Name:           "",
		Title:          "",
		Label:          "",
		Category:       "",
		Group:          "",
		Variant:        "",
		DefaultVariant: "",
		Schema:         map[string]any{},
	}

	if record == nil {
		return dto
	}

	schemaFieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"schema"})
	keyValue := strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"key", "component_key", "componentKey"}))
	if keyValue == "" {
		keyValue = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"slug"}))
	}

	nameValue := strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"name", "title", "label"}))
	titleValue := strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"title"}))
	labelValue := strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"label"}))

	dto.ID = strings.TrimSpace(record.Id)
	dto.Key = keyValue
	dto.ComponentKey = keyValue
	dto.Name = nameValue
	dto.Title = titleValue
	dto.Label = labelValue
	dto.Category = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"category"}))
	dto.Group = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"group"}))
	dto.Variant = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"variant"}))
	dto.DefaultVariant = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"defaultVariant", "default_variant"}))
	if schemaFieldName != "" {
		dto.Schema = normalizeNuvioPublicJSONValue(record.Get(schemaFieldName))
	}

	return dto
}

func buildNuvioCMSDashboardWebsiteSettingsDTO(rawSettings any, isAdmin bool) map[string]any {
	settings := parseNuvioSettingsObject(rawSettings)
	dto := map[string]any{}

	if featureFlagsRaw, ok := toStringAnyMap(settings["featureFlags"]); ok {
		featureFlags := map[string]any{}
		for _, featureKey := range nuvioCMSDashboardWebsiteFeatureFlagKeys {
			if value, ok := parseBoolValue(featureFlagsRaw[featureKey]); ok {
				featureFlags[featureKey] = value
			}
		}
		if len(featureFlags) > 0 {
			dto["featureFlags"] = featureFlags
		}
	}

	if contactFormRaw, ok := toStringAnyMap(settings["contactForm"]); ok {
		contactForm := map[string]any{}
		if value, ok := parseBoolValue(contactFormRaw["enabled"]); ok {
			contactForm["enabled"] = value
		}
		if message := strings.TrimSpace(parseStringValue(contactFormRaw["confirmationMessage"])); message != "" {
			contactForm["confirmationMessage"] = message
		}
		if fieldsRaw, ok := toStringAnyMap(contactFormRaw["fields"]); ok {
			fields := map[string]any{}
			if value, ok := parseBoolValue(fieldsRaw["phone"]); ok {
				fields["phone"] = value
			}
			if len(fields) > 0 {
				contactForm["fields"] = fields
			}
		}
		if notifications := sanitizeNuvioCMSDashboardEmailNotifications(contactFormRaw["emailNotifications"], true); len(notifications) > 0 {
			contactForm["emailNotifications"] = notifications
		}
		if len(contactForm) > 0 {
			dto["contactForm"] = contactForm
		}
	}

	if whatsappRaw, ok := toStringAnyMap(settings["whatsapp"]); ok {
		whatsapp := map[string]any{}
		if value, ok := parseBoolValue(whatsappRaw["enabled"]); ok {
			whatsapp["enabled"] = value
		}
		if phone := strings.TrimSpace(parseStringValue(whatsappRaw["phone"])); phone != "" {
			whatsapp["phone"] = phone
		}
		if message := strings.TrimSpace(parseStringValue(whatsappRaw["defaultMessage"])); message != "" {
			whatsapp["defaultMessage"] = message
		}
		if value, ok := parseBoolValue(whatsappRaw["showFloatingButton"]); ok {
			whatsapp["showFloatingButton"] = value
		}
		if notifications := sanitizeNuvioCMSDashboardEmailNotifications(whatsappRaw["emailNotifications"], true); len(notifications) > 0 {
			whatsapp["emailNotifications"] = notifications
		}
		if len(whatsapp) > 0 {
			dto["whatsapp"] = whatsapp
		}
	}

	if newsletterRaw, ok := toStringAnyMap(settings["newsletter"]); ok {
		newsletter := map[string]any{}
		if value, ok := parseBoolValue(newsletterRaw["doubleOptIn"]); ok {
			newsletter["doubleOptIn"] = value
		}
		if lifecycleRaw, ok := toStringAnyMap(newsletterRaw["lifecycle"]); ok {
			lifecycle := map[string]any{}
			if confirmationTemplate := sanitizeNuvioCMSDashboardTemplateObject(
				lifecycleRaw["confirmationTemplate"],
				false,
			); len(confirmationTemplate) > 0 {
				lifecycle["confirmationTemplate"] = confirmationTemplate
			}
			if len(lifecycle) > 0 {
				newsletter["lifecycle"] = lifecycle
			}
		}
		if len(newsletter) > 0 {
			dto["newsletter"] = newsletter
		}
	}

	if bookingRaw, ok := toStringAnyMap(settings["booking"]); ok {
		booking := map[string]any{}
		if value, ok := parseBoolValue(bookingRaw["enabled"]); ok {
			booking["enabled"] = value
		}
		confirmationMode := normalizeNuvioBookingConfirmationMode(bookingRaw["confirmationMode"])
		if confirmationMode != "" {
			booking["confirmationMode"] = confirmationMode
		}
		if notifications := sanitizeNuvioCMSDashboardBookingEmailNotifications(bookingRaw["emailNotifications"], true); len(notifications) > 0 {
			booking["emailNotifications"] = notifications
		}
		if visitorEmailsRaw, ok := toStringAnyMap(bookingRaw["visitorEmails"]); ok {
			visitorEmails := map[string]any{}
			if template := sanitizeNuvioCMSDashboardTemplateObject(visitorEmailsRaw["requestTemplate"], false); len(template) > 0 {
				visitorEmails["requestTemplate"] = template
			}
			if template := sanitizeNuvioCMSDashboardTemplateObject(visitorEmailsRaw["confirmationTemplate"], false); len(template) > 0 {
				visitorEmails["confirmationTemplate"] = template
			}
			if template := sanitizeNuvioCMSDashboardTemplateObject(visitorEmailsRaw["rescheduleTemplate"], false); len(template) > 0 {
				visitorEmails["rescheduleTemplate"] = template
			}
			if len(visitorEmails) > 0 {
				booking["visitorEmails"] = visitorEmails
			}
		}
		if rulesRaw, ok := toStringAnyMap(bookingRaw["rules"]); ok {
			rules := map[string]any{
				"minNoticeHours":       parseNuvioNonNegativeInt(rulesRaw["minNoticeHours"], 0),
				"bookingWindowDays":    parseNuvioNonNegativeInt(rulesRaw["bookingWindowDays"], 0),
				"bufferMinutes":        parseNuvioNonNegativeInt(rulesRaw["bufferMinutes"], 0),
				"calendarBlockingMode": normalizeNuvioBookingCalendarBlockingMode(rulesRaw["calendarBlockingMode"]),
			}
			booking["rules"] = rules
		}
		if len(booking) > 0 {
			dto["booking"] = booking
		}
	}

	if i18nRaw, ok := toStringAnyMap(settings["i18n"]); ok {
		i18n := map[string]any{}
		if value, ok := parseBoolValue(i18nRaw["enabled"]); ok {
			i18n["enabled"] = value
		}
		if languages := sanitizeNuvioCMSDashboardLanguages(i18nRaw["languages"]); len(languages) > 0 {
			i18n["languages"] = languages
		}
		if defaultLanguage := sanitizeNuvioCMSDashboardLanguageCode(i18nRaw["defaultLanguage"]); defaultLanguage != "" {
			i18n["defaultLanguage"] = defaultLanguage
		}
		if defaultLanguage := sanitizeNuvioCMSDashboardLanguageCode(i18nRaw["default_language"]); defaultLanguage != "" {
			if _, hasDefaultLanguage := i18n["defaultLanguage"]; !hasDefaultLanguage {
				i18n["defaultLanguage"] = defaultLanguage
			}
		}
		if len(i18n) > 0 {
			dto["i18n"] = i18n
		}
	}

	if previewRoutes := sanitizeNuvioCMSDashboardPreviewRoutes(settings["previewRoutes"]); len(previewRoutes) > 0 {
		dto["previewRoutes"] = previewRoutes
	}

	if reportsRaw, ok := toStringAnyMap(settings["reports"]); ok {
		if analyticsRaw, ok := toStringAnyMap(reportsRaw["analytics"]); ok {
			analytics := map[string]any{}
			if value, ok := parseBoolValue(analyticsRaw["enabled"]); ok {
				analytics["enabled"] = value
			}
			if isAdmin {
				provider := strings.ToLower(strings.TrimSpace(parseStringValue(analyticsRaw["provider"])))
				if provider == "umami" {
					analytics["provider"] = provider
				}
				if value, ok := parseBoolValue(analyticsRaw["scriptEnabled"]); ok {
					analytics["scriptEnabled"] = value
				}
				if siteID := strings.TrimSpace(parseStringValue(analyticsRaw["siteId"])); siteID != "" {
					analytics["siteId"] = siteID
				}
				if scriptURL := strings.TrimSpace(parseStringValue(analyticsRaw["scriptUrl"])); scriptURL != "" {
					if normalizedURL, err := normalizeNuvioAnalyticsURL(scriptURL); err == nil {
						analytics["scriptUrl"] = normalizedURL
					}
				}
				if eventsRaw, ok := toStringAnyMap(analyticsRaw["events"]); ok {
					events := map[string]any{}
					if value, ok := parseBoolValue(eventsRaw["scrollDepth"]); ok {
						events["scrollDepth"] = value
					}
					if len(events) > 0 {
						analytics["events"] = events
					}
				}
			}

			if len(analytics) > 0 {
				dto["reports"] = map[string]any{
					"analytics": analytics,
				}
			}
		}
	}

	return dto
}

func sanitizeNuvioCMSDashboardPreviewRoutes(raw any) map[string]string {
	rawRoutes, ok := toStringAnyMap(raw)
	if !ok {
		return map[string]string{}
	}

	routes := map[string]string{}
	for rawSlug, rawPath := range rawRoutes {
		pageSlug := strings.ToLower(strings.TrimSpace(rawSlug))
		if pageSlug == "" || strings.ContainsAny(pageSlug, "/\\?#") {
			continue
		}
		path := sanitizeNuvioCMSDashboardPreviewPath(rawPath)
		if path == "" {
			continue
		}
		routes[pageSlug] = path
	}
	return routes
}

func sanitizeNuvioCMSDashboardPreviewPath(raw any) string {
	path := strings.TrimSpace(parseStringValue(raw))
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "\\") {
		return ""
	}
	for _, char := range path {
		if char < 0x20 || char == 0x7f {
			return ""
		}
	}
	return path
}
func sanitizeNuvioCMSDashboardEmailNotifications(raw any, includeRecipients bool) map[string]any {
	notificationsRaw, ok := toStringAnyMap(raw)
	if !ok {
		return map[string]any{}
	}

	notifications := map[string]any{}
	if value, ok := parseBoolValue(notificationsRaw["enabled"]); ok {
		notifications["enabled"] = value
	}

	if template := sanitizeNuvioCMSDashboardTemplateObject(notificationsRaw["template"], true); len(template) > 0 {
		notifications["template"] = template
	}
	if includeRecipients {
		if recipients := sanitizeNuvioCMSDashboardEmailRecipients(notificationsRaw["to"]); len(recipients) > 0 {
			notifications["to"] = recipients
		}
		if recipients := sanitizeNuvioCMSDashboardEmailRecipients(notificationsRaw["cc"]); len(recipients) > 0 {
			notifications["cc"] = recipients
		}
	}

	return notifications
}

func sanitizeNuvioCMSDashboardBookingEmailNotifications(raw any, includeRecipients bool) map[string]any {
	notificationsRaw, ok := toStringAnyMap(raw)
	if !ok {
		return map[string]any{}
	}

	notifications := map[string]any{}
	if value, ok := parseBoolValue(notificationsRaw["enabled"]); ok {
		notifications["enabled"] = value
	}

	if template := sanitizeNuvioCMSDashboardTemplateObject(notificationsRaw["businessTemplate"], true); len(template) > 0 {
		notifications["businessTemplate"] = template
	}
	if includeRecipients {
		if recipients := sanitizeNuvioCMSDashboardEmailRecipients(notificationsRaw["to"]); len(recipients) > 0 {
			notifications["to"] = recipients
		}
		if recipients := sanitizeNuvioCMSDashboardEmailRecipients(notificationsRaw["cc"]); len(recipients) > 0 {
			notifications["cc"] = recipients
		}
	}

	return notifications
}

func sanitizeNuvioCMSDashboardEmailRecipients(raw any) []string {
	var rawItems []any
	switch typed := raw.(type) {
	case []any:
		rawItems = typed
	case []string:
		rawItems = make([]any, 0, len(typed))
		for _, value := range typed {
			rawItems = append(rawItems, value)
		}
	default:
		return []string{}
	}

	recipients := make([]string, 0, len(rawItems))
	seen := map[string]struct{}{}
	for _, rawItem := range rawItems {
		normalizedEmail, ok := normalizeNuvioEmail(parseStringValue(rawItem))
		if !ok {
			continue
		}
		if _, exists := seen[normalizedEmail]; exists {
			continue
		}

		seen[normalizedEmail] = struct{}{}
		recipients = append(recipients, normalizedEmail)
	}

	return recipients
}

func sanitizeNuvioCMSDashboardTemplateObject(raw any, includeLeadDetails bool) map[string]any {
	templateRaw, ok := toStringAnyMap(raw)
	if !ok {
		return map[string]any{}
	}

	template := map[string]any{}
	if value, ok := parseBoolValue(templateRaw["enabled"]); ok {
		template["enabled"] = value
	}
	if subject := strings.TrimSpace(parseStringValue(templateRaw["subject"])); subject != "" {
		template["subject"] = subject
	}
	if introText := strings.TrimSpace(parseStringValue(templateRaw["introText"])); introText != "" {
		template["introText"] = introText
	}
	if footerText := strings.TrimSpace(parseStringValue(templateRaw["footerText"])); footerText != "" {
		template["footerText"] = footerText
	}
	if includeLeadDetails {
		if value, ok := parseBoolValue(templateRaw["includeLeadDetails"]); ok {
			template["includeLeadDetails"] = value
		}
		if value, ok := parseBoolValue(templateRaw["includeAppointmentDetails"]); ok {
			template["includeAppointmentDetails"] = value
		}
	}

	return template
}

func sanitizeNuvioCMSDashboardLanguages(raw any) []map[string]string {
	rawItems := normalizeNuvioCMSDashboardAnySlice(raw)
	if len(rawItems) == 0 {
		return []map[string]string{}
	}

	seenByCode := map[string]struct{}{}
	result := make([]map[string]string, 0, len(rawItems))
	for _, rawItem := range rawItems {
		code := sanitizeNuvioCMSDashboardLanguageCodeFromEntry(rawItem)
		label := ""

		if entry, ok := toStringAnyMap(rawItem); ok {
			label = strings.TrimSpace(parseStringValue(entry["label"]))
			if label == "" {
				label = strings.TrimSpace(parseStringValue(entry["name"]))
			}
		}

		if code == "" {
			continue
		}
		if _, exists := seenByCode[code]; exists {
			continue
		}
		seenByCode[code] = struct{}{}

		language := map[string]string{
			"code": code,
		}
		if label != "" {
			language["label"] = label
		}

		result = append(result, language)
	}

	return result
}

func sanitizeNuvioCMSDashboardLanguageCodeFromEntry(raw any) string {
	if code := sanitizeNuvioCMSDashboardLanguageCode(raw); code != "" {
		return code
	}

	entry, ok := toStringAnyMap(raw)
	if !ok {
		return ""
	}

	for _, candidate := range []string{"code", "language", "lang", "locale", "value", "id", "key"} {
		if code := sanitizeNuvioCMSDashboardLanguageCode(entry[candidate]); code != "" {
			return code
		}
		if nestedEntry, ok := toStringAnyMap(entry[candidate]); ok {
			for _, nestedCandidate := range []string{"code", "language", "lang", "locale", "value", "id", "key"} {
				if code := sanitizeNuvioCMSDashboardLanguageCode(nestedEntry[nestedCandidate]); code != "" {
					return code
				}
			}
		}
	}

	return ""
}

func normalizeNuvioCMSDashboardAnySlice(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	case []string:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, item)
		}
		return result
	case []map[string]any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, item)
		}
		return result
	case []map[string]string:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			next := map[string]any{}
			for key, value := range item {
				next[key] = value
			}
			result = append(result, next)
		}
		return result
	default:
		return []any{}
	}
}

func sanitizeNuvioCMSDashboardLanguageCode(raw any) string {
	normalized := strings.ToLower(strings.TrimSpace(parseStringValue(raw)))
	if normalized == "" {
		return ""
	}

	for _, char := range normalized {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			continue
		}
		return ""
	}

	return normalized
}

func parseNuvioCMSDashboardStringByAliases(record *core.Record, aliases []string) string {
	return strings.TrimSpace(parseStringValue(readNuvioCMSDashboardValueByAliases(record, aliases)))
}

func parseNuvioCMSDashboardIntByAliases(record *core.Record, aliases []string, fallback int) int {
	value := readNuvioCMSDashboardValueByAliases(record, aliases)
	return parseNuvioNonNegativeInt(value, fallback)
}

func readNuvioCMSDashboardValueByAliases(record *core.Record, aliases []string) any {
	if record == nil {
		return nil
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}

		collection := record.Collection()
		if collection != nil && collection.Fields.GetByName(fieldName) == nil {
			continue
		}

		value := record.Get(fieldName)
		if value == nil {
			continue
		}

		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				continue
			}
		}

		return value
	}

	return nil
}

func toNuvioCMSDashboardSingleFileName(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []string:
		for _, item := range typed {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				return trimmed
			}
		}
	case []any:
		for _, item := range typed {
			if trimmed := strings.TrimSpace(parseStringValue(item)); trimmed != "" {
				return trimmed
			}
		}
	}

	return strings.TrimSpace(parseStringValue(value))
}

func buildNuvioCMSDashboardSEOFileValue(record *core.Record, aliases []string) any {
	if record == nil {
		return ""
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(record.Collection(), aliases)
	if fieldName == "" {
		return ""
	}

	rawValue := record.Get(fieldName)
	filename := toNuvioCMSDashboardSingleFileName(rawValue)
	if filename == "" {
		return ""
	}

	if ref, ok := parseNuvioCMSDashboardStoredAssetRef(filename); ok {
		return map[string]string{
			"collection": ref.Collection,
			"recordId":   ref.RecordID,
			"filename":   ref.Filename,
		}
	}

	field := record.Collection().Fields.GetByName(fieldName)
	if field != nil && field.Type() == core.FieldTypeFile {
		collectionName := strings.TrimSpace(record.Collection().Name)
		if collectionName == "" {
			collectionName = strings.TrimSpace(record.Collection().Id)
		}
		return map[string]string{
			"collection": collectionName,
			"recordId":   strings.TrimSpace(record.Id),
			"filename":   filename,
		}
	}

	return filename
}

func parseNuvioCMSDashboardStoredAssetRef(value string) (nuvioCMSBackofficeSEOAssetRef, bool) {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "{") {
		return nuvioCMSBackofficeSEOAssetRef{}, false
	}

	parsed := map[string]any{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nuvioCMSBackofficeSEOAssetRef{}, false
	}

	ref, ok, err := parseNuvioCMSBackofficeSEOAssetRef(parsed)
	if err != nil || !ok {
		return nuvioCMSBackofficeSEOAssetRef{}, false
	}

	return ref, true
}

func validateNuvioCMSBackofficeAssetUpload(uploadedFile *filesystem.File) (string, error) {
	if uploadedFile == nil {
		return "", fmt.Errorf("Missing file.")
	}
	if uploadedFile.Size <= 0 {
		return "", fmt.Errorf("File is empty.")
	}
	if uploadedFile.Size > nuvioCMSBackofficeAssetMaxFileSizeBytes {
		return "", fmt.Errorf("File exceeds the maximum allowed size of 8MB.")
	}

	reader, err := uploadedFile.Reader.Open()
	if err != nil {
		return "", fmt.Errorf("Failed to read uploaded file.")
	}
	defer reader.Close()

	detectedMimeType, err := mimetype.DetectReader(reader)
	if err != nil || detectedMimeType == nil {
		return "", fmt.Errorf("Failed to detect uploaded file type.")
	}

	for _, allowedMimeType := range nuvioCMSBackofficeAssetAllowedMimeTypes {
		if detectedMimeType.Is(allowedMimeType) {
			return allowedMimeType, nil
		}
	}

	return "", fmt.Errorf("Unsupported file type. Allowed types: image/jpeg, image/png, image/webp, image/gif.")
}

func computeNuvioCMSBackofficeAssetChecksum(uploadedFile *filesystem.File) (string, error) {
	if uploadedFile == nil {
		return "", fmt.Errorf("Missing file.")
	}

	reader, err := uploadedFile.Reader.Open()
	if err != nil {
		return "", fmt.Errorf("Failed to read uploaded file.")
	}
	defer reader.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", fmt.Errorf("Failed to read uploaded file.")
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func validateNuvioCMSBackofficeAssetMultipartFields(form *multipart.Form) error {
	if form == nil {
		return nil
	}

	for key := range form.Value {
		if strings.TrimSpace(key) == "websiteId" {
			continue
		}
		return fmt.Errorf("Unsupported field %q in upload payload.", key)
	}

	for key := range form.File {
		if strings.TrimSpace(key) == "file" {
			continue
		}
		return fmt.Errorf("Unsupported field %q in upload payload.", key)
	}

	return nil
}

func resolveNuvioCMSBackofficeAssetWebsiteID(record *core.Record, collection *core.Collection) string {
	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID != "" {
		return websiteID
	}

	if collection == nil {
		return ""
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"website", "site"})
	if fieldName == "" {
		return ""
	}

	return strings.TrimSpace(parseStringValue(record.Get(fieldName)))
}

func setNuvioCMSBackofficeAssetWebsite(record *core.Record, collection *core.Collection, websiteID string) {
	if record == nil || collection == nil {
		return
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"website", "site"})
	if fieldName == "" {
		return
	}

	field := collection.Fields.GetByName(fieldName)
	if field != nil && field.Type() == core.FieldTypeRelation {
		record.Set(fieldName, []string{websiteID})
		return
	}

	record.Set(fieldName, websiteID)
}

func setNuvioCMSBackofficeAssetMetadata(
	record *core.Record,
	collection *core.Collection,
	uploadedFile *filesystem.File,
	detectedMIME string,
	checksum string,
) {
	if record == nil || collection == nil || uploadedFile == nil {
		return
	}

	if fieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"originalName"}); fieldName != "" {
		record.Set(fieldName, strings.TrimSpace(uploadedFile.OriginalName))
	}
	if fieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"mimeType", "mime_type"}); fieldName != "" {
		record.Set(fieldName, strings.TrimSpace(detectedMIME))
	}
	if fieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"size"}); fieldName != "" {
		record.Set(fieldName, uploadedFile.Size)
	}
	if fieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"checksum"}); fieldName != "" {
		record.Set(fieldName, strings.TrimSpace(checksum))
	}
}

func buildNuvioCMSBackofficeAssetDTO(
	record *core.Record,
	assetsCollection *core.Collection,
) nuvioCMSBackofficeAssetDTO {
	dto := nuvioCMSBackofficeAssetDTO{}
	if record == nil {
		return dto
	}

	dto.ID = strings.TrimSpace(record.Id)
	dto.Website = resolveNuvioCMSBackofficeAssetWebsiteID(record, assetsCollection)
	dto.OriginalName = parseNuvioCMSDashboardStringByAliases(record, []string{"originalName"})
	dto.MimeType = parseNuvioCMSDashboardStringByAliases(record, []string{"mimeType", "mime_type"})
	dto.Size = int64(parseNuvioCMSDashboardIntByAliases(record, []string{"size"}, 0))
	dto.Created = strings.TrimSpace(record.GetDateTime("created").String())
	dto.Updated = strings.TrimSpace(record.GetDateTime("updated").String())

	collectionName := ""
	if assetsCollection != nil {
		collectionName = strings.TrimSpace(assetsCollection.Name)
	} else if collection := record.Collection(); collection != nil {
		collectionName = strings.TrimSpace(collection.Name)
	}
	dto.Collection = collectionName

	fileFieldName := "file"
	if assetsCollection != nil {
		if resolvedFileFieldName := resolveNuvioCollectionFieldNameByAliases(assetsCollection, []string{"file"}); resolvedFileFieldName != "" {
			fileFieldName = resolvedFileFieldName
		}
	}
	dto.Filename = toNuvioCMSDashboardSingleFileName(record.Get(fileFieldName))
	if dto.Filename != "" {
		dto.File = &nuvioCMSBackofficeAssetFileRefDTO{
			RecordID:   dto.ID,
			Filename:   dto.Filename,
			Collection: collectionName,
		}
	}

	return dto
}

// NUVIO CUSTOM END: Scoped CMS backoffice dashboard endpoint (A3.5.8B).
