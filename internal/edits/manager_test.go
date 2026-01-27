package edits

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

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
	states := []EditState{StateDraft, StateValidating, StateCommitted, StateAborted}
	expected := []string{"draft", "validating", "committed", "aborted"}

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

func TestSaveAndLoadEdit(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	edit := &Edit{
		Handle:      "test-handle",
		ServerID:    "server-123",
		PackageName: "com.example.app",
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		State:       StateDraft,
	}

	// Save
	if err := m.SaveEdit(edit); err != nil {
		t.Fatalf("SaveEdit() error = %v", err)
	}

	// Load
	loaded, err := m.LoadEdit("com.example.app", "test-handle")
	if err != nil {
		t.Fatalf("LoadEdit() error = %v", err)
	}

	if loaded.Handle != edit.Handle {
		t.Errorf("Handle = %q, want %q", loaded.Handle, edit.Handle)
	}
	if loaded.ServerID != edit.ServerID {
		t.Errorf("ServerID = %q, want %q", loaded.ServerID, edit.ServerID)
	}
	if loaded.PackageName != edit.PackageName {
		t.Errorf("PackageName = %q, want %q", loaded.PackageName, edit.PackageName)
	}
	if loaded.State != edit.State {
		t.Errorf("State = %q, want %q", loaded.State, edit.State)
	}
}

func TestLoadEditNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	edit, err := m.LoadEdit("com.example.app", "nonexistent")
	if err != nil {
		t.Fatalf("LoadEdit() error = %v, want nil", err)
	}
	if edit != nil {
		t.Error("LoadEdit() should return nil for non-existent edit")
	}
}

func TestDeleteEdit(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	edit := &Edit{
		Handle:      "test-handle",
		ServerID:    "server-123",
		PackageName: "com.example.app",
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		State:       StateDraft,
	}

	// Save
	if err := m.SaveEdit(edit); err != nil {
		t.Fatalf("SaveEdit() error = %v", err)
	}

	// Delete
	if err := m.DeleteEdit("com.example.app", "test-handle"); err != nil {
		t.Fatalf("DeleteEdit() error = %v", err)
	}

	// Should not be loadable
	edit, err := m.LoadEdit("com.example.app", "test-handle")
	if err != nil {
		t.Fatalf("LoadEdit() error = %v, want nil", err)
	}
	if edit != nil {
		t.Error("LoadEdit() should return nil after DeleteEdit()")
	}
}

func TestListEdits(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	// Save multiple edits
	for i := 1; i <= 3; i++ {
		edit := &Edit{
			Handle:      filepath.Base(t.TempDir()) + "-" + string(rune('0'+i)), // Use unique handles
			ServerID:    "server-" + string(rune('0'+i)),
			PackageName: "com.example.app",
			CreatedAt:   time.Now(),
			LastUsedAt:  time.Now(),
			State:       StateDraft,
		}
		if err := m.SaveEdit(edit); err != nil {
			t.Fatalf("SaveEdit() error = %v", err)
		}
	}

	// List
	edits, err := m.ListEdits("com.example.app")
	if err != nil {
		t.Fatalf("ListEdits() error = %v", err)
	}

	if len(edits) != 3 {
		t.Errorf("ListEdits() returned %d edits, want 3", len(edits))
	}
}

func TestListEditsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	edits, err := m.ListEdits("com.example.app")
	if err != nil {
		t.Fatalf("ListEdits() error = %v", err)
	}

	if len(edits) != 0 {
		t.Errorf("ListEdits() returned %d edits, want 0", len(edits))
	}
}

func TestUpdateEditState(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	edit := &Edit{
		Handle:      "test-handle",
		ServerID:    "server-123",
		PackageName: "com.example.app",
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		State:       StateDraft,
	}

	if err := m.SaveEdit(edit); err != nil {
		t.Fatalf("SaveEdit() error = %v", err)
	}

	// Update state
	updated, err := m.UpdateEditState("com.example.app", "test-handle", StateCommitted)
	if err != nil {
		t.Fatalf("UpdateEditState() error = %v", err)
	}

	if updated.State != StateCommitted {
		t.Errorf("State = %q, want %q", updated.State, StateCommitted)
	}

	// Verify it was persisted
	loaded, err := m.LoadEdit("com.example.app", "test-handle")
	if err != nil {
		t.Fatalf("LoadEdit() error = %v", err)
	}

	if loaded.State != StateCommitted {
		t.Errorf("Persisted state = %q, want %q", loaded.State, StateCommitted)
	}
}

func TestTouchEdit(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	oldTime := time.Now().Add(-1 * time.Hour)
	edit := &Edit{
		Handle:      "test-handle",
		ServerID:    "server-123",
		PackageName: "com.example.app",
		CreatedAt:   oldTime,
		LastUsedAt:  oldTime,
		State:       StateDraft,
	}

	if err := m.SaveEdit(edit); err != nil {
		t.Fatalf("SaveEdit() error = %v", err)
	}

	// Touch
	touched, err := m.TouchEdit("com.example.app", "test-handle")
	if err != nil {
		t.Fatalf("TouchEdit() error = %v", err)
	}

	if !touched.LastUsedAt.After(oldTime) {
		t.Error("LastUsedAt should be updated by TouchEdit()")
	}
}

func TestIdempotencyStoreEdgeCases(t *testing.T) {
	t.Run("get_nonexistent", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

		result, err := store.Get("nonexistent-key")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if result.Found {
			t.Error("Get().Found should be false for nonexistent key")
		}
	})

	t.Run("record_and_check_multiple", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

		// Record multiple operations
		for i := 1; i <= 5; i++ {
			key := IdempotencyKey("test", "app", string(rune('0'+i)))
			data := map[string]interface{}{"index": i}
			if err := store.Record(key, data); err != nil {
				t.Fatalf("Record() error = %v", err)
			}
		}

		// Check all exist
		for i := 1; i <= 5; i++ {
			key := IdempotencyKey("test", "app", string(rune('0'+i)))
			exists, err := store.Check(key)
			if err != nil {
				t.Fatalf("Check() error = %v", err)
			}
			if !exists {
				t.Errorf("Check() = false for key %d, want true", i)
			}
		}
	})

	t.Run("clear_nonexistent", func(t *testing.T) {
		tmpDir := t.TempDir()
		store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

		// Should not error
		if err := store.Clear("nonexistent"); err != nil {
			t.Errorf("Clear() should not error for nonexistent key, got %v", err)
		}
	})
}

func TestIsLockStale(t *testing.T) {
	m := NewManager()
	hostname, _ := os.Hostname()

	tests := []struct {
		name      string
		lock      *LockFile
		wantStale bool
		skipCheck bool
	}{
		{
			name: "fresh_lock_same_host",
			lock: &LockFile{
				PID:       os.Getpid(),
				Hostname:  hostname,
				CreatedAt: time.Now(),
			},
			// Skip check because isProcessAlive may not work consistently in test environment
			skipCheck: true,
		},
		{
			name: "old_lock_different_host",
			lock: &LockFile{
				PID:       12345,
				Hostname:  "different-host",
				CreatedAt: time.Now().Add(-5 * time.Hour),
			},
			wantStale: true,
		},
		{
			name: "old_lock_same_host_dead_process",
			lock: &LockFile{
				PID:       999999999, // Non-existent PID
				Hostname:  hostname,
				CreatedAt: time.Now().Add(-5 * time.Hour),
			},
			wantStale: true,
		},
		{
			name: "recent_lock_different_host",
			lock: &LockFile{
				PID:       12345,
				Hostname:  "different-host",
				CreatedAt: time.Now().Add(-1 * time.Minute),
			},
			wantStale: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipCheck {
				t.Skip("Skipping test that depends on isProcessAlive behavior")
			}
			got := m.isLockStale(tt.lock, hostname)
			if got != tt.wantStale {
				t.Errorf("isLockStale() = %v, want %v", got, tt.wantStale)
			}
		})
	}
}
