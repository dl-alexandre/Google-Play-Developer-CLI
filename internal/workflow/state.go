package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunStatus represents the current status of a workflow run.
type RunStatus string

const (
	// RunStatusPending indicates the run has not started.
	RunStatusPending RunStatus = "pending"
	// RunStatusRunning indicates the run is in progress.
	RunStatusRunning RunStatus = "running"
	// RunStatusCompleted indicates the run completed successfully.
	RunStatusCompleted RunStatus = "completed"
	// RunStatusFailed indicates the run failed.
	RunStatusFailed RunStatus = "failed"
	// RunStatusCancelled indicates the run was cancelled.
	RunStatusCancelled RunStatus = "cancelled"
)

// RunState represents the persisted state of a workflow run.
type RunState struct {
	RunID       string                `json:"runId"`
	Workflow    Workflow              `json:"workflow"`
	Status      RunStatus             `json:"status"`
	CurrentStep string                `json:"currentStep,omitempty"`
	StepResults []StepResult          `json:"stepResults"`
	StepOutputs map[string]StepOutput `json:"stepOutputs"`
	StartedAt   time.Time             `json:"startedAt"`
	FinishedAt  *time.Time            `json:"finishedAt,omitempty"`
	Error       string                `json:"error,omitempty"`
	Env         map[string]string     `json:"env,omitempty"`
}

// IsStepCompleted checks if a step has been completed in this run.
func (r *RunState) IsStepCompleted(stepName string) bool {
	for _, result := range r.StepResults {
		if result.Step.Name == stepName && result.Output.ExitCode == 0 {
			return true
		}
	}
	return false
}

// GetStepOutput retrieves the output from a completed step.
func (r *RunState) GetStepOutput(stepName string) (StepOutput, bool) {
	output, ok := r.StepOutputs[stepName]
	return output, ok
}

// GetNextPendingStep returns the first step that hasn't been completed yet.
func (r *RunState) GetNextPendingStep() *Step {
	for i := range r.Workflow.Steps {
		if !r.IsStepCompleted(r.Workflow.Steps[i].Name) {
			return &r.Workflow.Steps[i]
		}
	}
	return nil
}

// AddStepResult adds a step result to the run state.
func (r *RunState) AddStepResult(result StepResult) {
	r.StepResults = append(r.StepResults, result)
	r.StepOutputs[result.Step.Name] = result.Output
	if result.Output.ExitCode != 0 && result.Output.Error != "" {
		r.Error = result.Output.Error
	}
}

// StateManager handles persistence of workflow run states.
type StateManager struct {
	baseDir string
}

// NewStateManager creates a new state manager.
func NewStateManager(baseDir string) *StateManager {
	return &StateManager{
		baseDir: baseDir,
	}
}

// EnsureDefaultDirs creates the default directory structure if it doesn't exist.
func (sm *StateManager) EnsureDefaultDirs() error {
	dirs := []string{
		sm.baseDir,
		filepath.Join(sm.baseDir, "definitions"),
		filepath.Join(sm.baseDir, "runs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// Save persists a run state to disk.
func (sm *StateManager) Save(state *RunState) error {
	if err := sm.EnsureDefaultDirs(); err != nil {
		return err
	}

	path := sm.runStatePath(state.RunID)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal run state: %w", err)
	}

	// Write to temp file first, then rename for atomicity
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0640); err != nil {
		return fmt.Errorf("failed to write run state: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to save run state: %w", err)
	}

	return nil
}

// Load retrieves a run state from disk.
func (sm *StateManager) Load(runID string) (*RunState, error) {
	path := sm.runStatePath(runID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("run %s not found: %w", runID, err)
		}
		return nil, fmt.Errorf("failed to read run state: %w", err)
	}

	var state RunState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal run state: %w", err)
	}

	return &state, nil
}

// List returns all saved run states.
func (sm *StateManager) List() ([]RunState, error) {
	runsDir := filepath.Join(sm.baseDir, "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []RunState{}, nil
		}
		return nil, fmt.Errorf("failed to list runs: %w", err)
	}

	var states []RunState
	for _, entry := range entries {
		if entry.IsDir() || !hasJSONExtension(entry.Name()) {
			continue
		}

		runID := entry.Name()[:len(entry.Name())-5] // Remove .json
		state, err := sm.Load(runID)
		if err != nil {
			continue // Skip corrupted files
		}
		states = append(states, *state)
	}

	return states, nil
}

// Delete removes a run state from disk.
func (sm *StateManager) Delete(runID string) error {
	path := sm.runStatePath(runID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete run state: %w", err)
	}
	return nil
}

// GenerateRunID creates a unique run ID.
func GenerateRunID() string {
	return fmt.Sprintf("%d-%s", time.Now().Unix(), generateRandomID(8))
}

// NewRunState creates a new run state for a workflow.
func NewRunState(workflow Workflow) *RunState {
	return &RunState{
		RunID:       GenerateRunID(),
		Workflow:    workflow,
		Status:      RunStatusPending,
		StepResults: []StepResult{},
		StepOutputs: make(map[string]StepOutput),
		StartedAt:   time.Now(),
		Env:         workflow.Env,
	}
}

func (sm *StateManager) runStatePath(runID string) string {
	return filepath.Join(sm.baseDir, "runs", runID+".json")
}

func hasJSONExtension(name string) bool {
	return len(name) > 5 && name[len(name)-5:] == ".json"
}

func generateRandomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// ResumeInfo contains information needed to resume a workflow.
type ResumeInfo struct {
	CanResume      bool
	RunID          string
	Status         RunStatus
	FailedStep     string
	CompletedCount int
	TotalSteps     int
}

// GetResumeInfo returns information about whether a run can be resumed.
func (sm *StateManager) GetResumeInfo(runID string) (*ResumeInfo, error) {
	state, err := sm.Load(runID)
	if err != nil {
		return nil, err
	}

	info := &ResumeInfo{
		RunID:      runID,
		Status:     state.Status,
		TotalSteps: len(state.Workflow.Steps),
	}

	// Count completed steps
	for _, result := range state.StepResults {
		if result.Output.ExitCode == 0 {
			info.CompletedCount++
		} else if info.FailedStep == "" {
			info.FailedStep = result.Step.Name
		}
	}

	// Can resume if not already completed or cancelled
	info.CanResume = state.Status != RunStatusCompleted &&
		state.Status != RunStatusCancelled &&
		info.CompletedCount < info.TotalSteps

	return info, nil
}
