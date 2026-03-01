package apitest

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/androidpublisher/v3"
)

func TestNewMockClient(t *testing.T) {
	t.Parallel()

	client := NewMockClient()

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.PublisherResponses == nil {
		t.Error("expected PublisherResponses to be initialized")
	}

	if client.PublisherResponses.Edits == nil {
		t.Error("expected PublisherResponses.Edits to be initialized")
	}

	if client.PublisherResponses.Tracks == nil {
		t.Error("expected PublisherResponses.Tracks to be initialized")
	}

	if client.PublisherResponses.Bundles == nil {
		t.Error("expected PublisherResponses.Bundles to be initialized")
	}

	if client.PublisherResponses.APKs == nil {
		t.Error("expected PublisherResponses.APKs to be initialized")
	}

	if client.PublisherResponses.Testers == nil {
		t.Error("expected PublisherResponses.Testers to be initialized")
	}

	if client.PublisherResponses.Listings == nil {
		t.Error("expected PublisherResponses.Listings to be initialized")
	}

	if client.PublisherResponses.Images == nil {
		t.Error("expected PublisherResponses.Images to be initialized")
	}

	if client.ReportingResponses == nil {
		t.Error("expected ReportingResponses to be initialized")
	}

	if client.ReportingResponses.Vitals == nil {
		t.Error("expected ReportingResponses.Vitals to be initialized")
	}

	if client.GamesResponses == nil {
		t.Error("expected GamesResponses to be initialized")
	}

	if client.GamesResponses.Achievements == nil {
		t.Error("expected GamesResponses.Achievements to be initialized")
	}

	if client.GamesResponses.Scores == nil {
		t.Error("expected GamesResponses.Scores to be initialized")
	}

	if client.GamesResponses.Events == nil {
		t.Error("expected GamesResponses.Events to be initialized")
	}

	if client.GamesResponses.Players == nil {
		t.Error("expected GamesResponses.Players to be initialized")
	}

	if client.IntegrityResponses == nil {
		t.Error("expected IntegrityResponses to be initialized")
	}

	if client.CustomAppResponses == nil {
		t.Error("expected CustomAppResponses to be initialized")
	}

	if client.Calls == nil {
		t.Error("expected Calls to be initialized")
	}
}

func TestMockClientTrackCall(t *testing.T) {
	t.Parallel()

	client := NewMockClient()

	tests := []struct {
		name    string
		service string
		method  string
		args    map[string]interface{}
	}{
		{
			name:    "track single call",
			service: "AndroidPublisher",
			method:  "Edits.Insert",
			args:    map[string]interface{}{"packageName": "com.example.app"},
		},
		{
			name:    "track call without args",
			service: "PlayIntegrity",
			method:  "DecodeToken",
			args:    nil,
		},
		{
			name:    "track multiple calls",
			service: "AndroidPublisher",
			method:  "Edits.Commit",
			args:    map[string]interface{}{"editId": "123", "packageName": "com.example.app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client.TrackCall(tt.service, tt.method, tt.args)

			if len(client.Calls) == 0 {
				t.Fatal("expected at least one call recorded")
			}

			lastCall := client.Calls[len(client.Calls)-1]
			if lastCall.Service != tt.service {
				t.Errorf("expected service %q, got %q", tt.service, lastCall.Service)
			}
			if lastCall.Method != tt.method {
				t.Errorf("expected method %q, got %q", tt.method, lastCall.Method)
			}
		})
	}
}

func TestMockClientGetCallCount(t *testing.T) {
	t.Parallel()

	client := NewMockClient()

	client.TrackCall("AndroidPublisher", "Edits.Insert", nil)
	client.TrackCall("AndroidPublisher", "Edits.Insert", nil)
	client.TrackCall("AndroidPublisher", "Edits.Commit", nil)
	client.TrackCall("PlayIntegrity", "DecodeToken", nil)

	tests := []struct {
		name     string
		service  string
		method   string
		expected int
	}{
		{
			name:     "count specific service method",
			service:  "AndroidPublisher",
			method:   "Edits.Insert",
			expected: 2,
		},
		{
			name:     "count different method",
			service:  "AndroidPublisher",
			method:   "Edits.Commit",
			expected: 1,
		},
		{
			name:     "count different service",
			service:  "PlayIntegrity",
			method:   "DecodeToken",
			expected: 1,
		},
		{
			name:     "count non-existent method",
			service:  "AndroidPublisher",
			method:   "Edits.Get",
			expected: 0,
		},
		{
			name:     "count non-existent service",
			service:  "NonExistent",
			method:   "Method",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.GetCallCount(tt.service, tt.method)
			if got != tt.expected {
				t.Errorf("expected count %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestMockClientResetCalls(t *testing.T) {
	t.Parallel()

	client := NewMockClient()

	client.TrackCall("Service1", "Method1", nil)
	client.TrackCall("Service2", "Method2", nil)

	if len(client.Calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(client.Calls))
	}

	client.ResetCalls()

	if len(client.Calls) != 0 {
		t.Errorf("expected 0 calls after reset, got %d", len(client.Calls))
	}
}

func TestMockClientSetPublisherResponse(t *testing.T) {
	t.Parallel()

	client := NewMockClient()

	track := &androidpublisher.Track{
		Track: "internal",
	}

	editID := "edit-123"
	client.SetPublisherResponse(editID, track)

	got, ok := client.PublisherResponses.Tracks[editID]
	if !ok {
		t.Fatal("expected track to be set")
	}

	if got.Track != track.Track {
		t.Errorf("expected track %q, got %q", track.Track, got.Track)
	}
}

func TestMockClientSetPublisherResponseWithNilMap(t *testing.T) {
	t.Parallel()

	client := NewMockClient()
	client.PublisherResponses.Tracks = nil

	track := &androidpublisher.Track{
		Track: "production",
	}

	editID := "edit-456"
	client.SetPublisherResponse(editID, track)

	if client.PublisherResponses.Tracks == nil {
		t.Error("expected Tracks map to be initialized")
	}

	got, ok := client.PublisherResponses.Tracks[editID]
	if !ok {
		t.Fatal("expected track to be set")
	}

	if got.Track != track.Track {
		t.Errorf("expected track %q, got %q", track.Track, got.Track)
	}
}

func TestMockClientServiceAccessors(t *testing.T) {
	t.Parallel()

	client := NewMockClient()

	tests := []struct {
		name     string
		accessor func() error
	}{
		{
			name: "AndroidPublisher",
			accessor: func() error {
				_, err := client.AndroidPublisher()
				return err
			},
		},
		{
			name: "PlayReporting",
			accessor: func() error {
				_, err := client.PlayReporting()
				return err
			},
		},
		{
			name: "GamesManagement",
			accessor: func() error {
				_, err := client.GamesManagement()
				return err
			},
		},
		{
			name: "Games",
			accessor: func() error {
				_, err := client.Games()
				return err
			},
		},
		{
			name: "PlayIntegrity",
			accessor: func() error {
				_, err := client.PlayIntegrity()
				return err
			},
		},
		{
			name: "PlayCustomApp",
			accessor: func() error {
				_, err := client.PlayCustomApp()
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.accessor(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockClientAcquireRelease(t *testing.T) {
	t.Parallel()

	client := NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{
			name: "Acquire",
			fn: func() error {
				return client.Acquire(ctx)
			},
		},
		{
			name: "Release",
			fn: func() error {
				client.Release()
				return nil
			},
		},
		{
			name: "AcquireForUpload",
			fn: func() error {
				return client.AcquireForUpload(ctx)
			},
		},
		{
			name: "ReleaseForUpload",
			fn: func() error {
				client.ReleaseForUpload()
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockClientDoWithRetry(t *testing.T) {
	t.Parallel()

	client := NewMockClient()
	ctx := context.Background()

	tests := []struct {
		name        string
		fn          func() error
		expectError bool
	}{
		{
			name: "successful function",
			fn: func() error {
				return nil
			},
			expectError: false,
		},
		{
			name: "function returns error",
			fn: func() error {
				return errors.New("test error")
			},
			expectError: true,
		},
		{
			name: "function with multiple attempts still returns error",
			fn: func() error {
				return errors.New("persistent error")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.DoWithRetry(ctx, tt.fn)
			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockClientRetryConfig(t *testing.T) {
	t.Parallel()

	client := NewMockClient()
	cfg := client.RetryConfig()

	if cfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", cfg.MaxAttempts)
	}

	if cfg.InitialDelay != time.Second {
		t.Errorf("expected InitialDelay 1s, got %v", cfg.InitialDelay)
	}

	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("expected MaxDelay 30s, got %v", cfg.MaxDelay)
	}
}

func TestMockTokenSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tokenFunc   func() (*oauth2.Token, error)
		expectToken string
		expectError bool
	}{
		{
			name:        "default token",
			tokenFunc:   nil,
			expectToken: "mock-token",
			expectError: false,
		},
		{
			name: "custom token",
			tokenFunc: func() (*oauth2.Token, error) {
				return &oauth2.Token{
					AccessToken: "custom-token",
					TokenType:   "Bearer",
				}, nil
			},
			expectToken: "custom-token",
			expectError: false,
		},
		{
			name: "error from custom function",
			tokenFunc: func() (*oauth2.Token, error) {
				return nil, errors.New("token error")
			},
			expectToken: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mts := &MockTokenSource{
				TokenFunc: tt.tokenFunc,
			}

			token, err := mts.Token()
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if token.AccessToken != tt.expectToken {
				t.Errorf("expected token %q, got %q", tt.expectToken, token.AccessToken)
			}

			if token.TokenType != "Bearer" {
				t.Errorf("expected token type Bearer, got %q", token.TokenType)
			}
		})
	}
}

func TestMockClientConcurrency(t *testing.T) {
	t.Parallel()

	client := NewMockClient()

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		idx := i
		go func() {
			defer func() { done <- true }()

			service := "Service"
			method := "Method"
			args := map[string]interface{}{"idx": idx}

			client.TrackCall(service, method, args)
		}()
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	if len(client.Calls) != numGoroutines {
		t.Errorf("expected %d calls, got %d", numGoroutines, len(client.Calls))
	}
}

func TestErrMockNotConfigured(t *testing.T) {
	t.Parallel()

	if ErrMockNotConfigured == nil {
		t.Error("expected ErrMockNotConfigured to be non-nil")
	}

	if ErrMockNotConfigured.Error() != "mock response not configured" {
		t.Errorf("unexpected error message: %v", ErrMockNotConfigured)
	}
}
