package workflow

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// WatchFormat defines the output format for watch mode.
type WatchFormat string

const (
	// WatchFormatText displays simple text progress updates.
	WatchFormatText WatchFormat = "text"
	// WatchFormatJSON streams progress events as JSON lines.
	WatchFormatJSON WatchFormat = "json"
	// WatchFormatTUI displays a TUI-style progress bar.
	WatchFormatTUI WatchFormat = "tui"
)

// WorkflowEventType represents the type of workflow event.
type WorkflowEventType string

const (
	// EventWorkflowStarted is emitted when workflow execution begins.
	EventWorkflowStarted WorkflowEventType = "workflow_started"
	// EventWorkflowCompleted is emitted when workflow finishes successfully.
	EventWorkflowCompleted WorkflowEventType = "workflow_completed"
	// EventWorkflowFailed is emitted when workflow fails.
	EventWorkflowFailed WorkflowEventType = "workflow_failed"
	// EventStepStarted is emitted when a step begins execution.
	EventStepStarted WorkflowEventType = "step_started"
	// EventStepCompleted is emitted when a step completes successfully.
	EventStepCompleted WorkflowEventType = "step_completed"
	// EventStepFailed is emitted when a step fails.
	EventStepFailed WorkflowEventType = "step_failed"
	// EventStepSkipped is emitted when a step is skipped (condition not met or already completed).
	EventStepSkipped WorkflowEventType = "step_skipped"
)

// WorkflowEvent represents an event during workflow execution.
type WorkflowEvent struct {
	Type       WorkflowEventType `json:"type"`
	Timestamp  time.Time         `json:"timestamp"`
	Workflow   string            `json:"workflow"`
	RunID      string            `json:"runId"`
	StepName   string            `json:"stepName,omitempty"`
	StepNum    int               `json:"stepNum,omitempty"`
	TotalSteps int               `json:"totalSteps,omitempty"`
	Duration   time.Duration     `json:"duration,omitempty"`
	Error      string            `json:"error,omitempty"`
	ExitCode   int               `json:"exitCode,omitempty"`
	Message    string            `json:"message,omitempty"`
}

// WatcherOptions configures the watcher behavior.
type WatcherOptions struct {
	// Format specifies the output format (text, json, tui).
	Format WatchFormat
	// Output is the writer to write events to (defaults to stdout).
	Output io.Writer
	// UpdateInterval controls how often progress is updated in TUI mode.
	UpdateInterval time.Duration
	// ShowTimestamps includes timestamps in text output.
	ShowTimestamps bool
}

// DefaultWatcherOptions returns default watcher options.
func DefaultWatcherOptions() WatcherOptions {
	return WatcherOptions{
		Format:         WatchFormatText,
		Output:         os.Stdout,
		UpdateInterval: 100 * time.Millisecond,
		ShowTimestamps: false,
	}
}

// Watcher monitors workflow execution and emits progress events.
type Watcher struct {
	opts     WatcherOptions
	events   chan WorkflowEvent
	stop     chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
	started  bool
	workflow string
	runID    string
	outputMu sync.Mutex // Protects output writes to prevent races in tests
}

// NewWatcher creates a new workflow watcher with the given options.
func NewWatcher(opts WatcherOptions) *Watcher {
	if opts.Output == nil {
		opts.Output = os.Stdout
	}
	if opts.UpdateInterval <= 0 {
		opts.UpdateInterval = 100 * time.Millisecond
	}

	return &Watcher{
		opts:   opts,
		events: make(chan WorkflowEvent, 100),
		stop:   make(chan struct{}),
	}
}

// Start begins watching for workflow events.
func (w *Watcher) Start() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		return
	}

	w.started = true
	w.wg.Add(1)
	go w.eventLoop()
}

// Stop stops the watcher and waits for pending events to be processed.
func (w *Watcher) Stop() {
	w.mu.Lock()
	if !w.started {
		w.mu.Unlock()
		return
	}
	w.started = false
	w.mu.Unlock()

	close(w.stop)
	w.wg.Wait()

	// Drain any remaining events to ensure all output operations complete
	// This prevents race conditions when tests read from output buffers
	w.drainEvents()
}

// EmitWorkflowStarted emits an event when workflow execution begins.
func (w *Watcher) EmitWorkflowStarted(workflow, runID string, totalSteps int) {
	w.mu.RLock()
	if !w.started {
		w.mu.RUnlock()
		return
	}
	w.workflow = workflow
	w.runID = runID
	w.mu.RUnlock()

	event := WorkflowEvent{
		Type:       EventWorkflowStarted,
		Timestamp:  time.Now(),
		Workflow:   workflow,
		RunID:      runID,
		TotalSteps: totalSteps,
		Message:    fmt.Sprintf("Starting workflow: %s (%d steps)", workflow, totalSteps),
	}

	w.sendEvent(event)
}

// EmitWorkflowCompleted emits an event when workflow finishes successfully.
func (w *Watcher) EmitWorkflowCompleted(duration time.Duration) {
	event := WorkflowEvent{
		Type:      EventWorkflowCompleted,
		Timestamp: time.Now(),
		Workflow:  w.workflow,
		RunID:     w.runID,
		Duration:  duration,
		Message:   fmt.Sprintf("Workflow completed successfully: %s (duration: %v)", w.workflow, duration),
	}

	w.sendEvent(event)
}

// EmitWorkflowFailed emits an event when workflow fails.
func (w *Watcher) EmitWorkflowFailed(err error, duration time.Duration) {
	event := WorkflowEvent{
		Type:      EventWorkflowFailed,
		Timestamp: time.Now(),
		Workflow:  w.workflow,
		RunID:     w.runID,
		Duration:  duration,
		Error:     err.Error(),
		Message:   fmt.Sprintf("Workflow failed: %s - %v", w.workflow, err),
	}

	w.sendEvent(event)
}

// EmitStepStarted emits an event when a step begins execution.
func (w *Watcher) EmitStepStarted(stepName string, stepNum, totalSteps int) {
	event := WorkflowEvent{
		Type:       EventStepStarted,
		Timestamp:  time.Now(),
		Workflow:   w.workflow,
		RunID:      w.runID,
		StepName:   stepName,
		StepNum:    stepNum,
		TotalSteps: totalSteps,
		Message:    fmt.Sprintf("[%d/%d] Executing step: %s", stepNum, totalSteps, stepName),
	}

	w.sendEvent(event)
}

// EmitStepCompleted emits an event when a step completes successfully.
func (w *Watcher) EmitStepCompleted(stepName string, stepNum, totalSteps int, duration time.Duration) {
	event := WorkflowEvent{
		Type:       EventStepCompleted,
		Timestamp:  time.Now(),
		Workflow:   w.workflow,
		RunID:      w.runID,
		StepName:   stepName,
		StepNum:    stepNum,
		TotalSteps: totalSteps,
		Duration:   duration,
		Message:    fmt.Sprintf("[%d/%d] Step completed: %s (duration: %v)", stepNum, totalSteps, stepName, duration),
	}

	w.sendEvent(event)
}

// EmitStepFailed emits an event when a step fails.
func (w *Watcher) EmitStepFailed(stepName string, stepNum, totalSteps int, err error, exitCode int) {
	event := WorkflowEvent{
		Type:       EventStepFailed,
		Timestamp:  time.Now(),
		Workflow:   w.workflow,
		RunID:      w.runID,
		StepName:   stepName,
		StepNum:    stepNum,
		TotalSteps: totalSteps,
		Error:      err.Error(),
		ExitCode:   exitCode,
		Message:    fmt.Sprintf("[%d/%d] Step failed: %s (exit code: %d) - %v", stepNum, totalSteps, stepName, exitCode, err),
	}

	w.sendEvent(event)
}

// EmitStepSkipped emits an event when a step is skipped.
func (w *Watcher) EmitStepSkipped(stepName string, stepNum, totalSteps int, reason string) {
	event := WorkflowEvent{
		Type:       EventStepSkipped,
		Timestamp:  time.Now(),
		Workflow:   w.workflow,
		RunID:      w.runID,
		StepName:   stepName,
		StepNum:    stepNum,
		TotalSteps: totalSteps,
		Message:    fmt.Sprintf("[%d/%d] Skipping step: %s (%s)", stepNum, totalSteps, stepName, reason),
	}

	w.sendEvent(event)
}

func (w *Watcher) sendEvent(event WorkflowEvent) {
	select {
	case w.events <- event:
	case <-w.stop:
		// Watcher is stopping, drop event
	default:
		// Channel full, drop oldest event
		select {
		case <-w.events:
		default:
		}
		w.events <- event
	}
}

func (w *Watcher) eventLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.opts.UpdateInterval)
	defer ticker.Stop()

	// Track state for TUI mode
	var (
		currentStep    string
		stepNum        int
		totalSteps     int
		completedSteps int
		startTime      time.Time
	)

	for {
		select {
		case <-w.stop:
			// Process remaining events
			w.drainEvents()
			return

		case event := <-w.events:
			// Update state
			switch event.Type {
			case EventWorkflowStarted:
				startTime = event.Timestamp
				totalSteps = event.TotalSteps
			case EventStepStarted:
				currentStep = event.StepName
				stepNum = event.StepNum
			case EventStepCompleted, EventStepFailed:
				completedSteps++
			}

			// Output based on format
			switch w.opts.Format {
			case WatchFormatJSON:
				w.outputJSON(event)
			case WatchFormatTUI:
				w.outputTUI(currentStep, stepNum, totalSteps, completedSteps, startTime)
			default:
				w.outputText(event)
			}

		case <-ticker.C:
			// Update TUI display for long-running steps
			if w.opts.Format == WatchFormatTUI && currentStep != "" {
				w.outputTUI(currentStep, stepNum, totalSteps, completedSteps, startTime)
			}
		}
	}
}

func (w *Watcher) drainEvents() {
	for {
		select {
		case event := <-w.events:
			switch w.opts.Format {
			case WatchFormatJSON:
				w.outputJSON(event)
			default:
				w.outputText(event)
			}
		default:
			return
		}
	}
}

func (w *Watcher) outputText(event WorkflowEvent) {
	var sb strings.Builder

	if w.opts.ShowTimestamps {
		sb.WriteString(event.Timestamp.Format("15:04:05"))
		sb.WriteString(" ")
	}

	switch event.Type {
	case EventWorkflowStarted:
		sb.WriteString("▶ ")
		sb.WriteString(event.Message)
	case EventWorkflowCompleted:
		sb.WriteString("✓ ")
		sb.WriteString(event.Message)
	case EventWorkflowFailed:
		sb.WriteString("✗ ")
		sb.WriteString(event.Message)
	case EventStepStarted:
		sb.WriteString("→ ")
		sb.WriteString(event.Message)
	case EventStepCompleted:
		sb.WriteString("✓ ")
		sb.WriteString(event.Message)
	case EventStepFailed:
		sb.WriteString("✗ ")
		sb.WriteString(event.Message)
	case EventStepSkipped:
		sb.WriteString("⊘ ")
		sb.WriteString(event.Message)
	}

	sb.WriteString("\n")

	w.outputMu.Lock()
	_, _ = fmt.Fprint(w.opts.Output, sb.String())
	w.outputMu.Unlock()
}

func (w *Watcher) outputJSON(event WorkflowEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	w.outputMu.Lock()
	_, _ = fmt.Fprintln(w.opts.Output, string(data))
	w.outputMu.Unlock()
}

func (w *Watcher) outputTUI(currentStep string, stepNum, totalSteps, completedSteps int, startTime time.Time) {
	// Calculate progress
	progress := 0.0
	if totalSteps > 0 {
		progress = float64(completedSteps) / float64(totalSteps)
	}

	// Build progress bar
	barWidth := 30
	filled := int(progress * float64(barWidth))
	empty := barWidth - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	// Calculate elapsed time
	elapsed := time.Since(startTime)

	w.outputMu.Lock()
	defer w.outputMu.Unlock()

	// Clear previous line if not first output
	if stepNum > 0 || completedSteps > 0 {
		_, _ = fmt.Fprint(w.opts.Output, "\r\033[K")
	}

	// Format output
	if currentStep != "" && stepNum > 0 {
		_, _ = fmt.Fprintf(w.opts.Output, "[%s] %3.0f%% | %d/%d | %s | %v",
			bar,
			progress*100,
			completedSteps,
			totalSteps,
			currentStep,
			elapsed,
		)
	} else {
		_, _ = fmt.Fprintf(w.opts.Output, "[%s] %3.0f%% | %d/%d | %v",
			bar,
			progress*100,
			completedSteps,
			totalSteps,
			elapsed,
		)
	}

	// Final newline on completion
	if completedSteps == totalSteps && totalSteps > 0 {
		_, _ = fmt.Fprintln(w.opts.Output)
	}
}

// IsRunning returns true if the watcher is currently running.
func (w *Watcher) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.started
}
