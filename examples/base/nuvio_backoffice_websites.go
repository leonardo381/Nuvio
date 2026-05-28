package main

import (
	"net/http"
	"sort"
	"strings"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

const nuvioBackofficeWebsiteSelectorMaxScan = 5000

type nuvioBackofficeWebsiteSelectorDTO struct {
	ID          string            `json:"id"`
	DisplayName string            `json:"displayName"`
	Name        string            `json:"name,omitempty"`
	Title       string            `json:"title,omitempty"`
	Slug        string            `json:"slug,omitempty"`
	Domain      string            `json:"domain,omitempty"`
	Status      string            `json:"status,omitempty"`
	Logo        map[string]string `json:"logo,omitempty"`
}

// NUVIO CUSTOM START: Scoped backoffice websites selector endpoint (A3.5.2A).
func registerNuvioBackofficeWebsitesRoutes(e *core.ServeEvent) {
	backofficeGroup := e.Router.Group("/api/nuvio/backoffice").Bind(apis.RequireSuperuserAuth())

	backofficeGroup.GET("/websites", func(e *core.RequestEvent) error {
		websites, err := listNuvioBackofficeWebsitesForAuth(e.App, e.Auth)
		if err != nil {
			return err
		}

		result := make([]nuvioBackofficeWebsiteSelectorDTO, 0, len(websites))
		for _, website := range websites {
			result = append(result, buildNuvioBackofficeWebsiteSelectorDTO(website))
		}

		return e.JSON(http.StatusOK, result)
	})
}

func listNuvioBackofficeWebsitesForAuth(app core.App, authRecord *core.Record) ([]*core.Record, error) {
	if authRecord == nil {
		return nil, router.NewUnauthorizedError("The request requires valid record authorization token.", nil)
	}

	resolvedAuthRecord, err := resolveNuvioBackofficeWebsitesAuthRecord(app, authRecord)
	if err != nil {
		return nil, err
	}

	if apis.IsAdminSuperuser(resolvedAuthRecord) {
		websitesCollection, err := app.FindCachedCollectionByNameOrId(nuvioWebsitesCollectionID)
		if err != nil {
			return nil, err
		}

		records, err := app.FindRecordsByFilter(
			websitesCollection,
			"",
			"",
			nuvioBackofficeWebsiteSelectorMaxScan,
			0,
			nil,
		)
		if err != nil {
			return nil, err
		}

		sortNuvioBackofficeWebsiteRecords(records)
		return records, nil
	}

	if !apis.IsClientSuperuser(resolvedAuthRecord) {
		return nil, router.NewForbiddenError("The authorized record is not allowed to perform this action.", nil)
	}

	allowedWebsiteIDs := normalizeNuvioBackofficeWebsiteAccessIDs(
		resolvedAuthRecord.Get("websiteAccess"),
		resolvedAuthRecord.ExpandedAll("websiteAccess"),
	)
	if len(allowedWebsiteIDs) == 0 {
		return []*core.Record{}, nil
	}

	records, err := app.FindRecordsByIds(nuvioWebsitesCollectionID, allowedWebsiteIDs)
	if err != nil {
		return nil, err
	}

	sortNuvioBackofficeWebsiteRecords(records)
	return records, nil
}

func resolveNuvioBackofficeWebsitesAuthRecord(app core.App, authRecord *core.Record) (*core.Record, error) {
	if authRecord == nil {
		return nil, router.NewUnauthorizedError("The request requires valid record authorization token.", nil)
	}

	authID := strings.TrimSpace(authRecord.Id)
	if authID == "" {
		return authRecord, nil
	}

	freshRecord, err := app.FindRecordById(core.CollectionNameSuperusers, authID)
	if err != nil {
		return nil, err
	}

	return freshRecord, nil
}

func normalizeNuvioBackofficeWebsiteAccessIDs(rawValues ...any) []string {
	if len(rawValues) == 0 {
		return []string{}
	}

	unique := map[string]struct{}{}
	result := make([]string, 0, len(rawValues))

	appendID := func(rawID string) {
		normalizedID := strings.TrimSpace(rawID)
		if normalizedID == "" {
			return
		}
		if _, exists := unique[normalizedID]; exists {
			return
		}

		unique[normalizedID] = struct{}{}
		result = append(result, normalizedID)
	}

	var visit func(raw any)
	visit = func(raw any) {
		switch typed := raw.(type) {
		case nil:
			return
		case string:
			appendID(typed)
		case []string:
			for _, id := range typed {
				appendID(id)
			}
		case []any:
			for _, item := range typed {
				visit(item)
			}
		case *core.Record:
			if typed == nil {
				return
			}
			appendID(typed.Id)
		case []*core.Record:
			for _, record := range typed {
				if record == nil {
					continue
				}
				appendID(record.Id)
			}
		case map[string]any:
			for _, key := range []string{"id", "recordId", "websiteId"} {
				if value, ok := typed[key]; ok {
					visit(value)
				}
			}
		default:
			return
		}
	}

	for _, raw := range rawValues {
		visit(raw)
	}

	return result
}

func sortNuvioBackofficeWebsiteRecords(records []*core.Record) {
	sort.Slice(records, func(i int, j int) bool {
		left := normalizeNuvioBackofficeWebsiteSortKey(records[i])
		right := normalizeNuvioBackofficeWebsiteSortKey(records[j])
		if left == right {
			return strings.TrimSpace(records[i].Id) < strings.TrimSpace(records[j].Id)
		}
		return left < right
	})
}

func normalizeNuvioBackofficeWebsiteSortKey(record *core.Record) string {
	if record == nil {
		return ""
	}

	return strings.ToLower(strings.TrimSpace(resolveNuvioBackofficeWebsiteDisplayName(record)))
}

func resolveNuvioBackofficeWebsiteDisplayName(record *core.Record) string {
	if record == nil {
		return ""
	}

	for _, fieldName := range []string{"title", "name", "slug"} {
		value := strings.TrimSpace(record.GetString(fieldName))
		if value != "" {
			return value
		}
	}

	return strings.TrimSpace(record.Id)
}

func buildNuvioBackofficeWebsiteSelectorDTO(record *core.Record) nuvioBackofficeWebsiteSelectorDTO {
	dto := nuvioBackofficeWebsiteSelectorDTO{
		ID:          "",
		DisplayName: "",
		Name:        "",
		Title:       "",
		Slug:        "",
		Domain:      "",
		Status:      "",
	}

	if record == nil {
		return dto
	}

	dto.ID = strings.TrimSpace(record.Id)
	dto.DisplayName = resolveNuvioBackofficeWebsiteDisplayName(record)
	dto.Name = strings.TrimSpace(record.GetString("name"))
	dto.Title = strings.TrimSpace(record.GetString("title"))
	dto.Slug = strings.TrimSpace(record.GetString("slug"))
	dto.Domain = strings.TrimSpace(record.GetString("domain"))
	dto.Status = strings.TrimSpace(record.GetString("status"))

	logoFilename := strings.TrimSpace(record.GetString("logo"))
	if logoFilename != "" {
		collectionName := ""
		if collection := record.Collection(); collection != nil {
			collectionName = strings.TrimSpace(collection.Name)
		}

		dto.Logo = map[string]string{
			"recordId":   dto.ID,
			"filename":   logoFilename,
			"collection": collectionName,
		}
	}

	return dto
}

// NUVIO CUSTOM END: Scoped backoffice websites selector endpoint (A3.5.2A).
