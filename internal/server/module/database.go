package module

import (
	"context"
	"database/sql"

	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/impl/database"
	"github.com/nikimonax/laretz/internal/server/config"
)

// depends on *config.Config
func DatabaseModule() fx.Option {
	return fx.Module(
		"database",
		fx.Provide(
			provideDatabase,
		),
		fx.Invoke(
			registerDatabaseClose,
			registerDatabaseMigrate,
		),
	)
}

func provideDatabase(cfg *config.Config) (*sql.DB, error) {
	return database.Connect(cfg.Data.DatabaseFile())
}

func registerDatabaseClose(
	lc fx.Lifecycle,
	db *sql.DB,
) {
	lc.Append(fx.StopHook(db.Close))
}

func registerDatabaseMigrate(
	lc fx.Lifecycle,
	db *sql.DB,
) {
	lc.Append(fx.StartHook(func(ctx context.Context) error {
		return database.Migrate(ctx, db)
	}))
}
