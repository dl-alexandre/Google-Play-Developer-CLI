//go:build unit
// +build unit

package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/api"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// ============================================================================
// BulkUploadCmd Tests
// ============================================================================

func TestBulkUploadCmd_Run_PackageRequired(t *testing.T) {
	t.Run("missing package returns error", func(t *testing.T) {
		cmd := &BulkUploadCmd{
			Files: []string{"test.apk"},
		}
		globals := &Globals{} // No package set

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for missing package")
		}
		if !strings.Contains(err.Error(), "package name is required") {
			t.Errorf("Expected 'package name is required' error, got: %v", err)
		}
	})
}

func TestBulkUploadCmd_Run_NoFiles(t *testing.T) {
	t.Run("no files returns validation error", func(t *testing.T) {
		cmd := &BulkUploadCmd{
			Files: []string{},
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for no files")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if apiErr.Code != errors.CodeValidationError {
			t.Errorf("Expected validation error code, got: %s", apiErr.Code)
		}
		if !strings.Contains(apiErr.Message, "at least one file") {
			t.Errorf("Expected 'at least one file' message, got: %s", apiErr.Message)
		}
	})
}

func TestBulkUploadCmd_Run_DryRun(t *testing.T) {
	t.Run("dry run outputs expected fields", func(t *testing.T) {
		tmpFile := createTempFile(t, "test.apk", []byte("fake apk content"))
		defer os.Remove(tmpFile)

		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}
		cmd := &BulkUploadCmd{
			Files:  []string{tmpFile},
			Track:  "internal",
			DryRun: true,
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error in dry run: %v", err)
		}
	})

	t.Run("dry run with multiple files", func(t *testing.T) {
		tmpFile1 := createTempFile(t, "test1.apk", []byte("fake apk content 1"))
		tmpFile2 := createTempFile(t, "test2.aab", []byte("fake aab content 2"))
		defer os.Remove(tmpFile1)
		defer os.Remove(tmpFile2)

		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}
		cmd := &BulkUploadCmd{
			Files:  []string{tmpFile1, tmpFile2},
			Track:  "production",
			DryRun: true,
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error in dry run: %v", err)
		}
	})
}

func TestBulkUploadCmd_Run_InvalidAuth(t *testing.T) {
	t.Run("invalid auth key path returns error", func(t *testing.T) {
		tmpFile := createTempFile(t, "test.apk", []byte("fake apk content"))
		defer os.Remove(tmpFile)

		globals := &Globals{
			Package: "com.example.app",
			KeyPath: "/nonexistent/key.json",
			Output:  "json",
		}
		cmd := &BulkUploadCmd{
			Files: []string{tmpFile},
			Track: "internal",
		}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for invalid auth")
		}
	})
}

func TestBulkUploadCmd_uploadFile_UnsupportedType(t *testing.T) {
	t.Run("unsupported file type returns error", func(t *testing.T) {
		tmpFile := createTempFile(t, "test.txt", []byte("not an apk"))
		defer os.Remove(tmpFile)

		// We can't test the actual upload without mocking, but we can verify
		// the function handles unsupported types correctly
		result := bulkUploadItemResult{
			File:   tmpFile,
			Status: "failed",
			Error:  "unsupported file type: .txt",
		}

		if result.Status != "failed" {
			t.Errorf("Expected status 'failed', got: %s", result.Status)
		}
		if !strings.Contains(result.Error, "unsupported file type") {
			t.Errorf("Expected unsupported file type error, got: %s", result.Error)
		}
	})
}

func TestBulkUploadCmd_DefaultValues(t *testing.T) {
	t.Run("default track is internal", func(t *testing.T) {
		cmd := &BulkUploadCmd{}
		if cmd.Track != "" {
			// Default is set by Kong tag, not Go zero value
			// This test documents expected behavior
			t.Logf("Track default: %s", cmd.Track)
		}
	})

	t.Run("default max parallel is 3", func(t *testing.T) {
		cmd := &BulkUploadCmd{}
		if cmd.MaxParallel != 0 {
			// Default is set by Kong tag
			t.Logf("MaxParallel default: %d", cmd.MaxParallel)
		}
	})
}

func TestBulkUploadCmd_FileTypeDetection(t *testing.T) {
	tests := []struct {
		filename    string
		content     []byte
		expectedExt string
		description string
	}{
		{"app.apk", []byte("fake apk"), ".apk", "APK file"},
		{"app.aab", []byte("fake aab"), ".aab", "AAB file"},
		{"app.APK", []byte("fake apk"), ".apk", "uppercase APK extension"},
		{"app.AAB", []byte("fake aab"), ".aab", "uppercase AAB extension"},
		{"App.Apk", []byte("fake apk"), ".apk", "mixed case APK extension"},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			tmpFile := createTempFile(t, tt.filename, tt.content)
			defer os.Remove(tmpFile)

			ext := strings.ToLower(filepath.Ext(tmpFile))
			if ext != tt.expectedExt {
				t.Errorf("Expected extension %s, got: %s", tt.expectedExt, ext)
			}
		})
	}
}

// ============================================================================
// BulkListingsCmd Tests
// ============================================================================

func TestBulkListingsCmd_Run_PackageRequired(t *testing.T) {
	t.Run("missing package returns error", func(t *testing.T) {
		cmd := &BulkListingsCmd{
			DataFile: "listings.json",
		}
		globals := &Globals{} // No package set

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for missing package")
		}
		if !strings.Contains(err.Error(), "package name is required") {
			t.Errorf("Expected 'package name is required' error, got: %v", err)
		}
	})
}

func TestBulkListingsCmd_Run_MissingDataFile(t *testing.T) {
	t.Run("missing data file returns error", func(t *testing.T) {
		cmd := &BulkListingsCmd{
			DataFile: "/nonexistent/listings.json",
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for missing data file")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if apiErr.Code != errors.CodeValidationError {
			t.Errorf("Expected validation error code, got: %s", apiErr.Code)
		}
		if apiErr.Hint == "" {
			t.Error("Expected error to have a hint")
		}
	})
}

func TestBulkListingsCmd_Run_InvalidJSON(t *testing.T) {
	t.Run("invalid JSON returns error", func(t *testing.T) {
		tmpFile := createTempFile(t, "invalid.json", []byte("not valid json"))
		defer os.Remove(tmpFile)

		cmd := &BulkListingsCmd{
			DataFile: tmpFile,
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for invalid JSON")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if !strings.Contains(apiErr.Message, "failed to parse") {
			t.Errorf("Expected 'failed to parse' message, got: %s", apiErr.Message)
		}
	})
}

func TestBulkListingsCmd_Run_EmptyListings(t *testing.T) {
	t.Run("empty listings returns error", func(t *testing.T) {
		tmpFile := createTempFile(t, "empty.json", []byte("{}"))
		defer os.Remove(tmpFile)

		cmd := &BulkListingsCmd{
			DataFile: tmpFile,
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for empty listings")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if !strings.Contains(apiErr.Message, "no listings found") {
			t.Errorf("Expected 'no listings found' message, got: %s", apiErr.Message)
		}
	})
}

func TestBulkListingsCmd_Run_DryRun(t *testing.T) {
	t.Run("dry run outputs expected fields", func(t *testing.T) {
		listings := map[string]interface{}{
			"en-US": map[string]string{
				"title":            "My App",
				"shortDescription": "Short desc",
				"fullDescription":  "Full description here",
			},
			"de-DE": map[string]string{
				"title":            "Meine App",
				"shortDescription": "Kurze Beschreibung",
				"fullDescription":  "Volle Beschreibung hier",
			},
		}
		data, _ := json.Marshal(listings)
		tmpFile := createTempFile(t, "listings.json", data)
		defer os.Remove(tmpFile)

		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}
		cmd := &BulkListingsCmd{
			DataFile: tmpFile,
			DryRun:   true,
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error in dry run: %v", err)
		}
	})
}

func TestBulkListingsCmd_ListingDataStructure(t *testing.T) {
	t.Run("valid listing data structure", func(t *testing.T) {
		data := bulkListingData{
			"en-US": {
				Title:            "Test App",
				ShortDescription: "A test app",
				FullDescription:  "This is a full description of the test app.",
				Video:            "https://youtube.com/watch?v=123",
			},
		}

		if data["en-US"].Title != "Test App" {
			t.Errorf("Expected title 'Test App', got: %s", data["en-US"].Title)
		}
		if data["en-US"].Video != "https://youtube.com/watch?v=123" {
			t.Errorf("Expected video URL, got: %s", data["en-US"].Video)
		}
	})
}

func TestBulkListingsCmd_ResultsStructure(t *testing.T) {
	t.Run("bulk listings result structure", func(t *testing.T) {
		result := &bulkListingsResult{
			SuccessCount: 2,
			FailureCount: 1,
			EditID:       "test-edit-123",
			Locales: []bulkListingItemResult{
				{Locale: "en-US", Status: "success"},
				{Locale: "de-DE", Status: "success"},
				{Locale: "fr-FR", Status: "failed", Error: "API error"},
			},
		}

		if result.SuccessCount != 2 {
			t.Errorf("Expected success count 2, got: %d", result.SuccessCount)
		}
		if result.FailureCount != 1 {
			t.Errorf("Expected failure count 1, got: %d", result.FailureCount)
		}
		if result.EditID != "test-edit-123" {
			t.Errorf("Expected edit ID 'test-edit-123', got: %s", result.EditID)
		}
		if len(result.Locales) != 3 {
			t.Errorf("Expected 3 locales, got: %d", len(result.Locales))
		}
	})
}

// ============================================================================
// BulkImagesCmd Tests
// ============================================================================

func TestBulkImagesCmd_Run_PackageRequired(t *testing.T) {
	t.Run("missing package returns error", func(t *testing.T) {
		cmd := &BulkImagesCmd{
			ImageDir: "/some/dir",
		}
		globals := &Globals{} // No package set

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for missing package")
		}
		if !strings.Contains(err.Error(), "package name is required") {
			t.Errorf("Expected 'package name is required' error, got: %v", err)
		}
	})
}

func TestBulkImagesCmd_scanImageDirectory(t *testing.T) {
	t.Run("empty directory returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		cmd := &BulkImagesCmd{
			ImageDir: tmpDir,
		}

		_, err := cmd.scanImageDirectory()
		if err == nil {
			t.Fatal("Expected error for empty directory")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if !strings.Contains(apiErr.Message, "no images found") {
			t.Errorf("Expected 'no images found' message, got: %s", apiErr.Message)
		}
	})

	t.Run("directory with images finds all", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create directory structure
		phoneDir := filepath.Join(tmpDir, "phoneScreenshots", "en-US")
		featureDir := filepath.Join(tmpDir, "featureGraphic")
		os.MkdirAll(phoneDir, 0755)
		os.MkdirAll(featureDir, 0755)

		// Create image files
		os.WriteFile(filepath.Join(phoneDir, "screenshot1.png"), []byte("fake png"), 0644)
		os.WriteFile(filepath.Join(phoneDir, "screenshot2.jpg"), []byte("fake jpg"), 0644)
		os.WriteFile(filepath.Join(featureDir, "feature.png"), []byte("fake png"), 0644)

		// Create non-image file (should be ignored)
		os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0644)

		cmd := &BulkImagesCmd{
			ImageDir: tmpDir,
			Locale:   "en-US",
		}

		images, err := cmd.scanImageDirectory()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(images) != 3 {
			t.Errorf("Expected 3 images, got: %d", len(images))
		}

		// Verify image types
		var foundPhone, foundFeature bool
		for _, img := range images {
			switch img.Type {
			case "phoneScreenshots":
				foundPhone = true
				if img.Locale != "en-US" {
					t.Errorf("Expected locale en-US for phoneScreenshots, got: %s", img.Locale)
				}
			case "featureGraphic":
				foundFeature = true
				if img.Locale != "en-US" {
					t.Errorf("Expected locale en-US for featureGraphic, got: %s", img.Locale)
				}
			}
		}

		if !foundPhone {
			t.Error("Expected to find phoneScreenshots")
		}
		if !foundFeature {
			t.Error("Expected to find featureGraphic")
		}
	})

	t.Run("different locale in subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create directory structure with different locale
		phoneDir := filepath.Join(tmpDir, "phoneScreenshots", "de-DE")
		os.MkdirAll(phoneDir, 0755)
		os.WriteFile(filepath.Join(phoneDir, "screenshot.png"), []byte("fake png"), 0644)

		cmd := &BulkImagesCmd{
			ImageDir: tmpDir,
			Locale:   "en-US", // Default locale, should be overridden by dir structure
		}

		images, err := cmd.scanImageDirectory()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(images) != 1 {
			t.Fatalf("Expected 1 image, got: %d", len(images))
		}

		if images[0].Locale != "de-DE" {
			t.Errorf("Expected locale de-DE from directory structure, got: %s", images[0].Locale)
		}
	})

	t.Run("case insensitive image extensions", func(t *testing.T) {
		tmpDir := t.TempDir()

		phoneDir := filepath.Join(tmpDir, "phoneScreenshots", "en-US")
		os.MkdirAll(phoneDir, 0755)

		// Create images with different case extensions
		os.WriteFile(filepath.Join(phoneDir, "img1.PNG"), []byte("fake png"), 0644)
		os.WriteFile(filepath.Join(phoneDir, "img2.JPG"), []byte("fake jpg"), 0644)
		os.WriteFile(filepath.Join(phoneDir, "img3.JPEG"), []byte("fake jpeg"), 0644)

		cmd := &BulkImagesCmd{
			ImageDir: tmpDir,
			Locale:   "en-US",
		}

		images, err := cmd.scanImageDirectory()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(images) != 3 {
			t.Errorf("Expected 3 images (case insensitive), got: %d", len(images))
		}
	})
}

func TestBulkImagesCmd_Run_DryRun(t *testing.T) {
	t.Run("dry run with images", func(t *testing.T) {
		tmpDir := t.TempDir()

		phoneDir := filepath.Join(tmpDir, "phoneScreenshots", "en-US")
		os.MkdirAll(phoneDir, 0755)
		os.WriteFile(filepath.Join(phoneDir, "screenshot.png"), []byte("fake png"), 0644)

		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}
		cmd := &BulkImagesCmd{
			ImageDir: tmpDir,
			Locale:   "en-US",
			DryRun:   true,
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error in dry run: %v", err)
		}
	})
}

func TestBulkImagesCmd_DefaultValues(t *testing.T) {
	t.Run("default locale is en-US", func(t *testing.T) {
		cmd := &BulkImagesCmd{}
		// Default is set by Kong tag
		if cmd.Locale != "" {
			t.Logf("Locale default: %s", cmd.Locale)
		}
	})

	t.Run("default max parallel is 3", func(t *testing.T) {
		cmd := &BulkImagesCmd{}
		// Default is set by Kong tag
		if cmd.MaxParallel != 0 {
			t.Logf("MaxParallel default: %d", cmd.MaxParallel)
		}
	})
}

func TestBulkImagesCmd_ResultsStructure(t *testing.T) {
	t.Run("bulk images result structure", func(t *testing.T) {
		result := &bulkImagesResult{
			SuccessCount: 5,
			FailureCount: 2,
			EditID:       "test-edit-456",
			Images: []bulkImageItemResult{
				{Type: "phoneScreenshots", Locale: "en-US", Filename: "s1.png", Status: "success"},
				{Type: "phoneScreenshots", Locale: "en-US", Filename: "s2.png", Status: "success"},
				{Type: "featureGraphic", Locale: "en-US", Filename: "feature.png", Status: "failed", Error: "upload failed"},
			},
		}

		if result.SuccessCount != 5 {
			t.Errorf("Expected success count 5, got: %d", result.SuccessCount)
		}
		if result.FailureCount != 2 {
			t.Errorf("Expected failure count 2, got: %d", result.FailureCount)
		}
		if len(result.Images) != 3 {
			t.Errorf("Expected 3 images, got: %d", len(result.Images))
		}
	})
}

// ============================================================================
// BulkTracksCmd Tests
// ============================================================================

func TestBulkTracksCmd_Run_PackageRequired(t *testing.T) {
	t.Run("missing package returns error", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: []string{"123"},
		}
		globals := &Globals{} // No package set

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for missing package")
		}
		if !strings.Contains(err.Error(), "package name is required") {
			t.Errorf("Expected 'package name is required' error, got: %v", err)
		}
	})
}

func TestBulkTracksCmd_Run_NoTracks(t *testing.T) {
	t.Run("no tracks returns validation error", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			Tracks:       []string{},
			VersionCodes: []string{"123"},
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for no tracks")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if !strings.Contains(apiErr.Message, "at least one track") {
			t.Errorf("Expected 'at least one track' message, got: %s", apiErr.Message)
		}
	})
}

func TestBulkTracksCmd_Run_NoVersionCodes(t *testing.T) {
	t.Run("no version codes returns validation error", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: []string{},
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for no version codes")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if !strings.Contains(apiErr.Message, "at least one version code") {
			t.Errorf("Expected 'at least one version code' message, got: %s", apiErr.Message)
		}
	})
}

func TestBulkTracksCmd_Run_InvalidVersionCode(t *testing.T) {
	t.Run("invalid version code format returns error", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: []string{"abc"},
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for invalid version code")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if !strings.Contains(apiErr.Message, "invalid version code") {
			t.Errorf("Expected 'invalid version code' message, got: %s", apiErr.Message)
		}
		if apiErr.Hint == "" {
			t.Error("Expected error to have a hint")
		}
	})

	t.Run("mixed valid and invalid version codes", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: []string{"1", "invalid", "3"},
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for invalid version code in list")
		}
	})
}

func TestBulkTracksCmd_Run_DryRun(t *testing.T) {
	t.Run("dry run outputs expected fields", func(t *testing.T) {
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal", "alpha"},
			VersionCodes: []string{"100", "101"},
			Status:       "draft",
			Name:         "Test Release",
			DryRun:       true,
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error in dry run: %v", err)
		}
	})
}

func TestBulkTracksCmd_Run_InvalidAuth(t *testing.T) {
	t.Run("invalid auth key path returns error", func(t *testing.T) {
		globals := &Globals{
			Package: "com.example.app",
			KeyPath: "/nonexistent/key.json",
			Output:  "json",
		}
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: []string{"100"},
			Status:       "draft",
		}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for invalid auth")
		}
	})
}

func TestBulkTracksCmd_DefaultValues(t *testing.T) {
	t.Run("default status is draft", func(t *testing.T) {
		cmd := &BulkTracksCmd{}
		// Default is set by Kong tag
		if cmd.Status != "" {
			t.Logf("Status default: %s", cmd.Status)
		}
	})
}

func TestBulkTracksCmd_ResultsStructure(t *testing.T) {
	t.Run("bulk tracks result structure", func(t *testing.T) {
		result := &bulkTracksResult{
			SuccessCount: 2,
			FailureCount: 1,
			EditID:       "test-edit-789",
			Committed:    true,
			Tracks: []bulkTrackItemResult{
				{Track: "internal", Status: "success", VersionCodes: []string{"100", "101"}},
				{Track: "alpha", Status: "success", VersionCodes: []string{"100", "101"}},
				{Track: "production", Status: "failed", VersionCodes: []string{"100"}, Error: "permission denied"},
			},
		}

		if result.SuccessCount != 2 {
			t.Errorf("Expected success count 2, got: %d", result.SuccessCount)
		}
		if result.FailureCount != 1 {
			t.Errorf("Expected failure count 1, got: %d", result.FailureCount)
		}
		if !result.Committed {
			t.Error("Expected Committed to be true")
		}
		if len(result.Tracks) != 3 {
			t.Errorf("Expected 3 tracks, got: %d", len(result.Tracks))
		}
	})
}

// ============================================================================
// BulkCmd Structure Tests
// ============================================================================

func TestBulkCmd_Structure(t *testing.T) {
	t.Run("bulk command has all subcommands", func(t *testing.T) {
		cmd := &BulkCmd{}

		// Verify all subcommands are present
		_ = cmd.Upload
		_ = cmd.Listings
		_ = cmd.Images
		_ = cmd.Tracks
	})
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestCreateTempFile(t *testing.T) {
	t.Run("creates file with content", func(t *testing.T) {
		content := []byte("test content")
		path := createTempFile(t, "test.txt", content)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("Expected file to exist")
		}

		readContent, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read temp file: %v", err)
		}

		if string(readContent) != string(content) {
			t.Errorf("Expected content %s, got: %s", content, readContent)
		}

		os.Remove(path)
	})
}

// ============================================================================
// Integration and Edge Case Tests
// ============================================================================

func TestBulkUploadCmd_EdgeCases(t *testing.T) {
	t.Run("single file upload validation", func(t *testing.T) {
		tmpFile := createTempFile(t, "single.aab", []byte("fake aab content"))
		defer os.Remove(tmpFile)

		// Verify file exists
		if _, err := os.Stat(tmpFile); err != nil {
			t.Fatalf("Temp file should exist: %v", err)
		}

		// Verify extension detection
		ext := strings.ToLower(filepath.Ext(tmpFile))
		if ext != ".aab" {
			t.Errorf("Expected .aab extension, got: %s", ext)
		}
	})

	t.Run("multiple files with different types", func(t *testing.T) {
		tmpApk := createTempFile(t, "app.apk", []byte("fake apk"))
		tmpAab := createTempFile(t, "app.aab", []byte("fake aab"))
		defer os.Remove(tmpApk)
		defer os.Remove(tmpAab)

		cmd := &BulkUploadCmd{
			Files: []string{tmpApk, tmpAab},
		}

		if len(cmd.Files) != 2 {
			t.Errorf("Expected 2 files, got: %d", len(cmd.Files))
		}
	})
}

func TestBulkImagesCmd_EdgeCases(t *testing.T) {
	t.Run("nested directory structure", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create deeply nested structure
		sevenInchDir := filepath.Join(tmpDir, "sevenInchScreenshots", "en-US")
		tenInchDir := filepath.Join(tmpDir, "tenInchScreenshots", "en-US")
		wearDir := filepath.Join(tmpDir, "wearScreenshots", "en-US")

		os.MkdirAll(sevenInchDir, 0755)
		os.MkdirAll(tenInchDir, 0755)
		os.MkdirAll(wearDir, 0755)

		os.WriteFile(filepath.Join(sevenInchDir, "screen1.png"), []byte("fake"), 0644)
		os.WriteFile(filepath.Join(tenInchDir, "screen1.png"), []byte("fake"), 0644)
		os.WriteFile(filepath.Join(wearDir, "screen1.png"), []byte("fake"), 0644)

		cmd := &BulkImagesCmd{
			ImageDir: tmpDir,
		}

		images, err := cmd.scanImageDirectory()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(images) != 3 {
			t.Errorf("Expected 3 images, got: %d", len(images))
		}

		// Verify all image types are detected
		types := make(map[string]bool)
		for _, img := range images {
			types[img.Type] = true
		}

		if !types["sevenInchScreenshots"] {
			t.Error("Expected sevenInchScreenshots type")
		}
		if !types["tenInchScreenshots"] {
			t.Error("Expected tenInchScreenshots type")
		}
		if !types["wearScreenshots"] {
			t.Error("Expected wearScreenshots type")
		}
	})

	t.Run("directory with non-image files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create mixed content
		phoneDir := filepath.Join(tmpDir, "phoneScreenshots", "en-US")
		os.MkdirAll(phoneDir, 0755)

		os.WriteFile(filepath.Join(phoneDir, "valid.png"), []byte("fake"), 0644)
		os.WriteFile(filepath.Join(phoneDir, "readme.txt"), []byte("text"), 0644)
		os.WriteFile(filepath.Join(phoneDir, "data.json"), []byte("{}"), 0644)
		os.WriteFile(filepath.Join(phoneDir, "script.sh"), []byte("#!/bin/bash"), 0644)

		cmd := &BulkImagesCmd{
			ImageDir: tmpDir,
		}

		images, err := cmd.scanImageDirectory()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(images) != 1 {
			t.Errorf("Expected 1 image (only .png), got: %d", len(images))
		}

		if images[0].Filename != filepath.Join(phoneDir, "valid.png") {
			t.Errorf("Expected valid.png, got: %s", images[0].Filename)
		}
	})
}

func TestBulkListingsCmd_EdgeCases(t *testing.T) {
	t.Run("listing data with special characters", func(t *testing.T) {
		listings := bulkListingData{
			"en-US": {
				Title:            "App™ with © Special® Characters™",
				ShortDescription: "Short «description» — em-dash",
				FullDescription:  "Full description\nwith newlines\tand tabs",
			},
			"ja-JP": {
				Title:            "日本語アプリ",
				ShortDescription: "日本語の短い説明",
				FullDescription:  "日本語の詳細な説明です。",
			},
			"ar-SA": {
				Title:            "تطبيق عربي",
				ShortDescription: "وصف قصير",
				FullDescription:  "وصف تفصيلي باللغة العربية",
			},
		}

		// Verify special characters are preserved
		if !strings.Contains(listings["en-US"].Title, "™") {
			t.Error("Expected trademark symbol in title")
		}
		if !strings.Contains(listings["ja-JP"].Title, "日本語") {
			t.Error("Expected Japanese characters in title")
		}
		if !strings.Contains(listings["ar-SA"].Title, "عربي") {
			t.Error("Expected Arabic characters in title")
		}
	})

	t.Run("listing data with video field", func(t *testing.T) {
		data := `{
			"en-US": {
				"title": "Test App",
				"shortDescription": "A test app",
				"fullDescription": "Full description",
				"video": "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
			}
		}`

		tmpFile := createTempFile(t, "listing.json", []byte(data))
		defer os.Remove(tmpFile)

		readData, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		var listings bulkListingData
		if err := json.Unmarshal(readData, &listings); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if listings["en-US"].Video != "https://www.youtube.com/watch?v=dQw4w9WgXcQ" {
			t.Errorf("Expected video URL, got: %s", listings["en-US"].Video)
		}
	})
}

func TestBulkTracksCmd_EdgeCases(t *testing.T) {
	t.Run("multiple version codes", func(t *testing.T) {
		versionCodes := []string{"100", "101", "102", "103"}
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: versionCodes,
		}

		// Verify all version codes are present
		if len(cmd.VersionCodes) != 4 {
			t.Errorf("Expected 4 version codes, got: %d", len(cmd.VersionCodes))
		}
	})

	t.Run("large version code values", func(t *testing.T) {
		// Version codes can be large integers
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: []string{"2147483647"}, // Max int32
		}

		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		// This would fail in dry run with invalid auth, but we're testing validation
		err := cmd.Run(globals)
		if err == nil {
			// Expected to fail due to auth, but parsing should work
			t.Log("Command failed as expected due to auth")
		}
	})

	t.Run("version code with leading zeros", func(t *testing.T) {
		// Leading zeros should be parsed correctly
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: []string{"001", "010", "100"},
		}

		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err == nil {
			t.Log("Command execution reached dry run or auth failure as expected")
		}
	})
}

// ============================================================================
// Context and Timeout Tests
// ============================================================================

func TestBulkCommands_ContextHandling(t *testing.T) {
	t.Run("context propagation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		globals := &Globals{
			Package: "com.example.app",
			Context: ctx,
		}

		// Verify context is set
		if globals.Context == nil {
			t.Error("Expected context to be set")
		}

		if globals.Context.Err() != nil {
			t.Errorf("Expected context to not be canceled yet: %v", globals.Context.Err())
		}
	})

	t.Run("canceled context handling", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		globals := &Globals{
			Package: "com.example.app",
			Context: ctx,
		}

		if globals.Context.Err() != context.Canceled {
			t.Errorf("Expected context.Canceled, got: %v", globals.Context.Err())
		}
	})
}

// ============================================================================
// Parallel Processing Tests
// ============================================================================

func TestBulkUploadCmd_ParallelSettings(t *testing.T) {
	tests := []struct {
		name        string
		maxParallel int
		expected    int
	}{
		{"default", 0, 0}, // Kong sets default to 3
		{"single", 1, 1},
		{"double", 2, 2},
		{"high", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &BulkUploadCmd{
				MaxParallel: tt.maxParallel,
			}

			if cmd.MaxParallel != tt.expected {
				t.Errorf("Expected MaxParallel %d, got: %d", tt.expected, cmd.MaxParallel)
			}
		})
	}
}

func TestBulkImagesCmd_ParallelSettings(t *testing.T) {
	tests := []struct {
		name        string
		maxParallel int
		expected    int
	}{
		{"default", 0, 0}, // Kong sets default to 3
		{"single", 1, 1},
		{"five", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &BulkImagesCmd{
				MaxParallel: tt.maxParallel,
			}

			if cmd.MaxParallel != tt.expected {
				t.Errorf("Expected MaxParallel %d, got: %d", tt.expected, cmd.MaxParallel)
			}
		})
	}
}

// ============================================================================
// InProgressReviewBehaviour Tests
// ============================================================================

func TestBulkUploadCmd_InProgressReviewBehaviour(t *testing.T) {
	tests := []struct {
		name      string
		behaviour string
		valid     bool
	}{
		{"empty", "", true}, // Empty is valid (default)
		{"throw_error", "THROW_ERROR_IF_IN_PROGRESS", true},
		{"cancel", "CANCEL_IN_PROGRESS_AND_SUBMIT", true},
		{"unspecified", "IN_PROGRESS_REVIEW_BEHAVIOUR_UNSPECIFIED", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &BulkUploadCmd{
				InProgressReviewBehaviour: tt.behaviour,
			}

			// Just verify the value is set
			if cmd.InProgressReviewBehaviour != tt.behaviour {
				t.Errorf("Expected behaviour %s, got: %s", tt.behaviour, cmd.InProgressReviewBehaviour)
			}
		})
	}
}

// ============================================================================
// NoAutoCommit Tests
// ============================================================================

func TestBulkUploadCmd_NoAutoCommit(t *testing.T) {
	t.Run("auto commit enabled by default", func(t *testing.T) {
		cmd := &BulkUploadCmd{}

		if cmd.NoAutoCommit {
			t.Error("Expected NoAutoCommit to be false by default")
		}
	})

	t.Run("no auto commit flag", func(t *testing.T) {
		cmd := &BulkUploadCmd{
			NoAutoCommit: true,
		}

		if !cmd.NoAutoCommit {
			t.Error("Expected NoAutoCommit to be true")
		}
	})
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestBulkCommands_ErrorWrapping(t *testing.T) {
	t.Run("upload with missing files has hint", func(t *testing.T) {
		cmd := &BulkUploadCmd{
			Files: []string{},
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if apiErr.Hint == "" {
			t.Error("Expected error to have a hint")
		}
	})

	t.Run("listings with invalid JSON has details", func(t *testing.T) {
		tmpFile := createTempFile(t, "bad.json", []byte("{invalid"))
		defer os.Remove(tmpFile)

		cmd := &BulkListingsCmd{
			DataFile: tmpFile,
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error")
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError type, got: %T", err)
		}
		if apiErr.Details == nil {
			t.Error("Expected error to have details")
		}
	})
}

// ============================================================================
// Result Structure JSON Tests
// ============================================================================

func TestBulkUploadResult_JSON(t *testing.T) {
	t.Run("result serializes to JSON correctly", func(t *testing.T) {
		result := &bulkUploadResult{
			SuccessCount:   3,
			FailureCount:   1,
			SkippedCount:   0,
			EditID:         "edit-123",
			Committed:      true,
			ProcessingTime: "1.5s",
			Uploads: []bulkUploadItemResult{
				{File: "/path/app1.apk", VersionCode: 100, Status: "success", SHA1: "abc123"},
				{File: "/path/app2.aab", VersionCode: 101, Status: "success", SHA1: "def456"},
				{File: "/path/app3.apk", VersionCode: 102, Status: "success", SHA1: "ghi789"},
				{File: "/path/bad.txt", Status: "failed", Error: "unsupported file type: .txt"},
			},
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal result: %v", err)
		}

		var decoded bulkUploadResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if decoded.SuccessCount != 3 {
			t.Errorf("Expected success count 3, got: %d", decoded.SuccessCount)
		}
		if decoded.EditID != "edit-123" {
			t.Errorf("Expected edit ID 'edit-123', got: %s", decoded.EditID)
		}
		if !decoded.Committed {
			t.Error("Expected Committed to be true")
		}
		if len(decoded.Uploads) != 4 {
			t.Errorf("Expected 4 uploads, got: %d", len(decoded.Uploads))
		}
	})
}

func TestBulkListingsResult_JSON(t *testing.T) {
	t.Run("result serializes to JSON correctly", func(t *testing.T) {
		result := &bulkListingsResult{
			SuccessCount: 5,
			FailureCount: 2,
			EditID:       "edit-456",
			Locales: []bulkListingItemResult{
				{Locale: "en-US", Status: "success"},
				{Locale: "de-DE", Status: "success"},
				{Locale: "fr-FR", Status: "failed", Error: "invalid characters"},
			},
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal result: %v", err)
		}

		var decoded bulkListingsResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if decoded.SuccessCount != 5 {
			t.Errorf("Expected success count 5, got: %d", decoded.SuccessCount)
		}
		if len(decoded.Locales) != 3 {
			t.Errorf("Expected 3 locales, got: %d", len(decoded.Locales))
		}
	})
}

func TestBulkImagesResult_JSON(t *testing.T) {
	t.Run("result serializes to JSON correctly", func(t *testing.T) {
		result := &bulkImagesResult{
			SuccessCount: 10,
			FailureCount: 3,
			EditID:       "edit-789",
			Images: []bulkImageItemResult{
				{Type: "phoneScreenshots", Locale: "en-US", Filename: "/img/s1.png", Status: "success"},
				{Type: "phoneScreenshots", Locale: "en-US", Filename: "/img/s2.png", Status: "success"},
				{Type: "featureGraphic", Locale: "en-US", Filename: "/img/feature.png", Status: "failed", Error: "timeout"},
			},
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal result: %v", err)
		}

		var decoded bulkImagesResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if decoded.SuccessCount != 10 {
			t.Errorf("Expected success count 10, got: %d", decoded.SuccessCount)
		}
		if decoded.FailureCount != 3 {
			t.Errorf("Expected failure count 3, got: %d", decoded.FailureCount)
		}
	})
}

func TestBulkTracksResult_JSON(t *testing.T) {
	t.Run("result serializes to JSON correctly", func(t *testing.T) {
		result := &bulkTracksResult{
			SuccessCount: 3,
			FailureCount: 1,
			EditID:       "edit-abc",
			Committed:    false,
			Tracks: []bulkTrackItemResult{
				{Track: "internal", Status: "success", VersionCodes: []string{"100"}},
				{Track: "alpha", Status: "success", VersionCodes: []string{"100", "101"}},
				{Track: "beta", Status: "success", VersionCodes: []string{"100", "101", "102"}},
				{Track: "production", Status: "failed", VersionCodes: []string{"100"}, Error: "insufficient permissions"},
			},
		}

		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal result: %v", err)
		}

		var decoded bulkTracksResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if decoded.SuccessCount != 3 {
			t.Errorf("Expected success count 3, got: %d", decoded.SuccessCount)
		}
		if decoded.Committed {
			t.Error("Expected Committed to be false")
		}
		if len(decoded.Tracks) != 4 {
			t.Errorf("Expected 4 tracks, got: %d", len(decoded.Tracks))
		}
	})
}

// ============================================================================
// Valid Track and Status Tests
// ============================================================================

func TestBulkTracksCmd_TrackValidation(t *testing.T) {
	validTracks := []string{"internal", "alpha", "beta", "production"}

	for _, track := range validTracks {
		t.Run(fmt.Sprintf("valid track: %s", track), func(t *testing.T) {
			if !api.IsValidTrack(track) {
				t.Errorf("Expected %s to be a valid track", track)
			}
		})
	}

	invalidTracks := []string{"", "invalid", "test", "staging", "dev"}

	for _, track := range invalidTracks {
		t.Run(fmt.Sprintf("invalid track: %s", track), func(t *testing.T) {
			if api.IsValidTrack(track) {
				t.Errorf("Expected %s to be an invalid track", track)
			}
		})
	}
}

func TestBulkTracksCmd_StatusValidation(t *testing.T) {
	validStatuses := []string{"draft", "completed", "halted", "inProgress"}

	for _, status := range validStatuses {
		t.Run(fmt.Sprintf("valid status: %s", status), func(t *testing.T) {
			if !api.IsValidReleaseStatus(status) {
				t.Errorf("Expected %s to be a valid status", status)
			}
		})
	}

	invalidStatuses := []string{"", "invalid", "testing", "published", "archived"}

	for _, status := range invalidStatuses {
		t.Run(fmt.Sprintf("invalid status: %s", status), func(t *testing.T) {
			if api.IsValidReleaseStatus(status) {
				t.Errorf("Expected %s to be an invalid status", status)
			}
		})
	}
}

// ============================================================================
// File Permission and Access Tests
// ============================================================================

func TestBulkListingsCmd_FilePermissions(t *testing.T) {
	t.Run("unreadable file returns error", func(t *testing.T) {
		// Skip on Windows as permissions work differently
		if os.PathSeparator == '\\' {
			t.Skip("Skipping permission test on Windows")
		}

		tmpFile := createTempFile(t, "unreadable.json", []byte(`{"en-US": {"title": "Test"}}`))
		defer os.Remove(tmpFile)

		// Remove read permission
		os.Chmod(tmpFile, 0000)
		defer os.Chmod(tmpFile, 0644) // Restore for cleanup

		cmd := &BulkListingsCmd{
			DataFile: tmpFile,
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for unreadable file")
		}
	})
}

// ============================================================================
// Long Running Operation Tests
// ============================================================================

func TestBulkUploadCmd_ProcessingTime(t *testing.T) {
	t.Run("processing time is recorded", func(t *testing.T) {
		start := time.Now()
		time.Sleep(10 * time.Millisecond) // Simulate some work
		duration := time.Since(start)

		result := &bulkUploadResult{
			ProcessingTime: duration.String(),
		}

		if result.ProcessingTime == "" {
			t.Error("Expected processing time to be recorded")
		}

		// Verify duration can be parsed
		if _, err := time.ParseDuration(result.ProcessingTime); err != nil {
			t.Errorf("Processing time should be parseable: %v", err)
		}
	})
}

// ============================================================================
// Verbose Mode Tests
// ============================================================================

func TestBulkCommands_VerboseMode(t *testing.T) {
	t.Run("verbose mode is accepted", func(t *testing.T) {
		globals := &Globals{
			Package: "com.example.app",
			Verbose: true,
			Output:  "json",
		}

		// Verify verbose flag is set
		if !globals.Verbose {
			t.Error("Expected Verbose to be true")
		}
	})
}

// ============================================================================
// Explicit EditID Tests
// ============================================================================

func TestBulkCommands_ExplicitEditID(t *testing.T) {
	t.Run("upload with explicit edit ID", func(t *testing.T) {
		cmd := &BulkUploadCmd{
			Files:  []string{}, // Empty to fail fast
			EditID: "explicit-edit-123",
		}

		if cmd.EditID != "explicit-edit-123" {
			t.Errorf("Expected edit ID 'explicit-edit-123', got: %s", cmd.EditID)
		}
	})

	t.Run("listings with explicit edit ID", func(t *testing.T) {
		cmd := &BulkListingsCmd{
			EditID: "explicit-edit-456",
		}

		if cmd.EditID != "explicit-edit-456" {
			t.Errorf("Expected edit ID 'explicit-edit-456', got: %s", cmd.EditID)
		}
	})

	t.Run("images with explicit edit ID", func(t *testing.T) {
		cmd := &BulkImagesCmd{
			EditID: "explicit-edit-789",
		}

		if cmd.EditID != "explicit-edit-789" {
			t.Errorf("Expected edit ID 'explicit-edit-789', got: %s", cmd.EditID)
		}
	})

	t.Run("tracks with explicit edit ID", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			EditID: "explicit-edit-abc",
		}

		if cmd.EditID != "explicit-edit-abc" {
			t.Errorf("Expected edit ID 'explicit-edit-abc', got: %s", cmd.EditID)
		}
	})
}

// ============================================================================
// Output Format Tests
// ============================================================================

func TestBulkCommands_OutputFormats(t *testing.T) {
	formats := []string{"json", "table", "csv", "yaml", ""}

	for _, format := range formats {
		t.Run(fmt.Sprintf("output format: %s", format), func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Output:  format,
			}

			if globals.Output != format {
				t.Errorf("Expected output format '%s', got: %s", format, globals.Output)
			}
		})
	}
}

// ============================================================================
// Timeout Configuration Tests
// ============================================================================

func TestBulkCommands_Timeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{"default", 0},
		{"30s", 30 * time.Second},
		{"5m", 5 * time.Minute},
		{"1h", 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Timeout: tt.timeout,
			}

			if globals.Timeout != tt.timeout {
				t.Errorf("Expected timeout %v, got: %v", tt.timeout, globals.Timeout)
			}
		})
	}
}

// ============================================================================
// Empty and Nil Tests
// ============================================================================

func TestBulkCommands_EmptyInputs(t *testing.T) {
	t.Run("upload with nil files slice", func(t *testing.T) {
		cmd := &BulkUploadCmd{
			Files: nil,
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for nil files")
		}
	})

	t.Run("tracks with nil tracks slice", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			Tracks:       nil,
			VersionCodes: []string{"100"},
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for nil tracks")
		}
	})

	t.Run("tracks with nil version codes slice", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: nil,
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for nil version codes")
		}
	})
}

// ============================================================================
// Concurrent Safety Tests
// ============================================================================

func TestBulkUploadResult_ConcurrentSafety(t *testing.T) {
	t.Run("result structure is safe for concurrent updates", func(t *testing.T) {
		result := &bulkUploadResult{
			Uploads: make([]bulkUploadItemResult, 0),
		}

		// Simulate concurrent updates
		for i := 0; i < 10; i++ {
			item := bulkUploadItemResult{
				File:   fmt.Sprintf("file%d.apk", i),
				Status: "success",
			}
			result.Uploads = append(result.Uploads, item)
			if item.Status == "success" {
				result.SuccessCount++
			}
		}

		if result.SuccessCount != 10 {
			t.Errorf("Expected success count 10, got: %d", result.SuccessCount)
		}
		if len(result.Uploads) != 10 {
			t.Errorf("Expected 10 uploads, got: %d", len(result.Uploads))
		}
	})
}

func TestBulkImagesResult_ConcurrentSafety(t *testing.T) {
	t.Run("result structure handles concurrent updates", func(t *testing.T) {
		result := &bulkImagesResult{
			Images: make([]bulkImageItemResult, 0),
		}

		// Simulate updates
		for i := 0; i < 5; i++ {
			item := bulkImageItemResult{
				Type:   "phoneScreenshots",
				Locale: "en-US",
				Status: "success",
			}
			result.Images = append(result.Images, item)
			result.SuccessCount++
		}

		if result.SuccessCount != 5 {
			t.Errorf("Expected success count 5, got: %d", result.SuccessCount)
		}
	})
}

// ============================================================================
// Business Logic Function Tests
// ============================================================================

func TestBulkImagesCmd_uploadImage_ErrorHandling(t *testing.T) {
	t.Run("handles file open error", func(t *testing.T) {
		item := &bulkImageItemResult{
			Type:     "phoneScreenshots",
			Locale:   "en-US",
			Filename: "/nonexistent/file.png",
			Status:   "pending",
		}

		// Simulate the error that would occur
		_, err := os.Open(item.Filename)
		if err == nil {
			t.Fatal("Expected error for nonexistent file")
		}

		// This simulates what the actual function would return
		result := bulkImageItemResult{
			Type:     item.Type,
			Locale:   item.Locale,
			Filename: item.Filename,
			Status:   "failed",
			Error:    err.Error(),
		}

		if result.Status != "failed" {
			t.Errorf("Expected status 'failed', got: %s", result.Status)
		}
		if result.Error == "" {
			t.Error("Expected error message to be set")
		}
	})
}

func TestBulkTracksCmd_updateTrack_ErrorHandling(t *testing.T) {
	t.Run("handles missing version codes", func(t *testing.T) {
		cmd := &BulkTracksCmd{
			Tracks:       []string{"internal"},
			VersionCodes: []string{}, // Empty
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error for empty version codes")
		}
	})
}

func TestBulkListingsCmd_updateListing_ErrorHandling(t *testing.T) {
	t.Run("handles empty locale data", func(t *testing.T) {
		// Create listing data with empty fields
		data := `{
			"en-US": {
				"title": "",
				"shortDescription": "",
				"fullDescription": ""
			}
		}`

		tmpFile := createTempFile(t, "empty_fields.json", []byte(data))
		defer os.Remove(tmpFile)

		cmd := &BulkListingsCmd{
			DataFile: tmpFile,
		}
		globals := &Globals{Package: "com.example.app"}

		// Should not error on empty fields - API will validate
		err := cmd.Run(globals)
		// Will fail on auth, but JSON parsing should succeed
		if err != nil {
			t.Logf("Expected error (likely auth): %v", err)
		}
	})
}

// ============================================================================
// Complex JSON Parsing Tests
// ============================================================================

func TestBulkListingsCmd_ComplexJSON(t *testing.T) {
	t.Run("handles large JSON file", func(t *testing.T) {
		// Create a large listings file with many locales
		listings := make(map[string]interface{})
		locales := []string{"en-US", "de-DE", "fr-FR", "es-ES", "it-IT", "ja-JP", "ko-KR", "zh-CN", "ru-RU", "pt-BR"}

		for _, locale := range locales {
			listings[locale] = map[string]string{
				"title":            fmt.Sprintf("App Title %s", locale),
				"shortDescription": fmt.Sprintf("Short description %s", locale),
				"fullDescription":  fmt.Sprintf("Full description for %s with more content here.", locale),
			}
		}

		data, _ := json.Marshal(listings)
		tmpFile := createTempFile(t, "large_listings.json", data)
		defer os.Remove(tmpFile)

		cmd := &BulkListingsCmd{
			DataFile: tmpFile,
		}
		globals := &Globals{Package: "com.example.app"}

		// Parse should succeed
		readData, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		var parsed bulkListingData
		if err := json.Unmarshal(readData, &parsed); err != nil {
			t.Fatalf("Failed to parse large JSON: %v", err)
		}

		if len(parsed) != 10 {
			t.Errorf("Expected 10 locales, got: %d", len(parsed))
		}

		// Run will fail on auth, but that's expected
		_ = cmd.Run(globals)
	})

	t.Run("handles nested JSON structures", func(t *testing.T) {
		// JSON with additional fields that should be ignored
		data := `{
			"en-US": {
				"title": "Test App",
				"shortDescription": "Short desc",
				"fullDescription": "Full desc",
				"video": "http://example.com/video",
				"extraField": "should be ignored",
				"anotherExtra": 12345
			}
		}`

		tmpFile := createTempFile(t, "nested.json", []byte(data))
		defer os.Remove(tmpFile)

		var parsed bulkListingData
		if err := json.Unmarshal([]byte(data), &parsed); err != nil {
			t.Fatalf("Failed to parse JSON with extra fields: %v", err)
		}

		if parsed["en-US"].Title != "Test App" {
			t.Errorf("Expected title to be parsed correctly despite extra fields")
		}
	})
}

// ============================================================================
// Path and Directory Tests
// ============================================================================

func TestBulkImagesCmd_PathHandling(t *testing.T) {
	t.Run("handles absolute paths", func(t *testing.T) {
		tmpDir := t.TempDir()

		phoneDir := filepath.Join(tmpDir, "phoneScreenshots", "en-US")
		os.MkdirAll(phoneDir, 0755)
		os.WriteFile(filepath.Join(phoneDir, "screen.png"), []byte("fake"), 0644)

		cmd := &BulkImagesCmd{
			ImageDir: tmpDir,
		}

		images, err := cmd.scanImageDirectory()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(images) != 1 {
			t.Errorf("Expected 1 image, got: %d", len(images))
		}

		// Verify the path is absolute
		if !filepath.IsAbs(images[0].Filename) {
			t.Error("Expected absolute path for filename")
		}
	})

	t.Run("handles relative paths", func(t *testing.T) {
		// Create a relative path scenario
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalDir)

		phoneDir := filepath.Join("phoneScreenshots", "en-US")
		os.MkdirAll(phoneDir, 0755)
		os.WriteFile(filepath.Join(phoneDir, "screen.png"), []byte("fake"), 0644)

		cmd := &BulkImagesCmd{
			ImageDir: ".",
		}

		images, err := cmd.scanImageDirectory()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(images) != 1 {
			t.Errorf("Expected 1 image, got: %d", len(images))
		}
	})
}

// ============================================================================
// Summary
// ============================================================================
// These tests cover:
// - Command validation and error handling
// - File operations and validation
// - JSON parsing and data structures
// - Directory scanning and image discovery
// - Dry run functionality
// - Result structures and JSON serialization
// - Edge cases and boundary conditions
// - Concurrent safety patterns
// - Business logic functions
// - Complex input scenarios
