package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioComponentsCollectionID     = "pbc_184785686"
	nuvioCMSDashboardMaxScanRecords = 5000
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
		"logo":            {},
		"seoimage":        {},
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
)

type nuvioCMSDashboardWebsiteDTO struct {
	ID                      string         `json:"id"`
	DisplayName             string         `json:"displayName"`
	Name                    string         `json:"name,omitempty"`
	Title                   string         `json:"title,omitempty"`
	Slug                    string         `json:"slug,omitempty"`
	Domain                  string         `json:"domain,omitempty"`
	Logo                    string         `json:"logo,omitempty"`
	SEOTitle                string         `json:"seoTitle,omitempty"`
	SEODescription          string         `json:"seoDescription,omitempty"`
	SEOImage                string         `json:"seoImage,omitempty"`
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
	SEOSocialImage        string `json:"seo_social_image"`
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

// NUVIO CUSTOM START: Scoped CMS backoffice dashboard endpoint (A3.5.8B).
func registerNuvioCMSBackofficeRoutes(e *core.ServeEvent) {
	cmsGroup := e.Router.Group("/api/nuvio/cms").Bind(apis.RequireSuperuserAuth())

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

		if err := applyNuvioCMSBackofficeIdentityPatch(websiteRecord, payload, apis.IsAdminSuperuser(e.Auth)); err != nil {
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

		if err := applyNuvioCMSBackofficePageSEOPatch(pageRecord, pagesCollection, payload); err != nil {
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
		return translationMap, nil
	}

	switch typed := rawValue.(type) {
	case []any:
		if containsNuvioCMSBackofficeFileLikePayload(typed) {
			return nil, fmt.Errorf("File upload payloads are not supported in this endpoint yet")
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
			if err := setNuvioCMSBackofficePageSEOStringField(pageRecord, pagesCollection, []string{"seo_title", "seoTitle"}, strings.TrimSpace(parseStringValue(rawValue))); err != nil {
				return err
			}
			updatedFields++
		case "seodescription":
			if err := setNuvioCMSBackofficePageSEOStringField(pageRecord, pagesCollection, []string{"seo_description", "seoDescription"}, parseStringValue(rawValue)); err != nil {
				return err
			}
			updatedFields++
		case "seosocialimage":
			if err := setNuvioCMSBackofficePageSEOSocialImageField(pageRecord, pagesCollection, rawValue); err != nil {
				return err
			}
			updatedFields++
		case "seocanonicalurl":
			if err := setNuvioCMSBackofficePageSEOStringField(pageRecord, pagesCollection, []string{"seo_canonical_url", "seoCanonicalUrl"}, strings.TrimSpace(parseStringValue(rawValue))); err != nil {
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
			if err := setNuvioCMSBackofficePageSEOStringField(pageRecord, pagesCollection, []string{"seo_focus_keyword", "seoFocusKeyword"}, strings.TrimSpace(parseStringValue(rawValue))); err != nil {
				return err
			}
			updatedFields++
		case "seotranslations":
			normalizedTranslations, err := normalizeNuvioCMSBackofficeSEOTranslationsValue(rawValue)
			if err != nil {
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
	pageRecord *core.Record,
	pagesCollection *core.Collection,
	rawValue any,
) error {
	fieldName := resolveNuvioCollectionFieldNameByAliases(pagesCollection, []string{"seo_social_image", "seoSocialImage"})
	if fieldName == "" {
		return fmt.Errorf("Field %q is not available for this pages collection", "seo_social_image")
	}

	field := pagesCollection.Fields.GetByName(fieldName)
	if field != nil && field.Type() == core.FieldTypeFile {
		return fmt.Errorf("seo_social_image file mutation is not supported in this endpoint yet")
	}

	switch rawValue.(type) {
	case map[string]any, []any:
		return fmt.Errorf("seo_social_image expects a string value")
	}

	stringValue := ""
	switch typed := rawValue.(type) {
	case nil:
		stringValue = ""
	case string:
		stringValue = strings.TrimSpace(typed)
	default:
		return fmt.Errorf("seo_social_image expects a string value")
	}

	pageRecord.Set(fieldName, stringValue)
	return nil
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

func applyNuvioCMSBackofficeIdentityPatch(record *core.Record, payload map[string]any, isAdmin bool) error {
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

		stringValue := strings.TrimSpace(parseStringValue(rawValue))
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
			if err := applyNuvioCMSBackofficeContactFormSettingsPatch(currentSettings, rawValue); err != nil {
				return nil, err
			}
		case "whatsapp":
			if err := applyNuvioCMSBackofficeWhatsAppSettingsPatch(currentSettings, rawValue); err != nil {
				return nil, err
			}
		case "newsletter":
			if err := applyNuvioCMSBackofficeNewsletterSettingsPatch(currentSettings, rawValue); err != nil {
				return nil, err
			}
		case "booking":
			if err := applyNuvioCMSBackofficeBookingSettingsPatch(currentSettings, rawValue); err != nil {
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

func applyNuvioCMSBackofficeContactFormSettingsPatch(settings map[string]any, rawPatch any) error {
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
			contactFormSettings["confirmationMessage"] = strings.TrimSpace(parseStringValue(rawValue))
		case "fields":
			if err := applyNuvioCMSBackofficeContactFormFieldsPatch(contactFormSettings, rawValue); err != nil {
				return err
			}
		case "emailnotifications":
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(contactFormSettings, rawValue, "template", true); err != nil {
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

func applyNuvioCMSBackofficeWhatsAppSettingsPatch(settings map[string]any, rawPatch any) error {
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
			whatsAppSettings["defaultMessage"] = strings.TrimSpace(parseStringValue(rawValue))
		case "showfloatingbutton":
			value, ok := parseBoolValue(rawValue)
			if !ok {
				return fmt.Errorf("whatsapp.showFloatingButton must be a boolean")
			}
			whatsAppSettings["showFloatingButton"] = value
		case "emailnotifications":
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(whatsAppSettings, rawValue, "template", true); err != nil {
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
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	settings["newsletter"] = newsletterSettings
	return nil
}

func applyNuvioCMSBackofficeBookingSettingsPatch(settings map[string]any, rawPatch any) error {
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
			if err := applyNuvioCMSBackofficeEmailNotificationsPatch(bookingSettings, rawValue, "businessTemplate", true); err != nil {
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
		case "rules":
			if err := applyNuvioCMSBackofficeBookingRulesPatch(bookingSettings, rawValue); err != nil {
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
			if provider != "umami" {
				return fmt.Errorf("reports.analytics.provider must be umami")
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
		case "apiurl", "scripturl":
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
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
		default:
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(rawKey))
		}
	}

	parentSettings["emailNotifications"] = emailNotifications
	return nil
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
			templateSettings["subject"] = strings.TrimSpace(parseStringValue(rawValue))
		case "introtext":
			templateSettings["introText"] = strings.TrimSpace(parseStringValue(rawValue))
		case "footertext":
			templateSettings["footerText"] = strings.TrimSpace(parseStringValue(rawValue))
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

		code := strings.TrimSpace(parseStringValue(item["code"]))
		if code == "" {
			return nil, fmt.Errorf("i18n.languages entries require code")
		}

		language := map[string]string{
			"code": code,
		}
		if label := strings.TrimSpace(parseStringValue(item["label"])); label != "" {
			language["label"] = label
		}

		languages = append(languages, language)
	}

	return languages, nil
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

	dto.Logo = toNuvioCMSDashboardSingleFileName(record.Get("logo"))
	dto.SEOTitle = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seoTitle", "seo_title"}))
	dto.SEODescription = strings.TrimSpace(parseNuvioCMSDashboardStringByAliases(record, []string{"seoDescription", "seo_description"}))
	dto.SEOImage = toNuvioCMSDashboardSingleFileName(readNuvioCMSDashboardValueByAliases(record, []string{"seoImage", "seo_image"}))
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
	dto.SEOSocialImage = toNuvioCMSDashboardSingleFileName(readNuvioCMSDashboardValueByAliases(record, []string{"seo_social_image", "seoSocialImage"}))
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
		if notifications := sanitizeNuvioCMSDashboardEmailNotifications(contactFormRaw["emailNotifications"]); len(notifications) > 0 {
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
		if notifications := sanitizeNuvioCMSDashboardEmailNotifications(whatsappRaw["emailNotifications"]); len(notifications) > 0 {
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
		if notifications := sanitizeNuvioCMSDashboardBookingEmailNotifications(bookingRaw["emailNotifications"]); len(notifications) > 0 {
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
		if len(i18n) > 0 {
			dto["i18n"] = i18n
		}
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

func sanitizeNuvioCMSDashboardEmailNotifications(raw any) map[string]any {
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

	return notifications
}

func sanitizeNuvioCMSDashboardBookingEmailNotifications(raw any) map[string]any {
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

	return notifications
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
	rawItems := normalizeNuvioPublicAnySlice(raw)
	if len(rawItems) == 0 {
		return []map[string]string{}
	}

	result := make([]map[string]string, 0, len(rawItems))
	for _, rawItem := range rawItems {
		entry, ok := toStringAnyMap(rawItem)
		if !ok {
			continue
		}

		code := strings.TrimSpace(parseStringValue(entry["code"]))
		if code == "" {
			continue
		}

		language := map[string]string{
			"code": code,
		}
		if label := strings.TrimSpace(parseStringValue(entry["label"])); label != "" {
			language["label"] = label
		}

		result = append(result, language)
	}

	return result
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

// NUVIO CUSTOM END: Scoped CMS backoffice dashboard endpoint (A3.5.8B).
