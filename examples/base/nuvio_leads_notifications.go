package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioContactsCollectionID  = "pbc_1661203100"
	nuvioWhatsappCollectionID  = "pbc_1661203200"
	nuvioDefaultReplyToMaxSize = 1
)

type nuvioEmailNotificationsConfig struct {
	Enabled bool
	To      []string
	Cc      []string
}

type nuvioWebsiteContactFormConfig struct {
	FeatureAvailable    bool
	Enabled             bool
	PhoneFieldEnabled   bool
	ConfirmationMessage string
	EmailNotifications  nuvioEmailNotificationsConfig
}

type nuvioWebsiteWhatsappConfig struct {
	FeatureAvailable   bool
	Enabled            bool
	Phone              string
	DefaultMessage     string
	ShowFloatingButton bool
	EmailNotifications nuvioEmailNotificationsConfig
}

type nuvioContactSubmissionPayload struct {
	WebsiteID string `json:"websiteId"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
}

type nuvioWhatsappInteractionPayload struct {
	WebsiteID string `json:"websiteId"`
	Source    string `json:"source"`
	Page      string `json:"page"`
}

// NUVIO CUSTOM START: Contact form + WhatsApp interaction endpoints with independent email notifications.
func registerNuvioLeadsRoutes(e *core.ServeEvent) {
	contactSubmitHandler := func(e *core.RequestEvent) error {
		payload := nuvioContactSubmissionPayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid contact form payload.", err)
		}

		websiteID := strings.TrimSpace(payload.WebsiteID)
		if websiteID == "" {
			websiteID = strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		}
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		website, config, err := loadNuvioWebsiteContactFormConfig(e.App, websiteID)
		if err != nil {
			return e.BadRequestError("Failed to load website Contact Form settings.", err)
		}

		if !config.FeatureAvailable {
			return e.BadRequestError("Contact Form feature is unavailable for this website.", nil)
		}

		if !config.Enabled {
			return e.BadRequestError("Contact Form is disabled for this website.", nil)
		}

		name := strings.TrimSpace(payload.Name)
		if name == "" {
			return e.BadRequestError("Name is required.", nil)
		}

		email, ok := normalizeNuvioEmail(payload.Email)
		if !ok {
			return e.BadRequestError("A valid email is required.", nil)
		}

		message := strings.TrimSpace(payload.Message)
		if message == "" {
			return e.BadRequestError("Message is required.", nil)
		}

		phone := strings.TrimSpace(payload.Phone)
		if !config.PhoneFieldEnabled {
			phone = ""
		}

		subject := strings.TrimSpace(payload.Subject)

		contactsCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioContactsCollectionID)
		if err != nil {
			return e.BadRequestError("Contacts collection is missing.", err)
		}

		contactRecord := core.NewRecord(contactsCollection)
		contactRecord.Set("website", websiteID)
		contactRecord.Set("channel", "form")
		contactRecord.Set("name", name)
		contactRecord.Set("email", email)
		contactRecord.Set("phone", phone)
		contactRecord.Set("subject", subject)
		contactRecord.Set("message", message)
		contactRecord.Set("status", "new")

		if err := e.App.Save(contactRecord); err != nil {
			return e.BadRequestError("Failed to save contact submission.", err)
		}

		if notificationErr := maybeSendNuvioContactNotificationEmail(
			e.Request.Context(),
			website,
			config,
			nuvioContactSubmissionPayload{
				WebsiteID: websiteID,
				Name:      name,
				Email:     email,
				Phone:     phone,
				Subject:   subject,
				Message:   message,
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
		payload := nuvioWhatsappInteractionPayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid WhatsApp interaction payload.", err)
		}

		websiteID := strings.TrimSpace(payload.WebsiteID)
		if websiteID == "" {
			websiteID = strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		}
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		website, config, err := loadNuvioWebsiteWhatsappConfig(e.App, websiteID)
		if err != nil {
			return e.BadRequestError("Failed to load website WhatsApp settings.", err)
		}

		if !config.FeatureAvailable {
			return e.BadRequestError("WhatsApp feature is unavailable for this website.", nil)
		}

		if !config.Enabled {
			return e.BadRequestError("WhatsApp is disabled for this website.", nil)
		}

		whatsappCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioWhatsappCollectionID)
		if err != nil {
			return e.BadRequestError("Whatsapp collection is missing.", err)
		}

		source := strings.TrimSpace(payload.Source)
		page := strings.TrimSpace(payload.Page)

		interactionRecord := core.NewRecord(whatsappCollection)
		interactionRecord.Set("website", websiteID)
		interactionRecord.Set("source", source)
		interactionRecord.Set("page", page)

		if err := e.App.Save(interactionRecord); err != nil {
			return e.BadRequestError("Failed to save WhatsApp interaction.", err)
		}

		if notificationErr := maybeSendNuvioWhatsappNotificationEmail(
			e.Request.Context(),
			website,
			config,
			nuvioWhatsappInteractionPayload{
				WebsiteID: websiteID,
				Source:    source,
				Page:      page,
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
		fmt.Sprintf("Submitted at: %s", time.Now().UTC().Format(time.RFC3339)),
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
		Text:    strings.Join(textBodyLines, "\n"),
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
	subject := "New WhatsApp interaction"
	if websiteName != "" {
		subject = fmt.Sprintf("%s - %s", subject, websiteName)
	}

	textBodyLines := []string{
		fmt.Sprintf("Website: %s", websiteName),
		fmt.Sprintf("Tracked at: %s", time.Now().UTC().Format(time.RFC3339)),
		fmt.Sprintf("Source: %s", strings.TrimSpace(payload.Source)),
		fmt.Sprintf("Page: %s", strings.TrimSpace(payload.Page)),
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      config.EmailNotifications.To,
		Cc:      config.EmailNotifications.Cc,
		Subject: subject,
		Text:    strings.Join(textBodyLines, "\n"),
	})
}

// NUVIO CUSTOM END: Contact form + WhatsApp interaction endpoints with independent email notifications.
