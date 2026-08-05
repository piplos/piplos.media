-- +goose Up
-- +goose StatementBegin

-- Структурированные ссылки проекта (сайт, App Store, Google Play).
-- Формат: [{"url":"...","label":"...","kind":"website|google_play|app_store"}]
ALTER TABLE projects
ADD COLUMN IF NOT EXISTS links JSONB NOT NULL DEFAULT '[]'::jsonb;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE projects DROP COLUMN IF EXISTS links;

-- +goose StatementEnd
