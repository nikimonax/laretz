package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/internal/impl/database/generated"
)

func inTransaction(
	ctx context.Context,
	db *sql.DB,
	queries *generated.Queries,
	fn func(queries *generated.Queries) error,
) error {
	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	if err := fn(queries.WithTx(tx)); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func toNullInt64(value *domain.AssetID) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}

	return sql.NullInt64{Int64: *value, Valid: true}
}

func fromNullInt64(value sql.NullInt64) *domain.AssetID {
	if !value.Valid {
		return nil
	}

	assetID := domain.AssetID(value.Int64)

	return &assetID
}

func toNullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{}
	}

	return sql.NullString{String: value, Valid: true}
}
