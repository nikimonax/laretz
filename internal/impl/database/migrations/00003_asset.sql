-- +goose Up

CREATE TABLE IF NOT EXISTS asset (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    symbol     TEXT NOT NULL,
    name       TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_asset_updated_at
AFTER UPDATE ON asset
FOR EACH ROW
BEGIN
    UPDATE asset
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = OLD.id;
END;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS asset_link (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id          INTEGER NOT NULL REFERENCES asset (id) ON DELETE CASCADE,
    external_asset_id INTEGER NOT NULL REFERENCES external_asset (id),
    is_primary        INTEGER NOT NULL CHECK (is_primary IN (0, 1)),
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (asset_id, external_asset_id)
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_asset_link_updated_at
AFTER UPDATE ON asset_link
FOR EACH ROW
BEGIN
    UPDATE asset_link
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = OLD.id;
END;
-- +goose StatementEnd

CREATE UNIQUE INDEX IF NOT EXISTS asset_link_single_primary
    ON asset_link (asset_id)
    WHERE is_primary = 1;

CREATE VIRTUAL TABLE IF NOT EXISTS external_asset_fts USING fts5 (
    symbol,
    name,
    external_id,
    content='external_asset',
    content_rowid='id'
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS external_asset_fts_insert
AFTER INSERT ON external_asset
FOR EACH ROW
BEGIN
    INSERT INTO external_asset_fts (rowid, symbol, name, external_id)
    VALUES (new.id, new.symbol, new.name, new.external_id);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS external_asset_fts_delete
AFTER DELETE ON external_asset
FOR EACH ROW
BEGIN
    INSERT INTO external_asset_fts (external_asset_fts, rowid, symbol, name, external_id)
    VALUES ('delete', old.id, old.symbol, old.name, old.external_id);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS external_asset_fts_update
AFTER UPDATE ON external_asset
FOR EACH ROW
BEGIN
    INSERT INTO external_asset_fts (external_asset_fts, rowid, symbol, name, external_id)
    VALUES ('delete', old.id, old.symbol, old.name, old.external_id);
    INSERT INTO external_asset_fts (rowid, symbol, name, external_id)
    VALUES (new.id, new.symbol, new.name, new.external_id);
END;
-- +goose StatementEnd


-- +goose Down

-- +goose StatementBegin
DROP TRIGGER IF EXISTS external_asset_fts_insert;
DROP TRIGGER IF EXISTS external_asset_fts_delete;
DROP TRIGGER IF EXISTS external_asset_fts_update;
DROP TRIGGER IF EXISTS update_asset_link_updated_at;
DROP TRIGGER IF EXISTS update_asset_updated_at;
-- +goose StatementEnd

DROP TABLE IF EXISTS external_asset_fts;
DROP INDEX IF EXISTS asset_link_single_primary;
DROP TABLE IF EXISTS asset_link;
DROP TABLE IF EXISTS asset;
