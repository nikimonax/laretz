-- +goose Up

CREATE TABLE IF NOT EXISTS price_point (
    external_asset_id INTEGER NOT NULL REFERENCES external_asset (id) ON DELETE CASCADE,
    ts_hour           DATETIME NOT NULL,
    price_usd         TEXT NOT NULL,
    granularity       TEXT NOT NULL CHECK (granularity IN ('hour', 'day')),
    PRIMARY KEY (external_asset_id, ts_hour)
);

-- +goose Down

DROP TABLE IF EXISTS price_point;
