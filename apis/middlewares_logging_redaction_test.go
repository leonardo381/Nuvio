package apis

import (
	"strings"
	"testing"
)

func TestRedactRequestURIForLogsTokenAndEmail(t *testing.T) {
	uri := "/api/nuvio/newsletter/confirm?token=abc123&email=user@example.com&format=json"
	got := redactRequestURIForLogs(uri)

	if !strings.Contains(got, "token=[redacted]") {
		t.Fatalf("expected token query value to be redacted, got %q", got)
	}
	if !strings.Contains(got, "email=[redacted]") {
		t.Fatalf("expected email query value to be redacted, got %q", got)
	}
	if !strings.Contains(got, "format=json") {
		t.Fatalf("expected non-sensitive query key/value to remain visible, got %q", got)
	}
}

func TestRedactRequestURIForLogsPreservesNonSensitiveQueryKeys(t *testing.T) {
	uri := "/api/nuvio/leads/contact/submit?websiteSlug=demo&source=footer&page=%2Fhome"
	got := redactRequestURIForLogs(uri)

	if got != uri {
		t.Fatalf("expected non-sensitive query to remain unchanged, got %q", got)
	}
}

func TestRedactRequestURIForLogsDoesNotMutateOriginalInput(t *testing.T) {
	original := "/api/nuvio/newsletter/unsubscribe?token=secret-token&mode=json"
	input := original
	_ = redactRequestURIForLogs(input)

	if input != original {
		t.Fatalf("expected original request URI string to remain unchanged")
	}
}

func TestRedactRequestURIForLogsHandlesMalformedURI(t *testing.T) {
	uri := "/api/test?token=%ZZ&email=%GG&mode=raw"
	got := redactRequestURIForLogs(uri)

	expected := "/api/test?token=[redacted]&email=[redacted]&mode=raw"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestRedactURLQueryForLogs(t *testing.T) {
	rawURL := "https://example.com/site/newsletter/confirm?token=abc&state=xyz&utm=campaign"
	got := redactURLQueryForLogs(rawURL)

	if !strings.Contains(got, "token=[redacted]") {
		t.Fatalf("expected token query value to be redacted, got %q", got)
	}
	if !strings.Contains(got, "state=[redacted]") {
		t.Fatalf("expected state query value to be redacted, got %q", got)
	}
	if !strings.Contains(got, "utm=campaign") {
		t.Fatalf("expected non-sensitive query value to remain, got %q", got)
	}
}
