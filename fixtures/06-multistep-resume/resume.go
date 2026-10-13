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

func ResumeTaskWorkflow(cp *Checkpoint, steps []StepDefinition, runner StepRunner) (*Checkpoint, error) {
	if cp == nil {
		cp = &Checkpoint{}
	}

	completedMap := make(map[string]bool)
	for _, id := range cp.CompletedSteps {
		completedMap[id] = true
	}

	for _, step := range steps {
		if completedMap[step.ID] {
			continue // Skip already completed step
		}

		cp.CurrentStep = step.ID
		if err := runner.Execute(step); err != nil {
			cp.LastError = err.Error()
			return cp, err
		}

		cp.CompletedSteps = append(cp.CompletedSteps, step.ID)
		completedMap[step.ID] = true
	}

	cp.CurrentStep = ""
	cp.LastError = ""
	return cp, nil
}
