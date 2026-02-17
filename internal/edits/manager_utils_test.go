package edits

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSanitizeHandle(t *testing.T) {
	m := &Manager{}
	if got := m.sanitizeHandle(""); got != "default" {
		t.Fatalf("expected default, got %q", got)
	}
	if got := m.sanitizeHandle("a/b\\c:d"); got != "a_b_c_d" {
		t.Fatalf("expected sanitized handle, got %q", got)
	}
}

func TestEditPathAndPrefix(t *testing.T) {
	dir := t.TempDir()
	m := &Manager{editsDir: dir}

	prefix := m.editPrefix("com/example:app")
	if prefix != "com_example_app_" {
		t.Fatalf("expected prefix com_example_app_, got %q", prefix)
	}

	path := m.editPath("com/example:app", "handle/one")
	expected := filepath.Join(dir, "com_example_app_handle_one.json")
	if path != expected {
		t.Fatalf("expected path %q, got %q", expected, path)
	}
}

func TestIsEditExpired(t *testing.T) {
	m := &Manager{}
	now := time.Now()

	edit := &Edit{
		CreatedAt:  now.Add(-2 * time.Hour),
		LastUsedAt: now.Add(-30 * time.Minute),
	}
	if m.IsEditExpired(edit, now) {
		t.Fatalf("expected edit not expired")
	}

	edit.CreatedAt = now.Add(-editTTL - time.Minute)
	if !m.IsEditExpired(edit, now) {
		t.Fatalf("expected edit expired by ttl")
	}

	edit.CreatedAt = now.Add(-2 * time.Hour)
	edit.LastUsedAt = now.Add(-editIdleTTL - time.Minute)
	if !m.IsEditExpired(edit, now) {
		t.Fatalf("expected edit expired by idle ttl")
	}
}
