package logging

import (
	"bytes"
	"encoding/json"
	stdErrors "errors"
	"strings"
	"testing"
	"time"
)

func TestPIIRedactorAllowlistedFields(t *testing.T) {
	redactor := NewPIIRedactor()

	allowedFields := []string{
		"command", "duration", "durationMs", "exit_code", "exitCode",
		"package", "packageName", "track", "versionCode", "status",
	}

	for _, field := range allowedFields {
		t.Run(field, func(t *testing.T) {
			if !redactor.IsAllowedField(field) {
				t.Errorf("IsAllowedField(%q) = false, want true", field)
			}
		})
	}
}

func TestPIIRedactorSensitiveFields(t *testing.T) {
	redactor := NewPIIRedactor()

	sensitiveFields := []string{
		"email", "userName", "authorName", "reviewText", "text",
		"token", "purchaseToken", "accessToken", "refreshToken",
		"serviceAccountKey", "privateKey", "orderId",
	}

	for _, field := range sensitiveFields {
		t.Run(field, func(t *testing.T) {
			if !redactor.IsSensitiveField(field) {
				t.Errorf("IsSensitiveField(%q) = false, want true", field)
			}
		})
	}
}

func TestPIIRedactorRedactString(t *testing.T) {
	redactor := NewPIIRedactor()

	tests := []struct {
		name     string
		key      string
		value    string
		contains string // substring that should be in result
	}{
		{"email field redacted", "email", "user@example.com", "REDACTED"},
		{"token field redacted", "token", "ya29.abc123xyz", "REDACTED"},
		{"email pattern redacted", "message", "Contact user@example.com for help", "REDACTED"},
		{"ip pattern redacted", "log", "Request from 192.168.1.1 failed", "REDACTED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactor.Redact(tt.key, tt.value)
			if s, ok := result.(string); ok {
				if !strings.Contains(s, tt.contains) {
					t.Errorf("Redact(%q, %q) = %q, want to contain %q", tt.key, tt.value, s, tt.contains)
				}
			}
		})
	}
}

func TestPIIRedactorPreservesAllowedFields(t *testing.T) {
	redactor := NewPIIRedactor()

	// Non-sensitive fields should not be redacted
	value := "com.example.app"
	result := redactor.Redact("package", value)
	if result != value {
		t.Errorf("Redact('package', %q) = %q, want %q", value, result, value)
	}
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, false)

	logger.Info("test message", String("package", "com.example.app"))

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Output should contain 'INFO'")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Output should contain 'test message'")
	}
	if !strings.Contains(output, "com.example.app") {
		t.Error("Output should contain 'com.example.app'")
	}
}

func TestLoggerVerboseJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, true)

	logger.Info("test message", String("package", "com.example.app"))

	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Error("Verbose output should contain JSON level field")
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Error("Verbose output should contain JSON message field")
	}
}

func TestLoggerRedactsSensitiveFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, false)

	logger.Info("user login", String("email", "user@example.com"))

	output := buf.String()
	if strings.Contains(output, "user@example.com") {
		t.Error("Output should not contain raw email address")
	}
	if !strings.Contains(output, "REDACTED") {
		t.Error("Output should contain REDACTED")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, false)
	logger.SetLevel(LevelWarn)

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Output should not contain debug message")
	}
	if strings.Contains(output, "info message") {
		t.Error("Output should not contain info message")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Output should contain warn message")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Output should contain error message")
	}
}

func TestFieldCreators(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		f := String("key", "value")
		if f.Key != "key" || f.Value != "value" {
			t.Error("String field not created correctly")
		}
	})

	t.Run("Int", func(t *testing.T) {
		f := Int("count", 42)
		if f.Key != "count" || f.Value != 42 {
			t.Error("Int field not created correctly")
		}
	})

	t.Run("Int64", func(t *testing.T) {
		f := Int64("count", int64(99))
		if f.Key != "count" || f.Value != int64(99) {
			t.Error("Int64 field not created correctly")
		}
	})

	t.Run("Bool", func(t *testing.T) {
		f := Bool("enabled", true)
		if f.Key != "enabled" || f.Value != true {
			t.Error("Bool field not created correctly")
		}
	})

	t.Run("Duration", func(t *testing.T) {
		f := Duration("elapsed", 2*time.Second)
		if f.Key != "elapsed" || f.Value != "2s" {
			t.Error("Duration field not created correctly")
		}
	})

	t.Run("ErrNil", func(t *testing.T) {
		f := Err(nil)
		if f.Key != "error" || f.Value != nil {
			t.Error("Err field not created correctly for nil")
		}
	})

	t.Run("ErrNonNil", func(t *testing.T) {
		f := Err(stdErrors.New("boom"))
		if f.Key != "error" || f.Value != "boom" {
			t.Error("Err field not created correctly for error")
		}
	})

	t.Run("Sensitive", func(t *testing.T) {
		f := Sensitive("token", "secret123")
		if f.Key != "token" || !f.PII {
			t.Error("Sensitive field not marked as PII")
		}
	})
}

func TestLevelString(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("String() = %q, want %q", got, tt.expected)
		}
	}
}

type errWriter struct{}

func (errWriter) Write(p []byte) (int, error) {
	return 0, stdErrors.New("write error")
}

func TestLoggerVerboseJSONWriteError(t *testing.T) {
	logger := NewLogger(errWriter{}, true)
	logger.Info("test", String("package", "com.example.app"))
}

func TestLoggerVerboseJSONMarshalError(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, true)
	logger.Info("test", Field{Key: "bad", Value: func() {}})
	if buf.Len() != 0 {
		t.Fatal("Expected no output on marshal error")
	}
}

func TestLoggerPlainWriteError(t *testing.T) {
	logger := NewLogger(errWriter{}, false)
	logger.Info("test", String("package", "com.example.app"))
}

func TestPIIRedactorRedactNil(t *testing.T) {
	redactor := NewPIIRedactor()
	if redactor.Redact("email", nil) != nil {
		t.Fatal("Expected nil")
	}
}

func TestPIIRedactorRedactSlice(t *testing.T) {
	redactor := NewPIIRedactor()
	result := redactor.Redact("email", []string{"user@example.com"})
	list, ok := result.([]string)
	if !ok || len(list) != 1 {
		t.Fatal("Expected string slice")
	}
	if !strings.Contains(list[0], "REDACTED") {
		t.Fatal("Expected redacted value")
	}
}

func TestPIIRedactorRedactMap(t *testing.T) {
	redactor := NewPIIRedactor()
	result := redactor.Redact("payload", map[string]interface{}{
		"email":   "user@example.com",
		"package": "com.example.app",
	})
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected map")
	}
	if strings.Contains(m["email"].(string), "user@example.com") {
		t.Fatal("Expected email to be redacted")
	}
	if m["package"] != "com.example.app" {
		t.Fatal("Expected package to be preserved")
	}
}

func TestPIIRedactorRedactOtherType(t *testing.T) {
	redactor := NewPIIRedactor()
	if redactor.Redact("token", 123) != "[REDACTED]" {
		t.Fatal("Expected placeholder")
	}
}

func TestPIIRedactorRedactStringEmpty(t *testing.T) {
	redactor := NewPIIRedactor()
	if redactor.Redact("email", "") != "" {
		t.Fatal("Expected empty string")
	}
}

func TestPIIRedactorRedactStringShortSensitive(t *testing.T) {
	redactor := NewPIIRedactor()
	if redactor.Redact("token", "abc") != "[REDACTED]" {
		t.Fatal("Expected redacted short token")
	}
}

func TestPIIRedactorRedactStringLongSensitive(t *testing.T) {
	redactor := NewPIIRedactor()
	result := redactor.Redact("token", "abcdefgh")
	s, ok := result.(string)
	if !ok || !strings.Contains(s, "[REDACTED]") {
		t.Fatal("Expected redacted long token")
	}
}

func TestPIIRedactorRedactMapFunction(t *testing.T) {
	redactor := NewPIIRedactor()
	result := redactor.RedactMap(map[string]interface{}{
		"email":      "user@example.com",
		"package":    "com.example.app",
		"custom":     "user@example.com",
		"customNon":  123,
		"pageToken":  "safe",
		"secretKey":  "abc",
		"credential": "xyz",
	})
	if strings.Contains(result["email"].(string), "user@example.com") {
		t.Fatal("Expected email redacted")
	}
	if result["package"] != "com.example.app" {
		t.Fatal("Expected allowed field preserved")
	}
	if !strings.Contains(result["custom"].(string), "REDACTED") {
		t.Fatal("Expected custom string redacted")
	}
	if result["customNon"].(int) != 123 {
		t.Fatal("Expected custom non-string preserved")
	}
}

func TestDefaultLoggerHelpers(t *testing.T) {
	var buf bytes.Buffer
	_ = getDefaultLogger()
	logger := NewLogger(&buf, false)
	logger.SetLevel(LevelDebug)
	SetDefault(logger)
	Debug("debug", String("package", "com.example.app"))
	Info("info", String("package", "com.example.app"))
	Warn("warn", String("package", "com.example.app"))
	Error("error", String("package", "com.example.app"))
	output := buf.String()
	if !strings.Contains(output, "debug") {
		t.Fatal("Expected debug output")
	}
	if !strings.Contains(output, "info") {
		t.Fatal("Expected info output")
	}
	if !strings.Contains(output, "warn") {
		t.Fatal("Expected warn output")
	}
	if !strings.Contains(output, "error") {
		t.Fatal("Expected error output")
	}
}

func TestLoggerRedactsPIIFlag(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, false)
	logger.Info("test", Sensitive("token", "secret"))
	output := buf.String()
	if !strings.Contains(output, "REDACTED") {
		t.Fatal("Expected redacted value")
	}
}

func TestLoggerVerboseJSONStructure(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, true)
	logger.Info("test", String("package", "com.example.app"))
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("Expected JSON output: %v", err)
	}
	if parsed["message"] != "test" {
		t.Fatal("Expected message field")
	}
	if parsed["level"] != "INFO" {
		t.Fatal("Expected level field")
	}
}
