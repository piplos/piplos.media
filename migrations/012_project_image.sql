-- +goose Up
-- Превью проекта (URL картинки) для карточек в списке портфолио.
ALTER TABLE projects ADD COLUMN image TEXT NOT NULL DEFAULT '';

-- Превью для существующих проектов: иллюстрации из web/site/static/illustrations.
UPDATE projects SET image = '/illustrations/cat-analytics.png', updated_at = now() WHERE slug = 'analytics-dashboard' AND image = '';
UPDATE projects SET image = '/illustrations/cat-mobile.png', updated_at = now() WHERE slug = 'fintech-wallet' AND image = '';
UPDATE projects SET image = '/illustrations/cat-web.png', updated_at = now() WHERE slug = 'marketplace-platform' AND image = '';
UPDATE projects SET image = '/illustrations/cat-contact.png', updated_at = now() WHERE slug = 'hr-suite' AND image = '';
UPDATE projects SET image = '/illustrations/cat-launch.png', updated_at = now() WHERE slug = 'fleet-tracker' AND image = '';
UPDATE projects SET image = '/illustrations/cat-devops.png', updated_at = now() WHERE slug = 'cicd-panel' AND image = '';

-- +goose Down
ALTER TABLE projects DROP COLUMN image;
