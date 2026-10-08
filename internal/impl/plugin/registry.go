package plugin

import (
	"maps"
	"slices"

	"github.com/nikimonax/laretz/pkg/plugin"
)

type Registry struct {
	plugins map[PluginID]plugin.Plugin
}

func (r *Registry) Register(p plugin.Plugin) error {
	id := p.ID()

	_, exists := r.plugins[id]

	if exists {
		return NewErrPluginAlreadyRegistered(id)
	}

	r.plugins[id] = p
	return nil
}

func (r *Registry) Get(pluginID PluginID) (plugin.Plugin, error) {
	p, ok := r.plugins[pluginID]

	if !ok {
		return nil, NewErrPluginNotFound(pluginID)
	}

	return p, nil
}

func (r *Registry) GetAll() []plugin.Plugin {
	return slices.Collect(maps.Values(r.plugins))
}

func NewRegistry() *Registry {
	return &Registry{plugins: make(map[PluginID]plugin.Plugin)}
}

var registry *Registry

func init() {
	registry = NewRegistry()
}

func DefaultRegistry() *Registry {
	return registry
}
