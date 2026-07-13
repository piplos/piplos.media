-- +goose Up
-- SEO paths for portfolio cases must use project slug, not UUID.
UPDATE seo_pages sp
SET path = '/portfolio/' || p.slug,
    updated_at = now()
FROM projects p
WHERE sp.path = '/portfolio/' || p.id::text;

-- +goose Down
UPDATE seo_pages sp
SET path = '/portfolio/' || p.id::text,
    updated_at = now()
FROM projects p
WHERE sp.path = '/portfolio/' || p.slug;
