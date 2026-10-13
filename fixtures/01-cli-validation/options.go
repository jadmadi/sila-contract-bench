package clivalidation

import (
	"fmt"
	"strconv"
	"strings"
)

type RunOptions struct {
	Timeout int      `json:"timeout"`
	Format  string   `json:"format"`
	Strict  bool     `json:"strict"`
	Targets []string `json:"targets,omitempty"`
}

func ValidateAndParseRunOptions(args []string) (*RunOptions, error) {
	opts := &RunOptions{
		Timeout: 30,
		Format:  "text",
		Strict:  false,
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			if strings.HasPrefix(arg, "--timeout=") {
				valStr := strings.TrimPrefix(arg, "--timeout=")
				val, err := strconv.Atoi(valStr)
				if err != nil || val <= 0 {
					return nil, fmt.Errorf("INVALID_TIMEOUT: timeout must be positive integer, got %q", valStr)
				}
				opts.Timeout = val
			} else if strings.HasPrefix(arg, "--format=") {
				val := strings.TrimPrefix(arg, "--format=")
				if val != "json" && val != "text" {
					return nil, fmt.Errorf("INVALID_FORMAT: format must be json or text, got %q", val)
				}
				opts.Format = val
			} else if arg == "--strict" {
				opts.Strict = true
			} else {
				return nil, fmt.Errorf("UNKNOWN_FLAG: %s", arg)
			}
		} else {
			opts.Targets = append(opts.Targets, arg)
		}
	}

	return opts, nil
}
