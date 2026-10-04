-- +goose Up

CREATE TABLE IF NOT EXISTS portfolio (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_portfolio_updated_at 
AFTER UPDATE ON portfolio
FOR EACH ROW
BEGIN
    UPDATE portfolio 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = OLD.id;
END;
-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_portfolio_updated_at;
-- +goose StatementEnd

DROP TABLE IF EXISTS portfolio;
