package conditionaldependency

import (
	"strings"
	"testing"
)

func TestResolveExecutableTasks(t *testing.T) {
	tasks := []TaskNode{
		{ID: "T-001", DependsOn: nil},
		{ID: "T-002", DependsOn: []string{"T-001"}},
		{ID: "T-003", DependsOn: []string{"T-002"}},
		{ID: "T-004", DependsOn: []string{"T-001"}},
	}

	// 1. Initially only T-001 is executable
	completed := make(map[string]bool)
	exec, block, err := ResolveExecutableTasks(tasks, completed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec) != 1 || exec[0] != "T-001" {
		t.Fatalf("expected executable [T-001], got: %v", exec)
	}
	if len(block) != 3 {
		t.Fatalf("expected 3 blocked tasks, got: %v", block)
	}

	// 2. When T-001 is completed, T-002 and T-004 become executable
	completed["T-001"] = true
	exec, block, err = ResolveExecutableTasks(tasks, completed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec) != 2 || exec[0] != "T-002" || exec[1] != "T-004" {
		t.Fatalf("expected executable [T-002, T-004], got: %v", exec)
	}
	if len(block) != 1 || block[0] != "T-003" {
		t.Fatalf("expected blocked [T-003], got: %v", block)
	}
}

func TestOracleNegativeControls(t *testing.T) {
	// Negative 1: Cycle A -> B -> A
	cyclicTasks := []TaskNode{
		{ID: "A", DependsOn: []string{"B"}},
		{ID: "B", DependsOn: []string{"A"}},
	}
	_, _, err := ResolveExecutableTasks(cyclicTasks, nil)
	if err == nil || !strings.Contains(err.Error(), "CYCLE_DETECTED") {
		t.Fatalf("oracle must reject cycle with CYCLE_DETECTED, got: %v", err)
	}

	// Negative 2: Self dependency A -> A
	selfCycle := []TaskNode{
		{ID: "A", DependsOn: []string{"A"}},
	}
	_, _, err = ResolveExecutableTasks(selfCycle, nil)
	if err == nil || !strings.Contains(err.Error(), "CYCLE_DETECTED") {
		t.Fatalf("oracle must reject self cycle with CYCLE_DETECTED, got: %v", err)
	}
}
