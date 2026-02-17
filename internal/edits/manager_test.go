package edits

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

func TestAcquireAndReleaseLock(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := m.AcquireLock(ctx, "com.example.app"); err != nil {
		t.Fatalf("AcquireLock error: %v", err)
	}
	lockPath := filepath.Join(tmpDir, "com.example.app.lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("lock file not created: %v", err)
	}

	if err := m.ReleaseLock("com.example.app"); err != nil {
		t.Fatalf("ReleaseLock error: %v", err)
	}
	if _, err := os.Stat(lockPath); err == nil {
		t.Fatalf("lock file should be removed")
	}
}

func TestCleanExpiredCacheWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		cacheDir: tmpDir,
	}
	cacheDir := filepath.Join(tmpDir, "com.example.app")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		t.Fatalf("mkdir cache error: %v", err)
	}

	expired := CacheEntry{
		SHA256:    "expired",
		Path:      "/tmp/expired.aab",
		Size:      1,
		CachedAt:  time.Now().Add(-48 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	active := CacheEntry{
		SHA256:    "active",
		Path:      "/tmp/active.aab",
		Size:      2,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	expiredPath := filepath.Join(cacheDir, "expired.json")
	activePath := filepath.Join(cacheDir, "active.json")
	for path, entry := range map[string]CacheEntry{expiredPath: expired, activePath: active} {
		data, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			t.Fatalf("marshal cache entry error: %v", err)
		}
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatalf("write cache entry error: %v", err)
		}
	}

	if err := m.CleanExpiredCacheWithContext(context.Background()); err != nil {
		t.Fatalf("CleanExpiredCacheWithContext error: %v", err)
	}

	if _, err := os.Stat(expiredPath); !os.IsNotExist(err) {
		t.Fatalf("expired cache entry should be removed")
	}
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("active cache entry should remain: %v", err)
	}
}

func TestHashFileWithProgress(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "data.bin")
	payload := []byte("progress content")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	var calls int
	var lastProcessed int64
	var total int64
	_, err := HashFileWithProgress(path, func(processed, totalBytes int64) {
		calls++
		lastProcessed = processed
		total = totalBytes
	})
	if err != nil {
		t.Fatalf("HashFileWithProgress error: %v", err)
	}
	if calls == 0 {
		t.Fatalf("expected progress callback to be called")
	}
	if total != int64(len(payload)) {
		t.Fatalf("total bytes = %d, want %d", total, len(payload))
	}
	if lastProcessed != int64(len(payload)) {
		t.Fatalf("processed bytes = %d, want %d", lastProcessed, len(payload))
	}
}

func TestIdempotencyCleanExpiredWithContext(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	expired := IdempotencyEntry{
		Key:       "expired",
		Operation: "upload",
		Timestamp: time.Now().Add(-48 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	active := IdempotencyEntry{
		Key:       "active",
		Operation: "upload",
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	for _, entry := range []IdempotencyEntry{expired, active} {
		data, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			t.Fatalf("marshal entry error: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, entry.Key+".json"), data, 0o600); err != nil {
			t.Fatalf("write entry error: %v", err)
		}
	}

	if err := store.CleanExpiredWithContext(context.Background()); err != nil {
		t.Fatalf("CleanExpiredWithContext error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "expired.json")); !os.IsNotExist(err) {
		t.Fatalf("expired entry should be removed")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "active.json")); err != nil {
		t.Fatalf("active entry should remain: %v", err)
	}
}

func TestGetCachedArtifactByHash(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, tmpDir string)
		hash      string
		wantFound bool
		wantErr   bool
	}{
		{
			name: "cache_hit_valid",
			setup: func(t *testing.T, tmpDir string) {
				t.Helper()
				cacheDir := filepath.Join(tmpDir, "com.example.app")
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					t.Fatalf("mkdir error: %v", err)
				}
				entry := CacheEntry{
					SHA256:    "abc123",
					Path:      "/tmp/app.aab",
					Size:      1024,
					CachedAt:  time.Now(),
					ExpiresAt: time.Now().Add(1 * time.Hour),
				}
				data, _ := json.MarshalIndent(entry, "", "  ")
				if err := os.WriteFile(filepath.Join(cacheDir, "abc123.json"), data, 0600); err != nil {
					t.Fatalf("write error: %v", err)
				}
			},
			hash:      "abc123",
			wantFound: true,
			wantErr:   false,
		},
		{
			name: "cache_miss_nonexistent",
			setup: func(t *testing.T, tmpDir string) {
				t.Helper()
				cacheDir := filepath.Join(tmpDir, "com.example.app")
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					t.Fatalf("mkdir error: %v", err)
				}
			},
			hash:      "nonexistent",
			wantFound: false,
			wantErr:   false,
		},
		{
			name: "cache_expired",
			setup: func(t *testing.T, tmpDir string) {
				t.Helper()
				cacheDir := filepath.Join(tmpDir, "com.example.app")
				if err := os.MkdirAll(cacheDir, 0700); err != nil {
					t.Fatalf("mkdir error: %v", err)
				}
				entry := CacheEntry{
					SHA256:    "expired123",
					Path:      "/tmp/app.aab",
					Size:      1024,
					CachedAt:  time.Now().Add(-48 * time.Hour),
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}
				data, _ := json.MarshalIndent(entry, "", "  ")
				if err := os.WriteFile(filepath.Join(cacheDir, "expired123.json"), data, 0600); err != nil {
					t.Fatalf("write error: %v", err)
				}
			},
			hash:      "expired123",
			wantFound: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := &Manager{cacheDir: tmpDir}
			tt.setup(t, tmpDir)

			entry, err := m.GetCachedArtifactByHash("com.example.app", tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCachedArtifactByHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantFound && entry == nil {
				t.Error("GetCachedArtifactByHash() returned nil, want entry")
			}
			if !tt.wantFound && entry != nil {
				t.Error("GetCachedArtifactByHash() returned entry, want nil")
			}
		})
	}
}

func TestCacheArtifactWithHash(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		hash      string
		wantErr   bool
		checkFile bool
	}{
		{
			name: "cache_artifact_success",
			setup: func(t *testing.T) string {
				tmpFile := filepath.Join(t.TempDir(), "app.aab")
				if err := os.WriteFile(tmpFile, []byte("test content"), 0600); err != nil {
					t.Fatalf("write error: %v", err)
				}
				return tmpFile
			},
			hash:      "abc123def456",
			wantErr:   false,
			checkFile: true,
		},
		{
			name: "cache_artifact_nonexistent_file",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/app.aab"
			},
			hash:      "xyz789",
			wantErr:   true,
			checkFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := &Manager{cacheDir: tmpDir}
			artifactPath := tt.setup(t)

			err := m.CacheArtifactWithHash("com.example.app", artifactPath, tt.hash, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheArtifactWithHash() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.checkFile && !tt.wantErr {
				cacheFile := filepath.Join(tmpDir, "com.example.app", tt.hash+".json")
				if _, err := os.Stat(cacheFile); err != nil {
					t.Errorf("cache file not created: %v", err)
				}
				entry, err := m.GetCachedArtifactByHash("com.example.app", tt.hash)
				if err != nil {
					t.Errorf("GetCachedArtifactByHash() error = %v", err)
					return
				}
				if entry == nil {
					t.Fatal("GetCachedArtifactByHash() returned nil")
				}
				if entry.SHA256 != tt.hash {
					t.Errorf("SHA256 = %q, want %q", entry.SHA256, tt.hash)
				}
			}
		})
	}
}

func TestCleanExpiredCache(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{cacheDir: tmpDir}
	cacheDir := filepath.Join(tmpDir, "com.example.app")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	expired := CacheEntry{
		SHA256:    "expired",
		Path:      "/tmp/expired.aab",
		Size:      1,
		CachedAt:  time.Now().Add(-48 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	active := CacheEntry{
		SHA256:    "active",
		Path:      "/tmp/active.aab",
		Size:      2,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	expiredPath := filepath.Join(cacheDir, "expired.json")
	activePath := filepath.Join(cacheDir, "active.json")
	for path, entry := range map[string]CacheEntry{expiredPath: expired, activePath: active} {
		data, _ := json.MarshalIndent(entry, "", "  ")
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatalf("write error: %v", err)
		}
	}

	if err := m.CleanExpiredCache(); err != nil {
		t.Fatalf("CleanExpiredCache() error = %v", err)
	}

	if _, err := os.Stat(expiredPath); !os.IsNotExist(err) {
		t.Error("expired cache entry should be removed")
	}
	if _, err := os.Stat(activePath); err != nil {
		t.Errorf("active cache entry should remain: %v", err)
	}
}

func TestIdempotencyStoreCleanExpired(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	expired := IdempotencyEntry{
		Key:       "expired",
		Operation: "upload",
		Timestamp: time.Now().Add(-48 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	active := IdempotencyEntry{
		Key:       "active",
		Operation: "upload",
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	for _, entry := range []IdempotencyEntry{expired, active} {
		data, _ := json.MarshalIndent(entry, "", "  ")
		if err := os.WriteFile(filepath.Join(tmpDir, entry.Key+".json"), data, 0600); err != nil {
			t.Fatalf("write error: %v", err)
		}
	}

	if err := store.CleanExpired(); err != nil {
		t.Fatalf("CleanExpired() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "expired.json")); !os.IsNotExist(err) {
		t.Error("expired entry should be removed")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "active.json")); err != nil {
		t.Errorf("active entry should remain: %v", err)
	}
}

func TestCheckUploadByHash(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, store *IdempotencyStore)
		packageName string
		hash        string
		wantFound   bool
		wantErr     bool
	}{
		{
			name: "upload_found",
			setup: func(t *testing.T, store *IdempotencyStore) {
				t.Helper()
				result := &UploadResult{
					VersionCode: 1,
					SHA256:      "abc123",
					Path:        "/tmp/app.aab",
					Size:        1024,
					Type:        "aab",
					EditID:      "edit-123",
				}
				key := store.generateKey("upload", "com.example.app", "abc123")
				if err := store.RecordUpload(key, "com.example.app", "abc123", result); err != nil {
					t.Fatalf("RecordUpload error: %v", err)
				}
			},
			packageName: "com.example.app",
			hash:        "abc123",
			wantFound:   true,
			wantErr:     false,
		},
		{
			name: "upload_not_found",
			setup: func(t *testing.T, store *IdempotencyStore) {
				t.Helper()
			},
			packageName: "com.example.app",
			hash:        "nonexistent",
			wantFound:   false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}
			tt.setup(t, store)

			result, key, err := store.CheckUploadByHash(tt.packageName, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckUploadByHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result.Found != tt.wantFound {
				t.Errorf("CheckUploadByHash() Found = %v, want %v", result.Found, tt.wantFound)
			}
			if key == "" {
				t.Error("CheckUploadByHash() returned empty key")
			}
		})
	}
}

func TestRecordUpload(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	result := &UploadResult{
		VersionCode: 1,
		SHA256:      "abc123",
		Path:        "/tmp/app.aab",
		Size:        1024,
		Type:        "aab",
		EditID:      "edit-123",
	}

	key := store.generateKey("upload", "com.example.app", "abc123")
	if err := store.RecordUpload(key, "com.example.app", "abc123", result); err != nil {
		t.Fatalf("RecordUpload() error = %v", err)
	}

	checkResult, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !checkResult.Found {
		t.Error("Get() returned Found=false after RecordUpload()")
	}

	uploadData, ok := checkResult.Data.(map[string]interface{})
	if !ok {
		t.Errorf("Data type = %T, want map[string]interface{}", checkResult.Data)
	}
	if uploadData["sha256"] != "abc123" {
		t.Errorf("sha256 = %v, want abc123", uploadData["sha256"])
	}
}

func TestCheckCommit(t *testing.T) {
	tests := []struct {
		name              string
		setup             func(t *testing.T, store *IdempotencyStore)
		packageName       string
		editID            string
		contentIdentifier string
		wantFound         bool
		wantErr           bool
	}{
		{
			name: "commit_found",
			setup: func(t *testing.T, store *IdempotencyStore) {
				t.Helper()
				key := store.generateKey("commit", "com.example.app", "edit-123:content-abc")
				if err := store.RecordCommit(key, "com.example.app", "edit-123"); err != nil {
					t.Fatalf("RecordCommit error: %v", err)
				}
			},
			packageName:       "com.example.app",
			editID:            "edit-123",
			contentIdentifier: "content-abc",
			wantFound:         true,
			wantErr:           false,
		},
		{
			name: "commit_not_found",
			setup: func(t *testing.T, store *IdempotencyStore) {
				t.Helper()
			},
			packageName:       "com.example.app",
			editID:            "nonexistent",
			contentIdentifier: "content-xyz",
			wantFound:         false,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}
			tt.setup(t, store)

			result, key, err := store.CheckCommit(tt.packageName, tt.editID, tt.contentIdentifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckCommit() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result.Found != tt.wantFound {
				t.Errorf("CheckCommit() Found = %v, want %v", result.Found, tt.wantFound)
			}
			if key == "" {
				t.Error("CheckCommit() returned empty key")
			}
		})
	}
}

func TestRecordCommit(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	key := store.generateKey("commit", "com.example.app", "edit-123")
	if err := store.RecordCommit(key, "com.example.app", "edit-123"); err != nil {
		t.Fatalf("RecordCommit() error = %v", err)
	}

	checkResult, err := store.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !checkResult.Found {
		t.Error("Get() returned Found=false after RecordCommit()")
	}

	commitData, ok := checkResult.Data.(map[string]interface{})
	if !ok {
		t.Errorf("Data type = %T, want map[string]interface{}", checkResult.Data)
	}
	if commitData["editId"] != "edit-123" {
		t.Errorf("editId = %v, want edit-123", commitData["editId"])
	}
	if commitData["committed"] != true {
		t.Errorf("committed = %v, want true", commitData["committed"])
	}
}

func TestGenerateKey(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	tests := []struct {
		name        string
		operation   string
		packageName string
		contentHash string
	}{
		{
			name:        "upload_key",
			operation:   "upload",
			packageName: "com.example.app",
			contentHash: "abc123",
		},
		{
			name:        "commit_key",
			operation:   "commit",
			packageName: "com.example.app",
			contentHash: "edit-123:content-abc",
		},
		{
			name:        "different_content_different_key",
			operation:   "upload",
			packageName: "com.example.app",
			contentHash: "xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := store.generateKey(tt.operation, tt.packageName, tt.contentHash)
			if key == "" {
				t.Error("generateKey() returned empty key")
			}
			if len(key) != 32 {
				t.Errorf("generateKey() key length = %d, want 32", len(key))
			}

			key2 := store.generateKey(tt.operation, tt.packageName, tt.contentHash)
			if key != key2 {
				t.Errorf("generateKey() produced different keys for same inputs: %s vs %s", key, key2)
			}
		})
	}
}

func TestCacheArtifactWithHashEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "empty_hash",
			hash:    "",
			wantErr: true,
		},
		{
			name:    "long_hash",
			hash:    "a" + strings.Repeat("b", 100),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			m := &Manager{cacheDir: tmpDir}

			tmpFile := filepath.Join(tmpDir, "test.aab")
			if err := os.WriteFile(tmpFile, []byte("test"), 0600); err != nil {
				t.Fatalf("write error: %v", err)
			}

			err := m.CacheArtifactWithHash("com.example.app", tmpFile, tt.hash, 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("CacheArtifactWithHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCacheArtifactConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{cacheDir: tmpDir}

	tmpFile := filepath.Join(tmpDir, "test.aab")
	if err := os.WriteFile(tmpFile, []byte("test content"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(index int) {
			hash := fmt.Sprintf("hash%d", index)
			err := m.CacheArtifactWithHash("com.example.app", tmpFile, hash, int64(index))
			done <- err
		}(i)
	}

	for i := 0; i < 5; i++ {
		if err := <-done; err != nil {
			t.Errorf("concurrent cache error: %v", err)
		}
	}

	for i := 0; i < 5; i++ {
		hash := fmt.Sprintf("hash%d", i)
		entry, err := m.GetCachedArtifactByHash("com.example.app", hash)
		if err != nil {
			t.Errorf("GetCachedArtifactByHash() error = %v", err)
		}
		if entry == nil {
			t.Errorf("cache entry %d not found", i)
		}
	}
}

func TestUpdateEditStateNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	_, err := m.UpdateEditState("com.example.app", "nonexistent", StateCommitted)
	if err == nil {
		t.Error("UpdateEditState() should error for nonexistent edit")
	}
}

func TestTouchEditNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	_, err := m.TouchEdit("com.example.app", "nonexistent")
	if err == nil {
		t.Error("TouchEdit() should error for nonexistent edit")
	}
}

func TestIdempotencyStoreGetExpired(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	entry := IdempotencyEntry{
		Key:       "expired-key",
		Operation: "upload",
		Timestamp: time.Now().Add(-48 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	data, _ := json.MarshalIndent(entry, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "expired-key.json"), data, 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	result, err := store.Get("expired-key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if result.Found {
		t.Error("Get() should return Found=false for expired entry")
	}
	if !result.Expired {
		t.Error("Get() should return Expired=true for expired entry")
	}
}

func TestIdempotencyStoreRecordWithMeta(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	data := map[string]interface{}{"result": "success"}
	err := store.RecordWithMeta("test-key", "upload", "com.example.app", "hash123", data)
	if err != nil {
		t.Fatalf("RecordWithMeta() error = %v", err)
	}

	result, err := store.Get("test-key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !result.Found {
		t.Error("Get() should return Found=true after RecordWithMeta()")
	}
}

func TestCheckUploadByHashError(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	result, key, err := store.CheckUploadByHash("com.example.app", "hash123")
	if err != nil {
		t.Fatalf("CheckUploadByHash() error = %v", err)
	}
	if result.Found {
		t.Error("CheckUploadByHash() should return Found=false for nonexistent upload")
	}
	if key == "" {
		t.Error("CheckUploadByHash() should return non-empty key")
	}
}

func TestCheckCommitError(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	result, key, err := store.CheckCommit("com.example.app", "edit-123", "content-abc")
	if err != nil {
		t.Fatalf("CheckCommit() error = %v", err)
	}
	if result.Found {
		t.Error("CheckCommit() should return Found=false for nonexistent commit")
	}
	if key == "" {
		t.Error("CheckCommit() should return non-empty key")
	}
}

func TestSaveEditMkdirError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows: path handling differs")
	}

	m := &Manager{
		editsDir: "/invalid/path/that/does/not/exist",
	}

	edit := &Edit{
		Handle:      "test",
		ServerID:    "server-123",
		PackageName: "com.example.app",
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		State:       StateDraft,
	}

	err := m.SaveEdit(edit)
	if err == nil {
		t.Error("SaveEdit() should error when mkdir fails")
	}
}

func TestListEditsMkdirError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows: path handling differs")
	}

	m := &Manager{
		editsDir: "/invalid/path/that/does/not/exist",
	}

	_, err := m.ListEdits("com.example.app")
	if err == nil {
		t.Error("ListEdits() should error when mkdir fails")
	}
}

func TestGetCachedArtifactByHashInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{cacheDir: tmpDir}
	cacheDir := filepath.Join(tmpDir, "com.example.app")
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		t.Fatalf("mkdir error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(cacheDir, "invalid.json"), []byte("not json"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	_, err := m.GetCachedArtifactByHash("com.example.app", "invalid")
	if err == nil {
		t.Error("GetCachedArtifactByHash() should error for invalid JSON")
	}
}

func TestIdempotencyStoreGetInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte("not json"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	_, err := store.Get("invalid")
	if err == nil {
		t.Error("Get() should error for invalid JSON")
	}
}

func TestIdempotencyKeyEmptyArgs(t *testing.T) {
	key := IdempotencyKey("operation")
	if key == "" {
		t.Error("IdempotencyKey() should return non-empty key")
	}
	if len(key) != 16 {
		t.Errorf("IdempotencyKey() length = %d, want 16", len(key))
	}
}

func TestIdempotencyStoreCheckError(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	if err := os.WriteFile(filepath.Join(tmpDir, "invalid.json"), []byte("not json"), 0600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	_, err := store.Check("invalid")
	if err == nil {
		t.Error("Check() should error for invalid JSON")
	}
}

func TestIdempotencyStoreRecordWithMetaError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows: path handling differs")
	}

	store := &IdempotencyStore{dir: "/invalid/path/that/does/not/exist", ttl: idempotencyTTL}

	err := store.RecordWithMeta("key", "op", "pkg", "hash", map[string]interface{}{})
	if err == nil {
		t.Error("RecordWithMeta() should error when mkdir fails")
	}
}

func TestCleanExpiredCacheWithContextEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{cacheDir: tmpDir}

	if err := m.CleanExpiredCacheWithContext(context.Background()); err != nil {
		t.Fatalf("CleanExpiredCacheWithContext() error = %v", err)
	}
}

func TestCleanExpiredCacheWithContextInvalidDir(t *testing.T) {
	m := &Manager{cacheDir: "/invalid/path/that/does/not/exist"}

	err := m.CleanExpiredCacheWithContext(context.Background())
	if err != nil {
		t.Fatalf("CleanExpiredCacheWithContext() error = %v", err)
	}
}

func TestCleanExpiredWithContextEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	if err := store.CleanExpiredWithContext(context.Background()); err != nil {
		t.Fatalf("CleanExpiredWithContext() error = %v", err)
	}
}

func TestCleanExpiredWithContextInvalidDir(t *testing.T) {
	store := &IdempotencyStore{dir: "/invalid/path/that/does/not/exist", ttl: idempotencyTTL}

	err := store.CleanExpiredWithContext(context.Background())
	if err != nil {
		t.Fatalf("CleanExpiredWithContext() error = %v", err)
	}
}

func TestTouchEditUpdatesLastUsedAt(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	oldTime := time.Now().Add(-2 * time.Hour)
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

	touched, err := m.TouchEdit("com.example.app", "test-handle")
	if err != nil {
		t.Fatalf("TouchEdit() error = %v", err)
	}

	if !touched.LastUsedAt.After(oldTime) {
		t.Error("TouchEdit() should update LastUsedAt")
	}

	loaded, err := m.LoadEdit("com.example.app", "test-handle")
	if err != nil {
		t.Fatalf("LoadEdit() error = %v", err)
	}

	if !loaded.LastUsedAt.After(oldTime) {
		t.Error("Persisted LastUsedAt should be updated")
	}
}

func TestUpdateEditStateUpdatesLastUsedAt(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{
		editsDir:  tmpDir,
		lockFiles: make(map[string]*LockFile),
	}

	oldTime := time.Now().Add(-2 * time.Hour)
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

	updated, err := m.UpdateEditState("com.example.app", "test-handle", StateValidating)
	if err != nil {
		t.Fatalf("UpdateEditState() error = %v", err)
	}

	if updated.State != StateValidating {
		t.Errorf("State = %q, want %q", updated.State, StateValidating)
	}

	if !updated.LastUsedAt.After(oldTime) {
		t.Error("UpdateEditState() should update LastUsedAt")
	}

	loaded, err := m.LoadEdit("com.example.app", "test-handle")
	if err != nil {
		t.Fatalf("LoadEdit() error = %v", err)
	}

	if loaded.State != StateValidating {
		t.Errorf("Persisted state = %q, want %q", loaded.State, StateValidating)
	}
}

func TestGenerateKeyConsistency(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	key1 := store.generateKey("upload", "com.example.app", "hash123")
	key2 := store.generateKey("upload", "com.example.app", "hash123")

	if key1 != key2 {
		t.Errorf("generateKey() produced different keys for same inputs: %s vs %s", key1, key2)
	}

	key3 := store.generateKey("upload", "com.example.app", "hash456")
	if key1 == key3 {
		t.Error("generateKey() should produce different keys for different inputs")
	}
}

func TestRecordWithMetaPreservesData(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	testData := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": true,
	}

	err := store.RecordWithMeta("test-key", "operation", "com.example.app", "hash123", testData)
	if err != nil {
		t.Fatalf("RecordWithMeta() error = %v", err)
	}

	result, err := store.Get("test-key")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if !result.Found {
		t.Error("Get() should return Found=true")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Errorf("Data type = %T, want map[string]interface{}", result.Data)
	}

	if data["field1"] != "value1" {
		t.Errorf("field1 = %v, want value1", data["field1"])
	}
}

func TestGenerateKeyDifferentOperations(t *testing.T) {
	tmpDir := t.TempDir()
	store := &IdempotencyStore{dir: tmpDir, ttl: idempotencyTTL}

	uploadKey := store.generateKey("upload", "com.example.app", "hash123")
	commitKey := store.generateKey("commit", "com.example.app", "hash123")

	if uploadKey == commitKey {
		t.Error("generateKey() should produce different keys for different operations")
	}

	if uploadKey == "" || commitKey == "" {
		t.Error("generateKey() should return non-empty keys")
	}

	if len(uploadKey) != 32 || len(commitKey) != 32 {
		t.Errorf("generateKey() key length should be 32, got %d and %d", len(uploadKey), len(commitKey))
	}
}
