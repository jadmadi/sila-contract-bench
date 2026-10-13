# Task: Delimited Token Stream Parser

## Objective
Implement `ParseDelimitedTokens(input string, maxDepth int) (*TokenNode, error)` in Go.

## Requirements
1. Parse balanced curly braces `{}` and square brackets `[]`.
2. Each pair of braces or brackets defines a node. Braces define kind `"object"`, brackets define kind `"array"`.
3. Track nesting depth (starting at depth 0 for root, incrementing inside each opened delimiter):
   - If depth exceeds `maxDepth`, immediately return error `DEPTH_LIMIT_EXCEEDED: max depth <N> exceeded`.
4. Enforce strict closing delimiter matches:
   - `{` must be closed by `}`. If closed by `]`, return error `MISMATCHED_DELIMITER: expected }, got ]`.
   - `[` must be closed by `]`. If closed by `}`, return error `MISMATCHED_DELIMITER: expected ], got }`.
   - Closing a delimiter when none is open returns error `UNEXPECTED_CLOSING_DELIMITER`.
   - Reaching end of input with open delimiters returns error `UNTERMINATED_DELIMITER`.
5. Non-delimiter characters outside and inside are collected as text tokens in node content.

## Interfaces
```go
package parserboundary

type TokenNode struct {
	Kind     string       `json:"kind"` // "object", "array", or "root"
	Depth    int          `json:"depth"`
	Children []*TokenNode `json:"children,omitempty"`
}

func ParseDelimitedTokens(input string, maxDepth int) (*TokenNode, error)
```

## Constraints
- Pure Go standard library only (`fmt`, `strings`).
- Zero panics on malformed input.
