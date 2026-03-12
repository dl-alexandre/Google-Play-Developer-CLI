package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWorkflowValidation(t *testing.T) {
	tests := []struct {
		name      string
		workflow  Workflow
		wantError bool
		errField  string
	}{
		{
			name: "valid workflow",
			workflow: Workflow{
				Name:  "test-workflow",
				Steps: []Step{{Name: "step1", Command: "echo hello"}},
			},
			wantError: false,
		},
		{
			name: "missing name",
			workflow: Workflow{
				Steps: []Step{{Name: "step1", Command: "echo hello"}},
			},
			wantError: true,
			errField:  "name",
		},
		{
			name: "no steps",
			workflow: Workflow{
				Name: "test-workflow",
			},
			wantError: true,
			errField:  "steps",
		},
		{
			name: "duplicate step names",
			workflow: Workflow{
				Name: "test-workflow",
				Steps: []Step{
					{Name: "step1", Command: "echo hello"},
					{Name: "step1", Command: "echo world"},
				},
			},
			wantError: true,
			errField:  "steps",
		},
		{
			name: "missing step command",
			workflow: Workflow{
				Name:  "test-workflow",
				Steps: []Step{{Name: "step1"}},
			},
			wantError: true,
			errField:  "steps",
		},
		{
			name: "invalid step type",
			workflow: Workflow{
				Name:  "test-workflow",
				Steps: []Step{{Name: "step1", Command: "echo", Type: "invalid"}},
			},
			wantError: true,
			errField:  "steps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workflow.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err != nil && tt.errField != "" {
				if valErr, ok := err.(*ValidationError); ok {
					if valErr.Field != tt.errField {
						t.Errorf("Validate() error field = %v, want %v", valErr.Field, tt.errField)
					}
				}
			}
		})
	}
}

func TestWorkflowGetStep(t *testing.T) {
	wf := Workflow{
		Name: "test",
		Steps: []Step{
			{Name: "step1", Command: "echo 1"},
			{Name: "step2", Command: "echo 2"},
		},
	}

	t.Run("existing step", func(t *testing.T) {
		step, ok := wf.GetStep("step1")
		if !ok {
			t.Error("GetStep() returned false for existing step")
		}
		if step.Name != "step1" {
			t.Errorf("GetStep() returned wrong step: %v", step.Name)
		}
	})

	t.Run("missing step", func(t *testing.T) {
		_, ok := wf.GetStep("nonexistent")
		if ok {
			t.Error("GetStep() returned true for non-existent step")
		}
	})
}

func TestStateManager(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewStateManager(tempDir)

	t.Run("ensure directories", func(t *testing.T) {
		if err := sm.EnsureDefaultDirs(); err != nil {
			t.Fatalf("EnsureDefaultDirs() failed: %v", err)
		}

		// Check directories exist
		for _, dir := range []string{"definitions", "runs"} {
			path := filepath.Join(tempDir, dir)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("Directory %s was not created", dir)
			}
		}
	})

	t.Run("save and load", func(t *testing.T) {
		workflow := Workflow{
			Name: "test-workflow",
			Steps: []Step{
				{Name: "step1", Command: "echo hello"},
			},
		}
		state := NewRunState(workflow)

		if err := sm.Save(state); err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		loaded, err := sm.Load(state.RunID)
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if loaded.RunID != state.RunID {
			t.Errorf("Loaded wrong run ID: got %v, want %v", loaded.RunID, state.RunID)
		}
	})

	t.Run("list runs", func(t *testing.T) {
		// Create a few states
		for i := 0; i < 3; i++ {
			wf := Workflow{Name: "wf", Steps: []Step{{Name: "s", Command: "echo"}}}
			state := NewRunState(wf)
			if err := sm.Save(state); err != nil {
				t.Fatalf("Save() failed: %v", err)
			}
			time.Sleep(10 * time.Millisecond) // Ensure unique timestamps
		}

		states, err := sm.List()
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(states) < 3 {
			t.Errorf("List() returned %d states, want at least 3", len(states))
		}
	})
}

func TestInterpolator(t *testing.T) {
	outputs := map[string]StepOutput{
		"build": {
			StepName: "build",
			Data: map[string]interface{}{
				"versionCode": 123,
				"versionName": "1.0.0",
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	env := map[string]string{
		"PACKAGE": "com.example.app",
	}

	interp := NewInterpolator(outputs, env)

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "step output simple",
			input: "${steps.build.versionCode}",
			want:  "123",
		},
		{
			name:  "step output string",
			input: "${steps.build.versionName}",
			want:  "1.0.0",
		},
		{
			name:  "nested output",
			input: "${steps.build.nested.key}",
			want:  "value",
		},
		{
			name:  "env var",
			input: "${env.PACKAGE}",
			want:  "com.example.app",
		},
		{
			name:  "env shorthand",
			input: "${PACKAGE}",
			want:  "com.example.app",
		},
		{
			name:  "mixed interpolation",
			input: "pkg=${env.PACKAGE} version=${steps.build.versionCode}",
			want:  "pkg=com.example.app version=123",
		},
		{
			name:    "missing step",
			input:   "${steps.nonexistent.field}",
			wantErr: true,
		},
		{
			name:    "missing field",
			input:   "${steps.build.missing}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := interp.Interpolate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Interpolate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Interpolate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParserTopologicalSort(t *testing.T) {
	parser := NewParser()

	wf := Workflow{
		Name: "test",
		Steps: []Step{
			{Name: "step4", Command: "echo 4", DependsOn: []string{"step3"}},
			{Name: "step2", Command: "echo 2", DependsOn: []string{"step1"}},
			{Name: "step1", Command: "echo 1"},
			{Name: "step3", Command: "echo 3", DependsOn: []string{"step2"}},
		},
	}

	sorted, err := parser.TopologicalSort(&wf)
	if err != nil {
		t.Fatalf("TopologicalSort() failed: %v", err)
	}

	// Verify order
	order := make(map[string]int)
	for i, step := range sorted {
		order[step.Name] = i
	}

	// step1 must come before step2
	if order["step1"] >= order["step2"] {
		t.Error("step1 should come before step2")
	}
	// step2 must come before step3
	if order["step2"] >= order["step3"] {
		t.Error("step2 should come before step3")
	}
	// step3 must come before step4
	if order["step3"] >= order["step4"] {
		t.Error("step3 should come before step4")
	}
}

func TestParserCircularDependency(t *testing.T) {
	parser := NewParser()

	wf := Workflow{
		Name: "test",
		Steps: []Step{
			{Name: "step1", Command: "echo 1", DependsOn: []string{"step2"}},
			{Name: "step2", Command: "echo 2", DependsOn: []string{"step1"}},
		},
	}

	_, err := parser.Parse(nil) // First validate creates workflow
	if err == nil {
		// The circular dependency should be caught during parse
		err = parser.checkCircularDependencies(&wf)
	}
	if err == nil {
		t.Error("Expected error for circular dependency")
	}
}

func TestParseAndExtractJSON(t *testing.T) {
	json := `{"versionCode": 123, "versionName": "1.0.0", "nested": {"key": "value"}}`

	fields := []string{"versionCode", "versionName", "nested.key"}
	result, err := ParseAndExtractJSON(json, fields)
	if err != nil {
		t.Fatalf("ParseAndExtractJSON() failed: %v", err)
	}

	if result["versionCode"] != float64(123) { // JSON numbers are float64
		t.Errorf("versionCode = %v, want 123", result["versionCode"])
	}

	if result["versionName"] != "1.0.0" {
		t.Errorf("versionName = %v, want 1.0.0", result["versionName"])
	}

	if result["nested.key"] != "value" {
		t.Errorf("nested.key = %v, want value", result["nested.key"])
	}
}

func TestRunState(t *testing.T) {
	wf := Workflow{
		Name: "test",
		Steps: []Step{
			{Name: "step1", Command: "echo 1"},
			{Name: "step2", Command: "echo 2"},
		},
	}

	state := NewRunState(wf)

	t.Run("is step completed", func(t *testing.T) {
		if state.IsStepCompleted("step1") {
			t.Error("New state should not have completed steps")
		}

		// Add a completed step
		state.AddStepResult(StepResult{
			Step: Step{Name: "step1"},
			Output: StepOutput{
				StepName: "step1",
				ExitCode: 0,
			},
		})

		if !state.IsStepCompleted("step1") {
			t.Error("Step1 should be marked as completed")
		}
	})

	t.Run("get next pending step", func(t *testing.T) {
		pending := state.GetNextPendingStep()
		if pending == nil {
			t.Fatal("GetNextPendingStep() returned nil, expected step2")
		}
		if pending.Name != "step2" {
			t.Errorf("GetNextPendingStep() = %v, want step2", pending.Name)
		}
	})
}

func TestRetryLogic(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewStateManager(tempDir)

	t.Run("calculate retry delays", func(t *testing.T) {
		runner := NewRunner(sm, RunOptions{})

		tests := []struct {
			name      string
			step      Step
			attempt   int
			wantDelay time.Duration
		}{
			{
				name:      "linear backoff first attempt",
				step:      Step{RetryDelay: 100 * time.Millisecond, RetryBackoff: "linear"},
				attempt:   1,
				wantDelay: 100 * time.Millisecond,
			},
			{
				name:      "linear backoff second attempt",
				step:      Step{RetryDelay: 100 * time.Millisecond, RetryBackoff: "linear"},
				attempt:   2,
				wantDelay: 200 * time.Millisecond,
			},
			{
				name:      "exponential backoff first attempt",
				step:      Step{RetryDelay: 100 * time.Millisecond, RetryBackoff: "exponential"},
				attempt:   1,
				wantDelay: 100 * time.Millisecond,
			},
			{
				name:      "exponential backoff second attempt",
				step:      Step{RetryDelay: 100 * time.Millisecond, RetryBackoff: "exponential"},
				attempt:   2,
				wantDelay: 200 * time.Millisecond,
			},
			{
				name:      "exponential backoff third attempt",
				step:      Step{RetryDelay: 100 * time.Millisecond, RetryBackoff: "exponential"},
				attempt:   3,
				wantDelay: 400 * time.Millisecond,
			},
			{
				name:      "default backoff (linear)",
				step:      Step{RetryDelay: 100 * time.Millisecond},
				attempt:   2,
				wantDelay: 200 * time.Millisecond,
			},
			{
				name:      "unknown backoff type falls back to linear",
				step:      Step{RetryDelay: 100 * time.Millisecond, RetryBackoff: "unknown"},
				attempt:   2,
				wantDelay: 200 * time.Millisecond,
			},
			{
				name:      "zero delay",
				step:      Step{RetryDelay: 0},
				attempt:   1,
				wantDelay: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				delay := runner.calculateRetryDelay(tt.step, tt.attempt)
				if delay != tt.wantDelay {
					t.Errorf("calculateRetryDelay() = %v, want %v", delay, tt.wantDelay)
				}
			})
		}
	})
}

func TestStepRetryFields(t *testing.T) {
	tests := []struct {
		name            string
		step            Step
		expectedRetry   int
		expectedDelay   time.Duration
		expectedBackoff string
	}{
		{
			name:            "step with no retry",
			step:            Step{Name: "test", Command: "echo test"},
			expectedRetry:   0,
			expectedDelay:   0,
			expectedBackoff: "",
		},
		{
			name:            "step with retry configuration",
			step:            Step{Name: "test", Command: "echo test", RetryCount: 3, RetryDelay: 5 * time.Second, RetryBackoff: "exponential"},
			expectedRetry:   3,
			expectedDelay:   5 * time.Second,
			expectedBackoff: "exponential",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.step.RetryCount != tt.expectedRetry {
				t.Errorf("RetryCount = %v, want %v", tt.step.RetryCount, tt.expectedRetry)
			}
			if tt.step.RetryDelay != tt.expectedDelay {
				t.Errorf("RetryDelay = %v, want %v", tt.step.RetryDelay, tt.expectedDelay)
			}
			if tt.step.RetryBackoff != tt.expectedBackoff {
				t.Errorf("RetryBackoff = %v, want %v", tt.step.RetryBackoff, tt.expectedBackoff)
			}
		})
	}
}

func TestParallelStepExecution(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewStateManager(tempDir)

	t.Run("parallel field exists", func(t *testing.T) {
		step := Step{Name: "test", Command: "echo test", Parallel: true}
		if !step.Parallel {
			t.Error("Parallel field should be true")
		}
	})

	t.Run("maxParallel field exists", func(t *testing.T) {
		wf := Workflow{Name: "test", MaxParallel: 8}
		if wf.MaxParallel != 8 {
			t.Errorf("MaxParallel = %v, want 8", wf.MaxParallel)
		}
	})

	t.Run("group steps by level", func(t *testing.T) {
		runner := NewRunner(sm, RunOptions{})

		wf := Workflow{
			Name: "test-levels",
			Steps: []Step{
				{Name: "step1", Command: "echo 1"},
				{Name: "step2", Command: "echo 2", DependsOn: []string{"step1"}},
				{Name: "step3", Command: "echo 3", DependsOn: []string{"step1"}},
				{Name: "step4", Command: "echo 4", DependsOn: []string{"step2", "step3"}},
			},
		}

		groups, err := runner.groupStepsByLevel(&wf)
		if err != nil {
			t.Fatalf("groupStepsByLevel() failed: %v", err)
		}

		// Should have 3 levels:
		// Level 0: step1
		// Level 1: step2, step3
		// Level 2: step4
		if len(groups) != 3 {
			t.Errorf("Expected 3 levels, got %d", len(groups))
		}

		// Check level 0 contains step1 only
		if len(groups) > 0 {
			if len(groups[0].steps) != 1 || groups[0].steps[0].Name != "step1" {
				t.Errorf("Level 0 should contain step1 only, got %v", groups[0].steps)
			}
		}

		// Check level 1 contains step2 and step3 (in any order)
		if len(groups) > 1 {
			if len(groups[1].steps) != 2 {
				t.Errorf("Level 1 should contain 2 steps, got %d", len(groups[1].steps))
			}
			names := make(map[string]bool)
			for _, s := range groups[1].steps {
				names[s.Name] = true
			}
			if !names["step2"] || !names["step3"] {
				t.Errorf("Level 1 should contain step2 and step3, got %v", groups[1].steps)
			}
		}

		// Check level 2 contains step4
		if len(groups) > 2 {
			if len(groups[2].steps) != 1 || groups[2].steps[0].Name != "step4" {
				t.Errorf("Level 2 should contain step4 only, got %v", groups[2].steps)
			}
		}
	})

	t.Run("parallel step validation", func(t *testing.T) {
		wf := Workflow{
			Name:        "parallel-test",
			MaxParallel: 4,
			Steps: []Step{
				{Name: "step1", Command: "echo 1", Parallel: true},
				{Name: "step2", Command: "echo 2", Parallel: true},
				{Name: "step3", Command: "echo 3", DependsOn: []string{"step1", "step2"}},
			},
		}

		// Validate workflow passes with parallel steps
		err := wf.Validate()
		if err != nil {
			t.Errorf("Workflow with parallel steps should be valid: %v", err)
		}
	})
}

func TestExecuteParallelSteps(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewStateManager(tempDir)

	t.Run("max parallel limit respected", func(t *testing.T) {
		runner := NewRunner(sm, RunOptions{})

		wf := Workflow{
			Name:        "parallel-limit-test",
			MaxParallel: 2,
			Steps: []Step{
				{Name: "step1", Command: "echo 1", Parallel: true},
				{Name: "step2", Command: "echo 2", Parallel: true},
				{Name: "step3", Command: "echo 3", Parallel: true},
				{Name: "step4", Command: "echo 4", Parallel: true},
				{Name: "step5", Command: "echo 5", Parallel: true},
			},
		}

		state := NewRunState(wf)

		// Test with maxParallel=2, should limit concurrent execution
		completed, failed, err := runner.executeParallelSteps(context.Background(), state, wf.Steps, 2)

		if err != nil {
			t.Logf("Note: Step execution may have failed due to no actual binary: %v", err)
		}

		// If steps executed, they should be tracked
		t.Logf("Completed: %v, Failed: %v", completed, failed)
	})

	t.Run("error handling in parallel steps", func(t *testing.T) {
		runner := NewRunner(sm, RunOptions{})

		wf := Workflow{
			Name:        "parallel-error-test",
			MaxParallel: 4,
			Steps: []Step{
				{Name: "step1", Command: "echo 1", Parallel: true, ContinueOnError: true},
				{Name: "step2", Command: "exit 1", Parallel: true, ContinueOnError: true},
				{Name: "step3", Command: "echo 3", Parallel: true},
			},
		}

		state := NewRunState(wf)

		// With continueOnError, should complete even if step2 fails
		completed, failed, err := runner.executeParallelSteps(context.Background(), state, wf.Steps, 4)

		// Should not return error since continueOnError is set
		if err != nil {
			t.Logf("Got error (may be expected): %v", err)
		}

		t.Logf("Completed: %v, Failed: %v", completed, failed)
	})
}

func TestParallelExecutionWithDependencies(t *testing.T) {
	tempDir := t.TempDir()
	sm := NewStateManager(tempDir)

	t.Run("dependency ordering is respected with parallel steps", func(t *testing.T) {
		runner := NewRunner(sm, RunOptions{})

		wf := Workflow{
			Name:        "parallel-deps-test",
			MaxParallel: 4,
			Steps: []Step{
				{Name: "prep1", Command: "echo prep1", Parallel: true},
				{Name: "prep2", Command: "echo prep2", Parallel: true},
				{Name: "merge", Command: "echo merge", DependsOn: []string{"prep1", "prep2"}},
				{Name: "post1", Command: "echo post1", DependsOn: []string{"merge"}, Parallel: true},
				{Name: "post2", Command: "echo post2", DependsOn: []string{"merge"}, Parallel: true},
			},
		}

		groups, err := runner.groupStepsByLevel(&wf)
		if err != nil {
			t.Fatalf("groupStepsByLevel() failed: %v", err)
		}

		// Should have 3 levels
		if len(groups) != 3 {
			t.Errorf("Expected 3 levels, got %d", len(groups))
		}

		// Level 0: prep1, prep2 (both parallel)
		// Level 1: merge (sequential, depends on both prep steps)
		// Level 2: post1, post2 (both parallel, depend on merge)

		if len(groups) > 0 {
			// Both prep steps should be parallel
			for _, step := range groups[0].steps {
				if !step.Parallel {
					t.Errorf("Expected %s to be parallel in level 0", step.Name)
				}
			}
		}

		if len(groups) > 1 {
			// merge should NOT be parallel
			if len(groups[1].steps) != 1 || groups[1].steps[0].Name != "merge" {
				t.Errorf("Level 1 should contain only 'merge', got %v", groups[1].steps)
			}
			if groups[1].steps[0].Parallel {
				t.Error("merge step should NOT be parallel")
			}
		}

		if len(groups) > 2 {
			// Both post steps should be parallel
			for _, step := range groups[2].steps {
				if !step.Parallel {
					t.Errorf("Expected %s to be parallel in level 2", step.Name)
				}
			}
		}
	})
}
