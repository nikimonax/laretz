package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/impl/database/generated"
)

type SqliteQuoteRepository struct {
	db      *sql.DB
	queries *generated.Queries
}

// Coverage implements [interfaces.QuoteRepository].
func (repo *SqliteQuoteRepository) Coverage(ctx context.Context, externalAssetID domain.ExternalAssetID) ([]domain.PricePoint, error) {
	rows, err := repo.queries.ListPricePoints(ctx, externalAssetID)

	if err != nil {
		return nil, fmt.Errorf("list price points %d: %w", externalAssetID, err)
	}

	points := make([]domain.PricePoint, 0, len(rows))

	for _, row := range rows {
		points = append(points, domain.PricePoint{
			ExternalAssetID: row.ExternalAssetID,
			At:              row.TsHour,
			PriceUSD:        row.PriceUsd,
			Granularity:     domain.Granularity(row.Granularity),
		})
	}

	return points, nil
}

// Store implements [interfaces.QuoteRepository].
func (repo *SqliteQuoteRepository) Store(ctx context.Context, points []domain.PricePoint) error {
	return inTransaction(ctx, repo.db, repo.queries, func(queries *generated.Queries) error {
		for _, point := range points {
			params := generated.InsertPricePointParams{
				ExternalAssetID: point.ExternalAssetID,
				TsHour:          point.At,
				PriceUsd:        point.PriceUSD,
				Granularity:     string(point.Granularity),
			}

			if err := queries.InsertPricePoint(ctx, params); err != nil {
				return fmt.Errorf("insert price point %d@%s: %w", point.ExternalAssetID, point.At.Format(time.RFC3339), err)
			}
		}

		return nil
	})
}

func NewSqliteQuoteRepository(db *sql.DB) interfaces.QuoteRepository {
	return &SqliteQuoteRepository{db: db, queries: generated.New(db)}
}
