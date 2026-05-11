package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const nuvioDefaultResendFrom = "onboarding@resend.dev"

var nuvioEmailHTTPClient = &http.Client{
	Timeout: 20 * time.Second,
}

type resendSendEmailRequest struct {
	From        string                      `json:"from"`
	To          []string                    `json:"to,omitempty"`
	Cc          []string                    `json:"cc,omitempty"`
	Bcc         []string                    `json:"bcc,omitempty"`
	ReplyTo     []string                    `json:"reply_to,omitempty"`
	Subject     string                      `json:"subject"`
	Html        string                      `json:"html,omitempty"`
	Text        string                      `json:"text,omitempty"`
	Attachments []resendSendEmailAttachment `json:"attachments,omitempty"`
}

type resendSendEmailAttachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	ContentType string `json:"content_type,omitempty"`
}

type resendSendEmailResponse struct {
	ID string `json:"id"`
}

type resendBatchSendEmailResponse struct {
	Data   []resendSendEmailResponse     `json:"data"`
	Errors []resendBatchSendEmailFailure `json:"errors"`
}

type resendBatchSendEmailFailure struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

type nuvioBatchSendResult struct {
	SentCount       int
	FailedCount     int
	FailedIndexes   []int
	AmbiguousResult bool
}

type nuvioResendConfig struct {
	APIKey string
	From   string
}

type nuvioTransactionalEmailMessage struct {
	To      []string
	Cc      []string
	Bcc     []string
	ReplyTo []string
	Subject string
	HTML    string
	Text    string
	// Attachments expects raw, unencoded bytes. The helper will base64-encode
	// content according to Resend's attachments payload shape.
	Attachments []nuvioTransactionalEmailAttachment
}

type nuvioTransactionalEmailAttachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

func loadNuvioResendConfig() (nuvioResendConfig, error) {
	apiKey := strings.TrimSpace(os.Getenv("NUVIO_RESEND_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	}
	if apiKey == "" {
		return nuvioResendConfig{}, fmt.Errorf("Missing Resend API key (set NUVIO_RESEND_API_KEY or RESEND_API_KEY)")
	}

	fromAddress := strings.TrimSpace(os.Getenv("NUVIO_RESEND_FROM"))
	if fromAddress == "" {
		fromAddress = strings.TrimSpace(os.Getenv("RESEND_FROM"))
	}
	if fromAddress == "" {
		fromAddress = strings.TrimSpace(os.Getenv("NUVIO_RESEND_FROM_EMAIL"))
	}
	if fromAddress == "" {
		fromAddress = strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL"))
	}
	if fromAddress == "" {
		fromAddress = nuvioDefaultResendFrom
	}

	return nuvioResendConfig{
		APIKey: apiKey,
		From:   fromAddress,
	}, nil
}

func sendNuvioTransactionalEmailViaResend(
	ctx context.Context,
	config nuvioResendConfig,
	message nuvioTransactionalEmailMessage,
) error {
	subject := strings.TrimSpace(message.Subject)
	if subject == "" {
		return fmt.Errorf("Email subject is required")
	}

	if len(message.To) == 0 && len(message.Bcc) == 0 {
		return fmt.Errorf("Email recipients are required")
	}

	requestPayload := resendSendEmailRequest{
		From:    strings.TrimSpace(config.From),
		To:      message.To,
		Cc:      message.Cc,
		Bcc:     message.Bcc,
		ReplyTo: message.ReplyTo,
		Subject: subject,
	}

	htmlBody := strings.TrimSpace(message.HTML)
	textBody := strings.TrimSpace(message.Text)

	if htmlBody == "" && textBody == "" {
		return fmt.Errorf("Email body is required")
	}

	if htmlBody != "" {
		requestPayload.Html = htmlBody
	}

	if textBody != "" {
		requestPayload.Text = textBody
	}

	if len(message.Attachments) > 0 {
		attachments, err := buildNuvioResendAttachments(message.Attachments)
		if err != nil {
			return err
		}
		if len(attachments) > 0 {
			requestPayload.Attachments = attachments
		}
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
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(config.APIKey))

	response, err := nuvioEmailHTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("Failed to send email via Resend: %w", err)
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

		return fmt.Errorf("Resend rejected email send (%d): %s", response.StatusCode, message)
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

func sendNuvioTransactionalEmailBatchViaResend(
	ctx context.Context,
	config nuvioResendConfig,
	messages []nuvioTransactionalEmailMessage,
) (nuvioBatchSendResult, error) {
	result := nuvioBatchSendResult{}

	if len(messages) == 0 {
		return result, fmt.Errorf("Email batch is empty")
	}

	requestPayload := make([]resendSendEmailRequest, 0, len(messages))
	for i, message := range messages {
		subject := strings.TrimSpace(message.Subject)
		if subject == "" {
			return result, fmt.Errorf("Email subject is required for batch item %d", i+1)
		}

		if len(message.To) == 0 && len(message.Bcc) == 0 {
			return result, fmt.Errorf("Email recipients are required for batch item %d", i+1)
		}

		item := resendSendEmailRequest{
			From:    strings.TrimSpace(config.From),
			To:      message.To,
			Cc:      message.Cc,
			Bcc:     message.Bcc,
			ReplyTo: message.ReplyTo,
			Subject: subject,
		}

		htmlBody := strings.TrimSpace(message.HTML)
		textBody := strings.TrimSpace(message.Text)
		if htmlBody == "" && textBody == "" {
			return result, fmt.Errorf("Email body is required for batch item %d", i+1)
		}

		if htmlBody != "" {
			item.Html = htmlBody
		}
		if textBody != "" {
			item.Text = textBody
		}

		if len(message.Attachments) > 0 {
			return result, fmt.Errorf("Email attachments are not supported in batch item %d", i+1)
		}

		requestPayload = append(requestPayload, item)
	}

	rawPayload, err := json.Marshal(requestPayload)
	if err != nil {
		return result, fmt.Errorf("Failed to encode Resend batch payload: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.resend.com/emails/batch",
		bytes.NewBuffer(rawPayload),
	)
	if err != nil {
		return result, fmt.Errorf("Failed to build Resend batch request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+strings.TrimSpace(config.APIKey))

	response, err := nuvioEmailHTTPClient.Do(request)
	if err != nil {
		return result, fmt.Errorf("Failed to send batch email via Resend: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return result, fmt.Errorf("Failed reading Resend batch response: %w", err)
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

		return result, fmt.Errorf("Resend rejected batch email send (%d): %s", response.StatusCode, message)
	}

	decoded := resendBatchSendEmailResponse{}
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return result, fmt.Errorf("Failed to decode Resend batch success response: %w", err)
	}

	failedIndexesSet := map[int]struct{}{}
	invalidIndexedErrorsCount := 0
	for _, item := range decoded.Errors {
		if strings.TrimSpace(item.Message) == "" {
			continue
		}
		if item.Index >= 0 && item.Index < len(messages) {
			failedIndexesSet[item.Index] = struct{}{}
		} else {
			result.AmbiguousResult = true
			invalidIndexedErrorsCount++
		}
	}

	if len(failedIndexesSet) > 0 {
		result.FailedIndexes = make([]int, 0, len(failedIndexesSet))
		for idx := range failedIndexesSet {
			result.FailedIndexes = append(result.FailedIndexes, idx)
		}
		sort.Ints(result.FailedIndexes)
		result.FailedCount = len(result.FailedIndexes)
		result.SentCount = len(messages) - result.FailedCount

		if len(decoded.Data) > 0 && len(decoded.Data) != result.SentCount {
			result.AmbiguousResult = true
			result.SentCount = 0
			result.FailedCount = len(messages)
			result.FailedIndexes = nil
			return result, nil
		}

		if invalidIndexedErrorsCount > 0 {
			result.AmbiguousResult = true
			result.SentCount = 0
			result.FailedCount = len(messages)
			result.FailedIndexes = nil
		}

		return result, nil
	}

	if invalidIndexedErrorsCount > 0 {
		result.SentCount = 0
		result.FailedCount = len(messages)
		result.AmbiguousResult = true
		return result, nil
	}

	if len(decoded.Data) == 0 {
		result.SentCount = 0
		result.FailedCount = len(messages)
		result.AmbiguousResult = true
		return result, nil
	}

	sentIDsCount := 0
	for _, item := range decoded.Data {
		if strings.TrimSpace(item.ID) != "" {
			sentIDsCount++
		}
	}

	if sentIDsCount == 0 {
		result.SentCount = 0
		result.FailedCount = len(messages)
		result.AmbiguousResult = true
		return result, nil
	}

	if sentIDsCount != len(messages) {
		result.SentCount = 0
		result.FailedCount = len(messages)
		result.AmbiguousResult = true
		return result, nil
	}

	result.SentCount = len(messages)
	result.FailedCount = 0
	return result, nil
}

func buildNuvioResendAttachments(
	attachments []nuvioTransactionalEmailAttachment,
) ([]resendSendEmailAttachment, error) {
	result := make([]resendSendEmailAttachment, 0, len(attachments))

	for i, attachment := range attachments {
		filename := strings.TrimSpace(attachment.Filename)
		if filename == "" {
			return nil, fmt.Errorf("Attachment %d is missing filename", i+1)
		}

		if len(attachment.Content) == 0 {
			return nil, fmt.Errorf("Attachment %q is missing content", filename)
		}

		result = append(result, resendSendEmailAttachment{
			Filename:    filename,
			Content:     base64.StdEncoding.EncodeToString(attachment.Content),
			ContentType: strings.TrimSpace(attachment.ContentType),
		})
	}

	return result, nil
}
