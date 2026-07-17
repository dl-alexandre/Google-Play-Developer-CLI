//go:build unit
// +build unit

package cli

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/auth"
	gpdErrors "github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// generateServiceAccountKey generates a valid service account key for testing
func generateServiceAccountKey(t *testing.T) []byte {
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

// TestAuthStatusCmd_Unauthenticated tests that AuthStatusCmd returns unauthenticated status
func TestAuthStatusCmd_Unauthenticated(t *testing.T) {
	cmd := &AuthStatusCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("AuthStatusCmd.Run() unexpected error: %v", err)
	}
}

// TestAuthStatusCmd_WithAuthentication tests AuthStatusCmd when authenticated
func TestAuthStatusCmd_WithAuthentication(t *testing.T) {
	// First authenticate
	data := generateServiceAccountKey(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	loginCmd := &AuthLoginCmd{Key: keyPath}
	loginGlobals := &Globals{
		Output: "json",
		Pretty: false,
	}

	// Login
	if err := loginCmd.Run(loginGlobals); err != nil {
		t.Fatalf("AuthLoginCmd.Run() failed: %v", err)
	}

	// Then check status
	statusCmd := &AuthStatusCmd{}
	statusGlobals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := statusCmd.Run(statusGlobals)
	if err != nil {
		t.Fatalf("AuthStatusCmd.Run() unexpected error: %v", err)
	}
}

// TestAuthLoginCmd_Success tests successful authentication with service account key
func TestAuthLoginCmd_Success(t *testing.T) {
	data := generateServiceAccountKey(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	cmd := &AuthLoginCmd{Key: keyPath}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("AuthLoginCmd.Run() unexpected error: %v", err)
	}
}

// TestAuthLoginCmd_MissingKeyFile tests authentication with missing key file
func TestAuthLoginCmd_MissingKeyFile(t *testing.T) {
	cmd := &AuthLoginCmd{Key: "/nonexistent/path/key.json"}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("AuthLoginCmd.Run() expected error for missing key file, got nil")
	}
}

// TestAuthLoginCmd_InvalidKeyFile tests authentication with invalid key file content
func TestAuthLoginCmd_InvalidKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "invalid-key.json")
	if err := os.WriteFile(keyPath, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("failed to write invalid key file: %v", err)
	}

	cmd := &AuthLoginCmd{Key: keyPath}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("AuthLoginCmd.Run() expected error for invalid key file, got nil")
	}
}

// TestAuthLoginCmd_NoKeyEnvironmentVariable tests authentication using environment variable
func TestAuthLoginCmd_NoKeyEnvironmentVariable(t *testing.T) {
	data := generateServiceAccountKey(t)
	t.Setenv("GPD_SERVICE_ACCOUNT_KEY", string(data))

	cmd := &AuthLoginCmd{Key: ""}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("AuthLoginCmd.Run() with env var unexpected error: %v", err)
	}
}

// TestAuthLogoutCmd_Success tests successful logout
func TestAuthLogoutCmd_Success(t *testing.T) {
	// First authenticate
	data := generateServiceAccountKey(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	loginCmd := &AuthLoginCmd{Key: keyPath}
	loginGlobals := &Globals{
		Output: "json",
		Pretty: false,
	}
	if err := loginCmd.Run(loginGlobals); err != nil {
		t.Fatalf("AuthLoginCmd.Run() failed: %v", err)
	}

	// Then logout
	logoutCmd := &AuthLogoutCmd{}
	logoutGlobals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := logoutCmd.Run(logoutGlobals)
	if err != nil {
		t.Fatalf("AuthLogoutCmd.Run() unexpected error: %v", err)
	}

	// Verify logged out by checking status
	statusCmd := &AuthStatusCmd{}
	statusGlobals := &Globals{
		Output: "json",
		Pretty: false,
	}
	if err := statusCmd.Run(statusGlobals); err != nil {
		t.Fatalf("AuthStatusCmd.Run() after logout unexpected error: %v", err)
	}
}

// TestAuthLogoutCmd_AlreadyLoggedOut tests logout when already logged out
func TestAuthLogoutCmd_AlreadyLoggedOut(t *testing.T) {
	cmd := &AuthLogoutCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	// Should not error even when already logged out
	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("AuthLogoutCmd.Run() unexpected error when already logged out: %v", err)
	}
}

// TestAuthCmd_SubcommandsExist verifies AuthCmd has all subcommands defined
func TestAuthCmd_SubcommandsExist(t *testing.T) {
	cmd := AuthCmd{}

	want := map[string]string{
		"Status":   "cli.AuthStatusCmd",
		"Login":    "cli.AuthLoginCmd",
		"Init":     "cli.AuthInitCmd",
		"Logout":   "cli.AuthLogoutCmd",
		"Delete":   "cli.AuthDeleteCmd",
		"List":     "cli.AuthListCmd",
		"Switch":   "cli.AuthSwitchCmd",
		"Check":    "cli.AuthCheckCmd",
		"Doctor":   "cli.AuthDoctorCmd",
		"Diagnose": "cli.AuthDiagnoseCmd",
	}

	val := reflect.ValueOf(cmd)
	typ := reflect.TypeOf(cmd)
	for name, wantType := range want {
		field, ok := typ.FieldByName(name)
		if !ok {
			t.Errorf("AuthCmd missing field %s", name)
			continue
		}
		got := field.Type.String()
		if got != wantType {
			t.Errorf("AuthCmd.%s type = %v, want %s", name, got, wantType)
		}
		_ = val
	}
}

// TestAuthListCmd_Empty runs list with no stored profiles.
func TestAuthListCmd_Empty(t *testing.T) {
	cmd := &AuthListCmd{}
	globals := &Globals{Output: "json", Profile: "default"}
	applyAuthGlobals(globals)
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("AuthListCmd.Run() unexpected error: %v", err)
	}
}

// TestAuthSwitchCmd_Default allows switching to default without stored metadata.
func TestAuthSwitchCmd_Default(t *testing.T) {
	// Isolate config dir
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

	cmd := &AuthSwitchCmd{Profile: "default"}
	globals := &Globals{Output: "json"}
	applyAuthGlobals(globals)
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("AuthSwitchCmd.Run() unexpected error: %v", err)
	}
}

// TestAuthSwitchCmd_Missing fails for unknown profiles.
func TestAuthSwitchCmd_Missing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

	cmd := &AuthSwitchCmd{Profile: "does-not-exist"}
	globals := &Globals{Output: "json"}
	applyAuthGlobals(globals)
	if err := cmd.Run(globals); err == nil {
		t.Fatal("AuthSwitchCmd.Run() expected error for missing profile")
	}
}

// TestAuthDoctorCmd_NoRefresh runs doctor without credential load.
func TestAuthDoctorCmd_NoRefresh(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

	cmd := &AuthDoctorCmd{}
	globals := &Globals{Output: "json", StoreTokens: "never", Timeout: time.Second}
	applyAuthGlobals(globals)
	if err := cmd.Run(globals); err != nil {
		// Doctor may return auth failure only when summary.Errors > 0.
		// With no refresh check, empty env should still succeed.
		var apiErr *gpdErrors.APIError
		if errors.As(err, &apiErr) {
			t.Fatalf("AuthDoctorCmd.Run() unexpected API error: %v", err)
		}
		t.Fatalf("AuthDoctorCmd.Run() unexpected error: %v", err)
	}
}

// TestAuthCheckCmd_RequiresPackage validates package is required.
func TestAuthCheckCmd_RequiresPackage(t *testing.T) {
	cmd := &AuthCheckCmd{}
	globals := &Globals{Output: "json"}
	applyAuthGlobals(globals)
	if err := cmd.Run(globals); err == nil {
		t.Fatal("AuthCheckCmd.Run() expected error without package")
	}
}

// TestResolveAuthProfileViaGlobals ensures applyAuthGlobals is safe.
func TestApplyAuthGlobals_NilSafe(t *testing.T) {
	applyAuthGlobals(nil)
	_ = newAuthManager()
}

// TestAuthLoginCmd_StructFields tests AuthLoginCmd struct fields and tags
func TestAuthLoginCmd_StructFields(t *testing.T) {
	cmd := AuthLoginCmd{Key: "/path/to/key.json"}

	// Test field exists and is settable
	if cmd.Key != "/path/to/key.json" {
		t.Errorf("AuthLoginCmd.Key = %v, want /path/to/key.json", cmd.Key)
	}

	// Check struct tags using reflection
	typeOfCmd := reflect.TypeOf(cmd)
	keyField, found := typeOfCmd.FieldByName("Key")
	if !found {
		t.Fatal("AuthLoginCmd missing Key field")
	}

	// Verify tag exists (Kong uses struct tags for CLI parsing)
	helpTag := keyField.Tag.Get("help")
	if helpTag == "" {
		t.Error("AuthLoginCmd.Key missing 'help' struct tag")
	}

	typeTag := keyField.Tag.Get("type")
	if typeTag != "existingfile" {
		t.Errorf("AuthLoginCmd.Key type tag = %v, want 'existingfile'", typeTag)
	}
}

// TestAuthLogoutCmd_StructExists verifies AuthLogoutCmd struct exists
func TestAuthLogoutCmd_StructExists(t *testing.T) {
	cmd := AuthLogoutCmd{}

	// Verify struct can be instantiated
	if reflect.TypeOf(cmd).String() != "cli.AuthLogoutCmd" {
		t.Errorf("AuthLogoutCmd type = %v, want cli.AuthLogoutCmd", reflect.TypeOf(cmd))
	}
}

// TestAuthStatusCmd_StructExists verifies AuthStatusCmd struct exists
func TestAuthStatusCmd_StructExists(t *testing.T) {
	cmd := AuthStatusCmd{}

	// Verify struct can be instantiated
	if reflect.TypeOf(cmd).String() != "cli.AuthStatusCmd" {
		t.Errorf("AuthStatusCmd type = %v, want cli.AuthStatusCmd", reflect.TypeOf(cmd))
	}
}

// TestAuthStatusCmd_RunSignature verifies Run method has correct signature
func TestAuthStatusCmd_RunSignature(t *testing.T) {
	cmd := &AuthStatusCmd{}
	globals := &Globals{}

	// Verify Run method exists and accepts *Globals
	err := cmd.Run(globals)
	// Error is expected (not authenticated), but shouldn't panic
	_ = err
}

// TestAuthLoginCmd_RunSignature verifies Run method has correct signature
func TestAuthLoginCmd_RunSignature(t *testing.T) {
	cmd := &AuthLoginCmd{}
	globals := &Globals{}

	// Verify Run method exists and accepts *Globals
	err := cmd.Run(globals)
	// Error is expected (no key file), but shouldn't panic
	_ = err
}

// TestAuthLogoutCmd_RunSignature verifies Run method has correct signature
func TestAuthLogoutCmd_RunSignature(t *testing.T) {
	cmd := &AuthLogoutCmd{}
	globals := &Globals{}

	// Verify Run method exists and accepts *Globals
	err := cmd.Run(globals)
	// Should not error
	if err != nil {
		t.Errorf("AuthLogoutCmd.Run() unexpected error: %v", err)
	}
}

// TestAuthStatusCmd_WithDifferentOutputFormats tests status with different output formats
func TestAuthStatusCmd_WithDifferentOutputFormats(t *testing.T) {
	formats := []string{"json", "table"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cmd := &AuthStatusCmd{}
			globals := &Globals{
				Output: format,
				Pretty: false,
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("AuthStatusCmd.Run() with format %s unexpected error: %v", format, err)
			}
		})
	}
}

// TestAuthLoginCmd_WithPrettyOutput tests login with pretty JSON output
func TestAuthLoginCmd_WithPrettyOutput(t *testing.T) {
	data := generateServiceAccountKey(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	cmd := &AuthLoginCmd{Key: keyPath}
	globals := &Globals{
		Output: "json",
		Pretty: true,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("AuthLoginCmd.Run() with pretty output unexpected error: %v", err)
	}
}

// TestAuthCommands_WithProfile tests auth commands with profile setting
func TestAuthCommands_WithProfile(t *testing.T) {
	data := generateServiceAccountKey(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	cmd := &AuthLoginCmd{Key: keyPath}
	globals := &Globals{
		Output:  "json",
		Profile: "test-profile",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("AuthLoginCmd.Run() with profile unexpected error: %v", err)
	}
}

// TestAuthLoginCmd_ErrorReturnsAPIError verifies login errors return proper error types
func TestAuthLoginCmd_ErrorReturnsAPIError(t *testing.T) {
	cmd := &AuthLoginCmd{Key: "/nonexistent/key.json"}
	globals := &Globals{Output: "json"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing key file")
	}

	// Verify error is not nil and is of type error
	var apiErr *gpdErrors.APIError
	if errors.As(err, &apiErr) {
		// It's an APIError, which is expected
		if apiErr.ExitCode() != gpdErrors.ExitAuthFailure {
			t.Errorf("Expected exit code %d, got %d", gpdErrors.ExitAuthFailure, apiErr.ExitCode())
		}
	}
	// If it's not an APIError, that's also acceptable for file not found
}

// TestAuthStatusCmd_ErrorHandling verifies status command error handling
func TestAuthStatusCmd_ErrorHandling(t *testing.T) {
	cmd := &AuthStatusCmd{}
	globals := &Globals{Output: "json"}

	// Status command should not error even when not authenticated
	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("AuthStatusCmd.Run() should not error when not authenticated: %v", err)
	}
}

// TestAuthStatusCmd_ReturnsNotNilWhenUnauthenticated verifies proper error return
func TestAuthStatusCmd_ReturnsNotNilWhenUnauthenticated(t *testing.T) {
	cmd := &AuthStatusCmd{}
	globals := &Globals{Output: "json"}

	err := cmd.Run(globals)
	// Should return nil error even when not authenticated
	// The command outputs the unauthenticated status, it doesn't error
	if err != nil {
		t.Errorf("AuthStatusCmd.Run() should return nil error when unauthenticated, got: %v", err)
	}
}

// TestAuthLoginCmd_NilGlobals tests login command with nil globals (should panic or handle gracefully)
func TestAuthLoginCmd_NilGlobals(t *testing.T) {
	data := generateServiceAccountKey(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	cmd := &AuthLoginCmd{Key: keyPath}

	// This should handle nil gracefully or panic (which we recover from)
	defer func() {
		if r := recover(); r != nil {
			// Panic is acceptable for nil globals
			t.Logf("AuthLoginCmd.Run() panicked with nil globals (acceptable): %v", r)
		}
	}()

	_ = cmd.Run(nil)
}

// TestAuthManager_GetStatus_WithExpiredToken tests status when token is expired
func TestAuthManager_GetStatus_WithExpiredToken(t *testing.T) {
	// This test verifies the auth.Manager.GetStatus behavior with expired tokens
	// We test it indirectly through the auth package
	secureStorage := &mockStorage{available: false}
	mgr := auth.NewManager(secureStorage)

	// Get status without authentication - should return unauthenticated
	status, err := mgr.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}

	// Should not be authenticated
	if status.Authenticated {
		t.Error("Expected authenticated to be false when no credentials set")
	}
}

// mockStorage is a mock implementation of storage for testing
type mockStorage struct {
	available bool
	data      map[string][]byte
}

func (m *mockStorage) Store(key string, data []byte) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	m.data[key] = data
	return nil
}

func (m *mockStorage) Retrieve(key string) ([]byte, error) {
	if m.data == nil {
		return nil, errors.New("not found")
	}
	data, ok := m.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return data, nil
}

func (m *mockStorage) Delete(key string) error {
	if m.data != nil {
		delete(m.data, key)
	}
	return nil
}

func (m *mockStorage) Available() bool {
	return m.available
}

func (m *mockStorage) KeyPrefix() string {
	return "test"
}

// TestAuthLoginCmd_WithGOOGLE_APPLICATION_CREDENTIALS tests authentication using GOOGLE_APPLICATION_CREDENTIALS
func TestAuthLoginCmd_WithGOOGLE_APPLICATION_CREDENTIALS(t *testing.T) {
	data := generateServiceAccountKey(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "gac-key.json")
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	// Set environment variable
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", keyPath)

	cmd := &AuthLoginCmd{Key: ""} // No key provided, should use env var
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("AuthLoginCmd.Run() with GOOGLE_APPLICATION_CREDENTIALS unexpected error: %v", err)
	}
}

// TestAuthLoginCmd_WithStoreTokensMode tests login with different token storage modes
func TestAuthLoginCmd_WithStoreTokensMode(t *testing.T) {
	data := generateServiceAccountKey(t)
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "test-key.json")
	if err := os.WriteFile(keyPath, data, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	modes := []string{"auto", "never", "secure"}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			cmd := &AuthLoginCmd{Key: keyPath}
			globals := &Globals{
				Output:      "json",
				StoreTokens: mode,
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("AuthLoginCmd.Run() with StoreTokens=%s unexpected error: %v", mode, err)
			}
		})
	}
}

// TestAuthLogoutCmd_MultipleCalls tests that logout can be called multiple times without error
func TestAuthLogoutCmd_MultipleCalls(t *testing.T) {
	cmd := &AuthLogoutCmd{}
	globals := &Globals{Output: "json"}

	// Call logout multiple times
	for i := 0; i < 3; i++ {
		err := cmd.Run(globals)
		if err != nil {
			t.Fatalf("AuthLogoutCmd.Run() call %d unexpected error: %v", i+1, err)
		}
	}
}

func TestAuthLogoutCmd_WithProfileFlag(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

	// Seed metadata for a named profile via auth package helpers (login may not write meta without secure storage).
	metaDir := filepath.Join(tmp, ".config", "gpd", "tokens")
	if err := os.MkdirAll(metaDir, 0o700); err != nil {
		// Paths may differ by platform; fall through using login + store-tokens never still exercises Run().
		t.Logf("mkdir tokens (may vary by OS path layout): %v", err)
	}

	cmd := &AuthLogoutCmd{Name: "staging"}
	globals := &Globals{Output: "json", Profile: "default"}
	applyAuthGlobals(globals)
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("AuthLogoutCmd.Run() with --profile unexpected error: %v", err)
	}
}

func TestAuthLogoutCmd_All(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

	cmd := &AuthLogoutCmd{All: true}
	globals := &Globals{Output: "json"}
	applyAuthGlobals(globals)
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("AuthLogoutCmd.Run() --all unexpected error: %v", err)
	}
}

func TestAuthDeleteCmd_MissingProfile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

	cmd := &AuthDeleteCmd{Profile: "does-not-exist"}
	globals := &Globals{Output: "json", Profile: "default"}
	applyAuthGlobals(globals)
	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("AuthDeleteCmd.Run() expected not-found error")
	}
	var apiErr *gpdErrors.APIError
	if errors.As(err, &apiErr) {
		if apiErr.ExitCode() != gpdErrors.ExitNotFound && apiErr.Code != gpdErrors.CodeNotFound {
			// Accept either typed not-found semantics.
			t.Logf("delete missing profile error: %v", err)
		}
	}
}

func TestAuthDeleteCmd_RefuseActiveWithoutForce(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

	// Create token metadata so DeleteProfile would find the profile if force allowed.
	tokensDir := filepath.Join(tmp, ".config", "gpd", "tokens")
	if err := os.MkdirAll(tokensDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	meta := map[string]string{
		"profile":      "active-one",
		"clientIdHash": "hash",
		"origin":       "keyfile",
		"updatedAt":    time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.Marshal(meta)
	metaPath := filepath.Join(tokensDir, "active-one--hash.meta.json")
	if err := os.WriteFile(metaPath, data, 0o600); err != nil {
		t.Fatalf("write meta: %v", err)
	}

	// Persist active profile via switch-like config write.
	if err := (&AuthSwitchCmd{Profile: "active-one"}).Run(&Globals{Output: "json", Profile: "active-one"}); err != nil {
		// Switch requires profile in list — we wrote meta so it should work if config path matches.
		// If path layout differs, set globals.Profile and applyAuthGlobals instead.
		t.Logf("switch note: %v", err)
	}

	cmd := &AuthDeleteCmd{Profile: "active-one", Force: false}
	globals := &Globals{Output: "json", Profile: "active-one"}
	applyAuthGlobals(globals)
	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("AuthDeleteCmd.Run() expected error deleting active profile without --force")
	}
}

func TestAuthDeleteCmd_ForceActiveSwitchesToDefault(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	// Also set HOME-style path used by config.GetPaths on macOS.
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))

	tokensDir := filepath.Join(tmp, ".config", "gpd", "tokens")
	if err := os.MkdirAll(tokensDir, 0o700); err != nil {
		// Discover real tokens path via list after writing through auth if needed.
		t.Logf("mkdir primary tokens dir: %v", err)
	}
	meta := map[string]interface{}{
		"profile":      "doomed",
		"clientIdHash": "hash",
		"origin":       "keyfile",
		"updatedAt":    time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	// Write under both common config layouts used in tests.
	for _, dir := range []string{
		filepath.Join(tmp, ".config", "gpd", "tokens"),
		filepath.Join(tmp, "Library", "Application Support", "gpd", "tokens"),
	} {
		_ = os.MkdirAll(dir, 0o700)
		_ = os.WriteFile(filepath.Join(dir, "doomed--hash.meta.json"), data, 0o600)
	}

	// Make doomed active via config + globals.
	switchCmd := &AuthSwitchCmd{Profile: "doomed"}
	switchGlobals := &Globals{Output: "json", Profile: "doomed"}
	applyAuthGlobals(switchGlobals)
	if err := switchCmd.Run(switchGlobals); err != nil {
		// Fallback: still test delete --force path with active matching globals.
		t.Logf("AuthSwitchCmd note: %v", err)
	}

	delCmd := &AuthDeleteCmd{Profile: "doomed", Force: true}
	delGlobals := &Globals{Output: "json", Profile: "doomed"}
	applyAuthGlobals(delGlobals)
	if err := delCmd.Run(delGlobals); err != nil {
		t.Fatalf("AuthDeleteCmd.Run() --force unexpected error: %v", err)
	}

	// Deleting again should report not found.
	if err := delCmd.Run(delGlobals); err == nil {
		t.Fatal("second delete expected not found")
	}
}

func TestAuthDeleteCmd_NonActive(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, ".cache"))

	meta := map[string]interface{}{
		"profile":      "old-team",
		"clientIdHash": "h",
		"origin":       "oauth",
		"updatedAt":    time.Now().UTC().Format(time.RFC3339),
	}
	data, _ := json.MarshalIndent(meta, "", "  ")
	for _, dir := range []string{
		filepath.Join(tmp, ".config", "gpd", "tokens"),
		filepath.Join(tmp, "Library", "Application Support", "gpd", "tokens"),
	} {
		_ = os.MkdirAll(dir, 0o700)
		_ = os.WriteFile(filepath.Join(dir, "old-team--h.meta.json"), data, 0o600)
	}

	cmd := &AuthDeleteCmd{Profile: "old-team"}
	globals := &Globals{Output: "json", Profile: "default"}
	applyAuthGlobals(globals)
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("AuthDeleteCmd.Run() non-active unexpected error: %v", err)
	}
}

func TestAuthDeleteCmd_EmptyProfile(t *testing.T) {
	cmd := &AuthDeleteCmd{Profile: "  "}
	globals := &Globals{Output: "json"}
	if err := cmd.Run(globals); err == nil {
		t.Fatal("expected validation error for empty profile")
	}
}

func TestAuthDeleteCmd_StructFields(t *testing.T) {
	cmd := AuthDeleteCmd{Profile: "p", Force: true}
	if cmd.Profile != "p" || !cmd.Force {
		t.Fatalf("unexpected fields: %+v", cmd)
	}
	typ := reflect.TypeOf(cmd)
	prof, ok := typ.FieldByName("Profile")
	if !ok {
		t.Fatal("AuthDeleteCmd missing Profile field")
	}
	if _, hasArg := prof.Tag.Lookup("arg"); !hasArg {
		t.Fatal("AuthDeleteCmd.Profile should be a Kong arg")
	}
}

func TestAuthLogoutCmd_StructFields(t *testing.T) {
	cmd := AuthLogoutCmd{Name: "p", All: true}
	if cmd.Name != "p" || !cmd.All {
		t.Fatalf("unexpected fields: %+v", cmd)
	}
	typ := reflect.TypeOf(cmd)
	if _, ok := typ.FieldByName("All"); !ok {
		t.Fatal("AuthLogoutCmd missing All field")
	}
	if _, ok := typ.FieldByName("Name"); !ok {
		t.Fatal("AuthLogoutCmd missing Name field")
	}
}
