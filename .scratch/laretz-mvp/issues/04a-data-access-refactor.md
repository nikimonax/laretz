# 04a: Data-access рефакторинг: только репозитории (ADR-0005)

**What to build:** Один словарь data-доступа: порты — репозитории по концепции домена (ADR-0005). `internal/app/commands`, `internal/app/dto`, `interfaces.CommandHandler` и `internal/impl/handler` удаляются. Записи — через `AssetRepository` (Create/Rename/Relink/Delete) и `ExternalAssetRepository.Sync`; транзакции и `:execrows`→`ErrEntityNotFound` — внутри репозиториев; инварианты («rename вместе», source-match при re-link) — в use case; адаптер плагина отдаёт `domain.ExternalAsset` (единственная точка перевода внешнего вида). Server/views/sqlc/роуты не меняются.

**Blocked by:** 04 (решение выросло из ревью 04).

**Status:** done

- [x] Порты: `AssetRepository` + Create/Rename/Relink/Delete, `ExternalAssetRepository` + Sync (доменные типы, без dto); `interfaces/command.go` удалён
- [x] impl: `handler/` растворяется в `repository/` (tx.go переезжает; Create/Relink/Delete/Sync внутри транзакций), конструкторы репозиториев с `*sql.DB`
- [x] usecase: Import/Update/Delete/Sync зависят от репозиториев, копирки команд схлопнуты; инвариант rename — в use case
- [x] `commands/`, `dto/`, мёртвый `domain.AssetLink` удалены; fx-модуль storage без хендлеров
- [x] Доки: ADR-0005, AGENTS.md (impl-подпакеты), spec.md (конвенция, структура)
- [x] Гейты: go build, go vet, golangci-lint run, go test ./...

## Comments

- 2026-10-09: решение принято в ревью 04. Одна ось «порт = репозиторий по концепции»: агрегат, чью идентичность создаёт наша БД → CRUD-методы; реплика внешнего состояния → sync/upsert-методы с DTO-данными как доменными типами. Гибридный критерий (команды для «многостейтментных» записей) отвергнут как нестабильный — транзакции одинаково инкапсулируются репозиторием. `domain.ExternalAsset` сохраняет surrogate ID (аналогия OAuth: `sub` → внутренний id; bad tone «sub как PK» в схеме отсутствует — PK везде суррогатные); `Sync` принимает transient-сущности, хранилище присваивает ID/SyncedAt, слияние по (source, external_id).
- 2026-10-09: реализовано, гейты пройдены (build, vet, golangci-lint 0 issues, go test ./...), ожидает ревью; без коммита по требованию ревьюера. Отклонения и решения:
    - `AssetRepository.Create` возвращает `(domain.AssetID, error)` (не полный `domain.Asset`): sqlc `InsertAsset` — `:execlastid`, потребителя таймстампов созданной сущности нет (redirect на `/assets`).
    - Rename и Relink — два узких метода порта; смешанная форма (rename + relink в одном POST) сегодня атомарна, после рефакторинга — две транзакции подряд. В UI формы взаимоисключающи (rename-модалка против relink-формы), так что расхождение недостижимо через интерфейс; проверено по server/handler/asset_update.go.
    - Инвариант «symbol и name переименовываются вместе» переехал из sqlite-хендлера в `UpdateAssetUseCase` (сервер и так валидирует оба поля формы).
    - Логика `validateAssetRelink` перенесена в `SqliteAssetRepository.Relink` как есть (2 чтения + сравнение в транзакции); свёртка в один UPDATE с source-предикатом — возможное будущее углубление, не в этом тикете.
    - Server, views, sqlc-запросы, роуты — без изменений.
- 2026-10-11: закрыт. Верификация реализации против тикета и доков пройдена: порты, репозитории (транзакции, `*sql.DB`-конструкторы), use cases (инвариант rename), удаления (`commands/`, `dto/`, `domain.AssetLink`, `interfaces/command.go`, `impl/handler`), fx storage без хендлеров, ADR-0005/AGENTS.md/spec.md — расхождений нет; гейты повторно пройдены (build, vet, golangci-lint 0 issues, go test ./...). Замечание отложено: в ADR-0005 фраза «invariants stay in use cases» не оговаривает source-match при re-link, живущий в `SqliteAssetRepository.Relink` (отклонение зафиксировано в comments выше).
