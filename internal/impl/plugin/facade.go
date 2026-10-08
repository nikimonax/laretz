package plugin

import (
	"github.com/go-logr/logr"

	"github.com/nikimonax/laretz/pkg/plugin"
)

type Unmarshable interface {
	Unmarshal(v any) error
}

type Facade struct {
	configs map[PluginID]Unmarshable
	logger  logr.Logger

	providers ExternalAssetProviderRegistry
}

type BoundFacade struct {
	Facade

	pluginID PluginID
}

func (app *Facade) WithPlugin(pluginID PluginID) *BoundFacade {
	return &BoundFacade{Facade: *app, pluginID: pluginID}
}

// Logger implements [plugin.App].
func (app *BoundFacade) Logger() logr.Logger {
	return app.logger.WithValues("plugin", app.pluginID)
}

// GetConfig implements [plugin.App].
func (app *BoundFacade) GetConfig(cfg any) error {
	raw, ok := app.configs[app.pluginID]

	if ok {
		return raw.Unmarshal(cfg)
	}

	return nil
}

// SetConfig implements [plugin.App].
func (app *BoundFacade) SetConfig(_ any) error {
	// TODO
	panic("unimplemented")
}

// RegisterAssetProvider implements [plugin.App].
func (app *BoundFacade) RegisterAssetProvider(provider plugin.AssetProvider) error {
	return app.providers.Register(app.pluginID, NewExternalAssetProviderAdapter(provider, app.pluginID))
}

var _ plugin.App = (*BoundFacade)(nil)

func NewFacade(
	configs map[PluginID]Unmarshable,
	logger logr.Logger,
	providers ExternalAssetProviderRegistry,
) *Facade {
	return &Facade{configs: configs, logger: logger, providers: providers}
}
