// Package config provides configuration management for gpd.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config represents the gpd configuration.
type Config struct {
	DefaultPackage        string            `json:"defaultPackage,omitempty"`
	ServiceAccountKeyPath string            `json:"serviceAccountKeyPath,omitempty"`
	CredentialOrigin      string            `json:"credentialOrigin,omitempty"` // "adc", "keyfile", "env"
	OutputFormat          string            `json:"outputFormat,omitempty"`
	TimeoutSeconds        int               `json:"timeoutSeconds,omitempty"`
	StoreTokens           string            `json:"storeTokens,omitempty"` // "auto", "never", "secure"
	RateLimits            map[string]string `json:"rateLimits,omitempty"`
	TesterLimits          *TesterLimits     `json:"testerLimits,omitempty"`
}

// TesterLimits defines limits for different tester types.
type TesterLimits struct {
	Internal int `json:"internal"` // Default: 200
	Alpha    int `json:"alpha"`    // Default: -1 (unlimited)
	Beta     int `json:"beta"`     // Default: -1 (unlimited)
}

// DefaultTesterLimits returns the default tester limits.
func DefaultTesterLimits() *TesterLimits {
	return &TesterLimits{
		Internal: 200,
		Alpha:    -1,
		Beta:     -1,
	}
}

// Paths contains OS-specific configuration paths.
type Paths struct {
	ConfigDir  string
	CacheDir   string
	ConfigFile string
}

var runtimeGOOS = runtime.GOOS
var jsonMarshalIndent = json.MarshalIndent

// GetPaths returns the OS-appropriate configuration paths.
func GetPaths() Paths {
	return getPathsForOS(runtimeGOOS)
}

func getPathsForOS(goos string) Paths {
	var configDir, cacheDir string

	switch goos {
	case "darwin":
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, "Library", "Application Support", "gpd")
		cacheDir = filepath.Join(home, "Library", "Caches", "gpd")
	case "windows":
		configDir = filepath.Join(os.Getenv("APPDATA"), "gpd")
		cacheDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "gpd")
	default:
		home, _ := os.UserHomeDir()
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			configDir = filepath.Join(xdgConfig, "gpd")
		} else {
			configDir = filepath.Join(home, ".config", "gpd")
		}
		if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
			cacheDir = filepath.Join(xdgCache, "gpd")
		} else {
			cacheDir = filepath.Join(home, ".cache", "gpd")
		}
	}

	return Paths{
		ConfigDir:  configDir,
		CacheDir:   cacheDir,
		ConfigFile: filepath.Join(configDir, "config.json"),
	}
}

// GetLegacyConfigDir returns the legacy ~/.gpd directory path.
func GetLegacyConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gpd")
}

// Load loads the configuration from the config file.
func Load() (*Config, error) {
	paths := GetPaths()

	// Try primary config location
	cfg, err := loadFromFile(paths.ConfigFile)
	if err == nil {
		return cfg, nil
	}

	// Try legacy location
	legacyConfig := filepath.Join(GetLegacyConfigDir(), "config.json")
	cfg, err = loadFromFile(legacyConfig)
	if err == nil {
		return cfg, nil
	}

	// Return default config
	return DefaultConfig(), nil
}

func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves the configuration to the config file.
func (c *Config) Save() error {
	paths := GetPaths()

	// Ensure config directory exists
	if err := os.MkdirAll(paths.ConfigDir, 0700); err != nil {
		return err
	}

	data, err := jsonMarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(paths.ConfigFile, data, 0600)
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		OutputFormat:   "json",
		TimeoutSeconds: 30,
		StoreTokens:    "auto",
		TesterLimits:   DefaultTesterLimits(),
		RateLimits: map[string]string{
			"reviews.reply": "5s",
		},
	}
}

// DetectCI returns true if running in a CI environment.
func DetectCI() bool {
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"JENKINS_URL",
		"BUILDKITE",
		"CIRCLECI",
		"TRAVIS",
		"GITLAB_CI",
		"GPD_CI",
	}
	for _, env := range ciVars {
		if os.Getenv(env) != "" {
			return true
		}
	}
	return false
}

// Environment variable names for gpd configuration.
const (
	EnvServiceAccountKey = "GPD_SERVICE_ACCOUNT_KEY"
	EnvPackage           = "GPD_PACKAGE"
	EnvTimeout           = "GPD_TIMEOUT"
	EnvStoreTokens       = "GPD_STORE_TOKENS" //nolint:gosec // G101: This is an env var name, not credentials
	EnvCI                = "GPD_CI"
)

// GetEnvServiceAccountKey returns the service account key from environment.
func GetEnvServiceAccountKey() string {
	return os.Getenv(EnvServiceAccountKey)
}

// GetEnvPackage returns the default package from environment.
func GetEnvPackage() string {
	return os.Getenv(EnvPackage)
}

// GetEnvTimeout returns the timeout from environment.
func GetEnvTimeout() string {
	return os.Getenv(EnvTimeout)
}

// GetEnvStoreTokens returns the store tokens setting from environment.
func GetEnvStoreTokens() string {
	return os.Getenv(EnvStoreTokens)
}

// ValidTracks returns the list of valid track names.
func ValidTracks() []string {
	return []string{"internal", "alpha", "beta", "production"}
}

// IsValidTrack checks if a track name is valid.
func IsValidTrack(track string) bool {
	track = strings.ToLower(track)
	for _, t := range ValidTracks() {
		if t == track {
			return true
		}
	}
	return false
}

// NormalizeLocale converts locale codes from en_US to en-US format.
func NormalizeLocale(locale string) string {
	return strings.ReplaceAll(locale, "_", "-")
}

// InitProject initializes a new project configuration.
func InitProject(dir string) error {
	paths := GetPaths()

	// Create config directory
	if err := os.MkdirAll(paths.ConfigDir, 0700); err != nil {
		return err
	}

	// Create cache directory
	if err := os.MkdirAll(paths.CacheDir, 0700); err != nil {
		return err
	}

	// Create default config
	cfg := DefaultConfig()
	if err := cfg.Save(); err != nil {
		return err
	}

	// Create assets directory structure
	assetsDir := filepath.Join(dir, "assets")
	for _, locale := range []string{"en-US"} {
		for _, category := range []string{"phone", "tablet", "tv", "wear"} {
			catDir := filepath.Join(assetsDir, locale, category)
			if err := os.MkdirAll(catDir, 0755); err != nil {
				return err
			}
		}
	}

	// Create sample release-notes.json
	releaseNotes := map[string]string{
		"en-US": "Bug fixes and improvements.",
	}
	rnData, _ := json.MarshalIndent(releaseNotes, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "release-notes.json"), rnData, 0644); err != nil {
		return err
	}

	// Create .gitignore
	gitignore := `# gpd sensitive files
*.json.key
service-account*.json
.gpd/
`
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return err
	}

	return nil
}
