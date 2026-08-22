package mailer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/piplos/piplos.media/internal/config"
)

// SMTPConfig holds outbound mail server settings.
type SMTPConfig struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	From           string `json:"from"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

// Timeout returns the dial/send timeout.
func (c SMTPConfig) Timeout() time.Duration {
	if c.TimeoutSeconds <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// Ready reports whether outbound mail can be sent.
func (c SMTPConfig) Ready() bool {
	return c.Host != "" && c.From != ""
}

type smtpLoader interface {
	GetDecryptedValue(ctx context.Context, key string) (string, error)
}

// LoadSMTP reads and decrypts SMTP settings from the database.
// An unset setting ("") yields a zero config, not an error: callers use
// Ready() to detect the "not configured" state.
func LoadSMTP(ctx context.Context, repo smtpLoader) (SMTPConfig, error) {
	raw, err := repo.GetDecryptedValue(ctx, config.KeySMTP)
	if err != nil {
		return SMTPConfig{}, fmt.Errorf("load smtp settings: %w", err)
	}
	var cfg SMTPConfig
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
			return SMTPConfig{}, fmt.Errorf("parse smtp settings: %w", err)
		}
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	return cfg, nil
}
