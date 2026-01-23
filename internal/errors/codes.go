// Package errors provides standardized error types and exit codes for gpd.
package errors

import (
	"fmt"
	"net/http"
)

// Exit codes for gpd CLI following the requirements specification.
const (
	ExitSuccess          = 0 // Command succeeded
	ExitGeneralError     = 1 // Other API errors
	ExitAuthFailure      = 2 // Authentication failures
	ExitPermissionDenied = 3 // Permission denied
	ExitValidationError  = 4 // Input validation errors
	ExitRateLimited      = 5 // Rate limit exceeded (HTTP 429 or quota)
	ExitNetworkError     = 6 // Network errors (DNS, TLS, timeouts)
	ExitNotFound         = 7 // Resource not found
	ExitConflict         = 8 // Conflicts (edit exists, file lock contention)
)

// ErrorCode represents a string error code for structured error responses.
type ErrorCode string

const (
	CodeSuccess          ErrorCode = "SUCCESS"
	CodeGeneralError     ErrorCode = "GENERAL_ERROR"
	CodeAuthFailure      ErrorCode = "AUTH_FAILURE"
	CodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	CodeValidationError  ErrorCode = "VALIDATION_ERROR"
	CodeRateLimited      ErrorCode = "RATE_LIMITED"
	CodeNetworkError     ErrorCode = "NETWORK_ERROR"
	CodeNotFound         ErrorCode = "NOT_FOUND"
	CodeConflict         ErrorCode = "CONFLICT"
)

// APIError represents a structured error response.
type APIError struct {
	Code              ErrorCode   `json:"code"`
	Message           string      `json:"message"`
	Hint              string      `json:"hint,omitempty"`
	Details           interface{} `json:"details,omitempty"`
	HTTPStatus        int         `json:"httpStatus,omitempty"`
	RetryAfterSeconds int         `json:"retryAfterSeconds,omitempty"`
	Service           string      `json:"service,omitempty"`
	Operation         string      `json:"operation,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.Hint != "" {
		return fmt.Sprintf("%s: %s (hint: %s)", e.Code, e.Message, e.Hint)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ExitCode returns the appropriate exit code for this error.
func (e *APIError) ExitCode() int {
	switch e.Code {
	case CodeSuccess:
		return ExitSuccess
	case CodeAuthFailure:
		return ExitAuthFailure
	case CodePermissionDenied:
		return ExitPermissionDenied
	case CodeValidationError:
		return ExitValidationError
	case CodeRateLimited:
		return ExitRateLimited
	case CodeNetworkError:
		return ExitNetworkError
	case CodeNotFound:
		return ExitNotFound
	case CodeConflict:
		return ExitConflict
	default:
		return ExitGeneralError
	}
}

// NewAPIError creates a new APIError with the given parameters.
func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// WithHint adds a hint to the error.
func (e *APIError) WithHint(hint string) *APIError {
	e.Hint = hint
	return e
}

// WithDetails adds details to the error.
func (e *APIError) WithDetails(details interface{}) *APIError {
	e.Details = details
	return e
}

// WithHTTPStatus adds HTTP status to the error.
func (e *APIError) WithHTTPStatus(status int) *APIError {
	e.HTTPStatus = status
	return e
}

// WithRetryAfter adds retry-after seconds to the error.
func (e *APIError) WithRetryAfter(seconds int) *APIError {
	e.RetryAfterSeconds = seconds
	return e
}

// WithService adds service name to the error.
func (e *APIError) WithService(service string) *APIError {
	e.Service = service
	return e
}

// WithOperation adds operation name to the error.
func (e *APIError) WithOperation(operation string) *APIError {
	e.Operation = operation
	return e
}

// FromHTTPStatus creates an appropriate error code from HTTP status.
func FromHTTPStatus(status int) ErrorCode {
	switch {
	case status == http.StatusUnauthorized:
		return CodeAuthFailure
	case status == http.StatusForbidden:
		return CodePermissionDenied
	case status == http.StatusNotFound:
		return CodeNotFound
	case status == http.StatusConflict:
		return CodeConflict
	case status == http.StatusTooManyRequests:
		return CodeRateLimited
	case status >= 400 && status < 500:
		return CodeValidationError
	case status >= 500:
		return CodeGeneralError
	default:
		return CodeGeneralError
	}
}

// Common errors with hints.
var (
	ErrAuthNotConfigured = NewAPIError(CodeAuthFailure, "authentication not configured").
				WithHint("Run 'gpd auth status' to check authentication or set GPD_SERVICE_ACCOUNT_KEY environment variable")

	ErrServiceAccountInvalid = NewAPIError(CodeAuthFailure, "invalid service account key").
					WithHint("Ensure the service account key file is valid JSON and contains required fields")

	ErrPermissionDenied = NewAPIError(CodePermissionDenied, "permission denied").
				WithHint("Ensure the service account has required permissions in Google Play Console")

	ErrPackageRequired = NewAPIError(CodeValidationError, "package name is required").
				WithHint("Provide --package flag or set default package in config")

	ErrTrackInvalid = NewAPIError(CodeValidationError, "invalid track name").
			WithHint("Valid tracks are: internal, alpha, beta, production")

	ErrEditConflict = NewAPIError(CodeConflict, "edit transaction conflict").
			WithHint("Another process may be using this edit. Wait and retry, or use a different --edit-id")

	ErrFileLockTimeout = NewAPIError(CodeConflict, "file lock acquisition timeout").
				WithHint("Another gpd process may be running. Wait for it to complete or check for stale locks")
)
