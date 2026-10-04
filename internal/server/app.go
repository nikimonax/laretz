package server

import (
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/server/config"
	"github.com/nikimonax/laretz/internal/server/module"
	"github.com/nikimonax/laretz/internal/shared/logging"
)

func New(cfg *config.Config) (*fx.App, error) {
	if err := config.ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return fx.New(
		fx.Supply(cfg),
		module.LoggingModule(),
		module.DatabaseModule(),
		module.RepositoryModule(),
		module.UseCaseModule(),
		module.HTTPModule(),
		module.BrowserModule(),
		fx.WithLogger(logging.NewFxLogger),
	), nil
}
