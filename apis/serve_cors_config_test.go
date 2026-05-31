package apis

import (
	"reflect"
	"testing"
)

func TestResolveCORSAllowedOriginsExplicitWins(t *testing.T) {
	explicit := []string{"https://admin.example.com", "https://admin.example.com", "  https://app.example.com  "}
	envValue := "https://env.example.com"

	got := resolveCORSAllowedOrigins(explicit, envValue, false)
	want := []string{"https://admin.example.com", "https://app.example.com"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Expected explicit origins %v, got %v", want, got)
	}
}

func TestResolveCORSAllowedOriginsEnvFallback(t *testing.T) {
	explicit := []string{}
	envValue := " https://admin.example.com,https://site.example.com  https://site.example.com "

	got := resolveCORSAllowedOrigins(explicit, envValue, false)
	want := []string{"https://admin.example.com", "https://site.example.com"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Expected env origins %v, got %v", want, got)
	}
}

func TestResolveCORSAllowedOriginsDevFallback(t *testing.T) {
	got := resolveCORSAllowedOrigins(nil, "", true)
	want := []string{
		"http://localhost:*",
		"http://127.0.0.1:*",
		"https://localhost:*",
		"https://127.0.0.1:*",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Expected dev fallback %v, got %v", want, got)
	}
}

func TestResolveCORSAllowedOriginsProductionFallback(t *testing.T) {
	got := resolveCORSAllowedOrigins(nil, "", false)
	want := []string{
		"http://localhost:*",
		"http://127.0.0.1:*",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Expected production fallback %v, got %v", want, got)
	}
}

func TestHasExplicitCORSOrigins(t *testing.T) {
	if hasExplicitCORSOrigins(nil, "") {
		t.Fatal("Expected no explicit origins")
	}

	if !hasExplicitCORSOrigins([]string{"https://admin.example.com"}, "") {
		t.Fatal("Expected explicit origins from config list")
	}

	if !hasExplicitCORSOrigins(nil, "https://admin.example.com") {
		t.Fatal("Expected explicit origins from env")
	}
}
