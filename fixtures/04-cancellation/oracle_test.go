package taskcancellation

import (
	"context"
	"errors"
	"testing"
)

func TestExecutePipeline(t *testing.T) {
	ctx := context.Background()
	cleanupCalled := false
	executed := 0

	steps := []PipelineStep{
		{Name: "step1", Run: func(ctx context.Context) error { executed++; return nil }},
		{Name: "step2", Run: func(ctx context.Context) error { executed++; return nil }},
	}

	res, err := ExecutePipeline(ctx, steps, func() { cleanupCalled = true })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.CompletedCount != 2 || executed != 2 {
		t.Fatalf("expected 2 completed steps, got: %d", res.CompletedCount)
	}
	if cleanupCalled {
		t.Fatalf("cleanup must not be called on success")
	}
}

func TestOracleNegativeControls(t *testing.T) {
	// Negative 1: Canceled before starting
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cleanupCount := 0
	steps := []PipelineStep{
		{Name: "step1", Run: func(ctx context.Context) error { return nil }},
	}
	_, err := ExecutePipeline(ctx, steps, func() { cleanupCount++ })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if cleanupCount != 1 {
		t.Fatalf("cleanup must be called exactly once on initial cancellation, got: %d", cleanupCount)
	}

	// Negative 2: Cancellation midway through pipeline
	ctx, cancel = context.WithCancel(context.Background())
	step1Run := false
	step2Run := false
	cleanupCount = 0

	pipeline := []PipelineStep{
		{
			Name: "step1",
			Run: func(ctx context.Context) error {
				step1Run = true
				cancel() // cancel context during step 1
				return nil
			},
		},
		{
			Name: "step2",
			Run: func(ctx context.Context) error {
				step2Run = true
				return nil
			},
		},
	}

	_, err = ExecutePipeline(ctx, pipeline, func() { cleanupCount++ })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if !step1Run {
		t.Fatalf("step 1 should have run")
	}
	if step2Run {
		t.Fatalf("CRITICAL: step 2 must NOT run after context cancellation")
	}
	if cleanupCount != 1 {
		t.Fatalf("cleanup must be called exactly once on cancellation, got: %d", cleanupCount)
	}
}
