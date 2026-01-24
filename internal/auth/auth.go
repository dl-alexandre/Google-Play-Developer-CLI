// Package auth provides authentication management for gpd.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
)

// OAuth scopes required for gpd operations.
const (
	// ScopeAndroidPublisher is the scope for Android Publisher API
	// Used for: publish, reviews, monetization, purchases
	ScopeAndroidPublisher = "https://www.googleapis.com/auth/androidpublisher"

	// ScopePlayReporting is the scope for Play Developer Reporting API
	// Used for: analytics, vitals
	ScopePlayReporting = "https://www.googleapis.com/auth/playdeveloperreporting"
)

// CredentialOrigin indicates where credentials were obtained from.
type CredentialOrigin int

const (
	OriginADC CredentialOrigin = iota
	OriginKeyfile
	OriginEnvironment
)

func (o CredentialOrigin) String() string {
	switch o {
	case OriginADC:
		return "adc"
	case OriginKeyfile:
		return "keyfile"
	case OriginEnvironment:
		return "environment"
	default:
		return "unknown"
	}
}

// Credentials holds the authenticated credentials.
type Credentials struct {
	TokenSource oauth2.TokenSource
	Origin      CredentialOrigin
	KeyPath     string // Only for keyfile origin
	Email       string // Service account email
}

// Manager handles authentication operations.
type Manager struct {
	creds   *Credentials
	mu      sync.Mutex
	storage SecureStorage
}

// NewManager creates a new authentication manager.
func NewManager(storage SecureStorage) *Manager {
	return &Manager{
		storage: storage,
	}
}

// Authenticate attempts to obtain credentials from various sources.
func (m *Manager) Authenticate(ctx context.Context, keyPath string) (*Credentials, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	scopes := []string{ScopeAndroidPublisher, ScopePlayReporting}

	// Priority 1: Explicit key path
	if keyPath != "" {
		creds, err := m.authenticateFromKeyfile(ctx, keyPath, scopes)
		if err != nil {
			return nil, err
		}
		m.creds = creds
		return creds, nil
	}

	// Priority 2: Environment variable
	if envKey := config.GetEnvServiceAccountKey(); envKey != "" {
		creds, err := m.authenticateFromJSON(ctx, []byte(envKey), scopes)
		if err != nil {
			return nil, err
		}
		creds.Origin = OriginEnvironment
		m.creds = creds
		return creds, nil
	}

	// Priority 3: GOOGLE_APPLICATION_CREDENTIALS
	if gacPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); gacPath != "" {
		creds, err := m.authenticateFromKeyfile(ctx, gacPath, scopes)
		if err != nil {
			return nil, err
		}
		m.creds = creds
		return creds, nil
	}

	// Priority 4: Application Default Credentials
	creds, err := m.authenticateFromADC(ctx, scopes)
	if err != nil {
		return nil, errors.ErrAuthNotConfigured
	}
	m.creds = creds
	return creds, nil
}

func (m *Manager) authenticateFromKeyfile(ctx context.Context, path string, scopes []string) (*Credentials, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeAuthFailure, fmt.Sprintf("failed to read key file: %v", err)).
			WithHint("Check that the service account key file exists and is readable")
	}

	creds, err := m.authenticateFromJSON(ctx, data, scopes)
	if err != nil {
		return nil, err
	}
	creds.Origin = OriginKeyfile
	creds.KeyPath = path
	return creds, nil
}

func (m *Manager) authenticateFromJSON(ctx context.Context, jsonKey []byte, scopes []string) (*Credentials, error) {
	// Validate JSON structure
	var keyData struct {
		Type                    string `json:"type"`
		ProjectID               string `json:"project_id"`
		PrivateKeyID            string `json:"private_key_id"`
		PrivateKey              string `json:"private_key"`
		ClientEmail             string `json:"client_email"`
		ClientID                string `json:"client_id"`
		AuthURI                 string `json:"auth_uri"`
		TokenURI                string `json:"token_uri"`
		AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
		ClientX509CertURL       string `json:"client_x509_cert_url"`
	}
	if err := json.Unmarshal(jsonKey, &keyData); err != nil {
		return nil, errors.ErrServiceAccountInvalid.WithDetails(err.Error())
	}

	if keyData.Type != "service_account" {
		return nil, errors.NewAPIError(errors.CodeAuthFailure, "invalid credential type").
			WithHint("Credential file must be a service account key, not OAuth credentials")
	}

	jwtConfig, err := google.JWTConfigFromJSON(jsonKey, scopes...)
	if err != nil {
		return nil, errors.ErrServiceAccountInvalid.WithDetails(err.Error())
	}

	// Wrap token source with early refresh
	baseTokenSource := jwtConfig.TokenSource(ctx)
	wrappedTokenSource := &EarlyRefreshTokenSource{
		base:          baseTokenSource,
		refreshLeeway: 300 * time.Second, // 5 minutes before expiry
		clockSkew:     30 * time.Second,
	}

	return &Credentials{
		TokenSource: wrappedTokenSource,
		Origin:      OriginKeyfile,
		Email:       keyData.ClientEmail,
	}, nil
}

func (m *Manager) authenticateFromADC(ctx context.Context, scopes []string) (*Credentials, error) {
	creds, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeAuthFailure, "failed to find default credentials").
			WithHint("Set GOOGLE_APPLICATION_CREDENTIALS or configure Application Default Credentials")
	}

	// Wrap token source with early refresh
	wrappedTokenSource := &EarlyRefreshTokenSource{
		base:          creds.TokenSource,
		refreshLeeway: 300 * time.Second,
		clockSkew:     30 * time.Second,
	}

	return &Credentials{
		TokenSource: wrappedTokenSource,
		Origin:      OriginADC,
	}, nil
}

// GetTokenSource returns the current token source.
func (m *Manager) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.creds == nil {
		return nil, errors.ErrAuthNotConfigured
	}

	return m.creds.TokenSource, nil
}

// GetCredentials returns the current credentials.
func (m *Manager) GetCredentials() *Credentials {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.creds
}

// Clear clears the current credentials.
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.creds = nil
}

// EarlyRefreshTokenSource wraps a token source to refresh tokens early.
type EarlyRefreshTokenSource struct {
	base          oauth2.TokenSource
	refreshLeeway time.Duration
	clockSkew     time.Duration
	mu            sync.Mutex
	cachedToken   *oauth2.Token
}

// Token returns a token, refreshing early if needed.
func (s *EarlyRefreshTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if we have a valid cached token
	if s.cachedToken != nil && s.cachedToken.Valid() {
		// Check if we should refresh early
		expiryWithLeeway := s.cachedToken.Expiry.Add(-s.refreshLeeway).Add(-s.clockSkew)
		if time.Now().Before(expiryWithLeeway) {
			return s.cachedToken, nil
		}
	}

	// Get a new token
	token, err := s.base.Token()
	if err != nil {
		return nil, err
	}

	s.cachedToken = token
	return token, nil
}

// SecureStorage interface for platform-specific credential storage.
type SecureStorage interface {
	Store(key string, value []byte) error
	Retrieve(key string) ([]byte, error)
	Delete(key string) error
	Available() bool
}

// PermissionCheck represents a permission validation result.
type PermissionCheck struct {
	Surface   string `json:"surface"`
	HasAccess bool   `json:"hasAccess"`
	Error     string `json:"error,omitempty"`
	TestCall  string `json:"testCall"`
}

// CheckResult contains the results of permission validation.
type CheckResult struct {
	Valid       bool               `json:"valid"`
	Origin      string             `json:"origin"`
	Email       string             `json:"email,omitempty"`
	Permissions []*PermissionCheck `json:"permissions"`
}

// Status represents the current authentication status.
type Status struct {
	Authenticated bool   `json:"authenticated"`
	Origin        string `json:"origin,omitempty"`
	Email         string `json:"email,omitempty"`
	KeyPath       string `json:"keyPath,omitempty"`
	TokenValid    bool   `json:"tokenValid"`
	TokenExpiry   string `json:"tokenExpiry,omitempty"`
}

// GetStatus returns the current authentication status.
func (m *Manager) GetStatus(ctx context.Context) (*Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.creds == nil {
		return &Status{
			Authenticated: false,
		}, nil
	}

	status := &Status{
		Authenticated: true,
		Origin:        m.creds.Origin.String(),
		Email:         m.creds.Email,
		KeyPath:       m.creds.KeyPath,
	}

	// Check token validity
	token, err := m.creds.TokenSource.Token()
	if err != nil {
		status.TokenValid = false
	} else {
		status.TokenValid = token.Valid()
		if !token.Expiry.IsZero() {
			status.TokenExpiry = token.Expiry.Format(time.RFC3339)
		}
	}

	return status, nil
}
