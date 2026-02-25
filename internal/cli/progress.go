// Package cli provides progress indicator utilities for long-running operations.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ProgressCallback is called during file uploads to report progress.
type ProgressCallback func(transferred, total int64, speed float64)

// ProgressBar is a console progress bar for file uploads and long operations.
type ProgressBar struct {
	writer      io.Writer
	quiet       bool
	total       int64
	transferred int64
	startTime   time.Time
	lastUpdate  time.Time
	width       int
	mu          sync.Mutex
	done        atomic.Bool
}

// NewProgressBar creates a new progress bar.
// Set quiet to true to suppress output (respects --quiet flag).
func NewProgressBar(quiet bool) *ProgressBar {
	width := 50
	if w, err := getTerminalWidth(); err == nil && w > 20 {
		width = w - 30 // Leave room for text
		if width > 60 {
			width = 60
		}
	}

	return &ProgressBar{
		writer:    os.Stderr,
		quiet:     quiet,
		width:     width,
		startTime: time.Now(),
	}
}

// Start begins tracking progress for a transfer of totalBytes.
func (p *ProgressBar) Start(totalBytes int64) {
	if p.quiet {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.total = totalBytes
	p.transferred = 0
	p.startTime = time.Now()
	p.lastUpdate = p.startTime
	p.done.Store(false)

	p.render()
}

// Update updates the progress with bytes transferred.
func (p *ProgressBar) Update(transferred int64) {
	if p.quiet {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.transferred = transferred
	now := time.Now()

	// Throttle updates to avoid flickering (max 10 updates/sec)
	if now.Sub(p.lastUpdate) < 100*time.Millisecond {
		return
	}
	p.lastUpdate = now

	p.render()
}

// Finish completes the progress bar.
func (p *ProgressBar) Finish() {
	if p.quiet || p.done.Load() {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.done.Store(true)
	p.transferred = p.total
	p.render()
	_, _ = fmt.Fprintln(p.writer) // New line after progress bar
}

// render draws the progress bar.
func (p *ProgressBar) render() {
	if p.total <= 0 {
		// Indeterminate progress (spinner style)
		elapsed := time.Since(p.startTime)
		spinner := []string{"|", "/", "-", "\\"}
		idx := int(elapsed.Seconds()*10) % len(spinner)
		_, _ = fmt.Fprintf(p.writer, "\r%s Uploading...", spinner[idx])
		return
	}

	// Calculate percentage
	percent := float64(p.transferred) * 100.0 / float64(p.total)
	if percent > 100 {
		percent = 100
	}

	// Calculate transfer speed
	elapsed := time.Since(p.startTime).Seconds()
	var speed float64
	if elapsed > 0 {
		speed = float64(p.transferred) / elapsed
	}

	// Draw bar
	filled := int(percent * float64(p.width) / 100.0)
	if filled > p.width {
		filled = p.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.width-filled)

	// Format bytes
	transferredStr := formatBytes(p.transferred)
	totalStr := formatBytes(p.total)
	speedStr := formatBytes(int64(speed)) + "/s"

	// Format output: [██████░░░░] 45% 4.5 MB / 10 MB (2.3 MB/s)
	output := fmt.Sprintf("\r[%s] %.0f%% %s / %s (%s)",
		bar, percent, transferredStr, totalStr, speedStr)

	_, _ = fmt.Fprint(p.writer, output)
}

// Callback returns a ProgressCallback function that updates this bar.
func (p *ProgressBar) Callback() ProgressCallback {
	return func(transferred, total int64, speed float64) {
		if p.total == 0 {
			p.Start(total)
		}
		p.Update(transferred)
	}
}

// formatBytes converts bytes to human-readable format.
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// getTerminalWidth attempts to get the terminal width.
func getTerminalWidth() (int, error) {
	// Try to use terminal size detection
	// Fallback to 80 if unavailable
	return 80, nil
}

// ProgressReader wraps an io.Reader to report progress.
type ProgressReader struct {
	r     io.Reader
	total int64
	bar   *ProgressBar
	read  int64
}

// NewProgressReader creates a reader that reports progress.
func NewProgressReader(r io.Reader, total int64, bar *ProgressBar) *ProgressReader {
	return &ProgressReader{
		r:     r,
		total: total,
		bar:   bar,
	}
}

// Read implements io.Reader and updates progress.
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.r.Read(p)
	if n > 0 {
		pr.read += int64(n)
		pr.bar.Update(pr.read)
	}
	return n, err
}

// WithProgress executes a function with a progress callback, respecting the quiet flag.
func WithProgress(quiet bool, total int64, fn func(ProgressCallback) error) error {
	bar := NewProgressBar(quiet)
	bar.Start(total)

	callback := bar.Callback()
	err := fn(callback)

	bar.Finish()
	return err
}
