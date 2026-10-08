package coingecko

import (
	"net/http"
	"time"

	"github.com/nikimonax/laretz/pkg/lib/httplog"
	"github.com/nikimonax/laretz/pkg/plugin"
)

const PluginID = "coingecko"

type Plugin struct{}

// ID implements [plugin.Plugin].
func (p *Plugin) ID() string {
	return PluginID
}

// Init implements [plugin.Plugin].
func (p *Plugin) Init(app plugin.App) error {
	var cfg Config

	if err := app.GetConfig(&cfg); err != nil {
		return err
	}

	if cfg.APIKey == "" {
		app.Logger().Info("'api_key' not provided, configure coingecko plugin")
		return nil
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}

	transport := httplog.NewLoggingRoundTripper(http.DefaultTransport, app.Logger())
	httpClient := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	coingeckoClient := NewClient(httpClient, cfg.BaseURL, cfg.APIKey)
	provider := NewAssetProvider(coingeckoClient)

	return app.RegisterAssetProvider(provider)
}

func NewPlugin() plugin.Plugin {
	return new(Plugin)
}
