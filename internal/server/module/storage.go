package module

import (
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/impl/repository"
)

// depends on database module
func StorageModule() fx.Option {
	return fx.Module(
		"storage",
		fx.Provide(
			repository.NewSqlitePortfolioRepository,
			repository.NewSqliteAssetRepository,
			repository.NewSqliteExternalAssetRepository,
		),
	)
}
