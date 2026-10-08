package coingecko

import (
	"context"

	"github.com/nikimonax/laretz/pkg/plugin"
)

type AssetProvider struct {
	client *Client
}

// Provide implements [plugin.AssetProvider].
func (provider *AssetProvider) Provide(ctx context.Context) ([]plugin.Asset, error) {
	coins, err := provider.client.CoinsList(ctx)

	if err != nil {
		return nil, err
	}

	assets := make([]plugin.Asset, 0, len(coins))

	for _, coin := range coins {
		assets = append(assets, plugin.Asset{
			ID:     coin.ID,
			Symbol: coin.Symbol,
			Name:   coin.Name,
		})
	}

	return assets, nil
}

func NewAssetProvider(client *Client) plugin.AssetProvider {
	return &AssetProvider{client: client}
}
