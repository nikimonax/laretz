# Repositories are the only data-access ports

The application layer accesses data through one vocabulary: ports are repositories named after a domain concept (`PortfolioRepository`, `AssetRepository`, `ExternalAssetRepository`, …) with methods phrased as domain-language sentences. An aggregate whose identity is assigned by our storage (Asset, Portfolio, Operation) gets CRUD-shaped methods; a replica of external state (the ExternalAsset catalog, later price coverage) gets sync/upsert-shaped methods that accept the external data as domain types — identity is owned by the source and entries merge by `(source, external_id)`. Transactions, `:execrows`→`ErrEntityNotFound` mapping and SQL details live inside the repository implementations and are invisible to use cases; invariants stay in use cases. Use-case input commands (the CQS boundary server → use case, see the spec) are unaffected.

The previous split — repositories for reads and portfolio writes, generic `interfaces.CommandHandler[T]` sqlite handlers for asset writes and catalog sync — grew by accident: the command style was born with the catalog sync (a bulk upsert of identity-less provider dumps) and spread to asset writes because they needed multi-statement transactions. Both jobs were one job, and no criterion told new functionality where to go («why is DeleteAsset a command while DeletePortfolio is a repository method?»).

## Considered options

Full CQRS (generic `CommandHandler`/`QueryHandler` per operation, reads included) — rejected: roughly twice the types plus payload-copy ceremony for a single-user local SQLite app with no bus, no middleware and no separate read store.

Hybrid with a documented criterion (commands for multi-statement/invariant writes, repositories for simple CRUD) — rejected: the criterion is unstable. Transactions and invariants are equally encapsulated by a repository method (DeletePortfolio becomes a cascading delete, staying a repository method), so the classification flip-flops per feature — which was the problem being solved.

## Consequences

`internal/app/commands`, `internal/app/dto`, `interfaces.CommandHandler` and `internal/impl/handler` are gone; the `inTransaction` helper lives with the repositories, and repository constructors take `*sql.DB`. The plugin adapter translates `pkg.Asset` → `domain.ExternalAsset` at the plugin boundary — the single translation point, so ports speak domain types only. `domain.ExternalAsset` keeps the storage surrogate id (the OAuth analogy: `sub` is translated to an internal id and the app works with it; `(source, external_id)` is a unique constraint, never a primary key) — on the sync write path it is transient (zero id) and storage assigns id and `synced_at`. Import creates the Asset together with its primary AssetLink in one transaction inside the repository; deleting an Asset deletes its links the same way.
