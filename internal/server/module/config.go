package module

import (
	"fmt"

	"github.com/go-logr/logr"
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/server/config"
)

// depends on *config.Config and logging module
func ConfigModule() fx.Option {
	return fx.Module(
		"config",
		fx.Invoke(registerStartupConfigLog),
	)
}

func registerStartupConfigLog(
	lc fx.Lifecycle,
	cfg *config.Config,
	logger logr.Logger,
) {
	lc.Append(fx.StartHook(func() {
		logger.Info(
			"config",
			"source", cfg.Meta.Source,
			"data", cfg.Data.Dir,
			"server", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		)
	}))
}
