package handler

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nikimonax/laretz/internal/app/commands"
	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/impl/database/generated"
)

type SqliteSyncExternalAssetsCommandHandler struct {
	db      *sql.DB
	queries *generated.Queries
}

// Handle implements [interfaces.CommandHandler].
func (handler *SqliteSyncExternalAssetsCommandHandler) Handle(
	ctx context.Context,
	cmd commands.SyncExternalAssetsCommand,
) error {
	tx, err := handler.db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("upsert external assets: begin: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	queries := handler.queries.WithTx(tx)

	for _, asset := range cmd.Assets {
		params := generated.UpsertExternalAssetParams{
			Source:     asset.SourceID,
			ExternalID: asset.ExternalID,
			Symbol:     asset.Symbol,
			Name:       asset.Name,
		}

		if err := queries.UpsertExternalAsset(ctx, params); err != nil {
			return fmt.Errorf("upsert external asset: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("upsert external assets: commit: %w", err)
	}

	return nil
}

func NewSqliteSyncExternalAssetsCommandHandler(
	db *sql.DB,
) interfaces.CommandHandler[commands.SyncExternalAssetsCommand] {
	return &SqliteSyncExternalAssetsCommandHandler{db: db, queries: generated.New(db)}
}
