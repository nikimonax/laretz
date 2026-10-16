-- +goose Up

CREATE TABLE IF NOT EXISTS external_asset (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    source      TEXT NOT NULL,
    external_id TEXT NOT NULL,
    symbol      TEXT NOT NULL,
    name        TEXT NOT NULL,
    synced_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (source, external_id)
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_external_asset_synced_at
AFTER UPDATE ON external_asset
FOR EACH ROW
BEGIN
    UPDATE external_asset
    SET synced_at = CURRENT_TIMESTAMP
    WHERE id = OLD.id;
END;
-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_external_asset_synced_at;
-- +goose StatementEnd

DROP TABLE IF EXISTS external_asset;
