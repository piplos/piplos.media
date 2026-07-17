-- +goose Up
-- +goose StatementBegin

-- Технологический стек статьи (как у услуг): массив меток из каталога стека.
ALTER TABLE pages
ADD COLUMN tags TEXT[] NOT NULL DEFAULT '{}';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE pages DROP COLUMN IF EXISTS tags;

-- +goose StatementEnd
