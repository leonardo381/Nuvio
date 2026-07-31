package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

var errNuvioPublicWebsiteContextMissing = errors.New("missing_public_website_context")

const (
	nuvioContactsCollectionID  = "pbc_1661203100"
	nuvioWhatsappCollectionID  = "pbc_1661203200"
	nuvioDefaultReplyToMaxSize = 1
	nuvioTemplateSubjectMaxLen = 160
	nuvioTemplateTextMaxLen    = 4000

	nuvioPublicContactNameMaxLen    = 160
	nuvioPublicContactPhoneMaxLen   = 80
	nuvioPublicContactSubjectMaxLen = 200
	nuvioPublicContactMessageMaxLen = 4000
	nuvioPublicContactSourceMaxLen  = 120
	nuvioPublicContactPageMaxLen    = 200

	nuvioPublicWhatsappSourceMaxLen = 120
	nuvioPublicWhatsappPageMaxLen   = 200
)

const (
	nuvioContactSubmitRateLimitDefaultMaxRequests   = 5
	nuvioContactSubmitRateLimitDefaultWindowSeconds = 60
	nuvioContactSubmitRateLimitMaxEnv               = "NUVIO_CONTACT_SUBMIT_RATE_LIMIT_MAX"
	nuvioContactSubmitRateLimitWindowEnv            = "NUVIO_CONTACT_SUBMIT_RATE_LIMIT_WINDOW_SECONDS"
	nuvioContactSubmitRateLimitStoreKey             = "__nuvioContactSubmitRateLimiter__"
	nuvioContactSubmitRateLimitConfigStoreKey       = "__nuvioContactSubmitRateLimitConfig__"
	nuvioContactSubmitRateLimitMessage              = "Too many contact submissions. Please try again later."
)

type nuvioContactSubmitRateLimitConfig struct {
	MaxRequests int
	Window      time.Duration
}

type nuvioContactSubmitRateLimiter struct {
	mu          sync.Mutex
	clients     map[string]nuvioContactSubmitRateLimitWindow
	lastCleanup time.Time
}

type nuvioContactSubmitRateLimitWindow struct {
	remaining int
	resetAt   time.Time
}

// NUVIO CUSTOM START: Explicit public contact submit rate limiter.
func checkNuvioPublicContactSubmitRateLimit(e *core.RequestEvent) error {
	config := resolveNuvioContactSubmitRateLimitConfig(e.App)
	clientKey := buildNuvioContactSubmitRateLimitClientKey(e)
	limiter := getNuvioContactSubmitRateLimiter(e.App)

	allowed, retryAfter := limiter.Allow(clientKey, config, time.Now().UTC())
	if allowed {
		return nil
	}

	e.Response.Header().Set("Retry-After", strconv.Itoa(nuvioContactSubmitRetryAfterSeconds(retryAfter)))
	return e.TooManyRequestsError(nuvioContactSubmitRateLimitMessage, nil)
}

func getNuvioContactSubmitRateLimiter(app core.App) *nuvioContactSubmitRateLimiter {
	return app.Store().GetOrSet(nuvioContactSubmitRateLimitStoreKey, func() any {
		return &nuvioContactSubmitRateLimiter{
			clients: map[string]nuvioContactSubmitRateLimitWindow{},
		}
	}).(*nuvioContactSubmitRateLimiter)
}

func resolveNuvioContactSubmitRateLimitConfig(app core.App) nuvioContactSubmitRateLimitConfig {
	if raw := app.Store().Get(nuvioContactSubmitRateLimitConfigStoreKey); raw != nil {
		switch value := raw.(type) {
		case nuvioContactSubmitRateLimitConfig:
			return normalizeNuvioContactSubmitRateLimitConfig(value)
		case *nuvioContactSubmitRateLimitConfig:
			if value != nil {
				return normalizeNuvioContactSubmitRateLimitConfig(*value)
			}
		}
	}

	return normalizeNuvioContactSubmitRateLimitConfig(nuvioContactSubmitRateLimitConfig{
		MaxRequests: readNuvioPositiveIntEnv(nuvioContactSubmitRateLimitMaxEnv, nuvioContactSubmitRateLimitDefaultMaxRequests),
		Window: time.Duration(readNuvioPositiveIntEnv(
			nuvioContactSubmitRateLimitWindowEnv,
			nuvioContactSubmitRateLimitDefaultWindowSeconds,
		)) * time.Second,
	})
}

func normalizeNuvioContactSubmitRateLimitConfig(config nuvioContactSubmitRateLimitConfig) nuvioContactSubmitRateLimitConfig {
	if config.MaxRequests < 1 {
		config.MaxRequests = nuvioContactSubmitRateLimitDefaultMaxRequests
	}
	if config.Window <= 0 {
		config.Window = time.Duration(nuvioContactSubmitRateLimitDefaultWindowSeconds) * time.Second
	}

	return config
}

func readNuvioPositiveIntEnv(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}

	return parsed
}

func buildNuvioContactSubmitRateLimitClientKey(e *core.RequestEvent) string {
	method := http.MethodPost
	if e.Request != nil && e.Request.Method != "" {
		method = strings.ToUpper(strings.TrimSpace(e.Request.Method))
	}

	clientIP := strings.TrimSpace(e.RealIP())
	if clientIP == "" {
		clientIP = "unknown"
	}

	return method + " contact-submit|" + clientIP
}

func (limiter *nuvioContactSubmitRateLimiter) Allow(
	key string,
	config nuvioContactSubmitRateLimitConfig,
	now time.Time,
) (bool, time.Duration) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	if limiter.clients == nil {
		limiter.clients = map[string]nuvioContactSubmitRateLimitWindow{}
	}

	if limiter.lastCleanup.IsZero() || now.Sub(limiter.lastCleanup) >= config.Window {
		for clientKey, window := range limiter.clients {
			if !now.Before(window.resetAt) {
				delete(limiter.clients, clientKey)
			}
		}
		limiter.lastCleanup = now
	}

	window, exists := limiter.clients[key]
	if !exists || !now.Before(window.resetAt) {
		window = nuvioContactSubmitRateLimitWindow{
			remaining: config.MaxRequests,
			resetAt:   now.Add(config.Window),
		}
	}

	if window.remaining <= 0 {
		limiter.clients[key] = window
		return false, window.resetAt.Sub(now)
	}

	window.remaining--
	limiter.clients[key] = window
	return true, 0
}

func nuvioContactSubmitRetryAfterSeconds(duration time.Duration) int {
	if duration <= 0 {
		return 1
	}

	seconds := int(duration / time.Second)
	if duration%time.Second != 0 {
		seconds++
	}
	if seconds < 1 {
		return 1
	}

	return seconds
}

// NUVIO CUSTOM END: Explicit public contact submit rate limiter.

type nuvioEmailNotificationsConfig struct {
	Enabled bool
	To      []string
	Cc      []string
}

type nuvioContactNotificationTemplateConfig struct {
	Enabled            bool
	Subject            string
	IntroText          string
	FooterText         string
	IncludeLeadDetails bool
}

type nuvioWebsiteContactFormConfig struct {
	FeatureAvailable             bool
	Enabled                      bool
	PhoneFieldEnabled            bool
	ConfirmationMessage          string
	EmailNotifications           nuvioEmailNotificationsConfig
	BusinessNotificationTemplate nuvioContactNotificationTemplateConfig
}

type nuvioWebsiteWhatsappConfig struct {
	FeatureAvailable             bool
	Enabled                      bool
	Phone                        string
	DefaultMessage               string
	ShowFloatingButton           bool
	EmailNotifications           nuvioEmailNotificationsConfig
	BusinessNotificationTemplate nuvioContactNotificationTemplateConfig
}

type nuvioContactSubmissionPayload struct {
	WebsiteID   string `json:"websiteId"`
	WebsiteSlug string `json:"websiteSlug"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Subject     string `json:"subject"`
	Message     string `json:"message"`
	Source      string `json:"source"`
	Page        string `json:"page"`
}

type nuvioWhatsappInteractionPayload struct {
	WebsiteID   string `json:"websiteId"`
	WebsiteSlug string `json:"websiteSlug"`
	Source      string `json:"source"`
	Page        string `json:"page"`
}

// NUVIO CUSTOM START: Contact form + WhatsApp interaction endpoints with independent email notifications.
func registerNuvioLeadsRoutes(e *core.ServeEvent) {
	contactSubmitHandler := func(e *core.RequestEvent) error {
		if err := checkNuvioPublicContactSubmitRateLimit(e); err != nil {
			return err
		}

		payload := map[string]any{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid contact form payload.", nil)
		}

		if err := validateNuvioPublicPayloadKeys(
			payload,
			map[string]struct{}{
				"websiteId":   {},
				"website":     {},
				"websiteSlug": {},
				"slug":        {},
				"name":        {},
				"email":       {},
				"phone":       {},
				"subject":     {},
				"message":     {},
				"source":      {},
				"page":        {},
			},
		); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteRecord, websiteID, err := resolveNuvioPublicWebsiteFromPayload(e.App, payload, e.Request.URL.Query())
		if err != nil {
			return handleNuvioPublicWebsiteResolveError(e, err)
		}

		_, config, err := loadNuvioWebsiteContactFormConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			e.App.Logger().Error(
				"NUVIO contact settings load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load website Contact Form settings.", nil)
		}

		if !config.FeatureAvailable {
			return e.BadRequestError("Contact Form feature is unavailable for this website.", nil)
		}

		if !config.Enabled {
			return e.BadRequestError("Contact Form is disabled for this website.", nil)
		}

		name, err := validateNuvioPublicRequiredField(payload["name"], "Name", nuvioPublicContactNameMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		emailRaw, err := validateNuvioPublicRequiredField(payload["email"], "Email", 320)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		email, ok := normalizeNuvioEmail(emailRaw)
		if !ok {
			return e.BadRequestError("A valid email is required.", nil)
		}

		message, err := validateNuvioPublicRequiredField(payload["message"], "Message", nuvioPublicContactMessageMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		phone, err := validateNuvioPublicOptionalField(payload["phone"], "Phone", nuvioPublicContactPhoneMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		if !config.PhoneFieldEnabled {
			phone = ""
		}

		subject, err := validateNuvioPublicOptionalField(payload["subject"], "Subject", nuvioPublicContactSubjectMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		source, err := validateNuvioPublicOptionalField(payload["source"], "Source", nuvioPublicContactSourceMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		page, err := validateNuvioPublicOptionalField(payload["page"], "Page", nuvioPublicContactPageMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		contactsCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioContactsCollectionID)
		if err != nil {
			e.App.Logger().Error("NUVIO contact collection resolve failed", "error", err.Error())
			return e.BadRequestError("Contacts collection is missing.", nil)
		}

		contactRecord := core.NewRecord(contactsCollection)
		contactRecord.Set("website", websiteID)
		contactRecord.Set("channel", "form")
		contactRecord.Set("name", name)
		contactRecord.Set("email", email)
		contactRecord.Set("phone", phone)
		contactRecord.Set("subject", subject)
		contactRecord.Set("message", message)
		if sourceFieldName := resolveNuvioCollectionFieldNameByAliases(contactsCollection, []string{"source"}); sourceFieldName != "" {
			contactRecord.Set(sourceFieldName, source)
		}
		if pageFieldName := resolveNuvioCollectionFieldNameByAliases(contactsCollection, []string{"page"}); pageFieldName != "" {
			contactRecord.Set(pageFieldName, page)
		}
		contactRecord.Set("status", "new")

		if err := e.App.Save(contactRecord); err != nil {
			e.App.Logger().Error(
				"NUVIO contact submission save failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to save contact submission.", nil)
		}

		if notificationErr := maybeSendNuvioContactNotificationEmail(
			e.Request.Context(),
			websiteRecord,
			config,
			nuvioContactSubmissionPayload{
				WebsiteID:   websiteID,
				WebsiteSlug: strings.TrimSpace(websiteRecord.GetString("slug")),
				Name:        name,
				Email:       email,
				Phone:       phone,
				Subject:     subject,
				Message:     message,
				Source:      source,
				Page:        page,
			},
		); notificationErr != nil {
			e.App.Logger().Error(
				"NUVIO contact notification send failed",
				"websiteId",
				websiteID,
				"recordId",
				contactRecord.Id,
				"error",
				notificationErr.Error(),
			)
		}

		confirmationMessage := strings.TrimSpace(config.ConfirmationMessage)
		if confirmationMessage == "" {
			confirmationMessage = "Thank you for contacting us. We'll reply soon."
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":                  true,
			"confirmationMessage": confirmationMessage,
		})
	}

	whatsappInteractionHandler := func(e *core.RequestEvent) error {
		payload := map[string]any{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid WhatsApp interaction payload.", nil)
		}

		if err := validateNuvioPublicPayloadKeys(
			payload,
			map[string]struct{}{
				"websiteId":   {},
				"website":     {},
				"websiteSlug": {},
				"slug":        {},
				"source":      {},
				"page":        {},
			},
		); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		websiteRecord, websiteID, err := resolveNuvioPublicWebsiteFromPayload(e.App, payload, e.Request.URL.Query())
		if err != nil {
			return handleNuvioPublicWebsiteResolveError(e, err)
		}

		_, config, err := loadNuvioWebsiteWhatsappConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}
			e.App.Logger().Error(
				"NUVIO whatsapp settings load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load website WhatsApp settings.", nil)
		}

		if !config.FeatureAvailable {
			return e.BadRequestError("WhatsApp feature is unavailable for this website.", nil)
		}

		if !config.Enabled {
			return e.BadRequestError("WhatsApp is disabled for this website.", nil)
		}

		whatsappCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioWhatsappCollectionID)
		if err != nil {
			e.App.Logger().Error("NUVIO whatsapp collection resolve failed", "error", err.Error())
			return e.BadRequestError("Whatsapp collection is missing.", nil)
		}

		source, err := validateNuvioPublicOptionalField(payload["source"], "Source", nuvioPublicWhatsappSourceMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		page, err := validateNuvioPublicOptionalField(payload["page"], "Page", nuvioPublicWhatsappPageMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		if source == "" && page == "" {
			return e.BadRequestError("At least source or page is required.", nil)
		}

		interactionRecord := core.NewRecord(whatsappCollection)
		interactionRecord.Set("website", websiteID)
		interactionRecord.Set("source", source)
		interactionRecord.Set("page", page)
		if statusFieldName := resolveNuvioCollectionFieldNameByAliases(whatsappCollection, []string{"status"}); statusFieldName != "" {
			interactionRecord.Set(statusFieldName, "new")
		}

		if err := e.App.Save(interactionRecord); err != nil {
			e.App.Logger().Error(
				"NUVIO whatsapp interaction save failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to save WhatsApp interaction.", nil)
		}

		if notificationErr := maybeSendNuvioWhatsappNotificationEmail(
			e.Request.Context(),
			websiteRecord,
			config,
			nuvioWhatsappInteractionPayload{
				WebsiteID:   websiteID,
				WebsiteSlug: strings.TrimSpace(websiteRecord.GetString("slug")),
				Source:      source,
				Page:        page,
			},
		); notificationErr != nil {
			e.App.Logger().Error(
				"NUVIO whatsapp notification send failed",
				"websiteId",
				websiteID,
				"recordId",
				interactionRecord.Id,
				"error",
				notificationErr.Error(),
			)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok": true,
		})
	}

	leadsGroup := e.Router.Group("/api/nuvio/leads")
	leadsGroup.POST("/contact/submit", contactSubmitHandler)
	leadsGroup.POST("/whatsapp/interaction", whatsappInteractionHandler)

	// Keep compatibility with direct endpoint naming used by website integrations.
	e.Router.POST("/api/nuvio/contact/submit", contactSubmitHandler)
	e.Router.POST("/api/nuvio/whatsapp/interactions", whatsappInteractionHandler)

	// Register scoped backoffice Leads dashboard read endpoint.
	registerNuvioLeadsDashboardRoutes(e)
}

func validateNuvioPublicPayloadKeys(payload map[string]any, allowed map[string]struct{}) error {
	for rawKey := range payload {
		key := strings.TrimSpace(rawKey)
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("Field %q is not allowed in this endpoint.", key)
		}
	}

	return nil
}

func validateNuvioPublicRequiredField(raw any, fieldLabel string, maxLen int) (string, error) {
	normalized, err := validateNuvioPublicOptionalField(raw, fieldLabel, maxLen)
	if err != nil {
		return "", err
	}
	if normalized == "" {
		return "", fmt.Errorf("%s is required.", fieldLabel)
	}

	return normalized, nil
}

func validateNuvioPublicOptionalField(raw any, fieldLabel string, maxLen int) (string, error) {
	normalized, err := normalizeNuvioPublicFieldString(raw, fieldLabel)
	if err != nil {
		return "", err
	}
	if normalized == "" {
		return "", nil
	}

	if maxLen > 0 && len([]rune(normalized)) > maxLen {
		return "", fmt.Errorf("%s is too long. Maximum %d characters.", fieldLabel, maxLen)
	}

	return normalized, nil
}

func normalizeNuvioPublicFieldString(raw any, fieldLabel string) (string, error) {
	switch typed := raw.(type) {
	case nil:
		return "", nil
	case string:
		return strings.TrimSpace(typed), nil
	default:
		return "", fmt.Errorf("%s must be a string.", fieldLabel)
	}
}

func resolveNuvioPublicWebsiteFromPayload(
	app core.App,
	payload map[string]any,
	queryValues url.Values,
) (*core.Record, string, error) {
	websiteSlug, err := normalizeNuvioPublicFieldString(payload["websiteSlug"], "websiteSlug")
	if err != nil {
		return nil, "", err
	}
	if websiteSlug == "" {
		websiteSlug, err = normalizeNuvioPublicFieldString(payload["slug"], "slug")
		if err != nil {
			return nil, "", err
		}
	}
	if websiteSlug == "" {
		websiteSlug = strings.TrimSpace(queryValues.Get("websiteSlug"))
	}

	if websiteSlug != "" {
		website, err := findNuvioPublicWebsiteBySlugOrDomain(app, websiteSlug)
		if err != nil {
			return nil, "", err
		}

		return website, strings.TrimSpace(website.Id), nil
	}

	websiteID, err := normalizeNuvioPublicFieldString(payload["websiteId"], "websiteId")
	if err != nil {
		return nil, "", err
	}
	if websiteID == "" {
		websiteID, err = normalizeNuvioPublicFieldString(payload["website"], "website")
		if err != nil {
			return nil, "", err
		}
	}
	if websiteID == "" {
		websiteID = strings.TrimSpace(queryValues.Get("websiteId"))
	}
	if websiteID == "" {
		return nil, "", errNuvioPublicWebsiteContextMissing
	}

	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		return nil, "", err
	}

	return website, websiteID, nil
}

func handleNuvioPublicWebsiteResolveError(e *core.RequestEvent, err error) error {
	if errors.Is(err, errNuvioPublicWebsiteContextMissing) {
		return e.BadRequestError("Missing website context.", nil)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return e.NotFoundError("Website not found.", nil)
	}

	e.App.Logger().Error("NUVIO public website resolve failed", "error", err.Error())
	return e.BadRequestError("Failed to resolve website context.", nil)
}

func loadNuvioWebsiteContactFormConfig(
	app core.App,
	websiteID string,
) (*core.Record, nuvioWebsiteContactFormConfig, error) {
	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		return nil, nuvioWebsiteContactFormConfig{}, err
	}

	settings := parseNuvioSettingsObject(website.Get("settings"))
	config := nuvioWebsiteContactFormConfig{
		FeatureAvailable:    true,
		Enabled:             true,
		PhoneFieldEnabled:   true,
		ConfirmationMessage: "",
		EmailNotifications: nuvioEmailNotificationsConfig{
			Enabled: false,
			To:      []string{},
			Cc:      []string{},
		},
		BusinessNotificationTemplate: nuvioContactNotificationTemplateConfig{
			Enabled:            false,
			Subject:            "",
			IntroText:          "",
			FooterText:         "",
			IncludeLeadDetails: true,
		},
	}

	if featureFlags, ok := toStringAnyMap(settings["featureFlags"]); ok {
		if value, ok := parseBoolValue(featureFlags["contactForm"]); ok {
			config.FeatureAvailable = value
		}
	}

	if contactFormSettings, ok := toStringAnyMap(settings["contactForm"]); ok {
		if value, ok := parseBoolValue(contactFormSettings["enabled"]); ok {
			config.Enabled = value
		}

		config.ConfirmationMessage = strings.TrimSpace(parseStringValue(contactFormSettings["confirmationMessage"]))

		if fields, ok := toStringAnyMap(contactFormSettings["fields"]); ok {
			if value, ok := parseBoolValue(fields["phone"]); ok {
				config.PhoneFieldEnabled = value
			}
		}

		legacyDestination := strings.TrimSpace(parseStringValue(contactFormSettings["emailDestination"]))
		config.EmailNotifications = parseNuvioEmailNotificationsConfig(
			contactFormSettings["emailNotifications"],
			legacyDestination,
		)

		if emailNotificationsSettings, ok := toStringAnyMap(contactFormSettings["emailNotifications"]); ok {
			config.BusinessNotificationTemplate = parseNuvioContactNotificationTemplateConfig(
				emailNotificationsSettings["template"],
			)
		}
	}

	return website, config, nil
}

func loadNuvioWebsiteWhatsappConfig(
	app core.App,
	websiteID string,
) (*core.Record, nuvioWebsiteWhatsappConfig, error) {
	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		return nil, nuvioWebsiteWhatsappConfig{}, err
	}

	settings := parseNuvioSettingsObject(website.Get("settings"))
	config := nuvioWebsiteWhatsappConfig{
		FeatureAvailable:   true,
		Enabled:            false,
		Phone:              "",
		DefaultMessage:     "",
		ShowFloatingButton: false,
		EmailNotifications: nuvioEmailNotificationsConfig{
			Enabled: false,
			To:      []string{},
			Cc:      []string{},
		},
		BusinessNotificationTemplate: nuvioContactNotificationTemplateConfig{
			Enabled:            false,
			Subject:            "",
			IntroText:          "",
			FooterText:         "",
			IncludeLeadDetails: true,
		},
	}

	if featureFlags, ok := toStringAnyMap(settings["featureFlags"]); ok {
		if value, ok := parseBoolValue(featureFlags["whatsapp"]); ok {
			config.FeatureAvailable = value
		}
	}

	if whatsappSettings, ok := toStringAnyMap(settings["whatsapp"]); ok {
		if value, ok := parseBoolValue(whatsappSettings["enabled"]); ok {
			config.Enabled = value
		}

		config.Phone = strings.TrimSpace(parseStringValue(whatsappSettings["phone"]))
		config.DefaultMessage = strings.TrimSpace(parseStringValue(whatsappSettings["defaultMessage"]))
		if value, ok := parseBoolValue(whatsappSettings["showFloatingButton"]); ok {
			config.ShowFloatingButton = value
		}

		config.EmailNotifications = parseNuvioEmailNotificationsConfig(
			whatsappSettings["emailNotifications"],
			"",
		)

		if emailNotificationsSettings, ok := toStringAnyMap(whatsappSettings["emailNotifications"]); ok {
			config.BusinessNotificationTemplate = parseNuvioContactNotificationTemplateConfig(
				emailNotificationsSettings["template"],
			)
		}
	}

	return website, config, nil
}

func parseNuvioEmailNotificationsConfig(raw any, legacyTo string) nuvioEmailNotificationsConfig {
	config := nuvioEmailNotificationsConfig{
		Enabled: false,
		To:      []string{},
		Cc:      []string{},
	}

	hasExplicitConfig := false
	if settings, ok := toStringAnyMap(raw); ok {
		hasExplicitConfig = true
		if value, ok := parseBoolValue(settings["enabled"]); ok {
			config.Enabled = value
		}

		config.To = parseNuvioEmailList(settings["to"])
		config.Cc = parseNuvioEmailList(settings["cc"])
	}

	if len(config.To) == 0 {
		legacyEmail, ok := normalizeNuvioEmail(legacyTo)
		if ok {
			config.To = append(config.To, legacyEmail)
			if !hasExplicitConfig {
				config.Enabled = true
			}
		}
	}

	return config
}

func parseNuvioContactNotificationTemplateConfig(raw any) nuvioContactNotificationTemplateConfig {
	config := nuvioContactNotificationTemplateConfig{
		Enabled:            false,
		Subject:            "",
		IntroText:          "",
		FooterText:         "",
		IncludeLeadDetails: true,
	}

	settings, ok := toStringAnyMap(raw)
	if !ok {
		return config
	}

	if value, ok := parseBoolValue(settings["enabled"]); ok {
		config.Enabled = value
	}

	if value, ok := settings["subject"].(string); ok && value != "" {
		config.Subject = value
	}

	if value, ok := settings["introText"].(string); ok && value != "" {
		config.IntroText = value
	}

	if value, ok := settings["footerText"].(string); ok && value != "" {
		config.FooterText = value
	}

	if value, ok := parseBoolValue(settings["includeLeadDetails"]); ok {
		config.IncludeLeadDetails = value
	}

	return config
}

func parseNuvioEmailList(raw any) []string {
	normalized := []string{}
	seen := map[string]struct{}{}

	appendEmail := func(candidate string) {
		email, ok := normalizeNuvioEmail(candidate)
		if !ok {
			return
		}

		if _, exists := seen[email]; exists {
			return
		}

		seen[email] = struct{}{}
		normalized = append(normalized, email)
	}

	switch typed := raw.(type) {
	case []any:
		for _, item := range typed {
			switch candidate := item.(type) {
			case string:
				for _, piece := range splitNuvioEmailString(candidate) {
					appendEmail(piece)
				}
			default:
				if entry, ok := toStringAnyMap(candidate); ok {
					appendEmail(parseStringValue(entry["email"]))
					appendEmail(parseStringValue(entry["address"]))
					appendEmail(parseStringValue(entry["value"]))
				}
			}
		}
	case []string:
		for _, candidate := range typed {
			for _, piece := range splitNuvioEmailString(candidate) {
				appendEmail(piece)
			}
		}
	case string:
		for _, candidate := range splitNuvioEmailString(typed) {
			appendEmail(candidate)
		}
	default:
		appendEmail(parseStringValue(typed))
	}

	return normalized
}

func splitNuvioEmailString(raw string) []string {
	source := strings.TrimSpace(raw)
	if source == "" {
		return []string{}
	}

	normalized := strings.NewReplacer("\n", ",", ";", ",").Replace(source)
	parts := strings.Split(normalized, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

func buildNuvioDefaultContactNotificationEmail(
	websiteName string,
	payload nuvioContactSubmissionPayload,
	submittedAt time.Time,
) (string, string) {
	subjectPrefix := strings.TrimSpace(payload.Subject)
	if subjectPrefix == "" {
		subjectPrefix = "New contact form submission"
	}

	subject := subjectPrefix
	if websiteName != "" {
		subject = fmt.Sprintf("%s - %s", subjectPrefix, websiteName)
	}

	textBodyLines := []string{
		fmt.Sprintf("Website: %s", websiteName),
		fmt.Sprintf("Submitted at: %s", submittedAt.Format(time.RFC3339)),
		fmt.Sprintf("Name: %s", strings.TrimSpace(payload.Name)),
		fmt.Sprintf("Email: %s", strings.TrimSpace(payload.Email)),
	}

	trimmedPhone := strings.TrimSpace(payload.Phone)
	if trimmedPhone != "" {
		textBodyLines = append(textBodyLines, fmt.Sprintf("Phone: %s", trimmedPhone))
	}

	trimmedSubject := strings.TrimSpace(payload.Subject)
	if trimmedSubject != "" {
		textBodyLines = append(textBodyLines, fmt.Sprintf("Subject: %s", trimmedSubject))
	}

	textBodyLines = append(textBodyLines, "", "Message:", strings.TrimSpace(payload.Message))
	return subject, strings.Join(textBodyLines, "\n")
}

func sanitizeNuvioTemplateSubject(raw string) string {
	normalized := strings.NewReplacer("\r", " ", "\n", " ").Replace(raw)
	normalized = strings.Join(strings.Fields(normalized), " ")
	normalized = strings.TrimSpace(normalized)
	return truncateNuvioStringByRunes(normalized, nuvioTemplateSubjectMaxLen)
}

func sanitizeNuvioTemplateText(raw string) string {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimSpace(normalized)
	return truncateNuvioStringByRunes(normalized, nuvioTemplateTextMaxLen)
}

func truncateNuvioStringByRunes(raw string, maxLen int) string {
	if maxLen <= 0 || raw == "" {
		return ""
	}

	runes := []rune(raw)
	if len(runes) <= maxLen {
		return raw
	}

	return string(runes[:maxLen])
}

func sanitizeNuvioTemplateSingleLineValue(raw string, fallback string) string {
	normalized := strings.NewReplacer("\r", " ", "\n", " ").Replace(raw)
	normalized = strings.Join(strings.Fields(normalized), " ")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return fallback
	}
	return normalized
}

func sanitizeNuvioTemplateMultilineValue(raw string, fallback string) string {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return fallback
	}
	return normalized
}

func buildNuvioContactTemplateVariables(
	websiteName string,
	submittedAt time.Time,
	payload nuvioContactSubmissionPayload,
) map[string]string {
	displayWebsiteName := strings.TrimSpace(websiteName)
	if displayWebsiteName == "" {
		displayWebsiteName = "Website"
	}

	return map[string]string{
		"websiteName": displayWebsiteName,
		"submittedAt": submittedAt.Format(time.RFC3339),
		"leadSource":  "Contact form",
		"leadName":    sanitizeNuvioTemplateSingleLineValue(payload.Name, "Not provided"),
		"leadEmail":   sanitizeNuvioTemplateSingleLineValue(payload.Email, "Not provided"),
		"leadPhone":   sanitizeNuvioTemplateSingleLineValue(payload.Phone, "Not provided"),
		"leadSubject": sanitizeNuvioTemplateSingleLineValue(payload.Subject, "Not provided"),
		"leadMessage": sanitizeNuvioTemplateMultilineValue(payload.Message, "Not provided"),
	}
}

func replaceNuvioContactTemplateVariables(raw string, values map[string]string) string {
	replacer := strings.NewReplacer(
		"{{websiteName}}", values["websiteName"],
		"{{submittedAt}}", values["submittedAt"],
		"{{leadSource}}", values["leadSource"],
		"{{leadName}}", values["leadName"],
		"{{leadEmail}}", values["leadEmail"],
		"{{leadPhone}}", values["leadPhone"],
		"{{leadSubject}}", values["leadSubject"],
		"{{leadMessage}}", values["leadMessage"],
	)

	return replacer.Replace(raw)
}

func buildNuvioContactTemplateBodyDetails(values map[string]string) string {
	lines := []string{
		fmt.Sprintf("Website: %s", values["websiteName"]),
		fmt.Sprintf("Submitted at: %s", values["submittedAt"]),
		fmt.Sprintf("Lead source: %s", values["leadSource"]),
		fmt.Sprintf("Name: %s", values["leadName"]),
		fmt.Sprintf("Email: %s", values["leadEmail"]),
		fmt.Sprintf("Phone: %s", values["leadPhone"]),
		fmt.Sprintf("Subject: %s", values["leadSubject"]),
		"",
		"Message:",
		values["leadMessage"],
	}

	return strings.Join(lines, "\n")
}

func buildNuvioContactNotificationTemplateEmail(
	config nuvioContactNotificationTemplateConfig,
	websiteName string,
	submittedAt time.Time,
	payload nuvioContactSubmissionPayload,
) (string, string, bool) {
	if !config.Enabled {
		return "", "", false
	}

	values := buildNuvioContactTemplateVariables(websiteName, submittedAt, payload)
	subject := sanitizeNuvioTemplateSubject(
		replaceNuvioContactTemplateVariables(config.Subject, values),
	)
	if subject == "" {
		return "", "", false
	}

	sections := []string{}

	introText := sanitizeNuvioTemplateText(
		replaceNuvioContactTemplateVariables(config.IntroText, values),
	)
	if introText != "" {
		sections = append(sections, introText)
	}

	if config.IncludeLeadDetails {
		detailsText := sanitizeNuvioTemplateText(buildNuvioContactTemplateBodyDetails(values))
		if detailsText != "" {
			sections = append(sections, detailsText)
		}
	}

	footerText := sanitizeNuvioTemplateText(
		replaceNuvioContactTemplateVariables(config.FooterText, values),
	)
	if footerText != "" {
		sections = append(sections, footerText)
	}

	body := strings.TrimSpace(strings.Join(sections, "\n\n"))
	if body == "" {
		return "", "", false
	}

	return subject, body, true
}

func buildNuvioDefaultWhatsappNotificationEmail(
	websiteName string,
	payload nuvioWhatsappInteractionPayload,
	submittedAt time.Time,
) (string, string) {
	subject := "New WhatsApp interaction"
	if websiteName != "" {
		subject = fmt.Sprintf("%s - %s", subject, websiteName)
	}

	textBodyLines := []string{
		fmt.Sprintf("Website: %s", websiteName),
		fmt.Sprintf("Tracked at: %s", submittedAt.Format(time.RFC3339)),
		fmt.Sprintf("Source: %s", strings.TrimSpace(payload.Source)),
		fmt.Sprintf("Page: %s", strings.TrimSpace(payload.Page)),
	}

	return subject, strings.Join(textBodyLines, "\n")
}

func buildNuvioWhatsappTemplateVariables(
	websiteName string,
	submittedAt time.Time,
	config nuvioWebsiteWhatsappConfig,
	payload nuvioWhatsappInteractionPayload,
) map[string]string {
	displayWebsiteName := strings.TrimSpace(websiteName)
	if displayWebsiteName == "" {
		displayWebsiteName = "Website"
	}

	return map[string]string{
		"websiteName":    displayWebsiteName,
		"submittedAt":    submittedAt.Format(time.RFC3339),
		"leadSource":     "WhatsApp",
		"source":         sanitizeNuvioTemplateSingleLineValue(payload.Source, "Not provided"),
		"pageUrl":        sanitizeNuvioTemplateSingleLineValue(payload.Page, "Not provided"),
		"whatsappPhone":  sanitizeNuvioTemplateSingleLineValue(config.Phone, "Not configured"),
		"defaultMessage": sanitizeNuvioTemplateMultilineValue(config.DefaultMessage, "Not configured"),
	}
}

func replaceNuvioWhatsappTemplateVariables(raw string, values map[string]string) string {
	replacer := strings.NewReplacer(
		"{{websiteName}}", values["websiteName"],
		"{{submittedAt}}", values["submittedAt"],
		"{{leadSource}}", values["leadSource"],
		"{{source}}", values["source"],
		"{{pageUrl}}", values["pageUrl"],
		"{{whatsappPhone}}", values["whatsappPhone"],
		"{{defaultMessage}}", values["defaultMessage"],
	)

	return replacer.Replace(raw)
}

func buildNuvioWhatsappTemplateBodyDetails(values map[string]string) string {
	lines := []string{
		fmt.Sprintf("Website: %s", values["websiteName"]),
		fmt.Sprintf("Tracked at: %s", values["submittedAt"]),
		fmt.Sprintf("Lead source: %s", values["leadSource"]),
		fmt.Sprintf("Source: %s", values["source"]),
		fmt.Sprintf("Page URL: %s", values["pageUrl"]),
		fmt.Sprintf("WhatsApp phone: %s", values["whatsappPhone"]),
		"",
		"Default message:",
		values["defaultMessage"],
	}

	return strings.Join(lines, "\n")
}

func buildNuvioWhatsappNotificationTemplateEmail(
	config nuvioContactNotificationTemplateConfig,
	websiteName string,
	submittedAt time.Time,
	whatsappConfig nuvioWebsiteWhatsappConfig,
	payload nuvioWhatsappInteractionPayload,
) (string, string, bool) {
	if !config.Enabled {
		return "", "", false
	}

	values := buildNuvioWhatsappTemplateVariables(websiteName, submittedAt, whatsappConfig, payload)
	subject := sanitizeNuvioTemplateSubject(
		replaceNuvioWhatsappTemplateVariables(config.Subject, values),
	)
	if subject == "" {
		return "", "", false
	}

	sections := []string{}

	introText := sanitizeNuvioTemplateText(
		replaceNuvioWhatsappTemplateVariables(config.IntroText, values),
	)
	if introText != "" {
		sections = append(sections, introText)
	}

	if config.IncludeLeadDetails {
		detailsText := sanitizeNuvioTemplateText(buildNuvioWhatsappTemplateBodyDetails(values))
		if detailsText != "" {
			sections = append(sections, detailsText)
		}
	}

	footerText := sanitizeNuvioTemplateText(
		replaceNuvioWhatsappTemplateVariables(config.FooterText, values),
	)
	if footerText != "" {
		sections = append(sections, footerText)
	}

	body := strings.TrimSpace(strings.Join(sections, "\n\n"))
	if body == "" {
		return "", "", false
	}

	return subject, body, true
}

func maybeSendNuvioContactNotificationEmail(
	ctx context.Context,
	website *core.Record,
	config nuvioWebsiteContactFormConfig,
	payload nuvioContactSubmissionPayload,
) error {
	if !config.EmailNotifications.Enabled || len(config.EmailNotifications.To) == 0 {
		return nil
	}

	resendConfig, err := loadNuvioResendConfig()
	if err != nil {
		return err
	}

	websiteName := resolveWebsiteDisplayName(website)
	submittedAt := time.Now().UTC()
	subject, textBody := buildNuvioDefaultContactNotificationEmail(websiteName, payload, submittedAt)
	if templateSubject, templateBody, ok := buildNuvioContactNotificationTemplateEmail(
		config.BusinessNotificationTemplate,
		websiteName,
		submittedAt,
		payload,
	); ok {
		subject = templateSubject
		textBody = templateBody
	}

	replyTo := []string{}
	if normalizedReplyTo, ok := normalizeNuvioEmail(payload.Email); ok {
		replyTo = append(replyTo, normalizedReplyTo)
	}
	if len(replyTo) > nuvioDefaultReplyToMaxSize {
		replyTo = replyTo[:nuvioDefaultReplyToMaxSize]
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      config.EmailNotifications.To,
		Cc:      config.EmailNotifications.Cc,
		ReplyTo: replyTo,
		Subject: subject,
		Text:    textBody,
	})
}

func maybeSendNuvioWhatsappNotificationEmail(
	ctx context.Context,
	website *core.Record,
	config nuvioWebsiteWhatsappConfig,
	payload nuvioWhatsappInteractionPayload,
) error {
	if !config.EmailNotifications.Enabled || len(config.EmailNotifications.To) == 0 {
		return nil
	}

	resendConfig, err := loadNuvioResendConfig()
	if err != nil {
		return err
	}

	websiteName := resolveWebsiteDisplayName(website)
	submittedAt := time.Now().UTC()
	subject, textBody := buildNuvioDefaultWhatsappNotificationEmail(websiteName, payload, submittedAt)
	if templateSubject, templateBody, ok := buildNuvioWhatsappNotificationTemplateEmail(
		config.BusinessNotificationTemplate,
		websiteName,
		submittedAt,
		config,
		payload,
	); ok {
		subject = templateSubject
		textBody = templateBody
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      config.EmailNotifications.To,
		Cc:      config.EmailNotifications.Cc,
		Subject: subject,
		Text:    textBody,
	})
}

// NUVIO CUSTOM END: Contact form + WhatsApp interaction endpoints with independent email notifications.
