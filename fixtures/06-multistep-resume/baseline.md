# Task: Multi-Step Task Checkpoint and Resume

## Objective
Implement `ResumeTaskWorkflow(cp *Checkpoint, steps []StepDefinition, runner StepRunner) (*Checkpoint, error)` in Go.

## Requirements
1. `Checkpoint` records `CompletedSteps []string`, `CurrentStep string`, and `LastError string`.
2. Inspect `cp.CompletedSteps`:
   - Any step in `steps` whose `ID` is already in `cp.CompletedSteps` must be SKIPPED without calling `runner`.
3. For remaining uncompleted steps in sequential order:
   - Update `cp.CurrentStep = step.ID`.
   - Call `runner.Execute(step)`.
   - If `runner.Execute` succeeds:
     - Append `step.ID` to `cp.CompletedSteps`.
   - If `runner.Execute` fails:
     - Set `cp.LastError = err.Error()`.
     - Return the updated checkpoint and the error immediately without executing subsequent steps.
4. On full completion of all steps, set `cp.CurrentStep = ""` and `cp.LastError = ""`.

## Interfaces
```go
package multistepresume

type StepDefinition struct {
	ID    string
	Title string
}

type Checkpoint struct {
	CompletedSteps []string
	CurrentStep    string
	LastError      string
}

type StepRunner interface {
	Execute(step StepDefinition) error
}

func ResumeTaskWorkflow(cp *Checkpoint, steps []StepDefinition, runner StepRunner) (*Checkpoint, error)
```

## Constraints
- Pure Go standard library only.
- Must never re-execute a step that is already recorded in `CompletedSteps`.
