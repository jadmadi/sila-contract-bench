package clivalidation

import (
	"strings"
	"testing"
)

func TestValidateAndParseRunOptions(t *testing.T) {
	// 1. Defaults
	opts, err := ValidateAndParseRunOptions(nil)
	if err != nil {
		t.Fatalf("unexpected error on nil args: %v", err)
	}
	if opts.Timeout != 30 || opts.Format != "text" || opts.Strict != false || len(opts.Targets) != 0 {
		t.Fatalf("unexpected defaults: %+v", opts)
	}

	// 2. Custom values
	opts, err = ValidateAndParseRunOptions([]string{"--timeout=15", "--format=json", "--strict", "file1.go", "file2.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Timeout != 15 || opts.Format != "json" || !opts.Strict {
		t.Fatalf("parsed options mismatch: %+v", opts)
	}
	if len(opts.Targets) != 2 || opts.Targets[0] != "file1.go" || opts.Targets[1] != "file2.go" {
		t.Fatalf("targets mismatch: %+v", opts.Targets)
	}
}

func TestOracleNegativeControls(t *testing.T) {
	// Negative 1: timeout <= 0
	_, err := ValidateAndParseRunOptions([]string{"--timeout=0"})
	if err == nil || !strings.Contains(err.Error(), "INVALID_TIMEOUT") {
		t.Fatalf("oracle must reject timeout=0 with INVALID_TIMEOUT, got: %v", err)
	}
	_, err = ValidateAndParseRunOptions([]string{"--timeout=-5"})
	if err == nil || !strings.Contains(err.Error(), "INVALID_TIMEOUT") {
		t.Fatalf("oracle must reject timeout=-5 with INVALID_TIMEOUT, got: %v", err)
	}

	// Negative 2: invalid format
	_, err = ValidateAndParseRunOptions([]string{"--format=yaml"})
	if err == nil || !strings.Contains(err.Error(), "INVALID_FORMAT") {
		t.Fatalf("oracle must reject format=yaml with INVALID_FORMAT, got: %v", err)
	}

	// Negative 3: unknown flag
	_, err = ValidateAndParseRunOptions([]string{"--unknown-flag"})
	if err == nil || !strings.Contains(err.Error(), "UNKNOWN_FLAG") {
		t.Fatalf("oracle must reject unknown flag with UNKNOWN_FLAG, got: %v", err)
	}
}
