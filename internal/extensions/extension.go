// Package extensions provides the gpd extension system for installable subcommands.
package extensions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// Extension represents an installed gpd extension.
type Extension struct {
	Name        string    `json:"name" yaml:"name"`
	Version     string    `json:"version" yaml:"version"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	Author      string    `json:"author,omitempty" yaml:"author,omitempty"`
	Bin         string    `json:"bin,omitempty" yaml:"bin,omitempty"`
	Source      string    `json:"source" yaml:"source"` // GitHub repo path (owner/repo)
	InstalledAt time.Time `json:"installedAt" yaml:"installedAt"`
	UpdatedAt   time.Time `json:"updatedAt" yaml:"updatedAt"`
	Pinned      bool      `json:"pinned" yaml:"pinned"`                           // Whether auto-upgrade is disabled
	PinnedRef   string    `json:"pinnedRef,omitempty" yaml:"pinnedRef,omitempty"` // Tag or commit if pinned
	Type        string    `json:"type" yaml:"type"`                               // "binary" or "script"
}

// Manifest represents the .gpd-extension file in an extension repository.
type Manifest struct {
	Name        string `json:"name" yaml:"name"`
	Version     string `json:"version" yaml:"version"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Author      string `json:"author,omitempty" yaml:"author,omitempty"`
	Bin         string `json:"bin,omitempty" yaml:"bin,omitempty"` // Executable name (defaults to gpd-<name>)
	Homepage    string `json:"homepage,omitempty" yaml:"homepage,omitempty"`
}

// Validate checks if the manifest is valid.
func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("extension name is required")
	}

	// Name validation: alphanumeric + hyphens, no spaces
	for _, r := range m.Name {
		if !isValidExtensionNameChar(r) {
			return fmt.Errorf("extension name %q contains invalid character %q", m.Name, string(r))
		}
	}

	if m.Version == "" {
		return fmt.Errorf("extension version is required")
	}

	return nil
}

func isValidExtensionNameChar(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '-' || r == '_'
}

// DefaultBinName returns the default executable name for an extension.
func (m *Manifest) DefaultBinName() string {
	if m.Bin != "" {
		return m.Bin
	}
	return fmt.Sprintf("gpd-%s", m.Name)
}

// GetExtensionsDir returns the directory where extensions are installed.
func GetExtensionsDir() string {
	paths := getExtensionPathsForOS(runtime.GOOS)
	return paths.ExtensionsDir
}

// extensionPaths holds paths for the extension system.
type extensionPaths struct {
	ExtensionsDir string
}

func getExtensionPathsForOS(goos string) extensionPaths {
	var extensionsDir string

	switch goos {
	case "darwin":
		home := getHomeDir()
		extensionsDir = filepath.Join(home, "Library", "Application Support", "gpd", "extensions")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home := getHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		extensionsDir = filepath.Join(appData, "gpd", "extensions")
	default: // linux and others
		home := getHomeDir()
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			extensionsDir = filepath.Join(xdgData, "gpd", "extensions")
		} else {
			extensionsDir = filepath.Join(home, ".local", "share", "gpd", "extensions")
		}
	}

	return extensionPaths{
		ExtensionsDir: extensionsDir,
	}
}

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}

// List returns all installed extensions.
func List() ([]Extension, error) {
	dir := GetExtensionsDir()

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating extensions directory: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading extensions directory: %w", err)
	}

	var extensions []Extension
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		ext, err := LoadExtension(entry.Name())
		if err != nil {
			// Log warning but don't fail entirely
			continue
		}

		extensions = append(extensions, *ext)
	}

	return extensions, nil
}

// LoadExtension loads an extension by name from the extensions directory.
func LoadExtension(name string) (*Extension, error) {
	extDir := filepath.Join(GetExtensionsDir(), name)

	// Check if extension directory exists
	if _, err := os.Stat(extDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("extension %q not found", name)
	}

	// Load metadata file
	metaPath := filepath.Join(extDir, ".gpd-extension")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("reading extension metadata: %w", err)
	}

	var ext Extension
	if err := unmarshalMetadata(data, &ext); err != nil {
		return nil, fmt.Errorf("parsing extension metadata: %w", err)
	}

	return &ext, nil
}

// unmarshalMetadata unmarshals extension metadata (JSON or YAML).
func unmarshalMetadata(data []byte, ext *Extension) error {
	// Try JSON first
	if err := json.Unmarshal(data, ext); err == nil {
		return nil
	}

	// Fall back to YAML
	return yaml.Unmarshal(data, ext)
}

// IsInstalled checks if an extension is installed.
func IsInstalled(name string) bool {
	extDir := filepath.Join(GetExtensionsDir(), name)
	_, err := os.Stat(extDir)
	return err == nil
}

// GetExecutablePath returns the full path to an extension's executable.
func GetExecutablePath(name string) (string, error) {
	ext, err := LoadExtension(name)
	if err != nil {
		return "", err
	}

	extDir := filepath.Join(GetExtensionsDir(), name)
	binName := ext.Bin
	if binName == "" {
		binName = fmt.Sprintf("gpd-%s", ext.Name)
	}

	// Handle Windows executable extension
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	return filepath.Join(extDir, binName), nil
}
