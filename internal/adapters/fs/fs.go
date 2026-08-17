// Package fs provides the filesystem adapter for the application ports.
package fs

import (
	"os"
	"path/filepath"
)

// System is the production filesystem adapter.
type System struct{}

// New creates the production filesystem adapter.
func New() *System {
	return &System{}
}

// ReadFile reads the file at path.
func (s *System) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to path and creates missing parent directories.
func (s *System) WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
