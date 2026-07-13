package utils

import (
	"encoding/json"
	"strings"
	"testing"
)

var testKey = []byte("0123456789abcdef0123456789abcdef")

func TestEncryptDecryptRoundtrip(t *testing.T) {
	enc, err := Encrypt(testKey, "secret-value")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	dec, err := Decrypt(testKey, enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if dec != "secret-value" {
		t.Fatalf("roundtrip mismatch: %q", dec)
	}
}

func TestEncryptionKeyFromString(t *testing.T) {
	if _, err := EncryptionKeyFromString("short"); err == nil {
		t.Fatal("expected error for short key")
	}
	key, err := EncryptionKeyFromString(strings.Repeat("x", 40))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key))
	}
}

func TestEncryptDecryptField(t *testing.T) {
	enc, err := EncryptField(testKey, "hunter2")
	if err != nil {
		t.Fatalf("EncryptField: %v", err)
	}
	if !HasEncPrefix(enc) {
		t.Fatalf("expected enc:v1: prefix, got %q", enc)
	}
	dec, err := DecryptField(testKey, enc)
	if err != nil {
		t.Fatalf("DecryptField: %v", err)
	}
	if dec != "hunter2" {
		t.Fatalf("field roundtrip mismatch: %q", dec)
	}
	// Plaintext passthrough.
	plain, err := DecryptField(testKey, "plain")
	if err != nil || plain != "plain" {
		t.Fatalf("plaintext passthrough failed: %q, %v", plain, err)
	}
}

func TestEncryptDecryptJSONFields(t *testing.T) {
	src := `{"host":"smtp.example.com","password":"hunter2","port":587}`
	enc, err := EncryptJSONFields(testKey, src, []string{"password"})
	if err != nil {
		t.Fatalf("EncryptJSONFields: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(enc), &m); err != nil {
		t.Fatalf("unmarshal encrypted: %v", err)
	}
	if !HasEncPrefix(m["password"].(string)) {
		t.Fatalf("password not encrypted: %v", m["password"])
	}
	if m["host"] != "smtp.example.com" {
		t.Fatalf("host must stay plaintext: %v", m["host"])
	}

	dec, err := DecryptJSONFields(testKey, enc)
	if err != nil {
		t.Fatalf("DecryptJSONFields: %v", err)
	}
	var dm map[string]any
	if err := json.Unmarshal([]byte(dec), &dm); err != nil {
		t.Fatalf("unmarshal decrypted: %v", err)
	}
	if dm["password"] != "hunter2" {
		t.Fatalf("password not decrypted: %v", dm["password"])
	}
}

func TestMaskJSONFields(t *testing.T) {
	src := `{"username":"user","password":"hunter2","host":"h","empty":""}`
	masked := MaskJSONFields(src, []string{"username", "password", "empty"}, "****")
	var m map[string]any
	if err := json.Unmarshal([]byte(masked), &m); err != nil {
		t.Fatalf("unmarshal masked: %v", err)
	}
	if m["username"] != "****" || m["password"] != "****" {
		t.Fatalf("fields not masked: %v", m)
	}
	if m["empty"] != "" {
		t.Fatalf("empty field must stay empty: %v", m["empty"])
	}
	if m["host"] != "h" {
		t.Fatalf("host must stay intact: %v", m["host"])
	}
}
