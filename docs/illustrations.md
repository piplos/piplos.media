# Piplos Media — Brand Illustrations

Гайд по генерации иллюстраций в фирменном стиле (изометрический чёрный кот-маскот).

## Существующие иллюстрации

| Файл | Тема |
|------|------|
| `web/site/static/hero-cat-isometric.png` | Hero главной — кот на стеке интерфейсов |
| `web/site/static/illustrations/cat-web.png` | Веб-разработка — кот печатает на клавиатуре |
| `web/site/static/illustrations/cat-mobile.png` | Мобильная разработка — кот тапает лапой по экрану |
| `web/site/static/illustrations/cat-backend.png` | Backend/API — кот выглядывает из-за серверной стойки |
| `web/site/static/illustrations/cat-devops.png` | DevOps — кот толкает контейнер к CI/CD pipeline |
| `web/site/static/illustrations/cat-launch.png` | Запуск проекта — кот прыгает от радости у ракеты |
| `web/site/static/illustrations/cat-analytics.png` | Аналитика — кот в очках показывает на график |
| `web/site/static/illustrations/cat-contact.png` | Контакты — кот в гарнитуре лежит и слушает |
| `web/site/static/illustrations/cat-404.png` | Страница 404 — кот озадаченно смотрит на сломанную страницу |
| `web/site/static/illustrations/cat-500.png` | Страница 500 — испуганный кот у серверной стойки с кабелями |

Иллюстрации услуг (`data/uploads/services/{slug}.png`) — уменьшенные до 800px
копии: web ← cat-web, backend ← cat-backend, mobile ← cat-mobile,
data ← cat-analytics, devops ← cat-devops.

## Как сгенерировать новую иллюстрацию

Генераторы изображений не отдают настоящий альфа-канал, поэтому используется
двухшаговый процесс: генерация на однотонном magenta-фоне (chroma key) →
программное удаление фона.

### Шаг 1 — промпт

Всегда прикладывайте референс `web/site/static/hero-cat-isometric.png`
для консистентности стиля. Шаблон промпта (замените только блок `{SCENE}`):

```text
Same modern flat isometric illustration style as the reference image
(stylized black cat mascot with red-orange collar and hexagonal tag,
dark charcoal panels, red-orange #fd533f accents, subtle teal highlights,
floating dots and thin dashed connection lines). SCENE: {SCENE}.
IMPORTANT: the entire background must be one single flat uniform pure magenta
color #FF00FF (chroma key), completely solid, no gradient, no checkerboard,
no shadows cast onto the background, no vignette. Illustration elements
contain no magenta.
```

### Правило: кот всегда в действии

Кот — не статичная фигурка. В каждой иллюстрации задавайте ему **уникальную
позу и действие**, связанное с темой. Пишите действие ЗАГЛАВНЫМИ в начале
сцены (`ACTIVELY TYPING`, `PLAYFULLY TAPPING`, `PEEKING CURIOUSLY`…) и
описывайте положение тела, лап, хвоста и глаз (открыты/закрыты).

Использованные позы (не повторять для новых иллюстраций):

| Иллюстрация | Поза |
|-------------|------|
| hero | сидит с закрытыми глазами на стеке панелей |
| cat-web | печатает — стоит на задних лапах у клавиатуры, глаза открыты |
| cat-mobile | тапает лапой по экрану смартфона в игривом выпаде |
| cat-backend | выглядывает из-за серверной стойки, любопытные глаза |
| cat-devops | толкает контейнер обеими передними лапами |
| cat-launch | прыгает от радости, все лапы в воздухе |
| cat-analytics | в очках, стоит на задних лапах и показывает на график |
| cat-contact | в гарнитуре, лежит на животе, лапы скрещены |
| cat-404 | сидит озадаченно, голова наклонена вверх, знак вопроса над головой |
| cat-500 | отпрянул в испуге, одна лапа поднята, круглые шокированные глаза |

Идеи для новых поз: потягивается, спит клубком на панели, несёт что-то
в зубах, свисает лапой с платформы, умывается, охотится в приседе,
балансирует на стопке кубов.

Примеры `{SCENE}`:

- `the black cat is ACTIVELY TYPING — standing on its hind legs at a floating dark keyboard, front paws pressing keys, eyes open and focused on a large isometric browser window with a code editor. Theme: web development`
- `the black cat wearing a headset is LYING RELAXED ON ITS BELLY on a dark rounded platform, front paws crossed, head tilted attentively; a floating chat panel with message bubbles beside it. Theme: contact and support`
- `the black cat is JUMPING WITH JOY — all four paws off the ground, celebrating next to a launching red-orange isometric rocket. Theme: successful project launch`

Ключевые элементы стиля (не убирать из промпта):

- **Кот**: гладкий чёрный, красно-оранжевый ошейник с шестиугольным жетоном,
  оранжевые внутренние стороны ушей; поза и выражение — свои для каждой сцены
- **Палитра**: тёмный графит `#2a2b2f`/чёрный, акцент `#fd533f`,
  второстепенный teal `#1cd8d2`, белые детали
- **Композиция**: изометрия, парящие панели/иконки, тонкие пунктирные линии
  связей, мелкие плавающие точки-сферы
- **Aspect ratio**: 1:1

### Шаг 2 — удаление фона

```bash
python3 scripts/chroma_key.py input-chroma.png web/site/static/illustrations/cat-<name>.png
```

Скрипт (`scripts/chroma_key.py`, требуется `pillow` и `numpy`):

1. вычисляет расстояние каждого пикселя до чистого magenta `#FF00FF`
2. строит альфа-канал (< 60 — прозрачно, > 180 — непрозрачно, между — плавно)
3. убирает пурпурный каст везде (в палитре нет magenta, поэтому
   `min(R,B) > G` — всегда артефакт хромакея, включая «тени» на панелях)
4. обрезает по содержимому с отступом 8px и сохраняет RGBA PNG

### Шаг 3 — проверка

- Откройте PNG на тёмном и светлом фоне — не должно быть пурпурной каймы
- Убедитесь, что режим файла RGBA:
  `python3 -c "from PIL import Image; print(Image.open('file.png').mode)"`

## Использование на сайте

Файлы кладутся в `web/site/static/illustrations/` и подключаются как
`/illustrations/cat-<name>.png`. Фон прозрачный — картинки работают в обеих
темах без рамок и `border-radius`.
