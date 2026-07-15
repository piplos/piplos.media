# Карта редиректов: старый piplos.media → новый сайт

Источник: sitemap.xml старого сайта (409 URL) + краул навигации (about, team, разделы портфолио).
Все 217 слагов проектов и структура `/{lang}` (en|ru) совпадают со старым сайтом — контентные URL
переезжают без потерь. Ниже — только URL, изменившие адрес.

## Услуги (301, серверный redirect)

`web/site/src/routes/[lang]/services/[slug]/+page.server.ts` → `LEGACY_SLUGS`

| Старый URL                                  | Новый URL                  |
| ------------------------------------------- | -------------------------- |
| `/{lang}/services/web-application-development`   | `/{lang}/services/web`     |
| `/{lang}/services/custom-software-development`   | `/{lang}/services/backend` |
| `/{lang}/services/mobile-app-development`        | `/{lang}/services/mobile`  |
| `/{lang}/services/maintenance-and-support`       | `/{lang}/services/devops`  |
| `/{lang}/services/quality-assurance-and-testing` | `/{lang}/services/backend` |
| `/{lang}/services/ui-ux-design-services`         | `/{lang}/services/web`     |
| `/{lang}/services` (список)                      | `/{lang}#services` (308)   |

## Портфолио (301, серверные redirect)

Проекты: `[lang]/portfolio/[type]/[slug]/+page.server.ts` → `LEGACY_TYPES`.
Разделы: `[lang]/portfolio/[slug]/+page.server.ts` → `LEGACY_TYPE_FILTERS`
(фильтр выбран по фактическим категориям проектов каждого раздела).

| Старый URL                             | Новый URL                          |
| -------------------------------------- | ---------------------------------- |
| `/{lang}/portfolio/{type}/{slug}`      | `/{lang}/portfolio/{slug}`         |
| `/{lang}/portfolio/sites`              | `/{lang}/portfolio?filter=web`     |
| `/{lang}/portfolio/landing`            | `/{lang}/portfolio?filter=web`     |
| `/{lang}/portfolio/smm`                | `/{lang}/portfolio?filter=web`     |
| `/{lang}/portfolio/app`                | `/{lang}/portfolio?filter=mobile`  |
| `/{lang}/portfolio/soft`               | `/{lang}/portfolio?filter=backend` |

Типы старого сайта: `sites`, `soft`, `landing`, `app`, `smm`.

## Страницы без прямого аналога (301 на похожую страницу)

| Старый URL                     | Новый URL         | Обоснование                          |
| ------------------------------ | ----------------- | ------------------------------------ |
| `/{lang}/about`                | `/{lang}#about`   | секция «О компании» на главной       |
| `/{lang}/team`                 | `/{lang}#about`   | отдельной страницы команды нет       |
| `/{lang}/team/vacancy-designer`| `/{lang}#about`   | вакансий на новом сайте нет          |
| `/{lang}/team/vacancy-php`     | `/{lang}#about`   | вакансий на новом сайте нет          |

## URL без языкового префикса (клиентский redirect с определением языка)

Старый сайт отдавал 302 на `/ru/...`; новый определяет язык из localStorage/браузера
(`LangRedirect`, маршруты `web/site/src/routes/{about,team,order,portfolio,services,privacy,terms,cookies}`).

| Старый URL     | Новый URL                   |
| -------------- | --------------------------- |
| `/`            | `/{lang}`                   |
| `/about`       | `/{lang}#about`             |
| `/team`, `/team/*` | `/{lang}#about`         |
| `/order`       | `/{lang}/order`             |
| `/portfolio`   | `/{lang}/portfolio`         |
| `/portfolio/{slug}` | `/{lang}/portfolio/{slug}` |
| `/services`    | `/{lang}#services`          |
| `/privacy`     | `/{lang}/legal/privacy`     |
| `/terms`       | `/{lang}/legal/terms`       |
| `/cookies`     | `/{lang}/legal/cookies`     |
| `/legal`       | `/{lang}/legal/privacy`     |
| `/legal/{slug}`| `/{lang}/legal/{slug}`      |

## Без изменений

`/{lang}`, `/{lang}/portfolio`, `/{lang}/portfolio/{slug}` (217 слагов), `/{lang}/order`,
`/robots.txt`, `/sitemap.xml`.
