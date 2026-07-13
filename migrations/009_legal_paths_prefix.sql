-- +goose Up
-- Legal pages moved under the /legal/ prefix on the site.
UPDATE legal_pages SET path = '/legal/' || slug, updated_at = now();

-- +goose Down
UPDATE legal_pages SET path = '/' || slug, updated_at = now();
