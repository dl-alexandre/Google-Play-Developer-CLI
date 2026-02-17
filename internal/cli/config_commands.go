// Package cli provides config commands for gpd.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
	"github.com/dl-alexandre/gpd/internal/storage"
)

func validatePath(path string) error {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal")
	}
	return nil
}

func (c *CLI) addConfigCommands() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration commands",
		Long:  "Manage gpd configuration and system health.",
	}

	// config init
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize project configuration",
		Long:  "Scaffold config files, sample release-notes.json, assets/ layout, and .gitignore.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.configInit(cmd)
		},
	}

	// config doctor
	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose configuration and credential issues",
		Long:  "Check configuration, credentials, and system health.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.configDoctor(cmd)
		},
	}

	// config path
	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "Show configuration file locations",
		Long:  "Display the paths used for configuration and cache files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.configPath(cmd)
		},
	}

	// config get
	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a configuration value",
		Long:  "Get the value of a configuration key.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.configGet(cmd, args[0])
		},
	}

	// config set
	setCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Long:  "Set a configuration key to a value.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.configSet(cmd, args[0], args[1])
		},
	}

	// config print
	printCmd := &cobra.Command{
		Use:     "print",
		Short:   "Print resolved configuration",
		Long:    "Print the fully resolved configuration showing precedence.",
		Aliases: []string{"show"},
	}
	printCmd.Flags().Bool("resolved", false, "Show precedence resolution")
	printCmd.RunE = func(cmd *cobra.Command, args []string) error {
		resolved, _ := cmd.Flags().GetBool("resolved")
		return c.configPrint(cmd, resolved)
	}

	// config export
	var exportOutput string
	var exportIncludePaths bool
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export configuration to file",
		Long:  "Export safe configuration values to a JSON or YAML file (based on output extension).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.configExport(cmd, exportOutput, exportIncludePaths)
		},
	}
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "gpd-config.json", "Output file path")
	exportCmd.Flags().BoolVar(&exportIncludePaths, "include-paths", false, "Include serviceAccountKeyPath (warning: may be machine-specific)")

	// config import
	var importMerge bool
	importCmd := &cobra.Command{
		Use:   "import <file>",
		Short: "Import configuration from file",
		Long:  "Import configuration values from a JSON or YAML file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.configImport(cmd, args[0], importMerge)
		},
	}
	importCmd.Flags().BoolVar(&importMerge, "merge", true, "Merge with existing config (default: true, use --merge=false to replace)")

	// config completion
	completionCmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for gpd.

To load completions:

Bash:
  $ source <(gpd config completion bash)

Zsh:
  $ source <(gpd config completion zsh)

Fish:
  $ gpd config completion fish | source
`,
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.configCompletion(cmd, args[0])
		},
	}

	configCmd.AddCommand(initCmd, doctorCmd, pathCmd, getCmd, setCmd, printCmd, exportCmd, importCmd, completionCmd)
	c.rootCmd.AddCommand(configCmd)
}

func (c *CLI) configInit(_ *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if err := config.InitProject(cwd); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	paths := config.GetPaths()
	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"configDir":  paths.ConfigDir,
		"cacheDir":   paths.CacheDir,
		"configFile": paths.ConfigFile,
		"assetsDir":  filepath.Join(cwd, "assets"),
		"created": []string{
			paths.ConfigFile,
			filepath.Join(cwd, "release-notes.json"),
			filepath.Join(cwd, "assets"),
			filepath.Join(cwd, ".gitignore"),
		},
	})
	return c.Output(result.WithServices("config"))
}

type doctorResult struct {
	issues []string
	check  map[string]interface{}
}

func checkCredentials(envKey, gacPath string, parsedConfig *config.Config, configLoaded bool) doctorResult {
	result := doctorResult{
		check: map[string]interface{}{},
	}
	credentialsChecks := result.check

	// Check envServiceAccountKey
	if envKey != "" {
		valid, reason, fields := validateServiceAccountJSON([]byte(envKey))
		entry := map[string]interface{}{
			"set":   true,
			"valid": valid,
		}
		if !valid {
			entry["reason"] = reason
			if len(fields) > 0 {
				entry["missingFields"] = fields
			}
			result.issues = append(result.issues, "GPD_SERVICE_ACCOUNT_KEY is not a valid service account key")
		}
		credentialsChecks["envServiceAccountKey"] = entry
	} else {
		credentialsChecks["envServiceAccountKey"] = map[string]interface{}{"set": false}
	}

	// Check GOOGLE_APPLICATION_CREDENTIALS
	gacEntry := map[string]interface{}{"set": gacPath != ""}
	if gacPath != "" {
		gacEntry["path"] = gacPath
		if err := validatePath(gacPath); err != nil {
			gacEntry["valid"] = false
			gacEntry["error"] = err.Error()
			result.issues = append(result.issues, "GOOGLE_APPLICATION_CREDENTIALS has invalid path: "+err.Error())
		} else if _, err := os.Stat(gacPath); err != nil { // #nosec G703 -- Path validated above
			gacEntry["exists"] = false
			result.issues = append(result.issues, "GOOGLE_APPLICATION_CREDENTIALS points to a missing file")
		} else {
			gacEntry["exists"] = true
			data, err := os.ReadFile(gacPath) // #nosec G703 -- Path validated above
			if err != nil {
				gacEntry["readable"] = false
				result.issues = append(result.issues, "GOOGLE_APPLICATION_CREDENTIALS file is not readable")
			} else {
				valid, reason, fields := validateServiceAccountJSON(data)
				gacEntry["readable"] = true
				gacEntry["valid"] = valid
				if !valid {
					gacEntry["reason"] = reason
					if len(fields) > 0 {
						gacEntry["missingFields"] = fields
					}
					result.issues = append(result.issues, "GOOGLE_APPLICATION_CREDENTIALS does not contain a valid service account key")
				}
			}
		}
	}
	credentialsChecks["googleApplicationCredentials"] = gacEntry

	// Check serviceAccountKeyPath from config
	if configLoaded && parsedConfig.ServiceAccountKeyPath != "" {
		keyEntry := map[string]interface{}{"path": parsedConfig.ServiceAccountKeyPath}
		if _, err := os.Stat(parsedConfig.ServiceAccountKeyPath); err != nil {
			keyEntry["exists"] = false
			result.issues = append(result.issues, "serviceAccountKeyPath points to a missing file")
		} else {
			keyEntry["exists"] = true
			data, err := os.ReadFile(parsedConfig.ServiceAccountKeyPath)
			if err != nil {
				keyEntry["readable"] = false
				result.issues = append(result.issues, "serviceAccountKeyPath file is not readable")
			} else {
				valid, reason, fields := validateServiceAccountJSON(data)
				keyEntry["readable"] = true
				keyEntry["valid"] = valid
				if !valid {
					keyEntry["reason"] = reason
					if len(fields) > 0 {
						keyEntry["missingFields"] = fields
					}
					result.issues = append(result.issues, "serviceAccountKeyPath does not contain a valid service account key")
				}
			}
		}
		credentialsChecks["serviceAccountKeyPath"] = keyEntry
	} else {
		credentialsChecks["serviceAccountKeyPath"] = map[string]interface{}{"set": false}
	}

	return result
}

func checkConfigFile(path string) (config.Config, bool, doctorResult) {
	var parsedConfig config.Config
	result := doctorResult{
		check: map[string]interface{}{"path": path},
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		result.issues = append(result.issues, "Config file does not exist (run 'gpd config init')")
		result.check["exists"] = false
		return parsedConfig, false, result
	}

	result.check["exists"] = true
	data, err := os.ReadFile(path)
	if err != nil {
		result.issues = append(result.issues, "Config file is not readable")
		result.check["readable"] = false
		return parsedConfig, false, result
	}

	result.check["readable"] = true
	if err := json.Unmarshal(data, &parsedConfig); err != nil {
		result.issues = append(result.issues, "Config file contains invalid JSON")
		result.check["valid"] = false
		return parsedConfig, false, result
	}

	result.check["valid"] = true
	return parsedConfig, true, result
}

func (c *CLI) configDoctor(_ *cobra.Command) error {
	paths := config.GetPaths()
	issues := []string{}
	checks := map[string]interface{}{}

	// Check config directory
	if _, err := os.Stat(paths.ConfigDir); os.IsNotExist(err) {
		issues = append(issues, "Config directory does not exist")
		checks["configDir"] = map[string]interface{}{"exists": false, "path": paths.ConfigDir}
	} else {
		checks["configDir"] = map[string]interface{}{"exists": true, "path": paths.ConfigDir}
	}

	// Check config file
	parsedConfig, configLoaded, configFileIssues := checkConfigFile(paths.ConfigFile)
	issues = append(issues, configFileIssues.issues...)
	checks["configFile"] = configFileIssues.check

	// Check cache directory
	if _, err := os.Stat(paths.CacheDir); os.IsNotExist(err) {
		checks["cacheDir"] = map[string]interface{}{"exists": false, "path": paths.CacheDir}
	} else {
		checks["cacheDir"] = map[string]interface{}{"exists": true, "path": paths.CacheDir}
	}

	// Check secure storage
	secureStorage := storage.New()
	checks["secureStorage"] = map[string]interface{}{
		"available": secureStorage.Available(),
		"platform":  storage.Platform(),
	}
	if !secureStorage.Available() {
		issues = append(issues, "Secure storage not available on this platform")
	}

	storeTokensValue := config.GetEnvStoreTokens()
	if storeTokensValue == "" && configLoaded {
		storeTokensValue = parsedConfig.StoreTokens
	}
	if storeTokensValue != "" {
		checks["storeTokens"] = map[string]interface{}{
			"value":                  storeTokensValue,
			"secureStorageAvailable": secureStorage.Available(),
		}
		if storeTokensValue == "secure" && !secureStorage.Available() {
			issues = append(issues, "Store tokens is set to secure but secure storage is unavailable")
		}
	}

	// Check environment variables
	envChecks := map[string]interface{}{}
	envKey := config.GetEnvServiceAccountKey()
	if envKey != "" {
		envChecks["GPD_SERVICE_ACCOUNT_KEY"] = "set (value hidden)"
	} else {
		envChecks["GPD_SERVICE_ACCOUNT_KEY"] = "not set"
	}
	if pkg := config.GetEnvPackage(); pkg != "" {
		envChecks["GPD_PACKAGE"] = pkg
	} else {
		envChecks["GPD_PACKAGE"] = "not set"
	}
	gacPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	envChecks["GOOGLE_APPLICATION_CREDENTIALS"] = gacPath
	checks["environment"] = envChecks

	credentialsResult := checkCredentials(envKey, gacPath, &parsedConfig, configLoaded)
	issues = append(issues, credentialsResult.issues...)
	checks["credentials"] = credentialsResult.check

	// Check CI detection
	checks["ci"] = map[string]interface{}{
		"detected": config.DetectCI(),
	}

	// Check for multiple gpd binaries
	binaries := findGPDBinaries()
	checks["binaries"] = binaries
	if len(binaries) > 1 {
		issues = append(issues, fmt.Sprintf("Multiple gpd binaries found in PATH: %v", binaries))
	}

	result := output.NewResult(map[string]interface{}{
		"healthy": len(issues) == 0,
		"issues":  issues,
		"checks":  checks,
	})
	return c.Output(result.WithServices("config"))
}

func findGPDBinaries() []string {
	binaries := []string{}
	pathEnv := os.Getenv("PATH")
	var separator string
	if runtime.GOOS == "windows" {
		separator = ";"
	} else {
		separator = ":"
	}

	for _, dir := range strings.Split(pathEnv, separator) {
		if dir == "" {
			continue
		}
		if err := validatePath(dir); err != nil {
			continue
		}
		gpdPath := filepath.Join(dir, "gpd")
		if runtime.GOOS == "windows" {
			gpdPath = filepath.Join(dir, "gpd.exe")
		}
		if _, err := os.Stat(gpdPath); err == nil { // #nosec G703 -- Path validated above
			binaries = append(binaries, gpdPath)
		}
	}
	return binaries
}

func validateServiceAccountJSON(data []byte) (valid bool, email string, scopes []string) {
	var keyData struct {
		Type        string `json:"type"`
		ClientEmail string `json:"client_email"`
		ClientID    string `json:"client_id"`
		PrivateKey  string `json:"private_key"` // #nosec G117 -- Service account private key, required for OAuth
		TokenURI    string `json:"token_uri"`
	}
	if err := json.Unmarshal(data, &keyData); err != nil {
		return false, "invalid_json", nil
	}
	if keyData.Type != "service_account" {
		return false, "invalid_type", nil
	}
	missing := []string{}
	if keyData.ClientEmail == "" {
		missing = append(missing, "client_email")
	}
	if keyData.ClientID == "" {
		missing = append(missing, "client_id")
	}
	if keyData.PrivateKey == "" {
		missing = append(missing, "private_key")
	}
	if keyData.TokenURI == "" {
		missing = append(missing, "token_uri")
	}
	if len(missing) > 0 {
		return false, "missing_fields", missing
	}
	return true, "", nil
}

func (c *CLI) configPath(_ *cobra.Command) error {
	paths := config.GetPaths()
	if strings.EqualFold(c.outputFormat, "table") {
		t := tablewriter.NewWriter(c.stdout)
		t.Header([]string{"key", "path"})
		if err := t.Append([]string{"configDir", paths.ConfigDir}); err != nil {
			return err
		}
		if err := t.Append([]string{"cacheDir", paths.CacheDir}); err != nil {
			return err
		}
		if err := t.Append([]string{"configFile", paths.ConfigFile}); err != nil {
			return err
		}
		if err := t.Append([]string{"legacyDir", config.GetLegacyConfigDir()}); err != nil {
			return err
		}
		if err := t.Render(); err != nil {
			return err
		}
		return nil
	}

	result := output.NewResult(map[string]interface{}{
		"configDir":  paths.ConfigDir,
		"cacheDir":   paths.CacheDir,
		"configFile": paths.ConfigFile,
		"legacyDir":  config.GetLegacyConfigDir(),
	})
	return c.Output(result.WithServices("config"))
}

func (c *CLI) configGet(_ *cobra.Command, key string) error {
	cfg, err := config.Load()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	// Convert config to map for key lookup
	data, _ := json.Marshal(cfg)
	var cfgMap map[string]interface{}
	if err := json.Unmarshal(data, &cfgMap); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	value, ok := cfgMap[key]
	if !ok {
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("key not found: %s", key)))
	}

	result := output.NewResult(map[string]interface{}{
		"key":   key,
		"value": value,
	})
	return c.Output(result.WithServices("config"))
}

func (c *CLI) configSet(_ *cobra.Command, key, value string) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Convert config to map for key setting
	data, _ := json.Marshal(cfg)
	var cfgMap map[string]interface{}
	if err := json.Unmarshal(data, &cfgMap); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	cfgMap[key] = value

	// Convert back to config
	newData, _ := json.Marshal(cfgMap)
	if err := json.Unmarshal(newData, cfg); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if err := cfg.Save(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"key":   key,
		"value": value,
		"saved": true,
	})
	return c.Output(result.WithServices("config"))
}

func (c *CLI) configPrint(_ *cobra.Command, resolved bool) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if resolved {
		// Show precedence resolution
		result := output.NewResult(map[string]interface{}{
			"config": cfg,
			"precedence": map[string]interface{}{
				"1_flags":       "command-line flags",
				"2_environment": "environment variables",
				"3_config":      "configuration file",
				"4_defaults":    "built-in defaults",
			},
			"resolved": map[string]interface{}{
				"package":     c.packageName,
				"output":      c.outputFormat,
				"timeout":     c.timeout.String(),
				"storeTokens": c.storeTokens,
			},
		})
		return c.Output(result.WithServices("config"))
	}

	result := output.NewResult(cfg)
	return c.Output(result.WithServices("config"))
}

func (c *CLI) configCompletion(_ *cobra.Command, shell string) error {
	var err error
	switch shell {
	case "bash":
		err = c.rootCmd.GenBashCompletion(c.stdout)
	case "zsh":
		err = c.rootCmd.GenZshCompletion(c.stdout)
	case "fish":
		err = c.rootCmd.GenFishCompletion(c.stdout, true)
	default:
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("unsupported shell: %s", shell)).
			WithHint("Supported shells: bash, zsh, fish"))
	}

	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return nil
}

type ConfigExport struct {
	Version    string                 `json:"version"`
	ExportedAt string                 `json:"exportedAt"`
	Config     map[string]interface{} `json:"config"`
	Metadata   map[string]interface{} `json:"metadata"`
}

func (c *CLI) configExport(_ *cobra.Command, outputPath string, includePaths bool) error {
	cfg, err := config.Load()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to load config: %v", err)))
	}

	exportData := map[string]interface{}{}
	warnings := []string{}

	if cfg.DefaultPackage != "" {
		exportData["defaultPackage"] = cfg.DefaultPackage
	}
	if cfg.OutputFormat != "" {
		exportData["outputFormat"] = cfg.OutputFormat
	}
	if cfg.TimeoutSeconds > 0 {
		exportData["timeoutSeconds"] = cfg.TimeoutSeconds
	}
	if cfg.StoreTokens != "" {
		exportData["storeTokens"] = cfg.StoreTokens
	}
	if len(cfg.RateLimits) > 0 {
		exportData["rateLimits"] = cfg.RateLimits
	}
	if cfg.TesterLimits != nil {
		exportData["testerLimits"] = cfg.TesterLimits
	}

	if cfg.ServiceAccountKeyPath != "" {
		if includePaths {
			exportData["serviceAccountKeyPath"] = cfg.ServiceAccountKeyPath
			warnings = append(warnings, "serviceAccountKeyPath included - may be machine-specific")
		} else {
			warnings = append(warnings, "serviceAccountKeyPath not included - use --include-paths to export")
		}
	}

	export := ConfigExport{
		Version:    "1.0",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Config:     exportData,
		Metadata: map[string]interface{}{
			"platform":         runtime.GOOS,
			"credentialOrigin": cfg.CredentialOrigin,
			"warnings":         warnings,
		},
	}

	data, err := c.marshalConfigExport(outputPath, export)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to marshal config: %v", err)).
			WithHint("Use .json, .yaml, or .yml extension for config export"))
	}

	if err := os.WriteFile(outputPath, data, 0600); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to write file: %v", err)))
	}

	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"exported":   len(exportData),
		"outputPath": outputPath,
		"warnings":   warnings,
	})
	return c.Output(result)
}

func (c *CLI) configImport(_ *cobra.Command, inputPath string, merge bool) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("failed to read file: %v", err)))
	}

	var importData ConfigExport
	if err := c.unmarshalConfigExport(inputPath, data, &importData); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid config file: %v", err)).
			WithHint("Expected JSON/YAML format from 'gpd config export'"))
	}

	cfg := c.loadOrCreateConfig(merge)
	imported := c.applyImportedConfig(cfg, importData.Config)

	if err := cfg.Save(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to save config: %v", err)))
	}

	result := output.NewResult(map[string]interface{}{
		"success":  true,
		"imported": imported,
		"merge":    merge,
		"version":  importData.Version,
	})
	return c.Output(result)
}

func (c *CLI) loadOrCreateConfig(merge bool) *config.Config {
	if merge {
		cfg, err := config.Load()
		if err == nil {
			return cfg
		}
	}
	return &config.Config{}
}

func (c *CLI) applyImportedConfig(cfg *config.Config, data map[string]interface{}) []string {
	imported := []string{}

	if val, ok := data["defaultPackage"].(string); ok && val != "" {
		cfg.DefaultPackage = val
		imported = append(imported, "defaultPackage")
	}
	if val, ok := data["outputFormat"].(string); ok && val != "" {
		cfg.OutputFormat = val
		imported = append(imported, "outputFormat")
	}
	if val, ok := data["timeoutSeconds"].(float64); ok && val > 0 {
		cfg.TimeoutSeconds = int(val)
		imported = append(imported, "timeoutSeconds")
	}
	if val, ok := data["storeTokens"].(string); ok && val != "" {
		cfg.StoreTokens = val
		imported = append(imported, "storeTokens")
	}
	if val, ok := data["rateLimits"].(map[string]interface{}); ok && len(val) > 0 {
		cfg.RateLimits = c.parseRateLimits(val)
		imported = append(imported, "rateLimits")
	}
	if val, ok := data["testerLimits"].(map[string]interface{}); ok && len(val) > 0 {
		cfg.TesterLimits = c.parseTesterLimits(val)
		imported = append(imported, "testerLimits")
	}
	if val, ok := data["serviceAccountKeyPath"].(string); ok && val != "" {
		cfg.ServiceAccountKeyPath = val
		imported = append(imported, "serviceAccountKeyPath")
	}

	return imported
}

func (c *CLI) parseRateLimits(val map[string]interface{}) map[string]string {
	rateLimits := make(map[string]string)
	for k, v := range val {
		if strVal, ok := v.(string); ok {
			rateLimits[k] = strVal
		}
	}
	return rateLimits
}

func (c *CLI) parseTesterLimits(val map[string]interface{}) *config.TesterLimits {
	limits := config.DefaultTesterLimits()
	if internal, ok := val["internal"].(float64); ok {
		limits.Internal = int(internal)
	}
	if alpha, ok := val["alpha"].(float64); ok {
		limits.Alpha = int(alpha)
	}
	if beta, ok := val["beta"].(float64); ok {
		limits.Beta = int(beta)
	}
	return limits
}

func (c *CLI) marshalConfigExport(outputPath string, export ConfigExport) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(outputPath))
	if ext == ".yaml" || ext == ".yml" {
		return yaml.Marshal(export)
	}
	if ext == "" || ext == ".json" {
		return json.MarshalIndent(export, "", "  ")
	}
	return nil, fmt.Errorf("unsupported config export extension: %s", ext)
}

func (c *CLI) unmarshalConfigExport(inputPath string, data []byte, out *ConfigExport) error {
	ext := strings.ToLower(filepath.Ext(inputPath))
	if ext == ".yaml" || ext == ".yml" {
		return yaml.Unmarshal(data, out)
	}
	if ext == "" || ext == ".json" {
		return json.Unmarshal(data, out)
	}

	if err := json.Unmarshal(data, out); err == nil {
		return nil
	}
	return yaml.Unmarshal(data, out)
}
