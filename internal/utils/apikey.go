package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// apiKeyLivePrefix is the human-recognizable prefix of every agent API key.
const apiKeyLivePrefix = "pk_live_"

// GenerateAPIKey returns (rawKey, keyHash, keyPrefix). rawKey is shown to the
// admin exactly once at creation; keyHash is the SHA-256 hex digest stored in
// the DB; keyPrefix (first 16 chars of rawKey) is kept for display in lists.
func GenerateAPIKey() (raw, hash, prefix string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", "", fmt.Errorf("generate api key: %w", err)
	}
	raw = apiKeyLivePrefix + hex.EncodeToString(buf)
	return raw, HashAPIKey(raw), raw[:16], nil
}

// HashAPIKey returns the SHA-256 hex digest of a raw API key.
func HashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
