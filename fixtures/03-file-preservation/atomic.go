package filepreservation

import (
	"fmt"
	"os"
	"path/filepath"
)

func SafeAtomicUpdate(targetPath string, updateFn func(old []byte) ([]byte, error)) error {
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	var oldData []byte
	var perm os.FileMode = 0644

	if fi, err := os.Stat(targetPath); err == nil {
		perm = fi.Mode().Perm()
		data, err := os.ReadFile(targetPath)
		if err != nil {
			return fmt.Errorf("read old file: %w", err)
		}
		oldData = data
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat target: %w", err)
	}

	newData, err := updateFn(oldData)
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".atomic-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if err := tmpFile.Chmod(perm); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("chmod temp: %w", err)
	}

	if _, err := tmpFile.Write(newData); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("sync temp: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("rename temp to target: %w", err)
	}

	return nil
}
