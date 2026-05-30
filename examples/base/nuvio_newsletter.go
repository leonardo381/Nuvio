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
	"regexp"
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
	nuvioSubscriberGroupsCollectionID = "pbc_1661203600"
	nuvioNewsletterStatusPending      = "pending"
	nuvioNewsletterStatusActive       = "active"
	nuvioNewsletterStatusUnsubscribed = "unsubscribed"
	nuvioNewsletterTokenBytes         = 32
	nuvioNewsletterConfirmationTTL    = 72 * time.Hour
	nuvioNewsletterMaxSubscriberScan  = 5000
	nuvioNewsletterMaxNameLen         = 200
	nuvioNewsletterTemplateSubjectMax = 160
	nuvioNewsletterTemplateTextMax    = 4000
	nuvioNewsletterPublicBaseURLEnv   = "NUVIO_PUBLIC_BASE_URL"
	nuvioNewsletterCampaignBatchSize  = 100
	nuvioNewsletterBackofficeNameMax  = 200
	nuvioNewsletterBackofficeSlugMax  = 120
	nuvioNewsletterPublicEmailMaxLen  = 320
	nuvioNewsletterPublicTokenMaxLen  = 512
)

var (
	nuvioNewsletterBackofficeSubscribersCollectionAliases = []string{
		nuvioSubscribersCollectionID,
		"Subscribers",
		"subscribers",
	}
	nuvioNewsletterBackofficeCampaignsCollectionAliases = []string{
		nuvioCampaignsCollectionID,
		"Campaigns",
		"campaigns",
	}
	nuvioNewsletterBackofficeGroupsCollectionAliases = []string{
		nuvioSubscriberGroupsCollectionID,
		"SubscriberGroups",
		"subscribergroups",
		"subscriber_groups",
	}
	nuvioNewsletterBackofficeSubscribersSourceFieldAliases = []string{
		"source",
		"sourceLabel",
		"source_label",
		"origin",
	}
	nuvioNewsletterBackofficeSubscribersGroupsFieldAliases = []string{
		"groups",
		"groupIds",
		"subscriberGroups",
		"subscriber_groups",
	}
	nuvioNewsletterBackofficeCampaignRecipientsTypeFieldAliases = []string{
		"recipientsType",
		"recipientType",
		"recipients_type",
	}
	nuvioNewsletterBackofficeCampaignRecipientsIDsFieldAliases = []string{
		"recipientsIds",
		"recipientIds",
		"recipients_ids",
	}
	nuvioNewsletterBackofficeAllowedSubscriberStatus = []string{
		nuvioNewsletterStatusPending,
		nuvioNewsletterStatusActive,
		nuvioNewsletterStatusUnsubscribed,
	}
	nuvioNewsletterBackofficeAllowedCampaignStatus = []string{
		"draft",
		"sent",
	}
	nuvioNewsletterBackofficeAllowedRecipientsType = []string{
		"all",
		"manual",
	}
	nuvioNewsletterBackofficeAllowedCampaignCreateStatus = []string{
		"draft",
	}
	nuvioNewsletterBackofficeAllowedCampaignUpdateStatus = []string{
		"draft",
	}
	nuvioNewsletterBackofficeWebsiteFieldAliases = []string{
		"website",
		"site",
	}
	nuvioNewsletterBackofficeSubscriberEmailFieldAliases = []string{
		"email",
	}
	nuvioNewsletterBackofficeSubscriberNameFieldAliases = []string{
		"name",
	}
	nuvioNewsletterBackofficeSubscriberStatusFieldAliases = []string{
		"status",
	}
	nuvioNewsletterBackofficeGroupNameFieldAliases = []string{
		"name",
	}
	nuvioNewsletterBackofficeGroupSlugFieldAliases = []string{
		"slug",
	}
	nuvioNewsletterBackofficeCampaignSubjectFieldAliases = []string{
		"subject",
	}
	nuvioNewsletterBackofficeCampaignBodyFieldAliases = []string{
		"body",
	}
	nuvioNewsletterBackofficeCampaignStatusFieldAliases = []string{
		"status",
	}
	nuvioNewsletterBackofficeCampaignRecipientsCountFieldAliases = []string{
		"recipientsCount",
		"recipientCount",
		"recipients_count",
	}
	nuvioNewsletterBackofficeCampaignSentAtFieldAliases = []string{
		"sentAt",
		"sent_at",
	}
	nuvioNewsletterBackofficeLifecycleFieldAliases = map[string]struct{}{
		"confirmationtokenhash":      {},
		"confirmationtokenexpiresat": {},
		"unsubscribetokenhash":       {},
	}
	nuvioNewsletterBackofficeSubscriberCreateAllowedPayloadKeys = map[string]struct{}{
		"websiteid": {},
		"email":     {},
		"name":      {},
		"status":    {},
		"source":    {},
		"groups":    {},
	}
	nuvioNewsletterBackofficeSubscriberUpdateAllowedPayloadKeys = map[string]struct{}{
		"email":  {},
		"name":   {},
		"status": {},
		"source": {},
		"groups": {},
	}
	nuvioNewsletterBackofficeGroupCreateAllowedPayloadKeys = map[string]struct{}{
		"websiteid": {},
		"name":      {},
		"slug":      {},
	}
	nuvioNewsletterBackofficeCampaignCreateAllowedPayloadKeys = map[string]struct{}{
		"websiteid":      {},
		"subject":        {},
		"body":           {},
		"status":         {},
		"recipientstype": {},
		"recipientsids":  {},
	}
	nuvioNewsletterBackofficeCampaignUpdateAllowedPayloadKeys = map[string]struct{}{
		"subject":        {},
		"body":           {},
		"status":         {},
		"recipientstype": {},
		"recipientsids":  {},
	}
	nuvioNewsletterBackofficeSlugUnsafeCharsRegex = regexp.MustCompile(`[^a-z0-9-]+`)
	nuvioNewsletterBackofficeSlugMultiDashRegex   = regexp.MustCompile(`-+`)
)

type nuvioNewsletterConfirmationTemplateConfig struct {
	Enabled    bool
	Subject    string
	IntroText  string
	FooterText string
}

type nuvioWebsiteNewsletterConfig struct {
	FeatureAvailable     bool
	DoubleOptIn          bool
	ConfirmationTemplate nuvioNewsletterConfirmationTemplateConfig
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

type nuvioNewsletterInvitePayload struct {
	WebsiteID string `json:"websiteId"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Source    string `json:"source"`
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

type nuvioNewsletterBackofficeSubscriberDTO struct {
	ID             string   `json:"id"`
	Website        string   `json:"website"`
	Email          string   `json:"email"`
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	Source         string   `json:"source"`
	Groups         []string `json:"groups"`
	ConfirmedAt    string   `json:"confirmedAt"`
	UnsubscribedAt string   `json:"unsubscribedAt"`
	Created        string   `json:"created"`
	Updated        string   `json:"updated"`
}

type nuvioNewsletterBackofficeCampaignDTO struct {
	ID              string   `json:"id"`
	Website         string   `json:"website"`
	Subject         string   `json:"subject"`
	Body            string   `json:"body"`
	Status          string   `json:"status"`
	RecipientsType  string   `json:"recipientsType"`
	RecipientsIDs   []string `json:"recipientsIds"`
	RecipientsCount int      `json:"recipientsCount"`
	SentAt          string   `json:"sentAt"`
	Created         string   `json:"created"`
	Updated         string   `json:"updated"`
}

type nuvioNewsletterBackofficeGroupDTO struct {
	ID      string `json:"id"`
	Website string `json:"website"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

type nuvioNewsletterBackofficeDatasets struct {
	Subscribers []nuvioNewsletterBackofficeSubscriberDTO `json:"subscribers"`
	Campaigns   []nuvioNewsletterBackofficeCampaignDTO   `json:"campaigns"`
	Groups      []nuvioNewsletterBackofficeGroupDTO      `json:"groups"`
}

type nuvioNewsletterBackofficeSubscribersCapabilities struct {
	AllowedStatus  []string `json:"allowedStatus"`
	SupportsName   bool     `json:"supportsName"`
	SupportsSource bool     `json:"supportsSource"`
	SupportsGroups bool     `json:"supportsGroups"`
}

type nuvioNewsletterBackofficeCampaignCapabilities struct {
	AllowedStatus         []string `json:"allowedStatus"`
	AllowedRecipientsType []string `json:"allowedRecipientsType"`
}

type nuvioNewsletterBackofficeCapabilities struct {
	Subscribers nuvioNewsletterBackofficeSubscribersCapabilities `json:"subscribers"`
	Campaigns   nuvioNewsletterBackofficeCampaignCapabilities    `json:"campaigns"`
}

type nuvioNewsletterBackofficeDashboardResponse struct {
	State        string                                `json:"state"`
	WebsiteID    string                                `json:"websiteId"`
	Datasets     nuvioNewsletterBackofficeDatasets     `json:"datasets"`
	Capabilities nuvioNewsletterBackofficeCapabilities `json:"capabilities"`
}

// NUVIO CUSTOM START: Newsletter V1 send endpoint (server-side campaign dispatch via Resend).
func registerNuvioNewsletterRoutes(e *core.ServeEvent) {
	newsletterPublicGroup := e.Router.Group("/api/nuvio/newsletter")
	newsletterAdminGroup := e.Router.Group("/api/nuvio/newsletter").Bind(apis.RequireAdminSuperuserAuth())
	newsletterBackofficeGroup := e.Router.Group("/api/nuvio/newsletter/backoffice").Bind(apis.RequireSuperuserAuth())

	newsletterBackofficeGroup.GET("/dashboard", func(e *core.RequestEvent) error {
		websiteID := strings.TrimSpace(e.Request.URL.Query().Get("websiteId"))
		if websiteID == "" {
			return e.BadRequestError("Missing websiteId.", nil)
		}
		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}

		datasets, capabilities, err := loadNuvioNewsletterBackofficeDashboard(e.App, websiteID)
		if err != nil {
			e.App.Logger().Error(
				"NUVIO newsletter backoffice dashboard load failed",
				"websiteId",
				websiteID,
				"error",
				err.Error(),
			)
			return e.BadRequestError("Failed to load Newsletter dashboard data.", nil)
		}

		return e.JSON(http.StatusOK, nuvioNewsletterBackofficeDashboardResponse{
			State:        "ok",
			WebsiteID:    websiteID,
			Datasets:     datasets,
			Capabilities: capabilities,
		})
	})

	newsletterBackofficeGroup.POST("/subscribers", handleNuvioNewsletterBackofficeSubscriberCreate)
	newsletterBackofficeGroup.PATCH("/subscribers/{id}", handleNuvioNewsletterBackofficeSubscriberUpdate)
	newsletterBackofficeGroup.DELETE("/subscribers/{id}", handleNuvioNewsletterBackofficeSubscriberDelete)
	newsletterBackofficeGroup.POST("/subscribers/{id}/invite", handleNuvioNewsletterBackofficeSubscriberInvite)
	newsletterBackofficeGroup.POST("/groups", handleNuvioNewsletterBackofficeGroupCreate)
	newsletterBackofficeGroup.POST("/invite", func(e *core.RequestEvent) error {
		return handleNuvioNewsletterInvite(e, true)
	})
	newsletterBackofficeGroup.POST("/campaigns", handleNuvioNewsletterBackofficeCampaignCreate)
	newsletterBackofficeGroup.PATCH("/campaigns/{id}", handleNuvioNewsletterBackofficeCampaignUpdate)
	newsletterBackofficeGroup.DELETE("/campaigns/{id}", handleNuvioNewsletterBackofficeCampaignDelete)
	newsletterBackofficeGroup.POST("/campaigns/{id}/duplicate", handleNuvioNewsletterBackofficeCampaignDuplicate)
	newsletterBackofficeGroup.POST("/campaigns/{id}/send", handleNuvioNewsletterBackofficeCampaignSend)

	newsletterPublicGroup.POST("/subscribe", func(e *core.RequestEvent) error {
		payload := map[string]any{}
		if err := e.BindBody(&payload); err != nil {
			return e.BadRequestError("Invalid newsletter subscribe payload.", nil)
		}

		if err := validateNuvioPublicPayloadKeys(
			payload,
			map[string]struct{}{
				"websiteId":   {},
				"website":     {},
				"websiteSlug": {},
				"slug":        {},
				"email":       {},
				"name":        {},
			},
		); err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		_, websiteID, err := resolveNuvioPublicWebsiteFromPayload(e.App, payload, e.Request.URL.Query())
		if err != nil {
			return handleNuvioPublicWebsiteResolveError(e, err)
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

		emailRaw, err := validateNuvioPublicRequiredField(payload["email"], "Email", nuvioNewsletterPublicEmailMaxLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}

		email, ok := normalizeNuvioEmail(emailRaw)
		if !ok {
			return e.BadRequestError("A valid email is required.", nil)
		}

		name, err := validateNuvioPublicOptionalField(payload["name"], "Name", nuvioNewsletterMaxNameLen)
		if err != nil {
			return e.BadRequestError(err.Error(), nil)
		}
		name = sanitizeNuvioNewsletterName(name)
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

		logNuvioNewsletterPublicBaseURLMissing(e.App, website, "confirm")
		confirmPath, err := buildNuvioNewsletterConfirmPath(website)
		if err != nil {
			return e.InternalServerError("Unable to prepare confirmation link right now.", nil)
		}
		baseURL := resolveNuvioNewsletterPublicBaseURL(e.Request)
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
			config.ConfirmationTemplate,
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
		rawToken := normalizeNuvioNewsletterPublicToken(e.Request.URL.Query().Get("token"))
		if rawToken == "" {
			if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
				return renderNuvioNewsletterLifecycleHTML(
					e,
					http.StatusBadRequest,
					"Confirmation link expired",
					"This confirmation link is invalid, expired, or was already used. You can subscribe again from the website.",
				)
			}
			return e.BadRequestError("Missing confirmation token.", nil)
		}

		tokenHash := hashNuvioNewsletterToken(rawToken)
		if tokenHash == "" {
			if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
				return renderNuvioNewsletterLifecycleHTML(
					e,
					http.StatusBadRequest,
					"Confirmation link expired",
					"This confirmation link is invalid, expired, or was already used. You can subscribe again from the website.",
				)
			}
			return e.BadRequestError("Invalid or expired confirmation link.", nil)
		}

		if shouldRedirectNuvioNewsletterLifecycleToPublicPage(e.Request) {
			redirectURL, ok := tryBuildNuvioNewsletterLifecycleRedirectURL(
				e.App,
				e.Request,
				rawToken,
				tokenHash,
				func(app core.App, subscribersCollection *core.Collection, hash string) (*core.Record, error) {
					return findNuvioSubscriberByConfirmationTokenHash(app, subscribersCollection, hash)
				},
				buildNuvioNewsletterConfirmPath,
			)
			if ok {
				return e.Redirect(http.StatusTemporaryRedirect, redirectURL)
			}
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
				if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
					return renderNuvioNewsletterLifecycleHTML(
						e,
						http.StatusBadRequest,
						"Confirmation link expired",
						"This confirmation link is invalid, expired, or was already used. You can subscribe again from the website.",
					)
				}
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
			if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
				return renderNuvioNewsletterLifecycleHTML(
					e,
					http.StatusBadRequest,
					"Confirmation link expired",
					"This confirmation link is invalid, expired, or was already used. You can subscribe again from the website.",
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

		if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
			return renderNuvioNewsletterLifecycleHTML(
				e,
				http.StatusOK,
				"Newsletter subscription confirmed",
				"Thank you - your subscription has been confirmed.",
			)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":      true,
			"status":  nuvioNewsletterStatusActive,
			"message": "Subscription confirmed.",
		})
	})

	newsletterPublicGroup.GET("/unsubscribe", func(e *core.RequestEvent) error {
		rawToken := normalizeNuvioNewsletterPublicToken(e.Request.URL.Query().Get("token"))
		if rawToken == "" {
			if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
				return renderNuvioNewsletterLifecycleHTML(
					e,
					http.StatusBadRequest,
					"Invalid unsubscribe link",
					"This unsubscribe link is invalid.",
				)
			}
			return e.BadRequestError("Missing unsubscribe token.", nil)
		}

		tokenHash := hashNuvioNewsletterToken(rawToken)
		if tokenHash == "" {
			if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
				return renderNuvioNewsletterLifecycleHTML(
					e,
					http.StatusBadRequest,
					"Invalid unsubscribe link",
					"This unsubscribe link is invalid.",
				)
			}
			return e.BadRequestError("Invalid unsubscribe link.", nil)
		}

		if shouldRedirectNuvioNewsletterLifecycleToPublicPage(e.Request) {
			redirectURL, ok := tryBuildNuvioNewsletterLifecycleRedirectURL(
				e.App,
				e.Request,
				rawToken,
				tokenHash,
				func(app core.App, subscribersCollection *core.Collection, hash string) (*core.Record, error) {
					return findNuvioSubscriberByUnsubscribeTokenHash(app, subscribersCollection, hash)
				},
				buildNuvioNewsletterUnsubscribePath,
			)
			if ok {
				return e.Redirect(http.StatusTemporaryRedirect, redirectURL)
			}
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
				if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
					return renderNuvioNewsletterLifecycleHTML(
						e,
						http.StatusBadRequest,
						"Invalid unsubscribe link",
						"This unsubscribe link is invalid.",
					)
				}
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

			if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
				return renderNuvioNewsletterLifecycleHTML(
					e,
					http.StatusOK,
					"Already unsubscribed",
					"You are already unsubscribed from these newsletter updates.",
				)
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

		if shouldRenderNuvioNewsletterLifecycleHTML(e.Request) {
			return renderNuvioNewsletterLifecycleHTML(
				e,
				http.StatusOK,
				"You are unsubscribed",
				"You will no longer receive these newsletter emails.",
			)
		}

		return e.JSON(http.StatusOK, map[string]any{
			"ok":      true,
			"status":  nuvioNewsletterStatusUnsubscribed,
			"message": "You have been unsubscribed.",
		})
	})

	newsletterAdminGroup.POST("/invite", func(e *core.RequestEvent) error {
		return handleNuvioNewsletterInvite(e, false)
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

func handleNuvioNewsletterInvite(e *core.RequestEvent, enforceWebsiteAccess bool) error {
	payload := nuvioNewsletterInvitePayload{}
	if err := e.BindBody(&payload); err != nil {
		return e.BadRequestError("Invalid newsletter invite payload.", nil)
	}

	return executeNuvioNewsletterInvite(e, payload, enforceWebsiteAccess)
}

func executeNuvioNewsletterInvite(
	e *core.RequestEvent,
	payload nuvioNewsletterInvitePayload,
	enforceWebsiteAccess bool,
) error {
	websiteID := strings.TrimSpace(payload.WebsiteID)
	if websiteID == "" {
		return e.BadRequestError("Missing websiteId.", nil)
	}

	if enforceWebsiteAccess {
		if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
			return err
		}
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
	source := sanitizeNuvioNewsletterName(payload.Source)

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

	currentStatus := normalizeNuvioSubscriberStatus(subscriber.GetString("status"))
	if currentStatus == nuvioNewsletterStatusActive {
		return e.JSON(http.StatusOK, map[string]any{
			"ok":      true,
			"result":  "already_active",
			"status":  nuvioNewsletterStatusActive,
			"message": "This contact is already subscribed.",
		})
	}

	if currentStatus == nuvioNewsletterStatusUnsubscribed {
		return e.JSON(http.StatusOK, map[string]any{
			"ok":      true,
			"result":  "unsubscribed",
			"status":  nuvioNewsletterStatusUnsubscribed,
			"message": "This contact has unsubscribed and was not invited.",
		})
	}

	if err := ensureNuvioSubscriberUnsubscribeTokenHash(subscriber); err != nil {
		return e.InternalServerError("Failed to prepare subscriber lifecycle token.", nil)
	}

	now := time.Now().UTC()
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

	logNuvioNewsletterPublicBaseURLMissing(e.App, website, "confirm")
	confirmPath, err := buildNuvioNewsletterConfirmPath(website)
	if err != nil {
		return e.InternalServerError("Unable to prepare confirmation link right now.", nil)
	}
	baseURL := resolveNuvioNewsletterPublicBaseURL(e.Request)
	confirmURL, err := buildNuvioNewsletterLifecycleURL(baseURL, confirmPath, rawToken)
	if err != nil {
		return e.InternalServerError("Unable to prepare confirmation link right now.", nil)
	}

	if err := sendNuvioNewsletterConfirmationEmail(
		e.Request.Context(),
		website,
		email,
		confirmURL,
		config.ConfirmationTemplate,
	); err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter manual invite confirmation send failed",
			"websiteId",
			websiteID,
			"source",
			source,
			"error",
			err.Error(),
		)
		return e.InternalServerError("Unable to send confirmation email right now. Please try again.", nil)
	}

	result := "invited"
	message := "Newsletter invitation sent."
	if !isNewSubscriber && currentStatus == nuvioNewsletterStatusPending {
		result = "resent"
		message = "Confirmation email sent again."
	}

	return e.JSON(http.StatusOK, map[string]any{
		"ok":      true,
		"result":  result,
		"status":  nuvioNewsletterStatusPending,
		"message": message,
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

	logNuvioNewsletterPublicBaseURLMissing(app, website, "unsubscribe")
	unsubscribePath, err := buildNuvioNewsletterUnsubscribePath(website)
	if err != nil {
		return nil, err
	}
	baseURL := resolveNuvioNewsletterPublicBaseURL(request)
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
		FeatureAvailable:     true,
		DoubleOptIn:          false,
		ConfirmationTemplate: nuvioNewsletterConfirmationTemplateConfig{},
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

		config.ConfirmationTemplate = parseNuvioNewsletterConfirmationTemplateConfig(newsletterSettings)
	}

	return website, config, nil
}

func parseNuvioNewsletterConfirmationTemplateConfig(
	newsletterSettings map[string]any,
) nuvioNewsletterConfirmationTemplateConfig {
	template := nuvioNewsletterConfirmationTemplateConfig{}
	if len(newsletterSettings) == 0 {
		return template
	}

	lifecycleSettings, ok := toStringAnyMap(newsletterSettings["lifecycle"])
	if !ok {
		return template
	}

	templateSettings, ok := toStringAnyMap(lifecycleSettings["confirmationTemplate"])
	if !ok {
		return template
	}

	if value, ok := parseBoolValue(templateSettings["enabled"]); ok {
		template.Enabled = value
	}

	if value, ok := templateSettings["subject"].(string); ok {
		template.Subject = value
	}

	if value, ok := templateSettings["introText"].(string); ok {
		template.IntroText = value
	}

	if value, ok := templateSettings["footerText"].(string); ok {
		template.FooterText = value
	}

	return template
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
		"",
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

func normalizeNuvioNewsletterPublicToken(rawToken string) string {
	trimmed := strings.TrimSpace(rawToken)
	if trimmed == "" {
		return ""
	}
	if len([]rune(trimmed)) > nuvioNewsletterPublicTokenMaxLen {
		return ""
	}
	return trimmed
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

func hasNuvioNewsletterPublicBaseURLConfigured() bool {
	return resolveNuvioNewsletterPublicBaseURLFromEnv() != ""
}

func shouldRedirectNuvioNewsletterLifecycleToPublicPage(request *http.Request) bool {
	if request == nil {
		return false
	}

	if strings.EqualFold(strings.TrimSpace(request.URL.Query().Get("format")), "json") {
		return false
	}

	accept := strings.ToLower(strings.TrimSpace(request.Header.Get("Accept")))
	if accept == "" {
		return true
	}

	if strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*") {
		return true
	}

	if strings.Contains(accept, "application/json") {
		return false
	}

	return false
}

func shouldRenderNuvioNewsletterLifecycleHTML(request *http.Request) bool {
	if request == nil {
		return false
	}

	if strings.EqualFold(strings.TrimSpace(request.URL.Query().Get("format")), "json") {
		return false
	}

	accept := strings.ToLower(strings.TrimSpace(request.Header.Get("Accept")))
	if accept == "" {
		return true
	}

	if strings.Contains(accept, "text/html") || strings.Contains(accept, "*/*") {
		return true
	}

	return false
}

func tryBuildNuvioNewsletterLifecycleRedirectURL(
	app core.App,
	request *http.Request,
	rawToken string,
	tokenHash string,
	findSubscriberByTokenHash func(core.App, *core.Collection, string) (*core.Record, error),
	buildPath func(*core.Record) (string, error),
) (string, bool) {
	if app == nil || request == nil || findSubscriberByTokenHash == nil || buildPath == nil {
		return "", false
	}

	if !hasNuvioNewsletterPublicBaseURLConfigured() {
		return "", false
	}

	trimmedToken := strings.TrimSpace(rawToken)
	trimmedHash := strings.TrimSpace(tokenHash)
	if trimmedToken == "" || trimmedHash == "" {
		return "", false
	}

	subscribersCollection, err := app.FindCachedCollectionByNameOrId(nuvioSubscribersCollectionID)
	if err != nil {
		return "", false
	}

	subscriber, err := findSubscriberByTokenHash(app, subscribersCollection, trimmedHash)
	if err != nil || subscriber == nil {
		return "", false
	}

	websiteID := strings.TrimSpace(subscriber.GetString("website"))
	if websiteID == "" {
		return "", false
	}

	website, err := app.FindRecordById(nuvioWebsitesCollectionID, websiteID)
	if err != nil || website == nil {
		return "", false
	}

	path, err := buildPath(website)
	if err != nil || !strings.HasPrefix(path, "/site/") {
		return "", false
	}

	baseURL := resolveNuvioNewsletterPublicBaseURLFromEnv()
	redirectURL, err := buildNuvioNewsletterLifecycleURL(baseURL, path, trimmedToken)
	if err != nil {
		return "", false
	}

	return redirectURL, true
}

func renderNuvioNewsletterLifecycleHTML(
	e *core.RequestEvent,
	status int,
	title string,
	message string,
) error {
	if e == nil {
		return nil
	}

	safeTitle := html.EscapeString(strings.TrimSpace(title))
	if safeTitle == "" {
		safeTitle = "Newsletter update"
	}

	safeMessage := html.EscapeString(strings.TrimSpace(message))
	if safeMessage == "" {
		safeMessage = "Your newsletter request has been processed."
	}

	page := strings.Join([]string{
		"<!doctype html>",
		`<html lang="en">`,
		"<head>",
		`<meta charset="utf-8">`,
		`<meta name="viewport" content="width=device-width,initial-scale=1">`,
		`<title>` + safeTitle + `</title>`,
		`<style>`,
		`body{margin:0;font-family:ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#f3f4f6;color:#111827}`,
		`.wrap{min-height:100vh;display:flex;align-items:center;justify-content:center;padding:24px}`,
		`.card{max-width:560px;width:100%;background:#fff;border:1px solid #e5e7eb;border-radius:14px;padding:28px;box-shadow:0 10px 20px rgba(17,24,39,.06)}`,
		`h1{margin:0 0 10px;font-size:1.35rem;line-height:1.3}`,
		`p{margin:0;font-size:1rem;line-height:1.55;color:#374151}`,
		`</style>`,
		"</head>",
		"<body>",
		`<main class="wrap"><section class="card">`,
		`<h1>` + safeTitle + `</h1>`,
		`<p>` + safeMessage + `</p>`,
		`</section></main>`,
		"</body>",
		"</html>",
	}, "")

	return e.HTML(status, page)
}

func logNuvioNewsletterPublicBaseURLMissing(
	app core.App,
	website *core.Record,
	lifecycleType string,
) {
	if app == nil || hasNuvioNewsletterPublicBaseURLConfigured() || website == nil {
		return
	}

	slug := strings.TrimSpace(website.GetString("slug"))
	if slug == "" {
		return
	}

	normalizedLifecycleType := strings.TrimSpace(lifecycleType)
	if normalizedLifecycleType == "" {
		normalizedLifecycleType = "lifecycle"
	}

	app.Logger().Warn(
		"NUVIO_PUBLIC_BASE_URL is required for visitor-facing newsletter lifecycle links.",
		"env",
		nuvioNewsletterPublicBaseURLEnv,
		"lifecycle",
		normalizedLifecycleType,
		"websiteId",
		website.Id,
		"websiteSlug",
		slug,
	)
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

func buildNuvioNewsletterConfirmPath(website *core.Record) (string, error) {
	if website == nil {
		return "/api/nuvio/newsletter/confirm", nil
	}

	slug := strings.TrimSpace(website.GetString("slug"))
	if slug == "" {
		return "/api/nuvio/newsletter/confirm", nil
	}

	if !hasNuvioNewsletterPublicBaseURLConfigured() {
		return "/api/nuvio/newsletter/confirm", nil
	}

	return "/site/" + url.PathEscape(slug) + "/newsletter/confirm", nil
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
	templateConfig nuvioNewsletterConfirmationTemplateConfig,
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

	subject, textBody := buildNuvioDefaultNewsletterConfirmationEmail(websiteName, confirmURL)
	if customSubject, customBody, ok := buildNuvioCustomNewsletterConfirmationEmail(
		templateConfig,
		websiteName,
		normalizedEmail,
		confirmURL,
		time.Now().UTC(),
		subject,
	); ok {
		subject = customSubject
		textBody = customBody
	}

	return sendNuvioTransactionalEmailViaResend(ctx, resendConfig, nuvioTransactionalEmailMessage{
		To:      []string{normalizedEmail},
		Subject: subject,
		Text:    textBody,
	})
}

func buildNuvioDefaultNewsletterConfirmationEmail(websiteName string, confirmURL string) (string, string) {
	subject := "Confirm your subscription"
	textBody := strings.Join([]string{
		fmt.Sprintf("You requested a newsletter subscription for %s.", websiteName),
		"",
		"Please confirm your subscription by opening this link:",
		confirmURL,
		"",
		"If you did not request this, you can safely ignore this email.",
	}, "\n")

	return subject, textBody
}

func buildNuvioCustomNewsletterConfirmationEmail(
	templateConfig nuvioNewsletterConfirmationTemplateConfig,
	websiteName string,
	subscriberEmail string,
	confirmURL string,
	submittedAt time.Time,
	defaultSubject string,
) (string, string, bool) {
	if !templateConfig.Enabled {
		return "", "", false
	}

	values := buildNuvioNewsletterConfirmationTemplateVariables(
		websiteName,
		subscriberEmail,
		confirmURL,
		submittedAt,
	)

	subject := sanitizeNuvioNewsletterTemplateSubject(
		replaceNuvioNewsletterConfirmationTemplateVariables(templateConfig.Subject, values),
	)
	introText := sanitizeNuvioNewsletterTemplateText(
		replaceNuvioNewsletterConfirmationTemplateVariables(templateConfig.IntroText, values),
	)
	footerText := sanitizeNuvioNewsletterTemplateText(
		replaceNuvioNewsletterConfirmationTemplateVariables(templateConfig.FooterText, values),
	)

	if subject == "" && introText == "" && footerText == "" {
		return "", "", false
	}

	if subject == "" {
		subject = sanitizeNuvioNewsletterTemplateSubject(defaultSubject)
		if subject == "" {
			subject = "Confirm your subscription"
		}
	}

	requiredLines := []string{
		fmt.Sprintf("You requested a newsletter subscription for %s.", values["websiteName"]),
		"",
		"Please confirm your subscription by opening this link:",
		values["confirmationUrl"],
		"",
		"If you did not request this, you can safely ignore this email.",
	}

	lines := []string{}
	if introText != "" {
		lines = append(lines, introText, "")
	}
	lines = append(lines, requiredLines...)
	if footerText != "" {
		lines = append(lines, "", footerText)
	}

	textBody := strings.Join(lines, "\n")
	if strings.TrimSpace(textBody) == "" {
		return "", "", false
	}

	return subject, textBody, true
}

func buildNuvioNewsletterConfirmationTemplateVariables(
	websiteName string,
	subscriberEmail string,
	confirmURL string,
	submittedAt time.Time,
) map[string]string {
	displayWebsiteName := sanitizeNuvioNewsletterTemplateSingleLineValue(websiteName, "Website")
	displaySubscriberEmail := sanitizeNuvioNewsletterTemplateSingleLineValue(subscriberEmail, "subscriber")
	displayConfirmationURL := strings.TrimSpace(confirmURL)
	if displayConfirmationURL == "" {
		displayConfirmationURL = confirmURL
	}

	return map[string]string{
		"websiteName":     displayWebsiteName,
		"subscriberEmail": displaySubscriberEmail,
		"confirmationUrl": displayConfirmationURL,
		"submittedAt":     submittedAt.UTC().Format(time.RFC3339),
	}
}

func replaceNuvioNewsletterConfirmationTemplateVariables(raw string, values map[string]string) string {
	replacer := strings.NewReplacer(
		"{{websiteName}}", values["websiteName"],
		"{{subscriberEmail}}", values["subscriberEmail"],
		"{{confirmationUrl}}", values["confirmationUrl"],
		"{{submittedAt}}", values["submittedAt"],
	)

	return replacer.Replace(raw)
}

func sanitizeNuvioNewsletterTemplateSubject(raw string) string {
	normalized := strings.NewReplacer("\r", " ", "\n", " ").Replace(raw)
	normalized = strings.Join(strings.Fields(normalized), " ")
	normalized = strings.TrimSpace(normalized)
	return truncateNuvioNewsletterStringByRunes(normalized, nuvioNewsletterTemplateSubjectMax)
}

func sanitizeNuvioNewsletterTemplateText(raw string) string {
	normalized := strings.ReplaceAll(raw, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimSpace(normalized)
	return truncateNuvioNewsletterStringByRunes(normalized, nuvioNewsletterTemplateTextMax)
}

func sanitizeNuvioNewsletterTemplateSingleLineValue(raw string, fallback string) string {
	normalized := strings.NewReplacer("\r", " ", "\n", " ").Replace(raw)
	normalized = strings.Join(strings.Fields(normalized), " ")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return fallback
	}
	return normalized
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

func buildNuvioNewsletterUnsubscribePath(website *core.Record) (string, error) {
	if website == nil {
		return "/api/nuvio/newsletter/unsubscribe", nil
	}

	slug := strings.TrimSpace(website.GetString("slug"))
	if slug == "" {
		return "/api/nuvio/newsletter/unsubscribe", nil
	}

	if !hasNuvioNewsletterPublicBaseURLConfigured() {
		return "/api/nuvio/newsletter/unsubscribe", nil
	}

	return "/site/" + url.PathEscape(slug) + "/newsletter/unsubscribe", nil
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

func loadNuvioNewsletterBackofficeDashboard(
	app core.App,
	websiteID string,
) (nuvioNewsletterBackofficeDatasets, nuvioNewsletterBackofficeCapabilities, error) {
	datasets := nuvioNewsletterBackofficeDatasets{
		Subscribers: []nuvioNewsletterBackofficeSubscriberDTO{},
		Campaigns:   []nuvioNewsletterBackofficeCampaignDTO{},
		Groups:      []nuvioNewsletterBackofficeGroupDTO{},
	}

	subscribersCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(app, nuvioNewsletterBackofficeSubscribersCollectionAliases)
	if err != nil {
		return datasets, nuvioNewsletterBackofficeCapabilities{}, err
	}
	campaignsCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(app, nuvioNewsletterBackofficeCampaignsCollectionAliases)
	if err != nil {
		return datasets, nuvioNewsletterBackofficeCapabilities{}, err
	}
	groupsCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(app, nuvioNewsletterBackofficeGroupsCollectionAliases)
	if err != nil {
		return datasets, nuvioNewsletterBackofficeCapabilities{}, err
	}

	subscriberSourceFieldName := resolveNuvioCollectionFieldNameByAliases(subscribersCollection, nuvioNewsletterBackofficeSubscribersSourceFieldAliases)
	subscriberGroupsFieldName := resolveNuvioCollectionFieldNameByAliases(subscribersCollection, nuvioNewsletterBackofficeSubscribersGroupsFieldAliases)
	campaignRecipientsTypeFieldName := resolveNuvioCollectionFieldNameByAliases(campaignsCollection, nuvioNewsletterBackofficeCampaignRecipientsTypeFieldAliases)
	campaignRecipientsIDsFieldName := resolveNuvioCollectionFieldNameByAliases(campaignsCollection, nuvioNewsletterBackofficeCampaignRecipientsIDsFieldAliases)

	subscribers, err := findNuvioNewsletterBackofficeRecordsByWebsite(
		app,
		subscribersCollection,
		websiteID,
		[]string{"-created"},
	)
	if err != nil {
		return datasets, nuvioNewsletterBackofficeCapabilities{}, err
	}
	for _, record := range subscribers {
		datasets.Subscribers = append(datasets.Subscribers, buildNuvioNewsletterBackofficeSubscriberDTO(record, subscriberSourceFieldName, subscriberGroupsFieldName))
	}

	campaigns, err := findNuvioNewsletterBackofficeRecordsByWebsite(
		app,
		campaignsCollection,
		websiteID,
		[]string{"-sentAt,-updated,-created", "-updated,-created", "-created"},
	)
	if err != nil {
		return datasets, nuvioNewsletterBackofficeCapabilities{}, err
	}
	for _, record := range campaigns {
		datasets.Campaigns = append(datasets.Campaigns, buildNuvioNewsletterBackofficeCampaignDTO(record, campaignRecipientsTypeFieldName, campaignRecipientsIDsFieldName))
	}

	groups, err := findNuvioNewsletterBackofficeRecordsByWebsite(
		app,
		groupsCollection,
		websiteID,
		[]string{"+name", "+created"},
	)
	if err != nil {
		return datasets, nuvioNewsletterBackofficeCapabilities{}, err
	}
	for _, record := range groups {
		datasets.Groups = append(datasets.Groups, buildNuvioNewsletterBackofficeGroupDTO(record))
	}

	capabilities := nuvioNewsletterBackofficeCapabilities{
		Subscribers: nuvioNewsletterBackofficeSubscribersCapabilities{
			AllowedStatus:  resolveNuvioNewsletterBackofficeAllowedSelectValues(subscribersCollection, "status", nuvioNewsletterBackofficeAllowedSubscriberStatus),
			SupportsName:   resolveNuvioCollectionFieldNameByAliases(subscribersCollection, []string{"name"}) != "",
			SupportsSource: subscriberSourceFieldName != "",
			SupportsGroups: subscriberGroupsFieldName != "",
		},
		Campaigns: nuvioNewsletterBackofficeCampaignCapabilities{
			AllowedStatus:         resolveNuvioNewsletterBackofficeAllowedSelectValues(campaignsCollection, "status", nuvioNewsletterBackofficeAllowedCampaignStatus),
			AllowedRecipientsType: resolveNuvioNewsletterBackofficeAllowedSelectValues(campaignsCollection, campaignRecipientsTypeFieldName, nuvioNewsletterBackofficeAllowedRecipientsType),
		},
	}

	return datasets, capabilities, nil
}

func buildNuvioNewsletterBackofficeSubscriberDTO(
	record *core.Record,
	sourceFieldName string,
	groupsFieldName string,
) nuvioNewsletterBackofficeSubscriberDTO {
	if record == nil {
		return nuvioNewsletterBackofficeSubscriberDTO{Groups: []string{}}
	}

	source := ""
	if sourceFieldName != "" {
		source = strings.TrimSpace(parseStringValue(record.Get(sourceFieldName)))
	}

	groupIDs := []string{}
	if groupsFieldName != "" {
		groupIDs = parseNuvioRecipientIDs(record.Get(groupsFieldName))
	}

	return nuvioNewsletterBackofficeSubscriberDTO{
		ID:             strings.TrimSpace(record.Id),
		Website:        resolveNuvioPublicRelationID(record, "website", "site"),
		Email:          strings.TrimSpace(record.GetString("email")),
		Name:           strings.TrimSpace(record.GetString("name")),
		Status:         normalizeNuvioSubscriberStatus(record.GetString("status")),
		Source:         source,
		Groups:         groupIDs,
		ConfirmedAt:    strings.TrimSpace(record.GetString("confirmedAt")),
		UnsubscribedAt: strings.TrimSpace(record.GetString("unsubscribedAt")),
		Created:        strings.TrimSpace(record.GetString("created")),
		Updated:        strings.TrimSpace(record.GetString("updated")),
	}
}

func buildNuvioNewsletterBackofficeCampaignDTO(
	record *core.Record,
	recipientsTypeFieldName string,
	recipientsIDsFieldName string,
) nuvioNewsletterBackofficeCampaignDTO {
	if record == nil {
		return nuvioNewsletterBackofficeCampaignDTO{
			RecipientsIDs: []string{},
		}
	}

	recipientsType := ""
	for _, fieldName := range []string{
		strings.TrimSpace(recipientsTypeFieldName),
		"recipientsType",
		"recipientType",
		"recipients_type",
	} {
		if fieldName == "" {
			continue
		}
		value := strings.ToLower(strings.TrimSpace(record.GetString(fieldName)))
		if value != "" {
			recipientsType = value
			break
		}
	}
	if recipientsType == "" {
		recipientsType = "all"
	}

	recipientsIDs := []string{}
	for _, fieldName := range []string{
		strings.TrimSpace(recipientsIDsFieldName),
		"recipientsIds",
		"recipientIds",
		"recipients_ids",
	} {
		if fieldName == "" {
			continue
		}
		values := parseNuvioRecipientIDs(record.Get(fieldName))
		if len(values) > 0 {
			recipientsIDs = values
			break
		}
	}

	return nuvioNewsletterBackofficeCampaignDTO{
		ID:              strings.TrimSpace(record.Id),
		Website:         resolveNuvioPublicRelationID(record, "website", "site"),
		Subject:         strings.TrimSpace(record.GetString("subject")),
		Body:            strings.TrimSpace(record.GetString("body")),
		Status:          strings.ToLower(strings.TrimSpace(record.GetString("status"))),
		RecipientsType:  recipientsType,
		RecipientsIDs:   recipientsIDs,
		RecipientsCount: parseNuvioNonNegativeInt(record.Get("recipientsCount"), 0),
		SentAt:          strings.TrimSpace(record.GetString("sentAt")),
		Created:         strings.TrimSpace(record.GetString("created")),
		Updated:         strings.TrimSpace(record.GetString("updated")),
	}
}

func buildNuvioNewsletterBackofficeGroupDTO(record *core.Record) nuvioNewsletterBackofficeGroupDTO {
	if record == nil {
		return nuvioNewsletterBackofficeGroupDTO{}
	}

	return nuvioNewsletterBackofficeGroupDTO{
		ID:      strings.TrimSpace(record.Id),
		Website: resolveNuvioPublicRelationID(record, "website", "site"),
		Name:    strings.TrimSpace(record.GetString("name")),
		Slug:    strings.TrimSpace(record.GetString("slug")),
		Created: strings.TrimSpace(record.GetString("created")),
		Updated: strings.TrimSpace(record.GetString("updated")),
	}
}

func resolveNuvioNewsletterBackofficeAllowedSelectValues(
	collection *core.Collection,
	fieldName string,
	fallback []string,
) []string {
	if len(fallback) == 0 {
		return []string{}
	}

	seen := map[string]struct{}{}

	if collection != nil && strings.TrimSpace(fieldName) != "" {
		if field := collection.Fields.GetByName(fieldName); field != nil {
			if selectField, ok := field.(*core.SelectField); ok {
				for _, rawValue := range selectField.Values {
					normalizedValue := strings.ToLower(strings.TrimSpace(rawValue))
					if normalizedValue == "" {
						continue
					}
					for _, allowedValue := range fallback {
						if normalizedValue == allowedValue {
							seen[allowedValue] = struct{}{}
							break
						}
					}
				}
			}
		}
	}

	if len(seen) == 0 {
		for _, fallbackValue := range fallback {
			seen[fallbackValue] = struct{}{}
		}
	}

	ordered := make([]string, 0, len(fallback))
	for _, fallbackValue := range fallback {
		if _, exists := seen[fallbackValue]; exists {
			ordered = append(ordered, fallbackValue)
		}
	}

	return ordered
}

func findNuvioNewsletterBackofficeCollectionByAliases(app core.App, aliases []string) (*core.Collection, error) {
	var lastErr error

	for _, alias := range aliases {
		candidate := strings.TrimSpace(alias)
		if candidate == "" {
			continue
		}

		collection, err := app.FindCachedCollectionByNameOrId(candidate)
		if err == nil {
			return collection, nil
		}

		lastErr = err
		if !errors.Is(err, sql.ErrNoRows) {
			continue
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("collection not found for aliases")
}

func findNuvioNewsletterBackofficeRecordsByWebsite(
	app core.App,
	collection *core.Collection,
	websiteID string,
	sortCandidates []string,
) ([]*core.Record, error) {
	if collection == nil {
		return nil, fmt.Errorf("missing collection")
	}

	websiteFieldName := resolveNuvioCollectionFieldNameByAliases(collection, []string{"website", "site"})
	if websiteFieldName == "" {
		return nil, fmt.Errorf("missing website relation field")
	}

	filterExpr := websiteFieldName + "={:websiteId}"
	params := dbx.Params{
		"websiteId": websiteID,
	}

	sanitizedSortCandidates := make([]string, 0, len(sortCandidates)+1)
	for _, rawSort := range sortCandidates {
		sortExpr := strings.TrimSpace(rawSort)
		if sortExpr != "" {
			sanitizedSortCandidates = append(sanitizedSortCandidates, sortExpr)
		}
	}
	sanitizedSortCandidates = append(sanitizedSortCandidates, "")

	var lastErr error
	for _, sortExpr := range sanitizedSortCandidates {
		records, err := app.FindRecordsByFilter(
			collection,
			filterExpr,
			sortExpr,
			nuvioNewsletterMaxSubscriberScan,
			0,
			params,
		)
		if err == nil {
			return records, nil
		}

		lastErr = err
		if sortExpr != "" && strings.Contains(strings.ToLower(err.Error()), "invalid sort field") {
			continue
		}

		return nil, err
	}

	return nil, lastErr
}

func handleNuvioNewsletterBackofficeSubscriberCreate(e *core.RequestEvent) error {
	subscribersCollection, groupsCollection, subscriberSourceFieldName, subscriberGroupsFieldName, err := resolveNuvioNewsletterBackofficeSubscriberCollectionsAndFields(e.App)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice subscriber create resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to create subscriber.", nil)
	}

	payload, err := parseNuvioNewsletterBackofficePayloadMap(e)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	if err := validateNuvioNewsletterBackofficePayloadKeys(payload, nuvioNewsletterBackofficeSubscriberCreateAllowedPayloadKeys); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	websiteID := strings.TrimSpace(parseStringValue(payload["websiteId"]))
	if websiteID == "" {
		return e.BadRequestError("Missing websiteId.", nil)
	}
	if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
		return err
	}

	email, ok := normalizeNuvioEmail(parseStringValue(payload["email"]))
	if !ok {
		return e.BadRequestError("A valid email is required.", nil)
	}

	status, err := parseNuvioNewsletterBackofficeStatusValue(payload, "status", nuvioNewsletterBackofficeAllowedSubscriberStatus, nuvioNewsletterStatusPending)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	groupIDs, err := parseAndValidateNuvioNewsletterBackofficeGroupIDsByWebsite(e.App, groupsCollection, websiteID, payload["groups"])
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	record := core.NewRecord(subscribersCollection)
	setNuvioNewsletterBackofficeRelationField(record, subscribersCollection, nuvioNewsletterBackofficeWebsiteFieldAliases, websiteID)
	setNuvioNewsletterBackofficeStringField(record, subscribersCollection, nuvioNewsletterBackofficeSubscriberEmailFieldAliases, email)
	setNuvioNewsletterBackofficeStringField(record, subscribersCollection, nuvioNewsletterBackofficeSubscriberNameFieldAliases, sanitizeNuvioNewsletterName(parseStringValue(payload["name"])))
	setNuvioNewsletterBackofficeStringField(record, subscribersCollection, nuvioNewsletterBackofficeSubscriberStatusFieldAliases, status)

	if subscriberSourceFieldName != "" {
		record.Set(subscriberSourceFieldName, sanitizeNuvioNewsletterBackofficeTextValue(parseStringValue(payload["source"]), nuvioNewsletterTemplateSubjectMax))
	}
	if subscriberGroupsFieldName != "" {
		record.Set(subscriberGroupsFieldName, groupIDs)
	}

	if saveErr := e.App.Save(record); saveErr != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice subscriber create save failed",
			"websiteId",
			websiteID,
			"error",
			saveErr.Error(),
		)
		return e.BadRequestError("Failed to create subscriber.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state":      "ok",
		"subscriber": buildNuvioNewsletterBackofficeSubscriberDTO(record, subscriberSourceFieldName, subscriberGroupsFieldName),
	})
}

func handleNuvioNewsletterBackofficeSubscriberUpdate(e *core.RequestEvent) error {
	subscribersCollection, groupsCollection, subscriberSourceFieldName, subscriberGroupsFieldName, err := resolveNuvioNewsletterBackofficeSubscriberCollectionsAndFields(e.App)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice subscriber update resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to update subscriber.", nil)
	}

	record, err := resolveNuvioNewsletterBackofficeRecordWriteTarget(e, subscribersCollection, "Subscriber not found.")
	if err != nil {
		return err
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID == "" {
		return e.ForbiddenError("The authorized record is not allowed to perform this action.", nil)
	}

	payload, err := parseNuvioNewsletterBackofficePayloadMap(e)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if len(payload) == 0 {
		return e.BadRequestError("At least one subscriber field is required.", nil)
	}

	if err := validateNuvioNewsletterBackofficePayloadKeys(payload, nuvioNewsletterBackofficeSubscriberUpdateAllowedPayloadKeys); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	if rawEmail, hasEmail := payload["email"]; hasEmail {
		email, ok := normalizeNuvioEmail(parseStringValue(rawEmail))
		if !ok {
			return e.BadRequestError("A valid email is required.", nil)
		}
		setNuvioNewsletterBackofficeStringField(record, subscribersCollection, nuvioNewsletterBackofficeSubscriberEmailFieldAliases, email)
	}

	if rawName, hasName := payload["name"]; hasName {
		setNuvioNewsletterBackofficeStringField(record, subscribersCollection, nuvioNewsletterBackofficeSubscriberNameFieldAliases, sanitizeNuvioNewsletterName(parseStringValue(rawName)))
	}

	if _, hasStatus := payload["status"]; hasStatus {
		status, statusErr := parseNuvioNewsletterBackofficeStatusValue(payload, "status", nuvioNewsletterBackofficeAllowedSubscriberStatus, "")
		if statusErr != nil {
			return e.BadRequestError(statusErr.Error(), nil)
		}
		setNuvioNewsletterBackofficeStringField(record, subscribersCollection, nuvioNewsletterBackofficeSubscriberStatusFieldAliases, status)
	}

	if rawSource, hasSource := payload["source"]; hasSource && subscriberSourceFieldName != "" {
		record.Set(subscriberSourceFieldName, sanitizeNuvioNewsletterBackofficeTextValue(parseStringValue(rawSource), nuvioNewsletterTemplateSubjectMax))
	}

	if rawGroups, hasGroups := payload["groups"]; hasGroups {
		groupIDs, groupsErr := parseAndValidateNuvioNewsletterBackofficeGroupIDsByWebsite(e.App, groupsCollection, websiteID, rawGroups)
		if groupsErr != nil {
			return e.BadRequestError(groupsErr.Error(), nil)
		}
		if subscriberGroupsFieldName != "" {
			record.Set(subscriberGroupsFieldName, groupIDs)
		}
	}

	if saveErr := e.App.Save(record); saveErr != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice subscriber update save failed",
			"recordId",
			record.Id,
			"error",
			saveErr.Error(),
		)
		return e.BadRequestError("Failed to update subscriber.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state":      "ok",
		"subscriber": buildNuvioNewsletterBackofficeSubscriberDTO(record, subscriberSourceFieldName, subscriberGroupsFieldName),
	})
}

func handleNuvioNewsletterBackofficeSubscriberDelete(e *core.RequestEvent) error {
	subscribersCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(e.App, nuvioNewsletterBackofficeSubscribersCollectionAliases)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice subscriber delete collection resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to delete subscriber.", nil)
	}

	record, err := resolveNuvioNewsletterBackofficeRecordWriteTarget(e, subscribersCollection, "Subscriber not found.")
	if err != nil {
		return err
	}

	if deleteErr := e.App.Delete(record); deleteErr != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice subscriber delete failed",
			"recordId",
			record.Id,
			"error",
			deleteErr.Error(),
		)
		return e.BadRequestError("Failed to delete subscriber.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state": "ok",
	})
}

func handleNuvioNewsletterBackofficeSubscriberInvite(e *core.RequestEvent) error {
	subscribersCollection, _, subscriberSourceFieldName, _, err := resolveNuvioNewsletterBackofficeSubscriberCollectionsAndFields(e.App)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice subscriber invite resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to send subscriber confirmation.", nil)
	}

	record, err := resolveNuvioNewsletterBackofficeRecordWriteTarget(e, subscribersCollection, "Subscriber not found.")
	if err != nil {
		return err
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID == "" {
		return e.ForbiddenError("The authorized record is not allowed to perform this action.", nil)
	}

	email, ok := normalizeNuvioEmail(record.GetString("email"))
	if !ok {
		return e.BadRequestError("A valid email is required.", nil)
	}

	payload := nuvioNewsletterInvitePayload{
		WebsiteID: websiteID,
		Email:     email,
		Name:      sanitizeNuvioNewsletterName(record.GetString("name")),
		Source:    "manual_dashboard",
	}

	if subscriberSourceFieldName != "" {
		sourceCandidate := strings.TrimSpace(record.GetString(subscriberSourceFieldName))
		if sourceCandidate != "" {
			payload.Source = sourceCandidate
		}
	}

	return executeNuvioNewsletterInvite(e, payload, false)
}

func handleNuvioNewsletterBackofficeGroupCreate(e *core.RequestEvent) error {
	groupsCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(e.App, nuvioNewsletterBackofficeGroupsCollectionAliases)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice group create collection resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to create group.", nil)
	}

	payload, err := parseNuvioNewsletterBackofficePayloadMap(e)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	if err := validateNuvioNewsletterBackofficePayloadKeys(payload, nuvioNewsletterBackofficeGroupCreateAllowedPayloadKeys); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	websiteID := strings.TrimSpace(parseStringValue(payload["websiteId"]))
	if websiteID == "" {
		return e.BadRequestError("Missing websiteId.", nil)
	}
	if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
		return err
	}

	name := sanitizeNuvioNewsletterBackofficeTextValue(parseStringValue(payload["name"]), nuvioNewsletterBackofficeNameMax)
	if name == "" {
		return e.BadRequestError("Group name is required.", nil)
	}

	slugInput := parseStringValue(payload["slug"])
	if strings.TrimSpace(slugInput) == "" {
		slugInput = name
	}
	slug := normalizeNuvioNewsletterBackofficeSlug(slugInput)
	if slug == "" {
		return e.BadRequestError("Invalid group slug.", nil)
	}

	record := core.NewRecord(groupsCollection)
	setNuvioNewsletterBackofficeRelationField(record, groupsCollection, nuvioNewsletterBackofficeWebsiteFieldAliases, websiteID)
	setNuvioNewsletterBackofficeStringField(record, groupsCollection, nuvioNewsletterBackofficeGroupNameFieldAliases, name)
	setNuvioNewsletterBackofficeStringField(record, groupsCollection, nuvioNewsletterBackofficeGroupSlugFieldAliases, slug)

	if saveErr := e.App.Save(record); saveErr != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice group create save failed",
			"websiteId",
			websiteID,
			"error",
			saveErr.Error(),
		)
		return e.BadRequestError("Failed to create group.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state": "ok",
		"group": buildNuvioNewsletterBackofficeGroupDTO(record),
	})
}

func handleNuvioNewsletterBackofficeCampaignCreate(e *core.RequestEvent) error {
	campaignsCollection, subscribersCollection, recipientsTypeFieldName, recipientsIDsFieldName, err := resolveNuvioNewsletterBackofficeCampaignCollectionsAndFields(e.App)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign create resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to create campaign.", nil)
	}

	payload, err := parseNuvioNewsletterBackofficePayloadMap(e)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	if err := validateNuvioNewsletterBackofficePayloadKeys(payload, nuvioNewsletterBackofficeCampaignCreateAllowedPayloadKeys); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	websiteID := strings.TrimSpace(parseStringValue(payload["websiteId"]))
	if websiteID == "" {
		return e.BadRequestError("Missing websiteId.", nil)
	}
	if err := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); err != nil {
		return err
	}

	subject := sanitizeNuvioNewsletterBackofficeTextValue(parseStringValue(payload["subject"]), nuvioNewsletterTemplateSubjectMax)
	if subject == "" {
		return e.BadRequestError("Campaign subject is required.", nil)
	}

	body := sanitizeNuvioNewsletterBackofficeTextValue(parseStringValue(payload["body"]), nuvioNewsletterTemplateTextMax)
	if body == "" {
		return e.BadRequestError("Campaign body is required.", nil)
	}

	status, err := parseNuvioNewsletterBackofficeStatusValue(payload, "status", nuvioNewsletterBackofficeAllowedCampaignCreateStatus, "draft")
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	recipientsType, recipientsIDs, err := parseAndValidateNuvioNewsletterBackofficeCampaignRecipientsPayload(
		e.App,
		subscribersCollection,
		websiteID,
		payload,
		recipientsIDsFieldName,
	)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	record := core.NewRecord(campaignsCollection)
	setNuvioNewsletterBackofficeRelationField(record, campaignsCollection, nuvioNewsletterBackofficeWebsiteFieldAliases, websiteID)
	setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignSubjectFieldAliases, subject)
	setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignBodyFieldAliases, body)
	setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignStatusFieldAliases, status)
	if recipientsTypeFieldName != "" {
		record.Set(recipientsTypeFieldName, recipientsType)
	}
	if recipientsIDsFieldName != "" {
		record.Set(recipientsIDsFieldName, recipientsIDs)
	}
	setNuvioNewsletterBackofficeNumberField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignRecipientsCountFieldAliases, resolveNuvioNewsletterBackofficeRecipientsCount(recipientsType, recipientsIDs))
	setNuvioNewsletterBackofficeNullableStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignSentAtFieldAliases, "")

	if saveErr := e.App.Save(record); saveErr != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign create save failed",
			"websiteId",
			websiteID,
			"error",
			saveErr.Error(),
		)
		return e.BadRequestError("Failed to create campaign.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state":    "ok",
		"campaign": buildNuvioNewsletterBackofficeCampaignDTO(record, recipientsTypeFieldName, recipientsIDsFieldName),
	})
}

func handleNuvioNewsletterBackofficeCampaignUpdate(e *core.RequestEvent) error {
	campaignsCollection, subscribersCollection, recipientsTypeFieldName, recipientsIDsFieldName, err := resolveNuvioNewsletterBackofficeCampaignCollectionsAndFields(e.App)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign update resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to update campaign.", nil)
	}

	record, err := resolveNuvioNewsletterBackofficeRecordWriteTarget(e, campaignsCollection, "Campaign not found.")
	if err != nil {
		return err
	}

	if normalizeNuvioNewsletterCampaignStatus(record.GetString("status")) == "sent" {
		return e.BadRequestError("Sent campaigns are read-only.", nil)
	}

	payload, err := parseNuvioNewsletterBackofficePayloadMap(e)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if len(payload) == 0 {
		return e.BadRequestError("At least one campaign field is required.", nil)
	}
	if err := validateNuvioNewsletterBackofficePayloadKeys(payload, nuvioNewsletterBackofficeCampaignUpdateAllowedPayloadKeys); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID == "" {
		return e.ForbiddenError("The authorized record is not allowed to perform this action.", nil)
	}

	if rawSubject, hasSubject := payload["subject"]; hasSubject {
		subject := sanitizeNuvioNewsletterBackofficeTextValue(parseStringValue(rawSubject), nuvioNewsletterTemplateSubjectMax)
		if subject == "" {
			return e.BadRequestError("Campaign subject is required.", nil)
		}
		setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignSubjectFieldAliases, subject)
	}

	if rawBody, hasBody := payload["body"]; hasBody {
		body := sanitizeNuvioNewsletterBackofficeTextValue(parseStringValue(rawBody), nuvioNewsletterTemplateTextMax)
		if body == "" {
			return e.BadRequestError("Campaign body is required.", nil)
		}
		setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignBodyFieldAliases, body)
	}

	if _, hasStatus := payload["status"]; hasStatus {
		status, statusErr := parseNuvioNewsletterBackofficeStatusValue(payload, "status", nuvioNewsletterBackofficeAllowedCampaignUpdateStatus, "")
		if statusErr != nil {
			return e.BadRequestError(statusErr.Error(), nil)
		}
		setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignStatusFieldAliases, status)
	}

	resolvedType := resolveNuvioNewsletterBackofficeCampaignRecordRecipientsType(record, recipientsTypeFieldName)
	resolvedIDs := resolveNuvioNewsletterBackofficeCampaignRecordRecipientsIDs(record, recipientsIDsFieldName)

	if rawType, hasType := payload["recipientsType"]; hasType {
		nextType, typeErr := parseNuvioNewsletterBackofficeRecipientsType(rawType)
		if typeErr != nil {
			return e.BadRequestError(typeErr.Error(), nil)
		}
		resolvedType = nextType
	}

	if rawIDs, hasIDs := payload["recipientsIds"]; hasIDs {
		nextIDs, idsErr := validateNuvioNewsletterBackofficeSubscriberIDsByWebsite(e.App, subscribersCollection, websiteID, rawIDs, resolvedType)
		if idsErr != nil {
			return e.BadRequestError(idsErr.Error(), nil)
		}
		resolvedIDs = nextIDs
	}

	if _, hasType := payload["recipientsType"]; hasType {
		if recipientsTypeFieldName != "" {
			record.Set(recipientsTypeFieldName, resolvedType)
		}
	}
	if _, hasIDs := payload["recipientsIds"]; hasIDs {
		if recipientsIDsFieldName != "" {
			record.Set(recipientsIDsFieldName, resolvedIDs)
		}
	}
	if _, hasType := payload["recipientsType"]; hasType {
		setNuvioNewsletterBackofficeNumberField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignRecipientsCountFieldAliases, resolveNuvioNewsletterBackofficeRecipientsCount(resolvedType, resolvedIDs))
	} else if _, hasIDs := payload["recipientsIds"]; hasIDs {
		setNuvioNewsletterBackofficeNumberField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignRecipientsCountFieldAliases, resolveNuvioNewsletterBackofficeRecipientsCount(resolvedType, resolvedIDs))
	}

	if saveErr := e.App.Save(record); saveErr != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign update save failed",
			"recordId",
			record.Id,
			"error",
			saveErr.Error(),
		)
		return e.BadRequestError("Failed to update campaign.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state":    "ok",
		"campaign": buildNuvioNewsletterBackofficeCampaignDTO(record, recipientsTypeFieldName, recipientsIDsFieldName),
	})
}

func handleNuvioNewsletterBackofficeCampaignDelete(e *core.RequestEvent) error {
	campaignsCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(e.App, nuvioNewsletterBackofficeCampaignsCollectionAliases)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign delete collection resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to delete campaign.", nil)
	}

	record, err := resolveNuvioNewsletterBackofficeRecordWriteTarget(e, campaignsCollection, "Campaign not found.")
	if err != nil {
		return err
	}

	if deleteErr := e.App.Delete(record); deleteErr != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign delete failed",
			"recordId",
			record.Id,
			"error",
			deleteErr.Error(),
		)
		return e.BadRequestError("Failed to delete campaign.", nil)
	}

	return e.JSON(http.StatusOK, map[string]any{
		"state": "ok",
	})
}

func handleNuvioNewsletterBackofficeCampaignDuplicate(e *core.RequestEvent) error {
	campaignsCollection, subscribersCollection, recipientsTypeFieldName, recipientsIDsFieldName, err := resolveNuvioNewsletterBackofficeCampaignCollectionsAndFields(e.App)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign duplicate resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to duplicate campaign.", nil)
	}

	sourceRecord, err := resolveNuvioNewsletterBackofficeRecordWriteTarget(e, campaignsCollection, "Campaign not found.")
	if err != nil {
		return err
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(sourceRecord, "website", "site"))
	if websiteID == "" {
		return e.ForbiddenError("The authorized record is not allowed to perform this action.", nil)
	}

	recipientsType := resolveNuvioNewsletterBackofficeCampaignRecordRecipientsType(sourceRecord, recipientsTypeFieldName)
	recipientsIDsRaw := resolveNuvioNewsletterBackofficeCampaignRecordRecipientsIDs(sourceRecord, recipientsIDsFieldName)
	recipientsIDs, skippedRecipientsCount, err := sanitizeNuvioNewsletterBackofficeSubscriberIDsByWebsite(e.App, subscribersCollection, websiteID, recipientsIDsRaw, recipientsType)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}

	record := core.NewRecord(campaignsCollection)
	setNuvioNewsletterBackofficeRelationField(record, campaignsCollection, nuvioNewsletterBackofficeWebsiteFieldAliases, websiteID)
	setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignSubjectFieldAliases, sanitizeNuvioNewsletterBackofficeTextValue(sourceRecord.GetString("subject"), nuvioNewsletterTemplateSubjectMax))
	setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignBodyFieldAliases, sanitizeNuvioNewsletterBackofficeTextValue(sourceRecord.GetString("body"), nuvioNewsletterTemplateTextMax))
	setNuvioNewsletterBackofficeStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignStatusFieldAliases, "draft")
	if recipientsTypeFieldName != "" {
		record.Set(recipientsTypeFieldName, recipientsType)
	}
	if recipientsIDsFieldName != "" {
		record.Set(recipientsIDsFieldName, recipientsIDs)
	}
	setNuvioNewsletterBackofficeNumberField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignRecipientsCountFieldAliases, resolveNuvioNewsletterBackofficeRecipientsCount(recipientsType, recipientsIDs))
	setNuvioNewsletterBackofficeNullableStringField(record, campaignsCollection, nuvioNewsletterBackofficeCampaignSentAtFieldAliases, "")

	if saveErr := e.App.Save(record); saveErr != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign duplicate save failed",
			"sourceRecordId",
			sourceRecord.Id,
			"error",
			saveErr.Error(),
		)
		return e.BadRequestError("Failed to duplicate campaign.", nil)
	}

	response := map[string]any{
		"state":    "ok",
		"campaign": buildNuvioNewsletterBackofficeCampaignDTO(record, recipientsTypeFieldName, recipientsIDsFieldName),
	}
	if skippedRecipientsCount > 0 {
		response["skippedRecipientsCount"] = skippedRecipientsCount
	}

	return e.JSON(http.StatusOK, response)
}

func handleNuvioNewsletterBackofficeCampaignSend(e *core.RequestEvent) error {
	campaignsCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(e.App, nuvioNewsletterBackofficeCampaignsCollectionAliases)
	if err != nil {
		e.App.Logger().Error(
			"NUVIO newsletter backoffice campaign send collection resolve failed",
			"error",
			err.Error(),
		)
		return e.BadRequestError("Failed to send campaign.", nil)
	}

	record, err := resolveNuvioNewsletterBackofficeRecordWriteTarget(e, campaignsCollection, "Campaign not found.")
	if err != nil {
		return err
	}

	result, err := sendNuvioNewsletterCampaign(e.App, e.Request.Context(), record.Id, e.Request)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return e.NotFoundError("Campaign not found.", nil)
		}

		return e.BadRequestError(err.Error(), nil)
	}

	return e.JSON(http.StatusOK, result)
}

func resolveNuvioNewsletterBackofficeSubscriberCollectionsAndFields(
	app core.App,
) (*core.Collection, *core.Collection, string, string, error) {
	subscribersCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(app, nuvioNewsletterBackofficeSubscribersCollectionAliases)
	if err != nil {
		return nil, nil, "", "", err
	}

	groupsCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(app, nuvioNewsletterBackofficeGroupsCollectionAliases)
	if err != nil {
		return nil, nil, "", "", err
	}

	subscriberSourceFieldName := resolveNuvioCollectionFieldNameByAliases(subscribersCollection, nuvioNewsletterBackofficeSubscribersSourceFieldAliases)
	subscriberGroupsFieldName := resolveNuvioCollectionFieldNameByAliases(subscribersCollection, nuvioNewsletterBackofficeSubscribersGroupsFieldAliases)

	return subscribersCollection, groupsCollection, subscriberSourceFieldName, subscriberGroupsFieldName, nil
}

func resolveNuvioNewsletterBackofficeCampaignCollectionsAndFields(
	app core.App,
) (*core.Collection, *core.Collection, string, string, error) {
	campaignsCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(app, nuvioNewsletterBackofficeCampaignsCollectionAliases)
	if err != nil {
		return nil, nil, "", "", err
	}

	subscribersCollection, err := findNuvioNewsletterBackofficeCollectionByAliases(app, nuvioNewsletterBackofficeSubscribersCollectionAliases)
	if err != nil {
		return nil, nil, "", "", err
	}

	recipientsTypeFieldName := resolveNuvioCollectionFieldNameByAliases(campaignsCollection, nuvioNewsletterBackofficeCampaignRecipientsTypeFieldAliases)
	recipientsIDsFieldName := resolveNuvioCollectionFieldNameByAliases(campaignsCollection, nuvioNewsletterBackofficeCampaignRecipientsIDsFieldAliases)
	return campaignsCollection, subscribersCollection, recipientsTypeFieldName, recipientsIDsFieldName, nil
}

func parseNuvioNewsletterBackofficePayloadMap(e *core.RequestEvent) (map[string]any, error) {
	payload := map[string]any{}
	if err := e.BindBody(&payload); err != nil {
		return nil, fmt.Errorf("Invalid request payload")
	}
	return payload, nil
}

func validateNuvioNewsletterBackofficePayloadKeys(
	payload map[string]any,
	allowed map[string]struct{},
) error {
	for key := range payload {
		normalizedKey := strings.ToLower(strings.TrimSpace(key))
		if normalizedKey == "" {
			return fmt.Errorf("Invalid payload field")
		}

		lifecycleProbe := strings.ReplaceAll(normalizedKey, "_", "")
		lifecycleProbe = strings.ReplaceAll(lifecycleProbe, "-", "")
		if _, lifecycleField := nuvioNewsletterBackofficeLifecycleFieldAliases[lifecycleProbe]; lifecycleField {
			return fmt.Errorf("Lifecycle token fields are not allowed in this endpoint")
		}
		if _, ok := allowed[normalizedKey]; !ok {
			return fmt.Errorf("Field %q is not allowed in this endpoint", strings.TrimSpace(key))
		}
	}
	return nil
}

func resolveNuvioNewsletterBackofficeRecordWriteTarget(
	e *core.RequestEvent,
	collection *core.Collection,
	notFoundMessage string,
) (*core.Record, error) {
	recordID := strings.TrimSpace(e.Request.PathValue("id"))
	if recordID == "" {
		return nil, e.BadRequestError("Missing record id.", nil)
	}
	if collection == nil {
		return nil, e.BadRequestError("Failed to resolve backoffice collection.", nil)
	}

	record, err := e.App.FindRecordById(collection.Id, recordID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, e.NotFoundError(notFoundMessage, nil)
		}
		return nil, e.BadRequestError("Failed to load record.", nil)
	}

	websiteID := strings.TrimSpace(resolveNuvioPublicRelationID(record, "website", "site"))
	if websiteID == "" {
		return nil, e.ForbiddenError("The authorized record is not allowed to perform this action.", nil)
	}

	if accessErr := apis.RequireWebsiteAccessById(e.App, e.Auth, websiteID); accessErr != nil {
		return nil, accessErr
	}

	return record, nil
}

func parseNuvioNewsletterBackofficeStatusValue(
	payload map[string]any,
	fieldName string,
	allowed []string,
	defaultValue string,
) (string, error) {
	rawValue, hasValue := payload[fieldName]
	if !hasValue {
		if strings.TrimSpace(defaultValue) == "" {
			return "", fmt.Errorf("Status is required")
		}
		return strings.ToLower(strings.TrimSpace(defaultValue)), nil
	}

	normalizedValue := strings.ToLower(strings.TrimSpace(parseStringValue(rawValue)))
	if normalizedValue == "" {
		return "", fmt.Errorf("Status is required")
	}

	for _, allowedValue := range allowed {
		if normalizedValue == allowedValue {
			return normalizedValue, nil
		}
	}

	return "", fmt.Errorf("Invalid status value")
}

func parseAndValidateNuvioNewsletterBackofficeGroupIDsByWebsite(
	app core.App,
	groupsCollection *core.Collection,
	websiteID string,
	rawGroups any,
) ([]string, error) {
	groupIDs := parseNuvioRecipientIDs(rawGroups)
	if len(groupIDs) == 0 {
		return []string{}, nil
	}
	if groupsCollection == nil {
		return nil, fmt.Errorf("Subscriber groups are not configured")
	}

	normalizedGroupIDs := dedupeNuvioNewsletterBackofficeIDs(groupIDs)
	for _, groupID := range normalizedGroupIDs {
		groupRecord, err := app.FindRecordById(groupsCollection.Id, groupID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("Subscriber group not found")
			}
			return nil, fmt.Errorf("Failed to validate subscriber groups")
		}

		groupWebsiteID := strings.TrimSpace(resolveNuvioPublicRelationID(groupRecord, "website", "site"))
		if groupWebsiteID == "" || groupWebsiteID != websiteID {
			return nil, fmt.Errorf("Subscriber groups must belong to the selected website")
		}
	}

	return normalizedGroupIDs, nil
}

func parseAndValidateNuvioNewsletterBackofficeCampaignRecipientsPayload(
	app core.App,
	subscribersCollection *core.Collection,
	websiteID string,
	payload map[string]any,
	recipientsIDsFieldName string,
) (string, []string, error) {
	recipientsType := "manual"
	if rawType, hasType := payload["recipientsType"]; hasType {
		nextType, err := parseNuvioNewsletterBackofficeRecipientsType(rawType)
		if err != nil {
			return "", nil, err
		}
		recipientsType = nextType
	}

	recipientsIDsRaw := payload["recipientsIds"]
	if recipientsIDsFieldName != "" {
		if rawAlias, hasAlias := payload[recipientsIDsFieldName]; hasAlias {
			recipientsIDsRaw = rawAlias
		}
	}

	recipientsIDs, err := validateNuvioNewsletterBackofficeSubscriberIDsByWebsite(
		app,
		subscribersCollection,
		websiteID,
		recipientsIDsRaw,
		recipientsType,
	)
	if err != nil {
		return "", nil, err
	}

	return recipientsType, recipientsIDs, nil
}

func parseNuvioNewsletterBackofficeRecipientsType(rawValue any) (string, error) {
	normalizedValue := strings.ToLower(strings.TrimSpace(parseStringValue(rawValue)))
	if normalizedValue == "" {
		normalizedValue = "manual"
	}

	for _, allowed := range nuvioNewsletterBackofficeAllowedRecipientsType {
		if normalizedValue == allowed {
			return normalizedValue, nil
		}
	}

	return "", fmt.Errorf("Invalid recipientsType value")
}

func validateNuvioNewsletterBackofficeSubscriberIDsByWebsite(
	app core.App,
	subscribersCollection *core.Collection,
	websiteID string,
	rawRecipientsIDs any,
	recipientsType string,
) ([]string, error) {
	if recipientsType == "all" {
		return []string{}, nil
	}

	recipientsIDs := dedupeNuvioNewsletterBackofficeIDs(parseNuvioRecipientIDs(rawRecipientsIDs))
	if len(recipientsIDs) == 0 {
		return []string{}, nil
	}

	if subscribersCollection == nil {
		return nil, fmt.Errorf("Subscribers collection is not configured")
	}

	validatedRecipientsIDs := make([]string, 0, len(recipientsIDs))
	for _, subscriberID := range recipientsIDs {
		subscriberRecord, err := app.FindRecordById(subscribersCollection.Id, subscriberID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("Recipient subscriber not found")
			}
			return nil, fmt.Errorf("Failed to validate recipients")
		}

		subscriberWebsiteID := strings.TrimSpace(resolveNuvioPublicRelationID(subscriberRecord, "website", "site"))
		if subscriberWebsiteID == "" || subscriberWebsiteID != websiteID {
			return nil, fmt.Errorf("Recipients must belong to the selected website")
		}

		validatedRecipientsIDs = append(validatedRecipientsIDs, subscriberID)
	}

	return validatedRecipientsIDs, nil
}

func sanitizeNuvioNewsletterBackofficeSubscriberIDsByWebsite(
	app core.App,
	subscribersCollection *core.Collection,
	websiteID string,
	rawRecipientsIDs any,
	recipientsType string,
) ([]string, int, error) {
	if recipientsType == "all" {
		return []string{}, 0, nil
	}

	recipientsIDs := dedupeNuvioNewsletterBackofficeIDs(parseNuvioRecipientIDs(rawRecipientsIDs))
	if len(recipientsIDs) == 0 {
		return []string{}, 0, nil
	}

	if subscribersCollection == nil {
		return nil, 0, fmt.Errorf("Subscribers collection is not configured")
	}

	sanitizedRecipients := make([]string, 0, len(recipientsIDs))
	skippedCount := 0

	for _, subscriberID := range recipientsIDs {
		subscriberRecord, err := app.FindRecordById(subscribersCollection.Id, subscriberID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				skippedCount++
				continue
			}
			return nil, 0, fmt.Errorf("Failed to validate recipients")
		}

		subscriberWebsiteID := strings.TrimSpace(resolveNuvioPublicRelationID(subscriberRecord, "website", "site"))
		if subscriberWebsiteID == "" || subscriberWebsiteID != websiteID {
			skippedCount++
			continue
		}

		sanitizedRecipients = append(sanitizedRecipients, subscriberID)
	}

	return sanitizedRecipients, skippedCount, nil
}

func setNuvioNewsletterBackofficeRelationField(record *core.Record, collection *core.Collection, aliases []string, value string) {
	setNuvioNewsletterBackofficeStringField(record, collection, aliases, value)
}

func setNuvioNewsletterBackofficeStringField(record *core.Record, collection *core.Collection, aliases []string, value string) {
	if record == nil || collection == nil {
		return
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, aliases)
	if fieldName == "" {
		return
	}

	record.Set(fieldName, strings.TrimSpace(value))
}

func setNuvioNewsletterBackofficeNullableStringField(record *core.Record, collection *core.Collection, aliases []string, value string) {
	if record == nil || collection == nil {
		return
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, aliases)
	if fieldName == "" {
		return
	}

	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		record.Set(fieldName, "")
		return
	}

	record.Set(fieldName, normalizedValue)
}

func setNuvioNewsletterBackofficeNumberField(record *core.Record, collection *core.Collection, aliases []string, value int) {
	if record == nil || collection == nil {
		return
	}

	fieldName := resolveNuvioCollectionFieldNameByAliases(collection, aliases)
	if fieldName == "" {
		return
	}

	record.Set(fieldName, value)
}

func normalizeNuvioNewsletterBackofficeSlug(rawSlug string) string {
	slug := strings.ToLower(strings.TrimSpace(rawSlug))
	if slug == "" {
		return ""
	}

	slug = nuvioNewsletterBackofficeSlugUnsafeCharsRegex.ReplaceAllString(slug, "-")
	slug = nuvioNewsletterBackofficeSlugMultiDashRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return ""
	}

	if len([]rune(slug)) > nuvioNewsletterBackofficeSlugMax {
		slug = string([]rune(slug)[:nuvioNewsletterBackofficeSlugMax])
		slug = strings.Trim(slug, "-")
	}

	return slug
}

func sanitizeNuvioNewsletterBackofficeTextValue(raw string, maxLen int) string {
	normalized := strings.TrimSpace(raw)
	if normalized == "" || maxLen <= 0 {
		return normalized
	}

	runes := []rune(normalized)
	if len(runes) > maxLen {
		return strings.TrimSpace(string(runes[:maxLen]))
	}

	return normalized
}

func dedupeNuvioNewsletterBackofficeIDs(rawIDs []string) []string {
	if len(rawIDs) == 0 {
		return []string{}
	}

	seen := map[string]struct{}{}
	deduped := make([]string, 0, len(rawIDs))

	for _, rawID := range rawIDs {
		normalizedID := strings.TrimSpace(rawID)
		if normalizedID == "" {
			continue
		}

		if _, exists := seen[normalizedID]; exists {
			continue
		}

		seen[normalizedID] = struct{}{}
		deduped = append(deduped, normalizedID)
	}

	return deduped
}

func resolveNuvioNewsletterBackofficeCampaignRecordRecipientsType(
	record *core.Record,
	recipientsTypeFieldName string,
) string {
	for _, fieldName := range []string{
		strings.TrimSpace(recipientsTypeFieldName),
		"recipientsType",
		"recipientType",
		"recipients_type",
	} {
		if fieldName == "" {
			continue
		}

		value := strings.ToLower(strings.TrimSpace(record.GetString(fieldName)))
		if value == "all" || value == "manual" {
			return value
		}
	}

	return "manual"
}

func resolveNuvioNewsletterBackofficeCampaignRecordRecipientsIDs(
	record *core.Record,
	recipientsIDsFieldName string,
) []string {
	for _, fieldName := range []string{
		strings.TrimSpace(recipientsIDsFieldName),
		"recipientsIds",
		"recipientIds",
		"recipients_ids",
	} {
		if fieldName == "" {
			continue
		}

		values := dedupeNuvioNewsletterBackofficeIDs(parseNuvioRecipientIDs(record.Get(fieldName)))
		if len(values) > 0 {
			return values
		}
	}

	return []string{}
}

func resolveNuvioNewsletterBackofficeRecipientsCount(recipientsType string, recipientsIDs []string) int {
	if recipientsType != "manual" {
		return 0
	}
	return len(recipientsIDs)
}

func normalizeNuvioNewsletterCampaignStatus(rawStatus string) string {
	return strings.ToLower(strings.TrimSpace(rawStatus))
}

// NUVIO CUSTOM END: Newsletter V1 send endpoint (server-side campaign dispatch via Resend).
