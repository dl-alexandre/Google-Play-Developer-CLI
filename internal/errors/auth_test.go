package errors

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
)

func TestClassifyAuthErrorInvalidGrant(t *testing.T) {
	t.Parallel()
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
		},
		Body: []byte(`{"error":"invalid_grant","error_description":"expired"}`),
	}

	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
		return
	}
	if apiErr.Code != CodeAuthFailure {
		t.Fatalf("expected auth failure, got %s", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Fatalf("expected hint to be set")
	}
}

func TestClassifyAuthErrorGoogleAPI(t *testing.T) {
	t.Parallel()
	gapiErr := &googleapi.Error{
		Code:    http.StatusUnauthorized,
		Message: "unauthorized",
		Header:  http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
	}

	apiErr := ClassifyAuthError(gapiErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
		return
	}
	if apiErr.Code != CodeAuthFailure {
		t.Fatalf("expected auth failure, got %s", apiErr.Code)
	}
}

func TestClassifyAuthErrorNil(t *testing.T) {
	t.Parallel()
	apiErr := ClassifyAuthError(nil)
	if apiErr != nil {
		t.Fatalf("expected nil for nil error")
	}
}

func TestClassifyAuthErrorAlreadyAPIError(t *testing.T) {
	t.Parallel()
	original := NewAPIError(CodeValidationError, "test error")
	result := ClassifyAuthError(original)
	if result != original {
		t.Fatalf("expected same APIError instance")
	}
}

func TestClassifyAuthErrorOAuthTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		oauthError   string
		expectedCode ErrorCode
		expectHint   bool
		hintContains string
	}{
		{
			name:         "invalid_grant",
			oauthError:   "invalid_grant",
			expectedCode: CodeAuthFailure,
			expectHint:   true,
			hintContains: "Re-authenticate",
		},
		{
			name:         "invalid_client",
			oauthError:   "invalid_client",
			expectedCode: CodeAuthFailure,
			expectHint:   true,
			hintContains: "client credentials",
		},
		{
			name:         "unauthorized_client",
			oauthError:   "unauthorized_client",
			expectedCode: CodeAuthFailure,
			expectHint:   true,
			hintContains: "client credentials",
		},
		{
			name:         "access_denied",
			oauthError:   "access_denied",
			expectedCode: CodeAuthFailure,
			expectHint:   true,
			hintContains: "Re-authenticate",
		},
		{
			name:         "unknown_error",
			oauthError:   "unknown_error",
			expectedCode: CodeAuthFailure,
			expectHint:   true,
			hintContains: "Re-authenticate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retrieveErr := &oauth2.RetrieveError{
				Response: &http.Response{
					StatusCode: http.StatusBadRequest,
					Header:     http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
				},
				Body: []byte(`{"error":"` + tt.oauthError + `","error_description":"test description"}`),
			}

			apiErr := ClassifyAuthError(retrieveErr)
			if apiErr == nil {
				t.Fatal("expected APIError")
				return
			}
			if apiErr.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, apiErr.Code)
			}
			if tt.expectHint && apiErr.Hint == "" {
				t.Errorf("expected hint to be set")
			}
			if tt.hintContains != "" && !contains(apiErr.Hint, tt.hintContains) {
				t.Errorf("expected hint to contain %q, got %q", tt.hintContains, apiErr.Hint)
			}

			// Check details
			if apiErr.Details == nil {
				t.Error("expected details to be set")
			}
			details, ok := apiErr.Details.(map[string]interface{})
			if !ok {
				t.Fatal("expected details to be map")
				return
			}
			if details["oauthError"] != tt.oauthError {
				t.Errorf("expected oauthError %q in details, got %v", tt.oauthError, details["oauthError"])
			}
		})
	}
}

func TestClockSkewDetection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		timeOffset     time.Duration
		expectSkewHint bool
	}{
		{
			name:           "no_skew",
			timeOffset:     0,
			expectSkewHint: false,
		},
		{
			name:           "small_skew_4min",
			timeOffset:     4 * time.Minute,
			expectSkewHint: false,
		},
		{
			name:           "large_skew_6min",
			timeOffset:     6 * time.Minute,
			expectSkewHint: true,
		},
		{
			name:           "very_large_skew_1hour",
			timeOffset:     1 * time.Hour,
			expectSkewHint: true,
		},
		{
			name:           "negative_skew_10min",
			timeOffset:     -10 * time.Minute,
			expectSkewHint: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skewedTime := time.Now().Add(tt.timeOffset)
			retrieveErr := &oauth2.RetrieveError{
				Response: &http.Response{
					StatusCode: http.StatusBadRequest,
					Header:     http.Header{"Date": []string{skewedTime.UTC().Format(http.TimeFormat)}},
				},
				Body: []byte(`{"error":"invalid_grant"}`),
			}

			apiErr := ClassifyAuthError(retrieveErr)
			if apiErr == nil {
				t.Fatal("expected APIError")
				return
			}

			hasClockSkewHint := contains(apiErr.Hint, "clock")
			if hasClockSkewHint != tt.expectSkewHint {
				t.Errorf("expected clock skew hint: %v, got: %v (hint: %q)", tt.expectSkewHint, hasClockSkewHint, apiErr.Hint)
			}

			if tt.expectSkewHint {
				details, ok := apiErr.Details.(map[string]interface{})
				if !ok || details["clockSkewSeconds"] == nil {
					t.Error("expected clockSkewSeconds in details when skew detected")
				}
			}
		})
	}
}

func TestClockSkewNoDateHeader(t *testing.T) {
	t.Parallel()
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{},
		},
		Body: []byte(`{"error":"invalid_grant"}`),
	}

	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
		return
	}

	if contains(apiErr.Hint, "clock") {
		t.Error("should not have clock skew hint without Date header")
	}
}

func TestClockSkewNilHeader(t *testing.T) {
	t.Parallel()
	details := map[string]interface{}{}
	if addClockSkew(details, nil) {
		t.Fatal("expected no skew with nil header")
	}
	if _, ok := details["clockSkewSeconds"]; ok {
		t.Fatal("did not expect clockSkewSeconds")
	}
}

func TestClockSkewInvalidDate(t *testing.T) {
	t.Parallel()
	details := map[string]interface{}{}
	header := http.Header{"Date": []string{"not a date"}}
	if addClockSkew(details, header) {
		t.Fatal("expected no skew with invalid date")
	}
	if _, ok := details["clockSkewSeconds"]; ok {
		t.Fatal("did not expect clockSkewSeconds")
	}
}

func TestGoogleAPIErrorForbidden(t *testing.T) {
	t.Parallel()
	gapiErr := &googleapi.Error{
		Code:    http.StatusForbidden,
		Message: "forbidden",
		Header:  http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
	}

	apiErr := ClassifyAuthError(gapiErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
		return
	}
	if apiErr.Code != CodePermissionDenied {
		t.Errorf("expected permission denied, got %s", apiErr.Code)
	}
	if !contains(apiErr.Hint, "permissions") {
		t.Errorf("expected hint about permissions, got %q", apiErr.Hint)
	}
}

func TestGoogleAPIErrorWithClockSkew(t *testing.T) {
	t.Parallel()
	skewedTime := time.Now().Add(-10 * time.Minute)
	gapiErr := &googleapi.Error{
		Code:    http.StatusUnauthorized,
		Message: "unauthorized",
		Header:  http.Header{"Date": []string{skewedTime.UTC().Format(http.TimeFormat)}},
	}

	apiErr := ClassifyAuthError(gapiErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
		return
	}
	if !contains(apiErr.Hint, "clock") {
		t.Error("expected clock skew hint for Google API error with skew")
	}
}

func TestAppendHint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		current  string
		extra    string
		expected string
	}{
		{
			name:     "empty_current",
			current:  "",
			extra:    "new hint",
			expected: "new hint",
		},
		{
			name:     "append_to_existing",
			current:  "existing hint",
			extra:    "additional hint",
			expected: "existing hint; additional hint",
		},
		{
			name:     "duplicate_hint",
			current:  "check credentials",
			extra:    "check credentials",
			expected: "check credentials",
		},
		{
			name:     "contains_check",
			current:  "Re-authenticate and check credentials",
			extra:    "check credentials",
			expected: "Re-authenticate and check credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendHint(tt.current, tt.extra)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestOAuthErrorResponseParsing(t *testing.T) {
	t.Parallel()
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{},
		},
		Body: []byte(`{"error":"invalid_grant","error_description":"Token has been expired or revoked.","error_uri":"https://example.com/error"}`),
	}

	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
	}

	details, ok := apiErr.Details.(map[string]interface{})
	if !ok {
		t.Fatal("expected details to be map")
	}

	if details["oauthError"] != "invalid_grant" {
		t.Errorf("expected oauthError in details")
	}
	if details["oauthErrorDescription"] != "Token has been expired or revoked." {
		t.Errorf("expected oauthErrorDescription in details")
	}
}

func TestOAuthErrorInvalidJSON(t *testing.T) {
	t.Parallel()
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{},
		},
		Body: []byte(`not valid json`),
	}

	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatal("expected APIError even with invalid JSON")
	}
	if apiErr.Code != CodeAuthFailure {
		t.Errorf("expected auth failure, got %s", apiErr.Code)
	}
}

func TestOAuthErrorBodyNil(t *testing.T) {
	t.Parallel()
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{},
		},
		Body: nil,
	}
	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
	}
	details, ok := apiErr.Details.(map[string]interface{})
	if !ok {
		t.Fatal("expected details to be map")
	}
	if _, ok := details["oauthError"]; ok {
		t.Fatal("did not expect oauthError")
	}
}

func TestOAuthErrorSkewedInvalidClient(t *testing.T) {
	t.Parallel()
	skewedTime := time.Now().Add(-10 * time.Minute)
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Date": []string{skewedTime.UTC().Format(http.TimeFormat)}},
		},
		Body: []byte(`{"error":"invalid_client"}`),
	}
	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
	}
	if !contains(apiErr.Hint, "clock") {
		t.Fatal("expected clock skew hint")
	}
}

func TestOAuthErrorSkewedAccessDenied(t *testing.T) {
	t.Parallel()
	skewedTime := time.Now().Add(-10 * time.Minute)
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Date": []string{skewedTime.UTC().Format(http.TimeFormat)}},
		},
		Body: []byte(`{"error":"access_denied"}`),
	}
	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
	}
	if !contains(apiErr.Hint, "clock") {
		t.Fatal("expected clock skew hint")
	}
}

func TestOAuthErrorSkewedDefault(t *testing.T) {
	t.Parallel()
	skewedTime := time.Now().Add(-10 * time.Minute)
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Date": []string{skewedTime.UTC().Format(http.TimeFormat)}},
		},
		Body: []byte(`{"error":"unknown_error"}`),
	}
	apiErr := ClassifyAuthError(retrieveErr)
	if apiErr == nil {
		t.Fatal("expected APIError")
	}
	if !contains(apiErr.Hint, "clock") {
		t.Fatal("expected clock skew hint")
	}
}

func TestGenericError(t *testing.T) {
	t.Parallel()
	err := &testError{msg: "generic error"}
	apiErr := ClassifyAuthError(err)
	if apiErr == nil {
		t.Fatal("expected APIError")
	}
	if apiErr.Code != CodeAuthFailure {
		t.Errorf("expected auth failure for generic error, got %s", apiErr.Code)
	}
	if apiErr.Message != "generic error" {
		t.Errorf("expected message to be preserved, got %q", apiErr.Message)
	}
}

// Helper functions

func contains(s, substr string) bool {
	return s != "" && substr != "" && (s == substr || len(s) >= len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestClockSkewBoundaryConditions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		timeOffset     time.Duration
		expectSkewHint bool
	}{
		{
			name:           "exactly_5_minutes",
			timeOffset:     5 * time.Minute,
			expectSkewHint: false,
		},
		{
			name:           "5_minutes_plus_1_second",
			timeOffset:     5*time.Minute + 1*time.Second,
			expectSkewHint: true,
		},
		{
			name:           "4_minutes_59_seconds",
			timeOffset:     4*time.Minute + 59*time.Second,
			expectSkewHint: false,
		},
		{
			name:           "negative_4_minutes_59_seconds",
			timeOffset:     -(4*time.Minute + 59*time.Second),
			expectSkewHint: false,
		},
		{
			name:           "negative_5_minutes_plus_1_second",
			timeOffset:     -(5*time.Minute + 1*time.Second),
			expectSkewHint: true,
		},
		{
			name:           "large_future_skew",
			timeOffset:     -10 * time.Minute,
			expectSkewHint: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skewedTime := time.Now().Add(tt.timeOffset)
			retrieveErr := &oauth2.RetrieveError{
				Response: &http.Response{
					StatusCode: http.StatusBadRequest,
					Header:     http.Header{"Date": []string{skewedTime.UTC().Format(http.TimeFormat)}},
				},
				Body: []byte(`{"error":"invalid_grant"}`),
			}

			apiErr := ClassifyAuthError(retrieveErr)
			if apiErr == nil {
				t.Fatal("expected APIError")
			}

			hasClockSkewHint := contains(apiErr.Hint, "clock")
			if hasClockSkewHint != tt.expectSkewHint {
				t.Errorf("expected clock skew hint: %v, got: %v (hint: %q)", tt.expectSkewHint, hasClockSkewHint, apiErr.Hint)
			}
		})
	}
}

func TestHTTPStatusVariations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		statusCode   int
		expectedCode ErrorCode
	}{
		{
			name:         "rate_limited_429",
			statusCode:   http.StatusTooManyRequests,
			expectedCode: CodeRateLimited,
		},
		{
			name:         "internal_server_error_500",
			statusCode:   http.StatusInternalServerError,
			expectedCode: CodeGeneralError,
		},
		{
			name:         "bad_gateway_502",
			statusCode:   http.StatusBadGateway,
			expectedCode: CodeGeneralError,
		},
		{
			name:         "service_unavailable_503",
			statusCode:   http.StatusServiceUnavailable,
			expectedCode: CodeGeneralError,
		},
		{
			name:         "gateway_timeout_504",
			statusCode:   http.StatusGatewayTimeout,
			expectedCode: CodeGeneralError,
		},
		{
			name:         "not_found_404",
			statusCode:   http.StatusNotFound,
			expectedCode: CodeNotFound,
		},
		{
			name:         "conflict_409",
			statusCode:   http.StatusConflict,
			expectedCode: CodeConflict,
		},
		{
			name:         "bad_request_400",
			statusCode:   http.StatusBadRequest,
			expectedCode: CodeValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gapiErr := &googleapi.Error{
				Code:    tt.statusCode,
				Message: "test error",
				Header:  http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
			}

			apiErr := ClassifyAuthError(gapiErr)
			if apiErr == nil {
				t.Fatal("expected APIError")
			}
			if apiErr.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, apiErr.Code)
			}
		})
	}
}

func TestHTTPStatusWithHints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		statusCode   int
		expectHint   bool
		hintContains string
	}{
		{
			name:         "unauthorized_401",
			statusCode:   http.StatusUnauthorized,
			expectHint:   true,
			hintContains: "Re-authenticate",
		},
		{
			name:         "forbidden_403",
			statusCode:   http.StatusForbidden,
			expectHint:   true,
			hintContains: "permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gapiErr := &googleapi.Error{
				Code:    tt.statusCode,
				Message: "test error",
				Header:  http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
			}

			apiErr := ClassifyAuthError(gapiErr)
			if apiErr == nil {
				t.Fatal("expected APIError")
			}
			if tt.expectHint && apiErr.Hint == "" {
				t.Error("expected hint to be set")
			}
			if tt.hintContains != "" && !contains(apiErr.Hint, tt.hintContains) {
				t.Errorf("expected hint to contain %q, got %q", tt.hintContains, apiErr.Hint)
			}
		})
	}
}
