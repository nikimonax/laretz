# Canonical Assets are separated from Source Assets

Market data providers (CoinGecko, CoinMarketCap, …) identify the same real-world asset with different ids and names, and the project wants a plugin architecture over market data sources. The core tracks its own `Asset`; each `MarketDataSource` plugin owns its catalog of `ExternalAsset`s; an `AssetLink` (at most one per source, one primary) connects them, and prices are resolved through the primary link.

## Considered options

Storing provider ids directly on the asset — rejected: it locks the core to one provider's vocabulary and makes adding a second provider or re-linking painful.

## Consequences

Re-linking an Asset to a different ExternalAsset is a normal operation. Adding a new provider never touches core entities — only a new plugin behind the `AssetCatalog`/`QuoteFeed` ports.
