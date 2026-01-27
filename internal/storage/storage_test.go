package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/99designs/keyring"
)

func TestStorageUnavailable(t *testing.T) {
	s := NewWithKeyring(nil)
	if err := s.Store("k", []byte("v")); err != ErrStorageUnavailable {
		t.Fatalf("expected ErrStorageUnavailable, got %v", err)
	}
	if _, err := s.Retrieve("k"); err != ErrStorageUnavailable {
		t.Fatalf("expected ErrStorageUnavailable, got %v", err)
	}
	if err := s.Delete("k"); err != ErrStorageUnavailable {
		t.Fatalf("expected ErrStorageUnavailable, got %v", err)
	}
}

func TestStorageRoundTrip(t *testing.T) {
	s := NewWithKeyring(keyring.NewArrayKeyring(nil))
	key := fmt.Sprintf("gpd-test-%d", time.Now().UnixNano())
	value := []byte("value")

	if err := s.Store(key, value); err != nil {
		t.Fatalf("store error: %v", err)
	}
	got, err := s.Retrieve(key)
	if err != nil {
		t.Fatalf("retrieve error: %v", err)
	}
	if string(got) != "value" {
		t.Fatalf("unexpected value: %s", string(got))
	}
	if err := s.Delete(key); err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if _, err := s.Retrieve(key); err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
	if Platform() == "" {
		t.Fatalf("expected platform value")
	}
}

func TestNewKeyringOpenError(t *testing.T) {
	orig := openKeyring
	openKeyring = func(cfg keyring.Config) (keyring.Keyring, error) {
		return nil, ErrStorageUnavailable
	}
	t.Cleanup(func() { openKeyring = orig })
	s := New()
	if s.Available() {
		t.Fatalf("expected unavailable storage")
	}
}
