package rlm

import (
	"crypto/sha256"
	"encoding/hex"
)

// ContentKey returns the lowercase hex SHA-256 digest of data. Chunks and
// targets are referenced by this key so payloads stay out of the prompt.
func ContentKey(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
