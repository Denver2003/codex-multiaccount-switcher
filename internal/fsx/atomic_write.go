package fsx

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func EnsureDir(path string, perm os.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return fmt.Errorf("create directory %s: %w", path, err)
	}

	if err := os.Chmod(path, perm); err != nil {
		return fmt.Errorf("chmod directory %s: %w", path, err)
	}

	return nil
}

func AtomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir, 0o700); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", path, err)
	}

	tempPath := tempFile.Name()
	success := false
	defer func() {
		if !success {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("write temp file for %s: %w", path, err)
	}

	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("sync temp file for %s: %w", path, err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp file for %s: %w", path, err)
	}

	if err := os.Chmod(tempPath, perm); err != nil {
		return fmt.Errorf("chmod temp file for %s: %w", path, err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("rename temp file for %s: %w", path, err)
	}

	success = true
	return nil
}

func AtomicWriteJSON(path string, value any, perm os.FileMode) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON for %s: %w", path, err)
	}

	data = append(data, '\n')
	return AtomicWriteFile(path, data, perm)
}
