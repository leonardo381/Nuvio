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
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

const (
	nuvioSubscribersCollectionID = "pbc_1661203400"
	nuvioCampaignsCollectionID   = "pbc_1661203500"
)

var nuvioNewsletterHTTPClient = &http.Client{
	Timeout: 20 * time.Second,
}

type nuvioWebsiteNewsletterConfig struct {
	FeatureAvailable bool
}

type nuvioCampaignSendResult struct {
	CampaignID      string `json:"campaignId"`
	Status          string `json:"status"`
	RecipientsCount int    `json:"recipientsCount"`
	SentAt          string `json:"sentAt"`
}

type resendSendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to,omitempty"`
	Bcc     []string `json:"bcc,omitempty"`
	Subject string   `json:"subject"`
	Html    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

type resendSendEmailResponse struct {
	ID string `json:"id"`
}

// NUVIO CUSTOM START: Newsletter V1 send endpoint (server-side campaign dispatch via Resend).
func registerNuvioNewsletterRoutes(e *core.ServeEvent) {
	newsletterGroup := e.Router.Group("/api/nuvio/newsletter").Bind(apis.RequireSuperuserAuth())

	newsletterGroup.POST("/campaigns/send", func(e *core.RequestEvent) error {
		payload := struct {
			CampaignID string `json:"campaignId"`
		}{}

		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid campaign send payload.", err)
		}

		campaignID := strings.TrimSpace(payload.CampaignID)
		if campaignID == "" {
			campaignID = strings.TrimSpace(e.Request.URL.Query().Get("campaignId"))
		}
		if campaignID == "" {
			return e.BadRequestError("Missing campaignId.", nil)
		}

		result, err := sendNuvioNewsletterCampaign(e.App, e.Request.Context(), campaignID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Campaign not found.", nil)
			}

			return e.BadRequestError(err.Error(), nil)
		}

		return e.JSON(http.StatusOK, result)
	})
}

func sendNuvioNewsletterCampaign(app core.App, ctx context.Context, campaignID string) (*nuvioCampaignSendResult, error) {
	campaign, err := app.FindRecordById(nuvioCampaignsCollectionID, campaignID)
	if err != nil {
		return nil, err
	}

	status := strings.ToLower(strings.TrimSpace(campaign.GetString("status")))
	if status == "sent" {
		return nil, fmt.Errorf("Campaign is already sent")
	}

	websiteID := strings.TrimSpace(campaign.GetString("website"))
	if websiteID == "" {
		return nil, fmt.Errorf("Campaign is missing website relation")
	}

	newsletterConfig, err := loadNuvioWebsiteNewsletterConfig(app, websiteID)
	if err != nil {
		return nil, err
	}
	if !newsletterConfig.FeatureAvailable {
		return nil, fmt.Errorf("Newsletter feature is unavailable for this website")
	}

	subject := strings.TrimSpace(campaign.GetString("subject"))
	if subject == "" {
		return nil, fmt.Errorf("Campaign subject is required")
	}

	body := strings.TrimSpace(campaign.GetString("body"))
	if body == "" {
		return nil, fmt.Errorf("Campaign body is required")
	}

	recipientsType := strings.ToLower(strings.TrimSpace(campaign.GetString("recipientsType")))
	if recipientsType == "" {
		recipientsType = "all"
	}

	recipients, err := resolveNuvioCampaignRecipients(app, websiteID, recipientsType, campaign.Get("recipientsIds"))
	if err != nil {
		return nil, err
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("No active recipients found for this campaign")
	}

	resendAPIKey := strings.TrimSpace(os.Getenv("NUVIO_RESEND_API_KEY"))
	if resendAPIKey == "" {
		resendAPIKey = strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	}
	if resendAPIKey == "" {
		return nil, fmt.Errorf("Missing Resend API key (set NUVIO_RESEND_API_KEY or RESEND_API_KEY)")
	}

	fromAddress := strings.TrimSpace(os.Getenv("NUVIO_RESEND_FROM"))
	if fromAddress == "" {
		fromAddress = strings.TrimSpace(os.Getenv("RESEND_FROM"))
	}
	if fromAddress == "" {
		fromAddress = "onboarding@resend.dev"
	}

	if err := sendNuvioCampaignEmailViaResend(ctx, resendAPIKey, fromAddress, subject, body, recipients); err != nil {
		return nil, err
	}

	sentAt := time.Now().UTC().Format(time.RFC3339)
	campaign.Set("status", "sent")
	campaign.Set("sentAt", sentAt)
	campaign.Set("recipientsCount", len(recipients))

	if err := app.Save(campaign); err != nil {
		return nil, err
	}

	return &nuvioCampaignSendResult{
		CampaignID:      campaign.Id,
		Status:          "sent",
		RecipientsCount: len(recipients),
		SentAt:          sentAt,
	}, nil
}

func loadNuvioWebsiteNewsletterConfig(app core.App, websiteID string) (nuvioWebsiteNewsletterConfig, error) {
	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		return nuvioWebsiteNewsletterConfig{}, err
	}

	settings := parseNuvioSettingsObject(website.Get("settings"))

	config := nuvioWebsiteNewsletterConfig{
		FeatureAvailable: true,
	}

	if featureFlags, ok := toStringAnyMap(settings["featureFlags"]); ok {
		if value, ok := parseBoolValue(featureFlags["newsletter"]); ok {
			config.FeatureAvailable = value
		}
	}

	return config, nil
}

func resolveNuvioCampaignRecipients(
	app core.App,
	websiteID string,
	recipientsType string,
	rawRecipientsIDs any,
) ([]string, error) {
	recipients := map[string]struct{}{}

	switch recipientsType {
	case "manual":
		recipientIDs := parseNuvioRecipientIDs(rawRecipientsIDs)
		if len(recipientIDs) == 0 {
			return nil, fmt.Errorf("Campaign recipientsType is manual but recipientsIds is empty")
		}

		for _, subscriberID := range recipientIDs {
			subscriber, err := app.FindRecordById(nuvioSubscribersCollectionID, subscriberID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}
				return nil, err
			}

			if strings.TrimSpace(subscriber.GetString("website")) != websiteID {
				continue
			}

			if strings.ToLower(strings.TrimSpace(subscriber.GetString("status"))) != "active" {
				continue
			}

			email, ok := normalizeNuvioEmail(subscriber.GetString("email"))
			if ok {
				recipients[email] = struct{}{}
			}
		}
	default:
		subscribersCollection, err := app.FindCachedCollectionByNameOrId(nuvioSubscribersCollectionID)
		if err != nil {
			return nil, err
		}

		subscribers, err := app.FindRecordsByFilter(
			subscribersCollection,
			"website={:website} && status='active'",
			"-created",
			5000,
			0,
			dbx.Params{
				"website": websiteID,
			},
		)
		if err != nil {
			return nil, err
		}

		for _, subscriber := range subscribers {
			email, ok := normalizeNuvioEmail(subscriber.GetString("email"))
			if ok {
				recipients[email] = struct{}{}
			}
		}
	}

	result := make([]string, 0, len(recipients))
	for email := range recipients {
		result = append(result, email)
	}

	return result, nil
}

func parseNuvioRecipientIDs(raw any) []string {
	result := []string{}

	switch typed := raw.(type) {
	case []any:
		for _, item := range typed {
			value := strings.TrimSpace(parseStringValue(item))
			if value != "" {
				result = append(result, value)
			}
		}
	case []string:
		for _, item := range typed {
			value := strings.TrimSpace(item)
			if value != "" {
				result = append(result, value)
			}
		}
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return result
		}

		if strings.HasPrefix(trimmed, "[") {
			parsed := []string{}
			if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
				for _, item := range parsed {
					value := strings.TrimSpace(item)
					if value != "" {
						result = append(result, value)
					}
				}
			}
		}
	}

	return result
}

func normalizeNuvioEmail(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}

	parsed, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", false
	}

	email := strings.ToLower(strings.TrimSpace(parsed.Address))
	if email == "" {
		return "", false
	}

	return email, true
}

func sendNuvioCampaignEmailViaResend(
	ctx context.Context,
	apiKey string,
	from string,
	subject string,
	body string,
	recipients []string,
) error {
	requestPayload := resendSendEmailRequest{
		From:    from,
		To:      []string{from},
		Bcc:     recipients,
		Subject: subject,
	}

	trimmedBody := strings.TrimSpace(body)
	if strings.Contains(trimmedBody, "<") && strings.Contains(trimmedBody, ">") {
		requestPayload.Html = trimmedBody
	} else {
		requestPayload.Text = trimmedBody
	}

	rawPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return fmt.Errorf("Failed to encode Resend payload: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.resend.com/emails",
		bytes.NewBuffer(rawPayload),
	)
	if err != nil {
		return fmt.Errorf("Failed to build Resend request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+apiKey)

	response, err := nuvioNewsletterHTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("Failed to send campaign via Resend: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("Failed reading Resend response: %w", err)
	}

	if response.StatusCode >= 400 {
		message := strings.TrimSpace(string(responseBody))

		parsed := map[string]any{}
		if err := json.Unmarshal(responseBody, &parsed); err == nil {
			if parsedMessage, ok := parsed["message"].(string); ok && strings.TrimSpace(parsedMessage) != "" {
				message = strings.TrimSpace(parsedMessage)
			}
		}

		if message == "" {
			message = "Unknown Resend error"
		}

		return fmt.Errorf("Resend rejected campaign send (%d): %s", response.StatusCode, message)
	}

	decoded := resendSendEmailResponse{}
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return fmt.Errorf("Failed to decode Resend success response: %w", err)
	}

	if strings.TrimSpace(decoded.ID) == "" {
		return fmt.Errorf("Resend response missing message id")
	}

	return nil
}

// NUVIO CUSTOM END: Newsletter V1 send endpoint (server-side campaign dispatch via Resend).
