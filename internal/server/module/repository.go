package module

import (
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/impl/repository"
)

// depends on database module
func RepositoryModule() fx.Option {
	return fx.Module(
		"repository",
		fx.Provide(
			repository.NewSqlitePortfolioRepository,
		),
	)
}
