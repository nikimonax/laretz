package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nikimonax/laretz/internal/app"
	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/impl/database/generated"
)

type SqliteAssetRepository struct {
	db      *sql.DB
	queries *generated.Queries
}

// Get implements [interfaces.AssetRepository].
func (repo *SqliteAssetRepository) Get(ctx context.Context, id domain.AssetID) (domain.Asset, error) {
	row, err := repo.queries.GetAsset(ctx, id)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Asset{}, app.ErrEntityNotFound
	}

	if err != nil {
		return domain.Asset{}, fmt.Errorf("get asset %d: %w", id, err)
	}

	return toDomainAsset(row), nil
}

// List implements [interfaces.AssetRepository].
func (repo *SqliteAssetRepository) List(ctx context.Context) ([]domain.Asset, error) {
	rows, err := repo.queries.ListAssets(ctx)

	if err != nil {
		return nil, fmt.Errorf("list assets: %w", err)
	}

	assets := make([]domain.Asset, 0, len(rows))

	for _, row := range rows {
		assets = append(assets, toDomainAsset(generated.GetAssetRow(row)))
	}

	return assets, nil
}

// ListByPortfolioID implements [interfaces.AssetRepository].
func (repo *SqliteAssetRepository) ListByPortfolioID(ctx context.Context, portfolioID domain.PortfolioID) ([]domain.Asset, error) {
	rows, err := repo.queries.ListPortfolioAssets(ctx, portfolioID)

	if err != nil {
		return nil, fmt.Errorf("list assets of portfolio %d: %w", portfolioID, err)
	}

	assets := make([]domain.Asset, 0, len(rows))

	for _, row := range rows {
		assets = append(assets, toDomainAsset(generated.GetAssetRow(row)))
	}

	return assets, nil
}

// Create implements [interfaces.AssetRepository].
func (repo *SqliteAssetRepository) Create(
	ctx context.Context,
	externalAssetID domain.ExternalAssetID,
	symbol string,
	name string,
) (domain.AssetID, error) {
	var assetID domain.AssetID

	err := inTransaction(ctx, repo.db, repo.queries, func(queries *generated.Queries) error {
		if _, err := queries.GetExternalAsset(ctx, externalAssetID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return app.ErrEntityNotFound
			}
			return fmt.Errorf("get external asset %d: %w", externalAssetID, err)
		}

		id, err := queries.InsertAsset(ctx, generated.InsertAssetParams{
			Symbol: symbol,
			Name:   name,
		})

		if err != nil {
			return fmt.Errorf("insert asset: %w", err)
		}

		assetID = id

		if err := queries.InsertAssetLink(ctx, generated.InsertAssetLinkParams{
			AssetID:         assetID,
			ExternalAssetID: externalAssetID,
		}); err != nil {
			return fmt.Errorf("insert asset link: %w", err)
		}

		return nil
	})

	return assetID, err
}

// Rename implements [interfaces.AssetRepository].
func (repo *SqliteAssetRepository) Rename(ctx context.Context, id domain.AssetID, symbol string, name string) error {
	rows, err := repo.queries.UpdateAssetFields(ctx, generated.UpdateAssetFieldsParams{
		Symbol: symbol,
		Name:   name,
		ID:     id,
	})

	if err != nil {
		return fmt.Errorf("rename asset %d: %w", id, err)
	}

	if rows == 0 {
		return app.ErrEntityNotFound
	}

	return nil
}

// Relink implements [interfaces.AssetRepository].
func (repo *SqliteAssetRepository) Relink(ctx context.Context, id domain.AssetID, externalAssetID domain.ExternalAssetID) error {
	return inTransaction(ctx, repo.db, repo.queries, func(queries *generated.Queries) error {
		currentSource, err := queries.GetAssetPrimaryLinkSource(ctx, id)

		if errors.Is(err, sql.ErrNoRows) {
			return app.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("get asset %d primary link: %w", id, err)
		}

		target, err := queries.GetExternalAsset(ctx, externalAssetID)

		if errors.Is(err, sql.ErrNoRows) {
			return app.ErrEntityNotFound
		}

		if err != nil {
			return fmt.Errorf("get external asset %d: %w", externalAssetID, err)
		}

		if currentSource != target.Source {
			return fmt.Errorf(
				"relink asset %d to %s:%s: %w: linked to %q",
				id, target.Source, target.ExternalID, app.ErrSourceMismatch, currentSource,
			)
		}

		rows, err := queries.UpdateAssetPrimaryLink(ctx, generated.UpdateAssetPrimaryLinkParams{
			ExternalAssetID: externalAssetID,
			AssetID:         id,
		})

		if err != nil {
			return fmt.Errorf("relink asset %d: %w", id, err)
		}

		if rows == 0 {
			return app.ErrEntityNotFound
		}

		return nil
	})
}

// Delete implements [interfaces.AssetRepository].
func (repo *SqliteAssetRepository) Delete(ctx context.Context, id domain.AssetID) error {
	return inTransaction(ctx, repo.db, repo.queries, func(queries *generated.Queries) error {
		if err := queries.DeleteAssetLinks(ctx, id); err != nil {
			return fmt.Errorf("delete asset %d links: %w", id, err)
		}

		rows, err := queries.DeleteAsset(ctx, id)

		if err != nil {
			return fmt.Errorf("delete asset %d: %w", id, err)
		}

		if rows == 0 {
			return app.ErrEntityNotFound
		}

		return nil
	})
}

func NewSqliteAssetRepository(db *sql.DB) interfaces.AssetRepository {
	return &SqliteAssetRepository{db: db, queries: generated.New(db)}
}

func toDomainAsset(row generated.GetAssetRow) domain.Asset {
	return domain.Asset{
		ID:        row.ID,
		Symbol:    row.Symbol,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,

		External: domain.ExternalAsset{
			ID:         row.ExternalAssetID,
			SourceID:   row.Source,
			ExternalID: row.ExternalID,
			Symbol:     row.ExternalSymbol,
			Name:       row.ExternalName,
			SyncedAt:   row.SyncedAt,
		},
	}
}
