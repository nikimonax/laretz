-- +goose Up

CREATE INDEX IF NOT EXISTS idx_operation_portfolio_asset
    ON operation (portfolio_id, asset_id);

-- +goose Down

DROP INDEX IF EXISTS idx_operation_portfolio_asset;
