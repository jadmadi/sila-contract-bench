package conditionaldependency

import (
	"fmt"
	"sort"
)

type TaskNode struct {
	ID        string
	DependsOn []string
}

func ResolveExecutableTasks(tasks []TaskNode, completed map[string]bool) (executable []string, blocked []string, err error) {
	adj := make(map[string][]string)
	allNodes := make(map[string]bool)

	for _, t := range tasks {
		allNodes[t.ID] = true
		adj[t.ID] = append(adj[t.ID], t.DependsOn...)
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(curr string) bool
	hasCycle = func(curr string) bool {
		visited[curr] = true
		recStack[curr] = true

		for _, neighbor := range adj[curr] {
			if !visited[neighbor] {
				if hasCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[curr] = false
		return false
	}

	for id := range allNodes {
		if !visited[id] {
			if hasCycle(id) {
				return nil, nil, fmt.Errorf("CYCLE_DETECTED: cycle found in dependency graph")
			}
		}
	}

	// Evaluate eligibility
	for _, t := range tasks {
		if completed[t.ID] {
			continue // Already done
		}

		allMet := true
		for _, dep := range t.DependsOn {
			if !completed[dep] {
				allMet = false
				break
			}
		}

		if allMet {
			executable = append(executable, t.ID)
		} else {
			blocked = append(blocked, t.ID)
		}
	}

	sort.Strings(executable)
	sort.Strings(blocked)

	return executable, blocked, nil
}
