-- name: UpsertExternalAsset :exec
INSERT INTO external_asset (source, external_id, symbol, name)
VALUES (?, ?, ?, ?)
ON CONFLICT (source, external_id) DO UPDATE
SET symbol = excluded.symbol,
    name = excluded.name,
    synced_at = excluded.synced_at;
