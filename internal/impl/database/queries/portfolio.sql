-- name: CreatePortfolio :one
INSERT INTO portfolio (name)
VALUES (?)
RETURNING id, name, created_at, updated_at;

-- name: ListPortfolios :many
SELECT id, name, created_at, updated_at
FROM portfolio
ORDER BY created_at, id;

-- name: GetPortfolio :one
SELECT id, name, created_at, updated_at
FROM portfolio
WHERE id = ?;

-- name: RenamePortfolio :execrows
UPDATE portfolio
SET name = ?
WHERE id = ?;

-- name: DeletePortfolio :execrows
DELETE FROM portfolio
WHERE id = ?;
