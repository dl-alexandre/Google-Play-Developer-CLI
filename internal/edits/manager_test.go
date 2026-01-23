package edits

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIdempotencyKey(t *testing.T) {
	// Same inputs should produce same key
	key1 := IdempotencyKey("upload", "com.example.app", "abc123")
	key2 := IdempotencyKey("upload", "com.example.app", "abc123")

	if key1 != key2 {
		t.Errorf("Same inputs produced different keys: %s vs %s", key1, key2)
	}

	// Different inputs should produce different keys
	key3 := IdempotencyKey("upload", "com.example.app", "xyz789")
	if key1 == key3 {
		t.Error("Different inputs produced same key")
	}

	// Key should be fixed length
	if len(key1) != 16 {
		t.Errorf("IdempotencyKey length = %d, want 16", len(key1))
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := FormatBytes(tt.bytes); got != tt.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestHashFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// Hash should be consistent
	hash2, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile() second call error = %v", err)
	}

	if hash != hash2 {
		t.Error("Same file produced different hashes")
	}

	// Hash should be hex string
	if len(hash) != 64 { // SHA256 = 32 bytes = 64 hex chars
		t.Errorf("Hash length = %d, want 64", len(hash))
	}

	// Non-existent file should error
	_, err = HashFile(filepath.Join(tmpDir, "nonexistent.txt"))
	if err == nil {
		t.Error("HashFile() should error for non-existent file")
	}
}

func TestIdempotencyStore(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir}

	key := "test-key-123"

	// Initially should not exist
	exists, err := store.Check(key)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if exists {
		t.Error("Check() = true for new key, want false")
	}

	// Record the operation
	data := map[string]string{"result": "success"}
	if err := store.Record(key, data); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	// Now should exist
	exists, err = store.Check(key)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !exists {
		t.Error("Check() = false after Record(), want true")
	}

	// Get should return the data
	result, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !result.Found {
		t.Error("Get().Found = false, want true")
	}

	// Clear should remove it
	if err := store.Clear(key); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	exists, err = store.Check(key)
	if err != nil {
		t.Fatalf("Check() after Clear() error = %v", err)
	}
	if exists {
		t.Error("Check() = true after Clear(), want false")
	}
}

func TestEditState(t *testing.T) {
	// Test edit states
	states := []EditState{StateOpen, StateCommitted, StateAborted}
	expected := []string{"open", "committed", "aborted"}

	for i, state := range states {
		if string(state) != expected[i] {
			t.Errorf("EditState = %q, want %q", state, expected[i])
		}
	}
}

func TestLockFileDetection(t *testing.T) {
	// Test process alive detection (best effort)
	// Note: isProcessAlive is best-effort and may not work on all systems
	// We test it doesn't panic, but don't rely on specific results

	// Current process - just ensure no panic
	_ = isProcessAlive(os.Getpid())

	// Invalid PID - just ensure no panic
	_ = isProcessAlive(999999999)

	// Test that the function exists and is callable
	t.Log("isProcessAlive function exists and is callable")
}

func TestCacheEntryExpiry(t *testing.T) {
	// Test cache TTL constant
	if cacheTTL.Hours() != 24 {
		t.Errorf("cacheTTL = %v, want 24h", cacheTTL)
	}
}

func TestLockTimeouts(t *testing.T) {
	// Test lock timeout constants
	if lockTimeout.Seconds() != 30 {
		t.Errorf("lockTimeout = %v, want 30s", lockTimeout)
	}
	if staleLockAge.Hours() != 4 {
		t.Errorf("staleLockAge = %v, want 4h", staleLockAge)
	}
}
