# 03: Plugin API, coingecko и синк каталога

**What to build:** План изменён (см. ADR-0004): UI настроек интеграций в MVP нет, интеграции — фиксированный набор плагинов, включение — по наличию `api_key` в конфиге (`plugins.<id>.api_key`). Публичный plugin API в `pkg/plugin` (`Plugin`, `App`, `AssetProvider`), эталонный плагин coingecko в `plugin/coingecko` (клиент net/http, `x-cg-demo-api-key`, `/coins/list`), регистрация плагинов в `cmd/laretz/main.go` через `plugin.DefaultRegistry()`. Хост: реестр плагинов и фасад хоста в `internal/impl/plugin`, провайдеры каталога доступны ядру через порт `ExternalAssetProviderRegistry` (`internal/app/interfaces`). Синк каталога — вручную кнопкой «Sync assets» на `/` (диалог выбора источника → `POST /assets/sync`), а не на старте: `SyncExternalAssetsUseCase` → команда `SyncExternalAssetsCommand` → upsert ExternalAsset в одной транзакции (идемпотентен, `synced_at` обновляется). Миграция — `internal/impl/database/migrations`, sqlc — `internal/impl/database/queries` + `generated`, командный обработчик — `internal/impl/handler`, fx-модули — `internal/server/module`. См. спеку: «Use cases», ADR-0001, ADR-0004.

**Blocked by:** 02 (Портфели: CRUD end-to-end + sqlite фундамент).

**Status:** done

- [x] Plugin API в `pkg/plugin` (Plugin, App, AssetProvider) + реестр плагинов (`DefaultRegistry`), фасад хоста (секция конфига + скоуп-логгер) и `ExternalAssetProviderRegistry` с адаптером в `internal/impl/plugin`
- [x] coingecko-плагин в `plugin/coingecko`: клиент (net/http, `x-cg-demo-api-key`, `/coins/list`, лимит ответа, таймаут), AssetProvider, Config{base_url, api_key}; регистрация в `cmd/laretz/main.go`; без api_key провайдер не регистрируется
- [x] Миграция `external_asset` (+ триггер `synced_at`) + sqlc upsert (`ON CONFLICT (source, external_id)`) + `SqliteSyncExternalAssetsCommandHandler` (одна транзакция) в `internal/impl/handler`
- [x] Query провайдеров: порт `List` в `ExternalAssetProviderRegistry`, `ListExternalAssetProvidersUseCase`
- [x] Синк каталога — вручную кнопкой на `/`: диалог с выбором источника → `POST /assets/sync`; старт приложения от внешних сервисов не зависит (план менялся: стартового синка через fx lifecycle нет)
- [x] `pkg/lib/httplog`: RoundTripper логирует method/url/status/duration — заголовки не логируются, api_key не попадает в логи
- [x] «Sync assets» на `/` виден только при зарегистрированных провайдерах; пользователь слова «плагин» не видит
- [ ] Тесты (coingecko клиент/провайдер на fake-HTTP, SyncExternalAssets против sqlite, регистрация провайдеров через фасад) — перенесены в 09

## Comments

- 2026-10-09: закрыт по решению ревьюера; реализация принята как есть. Расхождения с исходным планом, зафиксированные в спеке и ADR-0004: страницы `/settings` и таблица `integration` не делаются (включение — `plugins.<id>.api_key` в конфиге); синк каталога — ручной кнопкой на `/`, а не на старте через fx lifecycle; plugin API вынесен в публичный `pkg/plugin`, плагин coingecko — в топ-левел `plugin/coingecko` (не `internal/impl`); sqlc-запросы переехали в `internal/impl/database/queries` + `generated`; командный обработчик синка — `internal/impl/handler`, а не repository. Подсказка про api_key в empty state `/` не делалась — кнопка синка просто скрыта, пока провайдеров нет. Тесты тикета перенесены в 09.
