# Laretz MVP — спека

Зафиксированные требования по итогам grill-сессии. Терминология — в корневом `CONTEXT.md`, ключевые решения — в `docs/adr/0001..0003`.

## Обзор

Однопользовательский трекер крипто-портфелей: MPA SSR desktop-приложение (локальный HTTP-сервер на `127.0.0.1` + браузер, без авторизации), хранение в sqlite — всё в каталоге данных `os.UserConfigDir()/laretz` (`~/.config/laretz` на linux, `%AppData%\laretz` на windows): БД `laretz.db` и `config.yaml` рядом. Пользователь заносит операции вручную, подключает интеграции поставщиков рыночных данных и по кнопке догружает цены, чтобы видеть историческую и текущую стоимость портфелей.

## Пользовательские сценарии

1. **Первый запуск** — `/` показывает empty state со ссылкой в `/settings`; там пользователь включает Integration (пока только coingecko) и вводит api-key.
2. **Синк каталога** — кнопка в настройках: плагин отдаёт каталог целиком, ядро кеширует его как ExternalAsset. Поиск при импорте идёт только по локальному кешу.
3. **Импорт актива** — `/assets/search?query=` ищет по каталогу; выбор результата создаёт Asset + первичный AssetLink. Пользователь говорит: «этот BTC связан с ценой coingecko:bitcoin».
4. **Операции** — пользователь вручную заносит Buy/Sell в портфель (модалка на странице портфеля). Transfer вне MVP.
5. **Цены** — кнопка FetchQuotes дозагружает недостающие диапазоны цен по HoldingInterval; портфель показывает график стоимости, позиции с текущей ценой и операции.

## Зафиксированные решения

- USD — единственная валюта расчётов (Quote Currency); парные сделки вне MVP.
- Operation хранит `quantity + total_usd` (как на биржах). Тройка price×qty=total решается JS на странице; сервер получает ровно двойку (qty, total). Поле price — UX-сахар, сервером игнорируется.
- Fee: amount + Asset (nullable), **не влияет на баланс** — хранится для будущих отчётов.
- Балансы могут уходить в минус; валидации «не хватает актива» нет.
- Полный CRUD операций/портфелей/активов, жёсткое удаление.
- Asset рождается только при импорте (вместе с линком); re-link разрешён; удаление актива — только если нет операций. Удаление портфеля — каскад операций с confirm-модалкой.
- Линков у Asset много (по одному на источник), один primary; цены — по primary.
- Цены в БД — только реально полученные от источника (hour ≤ 90д / day > 90д / пусто > 365д у coingecko demo); существующие PricePoint не перезаписываются, дыры не заполняются автоматически. Forward-fill — только для отображения (см. ADR-0002).
- Текущая цена — только для позиций с балансом ≠ 0, с меткой as-of.
- Каталог синкается по кнопке (при включении интеграции или вручную), без расписаний.
- Первый запуск без `/setup`: empty states ведут в настройки.
- Поиск по каталогу — SQLite FTS5 (bleve отклонён: второй storage не окупается на ~15k строк).
- Coingecko-клиент — свой тонкий `net/http` (go-gecko удалён), эндпоинты: `/coins/list`, `/coins/{id}/market_chart`.
- CSRF-токена нет; бинд на `127.0.0.1` + проверка Origin/Host на POST (см. Out of MVP).
- Конфигурация — yaml (`server`, `log`, `data`), при первом запуске создаётся дефолтный `config.yaml` в каталоге данных; путь переопределяется флагом `-c/--config` или `LARETZ_CONFIG`; `.env` подхватывается при старте (godotenv). Код конфига — `internal/server/config`.
- CLI — urfave/cli v3 в `cmd/laretz/cli`; флаги: `-o/--open` (открыть браузер), `-c/--config`.

## Use cases (CQS)

Commands:

| Command | Содержание |
|---|---|
| CreatePortfolio(name) | создание портфеля |
| UpdatePortfolio(id, name) | переименование |
| DeletePortfolio(id) | каскад операций |
| CreateOperation(portfolioId, assetId, type, executedAt, quantity, totalUsd, feeAmount?, feeAssetId?, note?) | создание операции |
| UpdateOperation(id, …) | редактирование (в т.ч. смена портфеля/актива) |
| DeleteOperation(id) | удаление |
| ImportAsset(externalAssetId, symbol, name) | Asset + primary AssetLink одним действием |
| UpdateAsset(id, symbol?, name?, externalAssetId?) | переименование и/или re-link |
| DeleteAsset(id) | блокируется при наличии операций |
| EnableIntegration(source, apiKey) | включение источника |
| UpdateIntegration(source, apiKey?, enabled?) | ключ/переключение |
| SyncCatalog(source) | полный дамп каталога → upsert ExternalAsset |
| FetchQuotes() | глобально: по всем портфелям/активам дозагрузка недостающих PricePoint |

Queries:

| Query | Содержание |
|---|---|
| ListPortfolios() | список + стоимость каждого (с флагами staleness) |
| GetPortfolio(id, opsPage, assetsPage) | итог, позиции (тек. цена, as-of), операции (пагинация) |
| GetPortfolioChart(id) | JSON-ряд для uPlot (форвард-филл по правилам отображения) |
| ListAssets() | глобальный список, без цен |
| SearchExternalAssets(query) | FTS5 по локальному каталогу |
| GetOperation(id) | префилл модалки редактирования |
| ListIntegrations() | доступные источники + их состояние (основа empty states) |

Конвенция кода: use-case — структура `*UseCase` в `internal/app/usecase` с методом `Execute(ctx, cmd)`; порты — интерфейсы в `internal/app/interfaces`; sentinel-ошибки — в `internal/app` (`ErrEntityNotFound`). Команды, меняющие строки, опираются на `:execrows` sqlc и мапят 0 строк в `ErrEntityNotFound`.

## Схема sqlite

- `portfolio(id, name, created_at, updated_at)`
- `asset(id, symbol, name, created_at, updated_at)`
- `external_asset(id, source, external_id, symbol, name, synced_at, updated_at)` UNIQUE(source, external_id)
- `asset_link(id, asset_id→asset, external_asset_id→external_asset, is_primary, updated_at)` UNIQUE(asset_id, external_asset_id); один primary на asset — partial unique index
- `operation(id, portfolio_id→, asset_id→, type CHECK(buy,sell), executed_at, quantity, total_usd, fee_amount?, fee_asset_id→asset?, note?, created_at, updated_at)`
- `integration(source PK, api_key?, enabled, catalog_synced_at?, updated_at)`
- `price_point(external_asset_id→, ts_hour, price_usd, granularity CHECK(hour,day))` PK(external_asset_id, ts_hour); дневная точка — час 00:00 UTC
- `external_asset_fts` — FTS5 поверх (symbol, name, external_id)

Децималы — `shopspring/decimal`, в БД TEXT. Время — UTC в sqlite `DATETIME`: `created_at`/`updated_at` — `DEFAULT CURRENT_TIMESTAMP`, `updated_at` поддерживается UPDATE-триггером на таблицу (конвенция миграций в `internal/impl/database/migrations`); в форме `datetime-local` (локальная зона браузера), будущее время запрещено.

## FetchQuotes — алгоритм

1. По всем (Portfolio, Asset) строится timeline баланса из операций.
2. HoldingInterval — максимальные интервалы, где баланс ≠ 0; границы — часы граничных операций, включительно; незакрытый интервал — до текущего часа.
3. Интервалы через primary link → ExternalAsset → требуемые часы; минус QuoteCoverage → недостающие часы.
4. Слоты сливаются в непрерывные диапазоны на ExternalAsset.
5. Адаптер запрашивает диапазоны (coingecko: ≤ 90д — hourly, дальше daily; > 365д — пусто), пауза между запросами; upsert `INSERT OR IGNORE`, ничего не перезаписывается.

## Правила отображения

- Стоимость часа = Σ (баланс × цена). Цена = PricePoint часа, иначе форвард-филл последней известной + бейдж «не актуальная».
- Актив без единой цены — исключается из суммы + бейдж «часть активов без цены».
- Отрицательный баланс уменьшает стоимость (как на бирже).

## Маршруты

Страницы (GET): `/`, `/portfolio/{id}` (итог, chart, позиции, операции; независимая пагинация `?assets_page=&ops_page=`), `/assets` (список без цен), `/assets/search?query=`, `/settings`.

Действия (POST + PRG, ошибки — перерисовка формы с переоткрытой модалкой через `views.Error{Message, Dialog}`):

- `POST /portfolio`, `POST /portfolio/{id}` (update), `DELETE /portfolio/{id}` — регистрируется настоящий DELETE-роут, форма шлёт POST с `_method=DELETE`, spoof-мидлвара переписывает метод
- `POST /portfolio/{id}/operation`, `POST /portfolio/{id}/operation/{id}` (update/delete аналогично)
- `POST /assets` (импорт), `POST /assets/{id}` (update/re-link/delete)
- `POST /settings/integrations`, `POST /settings/integrations/{source}`, `POST /settings/integrations/{source}/sync-catalog`
- `POST /quotes/fetch` — глобальная кнопка, PRG на текущую страницу
- `GET /api/portfolio/{id}/chart` — JSON для uPlot

Модалки — `<dialog id=...>` + делегированный клик по `data-dialog` с одним `showModal()`; переоткрытие после ошибки — `open?={ data.Error.Dialog == id }`. Удаление — method spoofing (ADR-0003). Middleware (в порядке подключения): origin-check → spoof → request log (logr).

## Структура

```
cmd/laretz/  main.go (one-liner → cli.Run()), cli/ (urfave/cli v3: флаги, .env, путь конфига)
internal/
  app/        use-case слой (CQS): usecase/ (*UseCase + Execute), interfaces/ (порты, напр. PortfolioRepository),
              errors.go (sentinel-ошибки)
  domain/     сущности; сервисы: BalanceTimeline, CoveragePlanner, Valuation
  impl/       infrastructure: database/ (Connect + goose-миграции, go:embed database/migrations),
              repository/ (sqlc: *.sql запросы + generated/, реализации портов),
              plugin/ (coingecko-адаптер, реестр MarketDataSource) — с тикета 03
  lib/        reusable, stdlib-only: spoof/, browser/; сюда же coingecko-клиент (net/http)
  shared/     cross-cutting c внешними зависимостями: logging/ (zap+logr, request-log middleware, fx/goose-адаптеры),
              validator/ (go-playground singleton)
  server/     composition root: app.go (fx.New), module/ (fx-модуль на concern: logging, database, repository,
              usecase, http, browser, ...), config/ (yaml: Server/Log/Data), handler/ (файл на действие,
              конструктор New*Handler), middleware/ (origin), renderer/ (ResponseRenderer + Provider/Mutate,
              фабрики Home/HomeWithError/NotFound/...), static/ (go:embed Pico.css, uPlot),
              views/{pages,components} (templ; общие типы — views/views.go)
```

Хендлеры и рендеринг: каждая страница — `*ResponseRenderer[PageData]` (Provider достаёт данные через запросы, Mutate дописывает ошибки), действия — отдельные handler-файлы, валидационная ошибка перерисовывает страницу с `views.Error`. fx-модули — только в `internal/server/module`; `internal/app` и `internal/domain` — обычные Go-пакеты; `internal/impl` зависит от app + domain.

## Стек

chi, templ, uber-go/fx, sqlc + goose (миграции вшиты go:embed), modernc.org/sqlite, shopspring/decimal, uber-go/zap + go-logr/logr, go-playground/validator, urfave/cli v3 + joho/godotenv, goccy/go-yaml, testify + mockery + moq, golangci-lint, Pico.css + uPlot (вендоренные, офлайн работает).

## Out of MVP

CSRF-токен, CORS-политика, transfer/кошельки, второй провайдер и UI переключения primary, авто-синк цен по расписанию, live-цены, PnL / cost basis, мультивалюта, страница актива, архив портфелей, bleve-поиск.

## Открытые вопросы

Нет.
