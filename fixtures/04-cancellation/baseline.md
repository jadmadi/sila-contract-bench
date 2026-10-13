# Task: Context Cancellation and Graceful Cleanup

## Objective
Implement `ExecutePipeline(ctx context.Context, steps []PipelineStep, onCleanup func()) (*PipelineResult, error)` in Go.

## Requirements
1. Execute `steps` sequentially in order.
2. Before and between every step, check `ctx.Err()`. If `ctx.Err() != nil`:
   - Immediately abort execution.
   - Invoke `onCleanup()` exactly once (if `onCleanup != nil`).
   - Return `nil, ctx.Err()`.
3. If a step function returns an error:
   - Abort execution immediately.
   - Invoke `onCleanup()` exactly once.
   - Return `nil, err`.
4. If all steps succeed:
   - Do NOT invoke `onCleanup()`.
   - Return `&PipelineResult{CompletedCount: len(steps)}`, `nil`.

## Interfaces
```go
package taskcancellation

import "context"

type PipelineStep struct {
	Name string
	Run  func(ctx context.Context) error
}

type PipelineResult struct {
	CompletedCount int
}

func ExecutePipeline(ctx context.Context, steps []PipelineStep, onCleanup func()) (*PipelineResult, error)
```

## Constraints
- Pure Go standard library only (`context`, `fmt`).
- Must not leak goroutines or execute trailing steps after cancellation.
