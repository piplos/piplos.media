-- +goose Up
-- Пользовательские страницы (раздел «Статьи» на сайте, создаются в админке:
-- Страницы → Новая страница). translations: {"en": {"title": ..., "description": ...,
-- "body": ...}, "ru": {...}}. publish_at — отложенная публикация: страница видна
-- на сайте только при published = TRUE и publish_at <= now() (или NULL).
CREATE TABLE pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    publish_at TIMESTAMPTZ,
    translations JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS pages;
