// Package workflow provides declarative workflow execution with step outputs and resumable state.
package workflow

import (
	"time"
)

// StepType represents the type of a workflow step.
type StepType string

const (
	// StepTypeShell executes a shell command.
	StepTypeShell StepType = "shell"
	// StepTypeGPD executes a gpd CLI command.
	StepTypeGPD StepType = "gpd"
)

// Workflow represents a declarative workflow definition.
type Workflow struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	MaxParallel int               `json:"maxParallel,omitempty"`
	Steps       []Step            `json:"steps"`
}

// Step represents a single workflow step.
type Step struct {
	Name            string            `json:"name"`
	Type            StepType          `json:"type,omitempty"`
	Command         string            `json:"command"`
	WorkingDir      string            `json:"workingDir,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	DependsOn       []string          `json:"dependsOn,omitempty"`
	CaptureOutputs  []string          `json:"captureOutputs,omitempty"`
	Condition       string            `json:"condition,omitempty"`
	ContinueOnError bool              `json:"continueOnError,omitempty"`
	Timeout         time.Duration     `json:"timeout,omitempty"`
	RetryCount      int               `json:"retryCount,omitempty"`
	RetryDelay      time.Duration     `json:"retryDelay,omitempty"`
	RetryBackoff    string            `json:"retryBackoff,omitempty"`
	Parallel        bool              `json:"parallel,omitempty"`
}

// StepOutput represents the captured output from a step execution.
type StepOutput struct {
	StepName   string                 `json:"stepName"`
	ExitCode   int                    `json:"exitCode"`
	Stdout     string                 `json:"stdout,omitempty"`
	Stderr     string                 `json:"stderr,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	StartedAt  time.Time              `json:"startedAt"`
	FinishedAt time.Time              `json:"finishedAt"`
	Duration   time.Duration          `json:"duration"`
	Error      string                 `json:"error,omitempty"`
	RetryCount int                    `json:"retryCount,omitempty"`
	Retries    int                    `json:"retries,omitempty"`
}

// StepResult represents the result of executing a single step.
type StepResult struct {
	Step   Step       `json:"step"`
	Output StepOutput `json:"output"`
}

// Validate performs basic validation on the workflow.
func (w *Workflow) Validate() error {
	if w.Name == "" {
		return ErrWorkflowNameRequired
	}

	if len(w.Steps) == 0 {
		return ErrNoStepsDefined
	}

	// Check for duplicate step names
	names := make(map[string]bool)
	for i, step := range w.Steps {
		if step.Name == "" {
			return &ValidationError{Field: "steps", Message: "step at index " + string(rune('0'+i)) + " missing name"}
		}
		if names[step.Name] {
			return &ValidationError{Field: "steps", Message: "duplicate step name: " + step.Name}
		}
		names[step.Name] = true

		if step.Command == "" {
			return &ValidationError{Field: "steps", Message: "step " + step.Name + " missing command"}
		}

		// Validate step type
		if step.Type != "" && step.Type != StepTypeShell && step.Type != StepTypeGPD {
			return &ValidationError{Field: "steps", Message: "step " + step.Name + " has invalid type: " + string(step.Type)}
		}

		// Validate dependencies exist
		for _, dep := range step.DependsOn {
			if !names[dep] {
				// Check if it's a forward reference
				found := false
				for _, s := range w.Steps {
					if s.Name == dep {
						found = true
						break
					}
				}
				if !found {
					return &ValidationError{Field: "steps", Message: "step " + step.Name + " has unknown dependency: " + dep}
				}
			}
		}
	}

	return nil
}

// GetStep returns a step by name.
func (w *Workflow) GetStep(name string) (*Step, bool) {
	for i := range w.Steps {
		if w.Steps[i].Name == name {
			return &w.Steps[i], true
		}
	}
	return nil, false
}

// ValidationError represents a workflow validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Common validation errors.
var (
	ErrWorkflowNameRequired = &ValidationError{Field: "name", Message: "workflow name is required"}
	ErrNoStepsDefined       = &ValidationError{Field: "steps", Message: "at least one step is required"}
)
