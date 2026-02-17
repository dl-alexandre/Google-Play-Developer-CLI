package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"testing"
	"time"

	"golang.org/x/oauth2"

	gpdErrors "github.com/dl-alexandre/gpd/internal/errors"
)

type errorTokenSource struct{}

func (e errorTokenSource) Token() (*oauth2.Token, error) {
	return nil, gpdErrors.ErrAuthNotConfigured
}

func serviceAccountJSON(t *testing.T) []byte {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("key gen error: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key error: %v", err)
	}
	block := &pem.Block{Type: "PRIVATE KEY", Bytes: der}
	privateKey := string(pem.EncodeToMemory(block))

	payload := map[string]string{
		"type":                        "service_account",
		"project_id":                  "test",
		"private_key_id":              "keyid",
		"private_key":                 privateKey,
		"client_email":                "test@example.com",
		"client_id":                   "123456789",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/test",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}
	return data
}

func TestCredentialOriginString(t *testing.T) {
	if OriginADC.String() != "adc" || OriginKeyfile.String() != "keyfile" || OriginEnvironment.String() != "environment" {
		t.Fatalf("unexpected origin string")
	}
	var unknown CredentialOrigin = 99
	if unknown.String() != "unknown" {
		t.Fatalf("expected unknown origin string")
	}
}

func TestAuthenticateKeyfile(t *testing.T) {
	data := serviceAccountJSON(t)
	path := filepathWithTempFile(t, data)

	m := NewManager(&memoryStorage{available: false})
	creds, err := m.Authenticate(context.Background(), path)
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if creds.Origin != OriginKeyfile || creds.KeyPath != path {
		t.Fatalf("unexpected credentials: %+v", creds)
	}
}

func TestAuthenticateKeyfileMissing(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	if _, err := m.authenticateFromKeyfile(context.Background(), "/no/such/file.json", []string{"scope"}); err == nil {
		t.Fatalf("expected error for missing keyfile")
	}
}

func TestAuthenticateEnvKey(t *testing.T) {
	data := serviceAccountJSON(t)
	t.Setenv("GPD_SERVICE_ACCOUNT_KEY", string(data))
	m := NewManager(&memoryStorage{available: false})
	creds, err := m.Authenticate(context.Background(), "")
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if creds.Origin != OriginEnvironment {
		t.Fatalf("expected environment origin, got %v", creds.Origin)
	}
}

func TestAuthenticateGACPath(t *testing.T) {
	data := serviceAccountJSON(t)
	path := filepathWithTempFile(t, data)
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", path)

	m := NewManager(&memoryStorage{available: false})
	creds, err := m.Authenticate(context.Background(), "")
	if err != nil {
		t.Fatalf("Authenticate error: %v", err)
	}
	if creds.Origin != OriginKeyfile {
		t.Fatalf("expected keyfile origin, got %v", creds.Origin)
	}
}

func TestAuthenticateADCNotConfigured(t *testing.T) {
	t.Setenv("GPD_SERVICE_ACCOUNT_KEY", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
	m := NewManager(&memoryStorage{available: false})
	_, err := m.Authenticate(context.Background(), "")
	var apiErr *gpdErrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != gpdErrors.CodeAuthFailure {
		t.Fatalf("expected auth failure, got %v", err)
	}
}

func TestAuthenticateADC(t *testing.T) {
	data := serviceAccountJSON(t)
	path := filepathWithTempFile(t, data)
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", path)
	m := NewManager(&memoryStorage{available: false})
	creds, err := m.authenticateFromADC(context.Background(), []string{ScopeAndroidPublisher})
	if err != nil {
		t.Fatalf("authenticateFromADC error: %v", err)
	}
	if creds.Origin != OriginADC || creds.TokenSource == nil {
		t.Fatalf("unexpected credentials: %+v", creds)
	}
}

func TestAuthenticateFromJSONInvalid(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	if _, err := m.authenticateFromJSON(context.Background(), []byte("bad"), []string{"scope"}, OriginKeyfile, ""); err == nil {
		t.Fatalf("expected error for invalid json")
	}
	invalidType := map[string]string{"type": "user"}
	data, _ := json.Marshal(invalidType)
	if _, err := m.authenticateFromJSON(context.Background(), data, []string{"scope"}, OriginKeyfile, ""); err == nil {
		t.Fatalf("expected error for invalid type")
	}
}

func TestSetStoreTokensEmpty(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	m.SetStoreTokens("")
	if m.storeTokensMode != "auto" {
		t.Fatalf("expected storeTokensMode auto, got %s", m.storeTokensMode)
	}
}

func TestGetStatus(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	m.creds = &Credentials{
		TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: "token",
			Expiry:      time.Now().Add(time.Hour),
		}),
		Origin:  OriginKeyfile,
		Email:   "test@example.com",
		KeyPath: "/path",
	}
	status, err := m.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if !status.Authenticated || !status.TokenValid || status.TokenExpiry == "" {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestGetStatusTokenError(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	m.creds = &Credentials{
		TokenSource: errorTokenSource{},
		Origin:      OriginKeyfile,
	}
	status, err := m.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if status.TokenValid {
		t.Fatalf("expected invalid token")
	}
}

func TestGetStatusUnauthenticated(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	status, err := m.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if status.Authenticated {
		t.Fatalf("expected unauthenticated")
	}
}

func TestGetTokenSourceAndClear(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	if _, err := m.GetTokenSource(context.Background()); err != gpdErrors.ErrAuthNotConfigured {
		t.Fatalf("expected ErrAuthNotConfigured, got %v", err)
	}
	m.creds = &Credentials{
		TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)}),
	}
	if _, err := m.GetTokenSource(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m.Clear()
	if m.GetCredentials() != nil {
		t.Fatalf("expected credentials cleared")
	}
}

func TestSetActiveProfile(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		expected string
	}{
		{"set profile", "myprofile", "myprofile"},
		{"empty profile defaults", "", "default"},
		{"another profile", "prod", "prod"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(&memoryStorage{available: false})
			m.SetActiveProfile(tt.profile)
			if got := m.GetActiveProfile(); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestGetActiveProfileDefault(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	if got := m.GetActiveProfile(); got != "default" {
		t.Fatalf("expected default profile, got %q", got)
	}
}

func TestGetActiveProfileAfterSet(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	m.SetActiveProfile("custom")
	if got := m.GetActiveProfile(); got != "custom" {
		t.Fatalf("expected custom profile, got %q", got)
	}
}

func TestAuthenticateWithDeviceCodeMissingClientID(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	_, err := m.AuthenticateWithDeviceCode(context.Background(), "", "secret", []string{"scope"}, nil, true)
	var apiErr *gpdErrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != gpdErrors.CodeAuthFailure {
		t.Fatalf("expected auth failure, got %v", err)
	}
}

func TestAuthenticateWithDeviceCodeMissingScopes(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	_, err := m.AuthenticateWithDeviceCode(context.Background(), "client-id", "secret", []string{}, nil, true)
	var apiErr *gpdErrors.APIError
	if !errors.As(err, &apiErr) || apiErr.Code != gpdErrors.CodeAuthFailure {
		t.Fatalf("expected auth failure, got %v", err)
	}
}

func TestAuthenticateWithDeviceCodeRequestFails(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := m.AuthenticateWithDeviceCode(ctx, "client-id", "secret", []string{"scope"}, nil, true)
	if err == nil {
		t.Fatalf("expected error for canceled context")
	}
}

func TestAuthenticateFromJSONMissingClientEmail(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	var payload map[string]string
	if err := json.Unmarshal(serviceAccountJSON(t), &payload); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	delete(payload, "client_email")
	invalidJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}
	_, err = m.authenticateFromJSON(context.Background(), invalidJSON, []string{"scope"}, OriginKeyfile, "")
	if err == nil {
		t.Fatalf("expected error for missing client_email")
	}
}

func TestAuthenticateFromJSONMissingClientID(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	var payload map[string]string
	if err := json.Unmarshal(serviceAccountJSON(t), &payload); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	delete(payload, "client_id")
	invalidJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}
	_, err = m.authenticateFromJSON(context.Background(), invalidJSON, []string{"scope"}, OriginKeyfile, "")
	if err == nil {
		t.Fatalf("expected error for missing client_id")
	}
}

func TestAuthenticateFromJSONMissingPrivateKey(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	var payload map[string]string
	if err := json.Unmarshal(serviceAccountJSON(t), &payload); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	delete(payload, "private_key")
	invalidJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}
	_, err = m.authenticateFromJSON(context.Background(), invalidJSON, []string{"scope"}, OriginKeyfile, "")
	if err == nil {
		t.Fatalf("expected error for missing private_key")
	}
}

func TestAuthenticateFromJSONMissingTokenURI(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	var payload map[string]string
	if err := json.Unmarshal(serviceAccountJSON(t), &payload); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	delete(payload, "token_uri")
	invalidJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}
	_, err = m.authenticateFromJSON(context.Background(), invalidJSON, []string{"scope"}, OriginKeyfile, "")
	if err == nil {
		t.Fatalf("expected error for missing token_uri")
	}
}

func TestCredentialOriginOAuth(t *testing.T) {
	if OriginOAuth.String() != "oauth" {
		t.Fatalf("expected oauth origin string")
	}
}

func TestGetActiveProfileEmptyAfterSet(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	m.SetActiveProfile("custom")
	m.SetActiveProfile("")
	if got := m.GetActiveProfile(); got != "default" {
		t.Fatalf("expected default profile after setting empty, got %q", got)
	}
}

func TestWrapTokenSourceWithoutStorage(t *testing.T) {
	m := NewManager(&memoryStorage{available: false})
	m.SetStoreTokens("never")
	baseTokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "token",
		Expiry:      time.Now().Add(time.Hour),
	})
	wrapped := m.wrapTokenSource(baseTokenSource, OriginKeyfile, "test@example.com", "client123", []string{"scope"})
	if wrapped == nil {
		t.Fatalf("expected wrapped token source")
	}
	token, err := wrapped.Token()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "token" {
		t.Fatalf("expected token to be preserved")
	}
}
