package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/impl/database/generated"
)

const searchExternalAssetsLimit = 50

const searchExternalAssetsSQL = `
SELECT ea.id, ea.source, ea.external_id, ea.symbol, ea.name, ea.synced_at
FROM external_asset_fts(?)
JOIN external_asset ea ON ea.id = external_asset_fts.rowid
ORDER BY (lower(ea.symbol) = lower(?)) DESC, bm25(external_asset_fts)
LIMIT ?
`

type SqliteExternalAssetRepository struct {
	db      *sql.DB
	queries *generated.Queries
}

// Count implements [interfaces.ExternalAssetRepository].
func (repo *SqliteExternalAssetRepository) Count(ctx context.Context) (int64, error) {
	count, err := repo.queries.CountExternalAssets(ctx)

	if err != nil {
		return 0, fmt.Errorf("count external assets: %w", err)
	}

	return count, nil
}

// Get implements [interfaces.ExternalAssetRepository].
func (repo *SqliteExternalAssetRepository) Get(ctx context.Context, id domain.ExternalAssetID) (domain.ExternalAsset, error) {
	row, err := repo.queries.GetExternalAsset(ctx, id)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.ExternalAsset{}, app.ErrEntityNotFound
	}

	if err != nil {
		return domain.ExternalAsset{}, fmt.Errorf("get external asset %d: %w", id, err)
	}

	return toDomainExternalAsset(row), nil
}

// Sync implements [interfaces.ExternalAssetRepository].
func (repo *SqliteExternalAssetRepository) Sync(ctx context.Context, catalog []domain.ExternalAsset) error {
	return inTransaction(ctx, repo.db, repo.queries, func(queries *generated.Queries) error {
		for _, asset := range catalog {
			params := generated.UpsertExternalAssetParams{
				Source:     asset.SourceID,
				ExternalID: asset.ExternalID,
				Symbol:     asset.Symbol,
				Name:       asset.Name,
			}

			if err := queries.UpsertExternalAsset(ctx, params); err != nil {
				return fmt.Errorf("upsert external asset %s:%s: %w", asset.SourceID, asset.ExternalID, err)
			}
		}

		return nil
	})
}

// Search implements [interfaces.ExternalAssetRepository].
func (repo *SqliteExternalAssetRepository) Search(ctx context.Context, query string) ([]domain.ExternalAsset, error) {
	match := ftsMatchQuery(query)

	if match == "" {
		return nil, nil
	}

	rows, err := repo.db.QueryContext(
		ctx,
		searchExternalAssetsSQL,
		match,
		strings.TrimSpace(query),
		searchExternalAssetsLimit,
	)

	if err != nil {
		return nil, fmt.Errorf("search external assets: %w", err)
	}

	defer func() { _ = rows.Close() }()

	assets := make([]domain.ExternalAsset, 0, searchExternalAssetsLimit)

	for rows.Next() {
		var row generated.ExternalAsset

		if err := rows.Scan(&row.ID, &row.Source, &row.ExternalID, &row.Symbol, &row.Name, &row.SyncedAt); err != nil {
			return nil, fmt.Errorf("search external assets: scan: %w", err)
		}

		assets = append(assets, toDomainExternalAsset(row))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search external assets: %w", err)
	}

	return assets, nil
}

func NewSqliteExternalAssetRepository(db *sql.DB) interfaces.ExternalAssetRepository {
	return &SqliteExternalAssetRepository{db: db, queries: generated.New(db)}
}

func ftsMatchQuery(query string) string {
	parts := make([]string, 0, 8)

	for _, term := range strings.Fields(query) {
		term = strings.ReplaceAll(term, `"`, `""`)
		parts = append(parts, `"`+term+`"*`)
	}

	return strings.Join(parts, " ")
}

func toDomainExternalAsset(row generated.ExternalAsset) domain.ExternalAsset {
	return domain.ExternalAsset{
		ID:         row.ID,
		SourceID:   row.Source,
		ExternalID: row.ExternalID,
		Symbol:     row.Symbol,
		Name:       row.Name,
		SyncedAt:   row.SyncedAt,
	}
}
