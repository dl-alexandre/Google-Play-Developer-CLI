package errors

import (
	"net/http"
	"testing"
)

func TestExitCodeMapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		code     ErrorCode
		expected int
	}{
		{"success", CodeSuccess, ExitSuccess},
		{"auth failure", CodeAuthFailure, ExitAuthFailure},
		{"permission denied", CodePermissionDenied, ExitPermissionDenied},
		{"validation error", CodeValidationError, ExitValidationError},
		{"rate limited", CodeRateLimited, ExitRateLimited},
		{"network error", CodeNetworkError, ExitNetworkError},
		{"not found", CodeNotFound, ExitNotFound},
		{"conflict", CodeConflict, ExitConflict},
		{"general error", CodeGeneralError, ExitGeneralError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.code, "test message")
			if got := err.ExitCode(); got != tt.expected {
				t.Errorf("ExitCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIErrorMessage(t *testing.T) {
	t.Parallel()
	err := NewAPIError(CodeAuthFailure, "authentication failed")
	if err.Error() != "AUTH_FAILURE: authentication failed" {
		t.Errorf("Error() = %v, want AUTH_FAILURE: authentication failed", err.Error())
	}

	err = err.WithHint("check your credentials")
	expected := "AUTH_FAILURE: authentication failed (hint: check your credentials)"
	if err.Error() != expected {
		t.Errorf("Error() with hint = %v, want %v", err.Error(), expected)
	}
}

func TestFromHTTPStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status   int
		expected ErrorCode
	}{
		{http.StatusUnauthorized, CodeAuthFailure},
		{http.StatusForbidden, CodePermissionDenied},
		{http.StatusNotFound, CodeNotFound},
		{http.StatusConflict, CodeConflict},
		{http.StatusTooManyRequests, CodeRateLimited},
		{http.StatusBadRequest, CodeValidationError},
		{http.StatusInternalServerError, CodeGeneralError},
		{http.StatusOK, CodeGeneralError},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			if got := FromHTTPStatus(tt.status); got != tt.expected {
				t.Errorf("FromHTTPStatus(%d) = %v, want %v", tt.status, got, tt.expected)
			}
		})
	}
}

func TestAPIErrorChaining(t *testing.T) {
	t.Parallel()
	err := NewAPIError(CodeValidationError, "invalid input").
		WithHint("check the format").
		WithDetails(map[string]string{"field": "package"}).
		WithHTTPStatus(http.StatusBadRequest).
		WithService("androidpublisher").
		WithOperation("edits.insert").
		WithRetryAfter(30)

	if err.Code != CodeValidationError {
		t.Errorf("Code = %v, want %v", err.Code, CodeValidationError)
	}
	if err.Message != "invalid input" {
		t.Errorf("Message = %v, want 'invalid input'", err.Message)
	}
	if err.Hint != "check the format" {
		t.Errorf("Hint = %v, want 'check the format'", err.Hint)
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %v, want %v", err.HTTPStatus, http.StatusBadRequest)
	}
	if err.Service != "androidpublisher" {
		t.Errorf("Service = %v, want 'androidpublisher'", err.Service)
	}
	if err.Operation != "edits.insert" {
		t.Errorf("Operation = %v, want 'edits.insert'", err.Operation)
	}
	if err.RetryAfterSeconds != 30 {
		t.Errorf("RetryAfterSeconds = %v, want 30", err.RetryAfterSeconds)
	}
}
