//go:build unit
// +build unit

package errors

import (
	"testing"
)

// FuzzAPIError fuzz tests API error creation
func FuzzAPIError(f *testing.F) {
	f.Add("Test error message", int(1))   // CodeGeneralError index
	f.Add("", int(0))                     // CodeSuccess index
	f.Add("Unicode error: エラー 🚨", int(4)) // CodeValidationError index

	f.Fuzz(func(t *testing.T, message string, codeIdx int) {
		// Map int to ErrorCode
		codes := []ErrorCode{
			CodeSuccess, CodeGeneralError, CodeAuthFailure,
			CodePermissionDenied, CodeValidationError, CodeRateLimited,
			CodeNetworkError, CodeNotFound, CodeConflict,
		}
		code := codes[codeIdx%len(codes)]

		// Should not panic
		err := NewAPIError(code, message)
		_ = err.Error()
		_ = err.ExitCode()
	})
}

// FuzzErrorWithMethods fuzz tests error method chaining
func FuzzErrorWithMethods(f *testing.F) {
	f.Add("base message", "hint text", "detail info")
	f.Add("", "", "")
	f.Add("complex message", "hint with unicode 🎉", "details with unicode 日本語")

	f.Fuzz(func(t *testing.T, message, hint, details string) {
		baseErr := NewAPIError(CodeGeneralError, message)

		// Should not panic
		_ = baseErr.WithHint(hint)
		_ = baseErr.WithDetails(details)
	})
}
