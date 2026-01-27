// Package api provides the unified API client for Google Play APIs.
package api

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/androidpublisher/v3"
	gamesmanagement "google.golang.org/api/gamesmanagement/v1management"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
	}
}

type Client struct {
	httpClient  *http.Client
	timeout     time.Duration
	retryConfig RetryConfig

	publisherOnce sync.Once
	publisherSvc  *androidpublisher.Service
	publisherErr  error

	reportingOnce sync.Once
	reportingSvc  *playdeveloperreporting.Service
	reportingErr  error

	gamesManagementOnce sync.Once
	gamesManagementSvc  *gamesmanagement.Service
	gamesManagementErr  error

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

func WithConcurrentCalls(n int) Option {
	return func(c *Client) {
		c.semaphore = make(chan struct{}, n)
	}
}

func WithRetryConfig(cfg RetryConfig) Option {
	return func(c *Client) {
		c.retryConfig = cfg
	}
}

func WithMaxRetryAttempts(n int) Option {
	return func(c *Client) {
		c.retryConfig.MaxAttempts = n
	}
}

func NewClient(ctx context.Context, tokenSource oauth2.TokenSource, opts ...Option) (*Client, error) {
	c := &Client{
		timeout:     30 * time.Second,
		semaphore:   make(chan struct{}, DefaultConcurrentCalls),
		retryConfig: DefaultRetryConfig(),
	}

	for _, opt := range opts {
		opt(c)
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	c.httpClient = &http.Client{
		Transport: &oauth2.Transport{
			Base:   transport,
			Source: tokenSource,
		},
		Timeout: c.timeout,
	}

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

// GamesManagement returns the Games Management API service.
func (c *Client) GamesManagement() (*gamesmanagement.Service, error) {
	c.gamesManagementOnce.Do(func() {
		c.gamesManagementSvc, c.gamesManagementErr = gamesmanagement.NewService(
			context.Background(),
			option.WithHTTPClient(c.httpClient),
		)
	})
	return c.gamesManagementSvc, c.gamesManagementErr
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

func (c *Client) ReleaseForUpload() {
	for i := 0; i < cap(c.semaphore); i++ {
		<-c.semaphore
	}
}

func (c *Client) DoWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < c.retryConfig.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if !isRetryableError(lastErr) {
			return lastErr
		}

		if attempt == c.retryConfig.MaxAttempts-1 {
			break
		}

		delay := c.calculateDelay(attempt, lastErr)
		fmt.Fprintf(os.Stderr, "Retrying request (attempt %d/%d) after %v due to: %v\n", attempt+2, c.retryConfig.MaxAttempts, delay, lastErr)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return lastErr
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if apiErr, ok := err.(*googleapi.Error); ok {
		if apiErr.Code == http.StatusTooManyRequests {
			return true
		}
		if apiErr.Code >= 500 && apiErr.Code < 600 {
			return true
		}
	}

	return false
}

func (c *Client) calculateDelay(attempt int, err error) time.Duration {
	return calculateDelayWithConfig(c.retryConfig, attempt, err)
}

func cryptoRandFloat64() float64 {
	var buf [8]byte
	_, err := randRead(buf[:])
	if err != nil {
		return 0.5
	}
	return float64(binary.BigEndian.Uint64(buf[:])&(1<<53-1)) / float64(1<<53)
}

var randRead = crand.Read

func calculateDelayWithConfig(cfg RetryConfig, attempt int, err error) time.Duration {
	if retryAfter := extractRetryAfter(err); retryAfter > 0 {
		if retryAfter > cfg.MaxDelay {
			return cfg.MaxDelay
		}
		return retryAfter
	}

	shift := attempt
	if shift > 62 {
		shift = 62
	}
	delay := cfg.InitialDelay * time.Duration(1<<shift)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}

	jitter := time.Duration(cryptoRandFloat64() * float64(delay) * 0.3)
	return delay + jitter
}

func extractRetryAfter(err error) time.Duration {
	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return 0
	}

	for key, values := range apiErr.Header {
		if http.CanonicalHeaderKey(key) == "Retry-After" && len(values) > 0 {
			if seconds, parseErr := strconv.Atoi(values[0]); parseErr == nil {
				return time.Duration(seconds) * time.Second
			}
			if t, parseErr := http.ParseTime(values[0]); parseErr == nil {
				return time.Until(t)
			}
		}
	}
	return 0
}

func (c *Client) RetryConfig() RetryConfig {
	return c.retryConfig
}

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
