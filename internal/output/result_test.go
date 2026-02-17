package output

import (
	"bytes"
	"encoding/json"
	stdErrors "errors"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/gpd/internal/errors"
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
	if result.Meta.HasMorePages == nil || !*result.Meta.HasMorePages {
		t.Errorf("HasMorePages = %v, want true", result.Meta.HasMorePages)
	}
}

func TestResultWithPaginationNoNext(t *testing.T) {
	result := NewResult(nil)
	result.WithPagination("token1", "")

	if result.Meta.HasMorePages == nil || *result.Meta.HasMorePages {
		t.Errorf("HasMorePages = %v, want false", result.Meta.HasMorePages)
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

func TestTableFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected []string // Strings that should appear in output
	}{
		{
			name: "slice_of_maps",
			data: []interface{}{
				map[string]interface{}{"name": "app1", "version": 1},
				map[string]interface{}{"name": "app2", "version": 2},
			},
			expected: []string{"name", "version", "app1", "app2"},
		},
		{
			name:     "single_map",
			data:     map[string]interface{}{"key1": "value1", "key2": "value2"},
			expected: []string{"key1", "value1", "key2", "value2"},
		},
		{
			name:     "empty_slice",
			data:     []interface{}{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			mgr := NewManager(&buf).SetFormat(FormatTable)

			result := NewResult(tt.data)
			if err := mgr.Write(result); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			output := buf.String()
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestTableFormatError(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatTable)

	err := errors.NewAPIError(errors.CodeValidationError, "test error")
	result := NewErrorResult(err)

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Errors should fall back to JSON even in table mode
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Error output should be valid JSON: %v", err)
	}
}

func TestMarkdownFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected []string
	}{
		{
			name: "slice_of_maps",
			data: []interface{}{
				map[string]interface{}{"name": "app1", "version": 1},
				map[string]interface{}{"name": "app2", "version": 2},
			},
			expected: []string{"|", "name", "version", "---", "app1", "app2"},
		},
		{
			name:     "single_map",
			data:     map[string]interface{}{"key": "value"},
			expected: []string{"- **key:**", "value"},
		},
		{
			name:     "empty_slice",
			data:     []interface{}{},
			expected: []string{"*No data*"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			mgr := NewManager(&buf).SetFormat(FormatMarkdown)

			result := NewResult(tt.data)
			if err := mgr.Write(result); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			output := buf.String()
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestMarkdownFormatError(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatMarkdown)

	err := errors.NewAPIError(errors.CodeValidationError, "test error").
		WithHint("test hint")
	result := NewErrorResult(err)

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "## Error") {
		t.Error("Markdown error output should contain '## Error' header")
	}
	if !strings.Contains(output, "test error") {
		t.Error("Markdown error output should contain error message")
	}
	if !strings.Contains(output, "test hint") {
		t.Error("Markdown error output should contain hint")
	}
}

func TestCSVFormat(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatCSV)

	data := []interface{}{
		map[string]interface{}{"name": "app1", "count": 10},
		map[string]interface{}{"name": "app2", "count": 20},
	}
	result := NewResult(data)

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines (header + data), got %d", len(lines))
	}

	// Check that it's comma-separated
	if !strings.Contains(lines[0], ",") {
		t.Error("CSV header should contain commas")
	}
}

func TestCSVFormatWithSpecialCharacters(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatCSV)

	data := []interface{}{
		map[string]interface{}{"name": "app, with comma", "desc": "contains \"quotes\""},
	}
	result := NewResult(data)

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	// Check that special characters are properly escaped
	if !strings.Contains(output, "\"") {
		t.Error("CSV should escape fields with special characters")
	}
}

func TestCSVFormatEmpty(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatCSV)

	result := NewResult([]interface{}{})
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Error("CSV output for empty data should be empty")
	}
}

func TestResultWithWarnings(t *testing.T) {
	result := NewResult(nil)
	result.WithWarnings("warning 1", "warning 2")

	if len(result.Meta.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(result.Meta.Warnings))
	}
	if result.Meta.Warnings[0] != "warning 1" {
		t.Errorf("First warning = %q, want 'warning 1'", result.Meta.Warnings[0])
	}
}

func TestResultWithPartial(t *testing.T) {
	result := NewResult(nil)
	result.WithPartial(100, 50, 200)

	if !result.Meta.Partial {
		t.Error("Partial should be true")
	}
	if result.Meta.ScannedCount != 100 {
		t.Errorf("ScannedCount = %d, want 100", result.Meta.ScannedCount)
	}
	if result.Meta.FilteredCount != 50 {
		t.Errorf("FilteredCount = %d, want 50", result.Meta.FilteredCount)
	}
	if result.Meta.TotalAvailable != 200 {
		t.Errorf("TotalAvailable = %d, want 200", result.Meta.TotalAvailable)
	}
}

func TestResultWithRetries(t *testing.T) {
	result := NewResult(nil)
	result.WithRetries(3)

	if result.Meta.Retries != 3 {
		t.Errorf("Retries = %d, want 3", result.Meta.Retries)
	}
}

func TestResultWithRequestID(t *testing.T) {
	result := NewResult(nil)
	result.WithRequestID("req-12345")

	if result.Meta.RequestID != "req-12345" {
		t.Errorf("RequestID = %q, want 'req-12345'", result.Meta.RequestID)
	}
}

func TestNewEmptyResult(t *testing.T) {
	result := NewEmptyResult()

	if result.Data != nil {
		t.Error("Data should be nil for empty result")
	}
	if result.Error != nil {
		t.Error("Error should be nil for empty result")
	}
	if result.ExitCode != errors.ExitSuccess {
		t.Errorf("ExitCode = %d, want %d", result.ExitCode, errors.ExitSuccess)
	}
}

func TestResultChaining(t *testing.T) {
	// Test that result methods return the result for chaining
	result := NewResult(map[string]interface{}{"test": true}).
		WithDuration(100*time.Millisecond).
		WithServices("service1", "service2").
		WithWarnings("warning1").
		WithRetries(2)

	if result.Meta.DurationMs != 100 {
		t.Error("Duration not set correctly in chain")
	}
	if len(result.Meta.Services) != 2 {
		t.Error("Services not set correctly in chain")
	}
	if len(result.Meta.Warnings) != 1 {
		t.Error("Warnings not set correctly in chain")
	}
	if result.Meta.Retries != 2 {
		t.Error("Retries not set correctly in chain")
	}
}

func TestSetFormatAndPretty(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatJSON).SetPretty(true)

	result := NewResult(map[string]interface{}{"test": true})
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	// Should be pretty JSON
	if !strings.Contains(output, "\n") {
		t.Error("Should have newlines for pretty JSON")
	}
}

func TestTableFormatNilData(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatTable)

	result := NewResult(nil)
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Should not panic - empty output is acceptable for nil data
	_ = buf.String()
}

func TestMarkdownFormatNilData(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatMarkdown)

	result := NewResult(nil)
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Should not panic - empty output is acceptable for nil data
	_ = buf.String()
}

func TestCSVFormatError(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatCSV)

	err := errors.NewAPIError(errors.CodeValidationError, "test error")
	result := NewErrorResult(err)

	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Errors should fall back to JSON even in CSV mode
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Error output should be valid JSON: %v", err)
	}
}

func TestSetFieldsAppliesProjection(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFields([]string{"data.key"})
	result := NewResult(map[string]interface{}{"key": "value"})
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}
	if _, ok := parsed["data"]; !ok {
		t.Error("Output missing data field")
	}
}

func TestTableFormatSliceWithNonMapFallbacksToJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatTable)
	result := NewResult([]interface{}{"value"})
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Fallback output should be JSON: %v", err)
	}
}

func TestTableFormatUnsupportedTypeFallbacksToJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatTable)
	result := NewResult(42)
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Fallback output should be JSON: %v", err)
	}
}

func TestMarkdownFormatUnsupportedTypeFallbacksToJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatMarkdown)
	result := NewResult(42)
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Fallback output should be JSON: %v", err)
	}
}

func TestMarkdownFormatErrorWithoutHint(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatMarkdown)
	err := errors.NewAPIError(errors.CodeValidationError, "test error")
	result := NewErrorResult(err)
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Hint") {
		t.Error("Markdown error output should not include hint")
	}
}

func TestCSVFormatNonSliceFallbacksToJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatCSV)
	result := NewResult(map[string]interface{}{"key": "value"})
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Fallback output should be JSON: %v", err)
	}
}

func TestCSVFormatSliceWithNonMapFallbacksToJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatCSV)
	result := NewResult([]interface{}{"value"})
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Fallback output should be JSON: %v", err)
	}
}

func TestWriteUnknownFormatDefaultsToJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(Format("unknown"))
	result := NewResult(map[string]interface{}{"key": "value"})
	if err := mgr.Write(result); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Output should be JSON: %v", err)
	}
}

func TestWriteJSONReturnsErrorOnInvalidData(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf)
	result := &Result{Data: func() {}, Meta: &Metadata{}}
	if err := mgr.Write(result); err == nil {
		t.Fatal("Expected error from JSON marshal")
	}
}

func TestWithMetaNilPaths(t *testing.T) {
	r1 := &Result{}
	r1.WithDuration(5 * time.Millisecond)
	if r1.Meta == nil || r1.Meta.DurationMs != 5 {
		t.Fatal("WithDuration should initialize meta")
	}

	r2 := &Result{}
	r2.WithServices("a", "b")
	if r2.Meta == nil || len(r2.Meta.Services) != 2 {
		t.Fatal("WithServices should initialize meta")
	}

	r3 := &Result{}
	r3.WithNoOp("reason")
	if r3.Meta == nil || !r3.Meta.NoOp || r3.Meta.NoOpReason != "reason" {
		t.Fatal("WithNoOp should initialize meta")
	}

	r4 := &Result{}
	r4.WithPagination("p", "n")
	if r4.Meta == nil || r4.Meta.PageToken != "p" || r4.Meta.NextPageToken != "n" {
		t.Fatal("WithPagination should initialize meta")
	}

	r5 := &Result{}
	r5.WithWarnings("w1", "w2")
	if r5.Meta == nil || len(r5.Meta.Warnings) != 2 {
		t.Fatal("WithWarnings should initialize meta")
	}

	r6 := &Result{}
	r6.WithPartial(1, 2, 3)
	if r6.Meta == nil || !r6.Meta.Partial {
		t.Fatal("WithPartial should initialize meta")
	}

	r7 := &Result{}
	r7.WithRetries(2)
	if r7.Meta == nil || r7.Meta.Retries != 2 {
		t.Fatal("WithRetries should initialize meta")
	}

	r8 := &Result{}
	r8.WithRequestID("req")
	if r8.Meta == nil || r8.Meta.RequestID != "req" {
		t.Fatal("WithRequestID should initialize meta")
	}
}

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) {
	return 0, stdErrors.New("write error")
}

type failAfterWriter struct {
	writes int
	limit  int
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes > w.limit {
		return 0, stdErrors.New("write error")
	}
	return len(p), nil
}

func TestWriteTableSliceWriteError(t *testing.T) {
	mgr := NewManager(errWriter{}).SetFormat(FormatTable)
	err := mgr.writeTableSlice([]interface{}{map[string]interface{}{"a": "b"}})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteTableSliceSeparatorWriteError(t *testing.T) {
	writer := &failAfterWriter{limit: 1}
	mgr := NewManager(writer).SetFormat(FormatTable)
	err := mgr.writeTableSlice([]interface{}{map[string]interface{}{"a": "b"}})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteTableSliceRowWriteError(t *testing.T) {
	writer := &failAfterWriter{limit: 2}
	mgr := NewManager(writer).SetFormat(FormatTable)
	err := mgr.writeTableSlice([]interface{}{map[string]interface{}{"a": "b"}})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteTableSliceSkipsInvalidRows(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatTable)
	err := mgr.writeTableSlice([]interface{}{map[string]interface{}{"a": "b"}, "skip"})
	if err != nil {
		t.Fatalf("writeTableSlice() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "a") {
		t.Fatal("Expected output to include header")
	}
}

func TestWriteTableMapWriteError(t *testing.T) {
	mgr := NewManager(errWriter{}).SetFormat(FormatTable)
	err := mgr.writeTableMap(map[string]interface{}{"a": "b"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteMarkdownTableWriteError(t *testing.T) {
	mgr := NewManager(errWriter{}).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdownTable([]interface{}{map[string]interface{}{"a": "b"}})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteMarkdownTableSeparatorWriteError(t *testing.T) {
	writer := &failAfterWriter{limit: 1}
	mgr := NewManager(writer).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdownTable([]interface{}{map[string]interface{}{"a": "b"}})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteMarkdownTableRowWriteError(t *testing.T) {
	writer := &failAfterWriter{limit: 2}
	mgr := NewManager(writer).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdownTable([]interface{}{map[string]interface{}{"a": "b"}})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteMarkdownTableFallbacksToJSON(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdownTable([]interface{}{"value"})
	if err != nil {
		t.Fatalf("writeMarkdownTable() error = %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Fallback output should be JSON: %v", err)
	}
}

func TestWriteMarkdownTableEmptyWriteError(t *testing.T) {
	mgr := NewManager(errWriter{}).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdownTable([]interface{}{})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteMarkdownTableSkipsInvalidRows(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdownTable([]interface{}{
		map[string]interface{}{"a": "b"},
		"skip",
	})
	if err != nil {
		t.Fatalf("writeMarkdownTable() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "| a |") {
		t.Fatal("Expected markdown table output")
	}
}

func TestWriteMarkdownMapWriteError(t *testing.T) {
	mgr := NewManager(errWriter{}).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdownMap(map[string]interface{}{"a": "b"})
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteMarkdownErrorWriteFailure(t *testing.T) {
	mgr := NewManager(errWriter{}).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdown(NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "bad")))
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteMarkdownErrorHintWriteFailure(t *testing.T) {
	writer := &failAfterWriter{limit: 1}
	mgr := NewManager(writer).SetFormat(FormatMarkdown)
	err := mgr.writeMarkdown(NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "bad").WithHint("hint")))
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteCSVWriteErrorOnHeader(t *testing.T) {
	mgr := NewManager(errWriter{}).SetFormat(FormatCSV)
	err := mgr.writeCSV(NewResult([]interface{}{map[string]interface{}{"a": "b"}}))
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteCSVWriteErrorOnRow(t *testing.T) {
	writer := &failAfterWriter{limit: 1}
	mgr := NewManager(writer).SetFormat(FormatCSV)
	err := mgr.writeCSV(NewResult([]interface{}{
		map[string]interface{}{"a": "b"},
	}))
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestWriteCSVSkipsInvalidRows(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatCSV)
	err := mgr.writeCSV(NewResult([]interface{}{
		map[string]interface{}{"a": "b"},
		"skip",
	}))
	if err != nil {
		t.Fatalf("writeCSV() error = %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "a") {
		t.Fatal("Expected CSV header")
	}
}

func TestWriteCSVNilData(t *testing.T) {
	var buf bytes.Buffer
	mgr := NewManager(&buf).SetFormat(FormatCSV)
	err := mgr.writeCSV(NewResult(nil))
	if err != nil {
		t.Fatalf("writeCSV() error = %v", err)
	}
	if buf.Len() != 0 {
		t.Fatal("Expected empty output")
	}
}
