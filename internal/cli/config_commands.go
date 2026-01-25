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

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
	"github.com/dl-alexandre/gpd/internal/storage"
)

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
		Use:   "print",
		Short: "Print resolved configuration",
		Long:  "Print the fully resolved configuration showing precedence.",
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
		Long:  "Export safe configuration values to a JSON file.",
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
		Long:  "Import configuration values from a JSON file.",
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
	if _, err := os.Stat(paths.ConfigFile); os.IsNotExist(err) {
		issues = append(issues, "Config file does not exist (run 'gpd config init')")
		checks["configFile"] = map[string]interface{}{"exists": false, "path": paths.ConfigFile}
	} else {
		checks["configFile"] = map[string]interface{}{"exists": true, "path": paths.ConfigFile}
	}

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

	// Check environment variables
	envChecks := map[string]interface{}{}
	if key := config.GetEnvServiceAccountKey(); key != "" {
		envChecks["GPD_SERVICE_ACCOUNT_KEY"] = "set (value hidden)"
	} else {
		envChecks["GPD_SERVICE_ACCOUNT_KEY"] = "not set"
	}
	if pkg := config.GetEnvPackage(); pkg != "" {
		envChecks["GPD_PACKAGE"] = pkg
	} else {
		envChecks["GPD_PACKAGE"] = "not set"
	}
	envChecks["GOOGLE_APPLICATION_CREDENTIALS"] = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	checks["environment"] = envChecks

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

func (c *CLI) configPath(_ *cobra.Command) error {
	paths := config.GetPaths()
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

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to marshal config: %v", err)))
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
	if err := json.Unmarshal(data, &importData); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid config file: %v", err)).
			WithHint("Expected JSON format from 'gpd config export'"))
	}

	var cfg *config.Config
	if merge {
		cfg, err = config.Load()
		if err != nil {
			cfg = &config.Config{}
		}
	} else {
		cfg = &config.Config{}
	}

	imported := []string{}

	if val, ok := importData.Config["defaultPackage"].(string); ok && val != "" {
		cfg.DefaultPackage = val
		imported = append(imported, "defaultPackage")
	}
	if val, ok := importData.Config["outputFormat"].(string); ok && val != "" {
		cfg.OutputFormat = val
		imported = append(imported, "outputFormat")
	}
	if val, ok := importData.Config["timeoutSeconds"].(float64); ok && val > 0 {
		cfg.TimeoutSeconds = int(val)
		imported = append(imported, "timeoutSeconds")
	}
	if val, ok := importData.Config["storeTokens"].(string); ok && val != "" {
		cfg.StoreTokens = val
		imported = append(imported, "storeTokens")
	}
	if val, ok := importData.Config["rateLimits"].(map[string]interface{}); ok && len(val) > 0 {
		rateLimits := make(map[string]string)
		for k, v := range val {
			if strVal, ok := v.(string); ok {
				rateLimits[k] = strVal
			}
		}
		cfg.RateLimits = rateLimits
		imported = append(imported, "rateLimits")
	}
	if val, ok := importData.Config["testerLimits"].(map[string]interface{}); ok && len(val) > 0 {
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
		cfg.TesterLimits = limits
		imported = append(imported, "testerLimits")
	}
	if val, ok := importData.Config["serviceAccountKeyPath"].(string); ok && val != "" {
		cfg.ServiceAccountKeyPath = val
		imported = append(imported, "serviceAccountKeyPath")
	}

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
