package usecase

import (
	"context"

	"github.com/nikimonax/laretz/internal/app/interfaces"
	"github.com/nikimonax/laretz/internal/domain"
)

// priceHistories resolves the PriceHistory of ExternalAssets through the
// QuoteRepository, loading each ExternalAsset's PricePoints once.
type priceHistories struct {
	quotes interfaces.QuoteRepository
	cache  map[domain.ExternalAssetID]domain.PriceHistory
}

func newPriceHistories(quotes interfaces.QuoteRepository) *priceHistories {
	return &priceHistories{quotes: quotes, cache: make(map[domain.ExternalAssetID]domain.PriceHistory)}
}

func (histories *priceHistories) of(ctx context.Context, externalAssetID domain.ExternalAssetID) (domain.PriceHistory, error) {
	if history, ok := histories.cache[externalAssetID]; ok {
		return history, nil
	}

	points, err := histories.quotes.Coverage(ctx, externalAssetID)

	if err != nil {
		return nil, err
	}

	history := domain.PriceHistory(points)
	histories.cache[externalAssetID] = history

	return history, nil
}

// buildValuation assembles the Valuation of one set of Operations: the
// Operations are grouped per Asset and paired with the Asset record and the
// PriceHistory of its primary ExternalAsset. Operations of unknown Assets
// are skipped.
func buildValuation(
	ctx context.Context,
	histories *priceHistories,
	operations []domain.Operation,
	assets []domain.Asset,
) (*domain.Valuation, error) {
	assetByID := make(map[domain.AssetID]domain.Asset, len(assets))

	for _, asset := range assets {
		assetByID[asset.ID] = asset
	}

	operationsByAsset := make(map[domain.AssetID][]domain.Operation)

	for _, operation := range operations {
		operationsByAsset[operation.AssetID] = append(operationsByAsset[operation.AssetID], operation)
	}

	valued := make([]domain.ValuedAsset, 0, len(operationsByAsset))

	for assetID, assetOperations := range operationsByAsset {
		asset, ok := assetByID[assetID]

		if !ok {
			continue
		}

		history, err := histories.of(ctx, asset.External.ID)

		if err != nil {
			return nil, err
		}

		valued = append(valued, domain.ValuedAsset{
			Asset:    asset,
			Timeline: domain.NewBalanceTimeline(assetOperations),
			Prices:   history,
		})
	}

	return domain.NewValuation(valued), nil
}
