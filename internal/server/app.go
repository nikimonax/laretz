package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nikimonax/laretz/internal/app/ports"
	"github.com/nikimonax/laretz/internal/lib/browser"
	"github.com/nikimonax/laretz/internal/server/handler"
)

type RunOptions struct {
	Port        int
	OpenBrowser bool
}

func Run(opts RunOptions) error {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = zapLogger.Sync() }()
	logger := zapr.NewLogger(zapLogger)

	var opener browser.Opener = browser.OpenerFunc(func(string) error { return nil })
	if opts.OpenBrowser {
		opener = browser.New()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	laretz := NewApp(Options{Port: opts.Port, Opener: opener, Logger: logger})
	if err := laretz.Err(); err != nil {
		logger.Error(err, "laretz failed to build")
		return err
	}
	if err := laretz.Start(ctx); err != nil {
		logger.Error(err, "laretz failed to start")
		return err
	}

	<-ctx.Done()
	logger.Info("shutting down")

	if err := laretz.Stop(context.Background()); err != nil {
		logger.Error(err, "laretz failed to stop")
		return err
	}
	return nil
}

type Options struct {
	Port   int
	Opener browser.Opener
	Logger logr.Logger
}

func NewApp(opts Options) *fx.App {
	return fx.New(
		fx.Supply(opts),
		fx.Provide(provideLogger, ports.NewSystemClock, handler.New),
		fx.Invoke(registerServer),
	)
}

func provideLogger(opts Options) logr.Logger {
	return opts.Logger
}

type serverParams struct {
	fx.In

	Opts    Options
	Handler http.Handler
}

func registerServer(lc fx.Lifecycle, p serverParams) {
	srv := &http.Server{
		Handler:           p.Handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var listenConfig net.ListenConfig
			listener, err := listenConfig.Listen(ctx, "tcp", fmt.Sprintf("127.0.0.1:%d", p.Opts.Port))
			if err != nil {
				return fmt.Errorf("listen on 127.0.0.1:%d: %w", p.Opts.Port, err)
			}
			appURL := "http://" + listener.Addr().String()
			p.Opts.Logger.Info("laretz is running", "url", appURL)

			go func() {
				if serveErr := srv.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
					p.Opts.Logger.Error(serveErr, "http server failed")
				}
			}()

			if openErr := p.Opts.Opener.Open(appURL); openErr != nil {
				p.Opts.Logger.Error(openErr, "opening browser (best effort failed)")
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
