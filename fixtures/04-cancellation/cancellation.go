package taskcancellation

import (
	"context"
)

type PipelineStep struct {
	Name string
	Run  func(ctx context.Context) error
}

type PipelineResult struct {
	CompletedCount int
}

func ExecutePipeline(ctx context.Context, steps []PipelineStep, onCleanup func()) (*PipelineResult, error) {
	if err := ctx.Err(); err != nil {
		if onCleanup != nil {
			onCleanup()
		}
		return nil, err
	}

	completed := 0
	for _, step := range steps {
		if err := ctx.Err(); err != nil {
			if onCleanup != nil {
				onCleanup()
			}
			return nil, err
		}

		if err := step.Run(ctx); err != nil {
			if onCleanup != nil {
				onCleanup()
			}
			return nil, err
		}
		completed++
	}

	return &PipelineResult{CompletedCount: completed}, nil
}
