-- Композитные настройки в стиле luna: один ключ = JSON-объект.
-- Чувствительные поля (AI.api_key, SMTP.username/password) шифруются
-- приложением при сохранении (AES-256-GCM, префикс enc:v1:).

-- +goose Up
-- +goose StatementBegin
INSERT INTO settings (key, value)
SELECT 'AI', json_build_object(
    'base_url', COALESCE((SELECT value FROM settings WHERE key = 'ai.base_url'), 'https://api.openai.com/v1/chat/completions'),
    'api_key', COALESCE((SELECT value FROM settings WHERE key = 'ai.api_key'), ''),
    'model', COALESCE((SELECT value FROM settings WHERE key = 'ai.model'), 'gpt-4o-mini'),
    'prompt', COALESCE((SELECT value FROM settings WHERE key = 'ai.translate_prompt'),
        'You are a professional translator for a software development agency website. Translate the values of the given JSON object into {target_language}. Keep the JSON keys unchanged, preserve technology names, brand names, numbers and formatting. Return only a valid JSON object.')
)::text
ON CONFLICT (key) DO NOTHING;
-- +goose StatementEnd

INSERT INTO settings (key, value)
VALUES ('SMTP', '{"host":"","port":587,"username":"","password":"","from":"","timeout_seconds":30}')
ON CONFLICT (key) DO NOTHING;

DELETE FROM settings WHERE key IN ('ai.base_url', 'ai.api_key', 'ai.model', 'ai.translate_prompt');

-- +goose Down
-- +goose StatementBegin
INSERT INTO settings (key, value)
SELECT kv.key, COALESCE(s.value::json ->> kv.field, '')
FROM (VALUES
    ('ai.base_url', 'base_url'),
    ('ai.api_key', 'api_key'),
    ('ai.model', 'model'),
    ('ai.translate_prompt', 'prompt')
) AS kv(key, field)
LEFT JOIN settings s ON s.key = 'AI'
ON CONFLICT (key) DO NOTHING;
-- +goose StatementEnd

DELETE FROM settings WHERE key IN ('AI', 'SMTP');
