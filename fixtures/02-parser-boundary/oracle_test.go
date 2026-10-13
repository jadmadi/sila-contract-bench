package parserboundary

import (
	"strings"
	"testing"
)

func TestParseDelimitedTokens(t *testing.T) {
	// 1. Valid nesting
	tree, err := ParseDelimitedTokens("{[{}][]}", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tree.Kind != "root" || len(tree.Children) != 1 {
		t.Fatalf("expected 1 root child, got: %+v", tree)
	}
	obj := tree.Children[0]
	if obj.Kind != "object" || obj.Depth != 1 || len(obj.Children) != 2 {
		t.Fatalf("expected 2 children in object, got: %+v", obj)
	}
}

func TestOracleNegativeControls(t *testing.T) {
	// Negative 1: Depth limit exceeded
	_, err := ParseDelimitedTokens("{{{}}}", 2)
	if err == nil || !strings.Contains(err.Error(), "DEPTH_LIMIT_EXCEEDED") {
		t.Fatalf("oracle must reject deep nesting with DEPTH_LIMIT_EXCEEDED, got: %v", err)
	}

	// Negative 2: Mismatched delimiter
	_, err = ParseDelimitedTokens("{]", 5)
	if err == nil || !strings.Contains(err.Error(), "MISMATCHED_DELIMITER") {
		t.Fatalf("oracle must reject {] with MISMATCHED_DELIMITER, got: %v", err)
	}
	_, err = ParseDelimitedTokens("[}", 5)
	if err == nil || !strings.Contains(err.Error(), "MISMATCHED_DELIMITER") {
		t.Fatalf("oracle must reject [} with MISMATCHED_DELIMITER, got: %v", err)
	}

	// Negative 3: Unterminated delimiter
	_, err = ParseDelimitedTokens("{[}", 5)
	if err == nil {
		t.Fatalf("oracle must reject {[} with error, got nil")
	}
	_, err = ParseDelimitedTokens("{{", 5)
	if err == nil || !strings.Contains(err.Error(), "UNTERMINATED_DELIMITER") {
		t.Fatalf("oracle must reject {{ with UNTERMINATED_DELIMITER, got: %v", err)
	}

	// Negative 4: Unexpected closing
	_, err = ParseDelimitedTokens("}", 5)
	if err == nil || !strings.Contains(err.Error(), "UNEXPECTED_CLOSING_DELIMITER") {
		t.Fatalf("oracle must reject } with UNEXPECTED_CLOSING_DELIMITER, got: %v", err)
	}
}
