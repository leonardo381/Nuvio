package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioWebsitesCollectionID         = "pbc_2619338178"
	nuvioReviewsCollectionID          = "pbc_1661203300"
	nuvioReviewsSourceGoogle          = "google"
	nuvioReviewsDefaultDashboardLimit = 20
	nuvioReviewsMaxDashboardScan      = 5000
)

var nuvioReviewsHTTPClient = &http.Client{
	Timeout: 12 * time.Second,
}

type nuvioGooglePlaceReview struct {
	ExternalID       string
	AuthorName       string
	Rating           float64
	Text             string
	RelativeTime     string
	PublishedAt      string
	ReviewerPhotoURL string
	ReviewURL        string
}

type nuvioGooglePlaceReviewsSnapshot struct {
	AverageRating float64
	TotalReviews  int
	OpenOnGoogle  string
	RecentReviews []nuvioGooglePlaceReview
}

type nuvioGooglePlacesDetailsResponse struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message"`
	Result       struct {
		Rating           float64 `json:"rating"`
		UserRatingsTotal int     `json:"user_ratings_total"`
		URL              string  `json:"url"`
		Reviews          []struct {
			AuthorName              string  `json:"author_name"`
			AuthorURL               string  `json:"author_url"`
			ProfilePhotoURL         string  `json:"profile_photo_url"`
			Rating                  float64 `json:"rating"`
			Text                    string  `json:"text"`
			RelativeTimeDescription string  `json:"relative_time_description"`
			Time                    int64   `json:"time"`
		} `json:"reviews"`
	} `json:"result"`
}

type nuvioWebsiteReviewsConfig struct {
	FeatureAvailable bool
	Enabled          bool
	GooglePlaceID    string
	ReviewLink       string
}

type nuvioReviewsDashboardSummary struct {
	AverageRating *float64 `json:"averageRating"`
	TotalReviews  int      `json:"totalReviews"`
	LastSyncAt    string   `json:"lastSyncAt"`
}

type nuvioReviewsDashboardReview struct {
	ID              string  `json:"id"`
	ExternalID      string  `json:"externalId"`
	AuthorName      string  `json:"authorName"`
	Rating          float64 `json:"rating"`
	Text            string  `json:"text"`
	PublishedAt     string  `json:"publishedAt"`
	RelativeTime    string  `json:"relativeTime"`
	ReviewerPhoto   string  `json:"reviewerPhotoUrl"`
	ReviewURL       string  `json:"reviewUrl"`
	SyncedAt        string  `json:"syncedAt"`
	Source          string  `json:"source"`
	WebsiteID       string  `json:"website"`
	WebsiteNameHint string  `json:"websiteNameHint"`
}

type nuvioReviewsDashboardPayload struct {
	WebsiteID         string                        `json:"websiteId"`
	WebsiteName       string                        `json:"websiteName"`
	State             string                        `json:"state"`
	FeatureAvailable  bool                          `json:"featureAvailable"`
	Enabled           bool                          `json:"enabled"`
	GooglePlaceID     string                        `json:"googlePlaceId"`
	OpenOnGoogle      string                        `json:"openOnGoogle"`
	ReviewRequestLink string                        `json:"reviewRequestLink"`
	Summary           nuvioReviewsDashboardSummary  `json:"summary"`
	Reviews           []nuvioReviewsDashboardReview `json:"reviews"`
	SyncedCount       int                           `json:"syncedCount,omitempty"`
}

type nuvioReviewsSyncResult struct {
	SyncedCount  int
	OpenOnGoogle string
}

// NUVIO CUSTOM START: Reusable Google Places reviews fetch utilities for collection sync.
func fetchGooglePlaceReviewsSnapshot(
	ctx context.Context,
	apiKey string,
	googlePlaceID string,
) (*nuvioGooglePlaceReviewsSnapshot, error) {
	trimmedKey := strings.TrimSpace(apiKey)
	trimmedPlaceID := strings.TrimSpace(googlePlaceID)

	if trimmedKey == "" {
		return nil, fmt.Errorf("missing Google Places API key")
	}

	if trimmedPlaceID == "" {
		return nil, fmt.Errorf("missing Google Place ID")
	}

	endpoint := "https://maps.googleapis.com/maps/api/place/details/json"
	query := url.Values{}
	query.Set("place_id", trimmedPlaceID)
	query.Set("fields", "rating,user_ratings_total,reviews,url")
	query.Set("key", trimmedKey)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Places request: %w", err)
	}

	response, err := nuvioReviewsHTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Google Places: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read Google Places response: %w", err)
	}

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("google Places request failed with status %d", response.StatusCode)
	}

	parsed := &nuvioGooglePlacesDetailsResponse{}
	if err := json.Unmarshal(body, parsed); err != nil {
		return nil, fmt.Errorf("failed to parse Google Places response: %w", err)
	}

	if strings.ToUpper(parsed.Status) != "OK" {
		if parsed.ErrorMessage != "" {
			return nil, fmt.Errorf("google Places error (%s): %s", parsed.Status, parsed.ErrorMessage)
		}

		return nil, fmt.Errorf("google Places error status: %s", parsed.Status)
	}

	snapshot := &nuvioGooglePlaceReviewsSnapshot{
		AverageRating: clampFloat(parsed.Result.Rating, 0, 5),
		TotalReviews:  maxInt(parsed.Result.UserRatingsTotal, 0),
		OpenOnGoogle:  strings.TrimSpace(parsed.Result.URL),
		RecentReviews: make([]nuvioGooglePlaceReview, 0, len(parsed.Result.Reviews)),
	}

	for _, review := range parsed.Result.Reviews {
		publishedAt := ""
		if review.Time > 0 {
			publishedAt = time.Unix(review.Time, 0).UTC().Format(time.RFC3339)
		}

		authorName := strings.TrimSpace(review.AuthorName)
		if authorName == "" {
			authorName = "Google user"
		}

		snapshot.RecentReviews = append(snapshot.RecentReviews, nuvioGooglePlaceReview{
			ExternalID:       deriveGoogleReviewExternalID(trimmedPlaceID, review.Time, authorName, review.Text),
			AuthorName:       authorName,
			Rating:           clampFloat(review.Rating, 0, 5),
			Text:             strings.TrimSpace(review.Text),
			RelativeTime:     strings.TrimSpace(review.RelativeTimeDescription),
			PublishedAt:      publishedAt,
			ReviewerPhotoURL: strings.TrimSpace(review.ProfilePhotoURL),
			ReviewURL:        strings.TrimSpace(review.AuthorURL),
		})
	}

	return snapshot, nil
}

func deriveGoogleReviewRequestLink(googlePlaceID string) string {
	trimmedPlaceID := strings.TrimSpace(googlePlaceID)
	if trimmedPlaceID == "" {
		return ""
	}

	return "https://search.google.com/local/writereview?placeid=" + url.QueryEscape(trimmedPlaceID)
}

func deriveGooglePlaceOpenLink(googlePlaceID string) string {
	trimmedPlaceID := strings.TrimSpace(googlePlaceID)
	if trimmedPlaceID == "" {
		return ""
	}

	return "https://www.google.com/maps/place/?q=place_id:" + url.QueryEscape(trimmedPlaceID)
}

func deriveGoogleReviewExternalID(placeID string, unixTime int64, authorName string, text string) string {
	source := strings.Join([]string{
		strings.TrimSpace(placeID),
		strconv.FormatInt(unixTime, 10),
		strings.TrimSpace(authorName),
		strings.TrimSpace(text),
	}, "|")

	hash := sha1.Sum([]byte(source))
	return "google_" + hex.EncodeToString(hash[:8])
}

func clampFloat(value float64, min float64, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// NUVIO CUSTOM END: Reusable Google Places reviews fetch utilities for collection sync.

// NUVIO CUSTOM START: Collection-backed dashboard Reviews module routes + sync.
func registerNuvioReviewsRoutes(e *core.ServeEvent) {
	reviewsGroup := e.Router.Group("/api/nuvio/reviews").Bind(apis.RequireSuperuserAuth())

	reviewsGroup.GET("/dashboard", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		payload, err := buildNuvioReviewsDashboardPayload(e.App, websiteID, nuvioReviewsDefaultDashboardLimit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}

			return e.BadRequestError("Failed to load Reviews dashboard data.", err)
		}

		return e.JSON(http.StatusOK, payload)
	})

	reviewsGroup.POST("/sync", func(e *core.RequestEvent) error {
		body := struct {
			WebsiteID string `json:"websiteId"`
		}{}

		if err := e.BindBody(&body); err != nil {
			return e.BadRequestError("Invalid sync payload.", err)
		}

		websiteID := strings.TrimSpace(body.WebsiteID)
		if websiteID == "" {
			websiteID = strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		}
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		syncResult, err := syncNuvioWebsiteReviews(e.App, e.Request.Context(), websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}

			return e.BadRequestError(err.Error(), nil)
		}

		payload, err := buildNuvioReviewsDashboardPayload(e.App, websiteID, nuvioReviewsDefaultDashboardLimit)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}

			return e.BadRequestError("Reviews synced, but failed to load dashboard data.", err)
		}

		if strings.TrimSpace(syncResult.OpenOnGoogle) != "" {
			payload.OpenOnGoogle = strings.TrimSpace(syncResult.OpenOnGoogle)
		}
		payload.SyncedCount = syncResult.SyncedCount

		return e.JSON(http.StatusOK, payload)
	})
}

func syncNuvioWebsiteReviews(app core.App, ctx context.Context, websiteID string) (*nuvioReviewsSyncResult, error) {
	website, config, err := loadNuvioWebsiteReviewsConfig(app, websiteID)
	if err != nil {
		return nil, err
	}

	if !config.FeatureAvailable {
		return nil, fmt.Errorf("Reviews feature is unavailable for this website")
	}

	if !config.Enabled {
		return nil, fmt.Errorf("Reviews are disabled for this website")
	}

	if strings.TrimSpace(config.GooglePlaceID) == "" {
		return nil, fmt.Errorf("Google Place ID is missing in website settings")
	}

	googleAPIKey := strings.TrimSpace(os.Getenv("NUVIO_GOOGLE_PLACES_API_KEY"))
	if googleAPIKey == "" {
		googleAPIKey = strings.TrimSpace(os.Getenv("GOOGLE_PLACES_API_KEY"))
	}
	if googleAPIKey == "" {
		return nil, fmt.Errorf("Google Places API key is missing (set NUVIO_GOOGLE_PLACES_API_KEY)")
	}

	snapshot, err := fetchGooglePlaceReviewsSnapshot(ctx, googleAPIKey, config.GooglePlaceID)
	if err != nil {
		return nil, err
	}

	reviewsCollection, err := app.FindCachedCollectionByNameOrId(nuvioReviewsCollectionID)
	if err != nil {
		return nil, err
	}

	syncedCount, err := upsertNuvioReviewsFromSnapshot(
		app,
		reviewsCollection,
		website.Id,
		config.GooglePlaceID,
		snapshot,
	)
	if err != nil {
		return nil, err
	}

	return &nuvioReviewsSyncResult{
		SyncedCount:  syncedCount,
		OpenOnGoogle: strings.TrimSpace(snapshot.OpenOnGoogle),
	}, nil
}

func upsertNuvioReviewsFromSnapshot(
	app core.App,
	reviewsCollection *core.Collection,
	websiteID string,
	googlePlaceID string,
	snapshot *nuvioGooglePlaceReviewsSnapshot,
) (int, error) {
	if snapshot == nil || reviewsCollection == nil {
		return 0, nil
	}

	syncedCount := 0
	syncedAt := time.Now().UTC().Format(time.RFC3339)

	for _, review := range snapshot.RecentReviews {
		externalID := strings.TrimSpace(review.ExternalID)
		if externalID == "" {
			externalID = deriveGoogleReviewExternalID(googlePlaceID, 0, review.AuthorName, review.Text)
		}

		existing, err := app.FindFirstRecordByFilter(
			reviewsCollection,
			"website={:website} && source={:source} && externalId={:externalId}",
			dbx.Params{
				"website":    websiteID,
				"source":     nuvioReviewsSourceGoogle,
				"externalId": externalID,
			},
		)

		record := existing
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return syncedCount, err
			}

			record = core.NewRecord(reviewsCollection)
		}

		authorName := strings.TrimSpace(review.AuthorName)
		if authorName == "" {
			authorName = "Google user"
		}

		reviewURL := strings.TrimSpace(review.ReviewURL)
		if reviewURL == "" {
			reviewURL = strings.TrimSpace(snapshot.OpenOnGoogle)
		}

		record.Set("website", websiteID)
		record.Set("source", nuvioReviewsSourceGoogle)
		record.Set("externalId", externalID)
		record.Set("authorName", authorName)
		record.Set("rating", clampFloat(review.Rating, 0, 5))
		record.Set("text", strings.TrimSpace(review.Text))
		record.Set("publishedAt", strings.TrimSpace(review.PublishedAt))
		record.Set("relativeTimeText", strings.TrimSpace(review.RelativeTime))
		record.Set("reviewerPhotoUrl", strings.TrimSpace(review.ReviewerPhotoURL))
		record.Set("reviewUrl", reviewURL)
		record.Set("syncedAt", syncedAt)

		if err := app.Save(record); err != nil {
			return syncedCount, err
		}

		syncedCount++
	}

	return syncedCount, nil
}

func buildNuvioReviewsDashboardPayload(
	app core.App,
	websiteID string,
	limit int,
) (*nuvioReviewsDashboardPayload, error) {
	website, config, err := loadNuvioWebsiteReviewsConfig(app, websiteID)
	if err != nil {
		return nil, err
	}

	reviewsCollection, err := app.FindCachedCollectionByNameOrId(nuvioReviewsCollectionID)
	if err != nil {
		return nil, err
	}

	allRecords, err := app.FindRecordsByFilter(
		reviewsCollection,
		"website={:website}",
		"-publishedAt,-syncedAt,-created",
		nuvioReviewsMaxDashboardScan,
		0,
		dbx.Params{"website": websiteID},
	)
	if err != nil {
		return nil, err
	}

	totalReviews := len(allRecords)
	averageRating := (*float64)(nil)
	if totalReviews > 0 {
		sum := 0.0
		for _, record := range allRecords {
			sum += clampFloat(record.GetFloat("rating"), 0, 5)
		}

		avg := math.Round((sum/float64(totalReviews))*10) / 10
		averageRating = &avg
	}

	lastSyncAt := resolveLatestSyncedAt(allRecords)

	if limit <= 0 {
		limit = nuvioReviewsDefaultDashboardLimit
	}

	visibleRecords := allRecords
	if len(visibleRecords) > limit {
		visibleRecords = visibleRecords[:limit]
	}

	reviews := make([]nuvioReviewsDashboardReview, 0, len(visibleRecords))
	websiteNameHint := resolveWebsiteDisplayName(website)
	for _, record := range visibleRecords {
		reviews = append(reviews, nuvioReviewsDashboardReview{
			ID:              record.Id,
			ExternalID:      strings.TrimSpace(record.GetString("externalId")),
			AuthorName:      strings.TrimSpace(record.GetString("authorName")),
			Rating:          clampFloat(record.GetFloat("rating"), 0, 5),
			Text:            strings.TrimSpace(record.GetString("text")),
			PublishedAt:     strings.TrimSpace(record.GetString("publishedAt")),
			RelativeTime:    strings.TrimSpace(record.GetString("relativeTimeText")),
			ReviewerPhoto:   strings.TrimSpace(record.GetString("reviewerPhotoUrl")),
			ReviewURL:       strings.TrimSpace(record.GetString("reviewUrl")),
			SyncedAt:        strings.TrimSpace(record.GetString("syncedAt")),
			Source:          strings.TrimSpace(record.GetString("source")),
			WebsiteID:       websiteID,
			WebsiteNameHint: websiteNameHint,
		})
	}

	reviewRequestLink := strings.TrimSpace(config.ReviewLink)
	if reviewRequestLink == "" {
		reviewRequestLink = deriveGoogleReviewRequestLink(config.GooglePlaceID)
	}

	payload := &nuvioReviewsDashboardPayload{
		WebsiteID:         websiteID,
		WebsiteName:       websiteNameHint,
		State:             resolveNuvioReviewsState(config, totalReviews),
		FeatureAvailable:  config.FeatureAvailable,
		Enabled:           config.Enabled,
		GooglePlaceID:     strings.TrimSpace(config.GooglePlaceID),
		OpenOnGoogle:      deriveGooglePlaceOpenLink(config.GooglePlaceID),
		ReviewRequestLink: reviewRequestLink,
		Summary: nuvioReviewsDashboardSummary{
			AverageRating: averageRating,
			TotalReviews:  totalReviews,
			LastSyncAt:    lastSyncAt,
		},
		Reviews: reviews,
	}

	return payload, nil
}

func loadNuvioWebsiteReviewsConfig(
	app core.App,
	websiteID string,
) (*core.Record, nuvioWebsiteReviewsConfig, error) {
	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		return nil, nuvioWebsiteReviewsConfig{}, err
	}

	settings := parseNuvioSettingsObject(website.Get("settings"))
	config := nuvioWebsiteReviewsConfig{
		FeatureAvailable: true,
		Enabled:          false,
		GooglePlaceID:    "",
		ReviewLink:       "",
	}

	if featureFlags, ok := toStringAnyMap(settings["featureFlags"]); ok {
		if value, ok := parseBoolValue(featureFlags["reviews"]); ok {
			config.FeatureAvailable = value
		}
	}

	if reviewsSettings, ok := toStringAnyMap(settings["reviews"]); ok {
		if value, ok := parseBoolValue(reviewsSettings["enabled"]); ok {
			config.Enabled = value
		}

		config.GooglePlaceID = strings.TrimSpace(parseStringValue(reviewsSettings["googlePlaceId"]))
		config.ReviewLink = strings.TrimSpace(parseStringValue(reviewsSettings["reviewLink"]))
	}

	return website, config, nil
}

func resolveNuvioReviewsState(config nuvioWebsiteReviewsConfig, totalReviews int) string {
	if !config.FeatureAvailable {
		return "feature_unavailable"
	}

	if !config.Enabled {
		return "disabled"
	}

	if strings.TrimSpace(config.GooglePlaceID) == "" {
		return "not_configured"
	}

	if totalReviews <= 0 {
		return "never_synced"
	}

	return "ready"
}

func resolveWebsiteDisplayName(website *core.Record) string {
	if website == nil {
		return ""
	}

	for _, field := range []string{"title", "name", "slug"} {
		value := strings.TrimSpace(website.GetString(field))
		if value != "" {
			return value
		}
	}

	return strings.TrimSpace(website.Id)
}

func resolveLatestSyncedAt(records []*core.Record) string {
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.000Z"}
	latestTime := time.Time{}
	latestRaw := ""

	for _, record := range records {
		raw := strings.TrimSpace(record.GetString("syncedAt"))
		if raw == "" {
			continue
		}

		parsed := time.Time{}
		parsedOK := false
		for _, layout := range layouts {
			if candidate, err := time.Parse(layout, raw); err == nil {
				parsed = candidate.UTC()
				parsedOK = true
				break
			}
		}

		if parsedOK {
			if latestTime.IsZero() || parsed.After(latestTime) {
				latestTime = parsed
				latestRaw = parsed.Format(time.RFC3339)
			}
			continue
		}

		if raw > latestRaw {
			latestRaw = raw
		}
	}

	return latestRaw
}

func parseNuvioSettingsObject(rawSettings any) map[string]any {
	if rawSettings == nil {
		return map[string]any{}
	}

	if direct, ok := toStringAnyMap(rawSettings); ok {
		return direct
	}

	if rawString, ok := rawSettings.(string); ok {
		normalized := strings.TrimSpace(rawString)
		if normalized == "" {
			return map[string]any{}
		}

		parsed := map[string]any{}
		if err := json.Unmarshal([]byte(normalized), &parsed); err == nil {
			return parsed
		}

		return map[string]any{}
	}

	encoded, err := json.Marshal(rawSettings)
	if err != nil {
		return map[string]any{}
	}

	parsed := map[string]any{}
	if err := json.Unmarshal(encoded, &parsed); err != nil {
		return map[string]any{}
	}

	return parsed
}

func toStringAnyMap(value any) (map[string]any, bool) {
	typed, ok := value.(map[string]any)
	if ok {
		return typed, true
	}

	typedInterface, ok := value.(map[string]interface{})
	if ok {
		return typedInterface, true
	}

	return nil, false
}

func parseBoolValue(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		normalized := strings.TrimSpace(strings.ToLower(typed))
		if normalized == "true" {
			return true, true
		}
		if normalized == "false" {
			return false, true
		}
	}

	return false, false
}

func parseStringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int32:
		return strconv.FormatInt(int64(typed), 10)
	case int16:
		return strconv.FormatInt(int64(typed), 10)
	case int8:
		return strconv.FormatInt(int64(typed), 10)
	case uint:
		return strconv.FormatUint(uint64(typed), 10)
	case uint64:
		return strconv.FormatUint(typed, 10)
	case uint32:
		return strconv.FormatUint(uint64(typed), 10)
	case uint16:
		return strconv.FormatUint(uint64(typed), 10)
	case uint8:
		return strconv.FormatUint(uint64(typed), 10)
	default:
		if value == nil {
			return ""
		}

		return fmt.Sprintf("%v", value)
	}
}

// NUVIO CUSTOM END: Collection-backed dashboard Reviews module routes + sync.
