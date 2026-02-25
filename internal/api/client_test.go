//go:build unit
// +build unit

package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
)

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultRetryConfig()
	if cfg.MaxAttempts != 3 {
		t.Fatalf("expected MaxAttempts 3, got %d", cfg.MaxAttempts)
	}
	if cfg.InitialDelay != time.Second {
		t.Fatalf("expected InitialDelay 1s, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Fatalf("expected MaxDelay 30s, got %v", cfg.MaxDelay)
	}
}

func TestAcquireRelease(t *testing.T) {
	t.Parallel()
	client := &Client{semaphore: make(chan struct{}, 1)}
	ctx := context.Background()
	if err := client.Acquire(ctx); err != nil {
		t.Fatalf("acquire error: %v", err)
	}
	if len(client.semaphore) != 1 {
		t.Fatalf("expected semaphore len 1, got %d", len(client.semaphore))
	}
	client.Release()
	if len(client.semaphore) != 0 {
		t.Fatalf("expected semaphore len 0, got %d", len(client.semaphore))
	}
}

func TestAcquireCanceled(t *testing.T) {
	t.Parallel()
	client := &Client{semaphore: make(chan struct{}, 1)}
	client.semaphore <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.Acquire(ctx); err == nil {
		t.Fatalf("expected error")
	}
	if len(client.semaphore) != 1 {
		t.Fatalf("expected semaphore unchanged, got %d", len(client.semaphore))
	}
}

func TestNewClientAndServices(t *testing.T) {
	t.Parallel()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)})
	client, err := NewClient(context.Background(), ts, WithTimeout(2*time.Second), WithMaxRetryAttempts(5), WithConcurrentCalls(2))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if client.timeout != 2*time.Second || client.retryConfig.MaxAttempts != 5 || cap(client.semaphore) != 2 {
		t.Fatalf("unexpected client config")
	}
	if _, err := client.AndroidPublisher(); err != nil {
		t.Fatalf("AndroidPublisher error: %v", err)
	}
	if _, err := client.PlayReporting(); err != nil {
		t.Fatalf("PlayReporting error: %v", err)
	}
	if _, err := client.GamesManagement(); err != nil {
		t.Fatalf("GamesManagement error: %v", err)
	}
}

func TestWithRetryConfig(t *testing.T) {
	t.Parallel()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "token", Expiry: time.Now().Add(time.Hour)})
	cfg := RetryConfig{MaxAttempts: 4, InitialDelay: time.Millisecond, MaxDelay: time.Second}
	client, err := NewClient(context.Background(), ts, WithRetryConfig(cfg))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if client.retryConfig.MaxAttempts != 4 {
		t.Fatalf("unexpected retry config")
	}
}

func TestAcquireForUploadAndRelease(t *testing.T) {
	t.Parallel()
	client := &Client{semaphore: make(chan struct{}, 2)}
	ctx := context.Background()
	if err := client.AcquireForUpload(ctx); err != nil {
		t.Fatalf("acquire for upload error: %v", err)
	}
	if len(client.semaphore) != cap(client.semaphore) {
		t.Fatalf("expected semaphore full, got %d", len(client.semaphore))
	}
	client.ReleaseForUpload()
	if len(client.semaphore) != 0 {
		t.Fatalf("expected semaphore empty, got %d", len(client.semaphore))
	}
}

func TestAcquireForUploadCanceled(t *testing.T) {
	t.Parallel()
	client := &Client{semaphore: make(chan struct{}, 1)}
	client.semaphore <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.AcquireForUpload(ctx); err == nil {
		t.Fatalf("expected error for canceled context")
	}
	if len(client.semaphore) != 1 {
		t.Fatalf("expected semaphore unchanged, got %d", len(client.semaphore))
	}
}

func TestAcquireForUploadPartialCancel(t *testing.T) {
	t.Parallel()
	client := &Client{semaphore: make(chan struct{}, 2)}
	client.semaphore <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		result <- client.AcquireForUpload(ctx)
	}()
	for len(client.semaphore) < 2 {
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-result; err == nil {
		t.Fatalf("expected error")
	}
	if len(client.semaphore) != 1 {
		t.Fatalf("expected one slot released, got %d", len(client.semaphore))
	}
}

func TestDoWithRetrySuccess(t *testing.T) {
	client := &Client{retryConfig: RetryConfig{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	ctx := context.Background()
	calls := 0
	err := client.DoWithRetry(ctx, func() error {
		calls++
		if calls == 1 {
			return &googleapi.Error{Code: http.StatusInternalServerError}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDoWithRetryNonRetryable(t *testing.T) {
	t.Parallel()
	client := &Client{retryConfig: RetryConfig{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	ctx := context.Background()
	calls := 0
	err := client.DoWithRetry(ctx, func() error {
		calls++
		return &googleapi.Error{Code: http.StatusBadRequest}
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDoWithRetryMaxAttempts(t *testing.T) {
	client := &Client{retryConfig: RetryConfig{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	ctx := context.Background()
	calls := 0
	err := client.DoWithRetry(ctx, func() error {
		calls++
		return &googleapi.Error{Code: http.StatusInternalServerError}
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestDoWithRetryCanceled(t *testing.T) {
	t.Parallel()
	client := &Client{retryConfig: RetryConfig{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.DoWithRetry(ctx, func() error { return nil }); err == nil {
		t.Fatalf("expected context error")
	}
}

func TestDoWithRetryCanceledDuringWait(t *testing.T) {
	t.Parallel()
	client := &Client{retryConfig: RetryConfig{MaxAttempts: 3, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := client.DoWithRetry(ctx, func() error {
		calls++
		if calls == 1 {
			cancel()
		}
		return &googleapi.Error{Code: http.StatusInternalServerError}
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestExtractRetryAfter(t *testing.T) {
	t.Parallel()
	apiErr := &googleapi.Error{Header: http.Header{"Retry-After": []string{"120"}}}
	if got := extractRetryAfter(apiErr); got != 120*time.Second {
		t.Fatalf("expected 120s, got %v", got)
	}

	ts := time.Now().Add(2 * time.Second).UTC()
	apiErr = &googleapi.Error{Header: http.Header{"Retry-After": []string{ts.Format(http.TimeFormat)}}}
	got := extractRetryAfter(apiErr)
	if got <= 0 {
		t.Fatalf("expected positive duration, got %v", got)
	}
}

func TestCalculateDelayWithConfig(t *testing.T) {
	t.Parallel()
	cfg := RetryConfig{MaxAttempts: 3, InitialDelay: time.Second, MaxDelay: 2 * time.Second}
	delay := calculateDelayWithConfig(cfg, 0, nil)
	if delay < time.Second || delay > 1300*time.Millisecond {
		t.Fatalf("expected delay within 1s-1.3s, got %v", delay)
	}

	apiErr := &googleapi.Error{Header: http.Header{"Retry-After": []string{"5"}}}
	delay = calculateDelayWithConfig(cfg, 1, apiErr)
	if delay != 2*time.Second {
		t.Fatalf("expected capped delay 2s, got %v", delay)
	}
}

func TestCalculateDelayWithConfigRetryAfterUnderMax(t *testing.T) {
	t.Parallel()
	cfg := RetryConfig{MaxAttempts: 3, InitialDelay: time.Second, MaxDelay: 5 * time.Second}
	apiErr := &googleapi.Error{Header: http.Header{"Retry-After": []string{"2"}}}
	delay := calculateDelayWithConfig(cfg, 1, apiErr)
	if delay != 2*time.Second {
		t.Fatalf("expected 2s, got %v", delay)
	}
}

func TestCalculateDelayWithConfigLargeAttempt(t *testing.T) {
	t.Parallel()
	cfg := RetryConfig{MaxAttempts: 3, InitialDelay: time.Second, MaxDelay: 2 * time.Second}
	delay := calculateDelayWithConfig(cfg, 100, nil)
	if delay < 0 || delay > 2*time.Second {
		t.Fatalf("unexpected delay: %v", delay)
	}
}

func TestCalculateDelayWithConfigCapsDelay(t *testing.T) {
	t.Parallel()
	cfg := RetryConfig{MaxAttempts: 3, InitialDelay: time.Second, MaxDelay: 2 * time.Second}
	delay := calculateDelayWithConfig(cfg, 3, nil)
	if delay < 2*time.Second || delay > 2600*time.Millisecond {
		t.Fatalf("unexpected delay: %v", delay)
	}
}

func TestValidTrackAndStatus(t *testing.T) {
	t.Parallel()
	if !IsValidTrack("internal") || IsValidTrack("nope") {
		t.Fatalf("unexpected track validation result")
	}
	if !IsValidReleaseStatus("draft") || IsValidReleaseStatus("bad") {
		t.Fatalf("unexpected release status validation result")
	}
}

func TestDefaultUploadOptions(t *testing.T) {
	t.Parallel()
	opts := DefaultUploadOptions()
	if opts.ChunkSize != 8*1024*1024 {
		t.Fatalf("expected 8MB chunk, got %d", opts.ChunkSize)
	}
}

func TestIsRetryableError(t *testing.T) {
	t.Parallel()
	if isRetryableError(nil) {
		t.Fatalf("expected false for nil error")
	}
	if isRetryableError(context.Canceled) {
		t.Fatalf("expected false for non-googleapi error")
	}
	if !isRetryableError(&googleapi.Error{Code: http.StatusTooManyRequests}) {
		t.Fatalf("expected true for 429")
	}
	if !isRetryableError(&googleapi.Error{Code: http.StatusInternalServerError}) {
		t.Fatalf("expected true for 5xx")
	}
}

func TestExtractRetryAfterInvalid(t *testing.T) {
	t.Parallel()
	apiErr := &googleapi.Error{Header: http.Header{"Retry-After": []string{"not-a-date"}}}
	if got := extractRetryAfter(apiErr); got != 0 {
		t.Fatalf("expected 0 for invalid header, got %v", got)
	}
}

func TestCryptoRandFloat64Error(t *testing.T) {
	// Not parallel due to global randRead modification
	orig := randRead
	randRead = func(_ []byte) (int, error) {
		return 0, context.Canceled
	}
	t.Cleanup(func() { randRead = orig })
	if got := cryptoRandFloat64(); got != 0.5 {
		t.Fatalf("expected 0.5 fallback, got %v", got)
	}
}

func TestRetryConfigGetter(t *testing.T) {
	t.Parallel()
	client := &Client{retryConfig: RetryConfig{MaxAttempts: 7}}
	if client.RetryConfig().MaxAttempts != 7 {
		t.Fatalf("unexpected retry config")
	}
}

func TestCalculateDelayMethod(t *testing.T) {
	t.Parallel()
	client := &Client{retryConfig: RetryConfig{InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	if delay := client.calculateDelay(0, nil); delay == 0 {
		t.Fatalf("expected non-zero delay")
	}
}

func TestGamesServiceInitialization(t *testing.T) {
	t.Parallel()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test"})
	client, err := NewClient(context.Background(), ts)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	svc, err := client.Games()
	if err != nil {
		t.Fatalf("Games() error: %v", err)
	}
	if svc == nil {
		t.Fatal("Games service should not be nil")
	}

	svc2, err := client.Games()
	if err != nil {
		t.Fatalf("Games() second call error: %v", err)
	}
	if svc != svc2 {
		t.Error("Games() should return the same service instance")
	}
}

func TestPlayIntegrityServiceInitialization(t *testing.T) {
	t.Parallel()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test"})
	client, err := NewClient(context.Background(), ts)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	svc, err := client.PlayIntegrity()
	if err != nil {
		t.Fatalf("PlayIntegrity() error: %v", err)
	}
	if svc == nil {
		t.Fatal("PlayIntegrity service should not be nil")
	}

	svc2, err := client.PlayIntegrity()
	if err != nil {
		t.Fatalf("PlayIntegrity() second call error: %v", err)
	}
	if svc != svc2 {
		t.Error("PlayIntegrity() should return the same service instance")
	}
}

func TestPlayCustomAppServiceInitialization(t *testing.T) {
	t.Parallel()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test"})
	client, err := NewClient(context.Background(), ts)
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}

	svc, err := client.PlayCustomApp()
	if err != nil {
		t.Fatalf("PlayCustomApp() error: %v", err)
	}
	if svc == nil {
		t.Fatal("PlayCustomApp service should not be nil")
	}

	svc2, err := client.PlayCustomApp()
	if err != nil {
		t.Fatalf("PlayCustomApp() second call error: %v", err)
	}
	if svc != svc2 {
		t.Error("PlayCustomApp() should return the same service instance")
	}
}

// ============================================================================
// Logging Transport Tests
// ============================================================================

func TestLoggingTransport_NonVerbose(t *testing.T) {
	t.Parallel()
	base := &testTransport{response: &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       http.NoBody,
		Header:     http.Header{},
	}}

	transport := &loggingTransport{
		base:    base,
		verbose: false,
	}

	req := httptest.NewRequest("GET", "http://example.com/test", http.NoBody)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}
	if !base.called {
		t.Error("Expected base transport to be called")
	}
}

func TestLoggingTransport_Verbose(t *testing.T) {
	// Not parallel - tests logging output
	base := &testTransport{response: &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       http.NoBody,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    httptest.NewRequest("GET", "http://example.com/test", nil),
	}}

	transport := &loggingTransport{
		base:    base,
		verbose: true,
	}

	req := httptest.NewRequest("GET", "http://example.com/test", http.NoBody)
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Content-Type", "application/json")

	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected non-nil response")
	}
	if !base.called {
		t.Error("Expected base transport to be called")
	}
}

func TestLoggingTransport_Error(t *testing.T) {
	t.Parallel()
	expectedErr := context.Canceled
	base := &testTransport{err: expectedErr}

	transport := &loggingTransport{
		base:    base,
		verbose: true,
	}

	req := httptest.NewRequest("GET", "http://example.com/test", http.NoBody)
	resp, err := transport.RoundTrip(req)

	if err != expectedErr {
		t.Errorf("Expected error %v, got: %v", expectedErr, err)
	}
	if resp != nil {
		t.Error("Expected nil response on error")
	}
}

func TestLoggingTransport_ErrorResponse(t *testing.T) {
	t.Parallel()
	base := &testTransport{response: &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "500 Internal Server Error",
		Body:       http.NoBody,
		Header:     http.Header{},
	}}

	transport := &loggingTransport{
		base:    base,
		verbose: true,
	}

	req := httptest.NewRequest("GET", "http://example.com/test", http.NoBody)
	resp, err := transport.RoundTrip(req)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got: %d", resp.StatusCode)
	}
}

func TestFormatHeaders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		headers  map[string]interface{}
		contains string
	}{
		{
			name:     "empty headers",
			headers:  map[string]interface{}{},
			contains: "{}",
		},
		{
			name:     "single header",
			headers:  map[string]interface{}{"Content-Type": "application/json"},
			contains: "Content-Type",
		},
		{
			name:     "multiple headers",
			headers:  map[string]interface{}{"Content-Type": "application/json", "Authorization": "Bearer token"},
			contains: "Content-Type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHeaders(tt.headers)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got: %s", tt.contains, result)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"empty string", "", `""`},
		{"non-empty string", "test", `"test"`},
		{"string slice", []string{"a", "b"}, "[array]"},
		{"int", 42, "[redacted]"},
		{"nil", nil, "[redacted]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %q, want %q", tt.value, result, tt.expected)
			}
		})
	}
}

func TestSummarizeResponseBody(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		contentLength int64
		expected      string
	}{
		{"with content", 100, "<body: 100 bytes>"},
		{"empty body", 0, "<empty body>"},
		{"chunked", -1, "<chunked body>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				ContentLength: tt.contentLength,
			}
			result := summarizeResponseBody(resp)
			if result != tt.expected {
				t.Errorf("summarizeResponseBody() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestWithVerbose_Context(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Test setting verbose to true
	ctxWithVerbose := WithVerbose(ctx, true)
	if !IsVerbose(ctxWithVerbose) {
		t.Error("Expected IsVerbose to return true")
	}

	// Test setting verbose to false
	ctxWithoutVerbose := WithVerbose(ctx, false)
	if IsVerbose(ctxWithoutVerbose) {
		t.Error("Expected IsVerbose to return false")
	}

	// Test default (no value in context)
	if IsVerbose(ctx) {
		t.Error("Expected IsVerbose to return false for context without value")
	}
}

func TestNewClient_WithVerboseLogging(t *testing.T) {
	t.Parallel()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test"})
	client, err := NewClient(context.Background(), ts, WithVerboseLogging(true))
	if err != nil {
		t.Fatalf("NewClient error: %v", err)
	}
	if !client.verbose {
		t.Error("Expected verbose to be true")
	}
}

// ============================================================================
// Test Transport Helper
// ============================================================================

type testTransport struct {
	response *http.Response
	err      error
	called   bool
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.called = true
	if t.err != nil {
		return nil, t.err
	}
	if t.response != nil && t.response.Request == nil {
		t.response.Request = req
	}
	return t.response, nil
}
