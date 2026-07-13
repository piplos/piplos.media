-- +goose Up
-- Legal documents (privacy, terms, cookies). SEO is not managed separately; pages are noindex.
CREATE TABLE legal_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    translations JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO legal_pages (slug, path, sort_order) VALUES
    ('privacy', '/privacy', 0),
    ('terms', '/terms', 1),
    ('cookies', '/cookies', 2);

-- +goose Down
DROP TABLE IF EXISTS legal_pages;
