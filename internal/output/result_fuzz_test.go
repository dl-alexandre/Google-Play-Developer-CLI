//go:build unit
// +build unit

package output

import (
	"testing"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// FuzzNewResult fuzz tests result creation with various data types
func FuzzNewResult(f *testing.F) {
	f.Add(`{"key": "value"}`)
	f.Add(`{}`)
	f.Add(`[]`)
	f.Add(`null`)
	f.Add(`{"unicode": "日本語"}`)

	f.Fuzz(func(t *testing.T, data string) {
		// Should not panic
		result := NewResult(data)
		_ = result.ExitCode
	})
}

// FuzzNewErrorResult fuzz tests error result creation
func FuzzNewErrorResult(f *testing.F) {
	f.Add("Test error message", int(1))   // CodeGeneralError index
	f.Add("", int(0))                     // CodeSuccess index
	f.Add("Unicode error: エラー 🚨", int(4)) // CodeValidationError index

	f.Fuzz(func(t *testing.T, message string, codeIdx int) {
		// Map int to ErrorCode
		codes := []errors.ErrorCode{
			errors.CodeSuccess, errors.CodeGeneralError, errors.CodeAuthFailure,
			errors.CodePermissionDenied, errors.CodeValidationError, errors.CodeRateLimited,
			errors.CodeNetworkError, errors.CodeNotFound, errors.CodeConflict,
		}
		code := codes[codeIdx%len(codes)]

		err := errors.NewAPIError(code, message)
		// Should not panic
		result := NewErrorResult(err)
		_ = result.ExitCode
	})
}

// FuzzParseFormat fuzz tests format parsing
func FuzzParseFormat(f *testing.F) {
	f.Add("json")
	f.Add("table")
	f.Add("csv")
	f.Add("markdown")
	f.Add("")
	f.Add("JSON")
	f.Add("invalid")
	f.Add("xml")

	f.Fuzz(func(t *testing.T, format string) {
		// Should not panic
		_ = ParseFormat(format)
	})
}
