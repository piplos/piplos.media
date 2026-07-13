-- +goose Up
-- group_id в stack_items = slug услуги (services.slug).
UPDATE stack_items SET group_id = 'web', updated_at = now() WHERE group_id = 'frontend';

-- +goose Down
UPDATE stack_items SET group_id = 'frontend', updated_at = now() WHERE group_id = 'web';
