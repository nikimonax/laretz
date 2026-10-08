# Laretz

A single-user crypto portfolio tracker: it records Buy/Sell operations and values portfolios using prices from pluggable market data sources.

## Language

### Portfolios & operations

**Portfolio**:
A named container of Operations belonging to the user.
_Avoid_: account, wallet

**Operation**:
A Buy or a Sell of an Asset inside a Portfolio at a point in time.
_Avoid_: trade, transaction, deal

**Fee**:
An optional cost attached to an Operation, denominated in some Asset. It never affects Balances.
_Avoid_: commission

**Balance**:
The quantity of an Asset in a Portfolio, derived from its Operations. May be negative.
_Avoid_: holdings, amount

**Position**:
An Asset held in a Portfolio together with its Balance at a point in time.
_Avoid_: holding

### Assets & sources

**Asset**:
The asset the core domain tracks. Exists independently of any market data source.
_Avoid_: canonical asset, coin, token, instrument

**ExternalAsset**:
An asset as a MarketDataSource identifies it: the source's own id and naming for it.
_Avoid_: provider asset, source asset, listing

**AssetLink**:
An association between an Asset and an ExternalAsset. At most one per MarketDataSource; exactly one is primary and supplies prices for the Asset.
_Avoid_: mapping, binding

**Primary Link**:
The AssetLink through which prices for an Asset are resolved.
_Avoid_: default binding, main mapping

**MarketDataSource**:
A pluggable external service that provides a Catalog of ExternalAssets and their Prices.
_Avoid_: price provider, plugin, vendor, feed

**Integration**:
A MarketDataSource the user enabled by providing its api_key in the config file. The only source-facing word shown to users.
_Avoid_: connection, account, plugin

**Catalog**:
The full set of ExternalAssets a MarketDataSource supports.
_Avoid_: coin list, market listing

### Pricing

**Quote Currency**:
The single currency all money values are expressed in: USD.
_Avoid_: base currency, display currency

**PricePoint**:
A price of an ExternalAsset at an hour boundary, expressed in the Quote Currency. May be hourly or daily; gaps are legitimate.
_Avoid_: quote, candle, rate

**HoldingInterval**:
A maximal span of time during which the Balance of an Asset in a Portfolio is non-zero. Prices are required only inside HoldingIntervals.
_Avoid_: position period, holding period

**QuoteCoverage**:
The set of hours for which PricePoints already exist for an ExternalAsset.
_Avoid_: coverage, price history

**FetchQuotes**:
Loading the PricePoints required by HoldingIntervals but missing from QuoteCoverage. Existing PricePoints are never overwritten and gaps are never filled in storage.
_Avoid_: price sync, refresh, backfill

**Current Price**:
The latest PricePoint available for an Asset through its Primary Link. Shown only for Positions with a non-zero Balance.
_Avoid_: live price, spot price

**Stale Price**:
A displayed price carried forward from the latest available PricePoint. A display-time device only; flagged as not current.
_Avoid_: cached price, outdated price
