package auth

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"golang.org/x/oauth2"

	gpdErrors "github.com/dl-alexandre/gpd/internal/errors"
)

type countingTokenSource struct {
	tokens []*oauth2.Token
	err    error
	calls  int
}

func (c *countingTokenSource) Token() (*oauth2.Token, error) {
	c.calls++
	if c.err != nil {
		return nil, c.err
	}
	if len(c.tokens) == 0 {
		return nil, nil
	}
	token := c.tokens[0]
	c.tokens = c.tokens[1:]
	return token, nil
}

func TestEarlyRefreshTokenSource(t *testing.T) {
	base := &countingTokenSource{
		tokens: []*oauth2.Token{
			{AccessToken: "a", Expiry: time.Now().Add(time.Hour)},
			{AccessToken: "b", Expiry: time.Now().Add(2 * time.Hour)},
		},
	}
	src := &EarlyRefreshTokenSource{
		base:          base,
		refreshLeeway: 0,
		clockSkew:     0,
	}
	token, err := src.Token()
	if err != nil || token.AccessToken != "a" {
		t.Fatalf("unexpected token: %v %v", token, err)
	}
	token, err = src.Token()
	if err != nil || token.AccessToken != "a" {
		t.Fatalf("expected cached token, got %v %v", token, err)
	}
	if base.calls != 1 {
		t.Fatalf("expected 1 base call, got %d", base.calls)
	}
	src.cachedToken.Expiry = time.Now().Add(-time.Minute)
	token, err = src.Token()
	if err != nil || token.AccessToken != "b" {
		t.Fatalf("expected refreshed token, got %v %v", token, err)
	}
}

func TestEarlyRefreshTokenSourceError(t *testing.T) {
	base := &countingTokenSource{err: os.ErrNotExist}
	src := &EarlyRefreshTokenSource{
		base:          base,
		refreshLeeway: 0,
		clockSkew:     0,
	}
	if _, err := src.Token(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPersistedTokenSource(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	storage := &memoryStorage{available: true}
	base := &countingTokenSource{
		tokens: []*oauth2.Token{
			{AccessToken: "token", Expiry: time.Now().Add(time.Hour)},
		},
	}
	meta := &TokenMetadata{
		Profile:      "default",
		ClientIDHash: "hash",
		Origin:       "keyfile",
		Email:        "test@example.com",
		Scopes:       []string{"a"},
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	src := &PersistedTokenSource{
		base:       base,
		storage:    storage,
		storageKey: "key",
		metadata:   meta,
	}
	token, err := src.Token()
	if err != nil || token == nil {
		t.Fatalf("unexpected token: %v %v", token, err)
	}
	data, err := storage.Retrieve("key")
	if err != nil {
		t.Fatalf("retrieve error: %v", err)
	}
	var stored StoredToken
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if stored.AccessToken != "token" || stored.Origin != "keyfile" {
		t.Fatalf("unexpected stored token: %+v", stored)
	}
	if meta.TokenExpiry == "" || meta.UpdatedAt == "" {
		t.Fatalf("expected metadata updated")
	}
	if _, err := os.Stat(tokenMetadataPath("key")); err != nil {
		t.Fatalf("expected metadata file, got %v", err)
	}
}

func TestWrapTokenSourceWithStorage(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	m := NewManager(&memoryStorage{available: true})
	m.SetStoreTokens("auto")

	base := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "token",
		Expiry:      time.Now().Add(time.Hour),
	})
	ts := m.wrapTokenSource(base, OriginKeyfile, "email", "clientid", []string{"scope"})
	if _, ok := ts.(*PersistedTokenSource); !ok {
		t.Fatalf("expected PersistedTokenSource")
	}
}

func TestPersistedTokenSourceError(t *testing.T) {
	src := &PersistedTokenSource{
		base: &countingTokenSource{err: gpdErrors.ErrAuthNotConfigured},
	}
	if _, err := src.Token(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPersistedTokenSourceNilToken(t *testing.T) {
	src := &PersistedTokenSource{
		base:    &countingTokenSource{},
		storage: nil,
	}
	token, err := src.Token()
	if err != nil || token != nil {
		t.Fatalf("expected nil token, got %v %v", token, err)
	}
}
