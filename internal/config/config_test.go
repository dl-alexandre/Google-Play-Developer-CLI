package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestDefaultTesterLimits(t *testing.T) {
	limits := DefaultTesterLimits()
	if limits.Internal != 200 || limits.Alpha != -1 || limits.Beta != -1 {
		t.Fatalf("unexpected limits: %+v", limits)
	}
}

func TestGetPathsAndLegacyDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths := GetPaths()
	if paths.ConfigDir == "" || paths.CacheDir == "" || paths.ConfigFile == "" {
		t.Fatal("expected non-empty paths")
	}
	if filepath.Dir(paths.ConfigFile) != paths.ConfigDir {
		t.Fatal("expected config file under config dir")
	}
	if GetLegacyConfigDir() != filepath.Join(home, ".gpd") {
		t.Fatal("unexpected legacy config dir")
	}
}

func TestGetPathsForOS(t *testing.T) {
	t.Setenv("APPDATA", filepath.Join(t.TempDir(), "appdata"))
	t.Setenv("LOCALAPPDATA", filepath.Join(t.TempDir(), "localappdata"))
	winPaths := getPathsForOS("windows")
	if !strings.Contains(winPaths.ConfigDir, "appdata") {
		t.Fatal("expected windows config dir from APPDATA")
	}
	if !strings.Contains(winPaths.CacheDir, "localappdata") {
		t.Fatal("expected windows cache dir from LOCALAPPDATA")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "xdgconfig"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, "xdgcache"))
	linuxPaths := getPathsForOS("linux")
	if !strings.Contains(linuxPaths.ConfigDir, "xdgconfig") {
		t.Fatal("expected linux config dir from XDG_CONFIG_HOME")
	}
	if !strings.Contains(linuxPaths.CacheDir, "xdgcache") {
		t.Fatal("expected linux cache dir from XDG_CACHE_HOME")
	}
}

func TestGetPathsForOSDefaultXDG(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	paths := getPathsForOS("linux")
	if !strings.Contains(paths.ConfigDir, filepath.Join(home, ".config")) {
		t.Fatal("expected default config dir")
	}
	if !strings.Contains(paths.CacheDir, filepath.Join(home, ".cache")) {
		t.Fatal("expected default cache dir")
	}
}

func TestLoadPrimaryConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths := GetPaths()
	if err := os.MkdirAll(paths.ConfigDir, 0700); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	cfg := &Config{DefaultPackage: "com.example.app", OutputFormat: "json"}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(paths.ConfigFile, data, 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.DefaultPackage != "com.example.app" {
		t.Fatalf("unexpected package: %q", loaded.DefaultPackage)
	}
}

func TestLoadLegacyConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	legacyDir := GetLegacyConfigDir()
	if err := os.MkdirAll(legacyDir, 0700); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	cfg := &Config{DefaultPackage: "legacy.app", OutputFormat: "json"}
	data, _ := json.Marshal(cfg)
	legacyFile := filepath.Join(legacyDir, "config.json")
	if err := os.WriteFile(legacyFile, data, 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.DefaultPackage != "legacy.app" {
		t.Fatalf("unexpected package: %q", loaded.DefaultPackage)
	}
}

func TestLoadDefaultConfigWhenMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.OutputFormat != "json" {
		t.Fatalf("unexpected output format: %q", loaded.OutputFormat)
	}
}

func TestLoadFromFileInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if _, err := loadFromFile(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigSaveWritesFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := &Config{DefaultPackage: "com.example.app"}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	paths := GetPaths()
	data, err := os.ReadFile(paths.ConfigFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("expected valid json")
	}
}

func TestConfigSaveWriteError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths := GetPaths()
	orig := osWriteFile
	osWriteFile = func(path string, data []byte, perm os.FileMode) error {
		if path == paths.ConfigFile {
			return os.ErrPermission
		}
		return orig(path, data, perm)
	}
	t.Cleanup(func() {
		osWriteFile = orig
	})
	cfg := &Config{DefaultPackage: "com.example.app"}
	if err := cfg.Save(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigSaveMkdirError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths := GetPaths()
	orig := osMkdirAll
	osMkdirAll = func(path string, perm os.FileMode) error {
		if path == paths.ConfigDir {
			return os.ErrPermission
		}
		return orig(path, perm)
	}
	t.Cleanup(func() {
		osMkdirAll = orig
	})
	cfg := &Config{DefaultPackage: "com.example.app"}
	if err := cfg.Save(); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigSaveMarshalError(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	orig := jsonMarshalIndent
	jsonMarshalIndent = func(v interface{}, prefix, indent string) ([]byte, error) {
		return nil, os.ErrInvalid
	}
	defer func() {
		jsonMarshalIndent = orig
	}()
	cfg := &Config{DefaultPackage: "com.example.app"}
	if err := cfg.Save(); err == nil {
		t.Fatal("expected error")
	}
}

func TestEnvAccessors(t *testing.T) {
	t.Setenv(EnvServiceAccountKey, "key")
	t.Setenv(EnvPackage, "pkg")
	t.Setenv(EnvTimeout, "10")
	t.Setenv(EnvStoreTokens, "auto")
	if GetEnvServiceAccountKey() != "key" {
		t.Fatal("unexpected service account key")
	}
	if GetEnvPackage() != "pkg" {
		t.Fatal("unexpected package")
	}
	if GetEnvTimeout() != "10" {
		t.Fatal("unexpected timeout")
	}
	if GetEnvStoreTokens() != "auto" {
		t.Fatal("unexpected store tokens")
	}
}

func TestInitProjectCreatesFiles(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	if err := InitProject(project); err != nil {
		t.Fatalf("InitProject failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, "assets", "en-US", "phone")); err != nil {
		t.Fatalf("assets dir missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, "release-notes.json")); err != nil {
		t.Fatalf("release-notes.json missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".gitignore")); err != nil {
		t.Fatalf(".gitignore missing: %v", err)
	}
	paths := GetPaths()
	if _, err := os.Stat(paths.ConfigFile); err != nil {
		t.Fatalf("config file missing: %v", err)
	}
}

func TestInitProjectConfigDirError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	paths := GetPaths()
	orig := osMkdirAll
	osMkdirAll = func(path string, perm os.FileMode) error {
		if path == paths.ConfigDir {
			return os.ErrPermission
		}
		return orig(path, perm)
	}
	t.Cleanup(func() {
		osMkdirAll = orig
	})
	if err := InitProject(project); err == nil {
		t.Fatal("expected error")
	}
}

func TestInitProjectCacheDirError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	paths := GetPaths()
	orig := osMkdirAll
	osMkdirAll = func(path string, perm os.FileMode) error {
		if path == paths.CacheDir {
			return os.ErrPermission
		}
		return orig(path, perm)
	}
	t.Cleanup(func() {
		osMkdirAll = orig
	})
	if err := InitProject(project); err == nil {
		t.Fatal("expected error")
	}
}

func TestInitProjectSaveError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	paths := GetPaths()
	orig := osWriteFile
	osWriteFile = func(path string, data []byte, perm os.FileMode) error {
		if path == paths.ConfigFile {
			return os.ErrPermission
		}
		return orig(path, data, perm)
	}
	t.Cleanup(func() {
		osWriteFile = orig
	})
	if err := InitProject(project); err == nil {
		t.Fatal("expected error")
	}
}

func TestInitProjectAssetsError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(project, "assets"), []byte("file"), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := InitProject(project); err == nil {
		t.Fatal("expected error")
	}
}

func TestInitProjectReleaseNotesError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(project, "release-notes.json")
	if err := os.WriteFile(path, []byte("existing"), 0400); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := InitProject(project); err == nil {
		t.Fatal("expected error")
	}
}

func TestInitProjectGitignoreError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(project, ".gitignore")
	if err := os.WriteFile(path, []byte("existing"), 0400); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	if err := InitProject(project); err == nil {
		t.Fatal("expected error")
	}
}
