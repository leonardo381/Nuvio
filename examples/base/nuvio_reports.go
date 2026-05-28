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
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioReportsTrafficPeriodThisMonth = "thisMonth"
	nuvioReportsTrafficPeriodLastMonth = "lastMonth"
	nuvioReportsTrafficPeriodLast30d   = "last30Days"
	nuvioReportsTrafficPeriodAllTime   = "allTime"
	nuvioReportsDashboardMaxScan       = 5000
)

var (
	nuvioReportsConversionEventNames = []string{
		"contact_form_submitted",
		"whatsapp_click",
		"booking_submitted",
		"newsletter_signup",
		"phone_click",
		"email_click",
		"directions_click",
	}
	nuvioReportsScrollDepthEventName  = "scroll_depth_reached"
	nuvioReportsScrollDepthThresholds = []int{25, 50, 75, 90}
)

var nuvioReportsHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

type nuvioWebsiteReportsAnalyticsConfig struct {
	FeatureAvailable  bool
	Provider          string
	Enabled           bool
	SiteID            string
	ScriptEnabled     bool
	ScriptURL         string
	EventsScrollDepth bool
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

type nuvioReportsTrafficConversionTotals struct {
	AllEvents        int `json:"allEvents"`
	UniqueEventTypes int `json:"uniqueEventTypes"`
}

type nuvioReportsTrafficConversionByType struct {
	Event string `json:"event"`
	Count int    `json:"count"`
}

type nuvioReportsTrafficConversionByPage struct {
	PageSlug string `json:"pageSlug"`
	Count    int    `json:"count"`
}

type nuvioReportsTrafficConversionBySourceBlock struct {
	SourceBlock string `json:"sourceBlock"`
	Count       int    `json:"count"`
}

type nuvioReportsTrafficConversionByCtaType struct {
	CtaType string `json:"ctaType"`
	Count   int    `json:"count"`
}

type nuvioReportsTrafficConversions struct {
	State         string                                       `json:"state"`
	Message       string                                       `json:"message,omitempty"`
	Totals        nuvioReportsTrafficConversionTotals          `json:"totals"`
	ByType        []nuvioReportsTrafficConversionByType        `json:"byType"`
	ByPage        []nuvioReportsTrafficConversionByPage        `json:"byPage"`
	BySourceBlock []nuvioReportsTrafficConversionBySourceBlock `json:"bySourceBlock"`
	ByCtaType     []nuvioReportsTrafficConversionByCtaType     `json:"byCtaType"`
}

type nuvioReportsTrafficScrollDepthThreshold struct {
	Depth int `json:"depth"`
	Count int `json:"count"`
}

type nuvioReportsTrafficScrollDepthByPage struct {
	PageSlug string `json:"pageSlug"`
	Count    int    `json:"count"`
}

type nuvioReportsTrafficScrollDepth struct {
	State      string                                    `json:"state"`
	Message    string                                    `json:"message,omitempty"`
	Thresholds []nuvioReportsTrafficScrollDepthThreshold `json:"thresholds"`
	ByPage     []nuvioReportsTrafficScrollDepthByPage    `json:"byPage"`
}

type nuvioReportsTrafficEngagement struct {
	ScrollDepth nuvioReportsTrafficScrollDepth `json:"scrollDepth"`
}

type nuvioReportsTrafficInsight struct {
	ID             string   `json:"id"`
	Severity       string   `json:"severity"`
	Area           string   `json:"area"`
	Title          string   `json:"title"`
	Evidence       []string `json:"evidence"`
	Recommendation string   `json:"recommendation"`
	Confidence     string   `json:"confidence"`
	TargetRoute    string   `json:"targetRoute,omitempty"`
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
	Conversions      nuvioReportsTrafficConversions       `json:"conversions"`
	Engagement       nuvioReportsTrafficEngagement        `json:"engagement"`
	Insights         []nuvioReportsTrafficInsight         `json:"insights"`
	FetchedAt        string                               `json:"fetchedAt,omitempty"`
}

type nuvioReportsDashboardFeatureFlags struct {
	Reports bool `json:"reports"`
}

type nuvioReportsDashboardWebsiteSEODefaults struct {
	SEOTitle       string `json:"seoTitle"`
	SEODescription string `json:"seoDescription"`
	SEOSocialImage bool   `json:"seoSocialImage"`
}

type nuvioReportsDashboardContactDTO struct {
	ID      string `json:"id"`
	Website string `json:"website"`
	Channel string `json:"channel"`
	Status  string `json:"status"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Subject string `json:"subject"`
	Message string `json:"message"`
	Created string `json:"created"`
}

type nuvioReportsDashboardWhatsappDTO struct {
	ID             string `json:"id"`
	Website        string `json:"website"`
	Status         string `json:"status"`
	Source         string `json:"source"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Message        string `json:"message"`
	DefaultMessage string `json:"defaultMessage"`
	Created        string `json:"created"`
}

type nuvioReportsDashboardAppointmentDTO struct {
	ID      string `json:"id"`
	Website string `json:"website"`
	Service string `json:"service"`
	Status  string `json:"status"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Date    string `json:"date"`
	Time    string `json:"time"`
	Created string `json:"created"`
}

type nuvioReportsDashboardBookingServiceDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type nuvioReportsDashboardSubscriberDTO struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Status  string `json:"status"`
	Created string `json:"created"`
}

type nuvioReportsDashboardCampaignDTO struct {
	ID              string `json:"id"`
	Subject         string `json:"subject"`
	Status          string `json:"status"`
	RecipientsCount any    `json:"recipientsCount"`
	SentAt          string `json:"sentAt"`
	Updated         string `json:"updated"`
	Created         string `json:"created"`
}

type nuvioReportsDashboardPageDTO struct {
	ID             string `json:"id"`
	Website        string `json:"website"`
	Title          string `json:"title"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	Path           string `json:"path"`
	URL            string `json:"url"`
	SEOTitle       string `json:"seo_title"`
	SEODescription string `json:"seo_description"`
	SEOSocialImage any    `json:"seo_social_image"`
	SEONoindex     any    `json:"seo_noindex"`
	Updated        string `json:"updated"`
	Created        string `json:"created"`
}

type nuvioReportsDashboardDatasets struct {
	Contacts        []nuvioReportsDashboardContactDTO        `json:"contacts"`
	Whatsapp        []nuvioReportsDashboardWhatsappDTO       `json:"whatsapp"`
	Appointments    []nuvioReportsDashboardAppointmentDTO    `json:"appointments"`
	BookingServices []nuvioReportsDashboardBookingServiceDTO `json:"bookingServices"`
	Subscribers     []nuvioReportsDashboardSubscriberDTO     `json:"subscribers"`
	Campaigns       []nuvioReportsDashboardCampaignDTO       `json:"campaigns"`
	Pages           []nuvioReportsDashboardPageDTO           `json:"pages"`
}

type nuvioReportsDashboardResponse struct {
	State              string                                  `json:"state"`
	WebsiteID          string                                  `json:"websiteId"`
	Period             nuvioReportsTrafficPeriod               `json:"period"`
	FeatureFlags       nuvioReportsDashboardFeatureFlags       `json:"featureFlags"`
	WebsiteSEODefaults nuvioReportsDashboardWebsiteSEODefaults `json:"websiteSeoDefaults"`
	Datasets           nuvioReportsDashboardDatasets           `json:"datasets"`
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
	Conversions      nuvioReportsTrafficConversions
	Engagement       nuvioReportsTrafficEngagement
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

type nuvioUmamiEventsStatsData struct {
	Events       any `json:"events"`
	UniqueEvents any `json:"uniqueEvents"`
}

type nuvioUmamiEventsStatsResponse struct {
	Data nuvioUmamiEventsStatsData `json:"data"`
}

type nuvioUmamiEventsSeriesRow struct {
	X any `json:"x"`
	Y any `json:"y"`
}

type nuvioUmamiEventDataValueRow struct {
	Value any `json:"value"`
	Total any `json:"total"`
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

	reportsGroup.GET("/dashboard", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}
		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		periodQuery, err := resolveNuvioReportsTrafficPeriod(strings.TrimSpace(e.Request.URL.Query().Get("period")), time.Now().UTC())
		if err != nil {
			return e.BadRequestError("Invalid period. Use thisMonth, lastMonth, last30Days, or allTime.", nil)
		}

		website, analyticsConfig, err := loadNuvioWebsiteReportsAnalyticsConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			return e.BadRequestError("Failed to load Reports dashboard settings.", nil)
		}

		datasets, err := loadNuvioReportsDashboardDatasets(e.App, websiteID)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO reports dashboard data load failed",
				"websiteId",
				websiteID,
				"period",
				periodQuery.Period.Key,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load Reports dashboard data.", nil)
		}

		response := nuvioReportsDashboardResponse{
			State:     "ok",
			WebsiteID: websiteID,
			Period:    periodQuery.Period,
			FeatureFlags: nuvioReportsDashboardFeatureFlags{
				Reports: analyticsConfig.FeatureAvailable,
			},
			WebsiteSEODefaults: buildNuvioReportsDashboardWebsiteSEODefaults(website),
			Datasets:           datasets,
		}

		return e.JSON(http.StatusOK, response)
	})

	reportsGroup.GET("/traffic", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}
		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
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

		umamiConfig, err := loadNuvioUmamiConfig()
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
			analyticsConfig.EventsScrollDepth,
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

		response := nuvioReportsTrafficResponse{
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
			Conversions:      trafficData.Conversions,
			Engagement:       trafficData.Engagement,
			Insights:         []nuvioReportsTrafficInsight{},
			FetchedAt:        time.Now().UTC().Format(time.RFC3339),
		}
		response.Insights = buildNuvioReportsTrafficInsights(
			response.State,
			response.Message,
			response.Summary,
			response.TopPages,
			response.Sources,
			response.Conversions,
			response.Engagement,
			true,
		)

		return e.JSON(http.StatusOK, response)
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
		FeatureAvailable:  true,
		Provider:          "",
		Enabled:           false,
		SiteID:            "",
		ScriptEnabled:     false,
		ScriptURL:         "",
		EventsScrollDepth: false,
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
			if eventsSettings, ok := toStringAnyMap(analyticsSettings["events"]); ok {
				if value, ok := parseBoolValue(eventsSettings["scrollDepth"]); ok {
					config.EventsScrollDepth = value
				}
			}
		}
	}

	return website, config, nil
}

func buildNuvioReportsDashboardWebsiteSEODefaults(website *core.Record) nuvioReportsDashboardWebsiteSEODefaults {
	if website == nil {
		return nuvioReportsDashboardWebsiteSEODefaults{}
	}

	seoTitle := strings.TrimSpace(website.GetString("seoTitle"))
	if seoTitle == "" {
		seoTitle = strings.TrimSpace(website.GetString("seo_title"))
	}

	seoDescription := strings.TrimSpace(website.GetString("seoDescription"))
	if seoDescription == "" {
		seoDescription = strings.TrimSpace(website.GetString("seo_description"))
	}

	seoImageValue := website.Get("seoImage")
	if seoImageValue == nil {
		seoImageValue = website.Get("seo_image")
	}

	return nuvioReportsDashboardWebsiteSEODefaults{
		SEOTitle:       seoTitle,
		SEODescription: seoDescription,
		SEOSocialImage: hasNuvioReportsDashboardFileValue(seoImageValue),
	}
}

func hasNuvioReportsDashboardFileValue(value any) bool {
	switch typed := value.(type) {
	case []string:
		for _, item := range typed {
			if strings.TrimSpace(item) != "" {
				return true
			}
		}
	case []any:
		for _, item := range typed {
			if hasNuvioReportsDashboardFileValue(item) {
				return true
			}
		}
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return false
		}

		if strings.HasPrefix(trimmed, "[") {
			var parsed []any
			if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
				return hasNuvioReportsDashboardFileValue(parsed)
			}
		}

		return true
	default:
		return value != nil
	}

	return false
}

func loadNuvioReportsDashboardDatasets(app core.App, websiteID string) (nuvioReportsDashboardDatasets, error) {
	datasets := nuvioReportsDashboardDatasets{
		Contacts:        make([]nuvioReportsDashboardContactDTO, 0),
		Whatsapp:        make([]nuvioReportsDashboardWhatsappDTO, 0),
		Appointments:    make([]nuvioReportsDashboardAppointmentDTO, 0),
		BookingServices: make([]nuvioReportsDashboardBookingServiceDTO, 0),
		Subscribers:     make([]nuvioReportsDashboardSubscriberDTO, 0),
		Campaigns:       make([]nuvioReportsDashboardCampaignDTO, 0),
		Pages:           make([]nuvioReportsDashboardPageDTO, 0),
	}

	contacts, err := findNuvioReportsDashboardRecordsByWebsite(app, nuvioContactsCollectionID, websiteID, "-created")
	if err != nil {
		return datasets, err
	}
	for _, record := range contacts {
		datasets.Contacts = append(datasets.Contacts, nuvioReportsDashboardContactDTO{
			ID:      strings.TrimSpace(record.Id),
			Website: resolveNuvioPublicRelationID(record, "website"),
			Channel: strings.TrimSpace(record.GetString("channel")),
			Status:  strings.TrimSpace(record.GetString("status")),
			Name:    strings.TrimSpace(record.GetString("name")),
			Email:   strings.TrimSpace(record.GetString("email")),
			Phone:   strings.TrimSpace(record.GetString("phone")),
			Subject: strings.TrimSpace(record.GetString("subject")),
			Message: strings.TrimSpace(record.GetString("message")),
			Created: strings.TrimSpace(record.GetString("created")),
		})
	}

	whatsapp, err := findNuvioReportsDashboardRecordsByWebsite(app, nuvioWhatsappCollectionID, websiteID, "-created")
	if err != nil {
		return datasets, err
	}
	for _, record := range whatsapp {
		datasets.Whatsapp = append(datasets.Whatsapp, nuvioReportsDashboardWhatsappDTO{
			ID:             strings.TrimSpace(record.Id),
			Website:        resolveNuvioPublicRelationID(record, "website"),
			Status:         strings.TrimSpace(record.GetString("status")),
			Source:         strings.TrimSpace(record.GetString("source")),
			Name:           strings.TrimSpace(record.GetString("name")),
			Email:          strings.TrimSpace(record.GetString("email")),
			Phone:          strings.TrimSpace(record.GetString("phone")),
			Message:        strings.TrimSpace(record.GetString("message")),
			DefaultMessage: strings.TrimSpace(record.GetString("defaultMessage")),
			Created:        strings.TrimSpace(record.GetString("created")),
		})
	}

	appointments, err := findNuvioReportsDashboardRecordsByWebsite(app, nuvioAppointmentsCollectionID, websiteID, "-created")
	if err != nil {
		return datasets, err
	}
	for _, record := range appointments {
		datasets.Appointments = append(datasets.Appointments, nuvioReportsDashboardAppointmentDTO{
			ID:      strings.TrimSpace(record.Id),
			Website: resolveNuvioPublicRelationID(record, "website"),
			Service: resolveNuvioPublicRelationID(record, "service"),
			Status:  strings.TrimSpace(record.GetString("status")),
			Name:    strings.TrimSpace(record.GetString("name")),
			Email:   strings.TrimSpace(record.GetString("email")),
			Phone:   strings.TrimSpace(record.GetString("phone")),
			Date:    strings.TrimSpace(record.GetString("date")),
			Time:    strings.TrimSpace(record.GetString("time")),
			Created: strings.TrimSpace(record.GetString("created")),
		})
	}

	bookingServices, err := findNuvioReportsDashboardRecordsByWebsite(app, nuvioBookingServicesCollectionID, websiteID, "+name")
	if err != nil {
		return datasets, err
	}
	for _, record := range bookingServices {
		datasets.BookingServices = append(datasets.BookingServices, nuvioReportsDashboardBookingServiceDTO{
			ID:   strings.TrimSpace(record.Id),
			Name: strings.TrimSpace(record.GetString("name")),
		})
	}

	subscribers, err := findNuvioReportsDashboardRecordsByWebsite(app, nuvioSubscribersCollectionID, websiteID, "-created")
	if err != nil {
		return datasets, err
	}
	for _, record := range subscribers {
		datasets.Subscribers = append(datasets.Subscribers, nuvioReportsDashboardSubscriberDTO{
			ID:      strings.TrimSpace(record.Id),
			Email:   strings.TrimSpace(record.GetString("email")),
			Status:  strings.TrimSpace(record.GetString("status")),
			Created: strings.TrimSpace(record.GetString("created")),
		})
	}

	campaigns, err := findNuvioReportsDashboardRecordsByWebsite(app, nuvioCampaignsCollectionID, websiteID, "-updated")
	if err != nil {
		return datasets, err
	}
	for _, record := range campaigns {
		datasets.Campaigns = append(datasets.Campaigns, nuvioReportsDashboardCampaignDTO{
			ID:              strings.TrimSpace(record.Id),
			Subject:         strings.TrimSpace(record.GetString("subject")),
			Status:          strings.TrimSpace(record.GetString("status")),
			RecipientsCount: record.Get("recipientsCount"),
			SentAt:          strings.TrimSpace(record.GetString("sentAt")),
			Updated:         strings.TrimSpace(record.GetString("updated")),
			Created:         strings.TrimSpace(record.GetString("created")),
		})
	}

	pages, err := findNuvioReportsDashboardRecordsByWebsite(app, nuvioPagesCollectionID, websiteID, "-updated")
	if err != nil {
		return datasets, err
	}
	for _, record := range pages {
		seoSocialImage := record.Get("seo_social_image")
		if seoSocialImage == nil {
			seoSocialImage = record.Get("seoSocialImage")
		}

		seoNoindex := record.Get("seo_noindex")
		if seoNoindex == nil {
			seoNoindex = record.Get("seoNoindex")
		}

		seoTitle := strings.TrimSpace(record.GetString("seo_title"))
		if seoTitle == "" {
			seoTitle = strings.TrimSpace(record.GetString("seoTitle"))
		}

		seoDescription := strings.TrimSpace(record.GetString("seo_description"))
		if seoDescription == "" {
			seoDescription = strings.TrimSpace(record.GetString("seoDescription"))
		}

		datasets.Pages = append(datasets.Pages, nuvioReportsDashboardPageDTO{
			ID:             strings.TrimSpace(record.Id),
			Website:        resolveNuvioPublicRelationID(record, "website"),
			Title:          strings.TrimSpace(record.GetString("title")),
			Name:           strings.TrimSpace(record.GetString("name")),
			Slug:           strings.TrimSpace(record.GetString("slug")),
			Path:           strings.TrimSpace(record.GetString("path")),
			URL:            strings.TrimSpace(record.GetString("url")),
			SEOTitle:       seoTitle,
			SEODescription: seoDescription,
			SEOSocialImage: seoSocialImage,
			SEONoindex:     seoNoindex,
			Updated:        strings.TrimSpace(record.GetString("updated")),
			Created:        strings.TrimSpace(record.GetString("created")),
		})
	}

	return datasets, nil
}

func findNuvioReportsDashboardRecordsByWebsite(
	app core.App,
	collectionID string,
	websiteID string,
	sortExpr string,
) ([]*core.Record, error) {
	collection, err := app.FindCachedCollectionByNameOrId(collectionID)
	if err != nil {
		return nil, err
	}

	filter := "website={:websiteId}"
	params := dbx.Params{
		"websiteId": websiteID,
	}

	records, err := app.FindRecordsByFilter(
		collection,
		filter,
		sortExpr,
		nuvioReportsDashboardMaxScan,
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
			nuvioReportsDashboardMaxScan,
			0,
			params,
		)
	}

	return nil, err
}

func resolveTrustedNuvioUmamiAPIURL() (string, error) {
	rawAPIURL := strings.TrimSpace(os.Getenv("NUVIO_UMAMI_API_URL"))

	if rawAPIURL == "" {
		return "", newNuvioReportsTrafficStateError(
			"provider_unconfigured",
			"Traffic analytics provider is not configured right now.",
			nil,
		)
	}

	normalizedAPIURL, err := normalizeNuvioAnalyticsURL(rawAPIURL)
	if err != nil {
		return "", newNuvioReportsTrafficStateError(
			"provider_unconfigured",
			"Traffic analytics provider is misconfigured right now.",
			err,
		)
	}

	return normalizedAPIURL, nil
}

func loadNuvioUmamiConfig() (nuvioUmamiConfig, error) {
	normalizedAPIURL, err := resolveTrustedNuvioUmamiAPIURL()
	if err != nil {
		return nuvioUmamiConfig{}, err
	}

	requestBaseURL, loginURL := resolveNuvioUmamiURLs(normalizedAPIURL)
	if requestBaseURL == "" {
		return nuvioUmamiConfig{}, newNuvioReportsTrafficStateError(
			"provider_unconfigured",
			"Traffic analytics provider is misconfigured right now.",
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
	response := nuvioReportsTrafficResponse{
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
		Conversions:      buildNuvioReportsDefaultConversions("unavailable", "Conversion event metrics are unavailable."),
		Engagement: nuvioReportsTrafficEngagement{
			ScrollDepth: buildNuvioReportsDefaultScrollDepth("unavailable", "Scroll depth metrics are unavailable."),
		},
		Insights: []nuvioReportsTrafficInsight{},
	}
	response.Insights = buildNuvioReportsTrafficInsights(
		response.State,
		response.Message,
		nil,
		nil,
		nil,
		response.Conversions,
		response.Engagement,
		false,
	)

	return response
}

func fetchNuvioUmamiTrafficData(
	ctx context.Context,
	config nuvioUmamiConfig,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
	scrollDepthEnabled bool,
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

	conversions, engagement, conversionIssueCount := fetchNuvioUmamiConversionAndEngagementData(
		ctx,
		config,
		authHeaders,
		siteID,
		periodQuery,
		scrollDepthEnabled,
	)
	partialIssueCount += conversionIssueCount

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
		Conversions:      conversions,
		Engagement:       engagement,
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

func buildNuvioReportsDefaultConversions(state string, message string) nuvioReportsTrafficConversions {
	byType := make([]nuvioReportsTrafficConversionByType, 0, len(nuvioReportsConversionEventNames))
	for _, eventName := range nuvioReportsConversionEventNames {
		byType = append(byType, nuvioReportsTrafficConversionByType{
			Event: eventName,
			Count: 0,
		})
	}

	return nuvioReportsTrafficConversions{
		State:   strings.TrimSpace(state),
		Message: strings.TrimSpace(message),
		Totals: nuvioReportsTrafficConversionTotals{
			AllEvents:        0,
			UniqueEventTypes: 0,
		},
		ByType:        byType,
		ByPage:        []nuvioReportsTrafficConversionByPage{},
		BySourceBlock: []nuvioReportsTrafficConversionBySourceBlock{},
		ByCtaType:     []nuvioReportsTrafficConversionByCtaType{},
	}
}

func buildNuvioReportsDefaultScrollDepth(state string, message string) nuvioReportsTrafficScrollDepth {
	thresholds := make([]nuvioReportsTrafficScrollDepthThreshold, 0, len(nuvioReportsScrollDepthThresholds))
	for _, depth := range nuvioReportsScrollDepthThresholds {
		thresholds = append(thresholds, nuvioReportsTrafficScrollDepthThreshold{
			Depth: depth,
			Count: 0,
		})
	}

	return nuvioReportsTrafficScrollDepth{
		State:      strings.TrimSpace(state),
		Message:    strings.TrimSpace(message),
		Thresholds: thresholds,
		ByPage:     []nuvioReportsTrafficScrollDepthByPage{},
	}
}

func fetchNuvioUmamiConversionAndEngagementData(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
	scrollDepthEnabled bool,
) (nuvioReportsTrafficConversions, nuvioReportsTrafficEngagement, int) {
	conversions := buildNuvioReportsDefaultConversions(
		"unavailable",
		"Conversion event metrics are unavailable.",
	)
	scrollDepthState := "disabled"
	scrollDepthMessage := "Scroll depth tracking is disabled."
	if scrollDepthEnabled {
		scrollDepthState = "unavailable"
		scrollDepthMessage = "Scroll depth metrics are unavailable."
	}
	engagement := nuvioReportsTrafficEngagement{
		ScrollDepth: buildNuvioReportsDefaultScrollDepth(scrollDepthState, scrollDepthMessage),
	}

	issueCount := 0

	_, statsErr := fetchNuvioUmamiEventsStats(ctx, config, authHeaders, siteID, periodQuery)
	if statsErr != nil {
		issueCount++
	}

	seriesRows, seriesErr := fetchNuvioUmamiEventsSeries(ctx, config, authHeaders, siteID, periodQuery)
	if seriesErr != nil {
		issueCount++
	}

	if statsErr == nil || seriesErr == nil {
		eventCountByName := map[string]int{}
		for _, row := range seriesRows {
			eventName := parseNuvioUmamiAnyString(row.X, "")
			if eventName == "" {
				continue
			}
			eventCountByName[eventName] += parseNuvioUmamiAnyInt(row.Y)
		}

		for index := range conversions.ByType {
			eventName := conversions.ByType[index].Event
			conversions.ByType[index].Count = eventCountByName[eventName]
		}
		totalEvents := 0
		uniqueEventTypes := 0
		for _, item := range conversions.ByType {
			if item.Count < 1 {
				continue
			}
			totalEvents += item.Count
			uniqueEventTypes++
		}
		conversions.Totals.AllEvents = totalEvents
		conversions.Totals.UniqueEventTypes = uniqueEventTypes

		conversions.State = "ok"
		conversions.Message = "Conversion event metrics are available."
		if statsErr != nil || seriesErr != nil {
			conversions.State = "partial"
			conversions.Message = "Some analytics event breakdowns are unavailable."
		}
	}

	conversionPropertiesIssueCount := 0
	conversionByPageMap := map[string]int{}
	conversionBySourceBlockMap := map[string]int{}
	conversionByCtaTypeMap := map[string]int{}

	for _, eventName := range nuvioReportsConversionEventNames {
		pageRows, err := fetchNuvioUmamiEventDataValues(
			ctx,
			config,
			authHeaders,
			siteID,
			periodQuery,
			eventName,
			"pageSlug",
		)
		if err != nil {
			conversionPropertiesIssueCount++
		} else {
			accumulateNuvioReportsValueRows(conversionByPageMap, pageRows)
		}

		sourceRows, err := fetchNuvioUmamiEventDataValues(
			ctx,
			config,
			authHeaders,
			siteID,
			periodQuery,
			eventName,
			"sourceBlock",
		)
		if err != nil {
			conversionPropertiesIssueCount++
		} else {
			accumulateNuvioReportsValueRows(conversionBySourceBlockMap, sourceRows)
		}

		ctaRows, err := fetchNuvioUmamiEventDataValues(
			ctx,
			config,
			authHeaders,
			siteID,
			periodQuery,
			eventName,
			"ctaType",
		)
		if err != nil {
			conversionPropertiesIssueCount++
		} else {
			accumulateNuvioReportsValueRows(conversionByCtaTypeMap, ctaRows)
		}
	}

	if conversionPropertiesIssueCount > 0 {
		issueCount += conversionPropertiesIssueCount
		if conversions.State == "ok" {
			conversions.State = "partial"
			conversions.Message = "Some analytics event breakdowns are unavailable."
		}
	}

	conversions.ByPage = convertNuvioReportsCountMapToByPageRows(conversionByPageMap)
	conversions.BySourceBlock = convertNuvioReportsCountMapToBySourceBlockRows(conversionBySourceBlockMap)
	conversions.ByCtaType = convertNuvioReportsCountMapToByCtaTypeRows(conversionByCtaTypeMap)
	if conversions.State == "unavailable" && (len(conversions.ByPage) > 0 ||
		len(conversions.BySourceBlock) > 0 ||
		len(conversions.ByCtaType) > 0) {
		conversions.State = "partial"
		conversions.Message = "Some analytics event breakdowns are unavailable."
	}

	if !scrollDepthEnabled {
		return conversions, engagement, issueCount
	}

	scrollDepthPropertyIssues := 0
	scrollDepthRows, err := fetchNuvioUmamiEventDataValues(
		ctx,
		config,
		authHeaders,
		siteID,
		periodQuery,
		nuvioReportsScrollDepthEventName,
		"depth",
	)
	if err != nil {
		scrollDepthPropertyIssues++
	} else {
		thresholdCounts := map[int]int{}
		for _, row := range scrollDepthRows {
			depthValue := parseNuvioReportsPropertyDepth(row.Value)
			if depthValue < 1 {
				continue
			}
			thresholdCounts[depthValue] += parseNuvioUmamiAnyInt(row.Total)
		}
		for index := range engagement.ScrollDepth.Thresholds {
			depth := engagement.ScrollDepth.Thresholds[index].Depth
			engagement.ScrollDepth.Thresholds[index].Count = thresholdCounts[depth]
		}
	}

	scrollDepthPageRows, err := fetchNuvioUmamiEventDataValues(
		ctx,
		config,
		authHeaders,
		siteID,
		periodQuery,
		nuvioReportsScrollDepthEventName,
		"pageSlug",
	)
	if err != nil {
		scrollDepthPropertyIssues++
	} else {
		byPageCounts := map[string]int{}
		accumulateNuvioReportsValueRows(byPageCounts, scrollDepthPageRows)
		engagement.ScrollDepth.ByPage = convertNuvioReportsCountMapToScrollDepthByPageRows(byPageCounts)
	}

	if scrollDepthPropertyIssues > 0 {
		issueCount += scrollDepthPropertyIssues
		engagement.ScrollDepth.State = "unavailable"
		engagement.ScrollDepth.Message = "Scroll depth metrics are unavailable."
		return conversions, engagement, issueCount
	}

	engagement.ScrollDepth.State = "ok"
	engagement.ScrollDepth.Message = "Scroll depth events are available."
	return conversions, engagement, issueCount
}

func fetchNuvioUmamiEventsStats(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) (nuvioUmamiEventsStatsResponse, error) {
	query := buildNuvioUmamiBaseQuery(periodQuery)
	endpointPath := buildNuvioUmamiWebsiteEndpointPath(siteID, "events/stats")
	requestURL := buildNuvioUmamiRequestURL(config.RequestBaseURL, endpointPath, query)

	payload := nuvioUmamiEventsStatsResponse{}
	if err := executeNuvioUmamiJSONRequest(ctx, http.MethodGet, requestURL, authHeaders, nil, &payload); err != nil {
		return nuvioUmamiEventsStatsResponse{}, err
	}

	return payload, nil
}

func fetchNuvioUmamiEventsSeries(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
) ([]nuvioUmamiEventsSeriesRow, error) {
	query := buildNuvioUmamiBaseQuery(periodQuery)
	if periodQuery.Unit != "" {
		query.Set("unit", periodQuery.Unit)
	}
	if periodQuery.Timezone != "" {
		query.Set("timezone", periodQuery.Timezone)
	}

	endpointPath := buildNuvioUmamiWebsiteEndpointPath(siteID, "events/series")
	requestURL := buildNuvioUmamiRequestURL(config.RequestBaseURL, endpointPath, query)

	rawPayload := any(nil)
	if err := executeNuvioUmamiJSONRequest(ctx, http.MethodGet, requestURL, authHeaders, nil, &rawPayload); err != nil {
		return nil, err
	}

	return parseNuvioUmamiEventsSeriesPayload(rawPayload)
}

func fetchNuvioUmamiEventDataValues(
	ctx context.Context,
	config nuvioUmamiConfig,
	authHeaders map[string]string,
	siteID string,
	periodQuery nuvioReportsTrafficPeriodQuery,
	eventName string,
	propertyName string,
) ([]nuvioUmamiEventDataValueRow, error) {
	query := buildNuvioUmamiBaseQuery(periodQuery)
	normalizedEventName := strings.TrimSpace(eventName)
	normalizedPropertyName := strings.TrimSpace(propertyName)
	if normalizedEventName != "" {
		query.Set("event", normalizedEventName)
		// Legacy compatibility for older Umami versions.
		query.Set("eventName", normalizedEventName)
	}
	if normalizedPropertyName != "" {
		query.Set("propertyName", normalizedPropertyName)
	}

	endpointPath := buildNuvioUmamiWebsiteEndpointPath(siteID, "event-data/values")
	requestURL := buildNuvioUmamiRequestURL(config.RequestBaseURL, endpointPath, query)

	rawPayload := any(nil)
	if err := executeNuvioUmamiJSONRequest(ctx, http.MethodGet, requestURL, authHeaders, nil, &rawPayload); err != nil {
		return nil, err
	}

	return parseNuvioUmamiEventDataValuesPayload(rawPayload)
}

func parseNuvioReportsPropertyDepth(value any) int {
	parsed := parseNuvioUmamiAnyInt(value)
	if parsed >= 1 && parsed <= 100 {
		return parsed
	}
	return 0
}

func parseNuvioUmamiEventsSeriesPayload(raw any) ([]nuvioUmamiEventsSeriesRow, error) {
	switch typed := raw.(type) {
	case []any:
		rows := make([]nuvioUmamiEventsSeriesRow, 0, len(typed))
		for _, item := range typed {
			rowMap, ok := toStringAnyMap(item)
			if !ok {
				return nil, newNuvioUmamiUnexpectedEventShapeError("events/series")
			}
			rows = append(rows, nuvioUmamiEventsSeriesRow{
				X: rowMap["x"],
				Y: rowMap["y"],
			})
		}
		return rows, nil
	case map[string]any:
		nested, exists := typed["data"]
		if !exists {
			return nil, newNuvioUmamiUnexpectedEventShapeError("events/series")
		}
		return parseNuvioUmamiEventsSeriesPayload(nested)
	default:
		return nil, newNuvioUmamiUnexpectedEventShapeError("events/series")
	}
}

func parseNuvioUmamiEventDataValuesPayload(raw any) ([]nuvioUmamiEventDataValueRow, error) {
	switch typed := raw.(type) {
	case []any:
		rows := make([]nuvioUmamiEventDataValueRow, 0, len(typed))
		for _, item := range typed {
			rowMap, ok := toStringAnyMap(item)
			if !ok {
				return nil, newNuvioUmamiUnexpectedEventShapeError("event-data/values")
			}
			rows = append(rows, nuvioUmamiEventDataValueRow{
				Value: rowMap["value"],
				Total: rowMap["total"],
			})
		}
		return rows, nil
	case map[string]any:
		nested, exists := typed["data"]
		if !exists {
			return nil, newNuvioUmamiUnexpectedEventShapeError("event-data/values")
		}
		return parseNuvioUmamiEventDataValuesPayload(nested)
	default:
		return nil, newNuvioUmamiUnexpectedEventShapeError("event-data/values")
	}
}

func newNuvioUmamiUnexpectedEventShapeError(endpoint string) error {
	normalizedEndpoint := strings.Trim(strings.TrimSpace(endpoint), "/")
	if normalizedEndpoint == "" {
		normalizedEndpoint = "events"
	}
	return fmt.Errorf("unexpected_umami_event_response_shape: %s", normalizedEndpoint)
}

func accumulateNuvioReportsValueRows(target map[string]int, rows []nuvioUmamiEventDataValueRow) {
	if len(rows) == 0 {
		return
	}

	for _, row := range rows {
		key := strings.TrimSpace(parseStringValue(row.Value))
		if key == "" {
			continue
		}
		target[key] += parseNuvioUmamiAnyInt(row.Total)
	}
}

func sortedNuvioReportsCountEntries(values map[string]int) []struct {
	Key   string
	Count int
} {
	rows := make([]struct {
		Key   string
		Count int
	}, 0, len(values))

	for key, count := range values {
		normalizedKey := strings.TrimSpace(key)
		if normalizedKey == "" || count < 1 {
			continue
		}
		rows = append(rows, struct {
			Key   string
			Count int
		}{
			Key:   normalizedKey,
			Count: count,
		})
	}

	sort.Slice(rows, func(i int, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].Key < rows[j].Key
	})

	return rows
}

func convertNuvioReportsCountMapToByPageRows(values map[string]int) []nuvioReportsTrafficConversionByPage {
	entries := sortedNuvioReportsCountEntries(values)
	rows := make([]nuvioReportsTrafficConversionByPage, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, nuvioReportsTrafficConversionByPage{
			PageSlug: entry.Key,
			Count:    entry.Count,
		})
	}
	return rows
}

func convertNuvioReportsCountMapToBySourceBlockRows(values map[string]int) []nuvioReportsTrafficConversionBySourceBlock {
	entries := sortedNuvioReportsCountEntries(values)
	rows := make([]nuvioReportsTrafficConversionBySourceBlock, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, nuvioReportsTrafficConversionBySourceBlock{
			SourceBlock: entry.Key,
			Count:       entry.Count,
		})
	}
	return rows
}

func convertNuvioReportsCountMapToByCtaTypeRows(values map[string]int) []nuvioReportsTrafficConversionByCtaType {
	entries := sortedNuvioReportsCountEntries(values)
	rows := make([]nuvioReportsTrafficConversionByCtaType, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, nuvioReportsTrafficConversionByCtaType{
			CtaType: entry.Key,
			Count:   entry.Count,
		})
	}
	return rows
}

func convertNuvioReportsCountMapToScrollDepthByPageRows(values map[string]int) []nuvioReportsTrafficScrollDepthByPage {
	entries := sortedNuvioReportsCountEntries(values)
	rows := make([]nuvioReportsTrafficScrollDepthByPage, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, nuvioReportsTrafficScrollDepthByPage{
			PageSlug: entry.Key,
			Count:    entry.Count,
		})
	}
	return rows
}

func buildNuvioReportsTrafficInsights(
	trafficState string,
	trafficMessage string,
	summary *nuvioReportsTrafficSummary,
	topPages []nuvioReportsTrafficTopPage,
	sources []nuvioReportsTrafficSource,
	conversions nuvioReportsTrafficConversions,
	engagement nuvioReportsTrafficEngagement,
	includeMetricInsights bool,
) []nuvioReportsTrafficInsight {
	insights := make([]nuvioReportsTrafficInsight, 0, 12)
	seen := map[string]struct{}{}

	addInsight := func(insight nuvioReportsTrafficInsight) {
		insight.ID = strings.TrimSpace(insight.ID)
		if insight.ID == "" {
			return
		}
		if _, exists := seen[insight.ID]; exists {
			return
		}
		insight.Severity = normalizeNuvioReportsInsightSeverity(insight.Severity)
		insight.Area = normalizeNuvioReportsInsightArea(insight.Area)
		insight.Confidence = normalizeNuvioReportsInsightConfidence(insight.Confidence)
		insight.Title = strings.TrimSpace(insight.Title)
		insight.Recommendation = strings.TrimSpace(insight.Recommendation)
		if insight.Title == "" || insight.Recommendation == "" {
			return
		}
		cleanEvidence := make([]string, 0, len(insight.Evidence))
		for _, line := range insight.Evidence {
			normalized := strings.TrimSpace(line)
			if normalized == "" {
				continue
			}
			cleanEvidence = append(cleanEvidence, normalized)
		}
		if len(cleanEvidence) == 0 {
			return
		}
		insight.Evidence = cleanEvidence
		insight.TargetRoute = strings.TrimSpace(insight.TargetRoute)

		seen[insight.ID] = struct{}{}
		insights = append(insights, insight)
	}

	normalizedTrafficState := strings.TrimSpace(strings.ToLower(trafficState))
	normalizedTrafficMessage := strings.TrimSpace(trafficMessage)

	switch normalizedTrafficState {
	case "analytics_disabled":
		addInsight(nuvioReportsTrafficInsight{
			ID:             "analytics_not_configured",
			Severity:       "medium",
			Area:           "data",
			Title:          "Traffic analytics are currently disabled.",
			Evidence:       []string{"Analytics is disabled for this website."},
			Recommendation: "Configure and enable Analytics in Website Settings to unlock traffic reporting.",
			Confidence:     "high",
		})
	case "provider_unconfigured", "analytics_not_configured":
		addInsight(nuvioReportsTrafficInsight{
			ID:             "analytics_not_configured",
			Severity:       "high",
			Area:           "data",
			Title:          "Traffic analytics are not configured yet.",
			Evidence:       []string{"Analytics provider settings are missing or incomplete."},
			Recommendation: "Configure Analytics in Website Settings to start collecting traffic data.",
			Confidence:     "high",
		})
	case "provider_auth_missing", "provider_auth_error", "provider_error", "provider_not_found":
		addInsight(nuvioReportsTrafficInsight{
			ID:             "analytics_not_configured",
			Severity:       "high",
			Area:           "data",
			Title:          "Traffic analytics connection needs attention.",
			Evidence:       []string{"Analytics provider credentials or connection are currently unavailable."},
			Recommendation: "Review Analytics credentials and provider settings in Website Settings.",
			Confidence:     "medium",
		})
	case "provider_unsupported":
		addInsight(nuvioReportsTrafficInsight{
			ID:             "analytics_not_configured",
			Severity:       "high",
			Area:           "data",
			Title:          "The selected analytics provider is unsupported.",
			Evidence:       []string{"Reports traffic currently supports Umami."},
			Recommendation: "Switch Analytics provider to Umami in Website Settings.",
			Confidence:     "high",
		})
	}

	if normalizedTrafficState == "ok" && normalizedTrafficMessage != "" {
		addInsight(nuvioReportsTrafficInsight{
			ID:             "traffic_partial",
			Severity:       "low",
			Area:           "data",
			Title:          "Some traffic metrics are temporarily unavailable.",
			Evidence:       []string{normalizedTrafficMessage},
			Recommendation: "If this persists, review your analytics configuration and provider connectivity.",
			Confidence:     "medium",
		})
	}

	if !includeMetricInsights {
		return sortAndLimitNuvioReportsInsights(insights, 10)
	}

	normalizedConversionState := strings.TrimSpace(strings.ToLower(conversions.State))
	conversionStateIsReliable := normalizedConversionState == "ok"
	conversionStateIsPartial := normalizedConversionState == "partial"
	if normalizedConversionState == "partial" || normalizedConversionState == "unavailable" {
		insightTitle := "Conversion event metrics are unavailable."
		if normalizedConversionState == "partial" {
			insightTitle = "Some conversion event breakdowns are unavailable."
		}
		addInsight(nuvioReportsTrafficInsight{
			ID:       "events_unavailable",
			Severity: "medium",
			Area:     "data",
			Title:    insightTitle,
			Evidence: []string{
				firstNonEmptyString(conversions.Message, "Conversion event metrics are unavailable."),
			},
			Recommendation: "Check Umami events configuration and API availability.",
			Confidence:     "medium",
		})
	}

	normalizedScrollDepthState := strings.TrimSpace(strings.ToLower(engagement.ScrollDepth.State))
	switch normalizedScrollDepthState {
	case "disabled":
		addInsight(nuvioReportsTrafficInsight{
			ID:             "scroll_depth_disabled",
			Severity:       "info",
			Area:           "engagement",
			Title:          "Scroll depth tracking is disabled.",
			Evidence:       []string{"Scroll depth events are currently turned off."},
			Recommendation: "Enable scroll depth events if you want engagement depth visibility.",
			Confidence:     "high",
		})
	case "partial", "unavailable":
		addInsight(nuvioReportsTrafficInsight{
			ID:       "scroll_depth_unavailable",
			Severity: "low",
			Area:     "engagement",
			Title:    "Scroll depth metrics are unavailable.",
			Evidence: []string{
				firstNonEmptyString(engagement.ScrollDepth.Message, "Scroll depth event data is not currently available."),
			},
			Recommendation: "Verify scroll depth tracking and Umami event-data availability.",
			Confidence:     "medium",
		})
	}

	visitors := 0
	pageviews := 0
	if summary != nil {
		visitors = summary.Visitors
		pageviews = summary.Pageviews
	}

	if conversionStateIsReliable && (visitors > 0 || pageviews > 0) && conversions.Totals.AllEvents == 0 {
		addInsight(nuvioReportsTrafficInsight{
			ID:       "no_conversion_events",
			Severity: "medium",
			Area:     "conversions",
			Title:    "No tracked conversion actions were recorded during this period.",
			Evidence: []string{
				fmt.Sprintf("Visitors: %d", visitors),
				fmt.Sprintf("Pageviews: %d", pageviews),
				"Tracked conversion events: 0",
			},
			Recommendation: "Review CTA placement and confirm conversion event tracking is enabled on key pages.",
			Confidence:     "medium",
		})
	}

	topConversionEvent, topConversionCount := findNuvioReportsTopConversionEvent(conversions.ByType)
	if topConversionEvent != "" && topConversionCount > 0 {
		if conversionStateIsReliable {
			addInsight(nuvioReportsTrafficInsight{
				ID:       "most_used_conversion_event",
				Severity: "info",
				Area:     "conversions",
				Title:    "The strongest tracked conversion action this period is clear.",
				Evidence: []string{
					fmt.Sprintf("Top conversion event: %s (%d)", topConversionEvent, topConversionCount),
				},
				Recommendation: recommendationForNuvioConversionEvent(topConversionEvent),
				Confidence:     "high",
			})
		} else if conversionStateIsPartial {
			addInsight(nuvioReportsTrafficInsight{
				ID:       "most_used_conversion_event",
				Severity: "info",
				Area:     "conversions",
				Title:    "Tracked actions were recorded.",
				Evidence: []string{
					fmt.Sprintf("Top tracked action: %s (%d).", topConversionEvent, topConversionCount),
					"Some conversion event breakdowns may be incomplete.",
				},
				Recommendation: "Use this as a directional signal and review detailed conversion tracking if numbers look incomplete.",
				Confidence:     "medium",
			})
		}
	}

	deepScrollCount := 0
	for _, item := range engagement.ScrollDepth.Thresholds {
		if item.Depth == 75 || item.Depth == 90 {
			deepScrollCount += item.Count
		}
	}

	if deepScrollCount > 0 {
		addInsight(nuvioReportsTrafficInsight{
			ID:       "scroll_depth_active",
			Severity: "info",
			Area:     "engagement",
			Title:    "Some visitors reached deeper parts of the page.",
			Evidence: []string{
				fmt.Sprintf("75%% and 90%% scroll events: %d", deepScrollCount),
			},
			Recommendation: "Continue testing content structure and CTA placement in deeper sections.",
			Confidence:     "medium",
		})
	}

	if deepScrollCount >= 5 && conversions.Totals.AllEvents == 0 {
		addInsight(nuvioReportsTrafficInsight{
			ID:       "high_scroll_low_conversion",
			Severity: "medium",
			Area:     "engagement",
			Title:    "Visitors may be engaging with content, but tracked actions are low.",
			Evidence: []string{
				fmt.Sprintf("75%% and 90%% scroll events: %d", deepScrollCount),
				"Tracked conversion events: 0",
			},
			Recommendation: "Consider reviewing CTA placement on high-engagement pages.",
			Confidence:     "low",
		})
	}

	if visitors > 0 || pageviews > 0 {
		addInsight(nuvioReportsTrafficInsight{
			ID:       "traffic_available",
			Severity: "info",
			Area:     "traffic",
			Title:    "Traffic activity was recorded this period.",
			Evidence: []string{
				fmt.Sprintf("Visitors: %d", visitors),
				fmt.Sprintf("Pageviews: %d", pageviews),
			},
			Recommendation: "Use this baseline to compare conversion and engagement performance.",
			Confidence:     "high",
		})
	}

	if len(topPages) > 0 {
		top := topPages[0]
		addInsight(nuvioReportsTrafficInsight{
			ID:       "top_page_available",
			Severity: "info",
			Area:     "traffic",
			Title:    "A top-performing page was identified.",
			Evidence: []string{
				fmt.Sprintf("Top page: %s", strings.TrimSpace(top.Page)),
				fmt.Sprintf("Pageviews: %d", top.Pageviews),
			},
			Recommendation: "Review this page's layout and CTA strategy to replicate winning patterns elsewhere.",
			Confidence:     "high",
		})
	}

	if len(sources) > 0 {
		topSource := sources[0]
		addInsight(nuvioReportsTrafficInsight{
			ID:       "sources_available",
			Severity: "info",
			Area:     "traffic",
			Title:    "A leading traffic source was identified.",
			Evidence: []string{
				fmt.Sprintf("Top source: %s", strings.TrimSpace(topSource.Source)),
				fmt.Sprintf("Visitors: %d", topSource.Visitors),
			},
			Recommendation: "Consider strengthening your presence on top-performing traffic sources.",
			Confidence:     "medium",
		})
	}

	return sortAndLimitNuvioReportsInsights(insights, 10)
}

func normalizeNuvioReportsInsightSeverity(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "high":
		return "high"
	case "medium":
		return "medium"
	case "low":
		return "low"
	default:
		return "info"
	}
}

func normalizeNuvioReportsInsightArea(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "data":
		return "data"
	case "conversions":
		return "conversions"
	case "engagement":
		return "engagement"
	default:
		return "traffic"
	}
}

func normalizeNuvioReportsInsightConfidence(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

func sortAndLimitNuvioReportsInsights(items []nuvioReportsTrafficInsight, limit int) []nuvioReportsTrafficInsight {
	if len(items) == 0 {
		return []nuvioReportsTrafficInsight{}
	}

	severityRank := map[string]int{
		"high":   0,
		"medium": 1,
		"low":    2,
		"info":   3,
	}
	areaRank := map[string]int{
		"data":        0,
		"conversions": 1,
		"engagement":  2,
		"traffic":     3,
	}

	sort.Slice(items, func(i int, j int) bool {
		leftSeverity := severityRank[items[i].Severity]
		rightSeverity := severityRank[items[j].Severity]
		if leftSeverity != rightSeverity {
			return leftSeverity < rightSeverity
		}
		leftArea := areaRank[items[i].Area]
		rightArea := areaRank[items[j].Area]
		if leftArea != rightArea {
			return leftArea < rightArea
		}
		return items[i].ID < items[j].ID
	})

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return items
}

func findNuvioReportsTopConversionEvent(items []nuvioReportsTrafficConversionByType) (string, int) {
	bestEvent := ""
	bestCount := 0
	for _, item := range items {
		eventName := strings.TrimSpace(item.Event)
		if eventName == "" || item.Count < 1 {
			continue
		}
		if item.Count > bestCount || (item.Count == bestCount && (bestEvent == "" || eventName < bestEvent)) {
			bestEvent = eventName
			bestCount = item.Count
		}
	}
	return bestEvent, bestCount
}

func recommendationForNuvioConversionEvent(eventName string) string {
	switch strings.TrimSpace(strings.ToLower(eventName)) {
	case "whatsapp_click":
		return "Keep your WhatsApp CTA visible on high-intent pages."
	case "contact_form_submitted":
		return "Review and follow up with new leads promptly."
	case "booking_submitted":
		return "Review booking requests and confirm availability quickly."
	case "newsletter_signup":
		return "Continue growing your audience with clear newsletter signup prompts."
	case "phone_click":
		return "Ensure phone inquiries are handled quickly during business hours."
	case "email_click":
		return "Ensure email inquiries receive timely responses."
	case "directions_click":
		return "Keep address and location details clear and up to date."
	default:
		return "Review which calls-to-action are driving the strongest response."
	}
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized != "" {
			return normalized
		}
	}
	return ""
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
