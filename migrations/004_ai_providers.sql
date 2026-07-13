-- AI providers (как в luna): модели в каталоге, ключи провайдеров и задача перевода.

-- +goose Up
CREATE TABLE IF NOT EXISTS ai_provider_models (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider     TEXT NOT NULL,
    model_id     TEXT NOT NULL,
    display_name TEXT NOT NULL,
    enabled      BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, model_id)
);

INSERT INTO ai_provider_models (provider, model_id, display_name) VALUES
    ('gemini', 'gemini-2.5-flash', 'Gemini 2.5 Flash'),
    ('gemini', 'gemini-2.5-flash-lite', 'Gemini 2.5 Flash Lite'),
    ('grok', 'grok-4-fast-non-reasoning', 'Grok 4 Fast'),
    ('openai', 'gpt-4o-mini', 'GPT-4o Mini'),
    ('openrouter', 'deepseek/deepseek-v3.2', 'DeepSeek V3.2'),
    ('openrouter', 'minimax/minimax-m2.5', 'MiniMax M2.5')
ON CONFLICT (provider, model_id) DO NOTHING;

INSERT INTO settings (key, value)
SELECT 'OPENAI', json_build_object(
    'enable', true,
    'apiKey', COALESCE(ai.value::json->>'api_key', ''),
    'rateLimit', 10,
    'timeoutSeconds', 120
)::text
FROM settings ai
WHERE ai.key = 'AI'
  AND NOT EXISTS (SELECT 1 FROM settings WHERE key = 'OPENAI');

INSERT INTO settings (key, value)
SELECT 'AI_TRANSLATION', json_build_object(
    'provider', 'openai',
    'model', COALESCE(ai.value::json->>'model', 'gpt-4o-mini'),
    'prompt', COALESCE(
        NULLIF(TRIM(ai.value::json->>'prompt'), ''),
        'You are a professional translator for a software development agency website. Translate the values of the given JSON object into {target_language}. Keep the JSON keys unchanged, preserve technology names, brand names, numbers and formatting. Return only a valid JSON object.'
    )
)::text
FROM settings ai
WHERE ai.key = 'AI'
  AND NOT EXISTS (SELECT 1 FROM settings WHERE key = 'AI_TRANSLATION');

INSERT INTO settings (key, value) VALUES
    ('GEMINI', '{"enable":false,"apiKey":"","rateLimit":10,"timeoutSeconds":60}'),
    ('GROK', '{"enable":false,"apiKey":"","rateLimit":10,"timeoutSeconds":30}'),
    ('OPENROUTER', '{"enable":false,"apiKey":"","rateLimit":10,"timeoutSeconds":30}')
ON CONFLICT (key) DO NOTHING;

DELETE FROM settings WHERE key = 'AI';

-- +goose Down
INSERT INTO settings (key, value)
SELECT 'AI', json_build_object(
    'base_url', 'https://api.openai.com/v1/chat/completions',
    'api_key', COALESCE(o.value::json->>'apiKey', ''),
    'model', COALESCE(t.value::json->>'model', 'gpt-4o-mini'),
    'prompt', COALESCE(t.value::json->>'prompt', '')
)::text
FROM settings o
LEFT JOIN settings t ON t.key = 'AI_TRANSLATION'
WHERE o.key = 'OPENAI'
ON CONFLICT (key) DO NOTHING;

DELETE FROM settings WHERE key IN ('GEMINI', 'GROK', 'OPENROUTER', 'OPENAI', 'AI_TRANSLATION');
DROP TABLE IF EXISTS ai_provider_models;
