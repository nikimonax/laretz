# 03: Integrations и синк каталога

**What to build:** Пользователь на `/settings` включает Integration (пока только coingecko), вводит api-key и по кнопке синкает Catalog: плагин отдаёт полный список активов, ядро кеширует их как ExternalAsset. После синка поиск при импорте (тикет 04) работает офлайн. Пользователь не видит слова «плагин» — только «интеграция». Хендлеры настроек — по образцу портфельных: `PageHandler` + `*ResponseRenderer` с Mutate для ошибок (`views.Error{Message, Dialog}`). Миграция — в `internal/impl/database/migrations`, sqlc-запросы — в `internal/impl/repository`, use-case'ы — в `internal/app/usecase` + порты в `internal/app/interfaces`, fx-модуль — в `internal/server/module`. См. спеку: «Use cases», ADR-0001.

**Blocked by:** 02 (Портфели: CRUD end-to-end + sqlite фундамент).

**Status:** ready-for-agent

- [ ] `/settings`: список доступных MarketDataSource, включение с api-key, редактирование ключа, вкл/выкл
- [ ] SyncCatalog: `/coins/list` через собственный coingecko-клиент (net/http), полный upsert ExternalAsset, `synced_at` обновляется, повторный синк идемпотентен
- [ ] Demo-key передаётся заголовком источника; ключ не попадает в логи
- [ ] Empty states: `/` и `/assets` без включённой интеграции ведут на `/settings`
- [ ] Порты AssetCatalog (и зачаток QuoteFeed) в `internal/app/interfaces`; coingecko-клиент и адаптер в `internal/impl/plugin`; fx-модуль в `internal/server/module`
- [ ] Тесты: адаптер каталога на fake-HTTP, команды Enable/Update/SyncCatalog
