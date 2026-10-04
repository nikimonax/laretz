package module

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nikimonax/laretz/internal/server/config"
	"github.com/nikimonax/laretz/internal/shared/logging"
)

// depends on *config.Config
func LoggingModule() fx.Option {
	return fx.Module(
		"loggin",
		fx.Provide(
			provideZapLogger,
			provideSugaredLogger,
			provideGooseLogger,
			provideLogrLogger,
		),
		fx.Invoke(
			registerGooseLogger,
			registerLoggerSync,
		),
	)
}

func provideZapLogger(cfg *config.Config) *zap.Logger {
	return logging.NewZapLogger(cfg.Log.Env, cfg.Log.Level)
}

func provideSugaredLogger(logger *zap.Logger) *zap.SugaredLogger {
	return logger.Sugar()
}

func provideLogrLogger(logger *zap.Logger) logr.Logger {
	return zapr.NewLogger(logger)
}

func provideGooseLogger(sugar *zap.SugaredLogger) goose.Logger {
	return logging.NewGooseLoggerAdapter(sugar)
}

func registerGooseLogger(lc fx.Lifecycle, logger goose.Logger) {
	lc.Append(fx.StartHook(func() { goose.SetLogger(logger) }))
}

func registerLoggerSync(lc fx.Lifecycle, logger *zap.Logger) {
	lc.Append(fx.StopHook(func() { _ = logger.Sync() }))
}
