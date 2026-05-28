package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

const (
	nuvioLeadsDashboardMaxScan              = 5000
	nuvioLeadsDashboardFollowUpNotesMaxLen  = 4000
	nuvioLeadsDashboardForbiddenAccessError = "The authorized record is not allowed to perform this action."
)

var (
	nuvioLeadsDashboardDateTimeLayouts = []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.000Z",
	}
	nuvioLeadsDashboardContactsCollectionAliases = []string{
		nuvioContactsCollectionID,
		"contacts",
		"contact",
		"Contacts",
	}
	nuvioLeadsDashboardWhatsappCollectionAliases = []string{
		nuvioWhatsappCollectionID,
		"whatsapp",
		"Whatsapp",
		"WhatsApp",
		"whatsapp_interactions",
		"whatsappInteractions",
		"whatsapp_clicks",
		"WhatsAppInteractions",
	}
	nuvioLeadsDashboardPageFieldAliases = []string{
		"page",
		"pagePath",
		"page_path",
		"pageSlug",
		"page_slug",
	}
	nuvioLeadsDashboardSourceFieldAliases = []string{
		"source",
		"sourceLabel",
		"source_label",
		"origin",
	}
	nuvioLeadsDashboardNotesFieldAliases = []string{
		"notes",
		"note",
		"internalNotes",
		"internal_notes",
	}
	nuvioLeadsDashboardLastContactedFieldAliases = []string{
		"lastContactedAt",
		"last_contacted_at",
		"lastContacted",
		"last_contacted",
	}
	nuvioLeadsDashboardDefaultMessageFieldAliases = []string{
		"defaultMessage",
		"default_message",
		"prefilledMessage",
		"prefilled_message",
	}
	nuvioLeadsDashboardKnownStatusValues = []string{
		"new",
		"read",
		"archived",
	}
)

type nuvioLeadsDashboardContactDTO struct {
	ID              string `json:"id"`
	Website         string `json:"website"`
	Channel         string `json:"channel"`
	Status          string `json:"status"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Subject         string `json:"subject"`
	Message         string `json:"message"`
	Page            string `json:"page"`
	Source          string `json:"source"`
	Notes           string `json:"notes"`
	LastContactedAt string `json:"lastContactedAt"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
}

type nuvioLeadsDashboardWhatsappDTO struct {
	ID              string `json:"id"`
	Website         string `json:"website"`
	Status          string `json:"status"`
	Source          string `json:"source"`
	Page            string `json:"page"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Message         string `json:"message"`
	DefaultMessage  string `json:"defaultMessage"`
	Notes           string `json:"notes"`
	LastContactedAt string `json:"lastContactedAt"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
}

type nuvioLeadsDashboardDatasets struct {
	Contacts []nuvioLeadsDashboardContactDTO  `json:"contacts"`
	Whatsapp []nuvioLeadsDashboardWhatsappDTO `json:"whatsapp"`
}

type nuvioLeadsDashboardDatasetCapabilities struct {
	SupportsStatus   bool     `json:"supportsStatus"`
	SupportsArchive  bool     `json:"supportsArchive"`
	AllowedStatus    []string `json:"allowedStatus"`
	SupportsFollowUp bool     `json:"supportsFollowUp"`
}

type nuvioLeadsDashboardCapabilities struct {
	Contacts nuvioLeadsDashboardDatasetCapabilities `json:"contacts"`
	Whatsapp nuvioLeadsDashboardDatasetCapabilities `json:"whatsapp"`
}

type nuvioLeadsDashboardResponse struct {
	State        string                          `json:"state"`
	WebsiteID    string                          `json:"websiteId"`
	Datasets     nuvioLeadsDashboardDatasets     `json:"datasets"`
	Capabilities nuvioLeadsDashboardCapabilities `json:"capabilities"`
}

// NUVIO CUSTOM START: Scoped Leads dashboard endpoint for authenticated backoffice usage.
func registerNuvioLeadsDashboardRoutes(e *core.ServeEvent) {
	leadsDashboardGroup := e.Router.Group("/api/nuvio/leads").Bind(apis.RequireSuperuserAuth())

	leadsDashboardGroup.GET("/dashboard", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		contactsCollection, err := findNuvioLeadsDashboardCollectionByAliases(
			e.App,
			nuvioLeadsDashboardContactsCollectionAliases,
		)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO leads dashboard contacts collection resolve failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load Leads dashboard data.", nil)
		}

		whatsappCollection, err := findNuvioLeadsDashboardCollectionByAliases(
			e.App,
			nuvioLeadsDashboardWhatsappCollectionAliases,
		)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO leads dashboard whatsapp collection resolve failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load Leads dashboard data.", nil)
		}

		datasets, err := loadNuvioLeadsDashboardDatasets(e.App, websiteID, contactsCollection, whatsappCollection)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO leads dashboard datasets load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load Leads dashboard data.", nil)
		}

		response := nuvioLeadsDashboardResponse{
			State:     "ok",
			WebsiteID: websiteID,
			Datasets:  datasets,
			Capabilities: nuvioLeadsDashboardCapabilities{
				Contacts: buildNuvioLeadsDashboardCapabilities(contactsCollection),
				Whatsapp: buildNuvioLeadsDashboardCapabilities(whatsappCollection),
			},
		}

		return e.JSON(http.StatusOK, response)
	})

	leadsDashboardGroup.PATCH("/contacts/{id}/status", func(e *core.RequestEvent) error {
		return handleNuvioLeadsDashboardStatusUpdate(
			e,
			nuvioLeadsDashboardContactsCollectionAliases,
			func(record *core.Record) any {
				return buildNuvioLeadsDashboardContactDTO(record)
			},
		)
	})

	leadsDashboardGroup.PATCH("/whatsapp/{id}/status", func(e *core.RequestEvent) error {
		return handleNuvioLeadsDashboardStatusUpdate(
			e,
			nuvioLeadsDashboardWhatsappCollectionAliases,
			func(record *core.Record) any {
				return buildNuvioLeadsDashboardWhatsappDTO(record)
			},
		)
	})

	leadsDashboardGroup.PATCH("/contacts/{id}/follow-up", func(e *core.RequestEvent) error {
		return handleNuvioLeadsDashboardFollowUpUpdate(
			e,
			nuvioLeadsDashboardContactsCollectionAliases,
			func(record *core.Record) any {
				return buildNuvioLeadsDashboardContactDTO(record)
			},
		)
	})

	leadsDashboardGroup.PATCH("/whatsapp/{id}/follow-up", func(e *core.RequestEvent) error {
		return handleNuvioLeadsDashboardFollowUpUpdate(
			e,
			nuvioLeadsDashboardWhatsappCollectionAliases,
			func(record *core.Record) any {
				return buildNuvioLeadsDashboardWhatsappDTO(record)
			},
		)
	})
}

func handleNuvioLeadsDashboardStatusUpdate(
	e *core.RequestEvent,
	collectionAliases []string,
	dtoBuilder func(*core.Record) any,
) error {
	collection, record, err := resolveNuvioLeadsDashboardWriteTarget(e, collectionAliases)
	if err != nil {
		return err
	}

	statusFieldName, statusValue, err := parseNuvioLeadsDashboardStatusPayload(e, collection)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	record.Set(statusFieldName, statusValue)
	if saveErr := e.App.Save(record); saveErr != nil {
		return e.BadRequestError("Failed to update lead.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state": "ok",
		"lead":  dtoBuilder(record),
	})
}

func handleNuvioLeadsDashboardFollowUpUpdate(
	e *core.RequestEvent,
	collectionAliases []string,
	dtoBuilder func(*core.Record) any,
) error {
	collection, record, err := resolveNuvioLeadsDashboardWriteTarget(e, collectionAliases)
	if err != nil {
		return err
	}

	updates, err := parseNuvioLeadsDashboardFollowUpPayload(e, collection)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	for fieldName, fieldValue := range updates {
		record.Set(fieldName, fieldValue)
	}

	if saveErr := e.App.Save(record); saveErr != nil {
		return e.BadRequestError("Failed to update lead.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state": "ok",
		"lead":  dtoBuilder(record),
	})
}

func resolveNuvioLeadsDashboardWriteTarget(
	e *core.RequestEvent,
	collectionAliases []string,
) (*core.Collection, *core.Record, error) {
	recordID := strings.TrimSpace(e.Request.PathValue("id"))
	if recordID == "" {
		return nil, nil, e.BadRequestError("Missing record id.", nil)
	}

	collection, err := findNuvioLeadsDashboardCollectionByAliases(e.App, collectionAliases)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO leads dashboard write collection resolve failed",
			"recordId",
			recordID,
			"error",
			err.Error(),
		)
		return nil, nil, e.BadRequestError("Failed to update lead.", nil)
	}

	record, err := e.App.FindRecordById(collection.Id, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, e.NotFoundError("Lead not found.", nil)
		}

		e.App.Logger().Error(
			"NUVIO leads dashboard write record load failed",
			"collection",
			collection.Name,
			"recordId",
			recordID,
			"error",
			err.Error(),
		)
		return nil, nil, e.BadRequestError("Failed to update lead.", nil)
	}

	if err := requireNuvioLeadsDashboardRecordWebsiteAccess(e, record); err != nil {
		return nil, nil, err
	}

	return collection, record, nil
}

func requireNuvioLeadsDashboardRecordWebsiteAccess(e *core.RequestEvent, record *core.Record) error {
	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID == "" {
		return router.NewForbiddenError(nuvioLeadsDashboardForbiddenAccessError, nil)
	}

	return apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID)
}

func parseNuvioLeadsDashboardStatusPayload(
	e *core.RequestEvent,
	collection *core.Collection,
) (string, string, error) {
	statusFieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"status"})
	if statusFieldName == "" {
		return "", "", fmt.Errorf("Status updates are not supported for this lead type")
	}

	payload := map[string]any{}
	if err := e.BindBody(&payload); err != nil {
		return "", "", fmt.Errorf("Invalid request payload")
	}

	if len(payload) == 0 {
		return "", "", fmt.Errorf("Status is required")
	}

	for key := range payload {
		if strings.TrimSpace(key) != "status" {
			return "", "", fmt.Errorf("Only status can be updated in this endpoint")
		}
	}

	statusValue := strings.ToLower(strings.TrimSpace(parseStringValue(payload["status"])))
	if statusValue == "" {
		return "", "", fmt.Errorf("Status is required")
	}

	if !slices.Contains(nuvioLeadsDashboardKnownStatusValues, statusValue) {
		return "", "", fmt.Errorf("Invalid status. Allowed values: new, read, archived")
	}

	allowedStatuses := resolveNuvioLeadsDashboardAllowedStatuses(collection, statusFieldName)
	if len(allowedStatuses) > 0 && !slices.Contains(allowedStatuses, statusValue) {
		return "", "", fmt.Errorf("Status is not supported for this lead type")
	}

	return statusFieldName, statusValue, nil
}

func parseNuvioLeadsDashboardFollowUpPayload(
	e *core.RequestEvent,
	collection *core.Collection,
) (map[string]any, error) {
	notesFieldName := resolveNuvioCollectionFieldNameByAliases(collection, nuvioLeadsDashboardNotesFieldAliases)
	lastContactedFieldName := resolveNuvioCollectionFieldNameByAliases(collection, nuvioLeadsDashboardLastContactedFieldAliases)

	payload := map[string]any{}
	if err := e.BindBody(&payload); err != nil {
		return nil, fmt.Errorf("Invalid request payload")
	}

	if len(payload) == 0 {
		return nil, fmt.Errorf("At least one follow-up field is required")
	}

	for key := range payload {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey != "notes" && normalizedKey != "lastContactedAt" {
			return nil, fmt.Errorf("Only notes and lastContactedAt can be updated in this endpoint")
		}
	}

	updates := map[string]any{}

	if rawNotes, hasNotes := payload["notes"]; hasNotes {
		if notesFieldName == "" {
			return nil, fmt.Errorf("Follow-up notes are not supported for this lead type")
		}

		normalizedNotes := strings.TrimSpace(parseStringValue(rawNotes))
		if len([]rune(normalizedNotes)) > nuvioLeadsDashboardFollowUpNotesMaxLen {
			return nil, fmt.Errorf("Notes are too long. Maximum %d characters", nuvioLeadsDashboardFollowUpNotesMaxLen)
		}

		updates[notesFieldName] = normalizedNotes
	}

	if rawLastContactedAt, hasLastContactedAt := payload["lastContactedAt"]; hasLastContactedAt {
		if lastContactedFieldName == "" {
			return nil, fmt.Errorf("lastContactedAt is not supported for this lead type")
		}

		normalizedLastContactedAt := ""
		if rawLastContactedAt != nil {
			lastContactedAtString := strings.TrimSpace(parseStringValue(rawLastContactedAt))
			if lastContactedAtString != "" {
				parsedLastContactedAt, err := normalizeNuvioLeadsDashboardDateTime(lastContactedAtString)
				if err != nil {
					return nil, fmt.Errorf("Invalid lastContactedAt. Use ISO date-time format")
				}
				normalizedLastContactedAt = parsedLastContactedAt
			}
		}

		updates[lastContactedFieldName] = normalizedLastContactedAt
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("At least one follow-up field is required")
	}

	return updates, nil
}

func normalizeNuvioLeadsDashboardDateTime(rawValue string) (string, error) {
	trimmedValue := strings.TrimSpace(rawValue)
	if trimmedValue == "" {
		return "", nil
	}

	for _, layout := range nuvioLeadsDashboardDateTimeLayouts {
		if parsed, err := time.Parse(layout, trimmedValue); err == nil {
			return parsed.UTC().Format(time.RFC3339), nil
		}
	}

	return "", fmt.Errorf("invalid date-time")
}

func loadNuvioLeadsDashboardDatasets(
	app core.App,
	websiteID string,
	contactsCollection *core.Collection,
	whatsappCollection *core.Collection,
) (nuvioLeadsDashboardDatasets, error) {
	datasets := nuvioLeadsDashboardDatasets{
		Contacts: make([]nuvioLeadsDashboardContactDTO, 0),
		Whatsapp: make([]nuvioLeadsDashboardWhatsappDTO, 0),
	}

	contactsRecords, err := findNuvioLeadsDashboardRecordsByWebsite(app, contactsCollection, websiteID, "-created")
	if err != nil {
		return datasets, err
	}

	for _, record := range contactsRecords {
		datasets.Contacts = append(datasets.Contacts, buildNuvioLeadsDashboardContactDTO(record))
	}

	whatsappRecords, err := findNuvioLeadsDashboardRecordsByWebsite(app, whatsappCollection, websiteID, "-created")
	if err != nil {
		return datasets, err
	}

	for _, record := range whatsappRecords {
		datasets.Whatsapp = append(datasets.Whatsapp, buildNuvioLeadsDashboardWhatsappDTO(record))
	}

	return datasets, nil
}

func buildNuvioLeadsDashboardContactDTO(record *core.Record) nuvioLeadsDashboardContactDTO {
	if record == nil {
		return nuvioLeadsDashboardContactDTO{}
	}

	return nuvioLeadsDashboardContactDTO{
		ID:              strings.TrimSpace(record.Id),
		Website:         resolveNuvioPublicRelationID(record, "website", "site"),
		Channel:         strings.TrimSpace(record.GetString("channel")),
		Status:          strings.TrimSpace(record.GetString("status")),
		Name:            strings.TrimSpace(record.GetString("name")),
		Email:           strings.TrimSpace(record.GetString("email")),
		Phone:           strings.TrimSpace(record.GetString("phone")),
		Subject:         strings.TrimSpace(record.GetString("subject")),
		Message:         strings.TrimSpace(record.GetString("message")),
		Page:            readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardPageFieldAliases),
		Source:          readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardSourceFieldAliases),
		Notes:           readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardNotesFieldAliases),
		LastContactedAt: readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardLastContactedFieldAliases),
		Created:         strings.TrimSpace(record.GetString("created")),
		Updated:         strings.TrimSpace(record.GetString("updated")),
	}
}

func buildNuvioLeadsDashboardWhatsappDTO(record *core.Record) nuvioLeadsDashboardWhatsappDTO {
	if record == nil {
		return nuvioLeadsDashboardWhatsappDTO{}
	}

	return nuvioLeadsDashboardWhatsappDTO{
		ID:              strings.TrimSpace(record.Id),
		Website:         resolveNuvioPublicRelationID(record, "website", "site"),
		Status:          strings.TrimSpace(record.GetString("status")),
		Source:          readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardSourceFieldAliases),
		Page:            readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardPageFieldAliases),
		Name:            strings.TrimSpace(record.GetString("name")),
		Email:           strings.TrimSpace(record.GetString("email")),
		Phone:           strings.TrimSpace(record.GetString("phone")),
		Message:         strings.TrimSpace(record.GetString("message")),
		DefaultMessage:  readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardDefaultMessageFieldAliases),
		Notes:           readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardNotesFieldAliases),
		LastContactedAt: readNuvioLeadsDashboardRecordStringByAliases(record, nuvioLeadsDashboardLastContactedFieldAliases),
		Created:         strings.TrimSpace(record.GetString("created")),
		Updated:         strings.TrimSpace(record.GetString("updated")),
	}
}

func buildNuvioLeadsDashboardCapabilities(collection *core.Collection) nuvioLeadsDashboardDatasetCapabilities {
	statusFieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"status"})
	allowedStatus := resolveNuvioLeadsDashboardAllowedStatuses(collection, statusFieldName)
	notesFieldName := resolveNuvioCollectionFieldNameByAliases(collection, nuvioLeadsDashboardNotesFieldAliases)
	lastContactedFieldName := resolveNuvioCollectionFieldNameByAliases(collection, nuvioLeadsDashboardLastContactedFieldAliases)

	return nuvioLeadsDashboardDatasetCapabilities{
		SupportsStatus:   statusFieldName != "",
		SupportsArchive:  slices.Contains(allowedStatus, "archived"),
		AllowedStatus:    allowedStatus,
		SupportsFollowUp: notesFieldName != "" || lastContactedFieldName != "",
	}
}

func resolveNuvioLeadsDashboardAllowedStatuses(collection *core.Collection, statusFieldName string) []string {
	if collection == nil || statusFieldName == "" {
		return []string{}
	}

	allowed := make([]string, 0, len(nuvioLeadsDashboardKnownStatusValues))
	seen := map[string]struct{}{}

	statusField := collection.Fields.GetByName(statusFieldName)
	selectField, isSelect := statusField.(*core.SelectField)
	if isSelect {
		for _, rawValue := range selectField.Values {
			normalizedValue := strings.TrimSpace(strings.ToLower(rawValue))
			if normalizedValue == "" {
				continue
			}
			if !slices.Contains(nuvioLeadsDashboardKnownStatusValues, normalizedValue) {
				continue
			}
			if _, exists := seen[normalizedValue]; exists {
				continue
			}

			seen[normalizedValue] = struct{}{}
		}
	}

	if len(seen) == 0 {
		for _, fallbackValue := range nuvioLeadsDashboardKnownStatusValues {
			seen[fallbackValue] = struct{}{}
		}
	}

	for _, orderedValue := range nuvioLeadsDashboardKnownStatusValues {
		if _, exists := seen[orderedValue]; exists {
			allowed = append(allowed, orderedValue)
		}
	}

	return allowed
}

func findNuvioLeadsDashboardCollectionByAliases(app core.App, aliases []string) (*core.Collection, error) {
	var lastErr error

	for _, alias := range aliases {
		candidate := strings.TrimSpace(alias)
		if candidate == "" {
			continue
		}

		collection, err := app.FindCachedCollectionByNameOrId(candidate)
		if err == nil {
			return collection, nil
		}

		lastErr = err
		if !errors.Is(err, sql.ErrNoRows) {
			continue
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("collection not found for aliases")
}

func findNuvioLeadsDashboardRecordsByWebsite(
	app core.App,
	collection *core.Collection,
	websiteID string,
	sortExpr string,
) ([]*core.Record, error) {
	if collection == nil {
		return nil, fmt.Errorf("missing collection")
	}

	filter := "website={:websiteId}"
	params := dbx.Params{
		"websiteId": websiteID,
	}

	records, err := app.FindRecordsByFilter(
		collection,
		filter,
		sortExpr,
		nuvioLeadsDashboardMaxScan,
		0,
		params,
	)
	if err == nil {
		return records, nil
	}

	if strings.TrimSpace(sortExpr) != "" && strings.Contains(strings.ToLower(err.Error()), "invalid sort field") {
		return app.FindRecordsByFilter(
			collection,
			filter,
			"",
			nuvioLeadsDashboardMaxScan,
			0,
			params,
		)
	}

	return nil, err
}

func readNuvioLeadsDashboardRecordStringByAliases(record *core.Record, aliases []string) string {
	if record == nil {
		return ""
	}

	for _, alias := range aliases {
		fieldName := strings.TrimSpace(alias)
		if fieldName == "" {
			continue
		}

		value := strings.TrimSpace(record.GetString(fieldName))
		if value != "" {
			return value
		}

		if fallback := strings.TrimSpace(parseStringValue(record.Get(fieldName))); fallback != "" {
			return fallback
		}
	}

	return ""
}

// NUVIO CUSTOM END: Scoped Leads dashboard endpoint for authenticated backoffice usage.
