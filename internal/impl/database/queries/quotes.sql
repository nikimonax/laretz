-- name: ListPricePoints :many
SELECT external_asset_id, ts_hour, price_usd, granularity
FROM price_point
WHERE external_asset_id = ?
ORDER BY ts_hour;

-- name: InsertPricePoint :exec
INSERT OR IGNORE INTO price_point (external_asset_id, ts_hour, price_usd, granularity)
VALUES (?, ?, ?, ?);
