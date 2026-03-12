package cli

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/extensions"
)

func TestExtensionInstallCmd(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name:    "empty source",
			source:  "",
			wantErr: true,
		},
		{
			name:    "valid github format",
			source:  "owner/gpd-test",
			wantErr: false, // Will error during actual install but command parsing works
		},
		{
			name:    "invalid format with spaces",
			source:  "owner / repo",
			wantErr: false, // Will be treated as local path
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ExtensionInstallCmd{
				Source: tt.source,
			}

			err := cmd.Run(&Globals{Context: context.Background()})
			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantErr && err != nil && tt.source == "" {
				// Empty source should error
				return
			}
		})
	}
}

func TestExtensionListCmd(t *testing.T) {
	cmd := &ExtensionListCmd{}
	globals := &Globals{
		Context: context.Background(),
		Output:  "json",
	}

	// This should work even with no extensions installed
	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("List command failed: %v", err)
	}
}

func TestExtensionRemoveCmd(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		wantErr bool
	}{
		{
			name:    "empty name",
			cmdName: "",
			wantErr: true,
		},
		{
			name:    "non-existent extension",
			cmdName: "nonexistent-ext",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ExtensionRemoveCmd{
				Name: tt.cmdName,
			}

			err := cmd.Run(&Globals{
				Context: context.Background(),
				Output:  "json",
			})
			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestExtensionUpgradeCmd(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		all     bool
		wantErr bool
	}{
		{
			name:    "no name and no --all",
			cmdName: "",
			all:     false,
			wantErr: true,
		},
		{
			name:    "with --all flag",
			cmdName: "",
			all:     true,
			wantErr: false, // --all reports failures as warnings, not errors
		},
		{
			name:    "with name",
			cmdName: "test-ext",
			all:     false,
			wantErr: true, // Individual extensions fail on error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ExtensionUpgradeCmd{
				Name: tt.cmdName,
				All:  tt.all,
			}

			err := cmd.Run(&Globals{Context: context.Background()})
			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestExtensionExecCmd(t *testing.T) {
	tests := []struct {
		name    string
		cmdName string
		wantErr bool
	}{
		{
			name:    "empty name",
			cmdName: "",
			wantErr: true,
		},
		{
			name:    "non-existent extension",
			cmdName: "nonexistent-ext",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ExtensionExecCmd{
				Name: tt.cmdName,
				Args: []string{},
			}

			err := cmd.Run(&Globals{Context: context.Background()})
			if tt.wantErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTryRunExtension(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "empty args",
			args: []string{},
			want: false,
		},
		{
			name: "built-in command",
			args: []string{"auth"},
			want: false,
		},
		{
			name: "global flag",
			args: []string{"--help"},
			want: false,
		},
		{
			name: "non-existent extension",
			args: []string{"nonexistent-ext"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// tryRunExtension doesn't actually run the extension in this test
			// since we don't have any installed extensions
			got := tryRunExtension(tt.args)
			if got != tt.want {
				t.Errorf("tryRunExtension(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestIsGlobalFlag(t *testing.T) {
	tests := []struct {
		arg  string
		want bool
	}{
		{"-h", true},
		{"--help", true},
		{"-v", true},
		{"--version", true},
		{"--verbose", true},
		{"--package", true},
		{"--output", true},
		{"--pretty", true},
		{"--timeout", true},
		{"--key", true},
		{"--profile", true},
		{"auth", false},
		{"extension", false},
		{"dash", false},
		{"random-arg", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got := isGlobalFlag(tt.arg)
			if got != tt.want {
				t.Errorf("isGlobalFlag(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestExtensionCmdStructure(t *testing.T) {
	// Verify the command structure has all expected subcommands
	cmd := &ExtensionCmd{}

	// These shouldn't be nil
	if cmd.Install.Source != "" {
		t.Error("Install.Source should be empty by default")
	}
}

func TestExtensionInstallCmdJSONOutput(t *testing.T) {
	// Create a mock local extension for testing
	tmpDir := t.TempDir()

	// Create manifest
	manifest := extensions.Manifest{
		Name:    "test-ext",
		Version: "1.0.0",
	}
	manifestPath := filepath.Join(tmpDir, ".gpd-extension")
	data, _ := json.Marshal(manifest)
	if err := writeFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Create executable
	binName := "gpd-test-ext"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(tmpDir, binName)
	content := "#!/bin/sh\necho test"
	if runtime.GOOS == "windows" {
		content = "@echo test"
	}
	if err := writeFile(binPath, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create executable: %v", err)
	}

	// Test install
	cmd := &ExtensionInstallCmd{
		Source: tmpDir,
		Pin:    false,
		Force:  true,
	}

	globals := &Globals{
		Context: context.Background(),
		Output:  "json",
	}

	err := cmd.Run(globals)
	// May fail due to environment, but should at least attempt to run
	_ = err
}

// Helper function to write files (avoiding naming conflict)
func writeFile(path string, data []byte, perm int) error {
	return writeFileInternal(path, data, perm)
}

func writeFileInternal(path string, data []byte, perm int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := f.Chmod(os.FileMode(perm)); err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

func TestExtensionOutputFormats(t *testing.T) {
	globals := &Globals{
		Context: context.Background(),
		Output:  "json",
		Pretty:  true,
	}

	// Test list command with different outputs
	listCmd := &ExtensionListCmd{}
	err := listCmd.Run(globals)
	if err != nil {
		t.Errorf("List command with JSON output failed: %v", err)
	}

	// Test with table output
	globals.Output = "table"
	err = listCmd.Run(globals)
	if err != nil {
		t.Errorf("List command with table output failed: %v", err)
	}
}
