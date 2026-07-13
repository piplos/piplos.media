// Package utils contains small shared helpers (encryption of settings values).
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
)

const aesKeyLen = 32

// EncPrefix marks a value as encrypted with AES-256-GCM (version 1).
const EncPrefix = "enc:v1:"

// Encrypt encrypts plaintext with AES-256-GCM. key must be exactly 32 bytes.
func Encrypt(key []byte, plaintext string) (string, error) {
	if len(key) != aesKeyLen {
		return "", errors.New("encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext. key must be exactly 32 bytes.
func Decrypt(key []byte, ciphertextB64 string) (string, error) {
	if len(key) != aesKeyLen {
		return "", errors.New("encryption key must be 32 bytes")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptionKeyFromString returns the first 32 bytes of s as the encryption key.
// Returns an error if s has fewer than 32 bytes (weak key in production).
func EncryptionKeyFromString(s string) ([]byte, error) {
	b := []byte(s)
	if len(b) < aesKeyLen {
		return nil, errors.New("encryption key must be at least 32 bytes")
	}
	return b[:aesKeyLen], nil
}

// HasEncPrefix reports whether value starts with the enc:v1: prefix.
func HasEncPrefix(value string) bool {
	return strings.HasPrefix(value, EncPrefix)
}

// EncryptField encrypts plaintext and returns it with the enc:v1: prefix.
// Empty plaintext is returned as-is (no encryption needed).
func EncryptField(key []byte, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	enc, err := Encrypt(key, plaintext)
	if err != nil {
		return "", err
	}
	return EncPrefix + enc, nil
}

// DecryptField decrypts a value with enc:v1: prefix.
// Values without the prefix are returned unchanged (plaintext passthrough).
func DecryptField(key []byte, value string) (string, error) {
	if !HasEncPrefix(value) {
		return value, nil
	}
	return Decrypt(key, value[len(EncPrefix):])
}

// DecryptJSONFields parses a JSON object string, decrypts any first-level string
// values that carry the enc:v1: prefix, and returns the modified JSON string.
func DecryptJSONFields(key []byte, jsonStr string) (string, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return "", err
	}
	for k, v := range m {
		s, ok := v.(string)
		if !ok || !HasEncPrefix(s) {
			continue
		}
		dec, err := Decrypt(key, s[len(EncPrefix):])
		if err != nil {
			return "", err
		}
		m[k] = dec
	}
	out, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// EncryptJSONFields parses a JSON object string, encrypts the specified fields
// (non-empty, without enc:v1: prefix), and returns the modified JSON string.
func EncryptJSONFields(key []byte, jsonStr string, fields []string) (string, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return "", err
	}
	fieldSet := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		fieldSet[f] = struct{}{}
	}
	for k, v := range m {
		if _, ok := fieldSet[k]; !ok {
			continue
		}
		s, ok := v.(string)
		if !ok || s == "" || HasEncPrefix(s) {
			continue
		}
		enc, err := EncryptField(key, s)
		if err != nil {
			return "", err
		}
		m[k] = enc
	}
	out, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// MaskJSONFields replaces the specified fields in a JSON string with mask.
// Non-empty string fields are replaced; other fields and types are left intact.
// Returns mask if the input is not valid JSON.
func MaskJSONFields(jsonStr string, fields []string, mask string) string {
	var m map[string]any
	if json.Unmarshal([]byte(jsonStr), &m) != nil {
		return mask
	}
	for _, f := range fields {
		if v, ok := m[f]; ok {
			if s, isStr := v.(string); isStr && s != "" {
				m[f] = mask
			}
		}
	}
	out, err := json.Marshal(m)
	if err != nil {
		return mask
	}
	return string(out)
}
