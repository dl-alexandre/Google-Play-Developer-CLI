package workflow

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewWatcher(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		opts := DefaultWatcherOptions()
		w := NewWatcher(opts)

		if w == nil {
			t.Fatal("NewWatcher() returned nil")
		}

		if w.opts.Format != WatchFormatText {
			t.Errorf("expected format %v, got %v", WatchFormatText, w.opts.Format)
		}

		if w.opts.UpdateInterval != 100*time.Millisecond {
			t.Errorf("expected update interval 100ms, got %v", w.opts.UpdateInterval)
		}
	})

	t.Run("custom options", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format:         WatchFormatJSON,
			Output:         &buf,
			UpdateInterval: 50 * time.Millisecond,
			ShowTimestamps: true,
		}
		w := NewWatcher(opts)

		if w.opts.Format != WatchFormatJSON {
			t.Errorf("expected format %v, got %v", WatchFormatJSON, w.opts.Format)
		}

		if w.opts.UpdateInterval != 50*time.Millisecond {
			t.Errorf("expected update interval 50ms, got %v", w.opts.UpdateInterval)
		}
	})

	t.Run("nil output defaults to stdout", func(t *testing.T) {
		opts := WatcherOptions{
			Format: WatchFormatText,
			Output: nil,
		}
		w := NewWatcher(opts)

		if w.opts.Output == nil {
			t.Error("expected Output to be set to default")
		}
	})
}

func TestWatcherLifecycle(t *testing.T) {
	t.Run("start and stop", func(t *testing.T) {
		opts := DefaultWatcherOptions()
		w := NewWatcher(opts)

		if w.IsRunning() {
			t.Error("watcher should not be running before Start()")
		}

		w.Start()

		if !w.IsRunning() {
			t.Error("watcher should be running after Start()")
		}

		w.Stop()

		if w.IsRunning() {
			t.Error("watcher should not be running after Stop()")
		}
	})

	t.Run("multiple start calls", func(t *testing.T) {
		opts := DefaultWatcherOptions()
		w := NewWatcher(opts)

		w.Start()
		w.Start() // Should not panic or cause issues

		if !w.IsRunning() {
			t.Error("watcher should be running")
		}

		w.Stop()
	})

	t.Run("stop without start", func(t *testing.T) {
		opts := DefaultWatcherOptions()
		w := NewWatcher(opts)

		// Should not panic
		w.Stop()

		if w.IsRunning() {
			t.Error("watcher should not be running")
		}
	})
}

func TestWatcherEvents(t *testing.T) {
	// Skip on Windows CI due to goroutine timing issues
	if os.Getenv("CI") == "true" && os.Getenv("RUNNER_OS") == "Windows" {
		t.Skip("Skipping watcher tests on Windows CI")
	}

	t.Run("emit workflow started", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format: WatchFormatText,
			Output: &buf,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("test-workflow", "run-123", 5)

		// Give time for event processing then stop before reading
		time.Sleep(50 * time.Millisecond)
		w.Stop()

		output := buf.String()
		if !strings.Contains(output, "test-workflow") {
			t.Errorf("expected output to contain workflow name, got: %s", output)
		}
		if !strings.Contains(output, "5 steps") {
			t.Errorf("expected output to contain step count, got: %s", output)
		}
	})

	t.Run("emit workflow completed", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format: WatchFormatText,
			Output: &buf,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("test-workflow", "run-123", 3)
		w.EmitWorkflowCompleted(5 * time.Second)
		w.Stop()

		output := buf.String()
		if !strings.Contains(output, "completed successfully") {
			t.Errorf("expected output to contain 'completed successfully', got: %s", output)
		}
	})

	t.Run("emit workflow failed", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format: WatchFormatText,
			Output: &buf,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("test-workflow", "run-123", 3)
		w.EmitWorkflowFailed(errors.New("something went wrong"), 2*time.Second)
		w.Stop()

		output := buf.String()
		if !strings.Contains(output, "failed") {
			t.Errorf("expected output to contain 'failed', got: %s", output)
		}
	})

	t.Run("emit step events", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format: WatchFormatText,
			Output: &buf,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("test-workflow", "run-123", 3)
		w.EmitStepStarted("build", 1, 3)
		w.EmitStepCompleted("build", 1, 3, 2*time.Second)
		w.EmitStepStarted("test", 2, 3)
		w.EmitStepFailed("test", 2, 3, errors.New("test failure"), 1)

		// Give time for event processing then stop before reading
		time.Sleep(100 * time.Millisecond)
		w.Stop()

		output := buf.String()
		if !strings.Contains(output, "Executing step: build") {
			t.Errorf("expected output to contain 'Executing step: build', got: %s", output)
		}
		if !strings.Contains(output, "Step completed: build") {
			t.Errorf("expected output to contain 'Step completed: build', got: %s", output)
		}
		if !strings.Contains(output, "Step failed: test") {
			t.Errorf("expected output to contain 'Step failed: test', got: %s", output)
		}
	})

	t.Run("emit step skipped", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format: WatchFormatText,
			Output: &buf,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("test-workflow", "run-123", 3)
		w.EmitStepSkipped("build", 1, 3, "already completed")

		// Give time for event processing then stop before reading
		time.Sleep(50 * time.Millisecond)
		w.Stop()

		output := buf.String()
		if !strings.Contains(output, "Skipping step: build") {
			t.Errorf("expected output to contain 'Skipping step: build', got: %s", output)
		}
		if !strings.Contains(output, "already completed") {
			t.Errorf("expected output to contain skip reason, got: %s", output)
		}
	})
}

func TestWatcherJSONFormat(t *testing.T) {
	t.Run("workflow events as JSON", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format: WatchFormatJSON,
			Output: &buf,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("test-workflow", "run-123", 3)
		w.EmitStepStarted("build", 1, 3)
		w.EmitStepCompleted("build", 1, 3, 2*time.Second)
		w.EmitWorkflowCompleted(5 * time.Second)

		w.Stop()

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		if len(lines) < 4 {
			t.Fatalf("expected at least 4 JSON lines, got %d", len(lines))
		}

		// Verify each line is valid JSON
		for i, line := range lines {
			var event WorkflowEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Errorf("line %d is not valid JSON: %v", i, err)
			}
		}
	})

	t.Run("JSON event structure", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format: WatchFormatJSON,
			Output: &buf,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("my-workflow", "run-456", 5)
		w.Stop()

		var event WorkflowEvent
		if err := json.Unmarshal(buf.Bytes(), &event); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		if event.Type != EventWorkflowStarted {
			t.Errorf("expected type %v, got %v", EventWorkflowStarted, event.Type)
		}
		if event.Workflow != "my-workflow" {
			t.Errorf("expected workflow 'my-workflow', got %s", event.Workflow)
		}
		if event.RunID != "run-456" {
			t.Errorf("expected run ID 'run-456', got %s", event.RunID)
		}
		if event.TotalSteps != 5 {
			t.Errorf("expected 5 total steps, got %d", event.TotalSteps)
		}
		if event.Timestamp.IsZero() {
			t.Error("expected non-zero timestamp")
		}
	})
}

func TestWatcherTimestamps(t *testing.T) {
	t.Run("text format with timestamps", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format:         WatchFormatText,
			Output:         &buf,
			ShowTimestamps: true,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("test-workflow", "run-123", 3)
		w.Stop()

		output := buf.String()
		// Check that output starts with timestamp pattern (HH:MM:SS)
		matched, err := regexp.MatchString(`^\d{2}:\d{2}:\d{2}`, output)
		if err != nil {
			t.Fatalf("regex error: %v", err)
		}
		if !matched {
			t.Errorf("expected output to start with timestamp, got: %s", output)
		}
	})

	t.Run("text format without timestamps", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format:         WatchFormatText,
			Output:         &buf,
			ShowTimestamps: false,
		}
		w := NewWatcher(opts)
		w.Start()

		w.EmitWorkflowStarted("test-workflow", "run-123", 3)
		w.Stop()

		output := buf.String()
		// Should not start with timestamp
		matched, err := regexp.MatchString(`^\d{2}:\d{2}:\d{2}`, output)
		if err != nil {
			t.Fatalf("regex error: %v", err)
		}
		if matched {
			t.Errorf("expected output without timestamp, got: %s", output)
		}
	})
}

func TestWatcherWhenNotRunning(t *testing.T) {
	t.Run("events dropped when not started", func(t *testing.T) {
		var buf bytes.Buffer
		opts := WatcherOptions{
			Format: WatchFormatText,
			Output: &buf,
		}
		w := NewWatcher(opts)
		// Don't start the watcher

		w.EmitWorkflowStarted("test-workflow", "run-123", 3)
		w.EmitStepStarted("build", 1, 3)

		// Output should be empty
		if buf.Len() > 0 {
			t.Errorf("expected no output when watcher not started, got: %s", buf.String())
		}
	})
}

func TestDefaultWatcherOptions(t *testing.T) {
	opts := DefaultWatcherOptions()

	if opts.Format != WatchFormatText {
		t.Errorf("expected default format %v, got %v", WatchFormatText, opts.Format)
	}

	if opts.Output == nil {
		t.Error("expected default output to be set")
	}

	if opts.UpdateInterval != 100*time.Millisecond {
		t.Errorf("expected default update interval 100ms, got %v", opts.UpdateInterval)
	}

	if opts.ShowTimestamps != false {
		t.Errorf("expected default ShowTimestamps to be false, got %v", opts.ShowTimestamps)
	}
}

func TestWatchFormatValues(t *testing.T) {
	formats := []WatchFormat{
		WatchFormatText,
		WatchFormatJSON,
		WatchFormatTUI,
	}

	expected := []string{"text", "json", "tui"}

	for i, format := range formats {
		if string(format) != expected[i] {
			t.Errorf("expected format %s, got %s", expected[i], string(format))
		}
	}
}

func TestWorkflowEventTypes(t *testing.T) {
	types := []WorkflowEventType{
		EventWorkflowStarted,
		EventWorkflowCompleted,
		EventWorkflowFailed,
		EventStepStarted,
		EventStepCompleted,
		EventStepFailed,
		EventStepSkipped,
	}

	expected := []string{
		"workflow_started",
		"workflow_completed",
		"workflow_failed",
		"step_started",
		"step_completed",
		"step_failed",
		"step_skipped",
	}

	for i, eventType := range types {
		if string(eventType) != expected[i] {
			t.Errorf("expected event type %s, got %s", expected[i], string(eventType))
		}
	}
}
