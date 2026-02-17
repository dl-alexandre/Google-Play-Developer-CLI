// Package migrate provides shared validation helpers for metadata migrations.
package migrate

import "unicode/utf8"

var limits = map[string]int{
	"title":            30,
	"shortDescription": 80,
	"fullDescription":  4000,
	"releaseNotes":     500,
}

// ValidationError represents a single metadata validation issue.
type ValidationError struct {
	Locale  string `json:"locale"`
	Field   string `json:"field"`
	Message string `json:"message"`
	Current int    `json:"current"`
	Limit   int    `json:"limit"`
}

// ValidateText returns a ValidationError when text exceeds Google Play limits.
func ValidateText(field, text string) *ValidationError {
	limit, ok := limits[field]
	if !ok || text == "" {
		return nil
	}
	current := utf8.RuneCountInString(text)
	if current <= limit {
		return nil
	}
	return &ValidationError{
		Field:   field,
		Message: "exceeds limit",
		Current: current,
		Limit:   limit,
	}
}
