package config

import (
	"os"
	"testing"
)

func TestIsValidTrack(t *testing.T) {
	validTracks := []string{"internal", "alpha", "beta", "production"}
	invalidTracks := []string{"custom", "staging", "dev", ""}

	for _, track := range validTracks {
		t.Run("valid_"+track, func(t *testing.T) {
			if !IsValidTrack(track) {
				t.Errorf("IsValidTrack(%q) = false, want true", track)
			}
		})
	}

	for _, track := range invalidTracks {
		t.Run("invalid_"+track, func(t *testing.T) {
			if IsValidTrack(track) {
				t.Errorf("IsValidTrack(%q) = true, want false", track)
			}
		})
	}
}

func TestNormalizeLocale(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en_US", "en-US"},
		{"en-US", "en-US"},
		{"zh_CN", "zh-CN"},
		{"pt_BR", "pt-BR"},
		{"fr", "fr"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := NormalizeLocale(tt.input); got != tt.expected {
				t.Errorf("NormalizeLocale(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDetectCI(t *testing.T) {
	// Save original env
	origCI := os.Getenv("CI")
	origGHA := os.Getenv("GITHUB_ACTIONS")
	defer func() {
		os.Setenv("CI", origCI)
		os.Setenv("GITHUB_ACTIONS", origGHA)
	}()

	// Clear CI variables
	os.Unsetenv("CI")
	os.Unsetenv("GITHUB_ACTIONS")
	os.Unsetenv("JENKINS_URL")
	os.Unsetenv("BUILDKITE")
	os.Unsetenv("CIRCLECI")
	os.Unsetenv("TRAVIS")
	os.Unsetenv("GITLAB_CI")
	os.Unsetenv("GPD_CI")

	if DetectCI() {
		t.Error("DetectCI() = true when no CI env vars set, want false")
	}

	// Set CI variable
	os.Setenv("CI", "true")
	if !DetectCI() {
		t.Error("DetectCI() = false with CI=true, want true")
	}
	os.Unsetenv("CI")

	// Test GITHUB_ACTIONS
	os.Setenv("GITHUB_ACTIONS", "true")
	if !DetectCI() {
		t.Error("DetectCI() = false with GITHUB_ACTIONS=true, want true")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.OutputFormat != "json" {
		t.Errorf("OutputFormat = %q, want 'json'", cfg.OutputFormat)
	}
	if cfg.TimeoutSeconds != 30 {
		t.Errorf("TimeoutSeconds = %d, want 30", cfg.TimeoutSeconds)
	}
	if cfg.StoreTokens != "auto" {
		t.Errorf("StoreTokens = %q, want 'auto'", cfg.StoreTokens)
	}
	if cfg.TesterLimits == nil {
		t.Error("TesterLimits = nil, want non-nil")
	}
	if cfg.TesterLimits.Internal != 200 {
		t.Errorf("TesterLimits.Internal = %d, want 200", cfg.TesterLimits.Internal)
	}
}

func TestValidTracks(t *testing.T) {
	tracks := ValidTracks()
	if len(tracks) != 4 {
		t.Errorf("ValidTracks() returned %d tracks, want 4", len(tracks))
	}

	expected := map[string]bool{
		"internal":   true,
		"alpha":      true,
		"beta":       true,
		"production": true,
	}

	for _, track := range tracks {
		if !expected[track] {
			t.Errorf("Unexpected track: %q", track)
		}
	}
}
