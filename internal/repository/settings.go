package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/site/internal/models"
	"github.com/piplos/site/internal/utils"
)

// sqlUpsertSetting upserts a settings row by key, refreshing updated_at on conflict.
const sqlUpsertSetting = `
	INSERT INTO settings (key, value, updated_at)
	VALUES ($1, $2, now())
	ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
`

func (r *Repository) encryptionKeyValid() bool {
	return len(r.encryptionKey) == 32
}

// ListSettings returns all settings. Values are raw (may contain enc:v1: prefixed secrets).
func (r *Repository) ListSettings(ctx context.Context) ([]models.Setting, error) {
	rows, err := r.pool.Query(ctx, "SELECT key, value, updated_at FROM settings ORDER BY key")
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	defer rows.Close()

	items := []models.Setting{}
	for rows.Next() {
		var s models.Setting
		if err := rows.Scan(&s.Key, &s.Value, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

// GetSetting returns the raw value for key ("" when missing). Key is normalized to uppercase.
func (r *Repository) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx, "SELECT value FROM settings WHERE key = $1", strings.ToUpper(key)).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return value, nil
}

// decryptValue decrypts a setting value using the unified enc:v1: mechanism.
// Standalone secrets (value starts with enc:v1:) are fully decrypted.
// Composite JSON (value contains enc:v1: inside) has its fields decrypted.
// Plain values are returned unchanged.
func (r *Repository) decryptValue(value string) (string, error) {
	if utils.HasEncPrefix(value) {
		if !r.encryptionKeyValid() {
			return "", fmt.Errorf("encryption key not set or invalid")
		}
		return utils.DecryptField(r.encryptionKey, value)
	}
	if strings.Contains(value, utils.EncPrefix) {
		if !r.encryptionKeyValid() {
			return "", fmt.Errorf("encryption key not set or invalid")
		}
		return utils.DecryptJSONFields(r.encryptionKey, value)
	}
	return value, nil
}

// GetDecryptedValue returns the plain value for a setting, decrypting enc:v1: fields if present.
func (r *Repository) GetDecryptedValue(ctx context.Context, key string) (string, error) {
	raw, err := r.GetSetting(ctx, key)
	if err != nil || raw == "" {
		return raw, err
	}
	return r.decryptValue(raw)
}

// normalizeMaskedSensitiveFields keeps old values for sensitive fields the client
// sent back masked ("****"), so a save without re-entering a secret preserves it.
func normalizeMaskedSensitiveFields(oldJSON, newJSON string, sensitiveFields []string) string {
	if len(sensitiveFields) == 0 {
		return newJSON
	}
	const maskedValue = "****"

	var oldM map[string]any
	if json.Unmarshal([]byte(oldJSON), &oldM) != nil {
		oldM = make(map[string]any)
	}
	var newM map[string]any
	if json.Unmarshal([]byte(newJSON), &newM) != nil {
		return newJSON
	}

	for _, f := range sensitiveFields {
		v, ok := newM[f]
		if !ok {
			continue
		}
		s, isStr := v.(string)
		if !isStr || s != maskedValue {
			continue
		}
		if ov, ok := oldM[f]; ok {
			newM[f] = ov
		} else {
			newM[f] = ""
		}
	}
	out, err := json.Marshal(newM)
	if err != nil {
		return newJSON
	}
	return string(out)
}

// SetCompositeSetting saves a composite JSON setting, encrypting the specified
// sensitive fields (enc:v1:, AES-256-GCM). Masked ("****") sensitive fields keep
// their stored values. Key is normalized to uppercase.
func (r *Repository) SetCompositeSetting(ctx context.Context, key, plainJSON string, sensitiveFields []string) error {
	key = strings.ToUpper(key)

	oldDecrypted, err := r.GetDecryptedValue(ctx, key)
	if err != nil {
		return fmt.Errorf("get current composite setting %s: %w", key, err)
	}
	plainJSON = normalizeMaskedSensitiveFields(oldDecrypted, plainJSON, sensitiveFields)
	if oldDecrypted == plainJSON {
		return nil
	}

	toStore := plainJSON
	if len(sensitiveFields) > 0 {
		if !r.encryptionKeyValid() {
			return fmt.Errorf("encryption key not set or invalid")
		}
		encrypted, encErr := utils.EncryptJSONFields(r.encryptionKey, plainJSON, sensitiveFields)
		if encErr != nil {
			return fmt.Errorf("encrypt composite setting: %w", encErr)
		}
		toStore = encrypted
	}

	if _, err := r.pool.Exec(ctx, sqlUpsertSetting, key, toStore); err != nil {
		return fmt.Errorf("set composite setting: %w", err)
	}
	return nil
}
