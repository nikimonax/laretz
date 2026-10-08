package module

import (
	"maps"

	"github.com/go-logr/logr"
	"go.uber.org/fx"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/impl/plugin"
	"github.com/nikimonax/laretz/internal/server/config"
)

// depends on logging & *plugin.Registry
func PluginModule() fx.Option {
	return fx.Module(
		"plugin",
		fx.Provide(
			provideExternalAssetProviderRegistry,
			provideExternalAssetProviderRegistryInterface,
			providePluginFacade,
		),
		fx.Invoke(applyPlugins),
	)
}

func provideExternalAssetProviderRegistry() *plugin.ExternalAssetProviderRegistry {
	return plugin.NewExternalAssetProviderRegistry()
}

func provideExternalAssetProviderRegistryInterface(
	registry *plugin.ExternalAssetProviderRegistry,
) interfaces.ExternalAssetProviderRegistry {
	return registry
}

func providePluginFacade(
	cfg *config.Config,
	logger logr.Logger,
	registry *plugin.ExternalAssetProviderRegistry,
) *plugin.Facade {
	configs := make(map[plugin.PluginID]plugin.Unmarshable, len(cfg.Plugins))

	for pluginID, raw := range maps.All(cfg.Plugins) {
		configs[pluginID] = &raw
	}

	return plugin.NewFacade(configs, logger, *registry)
}

func applyPlugins(facade *plugin.Facade, registry *plugin.Registry) error {
	for _, plugin := range registry.GetAll() {
		if err := plugin.Init(facade.WithPlugin(plugin.ID())); err != nil {
			return err
		}
	}
	return nil
}
