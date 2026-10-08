package server

import (
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/impl/plugin"
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
		fx.Supply(plugin.DefaultRegistry()),
		module.LoggingModule(),
		module.PluginModule(),
		module.DatabaseModule(),
		module.StorageModule(),
		module.UseCaseModule(),
		module.HTTPModule(),
		module.BrowserModule(),
		fx.WithLogger(logging.NewFxLogger),
	), nil
}
