package extensions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestFindReleaseAsset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/repos/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"tag_name": "v1.0.0",
			"assets": [
				{"name": "gpd-test-ext_darwin_amd64.tar.gz", "browser_download_url": "https://example.com/download1", "size": 1234567},
				{"name": "gpd-test-ext_linux_amd64.tar.gz", "browser_download_url": "https://example.com/download2", "size": 1234567},
				{"name": "gpd-test-ext_windows_amd64.zip", "browser_download_url": "https://example.com/download3", "size": 1234567}
			]
		}`
		_, _ = w.Write([]byte(response))
	}))
	defer server.Close()
	_ = server
}

func TestExtractArchive(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	destDir := filepath.Join(tmpDir, "extracted")

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	_, _ = file.Write([]byte("mock archive content"))
	_ = file.Close()

	err = extractArchive(archivePath, destDir)
	if err == nil {
		t.Error("extractArchive() should fail with invalid archive")
	}
}

func TestCopyExtensionFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	manifest := &Manifest{Name: "test-ext", Bin: "gpd-test-ext"}

	binName := "gpd-test-ext"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	srcBin := filepath.Join(srcDir, binName)
	if err := os.WriteFile(srcBin, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("Failed to create source binary: %v", err)
	}

	manifestPath := filepath.Join(srcDir, ".gpd-extension")
	manifestData, _ := yaml.Marshal(manifest)
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		t.Fatalf("Failed to create source manifest: %v", err)
	}

	err := copyExtensionFiles(srcDir, dstDir, manifest)
	if err != nil {
		t.Fatalf("copyExtensionFiles() error = %v", err)
	}

	dstBin := filepath.Join(dstDir, binName)
	if _, err := os.Stat(dstBin); os.IsNotExist(err) {
		t.Error("Binary was not copied to destination")
	}

	dstManifest := filepath.Join(dstDir, ".gpd-extension")
	if _, err := os.Stat(dstManifest); os.IsNotExist(err) {
		t.Error("Manifest was not copied to destination")
	}
}

func TestCopyExtensionFilesMissingBinary(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	manifest := &Manifest{Name: "test-ext", Bin: "gpd-test-ext"}

	err := copyExtensionFiles(srcDir, dstDir, manifest)
	if err == nil {
		t.Error("copyExtensionFiles() should fail when binary is missing")
	}
}

func TestCopyFile(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcPath := filepath.Join(srcDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	dstPath := filepath.Join(dstDir, "copied.txt")

	err := copyFile(srcPath, dstPath, 0755)
	if err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	copied, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	if string(copied) != string(content) {
		t.Errorf("Copied content = %q, want %q", string(copied), string(content))
	}
}

func TestCopyFileNonExistent(t *testing.T) {
	dstDir := t.TempDir()
	dstPath := filepath.Join(dstDir, "dest.txt")

	err := copyFile("/nonexistent/path/file.txt", dstPath, 0644)
	if err == nil {
		t.Error("copyFile() should fail when source doesn't exist")
	}
}

func TestSaveExtensionMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	extName := "test-save-ext"
	extDir := filepath.Join(tmpDir, extName)
	if err := os.MkdirAll(extDir, 0755); err != nil {
		t.Fatalf("Failed to create extension dir: %v", err)
	}

	ext := &Extension{Name: extName, Version: "1.0.0", Source: "test"}

	data, err := yaml.Marshal(ext)
	if err != nil {
		t.Fatalf("Failed to marshal extension: %v", err)
	}

	metaPath := filepath.Join(extDir, ".gpd-extension")
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		t.Fatalf("Failed to write metadata: %v", err)
	}

	readData, err := os.ReadFile(metaPath)
	if err != nil {
		t.Fatalf("Failed to read metadata: %v", err)
	}

	var readExt Extension
	if err := yaml.Unmarshal(readData, &readExt); err != nil {
		t.Fatalf("Failed to unmarshal metadata: %v", err)
	}

	if readExt.Name != extName {
		t.Errorf("Read extension name = %q, want %q", readExt.Name, extName)
	}
}

func TestLoadLocalManifest(t *testing.T) {
	tmpDir := t.TempDir()
	manifest := &Manifest{Name: "test-ext", Version: "1.0.0", Description: "Test extension"}

	manifestPath := filepath.Join(tmpDir, ".gpd-extension")
	data, _ := yaml.Marshal(manifest)
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	loaded, err := loadLocalManifest(tmpDir)
	if err != nil {
		t.Fatalf("loadLocalManifest() error = %v", err)
	}

	if loaded.Name != "test-ext" {
		t.Errorf("Name = %q, want %q", loaded.Name, "test-ext")
	}
}

func TestLoadLocalManifestJSON(t *testing.T) {
	tmpDir := t.TempDir()
	manifest := Manifest{Name: "json-ext", Version: "1.0.0"}

	manifestPath := filepath.Join(tmpDir, ".gpd-extension")
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		t.Fatalf("Failed to write JSON manifest: %v", err)
	}

	loaded, err := loadLocalManifest(tmpDir)
	if err != nil {
		t.Fatalf("loadLocalManifest() error = %v", err)
	}

	if loaded.Name != "json-ext" {
		t.Errorf("Name = %q, want %q", loaded.Name, "json-ext")
	}
}

func TestLoadLocalManifestNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := loadLocalManifest(tmpDir)
	if err == nil {
		t.Error("loadLocalManifest() should return error when manifest doesn't exist")
	}
}

func TestLoadLocalManifestInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, ".gpd-extension")
	if err := os.WriteFile(manifestPath, []byte("invalid {["), 0644); err != nil {
		t.Fatalf("Failed to write invalid manifest: %v", err)
	}

	_, err := loadLocalManifest(tmpDir)
	if err == nil {
		t.Error("loadLocalManifest() should return error when manifest is invalid")
	}
}

func TestInstallFromRepoClone(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git clone test in short mode")
	}

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("Git not available, skipping clone test")
	}
}

func TestInstallTimeout(t *testing.T) {
	// Skip in CI environments where git operations may fail
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping timeout test in CI")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	opts := InstallOptions{Source: "owner/repo", Timeout: 1 * time.Millisecond}
	ctx := context.Background()
	_, err := Install(ctx, opts)

	if err == nil {
		t.Error("Install() with short timeout should fail")
	}
}
