//go:build unit
// +build unit

package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestProgressBar(t *testing.T) {
	t.Run("quiet mode suppresses output", func(t *testing.T) {
		var buf bytes.Buffer
		bar := NewProgressBar(true)
		bar.writer = &buf
		bar.Start(1000)
		bar.Update(500)
		bar.Finish()

		if buf.String() != "" {
			t.Errorf("Expected no output in quiet mode, got: %q", buf.String())
		}
	})

	t.Run("shows progress bar", func(t *testing.T) {
		var buf bytes.Buffer
		bar := NewProgressBar(false)
		bar.writer = &buf
		bar.width = 20 // Small width for testing
		bar.Start(1000)
		bar.Finish()

		output := buf.String()
		// Just verify it renders a bar with percentage
		if !strings.Contains(output, "%") {
			t.Errorf("Expected output to contain %%, got: %q", output)
		}
		if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
			t.Errorf("Expected output to contain progress bar brackets, got: %q", output)
		}
	})

	t.Run("format bytes", func(t *testing.T) {
		tests := []struct {
			bytes    int64
			expected string
		}{
			{500, "500 B"},
			{1024, "1.0 KB"},
			{1024 * 1024, "1.00 MB"},
			{1024 * 1024 * 1024, "1.00 GB"},
		}

		for _, tc := range tests {
			result := formatBytes(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tc.bytes, result, tc.expected)
			}
		}
	})
}

func TestProgressCallback(t *testing.T) {
	t.Run("callback initializes and shows progress", func(t *testing.T) {
		var buf bytes.Buffer
		bar := NewProgressBar(false)
		bar.writer = &buf
		bar.width = 20

		callback := bar.Callback()
		// Callback should initialize the bar with total
		callback(500, 1000, 100.0)

		output := buf.String()
		// Callback calls Start which renders initial state (0%%)
		if !strings.Contains(output, "%") {
			t.Errorf("Expected callback to show progress, got: %q", output)
		}
	})
}

func TestProgressReader(t *testing.T) {
	t.Run("reads and reports progress", func(t *testing.T) {
		var buf bytes.Buffer
		bar := NewProgressBar(false)
		bar.writer = &buf
		bar.width = 20
		bar.Start(13) // "Hello, World!" is 13 bytes

		data := strings.NewReader("Hello, World!")
		reader := NewProgressReader(data, 13, bar)

		result, err := io.ReadAll(reader)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if string(result) != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got: %q", string(result))
		}
	})
}
