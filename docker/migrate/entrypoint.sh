#!/usr/bin/env sh

# goose runner for the migration image. Resolves the Postgres DSN the same way
# scripts/migrate does — DATABASE_URL → DSN → a DSN built from POSTGRES_* — then
# execs goose against the bundled /migrations. With no arguments it applies all
# pending migrations (`up`); pass any goose command to override (status, down,
# down-to, version, …).
set -e

dsn="${DATABASE_URL:-${DSN:-}}"
if [ -z "${dsn}" ]; then
    dsn="postgres://${POSTGRES_USER:-postgres}:${POSTGRES_PASSWORD:-postgres}@${POSTGRES_HOST:-db}:${POSTGRES_PORT:-5432}/${POSTGRES_DB:-piplos}?sslmode=disable"
fi

export GOOSE_DRIVER="${GOOSE_DRIVER:-postgres}"
export GOOSE_DBSTRING="${dsn}"
export GOOSE_MIGRATION_DIR="${GOOSE_MIGRATION_DIR:-/migrations}"

# No command given (e.g. `docker run --entrypoint`) → apply all pending.
if [ "$#" -eq 0 ]; then
    set -- up
fi

exec goose "$@"
