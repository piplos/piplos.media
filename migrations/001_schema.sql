-- +goose Up
-- Consolidated schema: all tables in their final shape plus deterministic
-- system configuration (languages, settings, AI provider models).
-- Content seeds live in 002_seed_catalog.sql and 003_seed_content.sql.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'manager' CHECK (role IN ('admin', 'manager')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Системные языки проекта. Контент хранит переводы по коду языка (JSONB).
CREATE TABLE languages (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0
);

INSERT INTO languages (code, name, is_default, enabled, sort_order) VALUES
    ('en', 'English', TRUE, TRUE, 0),
    ('ru', 'Русский', FALSE, TRUE, 1);

-- Портфолио. translations: {"en": {"title": ..., "subtitle": ..., "description": ...,
-- "challenge": ..., "solution": ..., "result": ..., "stack_detail": ...}, "ru": {...}}
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    category TEXT NOT NULL DEFAULT '',
    categories TEXT[] NOT NULL DEFAULT '{}',
    tags TEXT[] NOT NULL DEFAULT '{}',
    year INT NOT NULL DEFAULT 0,
    featured BOOLEAN NOT NULL DEFAULT FALSE,
    published BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    image TEXT NOT NULL DEFAULT '',
    translations JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Услуги. translations: {"en": {"title": ..., "description": ..., "body": ...}, ...}
CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    icon TEXT NOT NULL DEFAULT '',
    tags TEXT[] NOT NULL DEFAULT '{}',
    published BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    translations JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Технологический стек (названия технологий не переводятся).
-- group_id = slug услуги (services.slug). icon/icon_alt — пути в /uploads/stack/…
CREATE TABLE stack_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    group_id TEXT NOT NULL DEFAULT 'backend',
    published BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    icon TEXT NOT NULL DEFAULT '',
    icon_alt TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- SEO страниц. translations: {"en": {"title": ..., "description": ..., "keywords": ...,
-- "og_title": ..., "og_description": ..., "og_image": ...}, ...}
CREATE TABLE seo_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    path TEXT NOT NULL UNIQUE,
    translations JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Правовые документы (privacy, terms, cookies). Страницы noindex.
CREATE TABLE legal_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL UNIQUE,
    sort_order INT NOT NULL DEFAULT 0,
    translations JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Заявки с формы заказа на сайте.
CREATE TABLE leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    types TEXT[] NOT NULL DEFAULT '{}',
    project_name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    stack TEXT NOT NULL DEFAULT '',
    reference_urls TEXT NOT NULL DEFAULT '',
    budget INT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    timeline TEXT NOT NULL DEFAULT '',
    stage TEXT NOT NULL DEFAULT '',
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL,
    company TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    how_found TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    lang TEXT NOT NULL DEFAULT 'en',
    status TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'in_progress', 'done', 'spam')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_leads_status_created ON leads (status, created_at DESC);

-- Композитные настройки (luna-style): один ключ = JSON-объект.
-- Чувствительные поля (apiKey, SMTP.password) шифруются приложением (AES-256-GCM, enc:v1:).
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO settings (key, value) VALUES
    ('SMTP', '{"host":"","port":587,"username":"","password":"","from":"","timeout_seconds":30}'),
    ('OPENAI', '{"enable":true,"apiKey":"","rateLimit":10,"timeoutSeconds":120}'),
    ('GEMINI', '{"enable":false,"apiKey":"","rateLimit":10,"timeoutSeconds":60}'),
    ('GROK', '{"enable":false,"apiKey":"","rateLimit":10,"timeoutSeconds":30}'),
    ('OPENROUTER', '{"enable":false,"apiKey":"","rateLimit":10,"timeoutSeconds":30}'),
    ('AI_TRANSLATION', '{"provider":"openai","model":"gpt-4o-mini","prompt":"You are a professional translator for a software development agency website. Translate the values of the given JSON object into {target_language}. Keep the JSON keys unchanged, preserve technology names, brand names, numbers and formatting. Return only a valid JSON object."}');

-- Каталог AI-моделей провайдеров.
CREATE TABLE ai_provider_models (
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

-- +goose Down
DROP TABLE IF EXISTS ai_provider_models;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS leads;
DROP TABLE IF EXISTS legal_pages;
DROP TABLE IF EXISTS seo_pages;
DROP TABLE IF EXISTS stack_items;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS languages;
DROP TABLE IF EXISTS users;
