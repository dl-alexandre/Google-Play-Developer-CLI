package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"runtime"
	"testing"
	"time"
)

func setConfigEnv(t *testing.T, tempHome string) {
	t.Setenv("HOME", tempHome)
	switch runtime.GOOS {
	case "windows":
		t.Setenv("APPDATA", tempHome)
		t.Setenv("LOCALAPPDATA", tempHome)
	default:
		t.Setenv("XDG_CONFIG_HOME", tempHome)
		t.Setenv("XDG_CACHE_HOME", tempHome)
	}
}

func TestTokenStorageKey(t *testing.T) {
	if got := tokenStorageKey("default", ""); got != "default" {
		t.Fatalf("expected default, got %q", got)
	}
	if got := tokenStorageKey("profile", "abcd"); got != "profile--abcd" {
		t.Fatalf("expected profile--abcd, got %q", got)
	}
}

func TestClientIDHash(t *testing.T) {
	hash, last4 := clientIDHash("")
	if hash != "" || last4 != "" {
		t.Fatalf("expected empty hash and last4, got %q %q", hash, last4)
	}

	clientID := "client-id-12345"
	expectedHash := sha256.Sum256([]byte(clientID))
	expectedHex := hex.EncodeToString(expectedHash[:])
	expectedLast4 := "2345"

	hash, last4 = clientIDHash(clientID)
	if hash != expectedHex {
		t.Fatalf("expected hash %q, got %q", expectedHex, hash)
	}
	if last4 != expectedLast4 {
		t.Fatalf("expected last4 %q, got %q", expectedLast4, last4)
	}
}

func TestTokenMetadataReadWriteFind(t *testing.T) {
	tempHome := t.TempDir()
	setConfigEnv(t, tempHome)

	meta := &TokenMetadata{
		Profile:      "default",
		ClientIDHash: "hash",
		Origin:       "keyfile",
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	if err := writeTokenMetadata("key1", meta); err != nil {
		t.Fatalf("writeTokenMetadata error: %v", err)
	}
	path := tokenMetadataPath("key1")
	read, err := readTokenMetadata(path)
	if err != nil {
		t.Fatalf("readTokenMetadata error: %v", err)
	}
	if read.Profile != "default" || read.ClientIDHash != "hash" {
		t.Fatalf("unexpected metadata: %+v", read)
	}

	if err := writeTokenMetadata("key2", meta); err != nil {
		t.Fatalf("writeTokenMetadata error: %v", err)
	}
	if err := os.Chtimes(tokenMetadataPath("key1"), time.Now().Add(-time.Hour), time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("chtimes error: %v", err)
	}
	found, key, err := findTokenMetadata("default")
	if err != nil {
		t.Fatalf("findTokenMetadata error: %v", err)
	}
	if found == nil || key != "key2" {
		t.Fatalf("expected key2, got %q", key)
	}
}

func TestLoadStoredToken(t *testing.T) {
	storage := &memoryStorage{available: true}
	m := NewManager(storage)
	m.SetStoreTokens("auto")

	token := StoredToken{
		AccessToken: "token",
		Expiry:      time.Now().Add(time.Hour).Format(time.RFC3339),
	}
	data, _ := json.Marshal(token)
	if err := storage.Store("key", data); err != nil {
		t.Fatalf("store error: %v", err)
	}
	loaded, err := m.loadStoredToken("key")
	if err != nil || loaded == nil {
		t.Fatalf("expected token, got %v %v", loaded, err)
	}

	token.Expiry = time.Now().Add(-time.Hour).Format(time.RFC3339)
	data, _ = json.Marshal(token)
	_ = storage.Store("key", data)
	loaded, err = m.loadStoredToken("key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded != nil {
		t.Fatalf("expected nil for expired token")
	}
}

func TestStorageEnabledAndLocation(t *testing.T) {
	m := NewManager(&memoryStorage{available: true})
	if !m.storageEnabled() {
		t.Fatalf("expected storage enabled")
	}
	if m.TokenLocation() != "secure-storage" {
		t.Fatalf("expected secure-storage")
	}
	m.SetStoreTokens("never")
	if m.storageEnabled() {
		t.Fatalf("expected storage disabled")
	}
	if m.TokenLocation() != "memory" {
		t.Fatalf("expected memory location")
	}
}

func TestLoadTokenMetadata(t *testing.T) {
	tempHome := t.TempDir()
	setConfigEnv(t, tempHome)

	meta := &TokenMetadata{Profile: "default", ClientIDHash: "hash", UpdatedAt: time.Now().UTC().Format(time.RFC3339)}
	if err := writeTokenMetadata("key", meta); err != nil {
		t.Fatalf("writeTokenMetadata error: %v", err)
	}
	m := NewManager(&memoryStorage{available: true})
	got, err := m.LoadTokenMetadata("default")
	if err != nil || got == nil {
		t.Fatalf("LoadTokenMetadata error: %v", err)
	}
}

func TestFindTokenMetadataMissingDir(t *testing.T) {
	tempHome := t.TempDir()
	setConfigEnv(t, tempHome)
	meta, key, err := findTokenMetadata("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta != nil || key != "" {
		t.Fatalf("expected nil metadata")
	}
}
