package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google-play-cli/gpd/internal/errors"
)

func TestNewResult(t *testing.T) {
	data := map[string]interface{}{"key": "value"}
	result := NewResult(data)

	if result.Data == nil {
		t.Error("Data should not be nil")
	}
	if result.Error != nil {
		t.Error("Error should be nil for success result")
	}
	if result.Meta == nil {
		t.Error("Meta should not be nil")
	}
	if result.ExitCode != errors.ExitSuccess {
		t.Errorf("ExitCode = %d, want %d", result.ExitCode, errors.ExitSuccess)
	}
}

func TestNewErrorResult(t *testing.T) {
	err := errors.NewAPIError(errors.CodeValidationError, "invalid input")
	result := NewErrorResult(err)

	if result.Data != nil {
		t.Error("Data should be nil for error result")
	}
	if result.Error == nil {
		t.Error("Error should not be nil")
	}
	if result.ExitCode != errors.ExitValidationError {
		t.Errorf("ExitCode = %d, want %d", result.ExitCode, errors.ExitValidationError)
	}
}

func TestResultWithDuration(t *testing.T) {
	result := NewResult(nil)
	result.WithDuration(100 * time.Millisecond)

	if result.Meta.DurationMs != 100 {
		t.Errorf("DurationMs = %d, want 100", result.Meta.DurationMs)
	}
}

func TestResultWithServices(t *testing.T) {
	result := NewResult(nil)
	result.WithServices("androidpublisher", "playdeveloperreporting")

	if len(result.Meta.Services) != 2 {
		t.Errorf("Services count = %d, want 2", len(result.Meta.Services))
	}
}

func TestResultWithNoOp(t *testing.T) {
	result := NewResult(nil)
	result.WithNoOp("already uploaded")

	if !result.Meta.NoOp {
		t.Error("NoOp should be true")
	}
	if result.Meta.NoOpReason != "already uploaded" {
		t.Errorf("NoOpReason = %q, want 'already uploaded'", result.Meta.NoOpReason)
	}
}

func TestResultWithPagination(t *testing.T) {
	result := NewResult(nil)
	result.WithPagination("token1", "token2")

	if result.Meta.PageToken != "token1" {
		t.Errorf("PageToken = %q, want 'token1'", result.Meta.PageToken)
	}
	if result.Meta.NextPageToken != "token2" {
		t.Errorf("NextPageToken = %q, want 'token2'", result.Meta.NextPageToken)
	}
}

func TestOutputManagerJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf)

	result := NewResult(map[string]interface{}{"key": "value"})
	result.WithServices("test")

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Check envelope structure
	if _, ok := parsed["data"]; !ok {
		t.Error("Output missing 'data' field")
	}
	if _, ok := parsed["error"]; !ok {
		t.Error("Output missing 'error' field")
	}
	if _, ok := parsed["meta"]; !ok {
		t.Error("Output missing 'meta' field")
	}
}

func TestOutputManagerPrettyJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetPretty(true)

	result := NewResult(map[string]interface{}{"key": "value"})

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	// Pretty JSON should have newlines
	if !strings.Contains(output, "\n") {
		t.Error("Pretty JSON should contain newlines")
	}
}

func TestOutputManagerMinifiedJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf) // Default is minified

	result := NewResult(map[string]interface{}{"key": "value"})

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := strings.TrimSpace(buf.String())
	// Minified JSON should be a single line
	if strings.Count(output, "\n") > 0 {
		t.Error("Minified JSON should be a single line")
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected Format
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"table", FormatTable},
		{"TABLE", FormatTable},
		{"markdown", FormatMarkdown},
		{"md", FormatMarkdown},
		{"csv", FormatCSV},
		{"invalid", FormatJSON}, // Default
		{"", FormatJSON},        // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseFormat(tt.input); got != tt.expected {
				t.Errorf("ParseFormat(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestJSONEnvelopeStructure(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf)

	// Test success response
	result := NewResult(map[string]interface{}{"test": true})
	result.WithServices("androidpublisher")
	result.WithDuration(50 * time.Millisecond)

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var envelope struct {
		Data  map[string]interface{} `json:"data"`
		Error *errors.APIError       `json:"error"`
		Meta  *Metadata              `json:"meta"`
	}

	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		t.Fatalf("Failed to parse envelope: %v", err)
	}

	if envelope.Data == nil {
		t.Error("data should not be nil in success response")
	}
	if envelope.Error != nil {
		t.Error("error should be nil in success response")
	}
	if envelope.Meta == nil {
		t.Error("meta should not be nil")
	}
	if len(envelope.Meta.Services) == 0 {
		t.Error("meta.services should not be empty")
	}
}

func TestErrorEnvelopeStructure(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf)

	err := errors.NewAPIError(errors.CodeValidationError, "bad input").
		WithHint("check your parameters")
	result := NewErrorResult(err)

	if writeErr := mgr.Write(result); writeErr != nil {
		t.Fatalf("Write() error = %v", writeErr)
	}

	var envelope struct {
		Data  interface{}      `json:"data"`
		Error *errors.APIError `json:"error"`
		Meta  *Metadata        `json:"meta"`
	}

	if parseErr := json.Unmarshal(buf.Bytes(), &envelope); parseErr != nil {
		t.Fatalf("Failed to parse envelope: %v", parseErr)
	}

	if envelope.Data != nil {
		t.Error("data should be nil in error response")
	}
	if envelope.Error == nil {
		t.Error("error should not be nil in error response")
	}
	if envelope.Error.Code != errors.CodeValidationError {
		t.Errorf("error.code = %v, want %v", envelope.Error.Code, errors.CodeValidationError)
	}
	if envelope.Error.Message != "bad input" {
		t.Errorf("error.message = %v, want 'bad input'", envelope.Error.Message)
	}
	if envelope.Error.Hint != "check your parameters" {
		t.Errorf("error.hint = %v, want 'check your parameters'", envelope.Error.Hint)
	}
}
