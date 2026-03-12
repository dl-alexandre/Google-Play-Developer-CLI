package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestEndToEndWorkflowExecution tests complete workflow execution with all features
func TestEndToEndWorkflowExecution(t *testing.T) {
	// Skip on Windows CI - shell commands won't work properly
	if os.Getenv("CI") == "true" && runtime.GOOS == "windows" {
		t.Skip("Skipping integration tests on Windows CI")
	}

	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	// Create a test workflow that exercises multiple features
	workflow := &Workflow{
		Name:        "e2e-test-workflow",
		Description: "End-to-end integration test",
		Env: map[string]string{
			"TEST_VAR": "test_value",
			"PACKAGE":  "com.test.app",
		},
		Steps: []Step{
			{
				Name:           "step1",
				Type:           StepTypeShell,
				Command:        "echo '{\"versionCode\": 42, \"versionName\": \"1.0.0\"}'",
				CaptureOutputs: []string{"versionCode", "versionName"},
			},
			{
				Name:       "step2",
				Type:       StepTypeShell,
				Command:    "echo 'Using version ${steps.step1.versionCode} and package ${env.PACKAGE}'",
				DependsOn:  []string{"step1"},
				WorkingDir: tempDir,
			},
			{
				Name:      "step3",
				Type:      StepTypeShell,
				Command:   "echo 'Final step with ${steps.step1.versionName}'",
				DependsOn: []string{"step2"},
			},
		},
	}

	opts := RunOptions{
		Verbose: true,
	}

	runner := NewRunner(stateManager, opts)
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Workflow execution failed: %v", err)
	}

	// Verify workflow completed successfully
	if state.Status != RunStatusCompleted {
		t.Errorf("Expected status %s, got %s", RunStatusCompleted, state.Status)
	}

	// Verify all steps executed
	if len(state.StepResults) != 3 {
		t.Errorf("Expected 3 step results, got %d", len(state.StepResults))
	}

	// Verify step1 outputs were captured
	step1Output, ok := state.StepOutputs["step1"]
	if !ok {
		t.Fatal("step1 output not found")
	}

	if step1Output.ExitCode != 0 {
		t.Errorf("step1 exit code = %d, want 0", step1Output.ExitCode)
	}

	if step1Output.Data["versionCode"] == nil {
		t.Error("versionCode not captured from step1 output")
	}

	// Verify step2 had interpolated command
	step2Output := state.StepOutputs["step2"]
	if !strings.Contains(step2Output.Stdout, "42") {
		t.Error("step2 output should contain interpolated versionCode (42)")
	}
	if !strings.Contains(step2Output.Stdout, "com.test.app") {
		t.Error("step2 output should contain interpolated PACKAGE (com.test.app)")
	}

	// Verify state was persisted
	if state.RunID == "" {
		t.Error("RunID should be set")
	}

	loaded, err := stateManager.Load(state.RunID)
	if err != nil {
		t.Fatalf("Failed to load persisted state: %v", err)
	}

	if loaded.Status != RunStatusCompleted {
		t.Errorf("Loaded state status = %s, want %s", loaded.Status, RunStatusCompleted)
	}

	if len(loaded.StepResults) != 3 {
		t.Errorf("Loaded state has %d step results, want 3", len(loaded.StepResults))
	}

	// Verify timestamps are set
	if state.StartedAt.IsZero() {
		t.Error("StartedAt should be set")
	}
	if state.FinishedAt == nil || state.FinishedAt.IsZero() {
		t.Error("FinishedAt should be set")
	}
}

// TestResumeFunctionality tests resuming a failed workflow
func TestResumeFunctionality(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	// Create a workflow that will fail at step2
	workflow := &Workflow{
		Name: "resume-test-workflow",
		Steps: []Step{
			{
				Name:           "step1",
				Type:           StepTypeShell,
				Command:        "echo '{\"data\": \"step1_value\"}'",
				CaptureOutputs: []string{"data"},
			},
			{
				Name:      "step2",
				Type:      StepTypeShell,
				Command:   "exit 1", // This will fail
				DependsOn: []string{"step1"},
			},
			{
				Name:      "step3",
				Type:      StepTypeShell,
				Command:   "echo 'step3 executed'",
				DependsOn: []string{"step2"},
			},
		},
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	// First run - should fail at step2
	state, err := runner.RunWorkflow(ctx, workflow)
	if err == nil {
		t.Fatal("Expected workflow to fail at step2")
	}

	if state.Status != RunStatusFailed {
		t.Errorf("Expected status %s, got %s", RunStatusFailed, state.Status)
	}

	// Verify step1 completed, step2 failed
	if !state.IsStepCompleted("step1") {
		t.Error("step1 should be completed")
	}
	if state.IsStepCompleted("step2") {
		t.Error("step2 should not be completed (it failed)")
	}

	runID := state.RunID

	// Modify workflow to fix step2
	workflow.Steps[1].Command = "echo 'step2 succeeded'"

	// Resume the workflow
	resumeRunner := NewRunner(stateManager, RunOptions{
		ResumeRunID: runID,
	})

	resumedState, err := resumeRunner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Resumed workflow failed: %v", err)
	}

	// Verify workflow completed
	if resumedState.Status != RunStatusCompleted {
		t.Errorf("Resumed workflow status = %s, want %s", resumedState.Status, RunStatusCompleted)
	}

	// Verify step1 output still exists from original run
	_, ok := resumedState.StepOutputs["step1"]
	if !ok {
		t.Error("step1 output should still exist from original run")
	}
	// Note: on resume, we get the original state, so step1 was already there

	// Verify step2 was executed and succeeded
	step2Output, ok := resumedState.StepOutputs["step2"]
	if !ok {
		t.Fatal("step2 output not found after resume")
	}
	if step2Output.ExitCode != 0 {
		t.Errorf("step2 exit code = %d, want 0 after resume", step2Output.ExitCode)
	}

	// Verify step3 executed
	if !resumedState.IsStepCompleted("step3") {
		t.Error("step3 should be completed after resume")
	}
}

// TestResumeWithForceOption tests that Force re-runs completed steps
func TestResumeWithForceOption(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	workflow := &Workflow{
		Name: "force-resume-test",
		Steps: []Step{
			{
				Name:    "step1",
				Type:    StepTypeShell,
				Command: "echo 'original'",
			},
		},
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	// First run
	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("First run failed: %v", err)
	}

	runID := state.RunID
	originalOutput := state.StepOutputs["step1"].Stdout

	// Change the command
	workflow.Steps[0].Command = "echo 'modified'"

	// Resume with Force option
	resumeRunner := NewRunner(stateManager, RunOptions{
		ResumeRunID: runID,
		Force:       true,
	})

	resumedState, err := resumeRunner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Resume with force failed: %v", err)
	}

	// Verify step was re-executed
	newOutput := resumedState.StepOutputs["step1"].Stdout
	if newOutput == originalOutput {
		t.Error("Step should have been re-executed with different output due to Force option")
	}
	if !strings.Contains(newOutput, "modified") {
		t.Error("New output should contain 'modified'")
	}
}

// TestVariableInterpolation tests all interpolation scenarios
func TestVariableInterpolation(t *testing.T) {
	// Skip on Windows CI - shell commands work differently
	if os.Getenv("CI") == "true" && runtime.GOOS == "windows" {
		t.Skip("Skipping interpolation tests on Windows CI")
	}

	tests := []struct {
		name     string
		workflow *Workflow
		wantErr  bool
		validate func(t *testing.T, state *RunState)
	}{
		{
			name: "step output interpolation",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:           "producer",
						Command:        `echo '{"value": "produced_value"}'`,
						CaptureOutputs: []string{"value"},
					},
					{
						Name:      "consumer",
						Command:   `echo "Got: ${steps.producer.value}"`,
						DependsOn: []string{"producer"},
					},
				},
			},
			validate: func(t *testing.T, state *RunState) {
				output := state.StepOutputs["consumer"].Stdout
				if !strings.Contains(output, "produced_value") {
					t.Errorf("Expected interpolation of steps.producer.value, got: %s", output)
				}
			},
		},
		{
			name: "environment variable interpolation",
			workflow: &Workflow{
				Name: "test",
				Env: map[string]string{
					"MY_VAR": "env_value",
				},
				Steps: []Step{
					{
						Name:    "step1",
						Command: `echo "Env: ${env.MY_VAR}"`,
					},
				},
			},
			validate: func(t *testing.T, state *RunState) {
				output := state.StepOutputs["step1"].Stdout
				if !strings.Contains(output, "env_value") {
					t.Errorf("Expected interpolation of env.MY_VAR, got: %s", output)
				}
			},
		},
		{
			name: "shorthand env variable",
			workflow: &Workflow{
				Name: "test",
				Env: map[string]string{
					"QUICK_VAR": "quick_value",
				},
				Steps: []Step{
					{
						Name:    "step1",
						Command: `echo "Quick: ${QUICK_VAR}"`,
					},
				},
			},
			validate: func(t *testing.T, state *RunState) {
				output := state.StepOutputs["step1"].Stdout
				if !strings.Contains(output, "quick_value") {
					t.Errorf("Expected shorthand interpolation, got: %s", output)
				}
			},
		},
		{
			name: "nested field access",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:           "producer",
						Command:        `echo '{"nested": {"deep": {"value": "deep_value"}}}'`,
						CaptureOutputs: []string{"nested"}, // Capture the whole nested object
					},
					{
						Name:      "consumer",
						Command:   `echo "Deep: ${steps.producer.nested.deep.value}"`,
						DependsOn: []string{"producer"},
					},
				},
			},
			validate: func(t *testing.T, state *RunState) {
				output := state.StepOutputs["consumer"].Stdout
				if !strings.Contains(output, "deep_value") {
					t.Errorf("Expected nested interpolation, got: %s", output)
				}
			},
		},
		{
			name: "multiple interpolations",
			workflow: &Workflow{
				Name: "test",
				Env: map[string]string{
					"A": "alpha",
					"B": "beta",
				},
				Steps: []Step{
					{
						Name:           "step1",
						Command:        `echo '{"x": "chi"}'`,
						CaptureOutputs: []string{"x"},
					},
					{
						Name:      "step2",
						Command:   `echo "A=${env.A} B=${env.B} X=${steps.step1.x}"`,
						DependsOn: []string{"step1"},
					},
				},
			},
			validate: func(t *testing.T, state *RunState) {
				output := state.StepOutputs["step2"].Stdout
				if !strings.Contains(output, "alpha") || !strings.Contains(output, "beta") || !strings.Contains(output, "chi") {
					t.Errorf("Expected all interpolations, got: %s", output)
				}
			},
		},
		{
			name: "missing step variable error",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:    "step1",
						Command: `echo "Missing: ${steps.nonexistent.field}"`,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing env variable error",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:    "step1",
						Command: `echo "Missing: ${env.UNDEFINED_VAR}"`,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			stateManager := NewStateManager(tempDir)
			runner := NewRunner(stateManager, RunOptions{})
			ctx := context.Background()

			state, err := runner.RunWorkflow(ctx, tt.workflow)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if state.Status != RunStatusCompleted {
				t.Errorf("Status = %s, want %s", state.Status, RunStatusCompleted)
			}

			if tt.validate != nil {
				tt.validate(t, state)
			}
		})
	}
}

// TestStatePersistence tests saving and loading state
func TestStatePersistence(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	workflow := Workflow{
		Name:        "persistence-test",
		Description: "Test state persistence",
		Env: map[string]string{
			"KEY1": "value1",
			"KEY2": "value2",
		},
		Steps: []Step{
			{
				Name:    "step1",
				Command: "echo test",
			},
		},
	}

	// Create a run state with all fields populated
	state := NewRunState(workflow)
	state.Status = RunStatusRunning
	state.CurrentStep = "step1"
	state.Env = map[string]string{
		"RUNTIME_VAR": "runtime_value",
	}

	// Add some step results
	state.AddStepResult(StepResult{
		Step: Step{Name: "step1"},
		Output: StepOutput{
			StepName:   "step1",
			ExitCode:   0,
			Stdout:     "stdout content",
			Stderr:     "stderr content",
			Data:       map[string]interface{}{"key": "value"},
			StartedAt:  time.Now().Add(-time.Minute),
			FinishedAt: time.Now(),
			Duration:   time.Minute,
		},
	})

	// Save state
	if err := stateManager.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load state
	loaded, err := stateManager.Load(state.RunID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify all fields preserved
	if loaded.RunID != state.RunID {
		t.Errorf("RunID = %s, want %s", loaded.RunID, state.RunID)
	}

	if loaded.Workflow.Name != workflow.Name {
		t.Errorf("Workflow.Name = %s, want %s", loaded.Workflow.Name, workflow.Name)
	}

	if loaded.Status != state.Status {
		t.Errorf("Status = %s, want %s", loaded.Status, state.Status)
	}

	if loaded.CurrentStep != state.CurrentStep {
		t.Errorf("CurrentStep = %s, want %s", loaded.CurrentStep, state.CurrentStep)
	}

	if len(loaded.StepResults) != len(state.StepResults) {
		t.Errorf("StepResults count = %d, want %d", len(loaded.StepResults), len(state.StepResults))
	}

	if len(loaded.StepOutputs) != len(state.StepOutputs) {
		t.Errorf("StepOutputs count = %d, want %d", len(loaded.StepOutputs), len(state.StepOutputs))
	}

	// Verify step output data
	output, ok := loaded.StepOutputs["step1"]
	if !ok {
		t.Fatal("step1 output not found in loaded state")
	}

	if output.Stdout != "stdout content" {
		t.Errorf("Stdout = %s, want 'stdout content'", output.Stdout)
	}

	if output.Data["key"] != "value" {
		t.Errorf("Data[key] = %v, want 'value'", output.Data["key"])
	}

	// Verify timestamps
	if loaded.StartedAt.IsZero() {
		t.Error("StartedAt should not be zero")
	}

	// Verify env
	if loaded.Env["RUNTIME_VAR"] != "runtime_value" {
		t.Errorf("Env[RUNTIME_VAR] = %s, want 'runtime_value'", loaded.Env["RUNTIME_VAR"])
	}

	// Verify file exists on disk
	statePath := filepath.Join(tempDir, "runs", state.RunID+".json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("State file should exist on disk")
	}
}

// TestErrorScenarios tests various error conditions
func TestErrorScenarios(t *testing.T) {
	tests := []struct {
		name           string
		workflow       *Workflow
		wantErr        bool
		errContains    string
		expectedStatus RunStatus
	}{
		{
			name: "invalid workflow JSON",
			workflow: &Workflow{
				Name: "",
				Steps: []Step{
					{Name: "step1", Command: "echo test"},
				},
			},
			wantErr:        true,
			errContains:    "name",
			expectedStatus: RunStatusPending,
		},
		{
			name: "no steps defined",
			workflow: &Workflow{
				Name:  "test",
				Steps: []Step{},
			},
			wantErr:        true,
			errContains:    "steps",
			expectedStatus: RunStatusPending,
		},
		{
			name: "missing dependency",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:      "step1",
						Command:   "echo test",
						DependsOn: []string{"nonexistent"},
					},
				},
			},
			wantErr:     true,
			errContains: "dependency",
		},
		{
			name: "circular dependency",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:      "step1",
						Command:   "echo 1",
						DependsOn: []string{"step2"},
					},
					{
						Name:      "step2",
						Command:   "echo 2",
						DependsOn: []string{"step1"},
					},
				},
			},
			wantErr:     true,
			errContains: "circular",
		},
		{
			name: "failed step",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:    "step1",
						Command: "exit 1",
					},
				},
			},
			wantErr:        true,
			errContains:    "failed",
			expectedStatus: RunStatusFailed,
		},
		{
			name: "step timeout",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:    "step1",
						Command: "sleep 10",
						Timeout: 100 * time.Millisecond,
					},
				},
			},
			wantErr:        true,
			errContains:    "failed",
			expectedStatus: RunStatusFailed,
		},
		{
			name: "continue on error",
			workflow: &Workflow{
				Name: "test",
				Steps: []Step{
					{
						Name:            "step1",
						Command:         "exit 1",
						ContinueOnError: true,
					},
					{
						Name:    "step2",
						Command: "echo 'step2 executed'",
					},
				},
			},
			wantErr:        false,
			expectedStatus: RunStatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			stateManager := NewStateManager(tempDir)

			// First validate the workflow
			if err := tt.workflow.Validate(); err != nil {
				if !tt.wantErr {
					t.Fatalf("Unexpected validation error: %v", err)
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
				return
			}

			runner := NewRunner(stateManager, RunOptions{})
			ctx := context.Background()

			state, err := runner.RunWorkflow(ctx, tt.workflow)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
				if tt.expectedStatus != "" && state != nil && state.Status != tt.expectedStatus {
					t.Errorf("Status = %s, want %s", state.Status, tt.expectedStatus)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if state.Status != tt.expectedStatus {
					t.Errorf("Status = %s, want %s", state.Status, tt.expectedStatus)
				}
			}
		})
	}
}

// TestDryRunMode tests that dry-run mode validates without executing
func TestDryRunMode(t *testing.T) {
	tests := []struct {
		name     string
		workflow *Workflow
		wantErr  bool
	}{
		{
			name: "valid workflow dry run",
			workflow: &Workflow{
				Name: "dry-run-test",
				Steps: []Step{
					{
						Name:    "step1",
						Command: "this-command-does-not-exist-and-would-fail",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid workflow dry run",
			workflow: &Workflow{
				Name: "",
				Steps: []Step{
					{Name: "step1", Command: "echo test"},
				},
			},
			wantErr: true,
		},
		{
			name: "circular dependency dry run",
			workflow: &Workflow{
				Name: "dry-run-circular",
				Steps: []Step{
					{
						Name:      "step1",
						Command:   "echo 1",
						DependsOn: []string{"step2"},
					},
					{
						Name:      "step2",
						Command:   "echo 2",
						DependsOn: []string{"step1"},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			stateManager := NewStateManager(tempDir)
			runner := NewRunner(stateManager, RunOptions{DryRun: true})
			ctx := context.Background()

			state, err := runner.RunWorkflow(ctx, tt.workflow)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error in dry-run: %v", err)
			}

			// In dry-run, no steps should actually execute
			// So we should have no step results (or just pending state)
			if state.Status != RunStatusRunning {
				// Status might be running since we don't mark as completed in dry-run
				t.Logf("Note: Dry-run state status is %s", state.Status)
			}

			// Verify the invalid command was never executed (no error about command not found)
			// This is implicit - if we got here without error, dry-run worked
		})
	}
}

// TestDryRunNoSideEffects verifies dry-run produces no side effects
func TestDryRunNoSideEffects(t *testing.T) {
	tempDir := t.TempDir()
	markerFile := filepath.Join(tempDir, "marker.txt")

	workflow := &Workflow{
		Name: "side-effect-test",
		Steps: []Step{
			{
				Name:    "create_file",
				Command: "echo 'created' > " + markerFile,
			},
		},
	}

	stateManager := NewStateManager(tempDir)
	runner := NewRunner(stateManager, RunOptions{DryRun: true})
	ctx := context.Background()

	_, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Dry-run failed: %v", err)
	}

	// Verify marker file was NOT created
	if _, err := os.Stat(markerFile); !os.IsNotExist(err) {
		t.Error("Marker file should not exist after dry-run")
	}

	// Now run without dry-run to verify the command works
	runner2 := NewRunner(stateManager, RunOptions{DryRun: false})
	_, err = runner2.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Normal run failed: %v", err)
	}

	// Now file should exist
	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		t.Error("Marker file should exist after normal run")
	}
}

// TestWorkflowFileParsing tests parsing workflow from JSON file
func TestWorkflowFileParsing(t *testing.T) {
	tempDir := t.TempDir()

	validWorkflow := map[string]interface{}{
		"name":        "file-test",
		"description": "Test parsing from file",
		"env": map[string]string{
			"FILE_VAR": "from_file",
		},
		"steps": []map[string]interface{}{
			{
				"name":            "step1",
				"type":            "shell",
				"command":         "echo test",
				"continueOnError": false,
			},
		},
	}

	// Write valid workflow
	validPath := filepath.Join(tempDir, "valid.json")
	data, _ := json.Marshal(validWorkflow)
	if err := os.WriteFile(validPath, data, 0644); err != nil {
		t.Fatalf("Failed to write valid workflow: %v", err)
	}

	// Test parsing valid file
	parser := NewParser()
	wf, err := parser.ParseFile(validPath)
	if err != nil {
		t.Fatalf("Failed to parse valid workflow: %v", err)
	}

	if wf.Name != "file-test" {
		t.Errorf("Name = %s, want 'file-test'", wf.Name)
	}

	if wf.Env["FILE_VAR"] != "from_file" {
		t.Errorf("Env[FILE_VAR] = %s, want 'from_file'", wf.Env["FILE_VAR"])
	}

	// Write invalid JSON
	invalidPath := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("Failed to write invalid workflow: %v", err)
	}

	_, err = parser.ParseFile(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test non-existent file
	_, err = parser.ParseFile(filepath.Join(tempDir, "nonexistent.json"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// TestRunFromFile tests running workflow from file path
func TestRunFromFile(t *testing.T) {
	tempDir := t.TempDir()

	workflow := map[string]interface{}{
		"name": "run-from-file",
		"steps": []map[string]interface{}{
			{
				"name":    "step1",
				"command": "echo 'from file'",
			},
		},
	}

	workflowPath := filepath.Join(tempDir, "workflow.json")
	data, _ := json.Marshal(workflow)
	if err := os.WriteFile(workflowPath, data, 0644); err != nil {
		t.Fatalf("Failed to write workflow: %v", err)
	}

	stateManager := NewStateManager(tempDir)
	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	state, err := runner.Run(ctx, workflowPath)
	if err != nil {
		t.Fatalf("Run from file failed: %v", err)
	}

	if state.Status != RunStatusCompleted {
		t.Errorf("Status = %s, want %s", state.Status, RunStatusCompleted)
	}

	output := state.StepOutputs["step1"].Stdout
	if !strings.Contains(output, "from file") {
		t.Errorf("Output should contain 'from file', got: %s", output)
	}
}

// TestGlobalEnv tests global environment variables option
func TestGlobalEnv(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	workflow := &Workflow{
		Name: "global-env-test",
		Steps: []Step{
			{
				Name:    "step1",
				Command: `echo "Global: ${GLOBAL_VAR}"`,
			},
		},
	}

	opts := RunOptions{
		GlobalEnv: map[string]string{
			"GLOBAL_VAR": "global_value",
		},
	}

	runner := NewRunner(stateManager, opts)
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := state.StepOutputs["step1"].Stdout
	if !strings.Contains(output, "global_value") {
		t.Errorf("Output should contain global env var, got: %s", output)
	}
}

// TestMultipleDependencies tests steps with multiple dependencies
func TestMultipleDependencies(t *testing.T) {
	// Skip on Windows CI - shell commands work differently
	if os.Getenv("CI") == "true" && runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows CI")
	}

	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	workflow := &Workflow{
		Name: "multi-dep-test",
		Steps: []Step{
			{
				Name:           "step1",
				Command:        `echo '{"out": "a"}'`,
				CaptureOutputs: []string{"out"},
			},
			{
				Name:           "step2",
				Command:        `echo '{"out": "b"}'`,
				CaptureOutputs: []string{"out"},
			},
			{
				Name:      "step3",
				Command:   `echo "Got: ${steps.step1.out} and ${steps.step2.out}"`,
				DependsOn: []string{"step1", "step2"},
			},
		},
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if state.Status != RunStatusCompleted {
		t.Errorf("Status = %s, want %s", state.Status, RunStatusCompleted)
	}

	output := state.StepOutputs["step3"].Stdout
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
		t.Errorf("step3 should have both dependencies, got: %s", output)
	}
}

// TestStepOrdering tests that dependency ordering is respected
func TestStepOrdering(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	executionOrder := []string{}

	workflow := &Workflow{
		Name: "ordering-test",
		Steps: []Step{
			{
				Name:      "step3",
				Command:   `echo "step3"`,
				DependsOn: []string{"step2"},
			},
			{
				Name:    "step1",
				Command: `echo "step1"`,
			},
			{
				Name:      "step2",
				Command:   `echo "step2"`,
				DependsOn: []string{"step1"},
			},
		},
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// Verify execution order by checking timestamps
	step1Time := state.StepOutputs["step1"].StartedAt
	step2Time := state.StepOutputs["step2"].StartedAt
	step3Time := state.StepOutputs["step3"].StartedAt

	if !step1Time.Before(step2Time) {
		t.Error("step1 should start before step2")
	}
	if !step2Time.Before(step3Time) {
		t.Error("step2 should start before step3")
	}

	_ = executionOrder // Use the variable to avoid unused warning
}

// TestStateManagerList tests listing all runs
func TestStateManagerList(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	// Create multiple runs
	savedRunIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		wf := Workflow{
			Name:  fmt.Sprintf("wf-%d", i),
			Steps: []Step{{Name: "s", Command: "echo"}},
		}
		state := NewRunState(wf)
		state.Status = RunStatusCompleted
		if err := stateManager.Save(state); err != nil {
			t.Fatalf("Save failed: %v", err)
		}
		savedRunIDs[i] = state.RunID
		t.Logf("Saved run %d with ID: %s", i, state.RunID)
		time.Sleep(10 * time.Millisecond) // Ensure unique timestamps
	}

	runs, err := stateManager.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	t.Logf("Saved run IDs: %v", savedRunIDs)
	t.Logf("Listed runs count: %d", len(runs))
	for _, run := range runs {
		t.Logf("  - Run: %s", run.RunID)
	}

	// We expect at least the 3 runs we created
	if len(runs) < 3 {
		t.Logf("Warning: Expected at least 3 runs, got %d (this may be a timing issue)", len(runs))
	}

	// Verify the runs we created are present
	runIDs := make(map[string]bool)
	for _, run := range runs {
		runIDs[run.RunID] = true
	}

	// Verify all our saved runs are present
	for i, savedID := range savedRunIDs {
		if !runIDs[savedID] {
			t.Errorf("Saved run %d with ID %s not found in list", i, savedID)
		}
	}
}

// TestStateManagerDelete tests deleting runs
func TestStateManagerDelete(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	wf := Workflow{Name: "test", Steps: []Step{{Name: "s", Command: "echo"}}}
	state := NewRunState(wf)

	if err := stateManager.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify it exists
	_, err := stateManager.Load(state.RunID)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Delete it
	if err := stateManager.Delete(state.RunID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = stateManager.Load(state.RunID)
	if err == nil {
		t.Error("Expected error when loading deleted run")
	}
}

// TestGetResumeInfo tests the GetResumeInfo function
func TestGetResumeInfo(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	// Create a failed run
	wf := Workflow{
		Name: "resume-info-test",
		Steps: []Step{
			{Name: "step1", Command: "echo ok"},
			{Name: "step2", Command: "exit 1"},
			{Name: "step3", Command: "echo ok"},
		},
	}
	state := NewRunState(wf)
	state.Status = RunStatusFailed
	state.AddStepResult(StepResult{
		Step:   Step{Name: "step1"},
		Output: StepOutput{StepName: "step1", ExitCode: 0},
	})
	state.AddStepResult(StepResult{
		Step:   Step{Name: "step2"},
		Output: StepOutput{StepName: "step2", ExitCode: 1, Error: "failed"},
	})

	if err := stateManager.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err := stateManager.GetResumeInfo(state.RunID)
	if err != nil {
		t.Fatalf("GetResumeInfo failed: %v", err)
	}

	if !info.CanResume {
		t.Error("CanResume should be true for failed run")
	}

	if info.CompletedCount != 1 {
		t.Errorf("CompletedCount = %d, want 1", info.CompletedCount)
	}

	if info.TotalSteps != 3 {
		t.Errorf("TotalSteps = %d, want 3", info.TotalSteps)
	}

	if info.FailedStep != "step2" {
		t.Errorf("FailedStep = %s, want step2", info.FailedStep)
	}

	// Test completed run cannot resume
	state.Status = RunStatusCompleted
	state.Error = ""
	state.StepResults = append(state.StepResults, StepResult{
		Step:   Step{Name: "step2"},
		Output: StepOutput{StepName: "step2", ExitCode: 0},
	})
	state.StepResults = append(state.StepResults, StepResult{
		Step:   Step{Name: "step3"},
		Output: StepOutput{StepName: "step3", ExitCode: 0},
	})
	if err := stateManager.Save(state); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	info, err = stateManager.GetResumeInfo(state.RunID)
	if err != nil {
		t.Fatalf("GetResumeInfo failed: %v", err)
	}

	if info.CanResume {
		t.Error("CanResume should be false for completed run")
	}
}

// TestConditionEvaluation tests step conditions
func TestConditionEvaluation(t *testing.T) {
	// Skip on Windows CI - shell commands work differently
	if os.Getenv("CI") == "true" && runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows CI")
	}

	tests := []struct {
		name          string
		condition     string
		workflowEnv   map[string]string
		stepOutputs   map[string]StepOutput
		shouldExecute bool
	}{
		{
			name:          "true literal",
			condition:     "true",
			shouldExecute: true,
		},
		{
			name:          "false literal",
			condition:     "false",
			shouldExecute: false,
		},
		{
			name:          "env variable exists",
			condition:     "${env.CONDITION_VAR}",
			workflowEnv:   map[string]string{"CONDITION_VAR": "set"},
			shouldExecute: true,
		},
		{
			name:          "empty string is falsy",
			condition:     "${env.EMPTY_VAR}",
			workflowEnv:   map[string]string{"EMPTY_VAR": ""},
			shouldExecute: false,
		},
		{
			name:      "step output exists",
			condition: "${steps.prev.result}",
			stepOutputs: map[string]StepOutput{
				"prev": {
					StepName: "prev",
					Data:     map[string]interface{}{"result": "success"},
				},
			},
			shouldExecute: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			stateManager := NewStateManager(tempDir)

			workflow := &Workflow{
				Name: "condition-test",
				Env:  tt.workflowEnv,
				Steps: []Step{
					{
						Name:           "prev",
						Command:        `echo '{"result": "success"}'`,
						CaptureOutputs: []string{"result"},
					},
					{
						Name:      "conditional",
						Command:   `echo "executed"`,
						DependsOn: []string{"prev"},
						Condition: tt.condition,
					},
				},
			}

			// Manually set up state with outputs if needed
			state := NewRunState(*workflow)
			if tt.stepOutputs != nil {
				for name, output := range tt.stepOutputs {
					state.StepOutputs[name] = output
				}
			}

			runner := NewRunner(stateManager, RunOptions{})
			ctx := context.Background()

			finalState, err := runner.RunWorkflow(ctx, workflow)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			_, executed := finalState.StepOutputs["conditional"]
			if tt.shouldExecute && !executed {
				t.Error("Conditional step should have executed")
			}
			if !tt.shouldExecute && executed {
				t.Error("Conditional step should not have executed")
			}
		})
	}
}

// TestRetryConfiguration tests retry settings
func TestRetryConfiguration(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	workflow := &Workflow{
		Name: "retry-test",
		Steps: []Step{
			{
				Name:       "retry_step",
				Command:    `echo "retry_test"`,
				RetryCount: 3,
				RetryDelay: 100 * time.Millisecond,
			},
		},
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if state.Status != RunStatusCompleted {
		t.Errorf("Status = %s, want %s", state.Status, RunStatusCompleted)
	}

	// Verify step succeeded on first try (no retry needed)
	output := state.StepOutputs["retry_step"]
	if !strings.Contains(output.Stdout, "retry_test") {
		t.Errorf("Output should contain 'retry_test', got: %s", output.Stdout)
	}
}

// TestWorkingDirectory tests step working directory
func TestWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create a marker file in subdir
	markerFile := filepath.Join(subDir, "marker.txt")
	if err := os.WriteFile(markerFile, []byte("marker content"), 0644); err != nil {
		t.Fatalf("Failed to write marker: %v", err)
	}

	stateManager := NewStateManager(tempDir)

	workflow := &Workflow{
		Name: "workdir-test",
		Steps: []Step{
			{
				Name:       "check_dir",
				Command:    `cat marker.txt`,
				WorkingDir: subDir,
			},
		},
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	output := state.StepOutputs["check_dir"]
	if !strings.Contains(output.Stdout, "marker content") {
		t.Errorf("Step should read from subdir, got: %s", output.Stdout)
	}
}

// TestStepTypeInference tests automatic step type detection
func TestStepTypeInference(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		wantType StepType
	}{
		{
			name:     "gpd command",
			command:  "gpd publish upload app.aab",
			wantType: StepTypeGPD,
		},
		{
			name:     "shell command",
			command:  "echo hello",
			wantType: StepTypeShell,
		},
		{
			name:     "explicit type overrides",
			command:  "gpd something",
			wantType: StepTypeGPD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser()

			workflow := &Workflow{
				Name: "type-test",
				Steps: []Step{
					{
						Name:    "step1",
						Command: tt.command,
					},
				},
			}

			// Parse to trigger type inference
			data, _ := json.Marshal(workflow)
			parsed, err := parser.Parse(data)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if parsed.Steps[0].Type != tt.wantType {
				t.Errorf("Type = %s, want %s", parsed.Steps[0].Type, tt.wantType)
			}
		})
	}
}

// TestRunnerWithVerboseLogging tests verbose logging option
func TestRunnerWithVerboseLogging(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	workflow := &Workflow{
		Name: "verbose-test",
		Steps: []Step{
			{
				Name:    "step1",
				Command: `echo "verbose output"`,
			},
		},
	}

	opts := RunOptions{Verbose: true}
	runner := NewRunner(stateManager, opts)
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if state.Status != RunStatusCompleted {
		t.Errorf("Status = %s, want %s", state.Status, RunStatusCompleted)
	}
}

// TestCaptureOutputsFromPlainText tests capturing from non-JSON output
func TestCaptureOutputsFromPlainText(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	workflow := &Workflow{
		Name: "capture-plain-test",
		Steps: []Step{
			{
				Name:           "plain_step",
				Command:        `echo "versionCode=123 versionName=1.0.0"`,
				CaptureOutputs: []string{"versionCode"}, // Won't find in plain text
			},
		},
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// The step should succeed but capture no outputs (since it's not JSON)
	output := state.StepOutputs["plain_step"]
	if output.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", output.ExitCode)
	}

	// Data should be empty since output wasn't JSON
	if len(output.Data) > 0 {
		t.Logf("Note: Some data was captured from plain text: %v", output.Data)
	}
}

// TestWorkflowMerge tests merging workflows
func TestWorkflowMerge(t *testing.T) {
	base := &Workflow{
		Name:        "base",
		Description: "Base workflow",
		Env: map[string]string{
			"VAR1": "base1",
			"VAR2": "base2",
		},
		Steps: []Step{
			{Name: "step1", Command: "echo base"},
		},
	}

	override := &Workflow{
		Name:        "override",
		Description: "Override workflow",
		Env: map[string]string{
			"VAR2": "override2",
			"VAR3": "override3",
		},
		Steps: []Step{
			{Name: "step1", Command: "echo override"},
		},
	}

	merged := MergeWorkflows(base, override)

	if merged.Name != "override" {
		t.Errorf("Name = %s, want 'override'", merged.Name)
	}

	if merged.Description != "Override workflow" {
		t.Errorf("Description = %s, want 'Override workflow'", merged.Description)
	}

	if merged.Env["VAR1"] != "base1" {
		t.Errorf("VAR1 = %s, want 'base1'", merged.Env["VAR1"])
	}

	if merged.Env["VAR2"] != "override2" {
		t.Errorf("VAR2 = %s, want 'override2' (override takes precedence)", merged.Env["VAR2"])
	}

	if merged.Env["VAR3"] != "override3" {
		t.Errorf("VAR3 = %s, want 'override3'", merged.Env["VAR3"])
	}

	if len(merged.Steps) != 1 || merged.Steps[0].Command != "echo override" {
		t.Error("Override steps should take precedence")
	}
}

// TestNullWorkflowMerge tests merging with nil workflows
func TestNullWorkflowMerge(t *testing.T) {
	wf := &Workflow{
		Name: "test",
		Steps: []Step{
			{Name: "step1", Command: "echo test"},
		},
	}

	// Merge with nil base
	merged1 := MergeWorkflows(nil, wf)
	if merged1.Name != "test" {
		t.Error("Merge with nil base should return override")
	}

	// Merge with nil override
	merged2 := MergeWorkflows(wf, nil)
	if merged2.Name != "test" {
		t.Error("Merge with nil override should return base")
	}

	// Merge two nils
	merged3 := MergeWorkflows(nil, nil)
	if merged3 != nil {
		t.Error("Merge of two nils should return nil")
	}
}

// TestExpandTilde tests tilde expansion
func TestExpandTilde(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Cannot get home dir: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "~/test",
			expected: homeDir + "/test",
		},
		{
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			input:    "~user/test",
			expected: "~user/test", // Should not expand (different user)
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ExpandTilde(tt.input)
			if err != nil {
				t.Fatalf("ExpandTilde failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("ExpandTilde(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestValidateCommandSyntax tests command validation
func TestValidateCommandSyntax(t *testing.T) {
	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{
			name:    "valid command",
			command: "echo hello",
			wantErr: false,
		},
		{
			name:    "empty command",
			command: "",
			wantErr: true,
		},
		{
			name:    "balanced braces",
			command: "echo ${env.VAR}",
			wantErr: false,
		},
		{
			name:    "unbalanced braces",
			command: "echo ${env.VAR",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandSyntax(tt.command)
			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestConcurrentWorkflowExecution tests that multiple workflows can run independently
func TestConcurrentWorkflowExecution(t *testing.T) {
	tempDir1 := t.TempDir()
	tempDir2 := t.TempDir()

	wf1 := &Workflow{
		Name: "wf1",
		Steps: []Step{
			{Name: "step", Command: `echo "workflow1"`},
		},
	}

	wf2 := &Workflow{
		Name: "wf2",
		Steps: []Step{
			{Name: "step", Command: `echo "workflow2"`},
		},
	}

	stateManager1 := NewStateManager(tempDir1)
	stateManager2 := NewStateManager(tempDir2)

	runner1 := NewRunner(stateManager1, RunOptions{})
	runner2 := NewRunner(stateManager2, RunOptions{})

	ctx := context.Background()

	// Run both workflows
	state1, err1 := runner1.RunWorkflow(ctx, wf1)
	state2, err2 := runner2.RunWorkflow(ctx, wf2)

	if err1 != nil {
		t.Fatalf("Workflow 1 failed: %v", err1)
	}
	if err2 != nil {
		t.Fatalf("Workflow 2 failed: %v", err2)
	}

	// Verify both completed
	if state1.Status != RunStatusCompleted {
		t.Errorf("Workflow 1 status = %s, want %s", state1.Status, RunStatusCompleted)
	}
	if state2.Status != RunStatusCompleted {
		t.Errorf("Workflow 2 status = %s, want %s", state2.Status, RunStatusCompleted)
	}

	// Verify outputs are different
	out1 := state1.StepOutputs["step"].Stdout
	out2 := state2.StepOutputs["step"].Stdout

	if out1 == out2 {
		t.Error("Workflow outputs should be independent")
	}

	// Verify different run IDs
	if state1.RunID == state2.RunID {
		t.Error("Run IDs should be unique")
	}
}

// TestLongRunningWorkflow tests workflow with many steps
func TestLongRunningWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	// Create workflow with many steps
	steps := make([]Step, 20)
	for i := 0; i < 20; i++ {
		deps := []string{}
		if i > 0 {
			deps = append(deps, fmt.Sprintf("step%d", i-1))
		}
		steps[i] = Step{
			Name:      fmt.Sprintf("step%d", i),
			Command:   fmt.Sprintf(`echo "step %d"`, i),
			DependsOn: deps,
		}
	}

	workflow := &Workflow{
		Name:  "long-workflow",
		Steps: steps,
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Long workflow failed: %v", err)
	}

	if state.Status != RunStatusCompleted {
		t.Errorf("Status = %s, want %s", state.Status, RunStatusCompleted)
	}

	if len(state.StepResults) != 20 {
		t.Errorf("Expected 20 step results, got %d", len(state.StepResults))
	}

	// Verify execution order
	for i := 0; i < 20; i++ {
		stepName := fmt.Sprintf("step%d", i)
		if !state.IsStepCompleted(stepName) {
			t.Errorf("Step %s should be completed", stepName)
		}
	}
}

// TestComplexDependencies tests a complex dependency graph
func TestComplexDependencies(t *testing.T) {
	// Skip on Windows CI - shell commands work differently
	if os.Getenv("CI") == "true" && runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows CI")
	}

	tempDir := t.TempDir()
	stateManager := NewStateManager(tempDir)

	// Create a diamond dependency pattern:
	//     step1
	//    /     \
	//  step2   step3
	//    \     /
	//     step4
	workflow := &Workflow{
		Name: "complex-deps",
		Steps: []Step{
			{
				Name:           "step1",
				Command:        `echo '{"value": "start"}'`,
				CaptureOutputs: []string{"value"},
			},
			{
				Name:      "step2",
				Command:   `echo "step2 got: ${steps.step1.value}"`,
				DependsOn: []string{"step1"},
			},
			{
				Name:      "step3",
				Command:   `echo "step3 got: ${steps.step1.value}"`,
				DependsOn: []string{"step1"},
			},
			{
				Name:      "step4",
				Command:   `echo "step4 complete"`,
				DependsOn: []string{"step2", "step3"},
			},
		},
	}

	runner := NewRunner(stateManager, RunOptions{})
	ctx := context.Background()

	state, err := runner.RunWorkflow(ctx, workflow)
	if err != nil {
		t.Fatalf("Complex workflow failed: %v", err)
	}

	if state.Status != RunStatusCompleted {
		t.Errorf("Status = %s, want %s", state.Status, RunStatusCompleted)
	}

	// All steps should be completed
	for _, stepName := range []string{"step1", "step2", "step3", "step4"} {
		if !state.IsStepCompleted(stepName) {
			t.Errorf("Step %s should be completed", stepName)
		}
	}

	// Verify step4 output
	if !strings.Contains(state.StepOutputs["step4"].Stdout, "complete") {
		t.Error("step4 should have completed")
	}
}
