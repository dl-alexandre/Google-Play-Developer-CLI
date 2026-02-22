// Package cli provides Kong-compatible config commands for gpd.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
	"github.com/dl-alexandre/gpd/internal/storage"
	"gopkg.in/yaml.v3"
)

// Note: ConfigCmd is declared in kong_commands.go as a placeholder.
// The subcommands below implement the actual functionality for Kong CLI.

// ConfigCmd is defined in kong_commands.go. These are its subcommands:
// - Init: ConfigInitCmd
// - Doctor: ConfigDoctorCmd
// - Path: ConfigPathCmd
// - Get: ConfigGetCmd
// - Set: ConfigSetCmd
// - Print: ConfigPrintCmd
// - Export: ConfigExportCmd
// - Import: ConfigImportCmd
// - Completion: ConfigCompletionCmd

// ConfigInitCmd initializes project configuration.
type ConfigInitCmd struct{}

func (cmd *ConfigInitCmd) Run(globals *Globals) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}

	if err := config.InitProject(cwd); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, err.Error())
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

	return writeOutput(globals, result.WithServices("config"))
}

// ConfigDoctorCmd diagnoses configuration and credential issues.
type ConfigDoctorCmd struct{}

func (cmd *ConfigDoctorCmd) Run(globals *Globals) error {
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
	parsedConfig, configLoaded, configFileIssues := kongCheckConfigFile(paths.ConfigFile)
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

	credentialsResult := kongCheckDoctorCredentials(envKey, gacPath, &parsedConfig, configLoaded)
	issues = append(issues, credentialsResult.issues...)
	checks["credentials"] = credentialsResult.check

	// Check CI detection
	checks["ci"] = map[string]interface{}{
		"detected": config.DetectCI(),
	}

	// Check for multiple gpd binaries
	binaries := kongFindGPDBinaries()
	checks["binaries"] = binaries
	if len(binaries) > 1 {
		issues = append(issues, fmt.Sprintf("Multiple gpd binaries found in PATH: %v", binaries))
	}

	result := output.NewResult(map[string]interface{}{
		"healthy": len(issues) == 0,
		"issues":  issues,
		"checks":  checks,
	})
	return writeOutput(globals, result.WithServices("config"))
}

type kongDoctorResult struct {
	issues []string
	check  map[string]interface{}
}

func kongCheckConfigFile(path string) (config.Config, bool, kongDoctorResult) {
	var parsedConfig config.Config
	result := kongDoctorResult{
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

func kongCheckDoctorCredentials(envKey, gacPath string, parsedConfig *config.Config, configLoaded bool) kongDoctorResult {
	result := kongDoctorResult{
		check: map[string]interface{}{},
	}
	credentialsChecks := result.check

	if envKey != "" {
		valid, reason, fields := kongValidateServiceAccountJSON([]byte(envKey))
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

	gacEntry := map[string]interface{}{"set": gacPath != ""}
	if gacPath != "" {
		gacEntry["path"] = gacPath
		if err := kongValidatePath(gacPath); err != nil {
			gacEntry["valid"] = false
			gacEntry["error"] = err.Error()
			result.issues = append(result.issues, "GOOGLE_APPLICATION_CREDENTIALS has invalid path: "+err.Error())
		} else if _, err := os.Stat(gacPath); err != nil {
			gacEntry["exists"] = false
			result.issues = append(result.issues, "GOOGLE_APPLICATION_CREDENTIALS points to a missing file")
		} else {
			gacEntry["exists"] = true
			data, err := os.ReadFile(gacPath)
			if err != nil {
				gacEntry["readable"] = false
				result.issues = append(result.issues, "GOOGLE_APPLICATION_CREDENTIALS file is not readable")
			} else {
				valid, reason, fields := kongValidateServiceAccountJSON(data)
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
				valid, reason, fields := kongValidateServiceAccountJSON(data)
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

func kongFindGPDBinaries() []string {
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
		if err := kongValidatePath(dir); err != nil {
			continue
		}
		gpdPath := filepath.Join(dir, "gpd")
		if runtime.GOOS == "windows" {
			gpdPath = filepath.Join(dir, "gpd.exe")
		}
		if _, err := os.Stat(gpdPath); err == nil {
			binaries = append(binaries, gpdPath)
		}
	}
	return binaries
}

func kongValidateServiceAccountJSON(data []byte) (valid bool, email string, scopes []string) {
	var keyData struct {
		Type        string `json:"type"`
		ClientEmail string `json:"client_email"`
		ClientID    string `json:"client_id"`
		PrivateKey  string `json:"private_key"`
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

func kongValidatePath(path string) error {
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path contains directory traversal")
	}
	return nil
}

// ConfigPathCmd shows configuration file locations.
type ConfigPathCmd struct{}

func (cmd *ConfigPathCmd) Run(globals *Globals) error {
	paths := config.GetPaths()

	if strings.EqualFold(globals.Output, "table") {
		// Print table format manually
		fmt.Println("KEY\t\tPATH")
		fmt.Println("----\t\t----")
		fmt.Printf("configDir\t%s\n", paths.ConfigDir)
		fmt.Printf("cacheDir\t%s\n", paths.CacheDir)
		fmt.Printf("configFile\t%s\n", paths.ConfigFile)
		fmt.Printf("legacyDir\t%s\n", config.GetLegacyConfigDir())
		return nil
	}

	result := output.NewResult(map[string]interface{}{
		"configDir":  paths.ConfigDir,
		"cacheDir":   paths.CacheDir,
		"configFile": paths.ConfigFile,
		"legacyDir":  config.GetLegacyConfigDir(),
	})
	return writeOutput(globals, result.WithServices("config"))
}

// ConfigGetCmd gets a configuration value.
type ConfigGetCmd struct {
	Key string `arg:"" help:"Configuration key to get"`
}

func (cmd *ConfigGetCmd) Run(globals *Globals) error {
	cfg, err := config.Load()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}

	// Convert config to map for key lookup
	data, _ := json.Marshal(cfg)
	var cfgMap map[string]interface{}
	if err := json.Unmarshal(data, &cfgMap); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}

	value, ok := cfgMap[cmd.Key]
	if !ok {
		return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("key not found: %s", cmd.Key))
	}

	result := output.NewResult(map[string]interface{}{
		"key":   cmd.Key,
		"value": value,
	})
	return writeOutput(globals, result.WithServices("config"))
}

// ConfigSetCmd sets a configuration value.
type ConfigSetCmd struct {
	Key   string `arg:"" help:"Configuration key to set"`
	Value string `arg:"" help:"Value to set"`
}

func (cmd *ConfigSetCmd) Run(globals *Globals) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Convert config to map for key setting
	data, _ := json.Marshal(cfg)
	var cfgMap map[string]interface{}
	if err := json.Unmarshal(data, &cfgMap); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}

	cfgMap[cmd.Key] = cmd.Value

	// Convert back to config
	newData, _ := json.Marshal(cfgMap)
	if err := json.Unmarshal(newData, cfg); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}

	if err := cfg.Save(); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}

	result := output.NewResult(map[string]interface{}{
		"key":   cmd.Key,
		"value": cmd.Value,
		"saved": true,
	})
	return writeOutput(globals, result.WithServices("config"))
}

// ConfigPrintCmd prints the configuration.
type ConfigPrintCmd struct {
	Resolved bool `help:"Show precedence resolution"`
}

func (cmd *ConfigPrintCmd) Run(globals *Globals) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if cmd.Resolved {
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
				"package":     globals.Package,
				"output":      globals.Output,
				"timeout":     globals.Timeout.String(),
				"storeTokens": globals.StoreTokens,
			},
		})
		return writeOutput(globals, result.WithServices("config"))
	}

	result := output.NewResult(cfg)
	return writeOutput(globals, result.WithServices("config"))
}

// ConfigCompletionCmd generates shell completion scripts.
type ConfigCompletionCmd struct {
	Shell string `arg:"" help:"Shell type (bash, zsh, fish)" enum:"bash,zsh,fish"`
}

func (cmd *ConfigCompletionCmd) Run(globals *Globals) error {
	// Note: Kong doesn't have built-in shell completion like Cobra
	// We provide manual completion scripts
	switch cmd.Shell {
	case "bash":
		fmt.Println(`#!/bin/bash
# gpd bash completion
_gpd_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="auth config publish reviews vitals analytics purchases monetization permissions recovery apps games integrity migrate customapp grouping version --help --package --output --pretty --timeout --store-tokens --fields --quiet --verbose --key-path --profile"
    
    case "${prev}" in
        config)
            opts="init doctor path get set print export import completion"
            ;;
        auth)
            opts="login logout status check"
            ;;
        publish)
            opts="upload edit release"
            ;;
        reviews)
            opts="list get reply"
            ;;
        --output|-o)
            opts="json table markdown csv"
            ;;
        --store-tokens)
            opts="auto never secure"
            ;;
    esac
    
    COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
    return 0
}
complete -F _gpd_completions gpd`)
	case "zsh":
		fmt.Println(`#compdef gpd
# gpd zsh completion
_gpd() {
    local -a commands
    commands=(
        'auth:Authentication commands'
        'config:Configuration commands'
        'publish:Publishing commands'
        'reviews:Review management commands'
        'vitals:Android vitals commands'
        'analytics:Analytics commands'
        'purchases:Purchase verification commands'
        'monetization:Monetization commands'
        'permissions:Permissions management'
        'recovery:App recovery commands'
        'apps:App discovery commands'
        'games:Google Play Games services'
        'integrity:Play Integrity API commands'
        'migrate:Migration commands'
        'customapp:Custom app publishing'
        'grouping:App access grouping'
        'version:Show version information'
    )
    
    _arguments -C \\
        '(-p --package)'{-p,--package}'[App package name]:package:' \\
        '(-o --output)'{-o,--output}'[Output format]:format:(json table markdown csv)' \\
        '--pretty[Pretty print JSON output]' \\
        '--timeout[Network timeout]:timeout:' \\
        '--store-tokens[Token storage]:storage:(auto never secure)' \\
        '--fields[JSON field projection]:fields:' \\
        '--quiet[Suppress non-error output]' \\
        '(-v --verbose)'{-v,--verbose}'[Enable verbose logging]' \\
        '--key-path[Path to service account key file]:path:_files' \\
        '--profile[Configuration profile to use]:profile:' \\
        '1: :->command' \\
        '*::arg:->args'
    
    case "$line[1]" in
        config)
            local -a config_cmds
            config_cmds=(init doctor path get set print export import completion)
            _describe -t commands 'config command' config_cmds
            ;;
        auth)
            local -a auth_cmds
            auth_cmds=(login logout status check)
            _describe -t commands 'auth command' auth_cmds
            ;;
        publish)
            local -a publish_cmds
            publish_cmds=(upload edit release)
            _describe -t commands 'publish command' publish_cmds
            ;;
    esac
}
compdef _gpd gpd`)
	case "fish":
		fmt.Println(`# gpd fish completion
complete -c gpd -f

# Global flags
complete -c gpd -l package -s p -d "App package name"
complete -c gpd -l output -s o -d "Output format" -a "json table markdown csv"
complete -c gpd -l pretty -d "Pretty print JSON output"
complete -c gpd -l timeout -d "Network timeout"
complete -c gpd -l store-tokens -d "Token storage" -a "auto never secure"
complete -c gpd -l fields -d "JSON field projection"
complete -c gpd -l quiet -d "Suppress non-error output"
complete -c gpd -l verbose -s v -d "Enable verbose logging"
complete -c gpd -l key-path -d "Path to service account key file"
complete -c gpd -l profile -d "Configuration profile to use"

# Commands
complete -c gpd -n "__fish_use_subcommand" -a auth -d "Authentication commands"
complete -c gpd -n "__fish_use_subcommand" -a config -d "Configuration commands"
complete -c gpd -n "__fish_use_subcommand" -a publish -d "Publishing commands"
complete -c gpd -n "__fish_use_subcommand" -a reviews -d "Review management commands"
complete -c gpd -n "__fish_use_subcommand" -a vitals -d "Android vitals commands"
complete -c gpd -n "__fish_use_subcommand" -a analytics -d "Analytics commands"
complete -c gpd -n "__fish_use_subcommand" -a purchases -d "Purchase verification commands"
complete -c gpd -n "__fish_use_subcommand" -a monetization -d "Monetization commands"
complete -c gpd -n "__fish_use_subcommand" -a permissions -d "Permissions management"
complete -c gpd -n "__fish_use_subcommand" -a recovery -d "App recovery commands"
complete -c gpd -n "__fish_use_subcommand" -a apps -d "App discovery commands"
complete -c gpd -n "__fish_use_subcommand" -a games -d "Google Play Games services"
complete -c gpd -n "__fish_use_subcommand" -a integrity -d "Play Integrity API commands"
complete -c gpd -n "__fish_use_subcommand" -a migrate -d "Migration commands"
complete -c gpd -n "__fish_use_subcommand" -a customapp -d "Custom app publishing"
complete -c gpd -n "__fish_use_subcommand" -a grouping -d "App access grouping"
complete -c gpd -n "__fish_use_subcommand" -a version -d "Show version information"

# Config subcommands
complete -c gpd -n "__fish_seen_subcommand_from config" -a init -d "Initialize project configuration"
complete -c gpd -n "__fish_seen_subcommand_from config" -a doctor -d "Diagnose configuration issues"
complete -c gpd -n "__fish_seen_subcommand_from config" -a path -d "Show configuration paths"
complete -c gpd -n "__fish_seen_subcommand_from config" -a get -d "Get configuration value"
complete -c gpd -n "__fish_seen_subcommand_from config" -a set -d "Set configuration value"
complete -c gpd -n "__fish_seen_subcommand_from config" -a print -d "Print configuration"
complete -c gpd -n "__fish_seen_subcommand_from config" -a export -d "Export configuration"
complete -c gpd -n "__fish_seen_subcommand_from config" -a import -d "Import configuration"
complete -c gpd -n "__fish_seen_subcommand_from config" -a completion -d "Generate shell completion"`)
	default:
		return errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("unsupported shell: %s", cmd.Shell)).
			WithHint("Supported shells: bash, zsh, fish")
	}
	return nil
}

// ConfigExportCmd exports configuration to file.
type ConfigExportCmd struct {
	OutFile      string `help:"Output file path" default:"gpd-config.json" short:"o"`
	IncludePaths bool   `help:"Include serviceAccountKeyPath (warning: may be machine-specific)"`
}

func (cmd *ConfigExportCmd) Run(globals *Globals) error {
	cfg, err := config.Load()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to load config: %v", err))
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
		if cmd.IncludePaths {
			exportData["serviceAccountKeyPath"] = cfg.ServiceAccountKeyPath
			warnings = append(warnings, "serviceAccountKeyPath included - may be machine-specific")
		} else {
			warnings = append(warnings, "serviceAccountKeyPath not included - use --include-paths to export")
		}
	}

	export := configExport{
		Version:    "1.0",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Config:     exportData,
		Metadata: map[string]interface{}{
			"platform":         runtime.GOOS,
			"credentialOrigin": cfg.CredentialOrigin,
			"warnings":         warnings,
		},
	}

	data, err := marshalConfigExport(cmd.OutFile, export)
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to marshal config: %v", err)).
			WithHint("Use .json, .yaml, or .yml extension for config export")
	}

	if err := os.WriteFile(cmd.OutFile, data, 0600); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to write file: %v", err))
	}

	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"exported":   len(exportData),
		"outputPath": cmd.OutFile,
		"warnings":   warnings,
	})
	return writeOutput(globals, result)
}

type configExport struct {
	Version    string                 `json:"version"`
	ExportedAt string                 `json:"exportedAt"`
	Config     map[string]interface{} `json:"config"`
	Metadata   map[string]interface{} `json:"metadata"`
}

func marshalConfigExport(outputPath string, export configExport) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(outputPath))
	if ext == ".yaml" || ext == ".yml" {
		return yaml.Marshal(export)
	}
	if ext == "" || ext == ".json" {
		return json.MarshalIndent(export, "", "  ")
	}
	return nil, fmt.Errorf("unsupported config export extension: %s", ext)
}

// ConfigImportCmd imports configuration from file.
type ConfigImportCmd struct {
	File  string `arg:"" help:"Configuration file to import"`
	Merge bool   `help:"Merge with existing config" default:"true"`
}

func (cmd *ConfigImportCmd) Run(globals *Globals) error {
	data, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("failed to read file: %v", err))
	}

	var importData configExport
	if err := unmarshalConfigExport(cmd.File, data, &importData); err != nil {
		return errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid config file: %v", err)).
			WithHint("Expected JSON/YAML format from 'gpd config export'")
	}

	cfg := loadOrCreateConfig(cmd.Merge)
	imported := applyImportedConfig(cfg, importData.Config)

	if err := cfg.Save(); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to save config: %v", err))
	}

	result := output.NewResult(map[string]interface{}{
		"success":  true,
		"imported": imported,
		"merge":    cmd.Merge,
		"version":  importData.Version,
	})
	return writeOutput(globals, result)
}

func loadOrCreateConfig(merge bool) *config.Config {
	if merge {
		cfg, err := config.Load()
		if err == nil {
			return cfg
		}
	}
	return &config.Config{}
}

func applyImportedConfig(cfg *config.Config, data map[string]interface{}) []string {
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
		cfg.RateLimits = parseRateLimits(val)
		imported = append(imported, "rateLimits")
	}
	if val, ok := data["testerLimits"].(map[string]interface{}); ok && len(val) > 0 {
		cfg.TesterLimits = parseTesterLimits(val)
		imported = append(imported, "testerLimits")
	}
	if val, ok := data["serviceAccountKeyPath"].(string); ok && val != "" {
		cfg.ServiceAccountKeyPath = val
		imported = append(imported, "serviceAccountKeyPath")
	}

	return imported
}

func parseRateLimits(val map[string]interface{}) map[string]string {
	rateLimits := make(map[string]string)
	for k, v := range val {
		if strVal, ok := v.(string); ok {
			rateLimits[k] = strVal
		}
	}
	return rateLimits
}

func parseTesterLimits(val map[string]interface{}) *config.TesterLimits {
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

func unmarshalConfigExport(inputPath string, data []byte, out *configExport) error {
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

// writeOutput is a helper to write output using the output package
func writeOutput(globals *Globals, result *output.Result) error {
	out := output.NewManager(os.Stdout)
	out.SetFormat(output.ParseFormat(globals.Output))
	out.SetPretty(globals.Pretty)
	if globals.Fields != "" {
		out.SetFields(strings.Split(globals.Fields, ","))
	}
	return out.Write(result)
}
