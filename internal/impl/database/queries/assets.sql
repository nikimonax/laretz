-- name: UpsertExternalAsset :exec
INSERT INTO external_asset (source, external_id, symbol, name)
VALUES (?, ?, ?, ?)
ON CONFLICT (source, external_id) DO UPDATE
SET symbol = excluded.symbol,
    name = excluded.name,
    synced_at = excluded.synced_at;

-- name: CountExternalAssets :one
SELECT COUNT(*) FROM external_asset;

-- name: GetExternalAsset :one
SELECT id, source, external_id, symbol, name, synced_at
FROM external_asset
WHERE id = ?;

-- name: InsertAsset :execlastid
INSERT INTO asset (symbol, name)
VALUES (?, ?);

-- name: InsertAssetLink :exec
INSERT INTO asset_link (asset_id, external_asset_id, is_primary)
VALUES (?, ?, 1);

-- name: GetAsset :one
SELECT 
    a.id,
    a.symbol,
    a.name,
    a.created_at,
    a.updated_at,
    ea.id AS external_asset_id,
    ea.source,
    ea.external_id,
    ea.symbol AS external_symbol,
    ea.name AS external_name,
    ea.synced_at
FROM asset a
JOIN asset_link l ON l.asset_id = a.id AND l.is_primary = 1
JOIN external_asset ea ON ea.id = l.external_asset_id
WHERE a.id = ?;

-- name: ListAssets :many
SELECT 
    a.id,
    a.symbol,
    a.name,
    a.created_at,
    a.updated_at,
    ea.id AS external_asset_id,
    ea.source,
    ea.external_id,
    ea.symbol AS external_symbol,
    ea.name AS external_name,
    ea.synced_at
FROM asset a
JOIN asset_link l ON l.asset_id = a.id AND l.is_primary = 1
JOIN external_asset ea ON ea.id = l.external_asset_id
ORDER BY a.symbol, a.id;

-- name: ListPortfolioAssets :many
SELECT
    a.id,
    a.symbol,
    a.name,
    a.created_at,
    a.updated_at,
    ea.id AS external_asset_id,
    ea.source,
    ea.external_id,
    ea.symbol AS external_symbol,
    ea.name AS external_name,
    ea.synced_at
FROM asset a
JOIN asset_link l ON l.asset_id = a.id AND l.is_primary = 1
JOIN external_asset ea ON ea.id = l.external_asset_id
WHERE EXISTS (
    SELECT 1
    FROM operation o
    WHERE o.asset_id = a.id AND o.portfolio_id = ?
)
ORDER BY a.symbol, a.id;

-- name: UpdateAssetFields :execrows
UPDATE asset
SET symbol = ?, name = ?
WHERE id = ?;

-- name: GetAssetPrimaryLinkSource :one
SELECT ea.source
FROM asset_link l
JOIN external_asset ea ON ea.id = l.external_asset_id
WHERE l.asset_id = ? AND l.is_primary = 1;

-- name: UpdateAssetPrimaryLink :execrows
UPDATE asset_link
SET external_asset_id = ?
WHERE asset_id = ? AND is_primary = 1;

-- name: DeleteAssetLinks :exec
DELETE FROM asset_link
WHERE asset_id = ?;

-- name: DeleteAsset :execrows
DELETE FROM asset
WHERE id = ?;
