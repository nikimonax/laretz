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

type SqliteOperationRepository struct {
	queries *generated.Queries
}

// Get implements [interfaces.OperationRepository].
func (repo *SqliteOperationRepository) Get(ctx context.Context, id domain.OperationID) (domain.Operation, error) {
	row, err := repo.queries.GetOperation(ctx, id)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.Operation{}, app.ErrEntityNotFound
	}

	if err != nil {
		return domain.Operation{}, fmt.Errorf("get operation %d: %w", id, err)
	}

	return toDomainOperation(row), nil
}

// List implements [interfaces.OperationRepository].
func (repo *SqliteOperationRepository) List(ctx context.Context, portfolioID domain.PortfolioID) ([]domain.Operation, error) {
	rows, err := repo.queries.ListPortfolioOperations(ctx, portfolioID)

	if err != nil {
		return nil, fmt.Errorf("list all operations of portfolio %d: %w", portfolioID, err)
	}

	operations := make([]domain.Operation, 0, len(rows))

	for _, row := range rows {
		operations = append(operations, toDomainOperation(row))
	}

	return operations, nil
}

// ListPaged implements [interfaces.OperationRepository].
func (repo *SqliteOperationRepository) ListPaged(ctx context.Context, portfolioID domain.PortfolioID, limit, offset int) ([]domain.Operation, error) {
	rows, err := repo.queries.ListPortfolioOperationsPaged(ctx, generated.ListPortfolioOperationsPagedParams{
		PortfolioID: portfolioID,
		Limit:       int64(limit),
		Offset:      int64(offset),
	})

	if err != nil {
		return nil, fmt.Errorf("list operations of portfolio %d: %w", portfolioID, err)
	}

	operations := make([]domain.Operation, 0, len(rows))

	for _, row := range rows {
		operations = append(operations, toDomainOperation(row))
	}

	return operations, nil
}

// Count implements [interfaces.OperationRepository].
func (repo *SqliteOperationRepository) Count(ctx context.Context, portfolioID domain.PortfolioID) (int64, error) {
	count, err := repo.queries.CountOperations(ctx, portfolioID)

	if err != nil {
		return 0, fmt.Errorf("count operations of portfolio %d: %w", portfolioID, err)
	}

	return count, nil
}

// Create implements [interfaces.OperationRepository].
func (repo *SqliteOperationRepository) Create(ctx context.Context, operation domain.Operation) error {
	rows, err := repo.queries.InsertOperation(ctx, generated.InsertOperationParams{
		PortfolioID: operation.PortfolioID,
		AssetID:     operation.AssetID,
		Type:        string(operation.Type),
		ExecutedAt:  operation.ExecutedAt.UTC(),
		Quantity:    operation.Quantity,
		TotalUsd:    operation.TotalUSD,
		FeeAmount:   operation.FeeAmount,
		FeeAssetID:  toNullInt64(operation.FeeAssetID),
		Note:        toNullString(operation.Note),
	})

	if err != nil {
		return fmt.Errorf("create operation: %w", err)
	}

	if rows == 0 {
		return app.ErrEntityNotFound
	}

	return nil
}

// Update implements [interfaces.OperationRepository].
func (repo *SqliteOperationRepository) Update(ctx context.Context, operation domain.Operation) error {
	rows, err := repo.queries.UpdateOperation(ctx, generated.UpdateOperationParams{
		PortfolioID: operation.PortfolioID,
		AssetID:     operation.AssetID,
		Type:        string(operation.Type),
		ExecutedAt:  operation.ExecutedAt.UTC(),
		Quantity:    operation.Quantity,
		TotalUsd:    operation.TotalUSD,
		FeeAmount:   operation.FeeAmount,
		FeeAssetID:  toNullInt64(operation.FeeAssetID),
		Note:        toNullString(operation.Note),
		ID:          operation.ID,
	})

	if err != nil {
		return fmt.Errorf("update operation %d: %w", operation.ID, err)
	}

	if rows == 0 {
		return app.ErrEntityNotFound
	}

	return nil
}

// Delete implements [interfaces.OperationRepository].
func (repo *SqliteOperationRepository) Delete(ctx context.Context, id domain.OperationID) error {
	rows, err := repo.queries.DeleteOperation(ctx, id)

	if err != nil {
		return fmt.Errorf("delete operation %d: %w", id, err)
	}

	if rows == 0 {
		return app.ErrEntityNotFound
	}

	return nil
}

func NewSqliteOperationRepository(db *sql.DB) interfaces.OperationRepository {
	return &SqliteOperationRepository{queries: generated.New(db)}
}

func toDomainOperation(row generated.Operation) domain.Operation {
	return domain.Operation{
		ID:          row.ID,
		PortfolioID: row.PortfolioID,
		AssetID:     row.AssetID,
		Type:        domain.OperationType(row.Type),
		ExecutedAt:  row.ExecutedAt,
		Quantity:    row.Quantity,
		TotalUSD:    row.TotalUsd,
		FeeAmount:   row.FeeAmount,
		FeeAssetID:  fromNullInt64(row.FeeAssetID),
		Note:        row.Note.String,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
