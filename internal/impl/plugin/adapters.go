package plugin

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/dto"
	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/pkg/plugin"
)

func NewExternalAssetProviderAdapter(provider plugin.AssetProvider, pluginID PluginID) interfaces.ExternalAssetProvider {
	return interfaces.ExternalAssetProviderFunc(func(ctx context.Context) ([]dto.ExternalAsset, error) {
		assets, err := provider.Provide(ctx)

		if err != nil {
			return nil, err
		}

		externalAssets := make([]dto.ExternalAsset, 0, len(assets))
		for _, asset := range assets {
			externalAssets = append(externalAssets, dto.ExternalAsset{
				SourceID:   pluginID, // implicit casting plugin_id -> source_id
				ExternalID: asset.ID,
				Symbol:     asset.Symbol,
				Name:       asset.Name,
			})
		}

		return externalAssets, nil
	})
}
