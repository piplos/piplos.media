// Package config loads application configuration from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string

	JWTExpirationMinutes    int
	JWTRefreshExpirationHrs int

	CORSOrigins []string

	// Ключ шифрования секретов в таблице settings (AES-256-GCM, >= 32 байт).
	EncryptionKey string

	// Начальный администратор: создаётся при старте, если пользователей нет.
	AdminEmail    string
	AdminPassword string

	// Каталог загруженных файлов и публичный URL API для абсолютных ссылок на медиа.
	UploadDir    string
	PublicAPIURL string

	// Публичный URL админ-панели (ссылка на заявку в письме админам).
	AdminURL string
}

// JWTExpiration returns access token TTL.
func (c *Config) JWTExpiration() time.Duration {
	if c.JWTExpirationMinutes <= 0 {
		return 15 * time.Minute
	}
	return time.Duration(c.JWTExpirationMinutes) * time.Minute
}

// JWTRefreshExpiration returns refresh token TTL.
func (c *Config) JWTRefreshExpiration() time.Duration {
	if c.JWTRefreshExpirationHrs <= 0 {
		return 7 * 24 * time.Hour
	}
	return time.Duration(c.JWTRefreshExpirationHrs) * time.Hour
}

// Load reads configuration from the environment.
func Load() Config {
	return Config{
		Port:                    env("PORT", "3001"),
		DatabaseURL:             DatabaseURL(),
		JWTSecret:               env("JWT_SECRET", ""),
		JWTExpirationMinutes:    envInt("JWT_EXPIRATION_MINUTES", 15),
		JWTRefreshExpirationHrs: envInt("JWT_REFRESH_EXPIRATION_HOURS", 168),
		CORSOrigins:             envCSV("CORS_ORIGINS", []string{"*"}),
		EncryptionKey:           env("ENCRYPTION_KEY", ""),
		AdminEmail:              env("ADMIN_EMAIL", ""),
		AdminPassword:           env("ADMIN_PASSWORD", ""),
		UploadDir:               env("UPLOAD_DIR", "data/uploads"),
		PublicAPIURL:            env("PUBLIC_API_URL", ""),
		AdminURL:                env("ADMIN_URL", "http://localhost:5174"),
	}
}

// Validate checks required fields.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("database_url is required (set DATABASE_URL or POSTGRES_USER/POSTGRES_PASSWORD/POSTGRES_DB)")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("jwt_secret is required")
	}
	if len(c.EncryptionKey) < 32 {
		return fmt.Errorf("encryption_key is required (set ENCRYPTION_KEY, at least 32 bytes)")
	}
	return nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envCSV(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
