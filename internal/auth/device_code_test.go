package auth

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestNewDeviceCodeFlow(t *testing.T) {
	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		Scopes:       []string{"scope1", "scope2"},
	}

	flow := NewDeviceCodeFlow(config)
	if flow == nil {
		t.Fatal("expected non-nil flow")
	}
	if flow.config != config {
		t.Fatalf("expected config to be set")
	}
	if flow.response != nil {
		t.Fatalf("expected response to be nil initially")
	}
	if !flow.openBrowser {
		t.Fatalf("expected openBrowser to be true by default")
	}
}

func TestNewDeviceCodeFlowWithBrowserDisabled(t *testing.T) {
	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		Scopes:       []string{"scope1", "scope2"},
	}

	flow := NewDeviceCodeFlow(config, WithBrowserOpen(false))
	if flow.openBrowser {
		t.Fatalf("expected openBrowser to be false")
	}
}

func TestRequestDeviceCodeContextCanceled(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client",
		Scopes:   []string{"scope1"},
	}
	flow := NewDeviceCodeFlow(config)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := flow.RequestDeviceCode(ctx)
	if err == nil {
		t.Fatalf("expected error for canceled context")
	}
}

func TestPollForTokenWithoutDeviceCode(t *testing.T) {
	config := &oauth2.Config{ClientID: "test"}
	flow := NewDeviceCodeFlow(config)

	_, err := flow.PollForToken(context.Background())
	if err == nil || !strings.Contains(err.Error(), "device code not requested") {
		t.Fatalf("expected device code not requested error, got %v", err)
	}
}

func TestPollForTokenExpired(t *testing.T) {
	config := &oauth2.Config{ClientID: "test"}
	flow := NewDeviceCodeFlow(config)
	flow.response = &DeviceCodeResponse{
		DeviceCode: "device123",
		ExpiresIn:  -1,
		Interval:   1,
	}

	_, err := flow.PollForToken(context.Background())
	if err == nil || !strings.Contains(err.Error(), "expired") {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestPollForTokenContextCanceled(t *testing.T) {
	config := &oauth2.Config{ClientID: "test"}
	flow := NewDeviceCodeFlow(config)
	flow.response = &DeviceCodeResponse{
		DeviceCode: "device123",
		ExpiresIn:  3600,
		Interval:   1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := flow.PollForToken(ctx)
	if err == nil || err != context.Canceled {
		t.Fatalf("expected context canceled error, got %v", err)
	}
}

func TestDisplayDeviceCodePromptWithVerificationURL(t *testing.T) {
	response := &DeviceCodeResponse{
		UserCode:        "USER-CODE",
		VerificationURL: "https://example.com/auth",
	}
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, response, false)
	output := buf.String()
	if !strings.Contains(output, "USER-CODE") {
		t.Fatalf("expected USER-CODE in output, got %q", output)
	}
	if !strings.Contains(output, "https://example.com/auth") {
		t.Fatalf("expected verification URL in output, got %q", output)
	}
}

func TestDisplayDeviceCodePromptWithVerificationURI(t *testing.T) {
	response := &DeviceCodeResponse{
		UserCode:        "USER-CODE",
		VerificationURI: "https://example.com/auth",
	}
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, response, false)
	output := buf.String()
	if !strings.Contains(output, "USER-CODE") {
		t.Fatalf("expected USER-CODE in output, got %q", output)
	}
	if !strings.Contains(output, "https://example.com/auth") {
		t.Fatalf("expected verification URI in output, got %q", output)
	}
}

func TestDisplayDeviceCodePromptWithCompleteURL(t *testing.T) {
	response := &DeviceCodeResponse{
		UserCode:                "USER-CODE",
		VerificationURL:         "https://example.com/auth",
		VerificationURLComplete: "https://example.com/auth?code=USER-CODE",
	}
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, response, false)
	output := buf.String()
	if !strings.Contains(output, "Or visit:") {
		t.Fatalf("expected 'Or visit:' in output, got %q", output)
	}
	if !strings.Contains(output, "https://example.com/auth?code=USER-CODE") {
		t.Fatalf("expected complete URL in output, got %q", output)
	}
}

func TestDisplayDeviceCodePromptNilResponse(t *testing.T) {
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, nil, false)
	output := buf.String()
	if output != "" {
		t.Fatalf("expected empty output for nil response, got %q", output)
	}
}

func TestDisplayDeviceCodePromptNilWriter(t *testing.T) {
	response := &DeviceCodeResponse{UserCode: "CODE"}
	displayDeviceCodePrompt(nil, response, false)
}

func TestDeviceCodeResponseStructure(t *testing.T) {
	resp := &DeviceCodeResponse{
		DeviceCode:              "device123",
		UserCode:                "USER-CODE",
		VerificationURL:         "https://example.com/auth",
		VerificationURI:         "https://example.com/auth",
		VerificationURLComplete: "https://example.com/auth?code=USER-CODE",
		ExpiresIn:               1800,
		Interval:                5,
	}

	if resp.DeviceCode != "device123" {
		t.Fatalf("expected device code")
	}
	if resp.UserCode != "USER-CODE" {
		t.Fatalf("expected user code")
	}
	if resp.ExpiresIn != 1800 {
		t.Fatalf("expected expires in")
	}
	if resp.Interval != 5 {
		t.Fatalf("expected interval")
	}
}

func TestTokenResponseStructure(t *testing.T) {
	resp := &TokenResponse{
		AccessToken:  "access123",
		RefreshToken: "refresh123",
		ExpiresIn:    3600,
		TokenType:    "Bearer",
		Scope:        "scope1 scope2",
		Error:        "",
	}

	if resp.AccessToken != "access123" {
		t.Fatalf("expected access token")
	}
	if resp.RefreshToken != "refresh123" {
		t.Fatalf("expected refresh token")
	}
	if resp.ExpiresIn != 3600 {
		t.Fatalf("expected expires in")
	}
	if resp.TokenType != "Bearer" {
		t.Fatalf("expected token type")
	}
}

func TestDisplayDeviceCodePromptPreferVerificationURL(t *testing.T) {
	response := &DeviceCodeResponse{
		UserCode:        "USER-CODE",
		VerificationURL: "https://example.com/auth",
		VerificationURI: "https://example.com/auth-uri",
	}
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, response, false)
	output := buf.String()
	if !strings.Contains(output, "https://example.com/auth") {
		t.Fatalf("expected VerificationURL to be used, got %q", output)
	}
}

func TestDisplayDeviceCodePromptFallbackToURI(t *testing.T) {
	response := &DeviceCodeResponse{
		UserCode:        "USER-CODE",
		VerificationURI: "https://example.com/auth-uri",
	}
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, response, false)
	output := buf.String()
	if !strings.Contains(output, "https://example.com/auth-uri") {
		t.Fatalf("expected VerificationURI to be used, got %q", output)
	}
}

func TestDisplayDeviceCodePromptWriterError(t *testing.T) {
	response := &DeviceCodeResponse{
		UserCode:        "USER-CODE",
		VerificationURL: "https://example.com/auth",
	}

	failingWriter := &failingWriter{}
	displayDeviceCodePrompt(failingWriter, response, false)
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, io.ErrClosedPipe
}

func TestNewDeviceCodeFlowMultipleScopes(t *testing.T) {
	config := &oauth2.Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		Scopes:       []string{"scope1", "scope2", "scope3"},
	}

	flow := NewDeviceCodeFlow(config)
	if flow.config.Scopes[0] != "scope1" || flow.config.Scopes[1] != "scope2" || flow.config.Scopes[2] != "scope3" {
		t.Fatalf("expected all scopes to be set")
	}
}

func TestPollForTokenMinimumInterval(t *testing.T) {
	config := &oauth2.Config{ClientID: "test"}
	flow := NewDeviceCodeFlow(config)
	flow.response = &DeviceCodeResponse{
		DeviceCode: "device123",
		ExpiresIn:  3600,
		Interval:   1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-time.After(100 * time.Millisecond)
		cancel()
	}()

	_, err := flow.PollForToken(ctx)
	if err == nil {
		t.Fatalf("expected error for canceled context")
	}
}

func TestDisplayDeviceCodePromptAllFields(t *testing.T) {
	response := &DeviceCodeResponse{
		DeviceCode:              "device123",
		UserCode:                "USER-CODE",
		VerificationURL:         "https://example.com/auth",
		VerificationURI:         "https://example.com/auth-uri",
		VerificationURLComplete: "https://example.com/auth?code=USER-CODE",
		ExpiresIn:               1800,
		Interval:                5,
	}
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, response, false)
	output := buf.String()

	if !strings.Contains(output, "Authenticate with Google Play Developer CLI") {
		t.Fatalf("expected header in output")
	}
	if !strings.Contains(output, "USER-CODE") {
		t.Fatalf("expected user code in output")
	}
	if !strings.Contains(output, "Waiting for authorization") {
		t.Fatalf("expected waiting message in output")
	}
}

func TestDeviceCodeFlowConfigPreserved(t *testing.T) {
	config := &oauth2.Config{
		ClientID:     "my-client",
		ClientSecret: "my-secret",
		Scopes:       []string{"scope1"},
	}

	flow := NewDeviceCodeFlow(config)
	if flow.config.ClientID != "my-client" {
		t.Fatalf("expected client ID to be preserved")
	}
	if flow.config.ClientSecret != "my-secret" {
		t.Fatalf("expected client secret to be preserved")
	}
}

func TestDisplayDeviceCodePromptEmptyUserCode(t *testing.T) {
	response := &DeviceCodeResponse{
		UserCode:        "",
		VerificationURL: "https://example.com/auth",
	}
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, response, false)
	output := buf.String()
	if !strings.Contains(output, "https://example.com/auth") {
		t.Fatalf("expected URL in output even with empty user code")
	}
}

func TestDisplayDeviceCodePromptEmptyURL(t *testing.T) {
	response := &DeviceCodeResponse{
		UserCode:        "USER-CODE",
		VerificationURL: "",
		VerificationURI: "",
	}
	buf := &bytes.Buffer{}
	displayDeviceCodePrompt(buf, response, false)
	output := buf.String()
	if !strings.Contains(output, "USER-CODE") {
		t.Fatalf("expected user code in output even with empty URLs")
	}
}
