package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
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

type nuvioWebsiteNewsletterConfig struct {
	FeatureAvailable bool
}

type nuvioCampaignSendResult struct {
	CampaignID      string `json:"campaignId"`
	Status          string `json:"status"`
	RecipientsCount int    `json:"recipientsCount"`
	SentAt          string `json:"sentAt"`
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

	resendConfig, err := loadNuvioResendConfig()
	if err != nil {
		return nil, err
	}

	if err := sendNuvioCampaignEmailViaResend(ctx, resendConfig, subject, body, recipients); err != nil {
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
	config nuvioResendConfig,
	subject string,
	body string,
	recipients []string,
) error {
	trimmedBody := strings.TrimSpace(body)

	message := nuvioTransactionalEmailMessage{
		To:      []string{config.From},
		Bcc:     recipients,
		Subject: subject,
	}

	if strings.Contains(trimmedBody, "<") && strings.Contains(trimmedBody, ">") {
		message.HTML = trimmedBody
	} else {
		message.Text = trimmedBody
	}

	return sendNuvioTransactionalEmailViaResend(ctx, config, message)
}

// NUVIO CUSTOM END: Newsletter V1 send endpoint (server-side campaign dispatch via Resend).
