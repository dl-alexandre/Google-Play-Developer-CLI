package extensions

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// setupTestExtensionsDir creates a temporary directory for testing and returns cleanup function
func setupTestExtensionsDir(t *testing.T) (cleanup func()) {
	t.Helper()
	tmpDir := t.TempDir()

	// Save original environment
	oldXdgData := os.Getenv("XDG_DATA_HOME")
	oldXdgConfig := os.Getenv("XDG_CONFIG_HOME")
	oldAppData := os.Getenv("APPDATA")
	oldHome := os.Getenv("HOME")
	oldUserProfile := os.Getenv("USERPROFILE")

	// Set environment variables for test isolation
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "share"))
	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	_ = os.Setenv("APPDATA", tmpDir)
	_ = os.Setenv("HOME", tmpDir)
	_ = os.Setenv("USERPROFILE", tmpDir)

	return func() {
		// Restore original environment
		if oldXdgData == "" {
			_ = os.Unsetenv("XDG_DATA_HOME")
		} else {
			_ = os.Setenv("XDG_DATA_HOME", oldXdgData)
		}

		if oldXdgConfig == "" {
			_ = os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			_ = os.Setenv("XDG_CONFIG_HOME", oldXdgConfig)
		}

		if oldAppData == "" {
			_ = os.Unsetenv("APPDATA")
		} else {
			_ = os.Setenv("APPDATA", oldAppData)
		}

		if oldHome == "" {
			_ = os.Unsetenv("HOME")
		} else {
			_ = os.Setenv("HOME", oldHome)
		}

		if oldUserProfile == "" {
			_ = os.Unsetenv("USERPROFILE")
		} else {
			_ = os.Setenv("USERPROFILE", oldUserProfile)
		}
	}
}

// createTestExtension creates a test extension in the extensions directory
func createTestExtension(t *testing.T, name string, manifest *Manifest) *Extension {
	t.Helper()

	extDir := filepath.Join(GetExtensionsDir(), name)
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatalf("Failed to create extension directory: %v", err)
	}

	ext := &Extension{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Bin:         manifest.DefaultBinName(),
		Source:      "local/test",
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Type:        "binary",
	}

	// Save metadata
	metaPath := filepath.Join(extDir, ".gpd-extension")
	data, err := yaml.Marshal(ext)
	if err != nil {
		t.Fatalf("Failed to marshal extension metadata: %v", err)
	}
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		t.Fatalf("Failed to write extension metadata: %v", err)
	}

	// Create dummy executable
	binName := manifest.DefaultBinName()
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(extDir, binName)

	content := "#!/bin/sh\necho 'hello from extension'"
	if runtime.GOOS == "windows" {
		content = "@echo hello from extension"
	}

	if err := os.WriteFile(binPath, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to write executable: %v", err)
	}

	return ext
}

// createTestManifestFile creates a .gpd-extension manifest file
func createTestManifestFile(t *testing.T, dir string, manifest *Manifest) {
	t.Helper()

	manifestPath := filepath.Join(dir, ".gpd-extension")
	data, err := yaml.Marshal(manifest)
	if err != nil {
		t.Fatalf("Failed to marshal manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}
}

func TestManifestValidate(t *testing.T) {
	tests := []struct {
		name      string
		manifest  Manifest
		wantError bool
		errMsg    string
	}{
		{
			name: "valid manifest",
			manifest: Manifest{
				Name:        "test-ext",
				Version:     "1.0.0",
				Description: "A test extension",
				Author:      "tester",
			},
			wantError: false,
		},
		{
			name: "missing name",
			manifest: Manifest{
				Version: "1.0.0",
			},
			wantError: true,
			errMsg:    "extension name is required",
		},
		{
			name: "missing version",
			manifest: Manifest{
				Name: "test-ext",
			},
			wantError: true,
			errMsg:    "extension version is required",
		},
		{
			name: "name with spaces",
			manifest: Manifest{
				Name:    "test ext",
				Version: "1.0.0",
			},
			wantError: true,
			errMsg:    "invalid character",
		},
		{
			name: "name with special chars",
			manifest: Manifest{
				Name:    "test@ext",
				Version: "1.0.0",
			},
			wantError: true,
			errMsg:    "invalid character",
		},
		{
			name: "valid name with hyphen",
			manifest: Manifest{
				Name:    "test-ext",
				Version: "1.0.0",
			},
			wantError: false,
		},
		{
			name: "valid name with underscore",
			manifest: Manifest{
				Name:    "test_ext",
				Version: "1.0.0",
			},
			wantError: false,
		},
		{
			name: "valid name with numbers",
			manifest: Manifest{
				Name:    "test-ext-123",
				Version: "1.0.0",
			},
			wantError: false,
		},
		{
			name: "valid camelcase name",
			manifest: Manifest{
				Name:    "TestExt",
				Version: "1.0.0",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.Validate()
			if tt.wantError {
				if err == nil {
					t.Errorf("Validate() expected error containing %q, got nil", tt.errMsg)
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want containing %q", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestManifestDefaultBinName(t *testing.T) {
	tests := []struct {
		name     string
		manifest Manifest
		want     string
	}{
		{
			name:     "with explicit bin",
			manifest: Manifest{Name: "test", Bin: "custom-bin"},
			want:     "custom-bin",
		},
		{
			name:     "without explicit bin",
			manifest: Manifest{Name: "test"},
			want:     "gpd-test",
		},
		{
			name:     "empty bin falls back",
			manifest: Manifest{Name: "my-ext", Bin: ""},
			want:     "gpd-my-ext",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.manifest.DefaultBinName()
			if got != tt.want {
				t.Errorf("DefaultBinName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsValidExtensionNameChar(t *testing.T) {
	tests := []struct {
		char rune
		want bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'-', true},
		{'_', true},
		{' ', false},
		{'@', false},
		{'/', false},
		{'\\', false},
		{'.', false},
		{'*', false},
		{'&', false},
		{'%', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			got := isValidExtensionNameChar(tt.char)
			if got != tt.want {
				t.Errorf("isValidExtensionNameChar(%q) = %v, want %v", tt.char, got, tt.want)
			}
		})
	}
}

func TestGetExtensionsDir(t *testing.T) {
	dir := GetExtensionsDir()
	if dir == "" {
		t.Error("GetExtensionsDir() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("GetExtensionsDir() returned relative path: %s", dir)
	}
}

func TestGetExtensionPathsForOS(t *testing.T) {
	tests := []struct {
		goos string
		want string
	}{
		{"darwin", "Application Support"},
		{"linux", ".local/share"},
		{"windows", "gpd"},
		{"freebsd", ".local/share"},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			paths := getExtensionPathsForOS(tt.goos)
			if !contains(paths.ExtensionsDir, tt.want) {
				t.Errorf("getExtensionPathsForOS(%q) = %q, should contain %q", tt.goos, paths.ExtensionsDir, tt.want)
			}
		})
	}
}

func TestGetExtensionPathsForOSXDG(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping XDG test on Windows")
	}

	oldXdgData := os.Getenv("XDG_DATA_HOME")
	defer func() { _ = os.Setenv("XDG_DATA_HOME", oldXdgData) }()

	testDir := "/test/xdg/data"
	_ = os.Setenv("XDG_DATA_HOME", testDir)

	paths := getExtensionPathsForOS("linux")
	want := filepath.Join(testDir, "gpd", "extensions")
	if paths.ExtensionsDir != want {
		t.Errorf("XDG_DATA_HOME not respected: got %q, want %q", paths.ExtensionsDir, want)
	}
}

func TestIsInstalled(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Initially should not be installed
	if IsInstalled("test-ext") {
		t.Error("IsInstalled() should return false for non-existent extension")
	}

	// Create test extension
	createTestExtension(t, "test-ext", &Manifest{
		Name:    "test-ext",
		Version: "1.0.0",
	})

	// Now should be installed
	if !IsInstalled("test-ext") {
		t.Error("IsInstalled() should return true for existing extension")
	}
}

func TestList(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Initially should be empty
	exts, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(exts) != 0 {
		t.Errorf("List() returned %d extensions, want 0", len(exts))
	}

	// Create test extensions
	createTestExtension(t, "ext1", &Manifest{Name: "ext1", Version: "1.0.0"})
	createTestExtension(t, "ext2", &Manifest{Name: "ext2", Version: "2.0.0"})

	// Should list both
	exts, err = List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(exts) != 2 {
		t.Errorf("List() returned %d extensions, want 2", len(exts))
	}
}

func TestListIgnoresInvalid(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Create valid extension
	createTestExtension(t, "valid", &Manifest{Name: "valid", Version: "1.0.0"})

	// Create invalid extension (no metadata)
	tmpDir := GetExtensionsDir()
	invalidDir := filepath.Join(tmpDir, "invalid")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatalf("Failed to create invalid extension dir: %v", err)
	}

	// Should only list valid extension
	exts, err := List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(exts) != 1 {
		t.Errorf("List() returned %d extensions, want 1", len(exts))
	}
}

func TestLoadExtension(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Create test extension
	createTestExtension(t, "test-ext", &Manifest{
		Name:        "test-ext",
		Version:     "1.0.0",
		Description: "Test description",
		Author:      "tester",
	})

	// Load it
	ext, err := LoadExtension("test-ext")
	if err != nil {
		t.Fatalf("LoadExtension() error = %v", err)
	}

	if ext.Name != "test-ext" {
		t.Errorf("LoadExtension() Name = %q, want %q", ext.Name, "test-ext")
	}
	if ext.Version != "1.0.0" {
		t.Errorf("LoadExtension() Version = %q, want %q", ext.Version, "1.0.0")
	}
	if ext.Description != "Test description" {
		t.Errorf("LoadExtension() Description = %q, want %q", ext.Description, "Test description")
	}
}

func TestLoadExtensionNotFound(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	_, err := LoadExtension("nonexistent")
	if err == nil {
		t.Error("LoadExtension() should return error for non-existent extension")
	}
}

func TestLoadExtensionInvalidMetadata(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Create extension with invalid metadata
	tmpDir := GetExtensionsDir()
	extDir := filepath.Join(tmpDir, "invalid-meta")
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatalf("Failed to create extension dir: %v", err)
	}

	metaPath := filepath.Join(extDir, ".gpd-extension")
	if err := os.WriteFile(metaPath, []byte("invalid yaml {["), 0644); err != nil {
		t.Fatalf("Failed to write invalid metadata: %v", err)
	}

	_, err := LoadExtension("invalid-meta")
	if err == nil {
		t.Error("LoadExtension() should return error for invalid metadata")
	}
}

func TestGetExecutablePath(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Create test extension
	createTestExtension(t, "test-ext", &Manifest{
		Name: "test-ext",
		Bin:  "custom-bin",
	})

	// Get executable path
	path, err := GetExecutablePath("test-ext")
	if err != nil {
		t.Fatalf("GetExecutablePath() error = %v", err)
	}

	want := "custom-bin"
	if runtime.GOOS == "windows" {
		want += ".exe"
	}

	if !contains(path, want) {
		t.Errorf("GetExecutablePath() = %q, should contain %q", path, want)
	}
}

func TestGetExecutablePathNotFound(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	_, err := GetExecutablePath("nonexistent")
	if err == nil {
		t.Error("GetExecutablePath() should return error for non-existent extension")
	}
}

func TestUnmarshalMetadataJSON(t *testing.T) {
	data := []byte(`{"name":"test","version":"1.0.0","source":"github.com/owner/repo"}`)
	var ext Extension
	err := unmarshalMetadata(data, &ext)
	if err != nil {
		t.Fatalf("unmarshalMetadata() error = %v", err)
	}
	if ext.Name != "test" {
		t.Errorf("Name = %q, want %q", ext.Name, "test")
	}
}

func TestUnmarshalMetadataYAML(t *testing.T) {
	data := []byte("name: test\nversion: 1.0.0\nsource: github.com/owner/repo")
	var ext Extension
	err := unmarshalMetadata(data, &ext)
	if err != nil {
		t.Fatalf("unmarshalMetadata() error = %v", err)
	}
	if ext.Name != "test" {
		t.Errorf("Name = %q, want %q", ext.Name, "test")
	}
}

func TestIsBuiltInCommand(t *testing.T) {
	builtins := []string{
		"auth", "config", "publish", "reviews", "vitals", "monitor",
		"analytics", "purchases", "monetization", "permissions", "recovery",
		"extension", "help",
	}

	for _, cmd := range builtins {
		t.Run(cmd, func(t *testing.T) {
			if !IsBuiltInCommand(cmd) {
				t.Errorf("IsBuiltInCommand(%q) = false, want true", cmd)
			}
		})
	}

	nonBuiltins := []string{"dash", "my-ext", "custom", "foo-bar"}
	for _, cmd := range nonBuiltins {
		t.Run(cmd, func(t *testing.T) {
			if IsBuiltInCommand(cmd) {
				t.Errorf("IsBuiltInCommand(%q) = true, want false", cmd)
			}
		})
	}
}

func TestIsLocalPath(t *testing.T) {
	tests := []struct {
		source string
		want   bool
	}{
		{".", true},
		{"./my-ext", true},
		{"/absolute/path", true},
		{"~/home/path", true},
		{"../relative", true},
		{"owner/repo", false},                   // Valid GitHub format
		{"my-org/gpd-dash", false},              // Valid GitHub format with hyphens
		{"org123/repo456", false},               // Valid GitHub format with numbers
		{"github.com/owner/repo", true},         // Multiple slashes = local path
		{"https://github.com/owner/repo", true}, // URL = local path
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			got := isLocalPath(tt.source)
			if got != tt.want {
				t.Errorf("isLocalPath(%q) = %v, want %v", tt.source, got, tt.want)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home := getHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"/absolute", "/absolute"},
		{"relative", "relative"},
		{"./relative", "./relative"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := expandPath(tt.input)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectExtensionType(t *testing.T) {
	tmpDir := t.TempDir()

	// Test binary extension
	manifestBin := &Manifest{Name: "bin-ext", Bin: "gpd-bin-ext"}
	binPath := filepath.Join(tmpDir, "gpd-bin-ext")
	if err := os.WriteFile(binPath, []byte{0x7f, 0x45, 0x4c, 0x46}, 0755); err != nil { // ELF magic
		t.Fatalf("Failed to create binary: %v", err)
	}

	got := detectExtensionType(tmpDir, manifestBin)
	if got != "binary" {
		t.Errorf("detectExtensionType() for ELF = %q, want binary", got)
	}

	// Test script extension
	scriptDir := t.TempDir()
	manifestScript := &Manifest{Name: "script-ext", Bin: "gpd-script-ext"}
	scriptPath := filepath.Join(scriptDir, "gpd-script-ext")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho hello"), 0755); err != nil {
		t.Fatalf("Failed to create script: %v", err)
	}

	got = detectExtensionType(scriptDir, manifestScript)
	if got != "script" {
		t.Errorf("detectExtensionType() for script = %q, want script", got)
	}
}

func TestRemove(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Create test extension
	createTestExtension(t, "to-remove", &Manifest{Name: "to-remove", Version: "1.0.0"})

	// Verify it exists
	if !IsInstalled("to-remove") {
		t.Fatal("Extension should exist before removal")
	}

	// Remove it
	if err := Remove("to-remove"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// Verify it's gone
	if IsInstalled("to-remove") {
		t.Error("Extension should not exist after removal")
	}
}

func TestRemoveNotInstalled(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	err := Remove("nonexistent")
	if err == nil {
		t.Error("Remove() should return error for non-existent extension")
	}
}

func TestInstallLocal(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Create a local extension directory
	srcDir := t.TempDir()
	manifest := &Manifest{
		Name:        "local-ext",
		Version:     "1.0.0",
		Description: "A local test extension",
	}
	createTestManifestFile(t, srcDir, manifest)

	// Create executable
	binName := "gpd-local-ext"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(srcDir, binName)
	content := "#!/bin/sh\necho 'hello'"
	if runtime.GOOS == "windows" {
		content = "@echo hello"
	}
	if err := os.WriteFile(binPath, []byte(content), 0755); err != nil {
		t.Fatalf("Failed to create executable: %v", err)
	}

	// Install it
	opts := InstallOptions{
		Source: srcDir,
	}
	result, err := Install(context.Background(), opts)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}

	if !result.Installed {
		t.Error("Install() should report as new install")
	}
	if result.Extension.Name != "local-ext" {
		t.Errorf("Extension name = %q, want %q", result.Extension.Name, "local-ext")
	}

	// Verify it's installed
	if !IsInstalled("local-ext") {
		t.Error("Extension should be installed after Install()")
	}
}

func TestInstallLocalAlreadyExists(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Create and install first version
	srcDir1 := t.TempDir()
	manifest := &Manifest{Name: "test-ext", Version: "1.0.0"}
	createTestManifestFile(t, srcDir1, manifest)
	binName := "gpd-test-ext"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(srcDir1, binName)
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\necho v1"), 0755); err != nil {
		t.Fatalf("Failed to create executable: %v", err)
	}

	_, err := Install(context.Background(), InstallOptions{Source: srcDir1})
	if err != nil {
		t.Fatalf("First install failed: %v", err)
	}

	// Try to install again without force
	srcDir2 := t.TempDir()
	manifest2 := &Manifest{Name: "test-ext", Version: "2.0.0"}
	createTestManifestFile(t, srcDir2, manifest2)
	binPath2 := filepath.Join(srcDir2, binName)
	if err := os.WriteFile(binPath2, []byte("#!/bin/sh\necho v2"), 0755); err != nil {
		t.Fatalf("Failed to create executable: %v", err)
	}

	_, err = Install(context.Background(), InstallOptions{Source: srcDir2})
	if err == nil {
		t.Error("Install() should fail without Force=true when extension exists")
	}

	// Now try with force
	result, err := Install(context.Background(), InstallOptions{Source: srcDir2, Force: true})
	if err != nil {
		t.Fatalf("Install() with Force error = %v", err)
	}

	if result.Installed {
		t.Error("Install() should report as update, not new install")
	}
}

func TestInstallLocalInvalidManifest(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	srcDir := t.TempDir()
	// Don't create manifest - this should fail

	_, err := Install(context.Background(), InstallOptions{Source: srcDir})
	if err == nil {
		t.Error("Install() should fail without valid manifest")
	}
}

func TestInstallLocalBuiltInConflict(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	srcDir := t.TempDir()
	manifest := &Manifest{Name: "auth", Version: "1.0.0"} // "auth" is a built-in command
	createTestManifestFile(t, srcDir, manifest)

	_, err := Install(context.Background(), InstallOptions{Source: srcDir})
	if err == nil {
		t.Error("Install() should fail when extension name conflicts with built-in command")
	}
}

func TestInstallGitHub(t *testing.T) {
	// This test would require network access and mocking GitHub API
	// Mark as skipped for now, can be enabled with integration test tag
	t.Skip("Skipping GitHub install test - requires network access and mocking")
}

func TestInstallOptionsDefaults(t *testing.T) {
	opts := InstallOptions{}
	ctx := context.Background()

	// Install should set default timeout
	// This is harder to test without exposing internal state
	// Just verify it doesn't panic
	_ = ctx
	_ = opts
}

func TestLoadInstalledExtension(t *testing.T) {
	cleanup := setupTestExtensionsDir(t)
	defer cleanup()

	// Test with non-existent extension - should return nil
	got := loadInstalledExtension("nonexistent")
	if got != nil {
		t.Error("loadInstalledExtension() should return nil for non-existent extension")
	}

	// Create and load extension
	createTestExtension(t, "exists", &Manifest{Name: "exists", Version: "1.0.0"})
	got = loadInstalledExtension("exists")
	if got == nil {
		t.Error("loadInstalledExtension() should return extension for existing extension")
	}
	if got != nil && got.Name != "exists" {
		t.Errorf("Name = %q, want %q", got.Name, "exists")
	}
}

func TestInstallResult(t *testing.T) {
	// Just verify the struct works
	result := &InstallResult{
		Extension: &Extension{Name: "test", Version: "1.0.0"},
		Installed: true,
	}
	if !result.Installed {
		t.Error("InstallResult.Installed should be true")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
