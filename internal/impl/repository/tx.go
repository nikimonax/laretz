package repository

import (
	"context"
	"database/sql"
	"fmt"

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
