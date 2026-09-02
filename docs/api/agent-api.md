# Agent API — создание статей внешними агентами

API для внешних автоматизаций (Manus, n8n и др.): создание статей сразу на всех
языках сайта, загрузка обложек и планирование времени публикации.

## Базовые сведения

- **Base URL:** `https://api.piplos.media/v1` (локально: `http://localhost:3001/v1`)
- **Аутентификация:** `Authorization: Bearer <api-key>` — ключ вида `pk_live_...`
- **Формат:** JSON, даты в RFC 3339 / UTC (например `2026-09-01T09:00:00Z`)

Ключи выдаются администратором: `POST /v1/api-keys` (JWT админа) → в ответе
`{"api_key": {...}, "key": "pk_live_..."}`. Сырой ключ показывается **один раз** —
в БД хранится только SHA-256 хеш. Ключ можно отозвать (`POST /v1/api-keys/:id/revoke`)
или удалить (`DELETE /v1/api-keys/:id`); срок действия не ограничен.

Агентские методы не принимают JWT пользователей — только API-ключи.

## Статья: поля

| Поле | Тип | Описание |
|---|---|---|
| `slug` | string | Опционально. `a-z0-9` и `-`. Без указания генерируется из title языка по умолчанию (с транслитерацией кириллицы); при коллизии добавляется `-2`, `-3`… |
| `published` | bool | `false` → черновик; `true` → публикация (немедленная или отложенная через `publish_at`) |
| `publish_at` | string \| null | RFC 3339. В будущем + `published: true` → отложенная публикация; в прошлом или null → немедленная |
| `image` | string | Путь к обложке, например `/uploads/pages/<slug>/cover.webp` (см. «Загрузка картинок») |
| `tags` | string[] | **Обязательно.** Стек статьи: непустой список меток из каталога стека — см. `GET /v1/agent/stack`. Совпадение с `label` без учёта регистра; значения канонизируются к каталогу, дубликаты схлопываются, максимум 20 |
| `translations` | object | **Обязательно.** Каждое включённое зеркало языка: `title`, `description`, `body` |
| `seo` | object \| null | Опционально. `{ "translations": { "<lang>": { "title", "description", "keywords", "og_title", "og_description", "og_image" } } }` — обязательны `title` и `description` на каждом языке |

`body` — **Markdown** (публичный сайт рендерит его через goldmark). Не присылайте HTML.

## Валидация языков (строгая)

Каждое включённое зеркало языка должно иметь непустые `title`, `description`
и `body` (для SEO — `title`, `description`). Актуальный список языков:
`GET /v1/public/languages`. Нарушение → `422` с точным перечнем:

```json
{"error": "validation_failed", "message": "missing: en.title, ru.body"}
```

Лишние/опечатанные коды языков тоже отклоняются:
`{"error": "validation_failed", "message": "unsupported languages: de"}`.

## Каталог стека (теги статьи)

`GET /v1/agent/stack` → `{"stack": [{"slug", "label", "icon", "group_id", …}]}` —
опубликованные технологии сайта. `tags` статьи обязателен и собирается только
из этих значений (поле `label`); передавать можно без учёта регистра — сервер
приведёт к канонической форме каталога и уберёт дубликаты.

Нарушение → `422`:

```json
{"error": "validation_failed", "message": "missing: tags"}
{"error": "validation_failed", "message": "unknown stack tags: Nocat"}
```

## Методы

### Создать статью

`POST /v1/agent/articles` → `201`

```json
{
  "slug": "flutter-mobile-app-development",
  "published": true,
  "publish_at": "2026-09-01T09:00:00Z",
  "image": "/uploads/pages/flutter-mobile-app-development/cover.webp",
  "tags": ["Flutter"],
  "translations": {
    "en": {
      "title": "Flutter Mobile App Development",
      "description": "Why Flutter fits cross-platform projects.",
      "body": "# Flutter\n\nFull article text in Markdown…"
    },
    "ru": {
      "title": "Разработка мобильных приложений на Flutter",
      "description": "Почему Flutter подходит кроссплатформенным проектам.",
      "body": "# Flutter\n\nПолный текст статьи в Markdown…"
    }
  },
  "seo": {
    "translations": {
      "en": {"title": "Flutter Development — Piplos", "description": "Flutter app development services", "keywords": "flutter, mobile"},
      "ru": {"title": "Разработка на Flutter — Piplos", "description": "Услуги разработки мобильных приложений", "keywords": "flutter, мобильные приложения"}
    }
  }
}
```

Ответ:

```json
{
  "article": {
    "id": "0d9f…", "slug": "flutter-mobile-app-development",
    "published": true, "publish_at": "2026-09-01T09:00:00Z",
    "image": "/uploads/…", "tags": ["Flutter"],
    "translations": { "…": "…" },
    "created_at": "…", "updated_at": "…",
    "status": "scheduled",
    "seo": {"id": "…", "path": "/articles/flutter-mobile-app-development", "translations": {…}}
  }
}
```

### Список статей

`GET /v1/agent/articles` → `{"articles": [...]}` — все статьи, включая черновики
и отложенные (SEO в списке не отдаётся).

### Одна статья

`GET /v1/agent/articles/:id` → `{"article": {...}}` — с полем `seo`, если оно есть.

### Обновить статью

`PUT /v1/agent/articles/:id` → `{"article": {...}}`. Полная замена, валидация
языков такая же строгая, как при создании.

- пустой `slug` → оставить текущий; смена slug переносит SEO-запись на новый путь
  (`/articles/<slug>`); если новый путь занят другой записью → `409`
- блок `seo` не передан → существующая SEO-запись не меняется

### Удалить статью

`DELETE /v1/agent/articles/:id` → `{"ok": true}`. Удаляет также SEO-запись статьи.

### Загрузка картинок

`POST /v1/agent/uploads` — multipart/form-data:

- `file` — jpg/png/webp/gif/svg, до 5 МиБ
- `path` — папка, по конвенции `pages/<slug>`
- `name` — опциональное имя файла

Ответ содержит `url`/`path` (уже указывающие на WebP, если он сгенерирован):

```json
{"url": "/uploads/pages/my-slug/cover.webp", "path": "/pages/my-slug/cover.webp",
 "filename": "cover.webp", "mime": "image/webp", "webp_sizes": ["/uploads/pages/my-slug/cover-480.webp", …]}
```

Полученный `url` используйте как `image` статьи (и в Markdown внутри `body`).

## Планирование публикации

Статус в ответах (`status`) вычисляется на сервере:

| `published` | `publish_at` | `status` | Видна на сайте |
|---|---|---|---|
| `false` | любой | `draft` | нет |
| `true` | в будущем | `scheduled` | станет видимой в момент `publish_at` автоматически |
| `true` | null или в прошлом | `published` | да |

Отдельного вызова «опубликовать» не требуется: отложенная статья появляется
на сайте сама, когда наступает `publish_at`. Публикация `publish_at` в прошлом
не считается ошибкой — статья публикуется немедленно.

## Ошибки

| HTTP | `error` | Когда |
|---|---|---|
| 400 | `invalid_request` | Битой JSON, невалидный slug, битой `publish_at` (нужен RFC 3339) |
| 401 | `unauthorized` | Нет/неверный/отозванный API-ключ |
| 404 | `not_found` | Статья не найдена |
| 409 | `conflict` | Slug или SEO-путь заняты |
| 422 | `validation_failed` | Неполные переводы или теги вне каталога стека — см. `missing:` / `unknown stack tags:` в сообщении |
| 500 | `internal_error` | Сбой на сервере |

Формат всех ошибок: `{"error": "<code>", "message": "<human-readable>"}`.
