package plugin

import (
	"errors"
	"maps"
	"slices"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

var (
	ErrAssetProviderNotFound          = errors.New("asset provider not found")
	ErrAssetProviderAlreadyRegistered = errors.New("asset provider already registered")
)

type ExternalAssetProviderRegistry struct {
	providers map[domain.SourceID]interfaces.ExternalAssetProvider
}

// List implements [interfaces.ExternalAssetProviderRegistry].
func (registry *ExternalAssetProviderRegistry) List() []domain.SourceID {
	return slices.Sorted(maps.Keys(registry.providers))
}

// Find implements [interfaces.ExternalAssetProviderRegistry].
func (registry *ExternalAssetProviderRegistry) Find(sourceID domain.SourceID) (interfaces.ExternalAssetProvider, error) {
	provider, ok := registry.providers[sourceID]

	if !ok {
		return nil, ErrAssetProviderNotFound
	}

	return provider, nil
}

func (registry *ExternalAssetProviderRegistry) Register(sourceID domain.SourceID, provider interfaces.ExternalAssetProvider) error {
	_, exists := registry.providers[sourceID]

	if exists {
		return ErrAssetProviderAlreadyRegistered
	}

	registry.providers[sourceID] = provider
	return nil
}

var _ interfaces.ExternalAssetProviderRegistry = (*ExternalAssetProviderRegistry)(nil)

func NewExternalAssetProviderRegistry() *ExternalAssetProviderRegistry {
	return &ExternalAssetProviderRegistry{providers: make(map[domain.SourceID]interfaces.ExternalAssetProvider)}
}
