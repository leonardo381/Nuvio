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
