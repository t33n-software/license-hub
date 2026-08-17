// Package lockfile models the tenant-side template pin.
package lockfile

import (
	"encoding/json"
	"fmt"
)

// Lock pins a canonical template by path, version, and digest.
type Lock struct {
	Template string `json:"template"`
	Version  string `json:"version"`
	Digest   string `json:"digest"`
}

// Parse decodes and validates a lock file.
func Parse(data []byte) (Lock, error) {
	var lock Lock
	if err := json.Unmarshal(data, &lock); err != nil {
		return Lock{}, fmt.Errorf("parse lock file: %w", err)
	}
	if lock.Template == "" || lock.Version == "" || lock.Digest == "" {
		return Lock{}, fmt.Errorf("lock file requires template, version and digest")
	}
	return lock, nil
}
