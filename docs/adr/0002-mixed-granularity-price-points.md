# PricePoints store only what the source actually returned

The CoinGecko demo key serves hourly data only for ranges up to ~90 days, daily beyond that, and nothing older than ~365 days. PricePoints are written exactly as the source returned them: hourly slots, or daily points placed at the day's start hour; gaps are legitimate; existing points are never overwritten; no automatic rule ever synthesizes data into the DB. Forward-fill happens only at display time (charts, valuations) and is flagged as a Stale Price.

## Constraints not visible in the code

Price history depth and granularity are bounded by the free tier of the currently only Integration (CoinGecko demo key): hourly ≤ 90 days, daily ≤ ~365 days, nothing older.

## Consequences

Long-held assets get daily-granularity histories with gaps. Charts show flat carried segments across gaps. Portfolio valuation may report "part of assets without price".
