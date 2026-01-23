// Package api provides the unified API client for Google Play APIs.
package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/playdeveloperreporting/v1beta1"
)

// Client provides access to Google Play APIs.
type Client struct {
	httpClient *http.Client
	timeout    time.Duration

	// Lazy-initialized services
	publisherOnce sync.Once
	publisherSvc  *androidpublisher.Service
	publisherErr  error

	reportingOnce sync.Once
	reportingSvc  *playdeveloperreporting.Service
	reportingErr  error

	// Concurrency control
	semaphore chan struct{}
}

// DefaultConcurrentCalls is the default number of concurrent API calls.
const DefaultConcurrentCalls = 3

// Option configures the API client.
type Option func(*Client)

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// WithConcurrentCalls sets the maximum concurrent API calls.
func WithConcurrentCalls(n int) Option {
	return func(c *Client) {
		c.semaphore = make(chan struct{}, n)
	}
}

// NewClient creates a new API client with the given token source.
func NewClient(ctx context.Context, tokenSource oauth2.TokenSource, opts ...Option) (*Client, error) {
	c := &Client{
		timeout:   30 * time.Second,
		semaphore: make(chan struct{}, DefaultConcurrentCalls),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.httpClient = oauth2.NewClient(ctx, tokenSource)
	c.httpClient.Timeout = c.timeout

	return c, nil
}

// AndroidPublisher returns the Android Publisher API service.
func (c *Client) AndroidPublisher() (*androidpublisher.Service, error) {
	c.publisherOnce.Do(func() {
		c.publisherSvc, c.publisherErr = androidpublisher.NewService(
			context.Background(),
			option.WithHTTPClient(c.httpClient),
		)
	})
	return c.publisherSvc, c.publisherErr
}

// PlayReporting returns the Play Developer Reporting API service.
func (c *Client) PlayReporting() (*playdeveloperreporting.Service, error) {
	c.reportingOnce.Do(func() {
		c.reportingSvc, c.reportingErr = playdeveloperreporting.NewService(
			context.Background(),
			option.WithHTTPClient(c.httpClient),
		)
	})
	return c.reportingSvc, c.reportingErr
}

// Acquire acquires a semaphore slot for concurrent API calls.
func (c *Client) Acquire(ctx context.Context) error {
	select {
	case c.semaphore <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release releases a semaphore slot.
func (c *Client) Release() {
	<-c.semaphore
}

// AcquireForUpload acquires exclusive access for upload operations.
// Uploads are single-threaded for reliability.
func (c *Client) AcquireForUpload(ctx context.Context) error {
	// Acquire all slots to ensure exclusive access
	for i := 0; i < cap(c.semaphore); i++ {
		select {
		case c.semaphore <- struct{}{}:
		case <-ctx.Done():
			// Release any acquired slots
			for j := 0; j < i; j++ {
				<-c.semaphore
			}
			return ctx.Err()
		}
	}
	return nil
}

// ReleaseForUpload releases exclusive upload access.
func (c *Client) ReleaseForUpload() {
	for i := 0; i < cap(c.semaphore); i++ {
		<-c.semaphore
	}
}

// ValidTracks returns the list of valid track names.
var ValidTracks = []string{"internal", "alpha", "beta", "production"}

// IsValidTrack checks if a track name is valid.
func IsValidTrack(track string) bool {
	for _, t := range ValidTracks {
		if t == track {
			return true
		}
	}
	return false
}

// ReleaseStatus represents the status of a release.
type ReleaseStatus string

const (
	StatusDraft      ReleaseStatus = "draft"
	StatusCompleted  ReleaseStatus = "completed"
	StatusHalted     ReleaseStatus = "halted"
	StatusInProgress ReleaseStatus = "inProgress"
)

// IsValidReleaseStatus checks if a release status is valid.
func IsValidReleaseStatus(status string) bool {
	switch ReleaseStatus(status) {
	case StatusDraft, StatusCompleted, StatusHalted, StatusInProgress:
		return true
	default:
		return false
	}
}

// ReleaseConfig holds configuration for creating a release.
type ReleaseConfig struct {
	Track        string
	Name         string
	Status       ReleaseStatus
	VersionCodes []int64
	UserFraction float64
	ReleaseNotes map[string]string // locale -> text
}

// UploadOptions holds options for artifact uploads.
type UploadOptions struct {
	ChunkSize    int64                      // Chunk size for resumable uploads
	ProgressFunc func(current, total int64) // Progress callback
}

// DefaultUploadOptions returns the default upload options.
func DefaultUploadOptions() *UploadOptions {
	return &UploadOptions{
		ChunkSize: 8 * 1024 * 1024, // 8MB chunks
	}
}
