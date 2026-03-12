package workflow

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// RunOptions configures workflow execution.
type RunOptions struct {
	DryRun      bool
	ResumeRunID string
	Force       bool
	Verbose     bool
	GlobalEnv   map[string]string
	WorkingDir  string
	Watch       bool
	WatchFormat WatchFormat
}

type Runner struct {
	stateManager  *StateManager
	parser        *Parser
	options       RunOptions
	logger        Logger
	watcher       *Watcher
	workflowStart time.Time
}

type Logger interface {
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type DefaultLogger struct{}

func (l *DefaultLogger) Info(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[workflow] "+msg+"\n", args...)
}

func (l *DefaultLogger) Debug(msg string, args ...interface{}) {}

func (l *DefaultLogger) Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[workflow] ERROR: "+msg+"\n", args...)
}

func NewRunner(stateManager *StateManager, options RunOptions) *Runner {
	var logger Logger = &DefaultLogger{}
	if options.Verbose {
		logger = &VerboseLogger{}
	}
	runner := &Runner{
		stateManager: stateManager,
		parser:       NewParser(),
		options:      options,
		logger:       logger,
	}
	if options.Watch {
		watcherOpts := DefaultWatcherOptions()
		if options.WatchFormat != "" {
			watcherOpts.Format = options.WatchFormat
		}
		runner.watcher = NewWatcher(watcherOpts)
		runner.watcher.Start()
	}
	return runner
}

type VerboseLogger struct{}

func (l *VerboseLogger) Info(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[workflow] "+msg+"\n", args...)
}

func (l *VerboseLogger) Debug(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[workflow] DEBUG: "+msg+"\n", args...)
}

func (l *VerboseLogger) Error(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[workflow] ERROR: "+msg+"\n", args...)
}

func (r *Runner) Run(ctx context.Context, workflowPath string) (*RunState, error) {
	workflow, err := r.parser.ParseFile(workflowPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}
	return r.RunWorkflow(ctx, workflow)
}

type stepGroup struct {
	steps []Step
}

func (r *Runner) RunWorkflow(ctx context.Context, workflow *Workflow) (*RunState, error) {
	r.workflowStart = time.Now()

	if err := workflow.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	var state *RunState
	if r.options.ResumeRunID != "" {
		existing, err := r.stateManager.Load(r.options.ResumeRunID)
		if err != nil {
			return nil, fmt.Errorf("failed to resume run %s: %w", r.options.ResumeRunID, err)
		}
		state = existing
		state.Status = RunStatusRunning
		r.logger.Info("Resuming workflow run %s", state.RunID)
	} else {
		state = NewRunState(*workflow)
		state.Status = RunStatusRunning
		for k, v := range r.options.GlobalEnv {
			if state.Env == nil {
				state.Env = make(map[string]string)
			}
			state.Env[k] = v
		}
	}

	maxParallel := workflow.MaxParallel
	if maxParallel <= 0 {
		maxParallel = 4
	}

	groups, err := r.groupStepsByLevel(workflow)
	if err != nil {
		return nil, err
	}

	totalSteps := len(workflow.Steps)

	if r.watcher != nil {
		r.watcher.EmitWorkflowStarted(workflow.Name, state.RunID, totalSteps)
	}

	r.logger.Info("Starting workflow: %s (%d steps, maxParallel=%d)", workflow.Name, totalSteps, maxParallel)

	if r.options.DryRun {
		r.logger.Info("DRY RUN - no steps will be executed")
		if r.watcher != nil {
			r.watcher.EmitWorkflowCompleted(0)
			r.watcher.Stop()
		}
		return state, nil
	}

	stepNum := 0
	for groupIdx, group := range groups {
		var parallelSteps []Step
		var sequentialSteps []Step

		for _, step := range group.steps {
			if step.Parallel {
				parallelSteps = append(parallelSteps, step)
			} else {
				sequentialSteps = append(sequentialSteps, step)
			}
		}

		if len(parallelSteps) > 0 {
			r.logger.Info("Executing %d parallel steps (level %d/%d)", len(parallelSteps), groupIdx+1, len(groups))
			completed, failed, err := r.executeParallelSteps(ctx, state, parallelSteps, maxParallel)
			if err != nil {
				if r.watcher != nil {
					r.watcher.EmitWorkflowFailed(err, time.Since(r.workflowStart))
					r.watcher.Stop()
				}
				return state, err
			}
			if r.watcher != nil {
				for _, stepName := range completed {
					stepNum++
					r.watcher.EmitStepCompleted(stepName, stepNum, totalSteps, 0)
				}
				for stepName, errInfo := range failed {
					stepNum++
					r.watcher.EmitStepFailed(stepName, stepNum, totalSteps, errInfo.err, errInfo.exitCode)
				}
			}
		}

		for _, step := range sequentialSteps {
			stepNum++
			stepStartTime := time.Now()

			if state.IsStepCompleted(step.Name) && !r.options.Force {
				r.logger.Info("[%d/%d] Skipping completed step: %s", stepNum, totalSteps, step.Name)
				if r.watcher != nil {
					r.watcher.EmitStepSkipped(step.Name, stepNum, totalSteps, "already completed")
				}
				continue
			}

			if !r.dependenciesMet(state, step) {
				state.Status = RunStatusFailed
				state.Error = fmt.Sprintf("dependencies not met for step %s", step.Name)
				r.saveState(state)
				if r.watcher != nil {
					r.watcher.EmitWorkflowFailed(fmt.Errorf("dependencies not met for step %s", step.Name), time.Since(r.workflowStart))
					r.watcher.Stop()
				}
				return state, fmt.Errorf("dependencies not met for step %s", step.Name)
			}

			if step.Condition != "" {
				shouldRun, err := r.evaluateCondition(state, step.Condition)
				if err != nil {
					r.logger.Error("[%d/%d] Failed to evaluate condition for step %s: %v", stepNum, totalSteps, step.Name, err)
					if !step.ContinueOnError {
						state.Status = RunStatusFailed
						state.Error = fmt.Sprintf("condition evaluation failed for step %s: %v", step.Name, err)
						r.saveState(state)
						if r.watcher != nil {
							r.watcher.EmitWorkflowFailed(err, time.Since(r.workflowStart))
							r.watcher.Stop()
						}
						return state, err
					}
					continue
				}
				if !shouldRun {
					r.logger.Info("[%d/%d] Skipping step %s (condition not met)", stepNum, totalSteps, step.Name)
					if r.watcher != nil {
						r.watcher.EmitStepSkipped(step.Name, stepNum, totalSteps, "condition not met")
					}
					continue
				}
			}

			state.CurrentStep = step.Name
			r.saveState(state)
			r.logger.Info("[%d/%d] Executing step: %s", stepNum, totalSteps, step.Name)

			if r.watcher != nil {
				r.watcher.EmitStepStarted(step.Name, stepNum, totalSteps)
			}

			result := r.executeStep(ctx, state, step)
			state.AddStepResult(result)
			r.saveState(state)

			stepDuration := time.Since(stepStartTime)

			if result.Output.ExitCode != 0 && !step.ContinueOnError {
				state.Status = RunStatusFailed
				now := time.Now()
				state.FinishedAt = &now
				r.saveState(state)
				r.logger.Error("[%d/%d] Step %s failed with exit code %d", stepNum, totalSteps, step.Name, result.Output.ExitCode)
				if r.watcher != nil {
					r.watcher.EmitStepFailed(step.Name, stepNum, totalSteps, fmt.Errorf("%s", result.Output.Error), result.Output.ExitCode)
					r.watcher.EmitWorkflowFailed(fmt.Errorf("step %s failed: %s", step.Name, result.Output.Error), time.Since(r.workflowStart))
					r.watcher.Stop()
				}
				return state, fmt.Errorf("step %s failed: %s", step.Name, result.Output.Error)
			}

			if r.watcher != nil {
				r.watcher.EmitStepCompleted(step.Name, stepNum, totalSteps, stepDuration)
			}
		}
	}

	state.Status = RunStatusCompleted
	state.CurrentStep = ""
	now := time.Now()
	state.FinishedAt = &now
	r.saveState(state)

	workflowDuration := time.Since(r.workflowStart)
	r.logger.Info("Workflow completed successfully: %s", workflow.Name)

	if r.watcher != nil {
		r.watcher.EmitWorkflowCompleted(workflowDuration)
		r.watcher.Stop()
	}

	return state, nil
}

func (r *Runner) groupStepsByLevel(workflow *Workflow) ([]stepGroup, error) {
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)
	stepMap := make(map[string]Step)

	for _, step := range workflow.Steps {
		inDegree[step.Name] = 0
		stepMap[step.Name] = step
	}

	for _, step := range workflow.Steps {
		for _, dep := range step.DependsOn {
			inDegree[step.Name]++
			dependents[dep] = append(dependents[dep], step.Name)
		}
	}

	var groups []stepGroup
	remaining := make(map[string]bool)
	for _, step := range workflow.Steps {
		remaining[step.Name] = true
	}

	for len(remaining) > 0 {
		var currentGroup stepGroup
		var ready []string

		for stepName := range remaining {
			if inDegree[stepName] == 0 {
				ready = append(ready, stepName)
			}
		}

		if len(ready) == 0 {
			var remainingSteps []string
			for stepName := range remaining {
				remainingSteps = append(remainingSteps, stepName)
			}
			return nil, fmt.Errorf("circular dependency detected involving steps: %v", remainingSteps)
		}

		for _, stepName := range ready {
			currentGroup.steps = append(currentGroup.steps, stepMap[stepName])
			delete(remaining, stepName)
			for _, dependent := range dependents[stepName] {
				inDegree[dependent]--
			}
		}

		groups = append(groups, currentGroup)
	}

	return groups, nil
}

type stepErrorInfo struct {
	err      error
	exitCode int
}

func (r *Runner) executeParallelSteps(ctx context.Context, state *RunState, steps []Step, maxParallel int) ([]string, map[string]stepErrorInfo, error) {
	if len(steps) == 0 {
		return []string{}, map[string]stepErrorInfo{}, nil
	}

	if len(steps) == 1 {
		return r.executeSingleParallelStep(ctx, state, steps[0])
	}

	sem := make(chan struct{}, maxParallel)

	type stepResult struct {
		result StepResult
		step   Step
	}
	results := make(chan stepResult, len(steps))
	errChan := make(chan error, len(steps))

	var wg sync.WaitGroup
	var stopOnce sync.Once
	stopChan := make(chan struct{})
	var hasFatalError atomic.Value
	hasFatalError.Store(false)

	for _, step := range steps {
		wg.Add(1)
		go func(s Step) {
			defer wg.Done()

			select {
			case <-stopChan:
				return
			default:
			}

			sem <- struct{}{}
			defer func() { <-sem }()

			if !r.dependenciesMet(state, s) {
				errChan <- fmt.Errorf("dependencies not met for step %s", s.Name)
				if !s.ContinueOnError {
					stopOnce.Do(func() { close(stopChan) })
					hasFatalError.Store(true)
				}
				return
			}

			if s.Condition != "" {
				shouldRun, err := r.evaluateCondition(state, s.Condition)
				if err != nil {
					r.logger.Error("Failed to evaluate condition for step %s: %v", s.Name, err)
					if !s.ContinueOnError {
						errChan <- fmt.Errorf("condition evaluation failed for step %s: %v", s.Name, err)
						stopOnce.Do(func() { close(stopChan) })
						hasFatalError.Store(true)
						return
					}
				}
				if !shouldRun {
					r.logger.Info("Skipping step %s (condition not met)", s.Name)
					results <- stepResult{
						result: StepResult{
							Step: s,
							Output: StepOutput{
								StepName:  s.Name,
								ExitCode:  0,
								StartedAt: time.Now(),
							},
						},
						step: s,
					}
					return
				}
			}

			if state.IsStepCompleted(s.Name) && !r.options.Force {
				r.logger.Info("Skipping completed step: %s", s.Name)
				return
			}

			r.logger.Info("Executing step: %s", s.Name)
			result := r.executeStep(ctx, state, s)
			results <- stepResult{result: result, step: s}

			if result.Output.ExitCode != 0 && !s.ContinueOnError {
				errChan <- fmt.Errorf("step %s failed: %s", s.Name, result.Output.Error)
				stopOnce.Do(func() { close(stopChan) })
				hasFatalError.Store(true)
			}
		}(step)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errChan)
	}()

	var stepResults []StepResult
	var errors []error

	for result := range results {
		stepResults = append(stepResults, result.result)
	}

	for err := range errChan {
		errors = append(errors, err)
	}

	for _, result := range stepResults {
		state.AddStepResult(result)
		state.CurrentStep = result.Step.Name
	}
	r.saveState(state)

	var completed []string
	failed := make(map[string]stepErrorInfo)

	for _, result := range stepResults {
		if result.Output.ExitCode == 0 {
			completed = append(completed, result.Step.Name)
		} else {
			failed[result.Step.Name] = stepErrorInfo{
				err:      fmt.Errorf("%s", result.Output.Error),
				exitCode: result.Output.ExitCode,
			}
		}
	}

	if hasFatalError.Load().(bool) && len(errors) > 0 {
		return completed, failed, errors[0]
	}

	return completed, failed, nil
}

func (r *Runner) executeSingleParallelStep(ctx context.Context, state *RunState, step Step) ([]string, map[string]stepErrorInfo, error) {
	if state.IsStepCompleted(step.Name) && !r.options.Force {
		r.logger.Info("Skipping completed step: %s", step.Name)
		return []string{step.Name}, map[string]stepErrorInfo{}, nil
	}

	if !r.dependenciesMet(state, step) {
		state.Status = RunStatusFailed
		state.Error = fmt.Sprintf("dependencies not met for step %s", step.Name)
		r.saveState(state)
		return []string{}, map[string]stepErrorInfo{}, fmt.Errorf("dependencies not met for step %s", step.Name)
	}

	if step.Condition != "" {
		shouldRun, err := r.evaluateCondition(state, step.Condition)
		if err != nil {
			r.logger.Error("Failed to evaluate condition for step %s: %v", step.Name, err)
			if !step.ContinueOnError {
				state.Status = RunStatusFailed
				state.Error = fmt.Sprintf("condition evaluation failed for step %s: %v", step.Name, err)
				r.saveState(state)
				return []string{}, map[string]stepErrorInfo{}, err
			}
			return []string{}, map[string]stepErrorInfo{}, nil
		}
		if !shouldRun {
			r.logger.Info("Skipping step %s (condition not met)", step.Name)
			return []string{step.Name}, map[string]stepErrorInfo{}, nil
		}
	}

	state.CurrentStep = step.Name
	r.saveState(state)
	r.logger.Info("Executing step: %s", step.Name)

	result := r.executeStep(ctx, state, step)
	state.AddStepResult(result)
	r.saveState(state)

	if result.Output.ExitCode != 0 && !step.ContinueOnError {
		state.Status = RunStatusFailed
		now := time.Now()
		state.FinishedAt = &now
		r.saveState(state)
		r.logger.Error("Step %s failed with exit code %d", step.Name, result.Output.ExitCode)
		return []string{}, map[string]stepErrorInfo{step.Name: {err: fmt.Errorf("%s", result.Output.Error), exitCode: result.Output.ExitCode}}, fmt.Errorf("step %s failed: %s", step.Name, result.Output.Error)
	}

	if result.Output.ExitCode == 0 {
		return []string{step.Name}, map[string]stepErrorInfo{}, nil
	}
	return []string{}, map[string]stepErrorInfo{step.Name: {err: fmt.Errorf("%s", result.Output.Error), exitCode: result.Output.ExitCode}}, nil
}

func (r *Runner) executeStep(ctx context.Context, state *RunState, step Step) StepResult {
	startTime := time.Now()
	var lastResult StepResult
	var totalRetries int

	for attempt := 0; attempt <= step.RetryCount; attempt++ {
		if attempt > 0 {
			totalRetries = attempt
			delay := r.calculateRetryDelay(step, attempt)
			r.logger.Info("Retrying step %s (attempt %d/%d) after %v delay", step.Name, attempt, step.RetryCount, delay)
			time.Sleep(delay)
		}

		lastResult = r.executeStepOnce(ctx, state, step, startTime)

		if lastResult.Output.ExitCode == 0 {
			lastResult.Output.RetryCount = step.RetryCount
			lastResult.Output.Retries = totalRetries
			return lastResult
		}

		if attempt == step.RetryCount || step.ContinueOnError {
			break
		}

		r.logger.Error("Step %s failed (attempt %d/%d): %s", step.Name, attempt+1, step.RetryCount+1, lastResult.Output.Error)
	}

	lastResult.Output.RetryCount = step.RetryCount
	lastResult.Output.Retries = totalRetries
	return lastResult
}

func (r *Runner) calculateRetryDelay(step Step, attempt int) time.Duration {
	if step.RetryDelay <= 0 {
		return 0
	}

	switch step.RetryBackoff {
	case "exponential":
		return step.RetryDelay * time.Duration(1<<(attempt-1))
	case "linear", "":
		return step.RetryDelay * time.Duration(attempt)
	default:
		return step.RetryDelay * time.Duration(attempt)
	}
}

func (r *Runner) executeStepOnce(ctx context.Context, state *RunState, step Step, startTime time.Time) StepResult {
	r.logger.Info("Executing step: %s", step.Name)
	r.logger.Debug("Command: %s", step.Command)

	interpolator := NewInterpolator(state.StepOutputs, state.Env)

	command, err := interpolator.Interpolate(step.Command)
	if err != nil {
		return StepResult{
			Step: step,
			Output: StepOutput{
				StepName:   step.Name,
				ExitCode:   1,
				StartedAt:  startTime,
				FinishedAt: time.Now(),
				Duration:   time.Since(startTime),
				Error:      fmt.Sprintf("interpolation failed: %v", err),
			},
		}
	}

	env := make(map[string]string)
	for k, v := range state.Env {
		env[k] = v
	}
	for k, v := range step.Env {
		interpolated, err := interpolator.Interpolate(v)
		if err != nil {
			r.logger.Error("Failed to interpolate env var %s: %v", k, err)
			env[k] = v
		} else {
			env[k] = interpolated
		}
	}

	var cmd *exec.Cmd
	switch step.Type {
	case StepTypeGPD:
		gpdPath, err := r.findGpdBinary()
		if err != nil {
			return StepResult{
				Step: step,
				Output: StepOutput{
					StepName:   step.Name,
					ExitCode:   1,
					StartedAt:  startTime,
					FinishedAt: time.Now(),
					Duration:   time.Since(startTime),
					Error:      fmt.Sprintf("failed to find gpd binary: %v", err),
				},
			}
		}
		args := strings.Fields(strings.TrimPrefix(command, "gpd "))
		cmd = exec.CommandContext(ctx, gpdPath, args...)
	default:
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	workingDir := step.WorkingDir
	if workingDir == "" {
		workingDir = r.options.WorkingDir
	}
	if workingDir != "" {
		interpolated, err := interpolator.Interpolate(workingDir)
		if err == nil {
			cmd.Dir = interpolated
		} else {
			cmd.Dir = workingDir
		}
	}

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	if step.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)
		cmd.Dir = workingDir
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	r.logger.Debug("Step %s completed in %v (exit code: %d)", step.Name, duration, exitCode)

	data := make(map[string]interface{})
	if exitCode == 0 && len(step.CaptureOutputs) > 0 {
		captured, err := ParseAndExtractJSON(stdout, step.CaptureOutputs)
		if err == nil {
			data = captured
			r.logger.Debug("Captured outputs: %v", captured)
		} else {
			r.logger.Debug("Failed to capture outputs: %v", err)
		}
	}

	return StepResult{
		Step: step,
		Output: StepOutput{
			StepName:   step.Name,
			ExitCode:   exitCode,
			Stdout:     stdout,
			Stderr:     stderr,
			Data:       data,
			StartedAt:  startTime,
			FinishedAt: endTime,
			Duration:   duration,
			Error: func() string {
				if err != nil {
					return err.Error()
				}
				return ""
			}(),
		},
	}
}

func (r *Runner) findGpdBinary() (string, error) {
	gpdPath, err := exec.LookPath("gpd")
	if err == nil {
		return gpdPath, nil
	}

	possiblePaths := []string{
		"./gpd",
		"./bin/gpd",
		"../bin/gpd",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, err := os.Getwd()
			if err == nil {
				return absPath + "/" + path, nil
			}
			return path, nil
		}
	}

	return "", fmt.Errorf("gpd binary not found in PATH or common locations")
}

func (r *Runner) dependenciesMet(state *RunState, step Step) bool {
	for _, dep := range step.DependsOn {
		if !state.IsStepCompleted(dep) {
			return false
		}
	}
	return true
}

func (r *Runner) evaluateCondition(state *RunState, condition string) (bool, error) {
	interpolator := NewInterpolator(state.StepOutputs, state.Env)

	interpolated, err := interpolator.Interpolate(condition)
	if err != nil {
		return false, err
	}

	interpolated = strings.TrimSpace(interpolated)

	switch strings.ToLower(interpolated) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "":
		return false, nil
	default:
		return true, nil
	}
}

func (r *Runner) saveState(state *RunState) {
	if err := r.stateManager.Save(state); err != nil {
		r.logger.Error("Failed to save run state: %v", err)
	}
}

func (r *Runner) GetStateManager() *StateManager {
	return r.stateManager
}
