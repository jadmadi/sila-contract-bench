# Task: Conditional Dependency Resolver

## Objective
Implement `ResolveExecutableTasks(tasks []TaskNode, completed map[string]bool) (executable []string, blocked []string, err error)` in Go.

## Requirements
1. Each `TaskNode` has an `ID` and a list of dependencies `DependsOn []string`.
2. Cycle Detection:
   - If there is any dependency cycle (e.g. A -> B -> A, or self-dependency A -> A), return error `CYCLE_DETECTED: cycle found in dependency graph`.
3. Execution Eligibility:
   - For each task that is NOT already in `completed`:
     - If all of its `DependsOn` tasks are in `completed`, add its ID to `executable`.
     - Otherwise, add its ID to `blocked`.
4. Determinism:
   - Both `executable` and `blocked` return slices must be sorted alphabetically.

## Interfaces
```go
package conditionaldependency

type TaskNode struct {
	ID        string
	DependsOn []string
}

func ResolveExecutableTasks(tasks []TaskNode, completed map[string]bool) (executable []string, blocked []string, err error)
```

## Constraints
- Pure Go standard library only (`fmt`, `sort`).
- Zero panics. Cycles must be detected cleanly.
