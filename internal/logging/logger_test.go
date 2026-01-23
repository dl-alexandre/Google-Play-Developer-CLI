package logging

import (
	"bytes"
	"strings"
	"testing"
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

	t.Run("Bool", func(t *testing.T) {
		f := Bool("enabled", true)
		if f.Key != "enabled" || f.Value != true {
			t.Error("Bool field not created correctly")
		}
	})

	t.Run("Sensitive", func(t *testing.T) {
		f := Sensitive("token", "secret123")
		if f.Key != "token" || !f.PII {
			t.Error("Sensitive field not marked as PII")
		}
	})
}
