package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

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
