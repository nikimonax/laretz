package plugin

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
	"github.com/nikimonax/laretz/pkg/plugin"
)

func NewExternalAssetProviderAdapter(provider plugin.AssetProvider, pluginID PluginID) interfaces.ExternalAssetProvider {
	return interfaces.ExternalAssetProviderFunc(func(ctx context.Context) ([]domain.ExternalAsset, error) {
		assets, err := provider.Provide(ctx)

		if err != nil {
			return nil, err
		}

		externalAssets := make([]domain.ExternalAsset, 0, len(assets))
		for _, asset := range assets {
			externalAssets = append(externalAssets, domain.ExternalAsset{
				SourceID:   pluginID, // implicit casting plugin_id -> source_id
				ExternalID: asset.ID,
				Symbol:     asset.Symbol,
				Name:       asset.Name,
			})
		}

		return externalAssets, nil
	})
}

func NewQuoteFeedAdapter(feed plugin.QuoteProvider, pluginID PluginID) interfaces.QuoteFeed {
	return interfaces.QuoteFeedFunc(func(ctx context.Context, ranges []domain.QuoteRange) ([]domain.FetchedQuote, error) {
		pluginRanges := make([]plugin.QuoteRange, len(ranges))

		for i, quoteRange := range ranges {
			pluginRanges[i] = plugin.QuoteRange{
				ExternalID: quoteRange.ExternalID,
				From:       quoteRange.From,
				To:         quoteRange.To,
			}
		}

		quotes, err := feed.Fetch(ctx, pluginRanges)

		if err != nil {
			return nil, err
		}

		fetched := make([]domain.FetchedQuote, 0, len(quotes))

		for _, quote := range quotes {
			price, err := decimal.NewFromString(quote.Price)

			if err != nil {
				return nil, err
			}

			fetched = append(fetched, domain.FetchedQuote{
				SourceID:    pluginID, // implicit casting plugin_id -> source_id
				ExternalID:  quote.ExternalID,
				At:          quote.At,
				PriceUSD:    price,
				Granularity: domain.Granularity(quote.Granularity),
			})
		}

		return fetched, nil
	})
}
