package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateAPIKeyFormat(t *testing.T) {
	raw, hash, prefix, err := GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(raw, "pk_live_") {
		t.Fatalf("raw key %q must start with pk_live_", raw)
	}
	body := strings.TrimPrefix(raw, "pk_live_")
	if len(body) != 64 {
		t.Fatalf("raw key body length = %d, want 64 hex chars", len(body))
	}
	if _, err := hex.DecodeString(body); err != nil {
		t.Fatalf("key body is not hex: %v", err)
	}
	if len(prefix) != 16 || prefix != raw[:16] {
		t.Fatalf("prefix = %q, want first 16 chars of raw", prefix)
	}
	sum := sha256.Sum256([]byte(raw))
	if hash != hex.EncodeToString(sum[:]) {
		t.Fatal("hash is not SHA-256 of raw key")
	}
}

func TestGenerateAPIKeyUnique(t *testing.T) {
	raw1, _, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	raw2, _, _, err := GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if raw1 == raw2 {
		t.Fatal("two generated keys must differ")
	}
}

func TestHashAPIKey(t *testing.T) {
	// RFC 4231 test vector: SHA-256 of "abc".
	got := HashAPIKey("abc")
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Fatalf("HashAPIKey(%q) = %q, want %q", "abc", got, want)
	}
}
