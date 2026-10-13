package filepreservation

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeAtomicUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "data.txt")

	// 1. Initial write
	err := SafeAtomicUpdate(target, func(old []byte) ([]byte, error) {
		if old != nil {
			t.Fatalf("expected nil old data for new file")
		}
		return []byte("version 1"), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil || string(content) != "version 1" {
		t.Fatalf("expected 'version 1', got %q, err: %v", string(content), err)
	}

	// 2. Successful update
	err = SafeAtomicUpdate(target, func(old []byte) ([]byte, error) {
		return append(old, []byte(" -> version 2")...), nil
	})
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}

	content, _ = os.ReadFile(target)
	if string(content) != "version 1 -> version 2" {
		t.Fatalf("expected 'version 1 -> version 2', got %q", string(content))
	}
}

func TestOracleNegativeControls(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "important.json")
	initialContent := []byte(`{"critical":"data","do_not_lose":true}`)
	_ = os.WriteFile(target, initialContent, 0600)

	// Negative 1: updateFn fails
	testErr := errors.New("simulated network failure during update")
	err := SafeAtomicUpdate(target, func(old []byte) ([]byte, error) {
		return nil, testErr
	})
	if !errors.Is(err, testErr) {
		t.Fatalf("expected update error to be returned, got: %v", err)
	}

	// Verify original file is 100% byte-for-byte preserved
	currentContent, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("target file was deleted or unreadable: %v", err)
	}
	if !bytes.Equal(currentContent, initialContent) {
		t.Fatalf("CRITICAL: target file content was corrupted! got %q, want %q", currentContent, initialContent)
	}

	// Verify no stray temp files remain in directory
	entries, _ := os.ReadDir(tmpDir)
	if len(entries) != 1 || entries[0].Name() != "important.json" {
		t.Fatalf("stray temp files detected in dir: %v", entries)
	}
}
