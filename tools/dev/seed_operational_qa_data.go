package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	seedConfirmToken = "SEED_OPERATIONAL_QA_DATA"

	collectionWebsites            = "pbc_2619338178"
	collectionContacts            = "pbc_1661203100"
	collectionWhatsapp            = "pbc_1661203200"
	collectionBookingServices     = "pbc_1661203700"
	collectionBookingAvailability = "pbc_1661203800"
	collectionBookingExceptions   = "pbc_1778803400"
	collectionAppointments        = "pbc_1661203900"
	collectionSubscribers         = "pbc_1661203400"
	collectionSubscriberGroups    = "pbc_1661203600"
	collectionCampaigns           = "pbc_1661203500"
)

type seedOptions struct {
	BaseURL           string
	SuperuserEmail    string
	SuperuserPassword string
	WebsiteSlug       string
	WebsiteID         string
	Confirm           string
	AllowResetEnv     string
	DryRun            bool
}

type pbClient struct {
	baseURL string
	token   string
	client  *http.Client
}

type pbListResponse struct {
	Page       int                      `json:"page"`
	PerPage    int                      `json:"perPage"`
	TotalPages int                      `json:"totalPages"`
	TotalItems int                      `json:"totalItems"`
	Items      []map[string]interface{} `json:"items"`
}

type pbCollectionField struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type pbCollectionMeta struct {
	ID     string              `json:"id"`
	Name   string              `json:"name"`
	Fields []pbCollectionField `json:"fields"`
}

type collectionSummary struct {
	Created int
	Updated int
	Skipped int
}

type seedStats struct {
	ByCollection map[string]*collectionSummary
	SkippedNotes []string
}

type seedRunner struct {
	opts               seedOptions
	client             *pbClient
	collectionMeta     map[string]pbCollectionMeta
	collectionFieldMap map[string]map[string]bool
	stats              seedStats
}

type bookingServiceSeed struct {
	Key          string
	Name         string
	Duration     int
	Active       bool
	Description  string
	DisplayOrder int
	SyntheticID  string
	PersistedID  string
}

type subscriberGroupSeed struct {
	Key         string
	Name        string
	Slug        string
	SyntheticID string
	PersistedID string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := parseSeedOptions()
	if err != nil {
		return err
	}

	if opts.AllowResetEnv != "1" {
		return errors.New("NUVIO_ALLOW_DEV_RESET must be set to 1")
	}

	if opts.SuperuserEmail == "" {
		return errors.New("PB_SUPERUSER_EMAIL is required")
	}
	if opts.SuperuserPassword == "" {
		return errors.New("PB_SUPERUSER_PASSWORD is required")
	}

	if opts.WebsiteSlug == "" && opts.WebsiteID == "" {
		return errors.New("provide --websiteSlug or --websiteId")
	}
	if opts.WebsiteSlug != "" && opts.WebsiteID != "" {
		return errors.New("provide only one of --websiteSlug or --websiteId")
	}

	if opts.Confirm != "" && opts.Confirm != seedConfirmToken {
		return fmt.Errorf("invalid --confirm token; expected %s", seedConfirmToken)
	}

	opts.DryRun = opts.Confirm != seedConfirmToken

	if !isLikelyLocalBaseURL(opts.BaseURL) {
		if opts.DryRun {
			fmt.Printf("WARNING: base URL does not look local/dev: %s\n", opts.BaseURL)
		} else {
			return fmt.Errorf("write mode refused because base URL does not look local/dev: %s", opts.BaseURL)
		}
	}

	httpClient := &http.Client{Timeout: 20 * time.Second}
	client := &pbClient{
		baseURL: strings.TrimRight(opts.BaseURL, "/"),
		client:  httpClient,
	}

	if err := client.authSuperuser(opts.SuperuserEmail, opts.SuperuserPassword); err != nil {
		return fmt.Errorf("superuser auth failed: %w", err)
	}

	runner := &seedRunner{
		opts:               opts,
		client:             client,
		collectionMeta:     map[string]pbCollectionMeta{},
		collectionFieldMap: map[string]map[string]bool{},
		stats: seedStats{
			ByCollection: map[string]*collectionSummary{},
			SkippedNotes: []string{},
		},
	}

	if err := runner.preflightCollections(); err != nil {
		return err
	}

	websiteRecord, err := runner.resolveWebsite()
	if err != nil {
		return err
	}

	websiteID := strings.TrimSpace(toString(websiteRecord["id"]))
	websiteSlug := strings.TrimSpace(toString(websiteRecord["slug"]))

	fmt.Printf("Seed mode: %s\n", map[bool]string{true: "DRY-RUN", false: "WRITE"}[opts.DryRun])
	fmt.Printf("Target website: slug=%q id=%q\n", websiteSlug, websiteID)
	fmt.Println("Scope: Leads + Booking + Newsletter only (CMS content untouched)")

	if err := runner.seedLeads(websiteID); err != nil {
		return err
	}
	if err := runner.seedBooking(websiteID); err != nil {
		return err
	}
	if err := runner.seedNewsletter(websiteID); err != nil {
		return err
	}

	runner.printSummary(websiteSlug, websiteID)
	return nil
}

func parseSeedOptions() (seedOptions, error) {
	var opts seedOptions
	flag.StringVar(&opts.WebsiteSlug, "websiteSlug", "", "Target existing website slug")
	flag.StringVar(&opts.WebsiteID, "websiteId", "", "Target existing website id")
	flag.StringVar(&opts.Confirm, "confirm", "", "Confirmation token required for write mode")
	flag.Parse()

	opts.BaseURL = strings.TrimSpace(os.Getenv("NUVIO_QA_BASE_URL"))
	if opts.BaseURL == "" {
		opts.BaseURL = strings.TrimSpace(os.Getenv("PB_URL"))
	}
	if opts.BaseURL == "" {
		opts.BaseURL = "http://127.0.0.1:8090"
	}

	opts.SuperuserEmail = strings.TrimSpace(os.Getenv("PB_SUPERUSER_EMAIL"))
	opts.SuperuserPassword = strings.TrimSpace(os.Getenv("PB_SUPERUSER_PASSWORD"))
	opts.AllowResetEnv = strings.TrimSpace(os.Getenv("NUVIO_ALLOW_DEV_RESET"))

	return opts, nil
}

func isLikelyLocalBaseURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	if host == "" {
		return false
	}
	host = strings.ToLower(host)
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	if strings.HasSuffix(host, ".local") {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() {
		return true
	}
	if isPrivateIP(ip) {
		return true
	}
	return false
}

func isPrivateIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		switch {
		case ip4[0] == 10:
			return true
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			return true
		case ip4[0] == 192 && ip4[1] == 168:
			return true
		}
	}
	return false
}

func (c *pbClient) authSuperuser(email string, password string) error {
	payload := map[string]string{
		"identity": email,
		"password": password,
	}

	endpoints := []string{
		"/api/collections/_superusers/auth-with-password",
		"/api/admins/auth-with-password",
	}

	var lastErr error
	for _, endpoint := range endpoints {
		var resp map[string]interface{}
		err := c.requestJSON(http.MethodPost, endpoint, nil, payload, false, &resp)
		if err != nil {
			lastErr = err
			continue
		}

		token := strings.TrimSpace(toString(resp["token"]))
		if token == "" {
			lastErr = fmt.Errorf("no token in auth response from %s", endpoint)
			continue
		}
		c.token = token
		return nil
	}

	if lastErr == nil {
		lastErr = errors.New("unknown auth failure")
	}
	return lastErr
}

func (c *pbClient) requestJSON(
	method string,
	path string,
	query map[string]string,
	body interface{},
	allowNotFound bool,
	out interface{},
) error {
	u := c.baseURL + path
	if len(query) > 0 {
		values := url.Values{}
		for k, v := range query {
			values.Set(k, v)
		}
		u += "?" + values.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(c.token) != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return readErr
	}

	if resp.StatusCode == http.StatusNotFound && allowNotFound {
		return errNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed %s %s: status %d body=%s", method, path, resp.StatusCode, sanitizeErrorBody(raw))
	}

	if out == nil {
		return nil
	}
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

var errNotFound = errors.New("not found")

func sanitizeErrorBody(raw []byte) string {
	text := strings.TrimSpace(string(raw))
	if len(text) > 240 {
		return text[:240] + "..."
	}
	return text
}

func (r *seedRunner) preflightCollections() error {
	requiredCollections := []string{
		collectionWebsites,
		collectionContacts,
		collectionWhatsapp,
		collectionBookingServices,
		collectionBookingAvailability,
		collectionBookingExceptions,
		collectionAppointments,
		collectionSubscribers,
		collectionSubscriberGroups,
		collectionCampaigns,
	}

	for _, collectionID := range requiredCollections {
		meta, err := r.getCollectionMeta(collectionID)
		if err != nil {
			return fmt.Errorf("required collection missing or inaccessible (%s): %w", collectionID, err)
		}
		r.collectionMeta[collectionID] = meta
		r.collectionFieldMap[collectionID] = makeFieldMap(meta.Fields)
	}

	requiredSubscriberFields := []string{
		"confirmationTokenHash",
		"confirmationTokenExpiresAt",
		"unsubscribeTokenHash",
		"unsubscribedAt",
	}
	missingLifecycle := []string{}
	for _, field := range requiredSubscriberFields {
		if !r.hasField(collectionSubscribers, field) {
			missingLifecycle = append(missingLifecycle, field)
		}
	}
	if len(missingLifecycle) > 0 {
		return errors.New("Newsletter lifecycle migration is not applied. Apply migration 1779303900_updated_Subscribers_lifecycle.js before seeding newsletter lifecycle data.")
	}

	if !r.hasField(collectionAppointments, "archivedAt") {
		return errors.New("Appointments archivedAt field is missing. Apply migration 1779203800_updated_Appointments_archivedAt.js before seeding archived appointments.")
	}

	return nil
}

func (r *seedRunner) getCollectionMeta(collectionID string) (pbCollectionMeta, error) {
	var meta pbCollectionMeta
	err := r.client.requestJSON(
		http.MethodGet,
		"/api/collections/"+url.PathEscape(collectionID),
		nil,
		nil,
		false,
		&meta,
	)
	if err != nil {
		return pbCollectionMeta{}, err
	}
	return meta, nil
}

func makeFieldMap(fields []pbCollectionField) map[string]bool {
	result := map[string]bool{}
	for _, field := range fields {
		name := strings.TrimSpace(field.Name)
		if name != "" {
			result[name] = true
			result[strings.ToLower(name)] = true
		}
	}
	return result
}

func (r *seedRunner) hasField(collectionID string, fieldName string) bool {
	fieldMap := r.collectionFieldMap[collectionID]
	if fieldMap == nil {
		return false
	}
	if fieldMap[fieldName] {
		return true
	}
	return fieldMap[strings.ToLower(fieldName)]
}

func (r *seedRunner) resolveWebsite() (map[string]interface{}, error) {
	if r.opts.WebsiteID != "" {
		record, err := r.getRecordByID(collectionWebsites, r.opts.WebsiteID)
		if err != nil {
			return nil, fmt.Errorf("website not found by id %q: %w", r.opts.WebsiteID, err)
		}
		return record, nil
	}

	slug := strings.TrimSpace(r.opts.WebsiteSlug)
	filter := fmt.Sprintf("slug=%s", pbFilterString(slug))
	record, err := r.findFirstRecord(collectionWebsites, filter, "")
	if err == nil && record != nil {
		return record, nil
	}

	records, listErr := r.listRecords(collectionWebsites, "", 200, 1, "")
	if listErr != nil {
		return nil, fmt.Errorf("website slug lookup failed: %w", listErr)
	}
	for _, item := range records.Items {
		if strings.EqualFold(strings.TrimSpace(toString(item["slug"])), slug) {
			return item, nil
		}
	}

	return nil, fmt.Errorf("website not found by slug %q", slug)
}

func (r *seedRunner) getRecordByID(collectionID string, id string) (map[string]interface{}, error) {
	var record map[string]interface{}
	err := r.client.requestJSON(
		http.MethodGet,
		"/api/collections/"+url.PathEscape(collectionID)+"/records/"+url.PathEscape(strings.TrimSpace(id)),
		nil,
		nil,
		false,
		&record,
	)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *seedRunner) listRecords(collectionID string, filter string, perPage int, page int, sortExpr string) (pbListResponse, error) {
	query := map[string]string{
		"perPage": strconv.Itoa(perPage),
		"page":    strconv.Itoa(page),
	}
	if strings.TrimSpace(filter) != "" {
		query["filter"] = filter
	}
	if strings.TrimSpace(sortExpr) != "" {
		query["sort"] = sortExpr
	}

	var list pbListResponse
	err := r.client.requestJSON(
		http.MethodGet,
		"/api/collections/"+url.PathEscape(collectionID)+"/records",
		query,
		nil,
		false,
		&list,
	)
	return list, err
}

func (r *seedRunner) findFirstRecord(collectionID string, filter string, sortExpr string) (map[string]interface{}, error) {
	list, err := r.listRecords(collectionID, filter, 1, 1, sortExpr)
	if err != nil {
		return nil, err
	}
	if len(list.Items) == 0 {
		return nil, errNotFound
	}
	return list.Items[0], nil
}

func (r *seedRunner) createRecord(collectionID string, payload map[string]interface{}) (map[string]interface{}, error) {
	var created map[string]interface{}
	err := r.client.requestJSON(
		http.MethodPost,
		"/api/collections/"+url.PathEscape(collectionID)+"/records",
		nil,
		payload,
		false,
		&created,
	)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (r *seedRunner) updateRecord(collectionID string, recordID string, payload map[string]interface{}) (map[string]interface{}, error) {
	var updated map[string]interface{}
	err := r.client.requestJSON(
		http.MethodPatch,
		"/api/collections/"+url.PathEscape(collectionID)+"/records/"+url.PathEscape(recordID),
		nil,
		payload,
		false,
		&updated,
	)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *seedRunner) filterPayloadToKnownFields(collectionID string, payload map[string]interface{}) map[string]interface{} {
	fieldMap := r.collectionFieldMap[collectionID]
	if fieldMap == nil {
		return payload
	}

	filtered := map[string]interface{}{}
	for key, value := range payload {
		if fieldMap[key] || fieldMap[strings.ToLower(key)] {
			filtered[key] = value
		}
	}
	return filtered
}

func (r *seedRunner) upsertRecord(
	collectionID string,
	lookupFilter string,
	payload map[string]interface{},
	virtualID string,
) (string, string, error) {
	filteredPayload := r.filterPayloadToKnownFields(collectionID, payload)

	existing, err := r.findFirstRecord(collectionID, lookupFilter, "")
	if err != nil && !errors.Is(err, errNotFound) {
		return "", "", err
	}

	if existing != nil {
		recordID := strings.TrimSpace(toString(existing["id"]))
		if recordID == "" {
			return "", "", fmt.Errorf("existing record missing id in %s", collectionID)
		}

		if r.opts.DryRun {
			r.bumpCollectionStat(collectionID, "updated")
			return recordID, "would-update", nil
		}

		if _, err := r.updateRecord(collectionID, recordID, filteredPayload); err != nil {
			return "", "", err
		}
		r.bumpCollectionStat(collectionID, "updated")
		return recordID, "updated", nil
	}

	if r.opts.DryRun {
		r.bumpCollectionStat(collectionID, "created")
		if virtualID == "" {
			virtualID = "dryrun_" + shortHash(lookupFilter)
		}
		return virtualID, "would-create", nil
	}

	created, err := r.createRecord(collectionID, filteredPayload)
	if err != nil {
		return "", "", err
	}
	recordID := strings.TrimSpace(toString(created["id"]))
	if recordID == "" {
		return "", "", fmt.Errorf("created record missing id in %s", collectionID)
	}
	r.bumpCollectionStat(collectionID, "created")
	return recordID, "created", nil
}

func (r *seedRunner) bumpCollectionStat(collectionID string, action string) {
	entry := r.stats.ByCollection[collectionID]
	if entry == nil {
		entry = &collectionSummary{}
		r.stats.ByCollection[collectionID] = entry
	}
	switch action {
	case "created":
		entry.Created++
	case "updated":
		entry.Updated++
	case "skipped":
		entry.Skipped++
	}
}

func (r *seedRunner) seedLeads(websiteID string) error {
	contacts := buildSeedContacts(websiteID)
	for _, payload := range contacts {
		email := strings.TrimSpace(toString(payload["email"]))
		channel := strings.TrimSpace(toString(payload["channel"]))
		filter := fmt.Sprintf(
			"website=%s && email=%s && channel=%s",
			pbFilterString(websiteID),
			pbFilterString(email),
			pbFilterString(channel),
		)
		virtualID := "dry_contact_" + shortHash(email+"_"+channel)
		if _, _, err := r.upsertRecord(collectionContacts, filter, payload, virtualID); err != nil {
			return fmt.Errorf("contacts upsert failed (%s): %w", email, err)
		}
	}

	whatsapp := buildSeedWhatsappInteractions(websiteID)
	for _, payload := range whatsapp {
		source := strings.TrimSpace(toString(payload["source"]))
		page := strings.TrimSpace(toString(payload["page"]))
		filter := fmt.Sprintf(
			"website=%s && source=%s && page=%s",
			pbFilterString(websiteID),
			pbFilterString(source),
			pbFilterString(page),
		)
		virtualID := "dry_whatsapp_" + shortHash(source+"_"+page)
		if _, _, err := r.upsertRecord(collectionWhatsapp, filter, payload, virtualID); err != nil {
			return fmt.Errorf("whatsapp upsert failed (%s): %w", source, err)
		}
	}

	return nil
}

func (r *seedRunner) seedBooking(websiteID string) error {
	services := buildSeedBookingServices(websiteID)
	serviceIDByKey := map[string]string{}
	serviceByKey := map[string]bookingServiceSeed{}

	for _, service := range services {
		payload := map[string]interface{}{
			"website":         websiteID,
			"name":            service.Name,
			"durationMinutes": service.Duration,
			"active":          service.Active,
			"description":     service.Description,
			"displayOrder":    service.DisplayOrder,
		}

		filter := fmt.Sprintf("website=%s && name=%s", pbFilterString(websiteID), pbFilterString(service.Name))
		recordID, _, err := r.upsertRecord(collectionBookingServices, filter, payload, service.SyntheticID)
		if err != nil {
			return fmt.Errorf("booking service upsert failed (%s): %w", service.Name, err)
		}

		service.PersistedID = recordID
		serviceByKey[service.Key] = service
		serviceIDByKey[service.Key] = recordID
	}

	availability := buildSeedBookingAvailability(websiteID)
	for _, payload := range availability {
		filter := fmt.Sprintf(
			"website=%s && dayOfWeek=%s && startTime=%s && endTime=%s",
			pbFilterString(websiteID),
			pbFilterString(toString(payload["dayOfWeek"])),
			pbFilterString(toString(payload["startTime"])),
			pbFilterString(toString(payload["endTime"])),
		)
		virtualID := "dry_availability_" + shortHash(filter)
		if _, _, err := r.upsertRecord(collectionBookingAvailability, filter, payload, virtualID); err != nil {
			return fmt.Errorf("booking availability upsert failed (%s %s-%s): %w",
				toString(payload["dayOfWeek"]),
				toString(payload["startTime"]),
				toString(payload["endTime"]),
				err,
			)
		}
	}

	exceptions := buildSeedBookingExceptions(websiteID, time.Now().In(time.Local))
	for _, payload := range exceptions {
		filter := fmt.Sprintf(
			"website=%s && date=%s && type=%s && startTime=%s && endTime=%s",
			pbFilterString(websiteID),
			pbFilterString(toString(payload["date"])),
			pbFilterString(toString(payload["type"])),
			pbFilterString(toString(payload["startTime"])),
			pbFilterString(toString(payload["endTime"])),
		)
		virtualID := "dry_exception_" + shortHash(filter)
		if _, _, err := r.upsertRecord(collectionBookingExceptions, filter, payload, virtualID); err != nil {
			return fmt.Errorf("booking exception upsert failed (%s %s): %w", toString(payload["date"]), toString(payload["type"]), err)
		}
	}

	appointments := buildSeedAppointments(websiteID, serviceByKey, serviceIDByKey, time.Now().In(time.Local))
	for _, payload := range appointments {
		email := strings.TrimSpace(toString(payload["email"]))
		date := strings.TrimSpace(toString(payload["date"]))
		timeValue := strings.TrimSpace(toString(payload["time"]))
		filter := fmt.Sprintf(
			"website=%s && email=%s && date=%s && time=%s",
			pbFilterString(websiteID),
			pbFilterString(email),
			pbFilterString(date),
			pbFilterString(timeValue),
		)
		virtualID := "dry_appointment_" + shortHash(email+"_"+date+"_"+timeValue)
		if _, _, err := r.upsertRecord(collectionAppointments, filter, payload, virtualID); err != nil {
			return fmt.Errorf("appointment upsert failed (%s %s %s): %w", email, date, timeValue, err)
		}
	}

	return nil
}

func (r *seedRunner) seedNewsletter(websiteID string) error {
	groups := buildSeedSubscriberGroups(websiteID)
	groupIDs := map[string]string{}

	for _, group := range groups {
		payload := map[string]interface{}{
			"website": websiteID,
			"name":    group.Name,
			"slug":    group.Slug,
		}
		filter := fmt.Sprintf("website=%s && name=%s", pbFilterString(websiteID), pbFilterString(group.Name))
		recordID, _, err := r.upsertRecord(collectionSubscriberGroups, filter, payload, group.SyntheticID)
		if err != nil {
			return fmt.Errorf("subscriber group upsert failed (%s): %w", group.Name, err)
		}
		groupIDs[group.Key] = recordID
	}

	subscribers := buildSeedSubscribers(websiteID, groupIDs, time.Now().In(time.Local))
	activeSubscriberIDs := []string{}
	for _, payload := range subscribers {
		email := strings.TrimSpace(toString(payload["email"]))
		filter := fmt.Sprintf("website=%s && email=%s", pbFilterString(websiteID), pbFilterString(email))
		virtualID := "dry_subscriber_" + shortHash(email)
		recordID, _, err := r.upsertRecord(collectionSubscribers, filter, payload, virtualID)
		if err != nil {
			return fmt.Errorf("subscriber upsert failed (%s): %w", email, err)
		}
		if strings.TrimSpace(toString(payload["status"])) == "active" {
			activeSubscriberIDs = append(activeSubscriberIDs, recordID)
		}
	}

	sort.Strings(activeSubscriberIDs)
	campaigns := buildSeedCampaigns(websiteID, activeSubscriberIDs, time.Now().In(time.Local))
	for _, payload := range campaigns {
		subject := strings.TrimSpace(toString(payload["subject"]))
		filter := fmt.Sprintf("website=%s && subject=%s", pbFilterString(websiteID), pbFilterString(subject))
		virtualID := "dry_campaign_" + shortHash(subject)
		if _, _, err := r.upsertRecord(collectionCampaigns, filter, payload, virtualID); err != nil {
			return fmt.Errorf("campaign upsert failed (%s): %w", subject, err)
		}
	}

	return nil
}

func buildSeedContacts(websiteID string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 30)
	now := time.Now().UTC()

	for i := 1; i <= 30; i++ {
		status := "new"
		if i%3 == 2 {
			status = "read"
		} else if i%3 == 0 {
			status = "archived"
		}

		channel := "form"
		if i%5 == 0 {
			channel = "booking"
		}

		phone := fmt.Sprintf("+351910200%03d", i)
		if i%4 == 0 {
			phone = ""
		}

		notes := ""
		if i%2 == 0 {
			notes = fmt.Sprintf("[QA SEED] Follow-up note for contact %03d", i)
		}

		lastContactedAt := ""
		if status != "new" {
			lastContactedAt = now.AddDate(0, 0, -i).Format(time.RFC3339)
		}

		payload := map[string]interface{}{
			"website":         websiteID,
			"channel":         channel,
			"name":            fmt.Sprintf("[QA SEED] QA Lead Contact %03d", i),
			"email":           fmt.Sprintf("qa-seed-contact-%03d@example.com", i),
			"phone":           phone,
			"subject":         fmt.Sprintf("[QA SEED] Contact Subject %03d", i),
			"message":         fmt.Sprintf("[QA SEED] Contact message body %03d for operational QA dataset.", i),
			"status":          status,
			"notes":           notes,
			"lastContactedAt": lastContactedAt,
		}
		result = append(result, payload)
	}

	return result
}

func buildSeedWhatsappInteractions(websiteID string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 20)
	now := time.Now().UTC()

	sources := []string{"floating_button", "footer_link", "contact_section"}
	for i := 1; i <= 20; i++ {
		status := "new"
		if i%3 == 2 {
			status = "read"
		} else if i%3 == 0 {
			status = "archived"
		}

		notes := ""
		if i%2 == 1 {
			notes = fmt.Sprintf("[QA SEED] WhatsApp note %03d", i)
		}

		lastContactedAt := ""
		if status != "new" {
			lastContactedAt = now.AddDate(0, 0, -(i + 1)).Format(time.RFC3339)
		}

		payload := map[string]interface{}{
			"website":         websiteID,
			"source":          fmt.Sprintf("[QA SEED] %s_%03d", sources[(i-1)%len(sources)], i),
			"page":            fmt.Sprintf("/qa-seed/page/%02d", (i%8)+1),
			"status":          status,
			"notes":           notes,
			"lastContactedAt": lastContactedAt,
		}
		result = append(result, payload)
	}

	return result
}

func buildSeedBookingServices(websiteID string) []bookingServiceSeed {
	return []bookingServiceSeed{
		{
			Key:          "high",
			Name:         "[QA SEED] QA Booking Service High Priority",
			Duration:     30,
			Active:       true,
			Description:  "[QA SEED] High priority short service for operational booking tests.",
			DisplayOrder: 0,
			SyntheticID:  "dry_service_high",
		},
		{
			Key:          "normal",
			Name:         "[QA SEED] QA Booking Service Normal Priority",
			Duration:     45,
			Active:       true,
			Description:  "[QA SEED] Normal priority service used by most seed appointments.",
			DisplayOrder: 50,
			SyntheticID:  "dry_service_normal",
		},
		{
			Key:          "low",
			Name:         "[QA SEED] QA Booking Service Low Priority",
			Duration:     60,
			Active:       true,
			Description:  "[QA SEED] Low priority longer service for slot and conflict coverage.",
			DisplayOrder: 100,
			SyntheticID:  "dry_service_low",
		},
		{
			Key:          "inactive",
			Name:         "[QA SEED] QA Booking Service Inactive",
			Duration:     30,
			Active:       false,
			Description:  "[QA SEED] Inactive service for admin filtering and state checks.",
			DisplayOrder: 80,
			SyntheticID:  "dry_service_inactive",
		},
		{
			Key:          "extended",
			Name:         "[QA SEED] QA Booking Service Extended Session",
			Duration:     90,
			Active:       true,
			Description:  "[QA SEED] Extended session to test duration-sensitive slot and calendar displays.",
			DisplayOrder: 25,
			SyntheticID:  "dry_service_extended",
		},
	}
}

func buildSeedBookingAvailability(websiteID string) []map[string]interface{} {
	type window struct {
		day    string
		start  string
		end    string
		active bool
	}

	windows := []window{
		{"mon", "09:00", "12:00", true},
		{"mon", "13:00", "18:00", true},
		{"tue", "09:00", "12:00", true},
		{"tue", "13:00", "18:00", true},
		{"wed", "09:00", "12:00", true},
		{"thu", "09:00", "12:00", true},
		{"thu", "13:00", "18:00", true},
		{"fri", "09:00", "12:00", true},
		{"fri", "13:00", "17:00", true},
		{"sat", "09:00", "12:00", false},
		{"sun", "09:00", "12:00", false},
	}

	result := make([]map[string]interface{}, 0, len(windows))
	for _, w := range windows {
		result = append(result, map[string]interface{}{
			"website":   websiteID,
			"dayOfWeek": w.day,
			"startTime": w.start,
			"endTime":   w.end,
			"active":    w.active,
		})
	}
	return result
}

func buildSeedBookingExceptions(websiteID string, now time.Time) []map[string]interface{} {
	closedDate := now.AddDate(0, 0, 14).Format("2006-01-02")
	customDate := now.AddDate(0, 0, 15).Format("2006-01-02")
	inactiveDate := now.AddDate(0, 0, 16).Format("2006-01-02")
	secondaryCustomDate := now.AddDate(0, 0, 21).Format("2006-01-02")

	return []map[string]interface{}{
		{
			"website":   websiteID,
			"date":      closedDate,
			"type":      "closed",
			"startTime": "",
			"endTime":   "",
			"note":      "[QA SEED] Closed holiday coverage",
			"active":    true,
		},
		{
			"website":   websiteID,
			"date":      customDate,
			"type":      "customHours",
			"startTime": "14:00",
			"endTime":   "16:00",
			"note":      "[QA SEED] Custom hours afternoon coverage",
			"active":    true,
		},
		{
			"website":   websiteID,
			"date":      inactiveDate,
			"type":      "closed",
			"startTime": "",
			"endTime":   "",
			"note":      "[QA SEED] Inactive closed exception",
			"active":    false,
		},
		{
			"website":   websiteID,
			"date":      secondaryCustomDate,
			"type":      "customHours",
			"startTime": "09:30",
			"endTime":   "11:30",
			"note":      "[QA SEED] Secondary custom hours coverage",
			"active":    true,
		},
	}
}

func buildSeedAppointments(
	websiteID string,
	serviceByKey map[string]bookingServiceSeed,
	serviceIDByKey map[string]string,
	now time.Time,
) []map[string]interface{} {
	type appointmentSpec struct {
		Index         int
		ServiceKey    string
		Status        string
		DateOffset    int
		Time          string
		Notes         string
		InternalNotes string
		Archived      bool
	}

	specs := []appointmentSpec{
		{1, "normal", "pending", 10, "10:00", "[QA SEED] Pending future appointment 001", "", false},
		{2, "high", "pending", 11, "11:00", "[QA SEED] Pending future appointment 002", "[QA SEED] Internal pending note", false},
		{3, "high", "confirmed", 12, "09:00", "[QA SEED] Confirmed future appointment 001", "", false},
		{4, "normal", "confirmed", 13, "14:00", "[QA SEED] Confirmed future appointment 002", "[QA SEED] Internal confirmed note", false},
		{5, "low", "cancelled", 14, "15:00", "[QA SEED] Cancelled future appointment 001", "", false},
		{6, "extended", "cancelled", 15, "16:00", "[QA SEED] Cancelled future appointment 002", "[QA SEED] Internal cancelled note", false},
		{7, "high", "confirmed", -5, "10:00", "[QA SEED] Confirmed past appointment 001", "", false},
		{8, "normal", "confirmed", -10, "11:30", "[QA SEED] Confirmed past appointment 002", "", false},
		{9, "low", "cancelled", -7, "12:00", "[QA SEED] Cancelled past appointment 001", "", false},
		{10, "normal", "pending", 9, "13:00", "[QA SEED] Archived pending appointment", "", true},
		{11, "extended", "confirmed", 18, "09:30", "[QA SEED] Confirmed future appointment 003", "[QA SEED] Internal note for calendar tests", false},
		{12, "high", "cancelled", -2, "17:00", "[QA SEED] Archived cancelled appointment", "", true},
	}

	result := make([]map[string]interface{}, 0, len(specs))
	for _, spec := range specs {
		service := serviceByKey[spec.ServiceKey]
		serviceID := strings.TrimSpace(serviceIDByKey[spec.ServiceKey])
		dateValue := now.AddDate(0, 0, spec.DateOffset).Format("2006-01-02")
		confirmedAt := ""
		cancelledAt := ""
		if spec.Status == "confirmed" {
			confirmedAt = now.AddDate(0, 0, spec.DateOffset-1).Format(time.RFC3339)
		}
		if spec.Status == "cancelled" {
			cancelledAt = now.AddDate(0, 0, spec.DateOffset-1).Format(time.RFC3339)
		}
		rescheduledAt := ""
		if spec.Index%4 == 0 {
			rescheduledAt = now.AddDate(0, 0, spec.DateOffset-2).Format(time.RFC3339)
		}
		archivedAt := ""
		if spec.Archived {
			archivedAt = now.AddDate(0, 0, -1).Format(time.RFC3339)
		}

		phone := fmt.Sprintf("+351930500%03d", spec.Index)
		if spec.Index%3 == 0 {
			phone = ""
		}

		payload := map[string]interface{}{
			"website":                        websiteID,
			"service":                        serviceID,
			"name":                           fmt.Sprintf("[QA SEED] QA Appointment %03d", spec.Index),
			"email":                          fmt.Sprintf("qa-seed-appointment-%03d@example.com", spec.Index),
			"phone":                          phone,
			"date":                           dateValue,
			"time":                           spec.Time,
			"status":                         spec.Status,
			"notes":                          spec.Notes,
			"internalNotes":                  spec.InternalNotes,
			"confirmedAt":                    confirmedAt,
			"cancelledAt":                    cancelledAt,
			"rescheduledAt":                  rescheduledAt,
			"archivedAt":                     archivedAt,
			"serviceNameSnapshot":            service.Name,
			"serviceDurationMinutesSnapshot": service.Duration,
			"serviceDescriptionSnapshot":     service.Description,
		}
		result = append(result, payload)
	}
	return result
}

func buildSeedSubscriberGroups(websiteID string) []subscriberGroupSeed {
	return []subscriberGroupSeed{
		{
			Key:         "customers",
			Name:        "[QA SEED] QA Customers",
			Slug:        "qa-customers",
			SyntheticID: "dry_group_customers",
		},
		{
			Key:         "prospects",
			Name:        "[QA SEED] QA Prospects",
			Slug:        "qa-prospects",
			SyntheticID: "dry_group_prospects",
		},
		{
			Key:         "vip",
			Name:        "[QA SEED] QA VIP",
			Slug:        "qa-vip",
			SyntheticID: "dry_group_vip",
		},
	}
}

func buildSeedSubscribers(websiteID string, groupIDs map[string]string, now time.Time) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 30)
	groupKeys := [][]string{
		{"customers"},
		{"prospects"},
		{"vip"},
		{"customers", "vip"},
		{"prospects", "vip"},
	}

	for i := 1; i <= 30; i++ {
		status := "active"
		switch {
		case i <= 12:
			status = "active"
		case i <= 22:
			status = "pending"
		default:
			status = "unsubscribed"
		}

		confirmedAt := ""
		confirmationTokenHash := ""
		confirmationTokenExpiresAt := ""
		unsubscribeTokenHash := ""
		unsubscribedAt := ""

		email := fmt.Sprintf("qa-seed-subscriber-%03d@example.com", i)
		name := fmt.Sprintf("[QA SEED] QA Subscriber %03d", i)

		if status == "active" {
			confirmedAt = now.AddDate(0, 0, -(i + 2)).Format(time.RFC3339)
			unsubscribeTokenHash = deterministicHash("qa-seed-unsubscribe-active-" + strconv.Itoa(i))
		}
		if status == "pending" {
			confirmationTokenHash = deterministicHash("qa-seed-confirmation-pending-" + strconv.Itoa(i))
			confirmationTokenExpiresAt = now.Add(48 * time.Hour).Format(time.RFC3339)
		}
		if status == "unsubscribed" {
			confirmedAt = now.AddDate(0, 0, -(i + 5)).Format(time.RFC3339)
			unsubscribeTokenHash = deterministicHash("qa-seed-unsubscribe-final-" + strconv.Itoa(i))
			unsubscribedAt = now.AddDate(0, 0, -(i%7 + 1)).Format(time.RFC3339)
		}

		keys := groupKeys[(i-1)%len(groupKeys)]
		groups := make([]string, 0, len(keys))
		for _, key := range keys {
			id := strings.TrimSpace(groupIDs[key])
			if id != "" {
				groups = append(groups, id)
			}
		}

		payload := map[string]interface{}{
			"website":                    websiteID,
			"email":                      email,
			"name":                       name,
			"status":                     status,
			"confirmedAt":                confirmedAt,
			"groups":                     groups,
			"confirmationTokenHash":      confirmationTokenHash,
			"confirmationTokenExpiresAt": confirmationTokenExpiresAt,
			"unsubscribeTokenHash":       unsubscribeTokenHash,
			"unsubscribedAt":             unsubscribedAt,
		}
		result = append(result, payload)
	}

	return result
}

func buildSeedCampaigns(websiteID string, activeSubscriberIDs []string, now time.Time) []map[string]interface{} {
	firstManual := pickIDs(activeSubscriberIDs, 0, 6)
	secondManual := pickIDs(activeSubscriberIDs, 2, 5)

	return []map[string]interface{}{
		{
			"website":         websiteID,
			"subject":         "[QA SEED] QA Campaign Draft 001",
			"body":            "<p>[QA SEED] Draft campaign body for lifecycle verification.</p>",
			"status":          "draft",
			"recipientsType":  "all",
			"recipientsIds":   []string{},
			"recipientsCount": 0,
			"sentAt":          "",
		},
		{
			"website":         websiteID,
			"subject":         "[QA SEED] QA Campaign Draft Manual 001",
			"body":            "<p>[QA SEED] Draft manual-recipient campaign body.</p>",
			"status":          "draft",
			"recipientsType":  "manual",
			"recipientsIds":   firstManual,
			"recipientsCount": len(firstManual),
			"sentAt":          "",
		},
		{
			"website":         websiteID,
			"subject":         "[QA SEED] QA Campaign Sent 001",
			"body":            "<p>[QA SEED] Sent campaign body for reports and counters.</p>",
			"status":          "sent",
			"recipientsType":  "all",
			"recipientsIds":   []string{},
			"recipientsCount": len(activeSubscriberIDs),
			"sentAt":          now.AddDate(0, 0, -2).Format(time.RFC3339),
		},
		{
			"website":         websiteID,
			"subject":         "[QA SEED] QA Campaign Sent Manual 001",
			"body":            "<p>[QA SEED] Sent manual-recipient campaign body.</p>",
			"status":          "sent",
			"recipientsType":  "manual",
			"recipientsIds":   secondManual,
			"recipientsCount": len(secondManual),
			"sentAt":          now.AddDate(0, 0, -5).Format(time.RFC3339),
		},
	}
}

func pickIDs(ids []string, start int, max int) []string {
	if len(ids) == 0 || max <= 0 {
		return []string{}
	}
	if start < 0 {
		start = 0
	}
	if start >= len(ids) {
		start = 0
	}
	end := start + max
	if end > len(ids) {
		end = len(ids)
	}
	result := append([]string{}, ids[start:end]...)
	return result
}

func (r *seedRunner) printSummary(websiteSlug string, websiteID string) {
	fmt.Println("")
	fmt.Println("=== NUVIO QA Operational Seed Summary ===")
	fmt.Printf("Website slug: %s\n", websiteSlug)
	fmt.Printf("Website id:   %s\n", websiteID)
	fmt.Printf("Dry-run:      %t\n", r.opts.DryRun)
	fmt.Println("")

	ordered := []struct {
		ID    string
		Label string
	}{
		{collectionContacts, "Contacts"},
		{collectionWhatsapp, "Whatsapp"},
		{collectionBookingServices, "BookingServices"},
		{collectionBookingAvailability, "BookingAvailability"},
		{collectionBookingExceptions, "BookingExceptions"},
		{collectionAppointments, "Appointments"},
		{collectionSubscriberGroups, "SubscriberGroups"},
		{collectionSubscribers, "Subscribers"},
		{collectionCampaigns, "Campaigns"},
	}

	for _, item := range ordered {
		stats := r.stats.ByCollection[item.ID]
		if stats == nil {
			stats = &collectionSummary{}
		}
		fmt.Printf(
			"%s: created=%d updated=%d skipped=%d\n",
			item.Label,
			stats.Created,
			stats.Updated,
			stats.Skipped,
		)
	}

	if len(r.stats.SkippedNotes) > 0 {
		fmt.Println("")
		fmt.Println("Skipped notes:")
		for _, note := range r.stats.SkippedNotes {
			fmt.Printf("- %s\n", note)
		}
	}

	fmt.Println("")
	if r.opts.DryRun {
		fmt.Println("DRY-RUN completed: no writes were performed.")
	} else {
		fmt.Println("WRITE mode completed successfully.")
	}
}

func pbFilterString(raw string) string {
	return "'" + strings.ReplaceAll(strings.TrimSpace(raw), "'", "\\'") + "'"
}

func deterministicHash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func shortHash(raw string) string {
	full := deterministicHash(raw)
	if len(full) < 12 {
		return full
	}
	return full[:12]
}

func toString(v interface{}) string {
	switch typed := v.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}
