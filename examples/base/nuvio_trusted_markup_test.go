package main

import (
	"errors"
	"strings"
	"testing"
)

func TestTrustedMarkupSVGAllowsStaticIllustration(t *testing.T) {
	t.Parallel()

	input := `<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="Signal illustration"><title>Signal</title><defs><linearGradient id="signalGradient" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#60f0ff" stop-opacity="0.9"></stop><stop offset="1" stop-color="currentColor"></stop></linearGradient><clipPath id="signalClip"><rect width="24" height="24" rx="6"></rect></clipPath></defs><g clip-path="url(#signalClip)" fill="url(#signalGradient)"><path d="M4 12 C8 4 16 4 20 12" stroke="#53e5ff" stroke-width="1.5" stroke-linecap="round" fill="none"></path><circle cx="18" cy="12" r="2" opacity="0.8"></circle></g></svg>`

	clean, report, err := sanitizeTrustedMarkup(input, trustedSvgIllustration)
	if err != nil {
		t.Fatalf("expected safe SVG to pass, got %v", err)
	}
	if !report.Valid {
		t.Fatalf("expected report to be valid: %#v", report)
	}
	for _, fragment := range []string{"<svg", "viewBox=", "<linearGradient", `clip-path="url(#signalClip)"`, "<circle"} {
		if !strings.Contains(clean, fragment) {
			t.Fatalf("expected sanitized SVG to contain %q, got: %s", fragment, clean)
		}
	}
	assertNoTrustedMarkupUnsafeFragments(t, clean)
}

func TestTrustedMarkupSVGRejectsUnsafeContent(t *testing.T) {
	t.Parallel()

	cases := []string{
		`<svg onload="alert(1)"><path d="M0 0"></path></svg>`,
		`<svg onLoad="alert(1)"><path d="M0 0"></path></svg>`,
		`<svg><script>alert(1)</script></svg>`,
		`<svg><path fill="url(http://evil.test/gradient)"></path></svg>`,
		`<svg><path fill="url(javascript:alert(1))"></path></svg>`,
		`<svg><path fill="java&#x73;cript:alert(1)"></path></svg>`,
		`<svg><foreignObject><div>bad</div></foreignObject></svg>`,
		`<svg><iframe src="https://evil.test"></iframe></svg>`,
		`<svg><image href="https://evil.test/a.png"></image></svg>`,
		`<svg><style>.x{background:url(javascript:alert(1))}</style></svg>`,
		`<svg><form><input name="email"></input></form></svg>`,
		`<svg><path style="fill:red"></path></svg>`,
	}

	for _, input := range cases {
		clean, report, err := sanitizeTrustedMarkup(input, trustedSvgIllustration)
		if err == nil || !errors.Is(err, errNuvioTrustedMarkupUnsafe) {
			t.Fatalf("expected unsafe SVG to fail closed for %s, got clean=%q report=%#v err=%v", input, clean, report, err)
		}
		if clean != "" {
			t.Fatalf("expected unsafe SVG to return empty clean markup, got: %s", clean)
		}
		if report.Valid || len(report.Errors) == 0 {
			t.Fatalf("expected invalid report with errors, got: %#v", report)
		}
	}
}

func TestTrustedMarkupHTMLAllowsLimitedVisualMarkup(t *testing.T) {
	t.Parallel()

	input := `<div class="nuvio-visual" data-layer="signal" role="presentation"><span class="node" aria-hidden="true"></span><ul class="rail"><li><strong>Public site</strong></li><li><em>Managed layer</em></li></ul><small title="Decorative support">Illustration only</small></div>`

	clean, report, err := sanitizeTrustedMarkup(input, trustedHtmlIllustration)
	if err != nil {
		t.Fatalf("expected safe HTML visual markup to pass, got %v", err)
	}
	if !report.Valid {
		t.Fatalf("expected report to be valid: %#v", report)
	}
	for _, fragment := range []string{"<div", `data-layer="signal"`, "<span", "<ul", "<strong>Public site</strong>", "<small"} {
		if !strings.Contains(clean, fragment) {
			t.Fatalf("expected sanitized HTML to contain %q, got: %s", fragment, clean)
		}
	}
	assertNoTrustedMarkupUnsafeFragments(t, clean)
}

func TestTrustedMarkupHTMLRejectsUnsafeContent(t *testing.T) {
	t.Parallel()

	cases := []string{
		`<div onclick="alert(1)">Bad</div>`,
		`<div onLoad="alert(1)">Bad</div>`,
		`<div style="color:red">Bad</div>`,
		`<div title="java&#x73;cript:alert(1)">Bad</div>`,
		`<script>alert(1)</script>`,
		`<a href="javascript:alert(1)">Bad</a>`,
		`<iframe src="https://evil.test"></iframe>`,
		`<object data="https://evil.test"></object>`,
		`<img src="https://evil.test/a.png"></img>`,
		`<form action="/submit"><input name="email"></input></form>`,
		`<svg><circle cx="1" cy="1" r="1"></circle></svg>`,
	}

	for _, input := range cases {
		clean, report, err := sanitizeTrustedMarkup(input, trustedHtmlIllustration)
		if err == nil || !errors.Is(err, errNuvioTrustedMarkupUnsafe) {
			t.Fatalf("expected unsafe HTML to fail closed for %s, got clean=%q report=%#v err=%v", input, clean, report, err)
		}
		if clean != "" {
			t.Fatalf("expected unsafe HTML to return empty clean markup, got: %s", clean)
		}
		if report.Valid || len(report.Errors) == 0 {
			t.Fatalf("expected invalid report with errors, got: %#v", report)
		}
	}
}

func TestTrustedMarkupFailsClosedOnMalformedMarkup(t *testing.T) {
	t.Parallel()

	cases := []struct {
		profile trustedMarkupProfile
		input   string
	}{
		{trustedSvgIllustration, `<svg><path></svg>`},
		{trustedHtmlIllustration, `<div><span></div>`},
		{trustedSvgIllustration, `<g><path></path></g>`},
		{trustedHtmlIllustration, `plain text only`},
		{trustedSvgIllustration, `plain text only`},
	}

	for _, scenario := range cases {
		clean, report, err := sanitizeTrustedMarkup(scenario.input, scenario.profile)
		if err == nil || !errors.Is(err, errNuvioTrustedMarkupUnsafe) {
			t.Fatalf("expected malformed markup to fail closed for %s, got clean=%q report=%#v err=%v", scenario.input, clean, report, err)
		}
		if clean != "" {
			t.Fatalf("expected malformed markup to return empty clean markup, got: %s", clean)
		}
	}
}

func assertNoTrustedMarkupUnsafeFragments(t *testing.T, value string) {
	t.Helper()

	lower := strings.ToLower(value)
	for _, unsafe := range []string{"<script", "onload=", "onclick=", "onerror=", "javascript:", "data:", "vbscript:", "<iframe", "<object", "<embed", "<foreignobject", "<form", "<input", "<style", "http://evil", "https://evil"} {
		if strings.Contains(lower, unsafe) {
			t.Fatalf("sanitized markup still contains unsafe fragment %q: %s", unsafe, value)
		}
	}
}
