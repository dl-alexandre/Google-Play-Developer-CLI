// Package config provides configuration management for gpd.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
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
	ActiveProfile         string            `json:"activeProfile,omitempty"`
}

// TesterLimits defines limits for different tester types.
type TesterLimits struct {
	Internal int `json:"internal"` // Default: 200
	Alpha    int `json:"alpha"`    // Default: -1 (unlimited)
	Beta     int `json:"beta"`     // Default: -1 (unlimited)
}

// ConfigValidationResult holds validation warnings and errors.
type ConfigValidationResult struct {
	Warnings []string
	Errors   []string
}

func isValidPackageName(name string) bool {
	if name == "" {
		return false
	}
	if !unicode.IsLower(rune(name[0])) {
		return false
	}
	for _, r := range name {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' && r != '.' {
			return false
		}
	}
	return true
}

// Validate checks the config for invalid values.
func (c *Config) Validate() *ConfigValidationResult {
	result := &ConfigValidationResult{}

	if c.StoreTokens != "" && c.StoreTokens != "auto" && c.StoreTokens != "never" && c.StoreTokens != "secure" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("invalid storeTokens value %q, using default", c.StoreTokens))
		c.StoreTokens = "auto"
	}

	validFormats := map[string]bool{"json": true, "table": true, "markdown": true, "csv": true}
	if c.OutputFormat != "" && !validFormats[c.OutputFormat] {
		result.Warnings = append(result.Warnings, fmt.Sprintf("invalid outputFormat %q, using default", c.OutputFormat))
		c.OutputFormat = "json"
	}

	if c.TimeoutSeconds != 0 && (c.TimeoutSeconds < 5 || c.TimeoutSeconds > 300) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("timeoutSeconds %d out of range [5-300], using default", c.TimeoutSeconds))
		c.TimeoutSeconds = 30
	}

	if c.TesterLimits != nil {
		if c.TesterLimits.Internal < -1 {
			result.Warnings = append(result.Warnings, "testerLimits.internal cannot be negative, using default")
			c.TesterLimits.Internal = 200
		}
		if c.TesterLimits.Alpha < -1 {
			result.Warnings = append(result.Warnings, "testerLimits.alpha cannot be negative, using default")
			c.TesterLimits.Alpha = -1
		}
		if c.TesterLimits.Beta < -1 {
			result.Warnings = append(result.Warnings, "testerLimits.beta cannot be negative, using default")
			c.TesterLimits.Beta = -1
		}
	}

	if c.ActiveProfile != "" {
		for _, r := range c.ActiveProfile {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '-' && r != '_' {
				result.Errors = append(result.Errors, fmt.Sprintf("activeProfile contains invalid character %q", string(r)))
			}
		}
	}

	if c.DefaultPackage != "" && !isValidPackageName(c.DefaultPackage) {
		result.Warnings = append(result.Warnings, fmt.Sprintf("defaultPackage %q may be invalid", c.DefaultPackage))
	}

	return result
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
var osMkdirAll = os.MkdirAll
var osWriteFile = os.WriteFile

func getHomeDirForOS(goos string) string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if goos == "windows" {
		if home := os.Getenv("USERPROFILE"); home != "" {
			return home
		}
		if drive, path := os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"); drive != "" && path != "" {
			return drive + path
		}
	}
	home, _ := os.UserHomeDir()
	return home
}

// GetPaths returns the OS-appropriate configuration paths.
func GetPaths() Paths {
	return getPathsForOS(runtimeGOOS)
}

func getPathsForOS(goos string) Paths {
	var configDir, cacheDir string

	switch goos {
	case "darwin":
		home := getHomeDirForOS(goos)
		configDir = filepath.Join(home, "Library", "Application Support", "gpd")
		cacheDir = filepath.Join(home, "Library", "Caches", "gpd")
	case "windows":
		appData := os.Getenv("APPDATA")
		localAppData := os.Getenv("LOCALAPPDATA")
		if appData == "" || localAppData == "" {
			home := getHomeDirForOS(goos)
			if appData == "" && home != "" {
				appData = filepath.Join(home, "AppData", "Roaming")
			}
			if localAppData == "" && home != "" {
				localAppData = filepath.Join(home, "AppData", "Local")
			}
		}
		configDir = filepath.Join(appData, "gpd")
		cacheDir = filepath.Join(localAppData, "gpd")
	default:
		home := getHomeDirForOS(goos)
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
	home := getHomeDirForOS(runtimeGOOS)
	return filepath.Join(home, ".gpd")
}

// Load loads the configuration from the config file.
// Returns loadWarn for parse errors (not for missing files), returns error only for validation failures.
func Load() (cfg *Config, loadWarn error) {
	paths := GetPaths()

	cfg, loadErr := loadFromFile(paths.ConfigFile)
	if loadErr == nil {
		validatedCfg, valErr := validateAndNormalize(cfg, paths.ConfigFile)
		if valErr != nil {
			return nil, valErr
		}
		return validatedCfg, nil
	}

	// Try legacy location
	legacyConfig := filepath.Join(GetLegacyConfigDir(), "config.json")
	cfg, err := loadFromFile(legacyConfig)
	if err == nil {
		validatedCfg, valErr := validateAndNormalize(cfg, legacyConfig)
		if valErr != nil {
			return nil, valErr
		}
		return validatedCfg, nil
	}

	// Both configs failed - build warning for non-not-exist errors
	var warnings []string
	if !os.IsNotExist(loadErr) {
		warnings = append(warnings, fmt.Sprintf("failed to load config from %s: %v", paths.ConfigFile, loadErr))
	}
	if !os.IsNotExist(err) {
		warnings = append(warnings, fmt.Sprintf("failed to load config from %s: %v", legacyConfig, err))
	}

	if len(warnings) > 0 {
		return DefaultConfig(), fmt.Errorf("%s", strings.Join(warnings, "; "))
	}

	return DefaultConfig(), nil
}

func validateAndNormalize(cfg *Config, source string) (*Config, error) {
	validation := cfg.Validate()
	if len(validation.Errors) > 0 {
		return nil, fmt.Errorf("invalid config in %s: %s", source, strings.Join(validation.Errors, ", "))
	}
	// Warnings are handled by correcting the values in Validate()
	return cfg, nil
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
	if err := osMkdirAll(paths.ConfigDir, 0700); err != nil {
		return err
	}

	data, err := jsonMarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return osWriteFile(paths.ConfigFile, data, 0600)
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
	EnvOAuthClientID     = "GPD_CLIENT_ID"
	EnvOAuthClientSecret = "GPD_CLIENT_SECRET" //nolint:gosec // G101: This is an env var name, not credentials
	EnvPackage           = "GPD_PACKAGE"
	EnvAuthProfile       = "GPD_AUTH_PROFILE"
	EnvTimeout           = "GPD_TIMEOUT"
	EnvStoreTokens       = "GPD_STORE_TOKENS" //nolint:gosec // G101: This is an env var name, not credentials
	EnvCI                = "GPD_CI"
)

// GetEnvServiceAccountKey returns the service account key from environment.
func GetEnvServiceAccountKey() string {
	return os.Getenv(EnvServiceAccountKey)
}

func GetEnvOAuthClientID() string {
	return os.Getenv(EnvOAuthClientID)
}

func GetEnvOAuthClientSecret() string {
	return os.Getenv(EnvOAuthClientSecret)
}

// GetEnvPackage returns the default package from environment.
func GetEnvPackage() string {
	return os.Getenv(EnvPackage)
}

func GetEnvAuthProfile() string {
	return os.Getenv(EnvAuthProfile)
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
	if err := osMkdirAll(paths.ConfigDir, 0700); err != nil {
		return err
	}

	// Create cache directory
	if err := osMkdirAll(paths.CacheDir, 0700); err != nil {
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
			if err := osMkdirAll(catDir, 0755); err != nil {
				return err
			}
		}
	}

	// Create sample release-notes.json
	releaseNotes := map[string]string{
		"en-US": "Bug fixes and improvements.",
	}
	rnData, _ := json.MarshalIndent(releaseNotes, "", "  ")
	if err := osWriteFile(filepath.Join(dir, "release-notes.json"), rnData, 0644); err != nil {
		return err
	}

	// Create .gitignore
	gitignore := `# gpd sensitive files
*.json.key
service-account*.json
.gpd/
`
	if err := osWriteFile(filepath.Join(dir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return err
	}

	return nil
}
