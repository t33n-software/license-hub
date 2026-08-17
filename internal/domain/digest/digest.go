// Package digest computes canonical SHA-256 digests for template pinning.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 returns the sha256:<hex> digest of data.
func SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
