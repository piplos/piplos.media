-- +goose Up
-- +goose StatementBegin

-- Персональный флаг получения писем о новых заявках.
-- Меняется только администратором в настройках пользователей.
ALTER TABLE users
ADD COLUMN notify_leads BOOLEAN NOT NULL DEFAULT TRUE;

-- Шаблон письма о новой заявке (редактируется в админке: SMTP → Шаблон письма).
-- Body — Markdown (отправляется как HTML); переменные вида {{project_name}}
-- заменяются данными заявки.
INSERT INTO settings (key, value)
VALUES (
    'LEAD_EMAIL_TEMPLATE',
    json_build_object(
        'subject', '{{name}} — новая заявка ({{project_name}})',
        'body', E'## {{name}}\n{{email}} · {{phone}}  \n{{company}}\n\n---\n\n### Запрос\n\n| | |\n| --- | --- |\n| **Проект** | {{project_name}} |\n| **Тип** | {{types}} |\n| **Бюджет** | {{budget}} |\n| **Сроки** | {{timeline}} |\n| **Стадия** | {{stage}} |\n\n**Описание**  \n{{description}}\n\n**Стек:** {{stack}}  \n**Референсы:** {{references}}\n\n**Как нашли нас:** {{how_found}}  \n**Примечания:** {{notes}}\n\n_Заявка № {{id}} · {{created_at}} · {{lang}}_\n\n[Открыть в админке →]({{lead_url}})'
    )::text
)
ON CONFLICT (key) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users DROP COLUMN IF EXISTS notify_leads;
DELETE FROM settings WHERE key = 'LEAD_EMAIL_TEMPLATE';

-- +goose StatementEnd
