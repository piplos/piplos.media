-- +goose Up
-- Портфолио (ru + en). Данные управляются через админку / API.

INSERT INTO projects (slug, category, categories, tags, year, featured, published, sort_order, translations)
VALUES ('analytics-dashboard', 'saas', ARRAY['saas', 'web'], ARRAY['React', 'Node.js', 'PostgreSQL', 'Redis'], 2024, true, true, 0, $project_analytics_dashboard_json${"en":{"title":"Analytics Dashboard","subtitle":"SaaS Platform","description":"Real-time data visualization platform processing 10M+ events per day with sub-100ms query response. Scaled from 0 to 50k MAU in 8 months.","challenge":"The client needed a platform capable of ingesting millions of events per second from IoT devices and presenting actionable insights with near-zero latency.","solution":"We designed a distributed event pipeline using Kafka and ClickHouse, paired with a React-based dashboard with WebSocket-driven live updates.","result":"50k MAU in 8 months, 99.97% uptime, sub-100ms p95 query latency across all dashboards.","stack_detail":"React 18, TypeScript, Node.js, Kafka, ClickHouse, PostgreSQL, Redis, Docker, Kubernetes, AWS EKS"},"ru":{"title":"Аналитический Дашборд","subtitle":"SaaS Платформа","description":"Платформа визуализации данных в реальном времени, обрабатывающая 10M+ событий в день с временем ответа менее 100мс. Масштабирована с 0 до 50k MAU за 8 месяцев.","challenge":"Клиенту требовалась платформа, способная принимать миллионы событий в секунду от IoT-устройств и предоставлять аналитику с минимальной задержкой.","solution":"Мы разработали распределённый событийный пайплайн на Kafka и ClickHouse с React-дашбордом на WebSocket-соединениях.","result":"50k MAU за 8 месяцев, аптайм 99.97%, задержка p95 менее 100мс по всем дашбордам.","stack_detail":"React 18, TypeScript, Node.js, Kafka, ClickHouse, PostgreSQL, Redis, Docker, Kubernetes, AWS EKS"}}$project_analytics_dashboard_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    category = EXCLUDED.category,
    categories = EXCLUDED.categories,
    tags = EXCLUDED.tags,
    year = EXCLUDED.year,
    featured = EXCLUDED.featured,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO projects (slug, category, categories, tags, year, featured, published, sort_order, translations)
VALUES ('fintech-wallet', 'mobile', ARRAY['mobile', 'fintech'], ARRAY['React Native', 'Expo', 'Node.js', 'PostgreSQL'], 2024, true, true, 1, $project_fintech_wallet_json${"en":{"title":"Fintech Wallet","subtitle":"Mobile · Fintech","description":"Cross-platform mobile banking app with biometric auth, instant P2P transfers and AI-powered spending insights. 4.8★ on App Store.","challenge":"Building a PCI-DSS compliant mobile wallet that feels as fast as cash while supporting complex multi-currency P2P transfers.","solution":"React Native with Expo for cross-platform delivery, custom biometric auth flow, and a microservices backend with event sourcing for transaction integrity.","result":"4.8★ App Store rating, 200k downloads in 3 months, zero security incidents post-launch.","stack_detail":"React Native, Expo, TypeScript, Node.js, PostgreSQL, Redis, Stripe, AWS Lambda"},"ru":{"title":"Финтех Кошелёк","subtitle":"Мобильное · Финтех","description":"Кроссплатформенное мобильное банковское приложение с биометрической аутентификацией, мгновенными P2P-переводами и AI-аналитикой расходов. Рейтинг 4.8★ в App Store.","challenge":"Создание PCI-DSS совместимого мобильного кошелька с ощущением скорости наличных и поддержкой сложных мультивалютных P2P-переводов.","solution":"React Native с Expo для кроссплатформенной разработки, кастомный биометрический флоу и микросервисный бэкенд с event sourcing.","result":"Рейтинг 4.8★ в App Store, 200k загрузок за 3 месяца, ноль инцидентов безопасности после запуска.","stack_detail":"React Native, Expo, TypeScript, Node.js, PostgreSQL, Redis, Stripe, AWS Lambda"}}$project_fintech_wallet_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    category = EXCLUDED.category,
    categories = EXCLUDED.categories,
    tags = EXCLUDED.tags,
    year = EXCLUDED.year,
    featured = EXCLUDED.featured,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO projects (slug, category, categories, tags, year, featured, published, sort_order, translations)
VALUES ('marketplace-platform', 'ecommerce', ARRAY['ecommerce', 'web'], ARRAY['Next.js', 'TypeScript', 'PostgreSQL', 'Elasticsearch'], 2023, true, true, 2, $project_marketplace_platform_json${"en":{"title":"Marketplace Platform","subtitle":"E-commerce","description":"Multi-vendor marketplace with AI-powered recommendations, real-time inventory and 500+ active sellers at launch.","challenge":"Launching a marketplace from scratch with complex vendor onboarding, real-time inventory sync, and personalised search.","solution":"Next.js SSR storefront, Elasticsearch for product discovery, event-driven inventory management, and a custom ML recommendation engine.","result":"500+ sellers at launch, 2M+ product listings, 35% conversion lift from AI recommendations.","stack_detail":"Next.js, TypeScript, PostgreSQL, Elasticsearch, Redis, Python (ML), AWS, Stripe Connect"},"ru":{"title":"Маркетплейс","subtitle":"Электронная коммерция","description":"Мультивендорный маркетплейс с AI-рекомендациями, синхронизацией инвентаря в реальном времени и 500+ продавцами на старте.","challenge":"Запуск маркетплейса с нуля: сложный онбординг продавцов, синхронизация инвентаря в реальном времени и персонализированный поиск.","solution":"SSR-витрина на Next.js, Elasticsearch для поиска товаров, событийное управление инвентарём и кастомный ML-движок рекомендаций.","result":"500+ продавцов на старте, 2M+ товарных позиций, рост конверсии на 35% благодаря AI-рекомендациям.","stack_detail":"Next.js, TypeScript, PostgreSQL, Elasticsearch, Redis, Python (ML), AWS, Stripe Connect"}}$project_marketplace_platform_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    category = EXCLUDED.category,
    categories = EXCLUDED.categories,
    tags = EXCLUDED.tags,
    year = EXCLUDED.year,
    featured = EXCLUDED.featured,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO projects (slug, category, categories, tags, year, featured, published, sort_order, translations)
VALUES ('hr-suite', 'web', ARRAY['web', 'saas'], ARRAY['Vue', 'Node.js', 'PostgreSQL', 'Docker'], 2023, false, true, 3, $project_hr_suite_json${"en":{"title":"HR Suite","subtitle":"Enterprise Web App","description":"End-to-end HR management platform covering recruitment, onboarding, performance reviews and payroll for 5,000+ employees.","challenge":"Replacing a legacy HRIS with a modern platform that integrates with 12 existing enterprise systems without downtime.","solution":"Modular Vue.js SPA with a Node.js API gateway, event-driven integrations, and a phased migration strategy.","result":"Zero-downtime migration, 60% reduction in HR admin time, adopted by 5,000+ employees on day one.","stack_detail":"Vue 3, TypeScript, Node.js, PostgreSQL, RabbitMQ, Docker, Kubernetes"},"ru":{"title":"HR Suite","subtitle":"Корпоративное Веб-приложение","description":"Комплексная HR-платформа: рекрутинг, онбординг, оценка эффективности и расчёт зарплат для 5000+ сотрудников.","challenge":"Замена устаревшей HRIS на современную платформу с интеграцией 12 существующих корпоративных систем без простоев.","solution":"Модульное SPA на Vue.js с API-шлюзом на Node.js, событийными интеграциями и поэтапной стратегией миграции.","result":"Миграция без простоев, сокращение времени HR-администрирования на 60%, 5000+ сотрудников с первого дня.","stack_detail":"Vue 3, TypeScript, Node.js, PostgreSQL, RabbitMQ, Docker, Kubernetes"}}$project_hr_suite_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    category = EXCLUDED.category,
    categories = EXCLUDED.categories,
    tags = EXCLUDED.tags,
    year = EXCLUDED.year,
    featured = EXCLUDED.featured,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO projects (slug, category, categories, tags, year, featured, published, sort_order, translations)
VALUES ('fleet-tracker', 'saas', ARRAY['saas', 'mobile'], ARRAY['React', 'Go', 'PostgreSQL', 'WebSocket'], 2023, false, true, 4, $project_fleet_tracker_json${"en":{"title":"Fleet Tracker","subtitle":"IoT · SaaS","description":"Real-time fleet management system tracking 10,000+ vehicles with live GPS, route optimisation and predictive maintenance alerts.","challenge":"Processing high-frequency GPS telemetry from 10k+ vehicles simultaneously while keeping the map UI responsive.","solution":"Go-based telemetry ingestion service, WebSocket fan-out for live map updates, and a React frontend with canvas-based rendering.","result":"10k+ vehicles tracked in real time, 40% fuel cost reduction via route optimisation, 99.9% uptime.","stack_detail":"React, Go, PostgreSQL, TimescaleDB, WebSocket, Redis, Docker, GCP"},"ru":{"title":"Fleet Tracker","subtitle":"IoT · SaaS","description":"Система управления автопарком в реальном времени: отслеживание 10 000+ транспортных средств с GPS, оптимизацией маршрутов и предиктивным обслуживанием.","challenge":"Обработка высокочастотной GPS-телеметрии от 10k+ автомобилей одновременно с сохранением отзывчивости карты.","solution":"Сервис приёма телеметрии на Go, WebSocket fan-out для обновлений карты в реальном времени, React-фронтенд с canvas-рендерингом.","result":"10k+ ТС в реальном времени, снижение затрат на топливо на 40%, аптайм 99.9%.","stack_detail":"React, Go, PostgreSQL, TimescaleDB, WebSocket, Redis, Docker, GCP"}}$project_fleet_tracker_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    category = EXCLUDED.category,
    categories = EXCLUDED.categories,
    tags = EXCLUDED.tags,
    year = EXCLUDED.year,
    featured = EXCLUDED.featured,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

INSERT INTO projects (slug, category, categories, tags, year, featured, published, sort_order, translations)
VALUES ('cicd-panel', 'devops', ARRAY['devops', 'web'], ARRAY['React', 'Python', 'Docker', 'Kubernetes'], 2022, false, true, 5, $project_cicd_panel_json${"en":{"title":"CI/CD Control Panel","subtitle":"DevOps · Internal Tool","description":"Unified deployment dashboard for 200+ microservices across 4 environments with one-click rollbacks and automated canary releases.","challenge":"Engineering teams were losing hours each week navigating multiple CI/CD tools with no unified view of deployment health.","solution":"Custom React dashboard aggregating data from GitHub Actions, ArgoCD, and Datadog, with a Python backend for orchestration.","result":"Deployment time reduced by 70%, MTTR cut from 45 min to 8 min, adopted by 80+ engineers.","stack_detail":"React, TypeScript, Python, FastAPI, Docker, Kubernetes, ArgoCD, Datadog"},"ru":{"title":"CI/CD Панель","subtitle":"DevOps · Внутренний Инструмент","description":"Единая панель деплоя для 200+ микросервисов в 4 средах с откатами в один клик и автоматическими canary-релизами.","challenge":"Инженерные команды теряли часы в неделю, переключаясь между несколькими CI/CD инструментами без единого обзора.","solution":"Кастомный React-дашборд, агрегирующий данные из GitHub Actions, ArgoCD и Datadog, с Python-бэкендом для оркестрации.","result":"Время деплоя сокращено на 70%, MTTR — с 45 до 8 минут, 80+ инженеров перешли на новую систему.","stack_detail":"React, TypeScript, Python, FastAPI, Docker, Kubernetes, ArgoCD, Datadog"}}$project_cicd_panel_json$::jsonb)
ON CONFLICT (slug) DO UPDATE SET
    category = EXCLUDED.category,
    categories = EXCLUDED.categories,
    tags = EXCLUDED.tags,
    year = EXCLUDED.year,
    featured = EXCLUDED.featured,
    published = EXCLUDED.published,
    sort_order = EXCLUDED.sort_order,
    translations = EXCLUDED.translations,
    updated_at = now();

-- +goose Down
DELETE FROM projects WHERE slug IN ('analytics-dashboard', 'fintech-wallet', 'marketplace-platform', 'hr-suite', 'fleet-tracker', 'cicd-panel');
