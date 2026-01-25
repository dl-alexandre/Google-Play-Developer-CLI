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
	setEnv := func(key, value string) {
		if err := os.Setenv(key, value); err != nil {
			t.Fatalf("Setenv(%q) failed: %v", key, err)
		}
	}
	unsetEnv := func(key string) {
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("Unsetenv(%q) failed: %v", key, err)
		}
	}

	// Save original env
	origCI := os.Getenv("CI")
	origGHA := os.Getenv("GITHUB_ACTIONS")
	defer func() {
		setEnv("CI", origCI)
		setEnv("GITHUB_ACTIONS", origGHA)
	}()

	// Clear CI variables
	unsetEnv("CI")
	unsetEnv("GITHUB_ACTIONS")
	unsetEnv("JENKINS_URL")
	unsetEnv("BUILDKITE")
	unsetEnv("CIRCLECI")
	unsetEnv("TRAVIS")
	unsetEnv("GITLAB_CI")
	unsetEnv("GPD_CI")

	if DetectCI() {
		t.Error("DetectCI() = true when no CI env vars set, want false")
	}

	// Set CI variable
	setEnv("CI", "true")
	if !DetectCI() {
		t.Error("DetectCI() = false with CI=true, want true")
	}
	unsetEnv("CI")

	// Test GITHUB_ACTIONS
	setEnv("GITHUB_ACTIONS", "true")
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
