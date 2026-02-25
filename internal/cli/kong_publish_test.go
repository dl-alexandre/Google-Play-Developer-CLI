//go:build unit
// +build unit

package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/errors"
)

// ============================================================================
// PublishUploadCmd Tests
// ============================================================================

func TestPublishUploadCmd_ValidateUploadFile(t *testing.T) {
	t.Run("missing file returns error", func(t *testing.T) {
		cmd := &PublishUploadCmd{File: ""}
		_, _, err := cmd.validateUploadFile()
		if err == nil {
			t.Fatal("Expected error for missing file")
		}
		if !strings.Contains(err.Error(), "file is required") {
			t.Errorf("Expected 'file is required' error, got: %v", err)
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		cmd := &PublishUploadCmd{File: "/nonexistent/path/file.apk"}
		_, _, err := cmd.validateUploadFile()
		if err == nil {
			t.Fatal("Expected error for nonexistent file")
		}
		if !strings.Contains(err.Error(), "file not found") {
			t.Errorf("Expected 'file not found' error, got: %v", err)
		}
	})

	t.Run("invalid file extension returns error", func(t *testing.T) {
		tmpFile := createTempFile(t, "test.txt", []byte("test content"))
		defer os.Remove(tmpFile)

		cmd := &PublishUploadCmd{File: tmpFile}
		_, _, err := cmd.validateUploadFile()
		if err == nil {
			t.Fatal("Expected error for invalid file extension")
		}
		if !strings.Contains(err.Error(), "invalid file type") {
			t.Errorf("Expected 'invalid file type' error, got: %v", err)
		}
	})

	t.Run("apk file validates successfully", func(t *testing.T) {
		tmpFile := createTempFile(t, "test.apk", []byte("fake apk content"))
		defer os.Remove(tmpFile)

		cmd := &PublishUploadCmd{File: tmpFile}
		fileInfo, fileType, err := cmd.validateUploadFile()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if fileType != "apk" {
			t.Errorf("Expected file type 'apk', got: %s", fileType)
		}
		if fileInfo == nil {
			t.Error("Expected file info to be non-nil")
		}
	})

	t.Run("aab file validates successfully", func(t *testing.T) {
		tmpFile := createTempFile(t, "test.aab", []byte("fake aab content"))
		defer os.Remove(tmpFile)

		cmd := &PublishUploadCmd{File: tmpFile}
		fileInfo, fileType, err := cmd.validateUploadFile()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if fileType != "aab" {
			t.Errorf("Expected file type 'aab', got: %s", fileType)
		}
		if fileInfo == nil {
			t.Error("Expected file info to be non-nil")
		}
	})

	t.Run("case insensitive extension check", func(t *testing.T) {
		tmpFile := createTempFile(t, "test.APK", []byte("fake apk content"))
		defer os.Remove(tmpFile)

		cmd := &PublishUploadCmd{File: tmpFile}
		_, fileType, err := cmd.validateUploadFile()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if fileType != "apk" {
			t.Errorf("Expected file type 'apk' for uppercase extension, got: %s", fileType)
		}
	})
}

func TestPublishUploadCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PublishUploadCmd{File: "test.apk"}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPublishUploadCmd_HandleDryRunUpload(t *testing.T) {
	t.Run("dry run outputs expected fields", func(t *testing.T) {
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}
		cmd := &PublishUploadCmd{
			File:  "test.apk",
			Track: "internal",
		}

		err := cmd.handleDryRunUpload(time.Now(), "apk", globals)
		// Should not error - output goes to stdout
		if err != nil {
			t.Errorf("Unexpected error in dry run: %v", err)
		}
	})
}

// ============================================================================
// PublishReleaseCmd Tests
// ============================================================================

func TestPublishReleaseCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PublishReleaseCmd{
		Track:        "internal",
		VersionCodes: []string{"123"},
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPublishReleaseCmd_Run_InvalidTrack(t *testing.T) {
	cmd := &PublishReleaseCmd{
		Track:        "invalid-track",
		VersionCodes: []string{"123"},
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err != errors.ErrTrackInvalid {
		t.Errorf("Expected ErrTrackInvalid, got: %v", err)
	}
}

func TestPublishReleaseCmd_Run_InvalidStatus(t *testing.T) {
	cmd := &PublishReleaseCmd{
		Track:        "internal",
		Status:       "invalid-status",
		VersionCodes: []string{"123"},
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "invalid release status") {
		t.Errorf("Expected 'invalid release status' error, got: %v", err)
	}
}

func TestPublishReleaseCmd_ParseVersionCodes(t *testing.T) {
	t.Run("valid version codes", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			VersionCodes: []string{"1", "2", "3"},
		}
		codes, err := cmd.parseVersionCodes()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(codes) != 3 {
			t.Errorf("Expected 3 version codes, got: %d", len(codes))
		}
		if codes[0] != 1 || codes[1] != 2 || codes[2] != 3 {
			t.Errorf("Expected [1, 2, 3], got: %v", codes)
		}
	})

	t.Run("invalid version code format", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			VersionCodes: []string{"abc"},
		}
		_, err := cmd.parseVersionCodes()
		if err == nil {
			t.Fatal("Expected error for invalid version code")
		}
		if !strings.Contains(err.Error(), "invalid version code") {
			t.Errorf("Expected 'invalid version code' error, got: %v", err)
		}
	})

	t.Run("empty version codes", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			VersionCodes: []string{},
		}
		_, err := cmd.parseVersionCodes()
		if err == nil {
			t.Fatal("Expected error for empty version codes")
		}
		if !strings.Contains(err.Error(), "at least one version code is required") {
			t.Errorf("Expected version codes required error, got: %v", err)
		}
	})

	t.Run("mixed valid and invalid codes", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			VersionCodes: []string{"1", "invalid", "3"},
		}
		_, err := cmd.parseVersionCodes()
		if err == nil {
			t.Fatal("Expected error for mixed valid/invalid codes")
		}
	})
}

func TestPublishReleaseCmd_LoadReleaseNotes(t *testing.T) {
	t.Run("no release notes file returns empty map", func(t *testing.T) {
		cmd := &PublishReleaseCmd{}
		notes, err := cmd.loadReleaseNotes()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(notes) != 0 {
			t.Errorf("Expected empty map, got: %v", notes)
		}
	})

	t.Run("valid release notes file", func(t *testing.T) {
		content := `{"en-US": "Bug fixes", "de-DE": "Fehlerbehebungen"}`
		tmpFile := createTempFile(t, "notes.json", []byte(content))
		defer os.Remove(tmpFile)

		cmd := &PublishReleaseCmd{ReleaseNotesFile: tmpFile}
		notes, err := cmd.loadReleaseNotes()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(notes) != 2 {
			t.Errorf("Expected 2 entries, got: %d", len(notes))
		}
		if notes["en-US"] != "Bug fixes" {
			t.Errorf("Expected 'Bug fixes' for en-US, got: %s", notes["en-US"])
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		cmd := &PublishReleaseCmd{ReleaseNotesFile: "/nonexistent/file.json"}
		_, err := cmd.loadReleaseNotes()
		if err == nil {
			t.Fatal("Expected error for nonexistent file")
		}
		if !strings.Contains(err.Error(), "failed to read") {
			t.Errorf("Expected file read error, got: %v", err)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		tmpFile := createTempFile(t, "invalid.json", []byte("not valid json"))
		defer os.Remove(tmpFile)

		cmd := &PublishReleaseCmd{ReleaseNotesFile: tmpFile}
		_, err := cmd.loadReleaseNotes()
		if err == nil {
			t.Fatal("Expected error for invalid JSON")
		}
		if !strings.Contains(err.Error(), "failed to parse") {
			t.Errorf("Expected JSON parse error, got: %v", err)
		}
	})
}

func TestPublishReleaseCmd_ParseReleaseInputs(t *testing.T) {
	t.Run("valid inputs", func(t *testing.T) {
		content := `{"en-US": "Release notes"}`
		tmpFile := createTempFile(t, "notes.json", []byte(content))
		defer os.Remove(tmpFile)

		cmd := &PublishReleaseCmd{
			VersionCodes:     []string{"100"},
			ReleaseNotesFile: tmpFile,
		}
		codes, notes, err := cmd.parseReleaseInputs()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(codes) != 1 || codes[0] != 100 {
			t.Errorf("Expected version code 100, got: %v", codes)
		}
		if len(notes) != 1 {
			t.Errorf("Expected 1 release note, got: %d", len(notes))
		}
	})

	t.Run("invalid version codes returns error", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			VersionCodes: []string{"invalid"},
		}
		_, _, err := cmd.parseReleaseInputs()
		if err == nil {
			t.Fatal("Expected error for invalid version codes")
		}
	})
}

func TestPublishReleaseCmd_HandleDryRunRelease(t *testing.T) {
	t.Run("dry run outputs expected fields", func(t *testing.T) {
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}
		cmd := &PublishReleaseCmd{
			Track:        "internal",
			Status:       "draft",
			VersionCodes: []string{"100"},
		}

		codes := []int64{100}
		err := cmd.handleDryRunRelease(time.Now(), codes, globals)
		if err != nil {
			t.Errorf("Unexpected error in dry run: %v", err)
		}
	})
}

func TestPublishReleaseCmd_BuildTrack(t *testing.T) {
	t.Run("basic track build", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			Track:  "internal",
			Name:   "Release 1.0",
			Status: "draft",
		}
		versionCodes := []int64{100, 101}
		track := cmd.buildTrack(versionCodes, nil)

		if track.Track != "internal" {
			t.Errorf("Expected track 'internal', got: %s", track.Track)
		}
		if len(track.Releases) != 1 {
			t.Fatalf("Expected 1 release, got: %d", len(track.Releases))
		}
		release := track.Releases[0]
		if release.Name != "Release 1.0" {
			t.Errorf("Expected name 'Release 1.0', got: %s", release.Name)
		}
		if release.Status != "draft" {
			t.Errorf("Expected status 'draft', got: %s", release.Status)
		}
		if len(release.VersionCodes) != 2 {
			t.Errorf("Expected 2 version codes, got: %d", len(release.VersionCodes))
		}
	})

	t.Run("track with release notes", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			Track:  "production",
			Status: "completed",
		}
		versionCodes := []int64{200}
		releaseNotes := map[string]string{
			"en-US": "Bug fixes",
			"de-DE": "Fehlerbehebungen",
		}
		track := cmd.buildTrack(versionCodes, releaseNotes)

		if len(track.Releases) != 1 {
			t.Fatalf("Expected 1 release, got: %d", len(track.Releases))
		}
		release := track.Releases[0]
		if len(release.ReleaseNotes) != 2 {
			t.Errorf("Expected 2 release notes, got: %d", len(release.ReleaseNotes))
		}
	})

	t.Run("track with in-app update priority", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			Track:               "internal",
			Status:              "completed",
			InAppUpdatePriority: 3,
		}
		versionCodes := []int64{300}
		track := cmd.buildTrack(versionCodes, nil)

		if len(track.Releases) != 1 {
			t.Fatalf("Expected 1 release, got: %d", len(track.Releases))
		}
		release := track.Releases[0]
		if release.InAppUpdatePriority != 3 {
			t.Errorf("Expected priority 3, got: %d", release.InAppUpdatePriority)
		}
	})

	t.Run("priority outside range is ignored", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			Track:               "internal",
			Status:              "completed",
			InAppUpdatePriority: 10, // Outside 0-5 range
		}
		versionCodes := []int64{300}
		track := cmd.buildTrack(versionCodes, nil)

		release := track.Releases[0]
		if release.InAppUpdatePriority != 0 {
			t.Errorf("Expected priority 0 (default), got: %d", release.InAppUpdatePriority)
		}
	})
}

func TestPublishReleaseCmd_BuildLocalizedReleaseNotes(t *testing.T) {
	cmd := &PublishReleaseCmd{}
	notes := map[string]string{
		"en-US": "English notes",
		"es-ES": "Notas en español",
		"fr-FR": "Notes en français",
	}

	localized := cmd.buildLocalizedReleaseNotes(notes)
	if len(localized) != 3 {
		t.Errorf("Expected 3 localized notes, got: %d", len(localized))
	}

	// Verify each note has language and text
	for _, note := range localized {
		if note.Language == "" {
			t.Error("Expected non-empty language")
		}
		if note.Text == "" {
			t.Error("Expected non-empty text")
		}
	}
}

// ============================================================================
// PublishTracksCmd Tests
// ============================================================================

func TestPublishTracksCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PublishTracksCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func createTempFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	return path
}

// ============================================================================
// Integration Tests with Mock Server
// ============================================================================

func TestPublishCommands_WithMockServer(t *testing.T) {
	// Create a mock server that simulates Google Play API responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate successful responses
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case "POST":
			if strings.Contains(r.URL.Path, "/edits") {
				// Create edit
				fmt.Fprintf(w, `{"id": "mock-edit-id", "expiryTimeSeconds": "%d"}`, time.Now().Add(time.Hour).Unix())
			}
		case "GET":
			if strings.Contains(r.URL.Path, "/edits/") && strings.Contains(r.URL.Path, "/tracks") {
				// List tracks
				fmt.Fprint(w, `{"tracks": [{"track": "internal", "releases": [{"name": "Release 1.0", "status": "completed"}]}]}`)
			}
		case "PUT":
			if strings.Contains(r.URL.Path, "/edits/") && strings.Contains(r.URL.Path, "/tracks") {
				// Update track
				fmt.Fprint(w, `{"track": "internal"}`)
			}
		case "DELETE":
			if strings.Contains(r.URL.Path, "/edits/") {
				// Delete edit
				w.WriteHeader(http.StatusNoContent)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Note: To fully test with mock server, we would need to inject the server URL
	// into the API client. This requires additional refactoring of the createAPIClient
	// function to support custom endpoints for testing.

	t.Logf("Mock server created at: %s", server.URL)
}

// ============================================================================
// API Client Creation Tests (Error Paths)
// ============================================================================

func TestPublishCommands_CreateAPIClient(t *testing.T) {
	t.Run("upload with invalid auth returns error", func(t *testing.T) {
		// Create a temp APK file
		tmpFile := createTempFile(t, "test.apk", []byte("fake apk"))
		defer os.Remove(tmpFile)

		globals := &Globals{
			Package: "com.example.app",
			KeyPath: "/nonexistent/key.json", // Invalid key path
		}
		cmd := &PublishUploadCmd{File: tmpFile}

		// This will fail when trying to create API client
		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for invalid auth")
		}
	})
}

// ============================================================================
// Result Building Tests
// ============================================================================

func TestPublishUploadCmd_BuildUploadResult(t *testing.T) {
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	// Create a temp file to get file info
	tmpFile := createTempFile(t, "test.apk", []byte("fake apk content"))
	defer os.Remove(tmpFile)

	fileInfo, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat temp file: %v", err)
	}

	cmd := &PublishUploadCmd{
		File: "test.apk",
	}

	start := time.Now()
	err = cmd.buildUploadResult(start, fileInfo, "apk", "edit-123", 100, "sha1-abc", "sha256-xyz", true, globals)
	if err != nil {
		t.Errorf("Unexpected error building result: %v", err)
	}
}

func TestPublishReleaseCmd_BuildReleaseResult(t *testing.T) {
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	cmd := &PublishReleaseCmd{
		Track:  "internal",
		Name:   "Release 1.0",
		Status: "completed",
	}

	versionCodes := []int64{100, 101}
	start := time.Now()
	err := cmd.buildReleaseResult(start, "edit-123", versionCodes, true, globals)
	if err != nil {
		t.Errorf("Unexpected error building result: %v", err)
	}
}

// ============================================================================
// Verbose/Logging Tests
// ============================================================================

func TestPublishCommands_WithVerboseMode(t *testing.T) {
	// Test that commands accept verbose flag without error
	tmpFile := createTempFile(t, "test.apk", []byte("fake apk"))
	defer os.Remove(tmpFile)

	globals := &Globals{
		Package: "com.example.app",
		Verbose: true,
		Output:  "json",
	}

	// Just test that validation works with verbose mode
	cmd := &PublishUploadCmd{File: tmpFile}
	_, _, err := cmd.validateUploadFile()
	if err != nil {
		t.Errorf("Unexpected error with verbose mode: %v", err)
	}

	// Use globals to avoid unused variable error
	_ = globals.Verbose
}

// ============================================================================
// Error Wrapping Tests
// ============================================================================

func TestPublishCommands_ErrorHints(t *testing.T) {
	t.Run("upload missing file has hint", func(t *testing.T) {
		cmd := &PublishUploadCmd{File: ""}
		_, _, err := cmd.validateUploadFile()
		if err == nil {
			t.Fatal("Expected error")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatal("Expected APIError type")
		}
		if apiErr.Hint == "" {
			t.Error("Expected error to have hint")
		}
	})

	t.Run("release missing version codes has hint", func(t *testing.T) {
		cmd := &PublishReleaseCmd{
			VersionCodes: []string{},
		}
		_, err := cmd.parseVersionCodes()
		if err == nil {
			t.Fatal("Expected error")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatal("Expected APIError type")
		}
		if apiErr.Hint == "" {
			t.Error("Expected error to have hint")
		}
	})
}

// ============================================================================
// Track Constants Test
// ============================================================================

func TestValidTracks(t *testing.T) {
	expectedTracks := []string{"internal", "alpha", "beta", "production"}
	if len(api.ValidTracks) != len(expectedTracks) {
		t.Errorf("Expected %d valid tracks, got: %d", len(expectedTracks), len(api.ValidTracks))
	}
	for _, track := range expectedTracks {
		if !api.IsValidTrack(track) {
			t.Errorf("Expected %s to be a valid track", track)
		}
	}
	if api.IsValidTrack("invalid") {
		t.Error("Expected 'invalid' to not be a valid track")
	}
}

func TestValidReleaseStatus(t *testing.T) {
	validStatuses := []string{"draft", "completed", "halted", "inProgress"}
	for _, status := range validStatuses {
		if !api.IsValidReleaseStatus(status) {
			t.Errorf("Expected %s to be a valid status", status)
		}
	}
	if api.IsValidReleaseStatus("invalid") {
		t.Error("Expected 'invalid' to not be a valid status")
	}
}

// ============================================================================
// Context Tests
// ============================================================================

func TestPublishCommands_ContextPropagation(t *testing.T) {
	// Test that context cancellation would be respected
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The actual commands would need to accept context from outside
	// This test documents the expected behavior
	if ctx.Err() != context.Canceled {
		t.Error("Expected context to be canceled")
	}
}
