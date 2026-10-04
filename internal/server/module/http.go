package module

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/server/config"
	"github.com/nikimonax/laretz/internal/server/handler"
)

func HTTPModule() fx.Option {
	return fx.Module(
		"http",
		fx.Provide(
			handler.New,
			provideHTTPServer,
		),
		fx.Invoke(
			registerHTTPServerLifecycle,
		),
	)
}

func provideHTTPServer(router chi.Router) *http.Server {
	return &http.Server{
		Handler: router,
	}
}

func registerHTTPServerLifecycle(
	lc fx.Lifecycle,
	cfg *config.Config,
	srv *http.Server,
) {
	errs := make(chan error, 1)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lc := net.ListenConfig{}

			ln, err := lc.Listen(ctx, "tcp", cfg.Server.Listen())
			if err != nil {
				return err
			}

			go func() {
				err := srv.Serve(ln)

				if err != nil && err != http.ErrServerClosed {
					errs <- err
				}

				close(errs)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			shutdownErr := srv.Shutdown(ctx)

			select {
			case <-ctx.Done():
				return shutdownErr
			case serveErr := <-errs:
				return errors.Join(shutdownErr, serveErr)
			}
		},
	})
}
