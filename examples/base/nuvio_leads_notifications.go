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
	nuvioTemplateSubjectMaxLen = 160
	nuvioTemplateTextMaxLen    = 4000
)

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
