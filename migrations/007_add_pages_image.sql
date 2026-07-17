-- +goose Up
-- +goose StatementBegin

-- Превью статьи: фоновое изображение в карточке списка (как у проектов в портфолио).
ALTER TABLE pages
ADD COLUMN image TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE pages DROP COLUMN IF EXISTS image;

-- +goose StatementEnd
