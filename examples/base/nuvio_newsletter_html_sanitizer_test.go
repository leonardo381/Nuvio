package main

import (
	"strings"
	"testing"
)

func TestNuvioNewsletterCampaignEmailHTMLSanitizerRemovesUnsafeContent(t *testing.T) {
	t.Parallel()

	raw := `<p>Hello<script>alert(1)</script><img src=x onerror=alert(1)>` +
		`<a href="javascript:alert(1)" onclick="alert(2)">Click</a></p>`

	sanitized := sanitizeNuvioNewsletterCampaignEmailHTML(raw)

	if strings.Contains(strings.ToLower(sanitized), "<script") {
		t.Fatalf("expected script tag to be removed, got: %s", sanitized)
	}
	if strings.Contains(strings.ToLower(sanitized), "onerror=") || strings.Contains(strings.ToLower(sanitized), "onclick=") {
		t.Fatalf("expected event handlers to be removed, got: %s", sanitized)
	}
	if strings.Contains(strings.ToLower(sanitized), "javascript:") {
		t.Fatalf("expected javascript protocol to be removed, got: %s", sanitized)
	}
	if !strings.Contains(sanitized, "<p>") {
		t.Fatalf("expected paragraph content to remain, got: %s", sanitized)
	}
}

func TestNuvioNewsletterCampaignEmailHTMLSanitizerPreservesSafeFormatting(t *testing.T) {
	t.Parallel()

	raw := `<h3>News</h3><p><strong>Bold</strong> and <em>italic</em>.</p>` +
		`<ul><li>One</li><li>Two</li></ul>` +
		`<a href="https://example.com" target="_blank">Open</a>`

	sanitized := sanitizeNuvioNewsletterCampaignEmailHTML(raw)

	requiredFragments := []string{
		"<h3>News</h3>",
		"<strong>Bold</strong>",
		"<em>italic</em>",
		"<ul><li>One</li><li>Two</li></ul>",
		`<a href="https://example.com" target="_blank" rel="noopener noreferrer">Open</a>`,
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(sanitized, fragment) {
			t.Fatalf("expected sanitized html to contain %q, got: %s", fragment, sanitized)
		}
	}
}

func TestNuvioNewsletterCampaignMessageBuildSanitizesHTMLBeforeSend(t *testing.T) {
	t.Parallel()

	message := buildNuvioCampaignRecipientMessage(
		"subscriber@example.test",
		"Campaign subject",
		`<p>Safe content</p><script>alert(1)</script><a href="javascript:alert(2)">bad</a>`,
		"Demo Site",
		"https://example.test/unsubscribe?token=abc",
	)

	if strings.TrimSpace(message.HTML) == "" {
		t.Fatalf("expected HTML message for HTML campaign body")
	}
	if strings.Contains(strings.ToLower(message.HTML), "<script") {
		t.Fatalf("expected script tag to be removed, got: %s", message.HTML)
	}
	if strings.Contains(strings.ToLower(message.HTML), "javascript:") {
		t.Fatalf("expected javascript protocol to be removed, got: %s", message.HTML)
	}
	if !strings.Contains(message.HTML, "<p>Safe content</p>") {
		t.Fatalf("expected safe content to be preserved, got: %s", message.HTML)
	}
	if !strings.Contains(message.HTML, "Unsubscribe") {
		t.Fatalf("expected unsubscribe footer to be present, got: %s", message.HTML)
	}
}
