# Laretz MVP — спека

Зафиксированные требования по итогам grill-сессии. Терминология — в корневом `CONTEXT.md`, ключевые решения — в `docs/adr/0001..0003`.

## Обзор

Однопользовательский трекер крипто-портфелей: MPA SSR desktop-приложение (локальный HTTP-сервер на `127.0.0.1` + браузер, без авторизации), хранение в sqlite (`~/.local/share/laretz` на linux, appdata на windows). Пользователь заносит операции вручную, подключает интеграции поставщиков рыночных данных и по кнопке догружает цены, чтобы видеть историческую и текущую стоимость портфелей.

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

## Схема sqlite

- `portfolio(id, name, created_at, modified_at)`
- `asset(id, symbol, name, created_at, modified_at)`
- `external_asset(id, source, external_id, symbol, name, synced_at)` UNIQUE(source, external_id)
- `asset_link(id, asset_id→asset, external_asset_id→external_asset, is_primary, modified_at)` UNIQUE(asset_id, external_asset_id); один primary на asset — partial unique index
- `operation(id, portfolio_id→, asset_id→, type CHECK(buy,sell), executed_at, quantity, total_usd, fee_amount?, fee_asset_id→asset?, note?, created_at, modified_at)`
- `integration(source PK, api_key?, enabled, catalog_synced_at?, modified_at)`
- `price_point(external_asset_id→, ts_hour, price_usd, granularity CHECK(hour,day))` PK(external_asset_id, ts_hour); дневная точка — час 00:00 UTC
- `external_asset_fts` — FTS5 поверх (symbol, name, external_id)

Децималы — `shopspring/decimal`, в БД TEXT. Время — unix UTC; в форме `datetime-local` (локальная зона браузера), будущее время запрещено.

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

Страницы (GET): `/`, `/portfolio/:id` (итог, chart, позиции, операции; независимая пагинация `?assets_page=&ops_page=`), `/assets` (список без цен), `/assets/search?query=`, `/settings`.

Действия (POST + PRG, ошибки — перерисовка формы):

- `POST /portfolio`, `POST /portfolio/:id` (update; `_method=DELETE` — удаление с confirm-модалкой)
- `POST /portfolio/:id/operation`, `POST /portfolio/:id/operation/:id` (update/delete аналогично)
- `POST /assets` (импорт), `POST /assets/:id` (update/re-link/delete)
- `POST /settings/integrations`, `POST /settings/integrations/:source`, `POST /settings/integrations/:source/sync-catalog`
- `POST /quotes/fetch` — глобальная кнопка, PRG на текущую страницу
- `GET /api/portfolio/:id/chart` — JSON для uPlot

Модалки — `<dialog>` + одна строка inline JS (`showModal()`); разметка модалок присутствует в HTML страницы. Удаление — method spoofing (ADR-0003). Middleware: spoofing, origin/host check, logging (zap+logr).

## Структура

```
internal/
  app/      application: commands/, queries/, ports/ (AssetCatalog, QuoteFeed, Clock)
  domain/   сущности; сервисы: BalanceTimeline, CoveragePlanner, Valuation
  impl/     infrastructure: sqlite/ (sqlc + goose, modernc.org/sqlite, FTS5),
            plugin/ (coingecko-адаптер, реестр MarketDataSource)
  lib/      reusable: coingecko client (net/http), утилиты
  server/   app.go (fx composition root), handler/, middleware/,
            presenter/, model/ (view models), views/{pages,components} (templ),
            static/ (Pico.css, uPlot — go:embed)
cmd/laretz/ main.go
```

fx-модули по слоям; реестр MarketDataSource собирается из impl/plugin; internal/app зависит от internal/domain.

## Стек

chi, templ, uber-go/fx, sqlc + goose, modernc.org/sqlite, shopspring/decimal, uber-go/zap + go-logr/logr, go-playground/validator, testify + mockery + moq, golangci-lint, Pico.css + uPlot (вендоренные, офлайн работает).

## Out of MVP

CSRF-токен, CORS-политика, transfer/кошельки, второй провайдер и UI переключения primary, авто-синк цен по расписанию, live-цены, PnL / cost basis, мультивалюта, страница актива, архив портфелей, bleve-поиск.

## Открытые вопросы

Нет.
