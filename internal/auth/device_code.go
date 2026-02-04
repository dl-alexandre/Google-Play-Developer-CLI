package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURL         string `json:"verification_url"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURLComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	Error        string `json:"error,omitempty"`
}

const (
	// #nosec G101 -- OAuth endpoints, not credentials.
	deviceCodeEndpoint = "https://oauth2.googleapis.com/device/code"
	// #nosec G101 -- OAuth endpoints, not credentials.
	tokenEndpoint = "https://oauth2.googleapis.com/token"
	authPending   = "authorization_pending"
	slowDown      = "slow_down"
)

type DeviceCodeFlow struct {
	config   *oauth2.Config
	response *DeviceCodeResponse
}

func NewDeviceCodeFlow(config *oauth2.Config) *DeviceCodeFlow {
	return &DeviceCodeFlow{config: config}
}

func (f *DeviceCodeFlow) RequestDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", f.config.ClientID)
	data.Set("scope", strings.Join(f.config.Scopes, " "))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, deviceCodeEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create device code request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device code request failed: %s - %s", resp.Status, string(body))
	}

	var deviceResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode device code response: %w", err)
	}
	f.response = &deviceResp
	return &deviceResp, nil
}

func (f *DeviceCodeFlow) PollForToken(ctx context.Context) (*oauth2.Token, error) {
	if f.response == nil {
		return nil, fmt.Errorf("device code not requested")
	}

	interval := time.Duration(f.response.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	expiry := time.Now().Add(time.Duration(f.response.ExpiresIn) * time.Second)
	client := &http.Client{Timeout: 30 * time.Second}

	for {
		if time.Now().After(expiry) {
			return nil, fmt.Errorf("device code expired")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		token, errType, err := f.pollOnce(ctx, client)
		if err != nil {
			return nil, err
		}
		if token != nil {
			return token, nil
		}
		switch errType {
		case authPending:
			time.Sleep(interval)
		case slowDown:
			interval += 5 * time.Second
			time.Sleep(interval)
		default:
			time.Sleep(interval)
		}
	}
}

func (f *DeviceCodeFlow) pollOnce(ctx context.Context, client *http.Client) (*oauth2.Token, string, error) {
	data := url.Values{}
	data.Set("client_id", f.config.ClientID)
	if f.config.ClientSecret != "" {
		data.Set("client_secret", f.config.ClientSecret)
	}
	data.Set("device_code", f.response.DeviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create token poll request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to poll for token: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.Error != "" {
		switch tokenResp.Error {
		case authPending, slowDown:
			return nil, tokenResp.Error, nil
		case "expired_token":
			return nil, "", fmt.Errorf("device code has expired")
		case "access_denied":
			return nil, "", fmt.Errorf("user denied authorization")
		default:
			return nil, "", fmt.Errorf("token error: %s", tokenResp.Error)
		}
	}

	if tokenResp.AccessToken != "" {
		expiryDate := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
		return &oauth2.Token{
			AccessToken:  tokenResp.AccessToken,
			RefreshToken: tokenResp.RefreshToken,
			TokenType:    tokenResp.TokenType,
			Expiry:       expiryDate,
		}, "", nil
	}

	return nil, authPending, nil
}

func displayDeviceCodePrompt(w io.Writer, resp *DeviceCodeResponse) {
	if w == nil || resp == nil {
		return
	}
	verificationURL := resp.VerificationURL
	if verificationURL == "" {
		verificationURL = resp.VerificationURI
	}
	_, _ = fmt.Fprintln(w, "Authenticate with Google Play Developer CLI")
	_, _ = fmt.Fprintln(w, "1) Visit:", verificationURL)
	_, _ = fmt.Fprintln(w, "2) Enter code:", resp.UserCode)
	if resp.VerificationURLComplete != "" {
		_, _ = fmt.Fprintln(w, "Or visit:", resp.VerificationURLComplete)
	}
	_, _ = fmt.Fprintln(w, "Waiting for authorization...")
}
