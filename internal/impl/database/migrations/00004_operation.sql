-- +goose Up

CREATE TABLE IF NOT EXISTS operation (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    portfolio_id INTEGER NOT NULL REFERENCES portfolio (id) ON DELETE CASCADE,
    asset_id     INTEGER NOT NULL REFERENCES asset (id),
    type         TEXT NOT NULL CHECK (type IN ('buy', 'sell')),
    executed_at  DATETIME NOT NULL,
    quantity     TEXT NOT NULL,
    total_usd    TEXT NOT NULL,
    fee_amount   TEXT,
    fee_asset_id INTEGER REFERENCES asset (id),
    note         TEXT,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_operation_updated_at
AFTER UPDATE ON operation
FOR EACH ROW
BEGIN
    UPDATE operation
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = OLD.id;
END;
-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_operation_updated_at;
-- +goose StatementEnd

DROP TABLE IF EXISTS operation;