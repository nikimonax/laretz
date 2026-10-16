package module

import (
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/lib/browser"
	"github.com/nikimonax/laretz/internal/server/config"
)

// depends on *config.Config
func BrowserModule() fx.Option {
	return fx.Module(
		"browser",
		fx.Provide(provideOpener),
		fx.Invoke(registerBrowserOpen),
	)
}

func provideOpener() browser.Opener {
	return browser.New()
}

func registerBrowserOpen(
	lc fx.Lifecycle,
	opener browser.Opener,
	cfg *config.Config,
) {
	if !cfg.Server.OpenBrowser {
		return
	}

	lc.Append(fx.StartHook(func() error {
		return opener.Open(cfg.Server.BrowserOpenURL())
	}))
}
