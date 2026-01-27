package api

import (
	"context"
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
)

func TestDefaultRetryConfig(t *testing.T) {
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

func TestNewClientAndServices(t *testing.T) {
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

func TestAcquireForUploadAndRelease(t *testing.T) {
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

func TestDoWithRetryCanceled(t *testing.T) {
	client := &Client{retryConfig: RetryConfig{MaxAttempts: 2, InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.DoWithRetry(ctx, func() error { return nil }); err == nil {
		t.Fatalf("expected context error")
	}
}

func TestExtractRetryAfter(t *testing.T) {
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

func TestValidTrackAndStatus(t *testing.T) {
	if !IsValidTrack("internal") || IsValidTrack("nope") {
		t.Fatalf("unexpected track validation result")
	}
	if !IsValidReleaseStatus("draft") || IsValidReleaseStatus("bad") {
		t.Fatalf("unexpected release status validation result")
	}
}

func TestDefaultUploadOptions(t *testing.T) {
	opts := DefaultUploadOptions()
	if opts.ChunkSize != 8*1024*1024 {
		t.Fatalf("expected 8MB chunk, got %d", opts.ChunkSize)
	}
}

func TestIsRetryableError(t *testing.T) {
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
	apiErr := &googleapi.Error{Header: http.Header{"Retry-After": []string{"not-a-date"}}}
	if got := extractRetryAfter(apiErr); got != 0 {
		t.Fatalf("expected 0 for invalid header, got %v", got)
	}
}

func TestCryptoRandFloat64Error(t *testing.T) {
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
	client := &Client{retryConfig: RetryConfig{MaxAttempts: 7}}
	if client.RetryConfig().MaxAttempts != 7 {
		t.Fatalf("unexpected retry config")
	}
}

func TestCalculateDelayMethod(t *testing.T) {
	client := &Client{retryConfig: RetryConfig{InitialDelay: time.Millisecond, MaxDelay: time.Millisecond}}
	if delay := client.calculateDelay(0, nil); delay == 0 {
		t.Fatalf("expected non-zero delay")
	}
}
