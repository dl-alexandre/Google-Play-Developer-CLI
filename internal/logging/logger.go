// Package logging provides structured logging with PII redaction for gpd.
package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Level represents a log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Field represents a log field.
type Field struct {
	Key   string
	Value interface{}
	PII   bool // Mark as PII for redaction
}

// Logger provides structured logging with PII redaction.
type Logger struct {
	mu       sync.Mutex
	writer   io.Writer
	level    Level
	verbose  bool
	redactor *PIIRedactor
}

// NewLogger creates a new logger.
func NewLogger(w io.Writer, verbose bool) *Logger {
	return &Logger{
		writer:   w,
		level:    LevelInfo,
		verbose:  verbose,
		redactor: NewPIIRedactor(),
	}
}

// SetLevel sets the minimum log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(LevelDebug, msg, fields...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...Field) {
	l.log(LevelInfo, msg, fields...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(LevelWarn, msg, fields...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...Field) {
	l.log(LevelError, msg, fields...)
}

func (l *Logger) log(level Level, msg string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	// Apply PII redaction to fields
	redactedFields := make(map[string]interface{})
	for _, f := range fields {
		value := f.Value
		if f.PII || l.redactor.IsSensitiveField(f.Key) {
			value = l.redactor.Redact(f.Key, f.Value)
		}
		redactedFields[f.Key] = value
	}

	if l.verbose {
		// JSON format for verbose mode
		entry := map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"level":     level.String(),
			"message":   msg,
			"fields":    redactedFields,
		}
		data, _ := json.Marshal(entry)
		fmt.Fprintln(l.writer, string(data))
	} else {
		// Simple format for normal mode
		var parts []string
		for k, v := range redactedFields {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		fieldStr := ""
		if len(parts) > 0 {
			fieldStr = " " + strings.Join(parts, " ")
		}
		fmt.Fprintf(l.writer, "[%s] %s%s\n", level.String(), msg, fieldStr)
	}
}

// PIIRedactor handles PII redaction in log output.
type PIIRedactor struct {
	// Allowlisted fields (safe to log)
	allowedFields map[string]bool
	// Sensitive field patterns
	sensitiveFields map[string]bool
	// Regex patterns for detecting PII
	patterns []*regexp.Regexp
}

// NewPIIRedactor creates a new PII redactor.
func NewPIIRedactor() *PIIRedactor {
	r := &PIIRedactor{
		allowedFields: map[string]bool{
			"command":     true,
			"duration":    true,
			"durationMs":  true,
			"exit_code":   true,
			"exitCode":    true,
			"package":     true,
			"packageName": true,
			"track":       true,
			"versionCode": true,
			"status":      true,
			"action":      true,
			"editId":      true,
			"locale":      true,
			"format":      true,
			"pageSize":    true,
			"pageToken":   true,
			"startDate":   true,
			"endDate":     true,
			"metrics":     true,
			"dimensions":  true,
			"productId":   true,
			"productType": true,
			"environment": true,
			"success":     true,
			"error":       true,
			"hint":        true,
			"retries":     true,
			"noop":        true,
			"dryRun":      true,
		},
		sensitiveFields: map[string]bool{
			"email":             true,
			"userName":          true,
			"authorName":        true,
			"reviewText":        true,
			"text":              true,
			"token":             true,
			"purchaseToken":     true,
			"accessToken":       true,
			"refreshToken":      true,
			"serviceAccountKey": true,
			"privateKey":        true,
			"orderId":           true,
			"orderIds":          true,
			"groups":            true,
		},
	}

	// Compile regex patterns for PII detection
	r.patterns = []*regexp.Regexp{
		regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),       // Email
		regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`),               // IP address
		regexp.MustCompile(`\b\d{10,}\b`),                                          // Phone numbers (simplified)
		regexp.MustCompile(`ya29\.[a-zA-Z0-9_-]+`),                                 // Google access tokens
		regexp.MustCompile(`eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`), // JWT tokens
	}

	return r
}

// IsSensitiveField checks if a field is sensitive.
func (r *PIIRedactor) IsSensitiveField(key string) bool {
	if r.sensitiveFields[key] {
		return true
	}
	// Check for common sensitive key patterns
	lowerKey := strings.ToLower(key)
	return strings.Contains(lowerKey, "password") ||
		strings.Contains(lowerKey, "secret") ||
		strings.Contains(lowerKey, "key") ||
		strings.Contains(lowerKey, "token") ||
		strings.Contains(lowerKey, "credential")
}

// IsAllowedField checks if a field is safe to log.
func (r *PIIRedactor) IsAllowedField(key string) bool {
	return r.allowedFields[key]
}

// Redact redacts sensitive data from a value.
func (r *PIIRedactor) Redact(key string, value interface{}) interface{} {
	// Handle nil values
	if value == nil {
		return nil
	}

	// Handle string values
	if s, ok := value.(string); ok {
		return r.redactString(key, s)
	}

	// Handle slices of strings (like email lists)
	if slice, ok := value.([]string); ok {
		result := make([]string, len(slice))
		for i, s := range slice {
			result[i] = r.redactString(key, s)
		}
		return result
	}

	// Handle maps
	if m, ok := value.(map[string]interface{}); ok {
		result := make(map[string]interface{})
		for k, v := range m {
			if r.IsSensitiveField(k) {
				result[k] = r.Redact(k, v)
			} else {
				result[k] = v
			}
		}
		return result
	}

	// For other types, return redacted placeholder
	return "[REDACTED]"
}

func (r *PIIRedactor) redactString(key, value string) string {
	if value == "" {
		return ""
	}

	// Check for known sensitive field names
	if r.IsSensitiveField(key) {
		if len(value) <= 4 {
			return "[REDACTED]"
		}
		// Show first 2 and last 2 characters for debugging
		return value[:2] + "..." + value[len(value)-2:] + "[REDACTED]"
	}

	// Apply pattern-based redaction
	result := value
	for _, pattern := range r.patterns {
		result = pattern.ReplaceAllString(result, "[REDACTED]")
	}

	return result
}

// RedactMap redacts all sensitive fields in a map.
func (r *PIIRedactor) RedactMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		if r.IsSensitiveField(k) {
			result[k] = r.Redact(k, v)
		} else if r.IsAllowedField(k) {
			result[k] = v
		} else {
			// For unknown fields, check if value contains PII patterns
			if s, ok := v.(string); ok {
				result[k] = r.redactString(k, s)
			} else {
				result[k] = v
			}
		}
	}
	return result
}

// Global logger instance
var (
	defaultLogger     *Logger
	defaultLoggerOnce sync.Once
)

func getDefaultLogger() *Logger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewLogger(os.Stderr, false)
	})
	return defaultLogger
}

// SetDefault sets the default logger.
func SetDefault(l *Logger) {
	defaultLogger = l
}

// Debug logs a debug message using the default logger.
func Debug(msg string, fields ...Field) {
	getDefaultLogger().Debug(msg, fields...)
}

// Info logs an info message using the default logger.
func Info(msg string, fields ...Field) {
	getDefaultLogger().Info(msg, fields...)
}

// Warn logs a warning message using the default logger.
func Warn(msg string, fields ...Field) {
	getDefaultLogger().Warn(msg, fields...)
}

// Error logs an error message using the default logger.
func Error(msg string, fields ...Field) {
	getDefaultLogger().Error(msg, fields...)
}

// String creates a string field.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a boolean field.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field.
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value.String()}
}

// Err creates an error field.
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}

// Sensitive creates a field marked as PII.
func Sensitive(key string, value interface{}) Field {
	return Field{Key: key, Value: value, PII: true}
}
