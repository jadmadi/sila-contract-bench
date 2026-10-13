package parserboundary

import (
	"fmt"
)

type TokenNode struct {
	Kind     string       `json:"kind"` // "root", "object", or "array"
	Depth    int          `json:"depth"`
	Children []*TokenNode `json:"children,omitempty"`
}

func ParseDelimitedTokens(input string, maxDepth int) (*TokenNode, error) {
	root := &TokenNode{Kind: "root", Depth: 0}
	stack := []*TokenNode{root}

	for _, ch := range input {
		switch ch {
		case '{', '[':
			curr := stack[len(stack)-1]
			newDepth := curr.Depth + 1
			if newDepth > maxDepth {
				return nil, fmt.Errorf("DEPTH_LIMIT_EXCEEDED: max depth %d exceeded", maxDepth)
			}
			kind := "object"
			if ch == '[' {
				kind = "array"
			}
			node := &TokenNode{Kind: kind, Depth: newDepth}
			curr.Children = append(curr.Children, node)
			stack = append(stack, node)

		case '}':
			if len(stack) <= 1 {
				return nil, fmt.Errorf("UNEXPECTED_CLOSING_DELIMITER: unexpected } at root")
			}
			top := stack[len(stack)-1]
			if top.Kind != "object" {
				return nil, fmt.Errorf("MISMATCHED_DELIMITER: expected ], got }")
			}
			stack = stack[:len(stack)-1]

		case ']':
			if len(stack) <= 1 {
				return nil, fmt.Errorf("UNEXPECTED_CLOSING_DELIMITER: unexpected ] at root")
			}
			top := stack[len(stack)-1]
			if top.Kind != "array" {
				return nil, fmt.Errorf("MISMATCHED_DELIMITER: expected }, got ]")
			}
			stack = stack[:len(stack)-1]
		}
	}

	if len(stack) > 1 {
		return nil, fmt.Errorf("UNTERMINATED_DELIMITER: %d unclosed delimiter(s)", len(stack)-1)
	}

	return root, nil
}
