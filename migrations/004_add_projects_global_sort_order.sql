-- +goose Up
-- +goose StatementBegin

-- Сквозной порядок проектов на сайте (раздел «все проекты»).
-- Не зависит от sort_order, который задаёт порядок внутри группы (услуги).
ALTER TABLE projects
ADD COLUMN global_sort_order INT NOT NULL DEFAULT 0;

-- Бэкфилл: текущий видимый порядок «всех проектов» (sort_order, год, дата создания).
UPDATE projects
SET global_sort_order = ranked.rn
FROM (
    SELECT id, row_number() OVER (ORDER BY sort_order, year DESC, created_at DESC) - 1 AS rn
    FROM projects
) AS ranked
WHERE projects.id = ranked.id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE projects DROP COLUMN IF EXISTS global_sort_order;

-- +goose StatementEnd
