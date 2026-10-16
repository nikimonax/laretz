-- name: InsertOperation :execrows
INSERT INTO operation (
    portfolio_id,
    asset_id,
    type,
    executed_at,
    quantity,
    total_usd,
    fee_amount,
    fee_asset_id,
    note
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetOperation :one
SELECT
    id,
    portfolio_id,
    asset_id,
    type,
    executed_at,
    quantity,
    total_usd,
    fee_amount,
    fee_asset_id,
    note,
    created_at,
    updated_at
FROM operation
WHERE id = ?;

-- name: ListPortfolioOperations :many
SELECT
    id,
    portfolio_id,
    asset_id,
    type,
    executed_at,
    quantity,
    total_usd,
    fee_amount,
    fee_asset_id,
    note,
    created_at,
    updated_at
FROM operation
WHERE portfolio_id = ?
ORDER BY executed_at ASC, id ASC;

-- name: ListPortfolioOperationsPaged :many
SELECT
    id,
    portfolio_id,
    asset_id,
    type,
    executed_at,
    quantity,
    total_usd,
    fee_amount,
    fee_asset_id,
    note,
    created_at,
    updated_at
FROM operation
WHERE portfolio_id = ?
ORDER BY executed_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountOperations :one
SELECT COUNT(*)
FROM operation
WHERE portfolio_id = ?;

-- name: UpdateOperation :execrows
UPDATE operation
SET
    portfolio_id = ?,
    asset_id = ?,
    type = ?,
    executed_at = ?,
    quantity = ?,
    total_usd = ?,
    fee_amount = ?,
    fee_asset_id = ?,
    note = ?
WHERE id = ?;

-- name: DeleteOperation :execrows
DELETE FROM operation
WHERE id = ?;