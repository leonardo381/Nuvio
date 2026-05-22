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
	"net/url"
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
	ScriptURL        string
	APIURL           string
}

type nuvioUmamiConfig struct {
	APIBaseURL     string
	RequestBaseURL string
	LoginURL       string
	APIKey         string
	Username       string
	Password       string
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

type nuvioReportsTrafficEntryPage struct {
	Page      string `json:"page"`
	Visitors  int    `json:"visitors"`
	Pageviews int    `json:"pageviews"`
}

type nuvioReportsTrafficExitPage struct {
	Page      string `json:"page"`
	Visitors  int    `json:"visitors"`
	Pageviews int    `json:"pageviews"`
}

type nuvioReportsTrafficCountry struct {
	Country  string `json:"country"`
	Visitors int    `json:"visitors"`
}

type nuvioReportsTrafficRegion struct {
	Region   string `json:"region"`
	Visitors int    `json:"visitors"`
}

type nuvioReportsTrafficCity struct {
	City     string `json:"city"`
	Visitors int    `json:"visitors"`
}

type nuvioReportsTrafficBrowser struct {
	Browser  string `json:"browser"`
	Visitors int    `json:"visitors"`
}

type nuvioReportsTrafficOperatingSystem struct {
	OperatingSystem string `json:"operatingSystem"`
	Visitors        int    `json:"visitors"`
}

type nuvioReportsTrafficResponse struct {
	State            string                               `json:"state"`
	Message          string                               `json:"message,omitempty"`
	Provider         string                               `json:"provider,omitempty"`
	SiteID           string                               `json:"siteId,omitempty"`
	Period           nuvioReportsTrafficPeriod            `json:"period"`
	Summary          *nuvioReportsTrafficSummary          `json:"summary,omitempty"`
	TopPages         []nuvioReportsTrafficTopPage         `json:"topPages,omitempty"`
	Sources          []nuvioReportsTrafficSource          `json:"sources,omitempty"`
	Devices          []nuvioReportsTrafficDevice          `json:"devices,omitempty"`
	EntryPages       []nuvioReportsTrafficEntryPage       `json:"entryPages"`
	ExitPages        []nuvioReportsTrafficExitPage        `json:"exitPages"`
	Countries        []nuvioReportsTrafficCountry         `json:"countries"`
	Regions          []nuvioReportsTrafficRegion          `json:"regions"`
	Cities           []nuvioReportsTrafficCity            `json:"cities"`
	Browsers         []nuvioReportsTrafficBrowser         `json:"browsers"`
	OperatingSystems []nuvioReportsTrafficOperatingSystem `json:"operatingSystems"`
	FetchedAt        string                               `json:"fetchedAt,omitempty"`
}

type nuvioReportsTrafficPeriodQuery struct {
	Period    nuvioReportsTrafficPeriod
	StartAtMs int64
	EndAtMs   int64
	Unit      string
	Timezone  string
}

type nuvioReportsTrafficData struct {
	Summary          nuvioReportsTrafficSummary
	TopPages         []nuvioReportsTrafficTopPage
	Sources          []nuvioReportsTrafficSource
	Devices          []nuvioReportsTrafficDevice
	EntryPages       []nuvioReportsTrafficEntryPage
	ExitPages        []nuvioReportsTrafficExitPage
	Countries        []nuvioReportsTrafficCountry
	Regions          []nuvioReportsTrafficRegion
	Cities           []nuvioReportsTrafficCity
	Browsers         []nuvioReportsTrafficBrowser
	OperatingSystems []nuvioReportsTrafficOperatingSystem
}

type nuvioUmamiStatsResponse struct {
	Pageviews any `json:"pageviews"`
	Visitors  any `json:"visitors"`
	Visits    any `json:"visits"`
	Bounces   any `json:"bounces"`
	Totaltime any `json:"totaltime"`
}

type nuvioUmamiMetricRow struct {
	Name      any `json:"name"`
	Pageviews any `json:"pageviews"`
	Visitors  any `json:"visitors"`
}

type nuvioUmamiPageviewsPoint struct {
	X string `json:"x"`
	Y any    `json:"y"`
}

type nuvioUmamiPageviewsResponse struct {
	Pageviews []nuvioUmamiPageviewsPoint `json:"pageviews"`
	Sessions  []nuvioUmamiPageviewsPoint `json:"sessions"`
}

type nuvioReportsTrafficStateError struct {
	State   string
	Message string
	Cause   error
}

func (err *nuvioReportsTrafficStateError) Error() string {
	if err == nil {
		return ""
	}

	if err.Cause != nil {
		return err.Cause.Error()
	}

	if strings.TrimSpace(err.Message) != "" {
		return err.Message
	}

	return err.State
}

func (err *nuvioReportsTrafficStateError) Unwrap() error {
	if err == nil {
		return nil
	}

	return err.Cause
}

// NUVIO CUSTOM START: Reports traffic endpoint.
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
				"",
				"",
			))
		}

		if !analyticsConfig.Enabled {
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"analytics_disabled",
				"Traffic analytics are disabled for this website.",
				periodQuery.Period,
				"",
				"",
			))
		}

		provider := strings.ToLower(strings.TrimSpace(analyticsConfig.Provider))
		if provider == "" {
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"provider_unconfigured",
				"Traffic analytics provider is not configured right now.",
				periodQuery.Period,
				"",
				"",
			))
		}

		if provider != "umami" {
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"provider_unsupported",
				"Traffic analytics provider is not configured right now.",
				periodQuery.Period,
				provider,
				"",
			))
		}

		siteID := strings.TrimSpace(analyticsConfig.SiteID)
		if siteID == "" {
			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"analytics_not_configured",
				"Traffic analytics are not configured yet for this website.",
				periodQuery.Period,
				provider,
				"",
			))
		}

		umamiConfig, err := loadNuvioUmamiConfig(analyticsConfig)
		if err != nil {
			if stateErr, ok := unwrapNuvioReportsTrafficStateError(err); ok {
				return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
					stateErr.State,
					stateErr.Message,
					periodQuery.Period,
					provider,
					siteID,
				))
			}

			e.App.Logger().Error(
				"NUVIO reports analytics config error",
				"websiteId",
				websiteID,
				"provider",
				provider,
				"error",
				err.Error(),
			)

			return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
				"provider_error",
				"Traffic analytics are temporarily unavailable.",
				periodQuery.Period,
				provider,
				siteID,
			))
		}

		trafficData, partialIssueCount, err := fetchNuvioUmamiTrafficData(
			e.Request.Context(),
			umamiConfig,
			siteID,
			periodQuery,
		)
		if err != nil {
			if stateErr, ok := unwrapNuvioReportsTrafficStateError(err); ok {
				return e.JSON(http.StatusOK, buildNuvioReportsTrafficStateResponse(
					stateErr.State,
					stateErr.Message,
					periodQuery.Period,
					provider,
					siteID,
				))
			}

			e.App.Logger().Error(
				"NUVIO reports umami query failed",
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
				provider,
				siteID,
			))
		}

		responseMessage := ""
		if partialIssueCount > 0 {
			responseMessage = "Some traffic metrics are temporarily unavailable."
		}

		return e.JSON(http.StatusOK, nuvioReportsTrafficResponse{
			State:            "ok",
			Message:          responseMessage,
			Provider:         "umami",
			SiteID:           siteID,
			Period:           periodQuery.Period,
			Summary:          &trafficData.Summary,
			TopPages:         trafficData.TopPages,
			Sources:          trafficData.Sources,
			Devices:          trafficData.Devices,
			EntryPages:       trafficData.EntryPages,
			ExitPages:        trafficData.ExitPages,
			Countries:        trafficData.Countries,
			Regions:          trafficData.Regions,
			Cities:           trafficData.Cities,
			Browsers:         trafficData.Browsers,
			OperatingSystems: trafficData.OperatingSystems,
			FetchedAt:        time.Now().UTC().Format(time.RFC3339),
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
		Provider:         "",
		Enabled:          false,
		SiteID:           "",
		ScriptEnabled:    false,
		ScriptURL:        "",
		APIURL:           "",
	}

	if featureFlags, ok := toStringAnyMap(settings["featureFlags"]); ok {
		if value, ok := parseBoolValue(featureFlags["reports"]); ok {
			config.FeatureAvailable = value
		}
	}

	if reportsSettings, ok := toStringAnyMap(settings["reports"]); ok {
		if analyticsSettings, ok := toStringAnyMap(reportsSettings["analytics"]); ok {
			provider := strings.ToLower(strings.TrimSpace(parseStringValue(analyticsSettings["provider"])))
			if provider == "umami" {
				config.Provider = "umami"
			}
			if value, ok := parseBoolValue(analyticsSettings["enabled"]); ok {
				config.Enabled = value
			}
			config.SiteID = strings.TrimSpace(parseStringValue(analyticsSettings["siteId"]))
			if value, ok := parseBoolValue(analyticsSettings["scriptEnabled"]); ok {
				config.ScriptEnabled = value
			}
			config.ScriptURL = strings.TrimSpace(parseStringValue(analyticsSettings["scriptUrl"]))
			config.APIURL = strings.TrimSpace(parseStringValue(analyticsSettings["apiUrl"]))
		}
	}

	return website, config, nil
}

func loadNuvioUmamiConfig(analyticsConfig nuvioWebsiteReportsAnalyticsConfig) (nuvioUmamiConfig, error) {
	rawAPIURL := strings.TrimSpace(analyticsConfig.APIURL)
	if rawAPIURL == "" {
		rawAPIURL = strings.TrimSpace(os.Getenv("NUVIO_UMAMI_API_URL"))
	}

	if rawAPIURL == "" {
		return nuvioUmamiConfig{}, newNuvioReportsTrafficStateError(
			"provider_unconfigured",
			"Traffic analytics provider is not configured right now.",
			nil,
		)
	}

	normalizedAPIURL, err := normalizeNuvioAnalyticsURL(rawAPIURL)
	if err != nil {
		return nuvioUmamiConfig{}, newNuvioReportsTrafficStateError(
			"provider_unconfigured",
			"Traffic analytics provider is not configured right now.",
			err,
		)
	}

	requestBaseURL, loginURL := resolveNuvioUmamiURLs(normalizedAPIURL)
	if requestBaseURL == "" {
		return nuvioUmamiConfig{}, newNuvioReportsTrafficStateError(
			"provider_unconfigured",
			"Traffic analytics provider is not configured right now.",
			nil,
		)
	}

	apiKey := strings.TrimSpace(os.Getenv("NUVIO_UMAMI_API_KEY"))
	username := strings.TrimSpace(os.Getenv("NUVIO_UMAMI_USERNAME"))
	password := strings.TrimSpace(os.Getenv("NUVIO_UMAMI_PASSWORD"))

	if apiKey == "" && (username == "" || password == "") {
		return nuvioUmamiConfig{}, newNuvioReportsTrafficStateError(
			"provider_auth_missing",
			"Traffic analytics provider authentication is not configured yet.",
			nil,
		)
	}

	if apiKey == "" && loginURL == "" {
		return nuvioUmamiConfig{}, newNuvioReportsTrafficStateError(
			"provider_unconfigured",
			"Traffic analytics provider is not configured right now.",
			nil,
		)
	}

	return nuvioUmamiConfig{
		APIBaseURL:     normalizedAPIURL,
		RequestBaseURL: requestBaseURL,
		LoginURL:       loginURL,
		APIKey:         apiKey,
		Username:       username,
		Password:       password,
	}, nil
}

func resolveNuvioReportsTrafficPeriod(rawPeriod string, now time.Time) (nuvioReportsTrafficPeriodQuery, error) {
	periodKey := strings.TrimSpace(rawPeriod)
	if periodKey == "" {
		periodKey = nuvioReportsTrafficPeriodThisMonth
	}

	nowUTC := now.UTC()
	endDate := nowUTC.Format("2006-01-02")
	endAtMs := nowUTC.UnixMilli()

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
			StartAtMs: start.UnixMilli(),
			EndAtMs:   endAtMs,
			Unit:      "day",
			Timezone:  "UTC",
		}, nil
	case nuvioReportsTrafficPeriodLastMonth:
		currentMonthStart := time.Date(nowUTC.Year(), nowUTC.Month(), 1, 0, 0, 0, 0, time.UTC)
		lastMonthStart := currentMonthStart.AddDate(0, -1, 0)
		lastMonthEnd := currentMonthStart.AddDate(0, 0, -1)
		lastMonthEndAt := time.Date(lastMonthEnd.Year(), lastMonthEnd.Month(), lastMonthEnd.Day(), 23, 59, 59, int(time.Second-time.Millisecond), time.UTC)
		return nuvioReportsTrafficPeriodQuery{
			Period: nuvioReportsTrafficPeriod{
				Key:       nuvioReportsTrafficPeriodLastMonth,
				Label:     "Last month",
				StartDate: lastMonthStart.Format("2006-01-02"),
				EndDate:   lastMonthEnd.Format("2006-01-02"),
			},
			StartAtMs: lastMonthStart.UnixMilli(),
			EndAtMs:   lastMonthEndAt.UnixMilli(),
			Unit:      "day",
			Timezone:  "UTC",
		}, nil
	case nuvioReportsTrafficPeriodLast30d:
		start := nowUTC.AddDate(0, 0, -29)
		startAt := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
		return nuvioReportsTrafficPeriodQuery{
			Period: nuvioReportsTrafficPeriod{
				Key:       nuvioReportsTrafficPeriodLast30d,
				Label:     "Last 30 days",
				StartDate: startAt.Format("2006-01-02"),
				EndDate:   endDate,
			},
			StartAtMs: startAt.UnixMilli(),
			EndAtMs:   endAtMs,
			Unit:      "day",
			Timezone:  "UTC",
		}, nil
	case nuvioReportsTrafficPeriodAllTime:
		start := time.Unix(0, 0).UTC()
		return nuvioReportsTrafficPeriodQuery{
			Period: nuvioReportsTrafficPeriod{
				Key:   nuvioReportsTrafficPeriodAllTime,
				Label: "All time",
			},
			StartAtMs: start.UnixMilli(),
			EndAtMs:   endAtMs,
			Unit:      "month",
			Timezone:  "UTC",
		}, nil
	default:
		return nuvioReportsTrafficPeriodQuery{}, fmt.Errorf("invalid period")
	}
}

func buildNuvioReportsTrafficStateResponse(
	state string,
	message string,
	period nuvioReportsTrafficPeriod,
	provider string,
	siteID string,
) nuvioReportsTrafficResponse {
	return nuvioReportsTrafficResponse{
		State:            state,
		Message:          message,
		Provider:         strings.TrimSpace(provider),
		SiteID:           strings.TrimSpace(siteID),
		Period:           period,
		EntryPages:       []nuvioReportsTrafficEntryPage{},
		ExitPages:        []nuvioReportsTrafficExitPage{},
		Countries:        []nuvioReportsTrafficCountry{},
		Regions:          []nuvioReportsTrafficRegion{},
		Cities:           []nuvioReportsTrafficCity{},
		Browsers:         []nuvioReportsTrafficBrowser{},
		OperatingSystems: []nuvioReportsTrafficOperatingSystem{},
	}
}

func fetchNuvioUmamiTrafficData(
	ctx context.Context,
	config nuvioUmamiConfig,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) (*nuvioReportsTrafficData, int, error) {
	authHeaders, err := resolveNuvioUmamiAuthHeaders(ctx, config)
	if err != nil {
		return nil, 0, err
	}

	summary, err := fetchNuvioUmamiSummary(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		return nil, 0, err
	}

	partialIssueCount := 0

	topPages, err := fetchNuvioUmamiTopPages(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		topPages = []nuvioReportsTrafficTopPage{}
	}

	sources, err := fetchNuvioUmamiSources(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		sources = []nuvioReportsTrafficSource{}
	}

	devices, err := fetchNuvioUmamiDevices(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		devices = []nuvioReportsTrafficDevice{}
	}

	entryPages, err := fetchNuvioUmamiEntryPages(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		entryPages = []nuvioReportsTrafficEntryPage{}
	}

	exitPages, err := fetchNuvioUmamiExitPages(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		exitPages = []nuvioReportsTrafficExitPage{}
	}

	countries, err := fetchNuvioUmamiCountries(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		countries = []nuvioReportsTrafficCountry{}
	}

	regions, err := fetchNuvioUmamiRegions(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		regions = []nuvioReportsTrafficRegion{}
	}

	cities, err := fetchNuvioUmamiCities(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		cities = []nuvioReportsTrafficCity{}
	}

	browsers, err := fetchNuvioUmamiBrowsers(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		browsers = []nuvioReportsTrafficBrowser{}
	}

	operatingSystems, err := fetchNuvioUmamiOperatingSystems(ctx, config, authHeaders, siteID, periodQuery)
	if err != nil {
		partialIssueCount++
		operatingSystems = []nuvioReportsTrafficOperatingSystem{}
	}

	if err := fetchNuvioUmamiPageviewsSeries(ctx, config, authHeaders, siteID, periodQuery); err != nil {
		partialIssueCount++
	}

	return &nuvioReportsTrafficData{
		Summary:          summary,
		TopPages:         topPages,
		Sources:          sources,
		Devices:          devices,
		EntryPages:       entryPages,
		ExitPages:        exitPages,
		Countries:        countries,
		Regions:          regions,
		Cities:           cities,
		Browsers:         browsers,
		OperatingSystems: operatingSystems,
	}, partialIssueCount, nil
}

func fetchNuvioUmamiSummary(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) (nuvioReportsTrafficSummary, error) {
	query := buildNuvioUmamiBaseQuery(periodQuery)
	endpointPath := buildNuvioUmamiWebsiteEndpointPath(siteID, "stats")
	requestURL := buildNuvioUmamiRequestURL(config.RequestBaseURL, endpointPath, query)

	stats := nuvioUmamiStatsResponse{}
	if err := executeNuvioUmamiJSONRequest(ctx, http.MethodGet, requestURL, authHeaders, nil, &stats); err != nil {
		return nuvioReportsTrafficSummary{}, err
	}

	visitors := parseNuvioUmamiAnyInt(stats.Visitors)
	pageviews := parseNuvioUmamiAnyInt(stats.Pageviews)
	visits := parseNuvioUmamiAnyInt(stats.Visits)
	bounces, hasBounces := parseNuvioUmamiAnyFloat(stats.Bounces)
	totalTime, hasTotalTime := parseNuvioUmamiAnyFloat(stats.Totaltime)

	summary := nuvioReportsTrafficSummary{
		Visitors:  visitors,
		Pageviews: pageviews,
	}

	if visits > 0 && hasBounces {
		bounceRate := bounces / float64(visits)
		if bounceRate >= 0 {
			cleaned := roundNuvioFloat(bounceRate, 4)
			summary.BounceRate = &cleaned
		}
	}

	if visits > 0 && hasTotalTime {
		avgDuration := int(totalTime / float64(visits))
		if avgDuration >= 0 {
			summary.VisitDurationSeconds = &avgDuration
		}
	}

	return summary, nil
}

func fetchNuvioUmamiTopPages(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficTopPage, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "path", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficTopPage, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficTopPage{
			Page:      parseNuvioUmamiAnyString(row.Name, "/"),
			Visitors:  parseNuvioUmamiAnyInt(row.Visitors),
			Pageviews: parseNuvioUmamiAnyInt(row.Pageviews),
		})
	}

	return items, nil
}

func fetchNuvioUmamiSources(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficSource, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "referrer", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficSource, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficSource{
			Source:   parseNuvioUmamiAnyString(row.Name, "Direct"),
			Visitors: parseNuvioUmamiAnyInt(row.Visitors),
		})
	}

	return items, nil
}

func fetchNuvioUmamiDevices(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficDevice, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "device", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficDevice, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficDevice{
			Device:   parseNuvioUmamiAnyString(row.Name, "Unknown"),
			Visitors: parseNuvioUmamiAnyInt(row.Visitors),
		})
	}

	return items, nil
}

func fetchNuvioUmamiEntryPages(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficEntryPage, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "entry", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficEntryPage, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficEntryPage{
			Page:      parseNuvioUmamiAnyString(row.Name, "/"),
			Visitors:  parseNuvioUmamiAnyInt(row.Visitors),
			Pageviews: parseNuvioUmamiAnyInt(row.Pageviews),
		})
	}

	return items, nil
}

func fetchNuvioUmamiExitPages(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficExitPage, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "exit", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficExitPage, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficExitPage{
			Page:      parseNuvioUmamiAnyString(row.Name, "/"),
			Visitors:  parseNuvioUmamiAnyInt(row.Visitors),
			Pageviews: parseNuvioUmamiAnyInt(row.Pageviews),
		})
	}

	return items, nil
}

func fetchNuvioUmamiCountries(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficCountry, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "country", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficCountry, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficCountry{
			Country:  parseNuvioUmamiAnyString(row.Name, "Unknown"),
			Visitors: parseNuvioUmamiAnyInt(row.Visitors),
		})
	}

	return items, nil
}

func fetchNuvioUmamiRegions(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficRegion, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "region", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficRegion, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficRegion{
			Region:   parseNuvioUmamiAnyString(row.Name, "Unknown"),
			Visitors: parseNuvioUmamiAnyInt(row.Visitors),
		})
	}

	return items, nil
}

func fetchNuvioUmamiCities(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficCity, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "city", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficCity, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficCity{
			City:     parseNuvioUmamiAnyString(row.Name, "Unknown"),
			Visitors: parseNuvioUmamiAnyInt(row.Visitors),
		})
	}

	return items, nil
}

func fetchNuvioUmamiBrowsers(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficBrowser, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "browser", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficBrowser, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficBrowser{
			Browser:  parseNuvioUmamiAnyString(row.Name, "Unknown"),
			Visitors: parseNuvioUmamiAnyInt(row.Visitors),
		})
	}

	return items, nil
}

func fetchNuvioUmamiOperatingSystems(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioReportsTrafficOperatingSystem, error) {
	rows, err := fetchNuvioUmamiMetricsExpanded(ctx, config, authHeaders, siteID, periodQuery, "os", 10)
	if err != nil {
		return nil, err
	}

	items := make([]nuvioReportsTrafficOperatingSystem, 0, len(rows))
	for _, row := range rows {
		items = append(items, nuvioReportsTrafficOperatingSystem{
			OperatingSystem: parseNuvioUmamiAnyString(row.Name, "Unknown"),
			Visitors:        parseNuvioUmamiAnyInt(row.Visitors),
		})
	}

	return items, nil
}

func fetchNuvioUmamiMetricsExpanded(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
	metricType string,
	limit int,
) ([]nuvioUmamiMetricRow, error) {
	query := buildNuvioUmamiBaseQuery(periodQuery)
	query.Set("type", strings.TrimSpace(metricType))
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}

	endpointPath := buildNuvioUmamiWebsiteEndpointPath(siteID, "metrics/expanded")
	requestURL := buildNuvioUmamiRequestURL(config.RequestBaseURL, endpointPath, query)

	rows := []nuvioUmamiMetricRow{}
	if err := executeNuvioUmamiJSONRequest(ctx, http.MethodGet, requestURL, authHeaders, nil, &rows); err != nil {
		return nil, err
	}

	if rows == nil {
		return []nuvioUmamiMetricRow{}, nil
	}

	return rows, nil
}

func fetchNuvioUmamiPageviewsSeries(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) error {
	query := buildNuvioUmamiBaseQuery(periodQuery)
	if periodQuery.Unit != "" {
		query.Set("unit", periodQuery.Unit)
	}
	if periodQuery.Timezone != "" {
		query.Set("timezone", periodQuery.Timezone)
	}

	endpointPath := buildNuvioUmamiWebsiteEndpointPath(siteID, "pageviews")
	requestURL := buildNuvioUmamiRequestURL(config.RequestBaseURL, endpointPath, query)

	payload := nuvioUmamiPageviewsResponse{}
	if err := executeNuvioUmamiJSONRequest(ctx, http.MethodGet, requestURL, authHeaders, nil, &payload); err != nil {
		return err
	}

	return nil
}

func resolveNuvioUmamiAuthHeaders(ctx context.Context, config nuvioUmamiConfig) (map[string]string, error) {
	if strings.TrimSpace(config.APIKey) != "" {
		return map[string]string{
			"x-umami-api-key": strings.TrimSpace(config.APIKey),
		}, nil
	}

	username := strings.TrimSpace(config.Username)
	password := strings.TrimSpace(config.Password)
	if username == "" || password == "" {
		return nil, newNuvioReportsTrafficStateError(
			"provider_auth_missing",
			"Traffic analytics provider authentication is not configured yet.",
			nil,
		)
	}

	loginURL := strings.TrimSpace(config.LoginURL)
	if loginURL == "" {
		return nil, newNuvioReportsTrafficStateError(
			"provider_unconfigured",
			"Traffic analytics provider is not configured right now.",
			nil,
		)
	}

	requestBody := map[string]string{
		"username": username,
		"password": password,
	}

	responseBody := struct {
		Token string `json:"token"`
	}{}

	if err := executeNuvioUmamiJSONRequest(ctx, http.MethodPost, loginURL, map[string]string{}, requestBody, &responseBody); err != nil {
		return nil, err
	}

	token := strings.TrimSpace(responseBody.Token)
	if token == "" {
		return nil, newNuvioReportsTrafficStateError(
			"provider_auth_error",
			"Unable to authenticate with the analytics provider.",
			nil,
		)
	}

	return map[string]string{
		"Authorization": "Bearer " + token,
	}, nil
}

func executeNuvioUmamiJSONRequest(
	ctx context.Context,
	method string,
	requestURL string,
	headers map[string]string,
	requestBody any,
	target any,
) error {
	var bodyReader io.Reader
	if requestBody != nil {
		rawBody, err := json.Marshal(requestBody)
		if err != nil {
			return newNuvioReportsTrafficStateError(
				"provider_error",
				"Traffic analytics are temporarily unavailable.",
				fmt.Errorf("failed to encode Umami request payload: %w", err),
			)
		}
		bodyReader = bytes.NewBuffer(rawBody)
	}

	request, err := http.NewRequestWithContext(ctx, method, requestURL, bodyReader)
	if err != nil {
		return newNuvioReportsTrafficStateError(
			"provider_error",
			"Traffic analytics are temporarily unavailable.",
			fmt.Errorf("failed to build Umami request: %w", err),
		)
	}

	request.Header.Set("Accept", "application/json")
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		headerKey := strings.TrimSpace(key)
		headerValue := strings.TrimSpace(value)
		if headerKey == "" || headerValue == "" {
			continue
		}
		request.Header.Set(headerKey, headerValue)
	}

	response, err := nuvioReportsHTTPClient.Do(request)
	if err != nil {
		return newNuvioReportsTrafficStateError(
			"provider_error",
			"Traffic analytics are temporarily unavailable.",
			fmt.Errorf("failed to reach Umami provider: %w", err),
		)
	}
	defer response.Body.Close()

	responseRaw, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return newNuvioReportsTrafficStateError(
			"provider_error",
			"Traffic analytics are temporarily unavailable.",
			fmt.Errorf("failed reading Umami response: %w", err),
		)
	}

	if response.StatusCode >= http.StatusBadRequest {
		switch response.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return newNuvioReportsTrafficStateError(
				"provider_auth_error",
				"Unable to authenticate with the analytics provider.",
				nil,
			)
		case http.StatusNotFound:
			return newNuvioReportsTrafficStateError(
				"provider_not_found",
				"Analytics website was not found for this configuration.",
				nil,
			)
		default:
			providerMessage := parseNuvioUmamiProviderMessage(responseRaw)
			return newNuvioReportsTrafficStateError(
				"provider_error",
				"Traffic analytics are temporarily unavailable.",
				fmt.Errorf("umami request failed (%d): %s", response.StatusCode, providerMessage),
			)
		}
	}

	if target == nil {
		return nil
	}

	if len(strings.TrimSpace(string(responseRaw))) == 0 {
		return newNuvioReportsTrafficStateError(
			"provider_error",
			"Traffic analytics are temporarily unavailable.",
			fmt.Errorf("umami response is empty"),
		)
	}

	if err := json.Unmarshal(responseRaw, target); err != nil {
		return newNuvioReportsTrafficStateError(
			"provider_error",
			"Traffic analytics are temporarily unavailable.",
			fmt.Errorf("failed decoding Umami response: %w", err),
		)
	}

	return nil
}

func parseNuvioUmamiProviderMessage(raw []byte) string {
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

func buildNuvioUmamiBaseQuery(periodQuery nuvioReportsTrafficPeriodQuery) url.Values {
	query := url.Values{}
	query.Set("startAt", strconv.FormatInt(periodQuery.StartAtMs, 10))
	query.Set("endAt", strconv.FormatInt(periodQuery.EndAtMs, 10))
	return query
}

func buildNuvioUmamiWebsiteEndpointPath(siteID string, suffix string) string {
	safeSiteID := strings.TrimSpace(siteID)
	safeSuffix := strings.Trim(strings.TrimSpace(suffix), "/")
	if safeSuffix == "" {
		return "websites/" + url.PathEscape(safeSiteID)
	}
	return "websites/" + url.PathEscape(safeSiteID) + "/" + safeSuffix
}

func buildNuvioUmamiRequestURL(baseURL string, path string, query url.Values) string {
	normalizedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	normalizedPath := "/" + strings.Trim(strings.TrimSpace(path), "/")
	requestURL := normalizedBase + normalizedPath
	if query != nil {
		encoded := query.Encode()
		if encoded != "" {
			requestURL += "?" + encoded
		}
	}
	return requestURL
}

func resolveNuvioUmamiURLs(apiURL string) (requestBaseURL string, loginURL string) {
	normalized := strings.TrimRight(strings.TrimSpace(apiURL), "/")
	if normalized == "" {
		return "", ""
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", ""
	}

	path := strings.TrimRight(strings.TrimSpace(parsed.Path), "/")
	lowerPath := strings.ToLower(path)

	switch {
	case strings.HasSuffix(lowerPath, "/api") || strings.Contains(lowerPath, "/api/"):
		requestBaseURL = normalized
		loginURL = requestBaseURL + "/auth/login"
	case strings.HasPrefix(lowerPath, "/v1") || strings.Contains(lowerPath, "/v1/"):
		requestBaseURL = normalized
		loginURL = ""
	case lowerPath == "":
		requestBaseURL = normalized + "/api"
		loginURL = requestBaseURL + "/auth/login"
	default:
		requestBaseURL = normalized + "/api"
		loginURL = requestBaseURL + "/auth/login"
	}

	return requestBaseURL, loginURL
}

func normalizeNuvioAnalyticsURL(rawValue string) (string, error) {
	normalized := strings.TrimSpace(rawValue)
	if normalized == "" {
		return "", fmt.Errorf("missing url")
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("invalid url scheme")
	}

	if strings.TrimSpace(parsed.Host) == "" {
		return "", fmt.Errorf("missing url host")
	}

	parsed.Fragment = ""
	return strings.TrimRight(parsed.String(), "/"), nil
}

func parseNuvioUmamiAnyInt(value any) int {
	floatValue, ok := parseNuvioUmamiAnyFloat(value)
	if !ok {
		return 0
	}

	if floatValue < 0 {
		return 0
	}

	return int(floatValue)
}

func parseNuvioUmamiAnyFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case json.Number:
		if parsed, err := typed.Float64(); err == nil {
			return parsed, true
		}
	case string:
		if parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64); err == nil {
			return parsed, true
		}
	}

	return 0, false
}

func parseNuvioUmamiAnyString(value any, fallback string) string {
	normalized := strings.TrimSpace(parseStringValue(value))
	if normalized == "" {
		return fallback
	}

	return normalized
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

func newNuvioReportsTrafficStateError(state string, message string, cause error) error {
	return &nuvioReportsTrafficStateError{
		State:   strings.TrimSpace(state),
		Message: strings.TrimSpace(message),
		Cause:   cause,
	}
}

func unwrapNuvioReportsTrafficStateError(err error) (*nuvioReportsTrafficStateError, bool) {
	if err == nil {
		return nil, false
	}

	stateErr := &nuvioReportsTrafficStateError{}
	if errors.As(err, &stateErr) {
		if strings.TrimSpace(stateErr.State) == "" {
			stateErr.State = "provider_error"
		}
		if strings.TrimSpace(stateErr.Message) == "" {
			stateErr.Message = "Traffic analytics are temporarily unavailable."
		}
		return stateErr, true
	}

	return nil, false
}

// NUVIO CUSTOM END: Reports traffic endpoint.
