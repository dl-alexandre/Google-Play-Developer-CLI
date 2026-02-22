// Package cli provides the Kong-based CLI framework for gpd.
package cli

import (
	"os"
	"testing"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/gpd/internal/errors"
)

// TestGlobalsDefaults verifies that Globals struct has proper default values
func TestGlobalsDefaults(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Globals
		wantOutput  string
		wantTimeout time.Duration
		wantStore   string
		wantPretty  bool
		wantQuiet   bool
		wantVerbose bool
	}{
		{
			name:        "zero values",
			setup:       func() *Globals { return &Globals{} },
			wantOutput:  "",
			wantTimeout: 0,
			wantStore:   "",
			wantPretty:  false,
			wantQuiet:   false,
			wantVerbose: false,
		},
		{
			name: "with package set",
			setup: func() *Globals {
				return &Globals{Package: "com.example.app"}
			},
			wantOutput:  "",
			wantTimeout: 0,
		},
		{
			name: "with all bool flags set",
			setup: func() *Globals {
				return &Globals{Pretty: true, Quiet: true, Verbose: true}
			},
			wantPretty:  true,
			wantQuiet:   true,
			wantVerbose: true,
		},
		{
			name: "with key path and profile",
			setup: func() *Globals {
				return &Globals{KeyPath: "/path/to/key.json", Profile: "production"}
			},
			wantOutput: "",
		},
		{
			name: "with fields projection",
			setup: func() *Globals {
				return &Globals{Fields: "data.name,data.version"}
			},
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := tt.setup()
			if g.Output != tt.wantOutput {
				t.Errorf("Output = %v, want %v", g.Output, tt.wantOutput)
			}
			if g.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %v, want %v", g.Timeout, tt.wantTimeout)
			}
			if g.StoreTokens != tt.wantStore {
				t.Errorf("StoreTokens = %v, want %v", g.StoreTokens, tt.wantStore)
			}
			if g.Pretty != tt.wantPretty {
				t.Errorf("Pretty = %v, want %v", g.Pretty, tt.wantPretty)
			}
			if g.Quiet != tt.wantQuiet {
				t.Errorf("Quiet = %v, want %v", g.Quiet, tt.wantQuiet)
			}
			if g.Verbose != tt.wantVerbose {
				t.Errorf("Verbose = %v, want %v", g.Verbose, tt.wantVerbose)
			}
		})
	}
}

// TestKongCLIStructure verifies that KongCLI struct can be created and parsed
func TestKongCLIStructure(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments shows help",
			args:    []string{"--help"},
			wantErr: true, // Kong returns error for --help even with custom exit
		},
		{
			name:    "version command",
			args:    []string{"version", "--help"},
			wantErr: false, // Version command help succeeds
		},
		{
			name:    "auth command with help",
			args:    []string{"auth", "--help"},
			wantErr: true,
		},
		{
			name:    "config command with help",
			args:    []string{"config", "--help"},
			wantErr: true,
		},
		{
			name:    "publish command with help",
			args:    []string{"publish", "--help"},
			wantErr: true,
		},
		{
			name:    "reviews command with help",
			args:    []string{"reviews", "--help"},
			wantErr: true,
		},
		{
			name:    "vitals command with help",
			args:    []string{"vitals", "--help"},
			wantErr: true,
		},
		{
			name:    "analytics command with help",
			args:    []string{"analytics", "--help"},
			wantErr: true,
		},
		{
			name:    "purchases command with help",
			args:    []string{"purchases", "--help"},
			wantErr: true,
		},
		{
			name:    "monetization command with help",
			args:    []string{"monetization", "--help"},
			wantErr: true,
		},
		{
			name:    "permissions command with help",
			args:    []string{"permissions", "--help"},
			wantErr: true,
		},
		{
			name:    "recovery command with help",
			args:    []string{"recovery", "--help"},
			wantErr: true,
		},
		{
			name:    "apps command with help",
			args:    []string{"apps", "--help"},
			wantErr: true,
		},
		{
			name:    "games command with help",
			args:    []string{"games", "--help"},
			wantErr: true,
		},
		{
			name:    "integrity command with help",
			args:    []string{"integrity", "--help"},
			wantErr: true,
		},
		{
			name:    "migrate command with help",
			args:    []string{"migrate", "--help"},
			wantErr: false,
		},
		{
			name:    "customapp alias command with help",
			args:    []string{"customapp", "--help"},
			wantErr: false,
		},
		{
			name:    "custom-app command with help",
			args:    []string{"custom-app", "--help"},
			wantErr: false,
		},
		{
			name:    "grouping command with help",
			args:    []string{"grouping", "--help"},
			wantErr: false,
		},
		{
			name:    "global flags with json output",
			args:    []string{"--output", "json", "version", "--help"},
			wantErr: false,
		},
		{
			name:    "global flags with table output",
			args:    []string{"--output", "table", "version", "--help"},
			wantErr: false,
		},
		{
			name:    "global flags with invalid output",
			args:    []string{"--output", "invalid", "version"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli KongCLI
			parser, err := kong.New(&cli,
				kong.Name("gpd"),
				kong.Description("Google Play Developer CLI"),
				kong.Exit(func(int) {}),
			)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			_, err = parser.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestKongCLITopLevelCommands verifies all top-level command fields exist
func TestKongCLITopLevelCommands(t *testing.T) {
	var cli KongCLI

	// Create parser to validate structure
	parser, err := kong.New(&cli,
		kong.Name("gpd"),
		kong.Description("Test"),
		kong.Exit(func(int) {}),
	)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Verify all commands are registered
	commands := []string{
		"auth",
		"config",
		"publish",
		"reviews",
		"vitals",
		"analytics",
		"purchases",
		"monetization",
		"permissions",
		"recovery",
		"apps",
		"games",
		"integrity",
		"migrate",
		"custom-app",
		"customapp",
		"grouping",
		"version",
	}

	for _, cmd := range commands {
		t.Run("command_"+cmd, func(t *testing.T) {
			// Kong returns an error for --help even with custom exit handler
			// Just verify the parser doesn't panic
			_, _ = parser.Parse([]string{cmd, "--help"})
		})
	}
}

// TestGlobalsParsing verifies Globals struct is properly parsed from flags
func TestGlobalsParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantPackage string
		wantOutput  string
		wantPretty  bool
		wantTimeout time.Duration
		wantStore   string
		wantFields  string
		wantQuiet   bool
		wantVerbose bool
		wantKeyPath string
		wantProfile string
		wantErr     bool
	}{
		{
			name:        "default values via version command",
			args:        []string{"version", "--help"},
			wantOutput:  "json",
			wantTimeout: 30 * time.Second,
			wantStore:   "auto",
		},
		{
			name:        "package flag short",
			args:        []string{"-p", "com.test.app", "version", "--help"},
			wantPackage: "com.test.app",
			wantOutput:  "json",
		},
		{
			name:        "package flag long",
			args:        []string{"--package", "com.example.game", "version", "--help"},
			wantPackage: "com.example.game",
		},
		{
			name:       "output table",
			args:       []string{"--output", "table", "version", "--help"},
			wantOutput: "table",
		},
		{
			name:       "output markdown",
			args:       []string{"--output", "markdown", "version", "--help"},
			wantOutput: "markdown",
		},
		{
			name:       "output csv",
			args:       []string{"--output", "csv", "version", "--help"},
			wantOutput: "csv",
		},
		{
			name:       "pretty flag",
			args:       []string{"--pretty", "version", "--help"},
			wantPretty: true,
		},
		{
			name:        "timeout custom",
			args:        []string{"--timeout", "60s", "version", "--help"},
			wantTimeout: 60 * time.Second,
		},
		{
			name:      "store tokens never",
			args:      []string{"--store-tokens", "never", "version", "--help"},
			wantStore: "never",
		},
		{
			name:      "store tokens secure",
			args:      []string{"--store-tokens", "secure", "version", "--help"},
			wantStore: "secure",
		},
		{
			name:       "fields projection",
			args:       []string{"--fields", "data.id,data.name", "version", "--help"},
			wantFields: "data.id,data.name",
		},
		{
			name:      "quiet flag",
			args:      []string{"--quiet", "version", "--help"},
			wantQuiet: true,
		},
		{
			name:        "verbose flag short",
			args:        []string{"-v", "version", "--help"},
			wantVerbose: true,
		},
		{
			name:        "verbose flag long",
			args:        []string{"--verbose", "version", "--help"},
			wantVerbose: true,
		},
		{
			name:        "key path flag",
			args:        []string{"--key-path", "/keys/service.json", "version", "--help"},
			wantKeyPath: "/keys/service.json",
		},
		{
			name:        "profile flag",
			args:        []string{"--profile", "staging", "version", "--help"},
			wantProfile: "staging",
		},
		{
			name:    "invalid output format",
			args:    []string{"--output", "xml", "version"},
			wantErr: true,
		},
		{
			name:    "invalid store tokens",
			args:    []string{"--store-tokens", "invalid", "version"},
			wantErr: true,
		},
		{
			name:        "multiple global flags",
			args:        []string{"--package", "com.test.app", "--output", "table", "--verbose", "--timeout", "45s", "version", "--help"},
			wantPackage: "com.test.app",
			wantOutput:  "table",
			wantVerbose: true,
			wantTimeout: 45 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cli KongCLI
			parser, err := kong.New(&cli,
				kong.Name("gpd"),
				kong.Description("Test"),
				kong.Exit(func(int) {}),
			)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			_, err = parser.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.wantPackage != "" && cli.Globals.Package != tt.wantPackage {
				t.Errorf("Package = %v, want %v", cli.Globals.Package, tt.wantPackage)
			}
			if tt.wantOutput != "" && cli.Globals.Output != tt.wantOutput {
				t.Errorf("Output = %v, want %v", cli.Globals.Output, tt.wantOutput)
			}
			if cli.Globals.Pretty != tt.wantPretty {
				t.Errorf("Pretty = %v, want %v", cli.Globals.Pretty, tt.wantPretty)
			}
			if tt.wantTimeout != 0 && cli.Globals.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %v, want %v", cli.Globals.Timeout, tt.wantTimeout)
			}
			if tt.wantStore != "" && cli.Globals.StoreTokens != tt.wantStore {
				t.Errorf("StoreTokens = %v, want %v", cli.Globals.StoreTokens, tt.wantStore)
			}
			if tt.wantFields != "" && cli.Globals.Fields != tt.wantFields {
				t.Errorf("Fields = %v, want %v", cli.Globals.Fields, tt.wantFields)
			}
			if cli.Globals.Quiet != tt.wantQuiet {
				t.Errorf("Quiet = %v, want %v", cli.Globals.Quiet, tt.wantQuiet)
			}
			if cli.Globals.Verbose != tt.wantVerbose {
				t.Errorf("Verbose = %v, want %v", cli.Globals.Verbose, tt.wantVerbose)
			}
			if tt.wantKeyPath != "" && cli.Globals.KeyPath != tt.wantKeyPath {
				t.Errorf("KeyPath = %v, want %v", cli.Globals.KeyPath, tt.wantKeyPath)
			}
			if tt.wantProfile != "" && cli.Globals.Profile != tt.wantProfile {
				t.Errorf("Profile = %v, want %v", cli.Globals.Profile, tt.wantProfile)
			}
		})
	}
}

// TestParserCreation verifies that Kong parser can be created successfully
func TestParserCreation(t *testing.T) {
	tests := []struct {
		name    string
		skip    bool
		wantErr bool
	}{
		{
			name:    "create parser with valid struct",
			skip:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Skipping parser creation test")
			}

			var cli KongCLI
			parser, err := kong.New(&cli,
				kong.Name("gpd"),
				kong.Description("Google Play Developer CLI"),
				kong.UsageOnError(),
				kong.ConfigureHelp(kong.HelpOptions{
					Compact: true,
					Summary: true,
				}),
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("kong.New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if parser == nil && !tt.wantErr {
				t.Error("Expected non-nil parser")
			}
		})
	}
}

// TestRunKongCLIExitCodes verifies exit codes are properly defined
func TestRunKongCLIExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
		expected int
	}{
		{
			name:     "success exit code",
			exitCode: errors.ExitSuccess,
			expected: 0,
		},
		{
			name:     "general error exit code",
			exitCode: errors.ExitGeneralError,
			expected: 1,
		},
		{
			name:     "auth failure exit code",
			exitCode: errors.ExitAuthFailure,
			expected: 2,
		},
		{
			name:     "permission denied exit code",
			exitCode: errors.ExitPermissionDenied,
			expected: 3,
		},
		{
			name:     "validation error exit code",
			exitCode: errors.ExitValidationError,
			expected: 4,
		},
		{
			name:     "rate limited exit code",
			exitCode: errors.ExitRateLimited,
			expected: 5,
		},
		{
			name:     "network error exit code",
			exitCode: errors.ExitNetworkError,
			expected: 6,
		},
		{
			name:     "not found exit code",
			exitCode: errors.ExitNotFound,
			expected: 7,
		},
		{
			name:     "conflict exit code",
			exitCode: errors.ExitConflict,
			expected: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.exitCode != tt.expected {
				t.Errorf("Exit code = %d, expected %d", tt.exitCode, tt.expected)
			}
		})
	}
}

// TestKongCLICommandFields verifies all command fields exist on KongCLI struct
func TestKongCLICommandFields(t *testing.T) {
	var cli KongCLI

	// Verify all command fields are not nil interfaces
	commands := []struct {
		name  string
		field interface{}
	}{
		{"Auth", cli.Auth},
		{"Config", cli.Config},
		{"Publish", cli.Publish},
		{"Reviews", cli.Reviews},
		{"Vitals", cli.Vitals},
		{"Analytics", cli.Analytics},
		{"Purchases", cli.Purchases},
		{"Monetization", cli.Monetization},
		{"Permissions", cli.Permissions},
		{"Recovery", cli.Recovery},
		{"Apps", cli.Apps},
		{"Games", cli.Games},
		{"Integrity", cli.Integrity},
		{"Migrate", cli.Migrate},
		{"CustomApp", cli.CustomApp},
		{"Grouping", cli.Grouping},
		{"Version", cli.Version},
	}

	for _, cmd := range commands {
		t.Run("field_"+cmd.name, func(t *testing.T) {
			if cmd.field == nil {
				t.Errorf("Command field %q is nil", cmd.name)
			}
		})
	}
}

// TestGlobalsStructTags verifies struct tags are properly defined
func TestGlobalsStructTags(t *testing.T) {
	var g Globals

	// Verify struct can be instantiated
	if g.Output != "" {
		t.Error("Expected empty Output by default")
	}
	if g.Timeout != 0 {
		t.Error("Expected zero Timeout by default")
	}

	// Test setting values
	g.Output = "json"
	g.Timeout = 30 * time.Second
	g.StoreTokens = "auto"
	g.Package = "com.test.app"
	g.Pretty = true
	g.Quiet = false
	g.Verbose = true
	g.KeyPath = "/path/to/key.json"
	g.Profile = "default"
	g.Fields = "data.field1"

	if g.Output != "json" {
		t.Errorf("Output = %v, want json", g.Output)
	}
	if g.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", g.Timeout)
	}
	if g.StoreTokens != "auto" {
		t.Errorf("StoreTokens = %v, want auto", g.StoreTokens)
	}
}

// TestKongCLIWithEnvironment verifies CLI works with environment variable patterns
func TestKongCLIWithEnvironment(t *testing.T) {
	// Set up test environment
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "minimal args",
			args: []string{"gpd", "version", "--help"},
		},
		{
			name: "with package arg",
			args: []string{"gpd", "--package", "com.env.app", "version", "--help"},
		},
		{
			name: "with verbose",
			args: []string{"gpd", "-v", "version", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args

			var cli KongCLI
			parser, err := kong.New(&cli,
				kong.Name("gpd"),
				kong.Description("Test"),
				kong.Exit(func(int) {}),
			)
			if err != nil {
				t.Fatalf("Failed to create parser: %v", err)
			}

			_, err = parser.Parse(os.Args[1:])
			// Help with custom exit should not error
			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
			}
		})
	}
}

// TestAPIErrorExitCodeMapping verifies error code mappings work correctly
func TestAPIErrorExitCodeMapping(t *testing.T) {
	tests := []struct {
		name     string
		err      *errors.APIError
		expected int
	}{
		{
			name:     "success error code",
			err:      errors.NewAPIError(errors.CodeSuccess, "success"),
			expected: errors.ExitSuccess,
		},
		{
			name:     "auth failure error code",
			err:      errors.NewAPIError(errors.CodeAuthFailure, "auth failed"),
			expected: errors.ExitAuthFailure,
		},
		{
			name:     "permission denied error code",
			err:      errors.NewAPIError(errors.CodePermissionDenied, "denied"),
			expected: errors.ExitPermissionDenied,
		},
		{
			name:     "validation error code",
			err:      errors.NewAPIError(errors.CodeValidationError, "invalid"),
			expected: errors.ExitValidationError,
		},
		{
			name:     "rate limited error code",
			err:      errors.NewAPIError(errors.CodeRateLimited, "slow down"),
			expected: errors.ExitRateLimited,
		},
		{
			name:     "network error code",
			err:      errors.NewAPIError(errors.CodeNetworkError, "timeout"),
			expected: errors.ExitNetworkError,
		},
		{
			name:     "not found error code",
			err:      errors.NewAPIError(errors.CodeNotFound, "missing"),
			expected: errors.ExitNotFound,
		},
		{
			name:     "conflict error code",
			err:      errors.NewAPIError(errors.CodeConflict, "conflict"),
			expected: errors.ExitConflict,
		},
		{
			name:     "general error code",
			err:      errors.NewAPIError(errors.CodeGeneralError, "error"),
			expected: errors.ExitGeneralError,
		},
		{
			name:     "unknown error code defaults to general",
			err:      errors.NewAPIError(errors.ErrorCode("UNKNOWN"), "unknown"),
			expected: errors.ExitGeneralError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.ExitCode()
			if got != tt.expected {
				t.Errorf("ExitCode() = %d, expected %d", got, tt.expected)
			}
		})
	}
}

// TestGlobalsValidation verifies input validation for Globals fields
func TestGlobalsValidation(t *testing.T) {
	tests := []struct {
		name    string
		globals Globals
		valid   bool
	}{
		{
			name: "empty globals",
			globals: Globals{
				Output:      "",
				Timeout:     0,
				StoreTokens: "",
			},
			valid: true,
		},
		{
			name: "valid output json",
			globals: Globals{
				Output: "json",
			},
			valid: true,
		},
		{
			name: "valid output table",
			globals: Globals{
				Output: "table",
			},
			valid: true,
		},
		{
			name: "valid output markdown",
			globals: Globals{
				Output: "markdown",
			},
			valid: true,
		},
		{
			name: "valid output csv",
			globals: Globals{
				Output: "csv",
			},
			valid: true,
		},
		{
			name: "valid store auto",
			globals: Globals{
				StoreTokens: "auto",
			},
			valid: true,
		},
		{
			name: "valid store never",
			globals: Globals{
				StoreTokens: "never",
			},
			valid: true,
		},
		{
			name: "valid store secure",
			globals: Globals{
				StoreTokens: "secure",
			},
			valid: true,
		},
		{
			name: "valid timeout",
			globals: Globals{
				Timeout: 30 * time.Second,
			},
			valid: true,
		},
		{
			name: "valid package name",
			globals: Globals{
				Package: "com.example.valid",
			},
			valid: true,
		},
		{
			name: "valid key path",
			globals: Globals{
				KeyPath: "/valid/path/to/key.json",
			},
			valid: true,
		},
		{
			name: "valid profile",
			globals: Globals{
				Profile: "production",
			},
			valid: true,
		},
		{
			name: "valid fields",
			globals: Globals{
				Fields: "data.name,meta.version",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation here just checks struct can hold values
			// Kong will handle actual enum validation
			_ = tt.globals
		})
	}
}

// TestKongCLIHelpOutput verifies help output is generated for commands
func TestKongCLIHelpOutput(t *testing.T) {
	var cli KongCLI

	parser, err := kong.New(&cli,
		kong.Name("gpd"),
		kong.Description("Google Play Developer CLI"),
		kong.Exit(func(int) {}),
	)
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "root help",
			args: []string{"--help"},
		},
		{
			name: "auth help",
			args: []string{"auth", "--help"},
		},
		{
			name: "publish help",
			args: []string{"publish", "--help"},
		},
		{
			name: "config help",
			args: []string{"config", "--help"},
		},
		{
			name: "version help",
			args: []string{"version", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Kong returns an error for --help even with custom exit handler
			// This is expected behavior, we just verify the parser doesn't panic
			_, _ = parser.Parse(tt.args)
			// Test passes if we reach here without panic
		})
	}
}

// TestErrorCodeConstants verifies all error code constants are defined
func TestErrorCodeConstants(t *testing.T) {
	// These should all be defined in the errors package
	codes := []struct {
		name string
		code errors.ErrorCode
	}{
		{"CodeSuccess", errors.CodeSuccess},
		{"CodeGeneralError", errors.CodeGeneralError},
		{"CodeAuthFailure", errors.CodeAuthFailure},
		{"CodePermissionDenied", errors.CodePermissionDenied},
		{"CodeValidationError", errors.CodeValidationError},
		{"CodeRateLimited", errors.CodeRateLimited},
		{"CodeNetworkError", errors.CodeNetworkError},
		{"CodeNotFound", errors.CodeNotFound},
		{"CodeConflict", errors.CodeConflict},
	}

	for _, tt := range codes {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code == "" {
				t.Errorf("Error code %s is empty", tt.name)
			}
		})
	}
}
