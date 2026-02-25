//go:build unit
// +build unit

package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoldenFile_Compare(t *testing.T) {
	t.Parallel()

	// Create temp directory for test
	tmpDir := t.TempDir()
	goldenDir := filepath.Join(tmpDir, "testdata", "golden")
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("Failed to create golden dir: %v", err)
	}

	t.Run("compare matching content", func(t *testing.T) {
		t.Parallel()

		// Create golden file
		goldenPath := filepath.Join(goldenDir, "test1.txt")
		content := []byte("hello world")
		if err := os.WriteFile(goldenPath, content, 0644); err != nil {
			t.Fatalf("Failed to write golden file: %v", err)
		}

		gf := NewGolden(t, "test1.txt").WithDir(goldenDir)
		gf.update = false // Disable update mode

		if err := gf.Compare(content); err != nil {
			t.Errorf("Expected no error for matching content, got: %v", err)
		}
	})

	t.Run("compare mismatching content", func(t *testing.T) {
		t.Parallel()

		// Create golden file
		goldenPath := filepath.Join(goldenDir, "test2.txt")
		if err := os.WriteFile(goldenPath, []byte("expected"), 0644); err != nil {
			t.Fatalf("Failed to write golden file: %v", err)
		}

		gf := NewGolden(t, "test2.txt").WithDir(goldenDir)
		gf.update = false

		err := gf.Compare([]byte("actual"))
		if err == nil {
			t.Error("Expected error for mismatching content")
		}

		if _, ok := err.(*GoldenMismatch); !ok {
			t.Errorf("Expected GoldenMismatch error, got: %T", err)
		}
	})

	t.Run("compare non-existent golden file", func(t *testing.T) {
		t.Parallel()

		gf := NewGolden(t, "nonexistent.txt").WithDir(goldenDir)
		gf.update = false

		err := gf.Compare([]byte("content"))
		if err == nil {
			t.Error("Expected error for non-existent golden file")
		}

		if !os.IsNotExist(err) {
			// The error should mention that the file doesn't exist
			if !contains(err.Error(), "not found") {
				t.Errorf("Expected 'not found' error, got: %v", err)
			}
		}
	})
}

func TestGoldenFile_CompareString(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goldenDir := filepath.Join(tmpDir, "testdata", "golden")
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("Failed to create golden dir: %v", err)
	}

	// Create golden file
	goldenPath := filepath.Join(goldenDir, "string.txt")
	if err := os.WriteFile(goldenPath, []byte("test string"), 0644); err != nil {
		t.Fatalf("Failed to write golden file: %v", err)
	}

	gf := NewGolden(t, "string.txt").WithDir(goldenDir)
	gf.update = false

	if err := gf.CompareString("test string"); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestGoldenMismatch(t *testing.T) {
	t.Parallel()

	t.Run("error message", func(t *testing.T) {
		t.Parallel()

		m := &GoldenMismatch{
			GoldenFile: "test.golden",
			Expected:   "line1\nline2",
			Actual:     "line1\nline3",
		}

		err := m.Error()
		if !contains(err, "test.golden") {
			t.Error("Error should mention golden file path")
		}
	})

	t.Run("diff output", func(t *testing.T) {
		t.Parallel()

		m := &GoldenMismatch{
			GoldenFile: "test.golden",
			Expected:   "line1\nline2",
			Actual:     "line1\nline3",
		}

		diff := m.Diff()
		if !contains(diff, "Line 2:") {
			t.Error("Diff should show line numbers")
		}
		if !contains(diff, "line2") {
			t.Error("Diff should show expected content")
		}
		if !contains(diff, "line3") {
			t.Error("Diff should show actual content")
		}
	})
}

func TestGoldenFile_Read(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goldenDir := filepath.Join(tmpDir, "testdata", "golden")
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("Failed to create golden dir: %v", err)
	}

	// Create golden file
	goldenPath := filepath.Join(goldenDir, "read.txt")
	content := []byte("readable content")
	if err := os.WriteFile(goldenPath, content, 0644); err != nil {
		t.Fatalf("Failed to write golden file: %v", err)
	}

	gf := NewGolden(t, "read.txt").WithDir(goldenDir)

	read, err := gf.Read()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if string(read) != string(content) {
		t.Errorf("Read content mismatch: got %q, want %q", string(read), string(content))
	}

	readStr, err := gf.ReadString()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if readStr != string(content) {
		t.Errorf("ReadString mismatch: got %q, want %q", readStr, string(content))
	}
}

func TestGoldenFile_Exists(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	goldenDir := filepath.Join(tmpDir, "testdata", "golden")
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		t.Fatalf("Failed to create golden dir: %v", err)
	}

	// Create golden file
	goldenPath := filepath.Join(goldenDir, "exists.txt")
	if err := os.WriteFile(goldenPath, []byte("exists"), 0644); err != nil {
		t.Fatalf("Failed to write golden file: %v", err)
	}

	gf := NewGolden(t, "exists.txt").WithDir(goldenDir)
	if !gf.Exists() {
		t.Error("Exists() should return true for existing file")
	}

	gf2 := NewGolden(t, "notexists.txt").WithDir(goldenDir)
	if gf2.Exists() {
		t.Error("Exists() should return false for non-existing file")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
