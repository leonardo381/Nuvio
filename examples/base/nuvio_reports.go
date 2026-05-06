package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioReportsTrafficPeriodThisMonth = "thisMonth"
	nuvioReportsTrafficPeriodLastMonth = "lastMonth"
	nuvioReportsTrafficPeriodLast30d   = "last30Days"
	nuvioReportsTrafficPeriodAllTime   = "allTime"

	nuvioPlausibleDefaultBaseURL = "https://plausible.io"
)

var nuvioReportsHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

type nuvioWebsiteReportsAnalyticsConfig struct {
	FeatureAvailable bool
	Provider         string
	Enabled          bool
	SiteID           string
	ScriptEnabled    bool
}

type nuvioPlausibleConfig struct {
	APIKey  string
	BaseURL string
}

type nuvioReportsTrafficPeriod struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

type nuvioReportsTrafficSummary struct {
	Visitors             int      `json:"visitors"`
	Pageviews            int      `json:"pageviews"`
	BounceRate           *float64 `json:"bounceRate"`
	VisitDurationSeconds *int     `json:"visitDurationSeconds"`
}

type nuvioReportsTrafficTopPage struct {
	Page      string `json:"page"`
	Visitors  int    `json:"visitors"`
	Pageviews int    `json:"pageviews"`
}

type nuvioReportsTrafficSource struct {
	Source   string `json:"source"`
	Visitors int    `json:"visitors"`
}

type nuvioReportsTrafficDevice struct {
	Device   string `json:"device"`
	Visitors int    `json:"visitors"`
}

type nuvioReportsTrafficResponse struct {
	State     string                       `json:"state"`
	Message   string                       `json:"message,omitempty"`
	Provider  string                       `json:"provider,omitempty"`
	SiteID    string                       `json:"siteId,omitempty"`
	Period    nuvioReportsTrafficPeriod    `json:"period"`
	Summary   *nuvioReportsTrafficSummary  `json:"summary,omitempty"`
	TopPages  []nuvioReportsTrafficTopPage `json:"topPages,omitempty"`
	Sources   []nuvioReportsTrafficSource  `json:"sources,omitempty"`
	Devices   []nuvioReportsTrafficDevice  `json:"devices,omitempty"`
	FetchedAt string                       `json:"fetchedAt,omitempty"`
}

type nuvioReportsTrafficPeriodQuery struct {
	Period           nuvioReportsTrafficPeriod
	PlausibleRange   any
	FallbackAllRange any
}

type nuvioReportsTrafficData struct {
	Summary  nuvioReportsTrafficSummary
	TopPages []nuvioReportsTrafficTopPage
	Sources  []nuvioReportsTrafficSource
	Devices  []nuvioReportsTrafficDevice
}

type nuvioPlausibleQueryResponse struct {
	Results []nuvioPlausibleQueryResult `json:"results"`
}

type nuvioPlausibleQueryResult struct {
	Dimensions []any `json:"dimensions"`
	Metrics    []any `json:"metrics"`
}

// NUVIO CUSTOM START: Reports Phase 2B Plausible traffic endpoint.
func registerNuvioReportsRoutes(e *core.ServeEvent) {
	reportsGroup := e.Router.Group("/api/nuvio/reports").Bind(apis.RequireSuperuserAuth())

	reportsGroup.GET("/traffic", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		periodQuery, err := resolveNuvioReportsTrafficPeriod(strings.TrimSpace(e.Request.URL.Query().Get("period")), time.Now().UTC())
		if err != nil {
			return e.BadRequestError("Invalid period. Use thisMonth, lastMonth, last30Days, or allTime.", nil)
		}

		_, analyticsConfig, err := loadNuvioWebsiteReportsAnalyticsConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load Reports analytics settings.", nil)
		}

		if !analyticsConfig.FeatureAvailable {
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"feature_unavailable",
				"Reports feature is unavailable for this website.",
				periodQuery.Period,
			))
		}

		if !analyticsConfig.Enabled {
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"analytics_disabled",
				"Traffic analytics are disabled for this website.",
				periodQuery.Period,
			))
		}

		provider := strings.ToLower(strings.TrimSpace(analyticsConfig.Provider))
		if provider == "" {
			provider = "plausible"
		}
		if provider != "plausible" {
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"provider_unconfigured",
				"Traffic analytics provider is not configured right now.",
				periodQuery.Period,
			))
		}

		siteID := strings.TrimSpace(analyticsConfig.SiteID)
		if siteID == "" {
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"analytics_not_configured",
				"Traffic analytics are not configured yet for this website.",
				periodQuery.Period,
			))
		}

		plausibleConfig, err := loadNuvioPlausibleConfig()
		if err != nil {
			e.App.Logger().Error(
				"NUVIO reports plausible config missing",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"provider_unconfigured",
				"Traffic analytics provider is not configured right now.",
				periodQuery.Period,
			))
		}

		trafficData, err := fetchNuvioPlausibleTrafficData(
			e.Request.Context(),
			plausibleConfig,
			siteID,
			periodQuery,
		)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO reports plausible query failed",
				"websiteId",
				websiteID,
				"period",
				periodQuery.Period.Key,
				"siteId",
				siteID,
				"error",
				err.Error(),
			)
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"provider_error",
				"Traffic analytics are temporarily unavailable.",
				periodQuery.Period,
			))
		}

		return e.JSON(http.StatusOK, nuvioReportsTrafficResponse{
			State:     "ok",
			Provider:  "plausible",
			SiteID:    siteID,
			Period:    periodQuery.Period,
			Summary:   &trafficData.Summary,
			TopPages:  trafficData.TopPages,
			Sources:   trafficData.Sources,
			Devices:   trafficData.Devices,
			FetchedAt: time.Now().UTC().Format(time.RFC3339),
		})
	})
}

func loadNuvioWebsiteReportsAnalyticsConfig(
	app core.App,
	websiteID string,
) (*core.Record, nuvioWebsiteReportsAnalyticsConfig, error) {
	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		return nil, nuvioWebsiteReportsAnalyticsConfig{}, err
	}

	settings := parseNuvioSettingsObject(website.Get("settings"))
	config := nuvioWebsiteReportsAnalyticsConfig{
		FeatureAvailable: true,
		Provider:         "plausible",
		Enabled:          false,
		SiteID:           "",
		ScriptEnabled:    false,
	}

	if featureFlags, ok := toStringAnyMap(settings["featureFlags"]); ok {
		if value, ok := parseBoolValue(featureFlags["reports"]); ok {
			config.FeatureAvailable = value
		}
	}

	if reportsSettings, ok := toStringAnyMap(settings["reports"]); ok {
		if analyticsSettings, ok := toStringAnyMap(reportsSettings["analytics"]); ok {
			if provider := strings.TrimSpace(parseStringValue(analyticsSettings["provider"])); provider != "" {
				config.Provider = provider
			}
			if value, ok := parseBoolValue(analyticsSettings["enabled"]); ok {
				config.Enabled = value
			}
			config.SiteID = strings.TrimSpace(parseStringValue(analyticsSettings["siteId"]))
			if value, ok := parseBoolValue(analyticsSettings["scriptEnabled"]); ok {
				config.ScriptEnabled = value
			}
		}
	}

	return website, config, nil
}

func loadNuvioPlausibleConfig() (nuvioPlausibleConfig, error) {
	apiKey := strings.TrimSpace(os.Getenv("NUVIO_PLAUSIBLE_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("PLAUSIBLE_API_KEY"))
	}
	if apiKey == "" {
		return nuvioPlausibleConfig{}, fmt.Errorf("missing Plausible API key")
	}

	baseURL := strings.TrimSpace(os.Getenv("NUVIO_PLAUSIBLE_BASE_URL"))
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("PLAUSIBLE_BASE_URL"))
	}
	if baseURL == "" {
		baseURL = nuvioPlausibleDefaultBaseURL
	}

	baseURL = strings.TrimRight(baseURL, "/")
	if !strings.HasPrefix(strings.ToLower(baseURL), "http://") && !strings.HasPrefix(strings.ToLower(baseURL), "https://") {
		baseURL = "https://" + baseURL
	}

	return nuvioPlausibleConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
	}, nil
}

func resolveNuvioReportsTrafficPeriod(rawPeriod string, now time.Time) (nuvioReportsTrafficPeriodQuery, error) {
	periodKey := strings.TrimSpace(rawPeriod)
	if periodKey == "" {
		periodKey = nuvioReportsTrafficPeriodThisMonth
	}

	nowUTC := now.UTC()
	endDate := nowUTC.Format("2006-01-02")

	switch periodKey {
	case nuvioReportsTrafficPeriodThisMonth:
		start := time.Date(nowUTC.Year(), nowUTC.Month(), 1, 0, 0, 0, 0, time.UTC)
		return nuvioReportsTrafficPeriodQuery{
			Period: nuvioReportsTrafficPeriod{
				Key:       nuvioReportsTrafficPeriodThisMonth,
				Label:     "This month",
				StartDate: start.Format("2006-01-02"),
				EndDate:   endDate,
			},
			PlausibleRange: []string{start.Format("2006-01-02"), endDate},
		}, nil
	case nuvioReportsTrafficPeriodLastMonth:
		currentMonthStart := time.Date(nowUTC.Year(), nowUTC.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastMonthStart := currentMonthStart.AddDate(0, -1, 0)
		lastMonthEnd := currentMonthStart.AddDate(0, 0, -1)
		return nuvioReportsTrafficPeriodQuery{
			Period: nuvioReportsTrafficPeriod{
				Key:       nuvioReportsTrafficPeriodLastMonth,
				Label:     "Last month",
				StartDate: lastMonthStart.Format("2006-01-02"),
				EndDate:   lastMonthEnd.Format("2006-01-02"),
			},
			PlausibleRange: []string{
				lastMonthStart.Format("2006-01-02"),
				lastMonthEnd.Format("2006-01-02"),
			},
		}, nil
	case nuvioReportsTrafficPeriodLast30d:
		start := nowUTC.AddDate(0, 0, -29)
		return nuvioReportsTrafficPeriodQuery{
			Period: nuvioReportsTrafficPeriod{
				Key:       nuvioReportsTrafficPeriodLast30d,
				Label:     "Last 30 days",
				StartDate: start.Format("2006-01-02"),
				EndDate:   endDate,
			},
			PlausibleRange: []string{
				start.Format("2006-01-02"),
				endDate,
			},
		}, nil
	case nuvioReportsTrafficPeriodAllTime:
		fallbackStart := "1970-01-01"
		return nuvioReportsTrafficPeriodQuery{
			Period: nuvioReportsTrafficPeriod{
				Key:   nuvioReportsTrafficPeriodAllTime,
				Label: "All time",
			},
			PlausibleRange:   "all",
			FallbackAllRange: []string{fallbackStart, endDate},
		}, nil
	default:
		return nuvioReportsTrafficPeriodQuery{}, fmt.Errorf("invalid period")
	}
}

func buildNuvioReportsTrafficStateResponse(
	state string,
	message string,
	period nuvioReportsTrafficPeriod,
) nuvioReportsTrafficResponse {
	return nuvioReportsTrafficResponse{
		State:   state,
		Message: message,
		Period:  period,
	}
}

func fetchNuvioPlausibleTrafficData(
	ctx context.Context,
	config nuvioPlausibleConfig,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) (*nuvioReportsTrafficData, error) {
	summary, err := fetchNuvioPlausibleSummary(ctx, config, siteID, periodQuery)
	if err != nil {
		return nil, err
	}

	topPages, err := fetchNuvioPlausibleTopPages(ctx, config, siteID, periodQuery)
	if err != nil {
		return nil, err
	}

	sources, err := fetchNuvioPlausibleSources(ctx, config, siteID, periodQuery)
	if err != nil {
		return nil, err
	}

	devices, err := fetchNuvioPlausibleDevices(ctx, config, siteID, periodQuery)
	if err != nil {
		return nil, err
	}

	return &nuvioReportsTrafficData{
		Summary:  summary,
		TopPages: topPages,
		Sources:  sources,
		Devices:  devices,
	}, nil
}

func fetchNuvioPlausibleSummary(
	ctx context.Context,
	config nuvioPlausibleConfig,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) (nuvioReportsTrafficSummary, error) {
	trafficResult, err := queryNuvioPlausibleWithPeriodFallback(ctx, config, map[string]any{
		"site_id":    siteID,
		"metrics":    []string{"visitors", "pageviews"},
		"date_range": periodQuery.PlausibleRange,
	}, periodQuery)
	if err != nil {
		return nuvioReportsTrafficSummary{}, err
	}

	engagementResult, err := queryNuvioPlausibleWithPeriodFallback(ctx, config, map[string]any{
		"site_id":    siteID,
		"metrics":    []string{"bounce_rate", "visit_duration"},
		"date_range": periodQuery.PlausibleRange,
	}, periodQuery)
	if err != nil {
		return nuvioReportsTrafficSummary{}, err
	}

	summary := nuvioReportsTrafficSummary{
		Visitors:  0,
		Pageviews: 0,
	}

	if len(trafficResult.Results) > 0 {
		row := trafficResult.Results[0]
		summary.Visitors = parseNuvioPlausibleMetricAsInt(row.Metrics, 0)
		summary.Pageviews = parseNuvioPlausibleMetricAsInt(row.Metrics, 1)
	}

	if len(engagementResult.Results) > 0 {
		row := engagementResult.Results[0]
		if bounceRate, ok := parseNuvioPlausibleMetricAsFloat(row.Metrics, 0); ok {
			bounce := roundNuvioFloat(bounceRate, 2)
			summary.BounceRate = &bounce
		}
		if duration, ok := parseNuvioPlausibleMetricAsFloat(row.Metrics, 1); ok {
			seconds := int(duration)
			summary.VisitDurationSeconds = &seconds
		}
	}

	return summary, nil
}

func fetchNuvioPlausibleTopPages(
	ctx context.Context,
	config nuvioPlausibleConfig,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficTopPage, error) {
	result, err := queryNuvioPlausibleWithPeriodFallback(ctx, config, map[string]any{
		"site_id":    siteID,
		"metrics":    []string{"visitors", "pageviews"},
		"dimensions": []string{"event:page"},
		"date_range": periodQuery.PlausibleRange,
		"order_by":   [][]string{{"visitors", "desc"}},
		"pagination": map[string]int{"limit": 10, "offset": 0},
	}, periodQuery)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficTopPage, 0, len(result.Results))
	for _, row := range result.Results {
		page := normalizeNuvioPlausibleDimension(row.Dimensions, 0, "/")
		items = append(items, nuvioReportsTrafficTopPage{
			Page:      page,
			Visitors:  parseNuvioPlausibleMetricAsInt(row.Metrics, 0),
			Pageviews: parseNuvioPlausibleMetricAsInt(row.Metrics, 1),
		})
	}

	return items, nil
}

func fetchNuvioPlausibleSources(
	ctx context.Context,
	config nuvioPlausibleConfig,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficSource, error) {
	result, err := queryNuvioPlausibleWithPeriodFallback(ctx, config, map[string]any{
		"site_id":    siteID,
		"metrics":    []string{"visitors"},
		"dimensions": []string{"visit:source"},
		"date_range": periodQuery.PlausibleRange,
		"order_by":   [][]string{{"visitors", "desc"}},
		"pagination": map[string]int{"limit": 10, "offset": 0},
	}, periodQuery)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficSource, 0, len(result.Results))
	for _, row := range result.Results {
		source := normalizeNuvioPlausibleDimension(row.Dimensions, 0, "Direct")
		items = append(items, nuvioReportsTrafficSource{
			Source:   source,
			Visitors: parseNuvioPlausibleMetricAsInt(row.Metrics, 0),
		})
	}

	return items, nil
}

func fetchNuvioPlausibleDevices(
	ctx context.Context,
	config nuvioPlausibleConfig,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficDevice, error) {
	result, err := queryNuvioPlausibleWithPeriodFallback(ctx, config, map[string]any{
		"site_id":    siteID,
		"metrics":    []string{"visitors"},
		"dimensions": []string{"visit:device"},
		"date_range": periodQuery.PlausibleRange,
		"order_by":   [][]string{{"visitors", "desc"}},
		"pagination": map[string]int{"limit": 10, "offset": 0},
	}, periodQuery)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficDevice, 0, len(result.Results))
	for _, row := range result.Results {
		device := normalizeNuvioPlausibleDimension(row.Dimensions, 0, "Unknown")
		items = append(items, nuvioReportsTrafficDevice{
			Device:   device,
			Visitors: parseNuvioPlausibleMetricAsInt(row.Metrics, 0),
		})
	}

	return items, nil
}

func queryNuvioPlausibleWithPeriodFallback(
	ctx context.Context,
	config nuvioPlausibleConfig,
	query map[string]any,
	periodQuery nuvioReportsTrafficPeriodQuery,
) (*nuvioPlausibleQueryResponse, error) {
	response, err := queryNuvioPlausible(ctx, config, query)
	if err == nil {
		return response, nil
	}

	if periodQuery.Period.Key != nuvioReportsTrafficPeriodAllTime || periodQuery.FallbackAllRange == nil {
		return nil, err
	}

	fallbackQuery := cloneNuvioPlausibleQueryMap(query)
	fallbackQuery["date_range"] = periodQuery.FallbackAllRange
	return queryNuvioPlausible(ctx, config, fallbackQuery)
}

func queryNuvioPlausible(
	ctx context.Context,
	config nuvioPlausibleConfig,
	query map[string]any,
) (*nuvioPlausibleQueryResponse, error) {
	rawPayload, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to encode Plausible query: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(config.BaseURL, "/")+"/api/v2/query",
		bytes.NewBuffer(rawPayload),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build Plausible request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(config.APIKey))
	request.Header.Set("Content-Type", "application/json")

	response, err := nuvioReportsHTTPClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Plausible: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read Plausible response: %w", err)
	}

	if response.StatusCode >= 400 {
		return nil, fmt.Errorf("plausible rejected query (%d): %s", response.StatusCode, parseNuvioPlausibleProviderMessage(body))
	}

	decoded := &nuvioPlausibleQueryResponse{}
	if err := json.Unmarshal(body, decoded); err != nil {
		return nil, fmt.Errorf("failed to decode Plausible response: %w", err)
	}

	if decoded.Results == nil {
		decoded.Results = []nuvioPlausibleQueryResult{}
	}

	return decoded, nil
}

func parseNuvioPlausibleProviderMessage(raw []byte) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return "unknown provider error"
	}

	parsed := map[string]any{}
	if err := json.Unmarshal(raw, &parsed); err == nil {
		for _, key := range []string{"error", "message"} {
			if value := strings.TrimSpace(parseStringValue(parsed[key])); value != "" {
				return value
			}
		}
	}

	return trimmed
}

func cloneNuvioPlausibleQueryMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func normalizeNuvioPlausibleDimension(dimensions []any, index int, fallback string) string {
	if index < 0 || index >= len(dimensions) {
		return fallback
	}

	value := strings.TrimSpace(parseStringValue(dimensions[index]))
	if value == "" {
		return fallback
	}

	return value
}

func parseNuvioPlausibleMetricAsInt(metrics []any, index int) int {
	value, ok := parseNuvioPlausibleMetricAsFloat(metrics, index)
	if !ok {
		return 0
	}
	return int(value)
}

func parseNuvioPlausibleMetricAsFloat(metrics []any, index int) (float64, bool) {
	if index < 0 || index >= len(metrics) {
		return 0, false
	}

	switch typed := metrics[index].(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		if value, err := typed.Float64(); err == nil {
			return value, true
		}
	case string:
		if value, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			return value, true
		}
	}

	return 0, false
}

func roundNuvioFloat(value float64, precision int) float64 {
	if precision < 0 {
		precision = 0
	}

	pow := 1.0
	for i := 0; i < precision; i++ {
		pow *= 10
	}

	return float64(int(value*pow+0.5)) / pow
}

// NUVIO CUSTOM END: Reports Phase 2B Plausible traffic endpoint.
