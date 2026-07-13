-- +goose Up
-- Услуги с главной страницы сайта (ru + en). Источник: web/site/src/lib/i18n/*.json
-- Перегенерация: bun scripts/generate-services-seed.mjs

INSERT INTO services (slug, icon, tags, published, sort_order, translations)
VALUES ('web', '⬡', ARRAY['React', 'Svelte', 'Vue', 'Next.js'], true, 0, $service_web_json${"en":{"title":"Web Applications","description":"SPAs, dashboards, admin panels and enterprise platforms. React, Svelte, Vue and Next.js with TypeScript — from Figma prototypes to production UI."},"ru":{"title":"Веб-приложения","description":"SPA, дашборды, админ-панели и корпоративные платформы. React, Svelte, Vue и Next.js на TypeScript — от прототипа в Figma до продакшн-интерфейса."}}$service_web_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    icon = EXCLUDED.icon,
    tags = EXCLUDED.tags,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO services (slug, icon, tags, published, sort_order, translations)
VALUES ('mobile', '◈', ARRAY['Flutter', 'Swift', 'iOS', 'Android'], true, 1, $service_mobile_json${"en":{"title":"Mobile Apps","description":"Cross-platform apps with Flutter and native iOS with Swift. Performance-first UX, offline-ready flows and full App Store deployment."},"ru":{"title":"Мобильные приложения","description":"Кроссплатформенные приложения на Flutter и нативный iOS на Swift. Производительный UX, офлайн-сценарии и полный деплой в App Store."}}$service_mobile_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    icon = EXCLUDED.icon,
    tags = EXCLUDED.tags,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO services (slug, icon, tags, published, sort_order, translations)
VALUES ('backend', '⬢', ARRAY['Node.js', 'Go', 'Python', 'GraphQL'], true, 2, $service_backend_json${"en":{"title":"Backend & APIs","description":"Microservices, REST and GraphQL APIs, real-time systems and integrations. Node.js, Go, Python, Java and Rust — built to scale from day one."},"ru":{"title":"Бэкенд и API","description":"Микросервисы, REST и GraphQL API, real-time и интеграции. Node.js, Go, Python, Java и Rust — архитектура, готовая к росту с первого дня."}}$service_backend_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    icon = EXCLUDED.icon,
    tags = EXCLUDED.tags,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO services (slug, icon, tags, published, sort_order, translations)
VALUES ('data', '◫', ARRAY['PostgreSQL', 'ClickHouse', 'Redis', 'MySQL'], true, 3, $service_data_json${"en":{"title":"Data & Analytics","description":"Database design, query optimization, caching layers and analytics pipelines. OLTP and OLAP workloads with PostgreSQL, ClickHouse and Redis."},"ru":{"title":"Данные и аналитика","description":"Проектирование БД, оптимизация запросов, кэширование и аналитические пайплайны. OLTP и OLAP на PostgreSQL, ClickHouse и Redis."}}$service_data_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    icon = EXCLUDED.icon,
    tags = EXCLUDED.tags,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO services (slug, icon, tags, published, sort_order, translations)
VALUES ('devops', '⬟', ARRAY['Docker', 'Kubernetes', 'AWS', 'GitHub Actions'], true, 4, $service_devops_json${"en":{"title":"DevOps & Cloud","description":"CI/CD pipelines, container orchestration, cloud infrastructure and monitoring. Docker, Kubernetes, AWS and Terraform with GitHub Actions."},"ru":{"title":"DevOps и облако","description":"CI/CD, оркестрация контейнеров, облачная инфраструктура и мониторинг. Docker, Kubernetes, AWS и Terraform с GitHub Actions."}}$service_devops_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    icon = EXCLUDED.icon,
    tags = EXCLUDED.tags,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

-- +goose Down
DELETE FROM services WHERE slug IN ('web', 'mobile', 'backend', 'data', 'devops');
