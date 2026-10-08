# Integrations are config-driven; catalog sync is manual

The MVP ships a fixed, compile-time set of plugins behind the `pkg/plugin` API (`Plugin`, `App`, `AssetProvider`). An integration counts as enabled when its config section (`plugins.<id>.api_key` in `config.yaml`) provides a key; without a key the plugin registers no catalog provider and the sync button on `/` does not appear. The catalog is loaded only on demand — the «Sync assets» button upserts the whole catalog of the chosen source as ExternalAssets (`POST /assets/sync`). The host side (plugin registry, config facade, `ExternalAssetProviderRegistry`) lives in `internal/impl/plugin`; plugins are registered into `plugin.DefaultRegistry()` in `cmd/laretz/main.go`.

## Considered options

A `/settings` page storing api keys in the `integration` table (EnableIntegration / UpdateIntegration use cases) — rejected for MVP: the config file already exists, stays out of the repo, and a settings UI buys nothing for a single user.

Syncing every registered catalog at startup via fx lifecycle — rejected: a slow or failing external service would sit on the critical path of every launch for data the user may not need yet.

## Consequences

The `integration` table and the Enable/Update/ListIntegrations use cases are gone; the host facade hands each plugin its config section and a plugin-scoped logger. Adding an integration means compiling in a plugin and writing a config section. The user-facing word is still «интеграция» — «плагин» stays internal. Outbound HTTP calls to sources go through `pkg/lib/httplog`, which logs method/url/status/duration only, so keys never reach the logs.
