package config

import (
	"net"
	"net/url"
	"os"
)

// DatabaseURL returns DATABASE_URL when set, otherwise a DSN built from POSTGRES_*
// (host-side dev: go run, IDE). Empty when neither DATABASE_URL nor POSTGRES_USER
// is configured. Production compose sets DATABASE_URL explicitly (internal `db` host).
func DatabaseURL() string {
	if v := env("DATABASE_URL", ""); v != "" {
		return v
	}
	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return ""
	}
	pass := os.Getenv("POSTGRES_PASSWORD")
	dbName := env("POSTGRES_DB", "piplos")
	port := env("POSTGRES_PORT", "5432")
	host := env("POSTGRES_HOST", "localhost")

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + dbName,
	}
	q := u.Query()
	q.Set("sslmode", env("POSTGRES_SSLMODE", "disable"))
	u.RawQuery = q.Encode()
	return u.String()
}
