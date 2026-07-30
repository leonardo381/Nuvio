package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioPagesCollectionID               = "pbc_3945946014"
	nuvioBlocksCollectionID              = "pbc_4194232374"
	nuvioPublicContentMaxSitemapRecords  = 5000
	nuvioPublicContentMaxWebsiteScanRows = 5000
)

var nuvioPublicSlugPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

var (
	nuvioPublicContentAllowedQueryKeys = map[string]struct{}{
		"websiteslug": {},
		"pageslug":    {},
		"cmspreview":  {},
		"_cmspreview": {},
	}
	nuvioPublicSitemapAllowedQueryKeys = map[string]struct{}{}
	nuvioPublicSitemapExcludedStatuses = map[string]struct{}{
		"draft":       {},
		"disabled":    {},
		"inactive":    {},
		"archived":    {},
		"private":     {},
		"unpublished": {},
	}
)

type nuvioPublicContentResponse struct {
	Website map[string]any   `json:"website,omitempty"`
	Page    map[string]any   `json:"page,omitempty"`
	Blocks  []map[string]any `json:"blocks"`
}

type nuvioPublicSitemapDataResponse struct {
	Websites []map[string]any `json:"websites"`
	Pages    []map[string]any `json:"pages"`
}

// NUVIO CUSTOM START: Public content DTO endpoints (A3.1).
func registerNuvioPublicContentRoutes(e *core.ServeEvent) {
	publicGroup := e.Router.Group("/api/nuvio/public")

	publicGroup.GET("/content", func(e *core.RequestEvent) error {
		if err := validateNuvioPublicAllowedQueryKeys(e.Request.URL.Query(), nuvioPublicContentAllowedQueryKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteSlug := strings.TrimSpace(e.Request.URL.Query().Get("websiteSlug"))
		pageSlug := strings.TrimSpace(e.Request.URL.Query().Get("pageSlug"))

		if !isValidNuvioPublicSlug(websiteSlug) {
			return e.BadRequestError("Invalid website slug.", nil)
		}
		if pageSlug != "" && !isValidNuvioPublicSlug(pageSlug) {
			return e.BadRequestError("Invalid page slug.", nil)
		}

		websiteRecord, err := findNuvioPublicWebsiteBySlugOrDomain(e.App, websiteSlug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.InternalServerError("Failed to load public content.", nil)
		}

		response := nuvioPublicContentResponse{
			Website: buildNuvioPublicWebsiteDTO(websiteRecord),
			Blocks:  []map[string]any{},
		}

		if pageSlug == "" {
			return e.JSON(http.StatusOK, response)
		}

		pageRecord, err := findNuvioPublicPageBySlug(e.App, websiteRecord.Id, pageSlug)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Page not found.", nil)
			}
			return e.InternalServerError("Failed to load public content.", nil)
		}

		blockRecords, err := findNuvioPublicBlocksByPageID(e.App, pageRecord.Id)
		if err != nil {
			return e.InternalServerError("Failed to load public content.", nil)
		}

		response.Page = buildNuvioPublicPageDTO(pageRecord)
		response.Blocks = buildNuvioPublicBlocksDTO(blockRecords)

		return e.JSON(http.StatusOK, response)
	})

	publicGroup.GET("/sitemap-data", func(e *core.RequestEvent) error {
		if err := validateNuvioPublicAllowedQueryKeys(e.Request.URL.Query(), nuvioPublicSitemapAllowedQueryKeys); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websitesCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioWebsitesCollectionID)
		if err != nil {
			return e.InternalServerError("Failed to load sitemap data.", nil)
		}

		pagesCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioPagesCollectionID)
		if err != nil {
			return e.InternalServerError("Failed to load sitemap data.", nil)
		}

		websiteRecords, err := e.App.FindRecordsByFilter(
			websitesCollection,
			"id != ''",
			"slug",
			nuvioPublicContentMaxSitemapRecords,
			0,
			nil,
		)
		if err != nil {
			return e.InternalServerError("Failed to load sitemap data.", nil)
		}

		pageRecords, err := e.App.FindRecordsByFilter(
			pagesCollection,
			"id != ''",
			"slug",
			nuvioPublicContentMaxSitemapRecords,
			0,
			nil,
		)
		if err != nil {
			return e.InternalServerError("Failed to load sitemap data.", nil)
		}

		response := nuvioPublicSitemapDataResponse{
			Websites: make([]map[string]any, 0, len(websiteRecords)),
			Pages:    make([]map[string]any, 0, len(pageRecords)),
		}

		websiteSlugByID := make(map[string]string, len(websiteRecords))
		for _, website := range websiteRecords {
			if !isNuvioPublicSitemapRecordIndexable(website) {
				continue
			}

			websiteSlug := strings.TrimSpace(website.GetString("slug"))
			if !isValidNuvioPublicSlug(websiteSlug) {
				continue
			}

			websiteSlugByID[strings.TrimSpace(website.Id)] = websiteSlug
			response.Websites = append(response.Websites, buildNuvioPublicWebsiteSitemapDTO(website))
		}

		for _, page := range pageRecords {
			if !isNuvioPublicSitemapRecordIndexable(page) || isNuvioPublicSitemapPageExcluded(page) {
				continue
			}

			pageSlug := strings.TrimSpace(page.GetString("slug"))
			if !isValidNuvioPublicSlug(pageSlug) {
				continue
			}

			websiteSlug := resolveNuvioPublicSitemapPageWebsiteSlug(page, websiteSlugByID)
			if websiteSlug == "" {
				continue
			}

			response.Pages = append(response.Pages, buildNuvioPublicSitemapPageDTO(page, websiteSlug))
		}

		return e.JSON(http.StatusOK, response)
	})
}

func validateNuvioPublicAllowedQueryKeys(values url.Values, allowed map[string]struct{}) error {
	for key := range values {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if normalizedKey == "" {
			return errors.New("Invalid query parameter.")
		}
		if _, ok := allowed[normalizedKey]; !ok {
			return errors.New("Query parameter \"" + strings.TrimSpace(key) + "\" is not allowed.")
		}
	}

	return nil
}

func isValidNuvioPublicSlug(raw string) bool {
	return nuvioPublicSlugPattern.MatchString(strings.TrimSpace(raw))
}

func findNuvioPublicWebsiteBySlugOrDomain(app core.App, slugOrDomain string) (*core.Record, error) {
	websitesCollection, err := app.FindCachedCollectionByNameOrId(nuvioWebsitesCollectionID)
	if err != nil {
		return nil, err
	}

	normalizedCandidate := strings.TrimSpace(slugOrDomain)
	if normalizedCandidate == "" {
		return nil, sql.ErrNoRows
	}

	website, err := app.FindFirstRecordByFilter(
		websitesCollection,
		"slug={:slug}",
		dbx.Params{"slug": normalizedCandidate},
	)
	if err == nil {
		return website, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	candidates, err := app.FindRecordsByFilter(
		websitesCollection,
		"id != ''",
		"",
		nuvioPublicContentMaxWebsiteScanRows,
		0,
		nil,
	)
	if err != nil {
		return nil, err
	}

	for _, candidate := range candidates {
		if matchesNuvioPublicWebsiteDomain(candidate, normalizedCandidate) {
			return candidate, nil
		}
	}

	return nil, sql.ErrNoRows
}

func findNuvioPublicPageBySlug(app core.App, websiteID string, pageSlug string) (*core.Record, error) {
	pagesCollection, err := app.FindCachedCollectionByNameOrId(nuvioPagesCollectionID)
	if err != nil {
		return nil, err
	}

	websiteFieldName := resolveNuvioPublicPagesWebsiteFieldName(pagesCollection)
	if websiteFieldName == "" {
		return nil, sql.ErrNoRows
	}

	filter := websiteFieldName + "={:website} && slug={:slug}"
	return app.FindFirstRecordByFilter(
		pagesCollection,
		filter,
		dbx.Params{
			"website": strings.TrimSpace(websiteID),
			"slug":    strings.TrimSpace(pageSlug),
		},
	)
}

func findNuvioPublicBlocksByPageID(app core.App, pageID string) ([]*core.Record, error) {
	blocksCollection, err := app.FindCachedCollectionByNameOrId(nuvioBlocksCollectionID)
	if err != nil {
		return nil, err
	}

	trimmedPageID := strings.TrimSpace(pageID)
	if trimmedPageID == "" {
		return []*core.Record{}, nil
	}

	filterParts := []string{}
	filterParams := dbx.Params{
		"page": trimmedPageID,
	}

	if hasNuvioPublicCollectionField(blocksCollection, "page") {
		filterParts = append(filterParts, "page={:page}")
	}
	if hasNuvioPublicCollectionField(blocksCollection, "enabled") {
		filterParts = append(filterParts, "enabled={:enabled}")
		filterParams["enabled"] = true
	}

	filter := strings.Join(filterParts, " && ")
	if strings.TrimSpace(filter) == "" {
		return []*core.Record{}, nil
	}

	sortExpr := ""
	if hasNuvioPublicCollectionField(blocksCollection, "displayOrder") {
		sortExpr = "displayOrder"
	} else if hasNuvioPublicCollectionField(blocksCollection, "order") {
		sortExpr = "order"
	} else if hasNuvioPublicCollectionField(blocksCollection, "created") {
		sortExpr = "created"
	}

	blockRecords, err := app.FindRecordsByFilter(
		blocksCollection,
		filter,
		sortExpr,
		500,
		0,
		filterParams,
	)
	if err != nil {
		return nil, err
	}

	if hasNuvioPublicCollectionField(blocksCollection, "component") && len(blockRecords) > 0 {
		_ = app.ExpandRecords(blockRecords, []string{"component"}, nil)
	}

	return blockRecords, nil
}

func hasNuvioPublicCollectionField(collection *core.Collection, fieldName string) bool {
	if collection == nil {
		return false
	}
	return collection.Fields.GetByName(strings.TrimSpace(fieldName)) != nil
}

func readNuvioPublicBoolField(record *core.Record, fieldName string) any {
	if record == nil {
		return nil
	}
	if !hasNuvioPublicCollectionField(record.Collection(), fieldName) {
		return nil
	}
	return record.GetBool(fieldName)
}

func readNuvioPublicStringField(record *core.Record, fieldName string) string {
	if record == nil {
		return ""
	}
	if !hasNuvioPublicCollectionField(record.Collection(), fieldName) {
		return ""
	}
	return strings.TrimSpace(record.GetString(fieldName))
}
func resolveNuvioPublicPagesWebsiteFieldName(collection *core.Collection) string {
	if hasNuvioPublicCollectionField(collection, "website") {
		return "website"
	}
	if hasNuvioPublicCollectionField(collection, "site") {
		return "site"
	}
	return ""
}

func matchesNuvioPublicWebsiteDomain(record *core.Record, candidate string) bool {
	if record == nil {
		return false
	}

	normalizedCandidate := normalizeNuvioPublicDomainCandidate(candidate)
	if normalizedCandidate == "" {
		return false
	}

	domainFields := []string{
		"domain",
		"seo_canonical_domain",
		"public_url",
		"publicUrl",
		"url",
		"site_url",
		"website_url",
	}

	for _, fieldName := range domainFields {
		if normalizeNuvioPublicDomainCandidate(record.GetString(fieldName)) == normalizedCandidate {
			return true
		}
	}

	return false
}

func normalizeNuvioPublicDomainCandidate(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if normalized == "" {
		return ""
	}

	if strings.HasPrefix(normalized, "http://") {
		normalized = strings.TrimPrefix(normalized, "http://")
	}
	if strings.HasPrefix(normalized, "https://") {
		normalized = strings.TrimPrefix(normalized, "https://")
	}

	if slashIndex := strings.Index(normalized, "/"); slashIndex >= 0 {
		normalized = normalized[:slashIndex]
	}
	normalized = strings.TrimSpace(strings.TrimRight(normalized, "."))

	return normalized
}

func buildNuvioPublicWebsiteDTO(record *core.Record) map[string]any {
	if record == nil {
		return map[string]any{}
	}

	return map[string]any{
		"id":                       strings.TrimSpace(record.Id),
		"slug":                     strings.TrimSpace(record.GetString("slug")),
		"name":                     strings.TrimSpace(record.GetString("name")),
		"title":                    strings.TrimSpace(record.GetString("title")),
		"domain":                   strings.TrimSpace(record.GetString("domain")),
		"public_url":               strings.TrimSpace(record.GetString("public_url")),
		"publicUrl":                strings.TrimSpace(record.GetString("publicUrl")),
		"url":                      strings.TrimSpace(record.GetString("url")),
		"site_url":                 strings.TrimSpace(record.GetString("site_url")),
		"website_url":              strings.TrimSpace(record.GetString("website_url")),
		"logo":                     buildNuvioPublicFileRef(record, "logo"),
		"seoTitle":                 strings.TrimSpace(record.GetString("seoTitle")),
		"seoDescription":           strings.TrimSpace(record.GetString("seoDescription")),
		"seoImage":                 buildNuvioPublicFileRef(record, "seoImage"),
		"seo_title":                strings.TrimSpace(record.GetString("seo_title")),
		"seo_description":          strings.TrimSpace(record.GetString("seo_description")),
		"seo_image":                buildNuvioPublicFileRef(record, "seo_image"),
		"seo_canonical_domain":     strings.TrimSpace(record.GetString("seo_canonical_domain")),
		"seo_title_template":       strings.TrimSpace(record.GetString("seo_title_template")),
		"seo_title_separator":      strings.TrimSpace(record.GetString("seo_title_separator")),
		"business_name":            strings.TrimSpace(record.GetString("business_name")),
		"business_type":            strings.TrimSpace(record.GetString("business_type")),
		"business_phone":           strings.TrimSpace(record.GetString("business_phone")),
		"business_email":           strings.TrimSpace(record.GetString("business_email")),
		"business_address":         strings.TrimSpace(record.GetString("business_address")),
		"business_city":            strings.TrimSpace(record.GetString("business_city")),
		"business_postal_code":     strings.TrimSpace(record.GetString("business_postal_code")),
		"business_country":         strings.TrimSpace(record.GetString("business_country")),
		"business_service_area":    strings.TrimSpace(record.GetString("business_service_area")),
		"business_opening_hours":   strings.TrimSpace(record.GetString("business_opening_hours")),
		"business_google_place_id": strings.TrimSpace(record.GetString("business_google_place_id")),
		"business_social_profiles": strings.TrimSpace(record.GetString("business_social_profiles")),
		"business_price_range":     strings.TrimSpace(record.GetString("business_price_range")),
		"enabled":                  readNuvioPublicBoolField(record, "enabled"),
		"active":                   readNuvioPublicBoolField(record, "active"),
		"published":                readNuvioPublicBoolField(record, "published"),
		"status":                   readNuvioPublicStringField(record, "status"),
		"settings":                 buildNuvioPublicWebsiteSettingsDTO(record.Get("settings")),
	}
}

func buildNuvioPublicWebsiteSitemapDTO(record *core.Record) map[string]any {
	if record == nil {
		return map[string]any{}
	}

	return map[string]any{
		"slug":                 strings.TrimSpace(record.GetString("slug")),
		"domain":               strings.TrimSpace(record.GetString("domain")),
		"public_url":           strings.TrimSpace(record.GetString("public_url")),
		"publicUrl":            strings.TrimSpace(record.GetString("publicUrl")),
		"url":                  strings.TrimSpace(record.GetString("url")),
		"site_url":             strings.TrimSpace(record.GetString("site_url")),
		"website_url":          strings.TrimSpace(record.GetString("website_url")),
		"seo_canonical_domain": strings.TrimSpace(record.GetString("seo_canonical_domain")),
	}
}

func buildNuvioPublicPageDTO(record *core.Record) map[string]any {
	if record == nil {
		return map[string]any{}
	}

	websiteRelation := resolveNuvioPublicRelationID(record, "website", "site")

	return map[string]any{
		"id":                       strings.TrimSpace(record.Id),
		"slug":                     strings.TrimSpace(record.GetString("slug")),
		"title":                    strings.TrimSpace(record.GetString("title")),
		"name":                     strings.TrimSpace(record.GetString("name")),
		"website":                  websiteRelation,
		"websiteId":                websiteRelation,
		"seo_title":                strings.TrimSpace(record.GetString("seo_title")),
		"seo_description":          strings.TrimSpace(record.GetString("seo_description")),
		"seo_social_image":         buildNuvioPublicFileRef(record, "seo_social_image"),
		"seo_canonical_url":        strings.TrimSpace(record.GetString("seo_canonical_url")),
		"seo_noindex":              record.GetBool("seo_noindex"),
		"seo_exclude_from_sitemap": record.GetBool("seo_exclude_from_sitemap"),
		"seo_translations":         normalizeNuvioPublicJSONValue(record.Get("seo_translations")),
		"enabled":                  readNuvioPublicBoolField(record, "enabled"),
		"active":                   readNuvioPublicBoolField(record, "active"),
		"published":                readNuvioPublicBoolField(record, "published"),
		"status":                   readNuvioPublicStringField(record, "status"),
		"created":                  strings.TrimSpace(record.GetString("created")),
		"updated":                  strings.TrimSpace(record.GetString("updated")),
	}
}

func buildNuvioPublicSitemapPageDTO(record *core.Record, websiteSlug string) map[string]any {
	if record == nil {
		return map[string]any{}
	}

	return map[string]any{
		"websiteSlug":   strings.TrimSpace(websiteSlug),
		"slug":          strings.TrimSpace(record.GetString("slug")),
		"updated":       strings.TrimSpace(record.GetString("updated")),
		"updatedAt":     strings.TrimSpace(record.GetString("updatedAt")),
		"updated_at":    strings.TrimSpace(record.GetString("updated_at")),
		"modified":      strings.TrimSpace(record.GetString("modified")),
		"lastModified":  strings.TrimSpace(record.GetString("lastModified")),
		"last_modified": strings.TrimSpace(record.GetString("last_modified")),
	}
}

func isNuvioPublicSitemapRecordIndexable(record *core.Record) bool {
	if record == nil {
		return false
	}

	booleanFlags := []string{"enabled", "active", "published", "is_published", "isPublished"}
	for _, fieldName := range booleanFlags {
		if hasNuvioPublicCollectionField(record.Collection(), fieldName) && !record.GetBool(fieldName) {
			return false
		}
	}

	statusCandidates := []string{"status", "publication_status", "publishStatus"}
	for _, fieldName := range statusCandidates {
		if !hasNuvioPublicCollectionField(record.Collection(), fieldName) {
			continue
		}

		status := strings.ToLower(strings.TrimSpace(record.GetString(fieldName)))
		if _, isExcluded := nuvioPublicSitemapExcludedStatuses[status]; isExcluded {
			return false
		}
	}

	return true
}

func isNuvioPublicSitemapPageExcluded(record *core.Record) bool {
	if record == nil {
		return true
	}

	if hasNuvioPublicCollectionField(record.Collection(), "seo_noindex") && record.GetBool("seo_noindex") {
		return true
	}
	if hasNuvioPublicCollectionField(record.Collection(), "seo_exclude_from_sitemap") && record.GetBool("seo_exclude_from_sitemap") {
		return true
	}

	return false
}

func resolveNuvioPublicSitemapPageWebsiteSlug(record *core.Record, websiteSlugByID map[string]string) string {
	if record == nil || len(websiteSlugByID) == 0 {
		return ""
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID == "" {
		return ""
	}

	return strings.TrimSpace(websiteSlugByID[websiteID])
}

func buildNuvioPublicBlocksDTO(records []*core.Record) []map[string]any {
	if len(records) == 0 {
		return []map[string]any{}
	}

	result := make([]map[string]any, 0, len(records))

	for _, record := range records {
		if record == nil {
			continue
		}

		componentExpanded := map[string]any{}
		if componentRecord := record.ExpandedOne("component"); componentRecord != nil {
			componentExpanded = map[string]any{
				"id":            strings.TrimSpace(componentRecord.Id),
				"key":           strings.TrimSpace(componentRecord.GetString("key")),
				"name":          strings.TrimSpace(componentRecord.GetString("name")),
				"slug":          strings.TrimSpace(componentRecord.GetString("slug")),
				"component_key": strings.TrimSpace(componentRecord.GetString("component_key")),
			}
		}

		variant := strings.TrimSpace(record.GetString("variant"))
		if variant == "" {
			variant = strings.TrimSpace(record.GetString("component_variant"))
		}

		block := map[string]any{
			"id":            strings.TrimSpace(record.Id),
			"page":          resolveNuvioPublicRelationID(record, "page"),
			"slot":          strings.TrimSpace(record.GetString("slot")),
			"enabled":       record.GetBool("enabled"),
			"component":     resolveNuvioPublicRelationID(record, "component"),
			"component_key": strings.TrimSpace(record.GetString("component_key")),
			"variant":       variant,
			"props":         normalizeNuvioPublicJSONValue(record.Get("props")),
			"translations":  normalizeNuvioPublicJSONValue(record.Get("translations")),
			"title":         strings.TrimSpace(record.GetString("title")),
			"image":         buildNuvioPublicFileRef(record, "image"),
			"created":       strings.TrimSpace(record.GetString("created")),
			"updated":       strings.TrimSpace(record.GetString("updated")),
		}

		if len(componentExpanded) > 0 {
			block["expand"] = map[string]any{
				"component": componentExpanded,
			}
		}

		result = append(result, block)
	}

	return result
}

func buildNuvioPublicFileRef(record *core.Record, fieldName string) any {
	if record == nil {
		return ""
	}

	filename := strings.TrimSpace(record.GetString(fieldName))
	if filename == "" {
		return ""
	}

	collectionName := ""
	if collection := record.Collection(); collection != nil {
		collectionName = strings.TrimSpace(collection.Name)
	}

	if collectionName == "" {
		return filename
	}

	return map[string]string{
		"recordId":   strings.TrimSpace(record.Id),
		"filename":   filename,
		"collection": collectionName,
	}
}

func buildNuvioPublicWebsiteSettingsDTO(rawSettings any) map[string]any {
	settings := parseNuvioSettingsObject(rawSettings)
	publicSettings := map[string]any{}

	if featureFlagsRaw, ok := toStringAnyMap(settings["featureFlags"]); ok {
		featureFlags := map[string]any{}

		if value, ok := parseBoolValue(featureFlagsRaw["contactForm"]); ok {
			featureFlags["contactForm"] = value
		}
		if value, ok := parseBoolValue(featureFlagsRaw["whatsapp"]); ok {
			featureFlags["whatsapp"] = value
		}
		if value, ok := parseBoolValue(featureFlagsRaw["reports"]); ok {
			featureFlags["reports"] = value
		}

		if len(featureFlags) > 0 {
			publicSettings["featureFlags"] = featureFlags
		}
	}

	if i18nRaw, ok := toStringAnyMap(settings["i18n"]); ok {
		i18n := map[string]any{}

		if value, ok := parseBoolValue(i18nRaw["enabled"]); ok {
			i18n["enabled"] = value
		}

		normalizedLanguagesInput := normalizeNuvioPublicAnySlice(i18nRaw["languages"])
		if len(normalizedLanguagesInput) > 0 {
			normalizedLanguages := make([]map[string]string, 0, len(normalizedLanguagesInput))
			for _, item := range normalizedLanguagesInput {
				itemMap, ok := toStringAnyMap(item)
				if !ok {
					continue
				}

				code := strings.TrimSpace(parseStringValue(itemMap["code"]))
				label := strings.TrimSpace(parseStringValue(itemMap["label"]))
				if code == "" {
					continue
				}

				entry := map[string]string{"code": code}
				if label != "" {
					entry["label"] = label
				}
				normalizedLanguages = append(normalizedLanguages, entry)
			}

			if len(normalizedLanguages) > 0 {
				i18n["languages"] = normalizedLanguages
			}
		}

		if len(i18n) > 0 {
			publicSettings["i18n"] = i18n
		}
	}

	if contactFormRaw, ok := toStringAnyMap(settings["contactForm"]); ok {
		contactForm := map[string]any{}
		if value, ok := parseBoolValue(contactFormRaw["enabled"]); ok {
			contactForm["enabled"] = value
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

		if message := strings.TrimSpace(parseStringValue(contactFormRaw["confirmationMessage"])); message != "" {
			contactForm["confirmationMessage"] = message
		}

		if len(contactForm) > 0 {
			publicSettings["contactForm"] = contactForm
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

		if len(whatsapp) > 0 {
			publicSettings["whatsapp"] = whatsapp
		}
	}

	if reportsRaw, ok := toStringAnyMap(settings["reports"]); ok {
		if analyticsRaw, ok := toStringAnyMap(reportsRaw["analytics"]); ok {
			analytics := map[string]any{}

			if value, ok := parseBoolValue(analyticsRaw["enabled"]); ok {
				analytics["enabled"] = value
			}

			provider := strings.ToLower(strings.TrimSpace(parseStringValue(analyticsRaw["provider"])))
			if provider == "umami" {
				analytics["provider"] = "umami"
			}

			siteID := strings.TrimSpace(parseStringValue(analyticsRaw["siteId"]))
			if siteID != "" {
				analytics["siteId"] = siteID
			}

			if value, ok := parseBoolValue(analyticsRaw["scriptEnabled"]); ok {
				analytics["scriptEnabled"] = value
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

			if len(analytics) > 0 {
				publicSettings["reports"] = map[string]any{
					"analytics": analytics,
				}
			}
		}
	}

	return publicSettings
}

func resolveNuvioPublicRelationID(record *core.Record, fieldNames ...string) string {
	if record == nil {
		return ""
	}

	for _, fieldName := range fieldNames {
		normalizedField := strings.TrimSpace(fieldName)
		if normalizedField == "" {
			continue
		}

		relationIDs := record.GetStringSlice(normalizedField)
		for _, relationID := range relationIDs {
			trimmedID := strings.TrimSpace(relationID)
			if trimmedID != "" {
				return trimmedID
			}
		}

		if relationID := strings.TrimSpace(record.GetString(normalizedField)); relationID != "" {
			return relationID
		}
	}

	return ""
}

func normalizeNuvioPublicAnySlice(value any) []any {
	switch typed := value.(type) {
	case []any:
		return typed
	default:
		return []any{}
	}
}

func normalizeNuvioPublicJSONValue(value any) any {
	if value == nil {
		return map[string]any{}
	}

	if directMap, ok := toStringAnyMap(value); ok {
		return directMap
	}

	switch typed := value.(type) {
	case []any:
		return typed
	case string:
		normalized := strings.TrimSpace(typed)
		if normalized == "" {
			return map[string]any{}
		}

		var parsed any
		if err := json.Unmarshal([]byte(normalized), &parsed); err == nil {
			return parsed
		}
		return map[string]any{}
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return map[string]any{}
		}

		var parsed any
		if err := json.Unmarshal(encoded, &parsed); err != nil {
			return map[string]any{}
		}

		return parsed
	}
}

// NUVIO CUSTOM END: Public content DTO endpoints (A3.1).
