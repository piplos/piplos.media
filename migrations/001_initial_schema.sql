-- +goose Up
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
    translations JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Услуги. translations: {"en": {"title": ..., "description": ...}, ...}
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
CREATE TABLE stack_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    group_id TEXT NOT NULL DEFAULT 'backend',
    published BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
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

-- Настройки (в т.ч. AI-переводчик).
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO settings (key, value) VALUES
    ('ai.base_url', 'https://api.openai.com/v1/chat/completions'),
    ('ai.api_key', ''),
    ('ai.model', 'gpt-4o-mini'),
    ('ai.translate_prompt', 'You are a professional translator for a software development agency website. Translate the values of the given JSON object into {target_language}. Keep the JSON keys unchanged, preserve technology names, brand names, numbers and formatting. Return only a valid JSON object.');

-- +goose Down
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS leads;
DROP TABLE IF EXISTS seo_pages;
DROP TABLE IF EXISTS stack_items;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS languages;
DROP TABLE IF EXISTS users;
