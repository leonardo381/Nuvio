package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	nuvioSubscribersCollectionID      = "pbc_1661203400"
	nuvioCampaignsCollectionID        = "pbc_1661203500"
	nuvioNewsletterStatusPending      = "pending"
	nuvioNewsletterStatusActive       = "active"
	nuvioNewsletterStatusUnsubscribed = "unsubscribed"
	nuvioNewsletterTokenBytes         = 32
	nuvioNewsletterConfirmationTTL    = 72 * time.Hour
	nuvioNewsletterMaxSubscriberScan  = 5000
	nuvioNewsletterMaxNameLen         = 200
	nuvioNewsletterPublicBaseURLEnv   = "NUVIO_PUBLIC_BASE_URL"
	nuvioNewsletterCampaignBatchSize  = 100
)

type nuvioWebsiteNewsletterConfig struct {
	FeatureAvailable bool
	DoubleOptIn      bool
}

type nuvioCampaignSendResult struct {
	CampaignID      string `json:"campaignId"`
	Status          string `json:"status"`
	RecipientsCount int    `json:"recipientsCount"`
	SentCount       int    `json:"sentCount"`
	FailedCount     int    `json:"failedCount"`
	SentAt          string `json:"sentAt"`
}

type nuvioNewsletterSubscribePayload struct {
	WebsiteID string `json:"websiteId"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

type nuvioCampaignRecipient struct {
	Subscriber *core.Record
	Email      string
}

type nuvioPreparedCampaignRecipient struct {
	SubscriberID                  string
	Email                         string
	PreviousUnsubscribeTokenHash  string
	RotatedUnsubscribeTokenStored bool
	Message                       nuvioTransactionalEmailMessage
}

// NUVIO CUSTOM START: Newsletter V1 send endpoint (server-side campaign dispatch via Resend).
func registerNuvioNewsletterRoutes(e *core.ServeEvent) {
	newsletterPublicGroup := e.Router.Group("/api/nuvio/newsletter")
	newsletterAdminGroup := e.Router.Group("/api/nuvio/newsletter").Bind(apis.RequireSuperuserAuth())

	newsletterPublicGroup.POST("/subscribe", func(e *core.RequestEvent) error {
		payload := nuvioNewsletterSubscribePayload{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid newsletter subscribe payload.", nil)
		}

		websiteID := strings.TrimSpace(payload.WebsiteID)
		if websiteID == "" {
			websiteID = strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		}
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}

		website, config, err := loadNuvioWebsiteNewsletterConfig(e.App, websiteID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Website not found.", nil)
			}

			return e.BadRequestError("Failed to load Newsletter settings.", nil)
		}

		if !config.FeatureAvailable {
			return e.BadRequestError("Newsletter is unavailable for this website.", nil)
		}

		email, ok := normalizeNuvioEmail(payload.Email)
		if !ok {
			return e.BadRequestError("A valid email is required.", nil)
		}

		name := sanitizeNuvioNewsletterName(payload.Name)
		subscribersCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioSubscribersCollectionID)
		if err != nil {
			return e.InternalServerError("Newsletter subscribers collection is missing.", nil)
		}

		subscriber, err := findNuvioSubscriberByWebsiteEmail(
			e.App,
			subscribersCollection,
			websiteID,
			email,
		)
		isNewSubscriber := false
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return e.InternalServerError("Failed to load newsletter subscriber.", nil)
			}

			subscriber = core.NewRecord(subscribersCollection)
			isNewSubscriber = true
		}

		subscriber.Set("website", websiteID)
		subscriber.Set("email", email)
		if name != "" || isNewSubscriber {
			subscriber.Set("name", name)
		}

		if err := ensureNuvioSubscriberUnsubscribeTokenHash(subscriber); err != nil {
			return e.InternalServerError("Failed to prepare subscriber lifecycle token.", nil)
		}

		now := time.Now().UTC()
		nowISO := now.Format(time.RFC3339)
		currentStatus := normalizeNuvioSubscriberStatus(subscriber.GetString("status"))
		if currentStatus == "" {
			currentStatus = nuvioNewsletterStatusPending
		}

		if !config.DoubleOptIn || currentStatus == nuvioNewsletterStatusActive {
			if strings.TrimSpace(subscriber.GetString("confirmedAt")) == "" {
				subscriber.Set("confirmedAt", nowISO)
			}
			subscriber.Set("status", nuvioNewsletterStatusActive)
			subscriber.Set("confirmationTokenHash", "")
			subscriber.Set("confirmationTokenExpiresAt", "")
			subscriber.Set("unsubscribedAt", "")

			if err := e.App.Save(subscriber); err != nil {
				return e.InternalServerError("Failed to save newsletter subscriber.", nil)
			}

			return e.JSON(http.StatusOK, map[string]any{
				"ok":          true,
				"status":      nuvioNewsletterStatusActive,
				"doubleOptIn": config.DoubleOptIn,
			})
		}

		rawToken, tokenHash, err := generateNuvioNewsletterTokenPair()
		if err != nil {
			return e.InternalServerError("Failed to generate confirmation token.", nil)
		}

		expiresAt := now.Add(nuvioNewsletterConfirmationTTL).UTC()
		subscriber.Set("status", nuvioNewsletterStatusPending)
		subscriber.Set("confirmedAt", "")
		subscriber.Set("confirmationTokenHash", tokenHash)
		subscriber.Set("confirmationTokenExpiresAt", expiresAt.Format(time.RFC3339))
		subscriber.Set("unsubscribedAt", "")

		if err := e.App.Save(subscriber); err != nil {
			return e.InternalServerError("Failed to save newsletter subscriber.", nil)
		}

		baseURL := resolveNuvioNewsletterPublicBaseURL(e.Request)
		confirmPath := buildNuvioNewsletterConfirmPath(website)
		confirmURL, err := buildNuvioNewsletterLifecycleURL(
			baseURL,
			confirmPath,
			rawToken,
		)
		if err != nil {
			return e.InternalServerError("Unable to prepare confirmation link right now.", nil)
		}

		if err := sendNuvioNewsletterConfirmationEmail(
			e.Request.Context(),
			website,
			email,
			confirmURL,
		); err != nil {
			e.App.Logger().Error(
				"NUVIO newsletter confirmation email send failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.InternalServerError("Unable to send confirmation email right now. Please try again.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":          true,
			"status":      nuvioNewsletterStatusPending,
			"doubleOptIn": true,
		})
	})

	newsletterPublicGroup.GET("/confirm", func(e *core.RequestEvent) error {
		rawToken := strings.TrimSpace(e.Request.URL.Query().Get("token"))
		if rawToken == "" {
			return e.BadRequestError("Missing confirmation token.", nil)
		}

		tokenHash := hashNuvioNewsletterToken(rawToken)
		if tokenHash == "" {
			return e.BadRequestError("Invalid or expired confirmation link.", nil)
		}

		subscribersCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioSubscribersCollectionID)
		if err != nil {
			return e.InternalServerError("Newsletter subscribers collection is missing.", nil)
		}

		subscriber, err := findNuvioSubscriberByConfirmationTokenHash(
			e.App,
			subscribersCollection,
			tokenHash,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.BadRequestError("Invalid or expired confirmation link.", nil)
			}

			return e.InternalServerError("Failed to validate confirmation token.", nil)
		}

		expiresRaw := strings.TrimSpace(subscriber.GetString("confirmationTokenExpiresAt"))
		expiresAt, ok := parseNuvioNewsletterDateTime(expiresRaw)
		if !ok || time.Now().UTC().After(expiresAt) {
			subscriber.Set("confirmationTokenHash", "")
			subscriber.Set("confirmationTokenExpiresAt", "")
			if saveErr := e.App.Save(subscriber); saveErr != nil {
				e.App.Logger().Warn(
					"NUVIO newsletter expired token cleanup failed",
					"subscriberId",
					subscriber.Id,
					"error",
					saveErr.Error(),
				)
			}
			return e.BadRequestError("Invalid or expired confirmation link.", nil)
		}

		if err := ensureNuvioSubscriberUnsubscribeTokenHash(subscriber); err != nil {
			return e.InternalServerError("Failed to update subscriber lifecycle token.", nil)
		}

		nowISO := time.Now().UTC().Format(time.RFC3339)
		subscriber.Set("status", nuvioNewsletterStatusActive)
		subscriber.Set("confirmedAt", nowISO)
		subscriber.Set("confirmationTokenHash", "")
		subscriber.Set("confirmationTokenExpiresAt", "")
		subscriber.Set("unsubscribedAt", "")

		if err := e.App.Save(subscriber); err != nil {
			return e.InternalServerError("Failed to confirm newsletter subscription.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":      true,
			"status":  nuvioNewsletterStatusActive,
			"message": "Subscription confirmed.",
		})
	})

	newsletterPublicGroup.GET("/unsubscribe", func(e *core.RequestEvent) error {
		rawToken := strings.TrimSpace(e.Request.URL.Query().Get("token"))
		if rawToken == "" {
			return e.BadRequestError("Missing unsubscribe token.", nil)
		}

		tokenHash := hashNuvioNewsletterToken(rawToken)
		if tokenHash == "" {
			return e.BadRequestError("Invalid unsubscribe link.", nil)
		}

		subscribersCollection, err := e.App.FindCachedCollectionByNameOrId(nuvioSubscribersCollectionID)
		if err != nil {
			return e.InternalServerError("Newsletter subscribers collection is missing.", nil)
		}

		subscriber, err := findNuvioSubscriberByUnsubscribeTokenHash(
			e.App,
			subscribersCollection,
			tokenHash,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.BadRequestError("Invalid unsubscribe link.", nil)
			}

			return e.InternalServerError("Failed to validate unsubscribe token.", nil)
		}

		nowISO := time.Now().UTC().Format(time.RFC3339)
		if normalizeNuvioSubscriberStatus(subscriber.GetString("status")) == nuvioNewsletterStatusUnsubscribed {
			if strings.TrimSpace(subscriber.GetString("unsubscribedAt")) == "" {
				subscriber.Set("unsubscribedAt", nowISO)
				if err := e.App.Save(subscriber); err != nil {
					return e.InternalServerError("Failed to update newsletter subscriber.", nil)
				}
			}

			return e.JSON(http.StatusOK, map[string]any{
				"ok":                  true,
				"status":              nuvioNewsletterStatusUnsubscribed,
				"alreadyUnsubscribed": true,
				"message":             "You are already unsubscribed.",
			})
		}

		subscriber.Set("status", nuvioNewsletterStatusUnsubscribed)
		subscriber.Set("unsubscribedAt", nowISO)
		subscriber.Set("confirmationTokenHash", "")
		subscriber.Set("confirmationTokenExpiresAt", "")

		if err := e.App.Save(subscriber); err != nil {
			return e.InternalServerError("Failed to unsubscribe newsletter subscriber.", nil)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":      true,
			"status":  nuvioNewsletterStatusUnsubscribed,
			"message": "You have been unsubscribed.",
		})
	})

	newsletterAdminGroup.POST("/campaigns/send", func(e *core.RequestEvent) error {
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

		result, err := sendNuvioNewsletterCampaign(e.App, e.Request.Context(), campaignID, e.Request)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return e.NotFoundError("Campaign not found.", nil)
			}

			return e.BadRequestError(err.Error(), nil)
		}

		return e.JSON(http.StatusOK, result)
	})
}

func sendNuvioNewsletterCampaign(
	app core.App,
	ctx context.Context,
	campaignID string,
	request *http.Request,
) (*nuvioCampaignSendResult, error) {
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

	website, newsletterConfig, err := loadNuvioWebsiteNewsletterConfig(app, websiteID)
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

	baseURL := resolveNuvioNewsletterPublicBaseURL(request)
	unsubscribePath := buildNuvioNewsletterUnsubscribePath(website)
	websiteName := strings.TrimSpace(resolveWebsiteDisplayName(website))
	if websiteName == "" {
		websiteName = "Website"
	}

	sentCount := 0
	failedCount := 0
	preparedRecipients := make([]nuvioPreparedCampaignRecipient, 0, len(recipients))

	for _, recipient := range recipients {
		if recipient.Subscriber == nil {
			failedCount++
			app.Logger().Warn(
				"NUVIO newsletter campaign recipient missing subscriber record",
				"campaignId",
				campaignID,
				"recipientEmail",
				recipient.Email,
			)
			continue
		}

		previousTokenHash := strings.TrimSpace(recipient.Subscriber.GetString("unsubscribeTokenHash"))

		rawToken, tokenHash, err := generateNuvioNewsletterTokenPair()
		if err != nil {
			failedCount++
			app.Logger().Warn(
				"NUVIO newsletter unsubscribe token generation failed",
				"campaignId",
				campaignID,
				"subscriberId",
				recipient.Subscriber.Id,
				"error",
				err.Error(),
			)
			continue
		}

		unsubscribeURL, err := buildNuvioNewsletterLifecycleURL(baseURL, unsubscribePath, rawToken)
		if err != nil {
			failedCount++
			app.Logger().Warn(
				"NUVIO newsletter unsubscribe URL build failed",
				"campaignId",
				campaignID,
				"subscriberId",
				recipient.Subscriber.Id,
				"error",
				err.Error(),
			)
			continue
		}

		subscriber := recipient.Subscriber.Clone()
		subscriber.Set("unsubscribeTokenHash", tokenHash)
		if err := app.Save(subscriber); err != nil {
			failedCount++
			app.Logger().Warn(
				"NUVIO newsletter subscriber token hash save failed",
				"campaignId",
				campaignID,
				"subscriberId",
				recipient.Subscriber.Id,
				"error",
				err.Error(),
			)
			continue
		}

		preparedRecipients = append(preparedRecipients, nuvioPreparedCampaignRecipient{
			SubscriberID:                  recipient.Subscriber.Id,
			Email:                         recipient.Email,
			PreviousUnsubscribeTokenHash:  previousTokenHash,
			RotatedUnsubscribeTokenStored: true,
			Message: buildNuvioCampaignRecipientMessage(
				recipient.Email,
				subject,
				body,
				websiteName,
				unsubscribeURL,
			),
		})
	}

	if len(preparedRecipients) == 0 {
		return nil, fmt.Errorf("No active recipients could be prepared for sending")
	}

	for start := 0; start < len(preparedRecipients); start += nuvioNewsletterCampaignBatchSize {
		end := start + nuvioNewsletterCampaignBatchSize
		if end > len(preparedRecipients) {
			end = len(preparedRecipients)
		}

		chunk := preparedRecipients[start:end]
		chunkMessages := make([]nuvioTransactionalEmailMessage, 0, len(chunk))
		for _, prepared := range chunk {
			chunkMessages = append(chunkMessages, prepared.Message)
		}

		chunkResult, chunkErr := sendNuvioTransactionalEmailBatchViaResend(ctx, resendConfig, chunkMessages)
		if chunkErr != nil {
			app.Logger().Warn(
				"NUVIO newsletter campaign batch chunk send failed",
				"campaignId",
				campaignID,
				"chunkStart",
				start,
				"chunkSize",
				len(chunk),
				"error",
				chunkErr.Error(),
			)

			failedCount += len(chunk)
			restoreNuvioCampaignChunkUnsubscribeHashes(app, campaignID, chunk)
			continue
		}

		sentCount += chunkResult.SentCount
		failedCount += chunkResult.FailedCount

		if chunkResult.FailedCount > 0 {
			restoreFailedIndexes := chunkResult.FailedIndexes
			if len(restoreFailedIndexes) == 0 || chunkResult.AmbiguousResult {
				restoreFailedIndexes = make([]int, 0, len(chunk))
				for i := range chunk {
					restoreFailedIndexes = append(restoreFailedIndexes, i)
				}
			}

			restoreNuvioCampaignChunkUnsubscribeHashesByIndexes(
				app,
				campaignID,
				chunk,
				restoreFailedIndexes,
			)
		}
	}

	if sentCount == 0 {
		return nil, fmt.Errorf("Newsletter campaign send failed for all recipients")
	}

	if failedCount > 0 {
		app.Logger().Warn(
			"NUVIO newsletter campaign partial send",
			"campaignId",
			campaignID,
			"sentCount",
			sentCount,
			"failedCount",
			failedCount,
		)
	}

	sentAt := time.Now().UTC().Format(time.RFC3339)
	campaign.Set("status", "sent")
	campaign.Set("sentAt", sentAt)
	campaign.Set("recipientsCount", sentCount)

	if err := app.Save(campaign); err != nil {
		return nil, err
	}

	return &nuvioCampaignSendResult{
		CampaignID:      campaign.Id,
		Status:          "sent",
		RecipientsCount: sentCount,
		SentCount:       sentCount,
		FailedCount:     failedCount,
		SentAt:          sentAt,
	}, nil
}

func loadNuvioWebsiteNewsletterConfig(app core.App, websiteID string) (*core.Record, nuvioWebsiteNewsletterConfig, error) {
	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil {
		return nil, nuvioWebsiteNewsletterConfig{}, err
	}

	settings := parseNuvioSettingsObject(website.Get("settings"))

	config := nuvioWebsiteNewsletterConfig{
		FeatureAvailable: true,
		DoubleOptIn:      false,
	}

	if featureFlags, ok := toStringAnyMap(settings["featureFlags"]); ok {
		if value, ok := parseBoolValue(featureFlags["newsletter"]); ok {
			config.FeatureAvailable = value
		}
	}

	if newsletterSettings, ok := toStringAnyMap(settings["newsletter"]); ok {
		if value, ok := parseBoolValue(newsletterSettings["doubleOptIn"]); ok {
			config.DoubleOptIn = value
		}
	}

	return website, config, nil
}

func findNuvioSubscriberByWebsiteEmail(
	app core.App,
	subscribersCollection *core.Collection,
	websiteID string,
	email string,
) (*core.Record, error) {
	subscriber, err := app.FindFirstRecordByFilter(
		subscribersCollection,
		"website={:website} && email={:email}",
		dbx.Params{
			"website": websiteID,
			"email":   email,
		},
	)
	if err == nil {
		return subscriber, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	records, err := app.FindRecordsByFilter(
		subscribersCollection,
		"website={:website}",
		"-updated,-created",
		nuvioNewsletterMaxSubscriberScan,
		0,
		dbx.Params{
			"website": websiteID,
		},
	)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		normalizedEmail, ok := normalizeNuvioEmail(record.GetString("email"))
		if ok && normalizedEmail == email {
			return record, nil
		}
	}

	return nil, sql.ErrNoRows
}

func findNuvioSubscriberByConfirmationTokenHash(
	app core.App,
	subscribersCollection *core.Collection,
	tokenHash string,
) (*core.Record, error) {
	return app.FindFirstRecordByFilter(
		subscribersCollection,
		"confirmationTokenHash={:tokenHash}",
		dbx.Params{
			"tokenHash": tokenHash,
		},
	)
}

func findNuvioSubscriberByUnsubscribeTokenHash(
	app core.App,
	subscribersCollection *core.Collection,
	tokenHash string,
) (*core.Record, error) {
	return app.FindFirstRecordByFilter(
		subscribersCollection,
		"unsubscribeTokenHash={:tokenHash}",
		dbx.Params{
			"tokenHash": tokenHash,
		},
	)
}

func generateNuvioNewsletterTokenPair() (string, string, error) {
	tokenBytes := make([]byte, nuvioNewsletterTokenBytes)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", err
	}

	rawToken := hex.EncodeToString(tokenBytes)
	if rawToken == "" {
		return "", "", fmt.Errorf("failed to generate token")
	}

	return rawToken, hashNuvioNewsletterToken(rawToken), nil
}

func hashNuvioNewsletterToken(rawToken string) string {
	trimmed := strings.TrimSpace(rawToken)
	if trimmed == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(trimmed))
	return hex.EncodeToString(sum[:])
}

func ensureNuvioSubscriberUnsubscribeTokenHash(subscriber *core.Record) error {
	if subscriber == nil {
		return fmt.Errorf("subscriber is nil")
	}

	if strings.TrimSpace(subscriber.GetString("unsubscribeTokenHash")) != "" {
		return nil
	}

	_, tokenHash, err := generateNuvioNewsletterTokenPair()
	if err != nil {
		return err
	}

	subscriber.Set("unsubscribeTokenHash", tokenHash)
	return nil
}

func normalizeNuvioSubscriberStatus(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case nuvioNewsletterStatusPending, nuvioNewsletterStatusActive, nuvioNewsletterStatusUnsubscribed:
		return normalized
	default:
		return ""
	}
}

func parseNuvioNewsletterDateTime(raw string) (time.Time, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.000Z",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed.UTC(), true
		}
	}

	return time.Time{}, false
}

func resolveNuvioNewsletterPublicBaseURL(request *http.Request) string {
	if envBaseURL := resolveNuvioNewsletterPublicBaseURLFromEnv(); envBaseURL != "" {
		return envBaseURL
	}

	return resolveNuvioRequestBaseURL(request)
}

func resolveNuvioNewsletterPublicBaseURLFromEnv() string {
	rawValue := strings.TrimSpace(os.Getenv(nuvioNewsletterPublicBaseURLEnv))
	if rawValue == "" {
		return ""
	}

	return normalizeNuvioNewsletterBaseURL(rawValue)
}

func resolveNuvioRequestBaseURL(request *http.Request) string {
	if request == nil {
		return ""
	}

	host := normalizeNuvioRequestHost(firstNuvioHeaderValue(request.Header.Get("X-Forwarded-Host")))
	if host == "" {
		host = normalizeNuvioRequestHost(request.Host)
	}
	if host == "" {
		return ""
	}

	scheme := normalizeNuvioRequestScheme(firstNuvioHeaderValue(request.Header.Get("X-Forwarded-Proto")))
	if scheme == "" {
		if request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	if scheme == "" {
		scheme = "http"
	}

	return normalizeNuvioNewsletterBaseURL(scheme + "://" + host)
}

func firstNuvioHeaderValue(raw string) string {
	parts := strings.Split(raw, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeNuvioRequestScheme(raw string) string {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch normalized {
	case "http", "https":
		return normalized
	default:
		return ""
	}
}

func normalizeNuvioRequestHost(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	if strings.ContainsAny(trimmed, "/?#\\") {
		return ""
	}

	parsed, err := url.Parse("//" + trimmed)
	if err != nil {
		return ""
	}

	host := strings.TrimSpace(parsed.Host)
	if host == "" {
		return ""
	}

	return host
}

func normalizeNuvioNewsletterBaseURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil || !parsed.IsAbs() || strings.TrimSpace(parsed.Host) == "" {
		return ""
	}

	scheme := strings.ToLower(strings.TrimSpace(parsed.Scheme))
	if scheme != "http" && scheme != "https" {
		return ""
	}

	parsed.Scheme = scheme
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""

	normalizedPath := strings.TrimRight(parsed.Path, "/")
	if normalizedPath == "/" {
		normalizedPath = ""
	}
	parsed.Path = normalizedPath
	parsed.RawPath = ""

	return parsed.String()
}

func buildNuvioNewsletterConfirmPath(website *core.Record) string {
	if website == nil {
		return "/api/nuvio/newsletter/confirm"
	}

	slug := strings.TrimSpace(website.GetString("slug"))
	if slug == "" {
		return "/api/nuvio/newsletter/confirm"
	}

	return "/site/" + url.PathEscape(slug) + "/newsletter/confirm"
}

func buildNuvioNewsletterLifecycleURL(baseURL string, path string, rawToken string) (string, error) {
	trimmedBase := normalizeNuvioNewsletterBaseURL(baseURL)
	trimmedToken := strings.TrimSpace(rawToken)
	if trimmedBase == "" || trimmedToken == "" {
		return "", fmt.Errorf("missing base URL or token")
	}

	parsedBase, err := url.Parse(trimmedBase)
	if err != nil || parsedBase.Scheme == "" || parsedBase.Host == "" {
		return "", fmt.Errorf("invalid base URL")
	}

	query := url.Values{}
	query.Set("token", trimmedToken)

	ref := &url.URL{
		Path:     path,
		RawQuery: query.Encode(),
	}

	return parsedBase.ResolveReference(ref).String(), nil
}

func sanitizeNuvioNewsletterName(raw string) string {
	normalized := strings.NewReplacer("\r", " ", "\n", " ").Replace(raw)
	normalized = strings.Join(strings.Fields(normalized), " ")
	normalized = strings.TrimSpace(normalized)

	return truncateNuvioNewsletterStringByRunes(normalized, nuvioNewsletterMaxNameLen)
}

func truncateNuvioNewsletterStringByRunes(raw string, maxLen int) string {
	if maxLen <= 0 || raw == "" {
		return ""
	}

	runes := []rune(raw)
	if len(runes) <= maxLen {
		return raw
	}

	return string(runes[:maxLen])
}

func sendNuvioNewsletterConfirmationEmail(
	ctx context.Context,
	website *core.Record,
	recipientEmail string,
	confirmURL string,
) error {
	resendConfig, err := loadNuvioResendConfig()
	if err != nil {
		return err
	}

	normalizedEmail, ok := normalizeNuvioEmail(recipientEmail)
	if !ok {
		return fmt.Errorf("invalid subscriber email")
	}

	websiteName := strings.TrimSpace(resolveWebsiteDisplayName(website))
	if websiteName == "" {
		websiteName = "Website"
	}

	subject := "Confirm your subscription"
	textBody := strings.Join([]string{
		fmt.Sprintf("You requested a newsletter subscription for %s.", websiteName),
		"",
		"Please confirm your subscription by opening this link:",
		confirmURL,
		"",
		"If you did not request this, you can safely ignore this email.",
	}, "\n")

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      []string{normalizedEmail},
		Subject: subject,
		Text:    textBody,
	})
}

func resolveNuvioCampaignRecipients(
	app core.App,
	websiteID string,
	recipientsType string,
	rawRecipientsIDs any,
) ([]nuvioCampaignRecipient, error) {
	recipientsByEmail := map[string]nuvioCampaignRecipient{}

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

			if normalizeNuvioSubscriberStatus(subscriber.GetString("status")) != nuvioNewsletterStatusActive {
				continue
			}

			email, ok := normalizeNuvioEmail(subscriber.GetString("email"))
			if !ok {
				continue
			}

			if _, exists := recipientsByEmail[email]; exists {
				continue
			}

			recipientsByEmail[email] = nuvioCampaignRecipient{
				Subscriber: subscriber,
				Email:      email,
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
			nuvioNewsletterMaxSubscriberScan,
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
			if !ok {
				continue
			}

			if _, exists := recipientsByEmail[email]; exists {
				continue
			}

			recipientsByEmail[email] = nuvioCampaignRecipient{
				Subscriber: subscriber,
				Email:      email,
			}
		}
	}

	result := make([]nuvioCampaignRecipient, 0, len(recipientsByEmail))
	for _, recipient := range recipientsByEmail {
		result = append(result, recipient)
	}

	return result, nil
}

func restoreNuvioCampaignChunkUnsubscribeHashes(
	app core.App,
	campaignID string,
	chunk []nuvioPreparedCampaignRecipient,
) {
	indexes := make([]int, 0, len(chunk))
	for i := range chunk {
		indexes = append(indexes, i)
	}

	restoreNuvioCampaignChunkUnsubscribeHashesByIndexes(app, campaignID, chunk, indexes)
}

func restoreNuvioCampaignChunkUnsubscribeHashesByIndexes(
	app core.App,
	campaignID string,
	chunk []nuvioPreparedCampaignRecipient,
	indexes []int,
) {
	seen := map[int]struct{}{}
	for _, idx := range indexes {
		if idx < 0 || idx >= len(chunk) {
			app.Logger().Warn(
				"NUVIO newsletter campaign unsubscribe hash restore index out of range",
				"campaignId",
				campaignID,
				"index",
				idx,
				"chunkSize",
				len(chunk),
			)
			continue
		}

		if _, exists := seen[idx]; exists {
			continue
		}
		seen[idx] = struct{}{}

		prepared := chunk[idx]
		if !prepared.RotatedUnsubscribeTokenStored {
			continue
		}

		if err := restoreNuvioCampaignRecipientUnsubscribeHash(app, prepared); err != nil {
			app.Logger().Warn(
				"NUVIO newsletter campaign unsubscribe hash restore failed",
				"campaignId",
				campaignID,
				"subscriberId",
				prepared.SubscriberID,
				"recipientEmail",
				prepared.Email,
				"error",
				err.Error(),
			)
		}
	}
}

func restoreNuvioCampaignRecipientUnsubscribeHash(
	app core.App,
	prepared nuvioPreparedCampaignRecipient,
) error {
	subscriberID := strings.TrimSpace(prepared.SubscriberID)
	if subscriberID == "" {
		return fmt.Errorf("missing subscriber id")
	}

	subscriber, err := app.FindRecordById(nuvioSubscribersCollectionID, subscriberID)
	if err != nil {
		return err
	}

	subscriber.Set("unsubscribeTokenHash", strings.TrimSpace(prepared.PreviousUnsubscribeTokenHash))
	return app.Save(subscriber)
}

func parseNuvioRecipientIDs(raw any) []string {
	result := []string{}

	switch typed := raw.(type) {
	case []any:
		for _, item := range typed {
			result = append(result, parseNuvioRecipientIDs(item)...)
		}
	case []string:
		for _, item := range typed {
			value := strings.TrimSpace(item)
			if value != "" {
				result = append(result, value)
			}
		}
	case types.JSONRaw:
		result = append(result, parseNuvioRecipientIDs([]byte(typed))...)
	case []byte:
		trimmed := strings.TrimSpace(string(typed))
		if trimmed == "" {
			return result
		}

		decoded := any(nil)
		if err := json.Unmarshal(typed, &decoded); err != nil {
			return result
		}

		result = append(result, parseNuvioRecipientIDs(decoded)...)
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return result
		}

		if strings.HasPrefix(trimmed, "[") {
			decoded := any(nil)
			if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
				result = append(result, parseNuvioRecipientIDs(decoded)...)
				return result
			}
		} else {
			result = append(result, trimmed)
		}
	default:
		value := strings.TrimSpace(parseStringValue(typed))
		if value != "" {
			result = append(result, value)
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

func buildNuvioNewsletterUnsubscribePath(website *core.Record) string {
	if website == nil {
		return "/api/nuvio/newsletter/unsubscribe"
	}

	slug := strings.TrimSpace(website.GetString("slug"))
	if slug == "" {
		return "/api/nuvio/newsletter/unsubscribe"
	}

	return "/site/" + url.PathEscape(slug) + "/newsletter/unsubscribe"
}

func buildNuvioCampaignRecipientMessage(
	recipientEmail string,
	subject string,
	body string,
	websiteName string,
	unsubscribeURL string,
) nuvioTransactionalEmailMessage {
	trimmedBody := strings.TrimSpace(body)
	footerText := buildNuvioCampaignUnsubscribeFooterText(websiteName, unsubscribeURL)

	message := nuvioTransactionalEmailMessage{
		To:      []string{recipientEmail},
		Subject: subject,
	}

	if strings.Contains(trimmedBody, "<") && strings.Contains(trimmedBody, ">") {
		htmlFooter := buildNuvioCampaignUnsubscribeFooterHTML(websiteName, unsubscribeURL)
		if trimmedBody == "" {
			message.HTML = htmlFooter
		} else {
			message.HTML = trimmedBody + "\n\n" + htmlFooter
		}
		return message
	}

	if trimmedBody == "" {
		message.Text = footerText
	} else {
		message.Text = trimmedBody + "\n\n" + footerText
	}

	return message
}

func buildNuvioCampaignUnsubscribeFooterText(websiteName string, unsubscribeURL string) string {
	safeWebsiteName := strings.TrimSpace(websiteName)
	if safeWebsiteName == "" {
		safeWebsiteName = "Website"
	}

	safeURL := strings.TrimSpace(unsubscribeURL)
	return strings.Join([]string{
		"---",
		fmt.Sprintf("You are receiving this email because you subscribed to updates from %s.", safeWebsiteName),
		fmt.Sprintf("Unsubscribe: %s", safeURL),
	}, "\n")
}

func buildNuvioCampaignUnsubscribeFooterHTML(websiteName string, unsubscribeURL string) string {
	safeWebsiteName := strings.TrimSpace(websiteName)
	if safeWebsiteName == "" {
		safeWebsiteName = "Website"
	}

	escapedWebsiteName := html.EscapeString(safeWebsiteName)
	escapedURL := html.EscapeString(strings.TrimSpace(unsubscribeURL))
	return strings.Join([]string{
		"<hr>",
		"<p>You are receiving this email because you subscribed to updates from " + escapedWebsiteName + ".</p>",
		`<p>Unsubscribe: <a href="` + escapedURL + `">` + escapedURL + "</a></p>",
	}, "")
}

// NUVIO CUSTOM END: Newsletter V1 send endpoint (server-side campaign dispatch via Resend).
