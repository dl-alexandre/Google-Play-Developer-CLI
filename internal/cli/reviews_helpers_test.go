package cli

import (
	"testing"
	"time"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func TestStringValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		fallback string
		want     string
	}{
		{
			name:     "string value",
			value:    "hello",
			fallback: "default",
			want:     "hello",
		},
		{
			name:     "nil value",
			value:    nil,
			fallback: "default",
			want:     "default",
		},
		{
			name:     "int value",
			value:    42,
			fallback: "default",
			want:     "42",
		},
		{
			name:     "empty string value",
			value:    "",
			fallback: "default",
			want:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringValue(tt.value, tt.fallback)
			if got != tt.want {
				t.Errorf("stringValue() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "valid email",
			email: "user@example.com",
			want:  true,
		},
		{
			name:  "valid email with name",
			email: "User Name <user@example.com>",
			want:  true,
		},
		{
			name:  "invalid email",
			email: "not-an-email",
			want:  false,
		},
		{
			name:  "empty email",
			email: "",
			want:  false,
		},
		{
			name:  "email without domain",
			email: "user@",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("isValidEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "valid URL",
			url:  "https://example.com",
			want: true,
		},
		{
			name: "valid URL with path",
			url:  "https://example.com/path",
			want: true,
		},
		{
			name: "invalid URL no scheme",
			url:  "example.com",
			want: false,
		},
		{
			name: "invalid URL no host",
			url:  "https://",
			want: false,
		},
		{
			name: "empty URL",
			url:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidURL(tt.url)
			if got != tt.want {
				t.Errorf("isValidURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	items := []string{"a", "b", "c"}

	if !containsString(items, "a") {
		t.Error("containsString should find 'a'")
	}
	if !containsString(items, "b") {
		t.Error("containsString should find 'b'")
	}
	if containsString(items, "d") {
		t.Error("containsString should not find 'd'")
	}
	if containsString(nil, "a") {
		t.Error("containsString should return false for nil slice")
	}
}

func TestMergeVersionCodes(t *testing.T) {
	tests := []struct {
		name     string
		primary  []int64
		retain   []int64
		expected []int64
	}{
		{
			name:     "no duplicates",
			primary:  []int64{1, 2, 3},
			retain:   []int64{4, 5, 6},
			expected: []int64{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "with duplicates",
			primary:  []int64{1, 2, 3},
			retain:   []int64{2, 3, 4},
			expected: []int64{1, 2, 3, 4},
		},
		{
			name:     "all duplicates",
			primary:  []int64{1, 2},
			retain:   []int64{1, 2},
			expected: []int64{1, 2},
		},
		{
			name:     "empty primary",
			primary:  []int64{},
			retain:   []int64{1, 2},
			expected: []int64{1, 2},
		},
		{
			name:     "empty retain",
			primary:  []int64{1, 2},
			retain:   []int64{},
			expected: []int64{1, 2},
		},
		{
			name:     "both empty",
			primary:  []int64{},
			retain:   []int64{},
			expected: []int64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeVersionCodes(tt.primary, tt.retain)
			if len(got) != len(tt.expected) {
				t.Errorf("mergeVersionCodes() length = %d, want %d", len(got), len(tt.expected))
				return
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("mergeVersionCodes()[%d] = %d, want %d", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestOutputNewResult(t *testing.T) {
	result := output.NewResult(map[string]string{"key": "value"})
	if result.Data == nil {
		t.Error("Data should not be nil")
	}
	if result.Error != nil {
		t.Error("Error should be nil")
	}
	if result.Meta == nil {
		t.Error("Meta should not be nil")
	}
}

func TestOutputNewErrorResult(t *testing.T) {
	err := errors.NewAPIError(errors.CodeNotFound, "not found")
	result := output.NewErrorResult(err)
	if result.Data != nil {
		t.Error("Data should be nil")
	}
	if result.Error == nil {
		t.Error("Error should not be nil")
	}
}

func TestOutputNewEmptyResult(t *testing.T) {
	result := output.NewEmptyResult()
	if result.Data != nil {
		t.Error("Data should be nil")
	}
	if result.Error != nil {
		t.Error("Error should be nil")
	}
}

func TestOutputResultChaining(t *testing.T) {
	result := output.NewResult(nil).
		WithDuration(100 * time.Millisecond).
		WithServices("test").
		WithWarnings("warn1")

	if result.Meta.DurationMs != 100 {
		t.Errorf("DurationMs = %d, want 100", result.Meta.DurationMs)
	}
	if len(result.Meta.Services) != 1 {
		t.Errorf("Services = %d, want 1", len(result.Meta.Services))
	}
	if len(result.Meta.Warnings) != 1 {
		t.Errorf("Warnings = %d, want 1", len(result.Meta.Warnings))
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://example.com",
			wantErr: false,
		},
		{
			name:    "valid URL with path",
			url:     "https://example.com/path/to/resource",
			wantErr: false,
		},
		{
			name:    "valid URL with query",
			url:     "https://example.com/path?query=value",
			wantErr: false,
		},
		{
			name:    "http not allowed",
			url:     "http://example.com",
			wantErr: true,
		},
		{
			name:    "no scheme",
			url:     "example.com",
			wantErr: true,
		},
		{
			name:    "no host",
			url:     "https://",
			wantErr: true,
		},
		{
			name:    "loopback not allowed",
			url:     "https://127.0.0.1",
			wantErr: true,
		},
		{
			name:    "private IP not allowed",
			url:     "https://192.168.1.1",
			wantErr: true,
		},
		{
			name:    "public IP allowed",
			url:     "https://8.8.8.8",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePromoteInput(t *testing.T) {
	tests := []struct {
		name      string
		fromTrack string
		toTrack   string
		wantNil   bool
	}{
		{
			name:      "valid promote internal to beta",
			fromTrack: "internal",
			toTrack:   "beta",
			wantNil:   true,
		},
		{
			name:      "valid promote beta to production",
			fromTrack: "beta",
			toTrack:   "production",
			wantNil:   true,
		},
		{
			name:      "same track not allowed",
			fromTrack: "internal",
			toTrack:   "internal",
			wantNil:   false,
		},
		{
			name:      "invalid from track",
			fromTrack: "invalid",
			toTrack:   "production",
			wantNil:   false,
		},
		{
			name:      "invalid to track",
			fromTrack: "internal",
			toTrack:   "invalid",
			wantNil:   false,
		},
		{
			name:      "both invalid tracks",
			fromTrack: "foo",
			toTrack:   "bar",
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePromoteInput(tt.fromTrack, tt.toTrack)
			isNil := err == nil
			if isNil != tt.wantNil {
				t.Errorf("validatePromoteInput() = %v, wantNil %v", err, tt.wantNil)
			}
		})
	}
}

func TestConfigIsValidTrack(t *testing.T) {
	validTracks := []string{"internal", "beta", "alpha", "production", "INTERNAL", "BETA"}
	invalidTracks := []string{"", "invalid", "foobar"}

	for _, track := range validTracks {
		if !config.IsValidTrack(track) {
			t.Errorf("IsValidTrack(%q) = false, want true", track)
		}
	}

	for _, track := range invalidTracks {
		if config.IsValidTrack(track) {
			t.Errorf("IsValidTrack(%q) = true, want false", track)
		}
	}
}

func TestAPIErrorHelpers(t *testing.T) {
	err := errors.NewAPIError(errors.CodeValidationError, "test error")
	if err.Code != errors.CodeValidationError {
		t.Errorf("Code = %v, want %v", err.Code, errors.CodeValidationError)
	}

	errWithHint := err.WithHint("test hint")
	if errWithHint.Hint != "test hint" {
		t.Errorf("Hint = %v, want %v", errWithHint.Hint, "test hint")
	}

	errWithDetails := err.WithDetails(map[string]interface{}{"key": "value"})
	if errWithDetails.Details == nil {
		t.Error("Details should not be nil")
	}
}

func TestReviewsListParams(t *testing.T) {
	params := reviewsListParams{
		minRating:       3,
		maxRating:       5,
		scanLimit:       100,
		includeText:     true,
		translationLang: "en",
		pageSize:        50,
		pageToken:       "token123",
		all:             true,
	}

	if params.minRating != 3 {
		t.Errorf("minRating = %d, want 3", params.minRating)
	}
	if params.all != true {
		t.Error("all flag should be true")
	}
}

func TestReviewsListEdgeCases(t *testing.T) {
	params := reviewsListParams{
		scanLimit: 0,
		all:       false,
	}

	paramsAll := reviewsListParams{
		scanLimit: 1000,
		all:       true,
	}

	if !paramsAll.all {
		t.Error("all flag should be true")
	}

	_ = params
}

func TestProcessTemplate(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		vars    map[string]string
		want    string
		wantErr bool
	}{
		{
			name:    "no variables",
			text:    "Hello world",
			vars:    map[string]string{},
			want:    "Hello world",
			wantErr: false,
		},
		{
			name:    "single variable",
			text:    "Hello {{appName}}",
			vars:    map[string]string{"appName": "MyApp"},
			want:    "Hello MyApp",
			wantErr: false,
		},
		{
			name:    "multiple variables",
			text:    "{{appName}} rated {{rating}} stars",
			vars:    map[string]string{"appName": "MyApp", "rating": "5"},
			want:    "MyApp rated 5 stars",
			wantErr: false,
		},
		{
			name:    "missing variable",
			text:    "Hello {{appName}} and {{missing}}",
			vars:    map[string]string{"appName": "MyApp"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "repeated variable",
			text:    "{{appName}} is great. {{appName}} is the best!",
			vars:    map[string]string{"appName": "MyApp"},
			want:    "MyApp is great. MyApp is the best!",
			wantErr: false,
		},
		{
			name:    "empty variable name",
			text:    "Hello {{}}",
			vars:    map[string]string{},
			want:    "Hello {{}}",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processTemplate(tt.text, tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("processTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHashReply(t *testing.T) {
	hash1 := hashReply("review-123", "Thanks for your feedback!")
	hash2 := hashReply("review-123", "Thanks for your feedback!")
	hash3 := hashReply("review-123", "Different reply")
	hash4 := hashReply("review-456", "Thanks for your feedback!")

	if hash1 != hash2 {
		t.Error("hashReply() not deterministic for same inputs")
	}

	if hash1 == hash3 {
		t.Error("hashReply() should produce different hashes for different text")
	}

	if hash1 == hash4 {
		t.Error("hashReply() should produce different hashes for different review IDs")
	}

	if len(hash1) != 16 {
		t.Errorf("hashReply() expected 16 chars, got %d", len(hash1))
	}
}

func TestParseYear(t *testing.T) {
	tests := []struct {
		name string
		date string
		want int64
	}{
		{"valid date", "2024-03-15", 2024},
		{"single digit year", "2024-01-01", 2024},
		{"invalid format", "invalid", 0},
		{"empty string", "", 0},
		{"year only", "2024", 2024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseYear(tt.date); got != tt.want {
				t.Errorf("parseYear(%q) = %d, want %d", tt.date, got, tt.want)
			}
		})
	}
}

func TestParseMonth(t *testing.T) {
	tests := []struct {
		name string
		date string
		want int64
	}{
		{"valid date", "2024-03-15", 3},
		{"single digit month", "2024-01-01", 1},
		{"invalid format", "invalid", 0},
		{"empty string", "", 0},
		{"year only", "2024", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseMonth(tt.date); got != tt.want {
				t.Errorf("parseMonth(%q) = %d, want %d", tt.date, got, tt.want)
			}
		})
	}
}

func TestParseDay(t *testing.T) {
	tests := []struct {
		name string
		date string
		want int64
	}{
		{"valid date", "2024-03-15", 15},
		{"single digit day", "2024-01-01", 1},
		{"invalid format", "invalid", 0},
		{"empty string", "", 0},
		{"year only", "2024", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDay(tt.date); got != tt.want {
				t.Errorf("parseDay(%q) = %d, want %d", tt.date, got, tt.want)
			}
		})
	}
}

func TestFormatReportText(t *testing.T) {
	tests := []struct {
		name   string
		report string
		want   string
	}{
		{
			name:   "multiple lines",
			report: "line1\nline2\nline3",
			want:   "  1: line1\n  2: line2\n  3: line3",
		},
		{
			name:   "empty lines skipped",
			report: "line1\n\nline2",
			want:   "  1: line1\n  3: line2",
		},
		{
			name:   "whitespace only returns original",
			report: "   \n   ",
			want:   "   \n   ",
		},
		{
			name:   "single line",
			report: "single",
			want:   "  1: single",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatReportText(tt.report); got != tt.want {
				t.Errorf("formatReportText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTimeMillis(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    int64
		wantErr bool
	}{
		{
			name:    "milliseconds",
			value:   "1704067200000",
			want:    1704067200000,
			wantErr: false,
		},
		{
			name:    "RFC3339 format",
			value:   "2024-01-01T00:00:00Z",
			want:    1704067200000,
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid format",
			value:   "not-a-time",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeMillis(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeMillis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseTimeMillis() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidateDetailsPatchInput(t *testing.T) {
	tests := []struct {
		name            string
		email           string
		phone           string
		website         string
		defaultLanguage string
		wantNil         bool
	}{
		{
			name:            "valid email",
			email:           "test@example.com",
			phone:           "",
			website:         "",
			defaultLanguage: "",
			wantNil:         true,
		},
		{
			name:            "valid website",
			email:           "",
			phone:           "",
			website:         "https://example.com",
			defaultLanguage: "",
			wantNil:         true,
		},
		{
			name:            "all empty",
			email:           "",
			phone:           "",
			website:         "",
			defaultLanguage: "",
			wantNil:         false,
		},
		{
			name:            "invalid email",
			email:           "invalid",
			phone:           "",
			website:         "",
			defaultLanguage: "",
			wantNil:         false,
		},
		{
			name:            "invalid website",
			email:           "",
			phone:           "",
			website:         "not-a-url",
			defaultLanguage: "",
			wantNil:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDetailsPatchInput(tt.email, tt.phone, tt.website, tt.defaultLanguage)
			isNil := err == nil
			if isNil != tt.wantNil {
				t.Errorf("validateDetailsPatchInput() = %v, wantNil %v", err, tt.wantNil)
			}
		})
	}
}

func TestBuildUpdateMask(t *testing.T) {
	tests := []struct {
		name            string
		email           string
		phone           string
		website         string
		defaultLanguage string
		existingMask    string
		want            string
	}{
		{
			name:            "single field",
			email:           "test@example.com",
			phone:           "",
			website:         "",
			defaultLanguage: "",
			existingMask:    "",
			want:            "contactEmail",
		},
		{
			name:            "multiple fields",
			email:           "test@example.com",
			phone:           "123",
			website:         "",
			defaultLanguage: "",
			existingMask:    "",
			want:            "contactEmail,contactPhone",
		},
		{
			name:            "uses existing mask",
			email:           "test@example.com",
			phone:           "",
			website:         "",
			defaultLanguage: "",
			existingMask:    "customMask",
			want:            "customMask",
		},
		{
			name:            "empty fields",
			email:           "",
			phone:           "",
			website:         "",
			defaultLanguage: "",
			existingMask:    "",
			want:            "",
		},
		{
			name:            "all fields",
			email:           "test@example.com",
			phone:           "123",
			website:         "https://example.com",
			defaultLanguage: "en-US",
			existingMask:    "",
			want:            "contactEmail,contactPhone,contactWebsite,defaultLanguage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildUpdateMask(tt.email, tt.phone, tt.website, tt.defaultLanguage, tt.existingMask)
			if got != tt.want {
				t.Errorf("buildUpdateMask() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPermissionListString(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{
			name:  "string slice",
			value: []string{"a", "b", "c"},
			want:  "a,b,c",
		},
		{
			name:  "empty string slice",
			value: []string{},
			want:  "-",
		},
		{
			name:  "interface slice",
			value: []interface{}{"x", "y"},
			want:  "x,y",
		},
		{
			name:  "empty interface slice",
			value: []interface{}{},
			want:  "-",
		},
		{
			name:  "not a slice",
			value: "string",
			want:  "-",
		},
		{
			name:  "interface slice with empty",
			value: []interface{}{"", "a", ""},
			want:  "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := permissionListString(tt.value); got != tt.want {
				t.Errorf("permissionListString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJsonPath(t *testing.T) {
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"value": "found",
			},
			"array": []interface{}{"a", "b", "c"},
		},
		"simple": "hello",
	}

	tests := []struct {
		name  string
		value interface{}
		path  string
		want  string
	}{
		{
			name:  "nested path",
			value: data,
			path:  "level1.level2.value",
			want:  "found",
		},
		{
			name:  "simple path",
			value: data,
			path:  "simple",
			want:  "hello",
		},
		{
			name:  "nonexistent path",
			value: data,
			path:  "level1.missing",
			want:  "",
		},
		{
			name:  "nil value",
			value: nil,
			path:  "anything",
			want:  "",
		},
		{
			name:  "empty path",
			value: data,
			path:  "",
			want:  "",
		},
		{
			name:  "first level only",
			value: data,
			path:  "level1",
			want:  "map[array:[a b c] level2:map[value:found]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonPath(tt.value, tt.path); got != tt.want {
				t.Errorf("jsonPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   string
	}{
		{
			name:   "first non-empty",
			values: []string{"first", "second", "third"},
			want:   "first",
		},
		{
			name:   "skip empty strings",
			values: []string{"", "second", "third"},
			want:   "second",
		},
		{
			name:   "skip whitespace",
			values: []string{"   ", "second", "third"},
			want:   "second",
		},
		{
			name:   "all empty",
			values: []string{"", "", ""},
			want:   "",
		},
		{
			name:   "no values",
			values: []string{},
			want:   "",
		},
		{
			name:   "single value non-empty",
			values: []string{"only"},
			want:   "only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstNonEmpty(tt.values...); got != tt.want {
				t.Errorf("firstNonEmpty() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateBetaGroupTrack(t *testing.T) {
	tests := []struct {
		name    string
		track   string
		wantNil bool
	}{
		{
			name:    "valid internal",
			track:   "internal",
			wantNil: true,
		},
		{
			name:    "valid alpha",
			track:   "alpha",
			wantNil: true,
		},
		{
			name:    "valid beta",
			track:   "beta",
			wantNil: true,
		},
		{
			name:    "invalid track",
			track:   "production",
			wantNil: false,
		},
		{
			name:    "invalid empty",
			track:   "",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBetaGroupTrack(tt.track)
			isNil := err == nil
			if isNil != tt.wantNil {
				t.Errorf("validateBetaGroupTrack() = %v, wantNil %v", err, tt.wantNil)
			}
		})
	}
}
