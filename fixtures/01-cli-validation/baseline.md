# Task: CLI Option Validation and Parsing

## Objective
Implement `ValidateAndParseRunOptions(args []string) (*RunOptions, error)` in Go.

## Requirements
1. Parse flags from `args`:
   - `--timeout=<int>`: Timeout in seconds. Must be an integer greater than 0. Missing or <= 0 must return error `INVALID_TIMEOUT`.
   - `--format=<string>`: Output format. Must be either `json` or `text`. Any other value must return error `INVALID_FORMAT`. Defaults to `text` if omitted.
   - `--strict`: Boolean flag indicating strict mode. Defaults to false.
2. If an unrecognized flag is encountered (any argument starting with `--` other than the three supported flags), return error `UNKNOWN_FLAG: <flag>`.
3. Arguments not starting with `--` are treated as positional target arguments and stored in `Targets []string`.
4. Empty argument list is valid and returns default options (`Timeout: 30`, `Format: "text"`, `Strict: false`, `Targets: nil`).

## Interfaces
```go
package clivalidation

type RunOptions struct {
	Timeout int      `json:"timeout"`
	Format  string   `json:"format"`
	Strict  bool     `json:"strict"`
	Targets []string `json:"targets,omitempty"`
}

func ValidateAndParseRunOptions(args []string) (*RunOptions, error)
```

## Constraints
- Pure Go standard library only (`fmt`, `strconv`, `strings`). No external packages.
- Zero panic policy. All errors must be returned cleanly.
