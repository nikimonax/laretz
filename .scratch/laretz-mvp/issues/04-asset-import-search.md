# 04: Импорт Assets: поиск по каталогу

**What to build:** Пользовательский жест «я хочу этот BTC связать с ценой coingecko:bitcoin»: поиск по локальному каталогу на `/assets/search`, импорт создаёт Asset + primary AssetLink одним действием. Список `/assets` — без цен. Доступны переименование, re-link на другой ExternalAsset того же источника и удаление (блокируется при наличии операций — контракт, сквозная проверка придёт с операциями). См. спеку: «Зафиксированные решения», «Use cases».

**Blocked by:** 03 (Integrations и синк каталога).

**Status:** ready-for-agent

- [ ] Поиск по кешу каталога через FTS5: по symbol, name, external_id; результаты с указанием источника; офлайн работает
- [ ] ImportAsset: Asset + AssetLink создаются вместе; symbol/name предзаполняются из ExternalAsset и редактируемы; линк помечен primary
- [ ] `/assets`: список (symbol, name, source, действия), без цен; пустой каталог — подсказка синкануть
- [ ] Re-link: переливковка на другой ExternalAsset того же источника; у Asset ровно один primary
- [ ] DeleteAsset запрещён при наличии Operation (контракт в команде)
- [ ] Тесты: SearchExternalAssets (FTS5), ImportAsset, UpdateAsset (rename/re-link)
