package multistepresume

import (
	"errors"
	"testing"
)

type mockRunner struct {
	executed []string
	failAt   string
	failErr  error
}

func (m *mockRunner) Execute(step StepDefinition) error {
	m.executed = append(m.executed, step.ID)
	if m.failAt != "" && step.ID == m.failAt {
		if m.failErr != nil {
			return m.failErr
		}
		return errors.New("execution failed")
	}
	return nil
}

func TestResumeTaskWorkflow_FreshRun(t *testing.T) {
	steps := []StepDefinition{
		{ID: "step-1", Title: "First Step"},
		{ID: "step-2", Title: "Second Step"},
		{ID: "step-3", Title: "Third Step"},
	}

	runner := &mockRunner{}
	cp, err := ResumeTaskWorkflow(nil, steps, runner)
	if err != nil {
		t.Fatalf("unexpected error on fresh run: %v", err)
	}

	if len(cp.CompletedSteps) != 3 {
		t.Fatalf("expected 3 completed steps, got: %d", len(cp.CompletedSteps))
	}
	if cp.CurrentStep != "" {
		t.Fatalf("expected empty CurrentStep on completion, got: %s", cp.CurrentStep)
	}
	if cp.LastError != "" {
		t.Fatalf("expected empty LastError on completion, got: %s", cp.LastError)
	}
	if len(runner.executed) != 3 {
		t.Fatalf("expected runner to execute 3 steps, executed: %d", len(runner.executed))
	}
}

func TestResumeTaskWorkflow_ResumeFromCheckpoint(t *testing.T) {
	steps := []StepDefinition{
		{ID: "step-1", Title: "First Step"},
		{ID: "step-2", Title: "Second Step"},
		{ID: "step-3", Title: "Third Step"},
	}

	initialCP := &Checkpoint{
		CompletedSteps: []string{"step-1"},
		CurrentStep:    "",
		LastError:      "",
	}

	runner := &mockRunner{}
	cp, err := ResumeTaskWorkflow(initialCP, steps, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cp.CompletedSteps) != 3 {
		t.Fatalf("expected 3 completed steps, got: %d", len(cp.CompletedSteps))
	}
	// Check runner only executed step-2 and step-3
	if len(runner.executed) != 2 {
		t.Fatalf("expected runner to execute 2 steps, got: %d", len(runner.executed))
	}
	if runner.executed[0] != "step-2" || runner.executed[1] != "step-3" {
		t.Fatalf("expected runner to execute [step-2, step-3], got: %v", runner.executed)
	}
}

func TestOracleNegativeControls_HaltOnError(t *testing.T) {
	steps := []StepDefinition{
		{ID: "step-1", Title: "First Step"},
		{ID: "step-2", Title: "Failing Step"},
		{ID: "step-3", Title: "Third Step"},
	}

	runner := &mockRunner{
		failAt:  "step-2",
		failErr: errors.New("simulated error at step 2"),
	}

	cp, err := ResumeTaskWorkflow(nil, steps, runner)
	if err == nil {
		t.Fatalf("expected error from step 2, got nil")
	}

	// Must halt immediately: step-3 must not be executed
	if len(runner.executed) != 2 {
		t.Fatalf("expected runner to stop after step 2, executed: %v", runner.executed)
	}
	if len(cp.CompletedSteps) != 1 || cp.CompletedSteps[0] != "step-1" {
		t.Fatalf("expected CompletedSteps to only have step-1, got: %v", cp.CompletedSteps)
	}
	if cp.CurrentStep != "step-2" {
		t.Fatalf("expected CurrentStep to be step-2, got: %s", cp.CurrentStep)
	}
	if cp.LastError != "simulated error at step 2" {
		t.Fatalf("expected LastError to record error, got: %s", cp.LastError)
	}
}
