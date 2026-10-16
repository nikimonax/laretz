# 04: Импорт Assets: поиск по каталогу

**What to build:** Пользовательский жест «я хочу этот BTC связать с ценой coingecko:bitcoin»: поиск по локальному каталогу на `/assets/search`, импорт создаёт Asset + primary AssetLink одним действием. Список `/assets` — без цен. Доступны переименование, re-link на другой ExternalAsset того же источника и удаление (блокируется при наличии операций — контракт, сквозная проверка придёт с операциями). FTS5-таблица — в миграции `internal/impl/database/migrations`, поисковые sqlc-запросы — в `internal/impl/repository`. См. спеку: «Зафиксированные решения», «Use cases».

**Blocked by:** 03 (Plugin API, coingecko и синк каталога).

**Status:** done
- [x] Поиск по кешу каталога через FTS5: по symbol, name, external_id; результаты с указанием источника; офлайн работает
- [x] ImportAsset: Asset + AssetLink создаются вместе; symbol/name предзаполняются из ExternalAsset и редактируемы; линк помечен primary
- [x] `/assets`: список (symbol, name, source, действия), без цен; пустой каталог — подсказка синкануть
- [x] Re-link: переливковка на другой ExternalAsset того же источника; у Asset ровно один primary
- [x] DeleteAsset запрещён при наличии Operation (контракт в команде)

Тесты (SearchExternalAssets FTS5, ImportAsset, UpdateAsset rename/re-link) — в 09, единый тестовый проход в конце MVP.

## Comments

- 2026-10-09: реализовано, гейты пройдены (build, vet, golangci-lint, go test ./...), ожидает ревью; без коммита по требованию ревьюера. Отклонения и решения:
    - Поисковый запрос — рукописный SQL в `internal/impl/repository/external_asset.go` (не sqlc): sqlc не переваривает FTS5 TVF-вызов со вторым параметром в ORDER BY (молча выбрасывает параметр). Остальные запросы — в sqlc как обычно.
    - Ранжирование: точное совпадение symbol (case-insensitive) — первым, дальше bm25; без этого "btc" отдавал `$btc25` раньше `Bitcoin`. Лимит — 50 результатов.
    - FTS5-таблица — external content поверх external_asset, синк через INSERT/UPDATE/DELETE-триггеры в миграции 00003; проверено на живом каталоге (~17k строк coingecko, синк был в smoke-тесте).
    - Re-link на другой источник запрещён в транзакции командного обработчика (sentinel `ErrSourceMismatch`); в re-link-режиме страницы поиска кнопки чужих источников disabled.
    - DeleteAsset: sentinel `ErrAssetInUse` заведён, маппинг в 404-подобное сообщение на странице есть, сама проверка наличия операций придёт в 05 (таблицы operation ещё нет).
    - `POST /assets/{id}` обслуживает и rename (модалка на /assets), и re-link (форма на /assets/search) — различаются по наличию полей в форме; ошибки re-renderят соответствующую страницу (search — с сохранённым query/relink из hidden-полей).
    - Бойлерплейт begin/rollback/commit синк-обработчика вынесен в общий `inTransaction` (`internal/impl/handler/tx.go`), синк переведён на него.
    - В layout добавлена навигация Portfolios/Assets — иначе страница /assets недостижима.
